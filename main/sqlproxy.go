package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	opts, err := sqlproxy.NewOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating command line options: %v\n", err)
		os.Exit(1)
	}

	err = opts.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing command line options: %v\n", err)
		fmt.Fprintln(os.Stderr, "try 'mongosqld --help' for more information")
		os.Exit(1)
	}

	// set global log level
	log.SetVerbosity(opts.LogOpts)

	if opts.Version {
		common.PrintVersionAndGitspec("mongosqld", os.Stdout)
		os.Exit(0)
	}

	if opts.PrintHelp(os.Stdout) {
		os.Exit(0)
	}

	if len(opts.LogPath) > 0 {
		mode := os.O_WRONLY
		if _, err := os.Stat(opts.LogPath); err != nil {
			mode = mode | os.O_CREATE
		}

		if opts.LogAppend {
			mode = mode | os.O_APPEND
		} else {
			mode = mode | os.O_TRUNC
		}

		logfile, err := os.OpenFile(opts.LogPath, mode, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening log file: %v\n", err)
			os.Exit(1)
		}
		defer logfile.Close()

		log.SetWriter(logfile)
	}

	err = opts.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid options: %v\n", err)
		os.Exit(1)
	}

	cfg := &schema.Schema{}
	if len(opts.Schema) > 0 {
		err = cfg.LoadFile(opts.Schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
			os.Exit(1)
		}
	}

	if len(opts.SchemaDir) > 0 {
		err = cfg.LoadDir(opts.SchemaDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
			os.Exit(1)
		}
	}

	evaluator, err := sqlproxy.NewEvaluator(cfg, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to mongodb failed: %v\n", err)
		os.Exit(1)
	}

	logger := log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())

	logger.Logf(log.Always, "[initandlisten] connecting to mongodb at %v", opts.MongoOpts.MongoURI)

	svr, err := server.New(cfg, evaluator, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %v", err)
		os.Exit(1)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		logger.Logf(log.Info, "[initandlisten] got %s signal, now exiting...", sig.String())
		svr.Close()
		logger.Logf(log.Info, "[initandlisten] shutting down with code 0")
	}()

	svr.Run()
}
