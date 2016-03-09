package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"os"
	"testing"
)

var (
	dbOne               = "test"
	dbTwo               = "test2"
	tableOneName        = "foo"
	tableTwoName        = "bar"
	tableThreeName      = "baz"
	SSLTestKey          = "SQLPROXY_SSLTEST"
	NoOptimize          = "SQLPROXY_OPTIMIZE_OFF"
	TestSchemaSSLPrefix = []byte(`
ssl:
  allow_invalid_certs: true
  pem_key_file: "testdata/client.pem"
`)

	testSchema1 = []byte(`
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
	var sch1, sch3 []byte
	// ssl is turned on
	if len(os.Getenv(SSLTestKey)) > 0 {
		t.Logf("Testing with SSL turned on.")
		sch1 = []byte(string(TestSchemaSSLPrefix) + string(testSchema1))
		sch3 = []byte(string(TestSchemaSSLPrefix) + string(testSchema3))
	} else {
		sch1 = testSchema1
		sch3 = testSchema3
	}

	cfgOne, err := schema.ParseSchemaData(sch1)
	if err != nil {
		t.Fatalf("error parsing config1: %v", err)
		return nil
	}

	cfgThree, err := schema.ParseSchemaData(sch3)
	if err != nil {
		t.Fatalf("error parsing config3: %v", err)
		return nil
	}
	return &testEnv{cfgOne, cfgThree}
}
