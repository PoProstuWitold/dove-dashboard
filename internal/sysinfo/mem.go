package sysinfo

import (
	"strconv"
	"strings"
)

type MemInfo struct {
	TotalMiB    uint64  `json:"totalMiB"`
	UsedMiB     uint64  `json:"usedMiB"`
	FreeMiB     uint64  `json:"freeMiB"`
	UsedPercent float64 `json:"usedPercent"`
}

func GetMemInfo() MemInfo {
	data, err := ReadHostOrDefault("/proc/meminfo")
	if err != nil {
		return MemInfo{}
	}

	var total, available uint64

	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			total = extractKB(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			available = extractKB(line)
		}
	}

	used := total - available

	return MemInfo{
		TotalMiB:    total / 1024,
		UsedMiB:     used / 1024,
		FreeMiB:     available / 1024,
		UsedPercent: float64(used) / float64(total) * 100,
	}
}

func extractKB(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	val, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return val
}
