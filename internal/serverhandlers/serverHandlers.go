package serverhandlers

import (
	"bytes"
	"github.com/labstack/gommon/log"
	"html/template"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ServerHandler struct {
	MetricMapGauge   map[string]float64
	MetricMapCounter map[string]int64
}

const counter = "counter"
const gauge = "gauge"

func NewServerHandler() *ServerHandler {
	p := new(ServerHandler)
	p.MetricMapCounter = make(map[string]int64)
	p.MetricMapGauge = make(map[string]float64)
	return p
}

func (h *ServerHandler) NotFoundHandler(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func (h *ServerHandler) MainHandler(c echo.Context) error {

	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, "Internal ERROR")
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, *h)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, "Internal ERROR")
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return c.String(http.StatusOK, buf.String())
}

func (h *ServerHandler) GetHandler(c echo.Context) error {
	switch typeM := c.Param("type"); typeM {
	case counter:
		val, ok := h.MetricMapCounter[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatInt(val, 10))
	case gauge:
		val, ok := h.MetricMapGauge[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatFloat(val, 'f', -1, 64))
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}

func (h *ServerHandler) UpdateHandler(c echo.Context) error {

	switch typeM := c.Param("type"); typeM {
	case counter:
		val, err := strconv.ParseInt(c.Param("value"), 0, 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		h.MetricMapCounter[c.Param("metric")] += val
		return c.NoContent(http.StatusOK)
	case gauge:
		val, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		h.MetricMapGauge[c.Param("metric")] = val
		return c.NoContent(http.StatusOK)
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}

func (h *ServerHandler) UpdateJSONHandler(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.NoContent(http.StatusOK)
	}
	return c.NoContent(http.StatusNotFound)
}
