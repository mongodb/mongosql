package sqlproxy

var (
	dbOne        = "test"
	tableOneName = "simple"
	tableTwoName = "simple2"

	testSchemaSimple = []byte(
		`
schema:
-
  db: test
  tables:
  -
     table: bar
     collection: simple
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
     -
        Name: c
        MongoType: int
        SqlType: int
-
  db: foo
  tables:
  -
     table: bar
     collection: simple
     columns:
     -
        Name: c
        MongoType: int
        SqlType: int
     -
        Name: d
        MongoType: int
        SqlType: int
  -
     table: silly
     collection: simple2
     columns:
     -
        Name: e
        MongoType: int
        SqlType: int
     -
        Name: f
        MongoType: int
        SqlType: int
`)
)
