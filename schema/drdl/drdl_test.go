package drdl_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/schema/drdl"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
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

		So(cfg.Databases[1].Tables[0].Pipeline, ShouldResemble, bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$unwind", "$x")),
			bsonutil.NewD(bsonutil.NewDocElem("$sort", bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("b", int32(1)),
				bsonutil.NewDocElem("c", int32(-1)),
			)),
			),
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("b", int32(1)),
				bsonutil.NewDocElem("c", bsonutil.NewD(
					bsonutil.NewDocElem("$add", bsonutil.NewArray(
						"$a",
						int64(10),
					)),
				)),
			)),
			),
		))
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

		So(cfg.Databases[1].Tables[0].Pipeline, ShouldResemble, bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$unwind", "$x")),
			bsonutil.NewD(bsonutil.NewDocElem("$sort", bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("b", int32(1)),
				bsonutil.NewDocElem("c", int32(-1)),
			)),
			),
		))
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

func TestReadSchemaWithInvalidKey(t *testing.T) {
	req := require.New(t)

	var testSchemaWithInvalidKey = []byte(
		`
schema:
-
  db: test
  invalidkey:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
`)

	_, err := drdl.NewFromBytes(testSchemaWithInvalidKey)
	req.Equal(err, fmt.Errorf("unable to map key \"invalidkey\" to a struct field at line 5, column 2"))
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

func TestTable_MarshalBSON(t *testing.T) {
	type mongoStorableTable struct {
		SQLName   string         `bson:"sql_name"`
		MongoName string         `bson:"mongo_name"`
		Pipeline  string         `bson:"pipeline"`
		Columns   []*drdl.Column `bson:"columns"`
	}

	tests := []struct {
		name     string
		table    *drdl.Table
		expected mongoStorableTable
	}{
		{
			"table with empty pipeline and no columns",
			&drdl.Table{
				SQLName:   "foo",
				MongoName: "foo",
				Pipeline:  []bson.D{},
				Columns:   []*drdl.Column{},
			},
			mongoStorableTable{
				SQLName:   "foo",
				MongoName: "foo",
				Pipeline:  "[]",
				Columns:   []*drdl.Column{},
			},
		},
		{
			"table with empty pipeline and some columns",
			&drdl.Table{
				SQLName:   "foo",
				MongoName: "foo",
				Pipeline:  []bson.D{},
				Columns: []*drdl.Column{
					{"a", "int", "a", "int32"},
					{"a", "int", "a", "int32"},
				},
			},
			mongoStorableTable{
				SQLName:   "foo",
				MongoName: "foo",
				Pipeline:  "[]",
				Columns: []*drdl.Column{
					{"a", "int", "a", "int32"},
					{"a", "int", "a", "int32"},
				},
			},
		},
		{
			"table with pipeline and columns",
			&drdl.Table{
				SQLName:   "foo",
				MongoName: "foo",
				Pipeline:  []bson.D{{bson.E{Key: "$unwind", Value: "$x"}}},
				Columns: []*drdl.Column{
					{"a", "int", "a", "int32"},
					{"a", "int", "a", "int32"},
				},
			},
			mongoStorableTable{
				SQLName:   "foo",
				MongoName: "foo",
				Pipeline:  `[{"$unwind":"$x"}]`,
				Columns: []*drdl.Column{
					{"a", "int", "a", "int32"},
					{"a", "int", "a", "int32"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.table.MarshalBSON()
			if err != nil {
				t.Fatalf("failed to marshal actual table: %v", err)
			}

			expected, err := bson.Marshal(test.expected)
			if err != nil {
				t.Fatalf("failed to marshal expected table: %v", err)
			}

			if len(actual) != len(expected) {
				t.Fatalf("actual != expected:\nlen(actual) = len(expected) (%v != %v)\n", len(actual), len(expected))
			}

			for i, b := range actual {
				if b != expected[i] {
					t.Fatalf("actual != expected:\nactual[%v] != expected[%v] (%v != %v)\nactual:   %v\nexpected: %v\n", i, i, b, expected[i], actual, expected)
				}
			}
		})
	}
}
