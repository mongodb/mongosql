package mongodb

import (
	"context"
	"sync"

	"github.com/10gen/mongo-go-driver/yamgo/private/conn"
)

func newSessionConnPool(ctx context.Context, provider conn.Provider, max int32) (*sessionConnPool, error) {

	pool := &sessionConnPool{
		conns: make(chan *sessionConn, max),
	}

	for i := 0; i < int(max); i++ {
		conn, err := provider(ctx)
		if err != nil {
			pool.Close()
			return nil, err
		}

		pool.conns <- &sessionConn{conn, pool}
	}

	return pool, nil
}

type sessionConnPool struct {
	connsLock sync.Mutex
	conns     chan *sessionConn
	errLock   sync.Mutex
	err       error
}

func (p *sessionConnPool) Err() error {
	p.errLock.Lock()
	err := p.err
	p.errLock.Unlock()
	return err
}

func (p *sessionConnPool) Close() {
	p.connsLock.Lock()
	conns := p.conns
	p.conns = nil
	p.connsLock.Unlock()

	if conns == nil {
		return
	}

	p.errLock.Lock()
	err := p.err
	p.errLock.Unlock()

	close(conns)
	for c := range conns {
		if err != nil {
			// if the pool has an error, we need to ensure that any remaining
			// connections are pulled as we might know more than the underlying
			// pool and monitors.
			c.Connection.MarkDead()
		}
		c.Connection.Close()
	}
}

func (p *sessionConnPool) Get(ctx context.Context) (conn.Connection, error) {
	p.connsLock.Lock()
	conns := p.conns
	p.connsLock.Unlock()

	if conns == nil {
		return nil, conn.ErrPoolClosed
	}

	return p.getConn(ctx, conns)
}

func (p *sessionConnPool) getConn(ctx context.Context, conns <-chan *sessionConn) (conn.Connection, error) {
	select {
	case c := <-conns:
		if c == nil {
			return nil, conn.ErrPoolClosed
		}

		if c.Expired() {
			p.closeExpiredConn(c)
			return nil, conn.ErrPoolClosed
		}

		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *sessionConnPool) closeExpiredConn(c *sessionConn) {
	c.Connection.Close()

	// if the connection is expired, it is likely due to a network
	// outage since we don't set an idle time, which means the
	// connection experienced a castrophic network failure.
	p.errLock.Lock()
	p.err = conn.ErrPoolClosed
	p.errLock.Unlock()
	p.Close()
}

func (p *sessionConnPool) returnConn(c *sessionConn) error {
	if c.Expired() {
		p.closeExpiredConn(c)
		return conn.ErrPoolClosed
	}

	p.connsLock.Lock()
	defer p.connsLock.Unlock()

	if p.conns == nil {
		return c.Connection.Close()
	}

	p.conns <- c
	return nil
}

type sessionConn struct {
	conn.Connection
	p *sessionConnPool
}

func (c *sessionConn) Close() error {
	return c.p.returnConn(c)
}
