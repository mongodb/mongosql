package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	var testConfigData = []byte(
		`
# Sample BI-Connector Config File

# Address to lsiten on
addr : 127.0.0.1:4000

# Proxy user/pass
user : root
password : abc

log_level : error

schema :
- 
  db: test1
  tables:
  - 
     table: foo
     collection: test.foo
     columns:
     -
        name: a
        type: int
     -
        name: b
        type: string
-
  db: test2
  tables:
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        type: string
     -
        name: b
        type: int
     pipeline:
     -
        $unwind : "$x"
     -
        $limit : 10

  -
     table: bar2
     collection: test.bar2
`)

	cfg, err := ParseConfigData(testConfigData)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.LogLevel != "error" || cfg.User != "root" || cfg.Password != "abc" || cfg.Addr != "127.0.0.1:4000" {
		t.Fatal("Top Config not equal.")
	}

	if len(cfg.RawSchemas) != 2 {
		t.Fatal(cfg)
	}

	if len(cfg.RawSchemas[0].RawTables) != 1 {
		t.Fatal(len(cfg.RawSchemas[0].RawTables))
	}

	if len(cfg.RawSchemas[1].RawTables) != 2 {
		t.Fatal(len(cfg.RawSchemas[1].RawTables))
	}

	if cfg.RawSchemas[0].DB != "test1" {
		t.Fatalf("first db is wrong: %s", cfg.RawSchemas[0].DB)
	}

	if cfg.RawSchemas[0].RawTables[0].Table != "foo" || cfg.RawSchemas[0].RawTables[0].Collection != "test.foo" {
		t.Fatal("Table 0 (bar) basics wrong")
	}

	if cfg.Schemas["test1"].DB != "test1" {
		t.Fatal("map broken")
	}

	if cfg.Schemas["test1"].Tables["foo"].Collection != "test.foo" {
		t.Fatal("map broken 2")
	}

	if len(cfg.Schemas["test1"].Tables["foo"].Columns) != 2 {
		t.Fatal("test1.foo num columns wrong")
	}

	if cfg.Schemas["test1"].Tables["foo"].Columns[0].Name != "a" {
		t.Fatal("test1.foo.a name wrong")
	}

	testBar := cfg.Schemas["test2"].Tables["bar"]
	if len(testBar.Pipeline) != 2 {
		t.Fatal("test2.bar pipeline is wrong length")
	}
}

func TestConfigSubdir(t *testing.T) {
	var testConfigDataRoot = []byte(
		`
addr : 127.0.0.1:4000

# Proxy user/pass
user : root
password : abc

log_level : error

schema_dir : foo
`)

	var testConfigDataSub = []byte(
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
        type: int
     -
        name: b
        type: string
-
  db: test2
  tables:
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        type: string
     -
        name: b
        type: int
     pipeline:
     -
        $unwind : "$x"
     -
        $limit : 10

  -
     table: bar2
     collection: test.bar2
`)

	cfg, err := ParseConfigData(testConfigDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "foo" {
		t.Fatal("SchemaDir wrong")
	}

	err = cfg.ingestSubFile(testConfigDataSub)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.RawSchemas) != 2 {
		t.Fatal(cfg)
	}

	if len(cfg.RawSchemas[0].RawTables) != 1 {
		t.Fatal(len(cfg.RawSchemas[0].RawTables))
	}

	if len(cfg.RawSchemas[1].RawTables) != 2 {
		t.Fatal(len(cfg.RawSchemas[1].RawTables))
	}

	if cfg.RawSchemas[0].DB != "test1" {
		t.Fatalf("first db is wrong: %s", cfg.RawSchemas[0].DB)
	}

	if cfg.RawSchemas[0].RawTables[0].Table != "foo" || cfg.RawSchemas[0].RawTables[0].Collection != "test.foo" {
		t.Fatal("Table 0 (bar) basics wrong")
	}

	if cfg.Schemas["test1"].DB != "test1" {
		t.Fatal("map broken")
	}

	if cfg.Schemas["test1"].Tables["foo"].Collection != "test.foo" {
		t.Fatal("map broken 2")
	}

	if len(cfg.Schemas["test1"].Tables["foo"].Columns) != 2 {
		t.Fatal("test1.foo num columns wrong")
	}

	if cfg.Schemas["test1"].Tables["foo"].Columns[0].Name != "a" {
		t.Fatal("test1.foo.a name wrong")
	}

	testBar := cfg.Schemas["test2"].Tables["bar"]
	if len(testBar.Pipeline) != 2 {
		t.Fatal("test2.bar pipeline is wrong length")
	}
}

func TestConfigSubdir2(t *testing.T) {
	var testConfigDataRoot = []byte(
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
        type: int
     -
        name: b
        type: string
-
  db: test3
  tables:
  - 
     table: foo
     collection: test.foo
     columns:
     -
        name: a
        type: int
     -
        name: b
        type: string

`)

	var testConfigDataSub = []byte(
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
        type: string
     -
        name: b
        type: int
-
  db: test3
  tables:
  - 
     table: bar
     collection: test.foo
     columns:
     -
        name: a
        type: int
     -
        name: b
        type: string

`)

	cfg, err := ParseConfigData(testConfigDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "foo" {
		t.Fatal("SchemaDir wrong")
	}

	err = cfg.ingestSubFile(testConfigDataSub)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Schemas["test1"] == nil {
		t.Fatal("where is test1")
	}
	if cfg.Schemas["test2"] == nil {
		t.Fatal("where is test2")
	}
	if cfg.Schemas["test3"] == nil {
		t.Fatal("where is test3")
	}

	if len(cfg.RawSchemas) != 3 {
		t.Fatal(cfg)
	}

	if len(cfg.Schemas["test1"].RawTables) != 1 {
		t.Fatal("test1 wrong")
	}

	if len(cfg.Schemas["test2"].RawTables) != 1 {
		t.Fatal("test3 wrong")
	}

	if len(cfg.Schemas["test3"].RawTables) != 2 {
		t.Fatal("test3 wrong")
	}

}

func TestConfigSubdirConflict(t *testing.T) {
	var testConfigDataRoot = []byte(
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
        type: int
     -
        name: b
        type: string

`)

	var testConfigDataSub = []byte(
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
        type: int
     -
        name: b
        type: string

`)

	cfg, err := ParseConfigData(testConfigDataRoot)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "foo" {
		t.Fatal("SchemaDir wrong")
	}

	err = cfg.ingestSubFile(testConfigDataSub)
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
	cfg, err := ParseConfigFile("test_data/foo.conf")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SchemaDir != "sub" {
		t.Fatalf("SchemaDir wrong: (%s)", cfg.SchemaDir)
	}

	if len(cfg.RawSchemas) != 3 {
		t.Fatalf("num RawSchemas wrong: %d", len(cfg.RawSchemas))
	}
}
