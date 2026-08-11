package sysinfo

import (
	"net"
	"os"
	"path/filepath"
	"strings"
)

type NetInfo struct {
	Interface string `json:"interface"`
	Type      string `json:"type"`
}

func GetNetInfo() *NetInfo {
	iface := defaultInterface()
	if iface == "" {
		return nil
	}

	return &NetInfo{
		Interface: iface,
		Type:      interfaceType(iface),
	}
}

func defaultInterface() string {
	if data, err := ReadHostOrDefault("/proc/net/route"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 &&
				fields[1] == "00000000" &&
				usableInterface(fields[0]) {
				return fields[0]
			}
		}
	}

	entries, err := os.ReadDir(ResolveHostPath("/sys/class/net"))
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		iface := entry.Name()
		if usableInterface(iface) && interfaceUp(iface) {
			return iface
		}
	}

	return ""
}

func usableInterface(iface string) bool {
	if iface == "" ||
		iface == "lo" ||
		strings.HasPrefix(iface, "docker") ||
		strings.HasPrefix(iface, "veth") ||
		strings.HasPrefix(iface, "br-") {
		return false
	}

	// eth0 is valid on a normal host, but is usually Docker's interface
	// when the dashboard runs inside the existing container setup.
	if iface == "eth0" && hostMounted() {
		return false
	}

	return true
}

func interfaceUp(iface string) bool {
	data, err := os.ReadFile(
		ResolveHostPath(filepath.Join("/sys/class/net", iface, "operstate")),
	)
	return err == nil && strings.TrimSpace(string(data)) == "up"
}

func interfaceType(iface string) string {
	wireless := ResolveHostPath(filepath.Join("/sys/class/net", iface, "wireless"))

	if _, err := os.Stat(wireless); err == nil {
		return "wifi"
	}

	if strings.HasPrefix(iface, "wl") {
		return "wifi"
	}

	return "ethernet"
}

func interfaceIP(iface string) string {
	// net.InterfaceByName operates in the current network namespace.
	// Therefore this is reliable when running directly on the host,
	// but not for host interfaces while running inside Docker.
	if hostMounted() {
		return ""
	}

	device, err := net.InterfaceByName(iface)
	if err != nil {
		return ""
	}

	addrs, err := device.Addrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && ip.To4() != nil && !ip.IsLoopback() {
			return ip.String()
		}
	}

	return ""
}

func hostMounted() bool {
	_, err := os.Stat("/mnt/host")
	return err == nil
}
