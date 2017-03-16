package mongodb_test

import (
	"context"
	"os"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/client"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/schema"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testMongoHost = "127.0.0.1"
	testMongoPort = "27017"
)

func getSslOpts() *toolsoptions.SSL {
	sslOpts := &toolsoptions.SSL{}

	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		return &toolsoptions.SSL{
			UseSSL:              true,
			SSLPEMKeyFile:       "../testdata/resources/client.pem",
			SSLAllowInvalidCert: true,
		}
	}

	return sslOpts
}

func TestLoadInfo(t *testing.T) {
	Convey("Subject: LoadInfo", t, func() {
		opts, err := options.NewSqldOptions()
		So(err, ShouldBeNil)
		options.EnsureOptsNotNil(&opts)

		sslOpts := getSslOpts()

		*opts.MongoSSL = sslOpts.UseSSL
		*opts.MongoPEMKeyFile = sslOpts.SSLPEMKeyFile
		*opts.MongoAllowInvalidCerts = sslOpts.SSLAllowInvalidCert

		sp, err := client.NewSqldSessionProvider(opts)
		So(err, ShouldBeNil)

		s, err := sp.GetSession(context.Background())
		So(err, ShouldBeNil)
		defer s.Close()

		dbutils.DropDatabase(s.SelectedServer(), "mongodb_info_test")
		defer dbutils.DropDatabase(s.SelectedServer(), "mongodb_info_test")

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

		sch, err := schema.New([]byte(schemaString))
		So(err, ShouldBeNil)

		logger := log.GlobalLogger()
		info, err := mongodb.LoadInfo(&logger, s, sch, false)
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
