package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

var resultsByName = map[string]benchResult{}

var (
	resultType = flag.String("type", "integration", "the type of benchmark results being parsed ('unit' or 'integration')")
)

type benchResult struct {
	name          string
	typ           string
	queryTimes    []time.Duration
	pipelineTimes []time.Duration
}

func (b benchResult) perfJSON() string {
	values := []float64{}
	var average float64

	switch b.typ {
	case "queries", "unit":
		var sum float64
		for _, dur := range b.queryTimes {
			ms := -1000 * dur.Seconds()
			values = append(values, ms)
			sum += ms
		}
		average = sum / float64(len(values))

	case "overhead":
		var sum float64
		for idx, qt := range b.queryTimes {
			pt := b.pipelineTimes[idx]
			percentage := float64(100*(qt-pt)) / float64(pt)
			values = append(values, percentage)
			sum += percentage
		}
		average = sum / float64(len(values))
	}

	perf := fmt.Sprintf(`    { "name": %q, "results": { "1": { "ops_per_sec": %f, "ops_per_sec_values": [ `, b.name, average)
	for idx, val := range values {
		perf += fmt.Sprintf("%f", val)
		if idx < len(values)-1 {
			perf += ", "
		}
	}
	perf += ` ] } } }`

	return perf
}

func processIntegration(fields []string) {
	// the full name of the benchmark is the first field
	bench := fields[0]

	// split the subtest path
	benchParts := strings.Split(bench, "/")
	benchType := benchParts[1]
	benchName := benchParts[2]

	ns, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not convert latency to int: %v\n", err)
		os.Exit(1)
	}
	dur := time.Duration(ns)

	// fetch existing result for this name, or create a new one
	res, ok := resultsByName[benchName]
	if !ok {
		res = benchResult{
			name:          benchName,
			typ:           benchType,
			queryTimes:    []time.Duration{},
			pipelineTimes: []time.Duration{},
		}
	}

	// append the latency to the appropriate field of the result
	switch benchType {
	case "queries":
		res.queryTimes = append(res.queryTimes, dur)
	case "overhead":
		overheadComponent := benchParts[3]
		switch overheadComponent {
		case "query":
			res.queryTimes = append(res.queryTimes, dur)
		case "pipeline":
			res.pipelineTimes = append(res.pipelineTimes, dur)
		}
	}

	// re-insert the updated result into the map
	resultsByName[benchName] = res
}

func processUnit(fields []string) {
	// the full name of the benchmark is the first field
	benchName := fields[0]

	ns, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not convert latency to int: %v\n", err)
		os.Exit(1)
	}
	dur := time.Duration(ns)

	// fetch existing result for this name, or create a new one
	res, ok := resultsByName[benchName]
	if !ok {
		res = benchResult{
			name:       benchName,
			typ:        "unit",
			queryTimes: []time.Duration{},
		}
	}

	// include this measurement in the result
	res.queryTimes = append(res.queryTimes, dur)

	// re-insert the updated result into the map
	resultsByName[benchName] = res
}

func processLine(line string) {
	// split the line into fields
	fields := strings.Fields(line)

	name := fields[0]

	// ignore this line if the field isn't a test name
	if len(name) < 9 || name[:9] != "Benchmark" {
		return
	}

	// remove the "-<num_threads>" suffix from the benchmark name
	split := strings.Split(name, "-")
	fields[0] = strings.Join(split[:len(split)-1], "-")

	switch *resultType {
	case "unit":
		processUnit(fields)
	case "integration":
		processIntegration(fields)
	default:
		fmt.Fprintf(os.Stderr, "unknown result type %s\n", *resultType)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	fileName := "benchmarks.out"
	if *resultType == "unit" {
		fileName = "benchmarks-unit.out"
	}

	fileBytes, err := ioutil.ReadFile("testdata/artifacts/out/" + fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read benchmarks file: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(fileBytes))

	for scanner.Scan() {
		processLine(string(scanner.Bytes()))
	}

	printResults()
}

func printResults() {
	fmt.Println(`{"results": [`)
	var idx int
	for _, res := range resultsByName {
		fmt.Printf("%s", res.perfJSON())
		if idx < len(resultsByName)-1 {
			fmt.Printf(",")
		}
		fmt.Println()
		idx++
	}
	fmt.Println(`]}`)
}
