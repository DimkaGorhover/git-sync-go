package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func startServer(port int) error {

	http.Handle(`/metrics`, promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				ErrorLog:           log.StandardLogger(),
				Registry:           prometheus.DefaultRegisterer,
				DisableCompression: false,
				EnableOpenMetrics:  false,
			}),
	))

	http.HandleFunc(`/ping`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `PONG`)
	})

	http.HandleFunc(`/health`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc(`/health/live`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc(`/health/ready`, func(w http.ResponseWriter, r *http.Request) {
		// TODO: switch to ready after first clone
		w.WriteHeader(http.StatusOK)
	})

	log.WithFields(log.Fields{
		`port`: port,
	}).Info(`start http server`)

	addr := fmt.Sprintf(`:%d`, port)
	return http.ListenAndServe(addr, nil)
}
