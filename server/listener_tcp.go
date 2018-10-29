// +build windows

package server

import (
	"context"
	"net"
	"strconv"
	"strings"
)

// registerSignalListeners registers functions used to respond to specific user-issued signals.
// This does nothing on Windows.
func (s *Server) registerSignalListeners(ctx context.Context) {}

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
