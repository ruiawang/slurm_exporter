package collector

import (
	"os/exec"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// SlurmInfoCollector defines a Prometheus collector for Slurm binary and version information
type SlurmInfoCollector struct {
	slurmInfo *prometheus.Desc
	binaries  []string
	logger    log.Logger
}

// NewSlurmInfoCollector initializes a new SlurmInfoCollector
func NewSlurmInfoCollector(logger log.Logger) *SlurmInfoCollector {
	binaries := []string{
		"sinfo", "squeue", "sdiag", "scontrol",
		"sacct", "sbatch", "salloc", "srun",
	}
	labels := []string{"type", "binary", "version"}
	return &SlurmInfoCollector{
		slurmInfo: prometheus.NewDesc("slurm_info", "Information on Slurm version and binaries", labels, nil),
		binaries:  binaries,
		logger:    logger,
	}
}

// Describe sends the metric descriptions to Prometheus
func (c *SlurmInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.slurmInfo
}

// Collect gathers the Slurm information and sends it as a metric to Prometheus
func (c *SlurmInfoCollector) Collect(ch chan<- prometheus.Metric) {
	// Get the general Slurm version
	version, found := GetBinaryVersion(c.logger, "sinfo")
	versionValue := 0.0
	if found {
		versionValue = 1.0
	}
	// Send the general Slurm version as a metric
	ch <- prometheus.MustNewConstMetric(c.slurmInfo, prometheus.GaugeValue, versionValue, "general", "", version)

	// Check each binary and send their availability and version
	for _, binary := range c.binaries {
		binVersion, binFound := GetBinaryVersion(c.logger, binary)
		binValue := 0.0
		if binFound {
			binValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.slurmInfo, prometheus.GaugeValue, binValue, "binary", binary, binVersion)
	}
}

// getBinaryVersion checks if a Slurm binary is installed and returns its version string
func GetBinaryVersion(logger log.Logger, binary string) (string, bool) {
	cmd := exec.Command(binary, "--version")
	output, err := cmd.Output()
	if err != nil {
		level.Error(logger).Log("msg", "Binary not found", "binary", binary, "err", err)
		return "not_found", false
	}

	// Extract the version number from the output, e.g., "slurm 23.11.6"
	fields := strings.Fields(string(output))
	if len(fields) > 1 {
		return fields[1], true
	}
	return "unknown", true
}