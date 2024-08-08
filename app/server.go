package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer func(l net.Listener) {
		if err := l.Close(); err != nil {
			log.Fatal(err)
		}
	}(l)

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	router := NewRouter()
	router.Get("/", Ping)
	router.Get("/echo/{message}", Echo)
	router.Get("/user-agent", UserAgent)
	router.Serve(conn)
}
