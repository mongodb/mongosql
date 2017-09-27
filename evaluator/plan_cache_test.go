package evaluator

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	_ fmt.Stringer = nil
)

func TestCachePlanStage(t *testing.T) {
	Convey("A cache operator...", t, func() {
		cs := &CacheStage{}
		ctx := &ExecutionCtx{}

		Convey("should not open without rows", func() {
			_, err := cs.Open(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("should iterate through all rows contained sucessfully", func() {
			testCache := func(cs *CacheStage, ctx *ExecutionCtx, expected []Values) {
				iter, err := cs.Open(ctx)
				So(err, ShouldBeNil)

				row := &Row{}
				i := 0
				for iter.Next(row) {
					So(len(row.Data), ShouldEqual, len(expected[i]))
					So(row.Data, ShouldResemble, expected[i])
					row = &Row{}
					i++
				}
				So(i, ShouldEqual, len(expected))

				So(iter.Close(), ShouldBeNil)
				So(iter.Err(), ShouldBeNil)
			}

			expected := []Values{
				{{1, tableOneName, "a", SQLInt(1)}},
				{{1, tableOneName, "a", SQLInt(2)}},
				{{1, tableOneName, "a", SQLInt(3)}},
				{{1, tableOneName, "a", SQLInt(4)}},
				{{1, tableOneName, "a", SQLInt(5)}},
				{{1, tableOneName, "a", SQLInt(6)}},
				{{1, tableOneName, "a", SQLInt(7)}},
			}

			var rows []Row
			for _, values := range expected {
				rows = append(rows, Row{Data: values})
			}

			cs.rows = rows
			// Iterate through the cache twice to ensure the same values are obtained both times
			testCache(cs, ctx, expected)
			testCache(cs, ctx, expected)

		})

	})
}
