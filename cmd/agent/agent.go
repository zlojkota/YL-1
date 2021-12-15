package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/agentcollector"
	"github.com/zlojkota/YL-1/internal/collector"
)

type Worker struct {
	ServerAddr     *string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval *time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PoolInterval   *time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

var worker Worker

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

func init() {

}

func main() {

	var t collector.Collector
	var agent agentcollector.Agent

	err := env.Parse(&worker)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAA________ENV:")
	qwe, _ := json.Marshal(worker)
	fmt.Println(string(qwe))

	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		fmt.Println("ADDRESS not in ENV")
		worker.ServerAddr = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	} else {
		fmt.Println("ADDRESS IN ENV")
		_ = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	}
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); !ok {
		fmt.Println("REPORT_INTERVAL not in ENV")
		worker.ReportInterval = flag.Duration("r", 10*time.Second, "REPORT_INTERVAL")
	} else {
		fmt.Println("REPORT_INTERVAL IN ENV")
		_ = flag.Duration("r", 10*time.Second, "REPORT_INTERVAL")
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); !ok {
		fmt.Println("POLL_INTERVAL not in ENV")
		worker.PoolInterval = flag.Duration("p", 2*time.Second, "POLL_INTERVAL")
	} else {
		fmt.Println("POLL_INTERVAL IN ENV")
		_ = flag.Duration("p", 2*time.Second, "POLL_INTERVAL")
	}
	flag.Parse()

	fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAA________CMD:")
	ewq, _ := json.Marshal(worker)
	fmt.Println(string(ewq))

	agent.InitAgent(&worker, *worker.ServerAddr)
	t.Handle(*worker.PoolInterval, &agent)

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
