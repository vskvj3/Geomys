package main

import (
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/utils"
)

func main() {
	logger := utils.NewLogger("", true) // "app.log" is the log file, debugMode=true enables debug logs

	// Access home dir for config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Error getting home directory: " + err.Error())
		return
	}

	configPath := filepath.Join(homeDir, ".geomys", "geomys.conf")

	// Load configuration
	config, err := utils.LoadConfig(configPath)
	if err != nil {
		logger.Error("Error loading configuration: " + err.Error())
		return
	}

	// Attempt to bind to the configured port
	port := strconv.Itoa(config.Port)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Info("Port " + port + " unavailable. Selecting a random port...\n")
		listener, err = net.Listen("tcp", ":0") // Random port
		if err != nil {
			logger.Error("Error starting server: " + err.Error())
			return
		}
	}

	defer listener.Close()
	logger.Info("Server is listening on " + listener.Addr().String())

	server := network.NewServer()
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection: " + err.Error())
			continue
		}
		go server.HandleConnection(conn)
	}
}
