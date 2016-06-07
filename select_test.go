package sqlproxy_test

import (
	"os"
	"testing"

	. "github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	dbOne        = "test"
	tableOneName = "simple"
	tableTwoName = "simple2"
	SSLTestKey   = "SQLPROXY_SSLTEST"

	testSchemaSimple = []byte(
		`
schema:
-
  db: test
  tables:
  -
     table: bar
     collection: simple
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: d
        MongoType: int
        SqlType: int
     -
        Name: c
        MongoType: int
        SqlType: int
-
  db: foo
  tables:
  -
     table: bar
     collection: simple
     columns:
     -
        Name: c
        MongoType: int
        SqlType: int
     -
        Name: d
        MongoType: int
        SqlType: int
  -
     table: silly
     collection: simple2
     columns:
     -
        Name: e
        MongoType: int
        SqlType: int
     -
        Name: f
        MongoType: int
        SqlType: int
`)
)

type testEnv struct {
	eval          *Evaluator
	collectionOne *mgo.Collection
	collectionTwo *mgo.Collection
}

func (t testEnv) conn() evaluator.ConnectionCtx {
	return mockConnection{t.eval.Session(), t.collectionOne.Database.Name}
}

func (t testEnv) dbConn(db string) evaluator.ConnectionCtx {
	return mockConnection{t.eval.Session(), db}
}

type mockConnection struct {
	session *mgo.Session
	db      string
}

func (mc mockConnection) LastInsertId() int64 {
	return 1
}
func (mc mockConnection) RowCount() int64 {
	return 1
}
func (mc mockConnection) ConnectionId() uint32 {
	return 1
}
func (mc mockConnection) DB() string {
	return mc.db
}

func (mc mockConnection) Session() *mgo.Session {
	return mc.session
}

func setupEnv(t *testing.T) *testEnv {
	testOpts := Options{MongoURI: "localhost"}
	// ssl is turned on
	if len(os.Getenv(SSLTestKey)) > 0 {
		t.Logf("Testing with SSL turned on.")
		testOpts.MongoAllowInvalidCerts = true
		testOpts.MongoPEMFile = "testdata/client.pem"
		testOpts.MongoSSL = true
	}
	cfg, err := schema.New(testSchemaSimple)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
		return nil
	}

	eval, err := NewEvaluator(cfg, testOpts)
	if err != nil {
		t.Fatalf("failed to create evaluator: %v", err)
		return nil
	}

	collectionOne := eval.Session().DB(dbOne).C(tableOneName)
	collectionTwo := eval.Session().DB(dbOne).C(tableTwoName)
	return &testEnv{eval, collectionOne, collectionTwo}
}

func checkExpectedValues(count int, values [][]interface{}, expected map[interface{}][]evaluator.SQLExpr) {
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
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval
	Convey("With a star select query", t, func() {
		collectionOne.DropCollection()

		Convey("result set should be returned according to the schema order", func() {

			So(collectionOne.Insert(bson.M{"d": 5, "a": 6, "b": 7}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 15, "a": 16, "c": 17}), ShouldBeNil)

			names, values, err := eval.EvaluateRows("test", "select * from bar", nil, conn)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"a", "b", "d", "c"})
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 2)

			So(names[0], ShouldEqual, "a")
			So(names[1], ShouldEqual, "b")
			So(names[2], ShouldEqual, "d")
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
				names, values, err = eval.EvaluateRows("test", "select * from bar where a = 16", nil, conn)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 1)
				So(len(values[0]), ShouldEqual, 4)
				So(values[0][0], ShouldResemble, evaluator.SQLInt(16))
				So(values[0][1], ShouldResemble, evaluator.SQLNull)
				So(values[0][2], ShouldResemble, evaluator.SQLInt(15))
				So(values[0][3], ShouldResemble, evaluator.SQLInt(17))

				names, values, err = eval.EvaluateRows("", "select * from test.bar where a = 16", nil, conn)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 1)
				So(len(values[0]), ShouldEqual, 4)
				So(values[0][0], ShouldResemble, evaluator.SQLInt(16))
				So(values[0][1], ShouldResemble, evaluator.SQLNull)
				So(values[0][2], ShouldResemble, evaluator.SQLInt(15))
				So(values[0][3], ShouldResemble, evaluator.SQLInt(17))

				// A string literal should be usable as a matcher. "1" evaluates to true
				names, values, err = eval.EvaluateRows("", "select * from test.bar where '1'", nil, conn)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 2)

				// A string that can't be converted to a non-zero integer evaluates to false
				names, values, err = eval.EvaluateRows("", "select * from test.bar where 'xxx'", nil, conn)
				So(err, ShouldBeNil)
				So(len(values), ShouldEqual, 0)
			})
		})

		Convey("containing star expressions, columns in join context should be fully expanded", func() {

			So(collectionOne.Insert(bson.M{"d": 5, "a": 6, "b": 7}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 15, "a": 16, "c": 17}), ShouldBeNil)

			names, values, err := eval.EvaluateRows("", "select * from test.bar b1, test.bar b2", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 8)
			So(len(values), ShouldEqual, 4)
		})
	})
}

func TestSelectWithNonStar(t *testing.T) {

	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval
	Convey("With a non-star select query", t, func() {

		collectionOne.DropCollection()

		So(collectionOne.Insert(bson.M{"d": 5, "b": 6, "a": 7}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 3, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 4, "b": 4, "a": 3}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 1, "b": 4, "a": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 6, "b": 5, "a": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 7, "b": 6, "a": 5}), ShouldBeNil)

		Convey("selecting the fields in any order should return results as requested", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b, d from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))

			So(names, ShouldResemble, []string{"a", "b", "d"})

			names, values, err = eval.EvaluateRows("test", "select bar.* from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 4)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNull)

			So(names, ShouldResemble, []string{"a", "b", "d", "c"})

			names, values, err = eval.EvaluateRows("test", "select b, a from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))

			So(names, ShouldResemble, []string{"b", "a"})

			names, values, err = eval.EvaluateRows("test", "select bar.b, bar.a from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))

			So(names, ShouldResemble, []string{"b", "a"})

			names, values, err = eval.EvaluateRows("test", "select a, b from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

			So(names, ShouldResemble, []string{"a", "b"})

			names, values, err = eval.EvaluateRows("test", "select b, a, b from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(6))

			So(names, ShouldResemble, []string{"b", "a", "b"})

			names, values, err = eval.EvaluateRows("test", "select b, A, b from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 7)

			So(names, ShouldResemble, []string{"b", "A", "b"})
			So(len(values[0]), ShouldEqual, 3)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(6))
		})

		Convey("selecting fields with non-column names should return results as requested", func() {
			names, values, err := eval.EvaluateRows("test", "select a + b, sum(a) from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(13))
			So(values[0][1], ShouldResemble, evaluator.SQLFloat(27))

			names, values, err = eval.EvaluateRows("test", "select 1 from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 7)

			for _, value := range values[0] {
				So(value, ShouldResemble, evaluator.SQLInt(1))
			}
		})

	})

}

func TestSelectWithAliasing(t *testing.T) {

	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval
	Convey("With a non-star select query", t, func() {

		collectionOne.DropCollection()
		So(collectionOne.Insert(bson.M{"d": 5, "b": 6, "a": 7}), ShouldBeNil)

		Convey("aliased fields should return the aliased header", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b as c from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "c"})
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

			names, values, err = eval.EvaluateRows("test", "select a as d, b as c from bar b1", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"d", "c"})
			So(len(values[0]), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

		})

		Convey("aliased fields colliding with existing column names should also return the aliased header", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b as a from bar", nil, conn)
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

func TestSelectWithDistinct(t *testing.T) {

	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval
	Convey("With a select query containing a DISTINCT clause ", t, func() {

		collectionOne.DropCollection()
		So(collectionOne.Insert(bson.M{"d": 1, "b": 2, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 1, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 3, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 4, "b": 3, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 5, "b": 2, "a": 2}), ShouldBeNil)

		Convey("distinct column references should return distinct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT distinct a, b FROM bar order by a, b", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(len(values), ShouldEqual, 4)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[1][1], ShouldResemble, evaluator.SQLInt(2))
			So(values[2][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[2][1], ShouldResemble, evaluator.SQLInt(2))
			So(values[3][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[3][1], ShouldResemble, evaluator.SQLInt(3))
		})

		Convey("distinct column references with group by should return distinct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT distinct a, b FROM bar group by a order by a, b", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(len(values), ShouldEqual, 2)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(2))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[1][1], ShouldResemble, evaluator.SQLInt(3))
		})

		Convey("math within distinct column references should return distinct results", func() {
			names, values, err := eval.EvaluateRows("test", "SELECT distinct a+b FROM bar order by a+b", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)

			So(len(values), ShouldEqual, 4)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[2][0], ShouldResemble, evaluator.SQLInt(4))
			So(values[3][0], ShouldResemble, evaluator.SQLInt(5))
		})

		Convey("using aliases with distinct column references should return distinct results", func() {
			names, values, err := eval.EvaluateRows("test", "SELECT distinct a+b as f FROM bar order by f desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(4))
			So(values[2][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[3][0], ShouldResemble, evaluator.SQLInt(2))
		})

		Convey("distinct with aggregate functions and no grouping should return a single result", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT distinct sum(a-b), sum(b), a, a+b FROM bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			So(values[0][0], ShouldResemble, evaluator.SQLFloat(-3))
			So(values[0][1], ShouldResemble, evaluator.SQLFloat(11))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][3], ShouldResemble, evaluator.SQLInt(3))
		})

		Convey("distinct with aggregate functions and grouping should return distinct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT distinct sum(a-b), sum(b), b FROM bar group by b order by b", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 3)

			So(values[0][0], ShouldResemble, evaluator.SQLFloat(0))
			So(values[0][1], ShouldResemble, evaluator.SQLFloat(1))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(1))

			So(values[1][0], ShouldResemble, evaluator.SQLFloat(-1))
			So(values[1][1], ShouldResemble, evaluator.SQLFloat(4))
			So(values[1][2], ShouldResemble, evaluator.SQLInt(2))

			So(values[2][0], ShouldResemble, evaluator.SQLFloat(-2))
			So(values[2][1], ShouldResemble, evaluator.SQLFloat(6))
			So(values[2][2], ShouldResemble, evaluator.SQLInt(3))

		})

		Convey("distinct with single distinct aggregate functions and grouping should return distinct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT distinct sum(distinct a-b), sum(b), b FROM bar group by b order by b", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 3)

			So(values[0][0], ShouldResemble, evaluator.SQLFloat(0))
			So(values[0][1], ShouldResemble, evaluator.SQLFloat(1))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(1))

			So(values[1][0], ShouldResemble, evaluator.SQLFloat(-1))
			So(values[1][1], ShouldResemble, evaluator.SQLFloat(4))
			So(values[1][2], ShouldResemble, evaluator.SQLInt(2))

			So(values[2][0], ShouldResemble, evaluator.SQLFloat(-1))
			So(values[2][1], ShouldResemble, evaluator.SQLFloat(6))
			So(values[2][2], ShouldResemble, evaluator.SQLInt(3))

		})

		Convey("distinct with multiple distinct aggregate functions and grouping should return distinct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT distinct sum(distinct a-b), sum(distinct b), b FROM bar group by b order by b", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 3)

			So(values[0][0], ShouldResemble, evaluator.SQLFloat(0))
			So(values[0][1], ShouldResemble, evaluator.SQLFloat(1))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(1))

			So(values[1][0], ShouldResemble, evaluator.SQLFloat(-1))
			So(values[1][1], ShouldResemble, evaluator.SQLFloat(2))
			So(values[1][2], ShouldResemble, evaluator.SQLInt(2))

			So(values[2][0], ShouldResemble, evaluator.SQLFloat(-1))
			So(values[2][1], ShouldResemble, evaluator.SQLFloat(3))
			So(values[2][2], ShouldResemble, evaluator.SQLInt(3))

		})
	})
}

func TestSelectWithGroupBy(t *testing.T) {

	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing a GROUP BY clause", t, func() {

		collectionOne.DropCollection()
		So(collectionOne.Insert(bson.M{"d": 1, "b": 2, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 1, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 3, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 4, "b": 3, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 5, "b": 2, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 6, "b": nil, "a": 1}), ShouldBeNil)

		Convey("the result set should contain terms grouped accordingly", func() {

			names, values, err := eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLFloat(8),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("using multiple aggregation functions should produce correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, count(*), sum(bar.b) from bar group by a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "count(*)", "sum(bar.b)"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLInt(3),
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLInt(3),
					evaluator.SQLFloat(8),
				},
			}

			checkExpectedValues(3, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a, count(*), sum(bar.b) from bar group by 1", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "count(*)", "sum(bar.b)"})

			checkExpectedValues(3, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a from bar group by a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a"})

			expectedValues = map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{},
				evaluator.SQLInt(2): []evaluator.SQLExpr{},
			}
			checkExpectedValues(1, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a as zz from bar group by zz", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"zz"})

			expectedValues = map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{},
				evaluator.SQLInt(2): []evaluator.SQLExpr{},
			}
			checkExpectedValues(1, values, expectedValues)

		})

		Convey("no error should be returned if some select fields are unused in GROUP BY clause", func() {
			names, values, err := eval.EvaluateRows("test", "select a, b, sum(a) from bar group by a order by a", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 3)
			So(names, ShouldResemble, []string{"a", "b", "sum(a)"})

			So(len(values), ShouldEqual, 2)

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLInt(2),
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLInt(3),
					evaluator.SQLFloat(6),
				},
			}

			checkExpectedValues(3, values, expectedValues)

		})

		Convey("using aggregation function containing other complex expressions should produce correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, sum(a+b) from bar group by a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(a+b)"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(5),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLFloat(14),
				},
			}

			checkExpectedValues(2, values, expectedValues)
		})

		Convey("using aliased aggregation function should return aliased headers", func() {

			names, values, err := eval.EvaluateRows("test", "select a, sum(b) as sum from bar group by a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLFloat(8),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("grouping by aliased term should return aliased headers", func() {

			names, values, err := eval.EvaluateRows("test", "select a as f, sum(b) as sum from bar group by f", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"f", "sum"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLFloat(8),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("grouping by aliased term referencing aliased columns should return correct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`a`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM `bar` GROUP BY 3) `t0`", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"sum_a_ok"})
			So(values[0][0], ShouldResemble, evaluator.SQLFloat(9))

		})

		Convey("grouping by aliased term referencing aliased columns with a where clause should return correct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`a`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM `bar` GROUP BY 3) `t0` where havclause", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"sum_a_ok"})
			So(values[0][0], ShouldResemble, evaluator.SQLFloat(9))

			names, values, err = eval.EvaluateRows("test", "SELECT sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`a`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM `bar` GROUP BY 3) `t0` where not havclause", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 0)

		})

		Convey("grouping using distinct aggregation functions should return distinct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT a, sum(distinct b+d) FROM bar GROUP BY a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLFloat(13),
				},
			}

			checkExpectedValues(2, values, expectedValues)
		})

		Convey("grouping using multiple distinct aggregation functions should return distinct results", func() {

			names, values, err := eval.EvaluateRows("test", "SELECT a, sum(distinct b+d), sum(distinct b) FROM bar GROUP BY a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(3),
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLFloat(13),
					evaluator.SQLFloat(5),
				},
			}

			checkExpectedValues(3, values, expectedValues)
		})

		Convey("count with a star parameter should count all rows", func() {

			names, values, err := eval.EvaluateRows("test", "select count(*) from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
		})

		Convey("count with a column parameter should count only rows that contain a non-nullish b", func() {

			names, values, err := eval.EvaluateRows("test", "select count(b) from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))

		})

		Convey("count with a distinct column parameter should count only non-nullish distinct values", func() {

			names, values, err := eval.EvaluateRows("test", "select count(distinct b) from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)
			So(len(values[0]), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))

		})

		Convey("should handle aliased table definitions", func() {

			names, values, err := eval.EvaluateRows("test", "select a, sum(b) from bar as b group by a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(b)"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(3),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLFloat(8),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})
	})
}

func TestSelectWithHaving(t *testing.T) {
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing a HAVING clause", t, func() {

		collectionOne.DropCollection()
		So(collectionOne.Insert(bson.M{"d": 1, "b": 1, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 3, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 4, "b": 4, "a": 3}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 5, "b": 4, "a": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 6, "b": 5, "a": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 7, "b": 6, "a": 5}), ShouldBeNil)

		Convey("using the same select expression aggregate function should filter the result set accordingly", func() {

			names, values, err := eval.EvaluateRows("test", "select a, sum(b) from bar group by a having sum(b) > 3", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(b)"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(3): []evaluator.SQLExpr{
					evaluator.SQLFloat(4),
				},
				evaluator.SQLInt(4): []evaluator.SQLExpr{
					evaluator.SQLFloat(4),
				},
				evaluator.SQLInt(5): []evaluator.SQLExpr{
					evaluator.SQLFloat(11),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("using a different select expression aggregate function should filter the result set accordingly", func() {

			names, values, err := eval.EvaluateRows("test", "select a, sum(b) from bar group by a having count(b) > 1", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "sum(b)"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(5): []evaluator.SQLExpr{
					evaluator.SQLFloat(11),
				},
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLFloat(3),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("should work even if no group by clause exists", func() {

			names, values, err := eval.EvaluateRows("test", "select a, sum(b) from bar having sum(b) > 3", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "sum(b)"})

			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLFloat(25))
		})
	})
}

func TestSelectWithJoin(t *testing.T) {
	env := setupEnv(t)
	conn := env.conn()
	eval := env.eval
	collectionOne := eval.Session().DB("foo").C(tableOneName)
	collectionTwo := eval.Session().DB("foo").C(tableTwoName)

	Convey("With a non-star select query containing a join", t, func() {

		collectionOne.DropCollection()
		collectionTwo.DropCollection()

		So(collectionOne.Insert(bson.M{"c": 1, "d": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"c": 3, "d": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"c": 5, "d": 16}), ShouldBeNil)

		So(collectionTwo.Insert(bson.M{"e": 1, "f": 12}), ShouldBeNil)
		So(collectionTwo.Insert(bson.M{"e": 3, "f": 14}), ShouldBeNil)
		So(collectionTwo.Insert(bson.M{"e": 15, "f": 16}), ShouldBeNil)

		Convey("results should contain data from each of the joined tables", func() {

			Convey("for an inner join", func() {
				names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e", nil, conn)
				So(err, ShouldBeNil)
				So(len(names), ShouldEqual, 2)
				So(len(values), ShouldEqual, 2)

				So(names, ShouldResemble, []string{"c", "f"})

				expectedValues := map[interface{}][]evaluator.SQLExpr{
					evaluator.SQLInt(1): []evaluator.SQLExpr{
						evaluator.SQLInt(12),
					},
					evaluator.SQLInt(3): []evaluator.SQLExpr{
						evaluator.SQLInt(14),
					},
				}

				checkExpectedValues(2, values, expectedValues)
			})

			Convey("for an inner join with additional criteria", func() {
				names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e AND t2.f > 12", nil, conn)
				So(err, ShouldBeNil)
				So(len(names), ShouldEqual, 2)
				So(len(values), ShouldEqual, 1)

				So(names, ShouldResemble, []string{"c", "f"})

				expectedValues := map[interface{}][]evaluator.SQLExpr{
					evaluator.SQLInt(3): []evaluator.SQLExpr{
						evaluator.SQLInt(14),
					},
				}

				checkExpectedValues(2, values, expectedValues)
			})

			Convey("for an implicit join", func() {
				names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1, silly t2 where t1.c = t2.e", nil, conn)
				So(err, ShouldBeNil)
				So(len(names), ShouldEqual, 2)
				So(len(values), ShouldEqual, 2)

				So(names, ShouldResemble, []string{"c", "f"})

				expectedValues := map[interface{}][]evaluator.SQLExpr{
					evaluator.SQLInt(1): []evaluator.SQLExpr{
						evaluator.SQLInt(12),
					},
					evaluator.SQLInt(3): []evaluator.SQLExpr{
						evaluator.SQLInt(14),
					},
				}

				checkExpectedValues(2, values, expectedValues)
			})

			Convey("for an implicit join with additional criteria", func() {
				names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1, silly t2 where t1.c = t2.e AND t2.f > 12", nil, conn)
				So(err, ShouldBeNil)
				So(len(names), ShouldEqual, 2)
				So(len(values), ShouldEqual, 1)

				So(names, ShouldResemble, []string{"c", "f"})

				expectedValues := map[interface{}][]evaluator.SQLExpr{
					evaluator.SQLInt(3): []evaluator.SQLExpr{
						evaluator.SQLInt(14),
					},
				}

				checkExpectedValues(2, values, expectedValues)
			})

			Convey("for a left join", func() {
				names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 left join silly t2 on t1.c = t2.e", nil, conn)
				So(err, ShouldBeNil)
				So(len(names), ShouldEqual, 2)
				So(len(values), ShouldEqual, 3)

				So(names, ShouldResemble, []string{"c", "f"})

				expectedValues := map[interface{}][]evaluator.SQLExpr{
					evaluator.SQLInt(1): []evaluator.SQLExpr{
						evaluator.SQLInt(12),
					},
					evaluator.SQLInt(3): []evaluator.SQLExpr{
						evaluator.SQLInt(14),
					},
					evaluator.SQLInt(5): []evaluator.SQLExpr{
						evaluator.SQLNull,
					},
				}

				checkExpectedValues(2, values, expectedValues)
			})

			Convey("for a left join with additional criteria", func() {
				names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 left join silly t2 on t1.c = t2.e AND t2.f > 12", nil, conn)
				So(err, ShouldBeNil)
				So(len(names), ShouldEqual, 2)
				So(len(values), ShouldEqual, 3)

				So(names, ShouldResemble, []string{"c", "f"})

				expectedValues := map[interface{}][]evaluator.SQLExpr{
					evaluator.SQLInt(1): []evaluator.SQLExpr{
						evaluator.SQLNull,
					},
					evaluator.SQLInt(3): []evaluator.SQLExpr{
						evaluator.SQLInt(14),
					},
					evaluator.SQLInt(5): []evaluator.SQLExpr{
						evaluator.SQLNull,
					},
				}

				checkExpectedValues(2, values, expectedValues)
			})

			Convey("for a right join", func() {
				names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 right join silly t2 on t1.c = t2.e", nil, conn)
				So(err, ShouldBeNil)
				So(len(names), ShouldEqual, 2)
				So(len(values), ShouldEqual, 3)

				So(names, ShouldResemble, []string{"c", "f"})

				expectedValues := map[interface{}][]evaluator.SQLExpr{
					evaluator.SQLInt(1): []evaluator.SQLExpr{
						evaluator.SQLInt(12),
					},
					evaluator.SQLInt(3): []evaluator.SQLExpr{
						evaluator.SQLInt(14),
					},
					evaluator.SQLNull: []evaluator.SQLExpr{
						evaluator.SQLInt(16),
					},
				}

				checkExpectedValues(2, values, expectedValues)
			})

		})

		Convey("an error should be returned if derived table has no alias", func() {

			_, _, err := eval.EvaluateRows("foo", "select * from (select * from bar)", nil, conn)
			So(err, ShouldNotBeNil)

		})

		Convey("results should contain data from a derived table", func() {

			names, values, err := eval.EvaluateRows("foo", "select * from (select * from bar) as derived", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"c", "d"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLInt(2),
				},
				evaluator.SQLInt(3): []evaluator.SQLExpr{
					evaluator.SQLInt(4),
				},
				evaluator.SQLInt(5): []evaluator.SQLExpr{
					evaluator.SQLInt(16),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("results should be correct when basic table is joined with subquery", func() {
			// Note that this relies on the left deep nested join strategy
			names, values, err := eval.EvaluateRows("foo", "select * from bar join (select * from silly) as derived", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 9)

			So(names, ShouldResemble, []string{"c", "d", "e", "f"})

		})

		Convey("results should be correct when subquery is joined with subquery", func() {
			// Note that this relies on the left deep nested join strategy
			names, values, err := eval.EvaluateRows("foo", "select * from (select * from bar) as a join (select * from silly) as b", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 9)

			So(names, ShouldResemble, []string{"c", "d", "e", "f"})
		})

		Convey("where clause filtering should return only matched results", func() {

			names, values, err := eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e where t1.c > t2.e", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			So(names, ShouldResemble, []string{"c", "f"})
			checkExpectedValues(0, values, nil)

			names, values, err = eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e where t1.c = 3", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"c", "f"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(3): []evaluator.SQLExpr{
					evaluator.SQLInt(14),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvaluateRows("foo", "select t1.c, t2.f from bar t1 join silly t2 on t1.c = t2.e where t1.c = 3 or t2.f = 12", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"c", "f"})

			expectedValues = map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(3): []evaluator.SQLExpr{
					evaluator.SQLInt(14),
				},
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLInt(12),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

	})

}

func TestSelectFromSubquery(t *testing.T) {

	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval
	Convey("For a select statement with data from a subquery", t, func() {

		collectionOne.DropCollection()
		So(collectionOne.Insert(bson.M{"d": 5, "b": 6, "a": 7}), ShouldBeNil)

		Convey("an error should be returned if the subquery is unaliased", func() {

			_, _, err := eval.EvaluateRows("test", "select * from (select * from bar)", nil, conn)
			So(err, ShouldNotBeNil)
		})

		Convey("star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvaluateRows("test", "select * from (select * from bar) t0", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b", "d", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNull)
		})

		Convey("aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvaluateRows("test", "select d as x, c as y from (select d, c from bar) t0", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNull)
		})

		Convey("correctly qualified outer aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvaluateRows("test", "select t0.d as x, c as y from (select d, c from bar) t0", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNull)
		})

		Convey("correctly qualified outer and inner aliased non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvaluateRows("test", "select b as x, d as y from (select d as b, c as d from bar) t0", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"x", "y"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNull)
		})

		Convey("invalid (or invisible) column names in outer context should fail", func() {
			_, _, err := eval.EvaluateRows("test", "select da from (select * from (select d from bar) y) x", nil, conn)
			So(err, ShouldNotBeNil)
		})

		Convey("valid column names in outer context should pass", func() {
			names, values, err := eval.EvaluateRows("test", "select d from (select * from (select d from bar) y) x", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"d"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
		})

		Convey("aliased and valid column names in outer context should pass", func() {
			names, values, err := eval.EvaluateRows("test", "select d as c from (select * from (select d from bar) y) x", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
		})

		Convey("multiply aliased and valid column names in outer context should pass", func() {
			names, values, err := eval.EvaluateRows("test", "select b as d from (select d as b from (select d from bar) y) x", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"d"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
		})

		Convey("non-star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvaluateRows("test", "select d, c from (select * from bar) t0", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"d", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][1], ShouldResemble, evaluator.SQLNull)
		})

		Convey("incorrectly qualified aliased non-star select expressions should return the correct results in order", func() {
			_, _, err := eval.EvaluateRows("test", "select bar.d as x, c as y from (select * from bar) t0", nil, conn)
			So(err, ShouldNotBeNil)
		})

		Convey("unqualified star select expressions should return the correct results in order", func() {
			names, values, err := eval.EvaluateRows("test", "select * from (select * from bar) t0", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b", "d", "c"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][2], ShouldResemble, evaluator.SQLInt(5))
			So(values[0][3], ShouldResemble, evaluator.SQLNull)
		})

	})
}

func TestSelectWithRowValue(t *testing.T) {
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing a row value expression", t, func() {

		collectionOne.DropCollection()
		So(collectionOne.Insert(bson.M{"d": 1, "b": 1, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 4, "a": 3}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 4, "b": 4, "a": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 5, "b": 4, "a": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 6, "b": 5, "a": 6}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 7, "b": 6, "a": 7}), ShouldBeNil)

		Convey("degree 1 equality comparisons should return the correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where (a) = 3", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))

		})

		Convey("degree 1 inequality comparisons should return the correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where (a) > 5", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "b"})
			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(6): []evaluator.SQLExpr{
					evaluator.SQLInt(5),
				},
				evaluator.SQLInt(7): []evaluator.SQLExpr{
					evaluator.SQLInt(6),
				},
			}
			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a) < 2", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a) < (2)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 2 > (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 2 >= (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			// TODO: add ORDER BY clause
			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[1][1], ShouldResemble, evaluator.SQLInt(2))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 6 < (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(6))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 6 <= (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(5))
			So(values[1][0], ShouldResemble, evaluator.SQLInt(7))
			So(values[1][1], ShouldResemble, evaluator.SQLInt(6))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 6 <> (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b"})

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 6 = (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(5))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 6 in (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(6))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(5))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where 16 in (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a) in (a)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)

			_, _, err = eval.EvaluateRows("test", "select a, b from bar where (a) in 3", nil, conn)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvaluateRows("test", "select a, b from bar where a in 3", nil, conn)
			So(err, ShouldNotBeNil)
		})

		Convey("degree n equality comparisons should return the correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where (a, b) = (3, 4)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a, b) = (3, 5)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

		})

		Convey("degree n inequality comparisons should return the correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where (a, b) >= (3, 4)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 5)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a, b) > (3, 4)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a, b) < (4, 5)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a, b) <= (1, 2)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a, b) <> (1, 2)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 7)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where not (a, b) <> (1, 2)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvaluateRows("test", "select * from bar where (b-a, a+b) = (1, 7)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvaluateRows("test", "select * from bar where (a-b, a*b) > (0, 15)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvaluateRows("test", "select * from bar where (a-b, a*b) > (0, 17)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 3)

			names, values, err = eval.EvaluateRows("test", "select * from bar where (a+a*b, a*b) > (20, 15)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

		})

		Convey("comparisons using the IN operator should return the correct results", func() {
			_, _, err := eval.EvaluateRows("test", "select a, b from bar where (a, b) in (1, 2)", nil, conn)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvaluateRows("test", "select a, b from bar where a in 1", nil, conn)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvaluateRows("test", "select a, b from bar where (a) in 1", nil, conn)
			So(err, ShouldNotBeNil)

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where a in (1)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a in (1, 2)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (a) in (1, 2)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (b) in (4)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)
		})

		Convey("comparisons using the NOT IN operator should return the correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where (b) not in (4)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 4)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (b) not in (1, 2, 4, 5)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (b) not in (1, 2, 4, 5, 6)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where (b) not in (1, 2, 4, 5, 6) or a not in (1, 2, 3, 4, 5, 6)", nil, conn)
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
	env := setupEnv(t)
	conn := env.conn()
	eval := env.eval

	Convey("With a select expression that references no table...", t, func() {

		Convey("the result set should work on just the select expressions", func() {

			names, values, err := eval.EvaluateRows("test", "select 1, 3", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"1", "3"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(3))

			names, values, err = eval.EvaluateRows("test", "select 2*1, 3+5", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"2*1", "3+5"})
			So(values[0][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(8))

			// TODO: this works but can't test it since the execution context
			// isn't yet set
			//
			// names, values, err = eval.EvaluateRows("test", "select database()", nil, conn)
			// So(err, ShouldBeNil)
			// So(len(names), ShouldEqual, 1)
			// So(len(values), ShouldEqual, 1)
			//

		})

	})
}

func TestSelectWithWhere(t *testing.T) {
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select expression with a WHERE clause...", t, func() {

		collectionOne.DropCollection()

		So(collectionOne.Insert(bson.M{"d": 1, "b": 1, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 4, "a": 3}), ShouldBeNil)

		Convey("range filters should return the right results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where a between 1 and 3", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLInt(1),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLInt(2),
				},
				evaluator.SQLInt(3): []evaluator.SQLExpr{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a between 3 and 3", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues = map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(3): []evaluator.SQLExpr{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a between 3 and 1", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			So(names, ShouldResemble, []string{"a", "b"})

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a between 1 and 2 or a between 2 and 3", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues = map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLInt(1),
				},
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLInt(2),
				},
				evaluator.SQLInt(3): []evaluator.SQLExpr{
					evaluator.SQLInt(4),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a not between 1 and 2 and a not between 2 and 3", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a not between 1 and 2", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))
		})

		Convey("unary filter operators should return the right results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where a = ~-1", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a = ~-1 + 1", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(1))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(1))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a = +1 + 1", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(2))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(2))

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a = (~1 + 1 + (+4))", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))
		})

		Convey("complex filter operators should return the right results", func() {
			names, values, err := eval.EvaluateRows("test", "select a, b from bar where a > 1 AND a < b", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)
			So(values[0][0], ShouldResemble, evaluator.SQLInt(3))
			So(values[0][1], ShouldResemble, evaluator.SQLInt(4))
		})
	})
}

func TestSelectWithOrderBy(t *testing.T) {

	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing a ORDER BY clause", t, func() {

		collectionOne.DropCollection()

		Convey("with a single order by term, the result set should be sorted accordingly", func() {

			So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 1}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 4, "b": 10, "a": 3}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 1, "b": 1, "a": 1}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 3, "b": 2, "a": 2}), ShouldBeNil)

			names, values, err := eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a order by a", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLFloat(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLFloat(10)})

			names, values, err = eval.EvaluateRows("test", "select a from bar order by a asc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(3)})

			names, values, err = eval.EvaluateRows("test", "select a from bar order by a desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1)})

			names, values, err = eval.EvaluateRows("test", "select a from bar order by 1 desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1)})

			names, values, err = eval.EvaluateRows("test", "select a + b as cc from bar group by cc order by cc desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"cc"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(13)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(4)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(2)})

			names, values, err = eval.EvaluateRows("test", "select b as a from bar group by a order by a desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(10)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2)})

			names, values, err = eval.EvaluateRows("test", "select b as d from bar group by d order by b desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 1)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"d"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(10)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1)})

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a order by a desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLFloat(10)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLFloat(3)})

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a order by sum(bar.b)", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLFloat(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLFloat(10)})

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a order by 2", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLFloat(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLFloat(10)})

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) as c from bar group by a order by c", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLFloat(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLFloat(10)})

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) as c from bar group by a order by cd", nil, conn)
			So(err, ShouldNotBeNil)

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a order by 3", nil, conn)
			So(err, ShouldNotBeNil)

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a order by sum(b) asc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLFloat(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLFloat(10)})

			names, values, err = eval.EvaluateRows("test", "select * from bar order by 3", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a", "b", "d", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLNull})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLNull})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(3), evaluator.SQLNull})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(4), evaluator.SQLNull})

			names, values, err = eval.EvaluateRows("test", "select * from bar order by 3 desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a", "b", "d", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(4), evaluator.SQLNull})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(3), evaluator.SQLNull})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLNull})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLNull})

			names, values, err = eval.EvaluateRows("test", "select * from bar order by 2, 3 desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 4)

			So(names, ShouldResemble, []string{"a", "b", "d", "c"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLNull})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLInt(3), evaluator.SQLNull})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLNull})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLInt(4), evaluator.SQLNull})

			names, values, err = eval.EvaluateRows("test", "select a, sum(bar.b) from bar group by a order by sum(b) desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLFloat(10)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLFloat(3)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLFloat(2)})

		})

		Convey("with multiple order by terms, the result set should be sorted accordingly", func() {
			So(collectionOne.Insert(bson.M{"d": 1, "b": 1, "a": 1}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 1}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 3, "b": 2, "a": 2}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 4, "b": 10, "a": 3}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 5, "b": 3, "a": 4}), ShouldBeNil)
			So(collectionOne.Insert(bson.M{"d": 6, "b": 3, "a": 1}), ShouldBeNil)

			names, values, err := eval.EvaluateRows("test", "select a, b, sum(bar.b) from bar group by a, b order by a asc, sum(b) desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLFloat(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLFloat(1)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLFloat(10)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLFloat(3)})

			names, values, err = eval.EvaluateRows("test", "select a, b, sum(bar.b) from bar group by a, b order by a asc, sum(b) asc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLFloat(1)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLFloat(3)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLFloat(10)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLFloat(3)})

			names, values, err = eval.EvaluateRows("test", "select a, b, sum(bar.b) from bar group by a, b order by a desc, sum(b) asc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLFloat(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLFloat(10)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLFloat(1)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLFloat(3)})

			names, values, err = eval.EvaluateRows("test", "select a, b, sum(bar.b) from bar group by a, b order by a desc, sum(b) desc", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 3)
			So(len(values), ShouldEqual, 6)

			So(names, ShouldResemble, []string{"a", "b", "sum(bar.b)"})
			So(values[0], ShouldResemble, []interface{}{evaluator.SQLInt(4), evaluator.SQLInt(3), evaluator.SQLFloat(3)})
			So(values[1], ShouldResemble, []interface{}{evaluator.SQLInt(3), evaluator.SQLInt(10), evaluator.SQLFloat(10)})
			So(values[2], ShouldResemble, []interface{}{evaluator.SQLInt(2), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[3], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(3), evaluator.SQLFloat(3)})
			So(values[4], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(2), evaluator.SQLFloat(2)})
			So(values[5], ShouldResemble, []interface{}{evaluator.SQLInt(1), evaluator.SQLInt(1), evaluator.SQLFloat(1)})

			names, values, err = eval.EvaluateRows("test", "select a, a + b as c from bar order by c desc", nil, conn)
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
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing a case expression clause", t, func() {

		collectionOne.DropCollection()

		So(collectionOne.Insert(bson.M{"d": 1, "b": 1, "a": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 2, "a": 6}), ShouldBeNil)

		Convey("if a case matches, the correct result should be returned", func() {

			names, values, err := eval.EvaluateRows("test", "select a, (case when a > 5 then 'gt' else 'lt' end) as p from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "p"})

			expectedValues := [][]evaluator.SQLValue{
				[]evaluator.SQLValue{evaluator.SQLInt(5), evaluator.SQLVarchar("lt")},
				[]evaluator.SQLValue{evaluator.SQLInt(1), evaluator.SQLVarchar("lt")},
				[]evaluator.SQLValue{evaluator.SQLInt(6), evaluator.SQLVarchar("gt")},
			}

			for i, v := range expectedValues {
				So(values[i], ShouldResemble, []interface{}{v[0], v[1]})
			}

		})

		Convey("if no case matches, null should be returned", func() {

			names, values, err := eval.EvaluateRows("test", "select a, (case when a > 15 then 'gt' end) as p from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "p"})

			expectedValues := [][]evaluator.SQLValue{
				[]evaluator.SQLValue{evaluator.SQLInt(5), evaluator.SQLNull},
				[]evaluator.SQLValue{evaluator.SQLInt(1), evaluator.SQLNull},
				[]evaluator.SQLValue{evaluator.SQLInt(6), evaluator.SQLNull},
			}

			for i, v := range expectedValues {
				So(values[i], ShouldResemble, []interface{}{v[0], v[1]})
			}

		})

		Convey("if a simple select case matches, the correct match should be returned", func() {

			names, values, err := eval.EvaluateRows("test", "select a, (case a when 1 then 'one' else 'not one' end) as p from bar", nil, conn)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "p"})

			expectedValues := [][]evaluator.SQLValue{
				[]evaluator.SQLValue{evaluator.SQLInt(5), evaluator.SQLVarchar("not one")},
				[]evaluator.SQLValue{evaluator.SQLInt(1), evaluator.SQLVarchar("one")},
				[]evaluator.SQLValue{evaluator.SQLInt(6), evaluator.SQLVarchar("not one")},
			}

			for i, v := range expectedValues {
				So(values[i], ShouldResemble, []interface{}{v[0], v[1]})
			}

		})
	})
}

func TestSelectWithLimit(t *testing.T) {

	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing a limit expression", t, func() {

		collectionOne.DropCollection()

		So(collectionOne.Insert(bson.M{"d": 1, "b": 1, "a": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 2, "b": 2, "a": 1}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 3, "b": 2, "a": 6}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 4, "b": 2, "a": 6}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"d": 5, "b": 2, "a": 6}), ShouldBeNil)

		Convey("non-integer limits and/or row counts should return an error", func() {

			_, _, err := eval.EvaluateRows("test", "select * from bar limit 1,1.1", nil, conn)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvaluateRows("test", "select * from bar limit 1.1,1", nil, conn)
			So(err, ShouldNotBeNil)

			_, _, err = eval.EvaluateRows("test", "select * from bar limit 1.1,1.1", nil, conn)
			So(err, ShouldNotBeNil)

		})

		Convey("the number of results should match the limit", func() {

			names, values, err := eval.EvaluateRows("test", "select a from bar limit 1", nil, conn)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"a"})
			So(len(values), ShouldEqual, 1)

		})

		Convey("the offset should be skip the number of records specified", func() {

			names, values, err := eval.EvaluateRows("test", "select d from bar order by d limit 1, 1", nil, conn)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"d"})
			So(values, ShouldResemble, [][]interface{}{[]interface{}{evaluator.SQLInt(2)}})

			names, values, err = eval.EvaluateRows("test", "select d from bar order by d limit 3, 1", nil, conn)
			So(err, ShouldBeNil)
			So(names, ShouldResemble, []string{"d"})
			So(values, ShouldResemble, [][]interface{}{[]interface{}{evaluator.SQLInt(4)}})

		})
	})
}

func TestSelectWithExists(t *testing.T) {
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing an exists expression", t, func() {

		collectionOne.DropCollection()

		So(collectionOne.Insert(bson.M{"a": 2, "b": 4, "c": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 1, "b": 7, "c": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 6, "b": 1, "c": 9}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 5, "b": 3, "c": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 1, "b": 6, "c": 8}), ShouldBeNil)

		Convey("singly and multiply nested exist expressions from single tables should return correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select f1.a, f1.b from bar f1 where exists (select f2.b from bar f2 where f2.b < f1.a)", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 3)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLInt(4),
				},
				evaluator.SQLInt(6): []evaluator.SQLExpr{
					evaluator.SQLInt(1),
				},
				evaluator.SQLInt(5): []evaluator.SQLExpr{
					evaluator.SQLInt(3),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select f1.a, f1.b from bar f1 where exists (select f2.b from bar f2 where exists (select f3.c from bar f3 where exists (select * from bar f4 where f4.a > f2.b and f1.c = f3.a)))", nil, conn)

			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues = map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(2): []evaluator.SQLExpr{
					evaluator.SQLInt(4),
				},
				evaluator.SQLInt(1): []evaluator.SQLExpr{
					evaluator.SQLInt(7),
				},
			}

			checkExpectedValues(2, values, expectedValues)
		})

		Convey("singly and multiply nested exist expressions from joined tables should return correct results", func() {

			names, values, err := eval.EvaluateRows("test", "select f1.a, f9.c from bar f1 join bar f9 on f9.a > f1.b + 3 where exists (select f2.b from bar f2 join bar f10 on f2.c > f10.a where exists (select f3.c from bar f3 where exists (select * from bar, bar f5 where f5.a > f2.a and f10.c < f1.a*2)) and f2.a > f1.c and f1.a > f2.b or f9.c > 8 )", nil, conn)

			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "c"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(6): []evaluator.SQLExpr{
					evaluator.SQLInt(9),
				},
			}

			checkExpectedValues(2, values, expectedValues)
		})

	})
}

func TestSelectWithSubqueryWhere(t *testing.T) {
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing a subquery in the WHERE expression", t, func() {

		collectionOne.DropCollection()

		So(collectionOne.Insert(bson.M{"a": 2, "b": 4, "c": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 1, "b": 7, "c": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 6, "b": 1, "c": 9}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 5, "b": 3, "c": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 1, "b": 6, "c": 8}), ShouldBeNil)

		Convey("subqueries returning more than one row should error out", func() {

			_, _, err := eval.EvaluateRows("test", "select * from bar where (a, b) > (select a, b from bar where a < 2)", nil, conn)
			So(err, ShouldNotBeNil)

		})

		Convey("subqueries returning one matching row should not error out", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where a = (select max(a) from bar)", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(6): []evaluator.SQLExpr{
					evaluator.SQLInt(1),
				},
			}

			checkExpectedValues(2, values, expectedValues)

			names, values, err = eval.EvaluateRows("test", "select a, b from bar where a = (select a from bar where a = 5)", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 1)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues = map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(5): []evaluator.SQLExpr{
					evaluator.SQLInt(3),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

		Convey("subqueries returning no rows should return no results", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where (a, b) > (select a, b from bar where a = 0)", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 0)

			So(names, ShouldResemble, []string{"a", "b"})

		})

		Convey("subqueries returning exactly one row should be filtered correctly", func() {

			names, values, err := eval.EvaluateRows("test", "select a, b from bar where (a, b) > (select a, b from bar where a = 2)", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 2)

			So(names, ShouldResemble, []string{"a", "b"})

			expectedValues := map[interface{}][]evaluator.SQLExpr{
				evaluator.SQLInt(6): []evaluator.SQLExpr{
					evaluator.SQLInt(1),
				},
				evaluator.SQLInt(5): []evaluator.SQLExpr{
					evaluator.SQLInt(3),
				},
			}

			checkExpectedValues(2, values, expectedValues)

		})

	})

}

func TestSelectWithSubqueryInline(t *testing.T) {
	env := setupEnv(t)
	conn := env.conn()
	collectionOne, eval := env.collectionOne, env.eval

	Convey("With a select query containing an inline subquery as a data source", t, func() {

		collectionOne.DropCollection()

		So(collectionOne.Insert(bson.M{"a": 2, "b": 4, "c": 5}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 1, "b": 7, "c": 2}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 6, "b": 1, "c": 9}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 5, "b": 3, "c": 4}), ShouldBeNil)
		So(collectionOne.Insert(bson.M{"a": 1, "b": 6, "c": 8}), ShouldBeNil)

		Convey("star expressions combined with subqueries should return the correct columns", func() {

			names, values, err := eval.EvaluateRows("test", "select * from bar f, (select max(a) as maxloss, c from bar) m", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 6)
			So(len(values), ShouldEqual, 5)

			So(names, ShouldResemble, []string{"a", "b", "d", "c", "maxloss", "c"})
		})

		Convey("filtered star expressions combined with subqueries should return the correct columns", func() {

			names, values, err := eval.EvaluateRows("test", "select f.* from bar f, (select max(a) as maxloss from bar) m", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 5)

			So(names, ShouldResemble, []string{"a", "b", "d", "c"})

			names, values, err = eval.EvaluateRows("test", "select m.* from bar f, (select max(a) as maxloss, c from bar) m", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 2)
			So(len(values), ShouldEqual, 5)

			So(names, ShouldResemble, []string{"maxloss", "c"})

			names, values, err = eval.EvaluateRows("test", "select m.*, f.* from bar f, (select max(a) as maxloss, c from bar) m", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 6)
			So(len(values), ShouldEqual, 5)

			So(names, ShouldResemble, []string{"maxloss", "c", "a", "b", "d", "c"})

			names, values, err = eval.EvaluateRows("test", "select * from bar f, (select max(a) as maxloss, c from bar) m", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 6)
			So(len(values), ShouldEqual, 5)

			So(names, ShouldResemble, []string{"a", "b", "d", "c", "maxloss", "c"})

			names, values, err = eval.EvaluateRows("test", "select * from bar join (select * from bar) as derived order by bar.c", nil, conn)
			So(err, ShouldBeNil)

			So(len(names), ShouldEqual, 8)
			So(len(values), ShouldEqual, 25)

			So(names, ShouldResemble, []string{"a", "b", "d", "c", "a", "b", "d", "c"})

		})
	})
}
