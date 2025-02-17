package server

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	port     int
	listener net.Listener
}

func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	s.listener = listener

	log.Printf("Medis server started on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Failed to read from connection: %v", err)
			return
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			log.Printf("Failed to write to connection: %v", err)
			return
		}
	}
}
