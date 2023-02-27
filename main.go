package main

import (
	"net/http"
	// _ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/vetinari/brickd_exporter/collector"
)

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}
	if config.Collector.LogLevel == "" {
		config.Collector.LogLevel = "info"
	}
	lvl, err := log.ParseLevel(config.Collector.LogLevel)
	if err != nil {
		log.Errorf("failed to parse `log_level` %s: %s", config.Collector.LogLevel, err)
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)

	c := collector.NewCollector(
		config.Brickd.Address,
		config.Brickd.Password,
		config.Collector.CallbackPeriod,
		config.Collector.IgnoredUIDs,
		config.Collector.Labels,
		config.Collector.SensorLabels,
		config.Collector.Expire,
		config.MQTT,
	)

	prometheus.MustRegister(c)

	listenAddress := config.Listen.Address

	http.Handle(config.Listen.MetricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.Listen.MetricsPath, http.StatusFound)
	})

	log.Printf("Starting brickd exporter on %q", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("Cannot start WL exporter: %s", err)
	}
}
