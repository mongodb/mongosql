package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
	"github.com/mongodb/mongo-tools/common/log"
)

type LogLevel struct {
	level string
}

func (ll *LogLevel) IsQuiet() bool {
	return false
}

func (ll *LogLevel) Level() int {
	switch ll.level {
	case "always", "error":
		return log.Always
	case "info":
		return log.Info
	case "v", "debug":
		return log.DebugLow
	case "vv":
		return log.DebugHigh
	}
	return log.Info
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	opts, err := sqlproxy.NewOptions()
	if err != nil {
		os.Exit(1)
	}

	err = opts.Parse()
	if err != nil {
		os.Exit(1)
	}

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
			log.Logf(log.Always, "failed to open logfile: %v", err)
			os.Exit(1)
		}
		defer logfile.Close()

		log.SetWriter(logfile)
	}

	err = opts.Validate()
	if err != nil {
		log.Logf(log.Always, "invalid options: %v", err)
		os.Exit(1)
	}

	cfg := &schema.Schema{}
	if len(opts.Schema) > 0 {
		err = cfg.LoadFile(opts.Schema)
		if err != nil {
			log.Logf(log.Always, "failed to load schema: %v", err)
			os.Exit(1)
		}
	}
	if len(opts.SchemaDir) > 0 {
		err = cfg.LoadDir(opts.SchemaDir)
		if err != nil {
			log.Logf(log.Always, "failed to load schema: %v", err)
			os.Exit(1)
		}
	}

	log.SetVerbosity(opts)

	evaluator, err := sqlproxy.NewEvaluator(cfg, opts)
	if err != nil {
		log.Logf(log.Always, "error starting evaluator")
		log.Logf(log.Always, err.Error())
		os.Exit(1)
	}

	var svr *server.Server
	svr, err = server.New(cfg, evaluator, opts)
	if err != nil {
		log.Logf(log.Always, "error starting server")
		log.Logf(log.Always, err.Error())
		os.Exit(1)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		log.Logf(log.Info, "Got signal [%d] to exit.", sig)
		svr.Close()
	}()

	svr.Run()
}
