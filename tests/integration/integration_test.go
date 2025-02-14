package integration

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vskvj3/geomys/internal/core"
	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/utils"
)

// Helper function to send a serialized command and receive the deserialized response
func sendSerializedCommand(t *testing.T, conn net.Conn, command map[string]interface{}) map[string]interface{} {
	data, err := msgpack.Marshal(command)
	if err != nil {
		t.Fatalf("failed to serialize command: %v", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		t.Fatalf("failed to send command: %v", err)
	}

	reader := bufio.NewReader(conn)
	responseData := make([]byte, 1024)
	n, err := reader.Read(responseData)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	err = msgpack.Unmarshal(responseData[:n], &response)
	if err != nil {
		t.Fatalf("failed to deserialize response: %v", err)
	}

	return response
}

func TestIntegration(t *testing.T) {
	utils.NewLogger("", false)
	db := core.NewDatabase()
	commandHandler := core.NewCommandHandler(db)

	// Start the server
	go func() {
		utils.LoadConfig("configPath")
		server, _ := network.NewServer(nil, "6379", commandHandler)
		server.Start()
	}()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	t.Run("INCR command on existing integer key", func(t *testing.T) {
		setCommand := map[string]interface{}{"command": "SET", "key": "counter", "value": "10"}
		response := sendSerializedCommand(t, conn, setCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK}, got %v", response)
		}

		incrCommand := map[string]interface{}{"command": "INCR", "key": "counter", "offset": "5"}
		response = sendSerializedCommand(t, conn, incrCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK}, got %v", response)
		}
		if response["value"] != int8(15) {
			t.Errorf("expected value: 15 of int8, got %v of type %T", response["value"], response["value"])
		}
	})

	t.Run("INCR command on large existing integer key", func(t *testing.T) {
		setCommand := map[string]interface{}{"command": "SET", "key": "counter", "value": "999999999999"}
		response := sendSerializedCommand(t, conn, setCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK}, got %v", response)
		}

		incrCommand := map[string]interface{}{"command": "INCR", "key": "counter", "offset": "1"}
		response = sendSerializedCommand(t, conn, incrCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK}, got %v", response)
		}
		if response["value"] != uint64(1000000000000) {
			t.Errorf("expected value: 1000000000000 of float64, got %v of type %T", response["value"], response["value"])
		}
	})

	t.Run("LPOP command with empty list should return NOT_FOUND", func(t *testing.T) {
		lpopCommand := map[string]interface{}{"command": "LPOP", "key": "emptyList"}
		response := sendSerializedCommand(t, conn, lpopCommand)
		if response["status"] != "ERROR" {
			t.Errorf("expected {status: ERROR}, got %v", response)
		}
	})

	t.Run("RPOP command with empty list should return NOT_FOUND", func(t *testing.T) {
		rpopCommand := map[string]interface{}{"command": "RPOP", "key": "emptyList"}
		response := sendSerializedCommand(t, conn, rpopCommand)
		if response["status"] != "ERROR" {
			t.Errorf("expected {status: ERROR}, got %v", response)
		}
	})
}
