package algebrizer

var (
	dbName         = "test"
	tableOneName   = "foo"
	tableTwoName   = "bar"
	tableThreeName = "baz"

	testConfigSimple = []byte(
		`
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
     table: bar
     collection: test.bar
     columns:
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
)
