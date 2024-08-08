package main

import (
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/http-server-starter-go/app/context"
)

func Ping(ctx context.ServerContext, conn net.Conn) {
	if _, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		log.Println(err)
	}
}

func Echo(ctx context.ServerContext, conn net.Conn) {
	msg, err := ctx.GetParam("message")
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	responseBody := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(msg), msg)
	if _, err := conn.Write([]byte(responseBody)); err != nil {
		log.Fatal(err)
	}
}

func UserAgent(ctx context.ServerContext, conn net.Conn) {
	userAgent, err := ctx.GetUserAgent()
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	responseBody := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	if _, err := conn.Write([]byte(responseBody)); err != nil {
		log.Fatal(err)
	}
}
