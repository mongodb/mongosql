// Main package for the mongodrdl tool.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodrdl"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/util"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	opts, err := options.NewDrdlOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating command line options: %v\n", err)
		os.Exit(util.ExitError)
	}

	args, err := opts.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing command line options: %v\n", err)
		fmt.Fprintln(os.Stderr, "try 'mongodrdl --help' for more information")
		os.Exit(util.ExitBadOptions)
	}

	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "positional arguments not allowed: %v\n", args)
		fmt.Fprintln(os.Stderr, "try 'mongodrdl --help' for more information")
		os.Exit(util.ExitBadOptions)
	}

	log.SetVerbosity(opts.DrdlLog)

	if opts.Version {
		common.PrintVersionAndGitspec("mongodrdl", os.Stdout)
		os.Exit(util.ExitClean)
	}

	// print help, if specified
	if opts.PrintHelp(false) {
		os.Exit(util.ExitClean)
	}

	if err = opts.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintln(os.Stderr, "try 'mongodrdl --help' for more information")
		os.Exit(util.ExitBadOptions)
	}

	// connect directly, unless a replica set name is explicitly specified
	_, setName := util.ParseConnectionString(opts.Host)
	opts.Direct = (setName == "")
	opts.ReplicaSetName = setName

	schemaGen := mongodrdl.SchemaGenerator{
		ToolOptions:   opts,
		OutputOptions: opts.DrdlOutput,
		SampleOptions: opts.DrdlSample,
	}

	if err = schemaGen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(util.ExitError)
	}

	_, err = schemaGen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		if err == util.ErrTerminated {
			os.Exit(util.ExitKill)
		}
		os.Exit(util.ExitError)
	}
}
