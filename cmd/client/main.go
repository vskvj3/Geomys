package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// argParser parses and validates the command and its arguments
func argParser(input string) (map[string]interface{}, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, fmt.Errorf("no command entered")
	}

	command := strings.ToUpper(parts[0])
	request := map[string]interface{}{
		"command": command,
	}

	switch command {
	case "PING":
		// PING requires no additional arguments
		if len(parts) > 1 {
			return nil, fmt.Errorf("PING does not require any arguments")
		}

	case "ECHO":
		// ECHO requires a message
		if len(parts) < 2 {
			return nil, fmt.Errorf("ECHO requires a message")
		}
		request["message"] = strings.Join(parts[1:], " ")

	case "SET":
		// SET requires a key and a value
		if len(parts) < 3 {
			return nil, fmt.Errorf("SET requires a key and value")
		}
		request["key"] = parts[1]
		request["value"] = parts[2]

	case "GET":
		// GET requires a key
		if len(parts) < 2 {
			return nil, fmt.Errorf("GET requires a key")
		}
		request["key"] = parts[1]

	default:
		// Unknown command
		return nil, fmt.Errorf("unknown command: %s", command)
	}

	return request, nil
}

func main() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server. Type commands (e.g., PING, ECHO, SET key value, GET key) and press Enter.")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">> ")
		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		input = strings.TrimSpace(input)

		// Parse and validate the input
		request, err := argParser(input)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Serialize the request using MessagePack
		data, err := msgpack.Marshal(request)
		if err != nil {
			fmt.Println("Error serializing request:", err)
			continue
		}

		// Send the serialized request to the server
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Error sending to server:", err)
			return
		}

		// Read the server's response
		response := make([]byte, 4096)
		n, err := conn.Read(response)
		if err != nil {
			fmt.Println("Error reading from server:", err)
			return
		}

		// Deserialize the server's response
		var serverResponse map[string]interface{}
		err = msgpack.Unmarshal(response[:n], &serverResponse)
		if err != nil {
			fmt.Println("Error deserializing response:", err)
			continue
		}

		// Print the server's response
		if status, ok := serverResponse["status"].(string); ok && status == "OK" {
			if message, ok := serverResponse["message"].(string); ok {
				fmt.Println("Server:", message)
			} else if value, ok := serverResponse["value"]; ok {
				fmt.Println("Server:", value)
			} else {
				fmt.Println("Server: OK")
			}
		} else if status == "ERROR" {
			fmt.Println("Server Error:", serverResponse["message"])
		} else {
			fmt.Println("Unexpected server response:", serverResponse)
		}
	}
}
