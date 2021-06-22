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

	"google.golang.org/api/idtoken"

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

func Route(e *echo.Echo) {
	e.HTTPErrorHandler = middleware.HTTPErrorHandler
	e.Use(middleware.Logger)

	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	e.GET("*", func(c echo.Context) (err error) {
		client, err := NewClient(config.TargetURL, config.UseGoogleJWT)
		if err != nil {
			return err
		}

		req, _ := http.NewRequest("GET", config.TargetURL+c.Request().RequestURI, nil)
		req.Header = c.Request().Header.Clone()
		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)

		res, _ := client.Do(req)
		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})
}
