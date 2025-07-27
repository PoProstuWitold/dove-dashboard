package sysinfo

import (
	"fmt"
	"os/exec"
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
	}
}

func getKernelVersion() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(out))
}

func getHostname() string {
	out, err := exec.Command("hostname").Output()
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(out))
}

func getOSInfoFields() (name, id string) {
	out, err := exec.Command("cat", "/etc/os-release").Output()
	if err != nil {
		return "Unknown OS", "unknown"
	}

	lines := strings.SplitSeq(string(out), "\n")
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
	arch := runtime.GOARCH
	if arch == "amd64" {
		return "x64"
	}
	return arch
}

func getUptime() string {
	out, err := exec.Command("cat", "/proc/uptime").Output()
	if err != nil {
		return "Unknown"
	}

	fields := strings.Fields(string(out))
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
