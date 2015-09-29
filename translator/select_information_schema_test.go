package translator

import (
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestConfigScanOperatorSelect(t *testing.T) {

	Convey("using config data source should to filter columns", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
		So(err, ShouldBeNil)

		_, values, err := eval.EvalSelect("information_schema", "select * from columns", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 8)

		_, values, err = eval.EvalSelect("information_schema", "select * from columns where COLUMN_NAME = 'f'", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("test", "SELECT TABLE_NAME, TABLE_COMMENT, TABLE_TYPE, TABLE_SCHEMA FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND ( TABLE_TYPE='BASE TABLE' OR TABLE_TYPE='VIEW' )", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 0)

	})
}

func TestConfigScanOperatorTablesSelect(t *testing.T) {

	Convey("using config data source should to select tables", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
		So(err, ShouldBeNil)

		_, values, err := eval.EvalSelect("", "select * from information_schema.TABLES", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 3)

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE table_schema = 'test'", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE TABLE_SCHEMA = 'test'", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

		_, values, err = eval.EvalSelect("", "select TABLE_NAME from information_schema.TABLES", nil)
		So(err, ShouldBeNil)
		So(0, ShouldBeLessThan, len(string(values[0][0].(evaluator.SQLString))))

		_, values, err = eval.EvalSelect("", "select table_name from information_schema.TABLES", nil)
		So(err, ShouldBeNil)
		So(0, ShouldBeLessThan, len(string(values[0][0].(evaluator.SQLString))))

		_, values, err = eval.EvalSelect("", "select * from information_schema.TABLES WHERE table_schema LIKE 'test'", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)
	})
}

func TestConfigKeyColumnUsage(t *testing.T) {

	Convey("using config data source should to filter columns", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
		So(err, ShouldBeNil)

		_, values, err := eval.EvalSelect("information_schema", "select * from KEY_COLUMN_USAGE", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 0)

	})
}
