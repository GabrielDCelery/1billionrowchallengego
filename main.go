package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const bufferSize = 2048

func main() {
	filePath := "./measurements.txt"
	file, err := os.Open(filePath)

	if err != nil {
		log.Fatalf("failed to open %s, reason: %v", filePath, err)
	}

	defer file.Close()

	buffer := make([]byte, bufferSize)
	validBytesInBuffer := 0

	for {
		numOfBytesRead, errRead := file.Read(buffer[validBytesInBuffer:])

		validBytesInBuffer += numOfBytesRead

		if validBytesInBuffer >= bufferSize {
			log.Fatalf("exceeded buffer size of %d", bufferSize)
		}

		numOfBytesParsed, errParse := parse(buffer[:validBytesInBuffer])

		if errParse != nil {
			log.Fatalf("failed to parse")
		}

		if numOfBytesParsed > 0 {
			copy(buffer, buffer[numOfBytesParsed:validBytesInBuffer])
			validBytesInBuffer -= numOfBytesParsed
		}

		if errRead == io.EOF {
			break
		}

		if errRead != nil {
			log.Fatalf("failed to read file, reason: %v", errRead)
		}
	}
}

func parse(chunk []byte) (int, error) {
	fmt.Println(string(chunk))
	return len(chunk), nil
}
