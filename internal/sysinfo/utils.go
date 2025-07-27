package sysinfo

import (
	"os"
	"path/filepath"
)

// ReadHostOrDefault tries to read a file from mounted host (/mnt/host), and falls back to the local path if not available.
func ReadHostOrDefault(path string) ([]byte, error) {
	hostPath := filepath.Join("/mnt/host", path)

	// Check if the host file exists and is readable
	if fi, err := os.Stat(hostPath); err == nil && !fi.IsDir() {
		return os.ReadFile(hostPath)
	}

	// Fall back to the local path
	return os.ReadFile(path)
}

func ResolveHostPath(path string) string {
	hostRoot := "/mnt/host"
	if st, err := os.Stat(hostRoot); err == nil && st.IsDir() {
		return filepath.Join(hostRoot, path)
	}
	return path
}
