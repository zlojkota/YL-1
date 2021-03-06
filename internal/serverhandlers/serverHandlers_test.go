package serverhandlers

import (
	"github.com/zlojkota/YL-1/internal/memorystate"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestServerHandler_NotFoundHandler(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := new(ServerHandler)

		// Assertions
		if assert.NoError(t, h.NotFoundHandler(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
}

func TestServerHandler_GetUpdateHandlers(t *testing.T) {
	type requestType struct {
		name      string
		uri       string
		method    string
		typeName  string
		metric    string
		value     string
		code      int
		reqMethod string
	}

	tc := []requestType{
		{
			uri:       "/update/counter/testCounter/none",
			code:      http.StatusBadRequest,
			name:      "BadRequest POST /update/counter/testCounter/none",
			method:    "update",
			typeName:  "counter",
			metric:    "testCounter",
			value:     "none",
			reqMethod: http.MethodPost,
		},
		{
			uri:       "/update/counter/testCounter/100",
			code:      http.StatusOK,
			name:      "All OK POST /update/counter/testCounter/100",
			method:    "update",
			typeName:  "counter",
			metric:    "testCounter",
			value:     "100",
			reqMethod: http.MethodPost,
		},
		{
			uri:       "/update/gauge/testGauge/none",
			code:      http.StatusBadRequest,
			name:      "NotFound POST /update/gauge/testGauge/none",
			method:    "update",
			typeName:  "gauge",
			metric:    "testGauge",
			value:     "none",
			reqMethod: http.MethodPost,
		},
		{
			uri:       "/update/gauge/testGauge/100",
			code:      http.StatusOK,
			name:      "All OK POST /update/gauge/testGauge/100",
			method:    "update",
			typeName:  "gauge",
			metric:    "testGauge",
			value:     "100",
			reqMethod: http.MethodPost,
		},
		{
			uri:       "/update/unknown/testCounter/100",
			code:      http.StatusNotImplemented,
			name:      "NotImplement POST /update/unknown/testCounter/100",
			method:    "update",
			typeName:  "unknown",
			metric:    "testCounter",
			value:     "100",
			reqMethod: http.MethodPost,
		},
		{
			name:      "Get Unknown counter",
			uri:       "/value/counter/Testcounter321",
			method:    "value",
			typeName:  "counter",
			metric:    "Testcounter321",
			code:      http.StatusNotFound,
			reqMethod: http.MethodGet,
		},
		{
			name:      "Set Unknown counter",
			uri:       "/update/counter/Testcounter321/13",
			method:    "update",
			typeName:  "counter",
			metric:    "Testcounter321",
			value:     "13",
			code:      http.StatusOK,
			reqMethod: http.MethodPost,
		},
		{
			name:      "Get Known counter",
			uri:       "/value/counter/Testcounter321",
			method:    "value",
			typeName:  "counter",
			metric:    "Testcounter321",
			value:     "13",
			code:      http.StatusOK,
			reqMethod: http.MethodGet,
		},
		{
			name:      "Update Testcounter321 counter",
			uri:       "/update/counter/Testcounter321/13",
			method:    "update",
			typeName:  "counter",
			metric:    "Testcounter321",
			value:     "13",
			code:      http.StatusOK,
			reqMethod: http.MethodPost,
		},
		{
			name:      "Check Testcounter321 counter",
			uri:       "/value/counter/Testcounter321",
			method:    "value",
			typeName:  "counter",
			metric:    "Testcounter321",
			value:     "26",
			code:      http.StatusOK,
			reqMethod: http.MethodGet,
		},
		{
			name:      "Update Testcounter321 counter",
			uri:       "/update/counter/Testcounter321/100",
			method:    "update",
			typeName:  "counter",
			metric:    "Testcounter321",
			value:     "100",
			code:      http.StatusOK,
			reqMethod: http.MethodPost,
		},
		{
			name:      "Check Testcounter321 counter",
			uri:       "/value/counter/Testcounter321",
			method:    "value",
			typeName:  "counter",
			metric:    "Testcounter321",
			value:     "126",
			code:      http.StatusOK,
			reqMethod: http.MethodGet,
		},
		{
			name:      "Get other Unknown counter",
			uri:       "/value/counter/Testcounter123",
			method:    "value",
			typeName:  "counter",
			metric:    "Testcounter123",
			code:      http.StatusNotFound,
			reqMethod: http.MethodGet,
		},

		{
			name:      "Set Testgauge321 gauge",
			uri:       "/update/gauge/Testgauge321/13.3251",
			method:    "update",
			typeName:  "gauge",
			metric:    "Testgauge321",
			value:     "13.3251",
			code:      http.StatusOK,
			reqMethod: http.MethodPost,
		},
		{
			name:      "Get Testgauge321 gauge",
			uri:       "/value/gauge/Testgauge321",
			method:    "value",
			typeName:  "gauge",
			metric:    "Testgauge321",
			value:     "13.3251",
			code:      http.StatusOK,
			reqMethod: http.MethodGet,
		},
		{
			name:      "Update Testgauge321 gauge",
			uri:       "/update/gauge/Testgauge321/3551325.3",
			method:    "update",
			typeName:  "gauge",
			metric:    "Testgauge321",
			value:     "3551325.3",
			code:      http.StatusOK,
			reqMethod: http.MethodPost,
		},
		{
			name:      "Check Testgauge321 gauge",
			uri:       "/value/gauge/Testgauge321",
			method:    "value",
			typeName:  "gauge",
			metric:    "Testgauge321",
			value:     "3551325.3",
			code:      http.StatusOK,
			reqMethod: http.MethodGet,
		},
		{
			name:      "Update Testgauge321 gauge",
			uri:       "/update/gauge/Testgauge321/100",
			method:    "update",
			typeName:  "gauge",
			metric:    "Testgauge321",
			value:     "100",
			code:      http.StatusOK,
			reqMethod: http.MethodPost,
		},
		{
			name:      "Check Testgauge321 gauge",
			uri:       "/value/gauge/Testgauge321",
			method:    "value",
			typeName:  "gauge",
			metric:    "Testgauge321",
			value:     "100",
			code:      http.StatusOK,
			reqMethod: http.MethodGet,
		},
		{
			name:      "Get other Unknown gauge",
			uri:       "/value/gauge/Testgauge123",
			method:    "value",
			typeName:  "gauge",
			metric:    "Testgauge123",
			code:      http.StatusNotFound,
			reqMethod: http.MethodGet,
		},
	}

	h := new(ServerHandler)
	mem := new(memorystate.MemoryState)
	mem.InitHasher("")
	h.Init(mem)
	h.IndexPath = "../../internal/httpRoot/index.html"
	t.Run("Home page blank", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Assertions
		if assert.NoError(t, h.MainHandler(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotContains(t, rec.Body.String(), "http://localhost:8080")
		}
	})

	for _, itc := range tc {
		t.Run(itc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(itc.reqMethod, itc.uri, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			switch itc.reqMethod {
			case http.MethodPost:
				c.SetPath("/update/:type/:metric/:value")
				c.SetParamNames("method", "type", "metric", "value")
				c.SetParamValues(itc.method, itc.typeName, itc.metric, itc.value)
				// Assertions
				if assert.NoError(t, h.UpdateHandler(c)) {
					assert.Equal(t, itc.code, rec.Code)
				}
			case http.MethodGet:
				c.SetPath("/value/:type/:metric")
				c.SetParamNames("method", "type", "metric")
				c.SetParamValues(itc.method, itc.typeName, itc.metric)
				// Assertions
				if assert.NoError(t, h.GetHandler(c)) {
					assert.Equal(t, itc.code, rec.Code)
					if itc.value != "" {
						assert.Equal(t, itc.value, rec.Body.String())
					}
				}
			default:
				panic("Tests not valid!")
			}
		})
	}

	t.Run("Home page with data", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Assertions
		if assert.NoError(t, h.MainHandler(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), "http://localhost:8080")
		}
	})
}
