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
	Nodes         map[int32]string
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
func (c *GrpcClient) RequestVote(nodeID int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.Client.RequestVote(ctx, &pb.VoteRequest{NodeId: int32(nodeID)})
	if err != nil {
		log.Printf("Vote request failed: %v", err)
		return false
	}
	return resp.Granted
}

// SendHeartbeat sends a heartbeat to the current leader
func (c *GrpcClient) SendHeartbeat(nodeID int, port int32) bool {
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
		if _, exists := c.Nodes[newNodeID]; !exists {
			logger.Info(fmt.Sprintf("New node detected: Node %d at %s", newNodeID, addr))
		}
	}

	// Detect nodes leaving
	for existingNodeID := range c.Nodes {
		if _, stillAlive := resp.Nodes[existingNodeID]; !stillAlive {
			logger.Warn(fmt.Sprintf("Node %d has left the cluster", existingNodeID))
		}
	}

	// Update heartbeat and cluster state
	c.LastHeartbeat = time.Now()
	c.Nodes = resp.Nodes
	return resp.Success
}

// MonitorLeader continuously sends heartbeats and checks for leader failure
func (c *GrpcClient) MonitorLeader(server *GrpcServer) {
	logger := utils.GetLogger()
	for {
		success := c.SendHeartbeat(int(server.NodeID), server.Port)
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
	newLeader := server.NodeID
	logger.Info("Starting leader election...")

	// Identify the new leader (lowest node ID)
	for nodeID := range c.Nodes {
		if nodeID < newLeader {
			newLeader = nodeID
		}
	}

	// If this node is the new leader, assume leadership
	if newLeader == server.NodeID {
		logger.Info(fmt.Sprintf("I am the new Leader! (Leader ID: %d)", server.NodeID))
		c.LeaderID = int(server.NodeID)
		go server.MonitorFollowers()
		return
	}

	// Otherwise, connect to the new leader
	logger.Info(fmt.Sprintf("Node %d is elected as the new leader", newLeader))
	newLeaderAddr, exists := c.Nodes[newLeader]
	if !exists {
		logger.Error("New leader's address is unknown. Unable to connect.")
		return
	}

	// Close previous connection if necessary
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
	c.LeaderID = int(newLeader)
}
