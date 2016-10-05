// Main package for the mongosqld tool.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
	"github.com/10gen/sqlproxy/util"
)

func getConfig(opts options.SqldOptions) *schema.Schema {
	var err error

	cfg := &schema.Schema{}
	if len(opts.Schema) > 0 {
		err = cfg.LoadFile(opts.Schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
			os.Exit(util.ExitError)
		}
	}

	if len(opts.SchemaDir) > 0 {
		err = cfg.LoadDir(opts.SchemaDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
			os.Exit(util.ExitError)
		}
	}

	return cfg
}

func getOptions() options.SqldOptions {
	opts, err := options.NewSqldOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating command line options: %v\n", err)
		os.Exit(util.ExitError)
	}

	err = opts.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing command line options: %v\n", err)
		fmt.Fprintln(os.Stderr, "try 'mongosqld --help' for more information")
		os.Exit(util.ExitBadOptions)
	}

	err = opts.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid options: %v\n", err)
		os.Exit(util.ExitBadOptions)
	}

	return opts
}

func daemonize() {
	var args []string

	for _, arg := range os.Args {
		if arg == "--fork" {
			continue
		}
		args = append(args, arg)
	}

	cmd := exec.Command(os.Args[0], args...)
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start child process: %v\n", err)
		os.Exit(util.ExitError)
	}

	fmt.Fprintf(os.Stdout, "forked process: %d\n", cmd.Process.Pid)
	fmt.Fprintf(os.Stdout, "child process started successfully, parent exiting\n")
	os.Exit(util.ExitClean)
}

func setupLog(opts options.SqldOptions) (*os.File, *log.Logger) {
	var logfile *os.File
	var err error

	// set global log level
	log.SetVerbosity(opts.SqldLog)

	if len(opts.LogPath) > 0 {
		mode := os.O_WRONLY
		if _, err = os.Stat(opts.LogPath); err != nil {
			mode = mode | os.O_CREATE
		}

		if opts.LogAppend {
			mode = mode | os.O_APPEND
		} else {
			mode = mode | os.O_TRUNC
		}

		logfile, err = os.OpenFile(opts.LogPath, mode, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening log file: %v\n", err)
			os.Exit(util.ExitError)
		}

		log.SetWriter(logfile)
	}

	return logfile, log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	opts := getOptions()

	if opts.Version {
		common.PrintVersionAndGitspec("mongosqld", os.Stdout)
		os.Exit(util.ExitClean)
	}

	if opts.PrintHelp(os.Stdout) {
		os.Exit(util.ExitClean)
	}

	if opts.Fork {
		daemonize()
	}

	cfg := getConfig(opts)
	evaluator, err := sqlproxy.NewEvaluator(cfg, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to mongodb failed: %v\n", err)
		os.Exit(util.ExitError)
	}

	logfile, logger := setupLog(opts)
	if logfile != nil {
		defer logfile.Close()
	}

	logger.Logf(log.Always, "[initandlisten] connecting to mongodb at %v", opts.SqldMongoConnection.MongoURI)

	svr, err := server.New(cfg, evaluator, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %v\n", err)
		os.Exit(util.ExitError)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		logger.Logf(log.Info, "[signalProcessingThread] got %s signal, now exiting...", sig.String())
		svr.Close()
	}()

	svr.Run()
	logger.Logf(log.Info, "[signalProcessingThread] shutting down with code 0")
	os.Exit(util.ExitClean)
}
