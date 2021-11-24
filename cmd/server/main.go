package main

import (
	"log"
	"net/http"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Printf("Method: %s, URI: %s", r.Method, r.URL.RequestURI())
}

func main() {
	http.HandleFunc("/update/", HelloWorld)
	http.Handle("/update", http.NotFoundHandler())
	http.Handle("/", http.NotFoundHandler())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
