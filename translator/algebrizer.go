package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/util"
	"github.com/siddontang/mixer/sqlparser"
	"strconv"
	"strings"
)

// AlgebrizedQuery holds a name resolved form of a select query.
type AlgebrizedQuery struct {
	Collection interface{}
	Filter     interface{}
	Projection string
}

func algebrizeSelectStatement(stmt sqlparser.SelectStatement, pCtx *ParseCtx) (interface{}, error) {

	switch expr := stmt.(type) {

	case *sqlparser.Select:
		return getAlgebrizedQuery(expr, pCtx)

	default:
		return nil, fmt.Errorf("can't handle expression type %T", expr)
	}
	return nil, nil
}

// algebrizeExpr takes an expression and returns its algebrized form.
func algebrizeExpr(gExpr sqlparser.Expr, pCtx *ParseCtx) (interface{}, error) {
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
			t, err := algebrizeExpr(val, pCtx)
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
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
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
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("AndExpr error: %v", err)
		}
		return AndExpr{left, right}, nil

	case *sqlparser.OrExpr:
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("OrExpr error: %v", err)
		}
		return OrExpr{left, right}, nil

	case *sqlparser.ComparisonExpr:
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
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
		from, to, err := algebrizeLRExpr(expr.From, expr.To, pCtx)
		if err != nil {
			return nil, fmt.Errorf("RangeCond LR error: %v", err)
		}

		left, err := algebrizeExpr(expr.Left, pCtx)
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
		val, err := algebrizeExpr(expr.Expr, pCtx)
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
		val, err := algebrizeExpr(expr.Expr, pCtx)
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
		val, err := algebrizeExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("NotExpr error: %v", err)
		}

		return NotExpr{val}, err

	case *sqlparser.ParenBoolExpr:
		val, err := algebrizeExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("ParenBoolExpr error: %v", err)
		}

		return val, err

		//
		//  some nodes rely on SimpleSelect support
		//

	case *sqlparser.Subquery:
		val, err := algebrizeSelectStatement(expr.Select, pCtx)
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

// algebrizeTableExpr takes a table expression and returns its algebrized form.
func algebrizeTableExpr(tExpr sqlparser.TableExpr, pCtx *ParseCtx) (interface{}, error) {

	log.Logf(log.DebugLow, "table expr: %s (type is %T)", sqlparser.String(tExpr), tExpr)

	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:

		// TODO: ignoring index hints for now
		stExpr, err := algebrizeSimpleTableExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("AliasedTableExpr error: %v", err)
		}

		return []interface{}{stExpr, string(expr.As)}, nil

	case *sqlparser.ParenTableExpr:
		ptExpr, err := algebrizeTableExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("ParenTableExpr error: %v", err)
		}
		return ptExpr, nil

	case *sqlparser.JoinTableExpr:

		left, right, err := algebrizeLRTableExpr(expr.LeftExpr, expr.RightExpr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("JoinTableExpr LR error: %v", err)
		}

		criterion, err := algebrizeExpr(expr.On, pCtx)
		if err != nil {
			return nil, fmt.Errorf("JoinTableExpr On error: %v", err)
		}

		return []interface{}{left, criterion, right}, nil

	default:
		return nil, fmt.Errorf("can't handle table expression type %T", expr)
	}

}

// algebrizeSimpleTableExpr takes a simple table expression and returns its algebrized nodes.
func algebrizeSimpleTableExpr(stExpr sqlparser.SimpleTableExpr, pCtx *ParseCtx) (interface{}, error) {

	log.Logf(log.DebugLow, "simple table expr: %s (type is %T)", sqlparser.String(stExpr), stExpr)

	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		// TODO: ignoring qualifier for now
		return sqlparser.String(expr), nil

	case *sqlparser.Subquery:
		return algebrizeExpr(expr, pCtx)

	default:
		return nil, fmt.Errorf("can't handle simple table expression type %T", expr)
	}

}

// algebrizeLRExpr takes two leaf expressions and returns their translations.
func algebrizeLRExpr(lExpr, rExpr sqlparser.Expr, pCtx *ParseCtx) (interface{}, interface{}, error) {

	left, err := algebrizeExpr(lExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("lExpr error: %v", err)
	}

	right, err := algebrizeExpr(rExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("rExpr error: %v", err)
	}

	return left, right, nil
}

// algebrizeLRTableExpr takes two leaf table expressions and returns their translations.
func algebrizeLRTableExpr(lExpr, rExpr sqlparser.TableExpr, pCtx *ParseCtx) (interface{}, interface{}, error) {

	left, err := algebrizeTableExpr(lExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("lTableExpr error: %v", err)
	}

	right, err := algebrizeTableExpr(rExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("rTableExpr error: %v", err)
	}

	return left, right, nil
}

// getTableInfo takes a select expression and returns the table information.
func getTableInfo(tExprs sqlparser.TableExprs, pCtx *ParseCtx) ([]TableInfo, error) {
	tables := []TableInfo{}

	for _, tExpr := range tExprs {

		switch expr := tExpr.(type) {

		case *sqlparser.AliasedTableExpr:
			stExpr, err := algebrizeSimpleTableExpr(expr.Expr, pCtx)
			if err != nil {
				return nil, fmt.Errorf("getTableInfo error: %v", err)
			}

			if strVal, ok := stExpr.(string); ok {
				name := strings.TrimSpace(string(expr.As))
				if name == "" {
					name = strVal
				}
				table := TableInfo{
					Name: map[string]string{name: strVal},
				}
				tables = append(tables, table)
			} else {
				return nil, fmt.Errorf("unsupported simple table expression alias of %T", expr)
			}

		default:
			return nil, fmt.Errorf("can't handle table expression type %T", expr)
		}
	}
	return tables, nil
}

func parseColumnInfo(alias, name string, pCtx *ParseCtx) (columnInfo ColumnInfo) {
	if alias == "" {
		alias = name
	}
	if i := strings.Index(name, "."); i != -1 {
		if actual := pCtx.TableName(name[:i]); actual != "" {
			columnInfo.Table = actual
		} else {
			columnInfo.Table = name[:i]
		}
		columnInfo.Name = map[string]string{alias: name[i+1:]}
	} else {
		columnInfo.Name = map[string]string{alias: name}
	}
	if columnInfo.Table == "" {
		// TODO: join with multiple tables
		columnInfo.Table = pCtx.GetDefaultTable()
	}
	return
}

// getColumnInfo takes a select expression (and table information) and returns the column information.
func getColumnInfo(exprs sqlparser.SelectExprs, pCtx *ParseCtx) ([]ColumnInfo, error) {
	columns := []ColumnInfo{}

	for i, sExpr := range exprs {
		log.Logf(log.DebugLow, "handling parsed select expr %v: %#v", i, sExpr)

		switch expr := sExpr.(type) {

		case *sqlparser.StarExpr:
			log.Logf(log.DebugLow, "got star expression, fetching all columns")

			if columns == nil {
				return nil, fmt.Errorf("received multiple star expressions in select statement")
			}
			columns = nil

		case *sqlparser.NonStarExpr:
			c, err := algebrizeExpr(expr.Expr, nil)
			if err != nil {
				return nil, err
			}

			alias := strings.TrimSpace(string(expr.As))

			switch name := c.(type) {

			case ColName:
				column := parseColumnInfo(alias, name.Value, pCtx)
				columns = append(columns, column)

			case StrVal:
				column := parseColumnInfo(alias, name.Value, pCtx)
				columns = append(columns, column)

			case string:
				column := parseColumnInfo(alias, name, pCtx)
				columns = append(columns, column)

			default:
				return nil, fmt.Errorf("unsupported column type: %T", c)
			}

		default:
			return nil, fmt.Errorf("unreachable path")
		}
	}

	return columns, nil
}

func getAlgebrizedQuery(stmt *sqlparser.Select, pCtx *ParseCtx) (*AlgebrizedQuery, error) {

	ctx := &ParseCtx{Parent: pCtx}

	tableInfo, err := getTableInfo(stmt.From, pCtx)
	if err != nil {
		return nil, err
	}

	ctx.Table = tableInfo

	// handle select expressions like as aliasing
	// e.g. select FirstName as f, LastName as l from foo;
	columnInfo, err := getColumnInfo(stmt.SelectExprs, ctx)
	if err != nil {
		return nil, err
	}

	ctx.Column = columnInfo

	log.Logf(log.DebugLow, "ctxt: %#v", ctx)

	query := &AlgebrizedQuery{}

	if stmt.Where != nil {

		log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(stmt.Where.Expr), stmt.Where.Expr)

		filter, err := algebrizeExpr(stmt.Where.Expr, ctx)
		if err != nil {
			return nil, err
		}

		query.Filter = filter
	}

	if stmt.From != nil {
		if len(stmt.From) != 1 {
			return nil, fmt.Errorf("JOINS not yet supported")
		}

		c, err := algebrizeTableExpr(stmt.From[0], ctx)
		if err != nil {
			return nil, err
		}

		query.Collection = c
	}

	if stmt.Having != nil {
		return nil, fmt.Errorf("'HAVING' statement not yet supported")
	}

	return query, nil
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
