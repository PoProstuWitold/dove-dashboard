package sysinfo

import (
	"os/exec"
	"strconv"
	"strings"
)

type MemInfo struct {
	TotalMB     uint64  `json:"totalMB"`
	UsedMB      uint64  `json:"usedMB"`
	FreeMB      uint64  `json:"freeMB"`
	UsedPercent float64 `json:"usedPercent"`
}

func GetMemInfo() MemInfo {
	mem := MemInfo{}

	if out, err := exec.Command("free", "-m").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 4 {
				total, _ := strconv.ParseUint(fields[1], 10, 64)
				used, _ := strconv.ParseUint(fields[2], 10, 64)
				free, _ := strconv.ParseUint(fields[3], 10, 64)

				mem.TotalMB = total
				mem.UsedMB = used
				mem.FreeMB = free
				mem.UsedPercent = float64(used) / float64(total) * 100
			}
		}
	}

	return mem
}
