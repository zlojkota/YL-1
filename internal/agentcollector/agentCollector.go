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
	RequestServe(req *http.Request)
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

func (p *Agent) SendMetrics(metrics *[]collector.Metrics) {
	if p.agentCollector == nil {
		log.Error("No AgentCollector Init!")
		return
	}
	p.sendJSON = !(p.sendJSON)
	if p.sendJSON {
		for _, val := range *metrics {
			switch val.MType {
			case counter:
				val.Hash = p.hasher.Hash(&val)
			case gauge:
				val.Hash = p.hasher.Hash(&val)
			}
			jsonData, errEnc := json.Marshal(val)
			if errEnc != nil {
				log.Error(errEnc)
				return
			}
			url := fmt.Sprintf("%s/update/", p.serverAddr)
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
			url := fmt.Sprintf("%s/update/%s/%s/%s", p.serverAddr, val.MType, val.ID, strVal)
			res, err := http.NewRequest(http.MethodPost, url, nil)
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
