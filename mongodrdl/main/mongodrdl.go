// Main package for the mongodrdl tool.
package main

import (
	"github.com/10gen/sqlproxy/mongodrdl"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/common/util"
	"os"
)

func main() {
	// initialize command-line opts
	opts := options.New("mongodrdl", mongodrdl.Usage, options.EnabledOptions{true, true, true})

	outputOpts := &mongodrdl.OutputOptions{}
	opts.AddOptions(outputOpts)

	args, err := opts.Parse()
	if err != nil {
		log.Logf(log.Always, "error parsing command line options: %v", err)
		log.Logf(log.Always, "try 'mongodrdl --help' for more information")
		os.Exit(util.ExitBadOptions)
	}

	if len(args) > 0 {
		log.Logf(log.Always, "positional arguments not allowed: %v", args)
		log.Logf(log.Always, "try 'mongodrdl --help' for more information")
		os.Exit(util.ExitBadOptions)
	}

	// print help, if specified
	if opts.PrintHelp(false) {
		return
	}

	// init logger
	log.SetVerbosity(opts.Verbosity)

	// connect directly, unless a replica set name is explicitly specified
	_, setName := util.ParseConnectionString(opts.Host)
	opts.Direct = (setName == "")
	opts.ReplicaSetName = setName

	schemaGen := mongodrdl.SchemaGenerator{
		ToolOptions:   opts,
		OutputOptions: outputOpts,
	}

	if err = schemaGen.Init(); err != nil {
		log.Logf(log.Always, "Failed: %v", err)
		os.Exit(util.ExitError)
	}

	_, err = schemaGen.Generate()
	if err != nil {
		log.Logf(log.Always, "Failed: %v", err)
		if err == util.ErrTerminated {
			os.Exit(util.ExitKill)
		}
		os.Exit(util.ExitError)
	}
}
