/* Copyright 2017 Thomas Bourcey

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
	"log"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// SlurmInfoCollector defines a Prometheus collector for Slurm binary and version information
type SlurmInfoCollector struct {
	slurmInfo *prometheus.Desc
	binaries  []string
}

// NewSlurmInfoCollector initializes a new SlurmInfoCollector
func NewSlurmInfoCollector() *SlurmInfoCollector {
	binaries := []string{
		"sinfo", "squeue", "sdiag", "scontrol",
		"sacct", "sbatch", "salloc", "srun",
	}
	labels := []string{"type", "binary", "version"}
	return &SlurmInfoCollector{
		slurmInfo: prometheus.NewDesc("slurm_info", "Information on Slurm version and binaries", labels, nil),
		binaries:  binaries,
	}
}

// Describe sends the metric descriptions to Prometheus
func (c *SlurmInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.slurmInfo
}

// Collect gathers the Slurm information and sends it as a metric to Prometheus
func (c *SlurmInfoCollector) Collect(ch chan<- prometheus.Metric) {
	// Get the general Slurm version
	version, found := GetBinaryVersion("sinfo")
	versionValue := 0.0
	if found {
		versionValue = 1.0
	}
	// Send the general Slurm version as a metric
	ch <- prometheus.MustNewConstMetric(c.slurmInfo, prometheus.GaugeValue, versionValue, "general", "", version)

	// Check each binary and send their availability and version
	for _, binary := range c.binaries {
		binVersion, binFound := GetBinaryVersion(binary)
		binValue := 0.0
		if binFound {
			binValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.slurmInfo, prometheus.GaugeValue, binValue, "binary", binary, binVersion)
	}
}

// getBinaryVersion checks if a Slurm binary is installed and returns its version string
func GetBinaryVersion(binary string) (string, bool) {
	cmd := exec.Command(binary, "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Binary %s not found: %v", binary, err)
		return "not_found", false
	}

	// Extract the version number from the output, e.g., "slurm 23.11.6"
	fields := strings.Fields(string(output))
	if len(fields) > 1 {
		return fields[1], true
	}
	return "unknown", true
}
