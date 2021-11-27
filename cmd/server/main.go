package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func GaugeRoute(w http.ResponseWriter, r *http.Request) {
	//gaugeRoute := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "RandomValue"}
	uri, _ := url.Parse(r.URL.RequestURI())
	realPath := strings.Replace(uri.Path, "/update/gauge", "", 1)
	if string([]rune(realPath)[0]) == "/" {
		realPath = strings.Replace(realPath, "/", "", 1)
	}
	all := strings.Split(realPath, "/")
	log.Printf("Method: %s, URI: %s", r.Method, all)
	if len(all) == 2 {
		_, err := strconv.ParseFloat(all[1], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func CounterRoute(w http.ResponseWriter, r *http.Request) {
	uri, _ := url.Parse(r.URL.RequestURI())
	realPath := strings.Replace(uri.Path, "/update/counter", "", 1)
	if string([]rune(realPath)[0]) == "/" {
		realPath = strings.Replace(realPath, "/", "", 1)
	}
	all := strings.Split(realPath, "/")
	log.Printf("Method: %s, URI: %s", r.Method, all)
	if len(all) == 2 {
		_, err := strconv.ParseInt(all[1], 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	http.HandleFunc("/update/gauge/", GaugeRoute)
	http.Handle("/update/gauge", http.NotFoundHandler())
	http.HandleFunc("/update/counter/", CounterRoute)
	http.Handle("/update/counter", http.NotFoundHandler())
	http.Handle("/", http.NotFoundHandler())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
