package sysinfo

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"strings"
)

type CPUInfo struct {
	Name      string `json:"name"`      // e.g. AMD Ryzen 7 5800H with Radeon Graphics
	Cores     int    `json:"cores"`     // e.g. 4
	Threads   int    `json:"threads"`   // e.g. 8
	Frequency string `json:"frequency"` // e.g. 3.2 GHz
}

func GetCPUInfo() CPUInfo {
	name := getCPUName()
	cores := runtime.NumCPU() / 2
	threads := runtime.NumCPU()
	freq := getCPUClock()

	return CPUInfo{
		Name:      name,
		Cores:     cores,
		Threads:   threads,
		Frequency: freq,
	}
}

func getCPUName() string {
	data, err := ReadHostOrDefault("/proc/cpuinfo")
	if err != nil {
		return "Unknown"
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[1])
			if name != "" {
				return name
			}
		}
	}

	return "Unknown"
}

func getCPUClock() string {
	// 1. /sys/devices/.../cpuinfo_max_freq
	if data, err := ReadHostOrDefault("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"); err == nil {
		if mhz, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
			ghz := mhz / 1000000.0
			return fmt.Sprintf("%.2f", ghz)
		}
	}

	// 2. Fallback: current value from /proc/cpuinfo
	if out, err := ReadHostOrDefault("/proc/cpuinfo"); err == nil {
		for line := range strings.SplitSeq(string(out), "\n") {
			if strings.HasPrefix(line, "cpu MHz") {
				parts := strings.Split(line, ":")
				if len(parts) != 2 {
					continue
				}
				mhzStr := strings.TrimSpace(parts[1])
				if mhz, err := strconv.ParseFloat(mhzStr, 64); err == nil {
					ghz := math.Ceil(mhz/1000*100) / 100
					return fmt.Sprintf("%.2f", ghz)
				}
			}
		}
	}

	return "0.0"
}
