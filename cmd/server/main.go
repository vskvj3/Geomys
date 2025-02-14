package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vskvj3/geomys/internal/cluster"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/replicate"
	"github.com/vskvj3/geomys/internal/utils"
)

func main() {
	logger := utils.NewLogger("", true)

	// Load configurations
	config, err := loadConfig(logger)
	if err != nil {
		return
	}

	// Parse command-line arguments
	nodeID, port, bootstrap, joinAddr := parseFlags(config, logger)
	if nodeID == -1 {
		return
	}

	// Initialize database and command handler
	db := core.NewDatabase()
	commandHandler := core.NewCommandHandler(db)

	// Initialize gRPC servers
	clusterServer := cluster.NewGrpcServer(int32(nodeID), int32(port+1000))
	replicationServer := replicate.NewReplicationServer(commandHandler)

	// Determine node mode (Bootstrap, Join, or Standalone)
	startNodeMode(bootstrap, joinAddr, clusterServer, replicationServer, nodeID, logger)

	// Start TCP server for handling client requests
	startTCPServer(clusterServer, port, commandHandler, logger)

	// Block forever to keep the server running
	select {}
}

// Loads configuration from file
func loadConfig(logger *utils.Logger) (*utils.Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Error getting home directory: " + err.Error())
		return nil, err
	}
	configPath := filepath.Join(homeDir, ".geomys", "geomys.conf")
	utils.LoadConfig(configPath)

	config, err := utils.GetConfig()
	if err != nil {
		logger.Error("Error loading configuration: " + err.Error())
		return nil, err
	}
	logger.Info("Loaded configurations from " + configPath)
	return config, nil
}

// Parses command-line flags and determines node ID and port
func parseFlags(config *utils.Config, logger *utils.Logger) (int, int, bool, string) {
	nodeIdPtr := flag.String("node_id", "", "Node ID of the current node")
	portPtr := flag.String("port", "", "Port of the server")
	bootstrapPtr := flag.Bool("bootstrap", false, "Start cluster in bootstrap mode (leader)")
	joinPtr := flag.String("join", "", "Join an existing cluster (provide leader address in <ip:port>)")
	flag.Parse()

	// Validate bootstrap & join flags
	if *bootstrapPtr && *joinPtr != "" {
		logger.Error("Cannot use both -bootstrap and -join. Choose only one.")
		return -1, -1, false, ""
	}

	// Determine Node ID
	nodeID := config.NodeID
	if *nodeIdPtr != "" {
		parsedID, err := strconv.Atoi(*nodeIdPtr)
		if err != nil {
			logger.Error("Invalid node_id: must be an integer")
			return -1, -1, false, ""
		}
		nodeID = parsedID
		config.NodeID = nodeID
	}
	logger.Info(fmt.Sprintf("Node ID assigned: %d", nodeID))

	// Determine Port
	port := config.InternalPort
	if *portPtr != "" {
		parsedPort, err := strconv.Atoi(*portPtr)
		if err != nil {
			logger.Warn("Invalid port value. Using default from config.")
		} else {
			port = parsedPort
			config.InternalPort = port
			config.ExternalPort = port + 1000
		}
	}
	logger.Info("TCP Port assigned: " + strconv.Itoa(config.InternalPort))
	logger.Info("gRPC Port assigned: " + strconv.Itoa(config.ExternalPort))

	return nodeID, port, *bootstrapPtr, *joinPtr
}

// Determines and starts node mode (Bootstrap, Join, or Standalone)
func startNodeMode(bootstrap bool, joinAddr string, clusterServer *cluster.GrpcServer, replicationServer *replicate.ReplicationServer, nodeID int, logger *utils.Logger) {
	config, err := utils.GetConfig()
	if err != nil {
		logger.Error("Failed to load config: " + err.Error())
	}
	switch {
	case bootstrap:
		logger.Info("Starting in bootstrap mode (leader)...")
		config.IsLeader = true
		config.ClusterMode = true
		clusterServer.LeaderAddress = ""
		clusterServer.LeaderID = nodeID
		go clusterServer.StartServer(int(clusterServer.Port), replicationServer)
		go clusterServer.MonitorFollowers()

	case joinAddr != "":
		logger.Info("Joining existing cluster at " + joinAddr)
		config.IsLeader = false
		config.ClusterMode = true
		clusterServer.LeaderAddress = joinAddr
		go clusterServer.StartServer(int(clusterServer.Port), replicationServer)
		if err := joinCluster(joinAddr, clusterServer); err != nil {
			logger.Error("Failed to join cluster: " + err.Error())
		}

	default:
		logger.Info("Starting standalone node...")
	}
}

func startTCPServer(clusterServer *cluster.GrpcServer, port int, commandHandler *core.CommandHandler, logger *utils.Logger) {
	logger.Debug("Initializing TCP server on port " + strconv.Itoa(port))
	server, err := network.NewServer(clusterServer, strconv.Itoa(port), commandHandler)
	if err != nil {
		logger.Error("TCP Server creation failed: " + err.Error())
		return
	}
	logger.Debug("Starting TCP server...")
	go server.Start()
}

// Joins an existing cluster and starts monitoring leader
func joinCluster(leaderAddr string, clusterServer *cluster.GrpcServer) error {
	client, err := cluster.NewGrpcClient(leaderAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to leader: %v", err)
	}
	go client.MonitorLeader(clusterServer)
	return nil
}
