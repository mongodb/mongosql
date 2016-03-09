package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
)

func constructSelectExpressions(exprs map[string]SQLExpr, values ...string) (sExprs SelectExpressions) {
	for _, value := range values {

		expr := exprs[value]

		column := &Column{
			Name: expr.String(),
			View: value,
		}

		sExprs = append(sExprs, SelectExpression{
			Column:     column,
			Expr:       expr,
			Referenced: true,
		})
	}
	return
}

func constructOrderByKeys(exprs map[string]SQLExpr, values ...string) (keys []orderByKey) {
	sExprs := constructSelectExpressions(exprs, values...)

	for i := range sExprs {

		key := orderByKey{
			expr:      &sExprs[i],
			ascending: i%2 == 0,
		}

		keys = append(keys, key)
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

		return NewSQLExpr(selectStatement.Where.Expr, schema.Databases[dbOne].Tables)
	}
	return nil, fmt.Errorf("statement doesn't look like a 'SELECT'")
}
