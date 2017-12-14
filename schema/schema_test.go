package schema_test

import (
	"testing"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"

	"github.com/10gen/mongo-go-driver/bson"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

var (
	lgr = log.GlobalLogger()
)

func TestSchema(t *testing.T) {
	var testSchemaData = []byte(
		`
schema:
-
  db: test1
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar
-
  db: test2
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        SQLName: a
        MongoType: string
        SqlType: varchar
     -
        Name: b
        SQLName: b
        MongoType: int
        SqlType: int
     -
        Name: c
        MongoType: number
        SqlType: numeric
     pipeline:
     - $unwind : "$x"
     - $sort: { a: 1, b: 1, c: -1 }
     - $project: { a: 1, b: 1, c: { $add: ["$a", { $numberLong: "10" }] } }

  -
     table: bar2
     collection: bar2
`)

	Convey("Schema should parse correctly", t, func() {

		cfg, err := schema.New(testSchemaData, &lgr)
		So(err, ShouldBeNil)

		So(len(cfg.Databases), ShouldEqual, 2)
		So(cfg.Databases[0].Name, ShouldEqual, "test1")
		So(len(cfg.Databases[0].Tables), ShouldEqual, 1)
		So(len(cfg.Databases[1].Tables), ShouldEqual, 2)

		So(cfg.Databases[0].Tables[0].Name, ShouldEqual, "foo")
		So(cfg.Databases[0].Tables[0].CollectionName, ShouldEqual, "foo")
		So(cfg.Databases[0].Name, ShouldEqual, "test1")
		So(cfg.Databases[0].Tables[0].CollectionName, ShouldEqual, "foo")

		So(len(cfg.Databases[0].Tables[0].Columns), ShouldEqual, 2)

		So(cfg.Databases[0].Tables[0].Columns[0].SQLName, ShouldEqual, "a")

		So(cfg.Databases[1].Tables[0].Pipeline, ShouldResemble, []bson.D{
			{{"$unwind", "$x"}},
			{{"$sort", bson.D{{"a", int64(1)}, {"b", int64(1)}, {"c", int64(-1)}}}},
			{{"$project", bson.D{{"a", int64(1)}, {"b", int64(1)}, {"c", bson.D{{"$add", []interface{}{"$a", int64(10)}}}}}}}})
	})
}

func TestSchemaSubdir(t *testing.T) {
	var testSchemaDataSub = []byte(
		`
schema:
-
  db: test1
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar
     -
        Name: c
        MongoType: bson.Decimal128
        SqlType: numeric
-
  db: test2
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
     -
        Name: b
        MongoType: int
        SqlType: int
     pipeline:
     - $unwind : "$x"
     - $sort: { a: 1, b: 1, c: -1 }
-
  db: test3
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: c
        MongoType: string
        SqlType: varchar
     -
        Name: d
        MongoType: geo.2darray
        SQLName: d
        SqlType: numeric[]
  -
     table: bar2
     collection: bar2
`)

	Convey("Schema should parse correctly", t, func() {
		cfg := &schema.Schema{}

		err := cfg.Load(testSchemaDataSub, &lgr)
		So(err, ShouldBeNil)

		So(len(cfg.Databases), ShouldEqual, 3)
		So(len(cfg.Databases[0].Tables), ShouldEqual, 1)
		So(len(cfg.Databases[1].Tables), ShouldEqual, 1)
		So(len(cfg.Databases[2].Tables), ShouldEqual, 2)
		So(len(cfg.Databases[2].Tables[0].Columns), ShouldEqual, 3)
		So(len(cfg.Databases[2].Tables[1].Columns), ShouldEqual, 0)

		So(cfg.Databases[2].Tables[0].Columns[0].SQLName, ShouldEqual, "c")
		So(cfg.Databases[2].Tables[0].Columns[1].SQLName, ShouldEqual, "d_longitude")
		So(cfg.Databases[2].Tables[0].Columns[2].SQLName, ShouldEqual, "d_latitude")

		So(cfg.Databases[0].Name, ShouldEqual, "test1")

		So(cfg.Databases[0].Tables[0].Name, ShouldEqual, "foo")
		So(cfg.Databases[0].Tables[0].CollectionName, ShouldEqual, "foo")
		So(cfg.Databases[0].Name, ShouldEqual, "test1")
		So(cfg.Databases[0].Tables[0].CollectionName, ShouldEqual, "foo")

		So(len(cfg.Databases[0].Tables[0].Columns), ShouldEqual, 3)

		So(cfg.Databases[0].Tables[0].Columns[0].SQLName, ShouldEqual, "a")

		So(cfg.Databases[1].Tables[0].Pipeline, ShouldResemble, []bson.D{
			{{"$unwind", "$x"}},
			{{"$sort", bson.D{{"a", int64(1)}, {"b", int64(1)}, {"c", int64(-1)}}}}})
	})
}

func TestSchemaSubdir2(t *testing.T) {
	var testSchemaDataRoot = []byte(
		`
schema:
-
  db: test1
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar
-
  db: test3
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	var testSchemaDataSub = []byte(
		`
schema:
-
  db: test2
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: string
        SqlType: varchar
     -
        Name: b
        MongoType: int
        SqlType: int
-
  db: test3
  tables:
  -
     table: bar
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	cfg, err := schema.New(testSchemaDataRoot, &lgr)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.Load(testSchemaDataSub, &lgr)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Databases) != 3 {
		t.Fatal(cfg)
	}

	if len(cfg.Databases[0].Tables) != 1 {
		t.Fatal("test1 wrong")
	}

	if len(cfg.Databases[1].Tables) != 2 {
		t.Fatal("test3 wrong")
	}

	if len(cfg.Databases[2].Tables) != 1 {
		t.Fatal("test2 wrong")
	}

}

func TestSchemaSubdirConflict(t *testing.T) {
	var testSchemaDataRoot = []byte(
		`
schema:
-
  db: test3
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	var testSchemaDataSub = []byte(
		`
schema:
-
  db: test3
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: string
        SqlType: varchar

`)

	cfg, err := schema.New(testSchemaDataRoot, &lgr)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.Load(testSchemaDataSub, &lgr)
	if err == nil {
		t.Fatal("should have conflicted")
	}
}

func TestDuplicateColumnName(t *testing.T) {
	var testSchemaDataRoot = []byte(
		`
schema:
-
  db: test3
  tables:
  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SQLName: B
        SqlType: int
     -
        Name: b
        SQLName: B
        MongoType: string
        SqlType: varchar
`)

	_, err := schema.New(testSchemaDataRoot, &lgr)
	if err == nil {
		t.Fatal("should have conflicted")
	}
}

func TestReadFile(t *testing.T) {
	cfg := &schema.Schema{}
	err := cfg.LoadFile("testdata/foo.conf", &lgr)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.LoadDir("testdata/sub", &lgr)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Databases) != 3 {
		t.Fatalf("num RawDatabases wrong: %d", len(cfg.Databases))
	}
}

func TestAddTable(t *testing.T) {
	db := &schema.Database{Name: "test"}
	t1 := &schema.Table{Name: "t", CollectionName: "mongo_t1"}
	t2 := &schema.Table{Name: "T", CollectionName: "mongo_t1"}

	err := db.AddTable(t1, &lgr)
	if err != nil {
		t.Fatalf("AddTable call 1 failed: %v", err)
	}

	err = db.AddTable(t2, &lgr)
	if err != nil {
		t.Fatalf("AddTable call 2 failed: %v", err)
	}

	if len(db.Tables) != 2 {
		t.Fatalf("num tables wrong: %d", len(db.Tables))
	}

	if db.Tables[0].Name != "t" {
		t.Fatalf("first table name is: %v", db.Tables[0].Name)
	}

	if db.Tables[1].Name != "T_0" {
		t.Fatalf("second table name is: %v", db.Tables[1].Name)
	}
}

func TestAddColumn(t *testing.T) {
	table := &schema.Table{Name: "t", CollectionName: "mongo_t1"}
	columns := []*schema.Column{
		{SQLName: "c", Name: "c"},
		{SQLName: "C", Name: "C"},
		{SQLName: "ca", Name: "ca"},
		{SQLName: "cA", Name: "cA"},
		{SQLName: "Ca", Name: "Ca"},
	}

	for i, column := range columns {
		err := table.AddColumn(column, &lgr)
		if err != nil {
			t.Fatalf("AddColumn for column %v failed: %v", i+1, err)
		}
	}

	if len(table.Columns) != len(columns) {
		t.Fatalf("unexpected column count: %d", len(table.Columns))
	}

	expectedNames := []string{"c", "C_0", "ca", "cA_0", "Ca_1"}

	for i, column := range table.Columns {
		if table.Columns[i].SQLName != expectedNames[i] {
			t.Fatalf("column %v name is: %v expected %v", i+1,
				column.SQLName, expectedNames[i])
		}
	}
}

func TestTablePostProcess(t *testing.T) {
	columns := []*schema.Column{
		{SQLName: "abc", Name: "abc"},
		{SQLName: "def", Name: "def"},
		{SQLName: " ", Name: " "},
	}

	tableName := "mongo_t1"

	req := require.New(t)

	table := schema.NewTable(tableName, tableName, nil, columns, nil, nil, false)
	err := table.PostProcess(false, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(2, len(table.Columns), "whitespace column was not removed")
	req.Equal(table.Columns[0].Name, "abc")
	req.Equal(table.Columns[0].SQLName, "abc")
	req.Equal(table.Columns[1].Name, "def")
	req.Equal(table.Columns[1].SQLName, "def")

	table = schema.NewTable(tableName, tableName, nil, columns, nil, nil, true)
	err = table.PostProcess(false, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(3, len(table.Columns), "post processed tables should not get processed again")
	req.Equal(table.Columns[0].Name, "abc")
	req.Equal(table.Columns[0].SQLName, "abc")
	req.Equal(table.Columns[1].Name, "def")
	req.Equal(table.Columns[1].SQLName, "def")
	req.Equal(table.Columns[2].Name, " ")
	req.Equal(table.Columns[2].SQLName, " ")

	geoColumns := []*schema.Column{
		columns[0],
		columns[1],
		{
			SQLName:   "ghi",
			Name:      "ghi",
			MongoType: schema.MongoGeo2D,
		},
	}

	table = schema.NewTable(tableName, tableName, nil, geoColumns, nil, nil, false)
	err = table.PostProcess(false, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(4, len(table.Columns), "geo2d column was not remapped")
	req.Equal(table.Columns[0].Name, "abc")
	req.Equal(table.Columns[0].SQLName, "abc")
	req.Equal(table.Columns[1].Name, "def")
	req.Equal(table.Columns[1].SQLName, "def")
	req.Equal(table.Columns[2].Name, "ghi.0")
	req.Equal(table.Columns[2].SQLName, "ghi_longitude")
	req.Equal(table.Columns[3].Name, "ghi.1")
	req.Equal(table.Columns[3].SQLName, "ghi_latitude")
	geoColumns = []*schema.Column{
		columns[0],
		columns[1],
		{
			SQLName:   "ghi",
			Name:      "ghi",
			MongoType: schema.MongoGeo2D,
		},
		{
			SQLName: "ghi_longitude",
			Name:    "ghi_longitude",
		},
	}

	table = schema.NewTable(tableName, tableName, nil, geoColumns, nil, nil, false)
	err = table.PostProcess(false, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(5, len(table.Columns), "existing geo2d column was not remapped")
	req.Equal(table.Columns[0].Name, "abc")
	req.Equal(table.Columns[0].SQLName, "abc")
	req.Equal(table.Columns[1].Name, "def")
	req.Equal(table.Columns[1].SQLName, "def")
	req.Equal(table.Columns[2].Name, "ghi.0")
	req.Equal(table.Columns[2].SQLName, "ghi_longitude_0")
	req.Equal(table.Columns[3].Name, "ghi.1")
	req.Equal(table.Columns[3].SQLName, "ghi_latitude")
	req.Equal(table.Columns[4].Name, "ghi_longitude")
	req.Equal(table.Columns[4].SQLName, "ghi_longitude")

	parentPKs := []*schema.Column{
		{SQLName: "_id", Name: "_id"},
	}

	parentCols := []*schema.Column{
		{SQLName: "_id", Name: "_id"},
		{SQLName: "xyz", Name: "xyz"},
	}

	parent := schema.NewTable("parent", "parent", nil, parentCols, nil, parentPKs, false)
	table = schema.NewTable(tableName, tableName, nil, columns, parent, nil, false)
	err = table.PostProcess(true, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(4, len(table.Columns), "incorrect pre-join table column count")
	req.Equal(table.Columns[0].Name, "_id")
	req.Equal(table.Columns[0].SQLName, "_id")
	req.Equal(table.Columns[1].Name, "abc")
	req.Equal(table.Columns[1].SQLName, "abc")
	req.Equal(table.Columns[2].Name, "def")
	req.Equal(table.Columns[2].SQLName, "def")
	req.Equal(table.Columns[3].Name, "xyz")
	req.Equal(table.Columns[3].SQLName, "xyz")

	table = schema.NewTable(tableName, tableName, nil, columns, parent, nil, false)
	err = table.PostProcess(false, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(3, len(table.Columns), "incorrect non-pre-join table column count")
	req.Equal(table.Columns[0].Name, "_id")
	req.Equal(table.Columns[0].SQLName, "_id")
	req.Equal(table.Columns[1].Name, "abc")
	req.Equal(table.Columns[1].SQLName, "abc")
	req.Equal(table.Columns[2].Name, "def")
	req.Equal(table.Columns[2].SQLName, "def")

	parentPKs = []*schema.Column{
		{SQLName: "_id", Name: "_id"},
	}

	parentCols = []*schema.Column{
		{SQLName: "_id", Name: "_id"},
		{SQLName: "def", Name: "def_parent"},
	}

	parent = schema.NewTable("parent", "parent", nil, parentCols, nil, parentPKs, false)
	table = schema.NewTable(tableName, tableName, nil, columns, parent, nil, false)
	err = table.PostProcess(true, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(4, len(table.Columns), "incorrect pre-join table column (with conflict) count")
	req.Equal(table.Columns[0].Name, "_id")
	req.Equal(table.Columns[0].SQLName, "_id")
	req.Equal(table.Columns[1].Name, "abc")
	req.Equal(table.Columns[1].SQLName, "abc")
	req.Equal(table.Columns[2].Name, "def")
	req.Equal(table.Columns[2].SQLName, "def")
	req.Equal(table.Columns[3].Name, "def_parent")
	req.Equal(table.Columns[3].SQLName, "def_0")

	table = schema.NewTable(tableName, tableName, nil, columns, parent, nil, false)
	err = table.PostProcess(false, &lgr)
	req.Nilf(err, "error in post processing table: %v", err)
	req.Equal(3, len(table.Columns), "incorrect non-pre-join table column (with conflict) count")
	req.Equal(table.Columns[0].Name, "_id")
	req.Equal(table.Columns[0].SQLName, "_id")
	req.Equal(table.Columns[1].Name, "abc")
	req.Equal(table.Columns[1].SQLName, "abc")
	req.Equal(table.Columns[2].Name, "def")
	req.Equal(table.Columns[2].SQLName, "def")
}
