package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read the client input
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", err)
			return
		}

		// Trim the input and process commands
		message = strings.TrimSpace(message)
		fmt.Printf("Received: %s\n", message)

		// Handle the "PING" command
		switch strings.ToUpper(message) {
		case "PING":
			_, _ = conn.Write([]byte("PONG\n"))
		case "SET":
			_, _ = conn.Write([]byte("Not Defined!!\n"))
		case "GET":
			_, _ = conn.Write([]byte("Not Defined!!\n"))
		default:
			_, _ = conn.Write([]byte("Unknown Command\n"))
		}
	}
}

func main() {
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

		// Handle each client connection in a separate goroutine
		go handleConnection(conn)
	}
}
