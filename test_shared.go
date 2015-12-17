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
  name: test
  tables:
  -
     table: bar
     collection: test.simple
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
        sqlname: c
        sqltype: int
-
  name: foo
  tables:
  -
     table: bar
     collection: test.simple
     columns:
     -
        sqlname: c
        sqltype: int
     -
        sqlname: d
        sqltype: int
  -
     table: silly
     collection: test.simple2
     columns:
     -
        sqlname: e
        sqltype: int
     -
        sqlname: f
        sqltype: int
`)
)
