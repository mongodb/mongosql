package planner

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
)

// PlanQuery translates the SQL SELECT statement into an
// algebra tree representation of logical operators.
func PlanQuery(ss sqlparser.SelectStatement) (Operator, error) {

	log.Logf(log.DebugLow, "Planning query for %#v\n", ss)

	switch ast := ss.(type) {

	case *sqlparser.Select:
		return planSelectExpr(ast)

	case *sqlparser.Union:
		return nil, fmt.Errorf("union select statement not yet implemented")

	default:
		return nil, fmt.Errorf("unknown select statement: %T", ast)
	}
}

// planSelectExpr takes a select struct and returns a query execution plan for it.
func planSelectExpr(ast *sqlparser.Select) (operator Operator, err error) {
	// handles from and where expression
	if ast.From != nil {
		operator, err = planFromExpr(ast.From, ast.Where)
		if err != nil {
			return nil, err
		}
	}

	// handle select expressions
	selectOperator := &Select{source: operator}
	selectColumns := []SelectColumn{}

	if err = refColsInSelectStmt(&selectColumns, 0, ast); err != nil {
		return nil, err
	}

	for _, sc := range selectColumns {
		// TODO: check columns referenced in nested levels
		for _, column := range sc.Columns {
			selectOperator.Columns = append(selectOperator.Columns, *column)
		}
	}

	operator = selectOperator

	// handle group by expression
	if len(ast.GroupBy) != 0 {

		gb := &GroupBy{source: operator, fields: selectColumns}

		for _, valExpr := range ast.GroupBy {
			expr, ok := sqlparser.Expr(valExpr).(*sqlparser.ColName)
			if !ok {
				return nil, fmt.Errorf("unsupported group by term: %T", valExpr)
			}

			/*
				// a GROUP BY clause can't refer to nonaggregated columns in
				// the select list that are not named in the GROUP BY clause
				if !isValidGroupByTerm(selectColumns, expr) {
					return nil, fmt.Errorf("group by term '%v' not in select list", string(expr.Name))
				}

			*/

			gb.exprs = append(gb.exprs, expr)

		}

		operator = gb
	}

	return operator, nil

}

// isValidGroupByTerm returns true if the column expression is valid as a group
// by term within a select column context.
func isValidGroupByTerm(scs []SelectColumn, expr *sqlparser.ColName) bool {
	for _, sc := range scs {
		if sc.Level == 0 {
			for _, c := range sc.Columns {
				// TODO: support alias in group by term?
				if c.Table == string(expr.Qualifier) && c.Name == string(expr.Name) {
					return true
				}
			}
		}
	}
	return false
}

// refColsInSelectExpr adds any columns referenced in the select expression
// to s. level indicates what nesting level the columna appears in relative
// to s.
func refColsInSelectExpr(s *[]SelectColumn, level int, sExprs sqlparser.SelectExprs) error {
	columns := make([]*Column, 0)

	for _, sExpr := range sExprs {

		switch expr := sExpr.(type) {

		// mixture of star and non-star expression is acceptable
		case *sqlparser.StarExpr:
			continue

		case *sqlparser.NonStarExpr:

			refColumns, err := getReferencedColumns(s, level, expr.Expr)
			if err != nil {
				return err
			}

			// TODO: check columns against table config
			for _, column := range refColumns {
				column.View = string(expr.As)

				if column.View == "" {
					column.View = column.Name
				}

				columns = append(columns, column)
			}

		default:
			return fmt.Errorf("unknown SelectExprs in refColsInSelectExpr: %T", expr)
		}
	}

	*s = append(*s, SelectColumn{columns, level})

	return nil
}

// refColsInSelectStmt adds any columns referenced in the select statement
// to s. level indicates what nesting level the columna appears in relative
// to s.
func refColsInSelectStmt(s *[]SelectColumn, level int, ss sqlparser.SelectStatement) error {

	switch stmt := ss.(type) {

	case *sqlparser.Select:
		err := refColsInSelectExpr(s, level, stmt.SelectExprs)
		if err != nil {
			return err
		}

	case *sqlparser.Union:
		err := refColsInSelectStmt(s, level, stmt.Left)
		if err != nil {
			return err
		}

		err = refColsInSelectStmt(s, level, stmt.Right)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown SelectStatement in refColsInSelectStmt: %T", stmt)
	}

	return nil
}

// planFromExpr takes one or more table expressions and returns an Operator for the
// data source. If more than one table expression exists, it constructs a join
// operator (which is left deep for more than two table expressions).
func planFromExpr(tExpr sqlparser.TableExprs, where *sqlparser.Where) (Operator, error) {

	if len(tExpr) == 0 {
		return nil, fmt.Errorf("can't plan table expression with no tables")
	} else if len(tExpr) == 1 {
		return planTableExpr(tExpr[0], where)
	}

	var left, right Operator
	var err error

	if len(tExpr) == 2 {

		left, err = planTableExpr(tExpr[0], where)
		if err != nil {
			return nil, fmt.Errorf("error planning left table expr: %v", err)
		}

		right, err = planTableExpr(tExpr[1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning right table expr: %v", err)
		}

	} else {

		left, err = planFromExpr(tExpr[:len(tExpr)-1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning left forest: %v", err)
		}

		right, err = planTableExpr(tExpr[len(tExpr)-1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning right table leaf: %v", err)
		}

	}

	join := &Join{
		left:  left,
		right: right,
		kind:  sqlparser.AST_CROSS_JOIN,
	}

	return join, nil

}

// planTableExpr takes a table expression and returns an Operator that
// can iterate over its result set.
func planTableExpr(tExpr sqlparser.TableExpr, where *sqlparser.Where) (Operator, error) {
	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:
		// this is a simple table to get data from
		return planSimpleTableExpr(expr.Expr, where)

	case *sqlparser.ParenTableExpr:
		return planTableExpr(expr.Expr, where)

	case *sqlparser.JoinTableExpr:
		left, err := planTableExpr(expr.LeftExpr, where)
		if err != nil {
			return nil, fmt.Errorf("error on left join node: %v", err)
		}

		right, err := planTableExpr(expr.RightExpr, where)
		if err != nil {
			return nil, fmt.Errorf("error on right join node: %v", err)
		}

		join := &Join{
			left:  left,
			right: right,
			on:    expr.On,
			kind:  expr.Join,
		}

		return join, nil

	default:
		return nil, fmt.Errorf("can't handle table expression type %T", expr)
	}

	return nil, fmt.Errorf("unreachable in planTableExpr")
}

// planSimpleTableExpr takes a simple table expression and returns an operator that can iterate
// over its result set.
func planSimpleTableExpr(stExpr sqlparser.SimpleTableExpr, where *sqlparser.Where) (Operator, error) {
	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		ts := &TableScan{dbName: string(expr.Qualifier), tableName: string(expr.Name)}

		if where != nil {
			// create a matcher that can evaluate the WHERE expression
			matcher, err := BuildMatcher(where.Expr)
			if err != nil {
				return nil, err
			}

			// TODO currently the transformation is all-or-nothing either the entire query is
			// executed inside mongo or inside the matcher. Needs update to prune the matcher tree
			// so that the part of the query that can be expressed with MQL is extracted and passed
			// to mongo, and the rest of the filtering can be done by the (simplified) matcher
			if transformed, err := matcher.Transform(); err == nil {
				ts.filter = transformed
				ts.filterMatcher = matcher
				return ts, nil
			}
			return &MatchOperator{source: ts, matcher: matcher}, nil
		}
		return ts, nil

	case *sqlparser.Subquery:
		return PlanQuery(expr.Select)

	default:
		return nil, fmt.Errorf("can't yet handle simple table expression type %T", expr)
	}

}

func planExpr(sqlExpr sqlparser.Expr) (Operator, *Column, error) {

	log.Logf(log.DebugLow, "planExpr: %#v (type is %T)", sqlExpr, sqlExpr)

	switch expr := sqlExpr.(type) {

	case sqlparser.NumVal:
	case sqlparser.ValTuple:
	case *sqlparser.NullVal:
	case *sqlparser.ColName:
		ns := &Column{
			Table: string(expr.Qualifier),
			Name:  string(expr.Name),
		}

		return &Noop{}, ns, nil
	case sqlparser.StrVal:
	case *sqlparser.BinaryExpr:
	case *sqlparser.AndExpr:
	case *sqlparser.OrExpr:
	case *sqlparser.ComparisonExpr:
	case *sqlparser.RangeCond:
	case *sqlparser.NullCheck:
	case *sqlparser.UnaryExpr:
	case *sqlparser.NotExpr:
	case *sqlparser.ParenBoolExpr:
	case *sqlparser.Subquery:
		op, err := PlanQuery(expr.Select)
		b := make([]byte, 16)
		rand.Read(b)
		ns := &Column{
			Table: hex.EncodeToString(b),
		}
		return op, ns, err
	case sqlparser.ValArg:
	case *sqlparser.FuncExpr:
	case *sqlparser.CaseExpr:
	case *sqlparser.ExistsExpr:
	default:
		return nil, nil, fmt.Errorf("can't handle expression type %T", expr)
	}

	return nil, nil, nil
}

// getReferencedColumns accepts several exppressions and returns a slice
// of referenced columns in those expressions.
func getReferencedColumns(s *[]SelectColumn, l int, exprs ...sqlparser.Expr) ([]*Column, error) {

	log.Logf(log.DebugLow, "getReferencedColumns: %#v (type is %T)", exprs, exprs)

	columns := make([]*Column, 0)

	for _, e := range exprs {
		refColumns, err := referencedColumns(s, l, e)
		if err != nil {
			return nil, err
		}
		columns = append(columns, refColumns...)
	}

	return columns, nil
}

// referencedColumns returns a slice of referenced columns in an expression.
func referencedColumns(s *[]SelectColumn, l int, e sqlparser.Expr) ([]*Column, error) {

	log.Logf(log.DebugLow, "referencedColumns: %#v (type is %T)", e, e)

	switch expr := e.(type) {

	case sqlparser.NumVal:
	case sqlparser.ValTuple:
	case sqlparser.StrVal:
	case *sqlparser.NullVal:
	case *sqlparser.ColName:

		c := &Column{
			Table: string(expr.Qualifier),
			Name:  string(expr.Name),
		}
		return []*Column{c}, nil

	case *sqlparser.BinaryExpr:

		return getReferencedColumns(s, l, expr.Left, expr.Right)

	case *sqlparser.AndExpr:

		return getReferencedColumns(s, l, expr.Left, expr.Right)

	case *sqlparser.OrExpr:

		return getReferencedColumns(s, l, expr.Left, expr.Right)

	case *sqlparser.ParenBoolExpr:

		return referencedColumns(s, l, expr.Expr)

	case *sqlparser.ComparisonExpr:

		return getReferencedColumns(s, l, expr.Left, expr.Right)

	case *sqlparser.RangeCond:

		return getReferencedColumns(s, l, expr.From, expr.To, expr.Left)

	case *sqlparser.NullCheck:

		return referencedColumns(s, l, expr.Expr)

	case *sqlparser.UnaryExpr:

		return referencedColumns(s, l, expr.Expr)

	case *sqlparser.NotExpr:

		return referencedColumns(s, l, expr.Expr)

	case *sqlparser.Subquery:
		switch sbq := expr.Select.(type) {

		case *sqlparser.Select:
			if err := refColsInSelectStmt(s, l+1, sbq); err != nil {
				return nil, err
			}

		case *sqlparser.Union:
			if err := refColsInSelectStmt(s, l+1, sbq); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unknown Subquery type: %T", sbq)
		}

		// TODO: test
		return nil, nil

	case *sqlparser.FuncExpr:

		if err := refColsInSelectExpr(s, l, expr.Exprs); err != nil {
			return nil, err
		}
		// TODO: test
		return nil, nil

		// TODO: fill these in
	case sqlparser.ValArg:
		return nil, fmt.Errorf("referenced columns for ValArg for NYI")

	case *sqlparser.CaseExpr:
		return nil, fmt.Errorf("referenced columns for CaseExpr for NYI")
	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("referenced columns for ExistsExpr for NYI")
	default:
		return nil, fmt.Errorf("referenced columns NYI for: %T", expr)
	}

	return nil, fmt.Errorf("referenced columns (on %T) reached an unreachable point", e)
}
