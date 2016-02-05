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
			newV, err := NewSQLValue(v, "")
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, v)
		})

		Convey("nil value should return SQLNull", func() {
			v, err := NewSQLValue(nil, "")
			So(err, ShouldBeNil)
			So(v, ShouldResemble, SQLNull)
		})

		Convey("bson object id should return its string value", func() {
			v := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err := NewSQLValue(v, "")
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v.Hex())
		})

		Convey("string objects should return the string value", func() {
			v := "56a10dd56ce28a89a8ed6edb"
			newV, err := NewSQLValue(v, "")
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("int objects should return the int value", func() {
			v1 := int(6)
			newV, err := NewSQLValue(v1, "")
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v1)

			v2 := int32(6)
			newV, err = NewSQLValue(v2, "")
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v2)

			v3 := uint32(6)
			newV, err = NewSQLValue(v3, "")
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v3)
		})

		Convey("float objects should return the float value", func() {
			v := float64(6.3)
			newV, err := NewSQLValue(v, "")
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("time objects should return the appropriate value", func() {
			v := time.Date(2014, time.December, 31, 0, 0, 0, 0, schema.DefaultLocale)
			newV, err := NewSQLValue(v, "")
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v})

			v = time.Date(2014, time.December, 31, 10, 0, 0, 0, schema.DefaultLocale)
			newV, err = NewSQLValue(v, "")
			So(err, ShouldBeNil)

			sqlTimestamp, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTimestamp, ShouldResemble, SQLTimestamp{v})
		})
	})

	Convey("When creating a SQLValue with a column type specified calling NewSQLValue on a", t, func() {

		Convey("a SQLString/SQLVarchar column type should attempt to coerce to the SQLString type", func() {

			for _, t := range []string{schema.SQLString, schema.SQLVarchar} {

				newV, err := NewSQLValue(t, t)
				So(err, ShouldBeNil)
				So(newV, ShouldResemble, SQLString(t))

				newV, err = NewSQLValue(6, t)
				So(err, ShouldBeNil)
				So(newV, ShouldResemble, SQLString("6"))

				newV, err = NewSQLValue(6.6, t)
				So(err, ShouldBeNil)
				So(newV, ShouldResemble, SQLString("6.6"))

				newV, err = NewSQLValue(int64(6), t)
				So(err, ShouldBeNil)
				So(newV, ShouldResemble, SQLString("6"))

				_id := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
				newV, err = NewSQLValue(_id, t)
				So(err, ShouldBeNil)
				So(newV, ShouldResemble, SQLString(_id.Hex()))
			}
		})

		Convey("a SQLInt column type should attempt to coerce to the SQLInt type", func() {

			t := schema.SQLInt

			newV, err := NewSQLValue(true, t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(1))

			newV, err = NewSQLValue("6", t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(int(6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(int32(6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(int64(6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(float64(6.6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

		})

		Convey("a SQLFloat column type should attempt to coerce to the SQLFloat type", func() {

			t := schema.SQLFloat

			newV, err := NewSQLValue(true, t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(1))

			newV, err = NewSQLValue("6.6", t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6.6))

			newV, err = NewSQLValue(int(6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(int32(6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(int64(6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(float64(6.6), t)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6.6))

		})

		Convey("a SQLDate column type should attempt to coerce to the SQLDate type", func() {

			// Time type

			t := schema.SQLDate
			v1 := time.Date(2014, time.May, 11, 0, 0, 0, 0, schema.DefaultLocale)
			v2 := time.Date(2014, time.May, 11, 10, 32, 12, 0, schema.DefaultLocale)

			newV, err := NewSQLValue(v1, t)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			newV, err = NewSQLValue(v2, t)
			So(err, ShouldBeNil)

			sqlDate, ok = newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			// String type

			dates := []string{"2014-05-11", "2014-05-11 15:04:05", "2014-05-11 15:04:05.233"}

			for _, d := range dates {

				newV, err := NewSQLValue(d, t)
				So(err, ShouldBeNil)

				sqlDate, ok := newV.(SQLDate)
				So(ok, ShouldBeTrue)
				So(sqlDate, ShouldResemble, SQLDate{v1})

			}

			// invalid dates and those outside valid range
			// should return the default date
			dates = []string{"2014-12-44-44", "999-1-1", "10000-1-1"}

			for _, d := range dates {

				newV, err = NewSQLValue(d, t)
				So(err, ShouldBeNil)

				df, ok := newV.(SQLDate)
				So(ok, ShouldBeTrue)
				So(df, ShouldResemble, SQLDate{schema.DefaultTime})

			}

			v := bson.NewObjectId()
			newV, err = NewSQLValue(v, t)
			So(err, ShouldBeNil)

			sqlDate, ok = newV.(SQLDate)
			So(ok, ShouldBeTrue)

			expected := time.Date(v.Time().Year(), v.Time().Month(), v.Time().Day(), 0, 0, 0, 0, schema.DefaultLocale)
			So(sqlDate, ShouldResemble, SQLDate{expected})
		})

		Convey("a SQLTimestamp column type should attempt to coerce to the SQLTimestamp type", func() {

			// Time type

			t := schema.SQLTimestamp

			v1 := time.Date(2014, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)

			newV, err := NewSQLValue(v1, t)
			So(err, ShouldBeNil)

			sqlTs, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// String type

			newV, err = NewSQLValue("2014-05-11 15:04:05.000", t)
			So(err, ShouldBeNil)

			sqlTs, ok = newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// invalid dates and those outside valid range
			// should return the default date

			dates := []string{"2044-12-4", "1966-1-1", "43223-3223"}

			for _, d := range dates {

				newV, err = NewSQLValue(d, t)
				So(err, ShouldBeNil)

				df, ok := newV.(SQLTimestamp)
				So(ok, ShouldBeTrue)
				So(df, ShouldResemble, SQLTimestamp{schema.DefaultTime})

			}

			v := bson.NewObjectId()
			newV, err = NewSQLValue(v, t)
			So(err, ShouldBeNil)

			sqlTs, ok = newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)

			So(sqlTs, ShouldResemble, SQLTimestamp{v.Time().In(schema.DefaultLocale)})
		})

		Convey("a SQLYear column type should attempt to coerce to the SQLYear type", func() {

			// Time type

			t := schema.SQLYear

			v1 := time.Date(2014, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)

			newV, err := NewSQLValue(v1, t)
			So(err, ShouldBeNil)

			sqlYr, ok := newV.(SQLYear)
			So(ok, ShouldBeTrue)
			So(sqlYr, ShouldResemble, SQLYear{v1})

			// invalid dates and those outside valid range
			// should return the default date

			dates := []interface{}{"2044-12-4", "1966-1-1", "43223-3223", 5.5, -4, 100, 2200}

			for _, d := range dates {

				newV, err = NewSQLValue(d, t)
				So(err, ShouldBeNil)

				df, ok := newV.(SQLYear)
				So(ok, ShouldBeTrue)
				So(df, ShouldResemble, SQLYear{schema.DefaultTime})

			}

			v := bson.NewObjectId()
			newV, err = NewSQLValue(v, t)
			So(err, ShouldBeNil)

			sqlYr, ok = newV.(SQLYear)
			So(ok, ShouldBeTrue)

			So(sqlYr, ShouldResemble, SQLYear{v.Time().In(schema.DefaultLocale)})

			// valid dates should return the correct year
			testData := map[interface{}]int{
				"22": 2022,
				"00": 2000,
				"0":  2000,
				"14": 2014,
				"88": 1988,
				88:   1988,
				1:    2001,
				69:   2069,
				2155: 2155,
				1902: 1902,
			}

			for k, v := range testData {

				newV, err = NewSQLValue(k, t)
				So(err, ShouldBeNil)

				sqlYr, ok = newV.(SQLYear)
				So(ok, ShouldBeTrue)

				expected := time.Date(v, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)
				So(v, ShouldResemble, SQLYear{expected}.Time.Year())

			}

		})

	})
}

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
		schema, err := schema.ParseSchemaData(testSchema3)
		So(err, ShouldBeNil)
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where bar.a = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")})
	})
	Convey("Simple WHERE with implicit table names", t, func() {
		schema, err := schema.ParseSchemaData(testSchema3)
		So(err, ShouldBeNil)
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where a = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		schema, err := schema.ParseSchemaData(testSchema3)
		So(err, ShouldBeNil)
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where NOT((a = 'eliot') AND (b>1 OR a<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLNotExpr{
			&SQLAndExpr{
				&SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")},
				&SQLOrExpr{
					&SQLGreaterThanExpr{SQLFieldExpr{"bar", "b"}, SQLInt(1)},
					&SQLLessThanExpr{SQLFieldExpr{"bar", "a"}, SQLString("blah")},
				},
			},
		})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		schema, err := schema.ParseSchemaData(testSchema3)
		So(err, ShouldBeNil)
		matcher, err := getWhereSQLExprFromSQL(schema, "select * from bar where NOT((a = 'eliot') AND (b>13 OR a<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &SQLNotExpr{
			&SQLAndExpr{
				&SQLEqualsExpr{SQLFieldExpr{"bar", "a"}, SQLString("eliot")},
				&SQLOrExpr{
					&SQLGreaterThanExpr{SQLFieldExpr{"bar", "b"}, SQLInt(13)},
					&SQLLessThanExpr{SQLFieldExpr{"bar", "a"}, SQLString("blah")},
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

			_, err := NewSQLExpr(sqlValue)
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

			_, err := NewSQLExpr(sqlValue)
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

			_, err := NewSQLExpr(sqlValue)
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
				TableRows{{tableOneName, Values{{"a", "a", nil}}}},
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

			_, err := NewSQLExpr(sqlValue)
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
