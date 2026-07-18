package api

import (
	"dove-dashboard/internal/sysinfo"
	"encoding/json"
	"net/http"
)

func HandleStorage(w http.ResponseWriter, r *http.Request) {
	info := sysinfo.GetStorageInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
