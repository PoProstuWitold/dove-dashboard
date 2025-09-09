package sysinfo

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type OSInfo struct {
	Id       string `json:"id"`
	OSName   string `json:"os"`
	Arch     string `json:"arch"`
	Kernel   string `json:"kernel"`
	Uptime   string `json:"uptime"`
	Hostname string `json:"hostname"`
	Device   string `json:"device"` // np. "Lenovo 20W00122PB (ThinkPad T14 Gen 2i)"
}

func GetOSInfo() OSInfo {
	name, id := getOSInfoFields()

	return OSInfo{
		Id:       id,
		OSName:   name,
		Arch:     getArch(),
		Kernel:   getKernelVersion(),
		Uptime:   getUptime(),
		Hostname: getHostname(),
		Device:   getDeviceInfo(),
	}
}

func getKernelVersion() string {
	data, err := ReadHostOrDefault("/proc/sys/kernel/osrelease")
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(data))
}

func getHostname() string {
	data, err := ReadHostOrDefault("/etc/hostname")
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(data))
}

func getOSInfoFields() (name, id string) {
	data, err := ReadHostOrDefault("/etc/os-release")
	if err != nil {
		return "Unknown OS", "unknown"
	}

	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			name = strings.Trim(strings.SplitN(line, "=", 2)[1], `"`)
		}
		if strings.HasPrefix(line, "ID=") {
			id = strings.Trim(strings.SplitN(line, "=", 2)[1], `"`)
		}
	}

	if name == "" {
		name = "Unknown OS"
	}
	if id == "" {
		id = "unknown"
	}
	return
}

func getArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return runtime.GOARCH
}

func getUptime() string {
	data, err := ReadHostOrDefault("/proc/uptime")
	if err != nil {
		return "Unknown"
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "Unknown"
	}

	seconds, err := time.ParseDuration(fields[0] + "s")
	if err != nil {
		return "Unknown"
	}

	days := int(seconds.Hours()) / 24
	hours := int(seconds.Hours()) % 24
	minutes := int(seconds.Minutes()) % 60

	return formatDuration(days, hours, minutes)
}

func formatDuration(days, hours, minutes int) string {
	return strings.TrimSpace(
		strings.Join([]string{
			pluralize(days, "day"),
			pluralize(hours, "hour"),
			pluralize(minutes, "minute"),
		}, " "),
	)
}

func pluralize(n int, unit string) string {
	if n == 0 {
		return ""
	}
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

func getDeviceInfo() string {
	vendorData, _ := ReadHostOrDefault("/sys/devices/virtual/dmi/id/sys_vendor")
	nameData, _ := ReadHostOrDefault("/sys/devices/virtual/dmi/id/product_name")
	versionData, _ := ReadHostOrDefault("/sys/devices/virtual/dmi/id/product_version")

	vendor := strings.TrimSpace(string(vendorData))
	name := strings.TrimSpace(string(nameData))
	version := strings.TrimSpace(string(versionData))

	if vendor == "" && name == "" && version == "" {
		return "Unknown device"
	}

	if version != "" {
		return fmt.Sprintf("%s %s (%s)", vendor, name, version)
	}
	return fmt.Sprintf("%s %s", vendor, name)
}
