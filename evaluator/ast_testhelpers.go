package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
)

func constructSelectExpressions(exprs map[string]SQLExpr, values ...string) (sExprs SelectExpressions) {

	for i, value := range values {

		expr := exprs[value]

		column := &Column{
			Name: expr.String(),
			View: expr.String(),
		}

		sExprs = append(sExprs, SelectExpression{
			Column:     column,
			Expr:       expr,
			Referenced: (i % 2) == 0,
		})
	}

	return

}

func getWhereSQLExprFromSQL(schema *schema.Schema, sql string) (SQLExpr, error) {
	// Parse the statement, algebrize it, extract the WHERE clause and build a matcher from it.
	raw, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}
	if selectStatement, ok := raw.(*sqlparser.Select); ok {
		parseCtx, err := NewParseCtx(selectStatement, schema, dbOne)
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
