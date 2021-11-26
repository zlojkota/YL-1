package main

import (
	"log"
	"net/http"
)

func GaugeRoute(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Printf("Method: %s, URI: %s", r.Method, r.URL.RequestURI())
}

func CounterRoute(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Printf("Method: %s, URI: %s", r.Method, r.URL.RequestURI())
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
