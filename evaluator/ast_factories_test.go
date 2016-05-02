package evaluator

import (
	"testing"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

var (
	testCtx = &EvalCtx{Rows{
		{TableRows{{tableOneName, Values{{"a", "a", 1}, {"b", "b", 1}}}}},
		{TableRows{{tableOneName, Values{{"a", "a", 2}, {"b", "b", 2}}}}},
		{TableRows{{tableOneName, Values{{"a", "a", 3}, {"b", "b", 3}}}}},
		{TableRows{{tableOneName, Values{{"a", "a", 4}, {"b", "b", 1}}}}},
	},
		nil,
	}
)

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

func TestNewSQLExpr(t *testing.T) {

	schema, err := schema.New(testSchema3)
	if err != nil {
		panic(err)
	}
	tables := schema.Databases[dbOne].Tables
	columnTypeA := getColumnType(tables, "bar", "a")
	columnTypeB := getColumnType(tables, "bar", "b")

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
		expr, err := NewSQLExpr(sqlValue, tables)
		So(err, ShouldBeNil)
		_, ok := expr.(*SQLAggFunctionExpr)
		So(ok, ShouldBeTrue)
	})

	Convey("Calling NewSQLExpr on a single-valued tuple expression should return that value", t, func() {
		sqlValue := sqlparser.ValTuple{
			&sqlparser.ColName{
				Name:      []byte("a"),
				Qualifier: []byte(tableOneName),
			},
		}
		expr, err := NewSQLExpr(sqlValue, tables)
		So(err, ShouldBeNil)
		_, ok := expr.(SQLColumnExpr)
		So(ok, ShouldBeTrue)
	})

	Convey("Simple WHERE with explicit table names", t, func() {
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where bar.a = 4")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLEqualsExpr{SQLColumnExpr{"bar", "a", *columnTypeA}, SQLInt(4)})
	})
	Convey("Simple WHERE with implicit table names", t, func() {
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where a = 4")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLEqualsExpr{SQLColumnExpr{"bar", "a", *columnTypeA}, SQLInt(4)})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where NOT((a = 3) AND (b>1 OR a<8))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLNotExpr{
			&SQLAndExpr{
				&SQLEqualsExpr{SQLColumnExpr{"bar", "a", *columnTypeA}, SQLInt(3)},
				&SQLOrExpr{
					&SQLGreaterThanExpr{SQLColumnExpr{"bar", "b", *columnTypeB}, SQLInt(1)},
					&SQLLessThanExpr{SQLColumnExpr{"bar", "a", *columnTypeA}, SQLInt(8)},
				},
			},
		})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where NOT((a=4) AND (b>13 OR a<5))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLNotExpr{
			&SQLAndExpr{
				&SQLEqualsExpr{SQLColumnExpr{"bar", "a", *columnTypeA}, SQLInt(4)},
				&SQLOrExpr{
					&SQLGreaterThanExpr{SQLColumnExpr{"bar", "b", *columnTypeB}, SQLInt(13)},
					&SQLLessThanExpr{SQLColumnExpr{"bar", "a", *columnTypeA}, SQLInt(5)},
				},
			},
		})
	})
}

func TestAggFuncSum(t *testing.T) {
	Convey("When evaluating a 'sum' aggregation function", t, func() {

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

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
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(10))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("sumd")
			expr, err := NewSQLExpr(sqlValue, tables)
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

			_, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned even with unsummable values", func() {
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				TableRows{{tableOneName, Values{{"a", "a", "unsummable value"}}}},
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

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

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
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLFloat(2.5))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("avgd")
			expr, err := NewSQLExpr(sqlValue, tables)
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

			_, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned even with unsummable values", func() {
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				TableRows{{tableOneName, Values{{"a", "a", "nsummable value"}}}},
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

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

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
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("countd")
			expr, err := NewSQLExpr(sqlValue, tables)
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

			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("nil values should be skipped", func() {
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				TableRows{{tableOneName, Values{{"a", "a", nil}}}},
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

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

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
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("maxd")
			expr, err := NewSQLExpr(sqlValue, tables)
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

			_, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned in the presence of nil values", func() {
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				TableRows{{tableOneName, Values{{"a", "a", nil}}}},
			}

			evalCtx.Rows = append(evalCtx.Rows, unsummableRow)
			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})
	})
}

func TestScalarFuncIsNull(t *testing.T) {
	Convey("When evaluating the isnull function", t, func() {

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("isnull"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		Convey("only return 1 if the value is null", func() {
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(0))

			ctx := &EvalCtx{Rows{{TableRows{{tableOneName, Values{{"a", "a", nil}}}}}}, nil}
			value, err = funcExpr.Evaluate(ctx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(1))

		})
	})
}

func TestScalarFuncNot(t *testing.T) {
	Convey("When evaluating the not function", t, func() {

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("not"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		Convey("only return true if the value is false", func() {

			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(0))

			ctx := &EvalCtx{Rows{{TableRows{{tableOneName, Values{{"a", "a", false}}}}}}, nil}

			value, err = funcExpr.Evaluate(ctx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(1))

		})
	})
}

func TestScalarFuncPow(t *testing.T) {
	Convey("When evaluating the pow function", t, func() {

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

		sqlValue := &sqlparser.FuncExpr{
			Name: []byte("pow"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("a"),
						Qualifier: []byte(tableOneName),
					},
				},
				&sqlparser.NonStarExpr{
					Expr: &sqlparser.ColName{
						Name:      []byte("b"),
						Qualifier: []byte(tableOneName),
					},
				},
			},
		}

		expr, err := NewSQLExpr(sqlValue, tables)
		So(err, ShouldBeNil)
		funcExpr, ok := expr.(*SQLScalarFunctionExpr)
		So(ok, ShouldBeTrue)

		ctx := &EvalCtx{Rows{{TableRows{{tableOneName, Values{{"a", "a", 4}, {"b", "b", 3}}}}}}, nil}

		Convey("return the argument raised to the specified power", func() {
			value, err := funcExpr.Evaluate(ctx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLFloat(64))

			ctx.Rows[0].Data[0].Values[0].Data = 5
			ctx.Rows[0].Data[0].Values[1].Data = 2
			value, err = funcExpr.Evaluate(ctx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLFloat(25))

		})

		Convey("return an error if the argument and/or power isn't numeric", func() {
			ctx.Rows[0].Data[0].Values[0].Data = "a"
			ctx.Rows[0].Data[0].Values[1].Data = 3
			_, err := funcExpr.Evaluate(ctx)
			So(err, ShouldNotBeNil)

			ctx.Rows[0].Data[0].Values[0].Data = 4
			ctx.Rows[0].Data[0].Values[1].Data = "a"
			_, err = funcExpr.Evaluate(ctx)
			So(err, ShouldNotBeNil)

			ctx.Rows[0].Data[0].Values[0].Data = "a"
			ctx.Rows[0].Data[0].Values[1].Data = "a"
			_, err = funcExpr.Evaluate(ctx)
			So(err, ShouldNotBeNil)

		})
	})
}

func TestScalarFuncCast(t *testing.T) {
	Convey("When evaluating the cast function", t, func() {

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

		Convey("a timestamp argument should be cast correctly", func() {
			sqlValue := &sqlparser.FuncExpr{
				Name: []byte("cast"),
				Exprs: sqlparser.SelectExprs{
					&sqlparser.NonStarExpr{
						Expr: &sqlparser.ColName{
							Name:      []byte("a"),
							Qualifier: []byte(tableOneName),
						},
						As: sqlparser.StrVal([]byte("timestamp")),
					},
				},
			}

			arg := "2006-01-02 15:04:05"
			values := Values{{"a", "a", arg}}
			ctx := &EvalCtx{Rows{{TableRows{{tableOneName, values}}}}, nil}

			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			parsed, err := time.Parse(arg, arg)
			So(err, ShouldBeNil)

			value, err := funcExpr.Evaluate(ctx)
			So(err, ShouldBeNil)

			castExpr, ok := value.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(castExpr.Time, ShouldResemble, parsed)

		})

		Convey("a string argument should be cast correctly", func() {
			sqlValue := &sqlparser.FuncExpr{
				Name: []byte("cast"),
				Exprs: sqlparser.SelectExprs{
					&sqlparser.NonStarExpr{
						Expr: &sqlparser.ColName{
							Name:      []byte("a"),
							Qualifier: []byte(tableOneName),
						},
						As: sqlparser.StrVal([]byte("varchar")),
					},
				},
			}

			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			values := Values{{"a", "a", "hello"}}
			ctx := &EvalCtx{Rows{{TableRows{{tableOneName, values}}}}, nil}
			value, err := funcExpr.Evaluate(ctx)
			So(err, ShouldBeNil)

			castExpr, ok := value.(SQLVarchar)
			So(ok, ShouldBeTrue)
			So(castExpr, ShouldResemble, SQLVarchar("hello"))

		})

		Convey("a double argument should be cast correctly", func() {
			sqlValue := &sqlparser.FuncExpr{
				Name: []byte("cast"),
				Exprs: sqlparser.SelectExprs{
					&sqlparser.NonStarExpr{
						Expr: &sqlparser.ColName{
							Name:      []byte("a"),
							Qualifier: []byte(tableOneName),
						},
						As: sqlparser.StrVal([]byte("float64")),
					},
				},
			}

			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLScalarFunctionExpr)
			So(ok, ShouldBeTrue)

			values := Values{{"a", "a", 34}}
			ctx := &EvalCtx{Rows{{TableRows{{tableOneName, values}}}}, nil}
			value, err := funcExpr.Evaluate(ctx)
			So(err, ShouldBeNil)

			castExpr, ok := value.(SQLFloat)
			So(ok, ShouldBeTrue)
			So(castExpr, ShouldResemble, SQLFloat(34))

		})
	})
}

func TestAggFuncMin(t *testing.T) {
	Convey("When evaluating a 'min' aggregation function", t, func() {

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

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
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(1))
		})

		Convey("an error should be returned if the function is misspelt", func() {
			sqlValue.Name = []byte("mind")
			expr, err := NewSQLExpr(sqlValue, tables)
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

			_, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldNotBeNil)
		})

		Convey("a correct evaluation should be returned in the presence of nil values", func() {
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			evalRows := make([]Row, len(testCtx.Rows))
			copy(evalRows, testCtx.Rows)
			evalCtx := &EvalCtx{evalRows, nil}
			unsummableRow := Row{
				TableRows{{tableOneName, Values{{"a", "a", nil}}}},
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

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

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

			expr, err := NewSQLExpr(sqlValue, tables)
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

			expr, err := NewSQLExpr(sqlValue, tables)
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

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		tables := schema.Databases[dbOne].Tables

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
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(17))
		})

		Convey("a correct evaluation should be returned for avg", func() {
			sqlValue.Name = []byte("avg")
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLFloat(4.25))
		})

		Convey("a correct evaluation should be returned for count", func() {
			sqlValue.Name = []byte("count")
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(4))
		})

		Convey("a correct evaluation should be returned for max", func() {
			sqlValue.Name = []byte("max")
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(6))
		})

		Convey("a correct evaluation should be returned for min", func() {
			sqlValue.Name = []byte("min")
			expr, err := NewSQLExpr(sqlValue, tables)
			So(err, ShouldBeNil)
			funcExpr, ok := expr.(*SQLAggFunctionExpr)
			So(ok, ShouldBeTrue)

			value, err := funcExpr.Evaluate(testCtx)
			So(err, ShouldBeNil)
			So(value, ShouldResemble, SQLInt(2))
		})

	})
}
