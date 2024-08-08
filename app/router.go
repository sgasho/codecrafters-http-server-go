package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const Version = "HTTP/1.1"

type ContentType string

const (
	ContentTypePlainText   ContentType = "text/plain"
	ContentTypeOctetStream ContentType = "application/octet-stream"
)

type Method string

const (
	MethodGet  Method = "GET"
	MethodPost Method = "POST"
)

type Endpoint struct {
	Method     Method
	PathRegex  *regexp.Regexp
	ParamNames []string
	Handler    func(ctx ServerContext, conn net.Conn)
}

type Endpoints []*Endpoint

func (es Endpoints) FilterByMethod(method Method) Endpoints {
	selected := make(Endpoints, 0)
	for _, endpoint := range es {
		if endpoint.Method == method {
			selected = append(selected, endpoint)
		}
	}
	return selected
}

type Router struct {
	conn      net.Conn
	Endpoints Endpoints
}

func NewRouter() *Router {
	return &Router{}
}

type Encoding string

const (
	EncodingGZip Encoding = "gzip"
)

type Headers struct {
	Host           string
	UserAgent      string
	Accept         string
	AcceptEncoding Encoding
	ContentType    ContentType
	ContentLength  int
}

type Request struct {
	Headers *RequestHeaders
	Body    string
}

type RequestHeaders struct {
	method   Method
	path     string
	protocol string
	Headers  *Headers
}

func (r *Router) newRequest() (*Request, error) {
	reader := bufio.NewReader(r.conn)
	requestLineAndHeaders := make([]string, 0)
	// request line and headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\r\n" {
			break
		}
		requestLineAndHeaders = append(requestLineAndHeaders, strings.Trim(line, "\r\n"))
	}

	requestLine := requestLineAndHeaders[0]
	re, err := regexp.Compile(fmt.Sprintf(`^(\w+)\s+(/[^\s]*)\s+(HTTP/\d+\.\d+)$`))
	if err != nil {
		return nil, err
	}
	match := re.FindStringSubmatch(requestLine)
	if len(match) < 4 {
		return nil, errors.New("invalid request line format, could not find method")
	}

	hs := &Headers{}
	headers := requestLineAndHeaders[1:]
	for _, header := range headers {
		k, v := strings.Split(header, ":")[0], strings.Split(header, ": ")[1]
		switch k {
		case "Host":
			hs.Host = v
		case "User-Agent":
			hs.UserAgent = v
		case "Accept":
			hs.Accept = v
		case "Content-Type":
			hs.ContentType = ContentType(v)
		case "Content-Length":
			hs.ContentLength, err = strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
		case "Accept-Encoding":
			hs.AcceptEncoding = Encoding(v)
		default:
			log.Printf("parsing method for header key: %s is not implemented", k)
		}
	}

	requestBodyBuf := make([]byte, hs.ContentLength)
	if _, err := reader.Read(requestBodyBuf); err != nil {
		return nil, err
	}

	return &Request{
		Headers: &RequestHeaders{
			method:   Method(match[1]),
			path:     match[2],
			protocol: match[3],
			Headers:  hs,
		},
		Body: string(requestBodyBuf),
	}, nil
}

func (r *Router) Get(path string, handler func(ctx ServerContext, conn net.Conn)) {
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

func (r *Router) Post(path string, handler func(ctx ServerContext, conn net.Conn)) {
	pathRegexStr, paramNames, err := convertPathToRegexAndExtractParamNames(path)
	if err != nil {
		log.Fatal(err)
	}
	pathRegex, err := regexp.Compile(pathRegexStr)
	if err != nil {
		log.Fatal(err)
	}
	r.Endpoints = append(r.Endpoints, &Endpoint{
		Method:     MethodPost,
		PathRegex:  pathRegex,
		ParamNames: paramNames,
		Handler:    handler,
	})
}

func (r *Router) Serve(conn net.Conn) {
	r.conn = conn
	req, err := r.newRequest()
	if err != nil {
		log.Fatal(err)
	}

	for _, endpoint := range r.Endpoints.FilterByMethod(req.Headers.method) {
		if endpoint.PathRegex.MatchString(req.Headers.path) {
			ctx := Background()

			matches := endpoint.PathRegex.FindAllStringSubmatch(req.Headers.path, -1)
			for i, match := range matches {
				if len(match) <= 1 {
					continue
				}
				ctx.SetParam(endpoint.ParamNames[i], match[1])
			}
			ctx.SetUserAgent(req.Headers.Headers.UserAgent)
			ctx.SetEncoding(req.Headers.Headers.AcceptEncoding)
			if req.Headers.method == MethodPost {
				ctx.SetContentType(req.Headers.Headers.ContentType)
				ctx.SetContentLength(req.Headers.Headers.ContentLength)
				ctx.SetRequestBody(req.Body)
			}
			endpoint.Handler(ctx, conn)
			return
		}
	}

	RespondError(conn, StatusNotFound)
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
