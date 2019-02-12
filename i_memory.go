package main

import (
// 	"regexp"
	"os/exec"
	"strconv"
// 	"errors"
	"time"
	"strings"
	"fmt"
)

type MemStats struct {
	ps PS
}

func (_ *MemStats) Description() string {
	return "Read metrics about memory usage"
}

func (_ *MemStats) SampleConfig() string { return "" }

func (s *MemStats) Gather(acc Accumulator) error {
	output, err := exec.Command("top", "-n", "-u").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error getting Memory info: %s", err.Error())
	}

	now := time.Now()

	stats := string(output)
	rows := strings.Split(stats, "\n")
	data := strings.Fields(rows[4])
	total := memconv(data[1])
	free  := memconv(data[4])
	
	fields := map[string]interface{}{
		"total": total,
		"free":  free,
		"used":  total - free,
		"free_percent": 100 * float64(free) / float64(total),
		"used_percent": 100 * float64(total - free) / float64(total),
	}
	
	acc.AddCounter("mem", fields, nil, now)
	return nil
}

func memconv(memory string) int {
	if strings.Contains(memory, "G") {
		value, _ := strconv.Atoi(strings.TrimSuffix(memory, "G"))
		return value * 1024 * 1024 * 1024
	}
	if strings.Contains(memory, "M") {
		value, _ := strconv.Atoi(strings.TrimSuffix(memory, "M"))
		return value * 1024 * 1024
	}
	if strings.Contains(memory, "k") {
		value, _ := strconv.Atoi(strings.TrimSuffix(memory, "k"))
		return value * 1024
	}
	value, _ := strconv.Atoi(memory)
	return value
}
