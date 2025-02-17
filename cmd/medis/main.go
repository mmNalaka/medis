package main

import (
	"log"

	"github.com/mmnalaka/medis/internal/server"
)

const (
	REDIS_PORT = 6379
)

func main() {
	server := server.NewServer(REDIS_PORT)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start Medis server: %v", err)
	}
}
