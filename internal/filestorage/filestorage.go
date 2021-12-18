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
	Store         string
}

func (ss FileStorageState) SendDone() {
	ss.Done <- true
}

func (ss FileStorageState) WaitDone() {
	<-ss.Done
}

func (ss FileStorageState) Ping() bool {
	return false
}

func (ss *FileStorageState) Init(serverHandler *serverhandlers.ServerHandler, store string) {
	ss.ServerHandler = serverHandler
	ss.Done = make(chan bool)
	ss.Store = store
}

func (ss *FileStorageState) Restore() {
	file, err := os.OpenFile(ss.Store, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Error(err)
	}
	decoder := json.NewDecoder(file)
	mm := make(map[string]*collector.Metrics)
	err = decoder.Decode(&mm)
	if err != nil {
		return
	}
	ss.ServerHandler.SetMetricMap(mm)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}(file)
}

func (ss *FileStorageState) Run(storeInterval time.Duration) {
	tick := time.NewTicker(storeInterval)
	defer tick.Stop()
	for {
		select {
		case <-ss.Done:
			file, err := os.Create(ss.Store)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			err = encoder.Encode(ss.ServerHandler.MetricMap())
			if err != nil {
				return
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Error(err)
				}
			}(file)
			ss.Done <- true
			return
		case <-tick.C:
			file, err := os.Create(ss.Store)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			err = encoder.Encode(ss.ServerHandler.MetricMap())
			if err != nil {
				return
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Error(err)
				}
			}(file)
		}
	}

}
