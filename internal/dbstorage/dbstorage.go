package dbstorage

import (
	"database/sql"
	"github.com/zlojkota/YL-1/internal/collector"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"time"
)

type DataBaseStorageState struct {
	ServerHandler *serverhandlers.ServerHandler
	Done          chan bool
	db            *sql.DB
	store         string
	stopped       bool
}

func (ss DataBaseStorageState) SendDone() {
	ss.Done <- true
}

func (ss DataBaseStorageState) WaitDone() {
	if !ss.stopped {
		<-ss.Done
		ss.stopped = true
	}
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
	if _, err := ss.db.Exec("create table if not exists metrics( id varchar(256),mtype varchar(256), delta int, val double precision, hash varchar(256))"); err != nil {
		panic(err)
	}
	ss.store = store
	ss.stopped = false
}

func (ss *DataBaseStorageState) Restore() {

	rows, err := ss.db.Query("SELECT * FROM metrics")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	for rows.Next() {
		var m collector.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
		if err != nil {
			log.Error(err)
		}
		ss.ServerHandler.SetMetricMapItem(&m)
	}
}

func (ss *DataBaseStorageState) Run(storeInterval time.Duration) {
	tick := time.NewTicker(storeInterval)
	defer tick.Stop()
	for {
		select {
		case <-ss.Done:
			ss.SaveToStorageLast()
			ss.Done <- true
			ss.stopped = true
			return
		case <-tick.C:
			ss.SaveToStorage()
		}
	}

}

func (ss DataBaseStorageState) SaveToStorage() {

	mm := ss.ServerHandler.MetricMap()
	for _, val := range mm {
		var cnt int
		ss.db.QueryRow("SELECT count(id) FROM metrics WHERE id=$1 AND mtype=$2", val.ID, val.MType).Scan(&cnt)
		if cnt == 0 {
			ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5)", val.ID, val.MType, val.Delta, val.Value, val.Hash)
		} else {
			ss.db.Exec("UPDATE metrics set delta=$1, val=$2,hash=$3 where id=$4 AND mtype=$5", val.Delta, val.Value, val.Hash, val.ID, val.MType)
		}
	}
}

func (ss DataBaseStorageState) SaveToStorageLast() {

	dbLast, err := sql.Open("pgx", ss.store)
	if err != nil {
		panic(err)
	}
	mm := ss.ServerHandler.MetricMap()
	allSaved := false
	counter := 10
	for !allSaved {
		allSaved = true
		for _, val := range mm {
			var cnt int
			ss.db.QueryRow("SELECT count(id) FROM metrics WHERE id=$1 AND mtype=$2", val.ID, val.MType).Scan(&cnt)
			if cnt == 0 {
				ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5)", val.ID, val.MType, val.Delta, val.Value, val.Hash)
			} else {
				ss.db.Exec("UPDATE metrics set delta=$1, val=$2,hash=$3 where id=$4 AND mtype=$5", val.Delta, val.Value, val.Hash, val.ID, val.MType)
			}
		}
		for _, val := range mm {
			var cnt int
			dbLast.QueryRow("SELECT count(id) FROM metrics WHERE id=$1 AND mtype=$2 AND ((delta is null and val=$3) or (delta=$4 and val is null)) AND hash=$5", val.ID, val.MType, val.Value, val.Delta, val.Hash).Scan(&cnt)
			if cnt == 0 {
				allSaved = false
			}
		}
		log.Info("Wait save...")
		if counter == 0 {
			log.Error("Dont Save data.")
			allSaved = true
			ss.Done <- true
		}
	}
	if counter != 0 {
		log.Info("Saved last Data to DB")
	}
	ss.db.Close()
	log.Info("Primary DB connection close")
	dbLast.Close()
	log.Info("Testing DB connection close")
}
