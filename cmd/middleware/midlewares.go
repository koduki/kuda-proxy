package middleware

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

func ErrorHandler4Panic(c echo.Context) {
	if err := recover(); err != nil {
		log.Printf("[ERROR] %s\n", err)
		for depth := 0; ; depth++ {
			_, file, line, ok := runtime.Caller(depth)
			if !ok {
				break
			}
			log.Printf("======> %d: %v:%d", depth, file, line)

		}

		err2 := xerrors.New("error in main method")
		fmt.Printf("%+v\n", err2)

		err3 := xerrors.Errorf("error in main method: %w", err)

		fmt.Printf("%+v\n", err3)

		c.Error(echo.NewHTTPError(http.StatusInternalServerError, err))
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

		log := make(map[string]interface{})

		log["time"] = time.Now().Format("2006-01-02T00:00:00")
		log["method"] = c.Request().Method
		log["uri"] = c.Request().RequestURI
		log["headers"] = c.Request().Header
		log["query-params"] = c.QueryParams()

		// exec
		herr := next(c)
		if herr != nil {
			c.Error(herr)
		}

		status := c.Response().Status
		log["status"] = status
		if status == 200 {
			log["severity"] = "INFO"
		} else {
			log["severity"] = "ERROR"
			log["errors"] = fmt.Sprintf("%+v\n", herr)
		}

		// print log
		json, err := json.Marshal(log)
		if err != nil {
			c.Error(err)
		}
		fmt.Println(string(json))

		return nil
	}
}
