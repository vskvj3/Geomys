package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

func argParser(input string) (map[string]interface{}, error) {
	var parts []string
	var current string
	inQuotes := false

	for _, r := range input {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			if !inQuotes {
				// Closing quote ends the current token
				parts = append(parts, current)
				current = ""
			}
		case r == ' ' && !inQuotes:
			// Space outside quotes ends the current token
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		default:
			// Append character to the current token
			current += string(r)
		}
	}

	// Append the last token if it exists
	if current != "" {
		parts = append(parts, current)
	}

	if inQuotes {
		return nil, fmt.Errorf("unmatched quotes in input")
	}

	fmt.Printf("Parsed parts: %+q\n", parts)

	if len(parts) == 0 {
		return nil, fmt.Errorf("no command entered")
	}

	command := strings.ToUpper(parts[0])
	request := map[string]interface{}{
		"command": command,
	}

	switch command {
	case "PING":
		if len(parts) > 1 {
			return nil, fmt.Errorf("PING does not require any arguments")
		}

	case "ECHO":
		if len(parts) < 2 {
			return nil, fmt.Errorf("ECHO requires a message")
		}
		request["message"] = strings.Join(parts[1:], " ")

	case "SET":
		if len(parts) < 3 {
			return nil, fmt.Errorf("SET requires a key, value, and optional expiry")
		}
		request["key"] = parts[1]
		request["value"] = parts[2]
		if len(parts) > 3 {
			expiry, err := strconv.Atoi(parts[3])
			if err != nil {
				return nil, fmt.Errorf("invalid expiry value: %s", parts[3])
			}
			request["exp"] = expiry
		}

	case "GET":
		if len(parts) < 2 {
			return nil, fmt.Errorf("GET requires a key")
		}
		request["key"] = parts[1]

	case "INCR":
		if len(parts) < 3 {
			return nil, fmt.Errorf("INCR requires a key, and offset")
		}
		request["key"] = parts[1]
		request["offset"] = parts[2]

	default:
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
