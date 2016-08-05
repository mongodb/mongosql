package schema

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestSchema(t *testing.T) {
	var testSchemaData = []byte(
		`
schema:
-
  db: test1
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar
-
  db: test2
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
     -
        Name: b
        MongoType: int
        SqlType: int
     pipeline:
     - $unwind : "$x"
     - $sort: { a: 1, b: 1, c: -1 }
     - $project: { a: 1, b: 1, c: { $add: ["$a", { $numberLong: "10" }] } }

  -
     table: bar2
     collection: bar2
`)

	Convey("Schema should parse correctly", t, func() {

		cfg, err := New(testSchemaData)
		So(err, ShouldBeNil)

		So(len(cfg.RawDatabases), ShouldEqual, 2)
		So(len(cfg.RawDatabases[0].RawTables), ShouldEqual, 1)
		So(len(cfg.RawDatabases[1].RawTables), ShouldEqual, 2)
		So(cfg.RawDatabases[0].Name, ShouldEqual, "test1")

		So(cfg.RawDatabases[0].RawTables[0].Name, ShouldEqual, "foo")
		So(cfg.RawDatabases[0].RawTables[0].CollectionName, ShouldEqual, "foo")
		So(cfg.Databases["test1"].Name, ShouldEqual, "test1")
		So(cfg.Databases["test1"].Tables["foo"].CollectionName, ShouldEqual, "foo")

		So(len(cfg.Databases["test1"].Tables["foo"].RawColumns), ShouldEqual, 2)

		So(cfg.Databases["test1"].Tables["foo"].RawColumns[0].SqlName, ShouldEqual, "a")

		So(cfg.Databases["test2"].Tables["bar"].Pipeline, ShouldResemble, []bson.D{
			bson.D{{"$unwind", "$x"}},
			bson.D{{"$sort", bson.D{{"a", int64(1)}, {"b", int64(1)}, {"c", int64(-1)}}}},
			bson.D{{"$project", bson.D{{"a", int64(1)}, {"b", int64(1)}, {"c", bson.D{{"$add", []interface{}{"$a", int64(10)}}}}}}}})
	})
}

func TestSchemaSubdir(t *testing.T) {
	var testSchemaDataSub = []byte(
		`
schema:
-
  db: test1
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar
     -
        Name: c
        MongoType: bson.Decimal128
        SqlType: numeric
-
  db: test2
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
     -
        Name: b
        MongoType: int
        SqlType: int
     pipeline:
     - $unwind : "$x"
     - $sort: { a: 1, b: 1, c: -1 }
-
  db: test3
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: c
        MongoType: string
        SqlType: varchar
     -
        Name: d
        MongoType: geo.2darray
        SqlName: d
        SqlType: numeric[]
  -
     table: bar2
     collection: bar2
`)

	Convey("Schema should parse correctly", t, func() {
		cfg := &Schema{}

		err := cfg.Load(testSchemaDataSub)
		So(err, ShouldBeNil)

		So(len(cfg.RawDatabases), ShouldEqual, 3)
		So(len(cfg.RawDatabases[0].RawTables), ShouldEqual, 1)
		So(len(cfg.RawDatabases[1].RawTables), ShouldEqual, 1)
		So(len(cfg.RawDatabases[2].RawTables), ShouldEqual, 2)
		So(len(cfg.RawDatabases[2].RawTables[0].RawColumns), ShouldEqual, 3)
		So(len(cfg.RawDatabases[2].RawTables[1].RawColumns), ShouldEqual, 0)

		So(cfg.RawDatabases[2].RawTables[0].RawColumns[0].SqlName, ShouldEqual, "c")
		So(cfg.RawDatabases[2].RawTables[0].RawColumns[1].SqlName, ShouldEqual, "d_longitude")
		So(cfg.RawDatabases[2].RawTables[0].RawColumns[2].SqlName, ShouldEqual, "d_latitude")

		So(cfg.RawDatabases[0].Name, ShouldEqual, "test1")

		So(cfg.RawDatabases[0].RawTables[0].Name, ShouldEqual, "foo")
		So(cfg.RawDatabases[0].RawTables[0].CollectionName, ShouldEqual, "foo")
		So(cfg.Databases["test1"].Name, ShouldEqual, "test1")
		So(cfg.Databases["test1"].Tables["foo"].CollectionName, ShouldEqual, "foo")

		So(len(cfg.Databases["test1"].Tables["foo"].RawColumns), ShouldEqual, 3)

		So(cfg.Databases["test1"].Tables["foo"].RawColumns[0].SqlName, ShouldEqual, "a")

		So(cfg.Databases["test2"].Tables["bar"].Pipeline, ShouldResemble, []bson.D{
			bson.D{{"$unwind", "$x"}},
			bson.D{{"$sort", bson.D{{"a", int64(1)}, {"b", int64(1)}, {"c", int64(-1)}}}}})
	})
}

func TestSchemaSubdir2(t *testing.T) {
	var testSchemaDataRoot = []byte(
		`
schema:
-
  db: test1
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar
-
  db: test3
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	var testSchemaDataSub = []byte(
		`
schema:
-
  db: test2
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
     -
        Name: b
        MongoType: int
        SqlType: int
-
  db: test3
  tables:
  -
     table: bar
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	cfg, err := New(testSchemaDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.Load(testSchemaDataSub)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Databases["test1"] == nil {
		t.Fatal("where is test1")
	}
	if cfg.Databases["test2"] == nil {
		t.Fatal("where is test2")
	}
	if cfg.Databases["test3"] == nil {
		t.Fatal("where is test3")
	}

	if len(cfg.RawDatabases) != 3 {
		t.Fatal(cfg)
	}

	if len(cfg.Databases["test1"].RawTables) != 1 {
		t.Fatal("test1 wrong")
	}

	if len(cfg.Databases["test2"].RawTables) != 1 {
		t.Fatal("test3 wrong")
	}

	if len(cfg.Databases["test3"].RawTables) != 2 {
		t.Fatal("test3 wrong")
	}

}

func TestSchemaSubdirConflict(t *testing.T) {
	var testSchemaDataRoot = []byte(
		`
schema:
-
  db: test3
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	var testSchemaDataSub = []byte(
		`
schema:
-
  db: test3
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	cfg, err := New(testSchemaDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.Load(testSchemaDataSub)
	if err == nil {
		t.Fatal("should have conflicted")
	}

}

func TestReadFile(t *testing.T) {
	cfg := &Schema{}
	err := cfg.LoadFile("test_data/foo.conf")
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.LoadDir("test_data/sub")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.RawDatabases) != 3 {
		t.Fatalf("num RawDatabases wrong: %d", len(cfg.RawDatabases))
	}
}
