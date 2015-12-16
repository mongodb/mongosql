package evaluator

var (
	dbOne          = "test"
	dbTwo          = "test2"
	tableOneName   = "foo"
	tableTwoName   = "bar"
	tableThreeName = "baz"

	testConfig1 = []byte(`
url: localhost
log_level: vv
databases :
-
  db: test
  tables:
  -
     table: foo
     collection: test.foo
     columns:
     -
        sqlname: a
        sqltype: int
     -
        sqlname: b
        sqltype: int
     -
        sqlname: c
        sqltype: int
     -
        sqlname: _id
  -
     table: bar
     collection: test.bar
     columns:
     -
        sqlname: a
        sqltype: int
     -
        sqlname: b
        sqltype: int
     -
        sqlname: _id
-
  db: foo
  tables:
  -
     table: bar
     collection: foo.bar
     columns:
     -
        sqlname: c
        sqltype: int
     -
        sqlname: d
        sqltype: int
  -
     table: silly
     collection: foo.silly
     columns:
     -
        sqlname: e
        sqltype: int
     -
        sqlname: f
        sqltype: int
-
  db: test2
  tables:
  -
     table: foo
     collection: test2.foo
     columns:
     -
        sqlname: name
        sqltype: string
     -
        sqlname: orderid
        sqltype: int
     -
        sqlname: _id
  -
     table: bar
     collection: test2.bar
     columns:
     -
        sqlname: orderid
        sqltype: int
     -
        sqlname: amount
        sqltype: int
     -
        sqlname: _id
`)

	testConfig2 = []byte(`
url: localhost
log_level: vv
databases :
-
  db: test
  tables:
  -
     table: foo
     collection: test.foo
     columns:
     -
        sqlname: a
        sqltype: string
     -
        sqlname: x
        sqltype: string
     -
        sqlname: first
        sqltype: string
     -
        sqlname: last
        sqltype: string
     -
        sqlname: age
        sqltype: string
     -
        sqlname: b
        sqltype: string
  -
     table: orders
     collection: test.orders
     columns:
     -
        sqlname: customerid
        sqltype: string
     -
        sqlname: customername
        sqltype: string
     -
        sqlname: orderid
        sqltype: string
     -
        sqlname: orderdate
        sqltype: string
  -
     table: customers
     collection: test.customers
     columns:
     -
        sqlname: customerid
        sqltype: string
     -
        sqlname: customername
        sqltype: string
  -
     table: bar
     collection: test.bar
     columns:
     -
        sqlname: a
        sqltype: string
     -
        sqlname: z
        sqltype: string
  -
     table: baz
     collection: test.baz
     columns:
     -
        sqlname: b
        sqltype: string
`)

	testConfig3 = []byte(
		`
url: localhost
log_level: vv
databases :
-
  db: test
  tables:
  -
     table: bar
     collection: test.bar
     columns:
     -
        sqlname: a
        sqltype: int
     -
        sqlname: b
        sqltype: string
     -
        sqlname: _id
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
        sqlname: c
        sqltype: int
     -
        sqlname: d
        sqltype: string
  -
     table: silly
     collection: foo.silly
     columns:
     -
        sqlname: e
        sqltype: int
     -
        sqlname: f
        sqltype: string
`)
)
