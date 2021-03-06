package agentcollector

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zlojkota/YL-1/internal/hashhelper"

	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
)

const srvSheme = "http://"
const srvAddr = "localhost:8080"
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

func (p *Agent) InitAgentPTX(agentCollector AgentCollector, serverAddr string) {
	p.agentCollector = agentCollector
	if len(serverAddr) == 0 {
		p.serverAddr = srvSheme + srvAddr
	} else {
		p.serverAddr = fmt.Sprintf("%s%s", srvSheme, serverAddr)
	}
	p.sendJSON = false
}

func (p *Agent) InitAgentJSON(agentCollector AgentCollector, serverAddr string) {
	p.agentCollector = agentCollector
	if len(serverAddr) == 0 {
		p.serverAddr = srvSheme + srvAddr
	} else {
		p.serverAddr = fmt.Sprintf("%s%s", srvSheme, serverAddr)
	}
	p.sendJSON = true
}

func (p *Agent) MakeRequestJSON(metrics *collector.Metrics) (*http.Request, error) {

	metrics.Hash = p.hasher.Hash(metrics)
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

func (p *Agent) MakeRequestJSONBatch(metrics []*collector.Metrics) (*http.Request, error) {

	for _, val := range metrics {
		val.Hash = p.hasher.Hash(val)
	}

	jsonData, errEnc := json.Marshal(metrics)
	if errEnc != nil {
		return nil, errEnc
	}
	url := fmt.Sprintf("%s/updates/", p.serverAddr)
	body := bytes.NewReader(jsonData)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = make(http.Header)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (p *Agent) MakeRequestPLTX(metrics *collector.Metrics) (*http.Request, error) {
	var strVal string
	switch metrics.MType {
	case counter:
		if metrics.Delta == nil {
			return nil, errors.New("NIL counter value")
		}
		strVal = strconv.FormatUint(*metrics.Delta, 10)
	case gauge:
		if metrics.Value == nil {
			return nil, errors.New("NIL gauge value")
		}
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

func (p *Agent) MakeRequest(metrics []*collector.Metrics) {
	if p.agentCollector == nil {
		log.Error("No AgentCollector Init!")
		return
	}
	//for _, val := range metrics {
	//	var (
	//		req *http.Request
	//		err error
	//	)
	//	if p.sendJSON {
	//		req, err = p.MakeRequestJSON(val)
	//		if err != nil {
	//			log.Error(err)
	//			return
	//		}
	//	} else {
	//		req, err = p.MakeRequestPLTX(val)
	//		if err != nil {
	//			log.Error(err)
	//			return
	//		}
	//	}
	//	p.agentCollector.RequestSend(req)
	//}

	req, err := p.MakeRequestJSONBatch(metrics)
	if err != nil {
		log.Error(err)
		return
	}
	p.agentCollector.RequestSend(req)
}
