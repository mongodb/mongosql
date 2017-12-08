package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRowGeneratorStage(t *testing.T) {
	Convey("A row generator operator...", t, func() {
		selectIDs := []int{1}
		newColumn := evaluator.NewColumn(selectIDs[0], "", "", "", "rowCount", "", "rowCount",
			schema.SQLUint64, schema.SQLUint64, false)
		ctx := &evaluator.ExecutionCtx{}

		Convey("should iterate through all rows contained successfully with only empty rows", func() {
			rows := []evaluator.Row{}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)

			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &evaluator.Row{}
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
			rows := []evaluator.Row{
				{evaluator.Values{{1, "test1", "test", "rowCount1", evaluator.SQLInt(2)}}},
			}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &evaluator.Row{}
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
			rows := []evaluator.Row{
				{evaluator.Values{{1, "test1", "test", "rowCount", evaluator.SQLInt(0)}}},
			}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &evaluator.Row{}
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
			rows := []evaluator.Row{
				{evaluator.Values{{1, "test1", "test", "rowCount", evaluator.SQLInt(5)}}},
			}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &evaluator.Row{}
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
			rows := []evaluator.Row{
				{evaluator.Values{{1, "test1", "test", "rowCount", evaluator.SQLInt(5)}}},
				{evaluator.Values{{1, "test1", "test", "rowCount", evaluator.SQLInt(2)}}},
			}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &evaluator.Row{}
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
			rows := []evaluator.Row{
				{evaluator.Values{{1, "test1", "test", "rowCount", evaluator.SQLInt(0)}}},
				{evaluator.Values{{1, "test1", "test", "rowCount", evaluator.SQLInt(2)}}},
			}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)
			iter, err := rg.Open(ctx)
			So(err, ShouldBeNil)

			row := &evaluator.Row{}
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
