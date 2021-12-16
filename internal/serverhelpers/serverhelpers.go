package serverhelpers

import (
	"encoding/json"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"os"
	"time"
)

type StorageState struct {
	ServerHandler *serverhandlers.ServerHandler
	Done          chan bool
}

func (ssh *StorageState) SetServerHandler(serverHandler *serverhandlers.ServerHandler) {
	ssh.ServerHandler = serverHandler
	ssh.Done = make(chan bool)
}

func (ssh *StorageState) Restore(storeFile string) {
	file, err := os.OpenFile(storeFile, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Error(err)
	}
	decoder := json.NewDecoder(file)
	mm := make(map[string]*collector.Metrics)
	decoder.Decode(&mm)
	ssh.ServerHandler.SetMetricMap(mm)
	defer file.Close()
}

func (ssh *StorageState) Run(storeInterval time.Duration, storeFile string) {
	tick := time.NewTicker(storeInterval)
	defer tick.Stop()
	for {
		select {
		case <-ssh.Done:
			return
		case <-tick.C:
			file, err := os.Create(storeFile)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			encoder.Encode(ssh.ServerHandler.MetricMap())
			defer file.Close()
		}
	}

}
