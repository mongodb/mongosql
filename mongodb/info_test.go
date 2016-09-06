package mongodb_test

import (
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadInfo(t *testing.T) {
	Convey("Subject: LoadInfo", t, func() {

		s, err := mgo.Dial("mongodb://localhost:27017")
		So(err, ShouldBeNil)

		db := s.DB("mongodb_info_test")
		db.DropDatabase()
		defer db.DropDatabase()

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

		err = db.Run(bson.D{
			{"create", "one"},
		}, &struct{}{})
		So(err, ShouldBeNil)
		err = db.Run(bson.D{
			{"create", "two"},
			{"collation", bson.M{"locale": "fr"}},
		}, &struct{}{})
		So(err, ShouldBeNil)

		sch, err := schema.New([]byte(schemaString))
		So(err, ShouldBeNil)

		info, err := mongodb.LoadInfo(s, sch, false)
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

		if info.VersionAtLeast(3, 3) {
			So(two.Collation, ShouldNotBeNil)
			So(two.Collation.Locale, ShouldEqual, "fr")
		} else {
			So(two.Collation, ShouldBeNil)
		}
	})
}
