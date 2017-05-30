// Main package for the mongosqld tool.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
	"github.com/10gen/sqlproxy/util"
)

func loadConfig() *config.Config {

	cfg, err := config.Load(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(util.ExitError)
	}

	err = config.Validate(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(util.ExitError)
	}

	return cfg
}

func loadSchema(cfg *config.Schema) *schema.Schema {
	var err error

	fi, err := os.Stat(cfg.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
		os.Exit(util.ExitError)
	}

	s := &schema.Schema{}

	if fi.IsDir() {
		err = s.LoadDir(cfg.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
			os.Exit(util.ExitError)
		}
	} else {
		err = s.LoadFile(cfg.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading schema: %v\n", err)
			os.Exit(util.ExitError)
		}
	}

	return s
}

type verbosityLevel struct {
	level int
	quiet bool
}

func (v *verbosityLevel) Level() int {
	return v.level
}

func (v *verbosityLevel) IsQuiet() bool {
	return v.quiet
}

func setupLog(cfg *config.SystemLog) (*os.File, *log.Logger) {
	var logfile *os.File
	var err error

	// set global log level
	log.SetVerbosity(&verbosityLevel{
		level: cfg.Verbosity,
		quiet: cfg.Quiet,
	})

	if len(cfg.Path) > 0 {
		mode := os.O_WRONLY
		if _, err = os.Stat(cfg.Path); err != nil {
			mode = mode | os.O_CREATE
		}

		if cfg.LogAppend {
			mode = mode | os.O_APPEND
		} else {
			mode = mode | os.O_TRUNC
		}

		logfile, err = os.OpenFile(cfg.Path, mode, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening log file: %v\n", err)
			os.Exit(util.ExitError)
		}

		log.SetWriter(logfile)
	}

	return logfile, log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())
}

func logStartupInfo(cfg *config.Config) {
	controlLogger := log.NewComponentLogger(log.ControlComponent, log.GlobalLogger())

	controlLogger.Logf(log.Always, "[initandlisten] mongosqld version: %v", common.VersionStr)
	controlLogger.Logf(log.Always, "[initandlisten] git version: %v", common.Gitspec)
	controlLogger.Logf(log.Always, "[initandlisten] options: %v", config.ToJSON(cfg))

	// Production release version strings should not contain a "-", whereas all development releases should, e.g.
	// Production release: v2.0.1
	// Development release: v2.0.0-beta5 or v2.0.0-beta5-8-gfad1111
	if strings.Contains(common.VersionStr, "-") {
		controlLogger.Logf(log.Always, "[initandlisten]")
		controlLogger.Logf(log.Always, "[initandlisten] ** NOTE: This is a development version (%v) of mongosqld.", common.VersionStr)
		controlLogger.Logf(log.Always, "[initandlisten] **       Not recommended for production.")
		controlLogger.Logf(log.Always, "[initandlisten]")
	}

	if !cfg.Security.Enabled {
		controlLogger.Logf(log.Always, "[initandlisten] ** WARNING: Access control is not enabled for mongosqld.")
		controlLogger.Logf(log.Always, "[initandlisten]")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg := loadConfig()

	logStartupInfo(cfg)

	schema := loadSchema(&cfg.Schema)

	logfile, logger := setupLog(&cfg.SystemLog)
	if logfile != nil {
		defer logfile.Close()
	}

	sessionProvider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error establishing session provider: %v\n", err)
		os.Exit(util.ExitError)
	}
	defer sessionProvider.Close()

	svr, err := server.New(schema, sessionProvider, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %v\n", err)
		os.Exit(util.ExitError)
	}

	defer func() {
		if cfg.Net.UnixDomainSocket.Enabled {
			os.Remove(fmt.Sprintf("%s/mysql.sock", cfg.Net.UnixDomainSocket.PathPrefix))
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	util.PanicSafeGo(func() {
		sig := <-sc
		logger.Logf(log.Info, "[signalProcessingThread] got %s signal, now exiting...", sig.String())
		svr.Close()
	}, func(err interface{}) {
		logger.Errf(log.Info, "[signalProcessingThread] got error %s now exiting...", err)
	})

	svr.Run()
	logger.Logf(log.Info, "[signalProcessingThread] shutting down with code 0")
	os.Exit(util.ExitClean)
}
