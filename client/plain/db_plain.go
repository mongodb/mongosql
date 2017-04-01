package plain

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/cluster"
	"github.com/10gen/mongo-go-driver/readpref"

	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/options"
)

// PlainDBConnector is a connector for dialing the database, without SSL.
type PlainDBConnector struct {
	dialInfo *mongodb.DialInfo
}

func (v *PlainDBConnector) ConfigureDrdl(_ options.DrdlOptions, dialInfo *mongodb.DialInfo) error {
	v.dialInfo = dialInfo
	return nil
}

func (v *PlainDBConnector) ConfigureSqld(_ options.SqldOptions, dialInfo *mongodb.DialInfo) error {
	v.dialInfo = dialInfo
	return v.validate()
}

func (v *PlainDBConnector) GetNewSession(ctx context.Context, monitor *cluster.Monitor, readPreference *readpref.ReadPref) (*mongodb.Session, error) {
	session, err := v.dialInfo.Dial(ctx, monitor, readPreference)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (v *PlainDBConnector) validate() error {
	if v.dialInfo.Username != "" || v.dialInfo.Password != "" || v.dialInfo.AuthSource != "" {
		return fmt.Errorf("--mongo-uri may not contain any authentication information")
	}
	if v.dialInfo.Database != "" {
		return fmt.Errorf("--mongo-uri may not contain database name")
	}
	if v.dialInfo.AuthMechanism != "" {
		return fmt.Errorf("--mongo-uri may not contain any authentication mechanism")
	}
	if len(v.dialInfo.AuthMechanismProperties) != 0 {
		return fmt.Errorf("--mongo-uri may not contain any authentication mechanism properties")
	}
	return nil
}
