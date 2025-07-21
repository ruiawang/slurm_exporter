/* Copyright 2017 Victor Penso, Matteo Dessalvi

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

package collector

import (
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type CPUsMetrics struct {
	alloc float64
	idle  float64
	other float64
	total float64
}

func CPUsGetMetrics(logger log.Logger) (*CPUsMetrics, error) {
	data, err := CPUsData(logger)
	if err != nil {
		return nil, err
	}
	return ParseCPUsMetrics(data), nil
}

func ParseCPUsMetrics(input []byte) *CPUsMetrics {
	var cm CPUsMetrics
	if strings.Contains(string(input), "/") {
		splitted := strings.Split(strings.TrimSpace(string(input)), "/")
		cm.alloc, _ = strconv.ParseFloat(splitted[0], 64)
		cm.idle, _ = strconv.ParseFloat(splitted[1], 64)
		cm.other, _ = strconv.ParseFloat(splitted[2], 64)
		cm.total, _ = strconv.ParseFloat(splitted[3], 64)
	}
	return &cm
}

// Execute the sinfo command and return its output
func CPUsData(logger log.Logger) ([]byte, error) {
	return Execute(logger, "sinfo", []string{"-h", "-o", "%C"})
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewCPUsCollector(logger log.Logger) *CPUsCollector {
	return &CPUsCollector{
		alloc:  prometheus.NewDesc("slurm_cpus_alloc", "Allocated CPUs", nil, nil),
		idle:   prometheus.NewDesc("slurm_cpus_idle", "Idle CPUs", nil, nil),
		other:  prometheus.NewDesc("slurm_cpus_other", "Mix CPUs", nil, nil),
		total:  prometheus.NewDesc("slurm_cpus_total", "Total CPUs", nil, nil),
		logger: logger,
	}
}

type CPUsCollector struct {
	alloc  *prometheus.Desc
	idle   *prometheus.Desc
	other  *prometheus.Desc
	total  *prometheus.Desc
	logger log.Logger
}

// Send all metric descriptions
func (cc *CPUsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.alloc
	ch <- cc.idle
	ch <- cc.other
	ch <- cc.total
}
func (cc *CPUsCollector) Collect(ch chan<- prometheus.Metric) {
	cm, err := CPUsGetMetrics(cc.logger)
	if err != nil {
		_ = level.Error(cc.logger).Log("msg", "Failed to get CPUs metrics", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(cc.alloc, prometheus.GaugeValue, cm.alloc)
	ch <- prometheus.MustNewConstMetric(cc.idle, prometheus.GaugeValue, cm.idle)
	ch <- prometheus.MustNewConstMetric(cc.other, prometheus.GaugeValue, cm.other)
	ch <- prometheus.MustNewConstMetric(cc.total, prometheus.GaugeValue, cm.total)
}
