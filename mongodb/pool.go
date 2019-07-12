package mongodb

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

type provider func(context.Context) (driver.Connection, error)

func newSessionConnPool(ctx context.Context, prv provider, max int) (*sessionConnPool, error) {

	pool := &sessionConnPool{
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

	close(conns)
	for c := range conns {
		_ = c.Connection.Close()
	}
}

func (p *sessionConnPool) Get(ctx context.Context) (driver.Connection, error) {
	p.connsLock.Lock()
	conns := p.conns
	p.connsLock.Unlock()

	if conns == nil {
		return nil, topology.ErrPoolDisconnected
	}

	return p.getConn(ctx, conns)
}

func (p *sessionConnPool) getConn(ctx context.Context, conns <-chan *sessionConn) (driver.Connection, error) {
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

func (p *sessionConnPool) returnConn(c *sessionConn) error {
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
	p *sessionConnPool
}

func (c *sessionConn) Close() error {
	return c.p.returnConn(c)
}
