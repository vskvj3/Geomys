package cluster

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto" // Import generated gRPC code

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type GrpcServer struct {
	pb.UnimplementedNodeServiceServer
	NodeID     int32
	LeaderID   int
	VoteLock   sync.Mutex
	Heartbeats map[int]time.Time
	Nodes      []*pb.Node // List of known nodes
}

func NewGrpcServer(nodeID int32) *GrpcServer {
	return &GrpcServer{
		NodeID:     nodeID,
		LeaderID:   -1, // No leader initially
		Heartbeats: make(map[int]time.Time),
		Nodes:      []*pb.Node{},
	}
}

// Handle leader election votes
func (s *GrpcServer) RequestVote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	fmt.Printf("Node %d received vote request from Node %d\n", s.NodeID, req.NodeId)

	if req.NodeId > s.NodeID {
		s.LeaderID = int(req.NodeId)
		return &pb.VoteResponse{Granted: true}, nil
	}
	return &pb.VoteResponse{Granted: false}, nil
}

// Handle incoming heartbeat updates
// Handle incoming heartbeat updates
func (s *GrpcServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	fmt.Println("grpc_server: heartbeat request: ", req)

	// Get the request address (IP:Port)
	peerInfo, ok := peer.FromContext(ctx)
	var addr string
	if ok {
		addr = peerInfo.Addr.String()
		fmt.Printf("Received heartbeat from %s\n", addr)
	} else {
		fmt.Println("Failed to get peer address")
	}

	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	s.Heartbeats[int(req.NodeId)] = time.Now()

	// Check if the node is already in the list and update address if needed
	exists := false
	for _, node := range s.Nodes {
		if node.NodeId == req.NodeId {
			node.Address = addr // Update address if the node already exists
			exists = true
			break
		}
	}
	if !exists {
		s.Nodes = append(s.Nodes, &pb.Node{NodeId: req.NodeId, Address: addr})
	}

	return &pb.HeartbeatResponse{Success: true, Nodes: s.Nodes}, nil
}

// Start the gRPC server
func (s *GrpcServer) StartServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNodeServiceServer(grpcServer, s)

	fmt.Printf("Node %d started gRPC server on port %d\n", s.NodeID, port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Start monitoring leader status and trigger election if needed
func (e *GrpcServer) MonitorFollowers() {
	// for evry 15 seconds, clean up the nodes list and remove inactive nodes
	for {
		time.Sleep(15 * time.Second)
		e.cleanupInactiveNodes()
	}
}

// Cleanup nodes that have not sent heartbeats within 15 seconds
func (e *GrpcServer) cleanupInactiveNodes() {
	e.VoteLock.Lock()
	defer e.VoteLock.Unlock()

	activeNodes := []*pb.Node{}
	for _, node := range e.Nodes {
		if lastHeartbeat, exists := e.Heartbeats[int(node.NodeId)]; exists && time.Since(lastHeartbeat) <= 15*time.Second {
			activeNodes = append(activeNodes, node)
		} else {
			fmt.Printf("grpc_server.go: Removing inactive node %d\n", node.NodeId)
			delete(e.Heartbeats, int(node.NodeId))
		}
	}
	e.Nodes = activeNodes
}

// Initiate leader election
func (e *GrpcServer) StartLeaderElection() {
	e.VoteLock.Lock()
	defer e.VoteLock.Unlock()

	// Determine new leader (lowest ID)
	newLeader := e.NodeID
	for _, node := range e.Nodes {
		if int(node.NodeId) < int(newLeader) {
			newLeader = int32(node.NodeId)
		}
	}

	// If this node is the new leader, announce it
	if newLeader == e.NodeID {
		fmt.Printf("Inside grpc server: Node %d is now the new leader\n", e.NodeID)
		e.LeaderID = int(e.NodeID)
	} else {
		fmt.Printf("Inside grpc server: Node %d is elected as the new leader\n", newLeader)
		e.LeaderID = int(newLeader)
	}
}
