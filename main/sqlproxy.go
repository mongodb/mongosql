package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kardianos/service"

	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
)

type program struct {
	cfg *config.Config

	sessionProvider *mongodb.SessionProvider

	serviceLogger service.Logger
	logfile       *os.File
	controlLogger *log.Logger
	schema        *schema.Schema
	svr           *server.Server

	done chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.done = make(chan struct{})

	var err error
	if !service.Interactive() {
		p.serviceLogger, err = s.Logger(nil)
		if err != nil {
			return err
		}
	}

	err = config.Validate(p.cfg)
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Configuration was invalid: %v", s, err)
		}
		p.cleanup()
		return err
	}

	if !service.Interactive() {
		p.serviceLogger.Infof("%s starting with options: %s", s, config.ToJSON(p.cfg))
	}

	err = p.loadSchema()
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not load schema: %v", s, err)
		}
		p.cleanup()
		return err
	}

	err = p.initLog()
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not initialize logger: %v", s, err)
		}
		p.cleanup()
		return err
	}

	p.logStartupInfo()

	p.sessionProvider, err = mongodb.NewSqldSessionProvider(p.cfg)
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not create session provider: %v", s, err)
		}
		p.cleanup()
		return err
	}

	p.svr, err = server.New(p.schema, p.sessionProvider, p.cfg)
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not start server: %v", s, err)
		}
		p.cleanup()
		return err
	}

	util.PanicSafeGo(func() {
		p.svr.Run()
		p.controlLogger.Logf(log.Always, "[signalProcessingThread] shutting down")
		p.cleanup()
		p.done <- struct{}{}
	},
		func(err interface{}) {
			p.controlLogger.Logf(log.Always, "%v", err)
			p.cleanup()
			panic(err)
		},
	)

	return nil
}

func (p *program) Stop(s service.Service) error {
	if !service.Interactive() {
		p.serviceLogger.Infof("%s stopping", s)
	}

	p.svr.Close()
	select {
	case <-p.done:
	case <-time.After(5 * time.Second):
	}
	return nil
}

func (p *program) cleanup() {
	if p.sessionProvider != nil {
		p.sessionProvider.Close()
	}
	if p.logfile != nil {
		p.logfile.Close()
	}
	if p.cfg.Net.UnixDomainSocket.Enabled {
		os.Remove(fmt.Sprintf("%s/mysql.sock", p.cfg.Net.UnixDomainSocket.PathPrefix))
	}
}

func (p *program) initLog() error {

	log.SetVerbosity(&verbosityLevel{
		level: p.cfg.SystemLog.Verbosity,
		quiet: p.cfg.SystemLog.Quiet,
	})

	if len(p.cfg.SystemLog.Path) > 0 {
		mode := os.O_CREATE | os.O_WRONLY

		if p.cfg.SystemLog.LogAppend {
			mode = mode | os.O_APPEND
		} else {
			mode = mode | os.O_TRUNC
		}

		var err error
		p.logfile, err = os.OpenFile(p.cfg.SystemLog.Path, mode, 0666)
		if err != nil {
			return err
		}

		if service.Interactive() {
			fmt.Fprintf(os.Stdout, "log output directed to %s\n", p.cfg.SystemLog.Path)
		}

		log.SetWriter(p.logfile)
	} else if !service.Interactive() {
		return fmt.Errorf("when running as a service, a log path must be supplied")
	}

	p.controlLogger = log.NewComponentLogger(log.ControlComponent, log.GlobalLogger())
	return nil
}

func (p *program) loadConfig(args []string) error {
	cfg, err := config.Load(args)
	if err != nil {
		return err
	}

	p.cfg = cfg
	return nil
}

func (p *program) loadSchema() error {
	var err error

	fi, err := os.Stat(p.cfg.Schema.Path)
	if err != nil {
		return err
	}

	p.schema = &schema.Schema{}

	if fi.IsDir() {
		err = p.schema.LoadDir(p.cfg.Schema.Path)
		if err != nil {
			return err
		}
	} else {
		err = p.schema.LoadFile(p.cfg.Schema.Path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *program) logStartupInfo() {
	p.controlLogger.Logf(log.Always, "[initandlisten] mongosqld version: %v", config.VersionStr)
	p.controlLogger.Logf(log.Always, "[initandlisten] git version: %v", config.Gitspec)
	p.controlLogger.Logf(log.Always, "[initandlisten] options: %v", config.ToJSON(p.cfg))

	// Production release version strings should not contain a "-", whereas all development releases should, e.g.
	// Production release: v2.0.1
	// Development release: v2.0.0-beta5 or v2.0.0-beta5-8-gfad1111
	if strings.Contains(config.VersionStr, "-") {
		p.controlLogger.Logf(log.Always, "[initandlisten]")
		p.controlLogger.Logf(log.Always, "[initandlisten] ** NOTE: This is a development version (%v) of mongosqld.", config.VersionStr)
		p.controlLogger.Logf(log.Always, "[initandlisten] **       Not recommended for production.")
		p.controlLogger.Logf(log.Always, "[initandlisten]")
	}

	if !p.cfg.Security.Enabled {
		p.controlLogger.Logf(log.Always, "[initandlisten] ** WARNING: Access control is not enabled for mongosqld.")
		p.controlLogger.Logf(log.Always, "[initandlisten]")
	}
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

func main() {
	args := os.Args[1:]
	action := ""
	if len(args) > 0 && (args[0] == "install" || args[0] == "uninstall") {
		action = args[0]
		args = args[1:]
	}

	p := &program{}
	err := p.loadConfig(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start due to configuration error: %v\n", err)
		os.Exit(1)
	}

	svcConfig := &service.Config{
		Name:        p.cfg.ProcessManagement.Service.Name,
		DisplayName: p.cfg.ProcessManagement.Service.DisplayName,
		Description: p.cfg.ProcessManagement.Service.Description,
		Arguments:   args,
	}

	s, err := service.New(p, svcConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch action {
	case "install":
		err = config.Validate(p.cfg)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not install because configuration is invalid: %s\n", err)
			os.Exit(1)
		}

		err = s.Install()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "%s service installed\n", p.cfg.ProcessManagement.Service.Name)
		os.Exit(0)
	case "uninstall":
		err = s.Uninstall()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "%s service uninstalled\n", p.cfg.ProcessManagement.Service.Name)
		os.Exit(0)
	default:
		err = s.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
