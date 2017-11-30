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

	first := true

	for scanner.Scan() {
		line := string(scanner.Bytes())
		fields := strings.Fields(line)

		bench := fields[0]
		benchParts := strings.Split(bench, "/")
		if benchParts[0] != "BenchmarkIntegration" {
			continue
		}
		benchType := benchParts[1]
		benchName := benchParts[2]

		ns, err := strconv.Atoi(fields[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not convert latency to int: %v\n", err)
			os.Exit(1)
		}

		switch benchType {
		case "queries":
			if !first {
				fmt.Println(",")
			}
			first = false
			name := strings.Split(benchName, "-")[0]

			// divide the latency by 1000000 to get it in ms instead of ns.
			// multiply it by -1 so that we respect the "higher numbers are better"
			// assumption made by the evergreen perf module.
			printResult(name, ns/-1000000)

		}
	}
}

func printResult(name string, num int) {
	fmt.Printf(`    { "name": %q, "results": { "1": { "ops_per_sec": %d } } }`, name, num)
}
