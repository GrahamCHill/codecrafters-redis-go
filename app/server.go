package main

import (
	"fmt"
	"net"
	"os"
)

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
