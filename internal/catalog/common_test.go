package catalog_test

var testSchema = []byte(`
schema:
- db: test
  tables:
  - table: foo
    collection: fooCollection
    pipeline:
    - $unwind:
        includeArrayIndex: a_idx
        path: "$a"
    - $unwind:
        includeArrayIndex: a_idx_1
        path: "$a"
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlType: varchar
      SqlName: id
    - Name: a
      MongoType: int
      SqlType: int
      SqlName: value
    - Name: a_idx
      MongoType: int
      SqlType: int
      SqlName: idx1
    - Name: a_idx_1
      MongoType: int
      SqlType: int
      SqlName: idx2
  - table: bar
    collection: barCollection
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlType: varchar
      SqlName: id
    - Name: a
      MongoType: int
      SqlType: int
      SqlName: a
    - Name: b
      MongoType: string
      SqlType: varchar
      SqlName: b
    - Name: c
      MongoType: geo.2darray
      SqlType: numeric[]
      SqlName: c
    - Name: d
      MongoType: bson.Decimal128
      SqlType: decimal128
      SqlName: d
    - Name: e
      MongoType: bson.UUID
      SqlType: varchar
      SqlName: e
    - Name: f
      MongoType: number
      SqlType: numeric
      SqlName: f
`)

var testSchemaCreateTableFoo = "CREATE TABLE `foo` (\n" +
	"  `id` varchar(65535) COLLATE utf8_bin DEFAULT NULL COMMENT '{ \"name\": \"_id\" }',\n" +
	"  `idx1` bigint(20) DEFAULT NULL COMMENT '{ \"name\": \"a_idx\" }',\n" +
	"  `idx2` bigint(20) DEFAULT NULL COMMENT '{ \"name\": \"a_idx_1\" }',\n" +
	"  `value` bigint(20) DEFAULT NULL COMMENT '{ \"name\": \"a\" }',\n" +
	"  PRIMARY KEY (`id`,`idx1`,`idx2`)\n" +
	") ENGINE=MongoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin" +
	" COMMENT='{ \"collectionName\": \"fooCollection\" }'"

var testSchemaCreateTableBar = "CREATE TABLE `bar` (\n" +
	"  `id` varchar(10) COLLATE utf8_bin DEFAULT NULL COMMENT '{ \"name\": \"_id\" }',\n" +
	"  `a` bigint(20) DEFAULT NULL COMMENT '{ \"name\": \"a\" }',\n" +
	"  `b` varchar(10) COLLATE utf8_bin DEFAULT NULL COMMENT '{ \"name\": \"b\" }',\n" +
	"  `c_latitude` double DEFAULT NULL COMMENT '{ \"name\": \"c.1\" }',\n" +
	"  `c_longitude` double DEFAULT NULL COMMENT '{ \"name\": \"c.0\" }',\n" +
	"  `d` decimal(65,20) DEFAULT NULL COMMENT '{ \"name\": \"d\" }',\n" +
	"  `e` varchar(10) COLLATE utf8_bin DEFAULT NULL COMMENT '{ \"name\": \"e\" }',\n" +
	"  `f` double DEFAULT NULL COMMENT '{ \"name\": \"f\" }',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=MongoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin" +
	" COMMENT='{ \"collectionName\": \"barCollection\" }'"
