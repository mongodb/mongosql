package schema

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
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

	cfg, err := New(testSchemaData)
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

func TestSchemaSubdir(t *testing.T) {
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

	cfg := &Schema{}

	err := cfg.Load(testSchemaDataSub)
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

	type test struct {
		value  interface{}
		mType  MongoType
		result bool
	}

	runTests := func(tests []test) {
		for _, t := range tests {
			Convey(fmt.Sprintf("can compare %v to %v: %v", t.value, t.mType, t.result), func() {
				So(CanCompare(t.value, t.mType), ShouldEqual, t.result)
			})
		}
	}

	timeObject := time.Now()

	Convey("Subject: Numeric Comparison", t, func() {
		tests := []test{
			test{66, MongoInt, true},
			test{-1.2, MongoInt, true},
			test{4, MongoInt, true},
			test{int(8), MongoInt, true},
			test{float64(1.6), MongoInt, true},
			test{int8(11), MongoInt, true},
			test{int16(95), MongoInt, true},
			test{int32(46), MongoInt, true},
			test{int64(7), MongoInt, true},
			test{uint8(24), MongoInt, true},
			test{uint16(669), MongoInt, true},
			test{uint32(74), MongoInt, true},
			test{uint64(63), MongoInt, true},
			test{float32(32.2), MongoInt, true},
			test{float64(66.1), MongoInt, true},
			test{nil, MongoInt, true},
			test{"nil", MongoInt, false},
			test{timeObject, MongoInt, false},
			test{true, MongoInt, false},
			test{false, MongoInt, false},
		}
		Convey("Subject: MongoInt", func() {
			runTests(tests)
		})

		Convey("Subject: MongoInt64", func() {
			var int64Tests []test
			for _, test := range tests {
				test.mType = MongoInt64
				int64Tests = append(int64Tests, test)
			}
			runTests(int64Tests)
		})

		Convey("Subject: MongoFloat", func() {
			var float64Tests []test
			for _, test := range tests {
				test.mType = MongoFloat
				float64Tests = append(float64Tests, test)
			}
			runTests(float64Tests)
		})

		Convey("Subject: MongoDecimal", func() {
			var decimalTests []test
			for _, test := range tests {
				test.mType = MongoDecimal
				decimalTests = append(decimalTests, test)
			}
			runTests(decimalTests)
		})

		Convey("Subject: MongoGeo2D", func() {
			var geo2DTests []test
			for _, test := range tests {
				test.mType = MongoGeo2D
				geo2DTests = append(geo2DTests, test)
			}
			runTests(geo2DTests)
		})
	})

	Convey("Subject: MongoString Comparison", t, func() {
		tests := []test{
			test{66, MongoString, false},
			test{-1.2, MongoString, false},
			test{nil, MongoString, true},
			test{false, MongoString, false},
			test{true, MongoString, false},
			test{timeObject, MongoString, false},
		}
		runTests(tests)
	})

	Convey("Subject: MongoObjectId Comparison", t, func() {
		tests := []test{
			test{nil, MongoObjectId, true},
			test{"123412341234123412341234", MongoObjectId, true},
			test{"12341234123412341234123", MongoObjectId, false},
			test{"", MongoObjectId, false},
			test{66, MongoObjectId, false},
			test{false, MongoObjectId, false},
			test{true, MongoObjectId, false},
			test{timeObject, MongoObjectId, false},
			test{-1.2, MongoObjectId, false},
		}
		runTests(tests)
	})

	Convey("Subject: MongoBool Comparison", t, func() {
		tests := []test{
			test{nil, MongoBool, true},
			test{1, MongoBool, true},
			test{0, MongoBool, true},
			test{false, MongoBool, true},
			test{true, MongoBool, true},
			test{"", MongoBool, false},
			test{66, MongoBool, false},
			test{timeObject, MongoBool, false},
			test{-1.2, MongoBool, false},
		}
		runTests(tests)
	})

	Convey("Subject: MongoDate Comparison", t, func() {
		tests := []test{
			test{nil, MongoDate, true},
			test{"string", MongoDate, true},
			test{"", MongoDate, true},
			test{0, MongoDate, false},
			test{1, MongoDate, false},
			test{false, MongoDate, false},
			test{true, MongoDate, false},
			test{66, MongoDate, false},
			test{timeObject, MongoDate, false},
			test{-1.2, MongoDate, false},
		}
		runTests(tests)
	})

	Convey("Subject: MongoNone Comparison", t, func() {
		tests := []test{
			test{nil, MongoNone, false},
			test{"string", MongoNone, false},
			test{"", MongoNone, false},
			test{0, MongoNone, false},
			test{1, MongoNone, false},
			test{false, MongoNone, false},
			test{true, MongoNone, false},
			test{66, MongoNone, false},
			test{timeObject, MongoNone, false},
			test{-1.2, MongoNone, false},
		}
		runTests(tests)
	})

}
