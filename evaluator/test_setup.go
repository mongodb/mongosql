package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"testing"
)

var (
	dbOne          = "test"
	dbTwo          = "test2"
	tableOneName   = "foo"
	tableTwoName   = "bar"
	tableThreeName = "baz"
	SSLTestKey     = "SQLPROXY_SSLTEST"
	NoOptimize     = "SQLPROXY_OPTIMIZE_OFF"

	testSchema1 = []byte(`
schema:
-
  db: test
  tables:
  -
     table: foo
     collection: test.foo
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
        Name: c
        MongoType: int
        SqlType: int
     -
        Name: d.e
        MongoType: int
        SqlName: e
        SqlType: int
     -
        Name: d.f
        MongoType: int
        SqlName: f
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
     table: bar
     collection: test.bar
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
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
     table: baz
     collection: test.baz
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
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
-
  db: foo
  tables:
  -
     table: bar
     collection: foo.bar
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
     collection: foo.silly
     columns:
     -
        Name: e
        MongoType: int
        SqlType: int
     -
        Name: f
        MongoType: int
        SqlType: int
-
  db: test2
  tables:
  -
     table: foo
     collection: test2.foo
     columns:
     -
        Name: name
        MongoType: int
        SqlType: varchar
     -
        Name: orderid
        MongoType: int
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
     table: bar
     collection: test2.bar
     columns:
     -
        Name: orderid
        MongoType: int
        SqlType: int
     -
        Name: amount
        MongoType: int
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
`)

	testSchema2 = []byte(`
url: localhost
log_level: vv
schema:
-
  db: test
  tables:
  -
     table: foo
     collection: test.foo
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
     -
        Name: x
        MongoType: string
        SqlType: varchar
     -
        Name: first
        MongoType: string
        SqlType: varchar
     -
        Name: last
        MongoType: string
        SqlType: varchar
     -
        Name: age
        MongoType: string
        SqlType: varchar
     -
        Name: b
        MongoType: string
        SqlType: varchar
  -
     table: orders
     collection: test.orders
     columns:
     -
        Name: customerid
        MongoType: string
        SqlType: varchar
     -
        Name: customername
        MongoType: string
        SqlType: varchar
     -
        Name: orderid
        MongoType: string
        SqlType: varchar
     -
        Name: orderdate
        MongoType: string
        SqlType: varchar
  -
     table: customers
     collection: test.customers
     columns:
     -
        Name: customerid
        MongoType: string
        SqlType: varchar
     -
        Name: customername
        MongoType: string
        SqlType: varchar
  -
     table: bar
     collection: test.bar
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
     -
        Name: z
        MongoType: string
        SqlType: varchar
  -
     table: baz
     collection: test.baz
     columns:
     -
        Name: b
        MongoType: string
        SqlType: varchar
`)

	testSchema3 = []byte(
		`
url: localhost
log_level: vv
schema:
-
  db: test
  tables:
  -
     table: bar
     collection: test.bar
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar
     -
        Name: loc.1
        MongoType: string
        SqlName: c
        SqlType: varchar
     -
        Name: g
        MongoType: date
        SqlName: g
        SqlType: timestamp
     -
        Name: h
        MongoType: date
        SqlName: h
        SqlType: date
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
     table: foo
     collection: test.foo
-
  db: foo
  tables:
  -
     table: bar
     collection: foo.bar
     columns:
     -
        Name: c
        MongoType: int
        SqlType: int
     -
        Name: d
        MongoType: string
        SqlType: varchar
  -
     table: silly
     collection: foo.silly
     columns:
     -
        Name: e
        MongoType: int
        SqlType: int
     -
        Name: f
        MongoType: string
        SqlType: varchar
`)
)

type testEnv struct {
	cfgOne   *schema.Schema
	cfgThree *schema.Schema
}

func setupEnv(t *testing.T) *testEnv {
	cfgOne := schema.Must(schema.New(testSchema1))
	cfgThree := schema.Must(schema.New(testSchema3))
	return &testEnv{cfgOne, cfgThree}
}
