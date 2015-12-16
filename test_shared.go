package sqlproxy

var (
	dbOne        = "test"
	tableOneName = "simple"
	tableTwoName = "simple2"

	testConfigSimple = []byte(
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
        sql_name: a
        sql_type: int
     -
        sql_name: b
        sql_type: int
     -
        sql_name: _id
     -
        sql_name: c
        sql_type: int
-
  db: foo
  tables:
  -
     table: bar
     collection: test.simple
     columns:
     -
        sql_name: c
        sql_type: int
     -
        sql_name: d
        sql_type: int
  -
     table: silly
     collection: test.simple2
     columns:
     -
        sql_name: e
        sql_type: int
     -
        sql_name: f
        sql_type: int
`)
)
