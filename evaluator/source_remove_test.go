package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func sourceRemoveTest(operator *SourceRemove) {

	ctx := &ExecutionCtx{
		Schema:  cfgOne,
		Db:      dbOne,
		SrcRows: []*Row{&Row{}},
		Session: session,
	}

	operator.source = &TableScan{
		ctx:       ctx,
		tableName: tableOneName,
	}

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

func TestSourceRemoveOperator(t *testing.T) {

	Convey("A source remove operator...", t, func() {

		sourceRemove := &SourceRemove{hasSubquery: true}

		Convey("should remove the source row if the source operator contains a subquery", func() {

			sourceRemoveTest(sourceRemove)

		})

		Convey("should not remove the source row if the source operator does not contains a subquery", func() {

			sourceRemove.hasSubquery = false
			sourceRemoveTest(sourceRemove)

		})

	})
}
