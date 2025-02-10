package cluster

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
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
	Port       int32 // port the server is running
	VoteLock   sync.Mutex
	Heartbeats map[int]time.Time
	Nodes      map[int32]string // List of known nodes
}

func NewGrpcServer(nodeID int32, port int32) *GrpcServer {
	return &GrpcServer{
		NodeID:     nodeID,
		LeaderID:   -1, // No leader initially
		Port:       port,
		Heartbeats: make(map[int]time.Time),
		Nodes:      make(map[int32]string),
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
		clientAddr := peerInfo.Addr.String()
		fmt.Println(strings.Split(clientAddr, `:`)[0], req.Port)
		addr = strings.Split(clientAddr, `:`)[0] + ":" + req.Port
		fmt.Printf("Received heartbeat from %s\n", addr)
	} else {
		fmt.Println("Failed to get peer address")
	}

	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	s.Heartbeats[int(req.NodeId)] = time.Now()

	// Check if the node is already in the list and update address if needed
	s.Nodes[req.NodeId] = addr
	fmt.Println(s.Nodes)

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

	for nodeID, lastHeartbeat := range e.Heartbeats {
		if time.Since(lastHeartbeat) > 15*time.Second {
			fmt.Printf("grpc_server.go: Removing inactive node %d\n", nodeID)
			delete(e.Heartbeats, nodeID)
			delete(e.Nodes, int32(nodeID))
		}
	}
}
