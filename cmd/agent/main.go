package main

import (
	"fmt"
	"github.com/zlojkota/YL-1"
	"net/http"
	"time"
)

type Agent struct {
	duration time.Duration
}

func (p *Agent) agentInit(duration time.Duration) {
	p.duration = duration
}

func (p *Agent) HandleData(counter map[string]int64, gauge map[string]float64) {
	for key, val := range gauge {
		strval := fmt.Sprintf("%f", val)
		_, err := http.Post("http://localhost:8080/update/gauge/"+string(key)+"/"+strval, "text/plain", nil)
		if err != nil {
			fmt.Println(err)
		}
	}
	for key, val := range counter {
		strval := fmt.Sprintf("%d", val)
		_, err := http.Post("http://localhost:8080/update/counter/"+string(key)+"/"+strval, "text/plain", nil)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func main() {

	var t collector.Collector
	var agent Agent
	agent.agentInit(time.Second)
	t.Handle(2*time.Second, &agent)
	t.Run()

}
