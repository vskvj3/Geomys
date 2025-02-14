package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vskvj3/geomys/internal/cluster"
	"github.com/vskvj3/geomys/internal/cluster/election"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/utils"
)

func main() {
	logger := utils.NewLogger("", true)

	// Load Configurations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Error getting home directory: " + err.Error())
		return
	}
	configPath := filepath.Join(homeDir, ".geomys", "geomys.conf")
	utils.LoadConfig(configPath)

	config, err := utils.GetConfig()
	if err != nil {
		logger.Error("Error loading configuration: " + err.Error())
		return
	}
	logger.Info("Loaded configurations from " + configPath)

	// Parse Command-Line Flags
	nodeIdPtr := flag.String("node_id", "", "Node ID of the current node")
	portPtr := flag.String("port", "", "Port of the server")
	bootstrapPtr := flag.Bool("bootstrap", false, "Start cluster in bootstrap mode (leader)")
	joinPtr := flag.String("join", "", "Join an existing cluster (provide leader address in <ip:port>)")
	flag.Parse()

	if *bootstrapPtr && *joinPtr != "" {
		logger.Error("Cannot use both -bootstrap and -join. Choose only one.")
		return
	}

	// Determine Node ID
	nodeID := config.NodeID
	if *nodeIdPtr != "" {
		parsedID, err := strconv.Atoi(*nodeIdPtr)
		if err != nil {
			logger.Error("Invalid node_id: must be an integer")
			return
		}
		nodeID = parsedID
		config.NodeID = nodeID
	}
	logger.Info(fmt.Sprintf("Node ID assigned: %d", nodeID))

	// Determine Port
	port := config.InternalPort
	if *portPtr != "" {
		parsedPort, err := strconv.Atoi(*portPtr)
		if err == nil {
			port = parsedPort
			config.InternalPort = port
			config.ExternalPort = port + 1000
		} else {
			logger.Warn("Invalid port value. Using default from config.")
		}
	}
	logger.Info("TCP Port assigned: " + strconv.Itoa(config.InternalPort))
	logger.Info("gRPC Port assigned: " + strconv.Itoa(config.ExternalPort))

	// Initialize Core Components
	db := core.NewDatabase()
	commandHandler := core.NewCommandHandler(db)

	// Initialize Cluster and Replication Servers
	clusterServer := cluster.NewClusterServer(int32(nodeID), int32(config.ExternalPort))

	// Configure Node Mode (Bootstrap, Join, or Standalone)
	if *bootstrapPtr {
		logger.Info("Starting in bootstrap mode (leader)...")
		config.IsLeader = true
		config.ClusterMode = true
		clusterServer.SetLeaderAddress("")
		clusterServer.SetLeaderID(int32(nodeID))
		// start grpc server in leader mode
		go clusterServer.StartServer(commandHandler)

		// wait for 15 seconds, and start monitoring followers
		time.Sleep(15 * time.Second)
		go clusterServer.ElectionService.MonitorFollowers()
	} else if *joinPtr != "" {
		logger.Info("Joining existing cluster at " + *joinPtr)
		config.IsLeader = false
		config.ClusterMode = true
		clusterServer.SetLeaderAddress(*joinPtr)
		// start grpc server in follower mode
		go clusterServer.StartServer(commandHandler)

		// wait for 15 seconds, and start monitoring leader
		time.Sleep(15 * time.Second)

		client, err := election.NewElectionClient(*joinPtr)
		if err != nil {
			logger.Error("Failed to connect to leader: " + err.Error())
			return
		}
		go client.MonitorLeader(clusterServer.ElectionService)
	} else {
		config.IsLeader = true
		config.ClusterMode = false
		logger.Info("Starting standalone node...")
	}

	// Start TCP Server
	logger.Debug("Initializing TCP server on port " + strconv.Itoa(port))
	server, err := network.NewServer(clusterServer, strconv.Itoa(port), commandHandler)
	if err != nil {
		logger.Error("TCP Server creation failed: " + err.Error())
		return
	}
	logger.Debug("Starting TCP server...")
	go server.Start()

	// Block forever to keep the server running
	select {}
}
