package algebrizer

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"strings"
)

// AlgebrizeStatement takes a parsed SQL statement and returns an algebrized form of the query.
func AlgebrizeStatement(ss sqlparser.SelectStatement, ctx *ParseCtx) error {
	log.Logf(log.DebugLow, "pCtx AlgebrizeStatement: %#v\n\n", ctx)

	switch stmt := ss.(type) {

	case *sqlparser.Select:

		// algebrize 'FROM' clause
		if stmt.From != nil {
			for _, table := range stmt.From {
				err := algebrizeTableExpr(table, ctx)
				if err != nil {
					return err
				}
			}
		}

		// algebrize 'SELECT EXPRESSION' clause
		algebrizedSelectExprs, err := algebrizeSelectExprs(stmt.SelectExprs, ctx)
		if err != nil {
			return err
		}

		stmt.SelectExprs = algebrizedSelectExprs

		// algebrize 'WHERE' clause
		if stmt.Where != nil {
			algebrizedStmt, err := algebrizeExpr(stmt.Where.Expr, ctx)
			if err != nil {
				return err
			}

			stmt.Where.Expr = algebrizedStmt.(sqlparser.BoolExpr)
		}

		// algebrize 'GROUP BY' clause
		if len(stmt.GroupBy) != 0 {
			var algebrizedValExprs sqlparser.ValExprs

			for _, valExpr := range stmt.GroupBy {
				algebrizedValExpr, err := algebrizeExpr(valExpr, ctx)
				if err != nil {
					return err
				}
				algebrizedValExprs = append(algebrizedValExprs, algebrizedValExpr.(sqlparser.ValExpr))
			}

			stmt.GroupBy = []sqlparser.ValExpr(algebrizedValExprs)
		}

		log.Logf(log.DebugLow, "crazy ass end parse context: %#v\n\n", ctx)

		// algebrize group by -> having -> select
		// expressions -> into -> order by -> limit

	default:
		return fmt.Errorf("select statement type %T not yet supported", stmt)

	}

	return nil
}

func algebrizeSelectStatement(stmt sqlparser.SelectStatement, pCtx *ParseCtx) error {

	switch expr := stmt.(type) {

	case *sqlparser.Select:
		err := AlgebrizeStatement(expr, pCtx)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("can't handle expression type %T", expr)
	}
	return nil
}

// algebrizeSelectExprs takes a slice select expressions and returns its algebrized form.
func algebrizeSelectExprs(sExprs sqlparser.SelectExprs, pCtx *ParseCtx) (sqlparser.SelectExprs, error) {
	var algebrizeSelectExprs sqlparser.SelectExprs

	for _, sExpr := range sExprs {
		log.Logf(log.DebugLow, "sExpr: %s (type is %T)\npCtx: %#v\n\n", sqlparser.String(sExpr), sExpr, pCtx)

		switch expr := sExpr.(type) {

		// TODO: validate no mixture of star and non-star expression
		case *sqlparser.StarExpr:
			algebrizeSelectExprs = append(algebrizeSelectExprs, expr)

		case *sqlparser.NonStarExpr:
			nse, err := algebrizeExpr(expr.Expr, pCtx)
			if err != nil {
				return nil, err
			}
			algNSE := &sqlparser.NonStarExpr{nse, expr.As}
			algebrizeSelectExprs = append(algebrizeSelectExprs, algNSE)

		default:
			return nil, fmt.Errorf("unreachable path")
		}
	}
	return algebrizeSelectExprs, nil
}

// algebrizeExpr takes an expression and returns its algebrized form.
func algebrizeExpr(gExpr sqlparser.Expr, pCtx *ParseCtx) (sqlparser.Expr, error) {
	log.Logf(log.DebugLow, "expr: %#v (type is %T)\npCtx: %#v\n\n", gExpr, gExpr, pCtx)

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
		ns := string(expr.Name)
		qualifier := string(expr.Qualifier)
		if qualifier != "" {
			ns = fmt.Sprintf("%v.%v", qualifier, ns)
		}
		columnInfo, err := columnToCtx(pCtx, ns)
		if err != nil {
			return nil, fmt.Errorf("ColName error on %#v: %v %#v", ns, err, pCtx)
		}
		pCtx.Column = append(pCtx.Column, *columnInfo)

		expr.Name = []byte(columnInfo.Field)
		expr.Qualifier = []byte(columnInfo.Collection)
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
		nCtx, err := pCtx.ChildCtx(expr.Select)
		if err != nil {
			return nil, fmt.Errorf("error constructing Subquery parse context: %v", err)
		}

		err = algebrizeSelectStatement(expr.Select, nCtx)
		if err != nil {
			return nil, fmt.Errorf("Subquery error: %v", err)
		}
		return expr, nil

	case sqlparser.ValArg:
		return nil, fmt.Errorf("can't handle ValArg type %T", expr)

	case *sqlparser.FuncExpr:

		algebrizedSelectExprs, err := algebrizeSelectExprs(expr.Exprs, pCtx)
		if err != nil {
			return nil, err
		}
		expr.Exprs = algebrizedSelectExprs

		return expr, nil

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

		if expr.On != nil {
			criterion, err := algebrizeExpr(expr.On, pCtx)
			if err != nil {
				return fmt.Errorf("JoinTableExpr On error: %v", err)
			}
			expr.On = criterion.(sqlparser.BoolExpr)
		}

		return nil

	default:
		return fmt.Errorf("can't handle table expression type %T", expr)
	}

	return nil

}

// algebrizeSimpleTableExpr takes a simple table expression and returns its algebrized nodes.
func algebrizeSimpleTableExpr(stExpr sqlparser.SimpleTableExpr, pCtx *ParseCtx) (sqlparser.SimpleTableExpr, error) {

	log.Logf(log.DebugLow, "simple table expr: %s (type is %T)\npCtx: %#v\n\n", sqlparser.String(stExpr), stExpr, pCtx)

	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		tName, err := pCtx.TableName(string(expr.Qualifier), string(expr.Name))
		expr.Name = []byte(tName)

		return expr, err

	case *sqlparser.Subquery:
		nCtx, err := pCtx.ChildCtx(expr.Select)
		if err != nil {
			return nil, fmt.Errorf("error constructing new parse context: %v", err)
		}

		err = algebrizeSelectStatement(expr.Select, nCtx)
		if err != nil {
			return nil, fmt.Errorf("can't algebrize Subquery: %v", err)
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

// GetTableInfo takes a select expression and returns the table information.
func GetTableInfo(tExprs sqlparser.TableExprs, pCtx *ParseCtx) ([]TableInfo, error) {
	tables := []TableInfo{}

	for _, tExpr := range tExprs {

		switch expr := tExpr.(type) {

		case *sqlparser.AliasedTableExpr:
			stExpr, err := algebrizeSimpleTableExpr(expr.Expr, pCtx)
			if err != nil {
				return nil, fmt.Errorf("GetTableInfo AliasedTableExpr error: %v", err)
			}

			switch node := stExpr.(type) {

			case *sqlparser.TableName:
				tableName := strings.TrimSpace(string(node.Name))
				alias := string(expr.As)
				if alias == "" {
					alias = tableName
				}
				tables = append(tables, NewTableInfo(alias, tableName))
			case *sqlparser.Subquery:

				ctx, err := NewParseCtx(node.Select)
				if err != nil {
					return nil, fmt.Errorf("GetTableInfo Subquery ctx error: %v", err)
				}

				tables = append(tables, ctx.Table...)

			default:
				return nil, fmt.Errorf("GetTableInfo AliasedTableExpr type assert error: %v", err)
			}

		case *sqlparser.ParenTableExpr:
			tableInfo, err := GetTableInfo([]sqlparser.TableExpr{expr.Expr}, pCtx)
			if err != nil {
				return nil, fmt.Errorf("GetTableInfo ParenTableExpr error: %v", err)
			}
			tables = append(tables, tableInfo...)

		case *sqlparser.JoinTableExpr:
			var l sqlparser.TableExprs
			l = append(l, expr.LeftExpr)
			lInfo, err := GetTableInfo(l, pCtx)
			if err != nil {
				return nil, fmt.Errorf("JoinTableExpr LeftExpr error: %v", err)
			}

			var r sqlparser.TableExprs
			r = append(r, expr.RightExpr)
			rInfo, err := GetTableInfo(r, pCtx)
			if err != nil {
				return nil, fmt.Errorf("JoinTableExpr RightExpr error: %v", err)
			}

			tables = append(tables, lInfo...)
			tables = append(tables, rInfo...)
		default:
			return nil, fmt.Errorf("can't handle table expression type %T", expr)
		}
	}
	return tables, nil
}

// columnToCtx returns the column information given a parse context, the name
// of the field, and an optional alias.
func columnToCtx(pCtx *ParseCtx, column string, varAlias ...string) (*ColumnInfo, error) {

	columnInfo := &ColumnInfo{}

	var alias string

	if len(varAlias) != 0 && varAlias[0] != "" {
		alias = varAlias[0]
	}

	var tableName string
	var columnName string
	var err error

	// check if the column name is in the form
	// foo.baz - where foo is the table name
	// and baz is the actual column name.
	if i := strings.Index(column, "."); i != -1 {
		tableName = column[:i]
		columnName = column[i+1:]
	} else {
		columnName = column
	}

	// TODO: check if column is star expression
	if columnName == "" {
		return nil, fmt.Errorf("column name can not be empty: %v", column)
	}

	tableName, err = pCtx.GetCurrentTable(tableName)
	if err != nil {
		return nil, err
	}
	columnInfo.Collection = tableName

	// the column name itself could be an alias so handle this
	aliasInfo, err := pCtx.ColumnAlias(alias)
	if err != nil {
		columnInfo.Field = columnName
	} else {
		columnInfo.Field = aliasInfo.Field
		columnInfo.Alias = aliasInfo.Alias
	}

	// if the column isn't aliased, use the actual column name
	// as the alias
	if columnInfo.Alias == "" {
		columnInfo.Alias = columnName
	}

	return columnInfo, nil
}
