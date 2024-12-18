package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var reader = bufio.NewReader(os.Stdin)

func sendPing(conn net.Conn) bool {
	isActive := true
	_, err := fmt.Fprintln(conn, " ")

	if err!=nil {
		isActive = false
	}

	return isActive
}

func reconnect() net.Conn{
	fmt.Print("Would you like reconnect (y/n): ")
	yn, err := reader.ReadString('\n')
	if err!=nil {
		fmt.Println("Error: reading your answer")
		reconnect()
	}

	yn = strings.TrimSpace(yn)

	if yn == "y" {
		fmt.Println("Reconnecting...")

		newconn, err := net.Dial("tcp", "localhost:8080")
        if err != nil {
            fmt.Println("Error: creating the connection")
            reconnect()
        }
       
        go readAndWrite(newconn)
        return newconn
	} else {
		fmt.Println("Exiting program")
		os.Exit(1)
	}

    return nil
}

func readAndWrite(conn net.Conn) {
	readConn := bufio.NewReader(conn)
	for {
		message, err := readConn.ReadString('\n')
		if (err!=nil) {
			fmt.Println("\nError: reading, connection broken (press enter twice)")
			return
		}

		fmt.Println("Server: ", message)
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err!=nil {
		fmt.Println("Error: creating the connection")
		os.Exit(1)
	}
	defer conn.Close()

	go readAndWrite(conn)

	for {
		fmt.Print("Enter a command: ")

		command, err := reader.ReadString('\n')
		if (err!=nil) {
			fmt.Println("Error: reading your input")
			continue
		}

		if sendPing(conn) {
			command = strings.TrimSpace(command)
			if len(command) > 0 {
				fmt.Fprintln(conn, command)
			}
		} else {
			conn = reconnect()
		}
		

		
		time.Sleep(time.Duration(1)*time.Second)
	}
}