package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/pderyuga/httpfromtcp/internal/headers"
)

type State int

const (
	requestStateInitialized State = iota
	requestStateParsingHeaders
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       State
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state:   requestStateInitialized,
		Headers: make(headers.Headers),
	}
	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.state, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineString := string(data[:idx])

	requestLine, err := requestLineFromString(requestLineString)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parsedRequestLine := strings.Split(str, " ")
	if len(parsedRequestLine) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parsedRequestLine[0]
	for _, r := range method {
		if !unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return nil, fmt.Errorf("Invaild method")
		}
	}

	target := parsedRequestLine[1]

	httpVersion := parsedRequestLine[2]
	versionParts := strings.Split(httpVersion, "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}
	requestLine := RequestLine{
		HttpVersion:   version,
		RequestTarget: target,
		Method:        method,
	}

	return &requestLine, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		numBytes, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return totalBytesParsed, nil
		}
		totalBytesParsed += numBytes
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return numBytes, nil
	case requestStateParsingHeaders:
		numBytes, parsingHeadersDone, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if parsingHeadersDone {
			r.state = requestStateDone
		}
		return numBytes, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}
