package client

import (
	"context"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/options"

	"github.com/10gen/mongo-go-driver/cluster"
	"github.com/10gen/mongo-go-driver/readpref"
)

// DBConnector defines an interface for connecting to the database.
type DBConnector interface {
	ConfigureDrdl(options.DrdlOptions, *mongodb.DialInfo) error
	ConfigureSqld(options.SqldOptions, *mongodb.DialInfo) error
	GetNewSession(context.Context, *cluster.Monitor, *readpref.ReadPref) (*mongodb.Session, error)
}
