package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/pderyuga/httpfromtcp/internal/request"
	"github.com/pderyuga/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h HandlerError) Write(w io.Writer) error {
	messageLength := len(h.Message)

	err := response.WriteStatusLine(w, h.StatusCode)
	if err != nil {
		fmt.Println("Error writing status line:", err.Error())
		return err
	}

	defaultHeaders := response.GetDefaultHeaders(messageLength)

	err = response.WriteHeaders(w, defaultHeaders)
	if err != nil {
		fmt.Println("Error writing headers:", err.Error())
		return err
	}

	_, err = w.Write([]byte(h.Message))
	if err != nil {
		fmt.Println("Error writing body:", err.Error())
		return err
	}

	return nil
}

type Server struct {
	handler  Handler
	listener net.Listener
	port     int
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := Server{
		handler:  handler,
		listener: listener,
		port:     port,
	}

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *Server) listen() {
	fmt.Printf("Listening on port :%d\n", s.port)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)

	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Accepted connection from", conn.RemoteAddr())

	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := HandlerError{
			StatusCode: response.StatusBadrequest,
			Message:    err.Error(),
		}
		err = handlerError.Write(conn)
		if err != nil {
			fmt.Println("Error writing handler error:", err.Error())
		}
		return
	}

	buffer := bytes.NewBuffer([]byte{})

	handlerError := s.handler(buffer, req)
	if handlerError != nil {
		err := handlerError.Write(conn)
		if err != nil {
			fmt.Println("Error writing handler error:", err.Error())
		}
		return
	}

	err = response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		fmt.Println("Error writing status line:", err.Error())
		return
	}

	resBytes := buffer.Bytes()

	defaultHeaders := response.GetDefaultHeaders(len(resBytes))

	err = response.WriteHeaders(conn, defaultHeaders)
	if err != nil {
		fmt.Println("Error writing headers:", err.Error())
		return
	}

	_, err = conn.Write(resBytes)
	if err != nil {
		fmt.Println("Error writing body:", err.Error())
		return
	}

	fmt.Printf("Sent %d bytes as response\n", len(resBytes))

	fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
}
