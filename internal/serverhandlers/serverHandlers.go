package serverhandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
	"github.com/zlojkota/YL-1/internal/hashhelper"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type Stater interface {
	MetricMapMuxLock()
	MetricMapMuxUnlock()
	MetricMap() map[string]*collector.Metrics
	SetMetricMap(metricMap map[string]*collector.Metrics)
	MetricMapItem(item string) (*collector.Metrics, bool)
	SetMetricMapItem(metricMap *collector.Metrics)
	GetHaser() *hashhelper.Hasher
	InitHasher(hashKey string)
}

type Storager interface {
	Run(storeInterval time.Duration)
	Restore()
	SendDone()
	WaitDone()
	Init(store string)
	SetState(state Stater)
	Ping() bool
	StopStorage()
}

type ServerHandler struct {
	IndexPath string
	State     Stater
}

const counter = "counter"
const gauge = "gauge"

func (h *ServerHandler) Init(state Stater) {
	h.IndexPath = "./internal/httpRoot/index.html"
	h.State = state
	h.State.SetMetricMap(make(map[string]*collector.Metrics))
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
	mm := h.State.MetricMap()
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
		if val, ok := h.State.MetricMapItem(data.ID); ok {
			val.Hash = h.State.GetHaser().Hash(val)
			return c.JSON(http.StatusOK, val)
		} else {
			return c.NoContent(http.StatusNotFound)
		}
	} else {
		switch typeM := c.Param("type"); typeM {
		case counter:
			val, ok := h.State.MetricMapItem(c.Param("metric"))
			if !ok {
				return c.NoContent(http.StatusNotFound)
			}
			return c.String(http.StatusOK, strconv.FormatInt(*val.Delta, 10))
		case gauge:
			val, ok := h.State.MetricMapItem(c.Param("metric"))
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
			updateValue.Hash = h.State.GetHaser().HashC(updateValue.ID, val)
		case gauge:
			val, err := strconv.ParseFloat(c.Param("value"), 64)
			if err != nil {
				return c.NoContent(http.StatusBadRequest)
			}
			updateValue.Value = &val
			updateValue.Hash = h.State.GetHaser().HashG(updateValue.ID, val)
		default:
			return c.NoContent(http.StatusNotImplemented)
		}
	case "application/json":
		err := json.NewDecoder(c.Request().Body).Decode(&updateValue)
		if err != nil {
			return c.NoContent(http.StatusNotImplemented)
		}
		mmj, _ := json.Marshal(updateValue)
		fmt.Println("-REQUEST-", string(mmj))
		if !h.State.GetHaser().TestHash(&updateValue) {
			return c.NoContent(http.StatusBadRequest)
		}
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
	if _, ok := h.State.MetricMapItem(updateValue.ID); !ok {
		updateValue.Hash = h.State.GetHaser().Hash(&updateValue)
		mmj, _ := json.Marshal(updateValue)
		fmt.Println("-INSERT_DATA-", string(mmj))
		h.State.SetMetricMapItem(&updateValue)
		return c.NoContent(http.StatusOK)
	}
	switch updateValue.MType {
	case counter:
		metric, _ := h.State.MetricMapItem(updateValue.ID)
		delta := *metric.Delta + *updateValue.Delta
		h.State.SetMetricMapItem(&collector.Metrics{
			ID:    updateValue.ID,
			MType: updateValue.MType,
			Delta: &delta,
			Hash:  h.State.GetHaser().HashC(updateValue.ID, delta),
		})
		r, _ := h.State.MetricMapItem(updateValue.ID)
		mmj, _ := json.Marshal(r)
		fmt.Println("-UPDATE_DATA-", string(mmj))
		return c.NoContent(http.StatusOK)
	case gauge:
		h.State.SetMetricMapItem(&collector.Metrics{
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
