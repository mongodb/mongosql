package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSourceAppendOperator(t *testing.T) {
	runTest := func(operator *SourceAppendStage) {

		ctx := &ExecutionCtx{}

		ts := &BSONSourceStage{tableOneName, nil}

		operator.source = ts
		iter, err := operator.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}

		So(len(ctx.SrcRows), ShouldEqual, 0)

		for iter.Next(row) {
			So(len(ctx.SrcRows), ShouldEqual, 1)
		}

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("A source append operator...", t, func() {

		sourceAppend := &SourceAppendStage{}

		Convey("should always append the source row from the source operator", func() {

			runTest(sourceAppend)

		})

	})
}
