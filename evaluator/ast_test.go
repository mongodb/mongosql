package evaluator

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEvaluates(t *testing.T) {

	type test struct {
		sql    string
		result SQLExpr
	}

	runTests := func(ctx *EvalCtx, tests []test) {
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be %v", t.sql, t.result), func() {
				subject, err := getWhereSQLExprFromSQL("SELECT * FROM bar WHERE " + t.sql)
				So(err, ShouldBeNil)
				result, err := subject.Evaluate(ctx)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, t.result)
			})
		}
	}

	Convey("Subject: Evaluates", t, func() {
		evalCtx := &EvalCtx{
			[]Row{{
				Data: []TableRow{{
					"bar",
					Values{
						{"a", "a", 123},
						{"b", "b", 456},
						{"c", "c", nil}},
					nil}}}},
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
				test{"-1 / 1", SQLInt(-1)},
				test{"100 / 10", SQLInt(10)},
				test{"-10 / 10", SQLInt(-1)},
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

		Convey("Subject: SQLFieldExpr", func() {
			Convey("Should return the value of the field when it exists", func() {
				subject := SQLFieldExpr{"bar", "a"}
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, SQLInt(123))
			})

			Convey("Should return nil when the field is null", func() {
				subject := SQLFieldExpr{"bar", "c"}
				result, err := subject.Evaluate(evalCtx)
				So(err, ShouldBeNil)
				So(result, ShouldHaveSameTypeAs, SQLNull)
			})

			Convey("Should return nil when the field doesn't exists", func() {
				subject := SQLFieldExpr{"bar", "no_existy"}
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
				So(result, ShouldHaveSameTypeAs, SQLValues{})
				resultValues := result.(SQLValues)
				So(resultValues[0], ShouldEqual, SQLInt(10))
				So(resultValues[1], ShouldEqual, SQLInt(42))
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

		evalCtx := &EvalCtx{[]Row{}, nil}

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

func TestPartialEvaluation(t *testing.T) {
	Convey("Subject: Partial Evaluation", t, func() {

		Convey("Should partially evaluate an entire tree", func() {
			expr, err := getWhereSQLExprFromSQL("SELECT * FROM bar WHERE 3 + 3 = 6")
			So(err, ShouldBeNil)
			result, err := PartiallyEvaluate(expr)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, SQLTrue)
		})

		Convey("Should partially evaluate a subtree recursively", func() {
			expr, err := getWhereSQLExprFromSQL("SELECT * FROM bar WHERE 3 / (3 - 2) = a")
			So(err, ShouldBeNil)
			result, err := PartiallyEvaluate(expr)
			So(err, ShouldBeNil)
			So(result, ShouldHaveSameTypeAs, &SQLEqualsExpr{})
			eq := result.(*SQLEqualsExpr)
			So(eq.left, ShouldEqual, SQLInt(3))
			So(eq.right, ShouldHaveSameTypeAs, SQLFieldExpr{})
		})

		Convey("Should partially evaluate multiple subtrees", func() {
			expr, err := getWhereSQLExprFromSQL("SELECT * FROM bar WHERE 3 / (3 - 2) = a AND 4 - 2 = b")
			So(err, ShouldBeNil)
			result, err := PartiallyEvaluate(expr)
			So(err, ShouldBeNil)

			So(result, ShouldHaveSameTypeAs, &SQLAndExpr{})
			and := result.(*SQLAndExpr)

			So(and.left, ShouldHaveSameTypeAs, &SQLEqualsExpr{})
			leq := and.left.(*SQLEqualsExpr)
			So(leq.left, ShouldEqual, SQLInt(3))
			So(leq.right, ShouldHaveSameTypeAs, SQLFieldExpr{})

			So(and.right, ShouldHaveSameTypeAs, &SQLEqualsExpr{})
			req := and.right.(*SQLEqualsExpr)
			So(req.left, ShouldEqual, SQLInt(2))
			So(req.right, ShouldHaveSameTypeAs, SQLFieldExpr{})
		})
	})
}
