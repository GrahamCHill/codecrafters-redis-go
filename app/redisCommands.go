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
var configDir string
var configDBFilename string

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
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'SET' command\r\n"))
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
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'GET' command\r\n"))
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

func handleCONFIGCommand(conn net.Conn, elements []string) {
	if len(elements) < 3 || strings.ToUpper(elements[1]) != "GET" {
		_, err := conn.Write([]byte("-ERR wrong number of arguments for 'CONFIG' command\r\n"))
		if err != nil {
			log.Println("Failed to send CONFIG error response:", err)
		}
		return
	}

	param := elements[2]
	var value string
	switch strings.ToLower(param) {
	case "dir":
		value = configDir
	case "dbfilename":
		value = configDBFilename
	default:
		_, err := conn.Write([]byte("-ERR unknown configuration parameter\r\n"))
		if err != nil {
			log.Println("Failed to send unknown parameter response:", err)
		}
		return
	}

	response := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(param), param, len(value), value)
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Println("Failed to send CONFIG GET response:", err)
	}
}

func handleErrorCommand(conn net.Conn) {
	_, err := conn.Write([]byte("-ERR unknown command\r\n"))
	if err != nil {
		log.Println("Failed to send ECHO RESPONSE", err)
	}
}
