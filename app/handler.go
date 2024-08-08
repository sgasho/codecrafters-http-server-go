package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"
)

func Ping(ctx ServerContext, conn net.Conn) {
	RespondNoContent(conn, StatusOK)
}

func Echo(ctx ServerContext, conn net.Conn) {
	msg, err := ctx.GetParam("message")
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	ctx.SetContentType(ContentTypePlainText)
	Respond(ctx, conn, StatusOK, []byte(msg))
}

func UserAgent(ctx ServerContext, conn net.Conn) {
	userAgent, err := ctx.GetUserAgent()
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
			log.Println(err)
		}
	}

	ctx.SetContentType(ContentTypePlainText)
	Respond(ctx, conn, StatusOK, []byte(userAgent))
}

func GetFile(ctx ServerContext, conn net.Conn) {
	if len(os.Args) != 3 || os.Args[1] != "--directory" {
		RespondError(conn, StatusInternalServerError)
	}

	dir := os.Args[2]
	filename, err := ctx.GetParam("filename")
	if err != nil {
		RespondError(conn, StatusInternalServerError)
	}

	f, err := os.Open(fmt.Sprintf("%s%s", dir, filename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT) {
			RespondError(conn, StatusNotFound)
		}
		RespondError(conn, StatusInternalServerError)
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}(f)

	data, err := io.ReadAll(f)
	if err != nil {
		RespondError(conn, StatusInternalServerError)
	}

	ctx.SetContentType(ContentTypeOctetStream)
	Respond(ctx, conn, StatusOK, data)
}

func WriteFile(ctx ServerContext, conn net.Conn) {
	if len(os.Args) != 3 || os.Args[1] != "--directory" {
		RespondError(conn, StatusInternalServerError)
	}

	dir := os.Args[2]
	filename, err := ctx.GetParam("filename")
	if err != nil {
		RespondError(conn, StatusInternalServerError)
	}

	body, err := ctx.GetRequestBody()
	if err != nil {
		RespondError(conn, StatusBadRequest)
	}

	if err := os.WriteFile(dir+filename, []byte(body), 0644); err != nil {
		RespondError(conn, StatusInternalServerError)
	}

	RespondNoContent(conn, StatusCreated)
}
