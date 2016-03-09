package schema

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"path/filepath"
	"testing"
)

func TestSchema(t *testing.T) {
	var testSchemaData = []byte(
		`
# Sample BI-Connector Schema File

# Address to listen on
addr : 127.0.0.1:4000

# Proxy user/pass
user : root
password : abc

log_level : error

schema:
- 
  db: test1
  tables:
  - 
     table: foo
     collection: test.foo
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
     collection: test.bar
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
     -
        $unwind : "$x"
     -
        $limit : 10

  -
     table: bar2
     collection: test.bar2
`)

	cfg, err := ParseSchemaData(testSchemaData)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.LogLevel != "error" || cfg.User != "root" || cfg.Password != "abc" || cfg.Addr != "127.0.0.1:4000" {
		t.Fatal("Top Schema not equal.")
	}

	if len(cfg.RawDatabases) != 2 {
		t.Fatal(cfg)
	}

	if len(cfg.RawDatabases[0].RawTables) != 1 {
		t.Fatal(len(cfg.RawDatabases[0].RawTables))
	}

	if len(cfg.RawDatabases[1].RawTables) != 2 {
		t.Fatal(len(cfg.RawDatabases[1].RawTables))
	}

	if cfg.RawDatabases[0].Name != "test1" {
		t.Fatalf("first db is wrong: %s", cfg.RawDatabases[0].Name)
	}

	if cfg.RawDatabases[0].RawTables[0].Name != "foo" || cfg.RawDatabases[0].RawTables[0].FQNS != "test.foo" {
		t.Fatal("Table 0 (bar) basics wrong")
	}

	if cfg.Databases["test1"].Name != "test1" {
		t.Fatal("map broken")
	}

	if cfg.Databases["test1"].Tables["foo"].FQNS != "test.foo" {
		t.Fatal("map broken 2")
	}

	if len(cfg.Databases["test1"].Tables["foo"].RawColumns) != 2 {
		t.Fatal("test1.foo num columns wrong")
	}

	if cfg.Databases["test1"].Tables["foo"].RawColumns[0].SqlName != "a" {
		t.Fatal("test1.foo.a name wrong")
	}

	testBar := cfg.Databases["test2"].Tables["bar"]
	if len(testBar.Pipeline) != 2 {
		t.Fatal("test2.bar pipeline is wrong length")
	}
}

func TestSchemaSubdir(t *testing.T) {
	var testSchemaDataRoot = []byte(
		`
addr : 127.0.0.1:4000

# Proxy user/pass
user : root
password : abc

log_level : error

schema_dir : foo
`)

	var testSchemaDataSub = []byte(
		`
schema:
-
  db: test1
  tables:
  - 
     table: foo
     collection: test.foo
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
     collection: test.bar
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
     -
        $unwind : "$x"
     -
        $limit : 10

  -
     table: bar2
     collection: test.bar2
`)

	cfg, err := ParseSchemaData(testSchemaDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "foo" {
		t.Fatal("SchemaDir wrong")
	}

	err = cfg.ingestSubFile(testSchemaDataSub)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.RawDatabases) != 2 {
		t.Fatal(cfg)
	}

	if len(cfg.RawDatabases[0].RawTables) != 1 {
		t.Fatal(len(cfg.RawDatabases[0].RawTables))
	}

	if len(cfg.RawDatabases[1].RawTables) != 2 {
		t.Fatal(len(cfg.RawDatabases[1].RawTables))
	}

	if cfg.RawDatabases[0].Name != "test1" {
		t.Fatalf("first db is wrong: %s", cfg.RawDatabases[0].Name)
	}

	if cfg.RawDatabases[0].RawTables[0].Name != "foo" || cfg.RawDatabases[0].RawTables[0].FQNS != "test.foo" {
		t.Fatal("Table 0 (bar) basics wrong")
	}

	if cfg.Databases["test1"].Name != "test1" {
		t.Fatal("map broken")
	}

	if cfg.Databases["test1"].Tables["foo"].FQNS != "test.foo" {
		t.Fatal("map broken 2")
	}

	if len(cfg.Databases["test1"].Tables["foo"].RawColumns) != 2 {
		t.Fatal("test1.foo num columns wrong")
	}

	if cfg.Databases["test1"].Tables["foo"].RawColumns[0].SqlName != "a" {
		t.Fatal("test1.foo.a name wrong")
	}

	testBar := cfg.Databases["test2"].Tables["bar"]
	if len(testBar.Pipeline) != 2 {
		t.Fatal("test2.bar pipeline is wrong length")
	}
}

func TestSchemaSubdir2(t *testing.T) {
	var testSchemaDataRoot = []byte(
		`
addr : 127.0.0.1:4000

# Proxy user/pass
user : root
password : abc

log_level : error

schema_dir : foo
schema:
-
  db: test1
  tables:
  - 
     table: foo
     collection: test.foo
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
     collection: test.foo
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
     collection: test.bar
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
     collection: test.foo
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

	cfg, err := ParseSchemaData(testSchemaDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "foo" {
		t.Fatal("SchemaDir wrong")
	}

	err = cfg.ingestSubFile(testSchemaDataSub)
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
addr : 127.0.0.1:4000

# Proxy user/pass
user : root
password : abc

log_level : error

schema_dir : foo
schema:
-
  db: test3
  tables:
  - 
     table: foo
     collection: test.foo
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
     collection: test.foo
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

	cfg, err := ParseSchemaData(testSchemaDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "foo" {
		t.Fatal("SchemaDir wrong")
	}

	err = cfg.ingestSubFile(testSchemaDataSub)
	if err == nil {
		t.Fatal("should have conflicted")
	}

}

func _testComputeDirectory(t *testing.T, file string, dir string, correct string) {
	output := computeDirectory(file, dir)
	if output != correct {
		t.Fatalf("computeDirectory wrong (%s) (%s) -> (%s) != (%s)(correct)", file, dir, output, correct)
	}
}

func TestComputeDirectory(t *testing.T) {
	_testComputeDirectory(t, "asd", filepath.Join("/", "a"), filepath.Join("/", "a"))
	_testComputeDirectory(t, "foo.conf", filepath.Join("/", "a"), filepath.Join("/", "a"))
	_testComputeDirectory(t, "/b/foo.conf", filepath.Join("/", "a"), filepath.Join("/", "a"))
	_testComputeDirectory(t, "/b/foo.conf", "a", filepath.Join("/", "b", "a"))
}

func TestReadFile(t *testing.T) {
	cfg, err := ParseSchemaFile("test_data/foo.conf")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "sub" {
		t.Fatalf("SchemaDir wrong: (%s)", cfg.SchemaDir)
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
