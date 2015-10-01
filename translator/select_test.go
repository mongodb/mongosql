package translator

import (
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
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

			eval, err := NewEvalulator(cfg)
			So(err, ShouldBeNil)

			session := eval.getSession()
			defer session.Close()

			collection := session.DB("test").C("simple")
			collection.DropCollection()
			So(collection.Insert(bson.M{"_id": 5, "a": 6, "b": 7}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 15, "a": 16, "c": 17}), ShouldBeNil)

			names, values, err := eval.EvalSelect("test", "select * from bar", nil)
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
			So(values[0][3], ShouldResemble, evaluator.SQLNullValue{})

			So(values[1][0], ShouldEqual, 16)
			So(values[1][1], ShouldResemble, evaluator.SQLNullValue{})
			So(values[1][2], ShouldEqual, 15)
			So(values[1][3], ShouldEqual, 17)

			for _, row := range values {
				So(len(names), ShouldEqual, len(row))
			}

			Convey("result set should only contain records satisfying the WHERE clause", func() {

				names, values, err = eval.EvalSelect("test", "select * from bar where a = 16", nil)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 1)
				So(len(values[0]), ShouldEqual, 4)
				So(values[0][0], ShouldResemble, evaluator.SQLNumeric(16))
				So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
				So(values[0][2], ShouldResemble, evaluator.SQLNumeric(15))
				So(values[0][3], ShouldResemble, evaluator.SQLNumeric(17))

				names, values, err = eval.EvalSelect("", "select * from test.bar where a = 16", nil)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 1)
				So(len(values[0]), ShouldEqual, 4)
				So(values[0][0], ShouldResemble, evaluator.SQLNumeric(16))
				So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
				So(values[0][2], ShouldResemble, evaluator.SQLNumeric(15))
				So(values[0][3], ShouldResemble, evaluator.SQLNumeric(17))

			})
		})
	})
}

func TestSelectWithNonStar(t *testing.T) {

	Convey("With a non-star select query", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
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

			names, values, err := eval.EvalSelect("test", "select a, b, _id from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(6))
			So(values[0][2], ShouldResemble, evaluator.SQLNumeric(5))

			So(names, ShouldResemble, []string{"a", "b", "_id"})

			names, values, err = eval.EvalSelect("test", "select bar.* from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 4)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(6))
			So(values[0][2], ShouldResemble, evaluator.SQLNumeric(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNullValue{})

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})

			names, values, err = eval.EvalSelect("test", "select b, a from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(6))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(7))

			So(names, ShouldResemble, []string{"b", "a"})

			names, values, err = eval.EvalSelect("test", "select bar.b, bar.a from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(6))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(7))

			So(names, ShouldResemble, []string{"b", "a"})

			names, values, err = eval.EvalSelect("test", "select a, b from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(6))

			So(names, ShouldResemble, []string{"a", "b"})

			names, values, err = eval.EvalSelect("test", "select b, a, b from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(6))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][2], ShouldResemble, evaluator.SQLNumeric(6))

			So(names, ShouldResemble, []string{"b", "a", "b"})

			names, values, err = eval.EvalSelect("test", "select b, A, b from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)

			So(names, ShouldResemble, []string{"b", "a", "b"})
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(6))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][2], ShouldResemble, evaluator.SQLNumeric(6))

		})

		Convey("selecting fields with non-column names should return results as requested", func() {
			names, values, err := eval.EvalSelect("test", "select a + b, sum(a) from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(13))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(27))
		})
	})

}

func TestSelectWithAggregateFunction(t *testing.T) {

	Convey("With a non-star select query containing aggregate functions", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 1, "b": 6, "a": 7}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 2, "b": 6, "a": 7}), ShouldBeNil)
		So(collection.Insert(bson.M{"_id": 3, "b": 6, "a": 7}), ShouldBeNil)

		Convey("only one result set should be returned", func() {

			names, values, err := eval.EvalSelect("test", "select count(*) from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(3))

		})
	})

}

func TestSelectWithAliasing(t *testing.T) {

	Convey("With a non-star select query", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 5, "b": 6, "a": 7}), ShouldBeNil)

		Convey("aliased fields should return the aliased header", func() {

			names, values, err := eval.EvalSelect("test", "select a, b as c from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "c"})
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(6))

		})

		Convey("aliased fields colliding with existing column names should also return the aliased header", func() {

			names, values, err := eval.EvalSelect("test", "select a, b as a from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "a"})
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(6))

		})
	})
}

func TestSelectWithGroupBy(t *testing.T) {

	Convey("With a select query containing a GROUP BY clause", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
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

			names, values, err := eval.EvalSelect("test", "select a, sum(bar.b) from bar group by a", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(2): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("using multiple aggregation functions should produce correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, count(*), sum(bar.b) from bar group by a", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "count(*)", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(2),
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(2): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(1),
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(1),
					evaluator.SQLNumeric(4),
				},
			}

			checkExpectedValues(3, values, expectedValues)

		})

		Convey("an error should be returned if the some select fields are unused in GROUP BY clause", func() {
			_, _, err := eval.EvalSelect("test", "select a, b, sum(a) from bar group by a", nil)
			So(err, ShouldNotBeNil)
		})

		Convey("using aggregation function containing other complex expressions should produce correct results", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(a+b) from bar group by a", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.a+bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(5),
				},
				evaluator.SQLNumeric(2): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(5),
				},
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(7),
				},
			}

			checkExpectedValues(2, values, expectedValues)
		})

		Convey("using aliased aggregation function should return aliased headers", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(b) as sum from bar group by a", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(2): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("grouping by aliased term should return aliased headers", func() {

			names, values, err := eval.EvalSelect("test", "select a as f, sum(b) as sum from bar group by f", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"f", "sum"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(2): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(3),
				},
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("grouping by aliased term referencing aliased columns should return correct results", func() {

			names, values, err := eval.EvalSelect("test", "SELECT sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`a`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM `bar` GROUP BY 3) `t0`", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"sum_a_ok"})
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(7))

		})
	})
}

func TestSelectWithHaving(t *testing.T) {

	Convey("With a select query containing a HAVING clause", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
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

			names, values, err := eval.EvalSelect("test", "select a, sum(b) from bar group by a having sum(b) > 3", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(4),
				},
				evaluator.SQLNumeric(4): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(4),
				},
				evaluator.SQLNumeric(5): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(11),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("using a different select expression aggregate function should filter the result set accordingly", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(b) from bar group by a having count(b) > 1", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(5): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(11),
				},
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(3),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("should work even if no group by clause exists", func() {

			names, values, err := eval.EvalSelect("test", "select a, sum(b) from bar having sum(b) > 3", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(1))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(25))
		})
	})
}

func TestSelectWithJoin(t *testing.T) {

	Convey("With a non-star select query containing a join", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
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

			names, values, err := eval.EvalSelect("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"c", "f"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(12),
				},
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(14),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("an error should be returned if derived table has no alias", func() {

			_, _, err := eval.EvalSelect("foo", "select * from (select * from bar)", nil)
			So(err, ShouldNotBeNil)

		})

		Convey("results should contain data from a derived table", func() {

			names, values, err := eval.EvalSelect("foo", "select * from (select * from bar) as derived", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"c", "d"})

			expectedValues := map[interface{}][]evaluator.SQLNumeric{
				evaluator.SQLNumeric(1): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(2),
				},
				evaluator.SQLNumeric(3): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(4),
				},
				evaluator.SQLNumeric(5): []evaluator.SQLNumeric{
					evaluator.SQLNumeric(16),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("results should be correct when basic table is joined with subquery", func() {
			// Note that this relies on the left deep nested join strategy
			names, values, err := eval.EvalSelect("foo", "select * from bar join (select * from silly) as derived", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 9)

			So(names, ShouldResemble, []string{"c", "d", "e", "f"})

		})

		Convey("results should be correct when subquery is joined with subquery", func() {
			// Note that this relies on the left deep nested join strategy
			names, values, err := eval.EvalSelect("foo", "select * from (select * from bar) as a join (select * from silly) as b", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 9)

			So(names, ShouldResemble, []string{"c", "d", "e", "f"})
		})

	})

}

func TestSelectFromSubquery(t *testing.T) {

	Convey("For a select statement with data from a subquery", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
		So(err, ShouldBeNil)

		session := eval.getSession()
		defer session.Close()

		collection := session.DB("test").C("simple")
		collection.DropCollection()
		So(collection.Insert(bson.M{"_id": 5, "b": 6, "a": 7}), ShouldBeNil)

		Convey("an error should be returned if the subquery is unaliased", func() {

			_, _, err := eval.EvalSelect("test", "select * from (select * from bar)", nil)
			So(err, ShouldNotBeNil)
		})

		Convey("star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select * from (select * from bar) t0", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b", "_id", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(7))
			So(values[0][1], ShouldResemble, evaluator.SQLNumeric(6))
			So(values[0][2], ShouldResemble, evaluator.SQLNumeric(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select _id, c from (select * from bar) t0", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"_id", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select _id as x, c as y from (select * from bar) t0", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("correctly qualified aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvalSelect("test", "select t0._id as x, c as y from (select * from bar) t0", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLNumeric(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNullValue{})
		})

		Convey("incorrectly qualified aliased non-star select expressions should return the correct results in order", func() {
			_, _, err := eval.EvalSelect("test", "select bar._id as x, c as y from (select * from bar) t0", nil)
			So(err, ShouldNotBeNil)
		})
	})
}
