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
	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

func Ping(ctx context.ServerContext, conn net.Conn) {
	response.RespondNoContent(conn, response.StatusOK)
}

func Echo(ctx context.ServerContext, conn net.Conn) {
	msg, err := ctx.GetParam("message")
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	response.Respond(conn, response.StatusOK, response.ContentTypePlainText, []byte(msg))
}

func UserAgent(ctx context.ServerContext, conn net.Conn) {
	userAgent, err := ctx.GetUserAgent()
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	response.Respond(conn, response.StatusOK, response.ContentTypePlainText, []byte(userAgent))
}

func DoesFileExist(ctx context.ServerContext, conn net.Conn) {
	if len(os.Args) != 3 || os.Args[1] != "--directory" {
		response.RespondError(conn, response.StatusInternalServerError)
	}

	dir := os.Args[2]
	filename, err := ctx.GetParam("filename")
	if err != nil {
		response.RespondError(conn, response.StatusInternalServerError)
	}

	f, err := os.Open(fmt.Sprintf("%s%s", dir, filename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT) {
			response.RespondError(conn, response.StatusNotFound)
		}
		response.RespondError(conn, response.StatusInternalServerError)
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}(f)

	data, err := io.ReadAll(f)
	if err != nil {
		response.RespondError(conn, response.StatusInternalServerError)
	}

	response.Respond(conn, response.StatusOK, response.ContentTypeOctetStream, data)
}

func WriteFile(ctx context.ServerContext, conn net.Conn) {
	if len(os.Args) != 3 || os.Args[1] != "--directory" {
		response.RespondError(conn, response.StatusInternalServerError)
	}

	dir := os.Args[2]
	filename, err := ctx.GetParam("filename")
	if err != nil {
		response.RespondError(conn, response.StatusInternalServerError)
	}

	body, err := ctx.GetRequestBody()
	if err != nil {
		response.RespondError(conn, response.StatusBadRequest)
	}

	if err := os.WriteFile(dir+filename, []byte(body), 0644); err != nil {
		response.RespondError(conn, response.StatusInternalServerError)
	}
	response.RespondNoContent(conn, response.StatusCreated)
}
