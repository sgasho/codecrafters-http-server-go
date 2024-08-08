package context

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

const (
	headersPrefix = "headers"
	paramsPrefix  = "params"
)

type ServerContext interface {
	SetParam(key, value string)
	GetParam(param string) (string, error)
	SetUserAgent(userAgent string)
	GetUserAgent() (string, error)
	SetContentType(contentType response.ContentType)
	GetContentType() (response.ContentType, error)
	SetContentLength(length int)
	GetContentLength() (int, error)
	SetRequestBody(body string)
	GetRequestBody() (string, error)
}

type serverContext map[string]string

func (s serverContext) SetRequestBody(body string) {
	s["request-body"] = body
}

func (s serverContext) GetRequestBody() (string, error) {
	body, exists := s["request-body"]
	if !exists {
		return "", errors.New("request-body not set")
	}
	return body, nil
}

func (s serverContext) SetContentType(contentType response.ContentType) {
	s[fmt.Sprintf("%s.%s", headersPrefix, "content-type")] = string(contentType)
}

func (s serverContext) GetContentType() (response.ContentType, error) {
	value, exists := s[fmt.Sprintf("%s.content-type", headersPrefix)]
	if !exists {
		return "", errors.New("content-type not found")
	}
	return response.ContentType(value), nil
}

func (s serverContext) SetContentLength(length int) {
	s[fmt.Sprintf("%s.%s", headersPrefix, "content-length")] = strconv.Itoa(length)
}

func (s serverContext) GetContentLength() (int, error) {
	value, exists := s[fmt.Sprintf("%s.content-length", headersPrefix)]
	if !exists {
		return 0, errors.New("content-length not found")
	}
	length, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return length, nil
}

func Background() ServerContext {
	return serverContext{}
}

func (s serverContext) SetParam(key, value string) {
	s[fmt.Sprintf("%s.%s", paramsPrefix, key)] = value
}

func (s serverContext) GetParam(param string) (string, error) {
	value, exists := s[fmt.Sprintf("%s.%s", paramsPrefix, param)]
	if !exists {
		return "", fmt.Errorf("param %s not found", param)
	}
	return value, nil
}

func (s serverContext) SetUserAgent(userAgent string) {
	s[fmt.Sprintf("%s.%s", headersPrefix, "user-agent")] = userAgent
}

func (s serverContext) GetUserAgent() (string, error) {
	value, exists := s[fmt.Sprintf("%s.user-agent", headersPrefix)]
	if !exists {
		return "", errors.New("user-agent not found")
	}
	return value, nil
}
