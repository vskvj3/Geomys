package integration

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/vskvj3/geomys/internal/network"
)

// Helper function to send a command and receive the response
func sendCommand(t *testing.T, conn net.Conn, command string) string {
	_, err := conn.Write([]byte(command + "\n"))
	if err != nil {
		t.Fatalf("failed to send command: %v", err)
	}
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	return strings.TrimSpace(response)
}

func TestIntegration(t *testing.T) {
	// Start the server in a goroutine
	go func() {
		server := network.NewServer()
		listener, _ := net.Listen("tcp", ":6379")
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.HandleConnection(conn)
		}
	}()
	time.Sleep(100 * time.Millisecond) // Allow the server time to start

	// Connect to the server
	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Test PING command
	t.Run("PING command", func(t *testing.T) {
		response := sendCommand(t, conn, "PING")
		if response != "PONG" {
			t.Errorf("expected PONG, got %v", response)
		}
	})

	// Test ECHO command
	t.Run("ECHO command", func(t *testing.T) {
		response := sendCommand(t, conn, "ECHO Hello")
		if response != "Hello" {
			t.Errorf("expected Hello, got %v", response)
		}
	})

	// Test SET command
	t.Run("SET command", func(t *testing.T) {
		response := sendCommand(t, conn, "SET mykey myvalue")
		if response != "OK" {
			t.Errorf("expected OK, got %v", response)
		}
	})

	// Test GET command
	t.Run("GET command existing key", func(t *testing.T) {
		_ = sendCommand(t, conn, "SET mykey myvalue")
		response := sendCommand(t, conn, "GET mykey")
		if response != "myvalue" {
			t.Errorf("expected myvalue, got %v", response)
		}
	})

	t.Run("GET command non-existing key", func(t *testing.T) {
		response := sendCommand(t, conn, "GET nonexistent")
		if response != "Error: key not found" {
			t.Errorf("expected nil, got %v", response)
		}
	})
}
