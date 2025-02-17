package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	port     int
	listener net.Listener
	wg       sync.WaitGroup // WaitGroup to track active connections
}

func NewServer(port int) *Server {
	return &Server{port: port}
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
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Connection closed: %v", err)
			return
		}
		conn.Write(buf[:n]) // Echo back the received data
	}
}
