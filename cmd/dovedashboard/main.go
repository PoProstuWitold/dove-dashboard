package main

import (
	server "dovedashboard/internal"
	"log"
)

func main() {
	if err := server.Start(); err != nil {
		log.Fatalf("The Dove Dashboard server failed to start: %v", err)
	}
}
