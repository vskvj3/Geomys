package network

import (
	"fmt"
	"net"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/vskvj3/geomys/internal/core"
)

type Server struct {
	CommandHandler *core.CommandHandler
}

func NewServer() *Server {
	db := core.NewDatabase()
	handler := core.NewCommandHandler(db)
	return &Server{CommandHandler: handler}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Client disconnected:", err)
			return
		}

		// Deserialize the incoming request
		var request map[string]interface{}
		err = msgpack.Unmarshal(buffer[:n], &request)

		if err != nil {
			fmt.Println("Failed to decode request:", err)
			continue
		}

		s.CommandHandler.HandleCommand(conn, request)
	}
}
