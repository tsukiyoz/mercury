package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}
