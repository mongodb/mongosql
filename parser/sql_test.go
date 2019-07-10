package parser_test

import (
	"testing"

	"github.com/10gen/sqlproxy/parser"
)

// idPass does nothing, it is the id function. We use this to test
// that Children and ReplaceChild are correctly implemented for each
// node.
type idPass struct{}

// PreVisit is called for every node before its children are walked.
func (idPass) PreVisit(current parser.CST) (parser.CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (idPass) PostVisit(current parser.CST) (parser.CST, error) {
	return current, nil
}

var _ parser.Walker = idPass{}

func testParse(t *testing.T, sql string) parser.Statement {
	stmt, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("sql: %s err: %s", sql, err)
	}
	result, err := parser.Walk(idPass{}, stmt)
	if err != nil {
		t.Fatalf("failed to Walk CST with error: %s", err)
	}
	return result.(parser.Statement)
}

func testParseError(t *testing.T, sql string) {
	_, err := parser.Parse(sql)
	if err == nil {
		t.Fatalf("sql: %s should cause error", sql)
	}
}

func testParseErrorWithError(t *testing.T, sql string, errMessage string) {
	_, err := parser.Parse(sql)
	if err == nil {
		t.Fatalf("sql: %s should cause error", sql)
	}
	if err.Error() != errMessage {
		t.Fatalf("expected error message: '%s', got '%s'", errMessage, err.Error())
	}
}

type test struct {
	name       string
	input      string
	errMessage string
}

func tableTest(t *testing.T, tests []test) {
	for _, test := range tests {
		if test.errMessage != "" {
			t.Run(test.name, func(t *testing.T) {
				testParseErrorWithError(t, test.input, test.errMessage)
			})
		} else {
			t.Run(test.name, func(t *testing.T) {
				testParse(t, test.input)
			})
		}
	}
}

func TestWith(t *testing.T) {
	sql := "WITH cte1 AS (select * from conv_numbers) select * from cte1"
	testParse(t, sql)

	sql = "with hello AS (select * from num) select * from (select * from hello) as hi"
	testParse(t, sql)

	sql = "WITH cte (col1, col2) AS (SELECT 1, 2 UNION ALL SELECT 3, 4) SELECT col1, col2 FROM cte"
	testParse(t, sql)

	// No AS
	sql = "with hello (select * from num) select * from hello"
	testParseError(t, sql)

	// No comma breaking up the column list
	sql = "WITH cte (col1 col2) AS (SELECT 1, 2 UNION ALL SELECT 3, 4) SELECT col1, col2 FROM cte"
	testParseError(t, sql)

	// Multiple with clauses at the same level are not allowed
	sql = "WITH cte1 AS (SELECT * from num) WITH cte2 AS (SELECT * FROM num) SELECT * from cte1"
	testParseError(t, sql)

	// With in a subquery is valid
	sql = "select * from (with tbl as (select * from hello) select * from tbl) as hi"
	testParse(t, sql)

	// With on multiple levels
	sql = "with hello as (select * from num) select * from (with tbl as (select * from hello) select * from tbl) as hi"
	testParse(t, sql)

	// With on multiple levels, but bad syntax (no AS keyword)
	sql = "with hello as (select * from num) select * from (with tbl (select * from hello) select * from tbl) as hi"
	testParseError(t, sql)

	// With on multiple levels, but bad syntax (forgotten parentheses in subquery)
	sql = "with hello as (select * from num) select * from (with tbl as select * from hello) as hi"
	testParseError(t, sql)

	// With on multiple levels, but bad syntax (no query using the with in subquery)
	sql = "with hello as (select * from num) select * from (with tbl as (select * from hello)) as hi"
	testParseError(t, sql)

	// MySQL does allow nested withs
	sql = "with hello as (with abc as (select * from num) select * from abc) select * from hello"
	testParse(t, sql)

	// Multiple subqueries with With clauses
	sql = "select * from (with tbl1 as (select * from numbers) select * from tbl1) AS t1, (with tbl2 as (select * from numbers) select * from tbl2) AS t2"
	testParse(t, sql)

	// Parser does support RECURSIVE CTEs and the Algebrizer will throw the exception
	sql = "WITH RECURSIVE cte (n) AS (SELECT 1 UNION ALL SELECT n + 1 FROM cte WHERE n < 5 ) SELECT * FROM cte"
	testParse(t, sql)

	sql = "WITH RECURSIVE cte (n) (SELECT 1 UNION ALL SELECT n + 1 FROM cte WHERE n < 5 ) SELECT * FROM cte"
	testParseError(t, sql)

	// Recursion in the inner query
	sql = "select * from (with RECURSIVE r(n) as (select 1 UNION select n+1 from r where n < 5) select * from r) as tbl"
	testParse(t, sql)

	// Parenthesization of second select clause in UNION will be legal in sqlproxy
	sql = "with cte1 as (select * from conv_numbers_any_base) select * from cte1 UNION (select * from cte1)"
	testParse(t, sql)

	// Expected behavior is that cte1 is available to the first select but not the second (due to the parenthesization).
	sql = "(with cte1 as (select * from conv_numbers_any_base) select * from cte1) UNION select * from cte1"
	testParse(t, sql)

	// Expected behavior is that cte1 is available to both selects
	sql = "with cte1 as (select * from conv_numbers_any_base) select * from cte1 UNION select * from cte1"
	testParse(t, sql)

	// You can't create a CTE in a database's namespace. MySQL turns this into a parse error
	sql = "with test.cte1 as (select * from cte) select * from cte1"
	testParseError(t, sql)
}

func TestDual(t *testing.T) {
	// a where clause cannot follow a dual
	sql := "select * from dual where a = 3"
	testParseError(t, sql)

	// a order by clause cannot follow a dual
	sql = "select * from dual order by a"
	testParseError(t, sql)

	// a group by clause cannot follow a dual
	sql = "select * from dual group by a"
	testParseError(t, sql)

	// a group by clause and having clause cannot follow a dual
	sql = "select * from dual group by a having a = 3"
	testParseError(t, sql)

	// a limit clause cannot follow a dual
	sql = "select * from dual limit 1"
	testParseError(t, sql)

	// You cannot give dual an alias
	sql = "select * from dual as d"
	testParseError(t, sql)

	// having dual in a multi-table from list is incorrect
	sql = "select * from foo, bar, dual"
	testParseError(t, sql)

	// you cannot parenthesize dual
	sql = "select * from (dual)"
	testParseError(t, sql)

	// dual cannot appear elsewhere, like in the select list
	sql = "select dual from foo"
	testParseError(t, sql)

	// dual cannot appear elsewhere, like in the where clause
	sql = "select a from foo where dual = 0"
	testParseError(t, sql)

	// dual can appear in the select list if escaped
	sql = "select `dual` from foo"
	testParse(t, sql)

	// dual can appear in the where clause if escaped
	sql = "select a from foo where `dual` = 0"
	testParse(t, sql)
}

func TestAliases(t *testing.T) {
	sql := "select col1 a1 from foo"
	testParse(t, sql)

	sql = "select col1 `a1` from foo"
	testParse(t, sql)

	sql = "select col1 'a1' from foo"
	testParse(t, sql)

	sql = "select col1 \"a1\" from foo"
	testParse(t, sql)

	sql = "select col1 as `a1` from foo"
	testParse(t, sql)

	sql = "select col1 as 'a1' from foo"
	testParse(t, sql)

	sql = "select col1 as \"a1\" from foo"
	testParse(t, sql)

	sql = "select date '2007-01-01' from foo"
	testParse(t, sql)

	sql = "select date '2007-01-01' 'funny' from foo"
	testParse(t, sql)

	sql = "select date funny from foo"
	testParse(t, sql)
}

func TestDropDatabase(t *testing.T) {
	sql := "drop database foo"
	testParse(t, sql)

	sql = "drop database `foo`"
	testParse(t, sql)

	sql = "drop database if exists `foo`"
	testParse(t, sql)

	sql = "drop database if exists foo"
	testParse(t, sql)

	sql = "drop database `#funny`"
	testParse(t, sql)
}

func TestCreateDatabase(t *testing.T) {
	sql := "create database foo"
	testParse(t, sql)

	sql = "create database `foo`"
	testParse(t, sql)

	sql = "create database if not exists `foo`"
	testParse(t, sql)

	sql = "create database if not exists foo"
	testParse(t, sql)

	sql = "create database `#funny`"
	testParse(t, sql)
}

func TestDropTable(t *testing.T) {
	sql := "drop table foo"
	testParse(t, sql)

	sql = "drop table `foo`"
	testParse(t, sql)

	sql = "drop table `foo`.`bar`"
	testParse(t, sql)

	sql = "drop table if exists foo.bar"
	testParse(t, sql)

	sql = "drop temporary table foo"
	testParse(t, sql)

	sql = "drop table `foo` restrict"
	testParse(t, sql)

	sql = "drop table `foo` cascade"
	testParse(t, sql)

	sql = "drop table `#funny`"
	testParse(t, sql)

	sql = "drop temporary table if exists `#funny` cascade"
	testParse(t, sql)
}

func TestCreateTable(t *testing.T) {

	tests := []test{
		{
			name: "sqldump test",
			input: "CREATE TABLE `block` (" +
				"`pid` int(7) NOT NULL," +
				"`blk` varchar(7) NOT NULL UNIQUE," +
				"`brcv` varchar(7) DEFAULT NULL," +
				"`type` varchar(4) NOT NULL," +
				"bar int, " +
				"UNIQUE KEY `pid` (`pid` ASC, `blk` DESC, `brcv`)," +
				"`foo` double precision(65,33)" +
				") ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Blocked Punts, Field Goal Attempts, etc.'",
			errMessage: "",
		}, {
			name: "test table options without commas",
			input: "CREATE TABLE `block` (" +
				"`pid` int(7) NOT NULL" +
				") ENGINE=InnoDB, DEFAULT CHARSET=utf8 COMMENT='Blocked Punts, Field Goal Attempts, etc.'",
			errMessage: "",
		}, {
			name: "test single table option",
			input: "CREATE TABLE `block` (" +
				"`pid` int(7) NOT NULL" +
				") ENGINE=InnoDB",
			errMessage: "",
		}, {
			name: "test all types",
			input: "CREATE TABLE IF not EXIsTs foo(" +
				"a BOOL," +
				"b BOOLEAN," +
				"c BIT(1)," +
				"d BIT," +
				"e TINYINT(4)," +
				"f SMALLINT(15)," +
				"g INT(42)," +
				"f INTEGER," +
				"h BIGINT(42)," +
				"h2 DATE," +
				"i DATETIME(54)," +
				"j TIMESTAMP(42)," +
				"k VARCHAR(43)," +
				"k2 CHAR(43)," +
				"l TEXT(42)," +
				"m TINYTEXT(1)," +
				"m2 MEDIUMTEXT(153)," +
				"n LONGTEXT(143)," +
				"o DECIMAL(13,34)," +
				"o2 NUMERIC(13,34)," +
				"p FLOAT(13,14)," +
				"q DECIMAL(13,34)," +
				"r DOUBLE(15,45)," +
				"s DOUBLE PRECISION(1,2) NULL DEFAULT NULL," +
				"FULLTEXT KEY(k, l)," +
				"UNIQUE KEY f(f,g DESC,h ASC)" +
				")",
			errMessage: "",
		}, {
			name:       "doubles should have widths listed as (X,Y), not (X).",
			input:      "CREATE TABLE `block` (foo double precision(42))",
			errMessage: "unexpected RPAREN at position 47",
		}, {
			name:       "dates should not have fsp.",
			input:      "CREATE TABLE `block` (foo date(42))",
			errMessage: "unexpected LPAREN at position 32",
		}, {
			name:       "SERIAL is not supported",
			input:      "CREATE TABLE foo(bar SERIAL)",
			errMessage: "SERIAL is not supported at position 29",
		}, {
			name:       "AUTO_INCREMENT is not supported",
			input:      "CREATE TABLE foo(bar int AUTO_INCREMENT)",
			errMessage: "AUTO_INCREMENT is not supported at this time at position 41",
		}, {
			name:       "DEFAULT must be NULL",
			input:      "CREATE TABLE foo(bar int DEFAULT 3)",
			errMessage: "only NULL defaults are supported at position 36",
		}, {
			name:       "TEMPORARY not allowed",
			input:      "CREATE TEMPORARY TABLE foo(bar int)",
			errMessage: "temporary tables are not supported at position 37",
		}, {
			name:       "CREATE TABLE declarations cannot be empty",
			input:      "CREATE TABLE foo()",
			errMessage: "unexpected RPAREN at position 19",
		},
	}

	tableTest(t, tests)
}

func TestSet(t *testing.T) {
	sql := "set @@temp = 'gbk'"
	testParse(t, sql)

	sql = "set @@global.temp = 'gbk'"
	testParse(t, sql)

	sql = "set @@session.temp = 'gbk'"
	testParse(t, sql)

	sql = "set @@local.temp = 'gbk'"
	testParse(t, sql)

	sql = "set @temp = 'gbk'"
	testParse(t, sql)

	sql = "set @temp = 1"
	testParse(t, sql)

	sql = "set @temp = 0"
	testParse(t, sql)

	sql = "set @temp = TRUE"
	testParse(t, sql)

	sql = "set @temp = FALSE"
	testParse(t, sql)

	sql = "set @temp = ON"
	testParse(t, sql)

	sql = "set @temp = OFF"
	testParse(t, sql)
}

func TestSetNames(t *testing.T) {
	sql := "set names gbk"
	testParse(t, sql)

	sql = "set names gbk collate lks"
	testParse(t, sql)

	sql = "set names \"gbk\""
	testParse(t, sql)

	sql = "set names \"gbk\" collate lks"
	testParse(t, sql)

	sql = "set names 'gbk'"
	testParse(t, sql)

	sql = "set names 'gbk' collate lks"
	testParse(t, sql)

	sql = "set names 'gbk' collate 'lks'"
	testParse(t, sql)

	sql = "set names 'gbk' collate \"lks\""
	testParse(t, sql)
}

func TestSetCharset(t *testing.T) {
	sql := "set character set gbk"
	testParse(t, sql)

	sql = "set character set 'gbk'"
	testParse(t, sql)

	sql = "set character set \"gbk\""
	testParse(t, sql)
}

func TestSelect(t *testing.T) {
	sql := "select last_insert_id() as a"
	testParse(t, sql)
}

func TestOrderBy(t *testing.T) {
	sql := "select * from foo.tables order by a"
	testParse(t, sql)

	sql = "select * from foo.tables order by a > b"
	testParse(t, sql)
}

func TestLimit(t *testing.T) {
	sql := "select * from foo.tables limit 12"
	testParse(t, sql)

	sql = "select * from foo.tables limit 10, 12"
	testParse(t, sql)

	sql = "select * from foo.tables limit 12 offset 10"
	testParse(t, sql)
}

func TestAliasedWhere(t *testing.T) {
	sql := "select * from foo.tables where a"
	testParse(t, sql)

	sql = "select * from foo.tables where a or b"
	testParse(t, sql)

	sql = "select * from foo.tables where a > b"
	testParse(t, sql)

	sql = "SELECT sum_a_ok AS `sum_a_ok`, sum_b_ok AS `sum_b_ok` FROM ( SELECT SUM(`foo`.`a`) AS `sum_a_ok`, SUM(`foo`.`b`) AS `sum_b_ok`, (COUNT(1) > 0) AS `havclause`, 1 AS `_Tableau_const_expr` FROM `foo` GROUP BY 4 ) `t0` WHERE havclause"
	testParse(t, sql)

	sql = "select * from foo.tables"
	testParse(t, sql)
}

func TestBooleanLiteral(t *testing.T) {
	sql := "select * from foo.tables where a=true"
	testParse(t, sql)

	sql = "select false from foo.tables where false"
	testParse(t, sql)

	sql = "select false from foo.tables order by false"
	testParse(t, sql)
}

func TestBoolArithExpr(t *testing.T) {
	sql := "select (5>6)=1"
	testParse(t, sql)

	sql = "select (5>6)+1"
	testParse(t, sql)

	sql = "select (5>6)+(5<8)"
	testParse(t, sql)

	sql = "select 3+(5<8)"
	testParse(t, sql)

	sql = "select 3+(5>8)"
	testParse(t, sql)
}

func TestTimeConstructors(t *testing.T) {

	sql := "select * from foo.tables where a > (DATE '2014-06-01 00:00:00.000')"
	testParse(t, sql)

	sql = "select * from foo.tables where a > (TIME '2014-06-01 00:00:00.000')"
	testParse(t, sql)

	sql = "select * from foo.tables where a > (TIMESTAMP '2014-06-01 00:00:00.000')"
	testParse(t, sql)
}

func TestCastExpr(t *testing.T) {

	sql := "SELECT CAST(3/3 AS signed) from foo.tables"
	testParse(t, sql)

	sql = "SELECT CAST(3/3 AS datetime) from foo.tables"
	testParse(t, sql)

	sql = "SELECT (100 * (CASE WHEN 1 = 0 THEN NULL ELSE CAST((CASE WHEN (flights201406.cancelled = '1') THEN 1 ELSE 0 END) AS SQL_DOUBLE PRECISION) / 1 END)) from foo.tables"
	testParse(t, sql)
}

func TestCaseExpr(t *testing.T) {

	// simple case expression
	sql := "SELECT case x when 1 then 2 when 2 then 3 else 4 end from foo.tables"
	testParse(t, sql)

	// searched case expression
	sql = "SELECT case when a=1 then 2 when a=2 then 3 else 4 end from foo.tables"
	testParse(t, sql)

}

func TestSubqueryComparisons(t *testing.T) {
	sql := "SELECT * FROM foo WHERE a > any (SELECT a FROM foo)"
	testParse(t, sql)

	sql = "SELECT * FROM foo WHERE a = some (SELECT a FROM foo)"
	testParse(t, sql)

	sql = "SELECT * FROM foo WHERE a <> ALL (SELECT a FROM foo)"
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

func TestShow(t *testing.T) {
	sql := "show databases"
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

	sql = "show columns from `foo`"
	testParse(t, sql)

	sql = "show columns in `foo`"
	testParse(t, sql)
	if testParse(t, sql).(*parser.Show).Modifier != "" {
		t.Fatal("modifier wrong")
	}

	sql = "show full columns from `foo`"
	if testParse(t, sql).(*parser.Show).Modifier != "full" {
		t.Fatal("modifier wrong")
	}

	sql = "show columns from `foo` from `bar`"
	testParse(t, sql)
	if parser.String(testParse(t, sql).(*parser.Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", parser.String(testParse(t, sql).(*parser.Show).From))
	}

	sql = "show columns in `foo` in `bar`"
	testParse(t, sql)
	if parser.String(testParse(t, sql).(*parser.Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", parser.String(testParse(t, sql).(*parser.Show).From))
	}

	sql = "show fields from `foo`"
	testParse(t, sql)

	sql = "show fields in `foo`"
	testParse(t, sql)
	if testParse(t, sql).(*parser.Show).Modifier != "" {
		t.Fatal("modifier wrong")
	}

	sql = "show full fields from `foo`"
	if testParse(t, sql).(*parser.Show).Modifier != "full" {
		t.Fatal("modifier wrong")
	}

	sql = "show fields from `foo` from `bar`"
	testParse(t, sql)
	if parser.String(testParse(t, sql).(*parser.Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", parser.String(testParse(t, sql).(*parser.Show).From))
	}

	sql = "show fields in `foo` in `bar`"
	testParse(t, sql)
	if parser.String(testParse(t, sql).(*parser.Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", parser.String(testParse(t, sql).(*parser.Show).From))
	}
}

func TestExplain(t *testing.T) {
	sql := "explain foo"
	testParse(t, sql)

	sql = "explain foo a"
	testParse(t, sql)

	sql = "explain foo 'a%'"
	testParse(t, sql)

	sql = "describe foo 'a%'"
	testParse(t, sql)

	sql = "desc foo 'a%'"
	testParse(t, sql)

	sql = "explain foo 'a%' b"
	testParseError(t, sql)

	sql = "explain extended select * from foo"
	testParse(t, sql)

	sql = "explain partitions select * from foo"
	testParse(t, sql)

	sql = "explain format=json select * from foo"
	testParse(t, sql)

	sql = "explain format=traditional select * from foo"
	testParse(t, sql)

	sql = "explain partitions (select * from foo)"
	testParse(t, sql)

	sql = "describe partitions select * from foo"
	testParse(t, sql)

	sql = "desc partitions select * from foo"
	testParse(t, sql)

	sql = "explain bogus select * from foo"
	testParseError(t, sql)

	sql = "explain extended for connection bogus"
	testParseError(t, sql)

	sql = "explain extended for connection 1"
	testParse(t, sql)

	sql = "describe extended select * from foo"
	testParse(t, sql)

	sql = "desc extended select * from foo"
	testParse(t, sql)
}

func TestKeywordsAsIds(t *testing.T) {
	sql := "select a columns, b any from foo tables"
	testParse(t, sql)

	sql = "select a as columns, b as any from foo as tables"
	testParse(t, sql)

	sql = "select columns from (select sum(sum) from tables) tables"
	testParse(t, sql)

	sql = "select count, count(sum) from columns"
	testParse(t, sql)
}

func TestUsing(t *testing.T) {
	sql := "select bar.d, baz.a from bar join baz using (234)"
	testParseError(t, sql)

	sql = "select bar.d, baz.a from bar join baz using (77.9)"
	testParseError(t, sql)

	sql = "select bar.d, baz.a from bar join baz using ('c')"
	testParseError(t, sql)

	sql = "select bar.d, baz.a from bar join baz using (false)"
	testParseError(t, sql)

	sql = "select bar.d, baz.a from bar join baz using ()"
	testParseError(t, sql)

	// using not valid in straight_join
	sql = "select bar.d, baz.a from bar straight_join baz using (id)"
	testParseError(t, sql)
}

func TestFlush(t *testing.T) {
	sql := "flush logs"
	testParse(t, sql)

	sql = "flush sample"
	testParse(t, sql)

	sql = "flush tables"
	testParseError(t, sql)

	sql = "flush"
	testParseError(t, sql)
}

func TestNatural(t *testing.T) {
	sql := "select * from foo natural inner join bar"
	testParseError(t, sql)

	sql = "select * from foo natural outer join bar"
	testParseError(t, sql)

	sql = "select * from foo natural cross join bar"
	testParseError(t, sql)
}

func TestUnion(t *testing.T) {
	sql := "select d from bar union select g from baz"
	testParse(t, sql)

	sql = "select a from foo union select d from bar union select g from baz order by d"
	testParse(t, sql)

	sql = "select a from foo union select d from foo inner join bar on foo.id=bar.id union select g from baz order by a"
	testParse(t, sql)

	sql = "select * from (select a from foo union select d from foo inner join bar on foo.id=bar.id union select g from baz) as tmp order by a"
	testParse(t, sql)

	sql = "(select db from servers union select user from user)" // This one is supported in MySql 8.0.0
	testParse(t, sql)

	sql = "((select db from servers) union (select user from user))" // This one is supported in MySql 8.0.0
	testParse(t, sql)

	sql = "(select db from servers) union (select user from user)"
	testParse(t, sql)

	sql = "select * from (select a from foo union select d from foo) as tmp"
	testParse(t, sql)
}

func TestParenthesis(t *testing.T) {
	sql := "(select * from bar)"
	testParse(t, sql)

	sql = "(select * from bar) limit 2"
	testParse(t, sql)

	sql = "((select * from bar)) limit 2"
	testParse(t, sql)

	sql = "((select * from bar)) order by id limit 2"
	testParse(t, sql)

	sql = "(select * from bar limit 2)"
	testParse(t, sql)

	sql = "(select * from (select * from bar) limit 2)"
	testParse(t, sql)

	sql = "(select * from (((select * from bar))) limit 2) limit 3"
	testParse(t, sql)

	sql = "(((select * from (((select * from db))) s limit 3))) limit 1"
	testParse(t, sql)

	sql = "SELECT C_Name, cr.P_FirstName+\" \"+cr.P_SurName AS ClassRepresentativ, cr2.P_FirstName+\" \"+cr2.P_SurName AS ClassRepresentativ2nd FROM ((class INNER JOIN person AS cr ON class.C_P_ClassRep=cr.P_Nr) INNER JOIN person AS cr2 ON class.C_P_ClassRep2nd=cr2.P_Nr)"
	testParse(t, sql)

	sql = "(SELECT `Calcs`.`_id` AS `_id`, `Calcs`.`bool0_` AS `bool0_`, `Calcs`.`bool1_` AS `bool1_`, `Calcs`.`bool2_` AS `bool2_`, `Calcs`.`bool3_` AS `bool3_`, `Calcs`.`date0` AS `date0`, `Calcs`.`date1` AS `date1`, `Calcs`.`date2` AS `date2`, `Calcs`.`date3` AS `date3`, `Calcs`.`datetime0` AS `datetime0`, `Calcs`.`datetime1` AS `datetime1`, `Calcs`.`int0` AS `int0`, `Calcs`.`int1` AS `int1`, `Calcs`.`int2` AS `int2`, `Calcs`.`int3` AS `int3`, `Calcs`.`key` AS `key`, `Calcs`.`num0` AS `num0`, `Calcs`.`num1` AS `num1`, `Calcs`.`num2` AS `num2`, `Calcs`.`num3` AS `num3`, `Calcs`.`num4` AS `num4`, `Calcs`.`str0` AS `str0`, `Calcs`.`str1` AS `str1`, `Calcs`.`str2` AS `str2`, `Calcs`.`str3` AS `str3`, `Calcs`.`time0` AS `time0`, `Calcs`.`time1` AS `time1`, `Calcs`.`zzz` AS `zzz` FROM `Calcs`) LIMIT 0"
	testParse(t, sql)

	sql = "((select * from bar) order by id limit 2)"
	testParseError(t, sql)

	sql = "select * from (((select * from tuples)) s limit 2) limit 1"
	testParseError(t, sql)

	sql = "((select * from (((select * from bar))) d limit 2) limit 3)"
	testParseError(t, sql)
}

func TestAlter(t *testing.T) {
	correct := []string{
		"alter table mytable change oldcol newcol",
		"alter table mytable change column oldcol newcol",
		"alter table mytable drop mycolumn",
		"alter table mytable drop column mycolumn",
		"alter table mytable modify column mycolumn integer",
		"alter table mytable modify mycolumn integer",
		"alter table mytable rename newtbl",
		"alter table mytable rename to newtbl",
		"alter table mytable rename as newtbl",
	}

	for _, sql := range correct {
		testParse(t, sql)
	}

	incorrect := []string{
		"alter table mytable add onecolumn, add twocolumn",
		"alter table mytable add onecolumn,add twocolumn",
		"alter table mytable add mycolumn",
		"alter table mytable add column mycolumn",
		"alter table mytable",
		"alter table mytable add mycolumn,",
		"alter table mytable , add mycolumn",
		"alter table mytable add mycolumn, add",
		"alter mytable add mycolumn",
		"alter table mytable add column",
		"alter table mytable add onecolumn twocolumn",
		"alter table mytable change column",
		"alter table mytable drop",
		"alter table mytable drop column",
		"alter table mytable drop onecolumn twocolumn",
		"alter table mytable modify column integer",
		"alter table mytable modify column",
		"alter table mytable modify mycolumn",
		"alter table mytable rename to",
		"alter table mytable rename as",
	}

	for _, sql := range incorrect {
		testParseError(t, sql)
	}
}

func TestRename(t *testing.T) {
	correct := []string{
		"rename table oldtable to newtable",
		"rename table oldtable to newtable, thistable to thattable",
	}

	for _, sql := range correct {
		testParse(t, sql)
	}

	incorrect := []string{
		"rename table oldtable newtable",
		"rename table oldtable as newtable",
		"rename table oldtable to newtable,",
	}

	for _, sql := range incorrect {
		testParseError(t, sql)
	}
}

func TestSelectStraightJoin(t *testing.T) {
	sql := "select straight_join * from foo join bar"
	testParse(t, sql)

	sql = "select straight_join distinct * from foo join bar"
	testParse(t, sql)

	sql = "select distinct straight_join * from foo join bar"
	testParse(t, sql)

	sql = "with cte1 as (select * from foo) select straight_join * from cte1 join bar"
	testParse(t, sql)

	sql = "select straight_join 2+2"
	testParse(t, sql)
}

func TestComments(t *testing.T) {
	sql := "select /*comment*/ * from foo;"
	testParse(t, sql)

	sql = "select * from foo where a > 5 ;"
	testParse(t, sql)

	sql = "select * from foo; -- comment"
	testParse(t, sql)

	sql = "select * from foo; #comment"
	testParse(t, sql)

	sql = "select `a;` from `fo;o`; #comment"
	testParse(t, sql)

	sql = "select * /*this\n is \na \ncomment */ from foo;"
	testParse(t, sql)

	// We do not support comments of this form.
	sql = "select * from foo; // comment"
	testParseError(t, sql)

	// Double-dash must be followed by whitespace.
	sql = "select * from foo; --comment"
	testParseError(t, sql)

	// The following should error out because they comment out the semi-colon.
	sql = "select * from foo -- comment;"
	testParseError(t, sql)

	sql = "select * from foo #comment;"
	testParseError(t, sql)
}

func TestGroupConcat(t *testing.T) {
	sql := "select group_concat(a) from foo"
	testParse(t, sql)

	sql = "select group_concat(a, b) from foo"
	testParse(t, sql)

	sql = "select group_concat(distinct a, b) from foo"
	testParse(t, sql)

	sql = "select group_concat(distinct a, b order by a asc, b desc) from foo"
	testParse(t, sql)

	sql = "select group_concat(distinct a, b order by a asc, b desc separator \":\") from foo"
	testParse(t, sql)
}

func TestScalarFunctions(t *testing.T) {
	sql := "select trim('a') from foo"
	testParse(t, sql)

	sql = "select trim(both 'a' from 'b') from foo"
	testParse(t, sql)

	sql = "select trim('a' from 'b') from foo"
	testParse(t, sql)

	sql = "select trim('a', 'b', 'c') from foo"
	testParseError(t, sql)

	sql = "select trim('a', 'b', 'c', 'd') from foo"
	testParseError(t, sql)

	sql = "select current_timestamp() from foo"
	testParse(t, sql)

	sql = "select current_timestamp(4) from foo"
	testParse(t, sql)

	sql = "select current_timestamp(1, 2) from foo"
	testParseError(t, sql)
}
