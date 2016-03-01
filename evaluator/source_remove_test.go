package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSourceRemoveOperator(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne

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
			So(len(ctx.SrcRows), ShouldEqual, 0)
		}

		So(operator.Close(), ShouldBeNil)
		So(operator.Err(), ShouldBeNil)
	}

	Convey("A source remove operator...", t, func() {

		sourceRemove := &SourceRemove{}

		Convey("should always remove the source row from the source operator", func() {
			runTest(sourceRemove)
		})
	})
}
