package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	_ fmt.Stringer = nil
)

func TestCachePlanStage(t *testing.T) {
	Convey("A cache operator...", t, func() {
		ctx := &evaluator.ExecutionCtx{ConnectionCtx: createTestConnectionCtx(nil)}

		Convey("should not open without rows", func() {
			cs := &evaluator.CacheStage{}
			_, err := cs.Open(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("should iterate through all rows contained successfully", func() {
			testCache := func(cs *evaluator.CacheStage, ctx *evaluator.ExecutionCtx,
				expected []evaluator.Values) {
				iter, err := cs.Open(ctx)
				So(err, ShouldBeNil)

				row := &evaluator.Row{}
				i := 0
				for iter.Next(row) {
					So(len(row.Data), ShouldEqual, len(expected[i]))
					So(row.Data, ShouldResemble, expected[i])
					row = &evaluator.Row{}
					i++
				}
				So(i, ShouldEqual, len(expected))

				So(iter.Close(), ShouldBeNil)
				So(iter.Err(), ShouldBeNil)
			}

			expected := []evaluator.Values{
				{{SelectID: 1,
					Database: dbOne, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(1)}},
				{{SelectID: 1,
					Database: dbOne, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(2)}},
				{{SelectID: 1,
					Database: dbOne, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(3)}},
				{{SelectID: 1,
					Database: dbOne, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(4)}},
				{{SelectID: 1,
					Database: dbOne, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(5)}},
				{{SelectID: 1,
					Database: dbOne, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(6)}},
				{{SelectID: 1,
					Database: dbOne, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(7)}},
			}

			var rows []evaluator.Row
			for _, values := range expected {
				rows = append(rows, evaluator.Row{Data: values})
			}

			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			// Iterate through the cache twice to ensure the same values are obtained both times
			testCache(cs, ctx, expected)
			testCache(cs, ctx, expected)

		})

	})
}
