package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/mmnalaka/medis/internal/command"
)

type Server struct {
	port     int
	listener net.Listener
	wg       sync.WaitGroup // WaitGroup to track active connections
	handler  command.Handler
}

func NewServer(port int) *Server {
	return &Server{
		port:    port,
		handler: *command.NewHandler(),
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	s.listener = listener
	log.Printf("Server started on %s", addr)

	// Goroutine to handle shutdown when context is canceled
	go func() {
		<-ctx.Done() // Wait for cancellation signal
		log.Println("Shutting down server...")
		s.listener.Close() // Stop accepting new connections
		s.wg.Wait()        // Wait for active connections to finish
		log.Println("Server shutdown complete")
	}()

	// Accept loop for handling connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done(): // If context was canceled, exit gracefully
				return nil
			default:
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
		}

		// Track active connection
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// Handles a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done() // Decrement WaitGroup when function exits
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())

	// Example: Simple echo server
	reader := bufio.NewReader(conn)
	for {
		// Read the incommig command
		data, err := command.ReadCommand(reader)
		if err != nil {
			log.Printf("Failed to read command: %v", err)
			break
		}

		// Parse command
		cmd, err := command.ParseCommand(data)
		if err != nil {
			log.Printf("Failed to parse command: %v", err)
			break
		}

		// Handle the command
		respData := s.handler.Handle(cmd)

		// Write the response back to the client
		_, err = conn.Write(respData.Encode())
		if err != nil {
			log.Printf("Failed to write response: %v", err)
			break
		}
	}
}
