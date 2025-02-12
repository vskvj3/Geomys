package network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/replicate"
	"github.com/vskvj3/geomys/internal/utils"
)

type Server struct {
	CommandHandler *core.CommandHandler
}

func NewServer(leaderAddr string) (*Server, error) {
	fmt.Println("Leader address: " + leaderAddr)
	db := core.NewDatabase()
	handler := core.NewCommandHandler(db)
	logger := utils.GetLogger()
	config, err := utils.GetConfig()

	if err != nil {
		logger.Error("Loading config in server failed: " + err.Error())
	}

	//rebuild from persistence if standalone mode, else request sync from leader
	if !config.IsLeader && leaderAddr != "" {
		replicationClient, err := replicate.NewReplicationClient(leaderAddr)
		if err != nil {
			logger.Error("Replication client creation failed: " + err.Error())
		}
		logger.Info("Re-syncing from leader node...")
		replicationClient.SyncRequest(handler)
	} else {
		// only rebuild from persistence if no replication is happening (ie. leader node)
		if err := db.RebuildFromPersistence(); err != nil {
			logger.Warn("Could not read from persistence: " + err.Error())
		} else {
			logger.Info("Loaded data from persistence")
		}
	}

	// start database cleanup (to remove expired keys)
	db.StartCleanup(100 * time.Millisecond)

	return &Server{CommandHandler: handler}, nil
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer func() {
		logger := utils.GetLogger() // Use the singleton logger
		logger.Info("Client disconnected: " + conn.RemoteAddr().String())
		conn.Close()
	}()

	logger := utils.GetLogger() // Use the singleton logger

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)

		if err != nil {
			if errors.Is(err, io.EOF) {
				// Client closed the connection
				logger.Info("Client closed the connection: " + conn.RemoteAddr().String())
			} else {
				// Other errors
				logger.Error("Error reading from client: " + err.Error())
			}
			return
		}

		// Deserialize the incoming request
		var request map[string]interface{}
		err = msgpack.Unmarshal(buffer[:n], &request)

		if err != nil {
			logger.Error("Failed to decode request: " + err.Error())
			continue
		}

		// Log the received request (optional, for debugging)
		logger.Debug("Received request from client: " + conn.RemoteAddr().String())

		response, err := s.CommandHandler.HandleCommand(request)
		if err != nil {
			s.sendError(conn, err.Error())
		} else {
			s.sendResponse(conn, response)
		}
	}
}

// sendResponse serializes the response and sends it to the client
func (h *Server) sendResponse(conn net.Conn, response map[string]interface{}) {
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
func (h *Server) sendError(conn net.Conn, errorMessage string) {
	response := map[string]interface{}{"status": "ERROR", "message": errorMessage}
	h.sendResponse(conn, response)
}
