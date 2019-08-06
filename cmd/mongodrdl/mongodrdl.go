// Main package for the mongodrdl tool.
package main

import (
	"fmt"
	"os"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/mongodrdl"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
)

func main() {
	opts, err := mongodrdl.NewDrdlOptions()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error generating command line options: %v\n", err)
		os.Exit(procutil.ExitError)
	}

	err = opts.Parse(os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error parsing command line options: %v\n", err)
		_, _ = fmt.Fprintln(os.Stderr, "try 'mongodrdl --help' for more information")
		os.Exit(procutil.ExitBadOptions)
	}

	if opts.Version {
		config.PrintVersionAndGitspec("mongodrdl", os.Stdout)
		os.Exit(procutil.ExitClean)
	}

	// print help, if specified
	if opts.Help {
		fmt.Println(opts.HelpText())
		os.Exit(procutil.ExitClean)
	}

	err = opts.Validate()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\ntry 'mongodrdl --help' for more information", err)
		os.Exit(procutil.ExitBadOptions)
	}

	//// connect directly, unless a replica set name is explicitly specified
	//_, setName := procutil.ParseConnectionString(opts.Host)
	//opts.ReplicaSetName = setName

	verbosity := opts.DrdlLog.Level()
	if opts.DrdlLog.Quiet {
		verbosity = log.Quiet
	}
	log.SetVerbosity(verbosity)

	err = opts.Run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		if err == procutil.ErrTerminated {
			os.Exit(procutil.ExitKill)
		}
		os.Exit(procutil.ExitError)
	}
}
