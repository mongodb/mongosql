package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func main() {
	fileBytes, err := ioutil.ReadFile("testdata/artifacts/out/benchmarks.out")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read benchmarks file: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(fileBytes))

	var overheadComponent string
	var overheadNs float64
	first := true

	for scanner.Scan() {
		line := string(scanner.Bytes())
		fields := strings.Fields(line)

		// the full name of the benchmark is the first field
		bench := fields[0]

		// remove the "-<num_threads>" suffix from the benchmark name
		split := strings.Split(bench, "-")
		bench = strings.Join(split[:len(split)-1], "-")

		// split the subtest path, and ignore this line if the field isn't a test name
		benchParts := strings.Split(bench, "/")
		if benchParts[0] != "BenchmarkIntegration" {
			continue
		}

		benchType := benchParts[1]
		benchName := benchParts[2]

		ns, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not convert latency to int: %v\n", err)
			os.Exit(1)
		}

		switch benchType {
		case "queries":
			if len(benchParts) != 3 {
				fmt.Fprintf(os.Stderr, "expected subtest path with three components, found %d", len(benchParts))
				os.Exit(1)
			}

			if !first {
				fmt.Println(",")
			}
			first = false

			// divide the latency by 1000000 to get it in ms instead of ns.
			// multiply it by -1 so that we respect the "higher numbers are better"
			// assumption made by the evergreen perf module.
			printResult(benchName, ns/-1000000)

		case "overhead":
			var percentage float64

			if len(benchParts) != 4 {
				fmt.Fprintf(os.Stderr, "expected subtest path with four components, found %d", len(benchParts))
				os.Exit(1)
			}

			switch overheadComponent {
			case "":
				overheadComponent = benchParts[3]
				overheadNs = ns
				continue
			case "query":
				if !first {
					fmt.Println(",")
				}
				first = false
				percentage = 100 * (overheadNs - ns) / ns
				overheadComponent = ""
				overheadNs = 0
			default:
				fmt.Fprintf(os.Stderr, "unexpected benchmark-overhead component %s\n", overheadComponent)
				os.Exit(1)
			}

			printResult(benchName, percentage)
		}
	}
}

func printResult(name string, num float64) {
	fmt.Printf(`    { "name": %q, "results": { "1": { "ops_per_sec": %f } } }`, name, num)
}
