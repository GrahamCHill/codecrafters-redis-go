package main

import (
	"fmt"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	fmt.Println("tcp", "0.0.0.0:6379")

	// Listening on port 6379
	l, err := net.Listen("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {

		}
	}(l)
	fmt.Println("Server is listening on port 6379")

	for {
		// Accept an incoming connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection", err.Error())
			continue
		}

		// Handle the connection in a separate Go routine
		go handleConnection(conn)
	}
}

// Handle the client connection by sending a hardcoded response
func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// Hardcoded response for testing purposes
	response := "+PONG\r\n"

	// Write the response to client
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to send response", err.Error())
	}
}
