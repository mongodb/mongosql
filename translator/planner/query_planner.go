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

func planSelectExpr(ast *sqlparser.Select) (source Operator, err error) {
	// handles from and where expression
	if ast.From != nil {
		source, err = planFromExpr(ast.From, ast.Where)
		if err != nil {
			return nil, err
		}
	}

	// handle select expressions
	operator := &Select{source: source}
	sColumns := []SelectColumn{}

	if err = setReferencedColumns(ast, &sColumns, 0); err != nil {
		return nil, err
	}

	for _, sc := range sColumns {
		// TODO: check columns referenced in nested levels
		for _, column := range sc.Columns {
			operator.Columns = append(operator.Columns, *column)
		}
	}

	return operator, nil

}

// setReferencedColumns adds any columns referenced in the select statement
// to s. level indicates what nesting level the columna appears in relative
// to s.
func setReferencedColumns(ss sqlparser.SelectStatement, s *[]SelectColumn, level int) error {

	columns := make([]*Column, 0)

	switch ast := ss.(type) {

	case *sqlparser.Select:
		for _, sExpr := range ast.SelectExprs {

			switch expr := sExpr.(type) {

			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				continue

			case *sqlparser.NonStarExpr:

				refColumns, err := getReferencedColumn(expr.Expr, s, level)
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
				return fmt.Errorf("unreachable path")
			}
		}

	case *sqlparser.Union:
		err := setReferencedColumns(ast.Left, s, level)
		if err != nil {
			return err
		}

		err = setReferencedColumns(ast.Right, s, level)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown select statement: %T", ast)
	}

	*s = append(*s, SelectColumn{columns, level})

	return nil
}

// planFromExpr takes one or more table expressions and returns an Operator for the
// data source. If more than one table expression exists, it constructs a join
// operator (which is left deep for more than two table expressions).
func planFromExpr(tExpr sqlparser.TableExprs, where *sqlparser.Where) (Operator, error) {

	if len(tExpr) == 0 {
		return nil, fmt.Errorf("can plan table expression with no tables")
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

func random() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
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
		ns := &Column{
			Table: random(),
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

// getReferencedColumn returns a slice of referenced columns in an expression.
func getReferencedColumn(e sqlparser.Expr, s *[]SelectColumn, level int) ([]*Column, error) {

	log.Logf(log.DebugLow, "getReferencedColumn: %#v (type is %T)", e, e)

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

		return getReferencedColumns(s, level, expr.Left, expr.Right)

	case *sqlparser.AndExpr:

		return getReferencedColumns(s, level, expr.Left, expr.Right)

	case *sqlparser.OrExpr:

		return getReferencedColumns(s, level, expr.Left, expr.Right)

	case *sqlparser.ParenBoolExpr:

		return getReferencedColumn(expr.Expr, s, level)

	case *sqlparser.ComparisonExpr:

		return getReferencedColumns(s, level, expr.Left, expr.Right)

	case *sqlparser.RangeCond:

		return getReferencedColumns(s, level, expr.From, expr.To, expr.Left)

	case *sqlparser.NullCheck:

		return getReferencedColumn(expr.Expr, s, level)

	case *sqlparser.UnaryExpr:

		return getReferencedColumn(expr.Expr, s, level)

	case *sqlparser.NotExpr:

		return getReferencedColumn(expr.Expr, s, level)

	case *sqlparser.Subquery:
		switch ast := expr.Select.(type) {

		case *sqlparser.Select:
			if err := setReferencedColumns(ast, s, level+1); err != nil {
				return nil, err
			}

		case *sqlparser.Union:
			if err := setReferencedColumns(ast, s, level+1); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unknown select statement: %T", ast)
		}

		// TODO: fill these in
	case sqlparser.ValArg:
		return nil, fmt.Errorf("referenced columns for ValArg for NYI")
	case *sqlparser.FuncExpr:
		return nil, fmt.Errorf("referenced columns for FuncExpr for NYI")
	case *sqlparser.CaseExpr:
		return nil, fmt.Errorf("referenced columns for CaseExpr for NYI")
	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("referenced columns for ExistsExpr for NYI")
	default:
		return nil, fmt.Errorf("referenced columns NYI for: %T", expr)
	}

	return nil, fmt.Errorf("referenced columns (on %T) reached an unreachable point", e)
}

// getReferencedColumns accepts several exppressions and returns a slice
// of referenced columns in those expressions.
func getReferencedColumns(s *[]SelectColumn, l int, e ...sqlparser.Expr) ([]*Column, error) {
	columns := make([]*Column, 0)

	for _, expr := range e {
		refColumns, err := getReferencedColumn(expr, s, l)
		if err != nil {
			return nil, err
		}
		columns = append(columns, refColumns...)
	}

	return columns, nil
}
