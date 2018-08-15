// +build !windows

package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/variable"
)

const (
	baseMask = 0777 // octal literal
)

// signalHandler holds an OS signal and a handler function for that signal.
type signalHandler struct {
	signals  chan os.Signal
	handlers map[string]func() error
}

// registerSignalListeners registers functions used to respond to specific user-issued signals.
func (s *Server) registerSignalListeners() {
	var handlers = []struct {
		Signal  syscall.Signal
		Handler func() error
	}{
		{syscall.SIGUSR1, func() error { return s.RotateLogs() }},
	}

	sh := signalHandler{
		signals:  make(chan os.Signal, len(handlers)),
		handlers: make(map[string]func() error, len(handlers)),
	}

	for _, e := range handlers {
		signal.Notify(sh.signals, e.Signal)
		sh.handlers[e.Signal.String()] = e.Handler
	}

	go s.listenForSignals(sh)
}

func (s *Server) listenForSignals(sh signalHandler) {
	for {
		select {
		case <-s.lifetimeCtx.Done():
			return
		case sig := <-sh.signals:
			sigName := sig.String()
			s.logger.Debugf(log.Admin, "handling %s", sigName)
			if err := sh.handlers[sigName](); err != nil {
				s.logger.Debugf(log.Admin, "%s error: %s", sigName, err)
			}
			s.logger.Debugf(log.Admin, "successfully handled %s", sigName)
		}
	}
}

func (s *Server) populateListeners() error {
	var err error
	var listener net.Listener
	port := strconv.Itoa(s.cfg.Net.Port)
	for _, host := range s.cfg.Net.BindIP {
		listener, err = net.Listen("tcp", net.JoinHostPort(strings.TrimSpace(host), port))
		if err != nil {
			return err
		}
		s.listeners = append(s.listeners, listener)
	}

	if s.cfg.Net.UnixDomainSocket.Enabled {
		socket := fmt.Sprintf("%s/%s", s.cfg.Net.UnixDomainSocket.PathPrefix, "mysql.sock")
		s.variables.SetSystemVariable(variable.Socket, socket)

		permissions, err := strconv.ParseInt(s.cfg.Net.UnixDomainSocket.FilePermissions, 8, 64)
		if err != nil {
			return err
		}

		oldMask := syscall.Umask(int(baseMask - permissions))
		listener, err = net.Listen("unix", socket)
		if err != nil {
			if !isErrAddrInUse(err) {
				return err
			}

			c, dialErr := net.Dial("unix", socket)
			if dialErr == nil {
				// probably a server already listening
				// for connections, don't attempt to
				// unlink the socket file
				_ = c.Close()
				return err
			}

			// remove socket file
			_ = os.Remove(socket)
			listener, err = net.Listen("unix", socket)
			if err != nil {
				return err
			}
		}

		syscall.Umask(oldMask)
		s.listeners = append(s.listeners, listener)
	}

	return nil
}

// isErrAddrInUse returns true if err is of type syscall.EADDRINUSE
// and false otherwise.
func isErrAddrInUse(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
			return sysErr.Err == syscall.EADDRINUSE
		}
	}
	return false
}
