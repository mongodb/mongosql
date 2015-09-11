package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"strconv"
)

func BuildValue(gExpr sqlparser.Expr) (SQLValue, error) {
	switch expr := gExpr.(type) {
	case sqlparser.NumVal:
		f, err := strconv.ParseFloat(sqlparser.String(expr), 64)
		if err != nil {
			return nil, err
		}
		return SQLNumeric(f), nil
	case *sqlparser.ColName:
		return SQLField{string(expr.Qualifier), string(expr.Name)}, nil
	case sqlparser.StrVal:
		return SQLString(string([]byte(expr))), nil
	case *sqlparser.BinaryExpr:
		// look up the function in the function map
		funcImpl, ok := funcMap[string(expr.Operator)]
		if !ok {
			return nil, fmt.Errorf("can't find implementation for binary operator '%v'", expr.Operator)
		}

		left, err := BuildValue(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := BuildValue(expr.Right)
		if err != nil {
			return nil, err
		}
		return &SQLFuncValue{[]SQLValue{left, right}, funcImpl}, nil
	default:
		panic(fmt.Errorf("BuildValue expr not yet implemented for %T", expr))
		return nil, nil
	}
}

// BuildMatcher rewrites a boolean expression as a matcher.
func BuildMatcher(gExpr sqlparser.Expr) (Matcher, error) {
	log.Logf(log.DebugLow, "expr: %#v (type is %T)", gExpr, gExpr)

	switch expr := gExpr.(type) {
	case *sqlparser.AndExpr:
		left, err := BuildMatcher(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := BuildMatcher(expr.Right)
		if err != nil {
			return nil, err
		}
		return &And{[]Matcher{left, right}}, nil
	case *sqlparser.OrExpr:
		left, err := BuildMatcher(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := BuildMatcher(expr.Right)
		if err != nil {
			return nil, err
		}
		return &Or{[]Matcher{left, right}}, nil
	case *sqlparser.ComparisonExpr:
		left, err := BuildValue(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := BuildValue(expr.Right)
		if err != nil {
			return nil, err
		}
		switch expr.Operator {
		case sqlparser.AST_EQ:
			return &Equals{left, right}, nil
		case sqlparser.AST_LT:
			return &LessThan{left, right}, nil
		case sqlparser.AST_GT:
			return &GreaterThan{left, right}, nil
		case sqlparser.AST_LE:
			return &LessThanOrEqual{left, right}, nil
		case sqlparser.AST_GE:
			return &GreaterThanOrEqual{left, right}, nil
		case sqlparser.AST_NE:
			return &NotEquals{left, right}, nil
		case sqlparser.AST_LIKE:
			return &Like{left, right}, nil
		default:
			return &Equals{left, right}, fmt.Errorf("sql where clause not implemented: %s", sqlparser.String(expr))
		}
	case *sqlparser.RangeCond:
		// BETWEEN
		panic("not implemented")
	case *sqlparser.NullCheck:
		panic("not implemented")
	case *sqlparser.UnaryExpr:
		panic("not implemented")
	case *sqlparser.NotExpr:
		child, err := BuildMatcher(expr.Expr)
		if err != nil {
			return nil, err
		}
		return &Not{child}, nil
	case *sqlparser.ParenBoolExpr:
		child, err := BuildMatcher(expr.Expr)
		if err != nil {
			return nil, err
		}
		return child, nil
	case *sqlparser.Subquery:
		panic("not implemented: subquery")
		return nil, nil
	case sqlparser.ValArg:
		panic("not implemented: function")
		return nil, fmt.Errorf("can't handle ValArg type %T", expr)
	case *sqlparser.FuncExpr:
		panic("not implemented: function")
		return nil, nil
	case *sqlparser.CaseExpr:
		panic("not implemented: case")
		return nil, nil
	case *sqlparser.ExistsExpr:
		panic("not implemented: exists")
		return nil, nil
	default:
		panic(fmt.Errorf("not implemented: %v", expr))
		return nil, nil
	}

}
