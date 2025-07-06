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
	gpuAcct      = kingpin.Flag("gpus-acct", "Enable GPUs accounting").Default("false").Bool()
	toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":8080")
)

// Message to display on the root page
const indexHTML = `<html>
	<head><title>Slurm Exporter</title></head>
	<body>
		<h1>Slurm Exporter</h1>
		<p>Welcome to the Slurm Exporter. Click <a href='/metrics'>here</a> to see the metrics.</p>
	</body>
</html>`

func registerCollectors(logger log.Logger, gpuAcct bool) {
	collectors := []prometheus.Collector{
		collector.NewAccountsCollector(logger),
		collector.NewCPUsCollector(logger),
		collector.NewNodesCollector(logger),
		collector.NewNodeCollector(logger),
		collector.NewPartitionsCollector(logger),
		collector.NewQueueCollector(logger),
		collector.NewSchedulerCollector(logger),
		collector.NewFairShareCollector(logger),
		collector.NewUsersCollector(logger),
		collector.NewSlurmInfoCollector(logger),
	}

	// Register GPU collector if enabled
	if gpuAcct {
		collectors = append(collectors, collector.NewGPUsCollector(logger))
	}

	// Register all collectors
	for _, collector := range collectors {
		prometheus.MustRegister(collector)
	}
}

func main() {
	// Prometheus logging configuration
	promlogConfig := &promlog.Config{}
	kingpin.Version(version.Print("slurm_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	// Register version metrics
	prometheus.MustRegister(collectors.NewBuildInfoCollector())

	// Register collectors based on the GPU accounting flag
	registerCollectors(logger, *gpuAcct)

	// Log server startup details
	level.Info(logger).Log("msg", "Starting Server with GPUs Accounting", "enabled", *gpuAcct)

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
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
