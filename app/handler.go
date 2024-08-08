package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"

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
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
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
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	responseBody := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	if _, err := conn.Write([]byte(responseBody)); err != nil {
		log.Fatal(err)
	}
}

func DoesFileExist(ctx context.ServerContext, conn net.Conn) {
	if len(os.Args) != 3 || os.Args[1] != "--directory" {
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	dir := os.Args[2]
	filename, err := ctx.GetParam("filename")
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	f, err := os.Open(fmt.Sprintf("%s%s", dir, filename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT) {
			if _, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n")); err != nil {
				log.Println(err)
			}
		}
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}(f)

	data, err := io.ReadAll(f)
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	responseBody := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), string(data))
	if _, err := conn.Write([]byte(responseBody)); err != nil {
		log.Fatal(err)
	}
}
