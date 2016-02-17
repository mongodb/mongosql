package evaluator

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
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
		schema, err := schema.ParseSchemaData(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be %v", t.sql, t.result), func() {
				subject, err := getWhereSQLExprFromSQL(schema, "SELECT * FROM bar WHERE "+t.sql)
				So(err, ShouldBeNil)
				result, err := subject.Evaluate(ctx)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	Convey("Subject: Evaluates", t, func() {
		evalCtx := &EvalCtx{
			Rows{{
				Data: TableRows{{
					"bar",
					Values{
						{"a", "a", 123},
						{"b", "b", 456},
						{"c", "c", nil}},
				}}}},
			nil}

		Convey("Subject: SQLAddExpr", func() {
			tests := []test{
				test{"0 + 0", SQLInt(0)},
				test{"-1 + 1", SQLInt(0)},
				test{"10 + 32", SQLInt(42)},
				test{"-10 + -32", SQLInt(-42)},
			}

			runTests(evalCtx, tests)
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
				subject := SQLColumnExpr{"bar", "a"}
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, SQLInt(123))
			})

			Convey("Should return nil when the field is null", func() {
				subject := SQLColumnExpr{"bar", "c"}
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, SQLNull)
			})

			Convey("Should return nil when the field doesn't exists", func() {
				subject := SQLColumnExpr{"bar", "no_existy"}
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

		SkipConvey("Subject: SQLLikeExpr", func() {
		})

		Convey("Subject: SQLMultiplyExpr", func() {
			tests := []test{
				test{"0 * 0", SQLInt(0)},
				test{"-1 * 1", SQLInt(-1)},
				test{"10 * 32", SQLInt(320)},
				test{"-10 * -32", SQLInt(320)},
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

			Convey("Subject: CONCAT", func() {
				tests := []test{
					test{"CONCAT(NULL)", SQLNull},
					test{"CONCAT('A')", SQLString("A")},
					test{"CONCAT('A', 'B')", SQLString("AB")},
					test{"CONCAT('A', NULL, 'B')", SQLNull},
					test{"CONCAT('A', 123, 'B')", SQLString("A123B")},
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

			Convey("Subject: DAYNAME", func() {
				tests := []test{
					test{"DAYNAME(NULL)", SQLNull},
					test{"DAYNAME(14)", SQLNull},
					test{"DAYNAME('2016-01-01 00:00:00')", SQLString("Friday")},
					test{"DAYNAME('2016-1-1')", SQLString("Friday")},
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

			Convey("Subject: LCASE", func() {
				tests := []test{
					test{"LCASE(NULL)", SQLNull},
					test{"LCASE('sDg')", SQLString("sdg")},
					test{"LCASE(124)", SQLString("124")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: LENGTH", func() {
				tests := []test{
					test{"LENGTH(NULL)", SQLNull},
					test{"LENGTH('sDg')", SQLInt(3)},
					test{"LENGTH('世界')", SQLInt(6)},
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
					test{"LTRIM('   barbar')", SQLString("barbar")},
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
					test{"MONTHNAME('2016-1-01 10:23:52')", SQLString("January")},
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

			Convey("Subject: RTRIM", func() {
				tests := []test{
					test{"RTRIM(NULL)", SQLNull},
					test{"RTRIM('barbar   ')", SQLString("barbar")},
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
					test{"SUBSTRING('Quadratically', 5)", SQLString("ratically")},
					test{"SUBSTRING('Quadratically', 5, 6)", SQLString("ratica")},
					test{"SUBSTRING('Sakila', -3)", SQLString("ila")},
					test{"SUBSTRING('Sakila', -5, 3)", SQLString("aki")},
				}
				runTests(evalCtx, tests)
			})

			Convey("Subject: UCASE", func() {
				tests := []test{
					test{"UCASE(NULL)", SQLNull},
					test{"UCASE('sdg')", SQLString("SDG")},
					test{"UCASE(124)", SQLString("124")},
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
		})

		Convey("Subject: SQLUnaryMinusExpr", func() {
			tests := []test{
				test{"-10", SQLInt(-10)},
			}

			runTests(evalCtx, tests)
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

		evalCtx := &EvalCtx{Rows{}, nil}

		tests := [][]interface{}{
			[]interface{}{SQLInt(124), true},
			[]interface{}{SQLFloat(1235), true},
			[]interface{}{SQLString("512"), true},
			[]interface{}{SQLInt(0), false},
			[]interface{}{SQLFloat(0), false},
			[]interface{}{SQLString("000"), false},
			[]interface{}{SQLString("skdjbkjb"), false},
			[]interface{}{SQLString(""), false},
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

func TestOptimizeSQLExpr(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   SQLExpr
	}

	runTests := func(tests []test) {
		schema, err := schema.ParseSchemaData(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be optimized to %q", t.sql, t.expected), func() {
				e, err := getWhereSQLExprFromSQL(schema, "SELECT * FROM bar WHERE "+t.sql)
				So(err, ShouldBeNil)
				result, err := OptimizeSQLExpr(e)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	Convey("Subject: OptimizeSQLExpr", t, func() {

		tests := []test{
			test{"3 = a", "a = 3", &SQLEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 < a", "a > 3", &SQLGreaterThanExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 <= a", "a >= 3", &SQLGreaterThanOrEqualExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 > a", "a < 3", &SQLLessThanExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 >= a", "a <= 3", &SQLLessThanOrEqualExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 <> a", "a <> 3", &SQLNotEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 + 3 = 6", "true", SQLTrue},
			test{"3 / (3 - 2) = a", "a = 3", &SQLEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLFloat(3)}},
			test{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "a = 3", &SQLEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 / (3 - 2) = a AND 4 - 2 = b", "a = 3 AND b = 2",
				&SQLAndExpr{
					&SQLEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLFloat(3)},
					&SQLEqualsExpr{SQLColumnExpr{"bar", "b"}, SQLInt(2)}}},
			test{"3 + 3 = 6 OR a = 3", "true", SQLTrue},
			test{"3 + 3 = 5 OR a = 3", "a = 3", &SQLEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"3 + 3 = 5 AND a = 3", "false", SQLFalse},
			test{"3 + 3 = 6 AND a = 3", "a = 3", &SQLEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
			test{"a = (~1 + 1 + (+4))", "a = 3", &SQLEqualsExpr{SQLColumnExpr{"bar", "a"}, SQLInt(3)}},
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
		schema, err := schema.ParseSchemaData(testSchema3)

		lookupFieldName := createFieldNameLookup(schema.Databases["test"])

		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be translated to \"%s\"", t.sql, t.expected), func() {
				e, err := getWhereSQLExprFromSQL(schema, "SELECT * FROM bar WHERE "+t.sql)
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
		schema, err := schema.ParseSchemaData(testSchema3)
		lookupFieldName := createFieldNameLookup(schema.Databases["test"])
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be translated to \"%s\" and locally evaluate %q", t.sql, t.expected, t.localDesc), func() {
				e, err := getWhereSQLExprFromSQL(schema, "SELECT * FROM bar WHERE "+t.sql)
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
			partialTest{"a = 3 AND a < b", `{"a":3}`, "a < b", &SQLLessThanExpr{SQLColumnExpr{"bar", "a"}, SQLColumnExpr{"bar", "b"}}},
			partialTest{"a = 3 AND a < b AND b = 4", `{"$and":[{"a":3},{"b":4}]}`, "a < b", &SQLLessThanExpr{SQLColumnExpr{"bar", "a"}, SQLColumnExpr{"bar", "b"}}},
			partialTest{"a < b AND a = 3", `{"a":3}`, "a < b", &SQLLessThanExpr{SQLColumnExpr{"bar", "a"}, SQLColumnExpr{"bar", "b"}}},
			partialTest{"NOT (a = 3 AND a < b)", `{"a":{"$ne":3}}`, "NOT a < b", &SQLNotExpr{&SQLLessThanExpr{SQLColumnExpr{"bar", "a"}, SQLColumnExpr{"bar", "b"}}}},
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
		schema, err := schema.ParseSchemaData(testSchema3)
		lookupFieldName := createFieldNameLookup(schema.Databases["test"])
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be translated to \"%s\"", t.sql, t.expected), func() {
				e, err := getWhereSQLExprFromSQL(schema, "SELECT * FROM bar WHERE "+t.sql)
				So(err, ShouldBeNil)
				match, ok := TranslateExpr(e, lookupFieldName)
				So(ok, ShouldBeTrue)
				jsonResult, err := json.Marshal(match)
				So(err, ShouldBeNil)
				So(string(jsonResult), ShouldEqual, t.expected)
			})
		}
	}

	Convey("Subject: TranslateExpr", t, func() {

		tests := []test{
			test{"abs(a)", `{"$abs":"$a"}`},
			test{"concat(a, 'funny')", `{"$concat":["$a",{"$literal":"funny"}]}`},
			test{"dayname(a)", `{"$arrayElemAt":[["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"],{"$subtract":[{"$dayOfWeek":"$a"},1]}]}`},
			test{"dayofmonth(a)", `{"$dayOfMonth":"$a"}`},
			test{"dayofweek(a)", `{"$dayOfWeek":"$a"}`},
			test{"dayofyear(a)", `{"$dayOfYear":"$a"}`},
			test{"hour(a)", `{"$hour":"$a"}`},
			test{"lcase(a)", `{"$toLower":"$a"}`},
			test{"minute(a)", `{"$minute":"$a"}`},
			test{"second(a)", `{"$second":"$a"}`},
			test{"substring(a, 2)", `{"$substr":["$a",{"$literal":2},-1]}`},
			test{"substring(a, 2, 4)", `{"$substr":["$a",{"$literal":2},{"$literal":4}]}`},
			test{"month(a)", `{"$month":"$a"}`},
			test{"week(a)", `{"$week":"$a"}`},
			test{"ucase(a)", `{"$toUpper":"$a"}`},
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
}
