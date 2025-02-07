package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vskvj3/geomys/internal/cluster"
	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/utils"
)

func main() {
	logger := utils.NewLogger("", true)

	// Load configurations
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
			logger.Warn("Failed to copy internal port from flag to config")
		}
		externalPort, err := strconv.Atoi(*portPtr)
		externalPort += 1000
		config.ExternalPort = externalPort
		if err != nil {
			logger.Warn("Failed to copy internal port from flag to config")
		}
	}
	logger.Info("Port assigned: " + port)
	grpcPort, err := strconv.Atoi(port)
	if err != nil {
		logger.Error("Port must be an integer: " + err.Error())
	}
	grpcPort += 1000

	// Create clustering server instance
	clusterServer := cluster.NewElectionServer(int32(nodeID))

	if *bootstrapPtr {
		// Bootstrap Mode (Start the Leader Node)
		logger.Info("Starting in bootstrap mode (leader)...")
		clusterServer.LeaderID = nodeID
		go clusterServer.StartServer(grpcPort) // Start gRPC server for election
	} else if *joinPtr != "" {
		// Join Mode (Follower Node)
		logger.Info("Joining existing cluster at " + *joinPtr)
		go clusterServer.StartServer(grpcPort)  // Start gRPC server for election
		go joinCluster(*joinPtr, clusterServer) // Start heartbeat mechanism
	} else {
		// Standalone Mode
		logger.Info("Starting standalone node...")
	}

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

// Joins an existing cluster (Follower Node) and starts sending heartbeats
func joinCluster(leaderAddr string, clusterServer *cluster.GrpcServer) {
	logger := utils.NewLogger("", true)

	client, err := cluster.NewElectionClient(leaderAddr)
	if err != nil {
		logger.Error("Failed to connect to leader: " + err.Error())
		return
	}

	// Send heartbeats every 5 seconds
	for {
		success := client.SendHeartbeat(int(clusterServer.NodeID))
		if !success {
			logger.Warn("Failed to send heartbeat. Checking leader status...")

			// Detect if leader is down (no response for 3 cycles)
			clusterServer.VoteLock.Lock()
			lastHeartbeat, exists := clusterServer.Heartbeats[clusterServer.LeaderID]
			if !exists {
				logger.Warn("Cannot find last heartbeat...")
			}
			fmt.Println(clusterServer.Heartbeats)
			clusterServer.VoteLock.Unlock()

			if time.Since(lastHeartbeat) > 15*time.Second {
				logger.Warn("Leader appears to be down. Starting leader election...")
				startLeaderElection(clusterServer)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

// Starts leader election process
func startLeaderElection(clusterServer *cluster.GrpcServer) {
	logger := utils.NewLogger("", true)

	// Assume self as leader if no higher nodes exist
	clusterServer.VoteLock.Lock()
	newLeader := clusterServer.NodeID
	for nodeID := range clusterServer.Heartbeats {
		if nodeID > int(newLeader) {
			newLeader = int32(nodeID)
		}
	}
	clusterServer.VoteLock.Unlock()

	if newLeader == clusterServer.NodeID {
		logger.Info(fmt.Sprintf("Node %d is now the new leader", clusterServer.NodeID))
		clusterServer.LeaderID = int(clusterServer.NodeID)
	} else {
		logger.Info(fmt.Sprintf("Waiting for node %d to become leader", newLeader))
		clusterServer.LeaderID = int(newLeader)
	}
}
