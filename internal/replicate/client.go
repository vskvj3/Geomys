package replicate

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/replicate/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ReplicationClient is used by followers to communicate with the leader
type ReplicationClient struct {
	client proto.ReplicationServiceClient
}

// NewReplicationClient initializes a gRPC client connection
func NewReplicationClient(leaderAddress string) (*ReplicationClient, error) {
	conn, err := grpc.NewClient(leaderAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	return &ReplicationClient{
		client: proto.NewReplicationServiceClient(conn),
	}, nil
}

// Forward a write request from follower to leader
func (c *ReplicationClient) ForwardRequest(node_id int32, command *proto.Command) (*proto.CommandResponse, error) {
	req := &proto.CommandRequest{
		NodeId:  node_id,
		Command: command,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.ForwardRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *ReplicationClient) ReplicateRequest(command *proto.Command) (*proto.ReplicationAck, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.ReplicateRequest(ctx, command)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Helper function to replicate writes to all followers
func (c *ReplicationClient) ReplicateToFollowers(command *proto.Command, cluster ClusterNodeProvider) error {
	fmt.Println(command)
	for _, followerAddr := range cluster.GetFollowerNodes() {
		client, err := NewReplicationClient(followerAddr)
		if err != nil {
			log.Printf("Error connecting to follower %s: %v", followerAddr, err)
			continue
		}

		_, err = client.ReplicateRequest(command)
		if err != nil {
			log.Printf("Error replicating to follower %s: %v", followerAddr, err)
		} else {
			log.Printf("success")
		}
	}
	return nil
}

// SyncRequest is called when a follower restarts to get the latest data
func (c *ReplicationClient) SyncRequest(commandHandler *core.CommandHandler) error {
	req := &proto.SyncRequestMessage{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.SyncRequest(ctx, req)
	if err != nil {
		return err
	}

	// Process each received command
	for _, command := range resp.Commands {
		// Convert received gRPC Command into a map
		cmdMap := map[string]interface{}{
			"command": command.Command,
			"key":     command.Key,
			"value":   command.Value,
			"offset":  int(command.Offset), // Convert int32 to int
		}

		// Execute command in database
		_, err := commandHandler.HandleCommand(cmdMap)
		if err != nil {
			log.Printf("Error applying command during sync: %v", err)
		} else {
			log.Printf("Successfully applied command during sync: %+v", cmdMap)
		}
	}

	return nil
}
