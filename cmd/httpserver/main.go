package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/pderyuga/httpfromtcp/internal/headers"
	"github.com/pderyuga/httpfromtcp/internal/request"
	"github.com/pderyuga/httpfromtcp/internal/response"
	"github.com/pderyuga/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		route := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
		url := "https://httpbin.org" + route
		fmt.Println("Proxying to", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Error making GET request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		w.WriteStatusLine(response.StatusOK)

		h := response.GetDefaultHeaders(0)
		h.Remove("Content-Length")
		h.Override("Content-Type", "text/plain")
		h.Set("Transfer-Encoding", "chunked")
		h.Set("Tralier", "X-Content-SHA256, X-Content-Length")
		w.WriteHeaders(h)

		fullBody := make([]byte, 0)

		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				fmt.Printf("Received %d bytes:\n%s\n", n, string(buf[:n]))
				w.WriteChunkedBody(buf[:n])
				fullBody = append(fullBody, buf[:n]...)
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				fmt.Println("Error reading response body:", err)
				break
			}
		}
		w.WriteChunkedBodyDone()
		fmt.Println("Stream processed successfully")

		hashedBody := sha256.Sum256(fullBody)
		hashedBodyString := fmt.Sprintf("%x", hashedBody)

		trailers := headers.NewHeaders()
		trailers.Set("X-Content-SHA256", hashedBodyString)
		trailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))
		w.WriteTrailers(trailers)

		return
	}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		w.WriteStatusLine(response.StatusBadrequest)
		body := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		headers := response.GetDefaultHeaders(len(body))
		w.WriteHeaders(headers)
		w.WriteBody(body)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		w.WriteStatusLine(response.StatusInternalServerError)
		body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		headers := response.GetDefaultHeaders(len(body))
		w.WriteHeaders(headers)
		w.WriteBody(body)
		return
	}

	w.WriteStatusLine(response.StatusOK)
	body := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
	headers := response.GetDefaultHeaders(len(body))
	w.WriteHeaders(headers)
	w.WriteBody(body)
}
