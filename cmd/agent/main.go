package main

import (
	"fmt"
	"github.com/zlojkota/go-musthave-devops-tpl/internal/collector"
	"time"
)

type Agent struct {
	duration time.Duration
}

func (p *Agent) agentInit(duration time.Duration) {
	p.duration = duration
}

func (p *Agent) HandleData(data collector.Metric) {
	fmt.Println("Collected DATA:", data)
}

func main() {

	var t collector.Collector
	var agent Agent
	agent.agentInit(time.Second)
	t.Handle(2*time.Second, &agent)
	t.Run()

}
