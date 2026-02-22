package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		channel := getLinesChannel(conn)

		for line := range channel {
			fmt.Println(line)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}

func getLinesChannel(file io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer file.Close()
		defer close(lines)

		currentLine := ""

		for {
			bufferSize := 8
			buffer := make([]byte, bufferSize)
			count, err := file.Read(buffer)
			if err != nil {
				if currentLine != "" {
					lines <- currentLine
				}
				if err.Error() == "EOF" {
					break
				}
				log.Fatal(err)
			}

			str := string(buffer[:count])

			parts := strings.Split(str, "\n")
			for i, part := range parts {
				if i < len(parts)-1 {
					// last chunk before newline - add to current line, print  current line, and reset current line to empty string
					currentLine += part
					lines <- currentLine
					currentLine = ""
				} else {
					// not the last chunk before the newline, add to current line
					currentLine += part
				}
			}
		}
	}()

	return lines

}
