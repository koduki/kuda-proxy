package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
	"google.golang.org/api/workflowexecutions/v1"

	"kuda/cmd/config"
	"kuda/cmd/middleware"

	"github.com/labstack/echo/v4"
)

type (
	Response struct {
		Message string `json:"message"`
		Date    string `json:"date"`
	}
)

var Logger *log.Logger

var (
	optPort = flag.Int("p", 8080, "port number")
)

func main() {
	Logger = log.New(os.Stdout, "kuda:", log.LstdFlags)

	ParseArgs()
	e := echo.New()
	Route(e)
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(*optPort)))
}

func ParseArgs() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: hwrap [flags] command\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

func NewClient(TargetURL string, UseGoogleJWT bool) (*http.Client, error) {
	if UseGoogleJWT {
		ctx := context.Background()
		return idtoken.NewClient(ctx, TargetURL)
	} else {
		return &http.Client{
			Timeout: time.Second * 10,
		}, nil
	}
}

// Route is the main routing function for the server
func Route(e *echo.Echo) {
	e.HTTPErrorHandler = middleware.HTTPErrorHandler
	e.Use(middleware.Logger)

	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	e.POST("flow", func(c echo.Context) (err error) {
		log.Print("flow")

		requestBody, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			c.Error(err)
			return err
		}

		input := string(requestBody)
		fmt.Printf("flow input: %s\n", input)

		exe, err := ExecWorklow(c, config.WorkflowID, input)
		if err != nil {
			c.Error(err)
			return err
		}

		// 実行結果出力
		fmt.Println("workflow: " + exe.Name)
		return c.String(200, "success")
	})

	e.POST("forward", func(c echo.Context) (err error) {
		log.Print("forward")

		q, err := BodyToQuery(c.Request().Body)
		if err != nil {
			c.Error(err)
			return err
		}

		client, err := NewClient(config.TargetURL, config.UseGoogleJWT)
		if err != nil {
			c.Error(err)
			return err
		}

		req, err := http.NewRequest("GET", config.TargetURL+"/?"+q, nil)
		if err != nil {
			c.Error(err)
			return err
		}

		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)
		res, err := client.Do(req)
		if err != nil {
			c.Error(err)
			return err
		}

		// call logs
		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})

	e.GET("*", func(c echo.Context) (err error) {
		client, err := NewClient(config.TargetURL, config.UseGoogleJWT)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("GET", config.TargetURL+c.Request().RequestURI, c.Request().Body)
		if err != nil {
			return err
		}

		req.Header = c.Request().Header.Clone()
		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})

	e.POST("*", func(c echo.Context) (err error) {
		requestBody, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			c.Error(err)
			return
		}
		log.Printf("body :%s\n", string(requestBody))

		client, err := NewClient(config.TargetURL, config.UseGoogleJWT)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", config.TargetURL+c.Request().RequestURI, ioutil.NopCloser(strings.NewReader(string(requestBody))))
		if err != nil {
			return err
		}

		req.Header = c.Request().Header.Clone()
		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})

	e.DELETE("*", func(c echo.Context) (err error) {
		client, err := NewClient(config.TargetURL, config.UseGoogleJWT)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("DELETE", config.TargetURL+c.Request().RequestURI, c.Request().Body)
		if err != nil {
			return err
		}

		req.Header = c.Request().Header.Clone()
		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})
}

func BodyToQuery(body io.ReadCloser) (string, error) {
	requestBody, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	log.Println(string(requestBody))
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(requestBody), &data); err != nil {
		fmt.Println(err)
	}

	xs := data["targetParams"].(map[string]interface{})
	params := make([]string, 0, len(xs))
	for k, v := range xs {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}

	q := strings.Join(params, "&")
	log.Println(q)

	return q, nil
}

func ExecWorklow(c echo.Context, workflowID string, input string) (*workflowexecutions.Execution, error) {
	ctx := context.Background()
	workflowExecService, err := workflowexecutions.NewService(ctx)
	if err != nil {
		return nil, nil
	}

	exe, err := workflowExecService.Projects.Locations.Workflows.Executions.Create(
		workflowID, &workflowexecutions.Execution{
			Name:     workflowID,
			Argument: input,
		}).Do()
	if err != nil {
		return nil, nil
	}

	return exe, nil
}
