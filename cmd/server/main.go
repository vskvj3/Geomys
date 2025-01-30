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
	logger := utils.NewLogger("", true)

	// Try to load configs from homedir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Error getting home directory: " + err.Error())
		return
	}

	configPath := filepath.Join(homeDir, ".geomys", "geomys.conf")

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		logger.Error("Error loading configuration: " + err.Error())
		return
	}
	logger.Info("Loaded configurations from " + configPath)

	// Attempt to bind to the configured port
	port := strconv.Itoa(config.Port)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Warn("Port " + port + " unavailable. Selecting a random port...\n")
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			logger.Error("Error starting server: " + err.Error())
			return
		}
	}

	defer listener.Close()
	logger.Info("Server is listening on " + listener.Addr().String())

	server, err := network.NewServer()
	if err != nil {
		logger.Error("Server creation failed" + err.Error())
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection: " + err.Error())
			continue
		}
		logger.Info("Accepted client :" + conn.RemoteAddr().String())
		go server.HandleConnection(conn)
	}
}
