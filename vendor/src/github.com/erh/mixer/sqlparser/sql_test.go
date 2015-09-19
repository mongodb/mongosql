package sqlparser

import (
	"testing"
)

func testParse(t *testing.T, sql string) Statement{
	stmt, err := Parse(sql)
	if err != nil {
		t.Fatalf("sql: %s err: %s", sql, err)
	}
	return stmt
}

func TestSet(t *testing.T) {
	sql := "set names gbk"
	testParse(t, sql)
}

func TestSimpleSelect(t *testing.T) {
	sql := "select last_insert_id() as a"
	testParse(t, sql)
}

func TestFunnyNames(t *testing.T) {
	sql := "select * from columns"
	testParse(t, sql)

	sql = "select * from foo.columns"
	testParse(t, sql)

	sql = "select * from tables"
	testParse(t, sql)

	sql = "select * from foo.tables"
	testParse(t, sql)
}

func TestMixer(t *testing.T) {
	sql := `admin upnode("node1", "master", "127.0.0.1")`
	testParse(t, sql)

	sql = "show databases"
	testParse(t, sql)

	sql = "show tables from abc"
	testParse(t, sql)

	sql = "show tables from abc like a"
	testParse(t, sql)

	sql = "show tables from abc where a = 1"
	testParse(t, sql)

	sql = "show proxy abc"
	testParse(t, sql)

	sql = "show databases"
	testParse(t, sql)

	sql = "show databases like 'foo'"
	testParse(t, sql)

	sql = "SHOW VARIABLES"
	testParse(t, sql)
	
	sql = "show variables"
	testParse(t, sql)

	sql = "show variables LIKE 'foo'"
	testParse(t, sql)

	sql = "show columns from 'foo'"
	testParse(t, sql)

	sql = "show columns in 'foo'"
	testParse(t, sql)
	if testParse(t, sql).(*Show).Modifier != "" {
		t.Fatal("modifier wrong")
	}

	sql = "show full columns from 'foo'"
	if testParse(t, sql).(*Show).Modifier != "full" {
		t.Fatal("modifier wrong")
	}

	sql = "show columns from 'foo' from 'bar'"
	testParse(t, sql)
	if String(testParse(t, sql).(*Show).From) != "'foo'" {
		t.Fatalf("table wrong: %s", String(testParse(t, sql).(*Show).From))
	}
	if String(testParse(t, sql).(*Show).DBFilter) != "'bar'" {
		t.Fatalf("db wrong: %s", String(testParse(t, sql).(*Show).DBFilter))
	}

	sql = "show columns in 'foo' in 'bar'"
	testParse(t, sql)
	if String(testParse(t, sql).(*Show).From) != "'foo'" {
		t.Fatalf("table wrong: %s", String(testParse(t, sql).(*Show).From))
	}
	if String(testParse(t, sql).(*Show).DBFilter) != "'bar'" {
		t.Fatalf("db wrong: %s", String(testParse(t, sql).(*Show).DBFilter))
	}
}
