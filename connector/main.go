package main

import (
	"flag"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/proxy"
	"github.com/erh/mongo-sql-temp/translator"
	"github.com/mongodb/mongo-tools/common/log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var configFile *string = flag.String("config", "/etc/bi-connector.conf", "bi-connector config file")

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

	flag.Parse()

	if len(*configFile) == 0 {
		log.Logf(log.Always, "must use a config file")
		return
	}

	cfg, err := config.ParseConfigFile(*configFile)
	if err != nil {
		log.Logf(log.Always, "error parsing config file")
		log.Logf(log.Always, err.Error())
		return
	}

	log.SetVerbosity(&LogLevel{level: cfg.LogLevel})

	evaluator, err := translator.NewEvaluator(cfg)
	if err != nil {
		log.Logf(log.Always, "error starting evaluator")
		log.Logf(log.Always, err.Error())
		return
	}

	var svr *proxy.Server
	svr, err = proxy.NewServer(cfg, evaluator)
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

	log.Logf(log.Info, "Going to start running on %s.", cfg.Addr)

	svr.Run()
}
