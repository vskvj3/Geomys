package core

import (
	"errors"
	"strconv"
	"strings"

	"github.com/vskvj3/geomys/internal/persistence"
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
func (h *CommandHandler) HandleCommand(request map[string]interface{}) (map[string]interface{}, error) {
	disk, err := persistence.CreateOrReplacePersistence()
	if err != nil {
		return nil, errors.New("could not access disk: " + err.Error())
	}
	// Process the command
	command, ok := request["command"].(string)
	if !ok {
		return nil, errors.New("invalid or missing 'command' field")
	}

	command = strings.ToUpper(command)
	var response map[string]interface{}

	switch command {
	case "PING":
		response = map[string]interface{}{"status": "OK", "message": "PONG"}

	case "ECHO":
		message, ok := request["message"].(string)
		if !ok {
			return nil, errors.New("ECHO requires a 'message' field")
		}
		response = map[string]interface{}{"status": "OK", "message": message}

	case "SET":
		if err := disk.LogRequest(request); err != nil {
			return nil, errors.New("reuest logging to disk failed")
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
				return nil, errors.New("Invalid type for TTL: " + v.(string))
			}
		}

		if !keyOk || !valueOk {
			return nil, errors.New("SET requires 'key', 'value' fields")
		}

		if err := h.Database.Set(key, value, ttlMs); err != nil {
			return nil, errors.New("Set failed: " + err.Error())
		}
		response = map[string]interface{}{"status": "OK"}

	case "GET":
		key, ok := request["key"].(string)
		if !ok {
			return nil, errors.New("GET requires a 'key' field")
		}
		value, err := h.Database.Get(key)
		if err != nil {
			return nil, errors.New("Get failed: " + err.Error())
		}
		if value == "" {
			// If it does this, you are doing something very very wrong!!
			response = map[string]interface{}{"status": "NOT_FOUND"}
		} else {
			response = map[string]interface{}{"status": "OK", "value": value}
		}

	case "INCR":
		if err := disk.LogRequest(request); err != nil {
			return nil, errors.New("reuest logging to disk failed")
		}

		key, keyOk := request["key"].(string)
		offset, offsetOk := request["offset"].(string) // JSON numbers are unmarshaled as float64

		if !keyOk {
			return nil, errors.New("INCR requires a 'key' field")
		}
		if !offsetOk {
			return nil, errors.New("INCR requires an 'offset' field (integer)")
		}

		// Convert offset to int
		intOffset, err := strconv.Atoi(offset)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		// Call the Incr function
		newValue, err := h.Database.Incr(key, intOffset)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		// Send the success response
		response = map[string]interface{}{
			"status": "OK",
			"value":  newValue,
		}

	case "PUSH":
		if err := disk.LogRequest(request); err != nil {
			return nil, errors.New("Push failed: " + "Reuest logging to disk failed: " + err.Error())
		}

		key, keyOk := request["key"].(string)
		value, valueOk := request["value"].(string)

		if !keyOk || !valueOk {
			return nil, errors.New("PUSH requires 'key', 'value' fields")
		}

		if err := h.Database.Push(key, value); err != nil {
			return nil, errors.New("Push failed: " + err.Error())
		}
		response = map[string]interface{}{"status": "OK"}

	case "LPOP":
		if err := disk.LogRequest(request); err != nil {
			return nil, errors.New("reuest logging to disk failed")
		}

		key, ok := request["key"].(string)
		if !ok {
			return nil, errors.New("LPOP requires a 'key' field")
		}
		value, err := h.Database.Lpop(key)
		if err != nil {
			return nil, errors.New("Lpop failed: " + err.Error())
		}

		if value == "" {
			// If it does this, you are doing something very very wrong!!
			response = map[string]interface{}{"status": "NOT_FOUND"}
		} else {
			response = map[string]interface{}{"status": "OK", "value": value}
		}

	case "RPOP":
		if err := disk.LogRequest(request); err != nil {
			return nil, errors.New("reuest logging to disk failed")
		}

		key, ok := request["key"].(string)
		if !ok {
			return nil, errors.New("LPOP requires a 'key' field")
		}
		value, err := h.Database.Rpop(key)
		if err != nil {
			return nil, errors.New("Rpop failed: " + err.Error())
		}

		if value == "" {
			// If it does this, you are doing something very very wrong!!
			response = map[string]interface{}{"status": "NOT_FOUND"}
		} else {
			response = map[string]interface{}{"status": "OK", "value": value}
		}

	// warning: there should be some auth to perform this!!
	case "FLUSHDB":
		if err := disk.Clear(); err != nil {
			return nil, errors.New("Clearing persisted data failed: " + err.Error())
		}

		h.Database.Clear()

		response = map[string]interface{}{"status": "OK"}

	default:
		return nil, errors.New("unknown command")
	}

	// Send the response
	return response, nil
}
