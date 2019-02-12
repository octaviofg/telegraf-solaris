package main

import (
	"regexp"
	"os/exec"
	"strings"
	"strconv"
	"time"
	"fmt"
)

type CPUStats struct {
	ps             PS
	PerCPU         bool `toml:"percpu"`
	TotalCPU       bool `toml:"totalcpu"`
	CollectCPUTime bool `toml:"collect_cpu_time"`
	ReportActive   bool `toml:"report_active"`
}

func NewCPUStats(ps PS) *CPUStats {
	return &CPUStats{
		ps:             ps,
		CollectCPUTime: true,
		ReportActive:   true,
	}
}

func (_ *CPUStats) Description() string {
	return "Read metrics about cpu usage"
}

var sampleConfig = `
  ## Whether to report per-cpu stats or not
  percpu = true
  ## Whether to report total system cpu stats or not
  totalcpu = true
  ## If true, collect raw CPU time metrics.
  collect_cpu_time = false
  ## If true, compute and report the sum of all non-idle CPU states.
  report_active = false
`

func (_ *CPUStats) SampleConfig() string {
	return sampleConfig
}

func (s *CPUStats) Gather(acc Accumulator) error {
	var tidle, tsystem, tuser, temp, ncpus int64
	
	output, err := exec.Command("kstat", "-p", "-m", "cpu_stat").CombinedOutput()
	
// 	fmt.Printf("#%s#", output)
// 	fmt.Printf("#%s#", err)
	
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err.Error())
	}
	
	now := time.Now()
	
	stats := string(output)
	rows := strings.Split(stats, "\n")
	rows = rows[0: len(rows)-1]
	for _, row := range rows {
		data := strings.Fields(row)
		reg := regexp.MustCompile(".*:.*:.*:")
		field := reg.ReplaceAllString(data[0], "${1}")
		
		switch field {
		case "idle":
			temp, _ = strconv.ParseInt(data[1], 10, 0)
			tidle += temp
			ncpus++
		case "user":
			temp, _ = strconv.ParseInt(data[1], 10, 0)
			tuser += temp
		case "kernel":
			temp, _ = strconv.ParseInt(data[1], 10, 0)
			tsystem += temp
		}
	}
	
	tidle /= ncpus
	tuser /= ncpus
	tsystem /= ncpus

// 	fmt.Printf("i: %d, u: %d, k: %d, n: %d\n", tidle, tuser, tsystem, ncpus)

	fields := map[string]interface{}{
		"idle":   tidle,
		"user":   tuser,
		"kernel": tsystem,
	}
	
	tags := map[string]string{
		"cpu": "cpu-total",
	}
	
	acc.AddCounter("cpu", fields, tags, now)
	return nil
}
