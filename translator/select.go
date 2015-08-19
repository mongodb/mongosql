package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/util"
	"github.com/siddontang/mixer/sqlparser"
	"strconv"
)

func translateSelectStatement(stmt sqlparser.SelectStatement, pCtx *ParseCtx) (interface{}, error) {

	switch expr := stmt.(type) {

	case *sqlparser.Select:
		return getAlgebrizedQuery(expr, pCtx)

	default:
		return nil, fmt.Errorf("can't handle expression type %T", expr)
	}
	return nil, nil
}

// translateExpr takes an expression and returns its translated form.
func translateExpr(gExpr sqlparser.Expr, pCtx *ParseCtx) (interface{}, error) {
	log.Logf(log.DebugLow, "expr: %s (type is %T)", sqlparser.String(gExpr), gExpr)

	switch expr := gExpr.(type) {

	case sqlparser.NumVal:
		val, err := getNumVal(expr)
		if err != nil {
			return nil, fmt.Errorf("can't handle NumVal %v: %v", expr, err)
		}
		return NumVal{val}, err

	case sqlparser.ValTuple:
		vals := sqlparser.ValExprs(expr)
		tuple := ValTuple{}

		for i, val := range vals {
			t, err := translateExpr(val, pCtx)
			if err != nil {
				return nil, fmt.Errorf("can't handle ValExpr (%v) %v: %v", i, val, err)
			}
			tuple.Children = append(tuple.Children, t)
		}
		return tuple, nil

	case *sqlparser.NullVal:
		return NullVal{}, nil

		// TODO: regex lowercased
	case *sqlparser.ColName:
		return ColName{pCtx.ColumnName(sqlparser.String(expr))}, nil

	case sqlparser.StrVal:
		return StrVal{sqlparser.String(expr)}, nil

	case *sqlparser.BinaryExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("BinaryExpr LR error: %v", err)
		}

		// TODO ?: floats, complex values, strings
		leftVal, err := util.ToInt(left)
		if err != nil {
			return nil, fmt.Errorf("BinaryExpr leftVal error (%v): %v", left, err)
		}

		rightVal, err := util.ToInt(right)
		if err != nil {
			return nil, fmt.Errorf("BinaryExpr rightVal error (%v): %v", right, err)
		}
		return BinaryExpr{leftVal, expr.Operator, rightVal}, nil

	case *sqlparser.AndExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("AndExpr error: %v", err)
		}
		return AndExpr{left, right}, nil

	case *sqlparser.OrExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("OrExpr error: %v", err)
		}
		return OrExpr{left, right}, nil

	case *sqlparser.ComparisonExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("ComparisonExpr error: %v", err)
		}

		tLeft, okLeft := left.(string)
		tRight, okRight := right.(string)
		if okLeft && okRight {
			// TODO: test this
			return FieldComp{tLeft, oprtMap[expr.Operator], tRight}, nil
		}
		// TODO: verify structure
		return ComparisonExpr{left, oprtMap[expr.Operator], right}, nil

	case *sqlparser.RangeCond:
		from, to, err := translateLRExpr(expr.From, expr.To, pCtx)
		if err != nil {
			return nil, fmt.Errorf("RangeCond LR error: %v", err)
		}

		left, err := translateExpr(expr.Left, pCtx)
		if err != nil {
			return nil, fmt.Errorf("RangeCond key error: %v", err)
		}

		switch tLeft := left.(type) {
		case string:
			return RangeCond{from, tLeft, to}, nil
		default:
			return nil, fmt.Errorf("RangeCond key type error: %v (%T)", tLeft, tLeft)
		}

		// TODO: how is 'null' interpreted? exists? 'null'?
	case *sqlparser.NullCheck:
		val, err := translateExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("NullCheck error: %v", err)
		}

		switch tVal := val.(type) {
		case string:
			return NullCheck{tVal}, nil
		default:
			// TODO: can node not be a string?
			return nil, fmt.Errorf("NullCheck left type error: %v (%T)", val, tVal)
		}

	case *sqlparser.UnaryExpr:
		val, err := translateExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("UnaryExpr error: %v", err)
		}

		intVal, err := util.ToInt(val)
		if err != nil {
			return nil, fmt.Errorf("UnaryExpr conversion error (%v): %v", val, err)
		}

		switch expr.Operator {
		case sqlparser.AST_UPLUS:
			return UnaryExpr{intVal}, nil
		case sqlparser.AST_UMINUS:
			return UnaryExpr{-intVal}, nil
		case sqlparser.AST_TILDA:
			return UnaryExpr{^intVal}, nil
		default:
			return nil, fmt.Errorf("can't handle UnaryExpr operator type %T", expr.Operator)
		}

	case *sqlparser.NotExpr:
		val, err := translateExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("NotExpr error: %v", err)
		}

		return NotExpr{val}, err

	case *sqlparser.ParenBoolExpr:
		val, err := translateExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("ParenBoolExpr error: %v", err)
		}

		return val, err

		//
		//  some nodes rely on SimpleSelect support
		//

	case *sqlparser.Subquery:
		val, err := translateSelectStatement(expr.Select, pCtx)
		if err != nil {
			return nil, fmt.Errorf("Subquery error: %v", err)
		}
		return Subquery{val}, err

	case sqlparser.ValArg:
		return nil, fmt.Errorf("can't handle ValArg type %T", expr)

	case *sqlparser.FuncExpr:
		return nil, fmt.Errorf("can't handle FuncExpr type %T", expr)

		// TODO: might require resultset post-processing
	case *sqlparser.CaseExpr:
		return nil, fmt.Errorf("can't handle CaseExpr type %T", expr)

	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("can't handle ExistsExpr type %T", expr)

	default:
		return nil, fmt.Errorf("can't handle expression type %T", expr)
	}

}

// translateTableExpr takes a table expression and returns its translated form.
func translateTableExpr(tExpr sqlparser.TableExpr, pCtx *ParseCtx) (interface{}, error) {

	log.Logf(log.DebugLow, "table expr: %s (type is %T)", sqlparser.String(tExpr), tExpr)

	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:

		// TODO: ignoring index hints for now
		stExpr, err := translateSimpleTableExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("AliasedTableExpr error: %v", err)
		}

		return []interface{}{stExpr, string(expr.As)}, nil

	case *sqlparser.ParenTableExpr:
		ptExpr, err := translateTableExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("ParenTableExpr error: %v", err)
		}
		return ptExpr, nil

	case *sqlparser.JoinTableExpr:

		left, right, err := translateLRTableExpr(expr.LeftExpr, expr.RightExpr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("JoinTableExpr LR error: %v", err)
		}

		criterion, err := translateExpr(expr.On, pCtx)
		if err != nil {
			return nil, fmt.Errorf("JoinTableExpr On error: %v", err)
		}

		return []interface{}{left, criterion, right}, nil

	default:
		return nil, fmt.Errorf("can't handle table expression type %T", expr)
	}

}

// translateSimpleTableExpr takes a simple table expression and returns its translated form.
func translateSimpleTableExpr(stExpr sqlparser.SimpleTableExpr, pCtx *ParseCtx) (interface{}, error) {

	log.Logf(log.DebugLow, "simple table expr: %s (type is %T)", sqlparser.String(stExpr), stExpr)

	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		// TODO: ignoring qualifier for now
		return sqlparser.String(expr), nil

	case *sqlparser.Subquery:
		return translateExpr(expr, pCtx)

	default:
		return nil, fmt.Errorf("can't handle simple table expression type %T", expr)
	}

}

// translateLRExpr takes two leaf expressions and returns their translations.
func translateLRExpr(lExpr, rExpr sqlparser.Expr, pCtx *ParseCtx) (interface{}, interface{}, error) {

	left, err := translateExpr(lExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("lExpr error: %v", err)
	}

	right, err := translateExpr(rExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("rExpr error: %v", err)
	}

	return left, right, nil
}

// translateLRTableExpr takes two leaf table expressions and returns their translations.
func translateLRTableExpr(lExpr, rExpr sqlparser.TableExpr, pCtx *ParseCtx) (interface{}, interface{}, error) {

	left, err := translateTableExpr(lExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("lTableExpr error: %v", err)
	}

	right, err := translateTableExpr(rExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("rTableExpr error: %v", err)
	}

	return left, right, nil
}

// getNumVal takes a number value expression and returns a converted form of it.
func getNumVal(valExpr sqlparser.ValExpr) (interface{}, error) {

	switch val := valExpr.(type) {

	case sqlparser.StrVal:
		return sqlparser.String(val), nil

	case sqlparser.NumVal:
		// TODO: add other types
		f, err := strconv.ParseFloat(sqlparser.String(val), 64)
		if err != nil {
			return nil, err
		}

		return f, nil

	default:
		return nil, fmt.Errorf("not a literal type: %T", valExpr)
	}
}
