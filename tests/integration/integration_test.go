package integration

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vskvj3/geomys/internal/network"
	"github.com/vskvj3/geomys/internal/utils"
)

// Helper function to send a serialized command and receive the deserialized response
func sendSerializedCommand(t *testing.T, conn net.Conn, command map[string]interface{}) map[string]interface{} {
	// Serialize the command using MessagePack
	data, err := msgpack.Marshal(command)
	if err != nil {
		t.Fatalf("failed to serialize command: %v", err)
	}

	// Send the serialized command
	_, err = conn.Write(data)
	if err != nil {
		t.Fatalf("failed to send command: %v", err)
	}

	// Read the server's response
	reader := bufio.NewReader(conn)
	responseData := make([]byte, 1024) // Buffer size
	n, err := reader.Read(responseData)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	// Deserialize the response
	var response map[string]interface{}
	err = msgpack.Unmarshal(responseData[:n], &response)
	if err != nil {
		t.Fatalf("failed to deserialize response: %v", err)
	}

	return response
}

func TestIntegration(t *testing.T) {
	// Initialize the logger
	utils.NewLogger("", false)

	// Start the server in a goroutine
	go func() {
		server := network.NewServer()
		listener, _ := net.Listen("tcp", ":6379")
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.HandleConnection(conn)
		}
	}()
	time.Sleep(100 * time.Millisecond) // Allow the server time to start

	// Connect to the server
	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Test PING command
	t.Run("PING command", func(t *testing.T) {
		command := map[string]interface{}{"command": "PING"}
		response := sendSerializedCommand(t, conn, command)
		if response["status"] != "OK" || response["message"] != "PONG" {
			t.Errorf("expected {status: OK, message: PONG}, got %v", response)
		}
	})

	// Test ECHO command
	t.Run("ECHO command", func(t *testing.T) {
		command := map[string]interface{}{
			"command": "ECHO",
			"message": "Hello",
		}
		response := sendSerializedCommand(t, conn, command)
		if response["status"] != "OK" || response["message"] != "Hello" {
			t.Errorf("expected {status: OK, message: Hello}, got %v", response)
		}
	})

	// Test GET command for existing key
	t.Run("GET command existing key", func(t *testing.T) {
		setCommand := map[string]interface{}{
			"command": "SET",
			"key":     "mykey",
			"value":   "myvalue",
		}
		_ = sendSerializedCommand(t, conn, setCommand)

		getCommand := map[string]interface{}{
			"command": "GET",
			"key":     "mykey",
		}
		response := sendSerializedCommand(t, conn, getCommand)
		if response["status"] != "OK" || response["value"] != "myvalue" {
			t.Errorf("expected {status: OK, value: myvalue}, got %v", response)
		}
	})

	// Test GET command for non-existing key
	t.Run("GET command non-existing key", func(t *testing.T) {
		getCommand := map[string]interface{}{
			"command": "GET",
			"key":     "nonexistent",
		}
		response := sendSerializedCommand(t, conn, getCommand)
		if response["status"] != "ERROR" {
			t.Errorf("expected {status: ERROR}, got %v", response)
		}
	})

	t.Run("INCR command on existing integer key", func(t *testing.T) {
		// Set an initial value for the key
		setCommand := map[string]interface{}{
			"command": "SET",
			"key":     "counter",
			"value":   "10",
		}
		response := sendSerializedCommand(t, conn, setCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK} for SET command, got %v", response)
		}

		// Increment the key by 5
		incrCommand := map[string]interface{}{
			"command": "INCR",
			"key":     "counter",
			"offset":  "5",
		}
		response = sendSerializedCommand(t, conn, incrCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK} for INCR command, got %v", response)
		}
		if response["value"] != int8(15) { // JSON numbers are unmarshaled as float64
			t.Errorf("expected value: 15 of int8, got %v of type %v", response["value"], fmt.Sprintf("%T", response["value"]))
		}
	})

	t.Run("INCR command on large existing integer key", func(t *testing.T) {
		// Set an initial value for the key
		setCommand := map[string]interface{}{
			"command": "SET",
			"key":     "counter",
			"value":   "999999999999",
		}
		response := sendSerializedCommand(t, conn, setCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK} for SET command, got %v", response)
		}

		// Increment the key by 5
		incrCommand := map[string]interface{}{
			"command": "INCR",
			"key":     "counter",
			"offset":  "1",
		}
		response = sendSerializedCommand(t, conn, incrCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK} for INCR command, got %v", response)
		}
		if response["value"] != uint64(1000000000000) { // JSON numbers are unmarshaled as float64
			t.Errorf("expected value: 1000000000000 of int64, got %v of type %v", response["value"], fmt.Sprintf("%T", response["value"]))
		}
	})

	t.Run("INCR command on non-existing key", func(t *testing.T) {
		incrCommand := map[string]interface{}{
			"command": "INCR",
			"key":     "nonexistent",
			"offset":  "3",
		}
		response := sendSerializedCommand(t, conn, incrCommand)
		if response["status"] != "ERROR" {
			t.Errorf("expected {status: ERROR}, got %v", response)
		}
		if response["message"] != "key not found" {
			t.Errorf("expected error message: 'key not found', got %v", response["message"])
		}
	})

	t.Run("INCR command on non-integer value", func(t *testing.T) {
		// Set a non-integer value for the key
		setCommand := map[string]interface{}{
			"command": "SET",
			"key":     "stringValue",
			"value":   "hello",
		}
		response := sendSerializedCommand(t, conn, setCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK} for SET command, got %v", response)
		}

		// Try to increment the non-integer value
		incrCommand := map[string]interface{}{
			"command": "INCR",
			"key":     "stringValue",
			"offset":  "2",
		}
		response = sendSerializedCommand(t, conn, incrCommand)
		if response["status"] != "ERROR" {
			t.Errorf("expected {status: ERROR}, got %v", response)
		}
		if response["message"] != "value is not an integer" {
			t.Errorf("expected error message: 'value is not an integer', got %v", response["message"])
		}
	})

	t.Run("PUSH command", func(t *testing.T) {
		pushCommand := map[string]interface{}{
			"command": "PUSH",
			"key":     "listOne",
			"value":   "item1",
		}
		response := sendSerializedCommand(t, conn, pushCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK} for PUSH command, got %v", response)
		}

		// Push another item to the list
		pushCommand["value"] = "item2"
		response = sendSerializedCommand(t, conn, pushCommand)
		if response["status"] != "OK" {
			t.Errorf("expected {status: OK} for PUSH command, got %v", response)
		}
	})

	t.Run("LPOP command", func(t *testing.T) {
		// Push items to the list for testing
		sendSerializedCommand(t, conn, map[string]interface{}{
			"command": "PUSH",
			"key":     "listKey",
			"value":   "item1",
		})
		sendSerializedCommand(t, conn, map[string]interface{}{
			"command": "PUSH",
			"key":     "listKey",
			"value":   "item2",
		})

		// Pop the leftmost item
		lpopCommand := map[string]interface{}{
			"command": "LPOP",
			"key":     "listKey",
		}
		response := sendSerializedCommand(t, conn, lpopCommand)
		if response["status"] != "OK" || response["value"] != "item1" {
			t.Errorf("expected {status: OK, value: item1}, got %v", response)
		}

		// Pop the next leftmost item
		response = sendSerializedCommand(t, conn, lpopCommand)
		if response["status"] != "OK" || response["value"] != "item2" {
			t.Errorf("expected {status: OK, value: item2}, got %v", response)
		}

		// Try to pop from an empty list
		response = sendSerializedCommand(t, conn, lpopCommand)
		if response["status"] != "ERROR" {
			t.Errorf("expected {status: ERROR}, got %v", response)
		}
	})

	t.Run("RPOP command", func(t *testing.T) {
		// Push items to the list for testing
		sendSerializedCommand(t, conn, map[string]interface{}{
			"command": "PUSH",
			"key":     "listKey",
			"value":   "item1",
		})
		sendSerializedCommand(t, conn, map[string]interface{}{
			"command": "PUSH",
			"key":     "listKey",
			"value":   "item2",
		})

		// Pop the rightmost item
		rpopCommand := map[string]interface{}{
			"command": "RPOP",
			"key":     "listKey",
		}
		response := sendSerializedCommand(t, conn, rpopCommand)
		if response["status"] != "OK" || response["value"] != "item2" {
			t.Errorf("expected {status: OK, value: item2}, got %v", response)
		}

		// Pop the next rightmost item
		response = sendSerializedCommand(t, conn, rpopCommand)
		if response["status"] != "OK" || response["value"] != "item1" {
			t.Errorf("expected {status: OK, value: item1}, got %v", response)
		}

		// Try to pop from an empty list
		// In future, we migrate this test into status: NOT_FOUND
		// since item not found is not an error, it is condition, makes more sense?
		response = sendSerializedCommand(t, conn, rpopCommand)
		if response["status"] != "ERROR" {
			t.Errorf("expected {status: ERROR}, got %v", response)
		}
	})

}
