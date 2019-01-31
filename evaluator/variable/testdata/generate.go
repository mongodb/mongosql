package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"

	"github.com/10gen/sqlproxy/internal/generate"
)

func main() {
	if len(os.Args) != 3 {
		bail("must provide input and output file arguments")
	}
	inFile := os.Args[1]
	outFile := os.Args[2]

	vars := &generate.VariablesSpec{}
	err := vars.Parse(inFile)
	if err != nil {
		bail(err.Error())
	}

	buf := bytes.NewBuffer([]byte{})
	err = vars.Generate(buf)
	if err != nil {
		bail(err.Error())
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		_ = ioutil.WriteFile(outFile, buf.Bytes(), os.ModePerm)
		bail(err.Error())
	}

	err = ioutil.WriteFile(outFile, formatted, os.ModePerm)
	if err != nil {
		bail(err.Error())
	}
}

func bail(err string) {
	_, _ = fmt.Fprintf(os.Stderr, "fatal error: %s\n", err)
	os.Exit(1)
}
