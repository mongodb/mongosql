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
        name: a
        sqltype: int
     -
        name: b
        sqltype: int
     -
        name: c
        sqltype: int
     -
        name: d.e
        sqlname: e
        sqltype: int
     -
        name: d.f
        sqlname: f
        sqltype: int
     -
        name: _id
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        sqltype: int
     -
        name: b
        sqltype: int
     -
        name: _id
  -
     table: baz
     collection: test.baz
     columns:
     -
        name: a
        sqltype: int
     -
        name: b
        sqltype: int
     -
        name: _id
-
  db: foo
  tables:
  -
     table: bar
     collection: foo.bar
     columns:
     -
        name: c
        sqltype: int
     -
        name: d
        sqltype: int
  -
     table: silly
     collection: foo.silly
     columns:
     -
        name: e
        sqltype: int
     -
        name: f
        sqltype: int
-
  db: test2
  tables:
  -
     table: foo
     collection: test2.foo
     columns:
     -
        name: name
        sqltype: string
     -
        name: orderid
        sqltype: int
     -
        name: _id
  -
     table: bar
     collection: test2.bar
     columns:
     -
        name: orderid
        sqltype: int
     -
        name: amount
        sqltype: int
     -
        name: _id
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
        name: a
        sqltype: string
     -
        name: x
        sqltype: string
     -
        name: first
        sqltype: string
     -
        name: last
        sqltype: string
     -
        name: age
        sqltype: string
     -
        name: b
        sqltype: string
  -
     table: orders
     collection: test.orders
     columns:
     -
        name: customerid
        sqltype: string
     -
        name: customername
        sqltype: string
     -
        name: orderid
        sqltype: string
     -
        name: orderdate
        sqltype: string
  -
     table: customers
     collection: test.customers
     columns:
     -
        name: customerid
        sqltype: string
     -
        name: customername
        sqltype: string
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        sqltype: string
     -
        name: z
        sqltype: string
  -
     table: baz
     collection: test.baz
     columns:
     -
        name: b
        sqltype: string
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
        name: a
        sqltype: int
     -
        name: b
        sqltype: string
     -
        name: loc.1
        sqlname: c
        sqltype: string
     -
        name: _id
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
        name: c
        sqltype: int
     -
        name: d
        sqltype: string
  -
     table: silly
     collection: foo.silly
     columns:
     -
        name: e
        sqltype: int
     -
        name: f
        sqltype: string
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
