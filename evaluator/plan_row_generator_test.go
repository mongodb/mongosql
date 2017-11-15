package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRowGeneratorStage(t *testing.T) {
	Convey("A row generator operator...", t, func() {
		selectIDs := []int{1}
		newColumn := NewColumn(selectIDs[0], "", "", "", "rowCount", "", "rowCount",
			schema.SQLUint64, schema.SQLUint64, false)
		ctx := &ExecutionCtx{}
		cs := &CacheStage{}
		rg := NewRowGeneratorStage(cs, newColumn)

		Convey("should iterate through all rows contained successfully with only empty rows", func() {
			rows := []Row{}
			cs.rows = rows
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}
			i := 0
			for iter.Next(row) {
				So(row.Data, ShouldBeNil)
				i++
			}

			So(i, ShouldEqual, 0)
			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)
		})

		Convey("should iterate through all rows contained successfully with only one row having content of different field name", func() {
			rows := []Row{
				{Values{{1, "test1", "test", "rowCount1", SQLInt(2)}}},
			}
			cs.rows = rows
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}
			i := 0
			for iter.Next(row) {
				So(row.Data, ShouldBeNil)
				i++
			}

			So(i, ShouldEqual, 0)
			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldNotBeNil)
		})

		Convey("should iterate through all rows contained successfully with only one row having 0 value", func() {
			rows := []Row{
				{Values{{1, "test1", "test", "rowCount", SQLInt(0)}}},
			}
			cs.rows = rows
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}
			i := 0
			for iter.Next(row) {
				So(row.Data, ShouldBeNil)
				i++
			}

			So(i, ShouldEqual, 0)
			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)
		})

		Convey("should iterate through all rows contained successfully with only one row", func() {
			rows := []Row{
				{Values{{1, "test1", "test", "rowCount", SQLInt(5)}}},
			}
			cs.rows = rows
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}
			i := 0
			for iter.Next(row) {
				So(row.Data, ShouldBeNil)
				i++
			}

			So(i, ShouldEqual, 5)
			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)
		})

		Convey("should iterate through all rows contained successfully with two rows", func() {
			rows := []Row{
				{Values{{1, "test1", "test", "rowCount", SQLInt(5)}}},
				{Values{{1, "test1", "test", "rowCount", SQLInt(2)}}},
			}
			cs.rows = rows
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}
			i := 0
			for iter.Next(row) {
				So(row.Data, ShouldBeNil)
				i++
			}

			So(i, ShouldEqual, 7)
			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)
		})

		Convey("should iterate through all rows contained successfully with only two rows with at least one row with 0 value", func() {
			rows := []Row{
				{Values{{1, "test1", "test", "rowCount", SQLInt(0)}}},
				{Values{{1, "test1", "test", "rowCount", SQLInt(2)}}},
			}
			cs.rows = rows
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}
			i := 0
			for iter.Next(row) {
				So(row.Data, ShouldBeNil)
				i++
			}

			So(i, ShouldEqual, 2)
			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)
		})

	})
}
