package main

import (
	"fmt"
	"log"
	"os"
)

const filePath = "messages.txt"

func main() {
	file, err := os.Open(filePath)
	if err != nil {
		// Log the error and exit if the file cannot be read
		log.Fatalf("Unable to open %s: %v", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}

	fileSize := fileInfo.Size()

	fmt.Printf("Reading data from %s\n", filePath)
	fmt.Println("=====================================")

	for fileSize > 0 {
		bufferSize := 8
		buffer := make([]byte, bufferSize)
		count, err := file.Read(buffer)
		if err != nil && err.Error() != "EOF" { // Check for actual errors, not just end-of-file
			log.Fatal(err)
		}
		if err != nil && err.Error() == "EOF" {
			os.Exit(0)
		}
		fileSize -= int64(bufferSize)

		fmt.Printf("read: %s\n", string(buffer[:count]))
	}
}
