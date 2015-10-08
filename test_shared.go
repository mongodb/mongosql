package sqlproxy

var testConfigSimple = []byte(
	`
schema :
-
  url: localhost
  db: test
  tables:
  -
     table: bar
     collection: test.simple
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
        name: c
        type: string
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
     collection: test.simple2
     columns:
     -
        name: e
        type: int
     -
        name: f
        type: string
`)
