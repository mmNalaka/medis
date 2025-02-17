package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mmnalaka/medis/internal/server"
)

const (
	REDIS_PORT = 6379
)

func main() {
	// Create a context that will be canceled on interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensures that cancel() is called before the function exits

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)                      //  Creates a channel that will receive OS signals.
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // Registers the channel to receive interrupt (SIGINT) and termination (SIGTERM) signals.

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating shutdown", sig)
		cancel()
	}()

	server := server.NewServer(REDIS_PORT)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start Medis server: %v", err)
	}
}
