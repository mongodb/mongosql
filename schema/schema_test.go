package schema

import (
	"fmt"
	"testing"

	"gopkg.in/mgo.v2/bson"

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
     table: bar2
     collection: bar2
`)

	Convey("Schema should parse correctly", t, func() {
		cfg := &Schema{}

		err := cfg.Load(testSchemaDataSub)
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

func TestCanCompare(t *testing.T) {

	Convey("Subject: CanCompare", t, func() {

		type test struct {
			left         SQLType
			right        SQLType
			incomparable bool
		}

		runTests := func(tests []test) {
			for _, t := range tests {
				var incomparable string
				if !t.incomparable {
					incomparable = "not "
				}
				Convey(fmt.Sprintf("comparison between '%v' and '%v' should %vreturn an error", t.left, t.right, incomparable), func() {
					canCompare := CanCompare(t.left, t.right)
					if t.incomparable {
						So(canCompare, ShouldBeFalse)
					} else {
						So(canCompare, ShouldBeTrue)
					}
				})
			}
		}

		Convey("Subject: SQLInt", func() {
			tests := []test{
				{SQLInt, SQLInt, false},
				{SQLInt, SQLFloat, false},
				{SQLInt, SQLBoolean, false},
				{SQLInt, SQLNull, false},
				{SQLInt, SQLObjectID, true},
				{SQLInt, SQLVarchar, true},
				{SQLInt, SQLNone, false},
				{SQLInt, SQLDate, true},
				{SQLInt, SQLTimestamp, true},
			}
			runTests(tests)
		})

		Convey("Subject: SQLFloat", func() {
			tests := []test{
				{SQLFloat, SQLInt, false},
				{SQLFloat, SQLBoolean, false},
				{SQLFloat, SQLNull, false},
				{SQLFloat, SQLObjectID, true},
				{SQLFloat, SQLVarchar, true},
				{SQLFloat, SQLFloat, false},
				{SQLFloat, SQLNone, false},
				{SQLFloat, SQLDate, true},
				{SQLFloat, SQLTimestamp, true},
			}
			runTests(tests)
		})

		Convey("Subject: SQLBool", func() {
			tests := []test{
				{SQLBoolean, SQLFloat, false},
				{SQLBoolean, SQLNull, false},
				{SQLBoolean, SQLObjectID, true},
				{SQLBoolean, SQLVarchar, true},
				{SQLBoolean, SQLInt, false},
				{SQLBoolean, SQLBoolean, false},
				{SQLBoolean, SQLNone, false},
				{SQLBoolean, SQLDate, true},
				{SQLBoolean, SQLTimestamp, true},
			}
			runTests(tests)
		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				{SQLDate, SQLInt, true},
				{SQLDate, SQLFloat, true},
				{SQLDate, SQLBoolean, true},
				{SQLDate, SQLNull, false},
				{SQLDate, SQLObjectID, true},
				{SQLDate, SQLVarchar, false},
				{SQLDate, SQLNone, false},
				{SQLDate, SQLDate, false},
				{SQLDate, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				{SQLTimestamp, SQLInt, true},
				{SQLTimestamp, SQLFloat, true},
				{SQLTimestamp, SQLBoolean, true},
				{SQLTimestamp, SQLNull, false},
				{SQLTimestamp, SQLObjectID, true},
				{SQLTimestamp, SQLVarchar, false},
				{SQLTimestamp, SQLDate, false},
				{SQLTimestamp, SQLNone, false},
				{SQLTimestamp, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLNullValue", func() {
			tests := []test{
				{SQLNull, SQLInt, false},
				{SQLNull, SQLFloat, false},
				{SQLNull, SQLBoolean, false},
				{SQLNull, SQLObjectID, false},
				{SQLNull, SQLVarchar, false},
				{SQLNull, SQLNone, false},
				{SQLNull, SQLDate, false},
				{SQLNull, SQLTimestamp, false},
				{SQLNull, SQLNull, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				{SQLVarchar, SQLInt, true},
				{SQLVarchar, SQLFloat, true},
				{SQLVarchar, SQLBoolean, true},
				{SQLVarchar, SQLObjectID, true},
				{SQLVarchar, SQLVarchar, false},
				{SQLVarchar, SQLNone, false},
				{SQLVarchar, SQLDate, false},
				{SQLVarchar, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLObjectID", func() {

			tests := []test{
				{SQLObjectID, SQLInt, true},
				{SQLObjectID, SQLFloat, true},
				{SQLObjectID, SQLVarchar, false},
				{SQLObjectID, SQLBoolean, true},
				{SQLObjectID, SQLNone, false},
				{SQLObjectID, SQLDate, true},
				{SQLObjectID, SQLTimestamp, true},
				{SQLObjectID, SQLObjectID, false},
			}
			runTests(tests)
		})

	})
}

func TestIsSimilar(t *testing.T) {

	Convey("Subject: IsSimilar", t, func() {

		type test struct {
			left      SQLType
			right     SQLType
			isSimilar bool
		}

		runTests := func(tests []test) {
			for _, t := range tests {
				var isSimilar string
				if !t.isSimilar {
					isSimilar = "not "
				}
				Convey(fmt.Sprintf("similarity between '%v' and '%v' should %vbe true", t.left, t.right, isSimilar), func() {
					is := IsSimilar(t.left, t.right)
					if t.isSimilar {
						So(is, ShouldBeTrue)
					} else {
						So(is, ShouldBeFalse)
					}
				})
			}
		}

		Convey("Subject: SQLInt", func() {
			tests := []test{
				{SQLInt, SQLInt, true},
				{SQLInt, SQLFloat, true},
				{SQLInt, SQLInt64, true},
				{SQLInt, SQLArrNumeric, true},
				{SQLInt, SQLBoolean, false},
				{SQLInt, SQLNull, false},
				{SQLInt, SQLObjectID, false},
				{SQLInt, SQLVarchar, false},
				{SQLInt, SQLNone, false},
				{SQLInt, SQLDate, false},
				{SQLInt, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLFloat", func() {
			tests := []test{
				{SQLFloat, SQLInt, true},
				{SQLFloat, SQLInt64, true},
				{SQLFloat, SQLArrNumeric, true},
				{SQLFloat, SQLFloat, true},
				{SQLFloat, SQLBoolean, false},
				{SQLFloat, SQLNull, false},
				{SQLFloat, SQLObjectID, false},
				{SQLFloat, SQLVarchar, false},
				{SQLFloat, SQLNone, false},
				{SQLFloat, SQLDate, false},
				{SQLFloat, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLArrNumeric", func() {
			tests := []test{
				{SQLArrNumeric, SQLInt, true},
				{SQLArrNumeric, SQLInt64, true},
				{SQLArrNumeric, SQLArrNumeric, true},
				{SQLArrNumeric, SQLFloat, true},
				{SQLArrNumeric, SQLBoolean, false},
				{SQLArrNumeric, SQLNull, false},
				{SQLArrNumeric, SQLObjectID, false},
				{SQLArrNumeric, SQLVarchar, false},
				{SQLArrNumeric, SQLNone, false},
				{SQLArrNumeric, SQLDate, false},
				{SQLArrNumeric, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLInt64", func() {
			tests := []test{
				{SQLInt64, SQLInt, true},
				{SQLInt64, SQLInt64, true},
				{SQLInt64, SQLArrNumeric, true},
				{SQLInt64, SQLFloat, true},
				{SQLInt64, SQLBoolean, false},
				{SQLInt64, SQLNull, false},
				{SQLInt64, SQLObjectID, false},
				{SQLInt64, SQLVarchar, false},
				{SQLInt64, SQLNone, false},
				{SQLInt64, SQLDate, false},
				{SQLInt64, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLBool", func() {
			tests := []test{
				{SQLBoolean, SQLFloat, false},
				{SQLBoolean, SQLNull, false},
				{SQLBoolean, SQLObjectID, false},
				{SQLBoolean, SQLVarchar, false},
				{SQLBoolean, SQLInt, false},
				{SQLBoolean, SQLBoolean, false},
				{SQLBoolean, SQLNone, false},
				{SQLBoolean, SQLDate, false},
				{SQLBoolean, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				{SQLDate, SQLInt, false},
				{SQLDate, SQLFloat, false},
				{SQLDate, SQLBoolean, false},
				{SQLDate, SQLNull, false},
				{SQLDate, SQLObjectID, false},
				{SQLDate, SQLVarchar, false},
				{SQLDate, SQLNone, false},
				{SQLDate, SQLDate, true},
				{SQLDate, SQLTimestamp, true},
			}
			runTests(tests)
		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				{SQLTimestamp, SQLInt, false},
				{SQLTimestamp, SQLFloat, false},
				{SQLTimestamp, SQLBoolean, false},
				{SQLTimestamp, SQLNull, false},
				{SQLTimestamp, SQLObjectID, false},
				{SQLTimestamp, SQLVarchar, false},
				{SQLTimestamp, SQLDate, true},
				{SQLTimestamp, SQLNone, false},
				{SQLTimestamp, SQLTimestamp, true},
			}
			runTests(tests)
		})

		Convey("Subject: SQLNullValue", func() {
			tests := []test{
				{SQLNull, SQLInt, false},
				{SQLNull, SQLFloat, false},
				{SQLNull, SQLBoolean, false},
				{SQLNull, SQLObjectID, false},
				{SQLNull, SQLVarchar, false},
				{SQLNull, SQLNone, false},
				{SQLNull, SQLDate, false},
				{SQLNull, SQLTimestamp, false},
				{SQLNull, SQLNull, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				{SQLVarchar, SQLInt, false},
				{SQLVarchar, SQLFloat, false},
				{SQLVarchar, SQLBoolean, false},
				{SQLVarchar, SQLObjectID, false},
				{SQLVarchar, SQLVarchar, false},
				{SQLVarchar, SQLNone, false},
				{SQLVarchar, SQLDate, false},
				{SQLVarchar, SQLTimestamp, false},
			}
			runTests(tests)
		})

		Convey("Subject: SQLObjectID", func() {

			tests := []test{
				{SQLObjectID, SQLInt, false},
				{SQLObjectID, SQLFloat, false},
				{SQLObjectID, SQLVarchar, false},
				{SQLObjectID, SQLBoolean, false},
				{SQLObjectID, SQLNone, false},
				{SQLObjectID, SQLDate, false},
				{SQLObjectID, SQLTimestamp, false},
				{SQLObjectID, SQLObjectID, false},
			}
			runTests(tests)
		})

	})
}
