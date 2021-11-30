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

func (m *Collector) collect() {

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	m.gauge["Alloc"] = float64(rtm.Alloc)
	m.gauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	m.gauge["Frees"] = float64(rtm.Frees)
	m.gauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	m.gauge["GCSys"] = float64(rtm.GCSys)
	m.gauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	m.gauge["HeapIdle"] = float64(rtm.HeapIdle)
	m.gauge["HeapInuse"] = float64(rtm.HeapInuse)
	m.gauge["HeapObjects"] = float64(rtm.HeapObjects)
	m.gauge["HeapReleased"] = float64(rtm.HeapReleased)
	m.gauge["HeapSys"] = float64(rtm.HeapSys)
	m.gauge["LastGC"] = float64(rtm.LastGC)
	m.gauge["Lookups"] = float64(rtm.Lookups)
	m.gauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	m.gauge["MCacheSys"] = float64(rtm.MCacheSys)
	m.gauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	m.gauge["MSpanSys"] = float64(rtm.MSpanSys)
	m.gauge["Mallocs"] = float64(rtm.Mallocs)
	m.gauge["NextGC"] = float64(rtm.NextGC)
	m.gauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	m.gauge["NumGC"] = float64(rtm.NumGC)
	m.gauge["OtherSys"] = float64(rtm.OtherSys)
	m.gauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	m.gauge["StackInuse"] = float64(rtm.StackInuse)
	m.gauge["StackSys"] = float64(rtm.StackSys)
	m.gauge["Sys"] = float64(rtm.Sys)
	m.counter["PollCount"]++
	m.gauge["RandomValue"] = float64(rand.Float32())

}

func (p *Collector) Handle(duration time.Duration, handle CollectorHandle) {
	p.duration = duration
	p.handle = handle
}

func (p *Collector) Run() {
	p.gauge = map[string]float64{}
	p.counter = map[string]int64{}
	p.collect()
	if p.duration == 0 {
		p.duration = time.Second
	}
	tick := time.NewTicker(p.duration)
	defer tick.Stop()
	for {
		select {
		case <-p.Done:
			return
		case <-tick.C:
			p.collect()
			if p.handle == nil {
				fmt.Print(p)
			} else {
				p.handle.HandleData(p.counter, p.gauge)
			}
		}
	}
}
