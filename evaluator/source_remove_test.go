package evaluator

import (
	"github.com/10gen/sqlproxy/config"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func sourceRemoveTest(operator *SourceRemove) {

	cfg, err := config.ParseConfigData(testConfig1)
	So(err, ShouldBeNil)

	ctx := &ExecutionCtx{
		Config:  cfg,
		Db:      dbName,
		SrcRows: []*Row{&Row{}},
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
