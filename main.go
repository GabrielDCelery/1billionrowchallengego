package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sync"
)

const bufferSize = 1024 * 1024

type StationData struct {
	name string
	temp int
}

type StationAggregate struct {
	total int
	min   int
	max   int
	count int
}

func main() {
	var wg sync.WaitGroup
	filePath := "./measurements.txt"
	file, err := os.Open(filePath)

	if err != nil {
		log.Fatalf("failed to open %s, reason: %v", filePath, err)
	}

	defer file.Close()

	sdChan := make(chan StationData)
	sdAggregateMapChan := make(chan map[string]*StationAggregate)

	for range 10 {
		wg.Add(1)
		go worker(sdChan, sdAggregateMapChan, &wg)
	}

	go func() {
		wg.Wait()
		close(sdAggregateMapChan)
	}()

	buffer := make([]byte, bufferSize)

	validBytesInBuffer := 0

	for {
		numOfBytesRead, errRead := file.Read(buffer[validBytesInBuffer:])

		validBytesInBuffer += numOfBytesRead

		numOfBytesParsed, errParse := parse(buffer[:validBytesInBuffer], sdChan)

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

	close(sdChan)

	allStations := map[string]*StationAggregate{}

	for sdAggregate := range sdAggregateMapChan {
		for stationName, aggr := range sdAggregate {
			existing, ok := allStations[stationName]
			if !ok {
				allStations[stationName] = aggr
			} else {
				existing.total += aggr.total
				existing.count += aggr.count
				if aggr.min < existing.min {
					existing.min = aggr.min
				}
				if aggr.max > existing.max {
					existing.max = aggr.max
				}
			}
		}
	}

	for name, agg := range allStations {
		avgTmp := float64(agg.total) / float64(agg.count) / 10
		minTmp := float64(agg.min) / 10
		maxTmp := float64(agg.max) / 10
		fmt.Printf("%s=%.1f/%.1f/%.1f\n", name, minTmp, avgTmp, maxTmp)
	}
}

func parse(chunk []byte, sdChan chan StationData) (int, error) {
	numOfBytesParsed := 0
	for {
		lineEnd, isCompleteLine := findNextCRLF(chunk[numOfBytesParsed:])
		if !isCompleteLine {
			return numOfBytesParsed, nil
		}
		statsionData := parseStationData(chunk[numOfBytesParsed : numOfBytesParsed+lineEnd])
		sdChan <- statsionData
		numOfBytesParsed += lineEnd + 1
	}
}

func findNextCRLF(chunk []byte) (int, bool) {
	i := bytes.Index(chunk, []byte("\n"))
	return i, i != -1
}

func worker(sdChan chan StationData, sdAggregateMapChan chan map[string]*StationAggregate, wg *sync.WaitGroup) {
	stationAggregateMap := map[string]*StationAggregate{}
	defer wg.Done()
	for sd := range sdChan {
		v, ok := stationAggregateMap[sd.name]
		if !ok {
			v = &StationAggregate{
				total: 0,
				count: 0,
				min:   math.MaxInt,
				max:   math.MinInt,
			}
			stationAggregateMap[sd.name] = v
		}
		v.total += sd.temp
		v.count += 1
		if sd.temp < v.min {
			v.min = sd.temp
		}
		if sd.temp > v.max {
			v.max = sd.temp
		}
	}
	sdAggregateMapChan <- stationAggregateMap
}

func parseStationData(chunk []byte) StationData {
	i := bytes.Index(chunk, []byte(";"))
	sd := StationData{name: string(chunk[:i]), temp: parseTempAsInt(chunk[i+1:])}
	return sd
}

func parseTempAsInt(chunk []byte) int {
	total := 0
	decimal := 1
	for i := range chunk {
		char := chunk[len(chunk)-i-1]
		if char == '.' {
			continue
		}
		if char == '-' {
			total = -1 * total
			break
		}
		total += int(char-'0') * decimal
		decimal *= 10
	}
	return total
}
