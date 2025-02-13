package utils

import (
	"errors"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vskvj3/geomys/internal/replicate/proto"
)

// ConvertRequestToCommand converts a request map to a proto.Command
func ConvertRequestToCommand(request map[string]interface{}) (*proto.Command, error) {
	command, ok := request["command"].(string)
	if !ok {
		return nil, errors.New("invalid command format")
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

	return protoCommand, nil
}

// ConvertCommandToRequest converts a proto.Command back to a map
func ConvertCommandToRequest(cmd *proto.Command) map[string]interface{} {
	request := map[string]interface{}{
		"command": cmd.Command,
	}
	if cmd.Key != "" {
		request["key"] = cmd.Key
	}
	if cmd.Value != "" {
		request["value"] = cmd.Value
	}
	if cmd.Exp != 0 {
		request["exp"] = cmd.Exp
	}
	if cmd.Offset != 0 {
		request["offset"] = cmd.Offset
	}

	return request
}

// EncodeResponse serializes a response map into a byte slice
func EncodeResponse(response map[string]interface{}) ([]byte, error) {
	return msgpack.Marshal(response)
}

// DecodeRequest deserializes a byte slice into a request map
func DecodeRequest(data []byte) (map[string]interface{}, error) {
	var request map[string]interface{}
	err := msgpack.Unmarshal(data, &request)
	return request, err
}
