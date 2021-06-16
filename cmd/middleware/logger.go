package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log := make(map[string]interface{})

		log["severity"] = "INFO"
		log["time"] = time.Now().Format("2006-01-02T00:00:00")
		log["method"] = c.Request().Method
		log["uri"] = c.Request().RequestURI
		log["headers"] = c.Request().Header
		log["query-params"] = c.QueryParams()
		log["status"] = c.Response().Status

		json, _ := json.Marshal(log)
		fmt.Println(string(json))

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}
