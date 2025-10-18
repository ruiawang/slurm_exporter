package collector

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/slurm_exporter/internal/logger"
)

// NodeMetrics stores metrics for each node
type NodeMetrics struct {
	memAlloc   uint64
	memTotal   uint64
	cpuAlloc   uint64
	cpuIdle    uint64
	cpuOther   uint64
	cpuTotal   uint64
	nodeStatus string
	partitions []string
	reason     string
	user       string
	timestamp  string
}

func NodeGetMetrics(logger *logger.Logger) (map[string]*NodeMetrics, error) {
	data, err := NodeData(logger)
	if err != nil {
		return nil, err
	}
	return ParseNodeMetrics(data), nil
}

// ParseNodeMetrics takes the output of sinfo with node data
// It returns a map of metrics per node, including partitions
func ParseNodeMetrics(input []byte) map[string]*NodeMetrics {
	nodes := make(map[string]*NodeMetrics)
	lines := strings.Split(string(input), "\n")

	// Sort and remove all the duplicates from the 'sinfo' output
	sort.Strings(lines)
	linesUniq := RemoveDuplicates(lines)

	re := regexp.MustCompile(`\s{2,}`) // Split by two or more spaces, since Reason field can have singular spaces
	for _, line := range linesUniq {
		node := re.Split(line, -1)
		if len(node) < 9 {
			continue
		}
		nodeName := node[0]
		nodeStatus := node[4] // mixed, allocated, etc.
		partition := node[5]  // Partition name

		// Create new node metrics if it doesn't exist
		if _, exists := nodes[nodeName]; !exists {
			nodes[nodeName] = &NodeMetrics{0, 0, 0, 0, 0, 0, nodeStatus, []string{}, "", "", ""}
		}

		memAlloc, _ := strconv.ParseUint(node[1], 10, 64)
		memTotal, _ := strconv.ParseUint(node[2], 10, 64)

		cpuInfo := strings.Split(node[3], "/")
		cpuAlloc, _ := strconv.ParseUint(cpuInfo[0], 10, 64)
		cpuIdle, _ := strconv.ParseUint(cpuInfo[1], 10, 64)
		cpuOther, _ := strconv.ParseUint(cpuInfo[2], 10, 64)
		cpuTotal, _ := strconv.ParseUint(cpuInfo[3], 10, 64)

		nodes[nodeName].memAlloc = memAlloc
		nodes[nodeName].memTotal = memTotal
		nodes[nodeName].cpuAlloc = cpuAlloc
		nodes[nodeName].cpuIdle = cpuIdle
		nodes[nodeName].cpuOther = cpuOther
		nodes[nodeName].cpuTotal = cpuTotal

		reason := node[6]
		user := node[7]
		timestamp := node[8]

		nodes[nodeName].reason = reason
		nodes[nodeName].user = user
		nodes[nodeName].timestamp = timestamp
		// Add the partition if it's not already in the list
		nodes[nodeName].partitions = appendUnique(nodes[nodeName].partitions, partition)
	}

	return nodes
}

/*
NodeData executes the sinfo command to get detailed data for each node.
Expected sinfo output format: "NodeList,AllocMem,Memory,CPUsState,StateLong,Partition,Reason,UserLong,Timestamp".
*/
func NodeData(logger *logger.Logger) ([]byte, error) {
	args := []string{"-h", "-N", "-O", "NodeList:25,AllocMem,Memory,CPUsState,StateLong,Partition,Reason:30,UserLong,Timestamp"}
	return Execute(logger, "sinfo", args)
}

type NodeCollector struct {
	cpuAlloc   *prometheus.Desc
	cpuIdle    *prometheus.Desc
	cpuOther   *prometheus.Desc
	cpuTotal   *prometheus.Desc
	memAlloc   *prometheus.Desc
	memTotal   *prometheus.Desc
	nodeStatus *prometheus.Desc
	logger     *logger.Logger
}

func NewNodeCollector(logger *logger.Logger) *NodeCollector {
	labels := []string{"node", "status", "partition", "reason", "user", "timestamp"}
	return &NodeCollector{
		cpuAlloc:   prometheus.NewDesc("slurm_node_cpu_alloc", "Allocated CPUs per node", labels, nil),
		cpuIdle:    prometheus.NewDesc("slurm_node_cpu_idle", "Idle CPUs per node", labels, nil),
		cpuOther:   prometheus.NewDesc("slurm_node_cpu_other", "Other CPUs per node", labels, nil),
		cpuTotal:   prometheus.NewDesc("slurm_node_cpu_total", "Total CPUs per node", labels, nil),
		memAlloc:   prometheus.NewDesc("slurm_node_mem_alloc", "Allocated memory per node", labels, nil),
		memTotal:   prometheus.NewDesc("slurm_node_mem_total", "Total memory per node", labels, nil),
		nodeStatus: prometheus.NewDesc("slurm_node_status", "Node Status with partition", labels, nil),
		logger:     logger,
	}
}

func (nc *NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nc.cpuAlloc
	ch <- nc.cpuIdle
	ch <- nc.cpuOther
	ch <- nc.cpuTotal
	ch <- nc.memAlloc
	ch <- nc.memTotal
	ch <- nc.nodeStatus
}

func (nc *NodeCollector) Collect(ch chan<- prometheus.Metric) {
	nodes, err := NodeGetMetrics(nc.logger)
	if err != nil {
		nc.logger.Error("Failed to get node metrics", "err", err)
		return
	}
	for node, metrics := range nodes {
		for _, partition := range metrics.partitions {
			ch <- prometheus.MustNewConstMetric(nc.cpuAlloc, prometheus.GaugeValue, float64(metrics.cpuAlloc), node, metrics.nodeStatus, partition, metrics.reason, metrics.user, metrics.timestamp)
			ch <- prometheus.MustNewConstMetric(nc.cpuIdle, prometheus.GaugeValue, float64(metrics.cpuIdle), node, metrics.nodeStatus, partition, metrics.reason, metrics.user, metrics.timestamp)
			ch <- prometheus.MustNewConstMetric(nc.cpuOther, prometheus.GaugeValue, float64(metrics.cpuOther), node, metrics.nodeStatus, partition, metrics.reason, metrics.user, metrics.timestamp)
			ch <- prometheus.MustNewConstMetric(nc.cpuTotal, prometheus.GaugeValue, float64(metrics.cpuTotal), node, metrics.nodeStatus, partition, metrics.reason, metrics.user, metrics.timestamp)
			ch <- prometheus.MustNewConstMetric(nc.memAlloc, prometheus.GaugeValue, float64(metrics.memAlloc), node, metrics.nodeStatus, partition, metrics.reason, metrics.user, metrics.timestamp)
			ch <- prometheus.MustNewConstMetric(nc.memTotal, prometheus.GaugeValue, float64(metrics.memTotal), node, metrics.nodeStatus, partition, metrics.reason, metrics.user, metrics.timestamp)
			ch <- prometheus.MustNewConstMetric(nc.nodeStatus, prometheus.GaugeValue, 1, node, metrics.nodeStatus, partition, metrics.reason, metrics.user, metrics.timestamp)
		}
	}
}

// appendUnique adds a string to a slice if it doesn't already exist
func appendUnique(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
