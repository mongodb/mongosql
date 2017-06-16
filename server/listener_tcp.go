// +build windows

package server

import (
	"net"
	"strconv"
	"strings"
)

func (s *Server) populateListeners() error {
	port := strconv.Itoa(s.cfg.Net.Port)
	for _, host := range s.cfg.Net.BindIP {
		listener, err := net.Listen("tcp", net.JoinHostPort(strings.TrimSpace(host), port))
		if err != nil {
			return err
		}
		s.listeners = append(s.listeners, listener)
	}

	return nil
}
