package sqlproxy

var (
	dbOne        = "test"
	tableOneName = "simple"
	tableTwoName = "simple2"

	testSchemaSimple = []byte(
		`
url: localhost
log_level: vv
schema:
-
  db: test
  tables:
  -
     table: bar
     collection: test.simple
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
        name: c
        sqltype: int
-
  db: foo
  tables:
  -
     table: bar
     collection: test.simple
     columns:
     -
        name: c
        sqltype: int
     -
        name: d
        sqltype: int
  -
     table: silly
     collection: test.simple2
     columns:
     -
        name: e
        sqltype: int
     -
        name: f
        sqltype: int
`)
)
