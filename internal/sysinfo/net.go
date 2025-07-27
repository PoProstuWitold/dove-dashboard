package sysinfo

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	lastBenchmark time.Time
	benchmarkLock sync.Mutex
	benchmarkData []NetStats
)

type NetStats struct {
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	SpeedUpMbps   float64   `json:"speedUpMbps"`
	SpeedDownMbps float64   `json:"speedDownMbps"`
	Bandwidth     float64   `json:"bandwidth"`
	LastBenchmark time.Time `json:"lastBenchmark"`
}

func GetNetInfo() []NetStats {
	benchmarkLock.Lock()
	defer benchmarkLock.Unlock()

	if time.Since(lastBenchmark) < 4*time.Hour && len(benchmarkData) > 0 {
		for i := range benchmarkData {
			benchmarkData[i].LastBenchmark = lastBenchmark
		}
		return benchmarkData
	}

	benchmarkData = runHeavyBenchmark()
	lastBenchmark = time.Now()

	for i := range benchmarkData {
		benchmarkData[i].LastBenchmark = lastBenchmark
	}

	return benchmarkData
}

func getDefaultInterface() string {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "dev" && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}

func readInterfaceSpeed(name string) float64 {
	speedPath := fmt.Sprintf("/sys/class/net/%s/speed", name)
	data, err := os.ReadFile(speedPath)
	if err != nil {
		return 0
	}
	speed := strings.TrimSpace(string(data))
	val, _ := strconv.ParseFloat(speed, 64)
	return val
}

func detectInterfaceType(name string) string {
	if strings.HasPrefix(name, "w") {
		return "wireless"
	}
	return "wired"
}

func runHeavyBenchmark() []NetStats {
	defaultIface := getDefaultInterface()
	if defaultIface == "" {
		return nil
	}

	dlStart := time.Now()
	resp, err := http.Get("http://speedtest.tele2.net/100MB.zip")
	if err != nil {
		fmt.Printf("Download test error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	written, _ := io.Copy(io.Discard, resp.Body)
	dlDuration := time.Since(dlStart).Seconds()
	speedDownMbps := float64(written*8) / (dlDuration * 1_000_000)

	ulStart := time.Now()
	testData := bytes.NewReader(make([]byte, 50*1024*1024))

	req, err := http.NewRequest("PUT", "http://speedtest.tele2.net/upload.php", testData)
	if err != nil {
		fmt.Printf("Upload test setup error: %v\n", err)
		return nil
	}

	client := &http.Client{Timeout: 30 * time.Second}
	ulResp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Upload test error: %v\n", err)
		return nil
	}
	defer ulResp.Body.Close()

	ulDuration := time.Since(ulStart).Seconds()
	speedUpMbps := float64(50*8*1024*1024) / (ulDuration * 1_000_000)

	return []NetStats{
		{
			Name:          defaultIface,
			Type:          detectInterfaceType(defaultIface),
			SpeedDownMbps: speedDownMbps,
			SpeedUpMbps:   speedUpMbps,
			Bandwidth:     readInterfaceSpeed(defaultIface),
			LastBenchmark: time.Now(),
		},
	}
}
