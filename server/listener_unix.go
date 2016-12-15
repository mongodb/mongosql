// +build !windows

package server

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
)

const (
	BASEMASK = 0777 // octal literal
)

func (s *Server) populateListeners() error {
	listener, err := net.Listen("tcp", *s.opts.Addr)
	if err != nil {
		return err
	}
	s.listeners = append(s.listeners, listener)

	if s.opts.NoUnixSocket == nil || (s.opts.NoUnixSocket != nil && !*s.opts.NoUnixSocket) {
		socket := fmt.Sprintf("%s/%s", *s.opts.UnixSocketPrefix, "mysql.sock")
		s.variables.Socket = socket

		permissions, err := strconv.ParseInt(*s.opts.FilePermissions, 8, 64)
		if err != nil {
			return err
		}

		oldMask := syscall.Umask(int(BASEMASK - permissions))
		listener, err = net.Listen("unix", socket)
		syscall.Umask(oldMask)
		if err != nil {
			return err
		}

		s.listeners = append(s.listeners, listener)
	}

	return nil
}
