package sqlproxy

import (
	"github.com/10gen/sqlproxy/config"
	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func checkExpectedValues(count int, values [][]interface{}, expected map[interface{}][]evaluator.SQLNumeric) {
	for _, value := range values {
		So(len(value), ShouldEqual, count)
		x, ok := expected[value[0]]
		So(ok, ShouldBeTrue)
		for j := 0; j < count-1; j++ {
			So(value[j+1], ShouldResemble, x[j])
		}
	}
}

func TestSelectWithStar(t *testing.T) {
	Convey("With a star select query", t, func() {
		Convey("result set should be returned according to the schema order", func() {
			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)

			eval, err := NewEvaluator(cfg)
			So(err, ShouldBeNil)

			session := eval.getSession()
			defer session.Close()

			collection := session.DB("test").C("simple")
			collection.DropCollection()
			So(collection.Insert(bson.M{"_id": 5, "a": 6, "b": 7}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 15, "a": 16, "c": 17}), ShouldBeNil)

			names, values, err := eval.EvalSelect("test", "select * from bar", nil, nil)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 2)

			So(names[0], ShouldEqual, "a")
			So(names[1], ShouldEqual, "b")
			So(names[2], ShouldEqual, "_id")
			So(names[3], ShouldEqual, "c")

			So(values[0][0], ShouldEqual, 6)
			So(values[0][1], ShouldEqual, 7)
			So(values[0][2], ShouldEqual, 5)
			So(values[0][3], ShouldResemble, evaluator.SQLNull)

			So(values[1][0], ShouldEqual, 16)
			So(values[1][1], ShouldResemble, evaluator.SQLNull)
			So(values[1][2], ShouldEqual, 15)
			So(values[1][3], ShouldEqual, 17)

			for _, row := range values {
				So(len(names), ShouldEqual, len(row))
			}

			Convey("result set should only contain records satisfying the WHERE clause", func() {

				names, values, err = eval.EvalSelect("test", "select * from bar where a = 16", nil, nil)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 1)
				So(len(values[0]), ShouldEqual, 4)
				So(values[0][0], ShouldResemble, evaluator.SQLInt(16))
				So(values[0][1], ShouldResemble, evaluator.SQLNull)
				So(values[0][2], ShouldResemble, evaluator.SQLInt(15))
				So(values[0][3], ShouldResemble, evaluator.SQLInt(17))

				names, values, err = eval.EvalSelect("", "select * from test.bar where a = 16", nil, nil)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 1)
				So(len(values[0]), ShouldEqual, 4)
				So(values[0][0], ShouldResemble, evaluator.SQLInt(16))
				So(values[0][1], ShouldResemble, evaluator.SQLNull)
				So(values[0][2], ShouldResemble, evaluator.SQLInt(15))
				So(values[0][3], ShouldResemble, evaluator.SQLInt(17))
			})
		})
	})
}

func TestSelectWithNonStar(t *testing.T) {

	Convey("With a non-star select query", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 5, "b": 6, "a": 7}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 3, "a": 2}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 4, "b": 4, "a": 3}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 1, "b": 4, "a": 4}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 6, "b": 5, "a": 5}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 7, "b": 6, "a": 5}), ShouldBeNil)

		Convey("selecting the fields in any order should return results as requested", func() {

			names, values, err := eval.EvalSelect("test", "select a, b, _id from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))

			So(names, ShouldResemble, []string{"a", "b", "_id"})

			names, values, err = eval.EvalSelect("test", "select bar.* from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 4)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNull)

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})

			names, values, err = eval.EvalSelect("test", "select b, a from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))

			So(names, ShouldResemble, []string{"b", "a"})

			names, values, err = eval.EvalSelect("test", "select bar.b, bar.a from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))

			So(names, ShouldResemble, []string{"b", "a"})

			names, values, err = eval.EvalSelect("test", "select a, b from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

			So(names, ShouldResemble, []string{"a", "b"})

			names, values, err = eval.EvalSelect("test", "select b, a, b from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(6))

			So(names, ShouldResemble, []string{"b", "a", "b"})

			names, values, err = eval.EvalSelect("test", "select b, A, b from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)

			So(names, ShouldResemble, []string{"b", "a", "b"})
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(6))

		})

		Convey("selecting fields with non-column names should return results as requested", func() {
			names, values, err := eval.EvalSelect("test", "select a + b, sum(a) from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(13))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(27))
		})
	})

}

func TestSelectWithAggregateFunction(t *testing.T) {

	Convey("With a non-star select query containing aggregate functions", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 1, "b": 6, "a": 7}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 6, "a": 7}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 6, "a": 7}), ShouldBeNil)

		Convey("only one result set should be returned", func() {

			names, values, err := eval.EvalSelect("test", "select count(*) from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))

		})
	})

}

func TestSelectWithAliasing(t *testing.T) {

	Convey("With a non-star select query", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 5, "b": 6, "a": 7}), ShouldBeNil)

		Convey("aliased fields should return the aliased header", func() {

			names, values, err := eval.EvalSelect("test", "select a, b as c from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "c"})
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

			names, values, err = eval.EvalSelect("test", "select a as d, b as c from bar b1", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"d", "c"})
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

		})

		Convey("aliased fields colliding with existing column names should also return the aliased header", func() {

			names, values, err := eval.EvalSelect("test", "select a, b as a from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "a"})
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

		})
	})
}

func TestSelectWithGroupBy(t *testing.T) {

	Convey("With a select query containing a GROUP BY clause", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 3, "a": 2}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 4, "b": 4, "a": 3}), ShouldBeNil)

		Convey("the result set should contain terms grouped accordingly", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("using multiple aggregation functions should produce correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, count(*), sum(bar.b) from bar group by a", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "count(*)", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(2),
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{
					evaluator.SQLInt(1),
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(1),
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(3, values, expectedValues)

			names, values, err = eval.EvalSelect("test", "select a, count(*), sum(bar.b) from bar group by 1", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "count(*)", "sum(bar.b)"})

			checkExpectedValues(3, values, expectedValues)

			names, values, err = eval.EvalSelect("test", "select a from bar group by a", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a"})

			names, values, err = eval.EvalSelect("test", "select a as zz from bar group by zz", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"zz"})

			expectedValues = map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{},
			}
			checkExpectedValues(1, values, expectedValues)

		})

		Convey("no error should be returned if the some select fields are unused in GROUP BY clause", func() {
			names, values, err := eval.EvalSelect("test", "select a, b, sum(a) from bar group by a order by a", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 3)
			So(names, ShouldResemble, []string{"a", "b", "sum(bar.a)"})

			So(len(values), ShouldEqual, 3)

		})

		Convey("using aggregation function containing other complex expressions should produce correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(a+b) from bar group by a", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.a+bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(5),
				},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{
					evaluator.SQLInt(5),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(7),
				},
			}

			checkExpectedValues(2, values, expectedValues)
		})

		Convey("using aliased aggregation function should return aliased headers", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(b) as sum from bar group by a", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("grouping by aliased term should return aliased headers", func() {

			names, values, err := eval.EvalSelect("test", "select a as f, sum(b) as sum from bar group by f", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"f", "sum"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{
					evaluator.SQLInt(3),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("grouping by aliased term referencing aliased columns should return correct results", func() {

			names, values, err := eval.EvalSelect("test", "SELECT sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`a`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM `bar` GROUP BY 3) `t0`", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"sum_a_ok"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))

		})

		Convey("grouping by aliased term referencing aliased columns with a where clause should return correct results", func() {

			names, values, err := eval.EvalSelect("test", "SELECT sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`a`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM `bar` GROUP BY 3) `t0` where havclause", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"sum_a_ok"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))

			names, values, err = eval.EvalSelect("test", "SELECT sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`a`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM `bar` GROUP BY 3) `t0` where not havclause", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 0)

		})
	})
}

func TestSelectWithHaving(t *testing.T) {

	Convey("With a select query containing a HAVING clause", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 3, "a": 2}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 4, "b": 4, "a": 3}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 5, "b": 4, "a": 4}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 6, "b": 5, "a": 5}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 7, "b": 6, "a": 5}), ShouldBeNil)

		Convey("using the same select expression aggregate function should filter the result set accordingly", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(b) from bar group by a having sum(b) > 3", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
				evaluator.SQLInt(4): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
				evaluator.SQLInt(5): []evaluator.SQLNumeric{
					evaluator.SQLInt(11),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("using a different select expression aggregate function should filter the result set accordingly", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(b) from bar group by a having count(b) > 1", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(5): []evaluator.SQLNumeric{
					evaluator.SQLInt(11),
				},
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(3),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("should work even if no group by clause exists", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(b) from bar having sum(b) > 3", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(25))
		})
	})
}

func TestSelectWithJoin(t *testing.T) {

	Convey("With a non-star select query containing a join", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collectionOne := session.DB("test").C("simple")
		collectionTwo := session.DB("test").C("simple2")

		collectionOne.DropCollection()
		collectionTwo.DropCollection()

		So(collectionOne.Insert(bson.M{"c": 1, "d": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"c": 3, "d": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"c": 5, "d": 16}), ShouldBeNil)

		So(collectionTwo.Insert(bson.M{"e": 1, "f": 12}), ShouldBeNil)
		So(collectionTwo.Insert(bson.M{"e": 3, "f": 14}), ShouldBeNil)
		So(collectionTwo.Insert(bson.M{"e": 15, "f": 16}), ShouldBeNil)

		Convey("results should contain data from each of the joined tables", func() {

			names, values, err := eval.EvalSelect("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"c", "f"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(12),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(14),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("an error should be returned if derived table has no alias", func() {

			_, _, err := eval.EvalSelect("foo", "select * from (select * from bar)", nil, nil)
			So(err, ShouldNotBeNil)

		})

		Convey("results should contain data from a derived table", func() {

			names, values, err := eval.EvalSelect("foo", "select * from (select * from bar) as derived", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"c", "d"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(2),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
				evaluator.SQLInt(5): []evaluator.SQLNumeric{
					evaluator.SQLInt(16),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("results should be correct when basic table is joined with subquery", func() {
			// Note that this relies on the left deep nested join strategy
			names, values, err := eval.EvalSelect("foo", "select * from bar join (select * from silly) as derived", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 9)

			So(names, ShouldResemble, []string{"c", "d", "e", "f"})

		})

		Convey("results should be correct when subquery is joined with subquery", func() {
			// Note that this relies on the left deep nested join strategy
			names, values, err := eval.EvalSelect("foo", "select * from (select * from bar) as a join (select * from silly) as b", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 9)

			So(names, ShouldResemble, []string{"c", "d", "e", "f"})
		})

		Convey("where clause filtering should return only matched results", func() {

			names, values, err := eval.EvalSelect("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e where t1.c > t2.e", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			So(names, ShouldResemble, []string{"c", "f"})
			checkExpectedValues(0, values, nil)

			names, values, err = eval.EvalSelect("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e where t1.c = 3", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"c", "f"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(14),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvalSelect("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e where t1.c = 3 or t2.f = 12", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"c", "f"})

			expectedValues = map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(14),
				},
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(12),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

	})

}

func TestSelectFromSubquery(t *testing.T) {

	Convey("For a select statement with data from a subquery", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 5, "b": 6, "a": 7}), ShouldBeNil)

		Convey("an error should be returned if the subquery is unaliased", func() {

			_, _, err := eval.EvalSelect("test", "select * from (select * from bar)", nil, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select * from (select * from bar) t0", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select _id as x, c as y from (select _id, c from bar) t0", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("correctly qualified outer aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select t0._id as x, c as y from (select _id, c from bar) t0", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("correctly qualified outer and inner aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select b as x, d as y from (select _id as b, c as d from bar) t0", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("invalid (or invisible) column names in outer context should fail", func() {
			_, _, err := eval.EvalSelect("test", "select da from (select * from (select _id from bar) y) x", nil, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("valid column names in outer context should pass", func() {
			names, values, err := eval.EvalSelect("test", "select _id from (select * from (select _id from bar) y) x", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"_id"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
		})

		Convey("aliased and valid column names in outer context should pass", func() {
			names, values, err := eval.EvalSelect("test", "select _id as c from (select * from (select _id from bar) y) x", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
		})

		Convey("multiply aliased and valid column names in outer context should pass", func() {
			names, values, err := eval.EvalSelect("test", "select b as d from (select _id as b from (select _id from bar) y) x", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"d"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
		})

		Convey("non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select _id, c from (select * from bar) t0", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"_id", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("incorrectly qualified aliased non-star select expressions should return the correct results in order", func() {
			_, _, err := eval.EvalSelect("test", "select bar._id as x, c as y from (select * from bar) t0", nil, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("unqualified star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select * from (select * from bar) t0", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNullValue{})
		})

	})
}

func TestSelectWithRowValue(t *testing.T) {

	Convey("With a select query containing a row value expression", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 2}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 4, "a": 3}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 4, "b": 4, "a": 4}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 5, "b": 4, "a": 5}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 6, "b": 5, "a": 6}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 7, "b": 6, "a": 7}), ShouldBeNil)

		Convey("degree 1 equality comparisons should return the correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, b from bar where (a) = 3", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))

		})

		Convey("degree 1 inequality comparisons should return the correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, b from bar where (a) > 5", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "b"})
			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(6): []evaluator.SQLNumeric{
					evaluator.SQLInt(5),
				},
				evaluator.SQLInt(7): []evaluator.SQLNumeric{
					evaluator.SQLInt(6),
				},
			}
			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a) < 2", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a) < (2)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 2 > (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 2 >= (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			// TODO: add ORDER BY clause
			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[1][1], ShouldResemble, evaluator.SQLInt(2))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 6 < (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 6 <= (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(5))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[1][1], ShouldResemble, evaluator.SQLInt(6))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 6 <> (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b"})

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 6 = (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(5))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 6 in (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(5))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where 16 in (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a) in (a)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)

			_, _, err = eval.EvalSelect("test", "select a, b from bar where (a) in 3", nil, nil)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvalSelect("test", "select a, b from bar where a in 3", nil, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("degree n equality comparisons should return the correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, b from bar where (a, b) = (3, 4)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a, b) = (3, 5)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

		})

		Convey("degree n inequality comparisons should return the correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, b from bar where (a, b) >= (3, 4)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 5)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a, b) > (3, 4)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a, b) < (4, 5)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a, b) <= (1, 2)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a, b) <> (1, 2)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where not (a, b) <> (1, 2)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvalSelect("test", "select * from bar where (b-a, a+b) = (1, 7)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvalSelect("test", "select * from bar where (a-b, a*b) > (0, 15)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvalSelect("test", "select * from bar where (a-b, a*b) > (0, 17)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 3)

			names, values, err = eval.EvalSelect("test", "select * from bar where (a+a*b, a*b) > (20, 15)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

		})

		Convey("comparisons using the IN operator should return the correct results", func() {
			_, _, err := eval.EvalSelect("test", "select a, b from bar where (a, b) in (1, 2)", nil, nil)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvalSelect("test", "select a, b from bar where a in 1", nil, nil)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvalSelect("test", "select a, b from bar where (a) in 1", nil, nil)
			So(err, ShouldNotBeNil)

			names, values, err := eval.EvalSelect("test", "select a, b from bar where a in (1)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a in (1, 2)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (a) in (1, 2)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (b) in (4)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)
		})

		Convey("comparisons using the NOT IN operator should return the correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, b from bar where (b) not in (4)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (b) not in (1, 2, 4, 5)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (b) not in (1, 2, 4, 5, 6)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where (b) not in (1, 2, 4, 5, 6) or a not in (1, 2, 3, 4, 5, 6)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
		})

		//
		//
		// additional test cases to support:
		//
		// select a, b from foo where (3) > (true) ;
		//
		//

	})
}

func TestSelectWithoutTable(t *testing.T) {
	Convey("With a select expression that references no table...", t, func() {
		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		Convey("the result set should work on just the select expressions", func() {

			names, values, err := eval.EvalSelect("test", "select 1, 3", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"1", "3"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(3))

			names, values, err = eval.EvalSelect("test", "select 2*1, 3+5", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"2*1", "3+5"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(8))

			// TODO: this works but can't test it since the execution context
			// isn't yet set
			//
			// names, values, err = eval.EvalSelect("test", "select database()", nil, nil)
			// So(err, ShouldBeNil)
			// So(len(names), ShouldEqual, 1)
			// So(len(values), ShouldEqual, 1)
			//

		})

	})
}

func TestSelectWithWhere(t *testing.T) {
	Convey("With a select expression with a WHERE clause...", t, func() {
		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 2}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 4, "a": 3}), ShouldBeNil)

		Convey("range filters should return the right results", func() {

			names, values, err := eval.EvalSelect("test", "select a, b from bar where a between 1 and 3", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(1),
				},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{
					evaluator.SQLInt(2),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a between 3 and 3", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues = map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a between 3 and 1", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			So(names, ShouldResemble, []string{"a", "b"})

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a between 1 and 2 or a between 2 and 3", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues = map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLInt(1): []evaluator.SQLNumeric{
					evaluator.SQLInt(1),
				},
				evaluator.SQLInt(2): []evaluator.SQLNumeric{
					evaluator.SQLInt(2),
				},
				evaluator.SQLInt(3): []evaluator.SQLNumeric{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a not between 1 and 2 and a not between 2 and 3", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a not between 1 and 2", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))
		})

		Convey("unary filter operators should return the right results", func() {

			names, values, err := eval.EvalSelect("test", "select a, b from bar where a = ~-1", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a = ~-1 + 1", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a = +1 + 1", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(2))

			names, values, err = eval.EvalSelect("test", "select a, b from bar where a = (~1 + 1 + (+4))", nil, nil)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))
		})
	})
}

func TestSelectWithOrderBy(t *testing.T) {

	Convey("With a select query containing a ORDER BY clause", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()

		Convey("with a single order by term, the result set should be sorted accordingly", func() {

			So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 1}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 4, "b": 10, "a": 3}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 1}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 3, "b": 2, "a": 2}), ShouldBeNil)

			names, values, err := eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a order by a", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10)})

			names, values, err = eval.EvalSelect("test", "select a from bar order by a asc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(3)})

			names, values, err = eval.EvalSelect("test", "select a from bar order by a desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1)})

			names, values, err = eval.EvalSelect("test", "select a from bar order by 1 desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1)})

			names, values, err = eval.EvalSelect("test", "select a + b as c from bar group by c order by c desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(13)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(4)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(2)})

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a order by a desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a order by sum(bar.b)", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10)})

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a order by 2", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10)})

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) as c from bar group by a order by c", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10)})

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) as c from bar group by a order by cd", nil, nil)
			So(err, ShouldNotBeNil)

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a order by 3", nil, nil)
			So(err, ShouldNotBeNil)

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a order by sum(b) asc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10)})

			names, values, err = eval.EvalSelect("test", "select * from bar order by 3", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLNullValue{}})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLNullValue{}})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(3), evaluator.SQLNullValue{}})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(4), evaluator.SQLNullValue{}})

			names, values, err = eval.EvalSelect("test", "select * from bar order by 3 desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(4), evaluator.SQLNullValue{}})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(3), evaluator.SQLNullValue{}})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLNullValue{}})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLNullValue{}})

			names, values, err = eval.EvalSelect("test", "select * from bar order by 2, 3 desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLNullValue{}})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(3), evaluator.SQLNullValue{}})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLNullValue{}})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(4), evaluator.SQLNullValue{}})

			names, values, err = eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a order by sum(b) desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2)})

		})

		Convey("with multiple order by terms, the result set should be sorted accordingly", func() {
			So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 1}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 1}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 3, "b": 2, "a": 2}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 4, "b": 10, "a": 3}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 5, "b": 3, "a": 4}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 6, "b": 3, "a": 1}), ShouldBeNil)

			names, values, err := eval.EvalSelect("test", "select a, b, sum(bar.b) from bar group by a, b order by a asc, sum(b) desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(10)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLInt(3)})

			names, values, err = eval.EvalSelect("test", "select a, b, sum(bar.b) from bar group by a, b order by a asc, sum(b) asc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLInt(3)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(10)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLInt(3)})

			names, values, err = eval.EvalSelect("test", "select a, b, sum(bar.b) from bar group by a, b order by a desc, sum(b) asc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(10)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLInt(3)})

			names, values, err = eval.EvalSelect("test", "select a, b, sum(bar.b) from bar group by a, b order by a desc, sum(b) desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(10)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLInt(3)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1)})

			names, values, err = eval.EvalSelect("test", "select a, a + b as c from bar order by c desc", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(13)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(7)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(4)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(4)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2)})

		})

	})
}

func TestSelectWithCaseExpr(t *testing.T) {

	Convey("With a select query containing a case expression clause", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()

		So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 5}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 2, "a": 6}), ShouldBeNil)

		Convey("if a case matches, the correct result should be returned", func() {

			names, values, err := eval.EvalSelect("test", "select a, (case when a > 5 then 'gt' else 'lt' end) as p from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "p"})

			expectedValues := [][]evaluator.SQLValue{
				[]evaluator.SQLValue{evaluator.SQLInt(5), evaluator.SQLString("lt")},
				[]evaluator.SQLValue{evaluator.SQLInt(1), evaluator.SQLString("lt")},
				[]evaluator.SQLValue{evaluator.SQLInt(6), evaluator.SQLString("gt")},
			}

			for i, v := range expectedValues {
				caseValue := evaluator.SQLValues{[]evaluator.SQLValue{v[1]}}
				So(values[i], ShouldResemble, []interface{}{v[0], caseValue})
			}

		})

		Convey("if no case matches, null should be returned", func() {

			names, values, err := eval.EvalSelect("test", "select a, (case when a > 15 then 'gt' end) as p from bar", nil, nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "p"})

			expectedValues := [][]evaluator.SQLValue{
				[]evaluator.SQLValue{evaluator.SQLInt(5), evaluator.SQLNullValue{}},
				[]evaluator.SQLValue{evaluator.SQLInt(1), evaluator.SQLNullValue{}},
				[]evaluator.SQLValue{evaluator.SQLInt(6), evaluator.SQLNullValue{}},
			}

			for i, v := range expectedValues {
				caseValue := evaluator.SQLValues{[]evaluator.SQLValue{v[1]}}
				So(values[i], ShouldResemble, []interface{}{v[0], caseValue})
			}

		})
	})
}

func TestSelectWithLimit(t *testing.T) {

	Convey("With a select query containing a limit expression", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()

		So(collection.Insert(bson.M{"_id": 1, "b": 1, "a": 5}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 2, "a": 6}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 4, "b": 2, "a": 6}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 5, "b": 2, "a": 6}), ShouldBeNil)

		Convey("non-integer limits and/or row counts should return an error", func() {

			_, _, err := eval.EvalSelect("test", "select * from bar limit 1,1.1", nil, nil)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvalSelect("test", "select * from bar limit 1.1,1", nil, nil)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvalSelect("test", "select * from bar limit 1.1,1.1", nil, nil)
			So(err, ShouldNotBeNil)

		})

		Convey("the number of results should match the limit", func() {

			names, values, err := eval.EvalSelect("test", "select a from bar limit 1", nil, nil)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"a"})
			So(len(values), ShouldEqual, 1)

		})

		Convey("the offset should be skip the number of records specified", func() {

			names, values, err := eval.EvalSelect("test", "select _id from bar order by _id limit 1, 1", nil, nil)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"_id"})
			So(values, ShouldResemble, [][]interface{}{[]interface{}{evaluator.SQLInt(2)}})

			names, values, err = eval.EvalSelect("test", "select _id from bar order by _id limit 3, 1", nil, nil)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"_id"})
			So(values, ShouldResemble, [][]interface{}{[]interface{}{evaluator.SQLInt(4)}})

		})
	})
}
