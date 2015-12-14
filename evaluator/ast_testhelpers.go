package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
)

func getWhereSQLExprFromSQL(sql string) (SQLExpr, error) {
	// Parse the statement, algebrize it, extract the WHERE clause and build a matcher from it.
	raw, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}
	if selectStatement, ok := raw.(*sqlparser.Select); ok {
		cfg, err := config.ParseConfigData(testConfig3)
		So(err, ShouldBeNil)
		parseCtx, err := NewParseCtx(selectStatement, cfg, dbOne)
		if err != nil {
			return nil, err
		}

		parseCtx.Database = dbOne

		err = AlgebrizeStatement(selectStatement, parseCtx)
		if err != nil {
			return nil, err
		}

		return NewSQLExpr(selectStatement.Where.Expr)
	}
	return nil, fmt.Errorf("statement doesn't look like a 'SELECT'")
}
