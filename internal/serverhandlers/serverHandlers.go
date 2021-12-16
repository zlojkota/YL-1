package serverhandlers

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/gommon/log"
	"html/template"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type ServerHandler struct {
	MetricMap map[string]*Metrics
}

const counter = "counter"
const gauge = "gauge"

func NewServerHandler() *ServerHandler {
	p := new(ServerHandler)
	p.MetricMap = make(map[string]*Metrics)
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
		val, ok := h.MetricMap[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatInt(*val.Delta, 10))
	case gauge:
		val, ok := h.MetricMap[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatFloat(*val.Value, 'f', -1, 64))
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}

func (h *ServerHandler) GetJSONHandler(c echo.Context) error {
	var data Metrics
	err := json.NewDecoder(c.Request().Body).Decode(&data)
	if err != nil {
		return c.NoContent(http.StatusNotImplemented)
	}
	if val, ok := h.MetricMap[data.ID]; ok {
		return c.JSON(http.StatusOK, val)
	} else {
		return c.NoContent(http.StatusNotFound)
	}
}

func (h *ServerHandler) UpdateHandler(c echo.Context) error {
	switch c.Param("type") {
	case counter:
		val, err := strconv.ParseInt(c.Param("value"), 0, 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		updateValue := h.MetricMap[c.Param("metric")]
		if updateValue == nil {
			updateValue = &Metrics{
				ID:    c.Param("metric"),
				MType: counter,
				Delta: &val,
			}
		} else {
			*updateValue.Delta += val
		}
		h.MetricMap[c.Param("metric")] = updateValue
		return c.NoContent(http.StatusOK)
	case gauge:
		val, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		updateValue := h.MetricMap[c.Param("metric")]
		if updateValue == nil {
			updateValue = &Metrics{
				ID:    c.Param("metric"),
				MType: gauge,
				Value: &val,
			}
		} else {
			updateValue.Value = &val
		}
		h.MetricMap[c.Param("metric")] = updateValue
		return c.NoContent(http.StatusOK)
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}

func (h *ServerHandler) UpdateJSONHandler(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		var data Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&data)
		if err != nil {
			return c.NoContent(http.StatusNotImplemented)
		}

		switch data.MType {
		case counter:
			updateValue := h.MetricMap[data.ID]
			if updateValue == nil {
				updateValue = &Metrics{
					ID:    c.Param("metric"),
					MType: counter,
					Delta: data.Delta,
				}
			} else {
				*updateValue.Delta += *data.Delta
			}
			h.MetricMap[data.ID] = updateValue
			return c.NoContent(http.StatusOK)
		case gauge:
			h.MetricMap[data.ID] = &data
			return c.NoContent(http.StatusOK)
		default:
			return c.NoContent(http.StatusNotImplemented)
		}
	}
	return c.NoContent(http.StatusNotFound)
}

func (h *ServerHandler) GetterMetrics() map[string]*Metrics {
	return h.MetricMap
}

func (h *ServerHandler) SetterMetrics(metrics map[string]*Metrics) {
	h.MetricMap = metrics
}
