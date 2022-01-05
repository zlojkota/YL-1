package dbstorage

import (
	"database/sql"
	"errors"
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
	Finish       chan bool
	db           *sql.DB
	store        string
	state        serverhandlers.State
	metricMapMux sync.Mutex
	Hasher       *hashhelper.Hasher
}

func (ss *DataBaseStorageState) SendDone() {
	ss.Done <- true
}

func (ss *DataBaseStorageState) SendFinish() {
	ss.Finish <- true
}

func (ss *DataBaseStorageState) WaitDone() {
	<-ss.Done
}

func (ss *DataBaseStorageState) WaitFinish() {
	<-ss.Finish
}

func (ss *DataBaseStorageState) Ping() bool {
	if err := ss.db.Ping(); err != nil {
		log.Error(err)
		return false
	}

	return true
}

func (ss *DataBaseStorageState) Init(store string) {
	ss.Done = make(chan bool)
	ss.Finish = make(chan bool)
	var err error
	ss.db, err = sql.Open("pgx", store)
	if err != nil {
		panic(err)
	}
	if _, err := ss.db.Exec("create table if not exists metrics( id varchar(256),mtype varchar(256), delta bigint, val double precision, hash varchar(256) DEFAULT '')"); err != nil {
		panic(err)
	}
	if _, err := ss.db.Exec("create unique index if not exists metrics_id ON metrics(id,mtype);\n"); err != nil {
		panic(err)
	}
	ss.store = store
}

func (ss *DataBaseStorageState) SetState(state serverhandlers.State) {
	ss.state = state
}

func (ss *DataBaseStorageState) Restore() {

	rows, err := ss.db.Query("SELECT * FROM metrics")
	if err != nil {
		panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()
	if rows.Err() != nil {
		log.Error(err)
		panic(errors.New("restore data ERROR"))
	}
	for rows.Next() {
		var m collector.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
		if err != nil {
			log.Error(err)
			continue
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
			log.Info("Saving data and stopping db")
			ss.SaveToStorageLast()
			ss.SendFinish()
			return
		case <-tick.C:
			ss.SaveToStorage()
		}
	}

}

func (ss *DataBaseStorageState) SaveToStorage() {
	mm := ss.state.MetricMap()
	for _, val := range mm {
		_, err := ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5) ON CONFLICT (id,mtype) DO UPDATE set delta=$3, val=$4, hash=$5", val.ID, val.MType, val.Delta, val.Value, val.Hash)
		if err != nil {
			log.Error(err)
		}
	}
}

func (ss *DataBaseStorageState) SaveToStorageLast() {
	mm := ss.state.MetricMap()
	for _, val := range mm {
		_, err := ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5) ON CONFLICT (id,mtype) DO UPDATE set delta=$3, val=$4, hash=$5", val.ID, val.MType, val.Delta, val.Value, val.Hash)
		if err != nil {
			log.Error(err)
		}
	}
	err := ss.db.Close()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("Primary DB connection close")
}

func (ss *DataBaseStorageState) MetricMapMuxLock() {
	ss.metricMapMux.Lock()
}

func (ss *DataBaseStorageState) MetricMapMuxUnlock() {
	ss.metricMapMux.Unlock()
}

func (ss *DataBaseStorageState) MetricMap() map[string]*collector.Metrics {

	ss.MetricMapMuxLock()
	defer ss.MetricMapMuxUnlock()

	ret := make(map[string]*collector.Metrics)

	rows, err := ss.db.Query("SELECT * FROM metrics")
	if err != nil {
		panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	for rows.Next() {
		var m collector.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
		if err != nil {
			log.Error(err)
			continue
		}
		ret[m.ID] = &m
	}
	return ret
}

func (ss *DataBaseStorageState) SetMetricMap(metricMap map[string]*collector.Metrics) {
	ss.MetricMapMuxLock()
	defer ss.MetricMapMuxUnlock()

	if len(metricMap) != 0 {
		for _, val := range metricMap {
			_, err := ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5) ON CONFLICT (id,mtype) DO UPDATE set delta=$3, val=$4, hash=$5", val.ID, val.MType, val.Delta, val.Value, val.Hash)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (ss *DataBaseStorageState) MetricMapItem(item string) (*collector.Metrics, bool) {
	ss.MetricMapMuxLock()
	defer ss.MetricMapMuxUnlock()

	var val collector.Metrics
	rows, err := ss.db.Query("SELECT * FROM metrics WHERE id=$1", item)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&val.ID, &val.MType, &val.Delta, &val.Value, &val.Hash)
		if err != nil {
			log.Error(err)
			continue
		}
		val.Hash = ss.state.GetHaser().Hash(&val)
		return &val, true
	}
	return nil, false
}

func (ss *DataBaseStorageState) SetMetricMapItem(val *collector.Metrics) {
	ss.MetricMapMuxLock()
	defer ss.MetricMapMuxUnlock()

	_, err := ss.db.Exec("INSERT INTO metrics (id, mtype, delta, val, hash) values ($1,$2,$3,$4,$5) ON CONFLICT (id,mtype) DO UPDATE set delta=$3, val=$4, hash=$5", val.ID, val.MType, val.Delta, val.Value, val.Hash)
	if err != nil {
		log.Error(err)
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
	ss.SaveToStorageLast()
}
