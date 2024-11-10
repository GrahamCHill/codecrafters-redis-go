package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
)

func init() {
	// Parse command-line flags for dir and dbfilename
	flag.StringVar(&configDir, "dir", "/tmp", "Directory to store RDB files")
	flag.StringVar(&configDBFilename, "dbfilename", "dump.rdb", "RDB filename")
	flag.Parse()
}

// Handle the client connection by sending a response to command
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
					handlePINGCommand(conn)

				case "ECHO":
					handleECHOCommand(conn, elements)

				case "SET":
					handleSETCommand(conn, elements)

				case "GET":
					handleGETCommand(conn, elements)

				case "CONFIG":
					handleCONFIGCommand(conn, elements)

				default:
					handleErrorCommand(conn)
				}
			}
		}
	}
}
