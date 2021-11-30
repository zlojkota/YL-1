package serverheaders

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type ServerHandler struct {
	metricMapGauge   map[string]float64
	metricMapCounter map[string]int64
}

func NewServerHandler() *ServerHandler {
	p := new(ServerHandler)
	p.metricMapCounter = make(map[string]int64)
	p.metricMapGauge = make(map[string]float64)
	return p
}

func (h *ServerHandler) NotFoundHandler(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

/*
TODO Write Main handler
*/
func (h *ServerHandler) MainHandler(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *ServerHandler) GetHandler(c echo.Context) error {
	if c.Param("method") != "value" {
		return c.NoContent(http.StatusNotFound)
	}
	switch typeM := c.Param("type"); typeM {
	case "counter":
		val, ok := h.metricMapCounter[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatInt(val, 10))
	case "gauge":
		val, ok := h.metricMapGauge[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatFloat(val, 'g', -1, 64))
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}

func (h *ServerHandler) UpdateHandler(c echo.Context) error {

	if c.Param("method") != "update" {
		return c.NoContent(http.StatusNotFound)
	}
	switch typeM := c.Param("type"); typeM {
	case "counter":
		val, err := strconv.ParseInt(c.Param("value"), 0, 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		h.metricMapCounter[c.Param("metric")] = val
		return c.NoContent(http.StatusOK)
	case "gauge":
		val, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		h.metricMapGauge[c.Param("metric")] = val
		return c.NoContent(http.StatusOK)
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}
