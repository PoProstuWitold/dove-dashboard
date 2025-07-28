package sysinfo

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
	var stat syscall.Statfs_t
	err := syscall.Statfs(ResolveHostPath("/"), &stat)
	if err != nil {
		log.Println("statfs error:", err)
		return StorageInfo{}
	}

	total := stat.Blocks * uint64(stat.Bsize) / 1024 / 1024
	free := stat.Bfree * uint64(stat.Bsize) / 1024 / 1024
	used := total - free
	usedPercent := (float64(used) / float64(total)) * 100

	device, fsType := findRealRootMount()

	return StorageInfo{
		Device:      device,
		Mountpoint:  "/",
		FSType:      fsType,
		Type:        detectDiskType(device),
		TotalMB:     total,
		UsedMB:      used,
		FreeMB:      free,
		UsedPercent: usedPercent,
	}
}

func findRealRootMount() (string, string) {
	mountData, err := ReadHostOrDefault("/proc/mounts")
	if err != nil {
		log.Println("mounts read error:", err)
		return "/dev/root", "unknown"
	}

	lines := strings.Split(string(mountData), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && (fields[1] == "/" || fields[1] == "/mnt/host") && !strings.HasPrefix(fields[0], "overlay") {
			return fields[0], fields[2]
		}
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && !strings.HasPrefix(fields[2], "overlay") && !strings.HasPrefix(fields[2], "tmpfs") {
			return fields[0], fields[2]
		}
	}

	return "/dev/root", "unknown"
}

func detectDiskType(device string) string {
	devName := filepath.Base(device)

	if strings.HasPrefix(devName, "root") {
		if target, err := os.Readlink(ResolveHostPath("/dev/" + devName)); err == nil {
			devName = filepath.Base(target)
		}
	}

	for {
		if len(devName) > 0 && devName[len(devName)-1] >= '0' && devName[len(devName)-1] <= '9' {
			devName = devName[:len(devName)-1]
		} else if strings.HasSuffix(devName, "p") {
			devName = devName[:len(devName)-1]
		} else {
			break
		}
	}

	if strings.HasPrefix(devName, "nvme") {
		return "NVMe"
	}

	rotationalPath := ResolveHostPath("/sys/block/" + devName + "/queue/rotational")
	data, err := ReadHostOrDefault(rotationalPath)
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
