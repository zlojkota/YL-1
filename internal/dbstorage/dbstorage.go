package dbstorage

import (
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/collector"
	"time"

	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbStorageState struct {
	ServerHandler *serverhandlers.ServerHandler
	Done          chan bool
	Db            *gorm.DB
}

type MetricStruct struct {
	gorm.Model
	collector.Metrics
}

func (ss DbStorageState) SendDone() {
	ss.Done <- true
}

func (ss DbStorageState) WaitDone() {
	<-ss.Done
}

func (ss DbStorageState) Ping() bool {
	db, err := ss.Db.DB()
	if err != nil {
		log.Error(err)
		return false
	}
	if err := db.Ping(); err != nil {
		log.Error(err)
		return false
	} else {
		return true
	}
}

func (ss *DbStorageState) Init(serverHandler *serverhandlers.ServerHandler, store string) {
	ss.ServerHandler = serverHandler
	ss.Done = make(chan bool)
	var err error
	ss.Db, err = gorm.Open(postgres.Open(store), &gorm.Config{})
	if err != nil {
		panic(err)
	}

}

func (ss *DbStorageState) Restore() {

}

func (ss *DbStorageState) Run(storeInterval time.Duration) {
	tick := time.NewTicker(storeInterval)
	defer tick.Stop()
	for {
		select {
		case <-ss.Done:

			ss.Done <- true
			return
		case <-tick.C:

		}
	}

}
