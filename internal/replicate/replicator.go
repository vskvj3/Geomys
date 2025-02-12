package replicate

import (
	"context"
	"log"
	"strconv"

	"github.com/vskvj3/geomys/internal/persistence"
	"github.com/vskvj3/geomys/internal/replicate/proto"
)

// ReplicationServer handles leader-follower replication
type ReplicationServer struct {
	proto.UnimplementedReplicationServiceServer
}

func NewReplicationServer() *ReplicationServer {
	return &ReplicationServer{}
}

// ForwardRequest is called by followers when they receive a write request
func (s *ReplicationServer) ForwardRequest(ctx context.Context, req *proto.CommandRequest) (*proto.CommandResponse, error) {

	// Execute the command locally
	// CommandHandler needs commands as a interface
	// test := make(map[string]interface{}, 0)
	// response, err := commandHandler.HandleCommand(test)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(response)

	// // Send ReplicateRequest to all followers
	// err = s.replicateToFollowers(req)
	// if err != nil {
	// 	return nil, err
	// }
	testString := "test"
	return &proto.CommandResponse{Message: testString}, nil
}

// ReplicateRequest is called by the leader to sync a command to followers
func (s *ReplicationServer) ReplicateRequest(ctx context.Context, req *proto.CommandRequest) (*proto.ReplicationAck, error) {
	// test := make(map[string]interface{}, 0)
	// _, err := commandHandler.HandleCommand(test)
	// if err != nil {
	// 	return nil, err
	// }

	return &proto.ReplicationAck{Success: true}, nil
}

// SyncRequest is called when a follower restarts and wants the latest data
func (s *ReplicationServer) SyncRequest(ctx context.Context, req *proto.SyncRequestMessage) (*proto.SyncResponse, error) {
	// Load requests from persistence
	p, err := persistence.CreateOrReplacePersistence()
	if err != nil {
		return nil, err
	}
	requests, err := p.LoadRequests()
	if err != nil {
		return nil, err
	}

	// Convert loaded requests to []*proto.Command
	var commands []*proto.Command
	for _, req := range requests {
		command := &proto.Command{
			Command: req["command"].(string),
			Key:     req["key"].(string),
		}

		// Add optional fields if they exist
		if val, ok := req["value"].(string); ok {
			command.Value = val
		}
		if offsetStr, ok := req["offset"].(string); ok {
			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				log.Printf("Failed to parse offset: %v", err)
				continue // Skip invalid offsets
			}
			command.Offset = int32(offset)
		}

		commands = append(commands, command)
	}

	// Return response
	return &proto.SyncResponse{Commands: commands}, nil
}

// Helper function to replicate writes to all followers
func (s *ReplicationServer) replicateToFollowers(req *proto.CommandRequest, cluster ClusterNodeProvider) error {
	for _, followerAddr := range cluster.GetFollowerNodes() {
		client, err := NewReplicationClient(followerAddr)
		if err != nil {
			log.Printf("Error connecting to follower %s: %v", followerAddr, err)
			continue
		}

		err = client.ReplicateRequest(req)
		if err != nil {
			log.Printf("Error replicating to follower %s: %v", followerAddr, err)
		}
	}
	return nil
}
