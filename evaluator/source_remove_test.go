package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSourceRemoveOperator(t *testing.T) {

	runTest := func(operator *SourceRemove) {

		ctx := &ExecutionCtx{
			Schema:  cfgOne,
			Db:      dbOne,
			SrcRows: []*Row{&Row{}},
		}

		ts, err := NewBSONSource(ctx, tableOneName, nil)
		So(err, ShouldBeNil)

		operator.source = ts
		So(operator.Open(ctx), ShouldBeNil)

		row := &Row{}

		So(len(ctx.SrcRows), ShouldEqual, 1)

		for operator.Next(row) {
			if operator.hasSubquery {
				So(len(ctx.SrcRows), ShouldEqual, 0)
			} else {
				So(len(ctx.SrcRows), ShouldEqual, 1)
			}
		}

		So(operator.Close(), ShouldBeNil)
		So(operator.Err(), ShouldBeNil)
	}

	Convey("A source remove operator...", t, func() {

		sourceRemove := &SourceRemove{hasSubquery: true}

		Convey("should remove the source row if the source operator contains a subquery", func() {

			runTest(sourceRemove)

		})

		Convey("should not remove the source row if the source operator does not contains a subquery", func() {

			sourceRemove.hasSubquery = false
			runTest(sourceRemove)

		})

	})
}
