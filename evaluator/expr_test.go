package evaluator_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
)

func TestEvaluates(t *testing.T) {

	type test struct {
		sql    string
		result evaluator.SQLExpr
	}

	runTests := func(ctx *evaluator.EvalCtx, tests []test) {
		schema, err := schema.New(testSchema3, &lgr)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be %v", t.sql, t.result), func() {
				subject, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				result, err := subject.Evaluate(ctx)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	type typeTest struct {
		sql    string
		result schema.SQLType
	}

	runTypeTests := func(ctx *evaluator.EvalCtx, tests []typeTest) {
		schema, err := schema.New(testSchema3, &lgr)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be %v", t.sql, t.result), func() {
				subject, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				result := subject.Type()
				So(result, ShouldResemble, t.result)
			})
		}
	}

	execCtx := createTestExecutionCtx(nil)

	Convey("Subject: Evaluates", t, func() {
		evalCtx := evaluator.NewEvalCtx(execCtx, collation.Default, &evaluator.Row{evaluator.Values{
			{1, "test", "bar", "a", evaluator.SQLInt(123)},
			{1, "test", "bar", "b", evaluator.SQLInt(456)},
			{1, "test", "bar", "c", evaluator.SQLNull},
		}})

		Convey("Subject: SQLAddExpr", func() {
			tests := []test{
				{"0 + 0", evaluator.SQLInt(0)},
				{"-1 + 1", evaluator.SQLInt(0)},
				{"10 + 32", evaluator.SQLInt(42)},
				{"-10 + -32", evaluator.SQLInt(-42)},
				{"true + true", evaluator.SQLFloat(2)},
				{"true + true + false", evaluator.SQLFloat(2)},
				{"false + true + true", evaluator.SQLFloat(2)},
				{"true - '-1'", evaluator.SQLFloat(2)},
				{"true + '0'", evaluator.SQLFloat(1)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLAggFunctionExpr", func() {
			var t1, t2 time.Time

			t1 = time.Now()
			t2 = t1.Add(time.Hour)

			aggCtx := evaluator.NewEvalCtx(execCtx, collation.Default,
				&evaluator.Row{evaluator.Values{
					{1, "test", "bar", "a", evaluator.SQLNull},
					{1, "test", "bar", "b", evaluator.SQLInt(3)},
					{1, "test", "bar", "c", evaluator.SQLNull},
					{1, "test", "bar", "g", evaluator.SQLDate{t1}},
				}},
				&evaluator.Row{evaluator.Values{
					{1, "test", "bar", "a", evaluator.SQLInt(3)},
					{1, "test", "bar", "b", evaluator.SQLNull},
					{1, "test", "bar", "c", evaluator.SQLNull},
					{1, "test", "bar", "g", evaluator.SQLDate{t2}},
				}},
				&evaluator.Row{evaluator.Values{
					{1, "test", "bar", "a", evaluator.SQLInt(5)},
					{1, "test", "bar", "b", evaluator.SQLInt(6)},
					{1, "test", "bar", "c", evaluator.SQLNull},
					{1, "test", "bar", "g", evaluator.SQLNull},
				}},
			)

			Convey("Subject: AVG", func() {
				tests := []test{
					{"AVG(NULL)", evaluator.SQLNull},
					{"AVG(a)", evaluator.SQLFloat(4)},
					{"AVG(b)", evaluator.SQLFloat(4.5)},
					{"AVG(c)", evaluator.SQLNull},
					{"AVG('a')", evaluator.SQLFloat(0)},
					{"AVG(-20)", evaluator.SQLFloat(-20)},
					{"AVG(20)", evaluator.SQLFloat(20)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: COUNT", func() {
				tests := []test{
					{"COUNT(NULL)", evaluator.SQLInt(0)},
					{"COUNT(a)", evaluator.SQLInt(2)},
					{"COUNT(b)", evaluator.SQLInt(2)},
					{"COUNT(c)", evaluator.SQLInt(0)},
					{"COUNT(g)", evaluator.SQLInt(2)},
					{"COUNT('a')", evaluator.SQLInt(3)},
					{"COUNT(-20)", evaluator.SQLInt(3)},
					{"COUNT(20)", evaluator.SQLInt(3)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: MIN", func() {
				tests := []test{
					{"MIN(NULL)", evaluator.SQLNull},
					{"MIN(a)", evaluator.SQLInt(3)},
					{"MIN(b)", evaluator.SQLInt(3)},
					{"MIN(c)", evaluator.SQLNull},
					{"MIN('a')", evaluator.SQLVarchar("a")},
					{"MIN(-20)", evaluator.SQLInt(-20)},
					{"MIN(20)", evaluator.SQLInt(20)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: MAX", func() {
				tests := []test{
					{"MAX(NULL)", evaluator.SQLNull},
					{"MAX(a)", evaluator.SQLInt(5)},
					{"MAX(b)", evaluator.SQLInt(6)},
					{"MAX(c)", evaluator.SQLNull},
					{"MAX('a')", evaluator.SQLVarchar("a")},
					{"MAX(-20)", evaluator.SQLInt(-20)},
					{"MAX(20)", evaluator.SQLInt(20)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: SLEEP", func() {
				tests := []test{
					{"SLEEP(1)", evaluator.SQLInt(0)},
					{"SLEEP(1.5)", evaluator.SQLInt(0)},
					{"SLEEP(0)", evaluator.SQLInt(0)},
				}
				runTests(aggCtx, tests)

				Convey("Should error with negative input", func() {
					subject, err := evaluator.NewSQLScalarFunctionExpr(
						"sleep", []evaluator.SQLExpr{evaluator.SQLInt(-1)})
					So(err, ShouldBeNil)
					_, err = subject.Evaluate(evalCtx)
					So(err, ShouldNotBeNil)
				})

				Convey("Should error with null input", func() {
					subject, err := evaluator.NewSQLScalarFunctionExpr(
						"sleep",
						[]evaluator.SQLExpr{evaluator.SQLNull},
					)
					So(err, ShouldBeNil)
					_, err = subject.Evaluate(evalCtx)
					So(err, ShouldNotBeNil)
				})
			})

			Convey("Subject: SUM", func() {
				tests := []test{
					{"SUM(NULL)", evaluator.SQLNull},
					{"SUM(a)", evaluator.SQLFloat(8)},
					{"SUM(b)", evaluator.SQLFloat(9)},
					{"SUM(c)", evaluator.SQLNull},
					{"SUM('a')", evaluator.SQLFloat(0)},
					{"SUM(-20)", evaluator.SQLFloat(-60)},
					{"SUM(20)", evaluator.SQLFloat(60)},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: STDDEV_POP", func() {
				tests := []test{
					{"STD(NULL)", evaluator.SQLNull},
					{"STDDEV(a)", evaluator.SQLFloat(1)},
					{"STDDEV_POP(b)", evaluator.SQLFloat(1.5)},
					{"STD(c)", evaluator.SQLNull},
				}
				runTests(aggCtx, tests)
			})

			Convey("Subject: STDDEV_SAMP", func() {
				tests := []test{
					{"STDDEV_SAMP(NULL)", evaluator.SQLNull},
					{"STDDEV_SAMP(a)", evaluator.SQLFloat(1.4142135623730951)},
					{"STDDEV_SAMP(b)", evaluator.SQLFloat(2.1213203435596424)},
					{"STDDEV_SAMP(c)", evaluator.SQLNull},
				}
				runTests(aggCtx, tests)
			})

		})

		Convey("Subject: SQLAndExpr", func() {
			tests := []test{
				{"1 AND 1", evaluator.SQLTrue},
				{"1 AND 0", evaluator.SQLFalse},
				{"0 AND 1", evaluator.SQLFalse},
				{"0 AND 0", evaluator.SQLFalse},
				{"1 && 1", evaluator.SQLTrue},
				{"1 && 0", evaluator.SQLFalse},
				{"0 && 1", evaluator.SQLFalse},
				{"0 && 0", evaluator.SQLFalse},
				{"NULL && 0", evaluator.SQLFalse},
				{"NULL && 1", evaluator.SQLNull},
				{"NULL && NULL", evaluator.SQLNull},
				{"true AND true", evaluator.SQLTrue},
				{"true AND false", evaluator.SQLFalse},
				{"false AND true", evaluator.SQLFalse},
				{"false AND false", evaluator.SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLAssignmentExpr", func() {
			e := evaluator.NewSQLAssignmentExpr(
				evaluator.NewSQLVariableExpr(
					"test",
					variable.UserKind,
					variable.SessionScope,
					schema.SQLNone,
				),
				evaluator.NewSQLAddExpr(
					evaluator.SQLInt(1),
					evaluator.SQLInt(3),
				),
			)

			result, err := e.Evaluate(evalCtx)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, evaluator.SQLInt(4))
		})

		Convey("Subject: SQLBenchmarkExpr", func() {
			tests := []test{
				{"BENCHMARK(10, 1)", evaluator.SQLInt(0)},
				{"BENCHMARK(0, 10)", evaluator.SQLInt(0)},
				{"BENCHMARK(NULL, 0)", evaluator.SQLInt(0)},
			}
			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLDateTimeArithmetic", func() {

			Convey("Subject: Add", func() {
				tests := []test{
					{"DATE '2014-04-13' + 0", evaluator.SQLInt(20140413)},
					{"DATE '2014-04-13' + 2", evaluator.SQLInt(20140415)},
					{"TIME '11:04:13' + 0", evaluator.SQLDecimal128(decimal.NewFromFloat(110413))},
					{"TIME '11:04:13' + 2", evaluator.SQLDecimal128(decimal.NewFromFloat(110415))},
					{"TIME '11:04:13' + '2'", evaluator.SQLDecimal128(decimal.NewFromFloat(110415))},
					{"'2' + TIME '11:04:13'", evaluator.SQLDecimal128(decimal.NewFromFloat(110415))},
					{"TIMESTAMP '2014-04-13 11:04:13' + 0", evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110413))},
					{"TIMESTAMP '2014-04-13 11:04:13' + 2", evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110415))},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Subtract", func() {
				tests := []test{
					{"DATE '2014-04-13' - 0", evaluator.SQLInt(20140413)},
					{"DATE '2014-04-13' - 2", evaluator.SQLInt(20140411)},
					{"TIME '11:04:13' - 0", evaluator.SQLDecimal128(decimal.NewFromFloat(110413))},
					{"TIME '11:04:13' - 2", evaluator.SQLDecimal128(decimal.NewFromFloat(110411))},
					{"TIME '11:04:13' - '2'", evaluator.SQLDecimal128(decimal.NewFromFloat(110411))},
					{"TIMESTAMP '2014-04-13 11:04:13' - 0", evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110413))},
					{"TIMESTAMP '2014-04-13 11:04:13' - 2", evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110411))},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Multiply", func() {
				tests := []test{
					{"DATE '2014-04-13' * 0", evaluator.SQLInt(0)},
					{"DATE '2014-04-13' * 2", evaluator.SQLInt(40280826)},
					{"TIME '11:04:13' * 0", evaluator.SQLDecimal128(decimal.NewFromFloat(0))},
					{"TIME '11:04:13' * 2", evaluator.SQLDecimal128(decimal.NewFromFloat(220826))},
					{"TIME '11:04:13' * '2'", evaluator.SQLDecimal128(decimal.NewFromFloat(220826))},
					{"TIMESTAMP '2014-04-13 11:04:13' * 0", evaluator.SQLDecimal128(decimal.NewFromFloat(0))},
					{"TIMESTAMP '2014-04-13 11:04:13' * 2", evaluator.SQLDecimal128(decimal.NewFromFloat(40280826220826))},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Divide", func() {
				tests := []test{
					{"1.2 / 0.2", evaluator.SQLDecimal128(decimal.New(600000, -5))},
					{"1.2 / 0.23", evaluator.SQLDecimal128(decimal.New(521739, -5))},
					{"DATE '2014-04-13' / 0", evaluator.SQLNull},
					{"DATE '2014-04-13' / 2", evaluator.SQLFloat(10070206.5)},
					{"TIME '11:04:13' / 0", evaluator.SQLNull},
					{"TIME '11:04:13' / 2", evaluator.SQLDecimal128(decimal.New(552065000, -4))},
					{"TIME '11:04:13' / '2'", evaluator.SQLDecimal128(decimal.New(552065000, -4))},
					{"TIMESTAMP '2014-04-13 11:04:13' / 0", evaluator.SQLNull},
					{"TIMESTAMP '2014-04-13 11:04:13' / 2", evaluator.SQLDecimal128(decimal.New(100702065552065000, -4))},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Less Than", func() {
				tests := []test{
					{"DATE '2014-04-13' > 0", evaluator.SQLTrue},
					{"DATE '2014-04-13' > DATE '2014-04-14'", evaluator.SQLFalse},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Greater Than", func() {
				tests := []test{
					{"DATE '2014-04-13' > 0", evaluator.SQLTrue},
					{"DATE '2014-04-13' > DATE '2014-04-14'", evaluator.SQLFalse},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: Equal", func() {
				tests := []test{
					{"DATE '2014-04-13' = '0'", evaluator.SQLFalse},
					{"DATE '2014-04-13' = DATE '2014-04-13'", evaluator.SQLTrue},
				}
				runTests(evalCtx, tests)
			})
		})

		Convey("Subject: SQLCaseExpr", func() {
			tests := []test{
				{"CASE 3 WHEN 3 THEN 'three' WHEN 1 THEN 'one' ELSE 'else' END", evaluator.SQLVarchar("three")},
				{"CASE WHEN 5 > 3 THEN 'true' else 'false' END", evaluator.SQLVarchar("true")},
				{"CASE WHEN a = 123 THEN 'yes' else 'no' END", evaluator.SQLVarchar("yes")},
				{"CASE WHEN a = 245 THEN 'yes' END", evaluator.SQLNull},
			}
			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLDateTimeLiterals", func() {

			Convey("Subject: DATE", func() {
				dateTime, _ := time.Parse("2006-01-02", "2014-04-13")
				tests := []test{
					{"DATE '2014-04-13'", evaluator.SQLDate{Time: dateTime}},
					{"{d '2014-04-13'}", evaluator.SQLDate{Time: dateTime}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TIME", func() {
				dateTime, _ := time.Parse("15:04:05", "11:49:36")
				tests := []test{
					{"TIME '11:49:36'", evaluator.SQLTimestamp{Time: dateTime}},
					{"{t '11:49:36'}", evaluator.SQLTimestamp{Time: dateTime}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TIMESTAMP", func() {
				dateTime, _ := time.Parse("2006-01-02 15:04:05.999999999", "1997-01-31 09:26:50.124")
				tests := []test{
					{"TIMESTAMP '1997-01-31 09:26:50.124'", evaluator.SQLTimestamp{Time: dateTime}},
					{"{ts '1997-01-31 09:26:50.124'}", evaluator.SQLTimestamp{Time: dateTime}},
				}
				runTests(evalCtx, tests)
			})

		})

		Convey("Subject: SQLDivideExpr", func() {
			tests := []test{
				{"-1 / 1", evaluator.SQLFloat(-1)},
				{"100 / 10", evaluator.SQLFloat(10)},
				{"-10 / 10", evaluator.SQLFloat(-1)},
			}

			runTests(evalCtx, tests)

			Convey("The result should be SQLNull when dividing by zero", func() {
				subject := evaluator.NewSQLDivideExpr(
					evaluator.SQLInt(10),
					evaluator.SQLInt(0),
				)
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, evaluator.SQLNull)
			})
		})

		Convey("Subject: SQLEqualsExpr", func() {
			tests := []test{
				{"0 = 0", evaluator.SQLTrue},
				{"-1 = 1", evaluator.SQLFalse},
				{"10 = 10", evaluator.SQLTrue},
				{"-10 = -10", evaluator.SQLTrue},
				{"false = '0'", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		SkipConvey("Subject: SQLExistsExpr", func() {
		})

		Convey("Subject: SQLColumnExpr", func() {
			Convey("Should return the value of the field when it exists", func() {
				subject := evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt)
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, evaluator.SQLInt(123))
			})

			Convey("Should return nil when the field is null", func() {
				subject := evaluator.NewSQLColumnExpr(1, "test", "bar", "c", schema.SQLInt, schema.MongoInt)
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, evaluator.SQLNull)
			})

			Convey("Should return nil when the field doesn't exists", func() {
				subject := evaluator.NewSQLColumnExpr(1, "test", "bar", "no_existy", schema.SQLInt, schema.MongoInt)
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, evaluator.SQLNull)
			})
		})

		Convey("Subject: SQLGreaterThanExpr", func() {
			tests := []test{
				{"0 > 0", evaluator.SQLFalse},
				{"-1 > 1", evaluator.SQLFalse},
				{"1 > -1", evaluator.SQLTrue},
				{"11 > 10", evaluator.SQLTrue},
				{"true > '-1'", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLGreaterThanOrEqualExpr", func() {
			tests := []test{
				{"0 >= 0", evaluator.SQLTrue},
				{"-1 >= 1", evaluator.SQLFalse},
				{"1 >= -1", evaluator.SQLTrue},
				{"11 >= 10", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLIsExpr", func() {
			tests := []test{
				{"1 is true", evaluator.SQLTrue},
				{"null is true", evaluator.SQLFalse},
				{"null is unknown", evaluator.SQLTrue},
				{"1 is unknown", evaluator.SQLFalse},
				{"true is true", evaluator.SQLTrue},
				{"0 is false", evaluator.SQLTrue},
				{"1 is false", evaluator.SQLFalse},
				{"'1' is true", evaluator.SQLTrue},
				{"'0.0' is true", evaluator.SQLFalse},
				{"'cats' is false", evaluator.SQLTrue},
				{"DATE '2006-05-04' is false", evaluator.SQLFalse},
				{"TIMESTAMP '2008-04-06 15:32:23' is true", evaluator.SQLTrue},
				{"1 is null", evaluator.SQLFalse},
				{"null is null", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLIsNotExpr", func() {
			tests := []test{
				{"1 is not true", evaluator.SQLFalse},
				{"null is not true", evaluator.SQLTrue},
				{"null is not unknown", evaluator.SQLFalse},
				{"1 is not unknown", evaluator.SQLTrue},
				{"false is not true", evaluator.SQLTrue},
				{"0 is not false", evaluator.SQLFalse},
				{"1 is not false", evaluator.SQLTrue},
				{"'1' is not true", evaluator.SQLFalse},
				{"'0.0' is not true", evaluator.SQLTrue},
				{"'cats' is not false", evaluator.SQLFalse},
				{"DATE '2006-05-04' is not false", evaluator.SQLTrue},
				{"TIMESTAMP '2008-04-06 15:32:23' is not true", evaluator.SQLFalse},
				{"1 is not null", evaluator.SQLTrue},
				{"null is not null", evaluator.SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLIDivideExpr", func() {
			tests := []test{
				{"0 DIV 0", evaluator.SQLNull},
				{"0 DIV 5", evaluator.SQLInt(0)},
				{"5.5 DIV 2", evaluator.SQLInt(2)},
				{"-5 DIV 2", evaluator.SQLInt(-2)},
				{"NULL DIV 1", evaluator.SQLNull},
				{"1 DIV NULL", evaluator.SQLNull},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLInExpr", func() {
			tests := []test{
				{"0 IN(0)", evaluator.SQLTrue},
				{"-1 IN(1)", evaluator.SQLFalse},
				{"0 IN(10, 0)", evaluator.SQLTrue},
				{"-1 IN(1, 10)", evaluator.SQLFalse},
				{"NULL IN(0, 1)", evaluator.SQLNull},
				{"NULL IN(0, NULL)", evaluator.SQLNull},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLLessThanExpr", func() {
			tests := []test{
				{"0 < 0", evaluator.SQLFalse},
				{"-1 < 1", evaluator.SQLTrue},
				{"1 < -1", evaluator.SQLFalse},
				{"10 < 11", evaluator.SQLTrue},
				{"true < '5'", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLLessThanOrEqualExpr", func() {
			tests := []test{
				{"0 <= 0", evaluator.SQLTrue},
				{"-1 <= 1", evaluator.SQLTrue},
				{"1 <= -1", evaluator.SQLFalse},
				{"10 <= 11", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLLikeExpr", func() {
			tests := []test{
				{"'Á Â Ã Ä' LIKE '%'", evaluator.SQLTrue},
				{"'Á Â Ã Ä' LIKE 'Á Â Ã Ä'", evaluator.SQLTrue},
				{"'Á Â Ã Ä' LIKE 'Á%'", evaluator.SQLTrue},
				{"'a' LIKE 'a'", evaluator.SQLTrue},
				{"'Adam' LIKE 'am'", evaluator.SQLFalse},
				{"'Adam' LIKE 'adaM'", evaluator.SQLTrue}, // mixed case test
				{"'Adam' LIKE '%am%'", evaluator.SQLTrue},
				{"'Adam' LIKE 'Ada_'", evaluator.SQLTrue},
				{"'Adam' LIKE '__am'", evaluator.SQLTrue},
				{"'Clever' LIKE '%is'", evaluator.SQLFalse},
				{"'Adam is nice' LIKE '%xs '", evaluator.SQLFalse},
				{"'Adam is nice' LIKE '%is nice'", evaluator.SQLTrue},
				{"'abc' LIKE 'ABC'", evaluator.SQLTrue},    //case sensitive test
				{"'abc' LIKE 'ABC '", evaluator.SQLFalse},  // trailing space test
				{"'abc' LIKE ' ABC'", evaluator.SQLFalse},  // leading space test
				{"'abc' LIKE ' ABC '", evaluator.SQLFalse}, // padded space test
				{"'abc' LIKE 'ABC	'", evaluator.SQLFalse}, // trailing tab test
				{"'10' LIKE '1%'", evaluator.SQLTrue},
				{"'a   ' LIKE 'A   '", evaluator.SQLTrue},
				{"CURRENT_DATE() LIKE '2015-05-31%'", evaluator.SQLFalse},
				{"CURDATE() LIKE '2015-05-31%'", evaluator.SQLFalse},
				{"(DATE '2008-01-02') LIKE '2008-01%'", evaluator.SQLTrue},
				{"NOW() LIKE '" + strconv.Itoa(time.Now().Year()) + "%' ", evaluator.SQLTrue},
				{"10 LIKE '1%'", evaluator.SQLTrue},
				{"1.20 LIKE '1.2%'", evaluator.SQLTrue},
				{"NULL LIKE '1%'", evaluator.SQLNull},
				{"10 LIKE NULL", evaluator.SQLNull},
				{"NULL LIKE NULL", evaluator.SQLNull},
				{"'David_' LIKE 'David\\_'", evaluator.SQLTrue},
				{"'David%' LIKE 'David\\%'", evaluator.SQLTrue},
				{"'David_' LIKE 'David|_' ESCAPE '|'", evaluator.SQLTrue},
				{"'David\\_' LIKE 'David\\_' ESCAPE ''", evaluator.SQLTrue},
				{"'David_' LIKE 'David\\_' ESCAPE char(92)", evaluator.SQLTrue},
				{"'David_' LIKE 'David|_' {escape '|'}", evaluator.SQLTrue},
			}
			runTests(evalCtx, tests)
		})

		Convey("Subject: Mix Arithmetic and Boolean", func() {
			tests := []test{
				{"(5<6) + 1", evaluator.SQLInt(2)},
				{"(5<6) && (6>4)", evaluator.SQLTrue},
				{"(5<6) || (6>4)", evaluator.SQLTrue},
				{"(5<6) XOR (6>4)", evaluator.SQLFalse},
				{"(5<6)<7", evaluator.SQLTrue},
				{"1+(5<6)", evaluator.SQLInt(2)},
				{"1+(5>6)", evaluator.SQLInt(1)},
				{"1+(NULL>6)", evaluator.SQLNull},
				{"NULL+(5>6)", evaluator.SQLNull},
				{"20/(5<6)", evaluator.SQLFloat(20)},
				{"20*(5<6)", evaluator.SQLInt(20)},
				{"20/5<6", evaluator.SQLTrue},
				{"20*5<6", evaluator.SQLFalse},
				{"20+5<6", evaluator.SQLFalse},
				{"20-5<6", evaluator.SQLFalse},
				{"20+true", evaluator.SQLInt(21)},
				{"20+false", evaluator.SQLInt(20)},
			}
			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLModExpr", func() {
			tests := []test{
				{"0 % 0", evaluator.SQLNull},
				{"5 % 2", evaluator.SQLFloat(1)},
				{"5.5 % 2", evaluator.SQLFloat(1.5)},
				{"-5 % -3", evaluator.SQLFloat(-2)},
				{"5 MOD 2", evaluator.SQLFloat(1)},
				{"5.5 MOD 2", evaluator.SQLFloat(1.5)},
				{"-5 MOD -3", evaluator.SQLFloat(-2)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLMultiplyExpr", func() {
			tests := []test{
				{"0 * 0", evaluator.SQLInt(0)},
				{"-1 * 1", evaluator.SQLInt(-1)},
				{"10 * 32", evaluator.SQLInt(320)},
				{"-10 * -32", evaluator.SQLInt(320)},
				{"2.5 * 3", evaluator.SQLDecimal128(decimal.New(75, -1))},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNotEqualsExpr", func() {
			tests := []test{
				{"0 <> 0", evaluator.SQLFalse},
				{"-1 <> 1", evaluator.SQLTrue},
				{"10 <> 10", evaluator.SQLFalse},
				{"-10 <> -10", evaluator.SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNotExpr", func() {
			tests := []test{
				{"NOT 1", evaluator.SQLFalse},
				{"NOT 0", evaluator.SQLTrue},
				{"NOT true", evaluator.SQLFalse},
				{"NOT false", evaluator.SQLTrue},
				{"NOT NULL", evaluator.SQLNull},
				{"! 1", evaluator.SQLFalse},
				{"! 0", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNullSafeEqualsExpr", func() {
			tests := []test{
				{"0 <=> 0", evaluator.SQLTrue},
				{"-1 <=> 1", evaluator.SQLFalse},
				{"10 <=> 10", evaluator.SQLTrue},
				{"-10 <=> -10", evaluator.SQLTrue},
				{"1 <=> 1", evaluator.SQLTrue},
				{"NULL <=> NULL", evaluator.SQLTrue},
				{"1 <=> NULL", evaluator.SQLFalse},
				{"NULL <=> 1", evaluator.SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLOrExpr", func() {
			tests := []test{
				{"1 OR 1", evaluator.SQLTrue},
				{"1 OR 0", evaluator.SQLTrue},
				{"0 OR 1", evaluator.SQLTrue},
				{"NULL OR 1", evaluator.SQLTrue},
				{"NULL OR 0", evaluator.SQLNull},
				{"NULL OR NULL", evaluator.SQLNull},
				{"0 OR 0", evaluator.SQLFalse},
				{"true OR true", evaluator.SQLTrue},
				{"true OR false", evaluator.SQLTrue},
				{"false OR true", evaluator.SQLTrue},
				{"false OR false", evaluator.SQLFalse},
				{"1 || 1", evaluator.SQLTrue},
				{"1 || 0", evaluator.SQLTrue},
				{"0 || 1", evaluator.SQLTrue},
				{"0 || 0", evaluator.SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLXOrExpr", func() {
			tests := []test{
				{"1 XOR 1", evaluator.SQLFalse},
				{"1 XOR 0", evaluator.SQLTrue},
				{"0 XOR 1", evaluator.SQLTrue},
				{"0 XOR 0", evaluator.SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLNotRegexExpr", func() {
			tests := []test{
				{"'ABC123' NOT REGEXP 'AB'", evaluator.SQLFalse},
				{"'ABC123' NOT REGEXP 'ABD'", evaluator.SQLTrue},
				{"'ABC123' NOT REGEXP '[[:alpha:]]'", evaluator.SQLFalse},
				{"'fofo' NOT REGEXP '^fo'", evaluator.SQLFalse},
				{"'fofo' NOT REGEXP '^f.*$'", evaluator.SQLFalse},
				{"'pi' NOT REGEXP 'pi|apa'", evaluator.SQLFalse},
				{"'abcde' NOT REGEXP 'a[bcd]{2}e'", evaluator.SQLTrue},
				{"'abcde' NOT REGEXP 'a[bcd]{1,10}e'", evaluator.SQLFalse},
				{"null REGEXP 'abc'", evaluator.SQLNull},
				{"'a' REGEXP null", evaluator.SQLNull},
				{"2-1 NOT REGEXP '1'", evaluator.SQLFalse},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLRegexExpr", func() {
			tests := []test{
				{"'ABC123' REGEXP 'AB'", evaluator.SQLTrue},
				{"'ABC123' REGEXP 'ABD'", evaluator.SQLFalse},
				{"'ABC123' REGEXP '[[:alpha:]]'", evaluator.SQLTrue},
				{"'fofo' REGEXP '^fo'", evaluator.SQLTrue},
				{"'fofo' REGEXP '^f.*$'", evaluator.SQLTrue},
				{"'pi' REGEXP 'pi|apa'", evaluator.SQLTrue},
				{"'abcde' REGEXP 'a[bcd]{2}e'", evaluator.SQLFalse},
				{"'abcde' REGEXP 'a[bcd]{1,10}e'", evaluator.SQLTrue},
				{"null REGEXP 'abc'", evaluator.SQLNull},
				{"'a' REGEXP null", evaluator.SQLNull},
				{"2-1 REGEXP '1'", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLScalarFunctionExpr", func() {

			Convey("Subject: ABS", func() {
				tests := []test{
					{"ABS(NULL)", evaluator.SQLNull},
					{"ABS('C')", evaluator.SQLFloat(0)},
					{"ABS(-20)", evaluator.SQLFloat(20)},
					{"ABS(20)", evaluator.SQLFloat(20)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ACOS", func() {
				tests := []test{
					{"ACOS(NULL)", evaluator.SQLNull},
					{"ACOS(20)", evaluator.SQLNull},
					{"ACOS(-20)", evaluator.SQLNull},
					{"ACOS('C')", evaluator.SQLFloat(1.5707963267948966)},
					{"ACOS(0)", evaluator.SQLFloat(1.5707963267948966)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ASIN", func() {
				tests := []test{
					{"ASIN(NULL)", evaluator.SQLNull},
					{"ASIN(20)", evaluator.SQLNull},
					{"ASIN(-20)", evaluator.SQLNull},
					{"ASIN('C')", evaluator.SQLFloat(0)},
					{"ASIN(0)", evaluator.SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ATAN", func() {
				tests := []test{
					{"ATAN(NULL)", evaluator.SQLNull},
					{"ATAN(20)", evaluator.SQLFloat(1.5208379310729538)},
					{"ATAN(-20)", evaluator.SQLFloat(-1.5208379310729538)},
					{"ATAN('C')", evaluator.SQLFloat(0)},
					{"ATAN(0)", evaluator.SQLFloat(0)},
					{"ATAN(NULL, NULL)", evaluator.SQLNull},
					{"ATAN(-2, 2)", evaluator.SQLFloat(-0.7853981633974483)},
					{"ATAN('C', 2)", evaluator.SQLFloat(0)},
					{"ATAN(0, 2)", evaluator.SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ATAN2", func() {
				tests := []test{
					{"ATAN2(NULL, NULL)", evaluator.SQLNull},
					{"ATAN2(-2, 2)", evaluator.SQLFloat(-0.7853981633974483)},
					{"ATAN2('C', 2)", evaluator.SQLFloat(0)},
					{"ATAN2(0, 2)", evaluator.SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ADDDATE", func() {
				d, err := time.Parse("2006-01-02", "2003-01-02")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
				So(err, ShouldBeNil)
				t2, err := time.Parse("2006-01-02 15:04:05", "2003-02-03 10:28:06")
				So(err, ShouldBeNil)
				t3, err := time.Parse("2006-01-02 15:04:05", "2003-02-14 10:28:06")
				So(err, ShouldBeNil)
				d2, err := time.Parse("2006-01-02", "2003-11-30")
				So(err, ShouldBeNil)
				d3, err := time.Parse("2006-01-02", "2008-02-02")
				So(err, ShouldBeNil)

				tests := []test{
					{"ADDDATE(NULL, INTERVAL 1 YEAR)", evaluator.SQLNull},
					{"ADDDATE('2002-01-02', INTERVAL 1 YEAR)", evaluator.SQLTimestamp{Time: d}},
					{"ADDDATE('2003-08-31', INTERVAL 1 QUARTER)", evaluator.SQLTimestamp{Time: d2}},
					{"ADDDATE('2003-10-31', INTERVAL 1 MONTH)", evaluator.SQLTimestamp{Time: d2}},
					{"ADDDATE('2003-01-01', INTERVAL 1 DAY)", evaluator.SQLTimestamp{Time: d}},
					{"ADDDATE('2003-01-02 14:30:09', INTERVAL -2 HOUR)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 12:23:09', INTERVAL 7 MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 12:30:12', INTERVAL -3 SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 12:32:10', INTERVAL '-2:1' MINUTE_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 05:27:06', INTERVAL '7:3:3' HOUR_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 15:32:09', INTERVAL '-3:2' HOUR_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2002-12-31 10:27:05', INTERVAL '2 2:3:4' DAY_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2002-12-31 10:27:09', INTERVAL '2 2:3' DAY_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-01 08:30:09', INTERVAL '1 4' DAY_HOUR)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2000-09-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 12:33:09', INTERVAL '-3' HOUR_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"ADDDATE('2003-01-02 10:28:06', 32)", evaluator.SQLTimestamp{Time: t2}},
					{"ADDDATE('2003-01-02 10:28:06', 43)", evaluator.SQLTimestamp{Time: t3}},
					{"ADDDATE('2003-01-02 10:28:06.000', 43)", evaluator.SQLTimestamp{Time: t3}},
					{"ADDDATE('2003-01-02 10:28:06.000000', 43)", evaluator.SQLTimestamp{Time: t3}},
					{"ADDDATE('2008-01-02', 31)", evaluator.SQLTimestamp{Time: d3}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ASCII", func() {
				tests := []test{
					{"ASCII(NULL)", evaluator.SQLNull},
					{"ASCII('')", evaluator.SQLInt(0)},
					{"ASCII('A')", evaluator.SQLInt(65)},
					{"ASCII('AWESOME')", evaluator.SQLInt(65)},
					{"ASCII('¢')", evaluator.SQLInt(194)},
					{"ASCII('Č')", evaluator.SQLInt(196)}, // This is actually 268, but the first byte is 196
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CEIL", func() {
				tests := []test{
					{"CEIL(NULL)", evaluator.SQLNull},
					{"CEIL(20)", evaluator.SQLFloat(20)},
					{"CEIL(-20)", evaluator.SQLFloat(-20)},
					{"CEIL('C')", evaluator.SQLFloat(0)},
					{"CEIL(0.56)", evaluator.SQLFloat(1)},
					{"CEIL(-0.56)", evaluator.SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CEILING", func() {
				tests := []test{
					{"CEIL(NULL)", evaluator.SQLNull},
					{"CEIL(20)", evaluator.SQLFloat(20)},
					{"CEIL(-20)", evaluator.SQLFloat(-20)},
					{"CEIL('C')", evaluator.SQLFloat(0)},
					{"CEIL(0.56)", evaluator.SQLFloat(1)},
					{"CEIL(-0.56)", evaluator.SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CHAR", func() {
				tests := []test{
					{"CHAR(NULL)", evaluator.SQLVarchar("")},
					{"CHAR(77,121,83,81,'76')", evaluator.SQLVarchar("MySQL")},
					{"CHAR(77,121,NULL, 83, NULL, 81,'76')", evaluator.SQLVarchar("MySQL")},
					{"CHAR(256)", evaluator.SQLVarchar(string([]byte{1, 0}))},
					{"CHAR(512)", evaluator.SQLVarchar(string([]byte{2, 0}))},
					{"CHAR(513)", evaluator.SQLVarchar(string([]byte{2, 1}))},
					{"CHAR(256, 512)", evaluator.SQLVarchar(string([]byte{1, 0, 2, 0}))},
					{"CHAR(65537)", evaluator.SQLVarchar(string([]byte{1, 0, 1}))},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CHAR_LENGTH", func() {
				tests := []test{
					{"CHAR_LENGTH(NULL)", evaluator.SQLNull},
					{"CHAR_LENGTH('sDg')", evaluator.SQLInt(3)},
					{"CHAR_LENGTH('世界')", evaluator.SQLInt(2)},
					{"CHAR_LENGTH('')", evaluator.SQLInt(0)},

					{"CHARACTER_LENGTH(NULL)", evaluator.SQLNull},
					{"CHARACTER_LENGTH('sDg')", evaluator.SQLInt(3)},
					{"CHARACTER_LENGTH('世界')", evaluator.SQLInt(2)},
					{"CHARACTER_LENGTH('')", evaluator.SQLInt(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: COALESCE", func() {
				tests := []test{
					{"COALESCE(NULL)", evaluator.SQLNull},
					{"COALESCE('A')", evaluator.SQLVarchar("A")},
					{"COALESCE('A', NULL)", evaluator.SQLVarchar("A")},
					{"COALESCE('A', 'B')", evaluator.SQLVarchar("A")},
					{"COALESCE(NULL, 'A', NULL, 'B')", evaluator.SQLVarchar("A")},
					{"COALESCE(NULL, NULL, NULL)", evaluator.SQLNull},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"COALESCE(NULL, 1, 'A')", schema.SQLVarchar},
					{"COALESCE(NULL, 1, 23)", schema.SQLInt},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: CONCAT", func() {
				tests := []test{
					{"CONCAT(NULL)", evaluator.SQLNull},
					{"CONCAT('A')", evaluator.SQLVarchar("A")},
					{"CONCAT('A', 'B')", evaluator.SQLVarchar("AB")},
					{"CONCAT('A', NULL, 'B')", evaluator.SQLNull},
					{"CONCAT('A', 123, 'B')", evaluator.SQLVarchar("A123B")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CONCAT_WS", func() {
				tests := []test{
					{"CONCAT_WS(NULL, NULL)", evaluator.SQLNull},
					{"CONCAT_WS(',','A')", evaluator.SQLVarchar("A")},
					{"CONCAT_WS(',','A', 'B')", evaluator.SQLVarchar("A,B")},
					{"CONCAT_WS(',','A', NULL, 'B')", evaluator.SQLVarchar("A,B")},
					{"CONCAT_WS(',','A', 123, 'B')", evaluator.SQLVarchar("A,123,B")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CONNECTION_ID", func() {
				tests := []test{
					{"CONNECTION_ID()", evaluator.SQLUint32(42)},
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
					{"CONVERT(NULL, BINARY)", evaluator.SQLNull},
					{"CONVERT(NULL, BINARY(20))", evaluator.SQLNull},
					{"CONVERT(3, BINARY)", evaluator.SQLNull},
					{"CONVERT(3, BINARY(20))", evaluator.SQLNull},
					{"CONVERT('asgg', BINARY)", evaluator.SQLNull},
					{"CONVERT('asgg', BINARY(20))", evaluator.SQLNull},
					{"CONVERT(NULL, SIGNED)", evaluator.SQLNull},
					{"CONVERT(3, SIGNED)", evaluator.SQLInt(3)},
					{"CONVERT(3.4, SIGNED)", evaluator.SQLInt(3)},
					{"CONVERT(3.5, SIGNED INTEGER)", evaluator.SQLInt(4)},
					{"CONVERT(-3.4, SIGNED INTEGER)", evaluator.SQLInt(-3)},
					{"CONVERT(33245368230, SQL_BIGINT)", evaluator.SQLInt(33245368230)},
					{"CONVERT('janna', UNSIGNED INTEGER)", evaluator.SQLUint64(0)},
					{"CONVERT('423', UNSIGNED)", evaluator.SQLUint64(423)},
					{"CONVERT('-423', UNSIGNED)", evaluator.SQLUint64(0xfffffffffffffe59)},
					{"CONVERT('16a', SIGNED)", evaluator.SQLInt(0)},
					{"CONVERT(true, SIGNED)", evaluator.SQLInt(1)},
					{"CONVERT(false, SIGNED)", evaluator.SQLInt(0)},
					{"CONVERT(DATE '2006-05-11', SIGNED)", evaluator.SQLInt(20060511)},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', SIGNED)", evaluator.SQLInt(20060511123212)},
					{"CONVERT(NULL, DECIMAL)", evaluator.SQLNull},
					{"CONVERT(NULL, DECIMAL(12))", evaluator.SQLNull},
					{"CONVERT(NULL, DECIMAL(12,25))", evaluator.SQLNull},
					{"CONVERT('str', DECIMAL)", evaluator.SQLNull},
					{"CONVERT('str', DECIMAL(12))", evaluator.SQLNull},
					{"CONVERT('str', DECIMAL(12,25))", evaluator.SQLNull},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DECIMAL)", evaluator.SQLDecimal128(decimal.New(20060511123212, 0))},
					{"CONVERT('423', DECIMAL)", evaluator.SQLDecimal128(decimal.New(423, 0))},
					{"CONVERT('423', DECIMAL(12))", evaluator.SQLDecimal128(decimal.New(423, 0))},
					{"CONVERT('423', DECIMAL(12,25))", evaluator.SQLDecimal128(decimal.New(423, 0))},
					{"CONVERT(423, DECIMAL)", evaluator.SQLDecimal128(decimal.New(423, 0))},
					{"CONVERT(423, DECIMAL(12))", evaluator.SQLDecimal128(decimal.New(423, 0))},
					{"CONVERT(423, DECIMAL(12,25))", evaluator.SQLDecimal128(decimal.New(423, 0))},
					{"CONVERT(NULL, SQL_DOUBLE)", evaluator.SQLNull},
					{"CONVERT(3, SQL_DOUBLE)", evaluator.SQLFloat(3)},
					{"CONVERT(-3.4, SQL_DOUBLE)", evaluator.SQLFloat(-3.4)},
					{"CONVERT('janna', SQL_DOUBLE)", evaluator.SQLFloat(0)},
					{"CONVERT('4.4', SQL_DOUBLE)", evaluator.SQLFloat(4.4)},
					{"CONVERT('16a', SQL_DOUBLE)", evaluator.SQLFloat(0)},
					{"CONVERT(true, SQL_DOUBLE)", evaluator.SQLFloat(1)},
					{"CONVERT(false, SQL_DOUBLE)", evaluator.SQLFloat(0)},
					{"CONVERT(DATE '2006-05-11', SQL_DOUBLE)", evaluator.SQLFloat(20060511)},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', SQL_DOUBLE)", evaluator.SQLFloat(20060511123212)},
					{"CONVERT(NULL, CHAR)", evaluator.SQLNull},
					{"CONVERT(NULL, CHAR(3))", evaluator.SQLNull},
					{"CONVERT(3, CHAR)", evaluator.SQLVarchar("3")},
					{"CONVERT(3, CHAR(4))", evaluator.SQLVarchar("3")},
					{"CONVERT(-3.4, SQL_VARCHAR)", evaluator.SQLVarchar("-3.4")},
					{"CONVERT('janna', CHAR)", evaluator.SQLVarchar("janna")},
					{"CONVERT('janna', CHAR(12))", evaluator.SQLVarchar("janna")},
					{"CONVERT('16a', CHAR)", evaluator.SQLVarchar("16a")},
					{"CONVERT('16a', CHAR(20))", evaluator.SQLVarchar("16a")},
					{"CONVERT(true, CHAR)", evaluator.SQLVarchar("1")},
					{"CONVERT(true, CHAR(20))", evaluator.SQLVarchar("1")},
					{"CONVERT(false, CHAR)", evaluator.SQLVarchar("0")},
					{"CONVERT(false, CHAR(20))", evaluator.SQLVarchar("0")},
					{"CONVERT(DATE '2006-05-11', CHAR)", evaluator.SQLVarchar("2006-05-11")},
					{"CONVERT(DATE '2006-05-11', CHAR(20))", evaluator.SQLVarchar("2006-05-11")},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', CHAR)", evaluator.SQLVarchar("2006-05-11 12:32:12")},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', CHAR(20))", evaluator.SQLVarchar("2006-05-11 12:32:12")},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', NCHAR)", evaluator.SQLVarchar("2006-05-11 12:32:12")},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', NCHAR(20))", evaluator.SQLVarchar("2006-05-11 12:32:12")},
					{"CONVERT(NULL, DATE)", evaluator.SQLNull},
					{"CONVERT(3, DATE)", evaluator.SQLNull},
					{"CONVERT(-3.4, SQL_DATE)", evaluator.SQLNull},
					{"CONVERT('janna', DATE)", evaluator.SQLNull},
					{"CONVERT('2006-05-11', DATE)", evaluator.SQLDate{Time: d}},
					{"CONVERT(true, DATE)", evaluator.SQLNull},
					{"CONVERT(DATE '2006-05-11', DATE)", evaluator.SQLDate{Time: d}},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATE)", evaluator.SQLDate{Time: d}},
					{"CONVERT(NULL, DATETIME)", evaluator.SQLNull},
					{"CONVERT(-3.4, DATETIME)", evaluator.SQLNull},
					{"CONVERT('janna', DATETIME)", evaluator.SQLNull},
					{"CONVERT('2006-05-11', DATETIME)", evaluator.SQLTimestamp{Time: dt}},
					{"CONVERT(true, DATETIME)", evaluator.SQLNull},
					{"CONVERT(3, SQL_TIMESTAMP)", evaluator.SQLNull},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATETIME)", evaluator.SQLTimestamp{Time: t}},
					{"CONVERT(DATE '2006-05-11', SQL_TIMESTAMP)", evaluator.SQLTimestamp{Time: dt}},
					{"CONVERT('12:32:12', TIME)", evaluator.SQLTimestamp{Time: time.Date(0, 1, 1, 12, 32, 12, 0, time.UTC)}},
					{"CONVERT('2006-04-11 12:32:12', TIME)", evaluator.SQLTimestamp{Time: time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)}},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"CONVERT(DATE '2006-05-11', SIGNED)", schema.SQLInt},
					{"CONVERT(true, SQL_DOUBLE)", schema.SQLFloat},
					{"CONVERT('16a', CHAR)", schema.SQLVarchar},
					{"CONVERT('2006-05-11', DATE)", schema.SQLDate},
					{"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATETIME)", schema.SQLTimestamp},
					{"CONVERT(DATE '2006-05-11', SQL_TIMESTAMP)", schema.SQLTimestamp},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: COS", func() {
				tests := []test{
					{"COS(NULL)", evaluator.SQLNull},
					{"COS(20)", evaluator.SQLFloat(0.40808206181339196)},
					{"COS(-20)", evaluator.SQLFloat(0.40808206181339196)},
					{"COS('C')", evaluator.SQLFloat(1)},
					{"COS(0)", evaluator.SQLFloat(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: COT", func() {
				tests := []test{
					{"COT(NULL)", evaluator.SQLNull},
					{"COT(19)", evaluator.SQLFloat(6.596764247280111)},
					{"COT(-19)", evaluator.SQLFloat(-6.596764247280111)},
				}
				runTests(evalCtx, tests)

				Convey("Should error when out of range", func() {
					subject, err := evaluator.NewSQLScalarFunctionExpr("cot", []evaluator.SQLExpr{evaluator.SQLFloat(0)})
					So(err, ShouldBeNil)
					_, err = subject.Evaluate(evalCtx)
					So(err, ShouldNotBeNil)
				})

			})

			SkipConvey("Subject: CURRENT_DATE", func() {
				tests := []test{
					{"CURRENT_DATE()", evaluator.SQLDate{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: CURRENT_TIMESTAMP", func() {
				tests := []test{
					{"CURRENT_TIMESTAMP()", evaluator.SQLTimestamp{time.Now().UTC()}},
					{"CURRENT_TIMESTAMP", evaluator.SQLTimestamp{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: CURTIME", func() {
				tests := []test{
					{"CURRENT_TIMESTAMP()", evaluator.SQLTimestamp{time.Now().UTC()}},
					{"CURRENT_TIMESTAMP", evaluator.SQLTimestamp{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: UTC_TIMESTAMP", func() {
				tests := []test{
					{"UTC_TIMESTAMP()", evaluator.SQLTimestamp{time.Now().UTC()}},
					{"UTC_TIMESTAMP", evaluator.SQLTimestamp{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: UTC_DATE", func() {
				now := time.Now().In(time.UTC)
				t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
				tests := []test{
					{"UTC_DATE()", evaluator.SQLDate{t}},
					{"UTC_DATE", evaluator.SQLDate{t}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: CURRENT_USER/SESSION_USER/SYSTEM_USER/USER", func() {
				tests := []test{
					{"CURRENT_USER()", evaluator.SQLVarchar("test user")},
					{"SESSION_USER()", evaluator.SQLVarchar("test user")},
					{"SYSTEM_USER()", evaluator.SQLVarchar("test user")},
					{"USER()", evaluator.SQLVarchar("test user")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DATABASE/SCHEMA", func() {
				tests := []test{
					{"DATABASE()", evaluator.SQLVarchar("test")},
					{"SCHEMA()", evaluator.SQLVarchar("test")},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: NOW", func() {
				tests := []test{
					{"NOW()", evaluator.SQLTimestamp{time.Now().UTC()}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DATE", func() {
				fmtString := "2006-01-02"

				d, err := time.Parse(fmtString, "2016-03-01")
				So(err, ShouldBeNil)

				dExpected := evaluator.SQLDate{Time: d}

				preCutoff, err := time.Parse(fmtString, "2069-12-31")
				So(err, ShouldBeNil)

				postCutoff, err := time.Parse(fmtString, "1970-01-01")
				So(err, ShouldBeNil)

				jan112000, err := time.Parse(fmtString, "2000-01-11")
				So(err, ShouldBeNil)

				nov102000, err := time.Parse(fmtString, "2000-11-10")
				So(err, ShouldBeNil)

				nov102006, err := time.Parse(fmtString, "2006-11-10")
				So(err, ShouldBeNil)

				nov100116, err := time.Parse(fmtString, "0116-11-10")
				So(err, ShouldBeNil)

				may042000, err := time.Parse(fmtString, "2000-05-04")
				So(err, ShouldBeNil)

				tests := []test{
					// invalid inputs
					{"DATE(NULL)", evaluator.SQLNull},
					{"DATE(23)", evaluator.SQLNull},
					{"DATE('cat')", evaluator.SQLNull},
					{"DATE(6911)", evaluator.SQLNull},
					{"DATE(2017110722040)", evaluator.SQLNull},
					{"DATE(-50)", evaluator.SQLNull},
					{"DATE('')", evaluator.SQLNull},

					// explicitly labeling input as date/timestamp
					{"DATE(TIMESTAMP '2016-03-01 12:32:23')", dExpected},
					{"DATE(DATE '2016-03-01')", dExpected},

					// unlabeled string inputs
					{"DATE('2016-03-01 12:32:23')", dExpected},
					{"DATE('2016-03-01')", dExpected},
					{"DATE('20160301')", dExpected},

					// number inputs
					{"DATE(20160301)", dExpected},
					{"DATE(20160301123456)", dExpected},
					{"DATE(160301123456)", dExpected},
					{"DATE(160301)", dExpected},

					// numbers that are too short to pad
					{"DATE(1)", evaluator.SQLNull},
					{"DATE(11)", evaluator.SQLNull},

					// number inputs requiring padding
					{"DATE(111)", evaluator.SQLDate{Time: jan112000}},
					{"DATE(1110)", evaluator.SQLDate{Time: nov102000}},
					{"DATE(61110)", evaluator.SQLDate{Time: nov102006}},
					{"DATE(1161110)", evaluator.SQLDate{Time: nov100116}},
					{"DATE(504123025)", evaluator.SQLDate{Time: may042000}},
					{"DATE(1110123025)", evaluator.SQLDate{Time: nov102000}},
					{"DATE(61110123025)", evaluator.SQLDate{Time: nov102006}},
					{"DATE(61110123025.22)", evaluator.SQLDate{Time: nov102006}},
					{"DATE(1161110123025)", evaluator.SQLDate{Time: nov100116}},

					// alternate delimiters
					{"DATE('16-03-01')", dExpected},
					{"DATE('2016.03.01')", dExpected},

					// mixed delimiters
					{"DATE('2016@03.01')", dExpected},
					{"DATE('2016-03-01 12.32.23')", dExpected},

					// shortened form of single-digit values
					{"DATE('16-03-1')", dExpected},
					{"DATE('2016.3.1')", dExpected},
					{"DATE('16.3.1')", dExpected},

					// timestamp w/ fractional seconds
					{"DATE('2016-03-01 12.32.23.3333')", dExpected},

					// use T instead of space to separate
					{"DATE('2016-03-01T12.32.23.3333')", dExpected},

					// make sure behavior around year cutoff is correct -
					// 0-69 are intepreted as 2000-2069, while 70-99 are
					// interpreted as 1970-1999.
					{"DATE('69-12-31')", evaluator.SQLDate{Time: preCutoff}},
					{"DATE('70-01-01')", evaluator.SQLDate{Time: postCutoff}},
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
					{"DATE_ADD('2002-12-31 10:27:04.500000', INTERVAL '2 2:3:4.5' DAY_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},

					{"DATE_ADD('2003-01-02 10:28:05.500000', INTERVAL '2:2:3.5' DAY_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 10:28:05.500000', INTERVAL '2:2:3.5' HOUR_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2002-12-31 10:27:05', INTERVAL '2 2:3:4' DAY_SECOND)", evaluator.SQLTimestamp{Time: t}},

					{"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' DAY_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' HOUR_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' MINUTE_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 10:27:05', INTERVAL '2:3:4' DAY_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 10:27:05', INTERVAL '2:3:4' HOUR_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2002-12-31 10:27:09', INTERVAL '2 2:3' DAY_MINUTE)", evaluator.SQLTimestamp{Time: t}},

					{"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' DAY_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' HOUR_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' MINUTE_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' SECOND_MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' DAY_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' HOUR_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' MINUTE_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 15:32:09', INTERVAL '-3:2' DAY_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 15:32:09', INTERVAL '-3:2' HOUR_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2002-12-31 10:30:09', INTERVAL '2 2' DAY_HOUR)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2000-09-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)", evaluator.SQLTimestamp{Time: t}},

					{"DATE_ADD('2002-01-02', INTERVAL NULL YEAR)", evaluator.SQLNull},
					{"DATE_ADD(NULL, INTERVAL 1 YEAR)", evaluator.SQLNull},
					{"DATE_ADD('2002-01-02', INTERVAL 1 YEAR)", evaluator.SQLTimestamp{Time: d}},
					{"DATE_ADD('2003-08-31', INTERVAL 1 QUARTER)", evaluator.SQLTimestamp{Time: d2}},
					{"DATE_ADD('2003-10-31', INTERVAL 1 MONTH)", evaluator.SQLTimestamp{Time: d2}},
					{"DATE_ADD('2003-01-01', INTERVAL 1 DAY)", evaluator.SQLTimestamp{Time: d}},
					{"DATE_ADD('2003-01-02 14:30:09', INTERVAL -2 HOUR)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:23:09', INTERVAL 7 MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:30:12', INTERVAL -3 SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_ADD('2003-01-02 12:30:08.999999', INTERVAL 1 MICROSECOND)", evaluator.SQLTimestamp{Time: t}},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"DATE_ADD('2002-01-02', INTERVAL 1 YEAR)", schema.SQLTimestamp},
					{"DATE_ADD(DATE '2002-01-02', INTERVAL 1 HOUR)", schema.SQLTimestamp},
					{"DATE_ADD(TIMESTAMP '2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)", schema.SQLTimestamp},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: DATEDIFF", func() {
				tests := []test{
					{"DATEDIFF('2017-01-01', '2016-01-01 23:08:56')", evaluator.SQLInt(366)},
					{"DATEDIFF('2017-01-01', '2017-01-01')", evaluator.SQLInt(0)},
					{"DATEDIFF('2017-08-23 10:40:43', '2017-09-30 12:19:50')", evaluator.SQLInt(-38)},
					{"DATEDIFF(NULL, '2017-09-30 12:19:50')", evaluator.SQLNull},
					{"DATEDIFF('2002-09-07', '1700-08-02')", evaluator.SQLInt(106751)},
					{"DATEDIFF('1657-08-02', '2002-09-07')", evaluator.SQLInt(-106751)},
					{"DATEDIFF(20170823104043, '2017-09-30 12:19:50')", evaluator.SQLInt(-38)},
					{"DATEDIFF(20170823.09809, '2017-09-30 12:19:50')", evaluator.SQLInt(-38)},
					{"DATEDIFF('biconnectorisfun', '2017-09-30 12:19:50')", evaluator.SQLNull},
					{"DATEDIFF('2000-9-1', '2012-6-7')", evaluator.SQLInt(-4297)},
					{"DATEDIFF('00-09-1', '12-06-07')", evaluator.SQLInt(-4297)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DATE_FORMAT", func() {
				tests := []test{
					{"DATE_FORMAT('2009-10-04', NULL)", evaluator.SQLNull},
					{"DATE_FORMAT(NULL, '2009-10-04')", evaluator.SQLNull},
					{"DATE_FORMAT('2009-10-04 22:23:00', '%W %M 01 %Y')", evaluator.SQLVarchar("Sunday October 01 2009")},
					{"DATE_FORMAT('2009-10-04 22:23:00', '%W %M %Y')", evaluator.SQLVarchar("Sunday October 2009")},
					{"DATE_FORMAT('2007-10-04 22:23:00', '%H:01:%i:%s')", evaluator.SQLVarchar("22:01:23:00")},
					{"DATE_FORMAT('2007-10-04 22:23:00', '%H:%g:01%%:%i:%s%')", evaluator.SQLVarchar("22:%g:01%:23:00%")},
					{"DATE_FORMAT('2007-10-04 22:23:00', '%H:%i:%s')", evaluator.SQLVarchar("22:23:00")},
					{"DATE_FORMAT('1900-10-04 22:23:00', '%D %y %a %d %m %b %j')", evaluator.SQLVarchar("4th 00 Thu 04 10 Oct 277")},
					{"DATE_FORMAT('1997-10-04 22:23:00', '%H %k %I %r %T %S %w')", evaluator.SQLVarchar("22 22 10 10:23:00 PM 22:23:00 00 6")},
					{"DATE_FORMAT('1999-01-01', '%X %V')", evaluator.SQLVarchar("1998 52")},
					{"DATE_FORMAT('1989-05-14 01:03:01.232335','%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k|%l|%M|%m|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')", evaluator.SQLVarchar("Sun|May|5|14th|14|14|232335|01|01|01|03|134|1|1|May|05|AM|01:03:01 AM|01|01|01:03:01|20|19|20|19|Sunday|0|1989|1989|1989|89|%|1989")},
					{"DATE_FORMAT('1900-10-04 22:23:00', '%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k|%l|%M|%m|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')", evaluator.SQLVarchar("Thu|Oct|10|4th|04|4|000000|22|10|10|23|277|22|10|October|10|PM|10:23:00 PM|00|00|22:23:00|39|40|39|40|Thursday|4|1900|1900|1900|00|%|1900")},
					{"DATE_FORMAT('1983-07-05 23:22', '%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k|%l|%M|%m|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')", evaluator.SQLVarchar("Tue|Jul|7|5th|05|5|000000|23|11|11|22|186|23|11|July|07|PM|11:22:00 PM|00|00|23:22:00|27|27|27|27|Tuesday|2|1983|1983|1983|83|%|1983")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DATE_SUB, SUBDATE", func() {
				d, err := time.Parse("2006-01-02", "2003-01-02")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
				So(err, ShouldBeNil)
				t2, err := time.Parse("2006-01-02 15:04:05", "2007-12-02 12:00:00")
				So(err, ShouldBeNil)
				d2, err := time.Parse("2006-01-02", "2003-11-30")
				So(err, ShouldBeNil)

				tests := []test{
					{"DATE_SUB('2004-01-02', INTERVAL NULL YEAR)", evaluator.SQLNull},
					{"DATE_SUB(NULL, INTERVAL 1 YEAR)", evaluator.SQLNull},
					{"DATE_SUB('2004-01-02', INTERVAL 1 YEAR)", evaluator.SQLTimestamp{Time: d}},
					{"DATE_SUB('2003-04-02', INTERVAL 1 QUARTER)", evaluator.SQLTimestamp{Time: d}},
					{"DATE_SUB('2003-12-31', INTERVAL 1 MONTH)", evaluator.SQLTimestamp{Time: d2}},
					{"DATE_SUB('2003-01-03', INTERVAL 1 DAY)", evaluator.SQLTimestamp{Time: d}},
					{"SUBDATE('2004-01-02', INTERVAL 1 YEAR)", evaluator.SQLTimestamp{Time: d}},
					{"SUBDATE('2003-04-02', INTERVAL 1 QUARTER)", evaluator.SQLTimestamp{Time: d}},
					{"SUBDATE('2003-12-31', INTERVAL 1 MONTH)", evaluator.SQLTimestamp{Time: d2}},
					{"SUBDATE('2008-01-02 12:00:00', 31)", evaluator.SQLTimestamp{Time: t2}},
					{"SUBDATE('2016-01-02 12:00:00', 2953)", evaluator.SQLTimestamp{Time: t2}},
					{"DATE_SUB('2003-01-02 10:30:09', INTERVAL -2 HOUR)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-02 12:37:09', INTERVAL 7 MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-02 12:30:12', INTERVAL 3 SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-02 12:32:10', INTERVAL '2:1' MINUTE_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-02 19:33:12', INTERVAL '7:3:3' HOUR_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-02 15:32:09', INTERVAL '3:2' HOUR_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-04 14:33:13', INTERVAL '2 2:3:4' DAY_SECOND)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-04 14:33:09', INTERVAL '2 2:3' DAY_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-03 16:30:09', INTERVAL '1 4' DAY_HOUR)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2005-05-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-02 12:33:09', INTERVAL '3' HOUR_MINUTE)", evaluator.SQLTimestamp{Time: t}},
					{"DATE_SUB('2003-01-02 14:32:12', INTERVAL '2 2:3' DAY_SECOND)", evaluator.SQLTimestamp{Time: t}},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"DATE_SUB('2002-01-02', INTERVAL 1 YEAR)", schema.SQLTimestamp},
					{"DATE_SUB(DATE '2002-01-02', INTERVAL 1 HOUR)", schema.SQLTimestamp},
					{"DATE_SUB(TIMESTAMP '2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)", schema.SQLTimestamp},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: DAYNAME", func() {
				tests := []test{
					{"DAYNAME(NULL)", evaluator.SQLNull},
					{"DAYNAME(14)", evaluator.SQLNull},
					{"DAYNAME('2016-01-01 00:00:00')", evaluator.SQLVarchar("Friday")},
					{"DAYNAME('2016-1-1')", evaluator.SQLVarchar("Friday")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DAYOFMONTH", func() {
				tests := []test{
					{"DAYOFMONTH(NULL)", evaluator.SQLNull},
					{"DAYOFMONTH(14)", evaluator.SQLNull},
					{"DAYOFMONTH('2016-01-01')", evaluator.SQLInt(1)},
					{"DAYOFMONTH('2016-1-1')", evaluator.SQLInt(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DAYOFWEEK", func() {
				tests := []test{
					{"DAYOFWEEK(NULL)", evaluator.SQLNull},
					{"DAYOFWEEK(14)", evaluator.SQLNull},
					{"DAYOFWEEK('2016-01-01')", evaluator.SQLInt(6)},
					{"DAYOFWEEK('2016-1-1')", evaluator.SQLInt(6)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DAYOFYEAR", func() {
				tests := []test{
					{"DAYOFYEAR(NULL)", evaluator.SQLNull},
					{"DAYOFYEAR(14)", evaluator.SQLNull},
					{"DAYOFYEAR('2016-1-1')", evaluator.SQLInt(1)},
					{"DAYOFYEAR('2016-01-01')", evaluator.SQLInt(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: DEGREES", func() {
				tests := []test{
					{"DEGREES(NULL)", evaluator.SQLNull},
					{"DEGREES(20)", evaluator.SQLFloat(1145.9155902616465)},
					{"DEGREES(-20)", evaluator.SQLFloat(-1145.9155902616465)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ELT", func() {
				tests := []test{
					{"ELT(NULL, 'a', 'b')", evaluator.SQLNull},
					{"ELT(0, 'a', 'b')", evaluator.SQLNull},
					{"ELT(1, 'a', 'b')", evaluator.SQLVarchar("a")},
					{"ELT(2, 'a', 'b')", evaluator.SQLVarchar("b")},
					{"ELT(3, 'a', 'b', NULL)", evaluator.SQLNull},
					{"ELT(4, 'a', 'b', NULL)", evaluator.SQLNull},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: EXP", func() {
				tests := []test{
					{"EXP(NULL)", evaluator.SQLNull},
					{"EXP('sdg')", evaluator.SQLFloat(1)},
					{"EXP(0)", evaluator.SQLFloat(1)},
					{"EXP(2)", evaluator.SQLFloat(7.38905609893065)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: EXTRACT", func() {
				tests := []test{
					{"EXTRACT(YEAR FROM NULL)", evaluator.SQLNull},
					{"EXTRACT(YEAR FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(2006)},
					{"EXTRACT(QUARTER FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(2)},
					{"EXTRACT(WEEK FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(14)},
					{"EXTRACT(DAY FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(7)},
					{"EXTRACT(HOUR FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(7)},
					{"EXTRACT(MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(14)},
					{"EXTRACT(SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(23)},
					{"EXTRACT(MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(0)},
					{"EXTRACT(YEAR_MONTH FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(200604)},
					{"EXTRACT(DAY_HOUR FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(707)},
					{"EXTRACT(DAY_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(70714)},
					{"EXTRACT(DAY_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(7071423)},
					{"EXTRACT(DAY_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(7071423000000)},
					{"EXTRACT(HOUR_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(714)},
					{"EXTRACT(HOUR_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(71423)},
					{"EXTRACT(HOUR_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(71423000000)},
					{"EXTRACT(MINUTE_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(1423)},
					{"EXTRACT(MINUTE_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(1423000000)},
					{"EXTRACT(SECOND_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(23000000)},
					{"EXTRACT(SQL_TSI_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')", evaluator.SQLInt(14)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: FLOOR", func() {
				tests := []test{
					{"FLOOR(NULL)", evaluator.SQLNull},
					{"FLOOR('sdg')", evaluator.SQLFloat(0)},
					{"FLOOR(1.23)", evaluator.SQLFloat(1)},
					{"FLOOR(-1.23)", evaluator.SQLFloat(-2)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: FROM_DAYS", func() {
				t1 := time.Date(0001, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
				t2 := time.Date(2000, 7, 3, 0, 0, 0, 0, schema.DefaultLocale)
				t3 := time.Date(10000, 3, 15, 0, 0, 0, 0, schema.DefaultLocale)
				t4 := time.Date(0005, 6, 29, 0, 0, 0, 0, schema.DefaultLocale)
				t5 := time.Date(2112, 1, 8, 0, 0, 0, 0, schema.DefaultLocale)

				tests := []test{
					{"FROM_DAYS(NULL)", evaluator.SQLNull},
					{"FROM_DAYS('sdg')", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(1.23)", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(-1.23)", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(-223.33)", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(223.33)", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(365.33)", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(3652499.5)", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(-771399.216)", evaluator.SQLVarchar("0000-00-00")},
					{"FROM_DAYS(365.93)", evaluator.SQLDate{t1}},
					{"FROM_DAYS(343+23)", evaluator.SQLDate{t1}},
					{"FROM_DAYS(730669)", evaluator.SQLDate{t2}},
					{"FROM_DAYS(3652499.3)", evaluator.SQLDate{t3}},
					{"FROM_DAYS('2006-05-11')", evaluator.SQLDate{t4}},
					{"FROM_DAYS(771399.216)", evaluator.SQLDate{t5}},
				}

				runTests(evalCtx, tests)
			})

			Convey("Subject: FROM_UNIXTIME", func() {
				tests := []test{
					{"FROM_UNIXTIME(NULL)", evaluator.SQLNull},
					{"FROM_UNIXTIME(-1)", evaluator.SQLNull},
					{"FROM_UNIXTIME(1447430881) + 0", evaluator.SQLDecimal128(decimal.New(20151113160801, 0))},
					{"FROM_UNIXTIME(1447430881.5) + 0", evaluator.SQLDecimal128(decimal.New(20151113160802, 0))},
					{"CONCAT(FROM_UNIXTIME(1447430881), '')", evaluator.SQLVarchar("2015-11-13 16:08:01.000000")},
					{"CONCAT(FROM_UNIXTIME(1447430881.5), '')", evaluator.SQLVarchar("2015-11-13 16:08:02.000000")},
				}

				runTests(evalCtx, tests)
			})

			Convey("Subject: GREATEST", func() {
				d, err := time.Parse("2006-01-02", "2006-05-11")
				So(err, ShouldBeNil)
				t, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 12:32:23")
				So(err, ShouldBeNil)

				tests := []test{
					{"GREATEST(NULL, 1, 2)", evaluator.SQLNull},
					{"GREATEST(1,3,2)", evaluator.SQLInt(3)},
					{"GREATEST(2,2.3)", evaluator.SQLDecimal128(decimal.New(23, -1))},
					{"GREATEST('cats', '4', '2')", evaluator.SQLVarchar("cats")},
					{"GREATEST('dog', 'cats', 'bird')", evaluator.SQLVarchar("dog")},
					{"GREATEST('cat', 'bird', 2)", evaluator.SQLInt(2)},
					{"GREATEST('cat', 2.2)", evaluator.SQLDecimal128(decimal.New(22, -1))},
					{"GREATEST(false, true)", evaluator.SQLTrue},
					{"GREATEST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')", evaluator.SQLDate{Time: d}},
					{"GREATEST(DATE '2006-05-11', 14, 4235)", evaluator.SQLInt(20060511)},
					{"GREATEST(DATE '2006-05-11', 14, 20080622)", evaluator.SQLInt(20080622)},
					{"GREATEST(DATE '2006-05-11', 14, 20080622.1)", evaluator.SQLDecimal128(decimal.New(200806221, -1))},
					{"GREATEST(DATE '2006-05-11', 14, 4235.2)", evaluator.SQLDecimal128(decimal.New(20060511, 0))},
					{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')", evaluator.SQLTimestamp{Time: t}},
					{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)", evaluator.SQLInt(20060511123223)},
					{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)", evaluator.SQLDecimal128(decimal.New(200809231243453, -1))},
					{"GREATEST(DATE '2006-05-11', 'cat', '2007-04-11')", evaluator.SQLVarchar("2007-04-11")},
					{"GREATEST(DATE '2006-05-11', 20080912, '2007-04-11')", evaluator.SQLInt(20080912)},
					{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:45')", evaluator.SQLTimestamp{Time: t}},
					{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')", evaluator.SQLInt(20060511123223)},
					{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2008-09-13')", evaluator.SQLVarchar("2008-09-13")},
					{"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')", evaluator.SQLTimestamp{Time: t}},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"GREATEST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')", schema.SQLDate},
					{"GREATEST(1, 123.52, 'something')", schema.SQLDecimal128},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: HOUR", func() {
				tests := []test{
					{"HOUR(NULL)", evaluator.SQLNull},
					{"HOUR('sdg')", evaluator.SQLInt(0)},
					{"HOUR('10:23:52')", evaluator.SQLInt(10)},
					{"HOUR('10:61:52')", evaluator.SQLNull},
					{"HOUR('10:23:52.23.25.26')", evaluator.SQLInt(10)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: IF", func() {
				tests := []test{
					{"IF(1<2, 4, 5)", evaluator.SQLInt(4)},
					{"IF(1>2, 4, 5)", evaluator.SQLInt(5)},
					{"IF(1, 4, 5)", evaluator.SQLInt(4)},
					{"IF(-0, 4, 5)", evaluator.SQLInt(5)},
					{"IF(1-1, 4, 5)", evaluator.SQLInt(5)},
					{"IF('cat', 4, 5)", evaluator.SQLInt(5)},
					{"IF('3', 4, 5)", evaluator.SQLInt(4)},
					{"IF('0', 4, 5)", evaluator.SQLInt(5)},
					{"IF('-0.0', 4, 5)", evaluator.SQLInt(5)},
					{"IF('2.2', 4, 5)", evaluator.SQLInt(4)},
					{"IF('true', 4, 5)", evaluator.SQLInt(5)},
					{"IF(null, 4, 'cat')", evaluator.SQLVarchar("cat")},
					{"IF(true, 'dog', 'cat')", evaluator.SQLVarchar("dog")},
					{"IF(false, 'dog', 'cat')", evaluator.SQLVarchar("cat")},
					{"IF('ca.gh', 4, 5)", evaluator.SQLInt(5)},
					{"IF(current_timestamp(), 4, 5)", evaluator.SQLInt(4)}, // not being parsed as dates, being parsed as string
					{"IF(current_timestamp, 4, 5)", evaluator.SQLInt(4)},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"IF('ca.gh', 4, 5)", schema.SQLInt},
					{"IF('ca.gh', 4, 5.3)", schema.SQLDecimal128},
					{"IF('ca.gh', 'sdf', 5.2)", schema.SQLVarchar},
					{"IF('ca.gh', 'sdf', NULL)", schema.SQLVarchar},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: IFNULL", func() {
				tests := []test{
					{"IFNULL(1,0)", evaluator.SQLInt(1)},
					{"IFNULL(NULL,3)", evaluator.SQLInt(3)},
					{"IFNULL(NULL,NULL)", evaluator.SQLNull},
					{"IFNULL('cat', null)", evaluator.SQLVarchar("cat")},
					{"IFNULL(null, 'dog')", evaluator.SQLVarchar("dog")},
					{"IFNULL(1/0, 4)", evaluator.SQLInt(4)},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"IFNULL(4, 5)", schema.SQLInt},
					{"IFNULL(4, 5.3)", schema.SQLDecimal128},
					{"IFNULL('sdf', NULL)", schema.SQLVarchar},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: INTERVAL", func() {
				tests := []test{
					{"INTERVAL(1,0)", evaluator.SQLInt(1)},
					{"INTERVAL(NULL, 3)", evaluator.SQLInt(-1)},
					{"INTERVAL(NULL, NULL)", evaluator.SQLInt(-1)},
					{"INTERVAL(2, 1, 2, 3, 4)", evaluator.SQLInt(2)},
					{"INTERVAL('1.1', 0, 1.1, 2)", evaluator.SQLInt(2)},
					{"INTERVAL(-1, NULL, 4)", evaluator.SQLInt(1)},
					{"INTERVAL(4, 1, 2, 4)", evaluator.SQLInt(3)},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"INTERVAL(4, 5)", schema.SQLInt64},
					{"INTERVAL(4, 5.3)", schema.SQLInt64},
					{"INTERVAL(NULL, 4)", schema.SQLInt64},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: ISNULL", func() {
				tests := []test{
					{"ISNULL(a)", evaluator.SQLBool(0)},
					{"ISNULL(c)", evaluator.SQLBool(1)},
					{`ISNULL("")`, evaluator.SQLBool(0)},
					{`ISNULL(NULL)`, evaluator.SQLBool(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: INSERT", func() {
				tests := []test{
					{"INSERT('Quadratic', NULL, 4, 'What')", evaluator.SQLNull},
					{"INSERT('Quadratic', 3, 4, 'What')", evaluator.SQLVarchar("QuWhattic")},
					{"INSERT('Quadratic', -1, 4, 'What')", evaluator.SQLVarchar("Quadratic")},
					{"INSERT('Quadratic', 3, 100, 'What')", evaluator.SQLVarchar("QuWhat")},
					{"INSERT('Quadratic', 9, 4, 'What')", evaluator.SQLVarchar("QuadratiWhat")},
					{"INSERT('Quadratic', 8.5, 3.5, 'What')", evaluator.SQLVarchar("QuadratiWhat")},
					{"INSERT('Quadratic', 8.4, 3.4, 'What')", evaluator.SQLVarchar("QuadratWhat")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: INSTR", func() {
				tests := []test{
					{"INSTR(NULL, NULL)", evaluator.SQLNull},
					{"INSTR('sDg', 'D')", evaluator.SQLInt(2)},
					{"INSTR(124, 124)", evaluator.SQLInt(1)},
					{"INSTR('awesome','so')", evaluator.SQLInt(4)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LAST_DAY", func() {
				d1, err := time.Parse("2006-01-02", "2003-02-28")
				So(err, ShouldBeNil)
				d2, err := time.Parse("2006-01-02", "2004-02-29")
				So(err, ShouldBeNil)
				d3, err := time.Parse("2006-01-02", "2004-01-31")
				So(err, ShouldBeNil)

				tests := []test{
					{"LAST_DAY('')", evaluator.SQLNull},
					{"LAST_DAY(NULL)", evaluator.SQLNull},
					{"LAST_DAY('2003-03-32')", evaluator.SQLNull},
					{"LAST_DAY('2003-02-05')", evaluator.SQLDate{d1}},
					{"LAST_DAY('2004-02-05')", evaluator.SQLDate{d2}},
					{"LAST_DAY('2004-01-01 01:01:01')", evaluator.SQLDate{d3}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LCASE", func() {
				tests := []test{
					{"LCASE(NULL)", evaluator.SQLNull},
					{"LCASE('sDg')", evaluator.SQLVarchar("sdg")},
					{"LCASE(124)", evaluator.SQLVarchar("124")},
					{"LOWER(NULL)", evaluator.SQLNull},
					{"LOWER('')", evaluator.SQLVarchar("")},
					{"LOWER('A')", evaluator.SQLVarchar("a")},
					{"LOWER('awesome')", evaluator.SQLVarchar("awesome")},
					{"LOWER('AwEsOmE')", evaluator.SQLVarchar("awesome")},
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
					{"LEAST(NULL, 1, 2)", evaluator.SQLNull},
					{"LEAST(1,3,2)", evaluator.SQLInt(1)},
					{"LEAST(2,2.3)", evaluator.SQLDecimal128(decimal.New(2, 0))},
					{"LEAST('cats', '4', '2')", evaluator.SQLVarchar("2")},
					{"LEAST('dog', 'cats', 'bird')", evaluator.SQLVarchar("bird")},
					{"LEAST(false, true)", evaluator.SQLFalse},
					{"LEAST(DATE '2005-05-11', DATE '2006-05-11', DATE '2007-05-11')", evaluator.SQLDate{Time: d}},
					{"LEAST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')", evaluator.SQLTimestamp{Time: t}},
					{"LEAST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:23')", evaluator.SQLTimestamp{Time: t1}},
					{"LEAST('cat', 'bird', 2)", evaluator.SQLInt(0)},
					{"LEAST('cat', 2.2)", evaluator.SQLDecimal128(decimal.Zero)},
					{"LEAST(DATE '2006-05-11', 14, 4235)", evaluator.SQLInt(14)},
					{"LEAST(DATE '2006-05-11', 14, 20080622.1)", evaluator.SQLDecimal128(decimal.New(14, 0))},
					{"LEAST(DATE '2006-05-11', 14, 4235.2)", evaluator.SQLDecimal128(decimal.New(14, 0))},
					{"LEAST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)", evaluator.SQLInt(12)},
					{"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)", evaluator.SQLDecimal128(decimal.New(20060511123223, 0))},
					{"LEAST(DATE '2006-05-11', 'cat', '2007-04-11')", evaluator.SQLVarchar("cat")},
					{"LEAST(DATE '2006-05-11', 20080912, '2007-04-11')", evaluator.SQLInt(0)},
					{"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')", evaluator.SQLInt(20070823)},
					{"LEAST(TIMESTAMP '2006-05-11 10:32:23', '2008-09-13')", evaluator.SQLTimestamp{Time: t1}},
					{"LEAST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')", evaluator.SQLVarchar("2005-09-13")},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"LEAST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')", schema.SQLDate},
					{"LEAST(1, 123.52, 'something')", schema.SQLDecimal128},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: LEFT", func() {
				tests := []test{

					// if any argument null, should return null
					{"LEFT(NULL, NULL)", evaluator.SQLNull},
					{"LEFT('hi', NULL)", evaluator.SQLNull},
					{"LEFT(NULL, 5)", evaluator.SQLNull},

					// basic cases w/ string, int inputs and positive int length
					{"LEFT('sDgcdcdc', 4)", evaluator.SQLVarchar("sDgc")},
					{"LEFT(124, 2)", evaluator.SQLVarchar("12")},

					// negative lengths and 0 give empty string
					{"LEFT('hi', -1)", evaluator.SQLVarchar("")},
					{"LEFT('hi', 0)", evaluator.SQLVarchar("")},
					{"LEFT('hi', -2.5)", evaluator.SQLVarchar("")},

					// float lengths should be rounded to closest int
					{"LEFT('hello', 2.4)", evaluator.SQLVarchar("he")},
					{"LEFT('hello', 2.5)", evaluator.SQLVarchar("hel")},
					{"LEFT(1234, 2.3)", evaluator.SQLVarchar("12")},
					{"LEFT(1234, 2.5)", evaluator.SQLVarchar("123")},
					{"LEFT('yo', 2.5)", evaluator.SQLVarchar("yo")},

					// strings with spaces and symbols
					{"LEFT('  ', 1)", evaluator.SQLVarchar(" ")},
					{"LEFT('@!%', 2)", evaluator.SQLVarchar("@!")},

					// boolean for string
					{"LEFT(true, 3)", evaluator.SQLVarchar("1")},
					{"LEFT(false, 3)", evaluator.SQLVarchar("0")},

					// boolean for length
					{"LEFT('hello', true)", evaluator.SQLVarchar("h")},
					{"LEFT('hello', false)", evaluator.SQLVarchar("")},

					// string for length
					{"LEFT('hello', 'hi')", evaluator.SQLVarchar("")},

					// len > length of string
					{"LEFT('hi', 5)", evaluator.SQLVarchar("hi")},

					// string number as length
					{"LEFT('hello', '2')", evaluator.SQLVarchar("he")},
					{"LEFT('hello', '-3')", evaluator.SQLVarchar("")},

					// unlike with floats, string #s always round down
					{"LEFT('hello', '2.4')", evaluator.SQLVarchar("he")},
					{"LEFT('hello', '2.6')", evaluator.SQLVarchar("he")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LENGTH", func() {
				tests := []test{
					{"LENGTH(NULL)", evaluator.SQLNull},
					{"LENGTH('sDg')", evaluator.SQLInt(3)},
					{"LENGTH('世界')", evaluator.SQLInt(6)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LN", func() {
				tests := []test{
					{"LN(NULL)", evaluator.SQLNull},
					{"LN(1)", evaluator.SQLFloat(0)},
					{"LN(16.5)", evaluator.SQLFloat(2.803360380906535)},
					{"LN(-16.5)", evaluator.SQLNull},
					{"LOG(NULL)", evaluator.SQLNull},
					{"LOG(1)", evaluator.SQLFloat(0)},
					{"LOG(16.5)", evaluator.SQLFloat(2.803360380906535)},
					{"LOG(-16.5)", evaluator.SQLNull},
					{"LOG10(100)", evaluator.SQLFloat(2)},
					{"LOG(10,100)", evaluator.SQLFloat(2)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOCATE", func() {
				tests := []test{
					{"LOCATE(NULL, 'foobarbar')", evaluator.SQLNull},
					{"LOCATE('bar', NULL)", evaluator.SQLNull},
					{"LOCATE('bar', 'foobarbar')", evaluator.SQLInt(4)},
					{"LOCATE('xbar', 'foobarbar')", evaluator.SQLInt(0)},
					{"LOCATE('bar', 'foobarbar', 5)", evaluator.SQLInt(7)},
					{"LOCATE('bar', 'foobarbar', 4)", evaluator.SQLInt(4)},
					{"LOCATE('e', 'dvd', 6)", evaluator.SQLInt(0)},
					{"LOCATE('f', 'asdf', 4)", evaluator.SQLInt(4)},
					{"LOCATE('語', '日本語')", evaluator.SQLInt(3)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOG2", func() {
				tests := []test{
					{"LOG2(NULL)", evaluator.SQLNull},
					{"LOG2(4)", evaluator.SQLFloat(2)},
					{"LOG2(-100)", evaluator.SQLNull},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LOG10", func() {
				tests := []test{
					{"LOG10(NULL)", evaluator.SQLNull},
					{"LOG10('sdg')", evaluator.SQLNull},
					{"LOG10(2)", evaluator.SQLFloat(0.3010299956639812)},
					{"LOG10(100)", evaluator.SQLFloat(2)},
					{"LOG10(0)", evaluator.SQLNull},
					{"LOG10(-100)", evaluator.SQLNull},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LTRIM", func() {
				tests := []test{
					{"LTRIM(NULL)", evaluator.SQLNull},
					{"LTRIM('   barbar')", evaluator.SQLVarchar("barbar")},
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
				d3, err := time.Parse("2006-01-02", "0100-02-01")
				So(err, ShouldBeNil)

				tests := []test{
					{"MAKEDATE(NULL, 4)", evaluator.SQLNull},
					{"MAKEDATE(2004, 0)", evaluator.SQLNull},
					{"MAKEDATE(9999, 370)", evaluator.SQLNull},
					{"MAKEDATE('sdg', 32)", evaluator.SQLDate{Time: d}},
					{"MAKEDATE('2000.9', 32)", evaluator.SQLDate{Time: d}},
					{"MAKEDATE(1999.5, 32)", evaluator.SQLDate{Time: d}},
					{"MAKEDATE('2000.9', '32.9')", evaluator.SQLDate{Time: d}},
					{"MAKEDATE(1999.5, 31.5)", evaluator.SQLDate{Time: d}},
					{"MAKEDATE(2000, 32)", evaluator.SQLDate{Time: d}},
					{"MAKEDATE(12, 32)", evaluator.SQLDate{Time: d1}},
					{"MAKEDATE(77, 66)", evaluator.SQLDate{Time: d2}},
					{"MAKEDATE(99.5, 31.5)", evaluator.SQLDate{Time: d3}},
					{"MAKEDATE('100.9', '32.5')", evaluator.SQLDate{Time: d3}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MD5", func() {
				tests := []test{
					{"MD5(NULL)", evaluator.SQLNull},
					{"MD5(NULL + NULL)", evaluator.SQLNull},
					{"MD5('testing')", evaluator.SQLVarchar("ae2b1fca515949e5d54fb22b8ed95575")},
					{"MD5('hello')", evaluator.SQLVarchar("5d41402abc4b2a76b9719d911017c592")},
					{"MD5(12)", evaluator.SQLVarchar("c20ad4d76fe97759aa27a0c99bff6710")},
					{"MD5(6.23)", evaluator.SQLVarchar("fec8db978f6b7844b09d9bd54fb8335c")},
					{"MD5('12:STR.002234')", evaluator.SQLVarchar("81d56d5aeb92a55298af2f091e86ab61")},
					{"MD5(REPEAT('a', 30))", evaluator.SQLVarchar("59e794d45697b360e18ba972bada0123")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MICROSECOND", func() {
				tests := []test{
					{"MICROSECOND(NULL)", evaluator.SQLNull},
					{"MICROSECOND('')", evaluator.SQLNull},
					{"MICROSECOND('NULL')", evaluator.SQLInt(0)},
					{"MICROSECOND('hello')", evaluator.SQLInt(0)},
					{"MICROSECOND(TRUE)", evaluator.SQLInt(0)},
					{"MICROSECOND('true')", evaluator.SQLInt(0)},
					{"MICROSECOND('FALSE')", evaluator.SQLInt(0)},
					{"MICROSECOND('11:38:24')", evaluator.SQLInt(0)},
					{"MICROSECOND('11:38')", evaluator.SQLInt(0)},
					{"MICROSECOND('11 38 24')", evaluator.SQLInt(0)},
					{"MICROSECOND('11:38:24.000000')", evaluator.SQLInt(0)},
					{"MICROSECOND('11:38:24.000001')", evaluator.SQLInt(1)},
					{"MICROSECOND('11:38:24.123456')", evaluator.SQLInt(123456)},
					{"MICROSECOND('1978-9-22 1:58:59')", evaluator.SQLInt(0)},
					{"MICROSECOND('1978-9-22 1:58:59.00001')", evaluator.SQLInt(10)},
					{"MICROSECOND('1978-9-22 1:58:59.0000104')", evaluator.SQLInt(10)},
					{"MICROSECOND('12:STUFF.002234')", evaluator.SQLInt(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MID", func() {
				tests := []test{
					{"MID('foobarbar', 4, NULL)", evaluator.SQLNull},
					{"MID('Quadratically', 5, 6)", evaluator.SQLVarchar("ratica")},
					{"MID('Quadratically', 12, 2)", evaluator.SQLVarchar("ly")},
					{"MID('Sakila', -5, 3)", evaluator.SQLVarchar("aki")},
					{"MID('日本語', 2, 1)", evaluator.SQLVarchar("本")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MINUTE", func() {
				tests := []test{
					{"MINUTE(NULL)", evaluator.SQLNull},
					{"MINUTE('sdg')", evaluator.SQLInt(0)},
					{"MINUTE('10:23:52')", evaluator.SQLInt(23)},
					{"MINUTE('10:61:52')", evaluator.SQLNull},
					{"MINUTE('10:23:52.25.26.27.28')", evaluator.SQLInt(23)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MOD", func() {
				tests := []test{
					{"MOD(NULL, NULL)", evaluator.SQLNull},
					{"MOD(234, NULL)", evaluator.SQLNull},
					{"MOD(NULL, 10)", evaluator.SQLNull},
					{"MOD(234, 0)", evaluator.SQLNull},
					{"MOD(234, 10)", evaluator.SQLFloat(4)},
					{"MOD(253, 7)", evaluator.SQLFloat(1)},
					{"MOD(34.5, 3)", evaluator.SQLFloat(1.5)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MONTH", func() {
				tests := []test{
					{"MONTH(NULL)", evaluator.SQLNull},
					{"MONTH('sdg')", evaluator.SQLNull},
					{"MONTH('2016-1-01 10:23:52')", evaluator.SQLInt(1)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: MONTHNAME", func() {
				tests := []test{
					{"MONTHNAME(NULL)", evaluator.SQLNull},
					{"MONTHNAME('sdg')", evaluator.SQLNull},
					{"MONTHNAME('2016-1-01 10:23:52')", evaluator.SQLVarchar("January")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: NULLIF", func() {
				tests := []test{
					{"NULLIF(1,1)", evaluator.SQLNull},
					{"NULLIF(1,3)", evaluator.SQLInt(1)},
					{"NULLIF(null, null)", evaluator.SQLNull},
					{"NULLIF(null, 4)", evaluator.SQLNull},
					{"NULLIF(3, null)", evaluator.SQLInt(3)},
					//test{"NULLIF(3, '3')", evaluator.SQLNull},
					{"NULLIF('abc', 'abc')", evaluator.SQLNull},
					//test{"NULLIF('abc', 3)", evaluator.SQLVarchar("abc")},
					//test{"NULLIF('1', true)", evaluator.SQLNull},
					//test{"NULLIF('1', false)", evaluator.SQLVarchar("1")},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"NULLIF(3, null)", schema.SQLInt},
					{"NULLIF('abc', 'abc')", schema.SQLVarchar},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: PI", func() {
				tests := []test{
					{"PI()", evaluator.SQLFloat(3.141592653589793116)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: QUARTER", func() {
				tests := []test{
					{"QUARTER(NULL)", evaluator.SQLNull},
					{"QUARTER('sdg')", evaluator.SQLNull},
					{"QUARTER('2016-1-01 10:23:52')", evaluator.SQLInt(1)},
					{"QUARTER('2016-4-01 10:23:52')", evaluator.SQLInt(2)},
					{"QUARTER('2016-8-01 10:23:52')", evaluator.SQLInt(3)},
					{"QUARTER('2016-11-01 10:23:52')", evaluator.SQLInt(4)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: RADIANS", func() {
				tests := []test{
					{"RADIANS(NULL)", evaluator.SQLNull},
					{"RADIANS(1145.9155902616465)", evaluator.SQLFloat(20)},
					{"RADIANS(-1145.9155902616465)", evaluator.SQLFloat(-20)},
				}
				runTests(evalCtx, tests)
			})

			// NOTE: These values could change for different versions of Go.
			Convey("Subject: RAND", func() {
				tests := []test{
					{"RAND(NULL)", evaluator.SQLFloat(0.9451961492941164)},
					{"RAND('hello')", evaluator.SQLFloat(0.9451961492941164)},
					{"RAND(0)", evaluator.SQLFloat(0.9451961492941164)},
					{"RAND(1145.9155902616465)", evaluator.SQLFloat(0.16758646518059656)},
					{"RAND(-1145.9155902616465)", evaluator.SQLFloat(0.8321372077808122)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: REPEAT", func() {
				tests := []test{
					{"REPEAT(NULL, NULL)", evaluator.SQLNull},
					{"REPEAT(NULL, 3)", evaluator.SQLNull},
					{"REPEAT('apples', NULL)", evaluator.SQLNull},
					{"REPEAT('apples', -1)", evaluator.SQLVarchar("")},
					{"REPEAT('apples', 0)", evaluator.SQLVarchar("")},
					{"REPEAT('apples', 1)", evaluator.SQLVarchar("apples")},
					{"REPEAT('a', 5)", evaluator.SQLVarchar("aaaaa")},
					{"REPEAT(3, 5)", evaluator.SQLVarchar("33333")},
					{"REPEAT(FALSE, 5)", evaluator.SQLVarchar("00000")},
					{"REPEAT(FALSE, TRUE)", evaluator.SQLVarchar("0")},
					{"REPEAT('', 10)", evaluator.SQLVarchar("")},
					{"REPEAT(0, '4')", evaluator.SQLVarchar("0000")},
					{"REPEAT(NULL, 4)", evaluator.SQLNull},
					{"REPEAT(1.4, 3)", evaluator.SQLVarchar("1.41.41.4")},
					{"REPEAT('a', .3)", evaluator.SQLVarchar("")},
					{"REPEAT('a', 3.2)", evaluator.SQLVarchar("aaa")},
					{"REPEAT('a', 3.6)", evaluator.SQLVarchar("aaaa")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: REPLACE", func() {
				tests := []test{
					{"REPLACE(NULL, NULL, NULL)", evaluator.SQLNull},
					{"REPLACE('sDgcdcdc', 'D', 'd')", evaluator.SQLVarchar("sdgcdcdc")},
					{"REPLACE('www.mysql.com', 'w', 'Ww')", evaluator.SQLVarchar("WwWwWw.mysql.com")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: REVERSE", func() {
				tests := []test{
					{"REVERSE(NULL)", evaluator.SQLNull},
					{"REVERSE(3.14159265)", evaluator.SQLVarchar("56295141.3")},
					{"REVERSE(655)", evaluator.SQLVarchar("556")},
					{"REVERSE('www.mysql.com')", evaluator.SQLVarchar("moc.lqsym.www")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: RIGHT", func() {
				tests := []test{
					// if any argument null, should return null
					{"RIGHT(NULL, NULL)", evaluator.SQLNull},
					{"RIGHT('hi', NULL)", evaluator.SQLNull},
					{"RIGHT(NULL, 5)", evaluator.SQLNull},

					// basic cases w/ string, int inputs and positive int length
					{"RIGHT('sDgcdcdc', 4)", evaluator.SQLVarchar("dcdc")},
					{"RIGHT(124, 2)", evaluator.SQLVarchar("24")},

					// negative lengths and 0 give empty string
					{"RIGHT('hi', -1)", evaluator.SQLVarchar("")},
					{"RIGHT('hi', 0)", evaluator.SQLVarchar("")},
					{"RIGHT('hi', -2.5)", evaluator.SQLVarchar("")},

					// float lengths should be rounded to closest int
					{"RIGHT('hello', 2.4)", evaluator.SQLVarchar("lo")},
					{"RIGHT('hello', 2.5)", evaluator.SQLVarchar("llo")},
					{"RIGHT(1234, 2.3)", evaluator.SQLVarchar("34")},
					{"RIGHT(1234, 2.5)", evaluator.SQLVarchar("234")},
					{"RIGHT('yo', 2.5)", evaluator.SQLVarchar("yo")},

					// strings with spaces and symbols
					{"RIGHT('  ', 1)", evaluator.SQLVarchar(" ")},
					{"RIGHT('@!%', 2)", evaluator.SQLVarchar("!%")},

					// boolean for string
					{"RIGHT(true, 3)", evaluator.SQLVarchar("1")},
					{"RIGHT(false, 3)", evaluator.SQLVarchar("0")},

					// boolean for length
					{"RIGHT('hello', true)", evaluator.SQLVarchar("o")},
					{"RIGHT('hello', false)", evaluator.SQLVarchar("")},

					// string for length
					{"RIGHT('hello', 'hi')", evaluator.SQLVarchar("")},

					// len > length of string
					{"RIGHT('hi', 5)", evaluator.SQLVarchar("hi")},

					// string number as length
					{"RIGHT('hello', '2')", evaluator.SQLVarchar("lo")},
					{"RIGHT('hello', '-3')", evaluator.SQLVarchar("")},

					// unlike with floats, string #s always round down
					{"RIGHT('hello', '2.4')", evaluator.SQLVarchar("lo")},
					{"RIGHT('hello', '2.6')", evaluator.SQLVarchar("lo")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: ROUND", func() {
				tests := []test{
					{"ROUND(NULL, NULL)", evaluator.SQLNull},
					{"ROUND(NULL, 4)", evaluator.SQLNull},
					{"ROUND(-16.55555, 4)", evaluator.SQLFloat(-16.5556)},
					{"ROUND(4.56, 1)", evaluator.SQLFloat(4.6)},
					{"ROUND(-16.5, -1)", evaluator.SQLFloat(0)},
					{"ROUND(-16.5)", evaluator.SQLFloat(-17)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: RTRIM", func() {
				tests := []test{
					{"RTRIM(NULL)", evaluator.SQLNull},
					{"RTRIM('barbar   ')", evaluator.SQLVarchar("barbar")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LPAD", func() {
				tests := []test{

					// LPAD(str, len, padStr)

					// basic case
					{"LPAD('hello', 7, 'x')", evaluator.SQLVarchar("xxhello")},

					// nulls in various positions
					{"LPAD(NULL, 5, 'a')", evaluator.SQLNull},
					{"LPAD('hi', NULL, 'a')", evaluator.SQLNull},
					{"LPAD('hi', 5, NULL)", evaluator.SQLNull},
					{"LPAD(NULL, NULL, NULL)", evaluator.SQLNull},

					// str: empty
					{"LPAD('', 0, 'a')", evaluator.SQLVarchar("")},
					{"LPAD('', 1, 'a')", evaluator.SQLVarchar("a")},
					{"LPAD('', 7, 'ab')", evaluator.SQLVarchar("abababa")},

					// str: spaces and symbols
					{"LPAD(' hi', 4, 'x')", evaluator.SQLVarchar("x hi")},
					{"LPAD('  ', 5, ' ')", evaluator.SQLVarchar("     ")},
					{"LPAD('@!#_', 10, '.')", evaluator.SQLVarchar("......@!#_")},
					{"LPAD('I♥NY', 8, 'x')", evaluator.SQLVarchar("xxxxI♥NY")},
					{"LPAD('ƏŨ Ó€', 8, 'x')", evaluator.SQLVarchar("xxxƏŨ Ó€")},
					{"LPAD('⅓ ⅔ † ‡ µ ¢ £', 8, 'x')", evaluator.SQLVarchar("⅓ ⅔ † ‡ ")},
					{"LPAD('∞π∅≤≥≠≈', 8, 'x')", evaluator.SQLVarchar("x∞π∅≤≥≠≈")},
					{"LPAD('hello', 8, '♥')", evaluator.SQLVarchar("♥♥♥hello")},
					{"LPAD('hello', 8, 'ƏŨ')", evaluator.SQLVarchar("ƏŨƏhello")},

					// str type: numbers
					{"LPAD(5, 4, 'a')", evaluator.SQLVarchar("aaa5")},
					{"LPAD(10, 4, 'a')", evaluator.SQLVarchar("aa10")},
					{"LPAD(10.2, 4, 'a')", evaluator.SQLVarchar("10.2")},

					// str type: boolean
					{"LPAD(true, 4, 'a')", evaluator.SQLVarchar("aaa1")},
					{"LPAD(false, 4, 'a')", evaluator.SQLVarchar("aaa0")},

					// len < 0
					{"LPAD('hi', -1, 'a')", evaluator.SQLNull},

					// len = 0
					{"LPAD('hi', 0, 'a')", evaluator.SQLVarchar("")},

					// len <= len(str)
					{"LPAD('hello', 2, 'x')", evaluator.SQLVarchar("he")},
					{"LPAD('hello', 5, 'x')", evaluator.SQLVarchar("hello")},

					// len type: str
					{"LPAD('hello', '5', 'x')", evaluator.SQLVarchar("hello")},
					{"LPAD('hello', '5.6', 'x')", evaluator.SQLVarchar("hello")},
					{"LPAD('hello', '6', 'x')", evaluator.SQLVarchar("xhello")},
					{"LPAD('hello', '6.2', 'x')", evaluator.SQLVarchar("xhello")},
					// if can't be cast to #, then use length 0
					{"LPAD('hello', 'a', 'x')", evaluator.SQLVarchar("")},

					// len: floating point
					{"LPAD('hello', 5.4, 'x')", evaluator.SQLVarchar("hello")},
					{"LPAD('hello', 5.5, 'x')", evaluator.SQLVarchar("xhello")},

					// len float values close to 0 - round to closest int
					{"LPAD('hello', 0.4, 'x')", evaluator.SQLVarchar("")},
					{"LPAD('hello', 0.5, 'x')", evaluator.SQLVarchar("h")},
					{"LPAD('hello', -0.4, 'x')", evaluator.SQLVarchar("")},
					{"LPAD('hello', -0.5, 'x')", evaluator.SQLNull},

					// len string values close to 0 - always round toward 0
					{"LPAD('hello', '0.4', 'x')", evaluator.SQLVarchar("")},
					{"LPAD('hello', '0.5', 'x')", evaluator.SQLVarchar("")},
					{"LPAD('hello', '-0.4', 'x')", evaluator.SQLVarchar("")},
					{"LPAD('hello', '-0.5', 'x')", evaluator.SQLVarchar("")},

					// len: bool
					{"LPAD('hello', true, 'x')", evaluator.SQLVarchar("h")},
					{"LPAD('hello', false, 'x')", evaluator.SQLVarchar("")},

					// len(padStr) > 1
					{"LPAD('hello', 7, 'xy')", evaluator.SQLVarchar("xyhello")},
					{"LPAD('hello', 8, 'xy')", evaluator.SQLVarchar("xyxhello")},

					// padStr type: number
					{"LPAD('hello', 7, 1)", evaluator.SQLVarchar("11hello")},
					{"LPAD('hello', 10, 1.1)", evaluator.SQLVarchar("1.11.hello")},
					{"LPAD('hello', 10, -1)", evaluator.SQLVarchar("-1-1-hello")},

					// padStr type: boolean
					{"LPAD('hello', 7, true)", evaluator.SQLVarchar("11hello")},
					{"LPAD('hello', 10, false)", evaluator.SQLVarchar("00000hello")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: RPAD", func() {
				tests := []test{

					// RPAD(str, len, padStr)

					// basic case
					{"RPAD('hello', 7, 'x')", evaluator.SQLVarchar("helloxx")},

					// nulls in various positions
					{"RPAD(NULL, 5, 'a')", evaluator.SQLNull},
					{"RPAD('hi', NULL, 'a')", evaluator.SQLNull},
					{"RPAD('hi', 5, NULL)", evaluator.SQLNull},
					{"RPAD(NULL, NULL, NULL)", evaluator.SQLNull},

					// str: empty
					{"RPAD('', 0, 'a')", evaluator.SQLVarchar("")},
					{"RPAD('', 1, 'a')", evaluator.SQLVarchar("a")},
					{"RPAD('', 7, 'ab')", evaluator.SQLVarchar("abababa")},

					// str: spaces and symbols
					{"RPAD(' hi', 4, 'x')", evaluator.SQLVarchar(" hix")},
					{"RPAD('  ', 5, ' ')", evaluator.SQLVarchar("     ")},
					{"RPAD('@!#_', 10, '.')", evaluator.SQLVarchar("@!#_......")},
					{"RPAD('I♥NY', 8, 'x')", evaluator.SQLVarchar("I♥NYxxxx")},
					{"RPAD('ƏŨ Ó€', 8, 'x')", evaluator.SQLVarchar("ƏŨ Ó€xxx")},
					{"RPAD('⅓ ⅔ † ‡ µ ¢ £', 8, 'x')", evaluator.SQLVarchar("⅓ ⅔ † ‡ ")},
					{"RPAD('∞π∅≤≥≠≈', 8, 'x')", evaluator.SQLVarchar("∞π∅≤≥≠≈x")},
					{"RPAD('hello', 8, '♥')", evaluator.SQLVarchar("hello♥♥♥")},
					{"RPAD('hello', 8, 'ƏŨ')", evaluator.SQLVarchar("helloƏŨƏ")},

					// str type: numbers
					{"RPAD(5, 4, 'a')", evaluator.SQLVarchar("5aaa")},
					{"RPAD(10, 4, 'a')", evaluator.SQLVarchar("10aa")},
					{"RPAD(10.2, 4, 'a')", evaluator.SQLVarchar("10.2")},

					// str type: boolean
					{"RPAD(true, 4, 'a')", evaluator.SQLVarchar("1aaa")},
					{"RPAD(false, 4, 'a')", evaluator.SQLVarchar("0aaa")},

					// len < 0
					{"RPAD('hi', -1, 'a')", evaluator.SQLNull},

					// len = 0
					{"RPAD('hi', 0, 'a')", evaluator.SQLVarchar("")},

					// len <= len(str)
					{"RPAD('hello', 2, 'x')", evaluator.SQLVarchar("he")},
					{"RPAD('hello', 5, 'x')", evaluator.SQLVarchar("hello")},

					// len type: str
					{"RPAD('hello', '5', 'x')", evaluator.SQLVarchar("hello")},
					{"RPAD('hello', '5.6', 'x')", evaluator.SQLVarchar("hello")},
					{"RPAD('hello', '6', 'x')", evaluator.SQLVarchar("hellox")},
					{"RPAD('hello', '6.2', 'x')", evaluator.SQLVarchar("hellox")},
					// if can't be cast to #, then use length 0
					{"RPAD('hello', 'a', 'x')", evaluator.SQLVarchar("")},

					// len: floating point
					{"RPAD('hello', 5.4, 'x')", evaluator.SQLVarchar("hello")},
					{"RPAD('hello', 5.5, 'x')", evaluator.SQLVarchar("hellox")},

					// len float values close to 0 - round to closest int
					{"RPAD('hello', 0.4, 'x')", evaluator.SQLVarchar("")},
					{"RPAD('hello', 0.5, 'x')", evaluator.SQLVarchar("h")},
					{"RPAD('hello', -0.4, 'x')", evaluator.SQLVarchar("")},
					{"RPAD('hello', -0.5, 'x')", evaluator.SQLNull},

					// len string values close to 0 - always round toward 0
					{"RPAD('hello', '0.4', 'x')", evaluator.SQLVarchar("")},
					{"RPAD('hello', '0.5', 'x')", evaluator.SQLVarchar("")},
					{"RPAD('hello', '-0.4', 'x')", evaluator.SQLVarchar("")},
					{"RPAD('hello', '-0.5', 'x')", evaluator.SQLVarchar("")},

					// len: bool
					{"RPAD('hello', true, 'x')", evaluator.SQLVarchar("h")},
					{"RPAD('hello', false, 'x')", evaluator.SQLVarchar("")},

					// len(padStr) > 1
					{"RPAD('hello', 7, 'xy')", evaluator.SQLVarchar("helloxy")},
					{"RPAD('hello', 8, 'xy')", evaluator.SQLVarchar("helloxyx")},

					// padStr type: number
					{"RPAD('hello', 7, 1)", evaluator.SQLVarchar("hello11")},
					{"RPAD('hello', 10, 1.1)", evaluator.SQLVarchar("hello1.11.")},
					{"RPAD('hello', 10, -1)", evaluator.SQLVarchar("hello-1-1-")},

					// padStr type: boolean
					{"RPAD('hello', 7, true)", evaluator.SQLVarchar("hello11")},
					{"RPAD('hello', 10, false)", evaluator.SQLVarchar("hello00000")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SECOND", func() {
				tests := []test{
					{"SECOND(NULL)", evaluator.SQLNull},
					{"SECOND('sdg')", evaluator.SQLInt(0)},
					{"SECOND('10:23:52')", evaluator.SQLInt(52)},
					{"SECOND('10:61:52.24')", evaluator.SQLNull},
					{"SECOND('10:23:52.24.25.26.27')", evaluator.SQLInt(52)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SIGN", func() {
				tests := []test{
					{"SIGN(NULL)", evaluator.SQLNull},
					{"SIGN(-42)", evaluator.SQLInt(-1)},
					{"SIGN(0)", evaluator.SQLInt(0)},
					{"SIGN(42)", evaluator.SQLInt(1)},
					{"SIGN(42.0)", evaluator.SQLInt(1)},
					{"SIGN(-42.0)", evaluator.SQLInt(-1)},
					{"SIGN('hello world')", evaluator.SQLInt(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SIN", func() {
				tests := []test{
					{"SIN(NULL)", evaluator.SQLNull},
					{"SIN(19)", evaluator.SQLFloat(0.14987720966295234)},
					{"SIN(-19)", evaluator.SQLFloat(-0.14987720966295234)},
					{"SIN('C')", evaluator.SQLFloat(0)},
					{"SIN(0)", evaluator.SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SPACE", func() {
				tests := []test{
					{"SPACE(NULL)", evaluator.SQLNull},
					{"SPACE(5)", evaluator.SQLVarchar("     ")},
					{"SPACE(-3)", evaluator.SQLVarchar("")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SQRT", func() {
				tests := []test{
					{"SQRT(NULL)", evaluator.SQLNull},
					{"SQRT('sdg')", evaluator.SQLFloat(0)},
					{"SQRT(-16)", evaluator.SQLNull},
					{"SQRT(4)", evaluator.SQLFloat(2)},
					{"SQRT(20)", evaluator.SQLFloat(4.47213595499958)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SUBSTRING", func() {
				tests := []test{
					{"SUBSTRING(NULL, 4)", evaluator.SQLNull},
					{"SUBSTRING('foobarbar', NULL)", evaluator.SQLNull},
					{"SUBSTRING('foobarbar', 4, NULL)", evaluator.SQLNull},
					{"SUBSTRING('Quadratically', 5)", evaluator.SQLVarchar("ratically")},
					{"SUBSTRING('Quadratically', 5, 6)", evaluator.SQLVarchar("ratica")},
					{"SUBSTRING('Quadratically', 12, 2)", evaluator.SQLVarchar("ly")},
					{"SUBSTRING('Sakila', -3)", evaluator.SQLVarchar("ila")},
					{"SUBSTRING('Sakila', -5, 3)", evaluator.SQLVarchar("aki")},
					{"SUBSTRING('日本語', 2)", evaluator.SQLVarchar("本語")},
					{"SUBSTR(NULL, 4)", evaluator.SQLNull},
					{"SUBSTR('foobarbar', NULL)", evaluator.SQLNull},
					{"SUBSTR('foobarbar', 4, NULL)", evaluator.SQLNull},
					{"SUBSTR('Quadratically', 5)", evaluator.SQLVarchar("ratically")},
					{"SUBSTR('Quadratically', 5, 6)", evaluator.SQLVarchar("ratica")},
					{"SUBSTR('Sakila', -3)", evaluator.SQLVarchar("ila")},
					{"SUBSTR('Sakila', -5, 3)", evaluator.SQLVarchar("aki")},
					{"SUBSTR('日本語', 2)", evaluator.SQLVarchar("本語")},
					{"SUBSTR('five', 2, 2)", evaluator.SQLVarchar("iv")},
					{"SUBSTR('nine', 4, 9)", evaluator.SQLVarchar("e")},
					{"SUBSTR('five', 4, 3)", evaluator.SQLVarchar("e")},
					{"SUBSTR('five', -1, 1)", evaluator.SQLVarchar("e")},
					{"SUBSTR('five', 4, 0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA', 0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA', 0, 1)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA', 0, -1)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA', -1, 0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA', 1, 0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA', 0, 0)", evaluator.SQLVarchar("")},
					{"SUBSTRING(NULL from 4)", evaluator.SQLNull},
					{"SUBSTRING('foobarbar' from NULL)", evaluator.SQLNull},
					{"SUBSTRING('foobarbar' from 4 for NULL)", evaluator.SQLNull},
					{"SUBSTRING('Quadratically' FROM 5)", evaluator.SQLVarchar("ratically")},
					{"SUBSTRING('Quadratically' FROM  5 for 6)", evaluator.SQLVarchar("ratica")},
					{"SUBSTRING('Quadratically' from 12 FOR 2)", evaluator.SQLVarchar("ly")},
					{"SUBSTRING('Sakila' FROM -3)", evaluator.SQLVarchar("ila")},
					{"SUBSTRING('Sakila' from -5 for 3)", evaluator.SQLVarchar("aki")},
					{"SUBSTRING('日本語' FROM  2)", evaluator.SQLVarchar("本語")},
					{"SUBSTR(NULL FROM 4)", evaluator.SQLNull},
					{"SUBSTR('foobarbar' FROM NULL)", evaluator.SQLNull},
					{"SUBSTR('foobarbar' FROM 4 FOR NULL)", evaluator.SQLNull},
					{"SUBSTR('Quadratically' FROM  5)", evaluator.SQLVarchar("ratically")},
					{"SUBSTR('Quadratically' FROM  5 for 6)", evaluator.SQLVarchar("ratica")},
					{"SUBSTR('Sakila' from -3)", evaluator.SQLVarchar("ila")},
					{"SUBSTR('Sakila' from -5 for 3)", evaluator.SQLVarchar("aki")},
					{"SUBSTR('日本語' from 2)", evaluator.SQLVarchar("本語")},
					{"SUBSTR('five' from 2 for 2)", evaluator.SQLVarchar("iv")},
					{"SUBSTR('nine' from 4 for  9)", evaluator.SQLVarchar("e")},
					{"SUBSTR('five' FROM 4 FOR 3)", evaluator.SQLVarchar("e")},
					{"SUBSTR('five' FROM -1 FOR  1)", evaluator.SQLVarchar("e")},
					{"SUBSTR('five' FROM 4 FOR  0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA' FROM 0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA' FROM 0 FOR  1)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA' FROM 0 for  -1)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA' from -1 for  0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA' from 1 FOR 0)", evaluator.SQLVarchar("")},
					{"SUBSTR('ZBA' from 0 for 0)", evaluator.SQLVarchar("")},
					{"SUBSTR('this', -5.2)", evaluator.SQLVarchar("")},
					{"SUBSTR('this' from -5.2)", evaluator.SQLVarchar("")},
					{"SUBSTR('this', 2.632)", evaluator.SQLVarchar("is")},
					{"SUBSTR('this', '2.632')", evaluator.SQLVarchar("his")},
					{"SUBSTR('this', '2.1')", evaluator.SQLVarchar("his")},
					{"SUBSTR('this' from -2.632)", evaluator.SQLVarchar("his")},
					{"SUBSTR('this', 2.4, 1.4)", evaluator.SQLVarchar("h")},
					{"SUBSTR('this' from 2.4 for -1.4 )", evaluator.SQLVarchar("")},
					{"SUBSTR('this', 1.6, 2.6)", evaluator.SQLVarchar("his")},
					{"SUBSTR('this', 1.6, '2.6')", evaluator.SQLVarchar("hi")},
					{"SUBSTR('this', 1.6, '2.1')", evaluator.SQLVarchar("hi")},
					{"SUBSTR('this', -11.6)", evaluator.SQLVarchar("")},
					{"SUBSTR(NULL, -4)", evaluator.SQLNull},
					{"SUBSTR(NULL, -4, 2)", evaluator.SQLNull},
					{"SUBSTR('this' FROM NULL FOR 2)", evaluator.SQLNull},
					{"SUBSTR('this', 2, NULL )", evaluator.SQLNull},
					{"SUBSTR('this' FROM 3 FOR NULL)", evaluator.SQLNull},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: SUBSTRING_INDEX", func() {
				tests := []test{
					{"SUBSTRING_INDEX('www.cmysql.com', '.', NULL)", evaluator.SQLNull},
					{"SUBSTRING_INDEX('www.cmysql.com', '.', 0)", evaluator.SQLVarchar("")},
					{"SUBSTRING_INDEX('www.cmysql.com', '.', 1)", evaluator.SQLVarchar("www")},
					{"SUBSTRING_INDEX('www.cmysql.com', '.c', 1)", evaluator.SQLVarchar("www")},
					{"SUBSTRING_INDEX('www.cmysql.com', '.', 2)", evaluator.SQLVarchar("www.cmysql")},
					{"SUBSTRING_INDEX('www.cmysql.com', '.', 1000)", evaluator.SQLVarchar("www.cmysql.com")},
					{"SUBSTRING_INDEX('www.cmysql.com', '.c', 2)", evaluator.SQLVarchar("www.cmysql")},
					{"SUBSTRING_INDEX('www.cmysql.com', '.', -2)", evaluator.SQLVarchar("cmysql.com")},
					{"SUBSTRING_INDEX('www.cmysql.com', '.', -1)", evaluator.SQLVarchar("com")},
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
					{"STR_TO_DATE(NULL, 4)", evaluator.SQLNull},
					{"STR_TO_DATE('foobarbar', NULL)", evaluator.SQLNull},
					{"STR_TO_DATE('2016-04-03','%Y-%m-%d')", evaluator.SQLDate{d}},
					{"STR_TO_DATE('04,03,2016', '%m,%d,%Y')", evaluator.SQLDate{d}},
					{"STR_TO_DATE('04,03,a16', '%m,%d,a%y')", evaluator.SQLDate{d}},
					{"STR_TO_DATE('2016-04-03 12:22:22', '%Y-%m-%d %H:%i:%s')", evaluator.SQLTimestamp{t}},
					{"STR_TO_DATE('2016-04-03 12:22', '%Y-%m-%d %H:%i')", evaluator.SQLTimestamp{t2}},
					{"STR_TO_DATE('2005-04-02 12', '%Y-%m-%d %i')", evaluator.SQLTimestamp{t1}},
					{"STR_TO_DATE('Apr 03, 2016', '%b %d, %Y')", evaluator.SQLDate{d}},
					{"STR_TO_DATE('Tue 2016-04-03', '%a %Y-%m-%d')", evaluator.SQLDate{d}},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TAN", func() {
				tests := []test{
					{"TAN(NULL)", evaluator.SQLNull},
					{"TAN(19)", evaluator.SQLFloat(0.15158947061240008)},
					{"TAN(-19)", evaluator.SQLFloat(-0.15158947061240008)},
					{"TAN('C')", evaluator.SQLFloat(0)},
					{"TAN(0)", evaluator.SQLFloat(0)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TIME_TO_SEC", func() {

				tests := []test{
					{"TIME_TO_SEC(NULL)", evaluator.SQLNull},
					{"TIME_TO_SEC('22:23:00')", evaluator.SQLFloat(80580)},
					{"TIME_TO_SEC('12:34')", evaluator.SQLFloat(45240)},
					{"TIME_TO_SEC('00:39:38')", evaluator.SQLFloat(2378)},
					{"TIME_TO_SEC(1010103)", evaluator.SQLFloat(363663)},
					{"TIME_TO_SEC('2222')", evaluator.SQLFloat(1342)},
					{"TIME_TO_SEC(101010)", evaluator.SQLFloat(36610)},
					{"TIME_TO_SEC(-222)", evaluator.SQLFloat(-142)},
					{"TIME_TO_SEC('-22:33:32')", evaluator.SQLFloat(-81212)},
					{"TIME_TO_SEC(535911)", evaluator.SQLFloat(194351)},
					{"TIME_TO_SEC('-850:00:00')", evaluator.SQLFloat(-3020399)},
					{"TIME_TO_SEC('-838:59:59')", evaluator.SQLFloat(-3020399)},
					{"TIME_TO_SEC(CONCAT('48:2','4:59'))", evaluator.SQLFloat(174299)},
					{"TIME_TO_SEC(535959.9)", evaluator.SQLFloat(194399)},
					{"TIME_TO_SEC(534422333)", evaluator.SQLNull},
					{"TIME_TO_SEC(539911)", evaluator.SQLNull},
					{"TIME_TO_SEC(8991111)", evaluator.SQLNull},
					{"TIME_TO_SEC('-5359:11')", evaluator.SQLFloat(-3020399)},
					{"TIME_TO_SEC('2004-07-09 10:17:35')", evaluator.SQLFloat(37055)},
					{"TIME_TO_SEC('2004-07-09 10:17:35.238238')", evaluator.SQLFloat(37055)},
				}

				runTests(evalCtx, tests)
			})

			Convey("Subject: TIMEDIFF", func() {
				tests := []test{
					{"TIMEDIFF('2000:11:11 00:00:00', NULL)", evaluator.SQLNull},
					{"TIMEDIFF(NULL, '2000:11:11 00:00:00')", evaluator.SQLNull},
					{"TIMEDIFF('2000:09:11 00:00:00', '2000:09:31 00:00:01:323211')", evaluator.SQLNull},
					{"TIMEDIFF('2008-12-31 23:59:59.000001','2008-12-31 23:59:58.000001')", evaluator.SQLVarchar("00:00:01")},
					{"TIMEDIFF('2000:11:11 00:00:00', '2000:11:11 10:00:00.000231')", evaluator.SQLVarchar("-10:00:00.000231")},
					{"TIMEDIFF('2000:01:01 00:00:00','2000:01:01 00:00:00.000001')", evaluator.SQLVarchar("-00:00:00.000001")},
					{"TIMEDIFF('2008-12-31 23:59:59.000001','2008-12-30 01:01:01.000002')", evaluator.SQLVarchar("46:58:57.999999")},
				}

				runTests(evalCtx, tests)
			})

			Convey("Subject: TIMESTAMP", func() {
				t1, err := time.Parse("2006-01-02 15:04:05.000000", "2010-01-01 22:35:10.523236")
				So(err, ShouldBeNil)
				t2, err := time.Parse("2006-01-02 15:04:05.000000", "2010-01-01 23:33:11.400000")
				So(err, ShouldBeNil)
				t3, err := time.Parse("2006-01-02 15:04:05", "2004-01-01 00:00:00")
				So(err, ShouldBeNil)
				t4, err := time.Parse("2006-01-02 15:04:05", "2003-12-31 00:00:00")
				So(err, ShouldBeNil)
				t5, err := time.Parse("2006-01-02 15:04:05.000000", "2003-12-31 12:00:12.300000")
				So(err, ShouldBeNil)
				t6, err := time.Parse("2006-01-02 15:04:05", "2003-12-31 12:23:23")
				So(err, ShouldBeNil)
				t7, err := time.Parse("2006-01-02 15:04:05", "2010-01-01 12:33:23")
				So(err, ShouldBeNil)

				tests := []test{
					{"TIMESTAMP(NULL)", evaluator.SQLNull},
					{"TIMESTAMP(NULL, NULL)", evaluator.SQLNull},
					{"TIMESTAMP(NULL, '12:22:22')", evaluator.SQLNull},
					{"TIMESTAMP('2002-01-02', NULL)", evaluator.SQLNull},
					{"TIMESTAMP('2010-01-01 11:11:11', '11:71:11')", evaluator.SQLNull},
					{"TIMESTAMP('2010-01-01 11:11:11', '11:23:59.5232355')", evaluator.SQLTimestamp{Time: t1}},
					{"TIMESTAMP('2010-01-01 11:11:11', '12:22.4:12')", evaluator.SQLTimestamp{Time: t2}},
					{"TIMESTAMP('2003-12-31 12:00:00', '12:00:00')", evaluator.SQLTimestamp{Time: t3}},
					{"TIMESTAMP(20031231)", evaluator.SQLTimestamp{Time: t4}},
					{"TIMESTAMP('2003-12-31')", evaluator.SQLTimestamp{Time: t4}},
					{"TIMESTAMP('2003-12-31 12:00:00', '12.3:10:30')", evaluator.SQLTimestamp{Time: t5}},
					{"TIMESTAMP('2003-12-31 12:23:23')", evaluator.SQLTimestamp{Time: t6}},
					{"TIMESTAMP('2010-01-01 11:11:11', '12212')", evaluator.SQLTimestamp{Time: t7}},
					{"TIMESTAMP('2010-01-01 11:11:11', 12212)", evaluator.SQLTimestamp{Time: t7}},
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
				t2 := t.Add(time.Duration(15000) * time.Microsecond)

				tests := []test{
					{"TIMESTAMPADD(YEAR, 1, DATE '2002-01-02')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(YEAR, 0.5, DATE '2002-01-02')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(QUARTER, 1, DATE '2002-10-02')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(QUARTER, 0.5, DATE '2002-10-02')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(MONTH, 1, DATE '2002-12-02')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(MONTH, 0.5, DATE '2002-12-02')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(WEEK, 1, DATE '2002-12-26')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(WEEK, 0.5, DATE '2002-12-26')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(DAY, 1, DATE '2003-01-01')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(DAY, 0.5, DATE '2003-01-01')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(HOUR, 1, DATE '2003-01-02')", evaluator.SQLTimestamp{Time: dt}},
					{"TIMESTAMPADD(HOUR, 0.5, DATE '2003-01-02')", evaluator.SQLTimestamp{Time: dt}},
					{"TIMESTAMPADD(MINUTE, 60, DATE '2003-01-02')", evaluator.SQLTimestamp{Time: dt}},
					{"TIMESTAMPADD(MINUTE, 59.5, DATE '2003-01-02')", evaluator.SQLTimestamp{Time: dt}},
					{"TIMESTAMPADD(SECOND, 3600, DATE '2003-01-02')", evaluator.SQLTimestamp{Time: dt}},
					// No round test for SECOND, SECOND is not rounded.
					{"TIMESTAMPADD(MICROSECOND, 15000, TIMESTAMP '2003-01-02 12:30:09')", evaluator.SQLTimestamp{Time: t2}},
					{"TIMESTAMPADD(DAY, 1, TIMESTAMP '2003-01-01 12:30:09')", evaluator.SQLTimestamp{Time: t}},
					{"TIMESTAMPADD(WEEK, 2, TIMESTAMP '2002-12-19 12:30:09')", evaluator.SQLTimestamp{Time: t}},
					{"TIMESTAMPADD(SQL_TSI_YEAR, 2, TIMESTAMP '2001-01-02 12:30:09')", evaluator.SQLTimestamp{Time: t}},
					{"TIMESTAMPADD(SQL_TSI_QUARTER, 2, DATE '2002-07-02')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(SQL_TSI_MONTH, 1, TIMESTAMP '2002-12-02 12:30:09')", evaluator.SQLTimestamp{Time: t}},
					{"TIMESTAMPADD(SQL_TSI_WEEK, 1, DATE '2002-12-26')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(SQL_TSI_DAY, 1, DATE '2003-01-01')", evaluator.SQLTimestamp{Time: d}},
					{"TIMESTAMPADD(SQL_TSI_HOUR, 1, TIMESTAMP '2003-01-02 11:30:09')", evaluator.SQLTimestamp{Time: t}},
					{"TIMESTAMPADD(SQL_TSI_MINUTE, 1, TIMESTAMP '2003-01-02 12:29:09')", evaluator.SQLTimestamp{Time: t}},
					{"TIMESTAMPADD(SQL_TSI_SECOND, 1, TIMESTAMP '2003-01-02 12:30:08')", evaluator.SQLTimestamp{Time: t}},
				}
				runTests(evalCtx, tests)

				typeTests := []typeTest{
					{"TIMESTAMPADD(SQL_TSI_QUARTER, 2, DATE '2002-07-02')", schema.SQLTimestamp},
					{"TIMESTAMPADD(SQL_TSI_SECOND, 1, TIMESTAMP '2003-01-02 12:30:08')", schema.SQLTimestamp},
				}
				runTypeTests(evalCtx, typeTests)
			})

			Convey("Subject: TIMESTAMPDIFF", func() {
				tests := []test{
					{"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-02')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(YEAR, DATE '2002-01-02', DATE '2001-01-02')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(YEAR, DATE '2001-01-03', DATE '2002-01-02')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-03')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(QUARTER, DATE '2002-04-02', DATE '2002-01-02')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-06-02')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-07-02')", evaluator.SQLInt(2)},
					{"TIMESTAMPDIFF(QUARTER, DATE '2002-07-02', DATE '2002-01-02')", evaluator.SQLInt(-2)},
					{"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-01')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(MONTH, DATE '2002-02-01', DATE '2001-01-02')", evaluator.SQLInt(-12)},
					{"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-02')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(MONTH, DATE '2002-02-03', DATE '2002-01-02')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-16')", evaluator.SQLInt(2)},
					{"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-15')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(WEEK, DATE '2001-01-15', DATE '2001-01-02')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-17')", evaluator.SQLInt(2)},
					{"TIMESTAMPDIFF(DAY, DATE '2003-01-04', DATE '2003-01-16')", evaluator.SQLInt(12)},
					{"TIMESTAMPDIFF(DAY, DATE '2003-01-16', DATE '2003-01-04')", evaluator.SQLInt(-12)},
					{"TIMESTAMPDIFF(HOUR, DATE '2003-01-04', DATE '2003-01-06')", evaluator.SQLInt(48)},
					{"TIMESTAMPDIFF(MINUTE, DATE '2003-01-04', DATE '2003-01-06')", evaluator.SQLInt(2880)},
					{"TIMESTAMPDIFF(SECOND, DATE '2003-01-04', DATE '2003-01-05')", evaluator.SQLInt(86400)},
					{"TIMESTAMPDIFF(MICROSECOND, DATE '2003-01-04', DATE '2003-01-05')", evaluator.SQLInt(86400000000)},
					{"TIMESTAMPDIFF(MICROSECOND, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-02 13:40:33')", evaluator.SQLInt(90624000000)},
					{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', TIMESTAMP '2003-03-04 12:45:30')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', TIMESTAMP '2002-03-04 12:45:30')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-03-04 12:45:30', TIMESTAMP '2002-01-02 12:30:09')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2003-03-04 12:30:06', DATE '2002-03-04')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(SQL_TSI_YEAR, DATE '2004-03-04', TIMESTAMP '2003-03-04 12:30:06')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-01-01', TIMESTAMP '2002-04-01 12:30:06')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-04-01 12:30:06', DATE '2002-01-01')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-01-01 12:30:06', DATE '2002-04-01')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-04-01', TIMESTAMP '2002-01-01 12:30:06')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-01-01', TIMESTAMP '2002-03-01 12:30:09')", evaluator.SQLInt(2)},
					{"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-03-01 12:30:09', DATE '2002-01-01')", evaluator.SQLInt(-2)},
					{"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-03-01')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-03-01', TIMESTAMP '2002-01-01 12:30:09')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-08')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(SQL_TSI_WEEK, DATE '2002-01-01', TIMESTAMP '2002-01-08 12:30:09')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-08 12:30:09', DATE '2002-01-01')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(SQL_TSI_DAY, DATE '2002-01-01', TIMESTAMP '2002-01-02 12:30:09')", evaluator.SQLInt(1)},
					{"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-02 12:30:09', DATE '2002-01-01')", evaluator.SQLInt(-1)},
					{"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')", evaluator.SQLInt(0)},
					{"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')", evaluator.SQLInt(11)},
					{"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-02 11:02:33')", evaluator.SQLInt(22)},
					{"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-01 13:02:33')", evaluator.SQLInt(32)},
					{"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')", evaluator.SQLInt(689)},
					{"TIMESTAMPDIFF(SQL_TSI_SECOND, TIMESTAMP '2002-01-01 12:30:09', TIMESTAMP '2002-01-02 14:40:33')", evaluator.SQLInt(94224)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: TO_DAYS", func() {
				tests := []test{
					{"TO_DAYS(NULL)", evaluator.SQLNull},
					{"TO_DAYS('')", evaluator.SQLNull},
					{"TO_DAYS('0000-00-00')", evaluator.SQLNull},
					{"TO_DAYS('0000-01-01')", evaluator.SQLInt(1)},
					{"TO_DAYS('0000-11-11')", evaluator.SQLInt(315)},
					{"TO_DAYS('00-11-11')", evaluator.SQLInt(730800)},
					{"TO_DAYS('950501')", evaluator.SQLInt(728779)},
					{"TO_DAYS(950501)", evaluator.SQLInt(728779)},
					{"TO_DAYS('1995-05-01')", evaluator.SQLInt(728779)},
					{"TO_DAYS('2007-10-07')", evaluator.SQLInt(733321)},
					{"TO_DAYS(881111)", evaluator.SQLInt(726417)},
					{"TO_DAYS('2006-01-02')", evaluator.SQLInt(732678)},
					{"TO_DAYS('1452-04-15')", evaluator.SQLInt(530437)},
					{"TO_DAYS('4222-12-12')", evaluator.SQLInt(1542399)},
					{"TO_DAYS('2000-09-23 13:45:00')", evaluator.SQLInt(730751)},
					{"TO_DAYS('2000-09-24 13:45:00')", evaluator.SQLInt(730752)},
					{"TO_DAYS('2000-10-24 13:45:00')", evaluator.SQLInt(730782)},
				}

				runTests(evalCtx, tests)
			})

			Convey("Subject: TO_SECONDS", func() {
				tests := []test{
					{"TO_SECONDS(NULL)", evaluator.SQLNull},
					{"TO_SECONDS('')", evaluator.SQLNull},
					{"TO_SECONDS('0000-00-00')", evaluator.SQLNull},
					{"TO_SECONDS('0000-01-01')", evaluator.SQLInt(86400)},
					{"TO_SECONDS('0000-11-11')", evaluator.SQLInt(27216000)},
					{"TO_SECONDS('00-11-11')", evaluator.SQLInt(63141120000)},
					{"TO_SECONDS('950501')", evaluator.SQLInt(62966505600)},
					{"TO_SECONDS(950501)", evaluator.SQLInt(62966505600)},
					{"TO_SECONDS('1995-05-01')", evaluator.SQLInt(62966505600)},
					{"TO_SECONDS('2007-10-07')", evaluator.SQLInt(63358934400)},
					{"TO_SECONDS(881111)", evaluator.SQLInt(62762428800)},
					{"TO_SECONDS('2006-01-02')", evaluator.SQLInt(63303379200)},
					{"TO_SECONDS('1452-04-15')", evaluator.SQLInt(45829756800)},
					{"TO_SECONDS('4222-12-12')", evaluator.SQLInt(133263273600)},
					{"TO_SECONDS('2000-09-23 13:45:00')", evaluator.SQLInt(63136935900)},
					{"TO_SECONDS('2000-09-24 13:45:00')", evaluator.SQLInt(63137022300)},
					{"TO_SECONDS('2000-10-24 13:45:00')", evaluator.SQLInt(63139614300)},
					{"TO_SECONDS('2000-10-24 15:45:00')", evaluator.SQLInt(63139621500)},
					{"TO_SECONDS('2000-10-24 13:47:00')", evaluator.SQLInt(63139614420)},
					{"TO_SECONDS('2000-10-24 13:45:59')", evaluator.SQLInt(63139614359)},
				}

				runTests(evalCtx, tests)
			})

			Convey("Subject: TRIM", func() {
				tests := []test{
					{"TRIM(NULL)", evaluator.SQLNull},
					{"TRIM('   bar   ')", evaluator.SQLVarchar("bar")},
					{"TRIM(BOTH 'xyz' FROM 'xyzbarxyzxyz')", evaluator.SQLVarchar("bar")},
					{"TRIM(LEADING 'xyz' FROM 'xyzbarxyzxyz')", evaluator.SQLVarchar("barxyzxyz")},
					{"TRIM(TRAILING 'xyz' FROM 'xyzbarxyzxyz')", evaluator.SQLVarchar("xyzbar")},
					{"TRIM('xyz' FROM 'xyzbarxyzxyz')", evaluator.SQLVarchar("bar")},
				}

				runTests(evalCtx, tests)
			})

			Convey("Subject: TRUNCATE", func() {
				tests := []test{
					{"TRUNCATE(NULL, 2)", evaluator.SQLNull},
					{"TRUNCATE(1234.1234, NULL)", evaluator.SQLNull},
					{"TRUNCATE(1 / 0, 2)", evaluator.SQLNull},
					{"TRUNCATE(1234.1234, 1 / 0)", evaluator.SQLNull},
					{"TRUNCATE(1234.1234, 3)", evaluator.SQLFloat(1234.123)},
					{"TRUNCATE(1234.1234, 5)", evaluator.SQLFloat(1234.1234)},
					{"TRUNCATE(1234.1234, 0)", evaluator.SQLFloat(1234)},
					{"TRUNCATE(1234.1234, -3)", evaluator.SQLFloat(1000)},
					{"TRUNCATE(1234.1234, -5)", evaluator.SQLFloat(0)},
					{"TRUNCATE(-1234.1234, 3)", evaluator.SQLFloat(-1234.123)},
					{"TRUNCATE(-1234.1234, -3)", evaluator.SQLFloat(-1000)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: UCASE", func() {
				tests := []test{
					{"UCASE(NULL)", evaluator.SQLNull},
					{"UCASE('sdg')", evaluator.SQLVarchar("SDG")},
					{"UCASE(124)", evaluator.SQLVarchar("124")},
					{"UPPER(NULL)", evaluator.SQLNull},
					{"UPPER('')", evaluator.SQLVarchar("")},
					{"UPPER('a')", evaluator.SQLVarchar("A")},
					{"UPPER('AWESOME')", evaluator.SQLVarchar("AWESOME")},
					{"UPPER('AwEsOmE')", evaluator.SQLVarchar("AWESOME")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: UNIX_TIMESTAMP", func() {
				tests := []test{
					{"UNIX_TIMESTAMP(NULL)", evaluator.SQLNull},
					{"UNIX_TIMESTAMP('1923-12-12')", evaluator.SQLFloat(0)},
					/*
						These tests will fail if run on a server in a timezone
						different from EST (-05:00) - thus are flaky and commented out.
						test{"UNIX_TIMESTAMP('2015-11-13 10:20:19')", SQLUint64(1447428019)},
						test{"UNIX_TIMESTAMP('2017-03-27 03:00:00')", SQLUint64(1490598000)},
						test{"UNIX_TIMESTAMP('2012-11-17 12:00:00')", SQLUint64(1353171600)},
						test{"UNIX_TIMESTAMP('1985-03-21')", SQLUint64(480229200)},
						test{"UNIX_TIMESTAMP('1985')", SQLFloat(0)},
						test{"UNIX_TIMESTAMP('1985-12')", SQLFloat(0)},
						test{"UNIX_TIMESTAMP('1985-12-aa')", SQLFloat(0)},
						test{"UNIX_TIMESTAMP('1985-12-')", SQLFloat(0)},
						test{"UNIX_TIMESTAMP('1985-12-1')", SQLUint64(502261200)},
						test{"UNIX_TIMESTAMP('1985-12-01')", SQLUint64(502261200)},
					*/
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: WEEK", func() {
				tests := []test{
					{"WEEK(NULL)", evaluator.SQLNull},
					{"WEEK('sdg')", evaluator.SQLNull},
					{"WEEK('2016-1-01 10:23:52')", evaluator.SQLInt(0)},
					{"WEEK(DATE '2009-1-01')", evaluator.SQLInt(0)},
					{"WEEK(DATE '2009-1-01',0)", evaluator.SQLInt(0)},
					{"WEEK(DATE '2009-1-01','str')", evaluator.SQLInt(0)},
					{"WEEK(DATE '2009-1-01',1)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-1-01',2)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2009-1-01',3)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-1-01',4)", evaluator.SQLInt(0)},
					{"WEEK(DATE '2009-1-01',5)", evaluator.SQLInt(0)},
					{"WEEK(DATE '2009-1-01',6)", evaluator.SQLInt(53)},
					{"WEEK(DATE '2009-1-01',7)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2009-1-05')", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-1-05',1)", evaluator.SQLInt(2)},
					{"WEEK(DATE '2009-1-05',2)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-1-05',3)", evaluator.SQLInt(2)},
					{"WEEK(DATE '2009-1-05',4)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-1-05',5)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-1-05',6)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-1-05',7)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2009-12-31')", evaluator.SQLInt(52)},
					{"WEEK(DATE '2009-12-31',1)", evaluator.SQLInt(53)},
					{"WEEK(DATE '2009-12-31',2)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2009-12-31',3)", evaluator.SQLInt(53)},
					{"WEEK(DATE '2009-12-31',4)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2009-12-31',5)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2009-12-31',6)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2009-12-31',7)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2007-12-31')", evaluator.SQLInt(52)},
					{"WEEK(DATE '2007-12-31',1)", evaluator.SQLInt(53)},
					{"WEEK(DATE '2007-12-31',2)", evaluator.SQLInt(52)},
					{"WEEK(DATE '2007-12-31',3)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2007-12-31',4)", evaluator.SQLInt(53)},
					{"WEEK(DATE '2007-12-31',5)", evaluator.SQLInt(53)},
					{"WEEK(DATE '2007-12-31',6)", evaluator.SQLInt(1)},
					{"WEEK(DATE '2007-12-31',7)", evaluator.SQLInt(53)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: WEEKDAY", func() {
				tests := []test{
					{"WEEKDAY(NULL)", evaluator.SQLNull},
					{"WEEKDAY('sdg')", evaluator.SQLNull},
					{"WEEKDAY('2016-1-01 10:23:52')", evaluator.SQLInt(4)},
					{"WEEKDAY('2005-05-11')", evaluator.SQLInt(2)},
					{"WEEKDAY(DATE '2016-7-10')", evaluator.SQLInt(6)},
					{"WEEKDAY(DATE '2016-7-11')", evaluator.SQLInt(0)},
					{"WEEKDAY(TIMESTAMP '2016-7-13 21:22:23')", evaluator.SQLInt(2)},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: WEEKOFYEAR", func() {
				tests := []test{
					{"WEEKOFYEAR(NULL)", evaluator.SQLNull},
					{"WEEKOFYEAR('sdg')", evaluator.SQLNull},
					{"WEEKOFYEAR('2008-02-20')", evaluator.SQLInt(8)},
					{"WEEKOFYEAR('2009-01-01')", evaluator.SQLInt(1)},
					{"WEEKOFYEAR(DATE '2009-01-05')", evaluator.SQLInt(2)},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: YEAR", func() {
				tests := []test{
					{"YEAR(NULL)", evaluator.SQLNull},
					{"YEAR('sdg')", evaluator.SQLNull},
					{"YEAR('2016-1-01 10:23:52')", evaluator.SQLInt(53)},
				}
				runTests(evalCtx, tests)
			})

			SkipConvey("Subject: YEARWEEK", func() {
				tests := []test{
					{"YEARWEEK(NULL)", evaluator.SQLNull},
					{"YEARWEEK('sdg')", evaluator.SQLNull},
					{"YEARWEEK('2000-01-01')", evaluator.SQLInt(199252)},
					{"YEARWEEK('2001-01-01')", evaluator.SQLInt(200053)},
					{"YEARWEEK('2002-01-01')", evaluator.SQLInt(200152)},
					{"YEARWEEK('2003-01-01')", evaluator.SQLInt(200252)},
					{"YEARWEEK('2004-01-01')", evaluator.SQLInt(200352)},
					{"YEARWEEK('2005-01-01')", evaluator.SQLInt(200452)},
					{"YEARWEEK('2006-01-01')", evaluator.SQLInt(200601)},
					{"YEARWEEK('2000-01-06')", evaluator.SQLInt(200001)},
					{"YEARWEEK('2001-01-06')", evaluator.SQLInt(200053)},
					{"YEARWEEK('2002-01-06')", evaluator.SQLInt(200201)},
					{"YEARWEEK('2003-01-06')", evaluator.SQLInt(200301)},
					{"YEARWEEK('2004-01-06')", evaluator.SQLInt(200401)},
					{"YEARWEEK('2005-01-06')", evaluator.SQLInt(200501)},
					{"YEARWEEK('2006-01-06')", evaluator.SQLInt(200601)},
					{"YEARWEEK('2000-01-01',1)", evaluator.SQLInt(199252)},
					{"YEARWEEK('2001-01-01',1)", evaluator.SQLInt(200101)},
					{"YEARWEEK('2002-01-01',1)", evaluator.SQLInt(200201)},
					{"YEARWEEK('2003-01-01',1)", evaluator.SQLInt(200301)},
					{"YEARWEEK('2004-01-01',1)", evaluator.SQLInt(200401)},
					{"YEARWEEK('2005-01-01',1)", evaluator.SQLInt(200453)},
					{"YEARWEEK('2006-01-01',1)", evaluator.SQLInt(200552)},
					{"YEARWEEK('2000-01-06',1)", evaluator.SQLInt(200001)},
					{"YEARWEEK('2001-01-06',1)", evaluator.SQLInt(200101)},
					{"YEARWEEK('2002-01-06',1)", evaluator.SQLInt(200201)},
					{"YEARWEEK('2003-01-06',1)", evaluator.SQLInt(200301)},
					{"YEARWEEK('2004-01-06',1)", evaluator.SQLInt(200402)},
					{"YEARWEEK('2005-01-06',1)", evaluator.SQLInt(200501)},
					{"YEARWEEK('2006-01-06',1)", evaluator.SQLInt(200601)},
				}
				runTests(evalCtx, tests)
			})

		})

		Convey("Subject: SQLSubtractExpr", func() {
			tests := []test{
				{"0 - 0", evaluator.SQLInt(0)},
				{"-1 - 1", evaluator.SQLInt(-2)},
				{"10 - 32", evaluator.SQLInt(-22)},
				{"-10 - -32", evaluator.SQLInt(22)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLSubqueryCmpExpr", func() {
			Convey("Should not evaluate if the subquery returns a different number of columns than the left expression", func() {

				rows := []evaluator.Row{
					{evaluator.Values{{1, "", "test", "a", evaluator.SQLInt(1)}, {1, "", "test", "b", evaluator.SQLInt(2)}}},
					{evaluator.Values{{1, "", "test", "a", evaluator.SQLInt(2)}, {1, "", "test", "b", evaluator.SQLInt(4)}}},
				}

				cs := evaluator.NewCacheStage(0, rows, nil, nil)
				subqExpr := evaluator.NewSQLSubqueryExpr(false, false, cs)

				// Single SQLValue in left, two in subquery
				subCmpExpr := evaluator.NewSQLSubqueryCmpExpr(0, evaluator.SQLInt(1), subqExpr, "")
				_, err := subCmpExpr.Evaluate(evalCtx)
				So(err, ShouldNotBeNil)

				// Three SQLValues in left, two in subquery
				left := &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(3)}}
				subCmpExpr = evaluator.NewSQLSubqueryCmpExpr(0, left, subqExpr, "")
				_, err = subCmpExpr.Evaluate(evalCtx)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Subject: SQLTupleExpr", func() {
			Convey("Should evaluate all the expressions and return SQLValues", func() {
				subject := &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{evaluator.SQLInt(10), evaluator.NewSQLAddExpr(evaluator.SQLInt(30), evaluator.SQLInt(12))}}
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, &evaluator.SQLValues{})
				resultValues := result.(*evaluator.SQLValues)
				So(resultValues.Values[0], ShouldEqual, evaluator.SQLInt(10))
				So(resultValues.Values[1], ShouldEqual, evaluator.SQLInt(42))
			})
			Convey("Should evaluate to a single SQLValue if it contains only one value", func() {
				subject := &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{evaluator.SQLInt(10)}}
				sqlInt, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				intResult := sqlInt.(evaluator.SQLInt)
				So(intResult, ShouldEqual, evaluator.SQLInt(10))

				subject = &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{evaluator.SQLVarchar("10")}}
				sqlVarchar, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				varcharResult := sqlVarchar.(evaluator.SQLVarchar)
				So(varcharResult, ShouldEqual, evaluator.SQLVarchar("10"))
			})

			Convey("Should evaluate early if possible", func() {
				tests := []test{
					{"(1, 3) > (2, 4)", evaluator.SQLFalse},
					{"(1, 3) > ROW(2, 4)", evaluator.SQLFalse},
				}
				runTests(evalCtx, tests)
			})
		})

		Convey("Subject: SQLUnaryMinusExpr", func() {
			tests := []test{
				{"- 10", evaluator.SQLInt(-10)},
				{"- a", evaluator.SQLInt(-123)},
				{"- b", evaluator.SQLInt(-456)},
				{"- null", evaluator.SQLNull},
				{"- true", evaluator.SQLInt(-1)},
				{"- false", evaluator.SQLInt(0)},
				{"- date '2005-05-11'", evaluator.SQLInt(-20050511)},
				{"- timestamp '2005-05-11 12:22:04'", evaluator.SQLInt(-20050511122204)},
				{"- '4' ", evaluator.SQLFloat(-4)},
				{"- 6.7", evaluator.SQLDecimal128(decimal.New(-67, -1))},
				{"- '3.3'", evaluator.SQLFloat(-3.3)},
			}

			runTests(evalCtx, tests)
		})

		Convey("Subject: SQLVariableExpr", func() {
			tests := []test{
				{"@@autocommit", evaluator.SQLTrue},
				{"@@global.autocommit", evaluator.SQLTrue},
			}

			runTests(evalCtx, tests)

			Convey("Should error when unknown variable is used", func() {
				subject := evaluator.NewSQLVariableExpr(
					"blah",
					variable.SystemKind,
					variable.SessionScope,
					schema.SQLNone,
				)

				_, err := subject.Evaluate(evalCtx)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Subject: SQLUnaryPlusExpr", func() {
			tests := []test{
				{"+1", evaluator.SQLInt(1)},
				{"+'string'", evaluator.SQLVarchar("string")},
				{"+a", evaluator.SQLInt(123)},
			}

			runTests(evalCtx, tests)

		})

		SkipConvey("Subject: SQLUnaryTildeExpr", func() {
			// TODO: I'm not convinced we have this correct.
		})
	})
}

func TestSQLLikeExprConvertToPattern(t *testing.T) {
	test := func(syntax, expected string) {
		Convey(fmt.Sprintf("XXX LIKE '%s' should convert to pattern '%s'", syntax, expected), func() {
			pattern := evaluator.ConvertSQLValueToPattern(evaluator.SQLVarchar(syntax), '\\')
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

		evalCtx := evaluator.NewEvalCtx(nil, collation.Default)

		tests := [][]interface{}{
			{evaluator.SQLInt(124), true},
			{evaluator.SQLFloat(1235), true},
			{evaluator.SQLVarchar("512"), true},
			{evaluator.SQLInt(0), false},
			{evaluator.SQLFloat(0), false},
			{evaluator.SQLVarchar("000"), false},
			{evaluator.SQLVarchar("skdjbkjb"), false},
			{evaluator.SQLVarchar(""), false},
			{evaluator.SQLTrue, true},
			{evaluator.SQLFalse, false},
			{evaluator.NewSQLEqualsExpr(evaluator.SQLInt(42), evaluator.SQLInt(42)), true},
			{evaluator.NewSQLEqualsExpr(evaluator.SQLInt(42), evaluator.SQLInt(21)), false},
		}

		for _, t := range tests {
			Convey(fmt.Sprintf("Should evaluate %v(%T) to %v", t[0], t[0], t[1]), func() {
				m, err := evaluator.Matches(t[0].(evaluator.SQLExpr), evalCtx)
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
		sqlValue evaluator.SQLValue
	}

	runTests := func(tests []test) {
		for _, t := range tests {
			Convey(fmt.Sprintf("converting %v (%T) to '%v' should yield %v (%T)", t.input, t.input, t.sqlType, t.sqlValue, t.sqlValue), func() {
				val, _ := evaluator.NewSQLValue(t.input, t.sqlType, schema.SQLNone)
				So(val.String(), ShouldEqual, t.sqlValue.String())
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
		sqlVal              = evaluator.SQLInt(0)
		zeroTime            = time.Time{}
		defaultSQLDate      = evaluator.SQLDate{zeroTime}
		bsonDecimal128, _   = bson.ParseDecimal128("1.5")
	)

	Convey("Subject: NewSQLValue", t, func() {

		Convey("Subject: SQLNull", func() {
			tests := []test{
				{nil, schema.SQLBoolean, evaluator.SQLNull},
				{nil, schema.SQLDate, evaluator.SQLNull},
				{nil, schema.SQLDecimal128, evaluator.SQLNull},
				{nil, schema.SQLFloat, evaluator.SQLNull},
				{nil, schema.SQLInt, evaluator.SQLNull},
				{nil, schema.SQLInt64, evaluator.SQLNull},
				{nil, schema.SQLNumeric, evaluator.SQLNull},
				{nil, schema.SQLObjectID, evaluator.SQLNull},
				{nil, schema.SQLVarchar, evaluator.SQLNull},
			}

			runTests(tests)

		})

		Convey("Subject: SQLValue", func() {
			tests := []test{
				{sqlVal, schema.SQLBoolean, evaluator.SQLFalse},
				{sqlVal, schema.SQLDate, defaultSQLDate},
				{sqlVal, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.NewFromFloat(0))},
				{sqlVal, schema.SQLFloat, evaluator.SQLFloat(0)},
				{sqlVal, schema.SQLInt, evaluator.SQLInt(0)},
				{sqlVal, schema.SQLInt64, evaluator.SQLUint64(0)},
				{sqlVal, schema.SQLNumeric, evaluator.SQLFloat(0)},
				{sqlVal, schema.SQLObjectID, evaluator.SQLObjectID(strconv.FormatInt(int64(sqlVal), 10))},
				{sqlVal, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{sqlVal, schema.SQLNone, sqlVal},
			}

			runTests(tests)

		})

		Convey("Subject: SQLBoolean", func() {
			tests := []test{
				{false, schema.SQLBoolean, evaluator.SQLFalse},
				{true, schema.SQLBoolean, evaluator.SQLTrue},
				{floatVal, schema.SQLBoolean, evaluator.SQLBool(floatVal)},
				{0.0, schema.SQLBoolean, evaluator.SQLFalse},
				{objectIDVal, schema.SQLBoolean, evaluator.SQLTrue},
				{intVal, schema.SQLBoolean, evaluator.SQLBool(intVal)},
				{0, schema.SQLBoolean, evaluator.SQLFalse},
				{strFloatVal, schema.SQLBoolean, evaluator.SQLBool(3.23)},
				{"0.000", schema.SQLBoolean, evaluator.SQLFalse},
				{"1.0", schema.SQLBoolean, evaluator.SQLTrue},
				{strTimeVal, schema.SQLBoolean, evaluator.SQLFalse},
				{timeVal, schema.SQLBoolean, evaluator.SQLTrue},
			}

			runTests(tests)

		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				{false, schema.SQLDate, defaultSQLDate},
				{true, schema.SQLDate, defaultSQLDate},
				{floatVal, schema.SQLDate, defaultSQLDate},
				{0.0, schema.SQLDate, defaultSQLDate},
				{objectIDVal, schema.SQLDate, evaluator.SQLDate{objectIDVal.Time()}},
				{intVal, schema.SQLDate, defaultSQLDate},
				{0, schema.SQLDate, defaultSQLDate},
				{strFloatVal, schema.SQLDate, defaultSQLDate},
				{"0.000", schema.SQLDate, defaultSQLDate},
				{"1.0", schema.SQLDate, defaultSQLDate},
				{strTimeVal, schema.SQLDate, evaluator.SQLDate{strTimeValDate}},
				{timeVal, schema.SQLDate, evaluator.SQLDate{timeValParsed}},
			}

			runTests(tests)

		})

		Convey("Subject: SQLDecimal128", func() {
			tests := []test{
				{false, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{true, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(1, 0))},
				{floatVal, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.NewFromFloat(floatVal))},
				{0.0, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{objectIDVal, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{intVal, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.NewFromFloat(float64(intVal)))},
				{0, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{strFloatVal, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.NewFromFloat(floatVal + .1))},
				{"0.000", schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{"1.0", schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(1, 0))},
			}

			runTests(tests)

		})

		Convey("Subject: SQLFloat, SQLNumeric", func() {
			tests := []test{

				//
				// SQLFloat
				//
				{false, schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{true, schema.SQLFloat, evaluator.SQLFloat(1.0)},
				{floatVal, schema.SQLFloat, evaluator.SQLFloat(floatVal)},
				{0.0, schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{intVal, schema.SQLFloat, evaluator.SQLFloat(float64(intVal))},
				{0, schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{strFloatVal, schema.SQLFloat, evaluator.SQLFloat(3.23)},
				{"0.000", schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{"1.0", schema.SQLFloat, evaluator.SQLFloat(1.0)},
				{bsonDecimal128, schema.SQLFloat, evaluator.SQLFloat(1.5)},

				//
				// SQLNumeric
				//
				{false, schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{true, schema.SQLNumeric, evaluator.SQLFloat(1.0)},
				{floatVal, schema.SQLNumeric, evaluator.SQLFloat(floatVal)},
				{0.0, schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{intVal, schema.SQLNumeric, evaluator.SQLFloat(float64(intVal))},
				{0, schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{strFloatVal, schema.SQLNumeric, evaluator.SQLFloat(3.23)},
				{"0.000", schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{"1.0", schema.SQLNumeric, evaluator.SQLFloat(1.0)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLInt, SQLInt64", func() {
			tests := []test{
				{false, schema.SQLInt, evaluator.SQLInt(0)},
				{true, schema.SQLInt, evaluator.SQLInt(1)},
				{floatVal, schema.SQLInt, evaluator.SQLInt(int64(floatVal))},
				{0.0, schema.SQLInt, evaluator.SQLInt(0)},
				{intVal, schema.SQLInt, evaluator.SQLInt(intVal)},
				{0, schema.SQLInt, evaluator.SQLInt(0)},
				{strFloatVal, schema.SQLInt, evaluator.SQLInt(3)},
				{"0.000", schema.SQLInt, evaluator.SQLInt(0)},
				{"1.0", schema.SQLInt, evaluator.SQLInt(1)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLObjectID", func() {
			tests := []test{
				{false, schema.SQLObjectID, evaluator.SQLObjectID("0")},
				{true, schema.SQLObjectID, evaluator.SQLObjectID("1")},
				{floatVal, schema.SQLObjectID, evaluator.SQLObjectID(strconv.FormatFloat(floatVal, 'f', -1, 64))},
				{0.0, schema.SQLObjectID, evaluator.SQLObjectID("0")},
				{objectIDVal, schema.SQLObjectID, evaluator.SQLObjectID(objectIDVal.Hex())},
				{intVal, schema.SQLObjectID, evaluator.SQLObjectID(strconv.FormatInt(int64(intVal), 10))},
				{0, schema.SQLObjectID, evaluator.SQLObjectID("0")},
				{strFloatVal, schema.SQLObjectID, evaluator.SQLObjectID(strFloatVal)},
				{"0.000", schema.SQLObjectID, evaluator.SQLObjectID("0.000")},
				{"1.0", schema.SQLObjectID, evaluator.SQLObjectID("1.0")},
				{strTimeVal, schema.SQLObjectID, evaluator.SQLObjectID(strTimeVal)},
				{timeVal, schema.SQLObjectID, evaluator.SQLObjectID(bson.NewObjectIdWithTime(timeVal).Hex())},
			}

			runTests(tests)

		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				{false, schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{true, schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{floatVal, schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{0.0, schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{objectIDVal, schema.SQLTimestamp, evaluator.SQLTimestamp{objectIDVal.Time()}},
				{intVal, schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{0, schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{strFloatVal, schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{"0.000", schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{"1.0", schema.SQLTimestamp, evaluator.SQLTimestamp{zeroTime}},
				{strTimeVal, schema.SQLTimestamp, evaluator.SQLTimestamp{strTimeValParsed}},
				{timeVal, schema.SQLTimestamp, evaluator.SQLTimestamp{timeVal}},
			}

			runTests(tests)

		})

		Convey("Subject: SQLUint64", func() {
			tests := []test{
				{false, schema.SQLUint64, evaluator.SQLUint64(0)},
				{true, schema.SQLUint64, evaluator.SQLUint64(1)},
				{floatVal, schema.SQLUint64, evaluator.SQLUint64(uint64(floatVal))},
				{0.0, schema.SQLUint64, evaluator.SQLUint64(0)},
				{intVal, schema.SQLUint64, evaluator.SQLUint64(uint64(intVal))},
				{0, schema.SQLUint64, evaluator.SQLUint64(0)},
				{strFloatVal, schema.SQLUint64, evaluator.SQLUint64(3)},
				{"0.000", schema.SQLUint64, evaluator.SQLUint64(0)},
				{"1.0", schema.SQLUint64, evaluator.SQLUint64(1)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				{false, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{true, schema.SQLVarchar, evaluator.SQLVarchar("1")},
				{floatVal, schema.SQLVarchar, evaluator.SQLVarchar(strconv.FormatFloat(floatVal, 'f', -1, 64))},
				{0.0, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{objectIDVal, schema.SQLVarchar, evaluator.SQLVarchar(objectIDVal.Hex())},
				{intVal, schema.SQLVarchar, evaluator.SQLVarchar(strconv.FormatInt(int64(intVal), 10))},
				{0, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{strFloatVal, schema.SQLVarchar, evaluator.SQLVarchar(strFloatVal)},
				{"0.000", schema.SQLVarchar, evaluator.SQLVarchar("0.000")},
				{"1.0", schema.SQLVarchar, evaluator.SQLVarchar("1.0")},
				{strTimeVal, schema.SQLVarchar, evaluator.SQLVarchar(strTimeVal)},
				{timeVal, schema.SQLVarchar, evaluator.SQLVarchar(timeVal.Format(evaluator.DateTimeFormat))},
			}

			runTests(tests)

		})
	})

}

func TestNewSQLValueFromSQLColumnExpr(t *testing.T) {

	Convey("When creating a SQLValue with no column type specified calling NewSQLValueFromSQLColumnExpr on a", t, func() {

		Convey("SQLValue should return the same object passed in", func() {
			v := evaluator.SQLTrue
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLBoolean, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, v)
		})

		Convey("nil value should return SQLNull", func() {
			v, err := evaluator.NewSQLValueFromSQLColumnExpr(nil, schema.SQLNull, schema.MongoBool)
			So(err, ShouldBeNil)
			So(v, ShouldResemble, evaluator.SQLNull)
		})

		Convey("bson object id should return its string value", func() {
			v := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLVarchar, schema.MongoObjectID)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v.Hex())
		})

		Convey("string objects should return the string value", func() {
			v := "56a10dd56ce28a89a8ed6edb"
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("int objects should return the int value", func() {
			v1 := int(6)
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v1, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v1)

			v2 := int32(6)
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v2, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v2)

			v3 := uint32(6)
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v3, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v3)
		})

		Convey("float objects should return the float value", func() {
			v := float64(6.3)
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("time objects should return the appropriate value", func() {
			v := time.Date(2014, time.December, 31, 0, 0, 0, 0, schema.DefaultLocale)
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(evaluator.SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, evaluator.SQLDate{v})

			v = time.Date(2014, time.December, 31, 10, 0, 0, 0, schema.DefaultLocale)
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLTimestamp, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlTimestamp, ok := newV.(evaluator.SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTimestamp, ShouldResemble, evaluator.SQLTimestamp{v})
		})
	})

	Convey("When creating a SQLValue with a column type specified calling NewSQLValueFromSQLColumnExpr on a", t, func() {

		Convey("a SQLVarchar/SQLVarchar column type should attempt to coerce to the SQLVarchar type", func() {

			t := schema.SQLVarchar

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(t, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar(t))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(6, schema.SQLVarchar, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar("6"))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(6.6, schema.SQLVarchar, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar("6.6"))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLVarchar, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar("6"))

			_id := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(_id, schema.SQLVarchar, schema.MongoObjectID)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLObjectID(_id.Hex()))

		})

		Convey("a SQLInt column type should attempt to coerce to the SQLInt type", func() {

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(true, schema.SQLInt, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(1))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLInt, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(string("6"), schema.SQLInt, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

		})

		Convey("a SQLFloat column type should attempt to coerce to the SQLFloat type", func() {

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(true, schema.SQLFloat, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(1))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLFloat, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6.6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(string("6.6"), schema.SQLFloat, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6.6))

		})

		Convey("a SQLDecimal column type should attempt to coerce to the SQLDecimal type", func() {

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(true, schema.SQLDecimal128, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(1)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int(6), schema.SQLDecimal128, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLDecimal128, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLDecimal128, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLDecimal128, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6.6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(string("6.6"), schema.SQLDecimal128, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6.6)))

		})

		Convey("a SQLDate column type should attempt to coerce to the SQLDate type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 0, 0, 0, 0, schema.DefaultLocale)
			v2 := time.Date(2014, time.May, 11, 10, 32, 12, 0, schema.DefaultLocale)

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v1, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(evaluator.SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, evaluator.SQLDate{v1})

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v2, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok = newV.(evaluator.SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, evaluator.SQLDate{v1})

			// String type
			dates := []string{"2014-05-11", "2014-05-11 15:04:05", "2014-05-11 15:04:05.233"}

			for _, d := range dates {

				newV, err := evaluator.NewSQLValueFromSQLColumnExpr(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldBeNil)

				sqlDate, ok := newV.(evaluator.SQLDate)
				So(ok, ShouldBeTrue)
				So(sqlDate, ShouldResemble, evaluator.SQLDate{v1})

			}

			// invalid dates and those outside valid range
			// should return the default date
			dates = []string{"2014-12-44-44", "10000-1-1"}

			for _, d := range dates {
				newV, err = evaluator.NewSQLValueFromSQLColumnExpr(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldBeNil)

				_, ok := newV.(evaluator.SQLFloat)
				So(ok, ShouldBeTrue)
			}
		})

		Convey("a SQLTimestamp column type should attempt to coerce to the SQLTimestamp type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v1, schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok := newV.(evaluator.SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, evaluator.SQLTimestamp{v1})

			// String type
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr("2014-05-11 15:04:05.000", schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok = newV.(evaluator.SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, evaluator.SQLTimestamp{v1})

			// invalid dates should return the default date
			dates := []string{"2044-12-40", "1966-15-1", "43223-3223"}

			for _, d := range dates {
				newV, err = evaluator.NewSQLValueFromSQLColumnExpr(d, schema.SQLTimestamp, schema.MongoNone)
				So(err, ShouldBeNil)
				_, ok := newV.(evaluator.SQLFloat)
				So(ok, ShouldBeTrue)
			}
		})
	})
}

func TestReconcileSQLExpr(t *testing.T) {

	type test struct {
		sql             string
		reconciledLeft  evaluator.SQLExpr
		reconciledRight evaluator.SQLExpr
	}

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3, &lgr)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("Reconciliation for %q", t.sql), func() {
				e, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				left, right := evaluator.GetBinaryExprLeaves(e)
				left, right, err = evaluator.ReconcileSQLExprs(left, right)
				So(err, ShouldBeNil)
				So(left, ShouldResemble, t.reconciledLeft)
				So(right, ShouldResemble, t.reconciledRight)
			})
		}
	}

	exprConv := evaluator.NewSQLConvertExpr(evaluator.SQLVarchar("2010-01-01"), schema.SQLTimestamp, evaluator.SQLNone)
	exprA := evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt)
	exprB := evaluator.NewSQLColumnExpr(1, "test", "bar", "b", schema.SQLInt, schema.MongoInt)
	exprG := evaluator.NewSQLColumnExpr(1, "test", "bar", "g", schema.SQLTimestamp, schema.MongoDate)

	Convey("Subject: reconcileSQLExpr", t, func() {
		exprTime, err := evaluator.NewSQLScalarFunctionExpr("current_timestamp", []evaluator.SQLExpr{})
		So(err, ShouldBeNil)
		tests := []test{
			{"a = 3", exprA, evaluator.SQLInt(3)},
			{"g - '2010-01-01'", evaluator.NewSQLConvertExpr(exprG, schema.SQLDecimal128, evaluator.SQLNone), evaluator.NewSQLConvertExpr(evaluator.SQLVarchar("2010-01-01"), schema.SQLDecimal128, evaluator.SQLNone)},
			{"a in (3)", exprA, evaluator.SQLInt(3)},
			{"a in (2,3)", exprA, &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{evaluator.SQLInt(2), evaluator.SQLInt(3)}}},
			{"(a) in (3)", exprA, evaluator.SQLInt(3)},
			{"(a,b) in (2,3)", &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{exprA, exprB}}, &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{evaluator.SQLInt(2), evaluator.SQLInt(3)}}},
			{"g > '2010-01-01'", exprG, exprConv},
			{"a and b", exprA, exprB},
			{"a / b", exprA, exprB},
			{"'2010-01-01' and g", exprConv, exprG},
			{"g in ('2010-01-01',current_timestamp())", exprG, &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{exprConv, exprTime}}},
			{"g in ('2010-01-01',current_timestamp)", exprG, &evaluator.SQLTupleExpr{[]evaluator.SQLExpr{exprConv, exprTime}}},
		}

		runTests(tests)
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
			left     evaluator.SQLValue
			right    evaluator.SQLValue
			expected int
		}

		runTests := func(tests []test) {
			for _, t := range tests {
				Convey(fmt.Sprintf("comparing '%v' (%T) to '%v' (%T) should return %v", t.left, t.left, t.right, t.right, t.expected), func() {
					compareTo, err := evaluator.CompareTo(t.left, t.right, collation.Default)
					So(err, ShouldBeNil)
					So(compareTo, ShouldEqual, t.expected)
				})
			}
		}

		Convey("Subject: SQLInt", func() {
			tests := []test{
				{evaluator.SQLInt(1), evaluator.SQLInt(0), 1},
				{evaluator.SQLInt(1), evaluator.SQLInt(1), 0},
				{evaluator.SQLInt(1), evaluator.SQLInt(2), -1},
				{evaluator.SQLInt(1), evaluator.SQLUint32(1), 0},
				{evaluator.SQLInt(1), evaluator.SQLFloat(1), 0},
				{evaluator.SQLInt(1), evaluator.SQLFalse, 1},
				{evaluator.SQLInt(1), evaluator.SQLTrue, 0},
				{evaluator.SQLInt(1), evaluator.SQLNull, 1},
				{evaluator.SQLInt(1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLInt(1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLInt(1), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLInt(1), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLInt(1), evaluator.SQLDate{now}, -1},
				{evaluator.SQLInt(1), evaluator.SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLUint32", func() {
			tests := []test{
				{evaluator.SQLUint32(1), evaluator.SQLInt(0), 1},
				{evaluator.SQLUint32(1), evaluator.SQLInt(1), 0},
				{evaluator.SQLUint32(1), evaluator.SQLInt(2), -1},
				{evaluator.SQLUint32(1), evaluator.SQLUint32(1), 0},
				{evaluator.SQLUint32(1), evaluator.SQLFloat(1), 0},
				{evaluator.SQLUint32(1), evaluator.SQLFalse, 1},
				{evaluator.SQLUint32(1), evaluator.SQLTrue, 0},
				{evaluator.SQLUint32(1), evaluator.SQLNull, 1},
				{evaluator.SQLUint32(1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLUint32(1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLUint32(1), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLUint32(1), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLUint32(1), evaluator.SQLDate{now}, -1},
				{evaluator.SQLUint32(1), evaluator.SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLUint64", func() {
			tests := []test{
				{evaluator.SQLUint64(1), evaluator.SQLInt(0), 1},
				{evaluator.SQLUint64(1), evaluator.SQLInt(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLInt(2), -1},
				{evaluator.SQLUint64(1), evaluator.SQLUint32(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLUint64(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLFloat(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLFalse, 1},
				{evaluator.SQLUint64(1), evaluator.SQLTrue, 0},
				{evaluator.SQLUint64(1), evaluator.SQLNull, 1},
				{evaluator.SQLUint64(1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLUint64(1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLUint64(1), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLUint64(1), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLUint64(1), evaluator.SQLDate{now}, -1},
				{evaluator.SQLUint64(1), evaluator.SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLFloat", func() {
			tests := []test{
				{evaluator.SQLFloat(0.1), evaluator.SQLInt(0), 1},
				{evaluator.SQLFloat(1.1), evaluator.SQLInt(1), 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLInt(2), -1},
				{evaluator.SQLFloat(1.1), evaluator.SQLUint32(1), 1},
				{evaluator.SQLFloat(1.1), evaluator.SQLFloat(1), 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLFalse, 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLTrue, -1},
				{evaluator.SQLFloat(0.1), evaluator.SQLNull, 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLFloat(0.0), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLFloat(0.1), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLDate{now}, -1},
				{evaluator.SQLFloat(0.1), evaluator.SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLBool", func() {
			tests := []test{
				{evaluator.SQLTrue, evaluator.SQLInt(0), 1},
				{evaluator.SQLTrue, evaluator.SQLInt(1), 0},
				{evaluator.SQLTrue, evaluator.SQLInt(2), -1},
				{evaluator.SQLTrue, evaluator.SQLUint32(1), 0},
				{evaluator.SQLTrue, evaluator.SQLFloat(1), 0},
				{evaluator.SQLTrue, evaluator.SQLFalse, 1},
				{evaluator.SQLTrue, evaluator.SQLTrue, 0},
				{evaluator.SQLTrue, evaluator.SQLNull, 1},
				{evaluator.SQLTrue, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLTrue, evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLTrue, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLTrue, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLTrue, evaluator.SQLDate{now}, -1},
				{evaluator.SQLTrue, evaluator.SQLTimestamp{now}, -1},
				{evaluator.SQLFalse, evaluator.SQLInt(0), 0},
				{evaluator.SQLFalse, evaluator.SQLInt(1), -1},
				{evaluator.SQLFalse, evaluator.SQLInt(2), -1},
				{evaluator.SQLFalse, evaluator.SQLUint32(1), -1},
				{evaluator.SQLFalse, evaluator.SQLFloat(1), -1},
				{evaluator.SQLFalse, evaluator.SQLFalse, 0},
				{evaluator.SQLFalse, evaluator.SQLTrue, -1},
				{evaluator.SQLFalse, evaluator.SQLNull, 1},
				{evaluator.SQLFalse, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 0},
				{evaluator.SQLFalse, evaluator.SQLVarchar("bac"), 0},
				{evaluator.SQLFalse, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLFalse, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLFalse, evaluator.SQLDate{now}, -1},
				{evaluator.SQLFalse, evaluator.SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				{evaluator.SQLDate{now}, evaluator.SQLInt(0), 1},
				{evaluator.SQLDate{now}, evaluator.SQLInt(1), 1},
				{evaluator.SQLDate{now}, evaluator.SQLInt(2), 1},
				{evaluator.SQLDate{now}, evaluator.SQLUint32(1), 1},
				{evaluator.SQLDate{now}, evaluator.SQLFloat(1), 1},
				{evaluator.SQLDate{now}, evaluator.SQLFalse, 1},
				{evaluator.SQLDate{now}, evaluator.SQLDate{now.Add(diff)}, -1},
				{evaluator.SQLDate{now}, evaluator.SQLNull, 1},
				{evaluator.SQLDate{now}, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLDate{now}, evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLDate{now}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, 1},
				{evaluator.SQLDate{now}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLDate{now}, evaluator.SQLDate{now.Add(-diff)}, 1},
				{evaluator.SQLDate{now}, evaluator.SQLTimestamp{now.Add(diff)}, -1},
				{evaluator.SQLDate{now}, evaluator.SQLTimestamp{now.Add(-diff)}, 1},
				{evaluator.SQLDate{now}, evaluator.SQLDate{now}, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				{evaluator.SQLTimestamp{now}, evaluator.SQLInt(0), 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLInt(1), 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLInt(2), 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLUint32(1), 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLFloat(1), 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLFalse, 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLNull, 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLTimestamp{now}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, 1},
				{evaluator.SQLTimestamp{now}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLTimestamp{now.Add(diff)}, -1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLTimestamp{now.Add(-diff)}, 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLTimestamp{now}, 0},
				{evaluator.SQLTimestamp{now}, evaluator.SQLDate{now}, 0},
				{evaluator.SQLTimestamp{now.Add(sameDayDiff)}, evaluator.SQLDate{now}, 1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLDate{now.Add(diff)}, -1},
				{evaluator.SQLTimestamp{now}, evaluator.SQLDate{now.Add(-diff)}, 1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLNullValue", func() {
			tests := []test{
				{evaluator.SQLNull, evaluator.SQLInt(0), -1},
				{evaluator.SQLNull, evaluator.SQLInt(1), -1},
				{evaluator.SQLNull, evaluator.SQLInt(2), -1},
				{evaluator.SQLNull, evaluator.SQLUint32(1), -1},
				{evaluator.SQLNull, evaluator.SQLFloat(1), -1},
				{evaluator.SQLNull, evaluator.SQLFalse, -1},
				{evaluator.SQLNull, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{evaluator.SQLNull, evaluator.SQLVarchar("bac"), -1},
				{evaluator.SQLNull, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLNull, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLNull, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNull}}, 0},
				{evaluator.SQLNull, evaluator.SQLDate{now}, -1},
				{evaluator.SQLNull, evaluator.SQLTimestamp{now}, -1},
				{evaluator.SQLNull, evaluator.SQLNull, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				{evaluator.SQLVarchar("bac"), evaluator.SQLInt(0), 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLInt(1), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLInt(2), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLUint32(1), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLFloat(1), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLFalse, 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLVarchar("cba"), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLVarchar("bac"), 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLVarchar("abc"), 1},
				{evaluator.SQLVarchar("bac"), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLVarchar("bac"), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLVarchar("bac"), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLVarchar("bac")}}, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLValues", func() {
			tests := []test{
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLInt(0), 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLInt(1), 0},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLInt(2), -1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLUint32(1), 0},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLUint32(11), -1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLUint32(0), 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLFloat(1.1), -1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLFloat(0.1), 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLFalse, 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLVarchar("abc"), 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLNone, 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(-1)}}, 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(2)}}, -1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLDate{now}, -1},
				{&evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, evaluator.SQLTimestamp{now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLObjectID", func() {

			tests := []test{
				{evaluator.SQLObjectID(oid2), evaluator.SQLInt(0), 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLUint32(1), -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLFloat(1), -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLVarchar("cba"), 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLFalse, 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLTrue, -1},
				{evaluator.SQLObjectID(oid2), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLObjectID(oid2), &evaluator.SQLValues{[]evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLDate{now}, -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLTimestamp{now}, -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLObjectID(oid3), -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLObjectID(oid2), 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLObjectID(oid1), 1},
			}
			runTests(tests)
		})

	})
}

func TestIsTruthyIsFalsy(t *testing.T) {

	Convey("IsTruthy, IsFalsy", t, func() {
		d, err := time.Parse("2006-01-02", "2003-01-02")
		So(err, ShouldBeNil)
		t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
		So(err, ShouldBeNil)

		Convey("Subject: IsTruthy", func() {
			truthy := evaluator.IsTruthy(evaluator.SQLTimestamp{t})
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsTruthy(evaluator.SQLDate{d})
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsTruthy(evaluator.SQLInt(0))
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsTruthy(evaluator.SQLInt(1))
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsTruthy(evaluator.SQLVarchar("dsf"))
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsTruthy(evaluator.SQLVarchar("16"))
			So(truthy, ShouldBeTrue)
		})

		Convey("Subject: IsFalsy", func() {
			truthy := evaluator.IsFalsy(evaluator.SQLTimestamp{t})
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsFalsy(evaluator.SQLDate{d})
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsFalsy(evaluator.SQLInt(0))
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsFalsy(evaluator.SQLInt(1))
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsFalsy(evaluator.SQLVarchar("dsf"))
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsFalsy(evaluator.SQLVarchar("16"))
			So(truthy, ShouldBeFalse)
		})
	})
}

func TestIsUUID(t *testing.T) {
	Convey("IsUUID", t, func() {
		So(evaluator.IsUUID(schema.MongoUUID), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoUUIDCSharp), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoUUIDJava), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoUUIDOld), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoString), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoGeo2D), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoObjectID), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoBool), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoInt), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoInt64), ShouldBeFalse)
	})
}

func TestGetBinaryFromExpr(t *testing.T) {

	Convey("GetBinaryFromExpr", t, func() {

		expected := []byte{
			0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c,
			0x0d, 0x0e, 0x0f, 0x10,
		}

		Convey("Subject: invalid SQLExpr", func() {
			_, ok := evaluator.GetBinaryFromExpr(schema.MongoUUID, evaluator.SQLVarchar("3"))
			So(ok, ShouldBeFalse)
		})

		Convey("Subject: valid SQLExpr with dashes", func() {
			b, ok := evaluator.GetBinaryFromExpr(schema.MongoUUID, evaluator.SQLVarchar("01020304-0506-0708-090a-0b0c0d0e0f10"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x04)
			So(b.Data, ShouldResemble, expected)

			b, ok = evaluator.GetBinaryFromExpr(schema.MongoUUIDOld, evaluator.SQLVarchar("01020304-0506-0708-090a-0b0c0d0e0f10"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x03)
			So(b.Data, ShouldResemble, expected)
		})

		Convey("Subject: valid SQLExpr without dashes", func() {
			b, ok := evaluator.GetBinaryFromExpr(schema.MongoUUIDJava, evaluator.SQLVarchar("0807060504030201100f0e0d0c0b0a09"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x03)
			So(b.Data, ShouldResemble, expected)

			b, ok = evaluator.GetBinaryFromExpr(schema.MongoUUIDCSharp, evaluator.SQLVarchar("0403020106050807090a0b0c0d0e0f10"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x03)
			So(b.Data, ShouldResemble, expected)
		})
	})
}

func TestNormalizeUUID(t *testing.T) {

	Convey("NormalizeUUID", t, func() {
		expected := []byte{
			0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c,
			0x0d, 0x0e, 0x0f, 0x10,
		}

		Convey("Subject: standard UUID", func() {
			bytes := []byte{
				0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
				0x0d, 0x0e, 0x0f, 0x10,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUID, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

		Convey("Subject: old UUID", func() {
			bytes := []byte{
				0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
				0x0d, 0x0e, 0x0f, 0x10,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUIDOld, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

		Convey("Subject: C# Legacy UUID", func() {
			bytes := []byte{
				0x04, 0x03, 0x02, 0x01,
				0x06, 0x05, 0x08, 0x07,
				0x09, 0x0a, 0x0b, 0x0c,
				0x0d, 0x0e, 0x0f, 0x10,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUIDCSharp, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

		Convey("Subject: Java Legacy UUID", func() {
			bytes := []byte{
				0x08, 0x07, 0x06, 0x05,
				0x04, 0x03, 0x02, 0x01,
				0x10, 0x0f, 0x0e, 0x0d,
				0x0c, 0x0b, 0x0a, 0x09,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUIDJava, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

	})
}
