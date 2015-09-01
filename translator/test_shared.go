package translator

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
        a: int
        b: string
-
  url: localhost
  db: foo
  tables:
  -
     table: bar
     collection: test.simple
     columns:
        c: int
        d: string
  -
     table: silly
     collection: test.simple
     columns:
        e: int
        f: string



`)
