package mymiddleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/xerrors"
)

func logging(c echo.Context, callback func() error) {
	log := make(map[string]interface{})

	log["time"] = time.Now().Format("2006-01-02T00:00:00")
	log["method"] = c.Request().Method
	log["uri"] = c.Request().RequestURI
	log["headers"] = c.Request().Header
	log["query-params"] = c.QueryParams()

	// if requestBody, e := io.ReadAll(c.Request().Body); e != nil {
	// 	c.Error(e)
	// } else {
	// 	log["body-params"] = string(requestBody)
	// }

	err := callback()
	status := c.Response().Status
	log["status"] = status
	if status == 200 {
		log["severity"] = "INFO"
	} else {
		log["severity"] = "ERROR"
		log["errors"] = fmt.Sprintf("%+v\n", err)
	}

	// print log
	if json, e := json.Marshal(log); e != nil {
		c.Error(e)
	} else {
		fmt.Println(string(json))
	}
}

func ErrorHandler4Panic(c echo.Context) {
	if err := recover(); err != nil {
		log.Printf("[ERROR] %s\n", err)
		for depth := 0; ; depth++ {
			_, file, line, ok := runtime.Caller(depth)
			if !ok {
				break
			}
			log.Printf("\t %d: %v:%d", depth, file, line)
		}

		herr := echo.NewHTTPError(http.StatusInternalServerError, err)
		logging(c, func() error { return herr })
		c.Error(herr)
	}
}

func HTTPErrorHandler(err error, c echo.Context) {
	if he, ok := xerrors.Unwrap(err).(*echo.HTTPError); ok {
		c.JSON(he.Code, he)
	} else if he, ok := err.(*echo.HTTPError); ok {
		c.JSON(he.Code, he)
	}
}

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer ErrorHandler4Panic(c)

		logging(c, func() error {
			herr := next(c)
			if herr != nil {
				c.Error(herr)
			}
			return herr
		})

		return nil
	}
}
