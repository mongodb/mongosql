package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	lgr = log.GlobalLogger()
)

func TestMongoTable(t *testing.T) {
	config := schema.Must(schema.New(testSchema, &lgr))
	tblConfig := config.Databases[0].Tables[0]

	Convey("Subject: MongoTable", t, func() {
		t := catalog.NewMongoTable(tblConfig, catalog.BaseTable, collation.Default)

		So(string(t.Name()), ShouldEqual, "foo")
		So(t.CollectionName, ShouldEqual, "fooCollection")
		columns := t.Columns()
		So(len(columns), ShouldEqual, 4)

		column, err := t.Column("id")
		So(err, ShouldBeNil)
		So(string(column.Name()), ShouldEqual, "id")
		So(column.(*catalog.MongoColumn).MongoName, ShouldEqual, "_id")

		column, err = t.Column("value")
		So(err, ShouldBeNil)
		So(string(column.Name()), ShouldEqual, "value")
		So(column.(*catalog.MongoColumn).MongoName, ShouldEqual, "a")

		column, err = t.Column("idx1")
		So(err, ShouldBeNil)
		So(string(column.Name()), ShouldEqual, "idx1")
		So(column.(*catalog.MongoColumn).MongoName, ShouldEqual, "a_idx")

		column, err = t.Column("idx2")
		So(err, ShouldBeNil)
		So(string(column.Name()), ShouldEqual, "idx2")
		So(column.(*catalog.MongoColumn).MongoName, ShouldEqual, "a_idx_1")
	})
}
