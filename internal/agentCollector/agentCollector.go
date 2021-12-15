package agentCollector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
)

const serverAddr = "http://localhost:8080"
const counter = "counter"
const gauge = "gauge"

type AgentCollector interface {
	RequestServe(req *http.Request)
}

type Agent struct {
	sendJSON       bool
	agentCollector AgentCollector
}

func (p *Agent) InitAgent(agentCollector AgentCollector) {
	p.agentCollector = agentCollector
}

func (p *Agent) SendMetrics(metrics *[]collector.Metrics) {
	if p.agentCollector == nil {
		log.Error("No AgentCollector Init!")
		return
	}
	p.sendJSON = !(p.sendJSON)
	if p.sendJSON {
		for _, val := range *metrics {
			jsonData, errEnc := json.Marshal(val)
			if errEnc != nil {
				log.Error(errEnc)
				return
			}
			url := fmt.Sprintf("%s/update/", serverAddr)
			body := bytes.NewReader(jsonData)
			res, err := http.NewRequest(http.MethodPost, url, body)
			if err != nil {
				log.Error(err)
				return
			}
			res.Header = make(http.Header)
			res.Header.Set("Content-Type", "application/json")
			p.agentCollector.RequestServe(res)
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
			res, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				log.Error(err)
				return
			}
			path := fmt.Sprintf("/update/%s/%s/%s", val.MType, val.ID, strVal)
			res.URL.Path = path
			res.Header = make(http.Header)
			res.Header.Set("Content-Type", "text/plain")
			p.agentCollector.RequestServe(res)
		}
	}
}
