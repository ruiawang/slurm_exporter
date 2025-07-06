package collector

import (
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func PartitionsData(logger log.Logger) ([]byte, error) {
	cmd := exec.Command("sinfo", "-h", "-o%R,%C")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		level.Error(logger).Log("msg", "Failed to create stdout pipe", "err", err)
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		level.Error(logger).Log("msg", "Failed to start command", "err", err)
		return nil, err
	}
	out, _ := io.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		level.Error(logger).Log("msg", "Failed to wait for command", "err", err)
		return nil, err
	}
	return out, nil
}

func PartitionsPendingJobsData(logger log.Logger) ([]byte, error) {
	cmd := exec.Command("squeue", "-a", "-r", "-h", "-o%P", "--states=PENDING")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		level.Error(logger).Log("msg", "Failed to create stdout pipe", "err", err)
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		level.Error(logger).Log("msg", "Failed to start command", "err", err)
		return nil, err
	}
	out, _ := io.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		level.Error(logger).Log("msg", "Failed to wait for command", "err", err)
		return nil, err
	}
	return out, nil
}

type PartitionMetrics struct {
	allocated float64
	idle      float64
	other     float64
	pending   float64
	total     float64
}

func ParsePartitionsMetrics(logger log.Logger) (map[string]*PartitionMetrics, error) {
	partitions := make(map[string]*PartitionMetrics)
	partitionsData, err := PartitionsData(logger)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(partitionsData), "\n")
	for _, line := range lines {
		if strings.Contains(line, ",") {
			// name of a partition
			partition := strings.Split(line, ",")[0]
			_, key := partitions[partition]
			if !key {
				partitions[partition] = &PartitionMetrics{0, 0, 0, 0, 0}
			}
			states := strings.Split(line, ",")[1]
			allocated, _ := strconv.ParseFloat(strings.Split(states, "/")[0], 64)
			idle, _ := strconv.ParseFloat(strings.Split(states, "/")[1], 64)
			other, _ := strconv.ParseFloat(strings.Split(states, "/")[2], 64)
			total, _ := strconv.ParseFloat(strings.Split(states, "/")[3], 64)
			partitions[partition].allocated = allocated
			partitions[partition].idle = idle
			partitions[partition].other = other
			partitions[partition].total = total
		}
	}
	// get list of pending jobs by partition name
	pendingJobsData, err := PartitionsPendingJobsData(logger)
	if err != nil {
		return nil, err
	}
	list := strings.Split(string(pendingJobsData), "\n")
	for _, partition := range list {
		// accumulate the number of pending jobs
		_, key := partitions[partition]
		if key {
			partitions[partition].pending += 1
		}
	}

	return partitions, nil
}

type PartitionsCollector struct {
	allocated *prometheus.Desc
	idle      *prometheus.Desc
	other     *prometheus.Desc
	pending   *prometheus.Desc
	total     *prometheus.Desc
	logger    log.Logger
}

func NewPartitionsCollector(logger log.Logger) *PartitionsCollector {
	labels := []string{"partition"}
	return &PartitionsCollector{
		allocated: prometheus.NewDesc("slurm_partition_cpus_allocated", "Allocated CPUs for partition", labels, nil),
		idle:      prometheus.NewDesc("slurm_partition_cpus_idle", "Idle CPUs for partition", labels, nil),
		other:     prometheus.NewDesc("slurm_partition_cpus_other", "Other CPUs for partition", labels, nil),
		pending:   prometheus.NewDesc("slurm_partition_jobs_pending", "Pending jobs for partition", labels, nil),
		total:     prometheus.NewDesc("slurm_partition_cpus_total", "Total CPUs for partition", labels, nil),
		logger:    logger,
	}
}

func (pc *PartitionsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- pc.allocated
	ch <- pc.idle
	ch <- pc.other
	ch <- pc.pending
	ch <- pc.total
}

func (pc *PartitionsCollector) Collect(ch chan<- prometheus.Metric) {
	pm, err := ParsePartitionsMetrics(pc.logger)
	if err != nil {
		level.Error(pc.logger).Log("msg", "Failed to parse partitions metrics", "err", err)
		return
	}
	for p := range pm {
		if pm[p].allocated > 0 {
			ch <- prometheus.MustNewConstMetric(pc.allocated, prometheus.GaugeValue, pm[p].allocated, p)
		}
		if pm[p].idle > 0 {
			ch <- prometheus.MustNewConstMetric(pc.idle, prometheus.GaugeValue, pm[p].idle, p)
		}
		if pm[p].other > 0 {
			ch <- prometheus.MustNewConstMetric(pc.other, prometheus.GaugeValue, pm[p].other, p)
		}
		if pm[p].pending > 0 {
			ch <- prometheus.MustNewConstMetric(pc.pending, prometheus.GaugeValue, pm[p].pending, p)
		}
		if pm[p].total > 0 {
			ch <- prometheus.MustNewConstMetric(pc.total, prometheus.GaugeValue, pm[p].total, p)
		}
	}
}