package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, err := parseRequestLine(bytes)
	if err != nil {
		return nil, err
	}

	request := Request{
		RequestLine: *requestLine,
	}

	return &request, nil
}

func parseRequestLine(data []byte) (*RequestLine, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, fmt.Errorf("could not find CRLF in request-line")
	}
	requestLineString := string(data[:idx])

	parsedRequestLine := strings.Split(requestLineString, " ")
	if len(parsedRequestLine) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", requestLineString)
	}

	method := parsedRequestLine[0]
	for _, r := range method {
		if !unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return nil, fmt.Errorf("Invaild method")
		}
	}

	target := parsedRequestLine[1]

	httpVersion := parsedRequestLine[2]
	validProtocol := strings.HasPrefix(httpVersion, "HTTP")
	if !validProtocol {
		return nil, fmt.Errorf("Invalid protocol")
	}
	validVersion := strings.HasSuffix(httpVersion, "/1.1")
	if !validVersion {
		return nil, fmt.Errorf("Invalid version")
	}
	httpVersion = "1.1"

	requestLine := RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: target,
		Method:        method,
	}

	return &requestLine, nil

}
