package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var keyValue = make(map[string]string)
var mu sync.Mutex
var format = []string{"method", "key", "value"}
const waitTime = 1 * time.Minute

func loadData() {
	path, err := filepath.Abs("data/keyValueStorage.txt")
	if err!=nil {
		fmt.Println("Error: generating the absolute path")
	}

	dat, err := os.ReadFile(path)
	if err!=nil {
		fmt.Println("Error: reading the storage file ", err)
		os.Exit(1)
	}

	lst := strings.Split(string(dat), "\n")
	for _, line := range lst {
		parts := strings.Split(line, " ")
		if len(parts) > 1 {
			keyValue[parts[0]] = parts[1]
		}
	}

	fmt.Println("Current storage ", keyValue)
}

func storeData() {
	path, err := filepath.Abs("data/keyValueStorage.txt")
	if err!=nil {
		fmt.Println("Error: generating the absolute path")
	}
	
	f, err := os.Create(path)
	if err!=nil {
		fmt.Println("Error: creating a connection to storage")
	}
	defer f.Close()

	for key, value := range keyValue {
		kv := []byte(key + " " + value + "\n")
		_, err = f.Write(kv)
		if err!=nil {
			fmt.Println("Error: writing to storage")
		}
	}
}

func performTask(req map[string]string) string {
	mu.Lock()
	defer func () {
		storeData()

		mu.Unlock()
	} ()

	if len(req) < 1 {
		return "Invalid command"
	}

	req["method"] = strings.ToUpper(req["method"])

	if req["method"] == "GET" {
		if len(req) < 2 {
			return "Error: Get requires 1 argument"
		}

		if (keyValue[req["key"]] != "") {
			return keyValue[req["key"]]
		} else {
			return "Error: " + req["key"] + " not found"
		}		
	} else if req["method"] == "LIST" {
		res := ""
		for item, value := range keyValue {
			res += item + ": " + value + ", "
		}
		res += "\n"
		
		return res
	} else if req["method"] == "POST" {
		if len(req) < 3 {
			return "Error: Post requires 2 arguments"
		}

		if (keyValue[req["key"]] == "") {
			keyValue[req["key"]] = req["value"]
			return "OK"
		} else {
			return "Error: " + req["key"] + " already saved"
		}	
	} else if req["method"] == "PUT" {
		if len(req) < 3 {
			return "Error: Put requires 2 arguments"
		}

		if (keyValue[req["key"]] != "") {
			keyValue[req["key"]] = req["value"]
			return "Ok"
		} else {
			return "Error: " + req["key"] + " not found"
		}	
	} else if req["method"] == "DELETE" {
		if len(req) < 2 {
			return "Error: Delete requires 1 arguments"
		}

		if (keyValue[req["key"]] != "") {
			delete(keyValue, req["key"])
			return "Ok"
		} else {
			return "Error: " + req["key"] + " not found"
		}
	} else {
		return "Bad request"
	}
}

func formatRequest(req string) map[string]string {
    result := map[string]string{}
	req = strings.TrimSpace(req)
	params := strings.Split(req, " ")
	
	if len(params) < 4 {
		for i, param := range params {
			result[format[i]] = param
		}
	}

	return result
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		err := conn.SetDeadline(time.Now().Add(waitTime))
		if (err!=nil) {
			fmt.Println("Error: setting the deadline for connection")
		}

		request, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Timeout: Resource recovered")
			return
		}

		request = strings.TrimSpace(request)

		if len(request) > 0 {
			go fmt.Fprintln(conn, performTask(formatRequest(request)))
		}
		
		
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err!=nil {
		fmt.Println("405 Server Error: ", err)
		os.Exit(1)
	}
	fmt.Println("Server running on port 8080...")
	defer listener.Close()

	loadData()

	for {
		conn, err := listener.Accept()
		if err!=nil {
			fmt.Fprintf(conn, "405 Server Error: Problem during TCP handshake %v", err)
			continue
		}
		go handleRequest(conn)
	}
}