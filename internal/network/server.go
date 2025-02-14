package network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/vskvj3/geomys/internal/cluster"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/replicate"
	"github.com/vskvj3/geomys/internal/utils"
)

type Server struct {
	CommandHandler *core.CommandHandler
	grpcServer     *cluster.GrpcServer
	Port           string
}

func NewServer(grpcServer *cluster.GrpcServer, port string, handler *core.CommandHandler) (*Server, error) {
	logger := utils.GetLogger()

	// Load configuration
	config, err := utils.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	var leaderAddr string
	if grpcServer != nil {
		leaderAddr = grpcServer.LeaderAddress
	}

	if handler.Database == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	// Rebuild from persistence if standalone mode, else sync from leader
	if !config.IsLeader && leaderAddr != "" {
		replicationClient, err := replicate.NewReplicationClient(leaderAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create replication client: %v", err)
		}

		logger.Info("Re-syncing from leader at " + leaderAddr)
		if err := replicationClient.SyncRequest(handler); err != nil {
			return nil, fmt.Errorf("sync request failed: %v", err)
		}
	} else {
		if err := handler.Database.RebuildFromPersistence(); err != nil {
			logger.Warn("Could not read from persistence: " + err.Error())
		} else {
			logger.Info("Loaded data from persistence")
		}
	}

	handler.Database.StartCleanup(100 * time.Millisecond)
	logger.Info("TCP server initialized on port " + port)

	return &Server{CommandHandler: handler, grpcServer: grpcServer, Port: port}, nil
}

// Start the TCP server and listen for client connections
func (s *Server) Start() {
	logger := utils.GetLogger()

	// Attempt to bind to the configured port
	listener, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		logger.Warn("Port " + s.Port + " unavailable. Selecting a random port...")
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			logger.Error("Error starting server: " + err.Error())
			return
		}
	}
	defer listener.Close()
	logger.Info("Server is listening on " + listener.Addr().String())

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection: " + err.Error())
			continue
		}
		logger.Info("Accepted client: " + conn.RemoteAddr().String())
		go s.HandleConnection(conn)
	}
}

// Handle an incoming client connection
func (s *Server) HandleConnection(conn net.Conn) {
	defer func() {
		logger := utils.GetLogger()
		logger.Info("Client disconnected: " + conn.RemoteAddr().String())
		conn.Close()
	}()

	logger := utils.GetLogger()
	config, err := utils.GetConfig()
	if err != nil {
		logger.Error("Failed to load configuration: " + err.Error())
		return
	}

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("Client closed the connection: " + conn.RemoteAddr().String())
			} else {
				logger.Error("Error reading from client: " + err.Error())
			}
			return
		}

		request, err := utils.DecodeRequest(buffer[:n])

		if err != nil {
			logger.Error("Failed to decode request: " + err.Error())
			continue
		}

		logger.Debug("Received request from client: " + conn.RemoteAddr().String())

		command, err := utils.ConvertRequestToCommand(request)
		if err != nil {
			logger.Error("Request to command conversion failed")
		}

		var replicationClient *replicate.ReplicationClient

		// If not the leader and command is a write, forward it to the leader
		if !config.IsLeader && s.grpcServer != nil && isWriteCommand(command.Command) {
			logger.Info("Forwarding write request to leader node: " + s.grpcServer.LeaderAddress)

			replicationClient, err = replicate.NewReplicationClient(s.grpcServer.LeaderAddress)
			if err != nil {
				logger.Error("Replication client creation failed: " + err.Error())
				s.sendError(conn, "Failed to connect to leader")
				continue
			}

			response, err := replicationClient.ForwardRequest(int32(config.NodeID), command)
			if err != nil {
				logger.Error("Forward request failed: " + err.Error())
				s.sendError(conn, "Failed to forward request to leader")
				continue
			}

			logger.Debug("Got response for forward request: ")
			fmt.Println(response)
			s.sendResponse(conn, map[string]interface{}{"message": response.Message})
			continue
		}

		// Process command normally on the leader
		response, err := s.CommandHandler.HandleCommand(request)
		if err != nil {
			s.sendError(conn, err.Error())
		} else {
			s.sendResponse(conn, response)
			if config.IsLeader && s.grpcServer.LeaderID == int(s.grpcServer.NodeID) {
				replicate.ReplicateToFollowers(command, s.grpcServer)
			}

		}
	}
}

func isWriteCommand(command string) bool {
	writeCommands := map[string]bool{
		"SET":  true,
		"INCR": true,
		"PUSH": true,
		"LPOP": true,
		"RPOP": true,
	}
	return writeCommands[command]
}

// sendResponse serializes the response and sends it to the client
func (s *Server) sendResponse(conn net.Conn, response map[string]interface{}) {
	logger := utils.GetLogger()
	data, err := utils.EncodeResponse(response)
	if err != nil {
		logger.Error("Failed to encode response: " + err.Error())
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		logger.Error("Failed to send response: " + err.Error())
	}
}

// sendError sends an error message to the client
func (s *Server) sendError(conn net.Conn, errorMessage string) {
	response := map[string]interface{}{"status": "ERROR", "message": errorMessage}
	s.sendResponse(conn, response)
}
