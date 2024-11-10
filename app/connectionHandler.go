package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StoreItem struct {
	Value  string
	Expiry *time.Time // Pointer to time.Time; nil if no expiry
}

var dataStore = make(map[string]StoreItem)
var dataStoreLock = sync.RWMutex{}

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

func handlePINGCommand(conn net.Conn) {
	_, err := conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		log.Println("Failed to send PONG RESPONSE", err)
	}
}

func handleECHOCommand(conn net.Conn, elements []string) {
	if len(elements) < 2 {
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'ECHO' command\r\n"))
		if err != nil {
			log.Println("Failed to send ECHO RESPONSE", err)
		}
	}
	response := fmt.Sprintf("$%d\r\n%s\r\n", len(elements[1]), elements[1])
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Println("Failed to send ECHO RESPONSE", err)
	}

}

func handleSETCommand(conn net.Conn, elements []string) {
	if len(elements) < 3 {
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
		if err != nil {
			log.Println("Failed to send SET error response:", err)
		}
		return
	}

	key := elements[1]
	value := elements[2]
	var expiry *time.Time

	if len(elements) > 3 && strings.ToUpper(elements[3]) == "PX" {
		if len(elements) != 5 {
			_, err := conn.Write([]byte("-ERR syntax error\r\n"))
			if err != nil {
				log.Println("Failed to send SET syntax error response:", err)
			}
			return
		}

		expiryMillis, err := strconv.Atoi(elements[4])
		if err != nil || expiryMillis <= 0 {
			_, err := conn.Write([]byte("-ERR invalid PX value\r\n"))
			if err != nil {
				log.Println("Failed to send SET PX value error response:", err)
			}
			return
		}

		expiryTime := time.Now().Add(time.Duration(expiryMillis) * time.Millisecond)
		expiry = &expiryTime
	}

	dataStoreLock.Lock()
	dataStore[key] = StoreItem{Value: value, Expiry: expiry}
	dataStoreLock.Unlock()

	_, err := conn.Write([]byte("+OK\r\n"))
	if err != nil {
		log.Println("Failed to send SET response:", err)
	}
}

func handleGETCommand(conn net.Conn, elements []string) {
	if len(elements) != 2 {
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
		if err != nil {
			log.Println("Failed to send GET error response:", err)
		}
		return
	}

	key := elements[1]
	dataStoreLock.RLock()
	item, exists := dataStore[key]
	dataStoreLock.RUnlock()

	if !exists || (item.Expiry != nil && time.Now().After(*item.Expiry)) {
		if exists {
			// If the key exists but has expired, delete it
			dataStoreLock.Lock()
			delete(dataStore, key)
			dataStoreLock.Unlock()
		}
		_, err := conn.Write([]byte("$-1\r\n"))
		if err != nil {
			log.Println("Failed to send GET null response:", err)
		}
		return
	}

	response := fmt.Sprintf("$%d\r\n%s\r\n", len(item.Value), item.Value)
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Println("Failed to send GET response:", err)
	}
}
