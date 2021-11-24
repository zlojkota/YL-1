package collector

import (
	"fmt"
	//	"fmt"
	"math/rand"
	"runtime"
	"time"
	//	"time"
)

type gauge float64
type counter int64

type CollectorHandle interface {
	HandleData(data Metric)
}

type Metric struct {
	Alloc         gauge   `json:"alloc,omitempty"`
	BuckHashSys   gauge   `json:"buck_hash_sys,omitempty"`
	Frees         gauge   `json:"frees,omitempty"`
	GCCPUFraction gauge   `json:"gccpu_fraction,omitempty"`
	GCSys         gauge   `json:"gc_sys,omitempty"`
	HeapAlloc     gauge   `json:"heap_alloc,omitempty"`
	HeapIdle      gauge   `json:"heap_idle,omitempty"`
	HeapInuse     gauge   `json:"heap_inuse,omitempty"`
	HeapObjects   gauge   `json:"heap_objects,omitempty"`
	HeapReleased  gauge   `json:"heap_released,omitempty"`
	HeapSys       gauge   `json:"heap_sys,omitempty"`
	LastGC        gauge   `json:"last_gc,omitempty"`
	Lookups       gauge   `json:"lookups,omitempty"`
	MCacheInuse   gauge   `json:"m_cache_inuse,omitempty"`
	MCacheSys     gauge   `json:"m_cache_sys,omitempty"`
	MSpanInuse    gauge   `json:"m_span_inuse,omitempty"`
	MSpanSys      gauge   `json:"m_span_sys,omitempty"`
	Mallocs       gauge   `json:"mallocs,omitempty"`
	NextGC        gauge   `json:"next_gc,omitempty"`
	NumForcedGC   gauge   `json:"num_forced_gc,omitempty"`
	NumGC         gauge   `json:"num_gc,omitempty"`
	OtherSys      gauge   `json:"other_sys,omitempty"`
	PauseTotalNs  gauge   `json:"pause_total_ns,omitempty"`
	StackInuse    gauge   `json:"stack_inuse,omitempty"`
	StackSys      gauge   `json:"stack_sys,omitempty"`
	Sys           gauge   `json:"sys,omitempty"`
	PollCount     counter `json:"poll_count,omitempty"`
	RandomValue   gauge   `json:"random_value,omitempty"`
}

type Collector struct {
	m        Metric
	handle   CollectorHandle
	duration time.Duration
	Done     <-chan struct{}
}

func (m *Collector) collect() {

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	m.m.Alloc = gauge(rtm.Alloc)
	m.m.BuckHashSys = gauge(rtm.BuckHashSys)
	m.m.Frees = gauge(rtm.Frees)
	m.m.GCCPUFraction = gauge(rtm.GCCPUFraction)
	m.m.GCSys = gauge(rtm.GCSys)
	m.m.HeapAlloc = gauge(rtm.HeapAlloc)
	m.m.HeapIdle = gauge(rtm.HeapIdle)
	m.m.HeapInuse = gauge(rtm.HeapInuse)
	m.m.HeapObjects = gauge(rtm.HeapObjects)
	m.m.HeapReleased = gauge(rtm.HeapReleased)
	m.m.HeapSys = gauge(rtm.HeapSys)
	m.m.LastGC = gauge(rtm.LastGC)
	m.m.Lookups = gauge(rtm.Lookups)
	m.m.MCacheInuse = gauge(rtm.MCacheInuse)
	m.m.MCacheSys = gauge(rtm.MCacheSys)
	m.m.MSpanInuse = gauge(rtm.MSpanInuse)
	m.m.MSpanSys = gauge(rtm.MSpanSys)
	m.m.Mallocs = gauge(rtm.Mallocs)
	m.m.NextGC = gauge(rtm.NextGC)
	m.m.NumForcedGC = gauge(rtm.NumForcedGC)
	m.m.NumGC = gauge(rtm.NumGC)
	m.m.OtherSys = gauge(rtm.OtherSys)
	m.m.PauseTotalNs = gauge(rtm.PauseTotalNs)
	m.m.StackInuse = gauge(rtm.StackInuse)
	m.m.StackSys = gauge(rtm.StackSys)
	m.m.Sys = gauge(rtm.Sys)
	m.m.PollCount++
	m.m.RandomValue = gauge(rand.Float32())

}

func (p *Collector) String() string {
	return fmt.Sprintf("%v", p.m)
}

func (p *Collector) Handle(duration time.Duration, handle CollectorHandle) {
	p.duration = duration
	p.handle = handle
}

func (p *Collector) Run() {
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
				p.handle.HandleData(p.m)
			}
		}
	}
}
