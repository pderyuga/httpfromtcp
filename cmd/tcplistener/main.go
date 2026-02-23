package main

import (
	"fmt"
	"log"
	"net"

	"github.com/pderyuga/httpfromtcp/internal/request"
)

const port = "42069"

func main() {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer listener.Close()

	fmt.Printf("Listening on port :%s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Accepted connection from", conn.RemoteAddr())

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error parsing request: %s\n", err.Error())
		}
		fmt.Println("Request line:")
		fmt.Printf(" - Method: %s\n", request.RequestLine.Method)
		fmt.Printf(" - Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf(" - Version: %s\n", request.RequestLine.HttpVersion)

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
