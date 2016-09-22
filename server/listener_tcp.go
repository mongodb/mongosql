// +build windows

package server

import (
	"net"
)

func (s *Server) populateListeners() error {
	listener, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		return err
	}
	s.listeners = append(s.listeners, listener)

	return nil
}
