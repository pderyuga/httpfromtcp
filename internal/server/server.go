package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/pderyuga/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	port     int
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := Server{
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

	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		fmt.Println("Error writing status line:", err.Error())
		return
	}

	defaultHeaders := response.GetDefaultHeaders(0)

	err = response.WriteHeaders(conn, defaultHeaders)
	if err != nil {
		fmt.Println("Error writing headers:", err.Error())
		return
	}
	fmt.Println("Response sent!")

	fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
}
