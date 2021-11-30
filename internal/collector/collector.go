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

func (asiwantsoicallIliketheletterpsmall *Collector) collect() {

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	asiwantsoicallIliketheletterpsmall.gauge["Alloc"] = float64(rtm.Alloc)
	asiwantsoicallIliketheletterpsmall.gauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	asiwantsoicallIliketheletterpsmall.gauge["Frees"] = float64(rtm.Frees)
	asiwantsoicallIliketheletterpsmall.gauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	asiwantsoicallIliketheletterpsmall.gauge["GCSys"] = float64(rtm.GCSys)
	asiwantsoicallIliketheletterpsmall.gauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	asiwantsoicallIliketheletterpsmall.gauge["HeapIdle"] = float64(rtm.HeapIdle)
	asiwantsoicallIliketheletterpsmall.gauge["HeapInuse"] = float64(rtm.HeapInuse)
	asiwantsoicallIliketheletterpsmall.gauge["HeapObjects"] = float64(rtm.HeapObjects)
	asiwantsoicallIliketheletterpsmall.gauge["HeapReleased"] = float64(rtm.HeapReleased)
	asiwantsoicallIliketheletterpsmall.gauge["HeapSys"] = float64(rtm.HeapSys)
	asiwantsoicallIliketheletterpsmall.gauge["LastGC"] = float64(rtm.LastGC)
	asiwantsoicallIliketheletterpsmall.gauge["Lookups"] = float64(rtm.Lookups)
	asiwantsoicallIliketheletterpsmall.gauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	asiwantsoicallIliketheletterpsmall.gauge["MCacheSys"] = float64(rtm.MCacheSys)
	asiwantsoicallIliketheletterpsmall.gauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	asiwantsoicallIliketheletterpsmall.gauge["MSpanSys"] = float64(rtm.MSpanSys)
	asiwantsoicallIliketheletterpsmall.gauge["Mallocs"] = float64(rtm.Mallocs)
	asiwantsoicallIliketheletterpsmall.gauge["NextGC"] = float64(rtm.NextGC)
	asiwantsoicallIliketheletterpsmall.gauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	asiwantsoicallIliketheletterpsmall.gauge["NumGC"] = float64(rtm.NumGC)
	asiwantsoicallIliketheletterpsmall.gauge["OtherSys"] = float64(rtm.OtherSys)
	asiwantsoicallIliketheletterpsmall.gauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	asiwantsoicallIliketheletterpsmall.gauge["StackInuse"] = float64(rtm.StackInuse)
	asiwantsoicallIliketheletterpsmall.gauge["StackSys"] = float64(rtm.StackSys)
	asiwantsoicallIliketheletterpsmall.gauge["Sys"] = float64(rtm.Sys)
	asiwantsoicallIliketheletterpsmall.counter["PollCount"]++
	asiwantsoicallIliketheletterpsmall.gauge["RandomValue"] = float64(rand.Float32())

}

func (asiwantsoicallIliketheletterpsmall *Collector) Handle(duration time.Duration, handle CollectorHandle) {
	asiwantsoicallIliketheletterpsmall.duration = duration
	asiwantsoicallIliketheletterpsmall.handle = handle
}

func (asiwantsoicallIliketheletterpsmall *Collector) Run() {
	asiwantsoicallIliketheletterpsmall.gauge = map[string]float64{}
	asiwantsoicallIliketheletterpsmall.counter = map[string]int64{}
	asiwantsoicallIliketheletterpsmall.collect()
	if asiwantsoicallIliketheletterpsmall.duration == 0 {
		asiwantsoicallIliketheletterpsmall.duration = time.Second
	}
	tick := time.NewTicker(asiwantsoicallIliketheletterpsmall.duration)
	defer tick.Stop()
	for {
		select {
		case <-asiwantsoicallIliketheletterpsmall.Done:
			return
		case <-tick.C:
			asiwantsoicallIliketheletterpsmall.collect()
			if asiwantsoicallIliketheletterpsmall.handle == nil {
				fmt.Print(asiwantsoicallIliketheletterpsmall)
			} else {
				asiwantsoicallIliketheletterpsmall.handle.HandleData(asiwantsoicallIliketheletterpsmall.counter, asiwantsoicallIliketheletterpsmall.gauge)
			}
		}
	}
}
