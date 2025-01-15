package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/utils"
)

func main() {
	// Access home dir for config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	configPath := filepath.Join(homeDir, ".geomys", "geomys.conf")

	// Load configuration
	config, err := utils.LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	// Attempt to bind to the configured port
	port := strconv.Itoa(config.Port)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Port %s unavailable. Selecting a random port...\n", port)
		listener, err = net.Listen("tcp", ":0") // Random port
		if err != nil {
			fmt.Println("Error starting server:", err)
			return
		}
	}

	defer listener.Close()
	fmt.Printf("Server is listening on %s...\n", listener.Addr().String())

	server := network.NewServer()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go server.HandleConnection(conn)
	}
}
