package replicate

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/persistence"
	"github.com/vskvj3/geomys/internal/replicate/proto"
)

// ReplicationServer handles leader-follower replication
type ReplicationServer struct {
	proto.UnimplementedReplicationServiceServer
	CommandHandler *core.CommandHandler
}

func NewReplicationServer(handler *core.CommandHandler) *ReplicationServer {
	return &ReplicationServer{
		CommandHandler: handler,
	}
}

func (s *ReplicationServer) ForwardRequest(ctx context.Context, req *proto.CommandRequest) (*proto.CommandResponse, error) {
	// Convert gRPC request into a map
	commandData := map[string]interface{}{
		"command": strings.ToUpper(req.Command.Command),
	}
	if req.Command.Key != "" {
		commandData["key"] = req.Command.Key
	}
	if req.Command.Value != "" {
		commandData["value"] = req.Command.Value
	}
	if req.Command.Exp > 0 {
		commandData["exp"] = req.Command.Exp
	}
	if req.Command.Offset != 0 {
		commandData["offset"] = req.Command.Offset
	}

	// Execute the command using commandHandler
	response, err := s.CommandHandler.HandleCommand(commandData)
	if err != nil {
		return nil, err
	}

	// Extract message/value from response
	message := "OK"
	if msg, ok := response["message"].(string); ok {
		message = msg
	} else if val, ok := response["value"].(string); ok {
		message = val
	}

	// Return the final response
	return &proto.CommandResponse{Message: message}, nil
}

// ReplicateRequest is called by the leader to sync a command to followers
func (s *ReplicationServer) ReplicateRequest(ctx context.Context, command *proto.Command) (*proto.ReplicationAck, error) {
	test := make(map[string]interface{}, 0)
	_, err := s.CommandHandler.HandleCommand(test)
	if err != nil {
		return nil, err
	}
	// Convert gRPC request into a map
	commandData := map[string]interface{}{
		"command": strings.ToUpper(command.Command),
	}
	if command.Key != "" {
		commandData["key"] = command.Key
	}
	if command.Value != "" {
		commandData["value"] = command.Value
	}
	if command.Exp > 0 {
		commandData["exp"] = command.Exp
	}
	if command.Offset != 0 {
		commandData["offset"] = command.Offset
	}

	// Execute the command using commandHandler
	_, err = s.CommandHandler.HandleCommand(commandData)
	if err != nil {
		return nil, err
	}

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
