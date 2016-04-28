package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
)

type subqueryFinder struct {
	hasSq bool
}

// hasSubquery will take an expression and return true if it contains a subquery.
func hasSubquery(e SQLExpr) (bool, error) {

	sf := &subqueryFinder{}

	_, err := sf.Visit(e)
	if err != nil {
		return false, err
	}

	return sf.hasSq, nil
}

func (sf *subqueryFinder) Visit(e SQLExpr) (SQLExpr, error) {

	if sf.hasSq {
		return e, nil
	}

	switch e.(type) {

	case *SQLSubqueryExpr, *SQLSubqueryCmpExpr, *SQLExistsExpr:

		sf.hasSq = true

	default:

		return walk(sf, e)

	}

	return e, nil
}

type columnFinder struct {
	columns []*Column
	tables  map[string]*schema.Table
}

// referencedColumns will take an expression and return all the columns referenced in the expression
func referencedColumns(e SQLExpr, tables map[string]*schema.Table) ([]*Column, error) {

	cf := &columnFinder{tables: tables}

	_, err := cf.Visit(e)
	if err != nil {
		return nil, err
	}

	return cf.columns, nil
}

func (cf *columnFinder) Visit(e SQLExpr) (SQLExpr, error) {

	switch expr := e.(type) {

	case nil, SQLVarchar, SQLNullValue:

		return e, nil

	case *SQLScalarFunctionExpr:
		for _, funcArg := range expr.Exprs {
			_, err := cf.Visit(funcArg)
			if err != nil {
				return nil, err
			}
		}

	case *SQLAggFunctionExpr:

		for _, expr := range expr.Exprs {

			columns, err := referencedColumns(expr, cf.tables)
			if err != nil {
				return nil, err
			}

			cf.columns = append(cf.columns, columns...)
		}

	case SQLColumnExpr:

		column := &Column{
			Table:     string(expr.tableName),
			Name:      string(expr.columnName),
			View:      string(expr.columnName),
			MongoType: expr.columnType.MongoType,
			SQLType:   expr.columnType.SQLType,
		}

		cf.columns = append(cf.columns, column)

	case *SQLSubqueryCmpExpr:

		sExprs, err := referencedSelectExpressions(expr.value.stmt, cf.tables)
		if err != nil {
			return nil, err
		}

		_, err = cf.Visit(expr.left)
		if err != nil {
			return nil, err
		}

		cf.columns = append(cf.columns, SelectExpressions(sExprs).GetColumns()...)

	case *SQLSubqueryExpr:

		sExprs, err := referencedSelectExpressions(expr.stmt, cf.tables)
		if err != nil {
			return nil, err
		}

		cf.columns = append(cf.columns, SelectExpressions(sExprs).GetColumns()...)

	default:

		return walk(cf, expr)

	}

	return e, nil
}

type aggFunctionFinder struct {
	aggFuncs []*SQLAggFunctionExpr
}

// getAggFunctions will take an expression and return all
// aggregation functions it finds within the expression.
func getAggFunctions(e SQLExpr) ([]*SQLAggFunctionExpr, error) {

	af := &aggFunctionFinder{}

	_, err := af.Visit(e)
	if err != nil {
		return nil, err
	}

	return af.aggFuncs, nil
}

func (af *aggFunctionFinder) Visit(e SQLExpr) (SQLExpr, error) {

	switch typedE := e.(type) {

	case *SQLExistsExpr, SQLColumnExpr, SQLNullValue, SQLNumeric, SQLVarchar, *SQLSubqueryExpr:

		return e, nil

	case *SQLAggFunctionExpr:

		af.aggFuncs = append(af.aggFuncs, typedE)

	default:

		return walk(af, e)

	}

	return e, nil
}

// partiallyEvaluate will take an expression tree and partially evaluate any nodes that can
// evaluated without needing data from the database. If functions by using the
// nominateForPartialEvaluation function to gather candidates that are evaluatable. Then
// it walks the tree from top-down and, when it finds a candidate node, replaces the
// candidate node with the result of calling Evaluate on the candidate node.
func partiallyEvaluate(e SQLExpr) (SQLExpr, error) {
	candidates, err := nominateForPartialEvaluation(e)
	if err != nil {
		return nil, err
	}
	v := &partialEvaluator{candidates}
	return v.Visit(e)
}

type partialEvaluator struct {
	candidates map[SQLExpr]bool
}

func (pe *partialEvaluator) Visit(e SQLExpr) (SQLExpr, error) {
	if !pe.candidates[e] {
		return walk(pe, e)
	}

	// if we need an evaluation context, the partialEvaluatorNominator
	// is returning bad candidates.
	return e.Evaluate(nil)
}

// nominateForPartialEvaluation walks a SQLExpr tree from bottom up
// identifying nodes that are able to be evaluated without executing
// a query. It returns these identified nodes as candidates.
func nominateForPartialEvaluation(e SQLExpr) (map[SQLExpr]bool, error) {
	n := &partialEvaluatorNominator{
		candidates: make(map[SQLExpr]bool),
	}
	_, err := n.Visit(e)
	if err != nil {
		return nil, err
	}

	return n.candidates, nil
}

type partialEvaluatorNominator struct {
	blocked    bool
	candidates map[SQLExpr]bool
}

func (n *partialEvaluatorNominator) Visit(e SQLExpr) (SQLExpr, error) {
	oldBlocked := n.blocked
	n.blocked = false

	switch typedE := e.(type) {
	case *SQLExistsExpr:
		n.blocked = true
	case SQLColumnExpr:
		n.blocked = true
	case *SQLSubqueryCmpExpr:
		n.blocked = true
	case *SQLSubqueryExpr:
		n.blocked = true
	case *SQLAggFunctionExpr:
		n.blocked = true
	case *SQLScalarFunctionExpr:
		n.blocked = typedE.RequiresEvalCtx()
		if !n.blocked {
			_, err := walk(n, e)
			if err != nil {
				return nil, err
			}
		}
	default:
		_, err := walk(n, e)
		if err != nil {
			return nil, err
		}
	}

	if !n.blocked {
		n.candidates[e] = true
	}

	n.blocked = n.blocked || oldBlocked
	return e, nil
}

// normalize makes semantically equivalent expressions all
// look the same. For instance, it will make "3 > a" look like
// "a < 3".
func normalize(e SQLExpr) (SQLExpr, error) {
	v := &normalizer{}
	return v.Visit(e)
}

type normalizer struct{}

func (n *normalizer) Visit(e SQLExpr) (SQLExpr, error) {

	// walk the children first as they might get normalized
	// on the way up.
	e, err := walk(n, e)
	if err != nil {
		return nil, err
	}

	switch typedE := e.(type) {
	case *SQLAndExpr:
		if left, ok := typedE.left.(SQLValue); ok {
			matches, err := Matches(left, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return typedE.right, nil
			}
			return SQLFalse, nil
		}
		if right, ok := typedE.right.(SQLValue); ok {
			matches, err := Matches(right, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return typedE.left, nil
			}
			return SQLFalse, nil
		}
	case *SQLEqualsExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLEqualsExpr{typedE.right, typedE.left}, nil
		}
	case *SQLGreaterThanExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLLessThanExpr{typedE.right, typedE.left}, nil
		}
	case *SQLGreaterThanOrEqualExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLLessThanOrEqualExpr{typedE.right, typedE.left}, nil
		}
	case *SQLLessThanExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLGreaterThanExpr{typedE.right, typedE.left}, nil
		}
	case *SQLLessThanOrEqualExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLGreaterThanOrEqualExpr{typedE.right, typedE.left}, nil
		}
	case *SQLNotEqualsExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLNotEqualsExpr{typedE.right, typedE.left}, nil
		}
	case *SQLOrExpr:
		if left, ok := typedE.left.(SQLValue); ok {
			matches, err := Matches(left, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return SQLTrue, nil
			}
			return typedE.right, nil
		}
		if right, ok := typedE.right.(SQLValue); ok {
			matches, err := Matches(right, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return SQLTrue, nil
			}
			return typedE.left, nil
		}
	case *SQLTupleExpr:
		if len(typedE.Exprs) == 1 {
			return typedE.Exprs[0], nil
		}
	case *SQLValues:
		if len(typedE.Values) == 1 {
			return typedE.Values[0], nil
		}
	}

	return e, nil
}

func shouldFlip(n sqlBinaryNode) bool {
	if _, ok := n.left.(SQLValue); ok {
		if _, ok := n.right.(SQLValue); !ok {
			return true
		}
	}

	return false
}

type aggFunctionExprReplacer struct {
	tableName string
}

func replaceAggFunctionsWithColumns(tableName string, e SQLExpr) (SQLExpr, error) {
	v := &aggFunctionExprReplacer{tableName}
	return v.Visit(e)
}

func (v *aggFunctionExprReplacer) Visit(e SQLExpr) (SQLExpr, error) {
	switch typedE := e.(type) {
	case *SQLAggFunctionExpr:
		columnType := schema.ColumnType{typedE.Type(), schema.MongoNone}
		return SQLColumnExpr{v.tableName, typedE.String(), columnType}, nil
	default:
		return walk(v, e)
	}
}
