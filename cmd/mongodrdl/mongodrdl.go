// Main package for the mongodrdl tool.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/mongodrdl"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	opts, err := mongodrdl.NewDrdlOptions()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error generating command line options: %v\n", err)
		os.Exit(procutil.ExitError)
	}

	args, err := opts.Parse()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error parsing command line options: %v\n", err)
		_, _ = fmt.Fprintln(os.Stderr, "try 'mongodrdl --help' for more information")
		os.Exit(procutil.ExitBadOptions)
	}

	if len(args) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "positional arguments not allowed: %v\n", args)
		_, _ = fmt.Fprintln(os.Stderr, "try 'mongodrdl --help' for more information")
		os.Exit(procutil.ExitBadOptions)
	}

	if opts.Version {
		config.PrintVersionAndGitspec("mongodrdl", os.Stdout)
		os.Exit(procutil.ExitClean)
	}

	// print help, if specified
	if opts.PrintHelp(false) {
		os.Exit(procutil.ExitClean)
	}

	if err = opts.Validate(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		_, _ = fmt.Fprintln(os.Stderr, "try 'mongodrdl --help' for more information")
		os.Exit(procutil.ExitBadOptions)
	}

	// connect directly, unless a replica set name is explicitly specified
	_, setName := procutil.ParseConnectionString(opts.Host)
	opts.ReplicaSetName = setName

	verbosity := opts.DrdlLog.Level()
	if opts.DrdlLog.Quiet {
		verbosity = log.Quiet
	}
	log.SetVerbosity(verbosity)

	lg := log.NewComponentLogger(
		fmt.Sprintf("%-10v [schemaGeneration]", log.MongodrdlComponent),
		log.GlobalLogger(),
	)

	ctx := context.Background()

	err = mongodrdl.GenerateSchema(ctx, lg, *opts)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		if err == procutil.ErrTerminated {
			os.Exit(procutil.ExitKill)
		}
		os.Exit(procutil.ExitError)
	}
}
