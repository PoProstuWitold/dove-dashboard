package sysinfo

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type StorageInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	FSType      string  `json:"fsType"`
	Type        string  `json:"type"` // HDD / SSD / NVMe
	TotalMB     uint64  `json:"totalMB"`
	UsedMB      uint64  `json:"usedMB"`
	FreeMB      uint64  `json:"freeMB"`
	UsedPercent float64 `json:"usedPercent"`
}

func GetStorageInfo() StorageInfo {
	out, err := exec.Command("df", "-T", "-m", "/").Output()
	if err != nil {
		log.Println("df command error:", err)
		return StorageInfo{}
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		log.Println("df output too short:", lines)
		return StorageInfo{}
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 7 {
		log.Println("unexpected df field count:", fields)
		return StorageInfo{}
	}

	total, err1 := strconv.ParseUint(fields[2], 10, 64)
	used, err2 := strconv.ParseUint(fields[3], 10, 64)
	free, err3 := strconv.ParseUint(fields[4], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil || total == 0 {
		log.Println("failed to parse storage values:", fields)
		return StorageInfo{}
	}

	usedPercent := (float64(used) / float64(total)) * 100

	return StorageInfo{
		Device:      fields[0],
		FSType:      fields[1],
		TotalMB:     total,
		UsedMB:      used,
		FreeMB:      free,
		UsedPercent: usedPercent,
		Mountpoint:  fields[6],
		Type:        detectDiskType(fields[0]),
	}
}

func detectDiskType(device string) string {
	devName := filepath.Base(device)

	if strings.HasPrefix(devName, "mapper/") {
		return "Unknown (LVM)"
	}

	for len(devName) > 0 && devName[len(devName)-1] >= '0' && devName[len(devName)-1] <= '9' {
		devName = devName[:len(devName)-1]
	}

	if strings.HasPrefix(devName, "nvme") {
		return "NVMe"
	}

	rotationalPath := "/sys/block/" + devName + "/queue/rotational"
	data, err := os.ReadFile(rotationalPath)
	if err != nil {
		return "Unknown"
	}
	switch strings.TrimSpace(string(data)) {
	case "0":
		return "SSD"
	case "1":
		return "HDD"
	default:
		return "Unknown"
	}
}
