package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/context"
)

type Method string

const (
	MethodGet Method = "GET"
)

type Endpoint struct {
	Method     Method
	PathRegex  *regexp.Regexp
	ParamNames []string
	Handler    func(ctx context.ServerContext, conn net.Conn)
}

type Router struct {
	conn      net.Conn
	Endpoints []*Endpoint
}

func NewRouter() *Router {
	return &Router{}
}

type Headers struct {
	Host      string
	UserAgent string
	Accept    string
}

type RequestHeaders struct {
	method   Method
	path     string
	protocol string
	Headers  *Headers
}

func (r *Router) parseHeaders() (*RequestHeaders, error) {
	buf := make([]byte, 1024)
	if _, err := r.conn.Read(buf); err != nil {
		log.Fatal(err)
	}
	splitByDoubleCRLF := strings.Split(string(buf), "\r\n\r\n")
	requestLineAndHeaders := splitByDoubleCRLF[0]
	requestLineAndHeadersSlice := strings.Split(requestLineAndHeaders, "\r\n")
	requestLine := requestLineAndHeadersSlice[0]
	re, err := regexp.Compile(fmt.Sprintf(`^(\w+)\s+(/[^\s]*)\s+(HTTP/\d+\.\d+)$`))
	if err != nil {
		return nil, err
	}
	match := re.FindStringSubmatch(requestLine)
	if len(match) < 4 {
		return nil, errors.New("invalid request line format, could not find method")
	}

	hs := &Headers{}
	headers := requestLineAndHeadersSlice[1:]
	for _, header := range headers {
		k, v := strings.Split(header, ":")[0], strings.Split(header, ": ")[1]
		switch k {
		case "Host":
			hs.Host = v
		case "User-Agent":
			hs.UserAgent = v
		case "Accept":
			hs.Accept = v
		default:
			return nil, fmt.Errorf("parsing method for header key: %s is not implemented", k)
		}
	}

	return &RequestHeaders{
		method:   Method(match[1]),
		path:     match[2],
		protocol: match[3],
		Headers:  hs,
	}, nil
}

func (r *Router) Get(path string, handler func(ctx context.ServerContext, conn net.Conn)) {
	pathRegexStr, paramNames, err := convertPathToRegexAndExtractParamNames(path)
	if err != nil {
		log.Fatal(err)
	}
	pathRegex, err := regexp.Compile(pathRegexStr)
	if err != nil {
		log.Fatal(err)
	}
	r.Endpoints = append(r.Endpoints, &Endpoint{
		Method:     MethodGet,
		PathRegex:  pathRegex,
		ParamNames: paramNames,
		Handler:    handler,
	})
}

func (r *Router) Serve(conn net.Conn) {
	r.conn = conn
	headers, err := r.parseHeaders()
	if err != nil {
		log.Fatal(err)
	}

	for _, endpoint := range r.Endpoints {
		if endpoint.PathRegex.MatchString(headers.path) {
			ctx := context.Background()

			matches := endpoint.PathRegex.FindAllStringSubmatch(headers.path, -1)
			for i, match := range matches {
				if len(match) <= 1 {
					continue
				}
				ctx.SetParam(endpoint.ParamNames[i], match[1])
			}
			ctx.SetUserAgent(headers.Headers.UserAgent)
			endpoint.Handler(ctx, conn)
			return
		}
	}

	if _, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n")); err != nil {
		log.Println(err)
	}
}

func convertPathToRegexAndExtractParamNames(path string) (string, []string, error) {
	escapedPath := regexp.QuoteMeta(path)

	re, err := regexp.Compile(`\\\{[^/]+\\}`)
	if err != nil {
		return "", nil, err
	}
	regexPath := re.ReplaceAllStringFunc(escapedPath, func(m string) string {
		return "([^/]+)"
	})

	pathRegex := "^" + regexPath + "$"
	re, err = regexp.Compile(pathRegex)
	if err != nil {
		return "", nil, err
	}

	matches := re.FindAllStringSubmatch(path, -1)
	params := make([]string, 0)
	for _, match := range matches {
		if len(match) <= 1 {
			continue
		}
		params = append(params, strings.Trim(strings.Trim(match[1], "{"), "}"))
	}

	return pathRegex, params, nil
}
