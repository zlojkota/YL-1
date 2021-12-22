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
	state  serverhandlers.Stater
	Done   chan bool
	Finish chan bool
	store  string
}

func (ss *FileStorageState) SendDone() {
	ss.Done <- true
}

func (ss *FileStorageState) SendFinish() {
	ss.Finish <- true
}

func (ss *FileStorageState) WaitDone() {
	<-ss.Done
}

func (ss *FileStorageState) WaitFinish() {
	<-ss.Done
}

func (ss *FileStorageState) StopStorage() {
	ss.Done <- true
}
func (ss *FileStorageState) Ping() bool {
	return false
}

func (ss *FileStorageState) Init(store string) {
	ss.Done = make(chan bool)
	ss.Finish = make(chan bool)
	ss.store = store
}

func (ss *FileStorageState) SetState(state serverhandlers.Stater) {
	ss.state = state
}

func (ss *FileStorageState) Restore() {
	file, err := os.OpenFile(ss.store, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Error(err)
		return
	}
	decoder := json.NewDecoder(file)
	mm := make(map[string]*collector.Metrics)
	err = decoder.Decode(&mm)
	if err != nil {
		log.Error(err)
		return
	}
	ss.state.SetMetricMap(mm)
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
			file, err := os.Create(ss.store)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			err = encoder.Encode(ss.state.MetricMap())
			if err != nil {
				log.Error(err)
				return
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Error(err)
				}
			}(file)
			ss.SendFinish()
			return
		case <-tick.C:
			file, err := os.Create(ss.store)
			if err != nil {
				log.Error(err)
			}
			encoder := json.NewEncoder(file)
			err = encoder.Encode(ss.state.MetricMap())
			if err != nil {
				log.Error(err)
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
