package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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

func Route(e *echo.Echo) {
	e.Use(middleware.Logger)

	e.GET("*", func(c echo.Context) (err error) {

		var hparams []string
		for k, v := range c.QueryParams() {
			hparams = append(hparams, k+"="+v[0])
		}

		url := "https://kuda-target-dnb6froqha-uc.a.run.app/healthcheck"

		ctx := context.Background()
		client, err := idtoken.NewClient(ctx, url)
		if err != nil {
			return fmt.Errorf("idtoken.NewClient: %v", err)
		}

		req, _ := http.NewRequest("GET", url, nil)
		req.Header = c.Request().Header.Clone()
		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)

		res, _ := client.Do(req)

		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})
}
