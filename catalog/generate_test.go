package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateCreateTable(t *testing.T) {
	Convey("Subject: GenerateCreateTable", t, func() {
		config := schema.Must(schema.New(testSchema, &lgr))

		tblConfig := config.Databases[0].Tables[0]
		t := catalog.NewMongoTable(tblConfig, catalog.BaseTable, collation.Default)
		createTable := catalog.GenerateCreateTable(t, 0)
		So(createTable, ShouldEqual, testSchemaCreateTableFoo)

		tblConfig = config.Databases[0].Tables[1]
		t = catalog.NewMongoTable(tblConfig, catalog.BaseTable, collation.Default)
		createTable = catalog.GenerateCreateTable(t, 10)
		So(createTable, ShouldEqual, testSchemaCreateTableBar)
	})
}
