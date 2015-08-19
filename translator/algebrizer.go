package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
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

// getAlgebrizedQuery takes a parsed SQL statements and returns an algebrized form of the query.
func getAlgebrizedQuery(stmt *sqlparser.Select, pCtx *ParseCtx) error {

	ctx := &ParseCtx{Parent: pCtx}

	tableInfo, err := getTableInfo(stmt.From, pCtx)
	if err != nil {
		return err
	}

	ctx.Table = tableInfo

	// handle select expressions like as aliasing
	// e.g. select FirstName as f, LastName as l from foo;
	columnInfo, err := getColumnInfo(stmt.SelectExprs, ctx)
	if err != nil {
		return err
	}

	ctx.Column = columnInfo

	log.Logf(log.DebugLow, "ctxt: %#v", ctx)

	if stmt.Where != nil {

		log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(stmt.Where.Expr), stmt.Where.Expr)

		algebrizedStmt, err := algebrizeExpr(stmt.Where.Expr, ctx)
		if err != nil {
			return err
		}
		stmt.Where.Expr = algebrizedStmt.(sqlparser.BoolExpr)
	}

	if stmt.From != nil {
		if len(stmt.From) != 1 {
			return fmt.Errorf("JOINS not yet supported")
		}

		err := algebrizeTableExpr(stmt.From[0], ctx)
		if err != nil {
			return err
		}
	}

	if stmt.Having != nil {
		return fmt.Errorf("'HAVING' statement not yet supported")
	}

	return nil
}

func algebrizeSelectStatement(stmt sqlparser.SelectStatement, pCtx *ParseCtx) error {

	switch expr := stmt.(type) {

	case *sqlparser.Select:
		err := getAlgebrizedQuery(expr, pCtx)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("can't handle expression type %T", expr)
	}
	return nil
}

// algebrizeExpr takes an expression and returns its algebrized form.
func algebrizeExpr(gExpr sqlparser.Expr, pCtx *ParseCtx) (sqlparser.Expr, error) {
	log.Logf(log.DebugLow, "expr: %s (type is %T)", sqlparser.String(gExpr), gExpr)

	switch expr := gExpr.(type) {

	case sqlparser.NumVal:
		return expr, nil

	case sqlparser.ValTuple:
		vals := sqlparser.ValExprs(expr)
		tuple := sqlparser.ValTuple{}

		for i, val := range vals {
			agb, err := algebrizeExpr(val, pCtx)
			if err != nil {
				return nil, fmt.Errorf("can't handle ValExpr (%v) %v: %v", i, val, err)
			}
			tuple = append(tuple, agb.(sqlparser.ValExpr))
		}
		return tuple, nil

	case *sqlparser.NullVal:
		return nil, nil

		// TODO: regex lowercased
	case *sqlparser.ColName:
		expr = &sqlparser.ColName{
			Name:      []byte(pCtx.ColumnName(sqlparser.String(expr))),
			Qualifier: expr.Qualifier,
		}
		return expr, nil

	case sqlparser.StrVal:
		return expr, nil

	case *sqlparser.BinaryExpr:
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("BinaryExpr LR error: %v", err)
		}

		expr.Left = left.(sqlparser.Expr)
		expr.Right = right.(sqlparser.Expr)
		return expr, nil

	case *sqlparser.AndExpr:
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("AndExpr error: %v", err)
		}
		expr.Left = left.(sqlparser.BoolExpr)
		expr.Right = right.(sqlparser.BoolExpr)
		return expr, nil

	case *sqlparser.OrExpr:
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("OrExpr error: %v", err)
		}
		expr.Left = left.(sqlparser.BoolExpr)
		expr.Right = right.(sqlparser.BoolExpr)
		return expr, nil

	case *sqlparser.ComparisonExpr:
		left, right, err := algebrizeLRExpr(expr.Left, expr.Right, pCtx)
		if err != nil {
			return nil, fmt.Errorf("ComparisonExpr error: %v", err)
		}

		expr.Left = left.(sqlparser.ValExpr)
		expr.Right = right.(sqlparser.ValExpr)
		return expr, nil

	case *sqlparser.RangeCond:
		from, to, err := algebrizeLRExpr(expr.From, expr.To, pCtx)
		if err != nil {
			return nil, fmt.Errorf("RangeCond LR error: %v", err)
		}

		left, err := algebrizeExpr(expr.Left, pCtx)
		if err != nil {
			return nil, fmt.Errorf("RangeCond key error: %v", err)
		}

		expr.Left = left.(sqlparser.ValExpr)
		expr.To = to.(sqlparser.ValExpr)
		expr.From = from.(sqlparser.ValExpr)
		return expr, nil

	case *sqlparser.NullCheck:
		// TODO: how is 'null' interpreted? exists? 'null'?
		val, err := algebrizeExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("NullCheck error: %v", err)
		}
		expr.Expr = val.(sqlparser.ValExpr)
		return expr, nil

	case *sqlparser.UnaryExpr:
		val, err := algebrizeExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("UnaryExpr error: %v", err)
		}
		expr.Expr = val
		return expr, nil

	case *sqlparser.NotExpr:
		val, err := algebrizeExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("NotExpr error: %v", err)
		}

		expr.Expr = val.(sqlparser.BoolExpr)
		return expr, nil

	case *sqlparser.ParenBoolExpr:
		val, err := algebrizeExpr(expr.Expr, pCtx)
		if err != nil {
			return nil, fmt.Errorf("ParenBoolExpr error: %v", err)
		}

		expr.Expr = val.(sqlparser.BoolExpr)
		return expr, nil

		//
		//  some nodes rely on SimpleSelect support
		//

	case *sqlparser.Subquery:
		err := algebrizeSelectStatement(expr.Select, pCtx)
		if err != nil {
			return nil, fmt.Errorf("Subquery error: %v", err)
		}
		return expr, nil

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
func algebrizeTableExpr(tExpr sqlparser.TableExpr, pCtx *ParseCtx) error {

	log.Logf(log.DebugLow, "table expr: %s (type is %T)", sqlparser.String(tExpr), tExpr)

	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:

		// TODO: ignoring index hints for now
		ste, err := algebrizeSimpleTableExpr(expr.Expr, pCtx)
		if err != nil {
			return fmt.Errorf("AliasedTableExpr error: %v", err)
		}
		expr.Expr = ste

	case *sqlparser.ParenTableExpr:
		err := algebrizeTableExpr(expr.Expr, pCtx)
		if err != nil {
			return fmt.Errorf("ParenTableExpr error: %v", err)
		}
		return nil

	case *sqlparser.JoinTableExpr:

		left, right, err := algebrizeLRTableExpr(expr.LeftExpr, expr.RightExpr, pCtx)
		if err != nil {
			return fmt.Errorf("JoinTableExpr LR error: %v", err)
		}

		expr.LeftExpr = left.(sqlparser.TableExpr)
		expr.RightExpr = right.(sqlparser.TableExpr)

		criterion, err := algebrizeExpr(expr.On, pCtx)
		if err != nil {
			return fmt.Errorf("JoinTableExpr On error: %v", err)
		}

		expr.On = criterion.(sqlparser.BoolExpr)

		return nil

	default:
		return fmt.Errorf("can't handle table expression type %T", expr)
	}

	return nil

}

// algebrizeSimpleTableExpr takes a simple table expression and returns its algebrized nodes.
func algebrizeSimpleTableExpr(stExpr sqlparser.SimpleTableExpr, pCtx *ParseCtx) (sqlparser.SimpleTableExpr, error) {

	log.Logf(log.DebugLow, "simple table expr: %s (type is %T)", sqlparser.String(stExpr), stExpr)

	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		// TODO: ignoring qualifier for now
		expr.Name = []byte(pCtx.TableName(sqlparser.String(expr)))
		return expr, nil

	case *sqlparser.Subquery:
		err := algebrizeSelectStatement(expr.Select, pCtx)
		if err != nil {
			return nil, fmt.Errorf("can not algebrize subquery: %v", err)
		}
		return expr, nil

	default:
		return nil, fmt.Errorf("can't handle simple table expression type %T", expr)
	}

}

// algebrizeLRExpr takes two leaf expressions and returns them algebrized.
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
	err := algebrizeTableExpr(lExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("lTableExpr error: %v", err)
	}

	err = algebrizeTableExpr(rExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("rTableExpr error: %v", err)
	}

	return lExpr, rExpr, nil
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

			switch expr2 := stExpr.(type) {

			case *sqlparser.TableName:
				strVal := strings.TrimSpace(string(expr2.Name))
				name := string(expr.As)
				if name == "" {
					name = strVal
				}
				table := TableInfo{
					Name: map[string]string{name: strVal},
				}
				tables = append(tables, table)

			default:
				return nil, fmt.Errorf("unsupported simple table expression alias of %T", expr)
			}

		default:
			return nil, fmt.Errorf("can't handle table expression type %T", expr)
		}
	}
	return tables, nil
}

// parseColumnInfo returns the column information given a parse context, the name
// of the field, and an optional alias.
func parseColumnInfo(pCtx *ParseCtx, name string, varAlias ...string) (columnInfo ColumnInfo) {
	var alias string

	if len(varAlias) == 0 {
		alias = name
	} else {
		alias = varAlias[0]
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

			case *sqlparser.ColName:
				column := parseColumnInfo(pCtx, string(name.Name), alias)
				columns = append(columns, column)

			case sqlparser.StrVal:
				column := parseColumnInfo(pCtx, string(name), alias)
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
