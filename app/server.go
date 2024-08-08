package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	c, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	buf := make([]byte, 1024)
	if _, err := c.Read(buf); err != nil {
		log.Fatal(err)
	}
	headers := strings.Split(string(buf), "\r\n")
	requestLines := strings.Split(strings.Trim(headers[0], "\r\n"), " ")

	if requestLines[1] == "/" {
		if _, err := c.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
			log.Fatal(err)
		}
	} else {
		regex, err := regexp.Compile(`/echo/(\w+)`)
		if err != nil {
			log.Fatal(err)
		}
		echo := regex.FindStringSubmatch(requestLines[1])
		if len(echo) < 2 {
			if _, err := c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n")); err != nil {
				log.Fatal(err)
			}
		}
		responseBody := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo[1]), echo[1])
		if _, err := c.Write([]byte(responseBody)); err != nil {
			log.Fatal(err)
		}
	}
}
