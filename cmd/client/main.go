package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server. Type commands (e.g., PING) and press Enter.")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">> ")
		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Send the input to the server
		_, err = conn.Write([]byte(input))
		if err != nil {
			fmt.Println("Error sending to server:", err)
			return
		}

		// Read the server's response
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			return
		}

		fmt.Printf("Server: %s", response)
	}
}
