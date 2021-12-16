package collector

import (
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"time"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции

}

type CollectorHandle interface {
	SendMetrics(metrics *[]Metrics)
}

type Collector struct {
	handle       CollectorHandle
	poolinterval time.Duration
	Done         chan bool
	Metrics      []Metrics
	counter      int64
	randomvalue  float64
	rtm          runtime.MemStats
	rtmFloat     map[string]*float64
}

func (col *Collector) Handle(poolinterval time.Duration, handle CollectorHandle) {
	col.poolinterval = poolinterval
	col.handle = handle
	col.Done = make(chan bool)
	col.rtmFloat = make(map[string]*float64)
	col.Metrics = append(col.Metrics, Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: &col.counter,
	})
	col.Metrics = append(col.Metrics, Metrics{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &col.randomvalue,
	})
	ref := reflect.ValueOf(col.rtm)
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Field(i)
		temp := new(float64)
		switch field.Type().Name() {
		case "float64":
			*temp = field.Interface().(float64)
		case "uint64":
			*temp = float64(field.Interface().(uint64))
		case "uint32":
			*temp = float64(field.Interface().(uint32))
		default:
			continue
		}
		col.rtmFloat[ref.Type().Field(i).Name] = temp
		col.Metrics = append(col.Metrics, Metrics{
			ID:    ref.Type().Field(i).Name,
			MType: "gauge",
			Value: temp,
		})
	}
}

func (col *Collector) Run() {

	if col.poolinterval == 0 {
		col.poolinterval = time.Second
	}
	tick := time.NewTicker(col.poolinterval)
	defer tick.Stop()
	for {
		runtime.ReadMemStats(&col.rtm)
		ref := reflect.ValueOf(col.rtm)
		for i := 0; i < ref.NumField(); i++ {
			field := ref.Field(i)
			var temp float64
			switch field.Type().Name() {
			case "float64":
				temp = field.Interface().(float64)
			case "uint64":
				temp = float64(field.Interface().(uint64))
			case "uint32":
				temp = float64(field.Interface().(uint32))
			default:
				continue
			}
			r := col.rtmFloat[ref.Type().Field(i).Name]
			*r = temp
		}
		select {
		case <-col.Done:
			return
		case <-tick.C:

			if col.handle == nil {
				log.Println(col.Metrics)
			} else {
				col.handle.SendMetrics(&col.Metrics)
			}
		}
		col.counter++
		col.randomvalue = rand.Float64()
	}
}
