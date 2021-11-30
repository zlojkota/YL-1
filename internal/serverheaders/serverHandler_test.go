package serverheaders

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerHandler_NotFoundHandler(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := NewServerHandler()

		// Assertions
		if assert.NoError(t, h.NotFoundHandler(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
}

func TestServerHandler_UpdateHandler(t *testing.T) {
	type reqwestType struct {
		uri      string
		method   string
		typeName string
		metric   string
		value    string
		code     int
		name     string
	}

	tc := []reqwestType{
		{
			uri:      "/update/counter/testCounter/none",
			code:     http.StatusBadRequest,
			name:     "BadRequest POST /update/counter/testCounter/none",
			method:   "update",
			typeName: "counter",
			metric:   "testCounter",
			value:    "none",
		},
		{
			uri:      "/update/counter/testCounter/100",
			code:     http.StatusOK,
			name:     "All OK POST /update/counter/testCounter/100",
			method:   "update",
			typeName: "counter",
			metric:   "testCounter",
			value:    "100",
		},
		{
			uri:      "/update/gauge/testGauge/none",
			code:     http.StatusBadRequest,
			name:     "NotFound POST /update/gauge/testGauge/none",
			method:   "update",
			typeName: "gauge",
			metric:   "testGauge",
			value:    "none",
		},
		{
			uri:      "/update/gauge/testGauge/100",
			code:     http.StatusOK,
			name:     "All OK POST /update/gauge/testGauge/100",
			method:   "update",
			typeName: "gauge",
			metric:   "testGauge",
			value:    "100",
		},
		{
			uri:      "/update/unknown/testCounter/100",
			code:     http.StatusNotImplemented,
			name:     "NotImplement POST /update/unknown/testCounter/100",
			method:   "update",
			typeName: "unknown",
			metric:   "testCounter",
			value:    "100",
		},
		{
			uri:      "/updater/gauge/testGauge/100",
			code:     http.StatusNotFound,
			name:     "NotFound POST /updater/gauge/testGauge/100",
			method:   "updater",
			typeName: "gauge",
			metric:   "testGauge",
			value:    "100",
		},
	}

	for _, itc := range tc {
		t.Run(itc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, itc.uri, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:method/:type/:metric/:value")
			c.SetParamNames("method", "type", "metric", "value")
			c.SetParamValues(itc.method, itc.typeName, itc.metric, itc.value)
			h := NewServerHandler()

			// Assertions
			if assert.NoError(t, h.UpdateHandler(c)) {
				assert.Equal(t, itc.code, rec.Code)
			}
		})
	}
}
