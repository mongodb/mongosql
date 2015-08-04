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
-
  db: test2
  tables:
  -
     table: bar
     collection: test.bar
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
		t.Fatal("first db is wrong")
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

	testBar := cfg.Schemas["test2"].Tables["bar"]
	if len(testBar.Pipeline) != 2 {
		t.Fatal("test2.bar pipeline is wrong length")
	}
}
