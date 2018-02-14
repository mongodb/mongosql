package mongodb_test

import (
	"context"
	"os"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testMongoHost = "127.0.0.1"
	testMongoPort = "27017"
)

var (
	lgr = log.GlobalLogger()
)

func getSslOpts() *toolsoptions.SSL {
	sslOpts := &toolsoptions.SSL{}

	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		return &toolsoptions.SSL{
			UseSSL:              true,
			SSLPEMKeyFile:       "../testdata/resources/x509gen/client.pem",
			SSLAllowInvalidCert: true,
		}
	}

	return sslOpts
}

func TestLoadInfo(t *testing.T) {
	Convey("Subject: LoadInfo", t, func() {
		cfg := config.Default()

		sslOpts := getSslOpts()

		cfg.MongoDB.Net.SSL.Enabled = sslOpts.UseSSL
		cfg.MongoDB.Net.SSL.AllowInvalidCertificates = sslOpts.SSLAllowInvalidCert
		cfg.MongoDB.Net.SSL.PEMKeyFile = sslOpts.SSLPEMKeyFile

		sp, err := mongodb.NewSqldSessionProvider(cfg)
		So(err, ShouldBeNil)

		s, err := sp.Session(context.Background())
		So(err, ShouldBeNil)
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

		err = s.Run("mongodb_info_test", bson.D{
			{"create", "one"},
		}, &struct{}{})
		So(err, ShouldBeNil)
		err = s.Run("mongodb_info_test", bson.D{
			{"create", "two"},
			{"collation", bson.M{"locale": "fr"}},
		}, &struct{}{})
		So(err, ShouldBeNil)

		sch, err := schema.New([]byte(schemaString), &lgr)
		So(err, ShouldBeNil)

		logger := log.GlobalLogger()
		info, err := mongodb.LoadInfo(&logger, sp, s, sch, false)
		So(err, ShouldBeNil)

		So(info.Privileges, ShouldEqual, mongodb.AllPrivileges)

		So(len(info.Databases), ShouldBeGreaterThanOrEqualTo, 1)

		dbInfo, ok := info.Databases[mongodb.DatabaseName("mongodb_info_test")]
		So(ok, ShouldBeTrue)
		So(len(dbInfo.Collections), ShouldEqual, 2)
		So(dbInfo.Privileges, ShouldEqual, mongodb.AllPrivileges)

		one, ok := dbInfo.Collections[mongodb.CollectionName("one")]
		So(ok, ShouldBeTrue)
		So(one.Privileges, ShouldEqual, mongodb.AllPrivileges)
		So(one.Collation, ShouldBeNil)

		two, ok := dbInfo.Collections[mongodb.CollectionName("two")]
		So(ok, ShouldBeTrue)
		So(two.Privileges, ShouldEqual, mongodb.AllPrivileges)
		So(two.Collation, ShouldNotBeNil)
		if info.VersionAtLeast(3, 3) {
			So(two.Collation.Locale, ShouldEqual, "fr")
		}
	})
}

func TestVersionAtLeast(t *testing.T) {
	Convey("Subject: VersionAtLeast", t, func() {

		info := &mongodb.Info{
			VersionArray: []uint8{3, 2, 1},
		}

		So(info.VersionAtLeast(3, 2, 1), ShouldBeTrue)
		So(info.VersionAtLeast(3, 2, 2), ShouldBeFalse)
		So(info.VersionAtLeast(3, 3, 0), ShouldBeFalse)
		So(info.VersionAtLeast(4, 0, 0), ShouldBeFalse)
		So(info.VersionAtLeast(4, 4, 4), ShouldBeFalse)
		So(info.VersionAtLeast(3, 2, 0), ShouldBeTrue)
		So(info.VersionAtLeast(3, 0, 2), ShouldBeTrue)
		So(info.VersionAtLeast(2, 3, 3), ShouldBeTrue)

		Convey("With Compatible Version", func() {
			info = &mongodb.Info{
				VersionArray: []uint8{3, 0, 0},
			}
			info.SetCompatibleVersion("3.2.1")

			So(info.VersionAtLeast(3, 2, 1), ShouldBeTrue)
			So(info.VersionAtLeast(3, 2, 2), ShouldBeFalse)
			So(info.VersionAtLeast(3, 3, 0), ShouldBeFalse)
			So(info.VersionAtLeast(4, 0, 0), ShouldBeFalse)
			So(info.VersionAtLeast(4, 4, 4), ShouldBeFalse)
			So(info.VersionAtLeast(3, 2, 0), ShouldBeTrue)
			So(info.VersionAtLeast(3, 0, 2), ShouldBeTrue)
			So(info.VersionAtLeast(2, 3, 3), ShouldBeTrue)
		})

	})

}
