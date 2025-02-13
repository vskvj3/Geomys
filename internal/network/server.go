package network

import (
	"errors"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vskvj3/geomys/internal/cluster"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/replicate"
	"github.com/vskvj3/geomys/internal/replicate/proto"
	"github.com/vskvj3/geomys/internal/utils"
)

type Server struct {
	CommandHandler *core.CommandHandler
	grpcServer     *cluster.GrpcServer
	Port           string
}

func NewServer(grpcServer *cluster.GrpcServer, port string, handler *core.CommandHandler) (*Server, error) {
	logger := utils.GetLogger()
	config, err := utils.GetConfig()
	if err != nil {
		logger.Error("Loading config in server failed: " + err.Error())
	}
	var leaderAddr string
	if grpcServer != nil {
		leaderAddr = strconv.Itoa(grpcServer.LeaderID)
	} else {
		leaderAddr = ""
	}

	// Rebuild from persistence if standalone mode, else request sync from leader
	if !config.IsLeader && leaderAddr != "" {
		replicationClient, err := replicate.NewReplicationClient(leaderAddr)
		if err != nil {
			logger.Error("Replication client creation failed: " + err.Error())
		}
		logger.Info("Re-syncing from leader node...")
		replicationClient.SyncRequest(handler)
	} else {
		if err := handler.Database.RebuildFromPersistence(); err != nil {
			logger.Warn("Could not read from persistence: " + err.Error())
		} else {
			logger.Info("Loaded data from persistence")
		}
	}

	// Start database cleanup (to remove expired keys)
	handler.Database.StartCleanup(100 * time.Millisecond)

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

		// Deserialize request
		var request map[string]interface{}
		err = msgpack.Unmarshal(buffer[:n], &request)
		if err != nil {
			logger.Error("Failed to decode request: " + err.Error())
			continue
		}

		logger.Debug("Received request from client: " + conn.RemoteAddr().String())

		// Extract command type
		command, ok := request["command"].(string)
		if !ok {
			s.sendError(conn, "Invalid command format")
			continue
		}

		// If not the leader and command is a write, forward it to the leader
		if !config.IsLeader && s.grpcServer != nil && s.grpcServer.LeaderID != -1 && isWriteCommand(command) {
			logger.Info("Forwarding write request to leader node")

			replicationClient, err := replicate.NewReplicationClient(strconv.Itoa(s.grpcServer.LeaderID))
			if err != nil {
				logger.Error("Replication client creation failed: " + err.Error())
				s.sendError(conn, "Failed to connect to leader")
				continue
			}

			protoCommand := &proto.Command{Command: command}
			if key, ok := request["key"].(string); ok {
				protoCommand.Key = key
			}
			if value, ok := request["value"].(string); ok {
				protoCommand.Value = value
			}
			if exp, ok := request["exp"].(int64); ok {
				protoCommand.Exp = int32(exp)
			}
			if offset, ok := request["offset"].(int64); ok {
				protoCommand.Offset = int32(offset)
			}

			response, err := replicationClient.ForwardRequest(int32(config.NodeID), protoCommand)
			if err != nil {
				logger.Error("Forward request failed: " + err.Error())
				s.sendError(conn, "Failed to forward request to leader")
				continue
			}

			s.sendResponse(conn, map[string]interface{}{"message": response.Message})
			continue
		}

		// Process command normally on the leader
		response, err := s.CommandHandler.HandleCommand(request)
		if err != nil {
			s.sendError(conn, err.Error())
		} else {
			s.sendResponse(conn, response)
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
	data, err := msgpack.Marshal(response)
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
