package provider

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

type provider func(context.Context) (driver.Connection, error)

// NewSessionConnPool is the initializer for SessionConnPool.
func NewSessionConnPool(ctx context.Context, prv provider, max int) (*SessionConnPool, error) {

	pool := &SessionConnPool{
		conns: make(chan *sessionConn, max),
	}

	for i := 0; i < max; i++ {
		c, err := prv(ctx)
		if err != nil {
			pool.Close()
			return nil, err
		}

		pool.conns <- &sessionConn{c, pool}
	}

	return pool, nil
}

// SessionConnPool represents a pool of sessionConns.
type SessionConnPool struct {
	connsLock sync.Mutex
	conns     chan *sessionConn
	errLock   sync.Mutex
	err       error
}

// Err returns the err field of SessionConnPool.
func (p *SessionConnPool) Err() error {
	p.errLock.Lock()
	err := p.err
	p.errLock.Unlock()
	return err
}

// Close closes every sessionConn in the SessionConnPool.
func (p *SessionConnPool) Close() {
	p.connsLock.Lock()
	conns := p.conns
	p.conns = nil
	p.connsLock.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for c := range conns {
		_ = c.Connection.Close()
	}
}

// Get returns a sessionConn from the SessionConnPool.
func (p *SessionConnPool) Get(ctx context.Context) (driver.Connection, error) {
	p.connsLock.Lock()
	conns := p.conns
	p.connsLock.Unlock()

	if conns == nil {
		return nil, topology.ErrPoolDisconnected
	}

	return p.getConn(ctx, conns)
}

func (p *SessionConnPool) getConn(ctx context.Context, conns <-chan *sessionConn) (driver.Connection, error) {
	select {
	case c := <-conns:
		if c == nil {
			return nil, topology.ErrPoolDisconnected
		}

		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// returnConn sends a sessionConn back to its SessionConnPool's channel
// of connections. The connection is closed if the channel is nil.
func (p *SessionConnPool) returnConn(c *sessionConn) error {
	p.connsLock.Lock()
	defer p.connsLock.Unlock()

	if p.conns == nil {
		return c.Connection.Close()
	}

	p.conns <- c
	return nil
}

type sessionConn struct {
	driver.Connection
	p *SessionConnPool
}

// Close puts a sessionConn back in its SessionConnPool.
func (c *sessionConn) Close() error {
	return c.p.returnConn(c)
}

// CompressWireMessage handles compressing the provided wire message using the
// underlying driver.Connection if it is also a driver.Compressor.
func (c *sessionConn) CompressWireMessage(src, dst []byte) ([]byte, error) {
	if compressor, ok := c.Connection.(driver.Compressor); ok {
		return compressor.CompressWireMessage(src, dst)
	}

	// Cannot compress if the underlying driver.Connection is not a driver.Compressor.
	return append(dst, src...), nil
}

// noopCloseSessionConn wraps a connection and replaces the Close func with a no-op.
// This is used when passing a sessionConn to the driver during (*Session).Login to
// prevent any premature calls to (*sessionConn).Close from the driver.
type noopCloseSessionConn struct {
	driver.Connection
}

func (ncsc noopCloseSessionConn) Close() error {
	return nil
}
