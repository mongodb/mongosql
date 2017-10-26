// +build !windows

package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/10gen/sqlproxy/variable"
)

const (
	BASEMASK = 0777 // octal literal
)

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

		oldMask := syscall.Umask(int(BASEMASK - permissions))
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
				c.Close()
				return err
			}

			// remove socket file
			os.Remove(socket)
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
