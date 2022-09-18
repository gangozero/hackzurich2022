package main

import (
	"log"
	"net/http"
)

func main() {
	srv, err := newServer()
	if err != nil {
		log.Fatalf("Initialization error: %s", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/driver", srv.getDriverhandler())

	log.Print("Listening...")
	http.ListenAndServe(":8080", mux)
}
