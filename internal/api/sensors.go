package api

import (
	"dovedashboard/internal/sysinfo"
	"encoding/json"
	"net/http"
)

func HandleSensors(w http.ResponseWriter, r *http.Request) {
	info := sysinfo.GetSensors()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
