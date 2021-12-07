package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
)

const serverAddr = "http://localhost:8080"
const counter = "counter"
const gauge = "gauge"

type Agent struct {
	sendJSON bool
}

func (p *Agent) SendMetrics(metrics *[]collector.Metrics) {
	p.sendJSON = !(p.sendJSON)
	if p.sendJSON {
		for _, val := range *metrics {

			jsonData, errEnc := json.Marshal(val)
			if errEnc != nil {
				log.Error(errEnc)
				return
			}
			url := fmt.Sprintf("%s/update", serverAddr)
			body := bytes.NewReader(jsonData)
			res, err := http.Post(url, "application/json", body)
			if err != nil {
				log.Error(err)
				return
			}
			defer func() {
				err := res.Body.Close()
				if err != nil {
					log.Error(err)
					return
				}
			}()
		}
	} else {
		for _, val := range *metrics {
			var strVal string
			switch val.MType {
			case counter:
				strVal = strconv.FormatInt(*val.Delta, 10)
			case gauge:
				strVal = strconv.FormatFloat(*val.Value, 'f', -1, 64)
			}
			url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddr, val.MType, val.ID, strVal)
			res, err := http.Post(url, "text/plain", nil)
			if err != nil {
				log.Error(err)
				return
			}
			defer func() {
				err := res.Body.Close()
				if err != nil {
					log.Error(err)
					return
				}
			}()
		}
	}
}

func main() {

	var t collector.Collector
	var agent Agent
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
