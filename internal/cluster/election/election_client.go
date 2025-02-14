package election

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto"
	"github.com/vskvj3/geomys/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ElectionClient facilitates leader election
type ElectionClient struct {
	Conn          *grpc.ClientConn
	Client        pb.ElectionServiceClient
	LastHeartbeat time.Time
}

// NewGrpcClient establishes a connection with a node's gRPC server
func NewElectionClient(address string) (*ElectionClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	client := pb.NewElectionServiceClient(conn)
	return &ElectionClient{Conn: conn, Client: client}, nil
}

// RequestVote sends a leader election vote request to another node
func (c *ElectionClient) RequestVote(nodeID int32, term int32) int32 {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	resp, err := c.Client.RequestVote(ctx, &pb.VoteRequest{NodeId: nodeID, Term: term})
	if err != nil {
		log.Printf("Vote request failed from Node %d: %v", nodeID, err)
		return -1 // Return -1 to indicate failure
	}
	return resp.SmallestNode // Return the smallest node ID received
}

// SendHeartbeat sends a heartbeat to the current leader
func (c *ElectionClient) SendHeartbeat(server *ElectionServer) bool {
	nodeID := server.Cluster.GetNodeID()
	port := server.Cluster.GetPort()
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
		if _, exists := server.Cluster.GetNodes()[newNodeID]; !exists {
			logger.Info(fmt.Sprintf("New node detected: Node %d at %s", newNodeID, addr))
		}
	}

	// Detect nodes leaving
	for existingNodeID := range server.Cluster.GetNodes() {
		if _, stillAlive := resp.Nodes[existingNodeID]; !stillAlive {
			logger.Warn(fmt.Sprintf("Node %d has left the cluster", existingNodeID))
		}
	}

	// Update heartbeat and cluster state
	c.LastHeartbeat = time.Now()
	server.Cluster.SetNodes(resp.Nodes)
	return resp.Success
}

// MonitorLeader continuously sends heartbeats and checks for leader failure
func (c *ElectionClient) MonitorLeader(server *ElectionServer) {
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
				if server.Cluster.GetLeaderID() == server.Cluster.GetNodeID() {
					return
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

// StartLeaderElection initiates the leader election process
func (c *ElectionClient) StartLeaderElection(server *ElectionServer) {
	logger := utils.GetLogger()
	config, err := utils.GetConfig()
	if err != nil {
		logger.Error("Failed to load configs for leader election")
	}
	logger.Info("Starting leader election...")
	config.IsLeader = false

	// Map to track votes from other nodes
	votes := make(map[int32]bool)
	var smallestNode int32 = -1

	// Request votes from all other nodes
	for nodeID, addr := range server.Cluster.GetNodes() {
		if nodeID == server.Cluster.GetNodeID() {
			continue // Skip self
		}

		client, err := NewElectionClient(addr)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to connect to Node %d for vote request", nodeID))
			continue
		}

		// Request vote
		votedNode := client.RequestVote(server.Cluster.GetNodeID(), server.CurrentTerm)
		client.Conn.Close()

		// Ignore invalid votes
		if votedNode == -1 {
			continue
		}

		// Store vote
		votes[int32(votedNode)] = true

		// Track smallest voted node
		if smallestNode == -1 || int32(votedNode) < smallestNode {
			smallestNode = int32(votedNode)
		}
	}

	// If no votes were received, abort election
	if len(votes) == 0 {
		logger.Warn("No votes received. Retrying election...")
		time.Sleep(3 * time.Second)
		c.StartLeaderElection(server)
		return
	}

	// Ensure all nodes agree on the same leader
	for nodeID := range votes {
		if nodeID != smallestNode {
			logger.Warn("Nodes did not agree on the same leader. Restarting election...")
			time.Sleep(3 * time.Second)
			c.StartLeaderElection(server)
			return
		}
	}

	// Confirm the selected leader is alive
	if !c.ConfirmLeader(smallestNode, server) {
		logger.Warn(fmt.Sprintf("Leader %d is not responding. Restarting election...", smallestNode))
		delete(server.Cluster.GetNodes(), smallestNode)
		time.Sleep(3 * time.Second)
		c.StartLeaderElection(server)
		return
	}

	// Successfully elected leader
	logger.Info(fmt.Sprintf("Node %d is elected as the new leader!", smallestNode))
	server.Cluster.SetLeaderID(smallestNode)
	server.Cluster.SetLeaderAddress(server.Cluster.GetNodes()[smallestNode])

	if smallestNode == server.Cluster.GetNodeID() {
		logger.Info("I am the new leader! Managing followers...")
		config.IsLeader = true

		// wait for 15 seconds, and then start monitoring followers
		time.Sleep(5 * time.Second)
		go server.MonitorFollowers()
	} else {
		// Connect to the new leader
		newLeaderAddr, exists := server.Cluster.GetNodes()[smallestNode]
		if !exists {
			logger.Error("New leader's address is unknown.")
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
		c.Client = pb.NewElectionServiceClient(conn)
	}
}

// ConfirmLeader checks if the selected leader is still alive
func (c *ElectionClient) ConfirmLeader(leaderID int32, server *ElectionServer) bool {
	logger := utils.GetLogger()
	logger.Info(fmt.Sprintf("Confirming if Node %d is still alive...", leaderID))

	leaderAddr, exists := server.Cluster.GetNodes()[leaderID]
	if !exists {
		logger.Warn(fmt.Sprintf("Leader %d not found in nodes list.", leaderID))
		return false
	}

	client, err := NewElectionClient(leaderAddr)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to connect to Leader %d.", leaderID))
		return false
	}
	defer client.Conn.Close()

	// Send a simple heartbeat to confirm leader's availability
	resp := client.SendHeartbeat(server)
	return resp
}
