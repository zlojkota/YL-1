package agentcollector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zlojkota/YL-1/internal/hashhelper"
	"net/http"
	"strconv"

	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
)

//const serverAddr = "http://localhost:8080"
const counter = "counter"
const gauge = "gauge"

type AgentCollector interface {
	RequestSend(req *http.Request)
}
type Agent struct {
	sendJSON       bool
	agentCollector AgentCollector
	serverAddr     string
	hasher         hashhelper.Hasher
}

func (p *Agent) SetHasher(key string) {
	p.hasher.SetKey(key)
}

func (p *Agent) InitAgent(agentCollector AgentCollector, serverAddr ...string) {
	p.agentCollector = agentCollector
	if len(serverAddr) == 0 {
		p.serverAddr = "http://localhost:8080"
	} else {
		if serverAddr[0] == "" {
			p.serverAddr = "http://localhost:8080"
		} else {
			p.serverAddr = fmt.Sprintf("http://%s", serverAddr[0])
		}
	}

}

func (p *Agent) MakeRequestJSON(metrics collector.Metrics) (*http.Request, error) {
	switch metrics.MType {
	case counter:
		metrics.Hash = p.hasher.Hash(&metrics)
	case gauge:
		metrics.Hash = p.hasher.Hash(&metrics)
	}
	jsonData, errEnc := json.Marshal(metrics)
	if errEnc != nil {
		return nil, errEnc
	}
	url := fmt.Sprintf("%s/update/", p.serverAddr)
	body := bytes.NewReader(jsonData)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = make(http.Header)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (p *Agent) MakeRequestPLTX(metrics collector.Metrics) (*http.Request, error) {
	var strVal string
	switch metrics.MType {
	case counter:
		strVal = strconv.FormatInt(*metrics.Delta, 10)
	case gauge:
		strVal = strconv.FormatFloat(*metrics.Value, 'f', -1, 64)
	}
	url := fmt.Sprintf("%s/update/%s/%s/%s", p.serverAddr, metrics.MType, metrics.ID, strVal)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/update/%s/%s/%s", metrics.MType, metrics.ID, strVal)
	req.URL.Path = path
	req.Header = make(http.Header)
	req.Header.Set("Content-Type", "text/plain")
	return req, nil
}

func (p *Agent) MakeRequest(metrics *[]collector.Metrics) {
	if p.agentCollector == nil {
		log.Error("No AgentCollector Init!")
		return
	}
	p.sendJSON = !p.sendJSON
	if p.sendJSON {
		for _, val := range *metrics {
			req, err := p.MakeRequestJSON(val)
			if err != nil {
				log.Error(err)
				return
			}
			p.agentCollector.RequestSend(req)
		}
	} else {
		for _, val := range *metrics {
			req, err := p.MakeRequestPLTX(val)
			if err != nil {
				log.Error(err)
				return
			}
			p.agentCollector.RequestSend(req)
		}
	}
}
