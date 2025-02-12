package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Request struct {
	Command string      `msgpack:"command"`
	Message string      `msgpack:"message,omitempty"`
	Key     string      `msgpack:"key,omitempty"`
	Value   string      `msgpack:"value,omitempty"`
	Exp     int         `msgpack:"exp,omitempty"`
	Offset  interface{} `msgpack:"offset,omitempty"`
}

func argParser(input string) (Request, error) {
	parts := []string{}
	current := ""
	inQuotes := false

	for _, char := range input {
		switch char {
		case '"':
			inQuotes = !inQuotes
			if !inQuotes {
				parts = append(parts, current)
				current = ""
			}
		case ' ':
			if !inQuotes && current != "" {
				parts = append(parts, current)
				current = ""
			}
		default:
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	if inQuotes {
		return Request{}, errors.New("unmatched quotes in input")
	}

	if len(parts) == 0 {
		return Request{}, errors.New("no command entered")
	}

	command := strings.ToUpper(parts[0])
	req := Request{Command: command}

	switch command {
	case "PING":
		if len(parts) > 1 {
			return Request{}, errors.New("PING does not require any arguments")
		}

	case "ECHO":
		if len(parts) < 2 {
			return Request{}, errors.New("ECHO requires a message")
		}
		req.Message = strings.Join(parts[1:], " ")

	case "SET":
		if len(parts) < 3 {
			return Request{}, errors.New("SET requires a key, value, and optional expiry")
		}
		req.Key = parts[1]
		req.Value = parts[2]
		if len(parts) > 3 {
			exp, err := strconv.Atoi(parts[3])
			if err != nil {
				return Request{}, fmt.Errorf("invalid expiry value: %s", parts[3])
			}
			req.Exp = exp
		}

	case "GET":
		if len(parts) < 2 {
			return Request{}, errors.New("GET requires a key")
		}
		req.Key = parts[1]

	case "INCR":
		if len(parts) < 3 {
			return Request{}, errors.New("INCR requires a key and offset")
		}
		req.Key = parts[1]
		req.Offset = parts[2]

	case "PUSH":
		if len(parts) < 3 {
			return Request{}, errors.New("PUSH requires a key, value")
		}
		req.Key = parts[1]
		req.Value = parts[2]

	case "LPOP", "RPOP":
		if len(parts) < 2 {
			return Request{}, fmt.Errorf("%s requires a key", command)
		}
		req.Key = parts[1]

	case "FLUSHDB":
		if len(parts) > 1 {
			return Request{}, errors.New("FLUSHDB does not require any arguments")
		}

	default:
		return Request{}, fmt.Errorf("unknown command: %s", command)
	}

	return req, nil
}

func main() {
	// Define a port flag with a default value of 6379
	port := flag.Int("port", 6379, "Port number of the server")
	flag.Parse() // Parse the command-line flags

	serverAddr := fmt.Sprintf("localhost:%d", *port)

	// Connect to the server using the specified port
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Error connecting to server on port %d: %v\n", *port, err)
		return
	}
	defer conn.Close()

	fmt.Printf("Connected to server on port %d. Type commands (e.g., PING, ECHO, SET key value, GET key) and press Enter.\n", *port)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		req, err := argParser(userInput)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		data, err := msgpack.Marshal(req)
		if err != nil {
			fmt.Printf("Error serializing request: %v\n", err)
			continue
		}

		startTime := time.Now()
		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			continue
		}

		buffer := make([]byte, 4096)
		length, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			continue
		}

		responseTime := time.Since(startTime).Milliseconds()
		var serverResponse map[string]interface{}
		if err := msgpack.Unmarshal(buffer[:length], &serverResponse); err != nil {
			fmt.Printf("Error deserializing response: %v\n", err)
			continue
		}

		status, _ := serverResponse["status"].(string)
		switch status {
		case "OK":
			if message, ok := serverResponse["message"].(string); ok {
				fmt.Printf("Server: %s\n", message)
			} else if value, ok := serverResponse["value"].(string); ok {
				fmt.Printf("Server: %s\n", value)
			} else {
				fmt.Println("Server: OK")
			}
		case "ERROR":
			fmt.Printf("Server Error: %s\n", serverResponse["message"])
		default:
			fmt.Printf("Unexpected server response: %v\n", serverResponse)
		}

		fmt.Printf("Response time: %d ms\n", responseTime)
	}

	fmt.Println("Closing connection.")
}
