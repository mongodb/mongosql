package evaluator

import (
	"github.com/10gen/sqlproxy/config"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func sourceAppendTest(operator *SourceAppend) {

	cfg, err := config.ParseConfigData(testConfig1)
	So(err, ShouldBeNil)

	ctx := &ExecutionCtx{
		Config: cfg,
		Db:     dbName,
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
