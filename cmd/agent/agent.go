package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/agentcollector"
	"github.com/zlojkota/YL-1/internal/collector"
)

type Worker struct {
}

func (p *Worker) RequestServe(req *http.Request) {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer res.Body.Close()
	log.Print(req.URL.Path)
}

func main() {

	var t collector.Collector
	var agent agentcollector.Agent
	var worker Worker
	agent.InitAgent(&worker)
	t.Handle(2*time.Second, &agent)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan
		log.Error("Stopping")
		t.Done <- true
	}()
	t.Run()

}
