package core

import (
	"net"
	"strings"
)

type CommandHandler struct {
	Database *Database
}

// Create a new CommandHandler instance
func NewCommandHandler(db *Database) *CommandHandler {
	return &CommandHandler{Database: db}
}

// HandleCommand processes client commands and sends appropriate responses
func (h *CommandHandler) HandleCommand(conn net.Conn, message string) {
	// Parse the command and arguments
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
		key, value := parts[1], parts[2]
		if err := h.Database.Set(key, value); err != nil {
			_, _ = conn.Write([]byte("Error: " + err.Error() + "\n"))
		} else {
			_, _ = conn.Write([]byte("OK\n"))
		}

	case "GET":
		if len(parts) < 2 {
			_, _ = conn.Write([]byte("Error: GET requires a key\n"))
			return
		}
		key := parts[1]
		value, err := h.Database.Get(key)
		if err != nil {
			_, _ = conn.Write([]byte("Error: " + err.Error() + "\n"))
		} else {
			_, _ = conn.Write([]byte(value + "\n"))
		}

	default:
		_, _ = conn.Write([]byte("Unknown Command\n"))
	}
}
