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

func (this *Collector) collect() {

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	this.gauge["Alloc"] = float64(rtm.Alloc)
	this.gauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	this.gauge["Frees"] = float64(rtm.Frees)
	this.gauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	this.gauge["GCSys"] = float64(rtm.GCSys)
	this.gauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	this.gauge["HeapIdle"] = float64(rtm.HeapIdle)
	this.gauge["HeapInuse"] = float64(rtm.HeapInuse)
	this.gauge["HeapObjects"] = float64(rtm.HeapObjects)
	this.gauge["HeapReleased"] = float64(rtm.HeapReleased)
	this.gauge["HeapSys"] = float64(rtm.HeapSys)
	this.gauge["LastGC"] = float64(rtm.LastGC)
	this.gauge["Lookups"] = float64(rtm.Lookups)
	this.gauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	this.gauge["MCacheSys"] = float64(rtm.MCacheSys)
	this.gauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	this.gauge["MSpanSys"] = float64(rtm.MSpanSys)
	this.gauge["Mallocs"] = float64(rtm.Mallocs)
	this.gauge["NextGC"] = float64(rtm.NextGC)
	this.gauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	this.gauge["NumGC"] = float64(rtm.NumGC)
	this.gauge["OtherSys"] = float64(rtm.OtherSys)
	this.gauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	this.gauge["StackInuse"] = float64(rtm.StackInuse)
	this.gauge["StackSys"] = float64(rtm.StackSys)
	this.gauge["Sys"] = float64(rtm.Sys)
	this.counter["PollCount"]++
	this.gauge["RandomValue"] = float64(rand.Float32())

}

func (this *Collector) Handle(duration time.Duration, handle CollectorHandle) {
	this.duration = duration
	this.handle = handle
}

func (this *Collector) Run() {
	this.gauge = map[string]float64{}
	this.counter = map[string]int64{}
	this.collect()
	if this.duration == 0 {
		this.duration = time.Second
	}
	tick := time.NewTicker(this.duration)
	defer tick.Stop()
	for {
		select {
		case <-this.Done:
			return
		case <-tick.C:
			this.collect()
			if this.handle == nil {
				fmt.Print(this)
			} else {
				this.handle.HandleData(this.counter, this.gauge)
			}
		}
	}
}
