package api

import (
	"dovedashboard/internal/sysinfo"
	"encoding/json"
	"net/http"
)

func HandleOs(w http.ResponseWriter, r *http.Request) {
	info := sysinfo.GetOSInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
