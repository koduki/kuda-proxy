package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/api/idtoken"

	"github.com/labstack/echo/v4"

	"kuda/cmd/middleware"
)

type (
	Response struct {
		Message string `json:"message"`
		Date    string `json:"date"`
	}
)

type (
	Configuration struct {
		TargetURL    string
		UseGoogleJWT bool
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
		// x := 1
		// if true {
		// 	x = x * 5

		// }

		// y := 1 / (x - 5)
		// fmt.Println(y)

		return &http.Client{
			Timeout: time.Second * 10,
			// }, nil
		}, xerrors.Errorf("stacktrace: %w", echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials"))
		// }, echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")

	}
}

func Route(e *echo.Echo) {
	e.HTTPErrorHandler = middleware.HTTPErrorHandler
	e.Use(middleware.Logger)

	config := Configuration{
		TargetURL:    "https://www.google.com", //https://kuda-target-dnb6froqha-uc.a.run.app",
		UseGoogleJWT: false,
	}

	e.GET("*", func(c echo.Context) (err error) {

		var hparams []string
		for k, v := range c.QueryParams() {
			hparams = append(hparams, k+"="+v[0])
		}

		client, err := NewClient(config.TargetURL, config.UseGoogleJWT)
		if err != nil {
			// return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")

			return err
		}

		req, _ := http.NewRequest("GET", config.TargetURL, nil)
		req.Header = c.Request().Header.Clone()
		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)

		res, _ := client.Do(req)
		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})
}
