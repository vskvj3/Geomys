package cluster

import (
	"fmt"
	"net"
	"sync"

	pb "github.com/vskvj3/geomys/internal/cluster/proto"

	"github.com/vskvj3/geomys/internal/cluster/election"
	"github.com/vskvj3/geomys/internal/cluster/replication"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/utils"
	"google.golang.org/grpc"
)

// ClusterServer manages election and replication services
type ClusterServer struct {
	NodeID        int32
	LeaderID      int32
	LeaderAddress string
	Port          int32
	Nodes         map[int32]string
	mu            sync.Mutex

	ElectionService    *election.ElectionServer
	ReplicationService *replication.ReplicationServer
}

// NewClusterServer initializes the cluster server with election and replication services
func NewClusterServer(nodeID int32, port int32) *ClusterServer {
	cluster := &ClusterServer{
		NodeID: nodeID,
		Port:   port,
		Nodes:  make(map[int32]string),
	}

	return cluster
}

// StartServer initializes and starts the gRPC server
func (s *ClusterServer) StartServer(handler *core.CommandHandler) {
	logger := utils.GetLogger()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		logger.Error("Failed to start server: " + err.Error())
		return
	}

	grpcServer := grpc.NewServer()

	// Assign election and replication services
	s.ElectionService = election.NewElectionServer(s)
	s.ReplicationService = replication.NewReplicationServer(s, handler)

	// Register services with gRPC
	pb.RegisterElectionServiceServer(grpcServer, s.ElectionService)
	pb.RegisterReplicationServiceServer(grpcServer, s.ReplicationService)

	logger.Info(fmt.Sprintf("Node %d started gRPC server on port %d", s.NodeID, s.Port))

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server stopped: " + err.Error())
	}
}

/***************************************************************
*                  ClusterInterface Methods                    *
***************************************************************/
// Get node id of current node
func (c *ClusterServer) GetNodeID() int32 {
	return c.NodeID
}

// Change nodes list
func (c *ClusterServer) GetPort() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Port
}

// Get node id of leader node
func (c *ClusterServer) GetLeaderID() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.LeaderID
}

// Change leader id
func (c *ClusterServer) SetLeaderID(id int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LeaderID = id
}

// Get leader address
func (c *ClusterServer) GetLeaderAddress() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.LeaderAddress
}

// Change leader Address
func (c *ClusterServer) SetLeaderAddress(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LeaderAddress = addr
}

// Get Nodes
func (c *ClusterServer) GetNodes() map[int32]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Nodes
}

// Add a new node into node list
func (c *ClusterServer) AddNode(nodeID int32, addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Nodes[nodeID] = addr
}

// Change nodes list
func (c *ClusterServer) SetNodes(newNodes map[int32]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Nodes = newNodes
}
