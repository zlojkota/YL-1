package main

import (
	"bytes"
	"flag"
	"io/ioutil"
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
	HashKey        *string        `env:"KEY" envDefault:""`
}

func (p *Worker) RequestServe(req *http.Request) {

	var buf []byte
	if req.Body != nil {
		buff, bodyErr := ioutil.ReadAll(req.Body)
		if bodyErr != nil {
			log.Print("bodyErr ", bodyErr.Error())
			return
		}
		buf = buff
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	} else {
		buf = []byte("nil")
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		//log.Error(err)
		return
	} else {
		if res.StatusCode == http.StatusBadRequest {
			log.Print("URL:", req.URL.Path, ", Body:", string(buf))
		}
	}
	defer res.Body.Close()

}

func main() {

	var t collector.Collector
	var agent agentcollector.Agent
	var worker Worker

	err := env.Parse(&worker)
	if err != nil {
		log.Fatal(err)
	}

	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		worker.ServerAddr = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	} else {
		_ = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	}
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); !ok {
		worker.ReportInterval = flag.Duration("r", 10*time.Second, "REPORT_INTERVAL")
	} else {
		_ = flag.Duration("r", 10*time.Second, "REPORT_INTERVAL")
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); !ok {
		worker.PoolInterval = flag.Duration("p", 2*time.Second, "POLL_INTERVAL")
	} else {
		_ = flag.Duration("p", 2*time.Second, "POLL_INTERVAL")
	}
	if _, ok := os.LookupEnv("KEY"); !ok {
		worker.HashKey = flag.String("k", "", "KEY")
	} else {
		_ = flag.String("k", "", "k")
	}
	flag.Parse()

	agent.InitAgent(&worker, *worker.ServerAddr)
	agent.SetHasher(*worker.HashKey)

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
