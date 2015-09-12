package planner

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
