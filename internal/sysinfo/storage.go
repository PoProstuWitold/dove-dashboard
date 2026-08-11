package sysinfo

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

type StorageInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	FSType      string  `json:"fsType"`
	Type        string  `json:"type"`
	TotalMiB    uint64  `json:"totalMiB"`
	UsedMiB     uint64  `json:"usedMiB"`
	FreeMiB     uint64  `json:"freeMiB"`
	UsedPercent float64 `json:"usedPercent"`
}

type Mount struct {
	Mountpoint string
	Device     string
	FSType     string
}

var ignoredFilesystems = map[string]struct{}{
	"autofs": {}, "bpf": {}, "cgroup": {}, "cgroup2": {}, "configfs": {},
	"debugfs": {}, "devpts": {}, "devtmpfs": {}, "efivarfs": {}, "fusectl": {},
	"hugetlbfs": {}, "mqueue": {}, "nsfs": {}, "overlay": {}, "proc": {},
	"pstore": {}, "rpc_pipefs": {}, "securityfs": {}, "selinuxfs": {}, "squashfs": {},
	"sysfs": {}, "tmpfs": {}, "tracefs": {},
}

func GetStorageInfo() []StorageInfo {
	mounts := readMounts()
	result := make([]StorageInfo, 0, len(mounts))
	seen := make(map[string]struct{}, len(mounts))

	for _, mount := range mounts {
		if _, ok := seen[mount.Mountpoint]; ok {
			continue
		}
		seen[mount.Mountpoint] = struct{}{}

		if info, ok := storageInfo(mount); ok {
			result = append(result, info)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Mountpoint == "/" {
			return true
		}
		if result[j].Mountpoint == "/" {
			return false
		}
		return result[i].Mountpoint < result[j].Mountpoint
	})

	return result
}

func readMounts() []Mount {
	paths := []string{"/proc/self/mountinfo"}
	if _, err := os.Stat("/mnt/host/proc/1/mountinfo"); err == nil {
		paths = []string{"/mnt/host/proc/1/mountinfo", "/proc/self/mountinfo"}
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if mounts := parseMountInfo(string(data)); len(mounts) > 0 {
			return mounts
		}
	}

	return nil
}

func parseMountInfo(content string) []Mount {
	var mounts []Mount

	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		sep := -1
		for i := 6; i < len(fields); i++ {
			if fields[i] == "-" {
				sep = i
				break
			}
		}
		if sep < 0 || sep+2 >= len(fields) {
			continue
		}

		mountpoint := cleanMountField(fields[4])
		fstype := fields[sep+1]
		device := cleanMountField(fields[sep+2])

		if ignoredMount(fstype, mountpoint) {
			continue
		}

		mounts = append(mounts, Mount{
			Mountpoint: hostMountpoint(mountpoint),
			Device:     device,
			FSType:     fstype,
		})
	}

	return mounts
}

func ignoredMount(fstype, mountpoint string) bool {
	if _, ok := ignoredFilesystems[fstype]; ok {
		return true
	}

	if fstype == "fuse.portal" || fstype == "fuse.gvfsd-fuse" {
		return true
	}

	mp := hostMountpoint(mountpoint)
	return strings.HasPrefix(mp, "/proc/") ||
		strings.HasPrefix(mp, "/sys/") ||
		strings.HasPrefix(mp, "/dev/") ||
		strings.HasPrefix(mp, "/var/lib/docker/") ||
		strings.HasPrefix(mp, "/var/lib/containers/")
}

func cleanMountField(value string) string {
	return strings.NewReplacer(
		`\040`, " ",
		`\011`, "\t",
		`\012`, "\n",
		`\134`, `\`,
	).Replace(value)
}

func hostMountpoint(path string) string {
	if path == "/mnt/host" {
		return "/"
	}
	if strings.HasPrefix(path, "/mnt/host/") {
		return "/" + strings.TrimPrefix(path, "/mnt/host/")
	}
	return path
}

func storageInfo(mount Mount) (StorageInfo, bool) {
	var stat syscall.Statfs_t
	path := mount.Mountpoint

	if _, err := os.Stat("/mnt/host"); err == nil {
		if path == "/" {
			path = "/mnt/host"
		} else {
			path = filepath.Join("/mnt/host", strings.TrimPrefix(path, "/"))
		}
	}

	if err := syscall.Statfs(path, &stat); err != nil {
		return StorageInfo{}, false
	}

	total := stat.Blocks * uint64(stat.Bsize) / 1024 / 1024
	free := stat.Bfree * uint64(stat.Bsize) / 1024 / 1024
	used := total - free
	percent := 0.0
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}

	return StorageInfo{
		Device:      mount.Device,
		Mountpoint:  mount.Mountpoint,
		FSType:      mount.FSType,
		Type:        detectDiskType(mount.Device),
		TotalMiB:    total,
		UsedMiB:     used,
		FreeMiB:     free,
		UsedPercent: percent,
	}, true
}

func detectDiskType(device string) string {
	name := blockDeviceName(device)
	if name == "" {
		return "Unknown"
	}
	return classifyBlockDevice(name, map[string]bool{})
}

func blockDeviceName(device string) string {
	if !strings.HasPrefix(device, "/dev/") {
		return ""
	}

	path := ResolveHostPath(device)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}

	return filepath.Base(path)
}

func classifyBlockDevice(name string, visiting map[string]bool) string {
	if name == "" || visiting[name] {
		return "Unknown"
	}
	visiting[name] = true
	defer delete(visiting, name)

	if strings.HasPrefix(name, "nvme") {
		return "NVMe"
	}

	if slaves := blockSlaves(name); len(slaves) > 0 {
		types := make([]string, 0, len(slaves))
		for _, slave := range slaves {
			types = append(types, classifyBlockDevice(slave, visiting))
		}
		return mergeDiskTypes(types)
	}

	if parent := partitionParent(name); parent != "" {
		return classifyBlockDevice(parent, visiting)
	}

	data, err := os.ReadFile(ResolveHostPath(filepath.Join("/sys/class/block", name, "queue/rotational")))
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

func blockSlaves(name string) []string {
	entries, err := os.ReadDir(ResolveHostPath(filepath.Join("/sys/class/block", name, "slaves")))
	if err != nil {
		return nil
	}

	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Name())
	}
	return result
}

func partitionParent(name string) string {
	partition := ResolveHostPath(filepath.Join("/sys/class/block", name, "partition"))
	if _, err := os.Stat(partition); err != nil {
		return ""
	}

	path, err := filepath.EvalSymlinks(ResolveHostPath(filepath.Join("/sys/class/block", name)))
	if err != nil {
		return ""
	}

	parent := filepath.Base(filepath.Dir(path))
	if parent == "block" || parent == "virtual" || parent == name {
		return ""
	}
	return parent
}

func mergeDiskTypes(types []string) string {
	result := ""
	for _, current := range types {
		if current == "Unknown" {
			continue
		}
		if result == "" {
			result = current
			continue
		}
		if result != current {
			return "Mixed"
		}
	}
	if result == "" {
		return "Unknown"
	}
	return result
}
