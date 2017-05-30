// +build windows

package server

import (
	"net"
	"strconv"
)

func (s *Server) populateListeners() error {
	listener, err := net.Listen("tcp", net.JoinHostPort(s.cfg.Net.BindIP, strconv.Itoa(s.cfg.Net.Port)))
	if err != nil {
		return err
	}
	s.listeners = append(s.listeners, listener)
	return nil
}
