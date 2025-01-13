package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Server struct {
	mu    sync.Mutex
	store map[string]string
}

func NewServer() *Server {
	return &Server{
		store: make(map[string]string),
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", err)
			return
		}

		// Process the command
		message = strings.TrimSpace(message)
		parts := strings.SplitN(message, " ", 3)
		command := strings.ToUpper(parts[0])

		switch command {
		case "PING":
			_, _ = conn.Write([]byte("PONG\n"))
		case "SET":
			if len(parts) < 3 {
				_, _ = conn.Write([]byte("Error: SET requires a key and value\n"))
				continue
			}
			key, value := parts[1], parts[2]
			s.mu.Lock()
			s.store[key] = value
			s.mu.Unlock()
			_, _ = conn.Write([]byte("OK\n"))
		case "GET":
			if len(parts) < 2 {
				_, _ = conn.Write([]byte("Error: GET requires a key\n"))
				continue
			}
			key := parts[1]
			s.mu.Lock()
			value, exists := s.store[key]
			s.mu.Unlock()
			if exists {
				_, _ = conn.Write([]byte(value + "\n"))
			} else {
				_, _ = conn.Write([]byte("nil\n"))
			}
		default:
			_, _ = conn.Write([]byte("Unknown Command\n"))
		}
	}
}

func main() {
	server := NewServer()

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 6379...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go server.handleConnection(conn)
	}
}
