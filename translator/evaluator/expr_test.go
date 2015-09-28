package evaluator

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/translator/types"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	testCtx = &EvalCtx{Rows: []types.Row{
		{Data: []types.TableRow{{tableOneName, bson.D{{"a", 1}, {"b", 1}}, nil}}},
		{Data: []types.TableRow{{tableOneName, bson.D{{"a", 2}, {"b", 2}}, nil}}},
		{Data: []types.TableRow{{tableOneName, bson.D{{"a", 3}, {"b", 3}}, nil}}},
		{Data: []types.TableRow{{tableOneName, bson.D{{"a", 4}, {"b", 1}}, nil}}},
	}}
)

func TestNewExpr(t *testing.T) {
	Convey("Calling NewExpr on a sqlparser expression should produce the appropriate Expr", t, func() {
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
		expr, err := NewExpr(sqlValue)
		So(err, ShouldBeNil)
		_, ok := expr.(*FuncExpr)
		So(ok, ShouldBeTrue)
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
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(10))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("sumd")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
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

			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned even with unsummable values", func() {
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]types.Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows}
			unsummableRow := types.Row{
				[]types.TableRow{{tableOneName, bson.D{{"a", "nsummable value"}}, nil}},
			}
			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(10))
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
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(2.5))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("avgd")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
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

			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned even with unsummable values", func() {
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]types.Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows}
			unsummableRow := types.Row{
				[]types.TableRow{{tableOneName, bson.D{{"a", "nsummable value"}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(2.5))
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
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(4))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("countd")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
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

			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(4))
		})

		Convey("nil values should be skipped", func() {
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]types.Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows}
			unsummableRow := types.Row{
				[]types.TableRow{{tableOneName, bson.D{{"a", nil}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(4))
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
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(4))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("maxd")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
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

			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned in the presence of nil values", func() {
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]types.Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows}
			unsummableRow := types.Row{
				[]types.TableRow{{tableOneName, bson.D{{"a", nil}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(4))
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
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(1))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("mind")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
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

			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			_, err = funcExpr.Evaluate(testCtx)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned in the presence of nil values", func() {
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]types.Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows}
			unsummableRow := types.Row{
				[]types.TableRow{{tableOneName, bson.D{{"a", nil}}, nil}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(1))
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

			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(6))
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

			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(3))
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
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(17))
		})

		Convey("a correct evaluation should be returned for avg", func() {
			sqlValue.Name = []byte("avg")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(4.25))
		})

		Convey("a correct evaluation should be returned for count", func() {
			sqlValue.Name = []byte("count")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(4))
		})

		Convey("a correct evaluation should be returned for max", func() {
			sqlValue.Name = []byte("max")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(6))
		})

		Convey("a correct evaluation should be returned for min", func() {
			sqlValue.Name = []byte("min")
			expr, err := NewExpr(sqlValue)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*FuncExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLNumeric(2))
		})

	})
}
