package collector

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

type CollectorHandle interface {
	HandleData(counter map[string]int64, gauge map[string]float64)
}

type Collector struct {
	counter  map[string]int64
	gauge    map[string]float64
	handle   CollectorHandle
	duration time.Duration
	Done     <-chan struct{}
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (col *Collector) collect() {

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	col.gauge["Alloc"] = float64(rtm.Alloc)
	col.gauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	col.gauge["Frees"] = float64(rtm.Frees)
	col.gauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	col.gauge["GCSys"] = float64(rtm.GCSys)
	col.gauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	col.gauge["HeapIdle"] = float64(rtm.HeapIdle)
	col.gauge["HeapInuse"] = float64(rtm.HeapInuse)
	col.gauge["HeapObjects"] = float64(rtm.HeapObjects)
	col.gauge["HeapReleased"] = float64(rtm.HeapReleased)
	col.gauge["HeapSys"] = float64(rtm.HeapSys)
	col.gauge["LastGC"] = float64(rtm.LastGC)
	col.gauge["Lookups"] = float64(rtm.Lookups)
	col.gauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	col.gauge["MCacheSys"] = float64(rtm.MCacheSys)
	col.gauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	col.gauge["MSpanSys"] = float64(rtm.MSpanSys)
	col.gauge["Mallocs"] = float64(rtm.Mallocs)
	col.gauge["NextGC"] = float64(rtm.NextGC)
	col.gauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	col.gauge["NumGC"] = float64(rtm.NumGC)
	col.gauge["OtherSys"] = float64(rtm.OtherSys)
	col.gauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	col.gauge["StackInuse"] = float64(rtm.StackInuse)
	col.gauge["StackSys"] = float64(rtm.StackSys)
	col.gauge["Sys"] = float64(rtm.Sys)
	col.counter["PollCount"]++
	col.gauge["RandomValue"] = float64(rand.Float32())

}

func (col *Collector) Handle(duration time.Duration, handle CollectorHandle) {
	col.duration = duration
	col.handle = handle
}

func (col *Collector) Run() {
	col.gauge = map[string]float64{}
	col.counter = map[string]int64{}
	col.collect()
	if col.duration == 0 {
		col.duration = time.Second
	}
	tick := time.NewTicker(col.duration)
	defer tick.Stop()
	for {
		select {
		case <-col.Done:
			return
		case <-tick.C:
			col.collect()
			if col.handle == nil {
				fmt.Print(col)
			} else {
				col.handle.HandleData(col.counter, col.gauge)
			}
		}
	}
}
