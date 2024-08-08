package context

import (
	"errors"
	"fmt"
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
}

type serverContext map[string]string

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

func Background() ServerContext {
	return serverContext{}
}
