package serverhandlers

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
	"github.com/zlojkota/YL-1/internal/hashhelper"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
)

type ServerHandler struct {
	metricMap    map[string]*collector.Metrics
	metricMapMux sync.Mutex
	IndexPath    string
	hasher       hashhelper.Hasher
}

func (h *ServerHandler) SetHasher(key string) {
	h.hasher.SetKey(key)
}

func (h *ServerHandler) MetricMap() map[string]*collector.Metrics {
	return h.metricMap
}

func (h *ServerHandler) SetMetricMap(metricMap map[string]*collector.Metrics) {
	h.metricMapMux.Lock()
	h.metricMap = metricMap
	h.metricMapMux.Unlock()
}

func (h *ServerHandler) MetricMapItem(item string) (*collector.Metrics, bool) {
	res, ok := h.metricMap[item]
	return res, ok
}

func (h *ServerHandler) SetMetricMapItem(metricMap *collector.Metrics) {
	h.metricMapMux.Lock()
	h.metricMap[metricMap.ID] = metricMap
	h.metricMapMux.Unlock()
}

const counter = "counter"
const gauge = "gauge"

func NewServerHandler() *ServerHandler {
	p := new(ServerHandler)
	p.metricMap = make(map[string]*collector.Metrics)
	p.IndexPath = "./internal/httpRoot/index.html"
	return p
}

func (h *ServerHandler) NotFoundHandler(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func (h *ServerHandler) MainHandler(c echo.Context) error {

	t, err := template.ParseFiles(h.IndexPath)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, "Internal ERROR")
	}
	var buf bytes.Buffer
	mm := h.MetricMap()
	err = t.Execute(&buf, mm)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, "Internal ERROR")
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return c.String(http.StatusOK, buf.String())
}

func (h *ServerHandler) GetHandler(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		var data collector.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&data)
		if err != nil {
			return c.NoContent(http.StatusNotImplemented)
		}
		if val, ok := h.MetricMapItem(data.ID); ok {
			return c.JSON(http.StatusOK, val)
		} else {
			return c.NoContent(http.StatusNotFound)
		}
	} else {
		switch typeM := c.Param("type"); typeM {
		case counter:
			val, ok := h.MetricMapItem(c.Param("metric"))
			if !ok {
				return c.NoContent(http.StatusNotFound)
			}
			return c.String(http.StatusOK, strconv.FormatInt(*val.Delta, 10))
		case gauge:
			val, ok := h.MetricMapItem(c.Param("metric"))
			if !ok {
				return c.NoContent(http.StatusNotFound)
			}
			return c.String(http.StatusOK, strconv.FormatFloat(*val.Value, 'f', -1, 64))
		default:
			return c.NoContent(http.StatusNotImplemented)
		}
	}
}

func (h *ServerHandler) UpdateHandler(c echo.Context) error {

	var updateValue collector.Metrics
	contentType := c.Request().Header.Get("Content-Type")
	if contentType == "" &&
		c.Param("metric") != "" &&
		c.Param("value") != "" &&
		c.Param("type") != "" {

		contentType = "text/plain"
	}

	switch contentType {
	case "text/plain":
		updateValue.ID = c.Param("metric")
		updateValue.MType = c.Param("type")
		switch c.Param("type") {
		case counter:
			val, err := strconv.ParseInt(c.Param("value"), 0, 64)
			if err != nil {
				return c.NoContent(http.StatusBadRequest)
			}
			updateValue.Delta = &val
		case gauge:
			val, err := strconv.ParseFloat(c.Param("value"), 64)
			if err != nil {
				return c.NoContent(http.StatusBadRequest)
			}
			updateValue.Value = &val
		default:
			return c.NoContent(http.StatusNotImplemented)
		}
	case "application/json":
		err := json.NewDecoder(c.Request().Body).Decode(&updateValue)
		if err != nil {
			return c.NoContent(http.StatusNotImplemented)
		}
		if !h.hasher.TestHash(&updateValue) {
			return c.NoContent(http.StatusBadRequest)
		}
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
	if _, ok := h.MetricMapItem(updateValue.ID); !ok {
		h.SetMetricMapItem(&updateValue)
		return c.NoContent(http.StatusOK)
	}
	switch updateValue.MType {
	case counter:
		metric, _ := h.MetricMapItem(updateValue.ID)
		delta := *metric.Delta + *updateValue.Delta
		h.SetMetricMapItem(&collector.Metrics{
			ID:    updateValue.ID,
			MType: updateValue.MType,
			Delta: &delta,
			Hash:  updateValue.Hash,
		})
		return c.NoContent(http.StatusOK)
	case gauge:
		h.SetMetricMapItem(&collector.Metrics{
			ID:    updateValue.ID,
			MType: updateValue.MType,
			Value: updateValue.Value,
			Hash:  updateValue.Hash,
		})
		return c.NoContent(http.StatusOK)
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}
