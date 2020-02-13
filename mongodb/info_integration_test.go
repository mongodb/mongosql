//+build integration

package mongodb_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	mongoutil "github.com/10gen/sqlproxy/internal/testutil/mongodb"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodb/provider"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	toolsoptions "github.com/mongodb/mongo-tools-common/options"
)

var (
	lgr = log.GlobalLogger()
)

func getSslOpts() *toolsoptions.SSL {
	sslOpts := &toolsoptions.SSL{}

	if len(os.Getenv(mongoutil.SSLTestKey)) > 0 {
		return &toolsoptions.SSL{
			UseSSL:              true,
			SSLPEMKeyFile:       "../testdata/resources/x509gen/client.pem",
			SSLAllowInvalidCert: true,
		}
	}

	return sslOpts
}

func TestLoadInfo(t *testing.T) {
	req := require.New(t)

	cfg := config.Default()

	sslOpts := getSslOpts()

	cfg.Security.Enabled = false
	cfg.MongoDB.Net.SSL.Enabled = sslOpts.UseSSL
	cfg.MongoDB.Net.SSL.AllowInvalidCertificates = sslOpts.SSLAllowInvalidCert
	cfg.MongoDB.Net.SSL.PEMKeyFile = sslOpts.SSLPEMKeyFile

	sp, err := provider.NewSqldSessionProvider(cfg)
	req.Nil(err, "failed to get NewSqldSessionProvider")
	defer sp.Close()

	s, err := sp.Session(context.Background())
	req.NoError(err, "failed to get Session")
	defer s.Close()

	dbutils.DropDatabase(s, "mongodb_info_test")
	defer dbutils.DropDatabase(s, "mongodb_info_test")

	schemaString := `
schema:
-
  db: mongodb_info_test
  tables:
  -
    table: one
    collection: one
    columns:
    -
      Name: a
      MongoType: int
      SqlType: int
  -
    table: two
    collection: two
    columns:
    -
      Name: a
      MongoType: int
      SqlType: int
`

	err = s.Run(context.Background(), "mongodb_info_test", bsonutil.NewD(
		bsonutil.NewDocElem("create", "one"),
	), &struct{}{})
	req.NoError(err, "failed to run mongodb_info_test")
	err = s.Run(context.Background(), "mongodb_info_test", bsonutil.NewD(
		bsonutil.NewDocElem("create", "two"),
		bsonutil.NewDocElem("collation", bsonutil.NewM(bsonutil.NewDocElem("locale", "fr"))),
	), &struct{}{})
	req.NoError(err, "failed to run mongodb_info_test")

	drdlSchema, err := drdl.NewFromBytes([]byte(schemaString))
	req.NoError(err, "failed to load drdl")
	sch, err := schema.NewFromDRDL(lgr, drdlSchema)
	req.NoError(err, "failed to create schema from drdl")

	info, err := provider.LoadInfo(context.Background(), lgr, sp, s, sch, cfg)
	req.NoError(err, "failed to load info")

	req.True(len(info.Databases) >= 1,
		"there should be at least one database")

	dbInfo, ok := info.Databases[mongodb.DatabaseName("mongodb_info_test")]
	req.Equal(true, ok, "could not find collection one")
	req.Equal(len(dbInfo.Collections), 2,
		"should be two collections in mongodb_info_test")
	req.Equal(mongodb.AllPrivileges, dbInfo.Privileges,
		"mongodb_info_test privileges should be all privileges")

	one, ok := dbInfo.Collections[mongodb.CollectionName("one")]
	req.True(ok, "could not find mongodb_info_test.one")

	req.Equal(mongodb.AllPrivileges, one.Privileges,
		"mongodb_info_test.one privileges should be all privileges")
	req.Nil(one.Collation, "collation should be nil")

	two, ok := dbInfo.Collections[mongodb.CollectionName("two")]
	req.Equal(true, ok, "could not find collection two")
	req.Equal(mongodb.AllPrivileges, two.Privileges,
		"mongodb_info_test.two privileges should be all privileges")
	req.NotNil(two.Collation, "collation should not be nil")
	if info.VersionAtLeast(3, 3) {
		req.Equal(two.Collation.Locale, "fr",
			"collaction locale should be fr")
	}
}
