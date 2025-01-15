package core

import (
	"fmt"
	"net"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

type CommandHandler struct {
	Database *Database
}

// Create a new CommandHandler instance
func NewCommandHandler(db *Database) *CommandHandler {
	return &CommandHandler{Database: db}
}

// HandleCommand processes client commands and sends appropriate responses
func (h *CommandHandler) HandleCommand(conn net.Conn, request map[string]interface{}) {
	// Process the command
	command, ok := request["command"].(string)
	if !ok {
		h.sendError(conn, "Invalid or missing 'command' field")
		return
	}

	command = strings.ToUpper(command)
	var response map[string]interface{}

	switch command {
	case "PING":
		response = map[string]interface{}{"status": "OK", "message": "PONG"}

	case "ECHO":
		message, ok := request["message"].(string)
		if !ok {
			h.sendError(conn, "ECHO requires a 'message' field")
			return
		}
		response = map[string]interface{}{"status": "OK", "message": message}

	case "SET":
		fmt.Println(request)
		key, keyOk := request["key"].(string)
		value, valueOk := request["value"].(string)
		ttlMs := int64(0)

		// Check for expiration (exp) and handle its type
		if exp, ok := request["exp"]; ok {
			switch v := exp.(type) {
			case int8:
				ttlMs = int64(v)
			case int16:
				ttlMs = int64(v)
			case int32:
				ttlMs = int64(v)
			case int64:
				ttlMs = v
			case uint8:
				ttlMs = int64(v)
			case uint16:
				ttlMs = int64(v)
			case uint32:
				ttlMs = int64(v)
			default:
				fmt.Printf("Invalid type for TTL: %T\n", v)
			}
		}

		// Use key, value, and ttlMs as required
		fmt.Printf("Key: %s, Value: %s, TTL: %d ms\n", key, value, ttlMs)

		if !keyOk || !valueOk {
			h.sendError(conn, "SET requires 'key', 'value' fields")
			return
		}
		print(key, value, ttlMs)
		if err := h.Database.Set(key, value, ttlMs); err != nil {
			h.sendError(conn, err.Error())
			return
		}
		response = map[string]interface{}{"status": "OK"}

	case "GET":
		key, ok := request["key"].(string)
		if !ok {
			h.sendError(conn, "GET requires a 'key' field")
			return
		}
		value, err := h.Database.Get(key)
		if err != nil {
			h.sendError(conn, err.Error())
			return
		}
		if value == "" {
			response = map[string]interface{}{"status": "NOT_FOUND"}
		} else {
			response = map[string]interface{}{"status": "OK", "value": value}
		}

	default:
		h.sendError(conn, "Unknown command")
		return
	}

	// Send the response
	h.sendResponse(conn, response)
}

// sendResponse serializes the response and sends it to the client
func (h *CommandHandler) sendResponse(conn net.Conn, response map[string]interface{}) {
	data, err := msgpack.Marshal(response)
	if err != nil {
		fmt.Println("Failed to encode response:", err)
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Failed to send response:", err)
	}
}

// sendError sends an error message to the client
func (h *CommandHandler) sendError(conn net.Conn, errorMessage string) {
	response := map[string]interface{}{"status": "ERROR", "message": errorMessage}
	h.sendResponse(conn, response)
}
