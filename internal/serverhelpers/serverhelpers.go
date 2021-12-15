package serverhelpers

import (
	"encoding/json"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"os"
	"time"
)

type InterfaceStorageState interface {
	GetterMetrics() map[string]*serverhandlers.Metrics
	SetterMetrics(map[string]*serverhandlers.Metrics)
}

type StorageState struct {
	serverHandler *serverhandlers.ServerHandler
	Done          chan bool
}

func (ssh *StorageState) SetServerHandler(serverHandler *serverhandlers.ServerHandler) {
	ssh.serverHandler = serverHandler
}

func (ssh *StorageState) Restore(storeFile string) {
	file, err := os.OpenFile(storeFile, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Error(err)
	}
	decoder := json.NewDecoder(file)
	mm := make(map[string]*serverhandlers.Metrics)
	decoder.Decode(&mm)
	ssh.serverHandler.MetricMap = mm
	file.Close()
}

func (ssh *StorageState) Run(storeInterval time.Duration, storeFile string) {
	tick := time.NewTicker(storeInterval)
	ssh.Done = make(chan bool)
	defer tick.Stop()
	for {
		select {
		case <-ssh.Done:
			file, err := os.Create(storeFile)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			encoder.Encode(&ssh.serverHandler.MetricMap)
			file.Close()
			return
		case <-tick.C:
			file, err := os.Create(storeFile)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			encoder.Encode(&ssh.serverHandler.MetricMap)
			file.Close()
		}
	}

}
