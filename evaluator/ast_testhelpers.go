package evaluator

import (
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

func getBinaryExprLeaves(expr SQLExpr) (SQLExpr, SQLExpr) {
	switch typedE := expr.(type) {
	case *SQLAndExpr:
		return typedE.left, typedE.right
	case *SQLAddExpr:
		return typedE.left, typedE.right
	case *SQLSubtractExpr:
		return typedE.left, typedE.right
	case *SQLMultiplyExpr:
		return typedE.left, typedE.right
	case *SQLDivideExpr:
		return typedE.left, typedE.right
	case *SQLEqualsExpr:
		return typedE.left, typedE.right
	case *SQLLessThanExpr:
		return typedE.left, typedE.right
	case *SQLGreaterThanExpr:
		return typedE.left, typedE.right
	case *SQLLessThanOrEqualExpr:
		return typedE.left, typedE.right
	case *SQLGreaterThanOrEqualExpr:
		return typedE.left, typedE.right
	case *SQLLikeExpr:
		return typedE.left, typedE.right
	case *SQLSubqueryExpr:
		return nil, &SQLTupleExpr{typedE.Exprs()}
	//case *SQLSubqueryCmpExpr:
	// return typedE.left, &SQLTupleExpr{typedE.value.exprs}
	case *SQLInExpr:
		return typedE.left, typedE.right
	}
	return nil, nil
}

func getSQLExpr(schema *schema.Schema, dbName, tableName, sql string) (SQLExpr, error) {
	statement, err := sqlparser.Parse("select * from " + tableName + " where " + sql)
	if err != nil {
		return nil, err
	}

	selectStatement := statement.(sqlparser.SelectStatement)
	actualPlan, err := Algebrize(selectStatement, dbName, schema)
	if err != nil {
		return nil, err
	}

	expr := ((actualPlan.(*ProjectStage)).source.(*FilterStage)).matcher
	if conv, ok := expr.(*SQLConvertExpr); ok {
		expr = conv.expr
	}

	return expr, nil
}
