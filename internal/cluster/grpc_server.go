package cluster

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto" // Import generated gRPC code
	"github.com/vskvj3/geomys/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type GrpcServer struct {
	pb.UnimplementedNodeServiceServer
	NodeID     int32
	LeaderID   int
	Port       int32 // Server's port
	VoteLock   sync.Mutex
	Heartbeats map[int]time.Time
	Nodes      map[int32]string // Known nodes in the cluster
}

// NewGrpcServer initializes a new gRPC server node
func NewGrpcServer(nodeID int32, port int32) *GrpcServer {
	return &GrpcServer{
		NodeID:     nodeID,
		LeaderID:   -1, // No leader initially
		Port:       port,
		Heartbeats: make(map[int]time.Time),
		Nodes:      make(map[int32]string),
	}
}

// RequestVote handles leader election requests
func (s *GrpcServer) RequestVote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	logger := utils.GetLogger()
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	logger.Info(fmt.Sprintf("Node %d received vote request from Node %d", s.NodeID, req.NodeId))

	// Grant vote to higher node IDs
	if req.NodeId > s.NodeID {
		s.LeaderID = int(req.NodeId)
		return &pb.VoteResponse{Granted: true}, nil
	}
	return &pb.VoteResponse{Granted: false}, nil
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
func (s *GrpcServer) StartServer(port int) {
	logger := utils.GetLogger()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to start server: " + err.Error())
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNodeServiceServer(grpcServer, s)

	logger.Info(fmt.Sprintf("Node %d started gRPC server on port %d", s.NodeID, port))

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
