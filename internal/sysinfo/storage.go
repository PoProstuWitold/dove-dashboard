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
	TotalMiB    uint64  `json:"totalMiB"`
	UsedMiB     uint64  `json:"usedMiB"`
	FreeMiB     uint64  `json:"freeMiB"`
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
		TotalMiB:    total,
		UsedMiB:     used,
		FreeMiB:     free,
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

	phys := resolvePhysicalBlock(devName)

	if strings.HasPrefix(phys, "nvme") {
		return "NVMe"
	}

	if strings.HasPrefix(phys, "mmcblk") {
		return "SSD"
	}

	rotationalPath := ResolveHostPath("/sys/block/" + phys + "/queue/rotational")
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

func resolvePhysicalBlock(name string) string {
	for range 5 {
		slavesDir := ResolveHostPath("/sys/block/" + name + "/slaves")
		entries, err := os.ReadDir(slavesDir)
		if err != nil || len(entries) == 0 {
			return name
		}
		name = entries[0].Name()
	}
	return name
}
