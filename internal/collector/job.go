package collector

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/slurm_exporter/internal/logger"
)

// JobMetrics stores metrics for each job
type JobMetrics struct {
	jobCPUs    uint64
	jobName    string
	jobStatus  string
	jobReason  string
	user       string
	partitions []string
}

func JobGetMetrics(logger *logger.Logger) (map[string]*JobMetrics, error) {
	data, err := JobData(logger)
	if err != nil {
		return nil, err
	}
	return ParseJobMetrics(data), nil
}

// ParseJobMetrics takes the output of squeue with job data
// It returns a map of metrics per job, including partitions
// Expects squeue output format:"%P,%i,%j,%T,%C,%r,%u" (Partition,ID,Name,State,CPUs,Reason,User)
func ParseJobMetrics(input []byte) map[string]*JobMetrics {
	jobs := make(map[string]*JobMetrics)
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if strings.Contains(line, ",") {
			part := strings.Split(line, ",")[0]
			part = strings.TrimSpace(part)
			id := strings.Split(line, ",")[1]
			name := strings.Split(line, ",")[2]
			state := strings.Split(line, ",")[3]
			cores, _ := strconv.Atoi(strings.Split(line, ",")[4])
			reason := strings.Split(line, ",")[5]
			user := strings.Split(line, ",")[6]
			user = strings.TrimSpace(user)

			if _, exists := jobs[id]; !exists {
				jobs[id] = &JobMetrics{0, name, state, reason, user, []string{}}
			}

			jobs[id].jobCPUs = uint64(cores)
			jobs[id].jobName = name
			jobs[id].jobStatus = state
			jobs[id].jobReason = reason
			jobs[id].user = user

			// Add the partition if it's not already in the list
			jobs[id].partitions = appendUnique(jobs[id].partitions, part)
		}
	}

	return jobs
}

/*
JobData executes the squeue command to retrieve job information
Expected squeue output format: "%P,%i,%j,%T,%C,%r,%u" (Partition,State,CPUs,ID,Name,Reason,User).
*/
func JobData(logger *logger.Logger) ([]byte, error) {
	return Execute(logger, "squeue", []string{"-h", "-o", "%P,%T,%C,%i,%j,%r,%u"})
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm job metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */
func NewJobCollector(logger *logger.Logger) *JobCollector {
	labels := []string{"job", "name", "status", "reason", "partition"}
	return &JobCollector{
		jobCPUs:   prometheus.NewDesc("slurm_job_cpus", "CPUs allocated for job", labels, nil),
		jobID:     prometheus.NewDesc("slurm_job_id", "Job ID with partition", labels, nil),
		jobName:   prometheus.NewDesc("slurm_job_name", "Job Name with partition", labels, nil),
		jobStatus: prometheus.NewDesc("slurm_job_status", "Job Status with partition", labels, nil),
		jobReason: prometheus.NewDesc("slurm_job_reason", "Job Reason with partition", labels, nil),
		user:      prometheus.NewDesc("slurm_job_user", "Job User with partition", labels, nil),
		logger:    logger,
	}
}

type JobCollector struct {
	jobCPUs   *prometheus.Desc
	jobID     *prometheus.Desc
	jobName   *prometheus.Desc
	jobStatus *prometheus.Desc
	jobReason *prometheus.Desc
	user      *prometheus.Desc
	logger    *logger.Logger
}

func (jc *JobCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- jc.jobCPUs
	ch <- jc.jobID
	ch <- jc.jobName
	ch <- jc.jobStatus
	ch <- jc.jobReason
	ch <- jc.user
}

func (jc *JobCollector) Collect(ch chan<- prometheus.Metric) {
	jobs, err := JobGetMetrics(jc.logger)
	if err != nil {
		jc.logger.Error("Failed to get job metrics", "err", err)
		return
	}
	for job, metrics := range jobs {
		for _, partition := range metrics.partitions {
			ch <- prometheus.MustNewConstMetric(jc.jobCPUs, prometheus.GaugeValue, float64(metrics.jobCPUs), job, metrics.jobName, metrics.jobStatus, metrics.jobReason, partition)
			ch <- prometheus.MustNewConstMetric(jc.jobStatus, prometheus.GaugeValue, 1, job, metrics.jobName, metrics.jobStatus, metrics.jobReason, partition)
		}
	}
}
