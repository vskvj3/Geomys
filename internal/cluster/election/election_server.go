package election

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto"
	"github.com/vskvj3/geomys/internal/utils"
	"google.golang.org/grpc/peer"
)

// Interface to define clusterServer functions
type ClusterInterface interface {
	GetNodeID() int32
	GetLeaderID() int32
	SetLeaderID(id int32)
	GetLeaderAddress() string
	SetLeaderAddress(addr string)
	GetNodes() map[int32]string
	GetPort() int32
	AddNode(nodeID int32, addr string)
	RemodeNode(nodeID int32)
	SetNodes(map[int32]string)
}

// ElectionServer struct
type ElectionServer struct {
	pb.UnimplementedElectionServiceServer
	Cluster     ClusterInterface
	VoteLock    sync.Mutex
	Heartbeats  map[int32]time.Time
	CurrentTerm int32
}

// NewElectionServer initializes a new ElectionServer with a ClusterInterface
func NewElectionServer(cluster ClusterInterface) *ElectionServer {
	return &ElectionServer{
		Cluster:     cluster,
		Heartbeats:  make(map[int32]time.Time),
		CurrentTerm: -1,
	}
}

// RequestVote handles leader election requests
func (s *ElectionServer) RequestVote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	logger := utils.GetLogger()
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	logger.Info(fmt.Sprintf("Node %d received vote request from Node %d", s.Cluster.GetNodeID(), req.NodeId))

	// Check if there's an active leader
	if s.Cluster.GetLeaderID() > 0 {
		leaderAddr := s.Cluster.GetLeaderAddress()
		if leaderAddr != "" {
			client, err := NewElectionClient(leaderAddr)
			if err == nil {
				defer client.Conn.Close()
				if client.SendHeartbeat(s) {
					logger.Info(fmt.Sprintf("Returning current leader %d", s.Cluster.GetLeaderID()))
					return &pb.VoteResponse{SmallestNode: s.Cluster.GetLeaderID()}, nil
				}
			}
		}
	}

	// Find the smallest node ID (including self)
	smallestNode := s.Cluster.GetNodeID()
	for nodeID := range s.Cluster.GetNodes() {
		if nodeID < smallestNode {
			smallestNode = nodeID
		}
	}

	logger.Info(fmt.Sprintf("No active leader found. Returning smallest node %d", smallestNode))
	return &pb.VoteResponse{SmallestNode: smallestNode}, nil
}

// Heartbeat handles heartbeat messages from followers
func (s *ElectionServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
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
	if _, exists := s.Cluster.GetNodes()[req.NodeId]; !exists {
		isNewNode = true
	}

	// Update heartbeat and node list
	s.Heartbeats[req.NodeId] = time.Now()
	s.Cluster.AddNode(req.NodeId, addr)

	if isNewNode {
		logger.Info(fmt.Sprintf("New node joined: Node %d at %s", req.NodeId, addr))
	}

	return &pb.HeartbeatResponse{Success: true, Nodes: s.Cluster.GetNodes()}, nil
}

// MonitorFollowers periodically cleans up inactive nodes
func (s *ElectionServer) MonitorFollowers() {
	for {
		time.Sleep(5 * time.Second)
		s.cleanupInactiveNodes()
	}
}

// cleanupInactiveNodes removes nodes that have not sent heartbeats in the last 15 seconds
func (s *ElectionServer) cleanupInactiveNodes() {
	logger := utils.GetLogger()
	s.VoteLock.Lock()
	defer s.VoteLock.Unlock()

	for nodeID, lastHeartbeat := range s.Heartbeats {
		if time.Since(lastHeartbeat) > 15*time.Second {
			logger.Info(fmt.Sprintf("Node %d seems to be down. Removing from cluster", nodeID))
			delete(s.Heartbeats, nodeID)
			s.Cluster.RemodeNode(nodeID)
		}
	}
}
