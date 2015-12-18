package schema

import (
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
        name: a
        sqltype: int
     -
        name: b
        sqltype: string
-
  db: test2
  tables:
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        sqltype: string
     -
        name: b
        sqltype: int
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

	if cfg.RawDatabases[0].RawTables[0].Name != "foo" || cfg.RawDatabases[0].RawTables[0].CollectionName != "test.foo" {
		t.Fatal("Table 0 (bar) basics wrong")
	}

	if cfg.Databases["test1"].Name != "test1" {
		t.Fatal("map broken")
	}

	if cfg.Databases["test1"].Tables["foo"].CollectionName != "test.foo" {
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
        name: a
        sqltype: int
     -
        name: b
        sqltype: string
-
  db: test2
  tables:
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        sqltype: string
     -
        name: b
        sqltype: int
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

	if cfg.RawDatabases[0].RawTables[0].Name != "foo" || cfg.RawDatabases[0].RawTables[0].CollectionName != "test.foo" {
		t.Fatal("Table 0 (bar) basics wrong")
	}

	if cfg.Databases["test1"].Name != "test1" {
		t.Fatal("map broken")
	}

	if cfg.Databases["test1"].Tables["foo"].CollectionName != "test.foo" {
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
        name: a
        sqltype: int
     -
        name: b
        sqltype: string
-
  db: test3
  tables:
  - 
     table: foo
     collection: test.foo
     columns:
     -
        name: a
        sqltype: int
     -
        name: b
        sqltype: string

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
        name: a
        sqltype: string
     -
        name: b
        sqltype: int
-
  db: test3
  tables:
  - 
     table: bar
     collection: test.foo
     columns:
     -
        name: a
        sqltype: int
     -
        name: b
        sqltype: string

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
        name: a
        sqltype: int
     -
        name: b
        sqltype: string

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
        name: a
        sqltype: int
     -
        name: b
        sqltype: string

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
	_testComputeDirectory(t, "asd", "/a", "/a")
	_testComputeDirectory(t, "foo.conf", "/a", "/a")
	_testComputeDirectory(t, "/b/foo.conf", "/a", "/a")
	_testComputeDirectory(t, "/b/foo.conf", "a", "/b/a")
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
