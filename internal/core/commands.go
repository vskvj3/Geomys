package core

import (
	"net"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vskvj3/geomys/internal/persistence"
	"github.com/vskvj3/geomys/internal/utils"
)

type CommandHandler struct {
	Database    *Database
	Persistence *persistence.Persistence
}

// Create a new CommandHandler instance
func NewCommandHandler(db *Database) *CommandHandler {
	return &CommandHandler{Database: db}
}

// HandleCommand processes client commands and sends appropriate responses
func (h *CommandHandler) HandleCommand(conn net.Conn, request map[string]interface{}) {
	disk, err := persistence.CreateOrReplacePersistence()
	if err != nil {
		h.sendError(conn, "Could not access disk")
	}
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
		if err := disk.LogRequest(request); err != nil {
			h.sendError(conn, "Reuest logging to disk failed")
		}

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
				h.sendError(conn, "Invalid type for TTL: "+v.(string))
			}
		}

		if !keyOk || !valueOk {
			h.sendError(conn, "SET requires 'key', 'value' fields")
			return
		}

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
			// If it does this, you are doing something very very wrong!!
			response = map[string]interface{}{"status": "NOT_FOUND"}
		} else {
			response = map[string]interface{}{"status": "OK", "value": value}
		}

	case "INCR":
		if err := disk.LogRequest(request); err != nil {
			h.sendError(conn, "Reuest logging to disk failed")
		}

		key, keyOk := request["key"].(string)
		offset, offsetOk := request["offset"].(string) // JSON numbers are unmarshaled as float64

		if !keyOk {
			h.sendError(conn, "INCR requires a 'key' field")
			return
		}
		if !offsetOk {
			h.sendError(conn, "INCR requires an 'offset' field (integer)")
			return
		}

		// Convert offset to int
		intOffset, err := strconv.Atoi(offset)
		if err != nil {
			h.sendError(conn, err.Error())
			return
		}

		// Call the Incr function
		newValue, err := h.Database.Incr(key, intOffset)
		if err != nil {
			h.sendError(conn, err.Error())
			return
		}

		// Send the success response
		response = map[string]interface{}{
			"status": "OK",
			"value":  newValue,
		}

	case "PUSH":
		if err := disk.LogRequest(request); err != nil {
			h.sendError(conn, "Reuest logging to disk failed")
		}

		key, keyOk := request["key"].(string)
		value, valueOk := request["value"].(string)

		if !keyOk || !valueOk {
			h.sendError(conn, "PUSH requires 'key', 'value' fields")
			return
		}

		if err := h.Database.Push(key, value); err != nil {
			h.sendError(conn, err.Error())
			return
		}
		response = map[string]interface{}{"status": "OK"}

	case "LPOP":
		if err := disk.LogRequest(request); err != nil {
			h.sendError(conn, "Reuest logging to disk failed")
		}

		key, ok := request["key"].(string)
		if !ok {
			h.sendError(conn, "LPOP requires a 'key' field")
			return
		}
		value, err := h.Database.Lpop(key)
		if err != nil {
			h.sendError(conn, err.Error())
			return
		}

		if value == "" {
			// If it does this, you are doing something very very wrong!!
			response = map[string]interface{}{"status": "NOT_FOUND"}
		} else {
			response = map[string]interface{}{"status": "OK", "value": value}
		}

	case "RPOP":
		if err := disk.LogRequest(request); err != nil {
			h.sendError(conn, "Reuest logging to disk failed")
		}

		key, ok := request["key"].(string)
		if !ok {
			h.sendError(conn, "LPOP requires a 'key' field")
			return
		}
		value, err := h.Database.Rpop(key)
		if err != nil {
			h.sendError(conn, err.Error())
			return
		}

		if value == "" {
			// If it does this, you are doing something very very wrong!!
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
func (h *CommandHandler) sendError(conn net.Conn, errorMessage string) {
	response := map[string]interface{}{"status": "ERROR", "message": errorMessage}
	h.sendResponse(conn, response)
}
