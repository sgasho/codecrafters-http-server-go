package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"net"
	"strings"
)

type Status string

const (
	StatusOK                  Status = "200 OK"
	StatusCreated             Status = "201 Created"
	StatusBadRequest          Status = "400 Bad Request"
	StatusNotFound            Status = "404 Not Found"
	StatusInternalServerError Status = "500 Internal Server Error"
)

func responseHeaderString(contentType ContentType, encoding Encoding, data []byte) string {
	headerStrings := make([]string, 0)
	if contentType != "" {
		headerStrings = append(headerStrings, fmt.Sprintf("Content-Type: %s", contentType))
	}
	if encoding == EncodingGZip {
		headerStrings = append(headerStrings, fmt.Sprintf("Content-Encoding: %s", encoding))
	}
	headerStrings = append(headerStrings, fmt.Sprintf("Content-Length: %d", len(data)))
	return strings.Join(headerStrings, "\r\n")
}

func Respond(ctx ServerContext, conn net.Conn, status Status, data []byte) {
	contentType, err := ctx.GetContentType()
	if err != nil {
		log.Println(err)
	}
	encoding, err := ctx.GetEncoding()
	if err != nil {
		log.Println(err)
	}

	responseBody := fmt.Sprintf(
		"%s %s\r\n%s\r\n\r\n%s",
		Version, status, responseHeaderString(contentType, encoding, data), string(data),
	)

	if encoding == EncodingGZip {
		compressed, err := compressToGzip(data)
		if err != nil {
			RespondError(conn, StatusInternalServerError)
		}
		responseBody = fmt.Sprintf(
			"%s %s\r\n%s\r\n\r\n%s",
			Version, status, responseHeaderString(contentType, encoding, compressed), compressed,
		)
	}

	if _, err := conn.Write([]byte(responseBody)); err != nil {
		RespondError(conn, StatusInternalServerError)
	}
}

func RespondNoContent(conn net.Conn, status Status) {
	if _, err := conn.Write([]byte(fmt.Sprintf("%s %s\r\n\r\n", Version, status))); err != nil {
		log.Fatal(err)
	}
}

func RespondError(conn net.Conn, status Status) {
	if _, err := conn.Write([]byte(fmt.Sprintf("%s %s\r\n\r\n", Version, status))); err != nil {
		log.Fatal(err)
	}
}

func compressToGzip(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)

	if _, err := w.Write(input); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
