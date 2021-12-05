package collector

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

type CollectorHandle interface {
	SendMetrics(counter map[string]int64, gauge map[string]float64)
}

type Collector struct {
	counter      map[string]int64
	gauge        map[string]float64
	handle       CollectorHandle
	poolinterval time.Duration
	Done         chan bool
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

func (col *Collector) Handle(poolinterval time.Duration, handle CollectorHandle) {
	col.poolinterval = poolinterval
	col.handle = handle
	col.Done = make(chan bool)
}

func (col *Collector) Run() {
	col.gauge = map[string]float64{}
	col.counter = map[string]int64{}
	col.collect()
	if col.poolinterval == 0 {
		col.poolinterval = time.Second
	}
	tick := time.NewTicker(col.poolinterval)
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
				col.handle.SendMetrics(col.counter, col.gauge)
			}

		}
	}
}
