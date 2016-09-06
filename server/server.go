package server

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"runtime"
	"strings"
	"sync"

	sqlproxy "github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/shopspring/decimal"
)

// Server manages connections with clients.
type Server struct {
	sync.Mutex

	eval              *sqlproxy.Evaluator
	opts              sqlproxy.Options
	databases         map[string]*schema.Database
	activeConnections map[uint32]*conn
	variables         *variable.Container
	tlsConfig         *tls.Config

	connCount uint32

	running bool

	listener net.Listener
}

// New creates a NewServer.
func New(schema *schema.Schema, eval *sqlproxy.Evaluator, opts sqlproxy.Options) (*Server, error) {

	decimal.DivisionPrecision = 34

	s := &Server{
		eval:              eval,
		opts:              opts,
		running:           false,
		activeConnections: make(map[uint32]*conn),
		databases:         schema.Databases,
		variables:         variable.NewGlobalContainer(),
	}

	var err error
	netProto := "tcp"
	if strings.Contains(netProto, "/") {
		netProto = "unix"
	}
	s.listener, err = net.Listen(netProto, opts.Addr)
	if err != nil {
		return nil, err
	}

	if len(opts.SSLPEMFile) > 0 {
		s.tlsConfig = &tls.Config{
			InsecureSkipVerify: opts.SSLAllowInvalidCerts || opts.SSLCAFile == "",
		}
		cert, err := tls.LoadX509KeyPair(opts.SSLPEMFile, opts.SSLPEMFile)
		if err != nil {
			return nil, mysqlerrors.Unknownf("failed to load PEM file '%v': %v", opts.SSLPEMFile, err)
		}
		s.tlsConfig.Certificates = []tls.Certificate{cert}

		if len(opts.SSLCAFile) > 0 {
			caCert, err := ioutil.ReadFile(opts.SSLCAFile)
			if err != nil {
				return nil, mysqlerrors.Unknownf("failed to load CA file '%v': %v", opts.SSLCAFile, err)
			}
			s.tlsConfig.RootCAs = x509.NewCertPool()
			ok := s.tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
			if !ok {
				return nil, mysqlerrors.Unknownf("unable to append valid cert from PEM file '%v'", opts.SSLCAFile)
			}
		}
	}

	return s, nil
}

// Run starts the server and begins accepting connections.
func (s *Server) Run() {
	s.running = true

	log.Logf(log.Always, "waiting for connections at %v", s.listener.Addr())

	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Log(log.Always, err.Error())
			continue
		}

		go s.onConn(conn)
	}
}

// Close stops the server and stops accepting connections.
func (s *Server) Close() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *Server) onConn(c net.Conn) {
	conn := newConn(s, c)

	log.Logf(log.Info, "connection accepted from %v #%v", c.RemoteAddr(), conn.ConnectionId())

	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Logf(log.Always, "[conn%v] panic with %v: %v\n%s", conn.ConnectionId(), c.RemoteAddr(), err, buf)
		}

		conn.close()
	}()

	schema := s.eval.Schema()
	info, err := mongodb.LoadInfo(conn.session, &schema, s.opts.Auth)
	if err != nil {
		log.Logf(log.Always, "[conn%v] error retrieving information from MongoDB: %v", conn.ConnectionId(), err)
		c.Close()
		return
	}
	conn.variables.MongoDBInfo = info

	if err := conn.handshake(); err != nil {
		log.Logf(log.Always, "[conn%v] handshake error: %v", conn.ConnectionId(), err)
		c.Close()
		return
	}

	s.Lock()
	s.activeConnections[conn.ConnectionId()] = conn
	s.Unlock()

	conn.run()
}
