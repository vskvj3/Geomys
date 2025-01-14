package network

import (
	"bufio"
	"fmt"
	"net"

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
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", err)
			return
		}
		s.CommandHandler.HandleCommand(conn, message)
	}
}
