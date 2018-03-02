package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDynamicTable(t *testing.T) {
	Convey("Subject: DynamicTable", t, func() {
		t := catalog.NewDynamicTable("foo", catalog.BaseTable, func() []*catalog.DataRow {
			var rows []*catalog.DataRow
			for i := 0; i < 3; i++ {
				rows = append(rows, catalog.NewDataRow(i, i+1))
			}
			return rows
		})

		Convey("NewDynamicTable", func() {
			So(string(t.Name()), ShouldEqual, "foo")
			So(len(t.Columns()), ShouldEqual, 0)
		})

		Convey("AddColumn", func() {
			Convey("Should add the column if it doesn't already exist", func() {
				_, err := t.AddColumn("id", schema.SQLVarchar)
				So(err, ShouldBeNil)
				So(len(t.Columns()), ShouldEqual, 1)
			})

			Convey("Should return an error when an existing column has the same name", func() {
				_, err := t.AddColumn("id", schema.SQLVarchar)
				So(err, ShouldBeNil)

				_, err = t.AddColumn("id", schema.SQLVarchar)
				So(err, ShouldNotBeNil)

				_, err = t.AddColumn("ID", schema.SQLVarchar)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Column", func() {
			t.AddColumn("id", schema.SQLVarchar)

			Convey("Should return an error if the column doesn't exist", func() {
				_, err := t.Column("blah")
				So(err, ShouldNotBeNil)
			})

			Convey("Should return the column when it exists", func() {
				c, err := t.Column("id")
				So(err, ShouldBeNil)
				So(c, ShouldNotBeNil)
				So(string(c.Name()), ShouldEqual, "id")

				c, err = t.Column("ID")
				So(err, ShouldBeNil)
				So(c, ShouldNotBeNil)
				So(string(c.Name()), ShouldEqual, "id")
			})
		})

		Convey("Columns", func() {
			t.AddColumn("one", schema.SQLVarchar)
			t.AddColumn("two", schema.SQLVarchar)
			t.AddColumn("three", schema.SQLVarchar)

			So(len(t.Columns()), ShouldEqual, 3)
		})

		Convey("OpenReader", func() {
			t.AddColumn("one", schema.SQLInt)
			t.AddColumn("two", schema.SQLInt)

			reader, err := t.OpenReader()
			So(reader, ShouldNotBeNil)
			So(err, ShouldBeNil)

			row := &catalog.DataRow{}
			f, err := reader.Next(row)
			So(f, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(row.Values, ShouldResemble, []interface{}{0, 1})

			f, err = reader.Next(row)
			So(f, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(row.Values, ShouldResemble, []interface{}{1, 2})

			f, err = reader.Next(row)
			So(f, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(row.Values, ShouldResemble, []interface{}{2, 3})

			f, err = reader.Next(row)
			So(f, ShouldBeFalse)
			So(err, ShouldBeNil)

			err = reader.Close()
			So(err, ShouldBeNil)
		})

	})
}
