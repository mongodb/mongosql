package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSourceAppendOperator(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne

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
			So(len(ctx.SrcRows), ShouldEqual, 1)
		}

		So(operator.Close(), ShouldBeNil)
		So(operator.Err(), ShouldBeNil)
	}

	Convey("A source append operator...", t, func() {

		sourceAppend := &SourceAppend{}

		Convey("should always append the source row from the source operator", func() {

			runTest(sourceAppend)

		})

	})
}
