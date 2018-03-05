package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"

	"github.com/10gen/openssl"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
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

	err = p.initLog()
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not initialize logger: %v", s, err)
		}
		p.cleanup()
		return err
	}

	startupInfo := p.logStartupInfo()

	profile := p.cfg.Debug.ProfileScope
	if p.cfg.Debug.EnableProfiling == "cpu" && profile == "all" {
		filename := fmt.Sprintf("mongosqld_%s.pprof", time.Now().Format("2006-01-02-15-04-05.000000"))
		err := util.StartCPUProfile(filename)
		if err != nil {
			p.cleanup()
			return err
		}
	}

	p.sessionProvider, err = mongodb.NewSqldSessionProvider(p.cfg)
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not create session provider: %v", s, err)
		}
		p.cleanup()
		return err
	}

	err = p.loadSchema()
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not load schema: %v", s, err)
		}
		p.cleanup()
		return err
	}

	// If we're loading a really large DRDL file, it can create a lot of garbage.
	// Forcing garbage collection will allow the GC to catch up and help avoid
	// unnecessary OOM errors in such cases.
	runtime.GC()

	p.svr, err = server.New(p.schema, p.sessionProvider, p.cfg)
	if err != nil {
		if !service.Interactive() {
			p.serviceLogger.Errorf("%s failed to start. Could not start server: %v", s, err)
		}
		p.cleanup()
		return err
	}

	p.svr.StoreStartupInfo(startupInfo)

	util.PanicSafeGo(func() {
		p.svr.Run()
		p.controlLogger.Infof(log.Always, "[signalProcessingThread] shutting down")
		p.cleanup()
		p.done <- struct{}{}
	}, func(err interface{}) {
		p.controlLogger.Fatalf(log.Always, "%v", err)
		p.cleanup()
		panic(err)
	})

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
	log.Flush()

	if p.sessionProvider != nil {
		p.sessionProvider.Close()
	}
	if p.logfile != nil {
		p.logfile.Close()
	}
	if p.cfg.Net.UnixDomainSocket.Enabled {
		os.Remove(fmt.Sprintf("%s/mysql.sock", p.cfg.Net.UnixDomainSocket.PathPrefix))
	}
	util.StopCPUProfile()
}

func (p *program) initLog() error {

	verbosity := log.Verbosity(p.cfg.SystemLog.Verbosity)
	if p.cfg.SystemLog.Quiet {
		verbosity = log.Quiet
	}
	log.SetVerbosity(verbosity)

	if len(p.cfg.SystemLog.Path) > 0 {

		info, err := os.Stat(p.cfg.SystemLog.Path)
		if err == nil {
			//return an error if logPath is a directory
			if info.IsDir() {
				return fmt.Errorf("logPath \"%s\" should name a file, not a directory", p.cfg.SystemLog.Path)
			}

			// rename existing file at logPath, if necessary
			if !p.cfg.SystemLog.LogAppend {
				current := p.cfg.SystemLog.Path
				now := time.Now().Format(log.RotationTimeFormat)
				archive := fmt.Sprintf("%s.%s", current, now)

				msg := fmt.Sprintf("log file \"%s\" exists; moving to \"%s\"", current, archive)
				if service.Interactive() {
					fmt.Fprintln(os.Stderr, msg)
				} else {
					p.serviceLogger.Infof(msg)
				}

				err = os.Rename(current, archive)
				if err != nil {
					return err
				}
			}
		}

		err = log.SetOutputFile(p.cfg.SystemLog.Path, p.cfg.SystemLog.LogAppend, p.cfg.SystemLog.LogRotate)
		if err != nil {
			return err
		}

		if service.Interactive() {
			fmt.Fprintf(os.Stdout, "log output directed to %s\n", p.cfg.SystemLog.Path)
		}

	} else if !service.Interactive() {
		return fmt.Errorf("when running as a service, a log path must be supplied")
	}

	p.controlLogger = log.NewComponentLogger(log.ControlComponent, log.GlobalLogger())
	return nil
}

func (p *program) loadConfig(args []string) ([]string, error) {
	cfg, args, err := config.Load(args)
	if err != nil {
		return nil, err
	}

	p.cfg = cfg
	return args, nil
}

func (p *program) loadSchema() error {
	if p.cfg.Schema.Path != "" {
		fi, err := os.Stat(p.cfg.Schema.Path)
		if err != nil {
			return err
		}

		var drdlSchema *drdl.Schema

		if fi.IsDir() {
			drdlSchema, err = drdl.NewFromDir(p.cfg.Schema.Path)
			if err != nil {
				return err
			}
		} else {
			drdlSchema, err = drdl.NewFromFile(p.cfg.Schema.Path)
			if err != nil {
				return err
			}
		}

		schema, err := schema.NewFromDRDL(p.controlLogger, drdlSchema)
		if err != nil {
			return err
		}

		p.schema = schema
		return nil
	}

	return nil
}

func (p *program) logStartupInfo() []string {
	hostname, err := os.Hostname()
	if err != nil {
		p.controlLogger.Infof(log.Always, "can't get this server's hostname: %v", err)
	}
	startupInfo := []string{
		fmt.Sprintf("[initandlisten] mongosqld starting: version=%v pid=%v host=%v",
			config.VersionStr, os.Getpid(), hostname),
		fmt.Sprintf("[initandlisten] git version: %v", config.Gitspec),
		fmt.Sprintf("[initandlisten] OpenSSL version %v (built with %v)",
			openssl.Version, openssl.BuildVersion),
		fmt.Sprintf("[initandlisten] options: %v", config.ToJSON(p.cfg)),
	}

	// Production release version strings should not contain a "-", whereas
	// all development releases should, e.g. Production release: v2.0.1
	// Development release: v2.0.0-beta5 or v2.0.0-beta5-8-gfad1111
	if strings.Contains(config.VersionStr, "-") {
		startupInfo = append(startupInfo, fmt.Sprintf("[initandlisten]"),
			fmt.Sprintf("[initandlisten] ** NOTE: This is a development version (%v) of mongosqld.", config.VersionStr),
			fmt.Sprintf("[initandlisten] **       Not recommended for production."),
			fmt.Sprintf("[initandlisten]"),
		)
	}

	if !p.cfg.Security.Enabled {
		startupInfo = append(startupInfo,
			fmt.Sprintf("[initandlisten] ** WARNING: Access control is not enabled for mongosqld."),
			fmt.Sprintf("[initandlisten]"),
		)
	}

	for _, info := range startupInfo {
		p.controlLogger.Infof(log.Always, info)
	}

	return startupInfo
}

func main() {
	args := os.Args[1:]
	action := ""
	if len(args) > 0 && (args[0] == "install" || args[0] == "uninstall") {
		action = args[0]
		args = args[1:]
	}

	p := &program{}
	args, err := p.loadConfig(args)
	if err != nil {
		if err == config.ErrExitEarly {
			os.Exit(0)
		}
		if service.Interactive() {
			fmt.Fprintf(os.Stderr, "failed to start due to configuration error: %v%s", err, log.NewLine)
			fmt.Fprintln(os.Stderr, "try 'mongosqld --help' for more information")
			os.Exit(1)
		}
		cfg := config.Default()
		errorServiceConfig := &service.Config{
			Name:        cfg.ProcessManagement.Service.Name,
			DisplayName: cfg.ProcessManagement.Service.DisplayName,
			Description: cfg.ProcessManagement.Service.Description,
			Arguments:   args,
		}
		// We have no way to display any errors that occur here
		errorService, _ := service.New(p, errorServiceConfig)
		errLogger, _ := errorService.Logger(nil)
		errLogger.Errorf("failed to start due to configuration error: %v%s", err, log.NewLine)
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
			fmt.Fprintf(os.Stderr, "could not install because configuration is invalid: %s\n", err)
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
