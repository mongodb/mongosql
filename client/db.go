// Package client implements generic connection to MongoDB, and contains
// subpackages for specific methods of connection.
package client

import (
	"fmt"
	"sync"

	"github.com/10gen/sqlproxy/client/plain"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/password"
	"gopkg.in/mgo.v2"
)

type SessionProvider struct {
	masterSessionLock sync.Mutex
	masterSession     *mgo.Session
	connector         DBConnector
	flags             sessionFlag
}

type (
	sessionFlag uint32

	GetConnectorFunc func(opts options.Options) DBConnector
)

// Session flags.
const (
	DisableSocketTimeout sessionFlag = 0
)

var (
	GetConnectorFuncs = []GetConnectorFunc{}
)

func (self *SessionProvider) GetSession() (*mgo.Session, error) {
	self.masterSessionLock.Lock()
	defer self.masterSessionLock.Unlock()

	if self.masterSession != nil {
		return self.masterSession.Copy(), nil
	}

	var err error

	self.masterSession, err = self.connector.GetNewSession()
	if err != nil {
		return nil, fmt.Errorf("error connecting to db server: %v", err)
	}

	return self.masterSession.Copy(), nil
}

func (self *SessionProvider) Close() {
	self.masterSessionLock.Lock()
	defer self.masterSessionLock.Unlock()
	self.masterSession.Close()
}

func (self *SessionProvider) refresh() {
	if (self.flags & DisableSocketTimeout) > 0 {
		self.masterSession.SetSocketTimeout(0)
	}
}

func (self *SessionProvider) SetFlags(flagBits sessionFlag) {
	self.masterSessionLock.Lock()
	defer self.masterSessionLock.Unlock()

	self.flags = flagBits

	if self.masterSession != nil {
		self.refresh()
	}
}

func NewDrdlSessionProvider(opts options.DrdlOptions) (*SessionProvider, error) {
	provider := &SessionProvider{}

	if opts.DrdlAuth.ShouldAskForPassword() {
		opts.DrdlAuth.Password = password.Prompt()
	}

	provider.connector = getConnector(opts)

	err := provider.connector.ConfigureDrdl(opts)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func NewSqldSessionProvider(opts options.SqldOptions) (*SessionProvider, error) {
	provider := &SessionProvider{}

	provider.connector = getConnector(opts)

	err := provider.connector.ConfigureSqld(opts)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func getConnector(opts options.Options) DBConnector {
	for _, getConnectorFunc := range GetConnectorFuncs {
		if connector := getConnectorFunc(opts); connector != nil {
			return connector
		}
	}
	return &plain.PlainDBConnector{}
}
