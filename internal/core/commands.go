package core

import (
	"net"
	"strings"
)

type CommandHandler struct {
	Database *Database
}

func NewCommandHandler(db *Database) *CommandHandler {
	return &CommandHandler{Database: db}
}

func (h *CommandHandler) HandleCommand(conn net.Conn, message string) {
	parts := strings.SplitN(strings.TrimSpace(message), " ", 3)
	command := strings.ToUpper(parts[0])

	switch command {
	case "PING":
		_, _ = conn.Write([]byte("PONG\n"))
	case "ECHO":
		if len(parts) < 2 {
			_, _ = conn.Write([]byte("Error: ECHO requires a message\n"))
			return
		}
		response := strings.Join(parts[1:], " ")
		_, _ = conn.Write([]byte(response + "\n"))
	case "SET":
		if len(parts) < 3 {
			_, _ = conn.Write([]byte("Error: SET requires a key and value\n"))
			return
		}
		h.Database.Set(parts[1], parts[2])
		_, _ = conn.Write([]byte("OK\n"))
	case "GET":
		if len(parts) < 2 {
			_, _ = conn.Write([]byte("Error: GET requires a key\n"))
			return
		}
		value, exists := h.Database.Get(parts[1])
		if exists {
			_, _ = conn.Write([]byte(value + "\n"))
		} else {
			_, _ = conn.Write([]byte("nil\n"))
		}
	default:
		_, _ = conn.Write([]byte("Unknown Command\n"))
	}
}
