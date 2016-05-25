package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSourceRemoveOperator(t *testing.T) {
	runTest := func(operator *SourceRemoveStage) {

		ctx := &ExecutionCtx{
			SrcRows: []*Row{&Row{}},
		}

		ts := &BSONSourceStage{tableOneName, nil}

		operator.source = ts
		iter, err := operator.Open(ctx)
		So(err, ShouldBeNil)
		row := &Row{}

		So(len(ctx.SrcRows), ShouldEqual, 1)

		for iter.Next(row) {
			So(len(ctx.SrcRows), ShouldEqual, 0)
		}

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("A source remove operator...", t, func() {

		sourceRemove := &SourceRemoveStage{}

		Convey("should always remove the source row from the source operator", func() {
			runTest(sourceRemove)
		})
	})
}
