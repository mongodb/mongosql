package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo/address"
	"go.mongodb.org/mongo-driver/mongo/description"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

func TestNewSessionConnPool(t *testing.T) {
	t.Run("creates max connections", func(t *testing.T) {
		expectedMax := 3
		count := 0
		provider := func(context.Context) (driver.Connection, error) {
			count++
			return &mockConn{}, nil
		}

		_, err := NewSessionConnPool(context.Background(), provider, expectedMax)
		if err != nil {
			t.Fatalf("unexpected error creating pool: %v", err)
		}

		if count != expectedMax {
			t.Fatalf("expected pool to checkout %d connections eagerly, but only have %d", expectedMax, count)
		}
	})

	t.Run("returns an error when getting a connection errors", func(t *testing.T) {
		provider := func(context.Context) (driver.Connection, error) {
			return nil, fmt.Errorf("AHAHAHAH")
		}

		_, err := NewSessionConnPool(context.Background(), provider, 3)
		if err == nil {
			t.Fatal("expected an error, but got none")
		}
	})
}

func TestSessionConnPool_Get(t *testing.T) {
	t.Run("succeeds to get max connections", func(t *testing.T) {
		expectedMax := 5

		provider := func(context.Context) (driver.Connection, error) {
			return &mockConn{}, nil
		}

		p, err := NewSessionConnPool(context.Background(), provider, expectedMax)
		if err != nil {
			t.Fatalf("unexpected error creating pool: %v", err)
		}

		for i := 0; i < expectedMax; i++ {
			_, err = p.Get(context.Background())
			if err != nil {
				t.Fatalf("failed to get conn %d out of %d: %v", i+1, expectedMax, err)
			}
		}
	})

	t.Run("succeeds to get connections after they are closed and returned", func(t *testing.T) {
		expectedMax := 3

		provider := func(context.Context) (driver.Connection, error) {
			return &mockConn{}, nil
		}

		p, err := NewSessionConnPool(context.Background(), provider, expectedMax)
		if err != nil {
			t.Fatalf("unexpected error creating pool: %v", err)
		}

		conns := make([]driver.Connection, expectedMax)

		// check out all connections
		for i := 0; i < expectedMax; i++ {
			conns[i], err = p.Get(context.Background())
			if err != nil {
				t.Fatalf("failed to get conn %d out of %d: %v", i+1, expectedMax, err)
			}
		}

		// cannot check out more than that (cancel context to stop waiting)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(1 * time.Second)
			cancel()
		}()

		_, err = p.Get(ctx)
		if err == nil {
			t.Fatal("expected an error, but got none")
		}

		// close all connections
		for i, c := range conns {
			err = c.Close()
			if err != nil {
				t.Fatalf("unexpected error when closing connection %d: %v", i, err)
			}
		}

		// and finally successfully check out connections again
		for i := 0; i < expectedMax; i++ {
			_, err = p.Get(context.Background())
			if err != nil {
				t.Fatalf("failed to re-get conn %d out of %d: %v", i+1, expectedMax, err)
			}
		}
	})

	t.Run("succeeds to get a connection when one is returned", func(t *testing.T) {
		expectedMax := 3

		provider := func(context.Context) (driver.Connection, error) {
			return &mockConn{}, nil
		}

		p, err := NewSessionConnPool(context.Background(), provider, expectedMax)
		if err != nil {
			t.Fatalf("unexpected error creating pool: %v", err)
		}

		conns := make([]driver.Connection, expectedMax)

		// check out all connections
		for i := 0; i < expectedMax; i++ {
			conns[i], err = p.Get(context.Background())
			if err != nil {
				t.Fatalf("failed to get conn %d out of %d: %v", i+1, expectedMax, err)
			}
		}

		errChan := make(chan error, 1)
		// cannot check out more than that (cancel context to stop waiting)
		go func() {
			time.Sleep(1 * time.Second)
			err = conns[0].Close()
			if err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()

		err = <-errChan
		if err != nil {
			t.Fatalf("failed to close connection: %v", err)
		}

		_, err = p.Get(context.Background())
		if err != nil {
			t.Fatalf("failed to get connection after one became available: %v", err)
		}
	})

	t.Run("returns error when context is cancelled", func(t *testing.T) {
		provider := func(_ context.Context) (driver.Connection, error) {
			return &mockConn{}, nil
		}

		p, err := NewSessionConnPool(context.Background(), provider, 2)
		if err != nil {
			t.Fatalf("unexpected error creating pool: %v", err)
		}

		_, err = p.Get(context.Background())
		if err != nil {
			t.Fatalf("failed to get connection: %v", err)
		}
		_, err = p.Get(context.Background())
		if err != nil {
			t.Fatalf("failed to get connection: %v", err)
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
	})

	t.Run("returns an error when pool is closed", func(t *testing.T) {
		provider := func(context.Context) (driver.Connection, error) {
			return &mockConn{}, nil
		}

		p, err := NewSessionConnPool(context.Background(), provider, 2)
		if err != nil {
			t.Fatalf("unexpected error creating pool: %v", err)
		}

		p.Close()

		_, err = p.Get(context.Background())
		if err == nil {
			t.Fatal("expected an error, but got none")
		}

		if err != topology.ErrPoolDisconnected {
			t.Fatalf("expected error '%v' but got '%v'", topology.ErrPoolDisconnected, err)
		}
	})

	t.Run("returns an error when pool is closed after connections are checked out", func(t *testing.T) {
		provider := func(_ context.Context) (driver.Connection, error) {
			return &mockConn{}, nil
		}

		p, err := NewSessionConnPool(context.Background(), provider, 2)
		if err != nil {
			t.Fatalf("unexpected error creating pool: %v", err)
		}

		_, err = p.Get(context.Background())
		if err != nil {
			t.Fatalf("failed to get connection: %v", err)
		}
		_, err = p.Get(context.Background())
		if err != nil {
			t.Fatalf("failed to get connection: %v", err)
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
	})
}

type mockConn struct {
}

func (c *mockConn) WriteWireMessage(context.Context, []byte) error {
	return nil
}

func (c *mockConn) ReadWireMessage(ctx context.Context, dst []byte) ([]byte, error) {
	return nil, nil
}

func (c *mockConn) Description() description.Server {
	return description.Server{}
}

func (c *mockConn) Close() error {
	return nil
}

func (c *mockConn) ID() string {
	return ""
}

func (c *mockConn) Address() address.Address {
	return ""
}

func (c *mockConn) Stale() bool {
	return false
}
