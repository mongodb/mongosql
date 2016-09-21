package plain

import (
	"fmt"
	"time"

	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/util"
	"gopkg.in/mgo.v2"
)

// PlainDBConnector is a connector for dialing the database, without SSL.
type PlainDBConnector struct {
	dialInfo *mgo.DialInfo
}

func (v *PlainDBConnector) ConfigureDrdl(opts options.DrdlOptions) error {
	connectionAddrs := util.CreateConnectionAddrs(opts.Host, opts.Port)

	timeout := time.Duration(opts.Timeout) * time.Second

	v.dialInfo = &mgo.DialInfo{
		Addrs:          connectionAddrs,
		Timeout:        timeout,
		Direct:         opts.Direct,
		ReplicaSetName: opts.ReplicaSetName,
		Username:       opts.DrdlAuth.Username,
		Password:       opts.DrdlAuth.Password,
		Source:         opts.GetAuthenticationDatabase(),
		Mechanism:      opts.DrdlAuth.Mechanism,
	}

	return nil
}

func (v *PlainDBConnector) ConfigureSqld(opts options.SqldOptions) error {
	var err error

	v.dialInfo, err = mgo.ParseURL(opts.MongoURI)
	if err != nil {
		return err
	}

	if v.dialInfo.Username != "" || v.dialInfo.Password != "" {
		return fmt.Errorf("--mongo-uri may not contain any authentication information")
	}
	if v.dialInfo.Username != "" || v.dialInfo.Password != "" || v.dialInfo.Source != "" {
		return fmt.Errorf("--mongo-uri may not contain any authentication information")
	}
	if v.dialInfo.Database != "" {
		return fmt.Errorf("--mongo-uri may not contain database name")
	}
	if v.dialInfo.Mechanism != "" {
		return fmt.Errorf("--mongo-uri may not contain any authentication mechanism")
	}
	if v.dialInfo.Service != "" {
		return fmt.Errorf("unsupported: --mongo-uri may not contain GSSAPI service name")
	}
	if v.dialInfo.ServiceHost != "" {
		return fmt.Errorf("unsupported: --mongo-uri may not contain GSSAPI hostname")
	}

	v.dialInfo.Timeout = time.Duration(opts.MongoTimeout) * time.Second

	return nil
}

func (v *PlainDBConnector) GetNewSession() (*mgo.Session, error) {
	return mgo.DialWithInfo(v.dialInfo)
}
