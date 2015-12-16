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
        sql_name: a
        sql_type: int
     -
        sql_name: b
        sql_type: int
     -
        sql_name: c
        sql_type: int
     -
        sql_name: _id
  -
     table: bar
     collection: test.bar
     columns:
     -
        sql_name: a
        sql_type: int
     -
        sql_name: b
        sql_type: int
     -
        sql_name: _id
-
  db: foo
  tables:
  -
     table: bar
     collection: foo.bar
     columns:
     -
        sql_name: c
        sql_type: int
     -
        sql_name: d
        sql_type: int
  -
     table: silly
     collection: foo.silly
     columns:
     -
        sql_name: e
        sql_type: int
     -
        sql_name: f
        sql_type: int
-
  db: test2
  tables:
  -
     table: foo
     collection: test2.foo
     columns:
     -
        sql_name: name
        sql_type: string
     -
        sql_name: orderid
        sql_type: int
     -
        sql_name: _id
  -
     table: bar
     collection: test2.bar
     columns:
     -
        sql_name: orderid
        sql_type: int
     -
        sql_name: amount
        sql_type: int
     -
        sql_name: _id
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
        sql_name: a
        sql_type: string
     -
        sql_name: x
        sql_type: string
     -
        sql_name: first
        sql_type: string
     -
        sql_name: last
        sql_type: string
     -
        sql_name: age
        sql_type: string
     -
        sql_name: b
        sql_type: string
  -
     table: orders
     collection: test.orders
     columns:
     -
        sql_name: customerid
        sql_type: string
     -
        sql_name: customername
        sql_type: string
     -
        sql_name: orderid
        sql_type: string
     -
        sql_name: orderdate
        sql_type: string
  -
     table: customers
     collection: test.customers
     columns:
     -
        sql_name: customerid
        sql_type: string
     -
        sql_name: customername
        sql_type: string
  -
     table: bar
     collection: test.bar
     columns:
     -
        sql_name: a
        sql_type: string
     -
        sql_name: z
        sql_type: string
  -
     table: baz
     collection: test.baz
     columns:
     -
        sql_name: b
        sql_type: string
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
        sql_name: a
        sql_type: int
     -
        sql_name: b
        sql_type: string
     -
        sql_name: _id
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
        sql_name: c
        sql_type: int
     -
        sql_name: d
        sql_type: string
  -
     table: silly
     collection: foo.silly
     columns:
     -
        sql_name: e
        sql_type: int
     -
        sql_name: f
        sql_type: string
`)
)
