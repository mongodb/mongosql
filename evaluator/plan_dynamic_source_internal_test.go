package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDynamicSourceStage(t *testing.T) {

	tableName := "foo"
	table := catalog.NewDynamicTable(tableName, catalog.BaseTable, func() []*catalog.DataRow {
		return []*catalog.DataRow{
			catalog.NewDataRow(1, 2),
			catalog.NewDataRow(2, 3),
			catalog.NewDataRow(3, 4),
		}
	})

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	table.AddColumn("one", schema.SQLInt)
	table.AddColumn("two", schema.SQLInt)

	expected := []Values{
		Values{
			{SelectID: 1, Table: tableName, Name: "one", Data: SQLInt(1)},
			{SelectID: 1, Table: tableName, Name: "two", Data: SQLInt(2)},
		},
		Values{
			{SelectID: 1, Table: tableName, Name: "one", Data: SQLInt(2)},
			{SelectID: 1, Table: tableName, Name: "two", Data: SQLInt(3)},
		},
		Values{
			{SelectID: 1, Table: tableName, Name: "one", Data: SQLInt(3)},
			{SelectID: 1, Table: tableName, Name: "two", Data: SQLInt(4)},
		},
	}

	Convey("Subject: DynamicSourceStage", t, func() {
		db := &catalog.Database{}

		source := NewDynamicSourceStage(db, table, 1, tableName)

		connectionCtx := createTestConnectionCtx(testInfo)
		execCtx := &ExecutionCtx{
			ConnectionCtx: connectionCtx,
		}

		iter, err := source.Open(execCtx)
		So(err, ShouldBeNil)

		i := 0

		row := &Row{}
		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expected[i]))
			So(row.Data, ShouldResemble, expected[i])
			row = &Row{}
			i++
		}

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	})
}
