package api

import (
	"dovedashboard/internal/sysinfo"
	"encoding/json"
	"net/http"
)

func HandleCPU(w http.ResponseWriter, r *http.Request) {
	info := sysinfo.GetCPUInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
