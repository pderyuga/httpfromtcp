package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const serverAddr = "localhost:42069"

func main() {
	remoteAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatalf("Error resolving server address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", conn.RemoteAddr())

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		count, err := conn.Write([]byte(input))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Sent %d bytes to %s\n", count, conn.RemoteAddr())
	}

}
