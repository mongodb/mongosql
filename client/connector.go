package client

import (
	"github.com/10gen/sqlproxy/options"
	"gopkg.in/mgo.v2"
)

// DBConnector defines an interface for connecting to the database.
type DBConnector interface {
	ConfigureDrdl(options.DrdlOptions) error
	ConfigureSqld(options.SqldOptions) error
	GetNewSession() (*mgo.Session, error)
}
