package main

import (
	"fmt"
	"net"

	"github.com/vskvj3/geomys/internal/network"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 6379...")

	server := network.NewServer()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go server.HandleConnection(conn)
	}
}
