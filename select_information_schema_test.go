package sqlproxy_test

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaScanOperatorSelect(t *testing.T) {
	env := setupEnv(t)
	eval := env.eval
	conn := env.dbConn("fo")

	Convey("using config data source should to filter columns", t, func() {

		_, values, err := eval.EvalSelect("information_schema", "select * from columns", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 8)

		_, values, err = eval.EvalSelect("information_schema", "select * from columns where COLUMN_NAME = 'f'", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("test", "SELECT TABLE_NAME, TABLE_COMMENT, TABLE_TYPE, TABLE_SCHEMA FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND ( TABLE_TYPE='BASE TABLE' OR TABLE_TYPE='VIEW' )", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 0)

	})
}

func TestSchemaScanOperatorTablesSelect(t *testing.T) {
	env := setupEnv(t)
	eval := env.eval
	conn := env.conn()

	Convey("using config data source should to select tables", t, func() {

		_, values, err := eval.EvalSelect("", "select * from information_schema.TABLES", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 3)

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE table_schema = 'test'", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE TABLE_SCHEMA = 'test'", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("", "select TABLE_NAME from information_schema.TABLES", nil, conn)
		So(err, ShouldBeNil)
		So(0, ShouldBeLessThan, len(string(values[0][0].(evaluator.SQLVarchar))))

		_, values, err = eval.EvalSelect("", "select table_name from information_schema.TABLES", nil, conn)
		So(err, ShouldBeNil)
		So(0, ShouldBeLessThan, len(string(values[0][0].(evaluator.SQLVarchar))))

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE table_schema LIKE 'test'", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)
	})
}

func TestSchemaKeyColumnUsage(t *testing.T) {
	env := setupEnv(t)
	eval := env.eval
	conn := env.conn()

	Convey("using config data source should to filter columns", t, func() {

		_, values, err := eval.EvalSelect("information_schema", "select * from KEY_COLUMN_USAGE", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 0)

	})
}
