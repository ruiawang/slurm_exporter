/* Copyright 2017-2020 Victor Penso, Matteo Dessalvi, Joeri Hermans

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/sckyzo/slurm_exporter/collector"
)

var (
	// Flags for command-line configuration
	commandTimeout = kingpin.Flag("command.timeout", "Timeout for executing Slurm commands.").Default("5s").Duration()
	logLevel       = kingpin.Flag("log.level", "Only log messages with the given severity or above. One of: [debug, info, warn, error]").Default("info").Enum("debug", "info", "warn", "error")
	toolkitFlags   = webflag.AddFlags(kingpin.CommandLine, ":8080")

	// Map to store the state of collectors
	collectorState = make(map[string]*bool)
)

// Map of collector constructors
var collectorConstructors = map[string]func(logger log.Logger) prometheus.Collector{
	"accounts":   func(l log.Logger) prometheus.Collector { return collector.NewAccountsCollector(l) },
	"cpus":       func(l log.Logger) prometheus.Collector { return collector.NewCPUsCollector(l) },
	"nodes":      func(l log.Logger) prometheus.Collector { return collector.NewNodesCollector(l) },
	"node":       func(l log.Logger) prometheus.Collector { return collector.NewNodeCollector(l) },
	"partitions": func(l log.Logger) prometheus.Collector { return collector.NewPartitionsCollector(l) },
	"queue":      func(l log.Logger) prometheus.Collector { return collector.NewQueueCollector(l) },
	"scheduler":  func(l log.Logger) prometheus.Collector { return collector.NewSchedulerCollector(l) },
	"fairshare":  func(l log.Logger) prometheus.Collector { return collector.NewFairShareCollector(l) },
	"users":      func(l log.Logger) prometheus.Collector { return collector.NewUsersCollector(l) },
	"info":       func(l log.Logger) prometheus.Collector { return collector.NewSlurmInfoCollector(l) },
	"gpus":       func(l log.Logger) prometheus.Collector { return collector.NewGPUsCollector(l) },
	"reservations": func(l log.Logger) prometheus.Collector { return collector.NewReservationsCollector(l) },
}

// Message to display on the root page
const indexHTML = `<html>
	<head><title>Slurm Exporter</title></head>
	<body>
		<h1>Slurm Exporter</h1>
		<p>Welcome to the Slurm Exporter. Click <a href='/metrics'>here</a> to see the metrics.</p>
	</body>
</html>`

func registerCollectors(logger log.Logger) {
	for name, constructor := range collectorConstructors {
		if *collectorState[name] {
			prometheus.MustRegister(constructor(logger))
			_ = level.Info(logger).Log("msg", "Collector enabled", "collector", name)
		} else {
			_ = level.Info(logger).Log("msg", "Collector disabled", "collector", name)
		}
	}
}

func main() {
	// Dynamically create flags for each collector
	for name := range collectorConstructors {
		collectorState[name] = kingpin.Flag("collector."+name, "Enable the "+name+" collector.").Default("true").Bool()
	}

	// Prometheus logging configuration
	promlogConfig := &promlog.Config{}
	kingpin.Version(version.Print("slurm_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// Setup logger with the configured level
	promlogConfig.Level = &promlog.AllowedLevel{}
	if err := promlogConfig.Level.Set(*logLevel); err != nil {
		panic(err) // Should not happen due to kingpin's Enum validation
	}
	logger := promlog.New(promlogConfig)

	// Set the command timeout for the collector package.
	collector.SetCommandTimeout(*commandTimeout)

	// Register version metrics
	prometheus.MustRegister(collectors.NewBuildInfoCollector())

	// Register collectors based on flags
	registerCollectors(logger)

	// Log server startup details
	_ = level.Info(logger).Log("msg", "Starting Server...")
	_ = level.Info(logger).Log("msg", "Command timeout set", "timeout", *commandTimeout)

	// Define the root handler for '/'
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(indexHTML))
	})

	// Expose /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Create the HTTP server
	server := &http.Server{}

	// Use exporter toolkit to start the server (supports TLS, Basic Auth, etc.)
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		_ = level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
