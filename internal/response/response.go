package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pderyuga/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadrequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := ""

	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadrequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := make(headers.Headers)
	defaultHeaders.Set("Content-Length", strconv.Itoa(contentLen))
	defaultHeaders.Set("Connection", "close")
	defaultHeaders.Set("Content-Type", "text/plain")

	return defaultHeaders
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for name, header := range headers {
		headerString := fmt.Sprintf("%s: %s\r\n", name, header)
		_, err := w.Write([]byte(headerString))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}
