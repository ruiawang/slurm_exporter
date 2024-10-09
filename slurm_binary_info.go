/* Copyright 2024 Thomas Bourcey

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

/* Copyright 2024 Tom

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

// SlurmBinariesCollector collects metrics for Slurm binaries and their versions.
type SlurmBinariesCollector struct {
	binaryInfo *prometheus.Desc
	binaries   []string
}

// NewSlurmBinariesCollector initializes a new SlurmBinariesCollector with a list of Slurm binaries.
func NewSlurmBinariesCollector() *SlurmBinariesCollector {
	// List of all main Slurm binaries to check
	binaries := []string{
		"sinfo", "squeue", "sdiag", "scontrol",
		"sacct", "sbatch", "salloc", "srun",
	}
	// Metric description with labels: type="binary", binary_name="name", version="version"
	return &SlurmBinariesCollector{
		binaryInfo: prometheus.NewDesc("slurm_binary_info", "Information on Slurm binaries presence and versions", []string{"binary", "version"}, nil),
		binaries:   binaries,
	}
}

// Describe sends the metric descriptions to the Prometheus channel.
func (c *SlurmBinariesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.binaryInfo
}

// Collect gathers metrics for each Slurm binary.
func (c *SlurmBinariesCollector) Collect(ch chan<- prometheus.Metric) {
	for _, binary := range c.binaries {
		version, found := GetBinaryVersion(binary)
		binaryValue := 0.0
		if found {
			binaryValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.binaryInfo, prometheus.GaugeValue, binaryValue, binary, version)
	}
}

// GetBinaryVersion checks if a Slurm binary is installed and returns its version string.
func GetBinaryVersion(binary string) (string, bool) {
	// Check if the binary is available in PATH
	_, err := exec.LookPath(binary)
	if err != nil {
		log.Printf("Binary %s not found in PATH", binary)
		return "not_found", false
	}

	// Get the version of the binary using `--version` option
	cmd := exec.Command(binary, "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing %s: %v", binary, err)
		return "not_found", false
	}

	// Extract the version number, e.g., "slurm 23.11.6"
	fields := strings.Fields(string(output))
	if len(fields) > 1 {
		return fields[1], true
	}
	return "unknown", true
}
