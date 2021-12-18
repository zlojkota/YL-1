package filestorage

import (
	"encoding/json"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"os"
	"time"
)

type FileStorageState struct {
	ServerHandler *serverhandlers.ServerHandler
	Done          chan bool
}

func (ss FileStorageState) SendDone() {
	ss.Done <- true
}

func (ss FileStorageState) WaitDone() {
	<-ss.Done
}

func (ss *FileStorageState) SetServerHandler(serverHandler *serverhandlers.ServerHandler) {
	ss.ServerHandler = serverHandler
	ss.Done = make(chan bool)
}

func (ss *FileStorageState) Restore(storeFile string) {
	file, err := os.OpenFile(storeFile, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Error(err)
	}
	decoder := json.NewDecoder(file)
	mm := make(map[string]*collector.Metrics)
	decoder.Decode(&mm)
	ss.ServerHandler.SetMetricMap(mm)
	defer file.Close()
}

func (ss *FileStorageState) Run(storeInterval time.Duration, storeFile string) {
	tick := time.NewTicker(storeInterval)
	defer tick.Stop()
	for {
		select {
		case <-ss.Done:
			file, err := os.Create(storeFile)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			encoder.Encode(ss.ServerHandler.MetricMap())
			defer file.Close()
			ss.Done <- true
			return
		case <-tick.C:
			file, err := os.Create(storeFile)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			encoder.Encode(ss.ServerHandler.MetricMap())
			defer file.Close()
		}
	}

}
