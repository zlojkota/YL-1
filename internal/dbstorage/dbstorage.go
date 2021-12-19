package dbstorage

import (
	"database/sql"
	"github.com/zlojkota/YL-1/internal/collector"
	"github.com/zlojkota/YL-1/internal/hashhelper"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/gommon/log"
	"time"
)

type DataBaseStorageState struct {
	Done         chan bool
	db           *sql.DB
	store        string
	state        serverhandlers.Stater
	metricMapMux sync.Mutex
	Hasher       *hashhelper.Hasher
}

func (ss *DataBaseStorageState) SendDone() {
	ss.Done <- true
}

func (ss *DataBaseStorageState) WaitDone() {
	<-ss.Done
}

func (ss *DataBaseStorageState) Ping() bool {
	if err := ss.db.Ping(); err != nil {
		log.Error(err)
		return false
	} else {
		return true
	}
}

func (ss *DataBaseStorageState) Init(store string) {
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
}

func (ss *DataBaseStorageState) SetState(state serverhandlers.Stater) {
	ss.state = state
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
		ss.state.SetMetricMapItem(&m)
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
			return
		case <-tick.C:
			ss.SaveToStorage()
		}
	}

}

func (ss *DataBaseStorageState) SaveToStorage() {

	mm := ss.state.MetricMap()
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

func (ss *DataBaseStorageState) SaveToStorageLast() {

	dbLast, err := sql.Open("pgx", ss.store)
	if err != nil {
		panic(err)
	}
	mm := ss.state.MetricMap()
	allSaved := false
	counter := 100
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
		if !allSaved {
			log.Info("Wait save...")
			log.Info("reconnect to DB...")
			ss.db.Close()
			var err error
			ss.db, err = sql.Open("pgx", ss.store)
			if err != nil {
				log.Error(err)
			}
		}
		if counter == 0 {
			log.Error("Dont Save data.")
			allSaved = true
		} else {
			counter--
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

func (ss *DataBaseStorageState) MetricMapMuxLock() {
	ss.metricMapMux.Lock()
}

func (ss *DataBaseStorageState) MetricMapMuxUnlock() {
	ss.metricMapMux.Unlock()
}

func (ss *DataBaseStorageState) MetricMap() map[string]*collector.Metrics {
	var ret map[string]*collector.Metrics
	ret = make(map[string]*collector.Metrics)

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
		ret[m.ID] = &m
	}
	return ret
}

func (ss *DataBaseStorageState) SetMetricMap(metricMap map[string]*collector.Metrics) {

	if len(metricMap) != 0 {
		for _, val := range metricMap {
			var cnt int
			ss.db.QueryRow("SELECT count(id) FROM metrics WHERE id=$1 AND mtype=$2", val.ID, val.MType).Scan(&cnt)
			if cnt == 0 {
				ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5)", val.ID, val.MType, val.Delta, val.Value, val.Hash)
			} else {
				ss.db.Exec("UPDATE metrics set delta=$1, val=$2,hash=$3 where id=$4 AND mtype=$5", val.Delta, val.Value, val.Hash, val.ID, val.MType)
			}
		}
	}
}

func (ss *DataBaseStorageState) MetricMapItem(item string) (*collector.Metrics, bool) {
	var cnt int
	ss.db.QueryRow("SELECT count(id) FROM metrics WHERE id=$1 ", item).Scan(&cnt)
	if cnt == 0 {
		return nil, false
	} else {
		var val collector.Metrics
		ss.db.QueryRow("SELECT * FROM metrics WHERE id=$1", item).Scan(&val.ID, &val.MType, &val.Delta, &val.Value, &val.Hash)
		val.Hash = ss.state.GetHaser().Hash(&val)
		return &val, true
	}
}

func (ss *DataBaseStorageState) SetMetricMapItem(metricMap *collector.Metrics) {
	var cnt int
	ss.db.QueryRow("SELECT count(id) FROM metrics WHERE id=$1 AND mtype=$2", metricMap.ID, metricMap.MType).Scan(&cnt)
	if cnt == 0 {
		ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5)", metricMap.ID, metricMap.MType, metricMap.Delta, metricMap.Value, metricMap.Hash)
	} else {
		ss.db.Exec("UPDATE metrics set delta=$1, val=$2,hash=$3 where id=$4 AND mtype=$5", metricMap.Delta, metricMap.Value, metricMap.Hash, metricMap.ID, metricMap.MType)
	}
}

func (ss *DataBaseStorageState) GetHaser() *hashhelper.Hasher {
	return ss.Hasher
}

func (ss *DataBaseStorageState) InitHasher(hashKey string) {
	ss.Hasher = &hashhelper.Hasher{
		Key: hashKey,
	}
}

func (ss *DataBaseStorageState) StopStorage() {
	ss.db.Close()
	log.Info("Primary DB connection close")
}
