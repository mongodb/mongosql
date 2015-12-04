package sqlproxy

import (
	"github.com/10gen/sqlproxy/config"
	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type mockConnection struct {
	db string
}

func (mc mockConnection) LastInsertId() int64 {
	return 1
}
func (mc mockConnection) RowCount() int64 {
	return 1
}
func (mc mockConnection) ConnectionId() uint32 {
	return 1
}
func (mc mockConnection) DB() string {
	return mc.db
}

func TestConfigScanOperatorSelect(t *testing.T) {

	Convey("using config data source should to filter columns", t, func() {

		conn := mockConnection{"fo"}

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		_, values, err := eval.EvalSelect("information_schema", "select * from columns", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 8)

		_, values, err = eval.EvalSelect("information_schema", "select * from columns where COLUMN_NAME = 'f'", nil, conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("test", "SELECT TABLE_NAME, TABLE_COMMENT, TABLE_TYPE, TABLE_SCHEMA FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND ( TABLE_TYPE='BASE TABLE' OR TABLE_TYPE='VIEW' )", nil, &conn)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 0)

	})
}

func TestConfigScanOperatorTablesSelect(t *testing.T) {

	Convey("using config data source should to select tables", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		_, values, err := eval.EvalSelect("", "select * from information_schema.TABLES", nil, nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 3)

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE table_schema = 'test'", nil, nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE TABLE_SCHEMA = 'test'", nil, nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("", "select TABLE_NAME from information_schema.TABLES", nil, nil)
		So(err, ShouldBeNil)
		So(0, ShouldBeLessThan, len(string(values[0][0].(evaluator.SQLString))))

		_, values, err = eval.EvalSelect("", "select table_name from information_schema.TABLES", nil, nil)
		So(err, ShouldBeNil)
		So(0, ShouldBeLessThan, len(string(values[0][0].(evaluator.SQLString))))

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE table_schema LIKE 'test'", nil, nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)
	})
}

func TestConfigKeyColumnUsage(t *testing.T) {

	Convey("using config data source should to filter columns", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvaluator(cfg)
		So(err, ShouldBeNil)

		_, values, err := eval.EvalSelect("information_schema", "select * from KEY_COLUMN_USAGE", nil, nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 0)

	})
}
