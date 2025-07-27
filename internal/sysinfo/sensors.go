package sysinfo

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SensorReading struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Extra string  `json:"extra,omitempty"`
}

type SensorChip struct {
	Name     string          `json:"name"`
	Adapter  string          `json:"adapter"`
	Readings []SensorReading `json:"readings"`
}

func GetSensors() []SensorChip {
	var chips []SensorChip
	var current SensorChip

	reReading := regexp.MustCompile(`^([\w\s\.\-\+\(\)/]+):\s+\+?(-?[0-9.]+)\s*([A-Za-z°%]+)?\s*(\(.*\))?$`)

	out, err := exec.Command("sensors").Output()
	if err != nil {
		return chips
	}

	lines := strings.SplitSeq(string(out), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Name != "" {
				chips = append(chips, current)
				current = SensorChip{}
			}
			continue
		}

		if after, ok := strings.CutPrefix(line, "Adapter:"); ok {
			current.Adapter = strings.TrimSpace(after)
			continue
		}

		if !strings.Contains(line, ":") && current.Name == "" {
			current.Name = line
			continue
		}

		if matches := reReading.FindStringSubmatch(line); len(matches) >= 4 {
			value, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				continue
			}

			reading := SensorReading{
				Label: strings.TrimSpace(matches[1]),
				Value: value,
				Unit:  strings.TrimSpace(matches[3]),
			}

			if len(matches) == 5 {
				reading.Extra = strings.TrimSpace(matches[4])
			}

			current.Readings = append(current.Readings, reading)
		}
	}

	if current.Name != "" {
		chips = append(chips, current)
	}

	return chips
}
