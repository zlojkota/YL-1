package collector

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/shirou/gopsutil/v3/mem"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *uint64  `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции

}

type CollectorHandle interface {
	MakeRequest(metrics []*Metrics)
}

type Collector struct {
	handle       CollectorHandle
	poolinterval time.Duration
	Done         chan bool
	Metrics      []*Metrics
	counter      uint64
	randomvalue  float64
	memTotal     float64
	memFree      float64
	rtm          runtime.MemStats
	rtmFloat     map[string]*float64
	procFloat    map[string]*float64
	procIdles    *[]uint64
	procTotals   *[]uint64
}

func (col *Collector) Handle(poolinterval time.Duration, handle CollectorHandle) {
	col.poolinterval = poolinterval
	col.handle = handle
	col.Done = make(chan bool)
	col.rtmFloat = make(map[string]*float64)
	col.Metrics = append(col.Metrics, &Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: &col.counter,
	})
	col.Metrics = append(col.Metrics, &Metrics{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &col.randomvalue,
	})
	col.Metrics = append(col.Metrics, &Metrics{
		ID:    "TotalMemory",
		MType: "gauge",
		Value: &col.memTotal,
	})
	col.Metrics = append(col.Metrics, &Metrics{
		ID:    "FreeMemory",
		MType: "gauge",
		Value: &col.memFree,
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
		col.Metrics = append(col.Metrics, &Metrics{
			ID:    ref.Type().Field(i).Name,
			MType: "gauge",
			Value: temp,
		})
	}
	mem.VirtualMemory()
}

func (col Collector) collectRuntime(ctx context.Context) {
	if col.poolinterval == 0 {
		col.poolinterval = time.Second
	}
	tick := time.NewTicker(col.poolinterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
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
			col.counter++
			col.randomvalue = rand.Float64()
		}
	}
}

func (col Collector) getCpuUtilization() {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	var totals []uint64
	var idles []uint64
	cpuId := 0
	for _, line := range lines {
		fields := strings.Fields(line)
		str := ""
		if len(fields) != 0 {
			str = fields[0]
		}
		matched, _ := regexp.MatchString("cpu[0-9]+", str)
		if matched {
			numFields := len(fields)
			total := uint64(0)
			idle := uint64(0)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			totals = append(totals, total)
			idles = append(idles, idle)
			if col.procTotals != nil {
				idleTicks := float64(idle - (*col.procIdles)[cpuId])
				totalTicks := float64(total - (*col.procTotals)[cpuId])
				cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
				fmt.Printf("CPU%d usage is %f%%\n", cpuId, cpuUsage)
			}
			cpuId++
		}
	}
	col.procIdles = &idles
	col.procTotals = &totals
}

func (col Collector) collectProc(ctx context.Context) {
	if col.poolinterval == 0 {
		col.poolinterval = time.Second
	}
	tick := time.NewTicker(col.poolinterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			memProc, _ := mem.VirtualMemory()
			col.memTotal = float64(memProc.Total)
			col.memFree = float64(memProc.Free)
			col.getCpuUtilization()
		}
	}
}

func (col Collector) sendMertics(ctx context.Context) {
	if col.poolinterval == 0 {
		col.poolinterval = time.Second
	}
	tick := time.NewTicker(col.poolinterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if col.handle == nil {
				log.Println(col.Metrics)
			} else {
				col.handle.MakeRequest(col.Metrics)
			}
		}
	}
}

func (col *Collector) Run() {

	bctx, bctxCancel := context.WithCancel(context.Background())
	go col.collectRuntime(bctx)
	go col.collectProc(bctx)
	go col.sendMertics(bctx)

	select {
	case <-col.Done:
		bctxCancel()
		return
	}
}
