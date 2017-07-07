package mongodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/10gen/mongo-go-driver/conn"
	"github.com/10gen/mongo-go-driver/model"
	"github.com/10gen/mongo-go-driver/msg"
)

func TestNewSessionConnPool_Get_max_connections(t *testing.T) {
	count := 0
	provider := func(ctx context.Context) (conn.Connection, error) {
		count++
		return &mockConn{}, nil
	}

	_, err := newSessionConnPool(context.Background(), provider, 3)
	if err != nil {
		t.Fatal(err)
	}

	if count != 3 {
		t.Fatalf("expected pool to checkout 3 connections eagerly, but only have %d", count)
	}
}

func TestNewSessionConnPool_Return_an_error_when_getting_a_connection_errors(t *testing.T) {
	provider := func(ctx context.Context) (conn.Connection, error) {
		return nil, fmt.Errorf("AHAHAHAH")
	}

	_, err := newSessionConnPool(context.Background(), provider, 3)
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
}

func TestSessionConnPool_Get_when_context_is_cancelled(t *testing.T) {
	var created []*mockConn
	provider := func(_ context.Context) (conn.Connection, error) {
		created = append(created, &mockConn{})
		return created[len(created)-1], nil
	}

	p, err := newSessionConnPool(context.Background(), provider, 2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	_, err = p.Get(ctx)
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
}

func TestSessionConnPool_Get_when_pool_is_closed(t *testing.T) {
	var created []*mockConn
	provider := func(_ context.Context) (conn.Connection, error) {
		created = append(created, &mockConn{})
		return created[len(created)-1], nil
	}

	p, err := newSessionConnPool(context.Background(), provider, 2)
	if err != nil {
		t.Fatal(err)
	}

	p.Close()

	_, err = p.Get(context.Background())
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
}

func TestSessionConnPool_Get_when_pool_is_closed_2(t *testing.T) {
	var created []*mockConn
	provider := func(_ context.Context) (conn.Connection, error) {
		created = append(created, &mockConn{})
		return created[len(created)-1], nil
	}

	p, err := newSessionConnPool(context.Background(), provider, 2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	errChan := make(chan error, 1)

	go func() {
		_, err = p.Get(context.Background())
		errChan <- err
	}()

	time.Sleep(1 * time.Second)

	// close the pool while waiting for a get
	p.Close()

	err = <-errChan
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
}

func TestSessionConnPool_Get_a_connection_which_expired_in_the_pool(t *testing.T) {
	var created []*mockConn
	provider := func(_ context.Context) (conn.Connection, error) {
		created = append(created, &mockConn{})
		return created[len(created)-1], nil
	}

	p, err := newSessionConnPool(context.Background(), provider, 2)
	if err != nil {
		t.Fatal(err)
	}

	created[0].MarkDead()

	_, err = p.Get(context.Background())
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
}

func TestSessionConnPool_Connection_expired_when_checking_in(t *testing.T) {
	var created []*mockConn
	provider := func(_ context.Context) (conn.Connection, error) {
		created = append(created, &mockConn{})
		return created[len(created)-1], nil
	}

	p, err := newSessionConnPool(context.Background(), provider, 2)
	if err != nil {
		t.Fatal(err)
	}

	c, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	created[0].MarkDead()

	c.Close()

	if p.Err() == nil {
		t.Fatalf("expected error, but got none")
	}
}

type mockConn struct {
	dead bool
}

func (c *mockConn) Alive() bool {
	return !c.dead
}

func (c *mockConn) Close() error {
	c.dead = true
	return nil
}

func (c *mockConn) MarkDead() {
	c.dead = true
}

func (c *mockConn) Model() *model.Conn {
	return nil
}

func (c *mockConn) Expired() bool {
	return c.dead
}

func (c *mockConn) Read(ctx context.Context, responseTo int32) (msg.Response, error) {
	return nil, nil
}

func (c *mockConn) Write(ctx context.Context, reqs ...msg.Request) error {
	return nil
}
