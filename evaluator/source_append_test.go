package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func sourceAppendTest(operator *SourceAppend) {

	ctx := &ExecutionCtx{
		Config:  cfgOne,
		Db:      dbOne,
		Session: session,
	}

	operator.source = &TableScan{
		ctx:       ctx,
		tableName: tableOneName,
	}

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

func TestSourceAppendOperator(t *testing.T) {

	Convey("A source append operator...", t, func() {

		sourceAppend := &SourceAppend{hasSubquery: true}

		Convey("should append the source row if the source operator contains a subquery", func() {

			sourceAppendTest(sourceAppend)

		})

		Convey("should not append the source row if the source operator does not contains a subquery", func() {

			sourceAppend.hasSubquery = false
			sourceAppendTest(sourceAppend)

		})

	})
}
