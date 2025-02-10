package cluster

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto" // Import generated gRPC code
	"github.com/vskvj3/geomys/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	Conn          *grpc.ClientConn
	Client        pb.NodeServiceClient
	LeaderID      int
	LastHeartbeat time.Time
}

// NewGrpcClient establishes a connection with a node's gRPC server
func NewGrpcClient(address string) (*GrpcClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	client := pb.NewNodeServiceClient(conn)
	return &GrpcClient{Conn: conn, Client: client}, nil
}

// RequestVote sends a leader election vote request to another node
func (c *GrpcClient) RequestVote(nodeID int, term int32) int32 {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.Client.RequestVote(ctx, &pb.VoteRequest{NodeId: int32(nodeID), Term: term})
	if err != nil {
		log.Printf("Vote request failed from Node %d: %v", nodeID, err)
		return -1 // Return -1 to indicate failure
	}
	return resp.SmallestNode // Return the smallest node ID received
}

// SendHeartbeat sends a heartbeat to the current leader
func (c *GrpcClient) SendHeartbeat(server *GrpcServer) bool {
	nodeID := server.NodeID
	port := server.Port
	logger := utils.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.Client.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: int32(nodeID), Port: port})
	if err != nil {
		logger.Error("Heartbeat failed: " + err.Error())
		return false
	}

	// Detect new nodes joining
	for newNodeID, addr := range resp.Nodes {
		if _, exists := server.Nodes[newNodeID]; !exists {
			logger.Info(fmt.Sprintf("New node detected: Node %d at %s", newNodeID, addr))
		}
	}

	// Detect nodes leaving
	for existingNodeID := range server.Nodes {
		if _, stillAlive := resp.Nodes[existingNodeID]; !stillAlive {
			logger.Warn(fmt.Sprintf("Node %d has left the cluster", existingNodeID))
		}
	}

	// Update heartbeat and cluster state
	c.LastHeartbeat = time.Now()
	server.Nodes = resp.Nodes
	return resp.Success
}

// MonitorLeader continuously sends heartbeats and checks for leader failure
func (c *GrpcClient) MonitorLeader(server *GrpcServer) {
	logger := utils.GetLogger()
	for {
		success := c.SendHeartbeat(server)
		if !success {
			logger.Error("Failed to send heartbeat. Checking leader status...")

			// If last heartbeat is over 15 seconds old, assume leader is down
			if time.Since(c.LastHeartbeat) > 15*time.Second {
				logger.Warn("Leader appears to be down. Initiating leader election...")
				c.StartLeaderElection(server)

				// If this node becomes the leader, stop monitoring and start managing followers
				if c.LeaderID == int(server.NodeID) {
					return
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

// StartLeaderElection initiates the leader election process
func (c *GrpcClient) StartLeaderElection(server *GrpcServer) {
	logger := utils.GetLogger()
	logger.Info("Starting leader election...")

	// Increment election term
	server.CurrentTerm++

	// Collect votes from all nodes
	smallestNode := server.NodeID
	votes := make(map[int32]bool) // Track votes to check agreement

	for nodeID, addr := range server.Nodes {
		if nodeID == server.NodeID {
			continue // Skip self
		}

		client, err := NewGrpcClient(addr)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to connect to Node %d for vote request", nodeID))
			continue
		}

		// Request vote and get the smallest node ID
		votedNode := client.RequestVote(int(server.NodeID), server.CurrentTerm)
		if votedNode != -1 {
			votes[int32(votedNode)] = true // Store vote
			if votedNode < smallestNode {
				smallestNode = votedNode
			}
		}
		client.Conn.Close()
	}

	fmt.Println(votes)
	fmt.Println(smallestNode)
	// Check if all nodes agree on the same smallest node
	for nodeID := range votes {
		if nodeID != int32(smallestNode) {
			logger.Warn("Nodes did not agree on the same leader. Restarting election...")
			time.Sleep(3 * time.Second)
			c.StartLeaderElection(server) // Retry election
			return
		}
	}

	// Confirm the selected leader is still alive
	if !c.ConfirmLeader(smallestNode, server) {
		logger.Warn(fmt.Sprintf("Leader %d is not responding. Restarting election...", smallestNode))
		delete(server.Nodes, int32(smallestNode))
		time.Sleep(3 * time.Second)
		c.StartLeaderElection(server)
		return
	}

	// Successfully elected leader
	logger.Info(fmt.Sprintf("Node %d is elected as the new leader!", smallestNode))
	c.LeaderID = int(smallestNode)

	if smallestNode == server.NodeID {
		logger.Info("I am the new leader! Starting to manage followers...")
		go server.MonitorFollowers()
	} else {
		// Connect to the new leader
		newLeaderAddr, exists := server.Nodes[int32(smallestNode)]
		if !exists {
			logger.Error("New leader's address is unknown....")
			return
		}

		// Close previous connection
		if c.Conn != nil {
			c.Conn.Close()
		}

		conn, err := grpc.NewClient(newLeaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to connect to new leader (%s): %v", newLeaderAddr, err))
			return
		}

		// Update client connection
		c.Conn = conn
		c.Client = pb.NewNodeServiceClient(conn)
	}
}

// ConfirmLeader checks if the selected leader is still alive
func (c *GrpcClient) ConfirmLeader(leaderID int32, server *GrpcServer) bool {
	logger := utils.GetLogger()
	logger.Info(fmt.Sprintf("Confirming if Node %d is still alive...", leaderID))

	leaderAddr, exists := server.Nodes[leaderID]
	if !exists {
		logger.Warn(fmt.Sprintf("Leader %d not found in nodes list.", leaderID))
		return false
	}

	client, err := NewGrpcClient(leaderAddr)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to connect to Leader %d.", leaderID))
		return false
	}
	defer client.Conn.Close()

	// Send a simple heartbeat to confirm leader's availability
	resp := client.SendHeartbeat(server)
	return resp
}
