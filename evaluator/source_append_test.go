package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSourceAppendOperator(t *testing.T) {

	runTest := func(operator *SourceAppend) {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		ts, err := NewBSONSource(ctx, tableOneName, nil)
		So(err, ShouldBeNil)

		operator.source = ts
		So(operator.Open(ctx), ShouldBeNil)

		row := &Row{}

		So(len(ctx.SrcRows), ShouldEqual, 0)

		for operator.Next(row) {
			if operator.hasSubquery {
				So(len(ctx.SrcRows), ShouldEqual, 1)
			} else {
				So(len(ctx.SrcRows), ShouldEqual, 0)
			}
		}

		So(operator.Close(), ShouldBeNil)
		So(operator.Err(), ShouldBeNil)
	}

	Convey("A source append operator...", t, func() {

		sourceAppend := &SourceAppend{hasSubquery: true}

		Convey("should append the source row if the source operator contains a subquery", func() {

			runTest(sourceAppend)

		})

		Convey("should not append the source row if the source operator does not contains a subquery", func() {

			sourceAppend.hasSubquery = false
			runTest(sourceAppend)

		})

	})
}
