package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

const memoryLimit = 400.0 // The limit is actually 500 MB, but we try to stay under 400 MB.

func TestMemoryLimits(t *testing.T) {

	var tests = []string{
		"control",
		"large-documents",
		"deeply-nested-sub-documents",
		"many-fields",
		"many-databases",
		"many-collections",
		"many-documents",
		"many-arrays",
		"deeply-nested-arrays",
		"describe-table",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			maxMemUsed, err := getMaxMemory(test, t)
			if err != nil {
				t.Fatalf("failed to get max memory from gc trace: %v", err)
			}
			if maxMemUsed > memoryLimit {
				t.Errorf("Total memory used by mongosqld exceeded limit of %v MB: memory in use reached %v MB", memoryLimit, maxMemUsed)
			} else {
				t.Logf("mongosqld used %v MB of memory", maxMemUsed)
			}
		})
	}
}

func getMaxMemory(testName string, t *testing.T) (float64, error) {

	// Open the file containing the gc trace output for this dataset.
	fullFilePath := "../../artifacts/out/" + testName + ".out"
	inFile, err := os.OpenFile(fullFilePath, os.O_APPEND|os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}

	reader := bufio.NewReader(inFile)

	var line string
	var maxMemory float64 = 0
	var lineCount int
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			if lineCount == 0 {
				return 0, fmt.Errorf("results file was empty")
			}
			break
		}
		lineCount++

		if strings.HasPrefix(line, "gc") {
			parts := strings.Split(line, ",")
			if len(parts) < 3 {
				// Both mongosqld and the go gc write to this file, and
				// sometimes there are races between the two programs. If the
				// data is corrupted, we'd rather lose one of many data points
				// than fail the whole test.
				t.Log("corrupted gc data")
				continue
			}

			memInfo := strings.TrimLeft(parts[2], " ")
			memParts := strings.Split(memInfo, "->")
			if len(memParts) != 3 {
				t.Log("corrupted gc data")
				continue
			}

			// gctrace memory info takes the following form: {Total memory
			// before marking}->{Total memory after marking}->{Live memory}.
			// Since the live memory will always be <= the total memory, we'll
			// only check the total memory stats.
			memStrBefore := strings.TrimRight(memParts[0], "MB")
			memStrAfter := strings.TrimRight(memParts[1], "MB")

			memoryBeforeMarking, err := strconv.ParseFloat(memStrBefore, 32)
			if err != nil {
				return 0, err
			}
			memoryAfterMarking, err := strconv.ParseFloat(memStrAfter, 32)
			if err != nil {
				return 0, err
			}
			if memoryBeforeMarking > maxMemory {
				maxMemory = memoryBeforeMarking
			}
			if memoryAfterMarking > maxMemory {
				maxMemory = memoryAfterMarking
			}
		}
	}

	inFile.Close()
	return maxMemory, nil
}
