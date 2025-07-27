package api

import (
	"dovedashboard/internal/sysinfo"
	"encoding/json"
	"net/http"
)

func HandleMem(w http.ResponseWriter, r *http.Request) {
	info := sysinfo.GetMemInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
