package serverHeaders

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func DefaultHandler(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "OK")
}

func UpdateHandler(c echo.Context) error {

	if c.Param("method") != "update" {
		return c.JSON(http.StatusNotFound, "Not found 404")
	}
	if c.Param("type") != "counter" || c.Param("type") != "gauge" {
		return c.JSON(http.StatusNotFound, "Not found 404")
	}
	//c.Param("metric")
	//c.Param("value")
	log.Printf("Type: %s, Metric %s=%s", c.Param("type"), c.Param("metric"), c.Param("value"))
	return c.JSON(http.StatusOK, "OK")
}
