package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
)

// bsonDToValues takes a bson.D document and returns
// the corresponding values.
func bsonDToValues(tableName string, document bson.D) ([]Value, error) {
	values := []Value{}
	for _, v := range document {
		value, err := NewSQLValue(v.Value, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return nil, err
		}
		values = append(values, Value{tableName, v.Name, value})
	}
	return values, nil
}

func constructProjectedColumns(exprs map[string]SQLExpr, values ...string) (projectedColumns ProjectedColumns) {
	for _, value := range values {

		expr := exprs[value]

		column := &Column{
			Name: value,
		}

		projectedColumns = append(projectedColumns, ProjectedColumn{
			Column: column,
			Expr:   expr,
		})
	}
	return
}

func constructOrderByTerms(exprs map[string]SQLExpr, values ...string) (terms []*orderByTerm) {
	for i, v := range values {

		term := &orderByTerm{
			expr:      exprs[v],
			ascending: i%2 == 0,
		}

		terms = append(terms, term)
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
