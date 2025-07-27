package server

import (
	"dovedashboard/internal/api"
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed web/*
var content embed.FS

func Start() error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/cpu", api.HandleCPU)
	mux.HandleFunc("/api/mem", api.HandleMem)
	mux.HandleFunc("/api/storage", api.HandleStorage)
	mux.HandleFunc("/api/sensors", api.HandleSensors)
	mux.HandleFunc("/api/os", api.HandleOs)
	mux.HandleFunc("/api/net", api.HandleNet)

	// Static frontend
	subFS, err := fs.Sub(content, "web")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	log.Println("The Dove Dashboard server has started. GLHF!")
	return http.ListenAndServe(":2137", mux)
}
