// Main package for the mongosqld tool.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
	"github.com/10gen/sqlproxy/util"
)

func getSchema(opts options.SqldOptions) *schema.Schema {
	var err error

	cfg := &schema.Schema{}
	if len(*opts.Schema) > 0 {
		err = cfg.LoadFile(*opts.Schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
			os.Exit(util.ExitError)
		}
	}

	if len(*opts.SchemaDir) > 0 {
		err = cfg.LoadDir(*opts.SchemaDir)
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

func setupLog(opts options.SqldOptions) (*os.File, *log.Logger) {
	var logfile *os.File
	var err error

	// set global log level
	log.SetVerbosity(*opts.SqldLog)

	if len(*opts.LogPath) > 0 {
		mode := os.O_WRONLY
		if _, err = os.Stat(*opts.LogPath); err != nil {
			mode = mode | os.O_CREATE
		}

		if *opts.LogAppend {
			mode = mode | os.O_APPEND
		} else {
			mode = mode | os.O_TRUNC
		}

		logfile, err = os.OpenFile(*opts.LogPath, mode, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening log file: %v\n", err)
			os.Exit(util.ExitError)
		}

		log.SetWriter(logfile)
	}

	return logfile, log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())
}

func logStartupInfo(opts options.SqldOptions) {
	controlLogger := log.NewComponentLogger(log.ControlComponent, log.GlobalLogger())

	controlLogger.Logf(log.Always, "[initandlisten] mongosqld version: %v", common.VersionStr)
	controlLogger.Logf(log.Always, "[initandlisten] git version: %v", common.Gitspec)
	controlLogger.Logf(log.Always, "[initandlisten] arguments: %v", opts)

	// Production release version strings should not contain a "-", whereas all development releases should, e.g.
	// Production release: v2.0.1
	// Development release: v2.0.0-beta5 or v2.0.0-beta5-8-gfad1111
	if strings.Contains(common.VersionStr, "-") {
		controlLogger.Logf(log.Always, "[initandlisten]")
		controlLogger.Logf(log.Always, "[initandlisten] ** NOTE: This is a development version (%v) of mongosqld.", common.VersionStr)
		controlLogger.Logf(log.Always, "[initandlisten] **       Not recommended for production.")
		controlLogger.Logf(log.Always, "[initandlisten]")
	}

	if opts.Auth == nil || !*opts.Auth {
		controlLogger.Logf(log.Always, "[initandlisten] ** WARNING: Access control is not enabled for mongosqld.")
		controlLogger.Logf(log.Always, "[initandlisten]")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	opts := getOptions()

	if opts.Version != nil && *opts.Version {
		common.PrintVersionAndGitspec("mongosqld", os.Stdout)
		os.Exit(util.ExitClean)
	}

	if opts.PrintHelp(os.Stdout) {
		os.Exit(util.ExitClean)
	}

	logStartupInfo(opts)

	options.EnsureOptsNotNil(&opts)

	cfg := getSchema(opts)
	evaluator, err := sqlproxy.NewEvaluator(cfg, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to mongodb failed: %v\n", err)
		os.Exit(util.ExitError)
	}

	logfile, logger := setupLog(opts)
	if logfile != nil {
		defer logfile.Close()
	}

	logger.Logf(log.Always, "[initandlisten] connecting to mongodb at %v", *opts.SqldMongoConnection.MongoURI)

	svr, err := server.New(cfg, evaluator, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %v\n", err)
		os.Exit(util.ExitError)
	}

	defer func() {
		if opts.NoUnixSocket == nil || (opts.NoUnixSocket != nil && !*opts.NoUnixSocket) {
			os.Remove(fmt.Sprintf("%s/mysql.sock", *opts.UnixSocketPrefix))
		}
	}()

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
