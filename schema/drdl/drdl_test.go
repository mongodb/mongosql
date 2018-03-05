package drdl_test

import (
	"testing"

	"github.com/10gen/sqlproxy/schema/drdl"

	"github.com/10gen/mongo-go-driver/bson"
	. "github.com/smartystreets/goconvey/convey"
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
        SQLName: a
        MongoType: string
        SqlType: varchar
     -
        Name: b
        SQLName: b
        MongoType: int
        SqlType: int
     -
        Name: c
        MongoType: number
        SqlType: numeric
     pipeline:
     - $unwind : "$x"
     - $sort: { a: 1, b: 1, c: -1 }
     - $project: { a: 1, b: 1, c: { $add: ["$a", { $numberLong: "10" }] } }

  -
     table: bar2
     collection: bar2
`)

	Convey("Schema should parse correctly", t, func() {

		cfg, err := drdl.NewFromBytes(testSchemaData)
		So(err, ShouldBeNil)

		So(len(cfg.Databases), ShouldEqual, 2)
		So(cfg.Databases[0].Name, ShouldEqual, "test1")
		So(len(cfg.Databases[0].Tables), ShouldEqual, 1)
		So(len(cfg.Databases[1].Tables), ShouldEqual, 2)

		So(cfg.Databases[0].Tables[0].SQLName, ShouldEqual, "foo")
		So(cfg.Databases[0].Tables[0].MongoName, ShouldEqual, "foo")
		So(cfg.Databases[0].Name, ShouldEqual, "test1")
		So(cfg.Databases[0].Tables[0].MongoName, ShouldEqual, "foo")

		So(len(cfg.Databases[0].Tables[0].Columns), ShouldEqual, 2)

		So(cfg.Databases[0].Tables[0].Columns[0].SQLName, ShouldEqual, "")

		So(cfg.Databases[1].Tables[0].Pipeline, ShouldResemble, []bson.D{
			{{Name: "$unwind", Value: "$x"}},
			{{Name: "$sort", Value: bson.D{
				{Name: "a", Value: int64(1)},
				{Name: "b", Value: int64(1)},
				{Name: "c", Value: int64(-1)},
			}}},
			{{Name: "$project", Value: bson.D{
				{Name: "a", Value: int64(1)},
				{Name: "b", Value: int64(1)},
				{Name: "c", Value: bson.D{
					{Name: "$add", Value: []interface{}{"$a", bson.D{
						{Name: "$numberLong", Value: "10"},
					}}},
				}},
			}}},
		})
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
        SQLName: d
        SqlType: numeric[]
  -
     table: bar2
     collection: bar2
`)

	Convey("Schema should parse correctly", t, func() {

		cfg, err := drdl.NewFromBytes(testSchemaDataSub)
		So(err, ShouldBeNil)

		So(len(cfg.Databases), ShouldEqual, 3)
		So(len(cfg.Databases[0].Tables), ShouldEqual, 1)
		So(len(cfg.Databases[1].Tables), ShouldEqual, 1)
		So(len(cfg.Databases[2].Tables), ShouldEqual, 2)
		So(len(cfg.Databases[2].Tables[0].Columns), ShouldEqual, 2)
		So(len(cfg.Databases[2].Tables[1].Columns), ShouldEqual, 0)

		So(cfg.Databases[2].Tables[0].Columns[0].MongoName, ShouldEqual, "c")
		So(cfg.Databases[2].Tables[0].Columns[1].MongoName, ShouldEqual, "d")

		So(cfg.Databases[0].Name, ShouldEqual, "test1")

		So(cfg.Databases[0].Tables[0].SQLName, ShouldEqual, "foo")
		So(cfg.Databases[0].Tables[0].MongoName, ShouldEqual, "foo")
		So(cfg.Databases[0].Name, ShouldEqual, "test1")
		So(cfg.Databases[0].Tables[0].MongoName, ShouldEqual, "foo")

		So(len(cfg.Databases[0].Tables[0].Columns), ShouldEqual, 3)

		So(cfg.Databases[0].Tables[0].Columns[0].MongoName, ShouldEqual, "a")

		So(cfg.Databases[1].Tables[0].Pipeline, ShouldResemble, []bson.D{
			{{Name: "$unwind", Value: "$x"}},
			{{Name: "$sort", Value: bson.D{
				{Name: "a", Value: int64(1)},
				{Name: "b", Value: int64(1)},
				{Name: "c", Value: int64(-1)},
			}}},
		})
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

	cfg, err := drdl.NewFromBytes(testSchemaDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.Load(testSchemaDataSub)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Databases) != 3 {
		t.Fatal(cfg)
	}

	if len(cfg.Databases[0].Tables) != 1 {
		t.Fatal("test1 wrong")
	}

	if len(cfg.Databases[1].Tables) != 2 {
		t.Fatal("test3 wrong")
	}

	if len(cfg.Databases[2].Tables) != 1 {
		t.Fatal("test2 wrong")
	}

}

func TestReadFile(t *testing.T) {
	cfg := &drdl.Schema{}
	err := cfg.LoadFile("testdata/foo.conf")
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.LoadDir("testdata/sub")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Databases) != 3 {
		t.Fatalf("num RawDatabases wrong: %d", len(cfg.Databases))
	}
}
