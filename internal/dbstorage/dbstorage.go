package dbstorage

import (
	"database/sql"
	"fmt"
	"github.com/zlojkota/YL-1/internal/collector"

	//_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
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
	fmt.Println(store)
	if err != nil {
		panic(err)
	}
	if _, err := ss.db.Exec("create table if not exists metrics( id varchar(32),mtype varchar(32), delta int, val double precision, hash varchar(256))"); err != nil {
		panic(err)
	}
}

func (ss *DataBaseStorageState) Restore() {

	rows, err := ss.db.Query("SELECT * FROM metrics")
	if err != nil {
		panic(err)
	}
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
			ss.SaveToStorage()
			var res string
			ss.db.QueryRow("SELECT id FROM metrics where id like 'PopulateCounter%'").Scan(&res)
			fmt.Println("_____________", res, "___________")
			ss.db.Close()
			ss.Done <- true
			return
		case <-tick.C:
			ss.SaveToStorage()
		}
	}

}

func (ss DataBaseStorageState) SaveToStorage() {

	//fmt.Println("----------------------------Start save to db-------------------------------------------")
	mm := ss.ServerHandler.MetricMap()
	for _, val := range mm {
		var cnt int
		ss.db.QueryRow("SELECT count(id) FROM metrics WHERE id=$1", val.ID).Scan(&cnt)
		if cnt == 0 {
			ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5)", val.ID, val.MType, val.Delta, val.Value, val.Hash)
		} else {
			ss.db.Exec("UPDATE metrics set delta=$1, val=$2,hash=$3 where id=$4", val.Delta, val.Value, val.Hash, val.ID)
		}
		//fmt.Println("++++++", val.ID)
	}
	//fmt.Println("============================Stop save to db============================================")

}
