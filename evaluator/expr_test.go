package evaluator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
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
					test{"AVG('a')", SQLFloat(0)},
					test{"AVG(-20)", SQLFloat(-20)},
					test{"AVG(20)", SQLFloat(20)},
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

			Convey("Subject: MIN", func() {
				tests := []test{
					test{"MIN(NULL)", SQLNull},
					test{"MIN(a)", SQLInt(3)},
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
					test{"MAX(a)", SQLInt(5)},
					test{"MAX(b)", SQLInt(6)},
					test{"MAX(c)", SQLNull},
					test{"MAX('a')", SQLVarchar("a")},
					test{"MAX(-20)", SQLInt(-20)},
					test{"MAX(20)", SQLInt(20)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: SUM", func() {
				tests := []test{
					test{"SUM(NULL)", SQLNull},
					test{"SUM(a)", SQLFloat(8)},
					test{"SUM(b)", SQLFloat(9)},
					test{"SUM(c)", SQLNull},
					test{"SUM('a')", SQLFloat(0)},
					test{"SUM(-20)", SQLFloat(-60)},
					test{"SUM(20)", SQLFloat(60)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: STDDEV_POP", func() {
				tests := []test{
					test{"STD(NULL)", SQLNull},
					test{"STDDEV(a)", SQLFloat(1)},
					test{"STDDEV_POP(b)", SQLFloat(1.5)},
					test{"STD(c)", SQLNull},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: STDDEV_SAMP", func() {
				tests := []test{
					test{"STDDEV_SAMP(NULL)", SQLNull},
					test{"STDDEV_SAMP(a)", SQLFloat(1.4142135623730951)},
					test{"STDDEV_SAMP(b)", SQLFloat(2.1213203435596424)},
					test{"STDDEV_SAMP(c)", SQLNull},
				}
				runTests(aggCtx, tests)
			})

		})

		Convey("Subject: SQLAndExpr", func() {
			tests := []test{
				test{"1 AND 1", SQLTrue},
				test{"1 AND 0", SQLFalse},
				test{"0 AND 1", SQLFalse},
				test{"0 AND 0", SQLFalse},
				test{"1 && 1", SQLTrue},
				test{"1 && 0", SQLFalse},
				test{"0 && 1", SQLFalse},
				test{"0 && 0", SQLFalse},
				test{"NULL && 0", SQLFalse},
				test{"NULL && 1", SQLNull},
				test{"NULL && NULL", SQLNull},
				test{"true AND true", SQLTrue},
				test{"true AND false", SQLFalse},
				test{"false AND true", SQLFalse},
				test{"false AND false", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLAssignmentExpr", func() {
			e := &SQLAssignmentExpr{
				variable: &SQLVariableExpr{
					Name: "test",
					Kind: UserVariable,
				},
				expr: &SQLAddExpr{
					left:  SQLInt(1),
					right: SQLInt(3),
				},
			}

			result, err := e.Evaluate(evalCtx)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, SQLInt(4))
		})

		Convey("Subject: SQLDateTimeArithmetic", func() {

			Convey("Subject: Add", func() {
				tests := []test{
					test{"DATE '2014-04-13' + 0", SQLInt(20140413)},
					test{"DATE '2014-04-13' + 2", SQLInt(20140415)},
					test{"TIME '11:04:13' + 0", SQLInt(110413)},
					test{"TIME '11:04:13' + 2", SQLInt(110415)},
					test{"TIME '11:04:13' + '2'", SQLInt(110415)},
					test{"'2' + TIME '11:04:13'", SQLInt(110415)},
					test{"TIMESTAMP '2014-04-13 11:04:13' + 0", SQLInt(20140413110413)},
					test{"TIMESTAMP '2014-04-13 11:04:13' + 2", SQLInt(20140413110415)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Subtract", func() {
				tests := []test{
					test{"DATE '2014-04-13' - 0", SQLInt(20140413)},
					test{"DATE '2014-04-13' - 2", SQLInt(20140411)},
					test{"TIME '11:04:13' - 0", SQLInt(110413)},
					test{"TIME '11:04:13' - 2", SQLInt(110411)},
					test{"TIME '11:04:13' - '2'", SQLInt(110411)},
					test{"TIMESTAMP '2014-04-13 11:04:13' - 0", SQLInt(20140413110413)},
					test{"TIMESTAMP '2014-04-13 11:04:13' - 2", SQLInt(20140413110411)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Multiply", func() {
				tests := []test{
					test{"DATE '2014-04-13' * 0", SQLInt(0)},
					test{"DATE '2014-04-13' * 2", SQLInt(40280826)},
					test{"TIME '11:04:13' * 0", SQLInt(0)},
					test{"TIME '11:04:13' * 2", SQLInt(220826)},
					test{"TIME '11:04:13' * '2'", SQLInt(220826)},
					test{"TIMESTAMP '2014-04-13 11:04:13' * 0", SQLInt(0)},
					test{"TIMESTAMP '2014-04-13 11:04:13' * 2", SQLInt(40280826220826)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Divide", func() {
				tests := []test{
					test{"DATE '2014-04-13' / 0", SQLNull},
					test{"DATE '2014-04-13' / 2", SQLFloat(10070206.5)},
					test{"TIME '11:04:13' / 0", SQLNull},
					test{"TIME '11:04:13' / 2", SQLFloat(55206.5)},
					test{"TIME '11:04:13' / '2'", SQLFloat(55206.5)},
					test{"TIMESTAMP '2014-04-13 11:04:13' / 0", SQLNull},
					test{"TIMESTAMP '2014-04-13 11:04:13' / 2", SQLFloat(10070206555206.5)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Less Than", func() {
				tests := []test{
					test{"DATE '2014-04-13' > 0", SQLTrue},
					test{"DATE '2014-04-13' > DATE '2014-04-14'", SQLFalse},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Greater Than", func() {
				tests := []test{
					test{"DATE '2014-04-13' > 0", SQLTrue},
					test{"DATE '2014-04-13' > DATE '2014-04-14'", SQLFalse},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Equal", func() {
				tests := []test{
					test{"DATE '2014-04-13' = '0'", SQLFalse},
					test{"DATE '2014-04-13' = DATE '2014-04-13'", SQLTrue},
				}
				runTests(evalCtx, tests)
			})
		})

		Convey("Subject: SQLCaseExpr", func() {
			tests := []test{
				test{"CASE 3 WHEN 3 THEN 'three' WHEN 1 THEN 'one' ELSE 'else' END", SQLVarchar("three")},
				test{"CASE WHEN 5 > 3 THEN 'true' else 'false' END", SQLVarchar("true")},
				test{"CASE WHEN a = 123 THEN 'yes' else 'no' END", SQLVarchar("yes")},
				test{"CASE WHEN a = 245 THEN 'yes' END", SQLNull},
			}
			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLDateTimeLiterals", func() {

			Convey("Subject: DATE", func() {
				dateTime, _ := time.Parse("2006-01-02", "2014-04-13")
				tests := []test{
					test{"DATE '2014-04-13'", SQLDate{Time: dateTime}},
					test{"{d '2014-04-13'}", SQLDate{Time: dateTime}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TIME", func() {
				dateTime, _ := time.Parse("15:04:05", "11:49:36")
				tests := []test{
					test{"TIME '11:49:36'", SQLTimestamp{Time: dateTime}},
					test{"{t '11:49:36'}", SQLTimestamp{Time: dateTime}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TIMESTAMP", func() {
				dateTime, _ := time.Parse("2006-01-02 15:04:05.999999999", "1997-01-31 09:26:50.124")
				tests := []test{
					test{"TIMESTAMP '1997-01-31 09:26:50.124'", SQLTimestamp{Time: dateTime}},
					test{"{ts '1997-01-31 09:26:50.124'}", SQLTimestamp{Time: dateTime}},
				}
				runTests(evalCtx, tests)
			})

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

		Convey("Subject: SQLIsExpr", func() {
			tests := []test{
				test{"1 is true", SQLTrue},
				test{"null is true", SQLFalse},
				test{"null is unknown", SQLTrue},
				test{"1 is unknown", SQLFalse},
				test{"true is true", SQLTrue},
				test{"0 is false", SQLTrue},
				test{"1 is false", SQLFalse},
				test{"'1' is true", SQLTrue},
				test{"'0.0' is true", SQLFalse},
				test{"'cats' is false", SQLTrue},
				test{"DATE '2006-05-04' is false", SQLFalse},
				test{"TIMESTAMP '2008-04-06 15:32:23' is true", SQLTrue},
				test{"1 is null", SQLFalse},
				test{"null is null", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLIsNotExpr", func() {
			tests := []test{
				test{"1 is not true", SQLFalse},
				test{"null is not true", SQLTrue},
				test{"null is not unknown", SQLFalse},
				test{"1 is not unknown", SQLTrue},
				test{"false is not true", SQLTrue},
				test{"0 is not false", SQLFalse},
				test{"1 is not false", SQLTrue},
				test{"'1' is not true", SQLFalse},
				test{"'0.0' is not true", SQLTrue},
				test{"'cats' is not false", SQLFalse},
				test{"DATE '2006-05-04' is not false", SQLTrue},
				test{"TIMESTAMP '2008-04-06 15:32:23' is not true", SQLFalse},
				test{"1 is not null", SQLTrue},
				test{"null is not null", SQLFalse},
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
				test{"NULL IN(0, 1)", SQLNull},
				test{"NULL IN(0, NULL)", SQLNull},
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
				test{"'Á Â Ã Ä' LIKE '%'", SQLTrue},
				test{"'Á Â Ã Ä' LIKE 'Á Â Ã Ä'", SQLTrue},
				test{"'Á Â Ã Ä' LIKE 'Á%'", SQLTrue},
				test{"'a' LIKE 'a'", SQLTrue},
				test{"'Adam' LIKE 'am'", SQLFalse},
				test{"'Adam' LIKE 'adaM'", SQLTrue}, // mixed case test
				test{"'Adam' LIKE '%am%'", SQLTrue},
				test{"'Adam' LIKE 'Ada_'", SQLTrue},
				test{"'Adam' LIKE '__am'", SQLTrue},
				test{"'Clever' LIKE '%is'", SQLFalse},
				test{"'Adam is nice' LIKE '%xs '", SQLFalse},
				test{"'Adam is nice' LIKE '%is nice'", SQLTrue},
				test{"'abc' LIKE 'ABC'", SQLTrue},    //case sensitive test
				test{"'abc' LIKE 'ABC '", SQLFalse},  // trailing space test
				test{"'abc' LIKE ' ABC'", SQLFalse},  // leading space test
				test{"'abc' LIKE ' ABC '", SQLFalse}, // padded space test
				test{"'abc' LIKE 'ABC	'", SQLFalse}, // trailing tab test
				test{"'10' LIKE '1%'", SQLTrue},
				test{"'a   ' LIKE 'A   '", SQLTrue},
				test{"CURRENT_DATE() LIKE '2015-05-31%'", SQLFalse},
				test{"(DATE '2008-01-02') LIKE '2008-01%'", SQLTrue},
				test{"NOW() LIKE '" + strconv.Itoa(time.Now().Year()) + "%' ", SQLTrue},
				test{"10 LIKE '1%'", SQLTrue},
				test{"1.20 LIKE '1.2%'", SQLTrue},
				test{"NULL LIKE '1%'", SQLNull},
				test{"10 LIKE NULL", SQLNull},
				test{"NULL LIKE NULL", SQLNull},
				test{"'David_' LIKE 'David\\_'", SQLTrue},
				test{"'David%' LIKE 'David\\%'", SQLTrue},
			}
			runTests(evalCtx, tests)
		})

		Convey("Subject: Mix Arithmetic and Boolean", func() {
			tests := []test{
				test{"(5<6) + 1", SQLInt(2)},
				test{"(5<6) && (6>4)", SQLTrue},
				test{"(5<6) || (6>4)", SQLTrue},
				test{"(5<6) XOR (6>4)", SQLFalse},
				test{"(5<6)<7", SQLTrue},
				test{"1+(5<6)", SQLInt(2)},
				test{"1+(5>6)", SQLInt(1)},
				test{"1+(NULL>6)", SQLNull},
				test{"NULL+(5>6)", SQLNull},
				test{"20/(5<6)", SQLFloat(20)},
				test{"20*(5<6)", SQLInt(20)},
				test{"20/5<6", SQLTrue},
				test{"20*5<6", SQLFalse},
				test{"20+5<6", SQLFalse},
				test{"20-5<6", SQLFalse},
				test{"20+true", SQLInt(21)},
				test{"20+false", SQLInt(20)},
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
			tests := []test{
				test{"NOT 1", SQLFalse},
				test{"NOT 0", SQLTrue},
				test{"NOT true", SQLFalse},
				test{"NOT false", SQLTrue},
				test{"NOT NULL", SQLNull},
				test{"! 1", SQLFalse},
				test{"! 0", SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLOrExpr", func() {
			tests := []test{
				test{"1 OR 1", SQLTrue},
				test{"1 OR 0", SQLTrue},
				test{"0 OR 1", SQLTrue},
				test{"NULL OR 1", SQLTrue},
				test{"NULL OR 0", SQLNull},
				test{"NULL OR NULL", SQLNull},
				test{"0 OR 0", SQLFalse},
				test{"true OR true", SQLTrue},
				test{"true OR false", SQLTrue},
				test{"false OR true", SQLTrue},
				test{"false OR false", SQLFalse},
				test{"1 || 1", SQLTrue},
				test{"1 || 0", SQLTrue},
				test{"0 || 1", SQLTrue},
				test{"0 || 0", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLXOrExpr", func() {
			tests := []test{
				test{"1 XOR 1", SQLFalse},
				test{"1 XOR 0", SQLTrue},
				test{"0 XOR 1", SQLTrue},
				test{"0 XOR 0", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNotRegexExpr", func() {
			tests := []test{
				test{"'ABC123' NOT REGEXP 'AB'", SQLFalse},
				test{"'ABC123' NOT REGEXP 'ABD'", SQLTrue},
				test{"'ABC123' NOT REGEXP '[[:alpha:]]'", SQLFalse},
				test{"'fofo' NOT REGEXP '^fo'", SQLFalse},
				test{"'fofo' NOT REGEXP '^f.*$'", SQLFalse},
				test{"'pi' NOT REGEXP 'pi|apa'", SQLFalse},
				test{"'abcde' NOT REGEXP 'a[bcd]{2}e'", SQLTrue},
				test{"'abcde' NOT REGEXP 'a[bcd]{1,10}e'", SQLFalse},
				test{"null REGEXP 'abc'", SQLNull},
				test{"'a' REGEXP null", SQLNull},
				test{"2-1 NOT REGEXP '1'", SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLRegexExpr", func() {
			tests := []test{
				test{"'ABC123' REGEXP 'AB'", SQLTrue},
				test{"'ABC123' REGEXP 'ABD'", SQLFalse},
				test{"'ABC123' REGEXP '[[:alpha:]]'", SQLTrue},
				test{"'fofo' REGEXP '^fo'", SQLTrue},
				test{"'fofo' REGEXP '^f.*$'", SQLTrue},
				test{"'pi' REGEXP 'pi|apa'", SQLTrue},
				test{"'abcde' REGEXP 'a[bcd]{2}e'", SQLFalse},
				test{"'abcde' REGEXP 'a[bcd]{1,10}e'", SQLTrue},
				test{"null REGEXP 'abc'", SQLNull},
				test{"'a' REGEXP null", SQLNull},
				test{"2-1 REGEXP '1'", SQLTrue},
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

			Convey("Subject: CONVERT", func() {
				d, err := time.Parse("2006-01-02", "2006-05-11")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 12:32:12")
				So(err, ShouldBeNil)
				dt, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 00:00:00")
				So(err, ShouldBeNil)

				tests := []test{
					test{"CONVERT(NULL, SIGNED)", SQLNull},
					test{"CONVERT(3, SIGNED)", SQLInt(3)},
					test{"CONVERT(3.4, SIGNED)", SQLInt(3)},
					test{"CONVERT(3.5, SIGNED INTEGER)", SQLInt(4)},
					test{"CONVERT(-3.4, SIGNED INTEGER)", SQLInt(-3)},
					test{"CONVERT(33245368230, SQL_BIGINT)", SQLInt(33245368230)},
					test{"CONVERT('janna', UNSIGNED INTEGER)", SQLInt(0)},
					test{"CONVERT('423', UNSIGNED)", SQLInt(423)},
					test{"CONVERT('16a', SIGNED)", SQLInt(0)},
					test{"CONVERT(true, SIGNED)", SQLInt(1)},
					test{"CONVERT(false, SIGNED)", SQLInt(0)},
					test{"CONVERT(DATE '2006-05-11', SIGNED)", SQLInt(20060511)},
					test{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', SIGNED)", SQLInt(20060511123212)},
					test{"CONVERT(NULL, SQL_DOUBLE)", SQLNull},
					test{"CONVERT(3, SQL_DOUBLE)", SQLFloat(3)},
					test{"CONVERT(-3.4, SQL_DOUBLE)", SQLFloat(-3.4)},
					test{"CONVERT('janna', SQL_DOUBLE)", SQLFloat(0)},
					test{"CONVERT('4.4', SQL_DOUBLE)", SQLFloat(4.4)},
					test{"CONVERT('16a', SQL_DOUBLE)", SQLFloat(0)},
					test{"CONVERT(true, SQL_DOUBLE)", SQLFloat(1)},
					test{"CONVERT(false, SQL_DOUBLE)", SQLFloat(0)},
					test{"CONVERT(DATE '2006-05-11', SQL_DOUBLE)", SQLFloat(20060511)},
					test{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', SQL_DOUBLE)", SQLFloat(20060511123212)},
					test{"CONVERT(NULL, CHAR)", SQLNull},
					test{"CONVERT(3, CHAR)", SQLVarchar("3")},
					test{"CONVERT(-3.4, SQL_VARCHAR)", SQLVarchar("-3.4")},
					test{"CONVERT('janna', CHAR)", SQLVarchar("janna")},
					test{"CONVERT('16a', CHAR)", SQLVarchar("16a")},
					test{"CONVERT(true, CHAR)", SQLVarchar("1")},
					test{"CONVERT(false, CHAR)", SQLVarchar("0")},
					test{"CONVERT(DATE '2006-05-11', CHAR)", SQLVarchar("2006-05-11")},
					test{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', CHAR)", SQLVarchar("2006-05-11 12:32:12")},
					test{"CONVERT(NULL, DATE)", SQLNull},
					test{"CONVERT(3, DATE)", SQLNull},
					test{"CONVERT(-3.4, SQL_DATE)", SQLNull},
					test{"CONVERT('janna', DATE)", SQLNull},
					test{"CONVERT('2006-05-11', DATE)", SQLDate{Time: d}},
					test{"CONVERT(true, DATE)", SQLNull},
					test{"CONVERT(DATE '2006-05-11', DATE)", SQLDate{Time: d}},
					test{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATE)", SQLDate{Time: d}},
					test{"CONVERT(NULL, DATETIME)", SQLNull},
					test{"CONVERT(-3.4, DATETIME)", SQLNull},
					test{"CONVERT('janna', DATETIME)", SQLNull},
					test{"CONVERT('2006-05-11', DATETIME)", SQLTimestamp{Time: dt}},
					test{"CONVERT(true, DATETIME)", SQLNull},
					test{"CONVERT(3, SQL_TIMESTAMP)", SQLNull},
					test{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATETIME)", SQLTimestamp{Time: t}},
					test{"CONVERT(DATE '2006-05-11', SQL_TIMESTAMP)", SQLTimestamp{Time: dt}},
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
					test{"CURRENT_TIMESTAMP", SQLTimestamp{time.Now().UTC()}},
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

			Convey("Subject: DATE", func() {
				d, err := time.Parse("2006-01-02", "2016-03-01")
				So(err, ShouldBeNil)

				tests := []test{
					test{"DATE(NULL)", SQLNull},
					test{"DATE(23)", SQLNull},
					test{"DATE('cat')", SQLNull},
					test{"DATE(TIMESTAMP '2016-03-01 12:32:23')", SQLDate{Time: d}},
					test{"DATE(DATE '2016-03-01')", SQLDate{Time: d}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DATE_ADD", func() {
				d, err := time.Parse("2006-01-02", "2003-01-02")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
				So(err, ShouldBeNil)
				d2, err := time.Parse("2006-01-02", "2003-11-30")
				So(err, ShouldBeNil)

				tests := []test{
					test{"DATE_ADD('2002-01-02', INTERVAL 1 YEAR)", SQLDate{Time: d}},
					test{"DATE_ADD('2003-08-31', INTERVAL 1 QUARTER)", SQLDate{Time: d2}},
					test{"DATE_ADD('2003-10-31', INTERVAL 1 MONTH)", SQLDate{Time: d2}},
					test{"DATE_ADD('2003-01-01', INTERVAL 1 DAY)", SQLDate{Time: d}},
					test{"DATE_ADD('2003-01-02 14:30:09', INTERVAL -2 HOUR)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-02 12:23:09', INTERVAL 7 MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-02 12:30:12', INTERVAL -3 SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' MINUTE_SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-02 05:27:06', INTERVAL '7:3:3' HOUR_SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-02 15:32:09', INTERVAL '-3:2' HOUR_MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2002-12-31 10:27:05', INTERVAL '2 2:3:4' DAY_SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2002-12-31 10:27:09', INTERVAL '2 2:3' DAY_MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-01 08:30:09', INTERVAL '1 4' DAY_HOUR)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2000-09-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-02 12:33:09', INTERVAL '-3' HOUR_MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_ADD('2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)", SQLTimestamp{Time: t}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DATE_SUB", func() {
				d, err := time.Parse("2006-01-02", "2003-01-02")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
				So(err, ShouldBeNil)
				d2, err := time.Parse("2006-01-02", "2003-11-30")
				So(err, ShouldBeNil)

				tests := []test{
					test{"DATE_SUB('2004-01-02', INTERVAL 1 YEAR)", SQLDate{Time: d}},
					test{"DATE_SUB('2003-04-02', INTERVAL 1 QUARTER)", SQLDate{Time: d}},
					test{"DATE_SUB('2003-12-31', INTERVAL 1 MONTH)", SQLDate{Time: d2}},
					test{"DATE_SUB('2003-01-03', INTERVAL 1 DAY)", SQLDate{Time: d}},
					test{"DATE_SUB('2003-01-02 10:30:09', INTERVAL -2 HOUR)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-02 12:37:09', INTERVAL 7 MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-02 12:30:12', INTERVAL 3 SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-02 12:32:10', INTERVAL '2:1' MINUTE_SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-02 19:33:12', INTERVAL '7:3:3' HOUR_SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-02 15:32:09', INTERVAL '3:2' HOUR_MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-04 14:33:13', INTERVAL '2 2:3:4' DAY_SECOND)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-04 14:33:09', INTERVAL '2 2:3' DAY_MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-03 16:30:09', INTERVAL '1 4' DAY_HOUR)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2005-05-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-02 12:33:09', INTERVAL '3' HOUR_MINUTE)", SQLTimestamp{Time: t}},
					test{"DATE_SUB('2003-01-02 14:32:12', INTERVAL '2 2:3' DAY_SECOND)", SQLTimestamp{Time: t}},
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
					test{"EXP('sdg')", SQLFloat(1)},
					test{"EXP(0)", SQLFloat(1)},
					test{"EXP(2)", SQLFloat(7.38905609893065)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: EXTRACT", func() {
				tests := []test{
					test{"EXTRACT(YEAR FROM NULL)", SQLNull},
					test{"EXTRACT(YEAR FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(2006)},
					test{"EXTRACT(QUARTER FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(2)},
					test{"EXTRACT(WEEK FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(14)},
					test{"EXTRACT(DAY FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(7)},
					test{"EXTRACT(HOUR FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(7)},
					test{"EXTRACT(MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(14)},
					test{"EXTRACT(SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(23)},
					test{"EXTRACT(MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(0)},
					test{"EXTRACT(YEAR_MONTH FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(200604)},
					test{"EXTRACT(DAY_HOUR FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(707)},
					test{"EXTRACT(DAY_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(70714)},
					test{"EXTRACT(DAY_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(7071423)},
					test{"EXTRACT(DAY_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(7071423000000)},
					test{"EXTRACT(HOUR_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(714)},
					test{"EXTRACT(HOUR_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(71423)},
					test{"EXTRACT(HOUR_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(71423000000)},
					test{"EXTRACT(MINUTE_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(1423)},
					test{"EXTRACT(MINUTE_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(1423000000)},
					test{"EXTRACT(SECOND_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(23000000)},
					test{"EXTRACT(SQL_TSI_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", SQLInt(14)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: FLOOR", func() {
				tests := []test{
					test{"FLOOR(NULL)", SQLNull},
					test{"FLOOR('sdg')", SQLFloat(0)},
					test{"FLOOR(1.23)", SQLFloat(1)},
					test{"FLOOR(-1.23)", SQLFloat(-2)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: GREATEST", func() {
				d, err := time.Parse("2006-01-02", "2006-05-11")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 12:32:23")
				So(err, ShouldBeNil)

				tests := []test{
					test{"GREATEST(NULL, 1, 2)", SQLNull},
					test{"GREATEST(1,3,2)", SQLInt(3)},
					test{"GREATEST(2,2.3)", SQLFloat(2.3)},
					test{"GREATEST('cats', '4', '2')", SQLVarchar("cats")},
					test{"GREATEST('dog', 'cats', 'bird')", SQLVarchar("dog")},
					test{"GREATEST('cat', 'bird', 2)", SQLInt(2)},
					test{"GREATEST('cat', 2.2)", SQLFloat(2.2)},
					test{"GREATEST(false, true)", SQLBool(true)},
					test{"GREATEST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')", SQLDate{Time: d}},
					test{"GREATEST(DATE '2006-05-11', 14, 4235)", SQLInt(20060511)},
					test{"GREATEST(DATE '2006-05-11', 14, 20080622)", SQLInt(20080622)},
					test{"GREATEST(DATE '2006-05-11', 14, 20080622.1)", SQLFloat(20080622.1)},
					test{"GREATEST(DATE '2006-05-11', 14, 4235.2)", SQLFloat(20060511.0)},
					test{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')", SQLTimestamp{Time: t}},
					test{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)", SQLInt(20060511123223)},
					test{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)", SQLFloat(20080923124345.3)},
					test{"GREATEST(DATE '2006-05-11', 'cat', '2007-04-11')", SQLVarchar("2007-04-11")},
					test{"GREATEST(DATE '2006-05-11', 20080912, '2007-04-11')", SQLInt(20080912)},
					test{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:45')", SQLTimestamp{Time: t}},
					test{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')", SQLInt(20060511123223)},
					test{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2008-09-13')", SQLVarchar("2008-09-13")},
					test{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')", SQLTimestamp{Time: t}},
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

			Convey("Subject: IF", func() {
				tests := []test{
					test{"IF(1<2, 4, 5)", SQLInt(4)},
					test{"IF(1>2, 4, 5)", SQLInt(5)},
					test{"IF(1, 4, 5)", SQLInt(4)},
					test{"IF(-0, 4, 5)", SQLInt(5)},
					test{"IF(1-1, 4, 5)", SQLInt(5)},
					test{"IF('cat', 4, 5)", SQLInt(5)},
					test{"IF('3', 4, 5)", SQLInt(4)},
					test{"IF('0', 4, 5)", SQLInt(5)},
					test{"IF('-0.0', 4, 5)", SQLInt(5)},
					test{"IF('2.2', 4, 5)", SQLInt(4)},
					test{"IF('true', 4, 5)", SQLInt(5)},
					test{"IF(null, 4, 'cat')", SQLVarchar("cat")},
					test{"IF(true, 'dog', 'cat')", SQLVarchar("dog")},
					test{"IF(false, 'dog', 'cat')", SQLVarchar("cat")},
					test{"IF('ca.gh', 4, 5)", SQLInt(5)},
					test{"IF(current_timestamp(), 4, 5)", SQLInt(4)}, // not being parsed as dates, being parsed as string
					test{"IF(current_timestamp, 4, 5)", SQLInt(4)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: IFNULL", func() {
				tests := []test{
					test{"IFNULL(1,0)", SQLInt(1)},
					test{"IFNULL(NULL,3)", SQLInt(3)},
					test{"IFNULL(NULL,NULL)", SQLNull},
					test{"IFNULL('cat', null)", SQLVarchar("cat")},
					test{"IFNULL(null, 'dog')", SQLVarchar("dog")},
					test{"IFNULL(1/0, 4)", SQLInt(4)},
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

			Convey("Subject: LEAST", func() {
				d, err := time.Parse("2006-01-02", "2005-05-11")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 00:00:00")
				So(err, ShouldBeNil)
				t1, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 10:32:23")
				So(err, ShouldBeNil)

				tests := []test{
					test{"LEAST(NULL, 1, 2)", SQLNull},
					test{"LEAST(1,3,2)", SQLInt(1)},
					test{"LEAST(2,2.3)", SQLFloat(2.0)},
					test{"LEAST('cats', '4', '2')", SQLVarchar("2")},
					test{"LEAST('dog', 'cats', 'bird')", SQLVarchar("bird")},
					test{"LEAST(false, true)", SQLBool(false)},
					test{"LEAST(DATE '2005-05-11', DATE '2006-05-11', DATE '2007-05-11')", SQLDate{Time: d}},
					test{"LEAST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')", SQLTimestamp{Time: t}},
					test{"LEAST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:23')", SQLTimestamp{Time: t1}},
					test{"LEAST('cat', 'bird', 2)", SQLInt(0)},
					test{"LEAST('cat', 2.2)", SQLFloat(0)},
					test{"LEAST(DATE '2006-05-11', 14, 4235)", SQLInt(14)},
					test{"LEAST(DATE '2006-05-11', 14, 20080622.1)", SQLFloat(14.0)},
					test{"LEAST(DATE '2006-05-11', 14, 4235.2)", SQLFloat(14.0)},
					test{"LEAST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)", SQLInt(12)},
					test{"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)", SQLFloat(20060511123223.0)},
					test{"LEAST(DATE '2006-05-11', 'cat', '2007-04-11')", SQLVarchar("cat")},
					test{"LEAST(DATE '2006-05-11', 20080912, '2007-04-11')", SQLInt(0)},
					test{"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')", SQLInt(20070823)},
					test{"LEAST(TIMESTAMP '2006-05-11 10:32:23', '2008-09-13')", SQLTimestamp{Time: t1}},
					test{"LEAST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')", SQLVarchar("2005-09-13")},
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
					test{"LN(-16.5)", SQLNull},
					test{"LOG(NULL)", SQLNull},
					test{"LOG(1)", SQLFloat(0)},
					test{"LOG(16.5)", SQLFloat(2.803360380906535)},
					test{"LOG(-16.5)", SQLNull},
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
					test{"LOCATE('bar', 'foobarbar', 4)", SQLInt(4)},
					test{"LOCATE('e', 'dvd', 6)", SQLInt(0)},
					test{"LOCATE('f', 'asdf', 4)", SQLInt(4)},
					test{"LOCATE('語', '日本語')", SQLInt(3)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOG2", func() {
				tests := []test{
					test{"LOG2(NULL)", SQLNull},
					test{"LOG2(4)", SQLFloat(2)},
					test{"LOG2(-100)", SQLNull},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOG10", func() {
				tests := []test{
					test{"LOG10(NULL)", SQLNull},
					test{"LOG10('sdg')", SQLNull},
					test{"LOG10(2)", SQLFloat(0.3010299956639812)},
					test{"LOG10(100)", SQLFloat(2)},
					test{"LOG10(0)", SQLNull},
					test{"LOG10(-100)", SQLNull},
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

			Convey("Subject: MAKEDATE", func() {
				d, err := time.Parse("2006-01-02", "2000-02-01")
				So(err, ShouldBeNil)
				d1, err := time.Parse("2006-01-02", "2012-02-01")
				So(err, ShouldBeNil)
				d2, err := time.Parse("2006-01-02", "1977-03-07")
				So(err, ShouldBeNil)

				tests := []test{
					test{"MAKEDATE(NULL, 4)", SQLNull},
					test{"MAKEDATE(2004, 0)", SQLNull},
					test{"MAKEDATE('sdg', 32)", SQLDate{Time: d}},
					test{"MAKEDATE(2000, 32)", SQLDate{Time: d}},
					test{"MAKEDATE(12, 32)", SQLDate{Time: d1}},
					test{"MAKEDATE(77, 66)", SQLDate{Time: d2}},
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

			Convey("Subject: NULLIF", func() {
				tests := []test{
					test{"NULLIF(1,1)", SQLNull},
					test{"NULLIF(1,3)", SQLInt(1)},
					test{"NULLIF(null, null)", SQLNull},
					test{"NULLIF(null, 4)", SQLNull},
					test{"NULLIF(3, null)", SQLInt(3)},
					//test{"NULLIF(3, '3')", SQLNull},
					test{"NULLIF('abc', 'abc')", SQLNull},
					//test{"NULLIF('abc', 3)", SQLVarchar("abc")},
					//test{"NULLIF('1', true)", SQLNull},
					//test{"NULLIF('1', false)", SQLVarchar("1")},
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
					test{"SQRT('sdg')", SQLFloat(0)},
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
					test{"SUBSTRING('Quadratically', 12, 2)", SQLVarchar("ly")},
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

			Convey("Subject: STR_TO_DATE", func() {
				d, err := time.Parse("2006-01-02", "2016-04-03")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2016-04-03 12:22:22")
				So(err, ShouldBeNil)
				t1, err := time.Parse("2006-01-02 15:04:05", "2005-04-02 00:12:00")
				So(err, ShouldBeNil)
				t2, err := time.Parse("2006-01-02 15:04:05", "2016-04-03 12:22:00")
				So(err, ShouldBeNil)

				tests := []test{
					test{"STR_TO_DATE(NULL, 4)", SQLNull},
					test{"STR_TO_DATE('foobarbar', NULL)", SQLNull},
					test{"STR_TO_DATE('2016-04-03','%Y-%m-%d')", SQLDate{d}},
					test{"STR_TO_DATE('04,03,2016', '%m,%d,%Y')", SQLDate{d}},
					test{"STR_TO_DATE('04,03,a16', '%m,%d,a%y')", SQLDate{d}},
					test{"STR_TO_DATE('2016-04-03 12:22:22', '%Y-%m-%d %H:%i:%s')", SQLTimestamp{t}},
					test{"STR_TO_DATE('2016-04-03 12:22', '%Y-%m-%d %H:%i')", SQLTimestamp{t2}},
					test{"STR_TO_DATE('2005-04-02 12', '%Y-%m-%d %i')", SQLTimestamp{t1}},
					test{"STR_TO_DATE('Apr 03, 2016', '%b %d, %Y')", SQLDate{d}},
					test{"STR_TO_DATE('Tue 2016-04-03', '%a %Y-%m-%d')", SQLDate{d}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TIMESTAMPADD", func() {
				d, err := time.Parse("2006-01-02", "2003-01-02")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
				So(err, ShouldBeNil)
				dt, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 01:00:00")
				So(err, ShouldBeNil)

				tests := []test{
					test{"TIMESTAMPADD(YEAR, 1, DATE '2002-01-02')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(QUARTER, 1, DATE '2002-10-02')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(MONTH, 1, DATE '2002-12-02')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(WEEK, 1, DATE '2002-12-26')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(DAY, 1, DATE '2003-01-01')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(HOUR, 1, DATE '2003-01-02')", SQLTimestamp{Time: dt}},
					test{"TIMESTAMPADD(MINUTE, 60, DATE '2003-01-02')", SQLTimestamp{Time: dt}},
					test{"TIMESTAMPADD(SECOND, 3600, DATE '2003-01-02')", SQLTimestamp{Time: dt}},
					test{"TIMESTAMPADD(MICROSECOND, 1, TIMESTAMP '2003-01-02 12:30:09')", SQLTimestamp{Time: t}},
					test{"TIMESTAMPADD(DAY, 1, TIMESTAMP '2003-01-01 12:30:09')", SQLTimestamp{Time: t}},
					test{"TIMESTAMPADD(WEEK, 2, TIMESTAMP '2002-12-19 12:30:09')", SQLTimestamp{Time: t}},
					test{"TIMESTAMPADD(SQL_TSI_YEAR, 2, TIMESTAMP '2001-01-02 12:30:09')", SQLTimestamp{Time: t}},
					test{"TIMESTAMPADD(SQL_TSI_QUARTER, 2, DATE '2002-07-02')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(SQL_TSI_MONTH, 1, TIMESTAMP '2002-12-02 12:30:09')", SQLTimestamp{Time: t}},
					test{"TIMESTAMPADD(SQL_TSI_WEEK, 1, DATE '2002-12-26')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(SQL_TSI_DAY, 1, DATE '2003-01-01')", SQLDate{Time: d}},
					test{"TIMESTAMPADD(SQL_TSI_HOUR, 1, TIMESTAMP '2003-01-02 11:30:09')", SQLTimestamp{Time: t}},
					test{"TIMESTAMPADD(SQL_TSI_MINUTE, 1, TIMESTAMP '2003-01-02 12:29:09')", SQLTimestamp{Time: t}},
					test{"TIMESTAMPADD(SQL_TSI_SECOND, 1, TIMESTAMP '2003-01-02 12:30:08')", SQLTimestamp{Time: t}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TIMESTAMPDIFF", func() {
				tests := []test{
					test{"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-02')", SQLInt(1)},
					test{"TIMESTAMPDIFF(YEAR, DATE '2002-01-02', DATE '2001-01-02')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(YEAR, DATE '2001-01-03', DATE '2002-01-02')", SQLInt(0)},
					test{"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-03')", SQLInt(1)},
					test{"TIMESTAMPDIFF(QUARTER, DATE '2002-04-02', DATE '2002-01-02')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-06-02')", SQLInt(1)},
					test{"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-07-02')", SQLInt(2)},
					test{"TIMESTAMPDIFF(QUARTER, DATE '2002-07-02', DATE '2002-01-02')", SQLInt(-2)},
					test{"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-01')", SQLInt(0)},
					test{"TIMESTAMPDIFF(MONTH, DATE '2002-02-01', DATE '2001-01-02')", SQLInt(-12)},
					test{"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-02')", SQLInt(1)},
					test{"TIMESTAMPDIFF(MONTH, DATE '2002-02-03', DATE '2002-01-02')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-16')", SQLInt(2)},
					test{"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-15')", SQLInt(1)},
					test{"TIMESTAMPDIFF(WEEK, DATE '2001-01-15', DATE '2001-01-02')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-17')", SQLInt(2)},
					test{"TIMESTAMPDIFF(DAY, DATE '2003-01-04', DATE '2003-01-16')", SQLInt(12)},
					test{"TIMESTAMPDIFF(DAY, DATE '2003-01-16', DATE '2003-01-04')", SQLInt(-12)},
					test{"TIMESTAMPDIFF(HOUR, DATE '2003-01-04', DATE '2003-01-06')", SQLInt(48)},
					test{"TIMESTAMPDIFF(MINUTE, DATE '2003-01-04', DATE '2003-01-06')", SQLInt(2880)},
					test{"TIMESTAMPDIFF(SECOND, DATE '2003-01-04', DATE '2003-01-05')", SQLInt(86400)},
					test{"TIMESTAMPDIFF(MICROSECOND, DATE '2003-01-04', DATE '2003-01-05')", SQLInt(86400000000)},
					test{"TIMESTAMPDIFF(MICROSECOND, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-02 13:40:33')", SQLInt(90624000000)},
					test{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', TIMESTAMP '2003-03-04 12:45:30')", SQLInt(1)},
					test{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', TIMESTAMP '2002-03-04 12:45:30')", SQLInt(0)},
					test{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-03-04 12:45:30', TIMESTAMP '2002-01-02 12:30:09')", SQLInt(0)},
					test{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2003-03-04 12:30:06', DATE '2002-03-04')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(SQL_TSI_YEAR, DATE '2004-03-04', TIMESTAMP '2003-03-04 12:30:06')", SQLInt(0)},
					test{"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-01-01', TIMESTAMP '2002-04-01 12:30:06')", SQLInt(1)},
					test{"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-04-01 12:30:06', DATE '2002-01-01')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-01-01 12:30:06', DATE '2002-04-01')", SQLInt(0)},
					test{"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-04-01', TIMESTAMP '2002-01-01 12:30:06')", SQLInt(0)},
					test{"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-01-01', TIMESTAMP '2002-03-01 12:30:09')", SQLInt(2)},
					test{"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-03-01 12:30:09', DATE '2002-01-01')", SQLInt(-2)},
					test{"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-03-01')", SQLInt(1)},
					test{"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-03-01', TIMESTAMP '2002-01-01 12:30:09')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-08')", SQLInt(0)},
					test{"TIMESTAMPDIFF(SQL_TSI_WEEK, DATE '2002-01-01', TIMESTAMP '2002-01-08 12:30:09')", SQLInt(1)},
					test{"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-08 12:30:09', DATE '2002-01-01')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(SQL_TSI_DAY, DATE '2002-01-01', TIMESTAMP '2002-01-02 12:30:09')", SQLInt(1)},
					test{"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-02 12:30:09', DATE '2002-01-01')", SQLInt(-1)},
					test{"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')", SQLInt(0)},
					test{"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')", SQLInt(11)},
					test{"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-02 11:02:33')", SQLInt(22)},
					test{"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-01 13:02:33')", SQLInt(32)},
					test{"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')", SQLInt(689)},
					test{"TIMESTAMPDIFF(SQL_TSI_SECOND, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-02 14:40:33')", SQLInt(94224)},
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

			Convey("Subject: WEEK", func() {
				tests := []test{
					test{"WEEK(NULL)", SQLNull},
					test{"WEEK('sdg')", SQLNull},
					test{"WEEK('2016-1-01 10:23:52')", SQLInt(0)},
					test{"WEEK(DATE '2009-1-01')", SQLInt(0)},
					test{"WEEK(DATE '2009-1-01',0)", SQLInt(0)},
					test{"WEEK(DATE '2009-1-01','str')", SQLInt(0)},
					test{"WEEK(DATE '2009-1-01',1)", SQLInt(1)},
					test{"WEEK(DATE '2009-1-01',2)", SQLInt(52)},
					test{"WEEK(DATE '2009-1-01',3)", SQLInt(1)},
					test{"WEEK(DATE '2009-1-01',4)", SQLInt(0)},
					test{"WEEK(DATE '2009-1-01',5)", SQLInt(0)},
					test{"WEEK(DATE '2009-1-01',6)", SQLInt(53)},
					test{"WEEK(DATE '2009-1-01',7)", SQLInt(52)},
					test{"WEEK(DATE '2009-1-05')", SQLInt(1)},
					test{"WEEK(DATE '2009-1-05',1)", SQLInt(2)},
					test{"WEEK(DATE '2009-1-05',2)", SQLInt(1)},
					test{"WEEK(DATE '2009-1-05',3)", SQLInt(2)},
					test{"WEEK(DATE '2009-1-05',4)", SQLInt(1)},
					test{"WEEK(DATE '2009-1-05',5)", SQLInt(1)},
					test{"WEEK(DATE '2009-1-05',6)", SQLInt(1)},
					test{"WEEK(DATE '2009-1-05',7)", SQLInt(1)},
					test{"WEEK(DATE '2009-12-31')", SQLInt(52)},
					test{"WEEK(DATE '2009-12-31',1)", SQLInt(53)},
					test{"WEEK(DATE '2009-12-31',2)", SQLInt(52)},
					test{"WEEK(DATE '2009-12-31',3)", SQLInt(53)},
					test{"WEEK(DATE '2009-12-31',4)", SQLInt(52)},
					test{"WEEK(DATE '2009-12-31',5)", SQLInt(52)},
					test{"WEEK(DATE '2009-12-31',6)", SQLInt(52)},
					test{"WEEK(DATE '2009-12-31',7)", SQLInt(52)},
					test{"WEEK(DATE '2007-12-31')", SQLInt(52)},
					test{"WEEK(DATE '2007-12-31',1)", SQLInt(53)},
					test{"WEEK(DATE '2007-12-31',2)", SQLInt(52)},
					test{"WEEK(DATE '2007-12-31',3)", SQLInt(1)},
					test{"WEEK(DATE '2007-12-31',4)", SQLInt(53)},
					test{"WEEK(DATE '2007-12-31',5)", SQLInt(53)},
					test{"WEEK(DATE '2007-12-31',6)", SQLInt(1)},
					test{"WEEK(DATE '2007-12-31',7)", SQLInt(53)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: WEEKDAY", func() {
				tests := []test{
					test{"WEEKDAY(NULL)", SQLNull},
					test{"WEEKDAY('sdg')", SQLNull},
					test{"WEEKDAY('2016-1-01 10:23:52')", SQLInt(4)},
					test{"WEEKDAY('2005-05-11')", SQLInt(2)},
					test{"WEEKDAY(DATE '2016-7-10')", SQLInt(6)},
					test{"WEEKDAY(DATE '2016-7-11')", SQLInt(0)},
					test{"WEEKDAY(TIMESTAMP '2016-7-13 21:22:23')", SQLInt(2)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: WEEKOFYEAR", func() {
				tests := []test{
					test{"WEEKOFYEAR(NULL)", SQLNull},
					test{"WEEKOFYEAR('sdg')", SQLNull},
					test{"WEEKOFYEAR('2008-02-20')", SQLInt(8)},
					test{"WEEKOFYEAR('2009-01-01')", SQLInt(1)},
					test{"WEEKOFYEAR(DATE '2009-01-05')", SQLInt(2)},
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

			SkipConvey("Subject: YEARWEEK", func() {
				tests := []test{
					test{"YEARWEEK(NULL)", SQLNull},
					test{"YEARWEEK('sdg')", SQLNull},
					test{"YEARWEEK('2000-01-01')", SQLInt(199252)},
					test{"YEARWEEK('2001-01-01')", SQLInt(200053)},
					test{"YEARWEEK('2002-01-01')", SQLInt(200152)},
					test{"YEARWEEK('2003-01-01')", SQLInt(200252)},
					test{"YEARWEEK('2004-01-01')", SQLInt(200352)},
					test{"YEARWEEK('2005-01-01')", SQLInt(200452)},
					test{"YEARWEEK('2006-01-01')", SQLInt(200601)},
					test{"YEARWEEK('2000-01-06')", SQLInt(200001)},
					test{"YEARWEEK('2001-01-06')", SQLInt(200053)},
					test{"YEARWEEK('2002-01-06')", SQLInt(200201)},
					test{"YEARWEEK('2003-01-06')", SQLInt(200301)},
					test{"YEARWEEK('2004-01-06')", SQLInt(200401)},
					test{"YEARWEEK('2005-01-06')", SQLInt(200501)},
					test{"YEARWEEK('2006-01-06')", SQLInt(200601)},
					test{"YEARWEEK('2000-01-01',1)", SQLInt(199252)},
					test{"YEARWEEK('2001-01-01',1)", SQLInt(200101)},
					test{"YEARWEEK('2002-01-01',1)", SQLInt(200201)},
					test{"YEARWEEK('2003-01-01',1)", SQLInt(200301)},
					test{"YEARWEEK('2004-01-01',1)", SQLInt(200401)},
					test{"YEARWEEK('2005-01-01',1)", SQLInt(200453)},
					test{"YEARWEEK('2006-01-01',1)", SQLInt(200552)},
					test{"YEARWEEK('2000-01-06',1)", SQLInt(200001)},
					test{"YEARWEEK('2001-01-06',1)", SQLInt(200101)},
					test{"YEARWEEK('2002-01-06',1)", SQLInt(200201)},
					test{"YEARWEEK('2003-01-06',1)", SQLInt(200301)},
					test{"YEARWEEK('2004-01-06',1)", SQLInt(200402)},
					test{"YEARWEEK('2005-01-06',1)", SQLInt(200501)},
					test{"YEARWEEK('2006-01-06',1)", SQLInt(200601)},
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
				test{"- 10", SQLInt(-10)},
				test{"- a", SQLInt(-123)},
				test{"- b", SQLInt(-456)},
				test{"- null", SQLNull},
				test{"- true", SQLInt(-1)},
				test{"- false", SQLInt(0)},
				test{"- date '2005-05-11'", SQLInt(-20050511)},
				test{"- timestamp '2005-05-11 12:22:04'", SQLInt(-20050511122204)},
				test{"- '4' ", SQLFloat(-4)},
				test{"- 6.7", SQLFloat(-6.7)},
				test{"- '3.3'", SQLFloat(-3.3)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLVariableExpr", func() {
			tests := []test{
				test{"@@test_variable", SQLInt(123)},
			}

			runTests(evalCtx, tests)

			Convey("Should error when unknown variable is used", func() {
				subject := &SQLVariableExpr{
					Name: "blah",
					Kind: SessionVariable,
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

func TestSQLLikeExprConvertToPattern(t *testing.T) {
	test := func(syntax, expected string) {
		Convey(fmt.Sprintf("XXX LIKE '%s' should convert to pattern '%s'", syntax, expected), func() {
			pattern := convertSQLValueToPattern(SQLVarchar(syntax))
			So(pattern, ShouldEqual, expected)
		})
	}

	Convey("Subject: SQLLikeExpr.convertToPattern", t, func() {
		test("David", "^David$")
		test("Da\\vid", "^David$")
		test("Da\\\\vid", "^Da\\\\vid$")
		test("Da_id", "^Da.id$")
		test("Da\\_id", "^Da_id$")
		test("Da%d", "^Da.*d$")
		test("Da\\%d", "^Da%d$")
		test("Sto_. %ow", "^Sto.\\. .*ow$")
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

	type test struct {
		input    interface{}
		sqlType  schema.SQLType
		sqlValue SQLValue
	}

	runTests := func(tests []test) {
		for _, t := range tests {
			Convey(fmt.Sprintf("%v (%T) as '%v' should convert into %v (%T)", t.input, t.input, t.sqlType, t.sqlValue, t.sqlValue), func() {
				So(NewSQLValue(t.input, t.sqlType).String(), ShouldEqual, t.sqlValue.String())
			})
		}
	}

	getDate := func(t time.Time) time.Time {
		y, m, d := t.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, schema.DefaultLocale)
	}

	var (
		intVal              = 3
		floatVal            = 3.13
		strFloatVal         = "3.23"
		timeVal             = time.Now().In(schema.DefaultLocale)
		timeValParsed       = getDate(timeVal)
		strTimeVal          = "2006-01-01 11:10:12"
		strTimeValParsed, _ = time.Parse("2006-1-2 15:4:5", strTimeVal)
		strTimeValDate      = getDate(strTimeValParsed)
		objectIDVal         = bson.NewObjectId()
		sqlVal              = SQLInt(0)
		zeroTime            = time.Time{}
		defaultSQLDate      = SQLDate{zeroTime}
	)

	Convey("Subject: NewSQLValue", t, func() {

		Convey("Subject: SQLNull", func() {
			tests := []test{
				test{nil, schema.SQLBoolean, SQLNull},
				test{nil, schema.SQLDate, SQLNull},
				test{nil, schema.SQLDecimal128, SQLNull},
				test{nil, schema.SQLFloat, SQLNull},
				test{nil, schema.SQLInt, SQLNull},
				test{nil, schema.SQLInt64, SQLNull},
				test{nil, schema.SQLNumeric, SQLNull},
				test{nil, schema.SQLObjectID, SQLNull},
				test{nil, schema.SQLVarchar, SQLNull},
			}

			runTests(tests)

		})

		Convey("Subject: SQLValue", func() {
			tests := []test{
				test{sqlVal, schema.SQLBoolean, sqlVal},
				test{sqlVal, schema.SQLDate, sqlVal},
				test{sqlVal, schema.SQLDecimal128, sqlVal},
				test{sqlVal, schema.SQLFloat, sqlVal},
				test{sqlVal, schema.SQLInt, sqlVal},
				test{sqlVal, schema.SQLInt64, sqlVal},
				test{sqlVal, schema.SQLNumeric, sqlVal},
				test{sqlVal, schema.SQLObjectID, sqlVal},
				test{sqlVal, schema.SQLVarchar, sqlVal},
			}

			runTests(tests)

		})

		Convey("Subject: SQLBoolean", func() {
			tests := []test{
				test{false, schema.SQLBoolean, SQLFalse},
				test{true, schema.SQLBoolean, SQLTrue},
				test{floatVal, schema.SQLBoolean, SQLTrue},
				test{0.0, schema.SQLBoolean, SQLFalse},
				test{objectIDVal, schema.SQLBoolean, SQLTrue},
				test{intVal, schema.SQLBoolean, SQLTrue},
				test{0, schema.SQLBoolean, SQLFalse},
				test{strFloatVal, schema.SQLBoolean, SQLTrue},
				test{"0.000", schema.SQLBoolean, SQLFalse},
				test{"1.0", schema.SQLBoolean, SQLTrue},
				test{strTimeVal, schema.SQLBoolean, SQLFalse},
				test{timeVal, schema.SQLBoolean, SQLTrue},
			}

			runTests(tests)

		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				test{false, schema.SQLDate, defaultSQLDate},
				test{true, schema.SQLDate, defaultSQLDate},
				test{floatVal, schema.SQLDate, defaultSQLDate},
				test{0.0, schema.SQLDate, defaultSQLDate},
				test{objectIDVal, schema.SQLDate, SQLDate{objectIDVal.Time()}},
				test{intVal, schema.SQLDate, defaultSQLDate},
				test{0, schema.SQLDate, defaultSQLDate},
				test{strFloatVal, schema.SQLDate, defaultSQLDate},
				test{"0.000", schema.SQLDate, defaultSQLDate},
				test{"1.0", schema.SQLDate, defaultSQLDate},
				test{strTimeVal, schema.SQLDate, SQLDate{strTimeValDate}},
				test{timeVal, schema.SQLDate, SQLDate{timeValParsed}},
			}

			runTests(tests)

		})

		Convey("Subject: SQLDecimal128", func() {
			tests := []test{
				test{false, schema.SQLDecimal128, SQLDecimal128(decimal.New(0, 0))},
				test{true, schema.SQLDecimal128, SQLDecimal128(decimal.New(1, 0))},
				test{floatVal, schema.SQLDecimal128, SQLDecimal128(decimal.NewFromFloat(floatVal))},
				test{0.0, schema.SQLDecimal128, SQLDecimal128(decimal.New(0, 0))},
				test{objectIDVal, schema.SQLDecimal128, SQLDecimal128(decimal.New(0, 0))},
				test{intVal, schema.SQLDecimal128, SQLDecimal128(decimal.NewFromFloat(float64(intVal)))},
				test{0, schema.SQLDecimal128, SQLDecimal128(decimal.New(0, 0))},
				test{strFloatVal, schema.SQLDecimal128, SQLDecimal128(decimal.NewFromFloat(floatVal + .1))},
				test{"0.000", schema.SQLDecimal128, SQLDecimal128(decimal.New(0, 0))},
				test{"1.0", schema.SQLDecimal128, SQLDecimal128(decimal.New(1, 0))},
			}

			runTests(tests)

		})

		Convey("Subject: SQLFloat, SQLNumeric", func() {
			tests := []test{

				//
				// SQLFloat
				//
				test{false, schema.SQLFloat, SQLFloat(0.0)},
				test{true, schema.SQLFloat, SQLFloat(1.0)},
				test{floatVal, schema.SQLFloat, SQLFloat(floatVal)},
				test{0.0, schema.SQLFloat, SQLFloat(0.0)},
				test{intVal, schema.SQLFloat, SQLFloat(float64(intVal))},
				test{0, schema.SQLFloat, SQLFloat(0.0)},
				test{strFloatVal, schema.SQLFloat, SQLFloat(3.23)},
				test{"0.000", schema.SQLFloat, SQLFloat(0.0)},
				test{"1.0", schema.SQLFloat, SQLFloat(1.0)},

				//
				// SQLNumeric
				//
				test{false, schema.SQLNumeric, SQLFloat(0.0)},
				test{true, schema.SQLNumeric, SQLFloat(1.0)},
				test{floatVal, schema.SQLNumeric, SQLFloat(floatVal)},
				test{0.0, schema.SQLNumeric, SQLFloat(0.0)},
				test{intVal, schema.SQLNumeric, SQLFloat(float64(intVal))},
				test{0, schema.SQLNumeric, SQLFloat(0.0)},
				test{strFloatVal, schema.SQLNumeric, SQLFloat(3.23)},
				test{"0.000", schema.SQLNumeric, SQLFloat(0.0)},
				test{"1.0", schema.SQLNumeric, SQLFloat(1.0)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLInt, SQLInt64", func() {
			tests := []test{

				test{false, schema.SQLInt, SQLInt(0)},
				test{true, schema.SQLInt, SQLInt(1)},
				test{floatVal, schema.SQLInt, SQLInt(int64(floatVal))},
				test{0.0, schema.SQLInt, SQLInt(0)},
				test{intVal, schema.SQLInt, SQLInt(intVal)},
				test{0, schema.SQLInt, SQLInt(0)},
				test{strFloatVal, schema.SQLInt, SQLInt(3)},
				test{"0.000", schema.SQLInt, SQLInt(0)},
				test{"1.0", schema.SQLInt, SQLInt(1)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLObjectID", func() {
			tests := []test{
				test{false, schema.SQLObjectID, SQLObjectID("0")},
				test{true, schema.SQLObjectID, SQLObjectID("1")},
				test{floatVal, schema.SQLObjectID, SQLObjectID(strconv.FormatFloat(floatVal, 'f', -1, 64))},
				test{0.0, schema.SQLObjectID, SQLObjectID("0")},
				test{objectIDVal, schema.SQLObjectID, SQLObjectID(objectIDVal.Hex())},
				test{intVal, schema.SQLObjectID, SQLObjectID(strconv.FormatInt(int64(intVal), 10))},
				test{0, schema.SQLObjectID, SQLObjectID("0")},
				test{strFloatVal, schema.SQLObjectID, SQLObjectID(strFloatVal)},
				test{"0.000", schema.SQLObjectID, SQLObjectID("0.000")},
				test{"1.0", schema.SQLObjectID, SQLObjectID("1.0")},
				test{strTimeVal, schema.SQLObjectID, SQLObjectID(strTimeVal)},
				test{timeVal, schema.SQLObjectID, SQLObjectID(bson.NewObjectIdWithTime(timeVal).Hex())},
			}

			runTests(tests)

		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				test{false, schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{true, schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{floatVal, schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{0.0, schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{objectIDVal, schema.SQLTimestamp, SQLTimestamp{objectIDVal.Time()}},
				test{intVal, schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{0, schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{strFloatVal, schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{"0.000", schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{"1.0", schema.SQLTimestamp, SQLTimestamp{zeroTime}},
				test{strTimeVal, schema.SQLTimestamp, SQLTimestamp{strTimeValParsed}},
				test{timeVal, schema.SQLTimestamp, SQLTimestamp{timeVal}},
			}

			runTests(tests)

		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				test{false, schema.SQLVarchar, SQLVarchar("0")},
				test{true, schema.SQLVarchar, SQLVarchar("1")},
				test{floatVal, schema.SQLVarchar, SQLVarchar(strconv.FormatFloat(floatVal, 'f', -1, 64))},
				test{0.0, schema.SQLVarchar, SQLVarchar("0")},
				test{objectIDVal, schema.SQLVarchar, SQLVarchar(objectIDVal.Hex())},
				test{intVal, schema.SQLVarchar, SQLVarchar(strconv.FormatInt(int64(intVal), 10))},
				test{0, schema.SQLVarchar, SQLVarchar("0")},
				test{strFloatVal, schema.SQLVarchar, SQLVarchar(strFloatVal)},
				test{"0.000", schema.SQLVarchar, SQLVarchar("0.000")},
				test{"1.0", schema.SQLVarchar, SQLVarchar("1.0")},
				test{strTimeVal, schema.SQLVarchar, SQLVarchar(strTimeVal)},
				test{timeVal, schema.SQLVarchar, SQLVarchar(timeVal.String())},
			}

			runTests(tests)

		})

	})

}

func TestNewSQLValueFromSQLColumnExpr(t *testing.T) {

	Convey("When creating a SQLValue with no column type specified calling NewSQLValueFromSQLColumnExpr on a", t, func() {

		Convey("SQLValue should return the same object passed in", func() {
			v := SQLTrue
			newV, err := NewSQLValueFromSQLColumnExpr(v, schema.SQLBoolean, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, v)
		})

		Convey("nil value should return SQLNull", func() {
			v, err := NewSQLValueFromSQLColumnExpr(nil, schema.SQLNull, schema.MongoBool)
			So(err, ShouldBeNil)
			So(v, ShouldResemble, SQLNull)
		})

		Convey("bson object id should return its string value", func() {
			v := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err := NewSQLValueFromSQLColumnExpr(v, schema.SQLVarchar, schema.MongoObjectId)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v.Hex())
		})

		Convey("string objects should return the string value", func() {
			v := "56a10dd56ce28a89a8ed6edb"
			newV, err := NewSQLValueFromSQLColumnExpr(v, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("int objects should return the int value", func() {
			v1 := int(6)
			newV, err := NewSQLValueFromSQLColumnExpr(v1, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v1)

			v2 := int32(6)
			newV, err = NewSQLValueFromSQLColumnExpr(v2, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v2)

			v3 := uint32(6)
			newV, err = NewSQLValueFromSQLColumnExpr(v3, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v3)
		})

		Convey("float objects should return the float value", func() {
			v := float64(6.3)
			newV, err := NewSQLValueFromSQLColumnExpr(v, schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("time objects should return the appropriate value", func() {
			v := time.Date(2014, time.December, 31, 0, 0, 0, 0, schema.DefaultLocale)
			newV, err := NewSQLValueFromSQLColumnExpr(v, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v})

			v = time.Date(2014, time.December, 31, 10, 0, 0, 0, schema.DefaultLocale)
			newV, err = NewSQLValueFromSQLColumnExpr(v, schema.SQLTimestamp, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlTimestamp, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTimestamp, ShouldResemble, SQLTimestamp{v})
		})
	})

	Convey("When creating a SQLValue with a column type specified calling NewSQLValueFromSQLColumnExpr on a", t, func() {

		Convey("a SQLVarchar/SQLVarchar column type should attempt to coerce to the SQLVarchar type", func() {

			t := schema.SQLVarchar

			newV, err := NewSQLValueFromSQLColumnExpr(t, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar(t))

			newV, err = NewSQLValueFromSQLColumnExpr(6, schema.SQLVarchar, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6"))

			newV, err = NewSQLValueFromSQLColumnExpr(6.6, schema.SQLVarchar, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6.6"))

			newV, err = NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLVarchar, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6"))

			_id := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err = NewSQLValueFromSQLColumnExpr(_id, schema.SQLVarchar, schema.MongoObjectId)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLObjectID(_id.Hex()))

		})

		Convey("a SQLInt column type should attempt to coerce to the SQLInt type", func() {

			newV, err := NewSQLValueFromSQLColumnExpr(true, schema.SQLInt, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(1))

			newV, err = NewSQLValueFromSQLColumnExpr(int(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLInt, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

		})

		Convey("a SQLFloat column type should attempt to coerce to the SQLFloat type", func() {

			newV, err := NewSQLValueFromSQLColumnExpr(true, schema.SQLFloat, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(1))

			newV, err = NewSQLValueFromSQLColumnExpr(int(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLFloat, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6.6))

		})

		Convey("a SQLDate column type should attempt to coerce to the SQLDate type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 0, 0, 0, 0, schema.DefaultLocale)
			v2 := time.Date(2014, time.May, 11, 10, 32, 12, 0, schema.DefaultLocale)

			newV, err := NewSQLValueFromSQLColumnExpr(v1, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			newV, err = NewSQLValueFromSQLColumnExpr(v2, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok = newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			// String type
			dates := []string{"2014-05-11", "2014-05-11 15:04:05", "2014-05-11 15:04:05.233"}

			for _, d := range dates {

				newV, err := NewSQLValueFromSQLColumnExpr(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldBeNil)

				sqlDate, ok := newV.(SQLDate)
				So(ok, ShouldBeTrue)
				So(sqlDate, ShouldResemble, SQLDate{v1})

			}

			// invalid dates and those outside valid range
			// should return the default date
			dates = []string{"2014-12-44-44", "999-1-1", "10000-1-1"}

			for _, d := range dates {
				newV, err = NewSQLValueFromSQLColumnExpr(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldBeNil)

				_, ok := newV.(SQLFloat)
				So(ok, ShouldBeTrue)
			}
		})

		Convey("a SQLTimestamp column type should attempt to coerce to the SQLTimestamp type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)

			newV, err := NewSQLValueFromSQLColumnExpr(v1, schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// String type
			newV, err = NewSQLValueFromSQLColumnExpr("2014-05-11 15:04:05.000", schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok = newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// invalid dates should return the default date
			dates := []string{"2044-12-40", "1966-15-1", "43223-3223"}

			for _, d := range dates {
				newV, err = NewSQLValueFromSQLColumnExpr(d, schema.SQLTimestamp, schema.MongoNone)
				So(err, ShouldBeNil)
				_, ok := newV.(SQLFloat)
				So(ok, ShouldBeTrue)
			}
		})
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
	exprA := NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt)
	exprB := NewSQLColumnExpr(1, "bar", "b", schema.SQLInt, schema.MongoInt)
	exprG := NewSQLColumnExpr(1, "bar", "g", schema.SQLTimestamp, schema.MongoDate)

	Convey("Subject: reconcileSQLExpr", t, func() {

		tests := []test{
			test{"a = 3", exprA, SQLInt(3)},
			test{"g - '2010-01-01'", &SQLConvertExpr{exprG, schema.SQLInt}, &SQLConvertExpr{SQLVarchar("2010-01-01"), schema.SQLInt}},
			test{"a in (3)", exprA, SQLInt(3)},
			test{"a in (2,3)", exprA, &SQLTupleExpr{[]SQLExpr{SQLInt(2), SQLInt(3)}}},
			test{"(a) in (3)", exprA, SQLInt(3)},
			test{"(a,b) in (2,3)", &SQLTupleExpr{[]SQLExpr{exprA, exprB}}, &SQLTupleExpr{[]SQLExpr{SQLInt(2), SQLInt(3)}}},
			test{"g > '2010-01-01'", exprG, exprConv},
			test{"a and b", exprA, exprB},
			test{"a / b", exprA, exprB},
			test{"'2010-01-01' and g", exprConv, exprG},
			test{"g in ('2010-01-01',current_timestamp())", exprG, &SQLTupleExpr{[]SQLExpr{exprConv, exprTime}}},
			test{"g in ('2010-01-01',current_timestamp)", exprG, &SQLTupleExpr{[]SQLExpr{exprConv, exprTime}}},
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
				n, err := optimizeEvaluations(createTestEvalCtx(), e)
				So(err, ShouldBeNil)
				e = n.(SQLExpr)
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
			// This looks weird. It's because json.Marshal doesn't know how to deal with bson.DocElem which we used
			// because order matters for $regex and $options. However, the go driver does know and will handle this correctly.
			test{"a LIKE '%un%'", `{"a":[{"Name":"$regex","Value":"^.*un.*$"},{"Name":"$options","Value":"i"}]}`},
			test{"a REGEXP 'abc'", `{"a":{"$regex":{"Pattern":"abc","Options":""}}}`},
			test{"a REGEXP '(.* )?'", `{"a":{"$regex":{"Pattern":"(.* )?","Options":""}}}`},
			test{"a REGEXP 'a' OR 'b'", `{"a":{"$regex":{"Pattern":"a","Options":""}}}`},
			test{"a NOT REGEXP 'a' OR 'b'", `{"a":{"$nin":[{"Pattern":"a","Options":""}]}}`},
			test{"a NOT REGEXP 'abc'", `{"a":{"$nin":[{"Pattern":"abc","Options":""}]}}`},
			test{"a NOT REGEXP '(.* )?'", `{"a":{"$nin":[{"Pattern":"(.* )?","Options":""}]}}`},
		}

		runTests(tests)

		partialTests := []partialTest{
			partialTest{"a = 3 AND a < b", `{"a":3}`, "a < b", &SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLInt, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}},
			partialTest{"a = 3 AND a < b AND b = 4", `{"$and":[{"a":3},{"b":4}]}`, "a < b", &SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLInt, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}},
			partialTest{"a < b AND a = 3", `{"a":3}`, "a < b", &SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLInt, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}},
			partialTest{"NOT (a = 3 AND a < b)", `{"a":{"$ne":3}}`, "NOT a < b", &SQLNotExpr{&SQLLessThanExpr{NewSQLColumnExpr(1, tableTwoName, "a", schema.SQLInt, schema.MongoInt), NewSQLColumnExpr(1, tableTwoName, "b", schema.SQLInt, schema.MongoInt)}}},
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
			test{"concat(a, null)", `{"$concat":["$a",{"$literal":null}]}`},
			test{"concat(a, '')", `{"$concat":["$a",{"$literal":""}]}`},
			test{"concat_ws(',', a)", `{"$concat":[{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},{"$literal":""},"$a"]}]}`},
			test{"concat_ws(',', a, null)", `{"$concat":[{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},{"$literal":""},"$a"]},{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},{"$literal":""},{"$literal":","}]},{"$cond":[{"$eq":[{"$ifNull":[{"$literal":null},null]},null]},{"$literal":""},{"$literal":null}]}]}`},
			test{"dayname(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$arrayElemAt":[["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"],{"$subtract":[{"$dayOfWeek":"$a"},1]}]}]}`},
			test{"dayofmonth(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$dayOfMonth":"$a"}]}`},
			test{"dayofweek(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$dayOfWeek":"$a"}]}`},
			test{"dayofyear(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$dayOfYear":"$a"}]}`},
			test{"exp(a)", `{"$exp":"$a"}`},
			test{"extract(year from a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$year":"$a"}]}`},
			test{"floor(a)", `{"$floor":"$a"}`},
			test{"hour(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$hour":"$a"}]}`},
			test{"if(a, 2, 3)", `{"$cond":[{"$or":[{"$eq":[{"$ifNull":["$a",null]},null]},{"$eq":["$a",0]},{"$eq":["$a",false]}]},{"$literal":3},{"$literal":2}]}`},
			test{"ifnull(a, 1)", `{"$ifNull":["$a",{"$literal":1}]}`},
			test{"isnull(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},1,0]}`},
			test{"left(a, 2)", `{"$cond":[{"$or":[{"$eq":[{"$ifNull":["$a",null]},null]},{"$eq":[{"$ifNull":[{"$literal":2},null]},null]}]},null,{"$substr":["$a",0,{"$literal":2}]}]}`},
			test{"left('abcde', 0)", `{"$cond":[{"$or":[{"$eq":[{"$ifNull":[{"$literal":"abcde"},null]},null]},{"$eq":[{"$ifNull":[{"$literal":0},null]},null]}]},null,{"$substr":[{"$literal":"abcde"},0,{"$literal":0}]}]}`},
			test{"lcase(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$toLower":"$a"}]}`},
			test{"lower(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$toLower":"$a"}]}`},
			test{"log10(a)", `{"$cond":[{"$gt":["$a",0]},{"$log10":"$a"},{"$literal":null}]}`},
			test{"minute(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$minute":"$a"}]}`},
			test{"mod(a, 10)", `{"$mod":["$a",{"$literal":10}]}`},
			test{"month(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$month":"$a"}]}`},
			test{"monthname(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$arrayElemAt":[["January","February","March","April","May","June","July","August","September","October","November","December"],{"$subtract":[{"$month":"$a"},1]}]}]}`},
			test{"nullif(a, 1)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$cond":[{"$eq":["$a",{"$literal":1}]},null,"$a"]}]}`},
			test{"power(a, 10)", `{"$pow":["$a",{"$literal":10}]}`},
			test{"quarter(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$arrayElemAt":[[1,1,1,2,2,2,3,3,3,4,4,4],{"$subtract":[{"$month":"$a"},1]}]}]}`},
			test{"round(a, 5)", `{"$divide":[{"$cond":[{"$gte":["$a",0]},{"$floor":{"$add":[{"$multiply":["$a",100000]},0.5]}},{"$ceil":{"$subtract":[{"$multiply":["$a",100000]},0.5]}}]},100000]}`},
			test{"round(a, -5)", `{"$literal":0}`},

			test{"second(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$second":"$a"}]}`},
			test{"sqrt(a)", `{"$cond":[{"$gte":["$a",0]},{"$sqrt":"$a"},null]}`},
			test{"substring(a, 2)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$substr":["$a",1,-1]}]}`},
			test{"substring(a, 2, 4)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$substr":["$a",1,{"$literal":4}]}]}`},
			test{"substr(a, 2)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$substr":["$a",1,-1]}]}`},
			test{"substr(a, 2, 4)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$substr":["$a",1,{"$literal":4}]}]}`},
			test{"week(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$week":"$a"}]}`},
			test{"week(a, 0)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$week":"$a"}]}`},
			test{"weekday(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$mod":[{"$add":[{"$mod":[{"$subtract":[{"$dayOfWeek":"$a"},2]},7]},7]},7]}]}`},

			test{"ucase(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$toUpper":"$a"}]}`},
			test{"upper(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$toUpper":"$a"}]}`},
			test{"year(a)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$year":"$a"}]}`},

			test{"count(*)", `{"$size":{"$literal":"*"}}`},
			test{"count(a + b)", `{"$sum":{"$map":{"as":"i","in":{"$cond":[{"$eq":[{"$ifNull":["$$i",null]},null]},0,1]},"input":{"$add":["$a","$b"]}}}}`},
			test{"min(a + 4)", `{"$min":{"$add":["$a",{"$literal":4}]}}`},
			test{"sum(a * b)", `{"$sum":{"$multiply":["$a","$b"]}}`},
			test{"sum(a)", `{"$sum":"$a"}`},
			test{"sum(a < 1)", `{"$sum":{"$cond":[{"$or":[{"$eq":[{"$ifNull":["$a",null]},null]},{"$eq":[{"$ifNull":[{"$literal":1},null]},null]}]},null,{"$lt":["$a",{"$literal":1}]}]}}`},
			test{"std(a)", `{"$stdDevPop":"$a"}`},
			test{"stddev(a)", `{"$stdDevPop":"$a"}`},
			test{"stddev_samp(a)", `{"$stdDevSamp":"$a"}`},
			test{"a in (2,3,5)", `{"$cond":[{"$eq":[{"$ifNull":["$a",null]},null]},null,{"$cond":[{"$gt":[{"$size":{"$filter":{"as":"item","cond":{"$eq":["$$item","$a"]},"input":[{"$literal":2},{"$literal":3},{"$literal":5}]}}},{"$literal":0}]},true,{"$cond":[{"$eq":[false,true]},null,false]}]}]}`},
			test{"case when a > 1 then 'gt' else 'lt' end", `{"$cond":[{"$cond":[{"$or":[{"$eq":[{"$ifNull":["$a",null]},null]},{"$eq":[{"$ifNull":[{"$literal":1},null]},null]}]},null,{"$gt":["$a",{"$literal":1}]}]},{"$literal":"gt"},{"$literal":"lt"}]}`},
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
			sqlValueTest{SQLNull, `{"$literal":null}`},
			sqlValueTest{SQLDate{fakeTime}, fmt.Sprintf(`{"$literal":"%v"}`, fakeTime.Format(schema.DateFormat))},
			sqlValueTest{SQLTimestamp{fakeTime}, fmt.Sprintf(`{"$literal":"%v"}`, fakeTime.Format(schema.TimestampFormat))},
		}

		runSQLValueTests(sqlValueTests)
	})
}

func TestCompareTo(t *testing.T) {

	var (
		diff        = time.Duration(969 * time.Hour)
		sameDayDiff = time.Duration(1)
		now         = time.Now()
		oid1        = bson.NewObjectId().Hex()
		oid2        = bson.NewObjectId().Hex()
		oid3        = bson.NewObjectId().Hex()
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
				{SQLInt(1), SQLBool(false), 1},
				{SQLInt(1), SQLBool(true), 0},
				{SQLInt(1), SQLNull, 1},
				{SQLInt(1), SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLInt(1), SQLVarchar("bac"), 1},
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
				{SQLUint32(1), SQLBool(false), 1},
				{SQLUint32(1), SQLBool(true), 0},
				{SQLUint32(1), SQLNull, 1},
				{SQLUint32(1), SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLUint32(1), SQLVarchar("bac"), 1},
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
				{SQLFloat(0.1), SQLBool(false), 1},
				{SQLFloat(0.1), SQLBool(true), -1},
				{SQLFloat(0.1), SQLNull, 1},
				{SQLFloat(0.1), SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLFloat(0.1), SQLVarchar("bac"), 1},
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
				{SQLBool(true), SQLInt(1), 0},
				{SQLBool(true), SQLInt(2), -1},
				{SQLBool(true), SQLUint32(1), 0},
				{SQLBool(true), SQLFloat(1), 0},
				{SQLBool(true), SQLBool(false), 1},
				{SQLBool(true), SQLBool(true), 0},
				{SQLBool(true), SQLNull, 1},
				{SQLBool(true), SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{SQLBool(true), SQLVarchar("bac"), 1},
				{SQLBool(true), &SQLValues{[]SQLValue{SQLInt(1)}}, 0},
				{SQLBool(true), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLBool(true), SQLDate{now}, -1},
				{SQLBool(true), SQLTimestamp{now}, -1},
				{SQLBool(false), SQLInt(0), 0},
				{SQLBool(false), SQLInt(1), -1},
				{SQLBool(false), SQLInt(2), -1},
				{SQLBool(false), SQLUint32(1), -1},
				{SQLBool(false), SQLFloat(1), -1},
				{SQLBool(false), SQLBool(false), 0},
				{SQLBool(false), SQLBool(true), -1},
				{SQLBool(false), SQLNull, 1},
				{SQLBool(false), SQLObjectID("56e0750e1d857aea925a4ba1"), 0},
				{SQLBool(false), SQLVarchar("bac"), 0},
				{SQLBool(false), &SQLValues{[]SQLValue{SQLInt(1)}}, -1},
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
				{SQLTimestamp{now}, SQLDate{now}, 0},
				{SQLTimestamp{now.Add(sameDayDiff)}, SQLDate{now}, 1},
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
				{SQLVarchar("bac"), SQLInt(0), 0},
				{SQLVarchar("bac"), SQLInt(1), -1},
				{SQLVarchar("bac"), SQLInt(2), -1},
				{SQLVarchar("bac"), SQLUint32(1), -1},
				{SQLVarchar("bac"), SQLFloat(1), -1},
				{SQLVarchar("bac"), SQLBool(false), 0},
				{SQLVarchar("bac"), SQLObjectID("56e0750e1d857aea925a4ba1"), 0},
				{SQLVarchar("bac"), SQLVarchar("cba"), -1},
				{SQLVarchar("bac"), SQLVarchar("bac"), 0},
				{SQLVarchar("bac"), SQLVarchar("abc"), 1},
				{SQLVarchar("bac"), &SQLValues{[]SQLValue{SQLInt(1)}}, -1},
				{SQLVarchar("bac"), &SQLValues{[]SQLValue{SQLNone}}, 1},
				{SQLVarchar("bac"), &SQLValues{[]SQLValue{SQLVarchar("bac")}}, 0},
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
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLBool(false), 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{&SQLValues{[]SQLValue{SQLInt(1)}}, SQLVarchar("abc"), 1},
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
				{SQLObjectID(oid2), SQLInt(0), 0},
				{SQLObjectID(oid2), SQLUint32(1), -1},
				{SQLObjectID(oid2), SQLFloat(1), -1},
				{SQLObjectID(oid2), SQLVarchar("cba"), 0},
				{SQLObjectID(oid2), SQLBool(false), 0},
				{SQLObjectID(oid2), SQLBool(true), -1},
				{SQLObjectID(oid2), &SQLValues{[]SQLValue{SQLInt(1)}}, -1},
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

func TestIsTruthyIsFalsy(t *testing.T) {

	Convey("isTruthy, isFalsy", t, func() {
		d, err := time.Parse("2006-01-02", "2003-01-02")
		So(err, ShouldBeNil)
		t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
		So(err, ShouldBeNil)

		Convey("Subject: isTruthy", func() {
			truthy := isTruthy(SQLTimestamp{t})
			So(truthy, ShouldBeTrue)

			truthy = isTruthy(SQLDate{d})
			So(truthy, ShouldBeTrue)

			truthy = isTruthy(SQLInt(0))
			So(truthy, ShouldBeFalse)

			truthy = isTruthy(SQLInt(1))
			So(truthy, ShouldBeTrue)

			truthy = isTruthy(SQLVarchar("dsf"))
			So(truthy, ShouldBeFalse)

			truthy = isTruthy(SQLVarchar("16"))
			So(truthy, ShouldBeTrue)
		})

		Convey("Subject: isFalsy", func() {
			truthy := isFalsy(SQLTimestamp{t})
			So(truthy, ShouldBeFalse)

			truthy = isFalsy(SQLDate{d})
			So(truthy, ShouldBeFalse)

			truthy = isFalsy(SQLInt(0))
			So(truthy, ShouldBeTrue)

			truthy = isFalsy(SQLInt(1))
			So(truthy, ShouldBeFalse)

			truthy = isFalsy(SQLVarchar("dsf"))
			So(truthy, ShouldBeTrue)

			truthy = isFalsy(SQLVarchar("16"))
			So(truthy, ShouldBeFalse)
		})
	})
}
