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
schema :
-
  db: test
  tables:
  -
     table: foo
     collection: test.foo
     columns:
     -
        name: a
        type: int
     -
        name: b
        type: int
     -
        name: c
        type: int
     -
        name: _id
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        type: int
     -
        name: b
        type: int
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
        type: int
     -
        name: d
        type: int
  -
     table: silly
     collection: foo.silly
     columns:
     -
        name: e
        type: int
     -
        name: f
        type: int
-
  db: test2
  tables:
  -
     table: foo
     collection: test2.foo
     columns:
     -
        name: name
        type: string
     -
        name: orderid
        type: int
     -
        name: _id
  -
     table: bar
     collection: test2.bar
     columns:
     -
        name: orderid
        type: int
     -
        name: amount
        type: int
     -
        name: _id
`)

	testConfig2 = []byte(`
url: localhost
log_level: vv
schema :
-
  db: test
  tables:
  -
     table: foo
     collection: test.foo
     columns:
     -
        name: a
        type: string
     -
        name: x
        type: string
     -
        name: first
        type: string
     -
        name: last
        type: string
     -
        name: age
        type: string
     -
        name: b
        type: string
  -
     table: orders
     collection: test.orders
     columns:
     -
        name: customerid
        type: string
     -
        name: customername
        type: string
     -
        name: orderid
        type: string
     -
        name: orderdate
        type: string
  -
     table: customers
     collection: test.customers
     columns:
     -
        name: customerid
        type: string
     -
        name: customername
        type: string
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        type: string
     -
        name: z
        type: string
  -
     table: baz
     collection: test.baz
     columns:
     -
        name: b
        type: string
`)

	testConfig3 = []byte(
		`
url: localhost
log_level: vv
schema :
-
  db: test
  tables:
  -
     table: bar
     collection: test.bar
     columns:
     -
        name: a
        type: int
     -
        name: b
        type: string
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
        type: int
     -
        name: d
        type: string
  -
     table: silly
     collection: foo.silly
     columns:
     -
        name: e
        type: int
     -
        name: f
        type: string
`)
)
