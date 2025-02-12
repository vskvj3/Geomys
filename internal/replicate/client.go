package replicate

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/replicate/proto"

	"google.golang.org/grpc"
)

// ReplicationClient is used by followers to communicate with the leader
type ReplicationClient struct {
	client proto.ReplicationServiceClient
}

// NewReplicationClient initializes a gRPC client connection
func NewReplicationClient(leaderAddress string) (*ReplicationClient, error) {
	conn, err := grpc.Dial(leaderAddress, grpc.WithInsecure(), grpc.WithBlock())
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

func (*ReplicationClient) ReplicateRequest(req *proto.CommandRequest) error {
	fmt.Println("I dont know what this supposed to do")
	return nil
}
