package main

import (
	"fmt"
	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/proxy"
	"github.com/10gen/sqlproxy/schema"
	"github.com/jessevdk/go-flags"
	"github.com/mongodb/mongo-tools/common/log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

	opts := sqlproxy.Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Logf(log.Always, "failed to parse args: %v", err)
		os.Exit(1)
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
		return
	}

	var svr *proxy.Server
	svr, err = proxy.NewServer(cfg, evaluator, opts)
	if err != nil {
		log.Logf(log.Always, "error starting server")
		log.Logf(log.Always, err.Error())
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		log.Logf(log.Info, "Got signal [%d] to exit.", sig)
		svr.Close()
	}()

	log.Logf(log.Info, "Going to start running on %s.", opts.Addr)

	svr.Run()
}
