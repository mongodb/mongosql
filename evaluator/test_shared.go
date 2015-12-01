package evaluator

var (
	dbName         = "test"
	dbName2        = "test2"
	tableOneName   = "foo"
	tableTwoName   = "bar"
	tableThreeName = "baz"

	// from the planner package
	testConfig1 = []byte(`
schema :
-
  url: localhost
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
        type: int
     -
        name: _id
        type: int
  -
     table: foo
     collection: test.foo
-
  url: localhost
  db: foo
  tables:
  -
     table: bar
     collection: test.simple
     columns:
     -
        name: c
        type: int
     -
        name: d
        type: int
  -
     table: silly
     collection: test.simple
     columns:
     -
        name: e
        type: int
     -
        name: f
        type: int
-
  url: localhost
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
        type: int
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
        type: int
`)

	// from algebrizer package
	testConfig2 = []byte(`
schema :
-
  url: localhost
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
schema :
-
  url: localhost
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
        type: string
  -
     table: foo
     collection: test.foo
-
  url: localhost
  db: foo
  tables:
  -
     table: bar
     collection: test.simple
     columns:
     -
        name: c
        type: int
     -
        name: d
        type: string
  -
     table: silly
     collection: test.simple
     columns:
     -
        name: e
        type: int
     -
        name: f
        type: string
`)
)
