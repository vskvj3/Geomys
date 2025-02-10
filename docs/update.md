### Updated client.go

```go
package cluster

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto" // Import generated gRPC code

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	Conn          *grpc.ClientConn
	Client        pb.NodeServiceClient
	LeaderID      int
	KnownNodes    map[int]string // Stores node ID to address mapping
	LastHeartbeat time.Time
}

// Connect to a node's gRPC server
func NewGrpcClient(address string) (*GrpcClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	client := pb.NewNodeServiceClient(conn)
	return &GrpcClient{Conn: conn, Client: client, KnownNodes: make(map[int]string)}, nil
}

// Request vote from another node
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

// Send heartbeat to leader
func (c *GrpcClient) SendHeartbeat(nodeID int) bool {
	if c.LeaderID == -1 {
		log.Println("No leader known. Skipping heartbeat.")
		return false
	}

	leaderAddr, exists := c.KnownNodes[c.LeaderID]
	if !exists {
		log.Println("Leader address unknown. Skipping heartbeat.")
		return false
	}

	conn, err := grpc.Dial(leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to leader %d at %s: %v", c.LeaderID, leaderAddr, err)
		return false
	}
	defer conn.Close()

	serviceClient := pb.NewNodeServiceClient(conn)
	_, err = serviceClient.Heartbeat(context.Background(), &pb.HeartbeatRequest{
		NodeId: int32(nodeID),
	})

	if err != nil {
		log.Printf("Failed to send heartbeat to leader %d: %v", c.LeaderID, err)
		return false
	}

	c.LastHeartbeat = time.Now()
	return true
}

func (c *GrpcClient) MonitorLeader(server *GrpcServer) {
	for {
		if !c.SendHeartbeat(int(server.NodeID)) {
			log.Println("Failed to send heartbeat. Checking leader status...")
			if time.Since(c.LastHeartbeat) > 15*time.Second {
				log.Println("Leader appears to be down. Starting leader election...")
				server.StartLeaderElection()
			}
		}
		time.Sleep(5 * time.Second)
	}
}
```

---

### Updated server.go

```go
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
func (s *GrpcServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	s.Heartbeats[int(req.NodeId)] = time.Now()

	return &pb.HeartbeatResponse{Success: true}, nil
}

// Notify all nodes about new leader
func (s *GrpcServer) InformFollowers() {
	for _, node := range s.Nodes {
		if int(node.NodeId) == s.LeaderID {
			continue // Skip leader itself
		}
		conn, err := grpc.Dial(node.Address, grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to follower %d at %s: %v", node.NodeId, node.Address, err)
			continue
		}
		defer conn.Close()

		client := pb.NewNodeServiceClient(conn)
		_, err = client.UpdateLeader(context.Background(), &pb.LeaderUpdateRequest{
			LeaderId: int32(s.LeaderID),
		})
		if err != nil {
			log.Printf("Failed to update follower %d: %v", node.NodeId, err)
		}
	}
}

// Start leader election and notify followers
func (s *GrpcServer) StartLeaderElection() {
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	newLeader := s.NodeID
	for _, node := range s.Nodes {
		if int(node.NodeId) < int(newLeader) {
			newLeader = node.NodeId
		}
	}

	if newLeader == s.NodeID {
		fmt.Printf("Node %d is now the leader\n", s.NodeID)
		s.LeaderID = int(s.NodeID)
		go s.InformFollowers()
	} else {
		fmt.Printf("Node %d elected as the leader\n", newLeader)
		s.LeaderID = int(newLeader)
	}
}
```
