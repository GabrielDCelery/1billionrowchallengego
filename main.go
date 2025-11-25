package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const bufferSize = 1024 * 1024

func main() {
	var wg sync.WaitGroup
	filePath := "./measurements.txt"
	file, err := os.Open(filePath)

	if err != nil {
		log.Fatalf("failed to open %s, reason: %v", filePath, err)
	}

	defer file.Close()

	lines := make(chan []byte)

	for range 10 {
		go worker(lines, &wg)
	}

	buffer := make([]byte, bufferSize)

	validBytesInBuffer := 0

	for {
		numOfBytesRead, errRead := file.Read(buffer[validBytesInBuffer:])

		validBytesInBuffer += numOfBytesRead

		numOfBytesParsed, errParse := parse(buffer[:validBytesInBuffer], lines)

		if errParse != nil {
			log.Fatalf("failed to parse")
		}

		if numOfBytesParsed > 0 {
			copy(buffer, buffer[numOfBytesParsed:validBytesInBuffer])
			validBytesInBuffer -= numOfBytesParsed
		}

		if numOfBytesParsed == 0 && validBytesInBuffer >= bufferSize {
			log.Fatalf("line exceeds buffer size of %d bytes", bufferSize)
		}

		if errRead == io.EOF {
			break
		}

		if errRead != nil {
			log.Fatalf("failed to read file, reason: %v", errRead)
		}
	}

	wg.Wait()
}

func parse(chunk []byte, lines chan []byte) (int, error) {
	numOfBytesParsed := 0
	for {
		lineEnd, isCompleteLine := findNextCRLF(chunk[numOfBytesParsed:])
		if !isCompleteLine {
			return numOfBytesParsed, nil
		}
		line := make([]byte, 106)
		copy(line, chunk[numOfBytesParsed:numOfBytesParsed+lineEnd])
		lines <- line
		numOfBytesParsed += lineEnd + 1
	}
}

func findNextCRLF(chunk []byte) (int, bool) {
	i := bytes.Index(chunk, []byte("\n"))
	return i, i != -1
}

func worker(lines chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range lines {
		fmt.Println(string(line))
	}
}
