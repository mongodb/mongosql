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

	// select node is always on top of the tree
	selectNode := &Select{}

	switch ast := ss.(type) {

	case *sqlparser.Select:
		// interpret select expressions
		for _, sExpr := range ast.SelectExprs {
			switch expr := sExpr.(type) {
			// TODO: validate no mixture of star and non-star expression
			case *sqlparser.StarExpr:
				selectNode.isStar = true
				break

			case *sqlparser.NonStarExpr:
				plan, column, err := planExpr(expr.Expr)
				if err != nil {
					return nil, err
				}
				column.View = string(expr.As)
				if column.View == "" {
					column.View = column.Name
				}
				selectNode.Columns = append(selectNode.Columns, column)
				if _, ok := plan.(*Noop); !ok {
					selectNode.children = append(selectNode.children, plan)
				}
			default:
				return nil, fmt.Errorf("unreachable path")
			}
		}

		// interpret 'FROM' clause
		if ast.From != nil {
			plan, err := planFromExpr(ast.From, ast.Where)
			if err != nil {
				return nil, err
			}

			selectNode.children = append(selectNode.children, plan)
		}

		return selectNode, nil

	default:
		return nil, fmt.Errorf("%T select statement not yet implemented", ast)
	}
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

// planTableExpr takes a table expression and returns an Operator that can iterate over its result
// set.
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

			// Try to transform the match expression into mongo query language
			// if it is successful, use the query and skip the matcher

			// TODO currently the trasnformation is all-or-nothing either the entire query is
			// executed inside mongo or inside the matcher. Needs update to prune the matcher tree
			// so that the part of the query that can be expressed with MQL is extracted and passed
			// to mongo, and the rest of hte filtering can be done by the (simplified) matcher
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

	log.Logf(log.DebugLow, "planExpr: %s (type is %T)", sqlparser.String(sqlExpr), sqlExpr)

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
