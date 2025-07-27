package sysinfo

import (
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type CPUInfo struct {
	Brand     string `json:"brand"`     // e.g. AMD
	Model     string `json:"model"`     // e.g. Ryzen 7 5800H with Radeon Graphics
	Cores     int    `json:"cores"`     // e.g. 4
	Threads   int    `json:"threads"`   // e.g. 8
	Frequency string `json:"frequency"` // e.g. 3.2 GHz
}

func GetCPUInfo() CPUInfo {
	brand, model := getCPUBrandAndModel()
	cores := runtime.NumCPU() / 2
	threads := runtime.NumCPU()
	freq := getCPUClock()

	return CPUInfo{
		Brand:     brand,
		Model:     model,
		Cores:     cores,
		Threads:   threads,
		Frequency: freq,
	}
}

func getCPUBrandAndModel() (string, string) {
	out, err := exec.Command("lscpu").Output()
	if err != nil {
		return "Unknown", "Unknown"
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.HasPrefix(line, "Model name:") {
			full := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])

			words := strings.Fields(full)
			if len(words) > 1 {
				brand := words[0]
				model := strings.Join(words[1:], " ")
				return brand, model
			}
			return full, ""
		}
	}
	return "Unknown", "Unknown"
}

func getCPUClock() string {
	out, err := exec.Command("cat", "/proc/cpuinfo").Output()
	if err != nil {
		return "0.0"
	}

	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}
			mhzStr := strings.TrimSpace(parts[1])
			mhz, err := strconv.ParseFloat(mhzStr, 64)
			if err == nil {
				ghz := math.Ceil(mhz/1000*100) / 100
				return fmt.Sprintf("%.2f", ghz)
			}
		}
	}

	return "0.0"
}
