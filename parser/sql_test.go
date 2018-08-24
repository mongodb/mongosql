package parser_test

import (
	"testing"

	"github.com/10gen/sqlproxy/parser"
)

func testParse(t *testing.T, sql string) parser.Statement {
	stmt, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("sql: %s err: %s", sql, err)
	}
	return stmt
}

func testParseError(t *testing.T, sql string) parser.Statement {
	stmt, err := parser.Parse(sql)
	if err == nil {
		t.Fatalf("sql: %s should cause error", sql)
	}
	return stmt
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

func TestSimpleSelect(t *testing.T) {
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
