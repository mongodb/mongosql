package translator

import (
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewParseCtx(t *testing.T) {

	Convey("With a simple SQL statement...", t, func() {

		Convey("a new parse context should contain the right table and column aliases", func() {

			sql := "select age as a, height as h from foo f"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt := raw.(*sqlparser.Select)
			ctx, err := NewParseCtx(stmt)
			So(err, ShouldBeNil)

			So(len(ctx.Table), ShouldEqual, 1)
			tableAs := map[string]string{"f": "foo"}
			So(ctx.Table[0].Name, ShouldResemble, tableAs)

			So(len(ctx.Column), ShouldEqual, 2)
			columnInfo := []ColumnInfo{
				{
					Name:  map[string]string{"a": "age"},
					Table: "foo",
				},
				{
					Name:  map[string]string{"h": "height"},
					Table: "foo",
				},
			}

			for i, column := range ctx.Column {
				So(column, ShouldResemble, columnInfo[i])
			}
		})

	})
}
