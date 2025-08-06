package collector

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/slurm_exporter/internal/logger"
)

type GPUsMetrics struct {
	alloc       float64
	idle        float64
	other       float64
	total       float64
	utilization float64
}

func GPUsGetMetrics(logger *logger.Logger) (*GPUsMetrics, error) {
	return ParseGPUsMetrics(logger)
}



func ParseAllocatedGPUs(data []byte) float64 {
	var num_gpus = 0.0
	// sinfo -a -h --Format="Nodes: ,GresUsed:" --state=allocated
	// 3 gpu:2                                       # slurm>=20.11.8
	// 1 gpu:(null):3(IDX:0-7)                       # slurm 21.08.5
	// 13 gpu:A30:4(IDX:0-3),gpu:Q6K:4(IDX:0-3)      # slurm 21.08.5

	sinfo_lines := string(data)
	re := regexp.MustCompile(`gpu:(\(null\)|[^:(]*):?([0-9]+)(\([^)]*\))?`)
	if len(sinfo_lines) > 0 {
		for _, line := range strings.Split(sinfo_lines, "\n") {

			if len(line) > 0 && strings.Contains(line, "gpu:") {
				nodes := strings.Fields(line)[0]
				num_nodes, _ := strconv.ParseFloat(nodes, 64)
				node_active_gpus := strings.Fields(line)[1]
				num_node_active_gpus := 0.0
				for _, node_active_gpus_type := range strings.Split(node_active_gpus, ",") {
					if strings.Contains(node_active_gpus_type, "gpu:") {
						node_active_gpus_type = re.FindStringSubmatch(node_active_gpus_type)[2]
						num_node_active_gpus_type, _ := strconv.ParseFloat(node_active_gpus_type, 64)
						num_node_active_gpus += num_node_active_gpus_type
					}
				}
				num_gpus += num_nodes * num_node_active_gpus
			}
		}
	}

	return num_gpus
}

func ParseIdleGPUs(data []byte) float64 {
	var num_gpus = 0.0
	// sinfo -a -h --Format="Nodes: ,Gres: ,GresUsed:" --state=idle,allocated
	// 3 gpu:4 gpu:2                                       												# slurm 20.11.8
	// 1 gpu:8(S:0-1) gpu:(null):3(IDX:0-7)                       												# slurm 21.08.5
	// 13 gpu:A30:4(S:0-1),gpu:Q6K:40(S:0-1) gpu:A30:4(IDX:0-3),gpu:Q6K:4(IDX:0-3)       	# slurm 21.08.5

	sinfo_lines := string(data)
	re := regexp.MustCompile(`gpu:(\(null\)|[^:(]*):?([0-9]+)(\([^)]*\))?`)
	if len(sinfo_lines) > 0 {
		for _, line := range strings.Split(sinfo_lines, "\n") {

			if len(line) > 0 && strings.Contains(line, "gpu:") {
				fields := strings.Fields(line)
				nodes := fields[0]
				num_nodes, _ := strconv.ParseFloat(nodes, 64)

				if len(fields) == 1 {
					// Case where only one column is present, assume it's allocated GPUs
					num_gpus += 0 // No idle GPUs in this case
				} else if len(fields) == 2 {
					// Case where two columns are present
					node_gpus_str := fields[1]
					num_node_gpus := 0.0
					for _, node_gpus_type := range strings.Split(node_gpus_str, ",") {
						if strings.Contains(node_gpus_type, "gpu:") {
							submatch := re.FindStringSubmatch(node_gpus_type)
							if len(submatch) > 2 {
								gpu_count, _ := strconv.ParseFloat(submatch[2], 64)
								num_node_gpus += gpu_count
							}
						}
					}
					num_gpus += num_nodes * num_node_gpus
				} else if len(fields) >= 3 {
					// Original case with three or more columns
					node_gpus_str := fields[1]
					num_node_gpus := 0.0
					for _, node_gpus_type := range strings.Split(node_gpus_str, ",") {
						if strings.Contains(node_gpus_type, "gpu:") {
							submatch := re.FindStringSubmatch(node_gpus_type)
							if len(submatch) > 2 {
								gpu_count, _ := strconv.ParseFloat(submatch[2], 64)
								num_node_gpus += gpu_count
							}
						}
					}

					active_gpus_str := fields[2]
					num_node_active_gpus := 0.0
					for _, node_active_gpus_type := range strings.Split(active_gpus_str, ",") {
						if strings.Contains(node_active_gpus_type, "gpu:") {
							submatch := re.FindStringSubmatch(node_active_gpus_type)
							if len(submatch) > 2 {
								gpu_count, _ := strconv.ParseFloat(submatch[2], 64)
								num_node_active_gpus += gpu_count
							}
						}
					}
					num_gpus += num_nodes * (num_node_gpus - num_node_active_gpus)
				}
			}
		}
	}

	return num_gpus
}

func ParseTotalGPUs(data []byte) float64 {
	var num_gpus = 0.0
	// sinfo -a -h --Format="Nodes: ,Gres:"
	// 3 gpu:4                                       	# slurm 20.11.8
	// 1 gpu:8(S:0-1)                                	# slurm 21.08.5
	// 13 gpu:A30:4(S:0-1),gpu:Q6K:40(S:0-1)        	# slurm 21.08.5

	sinfo_lines := string(data)
	re := regexp.MustCompile(`gpu:(\(null\)|[^:(]*):?([0-9]+)(\([^)]*\))?`)
	if len(sinfo_lines) > 0 {
		for _, line := range strings.Split(sinfo_lines, "\n") {
			
			if len(line) > 0 && strings.Contains(line, "gpu:") {
				nodes := strings.Fields(line)[0]
				num_nodes, _ := strconv.ParseFloat(nodes, 64)
				node_gpus := strings.Fields(line)[1]
				num_node_gpus := 0.0
				for _, node_gpus_type := range strings.Split(node_gpus, ",") {
					if strings.Contains(node_gpus_type, "gpu:") {
						node_gpus_type = re.FindStringSubmatch(node_gpus_type)[2]
						num_node_gpus_type, _ := strconv.ParseFloat(node_gpus_type, 64)
						num_node_gpus += num_node_gpus_type
					}
				}
				num_gpus += num_nodes * num_node_gpus
			}
		}
	}

	return num_gpus
}

func ParseGPUsMetrics(logger *logger.Logger) (*GPUsMetrics, error) {
	var gm GPUsMetrics
	totalGPUsData, err := TotalGPUsData(logger)
	if err != nil {
		return nil, err
	}
	total_gpus := ParseTotalGPUs(totalGPUsData)

	allocatedGPUsData, err := AllocatedGPUsData(logger)
	if err != nil {
		return nil, err
	}
	allocated_gpus := ParseAllocatedGPUs(allocatedGPUsData)

	idleGPUsData, err := IdleGPUsData(logger)
	if err != nil {
		return nil, err
	}
	idle_gpus := ParseIdleGPUs(idleGPUsData)

	other_gpus := total_gpus - allocated_gpus - idle_gpus
	gm.alloc = allocated_gpus
	gm.idle = idle_gpus
	gm.other = other_gpus
	gm.total = total_gpus
	if total_gpus > 0 {
		gm.utilization = allocated_gpus / total_gpus
	}
	return &gm, nil
}

func AllocatedGPUsData(logger *logger.Logger) ([]byte, error) {
	args := []string{"-a", "-h", "--Format=Nodes: ,GresUsed:", "--state=allocated"}
	return Execute(logger, "sinfo", args)
}

func IdleGPUsData(logger *logger.Logger) ([]byte, error) {
	args := []string{"-a", "-h", "--Format=Nodes: ,Gres: ,GresUsed:", "--state=idle,allocated"}
	return Execute(logger, "sinfo", args)
}

func TotalGPUsData(logger *logger.Logger) ([]byte, error) {
	args := []string{"-a", "-h", "--Format=Nodes: ,Gres:"}
	return Execute(logger, "sinfo", args)
}


/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewGPUsCollector(logger *logger.Logger) *GPUsCollector {
	return &GPUsCollector{
		alloc:       prometheus.NewDesc("slurm_gpus_alloc", "Allocated GPUs", nil, nil),
		idle:        prometheus.NewDesc("slurm_gpus_idle", "Idle GPUs", nil, nil),
		other:       prometheus.NewDesc("slurm_gpus_other", "Other GPUs", nil, nil),
		total:       prometheus.NewDesc("slurm_gpus_total", "Total GPUs", nil, nil),
		utilization: prometheus.NewDesc("slurm_gpus_utilization", "Total GPU utilization", nil, nil),
		logger:      logger,
	}
}

type GPUsCollector struct {
	alloc       *prometheus.Desc
	idle        *prometheus.Desc
	other       *prometheus.Desc
	total       *prometheus.Desc
	utilization *prometheus.Desc
	logger      *logger.Logger
}


func (cc *GPUsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.alloc
	ch <- cc.idle
	ch <- cc.other
	ch <- cc.total
	ch <- cc.utilization
}
func (cc *GPUsCollector) Collect(ch chan<- prometheus.Metric) {
	cm, err := GPUsGetMetrics(cc.logger)
	if err != nil {
		cc.logger.Error("Failed to get GPUs metrics", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(cc.alloc, prometheus.GaugeValue, cm.alloc)
	ch <- prometheus.MustNewConstMetric(cc.idle, prometheus.GaugeValue, cm.idle)
	ch <- prometheus.MustNewConstMetric(cc.other, prometheus.GaugeValue, cm.other)
	ch <- prometheus.MustNewConstMetric(cc.total, prometheus.GaugeValue, cm.total)
	ch <- prometheus.MustNewConstMetric(cc.utilization, prometheus.GaugeValue, cm.utilization)
}