package response

import (
	"bytes"
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

type WriterState int

const (
	WritingStatusLine WriterState = iota
	WritingHeaders
	WritingBody
)

type Writer struct {
	Writer       io.Writer
	WriterState  WriterState
	BytesWritten int
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterState != WritingStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.WriterState)
	}

	statusLine := GetStatusLine(statusCode)
	_, err := w.Writer.Write(statusLine)
	if err != nil {
		return err
	}
	w.WriterState = WritingHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.WriterState != WritingHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.WriterState)
	}

	var buf bytes.Buffer

	for name, header := range headers {
		headerString := fmt.Sprintf("%s: %s\r\n", name, header)
		buf.WriteString(headerString)
	}
	buf.WriteString("\r\n")
	_, err := w.Writer.Write(buf.Bytes())
	if err != nil {
		return err
	}
	w.WriterState = WritingBody

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriterState != WritingBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.WriterState)
	}

	bytesWritten, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	w.BytesWritten = bytesWritten
	return bytesWritten, nil
}

func GetStatusLine(statusCode StatusCode) []byte {
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
	return []byte(statusLine)
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := make(headers.Headers)
	defaultHeaders.Set("Content-Length", strconv.Itoa(contentLen))
	defaultHeaders.Set("Connection", "close")
	defaultHeaders.Set("Content-Type", "text/html")

	return defaultHeaders
}
