package proxy

import (
	sqlproxy "github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"

	"net"
	"runtime"
	"strings"
)

type Server struct {
	cfg  *schema.Schema
	eval *sqlproxy.Evaluator

	running bool

	listener net.Listener

	databases map[string]*schema.Database
}

func (s *Server) GetDatabases() map[string]*schema.Database {
	return s.databases
}

func NewServer(cfg *schema.Schema, eval *sqlproxy.Evaluator) (*Server, error) {
	s := new(Server)

	s.cfg = cfg
	s.eval = eval
	s.running = false

	s.databases = cfg.Databases

	var err error
	netProto := "tcp"
	if strings.Contains(netProto, "/") {
		netProto = "unix"
	}
	s.listener, err = net.Listen(netProto, cfg.Addr)

	if err != nil {
		return nil, err
	}

	log.Logf(log.Always, "Server run MySql Protocol Listen(%s) at [%s]", netProto, cfg.Addr)
	return s, nil
}

func (s *Server) Run() error {
	s.running = true

	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Logf(log.Always, "accept error %s", err.Error())
			continue
		}

		go s.onConn(conn)
	}

	return nil
}

func (s *Server) Close() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *Server) onConn(c net.Conn) {
	conn := s.newConn(c)

	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Logf(log.Always, "onConn panic %v: %v\n%s", c.RemoteAddr().String(), err, buf)
		}

		conn.Close()
	}()

	if err := conn.Handshake(); err != nil {
		log.Logf(log.Always, "handshake error %s", err.Error())
		c.Close()
		return
	}

	conn.Run()

}
