package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/utils"
)

func main() {
	logger := utils.NewLogger("", true)

	// Load configurations from home directory
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

	// Parse command-line arguments
	nodeIdPtr := flag.String("node_id", "", "Node ID of the current node")
	portPtr := flag.String("port", "", "Port of server")
	bootstrapPtr := flag.Bool("bootstrap", false, "Start cluster in bootstrap mode (initial leader)")
	joinPtr := flag.String("join", "", "Join an existing cluster (provide leader address in <ip:port>)")
	flag.Parse()

	// Ensure only one of `bootstrap` or `join` is set
	if *bootstrapPtr && *joinPtr != "" {
		logger.Error("Cannot use both -bootstrap and -join. Choose only one.")
		return
	}

	// Determine Node ID
	nodeID := config.NodeID
	if *nodeIdPtr != "" {
		nodeID, err = strconv.Atoi(*nodeIdPtr)
		if err != nil {
			logger.Error("Invalid node_id: must be an integer")
			return
		}
	}
	logger.Info(fmt.Sprintf("Node ID assigned: %d", nodeID))

	// Determine Port
	port := strconv.Itoa(config.InternalPort)
	if *portPtr != "" {
		port = *portPtr
		config.InternalPort, err = strconv.Atoi(*portPtr)
		if err != nil {
			logger.Warn("Copying internal port from flag to config failed")
		}
	}
	logger.Info("Port assigned: " + port)

	// Handle Bootstrap Mode (Start the Leader Node)
	if *bootstrapPtr {
		logger.Info("Starting in bootstrap mode...")
		startGRPCServer(port)
		return
	}

	// Handle Join Mode (Connect to Leader)
	if *joinPtr != "" {
		logger.Info("Joining existing cluster at " + *joinPtr)
		joinCluster(*joinPtr, port)
		return
	}

	// If neither bootstrap nor join, start a standalone node
	logger.Info("Starting standalone node...")

	// Attempt to bind to the configured port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Warn("Port " + port + " unavailable. Selecting a random port...")
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			logger.Error("Error starting server: " + err.Error())
			return
		}
	}

	defer listener.Close()
	logger.Info("Server is listening on " + listener.Addr().String())

	// Create the network server
	server, err := network.NewServer()
	if err != nil {
		logger.Error("Server creation failed: " + err.Error())
		return
	}

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection: " + err.Error())
			continue
		}
		logger.Info("Accepted client: " + conn.RemoteAddr().String())
		go server.HandleConnection(conn)
	}
}

// Starts the gRPC server (Leader Node)
func startGRPCServer(port string) {
	logger := utils.NewLogger("", true)
	logger.Info("Starting gRPC server on port " + port)

}

// Joins an existing cluster (Follower Node)
func joinCluster(leaderAddr string, port string) {
	logger := utils.NewLogger("", true)
	logger.Info("Attempting to join cluster at " + leaderAddr + port)
}
