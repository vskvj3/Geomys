package cluster

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto" // Import generated gRPC code

	"google.golang.org/grpc"
)

type GrpcClient struct {
	Conn   *grpc.ClientConn
	Client pb.NodeServiceClient
}

// Connect to a node's gRPC server
func NewElectionClient(address string) (*GrpcClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewNodeServiceClient(conn)
	return &GrpcClient{Conn: conn, Client: client}, nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.Client.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: int32(nodeID)})
	if err != nil {
		log.Printf("Heartbeat failed: %v", err)
		return false
	}
	fmt.Println("grpc_client.go: sendHeartBeat", resp.Nodes)
	return resp.Success
}
