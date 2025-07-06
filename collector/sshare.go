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

func FairShareData(logger log.Logger) ([]byte, error) {
	cmd := exec.Command("sshare", "-n", "-P", "-o", "account,fairshare")
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

type FairShareMetrics struct {
	fairshare float64
}

func ParseFairShareMetrics(logger log.Logger) (map[string]*FairShareMetrics, error) {
	accounts := make(map[string]*FairShareMetrics)
	fairShareData, err := FairShareData(logger)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(fairShareData), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "  ") {
			if strings.Contains(line, "|") {
				account := strings.Trim(strings.Split(line, "|")[0], " ")
				_, key := accounts[account]
				if !key {
					accounts[account] = &FairShareMetrics{0}
				}
				fairshare, _ := strconv.ParseFloat(strings.Split(line, "|")[1], 64)
				accounts[account].fairshare = fairshare
			}
		}
	}
	return accounts, nil
}

type FairShareCollector struct {
	fairshare *prometheus.Desc
	logger    log.Logger
}

func NewFairShareCollector(logger log.Logger) *FairShareCollector {
	labels := []string{"account"}
	return &FairShareCollector{
		fairshare: prometheus.NewDesc("slurm_account_fairshare", "FairShare for account", labels, nil),
		logger:    logger,
	}
}

func (fsc *FairShareCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fsc.fairshare
}

func (fsc *FairShareCollector) Collect(ch chan<- prometheus.Metric) {
	fsm, err := ParseFairShareMetrics(fsc.logger)
	if err != nil {
		level.Error(fsc.logger).Log("msg", "Failed to parse fairshare metrics", "err", err)
		return
	}
	for f := range fsm {
		ch <- prometheus.MustNewConstMetric(fsc.fairshare, prometheus.GaugeValue, fsm[f].fairshare, f)
	}
}