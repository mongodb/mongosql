package evaluator

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"strconv"

	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func createFieldNameLookup(db *schema.Database) fieldNameLookup {

	return func(tableName, columnName string) (string, bool) {
		table := db.Tables[tableName]
		if table == nil {
			return "", false
		}

		column := table.SQLColumns[columnName]
		if column == nil {
			return "", false
		}

		return column.Name, true
	}
}

type fakeConnectionCtx struct{}

func (_ fakeConnectionCtx) LastInsertId() int64 {
	return 11
}
func (_ fakeConnectionCtx) RowCount() int64 {
	return 21
}
func (_ fakeConnectionCtx) ConnectionId() uint32 {
	return 42
}
func (_ fakeConnectionCtx) DB() string {
	return "test"
}
func (_ fakeConnectionCtx) Session() *mgo.Session {
	panic("Session is not supported in fakeConnectionCtx")
}
func (_ fakeConnectionCtx) User() string {
	return "test user"
}

func TestEvaluates(t *testing.T) {

	type test struct {
		sql    string
		result SQLExpr
	}

	runTests := func(ctx *EvalCtx, tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be %v", t.sql, t.result), func() {
				subject, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				result, err := subject.Evaluate(ctx)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	execCtx := &ExecutionCtx{
		ConnectionCtx: fakeConnectionCtx{},
	}

	Convey("Subject: Evaluates", t, func() {
		evalCtx := NewEvalCtx(execCtx, &Row{Values{
			{1, "bar", "a", 123},
			{1, "bar", "b", 456},
			{1, "bar", "c", nil},
		}})

		Convey("Subject: SQLAddExpr", func() {
			tests := []test{
				test{"0 + 0", SQLInt(0)},
				test{"-1 + 1", SQLInt(0)},
				test{"10 + 32", SQLInt(42)},
				test{"-10 + -32", SQLInt(-42)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLAggFunctionExpr", func() {
			var t1, t2 time.Time

			t1 = time.Now()
			t2 = t1.Add(time.Hour)

			aggCtx := NewEvalCtx(execCtx,
				&Row{Values{
					{1, "bar", "a", nil},
					{1, "bar", "b", 3},
					{1, "bar", "c", nil},
					{1, "bar", "g", t1},
				}},
				&Row{Values{
					{1, "bar", "a", 3},
					{1, "bar", "b", nil},
					{1, "bar", "c", nil},
					{1, "bar", "g", t2},
				}},
				&Row{Values{
					{1, "bar", "a", 5},
					{1, "bar", "b", 6},
					{1, "bar", "c", nil},
					{1, "bar", "g", nil},
				}},
			)

			Convey("Subject: AVG", func() {
				tests := []test{
					test{"AVG(NULL)", SQLNull},
					test{"AVG(a)", SQLFloat(4)},
					test{"AVG(b)", SQLFloat(4.5)},
					test{"AVG(c)", SQLNull},
					test{"AVG(g)", SQLFloat(0)},
					test{"AVG('a')", SQLFloat(0)},
					test{"AVG(-20)", SQLFloat(-20)},
					test{"AVG(20)", SQLFloat(20)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: SUM", func() {
				tests := []test{
					test{"SUM(NULL)", SQLNull},
					test{"SUM(a)", SQLFloat(8)},
					test{"SUM(b)", SQLFloat(9)},
					test{"SUM(c)", SQLNull},
					test{"SUM(g)", SQLFloat(0)},
					test{"SUM('a')", SQLFloat(0)},
					test{"SUM(-20)", SQLFloat(-60)},
					test{"SUM(20)", SQLFloat(60)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: MIN", func() {
				tests := []test{
					test{"MIN(NULL)", SQLNull},
					test{"MIN(a)", SQLFloat(3)},
					test{"MIN(b)", SQLInt(3)},
					test{"MIN(c)", SQLNull},
					test{"MIN('a')", SQLVarchar("a")},
					test{"MIN(-20)", SQLInt(-20)},
					test{"MIN(20)", SQLInt(20)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: MAX", func() {
				tests := []test{
					test{"MAX(NULL)", SQLNull},
					test{"MAX(a)", SQLFloat(5)},
					test{"MAX(b)", SQLInt(6)},
					test{"MAX(c)", SQLNull},
					test{"MAX('a')", SQLVarchar("a")},
					test{"MAX(-20)", SQLInt(-20)},
					test{"MAX(20)", SQLInt(20)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: COUNT", func() {
				tests := []test{
					test{"COUNT(NULL)", SQLInt(0)},
					test{"COUNT(a)", SQLInt(2)},
					test{"COUNT(b)", SQLInt(2)},
					test{"COUNT(c)", SQLInt(0)},
					test{"COUNT(g)", SQLInt(2)},
					test{"COUNT('a')", SQLInt(3)},
					test{"COUNT(-20)", SQLInt(3)},
					test{"COUNT(20)", SQLInt(3)},
				}
				runTests(aggCtx, tests)
			})

		})

		Convey("Subject: SQLAndExpr", func() {
			// INT-1040: boolean literals don't work
			tests := []test{
				test{"1 AND 1", SQLTrue},
				test{"1 AND 0", SQLFalse},
				test{"0 AND 1", SQLFalse},
				test{"0 AND 0", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLDivideExpr", func() {
			tests := []test{
				test{"-1 / 1", SQLFloat(-1)},
				test{"100 / 10", SQLFloat(10)},
				test{"-10 / 10", SQLFloat(-1)},
			}

			runTests(evalCtx, tests)

			Convey("The result should be SQLNull when dividing by zero", func() {
				subject := &SQLDivideExpr{SQLInt(10), SQLInt(0)}
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, SQLNull)
			})
		})

		Convey("Subject: SQLEqualsExpr", func() {
			tests := []test{
				test{"0 = 0", SQLTrue},
				test{"-1 = 1", SQLFalse},
				test{"10 = 10", SQLTrue},
				test{"-10 = -10", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		SkipConvey("Subject: SQLExistsExpr", func() {
		})

		Convey("Subject: SQLColumnExpr", func() {
			Convey("Should return the value of the field when it exists", func() {
				subject := NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt)
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, SQLInt(123))
			})

			Convey("Should return nil when the field is null", func() {
				subject := NewSQLColumnExpr(1, "bar", "c", schema.SQLInt, schema.MongoInt)
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, SQLNull)
			})

			Convey("Should return nil when the field doesn't exists", func() {
				subject := NewSQLColumnExpr(1, "bar", "no_existy", schema.SQLInt, schema.MongoInt)
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, SQLNull)
			})
		})

		Convey("Subject: SQLGreaterThanExpr", func() {
			tests := []test{
				test{"0 > 0", SQLFalse},
				test{"-1 > 1", SQLFalse},
				test{"1 > -1", SQLTrue},
				test{"11 > 10", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLGreaterThanOrEqualExpr", func() {
			tests := []test{
				test{"0 >= 0", SQLTrue},
				test{"-1 >= 1", SQLFalse},
				test{"1 >= -1", SQLTrue},
				test{"11 >= 10", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLIDivideExpr", func() {
			tests := []test{
				test{"0 DIV 0", SQLNull},
				test{"0 DIV 5", SQLInt(0)},
				test{"5.5 DIV 2", SQLInt(2)},
				test{"-5 DIV 2", SQLInt(-2)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLInExpr", func() {
			tests := []test{
				test{"0 IN(0)", SQLTrue},
				test{"-1 IN(1)", SQLFalse},
				test{"0 IN(10, 0)", SQLTrue},
				test{"-1 IN(1, 10)", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLLessThanExpr", func() {
			tests := []test{
				test{"0 < 0", SQLFalse},
				test{"-1 < 1", SQLTrue},
				test{"1 < -1", SQLFalse},
				test{"10 < 11", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLLessThanOrEqualExpr", func() {
			tests := []test{
				test{"0 <= 0", SQLTrue},
				test{"-1 <= 1", SQLTrue},
				test{"1 <= -1", SQLFalse},
				test{"10 <= 11", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLLikeExpr", func() {
			tests := []test{
				test{"'Á Â Ã Ä' LIKE '%'", SQLInt(1)},
				test{"'Á Â Ã Ä' LIKE 'Á Â Ã Ä'", SQLInt(1)},
				test{"'Á Â Ã Ä' LIKE 'Á%'", SQLInt(1)},
				test{"'a' LIKE 'a'", SQLInt(1)},
				test{"'Adam' LIKE 'am'", SQLInt(0)},
				test{"'Adam' LIKE 'adaM'", SQLInt(1)}, // mixed case test
				test{"'Adam' LIKE '%am%'", SQLInt(1)},
				test{"'Adam' LIKE 'Ada_'", SQLInt(1)},
				test{"'Adam' LIKE '__am'", SQLInt(1)},
				test{"'Clever' LIKE '%is'", SQLInt(0)},
				test{"'Adam is nice' LIKE '%xs '", SQLInt(0)},
				test{"'Adam is nice' LIKE '%is nice'", SQLInt(1)},
				test{"'abc' LIKE 'ABC'", SQLInt(1)},   //case sensitive test
				test{"'abc' LIKE 'ABC '", SQLInt(0)},  // trailing space test
				test{"'abc' LIKE ' ABC'", SQLInt(0)},  // leading space test
				test{"'abc' LIKE ' ABC '", SQLInt(0)}, // padded space test
				test{"'abc' LIKE 'ABC	'", SQLInt(0)}, // trailing tab test
				test{"'10' LIKE '1%'", SQLInt(1)},
				test{"'a   ' LIKE 'A   '", SQLInt(1)},
				test{"CURRENT_DATE() LIKE '2015-05-31%'", SQLInt(0)},
				test{"(DATE '2008-01-02') LIKE '2008-01%'", SQLInt(1)},
				test{"NOW() LIKE '" + strconv.Itoa(time.Now().Year()) + "%' ", SQLInt(1)},
				test{"10 LIKE '1%'", SQLInt(1)},
				test{"1.20 LIKE '1.2%'", SQLInt(1)},
				test{"NULL LIKE '1%'", SQLNull},
				test{"10 LIKE NULL", SQLNull},
				test{"NULL LIKE NULL", SQLNull},
			}
			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLModExpr", func() {
			tests := []test{
				test{"0 % 0", SQLNull},
				test{"5 % 2", SQLFloat(1)},
				test{"5.5 % 2", SQLFloat(1.5)},
				test{"-5 % -3", SQLFloat(-2)},
				test{"5 MOD 2", SQLFloat(1)},
				test{"5.5 MOD 2", SQLFloat(1.5)},
				test{"-5 MOD -3", SQLFloat(-2)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLMultiplyExpr", func() {
			tests := []test{
				test{"0 * 0", SQLInt(0)},
				test{"-1 * 1", SQLInt(-1)},
				test{"10 * 32", SQLInt(320)},
				test{"-10 * -32", SQLInt(320)},
				test{"2.5 * 3", SQLFloat(7.5)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNotEqualsExpr", func() {
			tests := []test{
				test{"0 <> 0", SQLFalse},
				test{"-1 <> 1", SQLTrue},
				test{"10 <> 10", SQLFalse},
				test{"-10 <> -10", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNotExpr", func() {
			// INT-1040: boolean literals don't work
			tests := []test{
				test{"NOT 1", SQLFalse},
				test{"NOT 0", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNullCmpExpr", func() {
			tests := []test{
				test{"1 IS NULL", SQLFalse},
				test{"NULL IS NULL", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLOrExpr", func() {
			// INT-1040: boolean literals don't work
			tests := []test{
				test{"1 OR 1", SQLTrue},
				test{"1 OR 0", SQLTrue},
				test{"0 OR 1", SQLTrue},
				test{"0 OR 0", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLScalarFunctionExpr", func() {

			Convey("Subject: ABS", func() {
				tests := []test{
					test{"ABS(NULL)", SQLNull},
					test{"ABS('C')", SQLFloat(0)},
					test{"ABS(-20)", SQLFloat(20)},
					test{"ABS(20)", SQLFloat(20)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ASCII", func() {
				tests := []test{
					test{"ASCII(NULL)", SQLNull},
					test{"ASCII('')", SQLInt(0)},
					test{"ASCII('A')", SQLInt(65)},
					test{"ASCII('AWESOME')", SQLInt(65)},
					test{"ASCII('¢')", SQLInt(194)},
					test{"ASCII('Č')", SQLInt(196)}, // This is actually 268, but the first byte is 196
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: COALESCE", func() {
				tests := []test{
					test{"COALESCE(NULL)", SQLNull},
					test{"COALESCE('A')", SQLVarchar("A")},
					test{"COALESCE('A', NULL)", SQLVarchar("A")},
					test{"COALESCE('A', 'B')", SQLVarchar("A")},
					test{"COALESCE(NULL, 'A', NULL, 'B')", SQLVarchar("A")},
					test{"COALESCE(NULL, NULL, NULL)", SQLNull},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CONCAT", func() {
				tests := []test{
					test{"CONCAT(NULL)", SQLNull},
					test{"CONCAT('A')", SQLVarchar("A")},
					test{"CONCAT('A', 'B')", SQLVarchar("AB")},
					test{"CONCAT('A', NULL, 'B')", SQLNull},
					test{"CONCAT('A', 123, 'B')", SQLVarchar("A123B")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CONCAT_WS", func() {
				tests := []test{
					test{"CONCAT_WS(NULL, NULL)", SQLNull},
					test{"CONCAT_WS(',','A')", SQLVarchar("A")},
					test{"CONCAT_WS(',','A', 'B')", SQLVarchar("A,B")},
					test{"CONCAT_WS(',','A', NULL, 'B')", SQLVarchar("A,B")},
					test{"CONCAT_WS(',','A', 123, 'B')", SQLVarchar("A,123,B")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CONNECTION_ID", func() {
				tests := []test{
					test{"CONNECTION_ID()", SQLUint32(42)},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: CURRENT_DATE", func() {
				tests := []test{
					test{"CURRENT_DATE()", SQLDate{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: CURRENT_TIMESTAMP", func() {
				tests := []test{
					test{"CURRENT_TIMESTAMP()", SQLTimestamp{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CURRENT_USER/SESSION_USER/SYSTEM_USER/USER", func() {
				tests := []test{
					test{"CURRENT_USER()", SQLVarchar("test user")},
					test{"SESSION_USER()", SQLVarchar("test user")},
					test{"SYSTEM_USER()", SQLVarchar("test user")},
					test{"USER()", SQLVarchar("test user")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DATABASE/SCHEMA", func() {
				tests := []test{
					test{"DATABASE()", SQLVarchar("test")},
					test{"SCHEMA()", SQLVarchar("test")},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: NOW", func() {
				tests := []test{
					test{"NOW()", SQLTimestamp{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DAYNAME", func() {
				tests := []test{
					test{"DAYNAME(NULL)", SQLNull},
					test{"DAYNAME(14)", SQLNull},
					test{"DAYNAME('2016-01-01 00:00:00')", SQLVarchar("Friday")},
					test{"DAYNAME('2016-1-1')", SQLVarchar("Friday")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DAYOFMONTH", func() {
				tests := []test{
					test{"DAYOFMONTH(NULL)", SQLNull},
					test{"DAYOFMONTH(14)", SQLNull},
					test{"DAYOFMONTH('2016-01-01')", SQLInt(1)},
					test{"DAYOFMONTH('2016-1-1')", SQLInt(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DAYOFWEEK", func() {
				tests := []test{
					test{"DAYOFWEEK(NULL)", SQLNull},
					test{"DAYOFWEEK(14)", SQLNull},
					test{"DAYOFWEEK('2016-01-01')", SQLInt(6)},
					test{"DAYOFWEEK('2016-1-1')", SQLInt(6)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DAYOFYEAR", func() {
				tests := []test{
					test{"DAYOFYEAR(NULL)", SQLNull},
					test{"DAYOFYEAR(14)", SQLNull},
					test{"DAYOFYEAR('2016-1-1')", SQLInt(1)},
					test{"DAYOFYEAR('2016-01-01')", SQLInt(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: EXP", func() {
				tests := []test{
					test{"EXP(NULL)", SQLNull},
					test{"EXP('sdg')", SQLNull},
					test{"EXP(0)", SQLFloat(1)},
					test{"EXP(2)", SQLFloat(7.38905609893065)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: FLOOR", func() {
				tests := []test{
					test{"FLOOR(NULL)", SQLNull},
					test{"FLOOR('sdg')", SQLNull},
					test{"FLOOR(1.23)", SQLFloat(1)},
					test{"FLOOR(-1.23)", SQLFloat(-2)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: HOUR", func() {
				tests := []test{
					test{"HOUR(NULL)", SQLNull},
					test{"HOUR('sdg')", SQLNull},
					test{"HOUR('10:23:52')", SQLInt(10)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ISNULL", func() {
				tests := []test{
					test{"ISNULL(a)", SQLInt(0)},
					test{"ISNULL(c)", SQLInt(1)},
					test{`ISNULL("")`, SQLInt(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: INSTR", func() {
				tests := []test{
					test{"INSTR(NULL, NULL)", SQLNull},
					test{"INSTR('sDg', 'D')", SQLInt(2)},
					test{"INSTR(124, 124)", SQLInt(1)},
					test{"INSTR('awesome','so')", SQLInt(4)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LCASE", func() {
				tests := []test{
					test{"LCASE(NULL)", SQLNull},
					test{"LCASE('sDg')", SQLVarchar("sdg")},
					test{"LCASE(124)", SQLVarchar("124")},
					test{"LOWER(NULL)", SQLNull},
					test{"LOWER('')", SQLVarchar("")},
					test{"LOWER('A')", SQLVarchar("a")},
					test{"LOWER('awesome')", SQLVarchar("awesome")},
					test{"LOWER('AwEsOmE')", SQLVarchar("awesome")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LEFT", func() {
				tests := []test{
					test{"LEFT(NULL, NULL)", SQLNull},
					test{"LEFT('sDgcdcdc', 4)", SQLVarchar("sDgc")},
					test{"LEFT(124, 2)", SQLVarchar("12")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LENGTH", func() {
				tests := []test{
					test{"LENGTH(NULL)", SQLNull},
					test{"LENGTH('sDg')", SQLInt(3)},
					test{"LENGTH('世界')", SQLInt(6)},
					test{"CHAR_LENGTH(NULL)", SQLNull},
					test{"CHAR_LENGTH('')", SQLInt(0)},
					test{"CHAR_LENGTH('A')", SQLInt(1)},
					test{"CHAR_LENGTH('AWESOME')", SQLInt(7)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LN", func() {
				tests := []test{
					test{"LN(NULL)", SQLNull},
					test{"LN(1)", SQLFloat(0)},
					test{"LN(16.5)", SQLFloat(2.803360380906535)},
					test{"LN(-16.5)", SQLFloat(0)},
					test{"LOG(NULL)", SQLNull},
					test{"LOG(1)", SQLFloat(0)},
					test{"LOG(16.5)", SQLFloat(2.803360380906535)},
					test{"LOG(-16.5)", SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOCATE", func() {
				tests := []test{
					test{"LOCATE(NULL, 'foobarbar')", SQLNull},
					test{"LOCATE('bar', NULL)", SQLNull},
					test{"LOCATE('bar', 'foobarbar')", SQLInt(4)},
					test{"LOCATE('xbar', 'foobarbar')", SQLInt(0)},
					test{"LOCATE('bar', 'foobarbar', 5)", SQLInt(7)},
					test{"LOCATE('語', '日本語')", SQLInt(3)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOG2", func() {
				tests := []test{
					test{"LOG2(NULL)", SQLNull},
					test{"LOG2(4)", SQLFloat(2)},
					test{"LOG2(-100)", SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOG10", func() {
				tests := []test{
					test{"LOG10(NULL)", SQLNull},
					test{"LOG10('sdg')", SQLNull},
					test{"LOG10(2)", SQLFloat(0.3010299956639812)},
					test{"LOG10(100)", SQLFloat(2)},
					test{"LOG10(0)", SQLFloat(0)},
					test{"LOG10(-100)", SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LTRIM", func() {
				tests := []test{
					test{"LTRIM(NULL)", SQLNull},
					test{"LTRIM('   barbar')", SQLVarchar("barbar")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MINUTE", func() {
				tests := []test{
					test{"MINUTE(NULL)", SQLNull},
					test{"MINUTE('sdg')", SQLNull},
					test{"MINUTE('10:23:52')", SQLInt(23)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MOD", func() {
				tests := []test{
					test{"MOD(NULL, NULL)", SQLNull},
					test{"MOD(234, NULL)", SQLNull},
					test{"MOD(NULL, 10)", SQLNull},
					test{"MOD(234, 0)", SQLNull},
					test{"MOD(234, 10)", SQLFloat(4)},
					test{"MOD(253, 7)", SQLFloat(1)},
					test{"MOD(34.5, 3)", SQLFloat(1.5)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MONTH", func() {
				tests := []test{
					test{"MONTH(NULL)", SQLNull},
					test{"MONTH('sdg')", SQLNull},
					test{"MONTH('2016-1-01 10:23:52')", SQLInt(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MONTHNAME", func() {
				tests := []test{
					test{"MONTHNAME(NULL)", SQLNull},
					test{"MONTHNAME('sdg')", SQLNull},
					test{"MONTHNAME('2016-1-01 10:23:52')", SQLVarchar("January")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: QUARTER", func() {
				tests := []test{
					test{"QUARTER(NULL)", SQLNull},
					test{"QUARTER('sdg')", SQLNull},
					test{"QUARTER('2016-1-01 10:23:52')", SQLInt(1)},
					test{"QUARTER('2016-4-01 10:23:52')", SQLInt(2)},
					test{"QUARTER('2016-8-01 10:23:52')", SQLInt(3)},
					test{"QUARTER('2016-11-01 10:23:52')", SQLInt(4)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: RIGHT", func() {
				tests := []test{
					test{"RIGHT(NULL, NULL)", SQLNull},
					test{"RIGHT('sDgcdcdc', 4)", SQLVarchar("dcdc")},
					test{"RIGHT(124, 2)", SQLVarchar("24")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ROUND", func() {
				tests := []test{
					test{"ROUND(NULL, NULL)", SQLNull},
					test{"ROUND(NULL, 4)", SQLNull},
					test{"ROUND(-16.55555, 4)", SQLFloat(-16.5556)},
					test{"ROUND(4.56, 1)", SQLFloat(4.6)},
					test{"ROUND(-16.5, -1)", SQLFloat(0)},
					test{"ROUND(-16.5)", SQLFloat(-17)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: RTRIM", func() {
				tests := []test{
					test{"RTRIM(NULL)", SQLNull},
					test{"RTRIM('barbar   ')", SQLVarchar("barbar")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SECOND", func() {
				tests := []test{
					test{"SECOND(NULL)", SQLNull},
					test{"SECOND('sdg')", SQLNull},
					test{"SECOND('10:23:52')", SQLInt(52)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SQRT", func() {
				tests := []test{
					test{"SQRT(NULL)", SQLNull},
					test{"SQRT('sdg')", SQLNull},
					test{"SQRT(-16)", SQLNull},
					test{"SQRT(4)", SQLFloat(2)},
					test{"SQRT(20)", SQLFloat(4.47213595499958)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SUBSTRING", func() {
				tests := []test{
					test{"SUBSTRING(NULL, 4)", SQLNull},
					test{"SUBSTRING('foobarbar', NULL)", SQLNull},
					test{"SUBSTRING('foobarbar', 4, NULL)", SQLNull},
					test{"SUBSTRING('Quadratically', 5)", SQLVarchar("ratically")},
					test{"SUBSTRING('Quadratically', 5, 6)", SQLVarchar("ratica")},
					test{"SUBSTRING('Sakila', -3)", SQLVarchar("ila")},
					test{"SUBSTRING('Sakila', -5, 3)", SQLVarchar("aki")},
					test{"SUBSTRING('日本語', 2)", SQLVarchar("本語")},
					test{"SUBSTR(NULL, 4)", SQLNull},
					test{"SUBSTR('foobarbar', NULL)", SQLNull},
					test{"SUBSTR('foobarbar', 4, NULL)", SQLNull},
					test{"SUBSTR('Quadratically', 5)", SQLVarchar("ratically")},
					test{"SUBSTR('Quadratically', 5, 6)", SQLVarchar("ratica")},
					test{"SUBSTR('Sakila', -3)", SQLVarchar("ila")},
					test{"SUBSTR('Sakila', -5, 3)", SQLVarchar("aki")},
					test{"SUBSTR('日本語', 2)", SQLVarchar("本語")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: UCASE", func() {
				tests := []test{
					test{"UCASE(NULL)", SQLNull},
					test{"UCASE('sdg')", SQLVarchar("SDG")},
					test{"UCASE(124)", SQLVarchar("124")},
					test{"UPPER(NULL)", SQLNull},
					test{"UPPER('')", SQLVarchar("")},
					test{"UPPER('a')", SQLVarchar("A")},
					test{"UPPER('AWESOME')", SQLVarchar("AWESOME")},
					test{"UPPER('AwEsOmE')", SQLVarchar("AWESOME")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: VERSION", func() {
				tests := []test{
					test{"VERSION()", SQLVarchar(common.VersionStr)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: WEEK", func() {
				tests := []test{
					test{"WEEK(NULL)", SQLNull},
					test{"WEEK('sdg')", SQLNull},
					test{"WEEK('2016-1-01 10:23:52')", SQLInt(53)},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: YEAR", func() {
				tests := []test{
					test{"YEAR(NULL)", SQLNull},
					test{"YEAR('sdg')", SQLNull},
					test{"YEAR('2016-1-01 10:23:52')", SQLInt(53)},
				}
				runTests(evalCtx, tests)
			})

		})

		Convey("Subject: SQLSubtractExpr", func() {
			tests := []test{
				test{"0 - 0", SQLInt(0)},
				test{"-1 - 1", SQLInt(-2)},
				test{"10 - 32", SQLInt(-22)},
				test{"-10 - -32", SQLInt(22)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLTupleExpr", func() {
			Convey("Should evaluate all the expressions and return SQLValues", func() {
				subject := &SQLTupleExpr{[]SQLExpr{SQLInt(10), &SQLAddExpr{SQLInt(30), SQLInt(12)}}}
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, &SQLValues{})
				resultValues := result.(*SQLValues)
				So(resultValues.Values[0], ShouldEqual, SQLInt(10))
				So(resultValues.Values[1], ShouldEqual, SQLInt(42))
			})
			Convey("Should evaluate to a single SQLValue if it contains only one value", func() {
				subject := &SQLTupleExpr{[]SQLExpr{SQLInt(10)}}
				sqlInt, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				intResult := sqlInt.(SQLInt)
				So(intResult, ShouldEqual, SQLInt(10))

				subject = &SQLTupleExpr{[]SQLExpr{SQLVarchar("10")}}
				sqlVarchar, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				varcharResult := sqlVarchar.(SQLVarchar)
				So(varcharResult, ShouldEqual, SQLVarchar("10"))
			})
		})

		Convey("Subject: SQLUnaryMinusExpr", func() {
			tests := []test{
				test{"-10", SQLInt(-10)},
				test{"-a", SQLFloat(-123)},
				test{"-b", SQLInt(-456)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLVariableExpr", func() {
			tests := []test{
				test{"@@max_allowed_packet", SQLInt(4194304)},
			}

			runTests(evalCtx, tests)

			Convey("Should error when unknown variable is used", func() {
				subject := &SQLVariableExpr{
					Name:         "blah",
					VariableType: SystemVariable,
				}
				_, err := subject.Evaluate(evalCtx)
				So(err, ShouldNotBeNil)
			})
		})

		SkipConvey("Subject: SQLUnaryPlusExpr", func() {
			// TODO: what this is supposed to do?
		})

		SkipConvey("Subject: SQLUnaryTildeExpr", func() {
			// TODO: I'm not convinced we have this correct.
		})
	})
}

func TestMatches(t *testing.T) {
	Convey("Subject: Matches", t, func() {

		evalCtx := NewEvalCtx(nil)

		tests := [][]interface{}{
			[]interface{}{SQLInt(124), true},
			[]interface{}{SQLFloat(1235), true},
			[]interface{}{SQLVarchar("512"), true},
			[]interface{}{SQLInt(0), false},
			[]interface{}{SQLFloat(0), false},
			[]interface{}{SQLVarchar("000"), false},
			[]interface{}{SQLVarchar("skdjbkjb"), false},
			[]interface{}{SQLVarchar(""), false},
			[]interface{}{SQLTrue, true},
			[]interface{}{SQLFalse, false},
			[]interface{}{&SQLEqualsExpr{SQLInt(42), SQLInt(42)}, true},
			[]interface{}{&SQLEqualsExpr{SQLInt(42), SQLInt(21)}, false},
		}

		for _, t := range tests {
			Convey(fmt.Sprintf("Should evaluate %v(%T) to %v", t[0], t[0], t[1]), func() {
				m, err := Matches(t[0].(SQLExpr), evalCtx)
				So(err, ShouldBeNil)
				So(m, ShouldEqual, t[1])
			})
		}
	})
}

func TestNewSQLValue(t *testing.T) {

	Convey("When creating a SQLValue with no column type specified calling NewSQLValue on a", t, func() {

		Convey("SQLValue should return the same object passed in", func() {
			v := SQLTrue
			newV, err := NewSQLValue(v, schema.SQLBoolean, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, v)
		})

		Convey("nil value should return SQLNull", func() {
			v, err := NewSQLValue(nil, schema.SQLNull, schema.MongoBool)
			So(err, ShouldBeNil)
			So(v, ShouldResemble, SQLNull)
		})

		Convey("bson object id should return its string value", func() {
			v := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err := NewSQLValue(v, schema.SQLVarchar, schema.MongoObjectId)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v.Hex())
		})

		Convey("string objects should return the string value", func() {
			v := "56a10dd56ce28a89a8ed6edb"
			newV, err := NewSQLValue(v, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("int objects should return the int value", func() {
			v1 := int(6)
			newV, err := NewSQLValue(v1, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v1)

			v2 := int32(6)
			newV, err = NewSQLValue(v2, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v2)

			v3 := uint32(6)
			newV, err = NewSQLValue(v3, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v3)
		})

		Convey("float objects should return the float value", func() {
			v := float64(6.3)
			newV, err := NewSQLValue(v, schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("time objects should return the appropriate value", func() {
			v := time.Date(2014, time.December, 31, 0, 0, 0, 0, schema.DefaultLocale)
			newV, err := NewSQLValue(v, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v})

			v = time.Date(2014, time.December, 31, 10, 0, 0, 0, schema.DefaultLocale)
			newV, err = NewSQLValue(v, schema.SQLTimestamp, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlTimestamp, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTimestamp, ShouldResemble, SQLTimestamp{v})
		})
	})

	Convey("When creating a SQLValue with a column type specified calling NewSQLValue on a", t, func() {

		Convey("a SQLVarchar/SQLVarchar column type should attempt to coerce to the SQLVarchar type", func() {

			t := schema.SQLVarchar

			newV, err := NewSQLValue(t, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar(t))

			newV, err = NewSQLValue(6, schema.SQLVarchar, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6"))

			newV, err = NewSQLValue(6.6, schema.SQLVarchar, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6.6"))

			newV, err = NewSQLValue(int64(6), schema.SQLVarchar, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6"))

			_id := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err = NewSQLValue(_id, schema.SQLVarchar, schema.MongoObjectId)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLObjectID(_id.Hex()))

		})

		Convey("a SQLInt column type should attempt to coerce to the SQLInt type", func() {

			_, err := NewSQLValue(true, schema.SQLInt, schema.MongoBool)
			So(err, ShouldNotBeNil)

			_, err = NewSQLValue("6", schema.SQLInt, schema.MongoString)
			So(err, ShouldNotBeNil)

			newV, err := NewSQLValue(int(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(int32(6), schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(int64(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(float64(6.6), schema.SQLInt, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

		})

		Convey("a SQLFloat column type should attempt to coerce to the SQLFloat type", func() {

			_, err := NewSQLValue(true, schema.SQLFloat, schema.MongoBool)
			So(err, ShouldNotBeNil)

			_, err = NewSQLValue("6.6", schema.SQLFloat, schema.MongoString)
			So(err, ShouldNotBeNil)

			newV, err := NewSQLValue(int(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(int32(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(int64(6), schema.SQLFloat, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(float64(6.6), schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6.6))

		})

		Convey("a SQLDate column type should attempt to coerce to the SQLDate type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 0, 0, 0, 0, schema.DefaultLocale)
			v2 := time.Date(2014, time.May, 11, 10, 32, 12, 0, schema.DefaultLocale)

			newV, err := NewSQLValue(v1, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			newV, err = NewSQLValue(v2, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok = newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			// String type
			dates := []string{"2014-05-11", "2014-05-11 15:04:05", "2014-05-11 15:04:05.233"}

			for _, d := range dates {

				newV, err := NewSQLValue(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldBeNil)

				sqlDate, ok := newV.(SQLDate)
				So(ok, ShouldBeTrue)
				So(sqlDate, ShouldResemble, SQLDate{v1})

			}

			// invalid dates and those outside valid range
			// should return the default date
			dates = []string{"2014-12-44-44", "999-1-1", "10000-1-1"}

			for _, d := range dates {
				_, err = NewSQLValue(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("a SQLTimestamp column type should attempt to coerce to the SQLTimestamp type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)

			newV, err := NewSQLValue(v1, schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// String type
			newV, err = NewSQLValue("2014-05-11 15:04:05.000", schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok = newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// invalid dates should return the default date
			dates := []string{"2044-12-40", "1966-15-1", "43223-3223"}

			for _, d := range dates {
				_, err = NewSQLValue(d, schema.SQLTimestamp, schema.MongoNone)
				So(err, ShouldNotBeNil)
			}
		})
	})
}

func TestOptimizeSQLExpr(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   SQLExpr
	}

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be optimized to %q", t.sql, t.expected), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				result, err := OptimizeSQLExpr(e)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	Convey("Subject: OptimizeSQLExpr", t, func() {

		tests := []test{
			test{"3 = a", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 < a", "a > 3", &SQLGreaterThanExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 <= a", "a >= 3", &SQLGreaterThanOrEqualExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 > a", "a < 3", &SQLLessThanExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 >= a", "a <= 3", &SQLLessThanOrEqualExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 <> a", "a <> 3", &SQLNotEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 + 3 = 6", "true", SQLTrue},
			test{"3 / (3 - 2) = a", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLFloat(3)}},
			test{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 / (3 - 2) = a AND 4 - 2 = b", "a = 3 AND b = 2",
				&SQLAndExpr{
					&SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLFloat(3)},
					&SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "b", schema.SQLInt, schema.MongoInt), SQLInt(2)}}},
			test{"3 + 3 = 6 OR a = 3", "true", SQLTrue},
			test{"3 + 3 = 5 OR a = 3", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"3 + 3 = 5 AND a = 3", "false", SQLFalse},
			test{"3 + 3 = 6 AND a = 3", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"a = (~1 + 1 + (+4))", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt), SQLInt(3)}},
			test{"DAYNAME('2016-1-1')", "Friday", SQLVarchar("Friday")},
			test{"(8-7)", "1", SQLInt(1)},
		}

		runTests(tests)
	})
}

func TestReconcileSQLExpr(t *testing.T) {

	type test struct {
		sql             string
		reconciledLeft  SQLExpr
		reconciledRight SQLExpr
	}

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be reconciled to %#v and %#v", t.sql, t.reconciledLeft, t.reconciledRight), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				left, right := getBinaryExprLeaves(e)
				left, right, err = reconcileSQLExprs(left, right)
				So(err, ShouldBeNil)
				So(left, ShouldResemble, t.reconciledLeft)
				So(right, ShouldResemble, t.reconciledRight)
			})
		}
	}

	exprConv := &SQLConvertExpr{SQLVarchar("2010-01-01"), schema.SQLTimestamp}
	exprTime := &SQLScalarFunctionExpr{"current_timestamp", []SQLExpr{}}
	exprA := NewSQLColumnExpr(1, "bar", "a", schema.SQLNumeric, schema.MongoInt)
	exprB := NewSQLColumnExpr(1, "bar", "b", schema.SQLInt, schema.MongoInt)
	exprG := NewSQLColumnExpr(1, "bar", "g", schema.SQLTimestamp, schema.MongoDate)

	Convey("Subject: reconcileSQLExpr", t, func() {

		tests := []test{
			test{"a = 3", exprA, SQLInt(3)},
			test{"g - '2010-01-01'", exprG, exprConv},
			test{"a in (3)", exprA, SQLInt(3)},
			test{"a in (2,3)", exprA, &SQLTupleExpr{[]SQLExpr{SQLInt(2), SQLInt(3)}}},
			test{"(a) in (3)", exprA, SQLInt(3)},
			test{"(a,b) in (2,3)", &SQLTupleExpr{[]SQLExpr{exprA, exprB}}, &SQLTupleExpr{[]SQLExpr{SQLInt(2), SQLInt(3)}}},
			test{"g > '2010-01-01'", exprG, exprConv},
			test{"a and b", exprA, exprB},
			test{"a / b", exprA, exprB},
			test{"'2010-01-01' and g", exprConv, exprG},
			test{"g in ('2010-01-01',current_timestamp())", exprG, &SQLTupleExpr{[]SQLExpr{exprConv, exprTime}}},
		}

		runTests(tests)
	})

}

func TestTranslatePredicate(t *testing.T) {

	type test struct {
		sql      string
		expected string
	}

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		lookupFieldName := createFieldNameLookup(schema.Databases[dbOne])

		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be translated to \"%s\"", t.sql, t.expected), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				e, err = OptimizeSQLExpr(e)
				So(err, ShouldBeNil)
				match, local := TranslatePredicate(e, lookupFieldName)
				jsonResult, err := json.Marshal(match)
				So(err, ShouldBeNil)
				So(string(jsonResult), ShouldEqual, t.expected)
				So(local, ShouldBeNil)
			})
		}
	}

	type partialTest struct {
		sql       string
		expected  string
		localDesc string
		local     SQLExpr
	}

	runPartialTests := func(tests []partialTest) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		lookupFieldName := createFieldNameLookup(schema.Databases[dbOne])

		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be translated to \"%s\" and locally evaluate %q", t.sql, t.expected, t.localDesc), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				match, local := TranslatePredicate(e, lookupFieldName)
				jsonResult, err := json.Marshal(match)
				So(err, ShouldBeNil)
				So(string(jsonResult), ShouldEqual, t.expected)
				So(local, ShouldResemble, t.local)
			})
		}
	}

	Convey("Subject: TranslatePredicate", t, func() {
		tests := []test{
			test{"a = 3", `{"a":3}`},
			test{"a > 3", `{"a":{"$gt":3}}`},
			test{"a >= 3", `{"a":{"$gte":3}}`},
			test{"a < 3", `{"a":{"$lt":3}}`},
			test{"a <= 3", `{"a":{"$lte":3}}`},
			test{"a <> 3", `{"a":{"$ne":3}}`},
			test{"a > 3 AND a < 10", `{"$and":[{"a":{"$gt":3}},{"a":{"$lt":10}}]}`},
			test{"(a > 3 AND a < 10) AND b = 10", `{"$and":[{"a":{"$gt":3}},{"a":{"$lt":10}},{"b":10}]}`},
			test{"a > 3 AND (a < 10 AND b = 10)", `{"$and":[{"a":{"$gt":3}},{"a":{"$lt":10}},{"b":10}]}`},
			test{"a > 3 OR a < 10", `{"$or":[{"a":{"$gt":3}},{"a":{"$lt":10}}]}`},
			test{"(a > 3 OR a < 10) OR b = 10", `{"$or":[{"a":{"$gt":3}},{"a":{"$lt":10}},{"b":10}]}`},
			test{"a > 3 OR (a < 10 OR b = 10)", `{"$or":[{"a":{"$gt":3}},{"a":{"$lt":10}},{"b":10}]}`},
			test{"(a > 3 AND a < 10) OR b = 10", `{"$or":[{"$and":[{"a":{"$gt":3}},{"a":{"$lt":10}}]},{"b":10}]}`},
			test{"a > 3 AND (a < 10 OR b = 10)", `{"$and":[{"a":{"$gt":3}},{"$or":[{"a":{"$lt":10}},{"b":10}]}]}`},
			test{"a IN(1,3,5)", `{"a":{"$in":[1,3,5]}}`},
			test{"g IN('2016-02-03 12:23:11.392')", `{"g":{"$in":["2016-02-03T12:23:11.392Z"]}}`},
			test{"h IN('2016-02-03 12:23:11.392')", `{"h":{"$in":["2016-02-03T00:00:00Z"]}}`},
			test{"a IS NULL", `{"a":null}`},
			test{"NOT (a > 3)", `{"a":{"$not":{"$gt":3}}}`},
			test{"NOT (NOT (a > 3))", `{"a":{"$gt":3}}`},
			test{"NOT (a = 3)", `{"a":{"$ne":3}}`},
			test{"NOT (a <> 3)", `{"a":3}`},
			test{"NOT (a NOT IN (1,3,5))", `{"a":{"$in":[1,3,5]}}`},
			test{"a NOT IN (1,3,5)", `{"a":{"$nin":[1,3,5]}}`},
			test{"NOT a IN (1,3,5)", `{"a":{"$nin":[1,3,5]}}`},
			test{"NOT (a > 3 AND a < 10)", `{"$nor":[{"$and":[{"a":{"$gt":3}},{"a":{"$lt":10}}]}]}`},
			test{"NOT (NOT (a > 3 AND a < 10))", `{"$or":[{"$and":[{"a":{"$gt":3}},{"a":{"$lt":10}}]}]}`},
			test{"NOT (a > 3 OR a < 10)", `{"$nor":[{"a":{"$gt":3}},{"a":{"$lt":10}}]}`},
		}

		runTests(tests)

		partialTests := []partialTest{
			partialTest{"a = 3 AND a < b", `{"a":3}`, "a < b", &SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLNumeric, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}},
			partialTest{"a = 3 AND a < b AND b = 4", `{"$and":[{"a":3},{"b":4}]}`, "a < b", &SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLNumeric, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}},
			partialTest{"a < b AND a = 3", `{"a":3}`, "a < b", &SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLNumeric, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}},
			partialTest{"NOT (a = 3 AND a < b)", `{"a":{"$ne":3}}`, "NOT a < b", &SQLNotExpr{&SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLNumeric, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}}},
		}

		runPartialTests(partialTests)
	})
}

func TestTranslateExpr(t *testing.T) {

	type test struct {
		sql      string
		expected string
	}

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		lookupFieldName := createFieldNameLookup(schema.Databases[dbOne])
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be translated to \"%s\"", t.sql, t.expected), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				translated, ok := TranslateExpr(e, lookupFieldName)
				So(ok, ShouldBeTrue)
				jsonResult, err := json.Marshal(translated)
				So(err, ShouldBeNil)
				So(string(jsonResult), ShouldResembleDiffed, t.expected)
			})
		}
	}

	Convey("Subject: TranslateExpr", t, func() {

		tests := []test{
			test{"abs(a)", `{"$abs":"$a"}`},
			test{"concat(a, 'funny')", `{"$concat":["$a",{"$literal":"funny"}]}`},
			test{"concat(a, null)", `{"$concat":["$a"]}`},
			test{"concat(a, '')", `{"$concat":["$a",{"$literal":""}]}`},
			test{"concat_ws(',', a)", `{"$concat":["$a"]}`},
			test{"concat_ws(',', a, null)", `{"$concat":["$a"]}`},
			test{"concat_ws(',', a, 'funny')", `{"$concat":["$a",{"$literal":","},{"$literal":"funny"}]}`},
			test{"concat_ws(',', a, b)", `{"$concat":["$a",{"$literal":","},"$b"]}`},
			test{"concat_ws(',', a, b, 'hey')", `{"$concat":["$a",{"$literal":","},"$b",{"$literal":","},{"$literal":"hey"}]}`},
			test{"dayname(a)", `{"$arrayElemAt":[["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"],{"$subtract":[{"$dayOfWeek":"$a"},1]}]}`},
			test{"dayofmonth(a)", `{"$dayOfMonth":"$a"}`},
			test{"dayofweek(a)", `{"$dayOfWeek":"$a"}`},
			test{"dayofyear(a)", `{"$dayOfYear":"$a"}`},
			test{"exp(a)", `{"$exp":"$a"}`},
			test{"floor(a)", `{"$floor":"$a"}`},
			test{"hour(a)", `{"$hour":"$a"}`},
			test{"isnull(a)", `{"$cond":[{"$eq":["$a",null]},1,0]}`},
			test{"left(a, 2)", `{"$substr":["$a",0,{"$literal":2}]}`},
			test{"left('abcde', 0)", `{"$substr":[{"$literal":"abcde"},0,{"$literal":0}]}`},
			test{"lcase(a)", `{"$toLower":"$a"}`},
			test{"lower(a)", `{"$toLower":"$a"}`},
			test{"log10(a)", `{"$log10":"$a"}`},
			test{"minute(a)", `{"$minute":"$a"}`},
			test{"mod(a, 10)", `{"$mod":["$a",{"$literal":10}]}`},
			test{"month(a)", `{"$month":"$a"}`},
			test{"monthname(a)", `{"$arrayElemAt":[["January","February","March","April","May","June","July","August","September","October","November","December"],{"$subtract":[{"$month":"$a"},1]}]}`},
			test{"power(a, 10)", `{"$pow":["$a",{"$literal":10}]}`},
			test{"quarter(a)", `{"$arrayElemAt":[[1,1,1,2,2,2,3,3,3,4,4,4],{"$subtract":[{"$month":"$a"},1]}]}`},
			test{"round(a, 5)", `{"$divide":[{"$cond":[{"$gte":["$a",0]},{"$floor":{"$add":[{"$multiply":["$a",100000]},0.5]}},{"$floor":{"$subtract":[{"$multiply":["$a",100000]},0.5]}}]},100000]}`},
			test{"round(a, -5)", `{"$literal":0}`},
			test{"second(a)", `{"$second":"$a"}`},
			test{"sqrt(a)", `{"$sqrt":"$a"}`},
			test{"substring(a, 2)", `{"$substr":["$a",{"$literal":2},-1]}`},
			test{"substring(a, 2, 4)", `{"$substr":["$a",{"$literal":2},{"$literal":4}]}`},
			test{"substr(a, 2)", `{"$substr":["$a",{"$literal":2},-1]}`},
			test{"substr(a, 2, 4)", `{"$substr":["$a",{"$literal":2},{"$literal":4}]}`},
			test{"week(a)", `{"$week":"$a"}`},
			test{"ucase(a)", `{"$toUpper":"$a"}`},
			test{"upper(a)", `{"$toUpper":"$a"}`},
			//test{"week(a, 3)", `{"$week":"$a"}`}, Not support second argument
			//test{"year(a)", `{"$year":"$a"}`}, Parser error

			test{"sum(a * b)", `{"$sum":{"$multiply":["$a","$b"]}}`},
			test{"sum(a)", `{"$sum":"$a"}`},
			test{"sum(a < 1)", `{"$sum":{"$lt":["$a",{"$literal":1}]}}`},
			test{"min(a + 4)", `{"$min":{"$add":["$a",{"$literal":4}]}}`},
			test{"count(*)", `{"$size":{"$literal":"*"}}`},
			test{"count(a + b)", `{"$sum":{"$map":{"as":"i","in":{"$cond":[{"$eq":[{"$ifNull":["$$i",null]},null]},0,1]},"input":{"$add":["$a","$b"]}}}}`},
		}

		runTests(tests)

	})

	type sqlValueTest struct {
		sqlValue SQLValue
		expected string
	}

	runSQLValueTests := func(tests []sqlValueTest) {
		schema, err := schema.New(testSchema3)
		lookupFieldName := createFieldNameLookup(schema.Databases["test"])
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be translated to \"%s\"", t.sqlValue, t.expected), func() {
				match, ok := TranslateExpr(t.sqlValue, lookupFieldName)
				So(ok, ShouldBeTrue)
				jsonResult, err := json.Marshal(match)
				So(err, ShouldBeNil)
				So(string(jsonResult), ShouldEqual, t.expected)
			})
		}
	}

	fakeTime := time.Now()

	Convey("Subject: TranslateExpr with SQLValue", t, func() {

		sqlValueTests := []sqlValueTest{
			sqlValueTest{SQLTrue, `{"$literal":true}`},
			sqlValueTest{SQLFalse, `{"$literal":false}`},
			sqlValueTest{SQLFloat(1.1), `{"$literal":1.1}`},
			sqlValueTest{SQLInt(11), `{"$literal":11}`},
			sqlValueTest{SQLUint32(32), `{"$literal":32}`},
			sqlValueTest{SQLVarchar("vc"), `{"$literal":"vc"}`},
			sqlValueTest{SQLNull, "null"},
			sqlValueTest{SQLDate{fakeTime}, fmt.Sprintf(`{"$literal":"%v"}`, fakeTime.Format(schema.DateFormat))},
			sqlValueTest{SQLTimestamp{fakeTime}, fmt.Sprintf(`{"$literal":"%v"}`, fakeTime.Format(schema.TimestampFormat))},
		}

		runSQLValueTests(sqlValueTests)
	})
}

func TestCompareTo(t *testing.T) {

	var (
		diff = time.Duration(969 * time.Hour)
		now  = time.Now()
		oid1 = bson.NewObjectId().Hex()
		oid2 = bson.NewObjectId().Hex()
		oid3 = bson.NewObjectId().Hex()
	)

	Convey("Subject: CompareTo", t, func() {

		type test struct {
			left     SQLValue
			right    SQLValue
			expected int
		}

		runTests := func(tests []test) {
			for _, t := range tests {
				Convey(fmt.Sprintf("comparing '%v' (%T) to '%v' (%T) should return %v", t.left, t.left, t.right, t.right, t.expected), func() {
					compareTo, err := CompareTo(t.left, t.right)
					So(err, ShouldBeNil)
					So(compareTo, ShouldEqual, t.expected)
				})
			}
		}

		Convey("Subject: SQLInt", func() {
			tests := []test{
				{SQLInt(1), SQLInt(0), 1},
				{SQLInt(1), SQLInt(1), 0},
				{SQLInt(1), SQLInt(2), -1},
				{SQLInt(1), SQLUint32(1), 0},
				{SQLInt(1), SQLFloat(1), 0},
				{SQLInt(1), SQLBool(false), -1},
				{SQLInt(1), SQLBool(true), -1},
				{SQLInt(1), SQLNull, 1},
				{SQLInt(1), SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{SQLInt(1), SQLVarchar("bac"), -1},
				{SQLInt(1), &SQLValues{[]SQLValue{SQLInt(1)}}, 0},
				{SQLInt(1), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLInt(1), SQLDate{now}, -1},
				{SQLInt(1), SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLUint32", func() {
			tests := []test{
				{SQLUint32(1), SQLInt(0), 1},
				{SQLUint32(1), SQLInt(1), 0},
				{SQLUint32(1), SQLInt(2), -1},
				{SQLUint32(1), SQLUint32(1), 0},
				{SQLUint32(1), SQLFloat(1), 0},
				{SQLUint32(1), SQLBool(false), -1},
				{SQLUint32(1), SQLBool(true), -1},
				{SQLUint32(1), SQLNull, 1},
				{SQLUint32(1), SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{SQLUint32(1), SQLVarchar("bac"), -1},
				{SQLUint32(1), &SQLValues{[]SQLValue{SQLInt(1)}}, 0},
				{SQLUint32(1), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLUint32(1), SQLDate{now}, -1},
				{SQLUint32(1), SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLFloat", func() {
			tests := []test{
				{SQLFloat(0.1), SQLInt(0), 1},
				{SQLFloat(1.1), SQLInt(1), 1},
				{SQLFloat(0.1), SQLInt(2), -1},
				{SQLFloat(1.1), SQLUint32(1), 1},
				{SQLFloat(1.1), SQLFloat(1), 1},
				{SQLFloat(0.1), SQLBool(false), -1},
				{SQLFloat(0.1), SQLBool(true), -1},
				{SQLFloat(0.1), SQLNull, 1},
				{SQLFloat(0.1), SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{SQLFloat(0.1), SQLVarchar("bac"), -1},
				{SQLFloat(0.0), &SQLValues{[]SQLValue{SQLInt(1)}}, -1},
				{SQLFloat(0.1), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLFloat(0.1), SQLDate{now}, -1},
				{SQLFloat(0.1), SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLBool", func() {
			tests := []test{
				{SQLBool(true), SQLInt(0), 1},
				{SQLBool(true), SQLInt(1), 1},
				{SQLBool(true), SQLInt(2), 1},
				{SQLBool(true), SQLUint32(1), 1},
				{SQLBool(true), SQLFloat(1), 1},
				{SQLBool(true), SQLBool(false), 1},
				{SQLBool(true), SQLBool(true), 0},
				{SQLBool(true), SQLNull, 1},
				{SQLBool(true), SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLBool(true), SQLVarchar("bac"), 1},
				{SQLBool(true), &SQLValues{[]SQLValue{SQLInt(1)}}, 1},
				{SQLBool(true), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLBool(true), SQLDate{now}, -1},
				{SQLBool(true), SQLTimestamp{now}, -1},
				{SQLBool(false), SQLInt(0), 1},
				{SQLBool(false), SQLInt(1), 1},
				{SQLBool(false), SQLInt(2), 1},
				{SQLBool(false), SQLUint32(1), 1},
				{SQLBool(false), SQLFloat(1), 1},
				{SQLBool(false), SQLBool(false), 0},
				{SQLBool(false), SQLBool(true), -1},
				{SQLBool(false), SQLNull, 1},
				{SQLBool(false), SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLBool(false), SQLVarchar("bac"), 1},
				{SQLBool(false), &SQLValues{[]SQLValue{SQLInt(1)}}, 1},
				{SQLBool(false), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLBool(false), SQLDate{now}, -1},
				{SQLBool(false), SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				{SQLDate{now}, SQLInt(0), 1},
				{SQLDate{now}, SQLInt(1), 1},
				{SQLDate{now}, SQLInt(2), 1},
				{SQLDate{now}, SQLUint32(1), 1},
				{SQLDate{now}, SQLFloat(1), 1},
				{SQLDate{now}, SQLBool(false), 1},
				{SQLDate{now}, SQLDate{now.Add(diff)}, -1},
				{SQLDate{now}, SQLNull, 1},
				{SQLDate{now}, SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLDate{now}, SQLVarchar("bac"), 1},
				{SQLDate{now}, &SQLValues{[]SQLValue{SQLInt(1)}}, 1},
				{SQLDate{now}, &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLDate{now}, SQLDate{now.Add(-diff)}, 1},
				{SQLDate{now}, SQLTimestamp{now.Add(diff)}, -1},
				{SQLDate{now}, SQLTimestamp{now.Add(-diff)}, 1},
				{SQLDate{now}, SQLDate{now}, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				{SQLTimestamp{now}, SQLInt(0), 1},
				{SQLTimestamp{now}, SQLInt(1), 1},
				{SQLTimestamp{now}, SQLInt(2), 1},
				{SQLTimestamp{now}, SQLUint32(1), 1},
				{SQLTimestamp{now}, SQLFloat(1), 1},
				{SQLTimestamp{now}, SQLBool(false), 1},
				{SQLTimestamp{now}, SQLNull, 1},
				{SQLTimestamp{now}, SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLTimestamp{now}, SQLVarchar("bac"), 1},
				{SQLTimestamp{now}, &SQLValues{[]SQLValue{SQLInt(1)}}, 1},
				{SQLTimestamp{now}, &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLTimestamp{now}, SQLTimestamp{now.Add(diff)}, -1},
				{SQLTimestamp{now}, SQLTimestamp{now.Add(-diff)}, 1},
				{SQLTimestamp{now}, SQLTimestamp{now}, 0},
				{SQLTimestamp{now}, SQLDate{now}, 1},
				{SQLTimestamp{now}, SQLDate{now.Add(diff)}, -1},
				{SQLTimestamp{now}, SQLDate{now.Add(-diff)}, 1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLNullValue", func() {
			tests := []test{
				{SQLNull, SQLInt(0), -1},
				{SQLNull, SQLInt(1), -1},
				{SQLNull, SQLInt(2), -1},
				{SQLNull, SQLUint32(1), -1},
				{SQLNull, SQLFloat(1), -1},
				{SQLNull, SQLBool(false), -1},
				{SQLNull, SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{SQLNull, SQLVarchar("bac"), -1},
				{SQLNull, &SQLValues{[]SQLValue{SQLInt(1)}}, -1},
				{SQLNull, &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLNull, &SQLValues{[]SQLValue{SQLNull}}, 0},
				{SQLNull, SQLDate{now}, -1},
				{SQLNull, SQLTimestamp{now}, -1},
				{SQLNull, SQLNull, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				{SQLVarchar("bac"), SQLInt(0), 1},
				{SQLVarchar("bac"), SQLInt(1), 1},
				{SQLVarchar("bac"), SQLInt(2), 1},
				{SQLVarchar("bac"), SQLUint32(1), 1},
				{SQLVarchar("bac"), SQLFloat(1), 1},
				{SQLVarchar("bac"), SQLBool(false), -1},
				{SQLVarchar("bac"), SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{SQLVarchar("bac"), SQLVarchar("cba"), -1},
				{SQLVarchar("bac"), SQLVarchar("bac"), 0},
				{SQLVarchar("bac"), SQLVarchar("abc"), 1},
				{SQLVarchar("bac"), &SQLValues{[]SQLValue{SQLInt(1)}}, 1},
				{SQLVarchar("bac"), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLVarchar("bac"), &SQLValues{[]SQLValue{SQLVarchar("bac")}}, 0},
				{SQLVarchar("bac"), SQLDate{now}, -1},
				{SQLVarchar("bac"), SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLValues", func() {
			tests := []test{
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLInt(0), 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLInt(1), 0},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLInt(2), -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLUint32(1), 0},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLUint32(11), -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLUint32(0), 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLFloat(1.1), -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLFloat(0.1), 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLBool(false), -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLVarchar("abc"), -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLNone, 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, &SQLValues{[]SQLValue{SQLInt(1)}}, 0},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, &SQLValues{[]SQLValue{SQLInt(-1)}}, 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, &SQLValues{[]SQLValue{SQLInt(2)}}, -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, &SQLValues{[]SQLValue{SQLNone}}, 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLDate{now}, -1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLObjectID", func() {

			tests := []test{
				{SQLObjectID(oid2), SQLInt(0), 1},
				{SQLObjectID(oid2), SQLUint32(1), 1},
				{SQLObjectID(oid2), SQLFloat(1), 1},
				{SQLObjectID(oid2), SQLVarchar("cba"), 1},
				{SQLObjectID(oid2), SQLBool(false), -1},
				{SQLObjectID(oid2), SQLBool(true), -1},
				{SQLObjectID(oid2), &SQLValues{[]SQLValue{SQLInt(1)}}, 1},
				{SQLObjectID(oid2), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLObjectID(oid2), SQLDate{now}, -1},
				{SQLObjectID(oid2), SQLTimestamp{now}, -1},
				{SQLObjectID(oid2), SQLObjectID(oid3), -1},
				{SQLObjectID(oid2), SQLObjectID(oid2), 0},
				{SQLObjectID(oid2), SQLObjectID(oid1), 1},
			}
			runTests(tests)
		})

	})
}
