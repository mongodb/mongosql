package evaluator

import (
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	testCtx = &EvalCtx{[]Row{
		{[]TableRow{{tableOneName, Values{{"a", "a", 1}, {"b", "b", 1}}, nil}}},
		{[]TableRow{{tableOneName, Values{{"a", "a", 2}, {"b", "b", 2}}, nil}}},
		{[]TableRow{{tableOneName, Values{{"a", "a", 3}, {"b", "b", 3}}, nil}}},
		{[]TableRow{{tableOneName, Values{{"a", "a", 4}, {"b", "b", 1}}, nil}}},
	},
		nil,
	}
)

func TestNewSQLExpr(t *testing.T) {
	Convey("Calling NewSQLExpr on a sqlparser expression should produce the appropriate Expr", t, func() {
		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("sum"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}
		expr, err := NewSQLExpr(sqlValue)
		So(err, ShouldBeNil)
		_, ok := expr.(*SQLAggFunctionExpr)
		So(ok, ShouldBeTrue)
	})

	Convey("Simple WHERE with explicit table names", t, func() {
		matcher, err := getSQLExprFromSQL("select * from bar where bar.a = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")})
	})
	Convey("Simple WHERE with implicit table names", t, func() {
		matcher, err := getSQLExprFromSQL("select * from bar where a = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getSQLExprFromSQL("select * from bar where NOT((a = 'eliot') AND (b>1 OR a<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLNotExpr{
			&SQLAndExpr{
				[]SQLExpr{
					&SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")},
					&SQLOrExpr{
						[]SQLExpr{
							&SQLGreaterThanExpr{SQLFieldExpr{"bar", "b"}, SQLInt(1)},
							&SQLLessThanExpr{SQLFieldExpr{"bar", "a"}, SQLString("blah")},
						},
					},
				},
			},
		})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getSQLExprFromSQL("select * from bar where NOT((a = 'eliot') AND (b>13 OR a<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLNotExpr{
			&SQLAndExpr{
				[]SQLExpr{
					&SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")},
					&SQLOrExpr{
						[]SQLExpr{
							&SQLGreaterThanExpr{SQLFieldExpr{"bar", "b"}, SQLInt(13)},
							&SQLLessThanExpr{SQLFieldExpr{"bar", "a"}, SQLString("blah")},
						},
					},
				},
			},
		})
	})
}

func TestAggFuncSum(t *testing.T) {
	Convey("When evaluating a 'sum' aggregation function", t, func() {

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("sum"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		Convey("a correct evaluation should be returned", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(10))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("sumd")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("an error should be returned if the a star expression is used", func() {
			sqlValue.Exprs = sqlparser.SelectExprs{
				&sqlparser.StarExpr{
					TableName: []byte(tableOneName),
				},
			}

			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned even with unsummable values", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				[]TableRow{{tableOneName, Values{{"a", "a", "unsummable value"}}, nil}},
			}
			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(10))
		})
	})
}

func TestAggFuncAvg(t *testing.T) {
	Convey("When evaluating a 'avg' aggregation function", t, func() {

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("avg"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		Convey("a correct evaluation should be returned", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLFloat(2.5))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("avgd")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("an error should be returned if the a star expression is used", func() {
			sqlValue.Exprs = sqlparser.SelectExprs{
				&sqlparser.StarExpr{
					TableName: []byte(tableOneName),
				},
			}

			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned even with unsummable values", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				[]TableRow{{tableOneName, Values{{"a", "a", "nsummable value"}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLFloat(2.5))
		})
	})
}

func TestAggFuncCount(t *testing.T) {
	Convey("When evaluating a 'count' aggregation function", t, func() {

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("count"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		Convey("a correct evaluation should be returned", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("countd")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned even with star expressions", func() {
			sqlValue.Exprs = sqlparser.SelectExprs{
				&sqlparser.StarExpr{
					TableName: []byte(tableOneName),
				},
			}

			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("nil values should be skipped", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				[]TableRow{{tableOneName, Values{{"a", "a", nil}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})
	})
}

func TestAggFuncMax(t *testing.T) {
	Convey("When evaluating a 'max' aggregation function", t, func() {

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("max"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		Convey("a correct evaluation should be returned", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("maxd")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("an error should be returned if the a star expression is used", func() {
			sqlValue.Exprs = sqlparser.SelectExprs{
				&sqlparser.StarExpr{
					TableName: []byte(tableOneName),
				},
			}

			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned in the presence of nil values", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				[]TableRow{{tableOneName, Values{{"a", "a", nil}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})
	})
}

func TestAggFuncMin(t *testing.T) {
	Convey("When evaluating a 'min' aggregation function", t, func() {

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("min"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		Convey("a correct evaluation should be returned", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(1))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("mind")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("an error should be returned if the a star expression is used", func() {
			sqlValue.Exprs = sqlparser.SelectExprs{
				&sqlparser.StarExpr{
					TableName: []byte(tableOneName),
				},
			}

			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned in the presence of nil values", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				[]TableRow{{tableOneName, Values{{"a", "a", nil}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(1))
		})
	})
}

func TestAggFuncDistinct(t *testing.T) {
	Convey("When evaluating a distinct aggregation function", t, func() {

		Convey("a correct evaluation should be returned for sum", func() {
			sqlValue := &sqlparser.FuncExpr{
				Name:     []byte("sum"),
				Distinct: true,
				Exprs: sqlparser.SelectExprs{
					&sqlparser.NonStarExpr{
						Expr: &sqlparser.ColName{
							Name:      []byte("b"),
							Qualifier: []byte(tableOneName),
						},
					},
				},
			}

			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(6))
		})

		Convey("a correct evaluation should be returned for count", func() {
			sqlValue := &sqlparser.FuncExpr{
				Name:     []byte("count"),
				Distinct: true,
				Exprs: sqlparser.SelectExprs{
					&sqlparser.NonStarExpr{
						Expr: &sqlparser.ColName{
							Name:      []byte("b"),
							Qualifier: []byte(tableOneName),
						},
					},
				},
			}

			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(3))
		})

	})
}

func TestAggFuncComplex(t *testing.T) {
	Convey("When evaluating an aggregation function with a complex expression", t, func() {

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("sum"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.BinaryExpr{
						Operator: sqlparser.AST_PLUS,
						Left: &sqlparser.ColName{
							Name:      []byte("a"),
							Qualifier: []byte(tableOneName),
						},
						Right: &sqlparser.ColName{
							Name:      []byte("b"),
							Qualifier: []byte(tableOneName),
						},
					},
				},
			},
		}

		Convey("a correct evaluation should be returned for sum", func() {
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(17))
		})

		Convey("a correct evaluation should be returned for avg", func() {
			sqlValue.Name = []byte("avg")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLFloat(4.25))
		})

		Convey("a correct evaluation should be returned for count", func() {
			sqlValue.Name = []byte("count")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("a correct evaluation should be returned for max", func() {
			sqlValue.Name = []byte("max")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(6))
		})

		Convey("a correct evaluation should be returned for min", func() {
			sqlValue.Name = []byte("min")
			expr, err := NewSQLExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(2))
		})

	})
}
