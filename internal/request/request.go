package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type State int

const (
	initialized State = iota
	done
)

type Request struct {
	RequestLine RequestLine
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
		state: initialized,
	}
	for req.state != done {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != done {
					return nil, fmt.Errorf("incomplete request")
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

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case initialized:
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = done
		return numBytes, nil
	case done:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}

}
