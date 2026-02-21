package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const filePath = "messages.txt"

func main() {
	file, err := os.Open(filePath)
	if err != nil {
		// Log the error and exit if the file cannot be read
		log.Fatalf("Unable to open %s: %v", filePath, err)
	}
	defer file.Close()

	channel := getLinesChannel(file)

	fmt.Printf("Reading data from %s\n", filePath)
	fmt.Println("=====================================")

	for line := range channel {
		fmt.Printf("read: %s\n", line)
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
			if err != nil && err.Error() != "EOF" { // Check for actual errors, not just end-of-file
				log.Fatal(err)
			}
			if err != nil && err.Error() == "EOF" {
				break
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
