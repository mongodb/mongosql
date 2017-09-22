package parser

import (
	"testing"
)

func testParse(t *testing.T, sql string) Statement {
	stmt, err := Parse(sql)
	if err != nil {
		t.Fatalf("sql: %s err: %s", sql, err)
	}
	return stmt
}

func testParseError(t *testing.T, sql string) Statement {
	stmt, err := Parse(sql)
	if err == nil {
		t.Fatalf("sql: %s should cause error", sql)
	}
	return stmt
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
}

func TestSetNames(t *testing.T) {
	sql := "set names gbk"
	testParse(t, sql)

	sql = "set names gbk collate lks"
	testParse(t, sql)
}

func TestSetCharset(t *testing.T) {
	sql := "set character set gbk"
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
	if testParse(t, sql).(*Show).Modifier != "" {
		t.Fatal("modifier wrong")
	}

	sql = "show full columns from `foo`"
	if testParse(t, sql).(*Show).Modifier != "full" {
		t.Fatal("modifier wrong")
	}

	sql = "show columns from `foo` from `bar`"
	testParse(t, sql)
	if String(testParse(t, sql).(*Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", String(testParse(t, sql).(*Show).From))
	}

	sql = "show columns in `foo` in `bar`"
	testParse(t, sql)
	if String(testParse(t, sql).(*Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", String(testParse(t, sql).(*Show).From))
	}

	sql = "show fields from `foo`"
	testParse(t, sql)

	sql = "show fields in `foo`"
	testParse(t, sql)
	if testParse(t, sql).(*Show).Modifier != "" {
		t.Fatal("modifier wrong")
	}

	sql = "show full fields from `foo`"
	if testParse(t, sql).(*Show).Modifier != "full" {
		t.Fatal("modifier wrong")
	}

	sql = "show fields from `foo` from `bar`"
	testParse(t, sql)
	if String(testParse(t, sql).(*Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", String(testParse(t, sql).(*Show).From))
	}

	sql = "show fields in `foo` in `bar`"
	testParse(t, sql)
	if String(testParse(t, sql).(*Show).From) != "bar.foo" {
		t.Fatalf("table wrong: %s", String(testParse(t, sql).(*Show).From))
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
}

func TestFlush(t *testing.T) {
	sql := "flush logs"
	testParse(t, sql)

	sql = "flush tables"
	testParseError(t, sql)

	sql = "flush"
	testParseError(t, sql)
}

func TestNatural(t *testing.T) {
	sql := "select * from bar natural join baz using (id)"
	testParseError(t, sql)

	sql = "select * from bar natural join baz on bar.id=baz.id"
	testParseError(t, sql)

	sql = "select * from foo natural inner join bar"
	testParseError(t, sql)

	sql = "select * from foo natural outer join bar"
	testParseError(t, sql)

	sql = "select * from foo natural cross join bar"
	testParseError(t, sql)

	sql = "select * from foo natural left join bar using (id)"
	testParseError(t, sql)

	sql = "select * from foo natural right join bar using (id)"
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
