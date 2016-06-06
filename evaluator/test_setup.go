package evaluator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/kr/pretty"
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
     collection: foo
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
        Name: g
        MongoType: bool
        SqlName: g
        SqlType: boolean
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
     table: bar
     collection: bar
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
     collection: baz
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
     collection: bar
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
     collection: silly
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
     collection: foo
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
     collection: bar
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
     collection: foo
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
     collection: orders
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
     collection: customers
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
     collection: bar
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
     collection: baz
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
     collection: bar
     columns:
     -
        Name: a
        MongoType: int
        SqlType: numeric
     -
        Name: b
        MongoType: int
        SqlType: int
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
     collection: foo
-
  db: foo
  tables:
  -
     table: bar
     collection: bar
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
     collection: silly
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

// ShouldResembleDiffed returns a blank string if its arguments resemble each other, and returns a
// list of pretty-printed diffs between the objects if they do not match.
func ShouldResembleDiffed(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("Assertion requires 1 expected value, you provided %v", len(expected))
	}
	diffs := pretty.Diff(actual, expected[0])
	if len(diffs) == 0 {
		return "" // assertion passed
	}
	delim := "\n\t- "
	return delim + strings.Join(diffs, delim)
}
