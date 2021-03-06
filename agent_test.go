package yl1

import (
	"github.com/zlojkota/YL-1/internal/memorystate"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/zlojkota/YL-1/internal/agentcollector"
	"github.com/zlojkota/YL-1/internal/collector"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
)

const iterations = 200

type Worker struct {
	t *testing.T
	h *serverhandlers.ServerHandler
	e *echo.Echo
}

func (p *Worker) RequestSend(req *http.Request) {

	p.t.Run("Handling reqwest-response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := p.e.NewContext(req, rec)
		r := p.e.Router()
		r.Find(http.MethodPost, req.URL.Path, c)
		// Assertions
		val := req.Header.Get("Content-Type")
		if val == "application/json" {
			if assert.NoError(p.t, p.h.UpdateBATCHHandler(c), "Error application/json") {
				if req.URL.Path == "/update/" {
					assert.Equal(p.t, http.StatusOK, rec.Code, "Not valid Post update application/json")
				}
			}
		} else {
			c.SetPath("/update/:type/:metric/:value")
			if assert.NoError(p.t, p.h.UpdateHandler(c), "Error plain/text") {
				assert.Equal(p.t, http.StatusOK, rec.Code, "Not valid Get update text/plain")
			}
		}
	})

}

func (p *Worker) InitWorker(t *testing.T) {
	p.t = t
	p.h = new(serverhandlers.ServerHandler)
	mem := new(memorystate.MemoryState)
	mem.InitHasher("")
	p.h.Init(mem)
	p.e = echo.New()
	p.e.GET("/*", p.h.NotFoundHandler)
	p.e.POST("/*", p.h.NotFoundHandler)

	// update Handler
	p.e.POST("/update/:type/:metric/:value", p.h.UpdateHandler)
	p.e.POST("/update/", p.h.UpdateHandler)
	p.e.POST("/updates/", p.h.UpdateBATCHHandler)

	// homePage Handler
	p.e.GET("/", p.h.MainHandler)

	// getValue Handler
	p.e.GET("/value/:type/:metric", p.h.GetHandler)
	p.e.POST("/value/", p.h.GetHandler)

}

func TestAllapp(t *testing.T) {

	var col collector.Collector
	var agent agentcollector.Agent
	var worker Worker
	worker.InitWorker(t)
	agent.InitAgentJSON(&worker, "")
	col.Handle(2*time.Millisecond, &agent)
	go func() {
		col.Run()
	}()
	t.Run("Check update value", func(t *testing.T) {
		var (
			oldMapGauge   map[string]float64
			oldMapCounter map[string]uint64
		)
		oldMapGauge = make(map[string]float64)
		oldMapCounter = make(map[string]uint64)

		tick := time.NewTicker(4 * time.Millisecond)
		defer tick.Stop()
		iter := iterations
		loop := true
		updatedCounter := false
		updatedGauge := false

		for loop {
			<-tick.C
			newMapGauge := make(map[string]float64)
			newMapCounter := make(map[string]uint64)
			mm := worker.h.State.MetricMap()
			worker.h.State.MetricMapMuxLock()
			for _, val := range mm {
				switch val.MType {
				case "gauge":
					newMapGauge[val.ID] = *val.Value
				case "counter":
					newMapCounter[val.ID] = *val.Delta
				}
			}
			worker.h.State.MetricMapMuxUnlock()
			if iter == 0 {
				col.Done <- true
				loop = false
			} else if iter != iterations {
				for key, val := range newMapCounter {
					oldVal, ok := oldMapCounter[key]
					if val != oldVal && ok {
						updatedCounter = true
					}
				}
				for key, val := range newMapGauge {
					oldVal, ok := oldMapGauge[key]
					if val != oldVal && ok {
						updatedGauge = true
					}
				}
			}
			for key, val := range newMapCounter {
				oldMapCounter[key] = val
			}
			for key, val := range newMapGauge {
				oldMapGauge[key] = val
			}
			iter--
		}
		assert.True(t, updatedCounter, "Metric counter notUpdated")
		assert.True(t, updatedGauge, "Metric gauge notUpdated")
	})
	t.Run("Check Not update value", func(t *testing.T) {
		var (
			oldMapGauge   map[string]float64
			oldMapCounter map[string]uint64
		)
		oldMapGauge = make(map[string]float64)
		oldMapCounter = make(map[string]uint64)

		tick := time.NewTicker(4 * time.Millisecond)
		defer tick.Stop()

		updatedCounter := false
		updatedGauge := false
		iter := iterations
		loop := true
		for loop {
			<-tick.C
			newMapGauge := make(map[string]float64)
			newMapCounter := make(map[string]uint64)
			mm := worker.h.State.MetricMap()
			for _, val := range mm {
				switch val.MType {
				case "gauge":
					newMapGauge[val.ID] = *val.Value
				case "counter":
					newMapCounter[val.ID] = *val.Delta
				}
			}

			if iter == 0 {
				loop = false
			} else if iter != iterations {
				for key, val := range newMapCounter {
					oldVal, ok := oldMapCounter[key]
					if val != oldVal && ok {
						updatedCounter = true
					}
				}
				for key, val := range newMapGauge {
					oldVal, ok := oldMapGauge[key]
					if val != oldVal && ok {
						updatedGauge = true
					}
				}

			}
			for key, val := range newMapCounter {
				oldMapCounter[key] = val
			}
			for key, val := range newMapGauge {
				oldMapGauge[key] = val
			}
			iter--
		}
		assert.False(t, updatedCounter, "Metric counter isUpdated")
		assert.False(t, updatedGauge, "Metric gauge isUpdated")
	})
}
