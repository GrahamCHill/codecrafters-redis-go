package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var dataStore = make(map[string]string)

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

	buff := make([]byte, 1024)

	for {
		length, err := conn.Read(buff)
		if err != nil {
			log.Println("Connection read error or closed:", err)
			return
		}

		rawData := string(buff[:length])
		lines := strings.Split(rawData, "\n")

		if len(lines) > 0 && strings.HasPrefix(lines[0], "*") {
			var elements []string
			for i := 1; i < len(lines); i++ {
				if strings.HasPrefix(lines[i], "$") {
					elementLength, err := strconv.Atoi(strings.Trim(lines[i][1:], "\r"))
					if err != nil {
						log.Println("Error parsing element length:", err)
						return
					}
					if i+1 < len(lines) && len(strings.Trim(lines[i+1], "\r")) == elementLength {
						elements = append(elements, strings.Trim(lines[i+1], "\r"))
						i++ // Skip the next line as it is part of the current element
					}
				}
			}

			// Handling Multiple Redis commands
			if len(elements) > 0 {
				command := strings.ToUpper(elements[0])

				switch command {
				case "PING":
					_, err := conn.Write([]byte("+PONG\r\n"))
					if err != nil {
						log.Println("Failed to send PONG RESPONSE", err)
					}

				case "ECHO":
					if len(elements) == 2 {
						response := fmt.Sprintf("$%d\r\n%s\r\n", len(elements[1]), elements[1])
						_, err := conn.Write([]byte(response))
						if err != nil {
							log.Println("Failed to send ECHO RESPONSE", err)
						}
					} else {
						_, err := conn.Write([]byte("-ERR wong number of arguments provided for 'ECHO' command\r\n"))
						if err != nil {
							log.Println("Failed to send ECHO error RESPONSE", err)
						}
					}

				case "SET":
					if len(elements) == 3 {
						key := elements[1]
						value := elements[2]
						dataStore[key] = value

						_, err := conn.Write([]byte("+OK\r\n"))
						if err != nil {
							log.Println("Failed to send SET RESPONSE", err)
						}
					} else {
						_, err := conn.Write([]byte("-ERR wong number of arguments provided for 'SET' command\r\n"))
						if err != nil {
							log.Println("Failed to send SET RESPONSE", err)
						}
					}

				case "GET":
					if len(elements) == 2 {
						key := elements[1]
						value, exists := dataStore[key]
						if exists {
							response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
							_, err := conn.Write([]byte(response))
							if err != nil {
								log.Println("Failed to send GET RESPONSE", err)
							}
						} else {
							_, err := conn.Write([]byte("$-1\r\n"))
							if err != nil {
								log.Println("Failed to send GET NULL RESPONSE", err)
							}
						}
					} else {
						_, err := conn.Write([]byte("-ERR wong number of arguments provided for 'GET' command\r\n"))
						if err != nil {
							log.Println("Failed to send GET NULL RESPONSE", err)
						}
					}
				default:
					_, err := conn.Write([]byte("-ERR unknown command\r\n"))
					if err != nil {
						log.Println("Failed to send ECHO RESPONSE", err)
					}
				}
			}
		}
	}
}
