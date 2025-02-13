package replicate

import (
	"context"

	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/persistence"
	"github.com/vskvj3/geomys/internal/replicate/proto"
	"github.com/vskvj3/geomys/internal/utils"
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

func (s *ReplicationServer) ForwardRequest(ctx context.Context, command *proto.CommandRequest) (*proto.CommandResponse, error) {
	requestMap := utils.ConvertCommandToRequest(command.Command)

	response, err := s.CommandHandler.HandleCommand(requestMap)
	if err != nil {
		return nil, err
	}

	// Extract message/value from response
	protoResponse := proto.CommandResponse{Status: response["status"].(string)}

	if msg, ok := response["message"].(string); ok {
		protoResponse.Message = msg
	}
	if val, ok := response["value"].(string); ok {
		protoResponse.Value = val
	}

	// Return the final response
	return &protoResponse, nil
}

// ReplicateRequest is called by the leader to sync a command to followers
func (s *ReplicationServer) ReplicateRequest(ctx context.Context, command *proto.Command) (*proto.ReplicationAck, error) {
	test := make(map[string]interface{}, 0)
	_, err := s.CommandHandler.HandleCommand(test)
	if err != nil {
		return nil, err
	}
	// Convert gRPC request into a map
	requestMap := utils.ConvertCommandToRequest(command)

	// Execute the command using commandHandler
	_, err = s.CommandHandler.HandleCommand(requestMap)
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
		command, err := utils.ConvertRequestToCommand(req)
		if err != nil {
			utils.GetLogger().Error("Reuest conversion failed: " + err.Error())
		}

		commands = append(commands, command)
	}

	// Return response
	return &proto.SyncResponse{Commands: commands}, nil
}
