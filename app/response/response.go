package response

import (
	"fmt"
	"log"
	"net"
)

const Version = "HTTP/1.1"

type ContentType string

const (
	ContentTypePlainText   ContentType = "text/plain"
	ContentTypeOctetStream ContentType = "application/octet-stream"
)

type Status string

const (
	StatusOK                  Status = "200 OK"
	StatusCreated             Status = "201 Created"
	StatusBadRequest          Status = "400 Bad Request"
	StatusNotFound            Status = "404 Not Found"
	StatusInternalServerError Status = "500 Internal Server Error"
)

func responseHeaderString(contentType ContentType, data []byte) string {
	return fmt.Sprintf("Content-Type: %s\r\nContent-Length: %d", string(contentType), len(data))
}

func Respond(conn net.Conn, status Status, contentType ContentType, data []byte) {
	responseBody := fmt.Sprintf(
		"%s %s\r\n%s\r\n\r\n%s",
		Version, status, responseHeaderString(contentType, data), string(data),
	)
	if _, err := conn.Write([]byte(responseBody)); err != nil {
		log.Fatal(err)
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
