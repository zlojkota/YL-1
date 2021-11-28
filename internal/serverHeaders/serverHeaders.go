package serverHeaders

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func NotFoundHandler(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func UpdateHandler(c echo.Context) error {

	if c.Param("method") != "update" {
		return c.NoContent(http.StatusNotFound)
	}
	switch typeM := c.Param("type"); typeM {
	case "counter":
		_, err := strconv.ParseInt(c.Param("value"), 0, 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusOK)
	case "gauge":
		_, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusOK)
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}
