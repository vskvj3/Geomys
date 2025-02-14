package replication

import (
	"context"
	"fmt"

	"github.com/vskvj3/geomys/internal/cluster/proto"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/persistence"
	"github.com/vskvj3/geomys/internal/utils"
)

// interface to implement clusterServer functions
type ClusterInterface interface {
	GetNodeID() int32
	GetLeaderID() int32
	SetLeaderID(id int32)
	GetLeaderAddress() string
	SetLeaderAddress(addr string)
	GetNodes() map[int32]string
	GetPort() int32
	AddNode(nodeID int32, addr string)
	SetNodes(map[int32]string)
}

// ReplicationServer handles leader-follower replication
type ReplicationServer struct {
	proto.UnimplementedReplicationServiceServer
	Cluster        ClusterInterface
	CommandHandler *core.CommandHandler
}

func NewReplicationServer(server ClusterInterface, handler *core.CommandHandler) *ReplicationServer {
	return &ReplicationServer{
		CommandHandler: handler,
		Cluster:        server,
	}
}
func (s *ReplicationServer) ForwardRequest(ctx context.Context, command *proto.CommandRequest) (*proto.CommandResponse, error) {
	logger := utils.GetLogger()
	logger.Debug("Recieved forward request")
	fmt.Println(command)
	requestMap := utils.ConvertCommandToRequest(command.Command)

	// copy the request into database
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

	// replicate the request to followers
	ReplicateToFollowers(command.Command, s)

	// Return the final response
	return &protoResponse, nil
}

// ReplicateRequest is called by the leader to sync a command to followers
func (s *ReplicationServer) ReplicateRequest(ctx context.Context, command *proto.Command) (*proto.ReplicationAck, error) {

	// Convert gRPC request into a map
	requestMap := utils.ConvertCommandToRequest(command)

	// Execute the command using commandHandler
	_, err := s.CommandHandler.HandleCommand(requestMap)
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
