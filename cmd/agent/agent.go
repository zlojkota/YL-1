package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
)

type Agent struct {
	poolinterval time.Duration
}

func (p *Agent) agentInit(poolinterval time.Duration) {
	p.poolinterval = poolinterval
}

func (p *Agent) SendMetrics(counter map[string]int64, gauge map[string]float64) {
	const serverAddr = "http://localhost:8080"
	for key, val := range gauge {
		strval := strconv.FormatFloat(val, 'f', -1, 64)
		url := fmt.Sprintf("%s/update/gauge/%s/%s", serverAddr, key, strval)
		res, err := http.Post(url, "text/plain", nil)
		if err != nil {
			log.Error(err)
			return
		}
		defer res.Body.Close()
	}
	for key, val := range counter {
		strval := strconv.FormatInt(val, 10)
		res, err := http.Post("http://localhost:8080/update/counter/"+string(key)+"/"+strval, "text/plain", nil)
		if err != nil {
			log.Error(err)
			return
		}
		defer res.Body.Close()
	}

}

func main() {

	var t collector.Collector
	var agent Agent
	agent.agentInit(time.Second)
	t.Handle(2*time.Second, &agent)
	t.Run()

}
