package dbstorage

import (
	"database/sql"
	_ "github.com/jackc/pgx/v4"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"time"
)

type DataBaseStorageState struct {
	ServerHandler *serverhandlers.ServerHandler
	Done          chan bool
	db            *sql.DB
}

func (ss DataBaseStorageState) SendDone() {
	ss.Done <- true
}

func (ss DataBaseStorageState) WaitDone() {
	<-ss.Done
}

func (ss DataBaseStorageState) Ping() bool {
	if err := ss.db.Ping(); err != nil {
		log.Error(err)
		return false
	} else {
		return true
	}
}

func (ss *DataBaseStorageState) Init(serverHandler *serverhandlers.ServerHandler, store string) {
	ss.ServerHandler = serverHandler
	ss.Done = make(chan bool)
	var err error
	ss.db, err = sql.Open("pgx", store)
	if err != nil {
		panic(err)
	}

}

func (ss *DataBaseStorageState) Restore() {

}

func (ss *DataBaseStorageState) Run(storeInterval time.Duration) {
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
