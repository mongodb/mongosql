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
        type: int
     -
        name: _id

     -
        name: c
        type: int
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
     collection: test.simple2
     columns:
     -
        name: e
        type: int
     -
        name: f
        type: int
`)
