package main

import (
	"net/http"

	"gitlab.mgmt.arms-dev.net/go-common/healthcheck"
	"gitlab.mgmt.arms-dev.net/go-common/logger"
)

func main() {

	log := logger.NewEntry().WithField("context", "go-http")
	hc := healthcheck.NewServer(healthcheck.Opts{})

	mux := http.NewServeMux()
	mux.Handle("/liveness", hc.Liveness())
	mux.Handle("/readiness", hc.Readiness())
	mux.HandleFunc("/", defaultHandler)

	logger.InfoAlways(log, "Listening for connections...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func defaultHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}
