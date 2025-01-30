package network

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/utils"
)

type Server struct {
	CommandHandler *core.CommandHandler
}

func NewServer() (*Server, error) {
	db := core.NewDatabase()
	db.StartCleanup(100 * time.Millisecond)

	handler := core.NewCommandHandler(db)
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

		s.CommandHandler.HandleCommand(conn, request)
	}
}
