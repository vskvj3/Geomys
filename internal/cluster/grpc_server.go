package cluster

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto" // Import generated gRPC code
	pbr "github.com/vskvj3/geomys/internal/replicate/proto"

	"github.com/vskvj3/geomys/internal/replicate"
	"github.com/vskvj3/geomys/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type GrpcServer struct {
	pb.UnimplementedNodeServiceServer
	NodeID        int32
	LeaderID      int
	LeaderAddress string
	Port          int32 // Server's port
	VoteLock      sync.Mutex
	Heartbeats    map[int]time.Time
	CurrentTerm   int32
	Nodes         map[int32]string // Known nodes in the cluster
}

// NewGrpcServer initializes a new gRPC server node
func NewGrpcServer(nodeID int32, port int32) *GrpcServer {
	return &GrpcServer{
		NodeID:      nodeID,
		LeaderID:    -1, // No leader initially
		Port:        port,
		Heartbeats:  make(map[int]time.Time),
		Nodes:       make(map[int32]string),
		CurrentTerm: -1,
	}
}

// RequestVote handles leader election requests
func (s *GrpcServer) RequestVote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	logger := utils.GetLogger()
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	logger.Info(fmt.Sprintf("Node %d received vote request from Node %d", s.NodeID, req.NodeId))

	// Check if there's an active leader
	if s.LeaderID > 0 {
		leaderAddr, exists := s.Nodes[int32(s.LeaderID)]
		if exists {
			client, err := NewGrpcClient(leaderAddr)
			if err == nil {
				defer client.Conn.Close()
				if client.SendHeartbeat(s) {
					// Leader is still alive, return it
					logger.Info(fmt.Sprintf("Returning current leader %d", s.LeaderID))
					return &pb.VoteResponse{SmallestNode: int32(s.LeaderID)}, nil
				}
			}
		}
	}

	// Find the smallest node ID (including self)
	smallestNode := s.NodeID
	for nodeID := range s.Nodes {
		if nodeID < smallestNode {
			smallestNode = nodeID
		}
	}

	logger.Info(fmt.Sprintf("No active leader found. Returning smallest node %d", smallestNode))
	return &pb.VoteResponse{SmallestNode: smallestNode}, nil
}

// Heartbeat handles heartbeat messages from followers
func (s *GrpcServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	logger := utils.GetLogger()
	peerInfo, ok := peer.FromContext(ctx)

	var addr string
	if ok {
		clientAddr := peerInfo.Addr.String()
		addr = fmt.Sprintf("%s:%d", strings.Split(clientAddr, ":")[0], req.Port)
	} else {
		logger.Error("Failed to retrieve peer address")
		return nil, fmt.Errorf("unable to determine peer address")
	}

	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	// Detect new nodes joining
	isNewNode := false
	if _, exists := s.Nodes[req.NodeId]; !exists {
		isNewNode = true
	}

	// Update heartbeat and node list
	s.Heartbeats[int(req.NodeId)] = time.Now()
	s.Nodes[req.NodeId] = addr

	if isNewNode {
		logger.Info(fmt.Sprintf("New node joined: Node %d at %s", req.NodeId, addr))
	}

	return &pb.HeartbeatResponse{Success: true, Nodes: s.Nodes}, nil
}

// StartServer initializes and starts the gRPC server
func (s *GrpcServer) StartServer(port int, repliServer *replicate.ReplicationServer) {
	logger := utils.GetLogger()

	// Start listening on the given port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to start server: " + err.Error())
		return
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register election (NodeService) and replication (ReplicationService) services
	pb.RegisterNodeServiceServer(grpcServer, s)
	pbr.RegisterReplicationServiceServer(grpcServer, repliServer)

	logger.Info(fmt.Sprintf("Node %d started gRPC server on port %d", s.NodeID, port))

	// Start the server in a goroutine so it can be stopped gracefully
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server stopped: " + err.Error())
	}
}

// MonitorFollowers periodically cleans up inactive nodes
func (s *GrpcServer) MonitorFollowers() {
	for {
		time.Sleep(5 * time.Second)
		s.cleanupInactiveNodes()
	}
}

// cleanupInactiveNodes removes nodes that have not sent heartbeats in the last 15 seconds
func (s *GrpcServer) cleanupInactiveNodes() {
	logger := utils.GetLogger()
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	for nodeID, lastHeartbeat := range s.Heartbeats {
		if time.Since(lastHeartbeat) > 15*time.Second {
			logger.Info(fmt.Sprintf("Node %d seems to be down. Removing from cluster", nodeID))
			delete(s.Heartbeats, nodeID)
			delete(s.Nodes, int32(nodeID))
		}
	}
}

// in cluster/grpc_server.go
func (s *GrpcServer) GetFollowerNodes() map[int32]string {
	return s.Nodes
}
