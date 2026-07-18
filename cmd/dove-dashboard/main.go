package main

import (
	server "dove-dashboard/internal"
	"log"
)

func main() {
	if err := server.Start(); err != nil {
		log.Fatalf("The Dove Dashboard server failed to start: %v", err)
	}
}
