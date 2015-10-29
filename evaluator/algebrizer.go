package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
)

// AlgebrizeStatement takes a parsed SQL statement and returns an algebrized form of the query.
func AlgebrizeStatement(ss sqlparser.SelectStatement, pCtx *ParseCtx) error {
	log.Logf(log.DebugLow, "pCtx AlgebrizeStatement: %v\n\n", pCtx.String())

	switch stmt := ss.(type) {

	case *sqlparser.Select:

		// algebrize 'FROM' clause
		if stmt.From != nil {
			for _, table := range stmt.From {
				pCtx.Expr = stmt.From
				err := algebrizeTableExpr(table, pCtx)
				if err != nil {
					return err
				}
			}
		}

		// algebrize 'WHERE' clause
		if stmt.Where != nil {
			pCtx.Expr = stmt.Where
			algebrizedStmt, err := algebrizeExpr(stmt.Where.Expr, pCtx)
			if err != nil {
				return err
			}

			stmt.Where.Expr = algebrizedStmt.(sqlparser.BoolExpr)
		}

		// algebrize 'SELECT EXPRESSION' clause
		pCtx.Expr = stmt.SelectExprs
		algebrizedSelectExprs, err := algebrizeSelectExprs(stmt.SelectExprs, pCtx)
		if err != nil {
			return err
		}

		stmt.SelectExprs = algebrizedSelectExprs

		// algebrize 'GROUP BY' clause
		if len(stmt.GroupBy) != 0 {
			var algebrizedValExprs sqlparser.ValExprs

			pCtx.Expr = stmt.GroupBy
			for _, valExpr := range stmt.GroupBy {
				algebrizedValExpr, err := algebrizeExpr(valExpr, pCtx)
				if err != nil {
					return err
				}
				algebrizedValExprs = append(algebrizedValExprs, algebrizedValExpr.(sqlparser.ValExpr))
			}

			stmt.GroupBy = []sqlparser.ValExpr(algebrizedValExprs)
		}

		// algebrize 'HAVING' clause
		if stmt.Having != nil {
			pCtx.Expr = stmt.Having
			algebrizedStmt, err := algebrizeExpr(stmt.Having.Expr, pCtx)
			if err != nil {
				return err
			}

			stmt.Having.Expr = algebrizedStmt.(sqlparser.BoolExpr)
		}

		// algebrize 'ORDER BY' clause
		if len(stmt.OrderBy) != 0 {
			for _, orderBy := range stmt.OrderBy {
				algebrizedStmt, err := algebrizeExpr(orderBy.Expr, pCtx)
				if err != nil {
					return err
				}

				orderBy.Expr = algebrizedStmt.(sqlparser.ValExpr)
			}
		}

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
		log.Logf(log.DebugLow, "sExpr: %s (type is %T)\npCtx: %v\n\n", sqlparser.String(sExpr), sExpr, pCtx.String())

		switch expr := sExpr.(type) {

		// TODO: validate no mixture of star and non-star expression
		case *sqlparser.StarExpr:
			// validate table name if present
			if string(expr.TableName) != "" {
				_, err := pCtx.TableInfo(pCtx.Database, string(expr.TableName))
				if err != nil {
					return nil, err
				}
			}
			algebrizeSelectExprs = append(algebrizeSelectExprs, expr)

		case *sqlparser.NonStarExpr:
			nonStarExpr := expr
			if pCtx.NonStarAlias == "" {
				pCtx.NonStarAlias = string(expr.As)
			}

			pCtx.RefColumn = false
			nse, err := algebrizeExpr(expr.Expr, pCtx)
			if err != nil {
				return nil, err
			}

			// if this expression doesn't reference a column, add it
			// to the parse context provided we're not parsing a
			// subquery or a select expression within a function
			_, inFuncExpr := pCtx.Expr.(*sqlparser.FuncExpr)
			_, isSubquery := nonStarExpr.Expr.(*sqlparser.Subquery)
			if !pCtx.RefColumn && !inFuncExpr && !isSubquery {
				nonStarAlias := string(nonStarExpr.As)
				nonStarName := sqlparser.String(nonStarExpr.Expr)

				if nonStarAlias == "" {
					nonStarAlias = nonStarName
				}

				column := ColumnInfo{nonStarName, nonStarAlias, ""}
				pCtx.Columns = append(pCtx.Columns, column)
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
	log.Logf(log.DebugLow, "expr: %#v (type is %T)\npCtx: %v\n\n", gExpr, gExpr, pCtx.String())

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
		columnInfo, err := columnToCtx(pCtx, expr)
		if err != nil {
			return nil, err
		}

		if pCtx.NonStarAlias != "" {
			columnInfo.Alias = pCtx.NonStarAlias
			pCtx.NonStarAlias = ""
		}

		if !pCtx.Columns.Contains(columnInfo) {
			pCtx.Columns = append(pCtx.Columns, *columnInfo)
		}

		expr.Name = []byte(columnInfo.Name)

		// ensure all columns include a table name
		if len(expr.Qualifier) == 0 {
			expr.Qualifier = []byte(columnInfo.Table)
		}

		pCtx.RefColumn = true
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
		// set the current expression being parsed to this function to
		// prevent treating nested select expressions as top-level column
		// references
		fExpr := pCtx.Expr
		pCtx.Expr = expr
		algebrizedSelectExprs, err := algebrizeSelectExprs(expr.Exprs, pCtx)
		if err != nil {
			return nil, err
		}
		pCtx.Expr = fExpr
		expr.Exprs = algebrizedSelectExprs

		return expr, nil

	case *sqlparser.CaseExpr:

		if expr.Else != nil {
			whenElse, err := algebrizeExpr(expr.Else, pCtx)
			if err != nil {
				return nil, fmt.Errorf("CaseExpr Else error: %v", err)
			}
			expr.Else = whenElse.(sqlparser.ValExpr)
		}

		if expr.Expr != nil {
			whenExpr, err := algebrizeExpr(expr.Expr, pCtx)
			if err != nil {
				return nil, fmt.Errorf("CaseExpr Expr error: %v", err)
			}
			expr.Expr = whenExpr.(sqlparser.ValExpr)
		}

		for _, when := range expr.Whens {

			c, err := algebrizeExpr(when.Cond, pCtx)
			if err != nil {
				return nil, fmt.Errorf("CaseExpr Cond error: %v", err)
			}

			v, err := algebrizeExpr(when.Val, pCtx)
			if err != nil {
				return nil, fmt.Errorf("CaseExpr Val error: %v", err)
			}

			when.Cond = c.(sqlparser.BoolExpr)
			when.Val = v.(sqlparser.ValExpr)
		}

		return expr, nil

	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("can't handle ExistsExpr type %T", expr)

	case nil:

		return nil, nil

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
		table, err := pCtx.TableInfo(string(expr.Qualifier), string(expr.Name))
		if err != nil {
			expr.Name = []byte(string(expr.Name))
		} else {
			expr.Name = []byte(table.Name)
		}

		return expr, nil

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

			alias := string(expr.As)

			switch node := stExpr.(type) {

			case *sqlparser.TableName:
				tableName := string(node.Name)

				if alias == "" {
					alias = tableName
				}

				db := string(node.Qualifier)
				if db == "" {
					db = pCtx.Database
				}

				// handles cases where the database is expressed in the FROM clause e.g.
				// EvalSelect("", "select * from db.table", nil)
				if pCtx.Database == "" {
					pCtx.Database = db
				}

				table := NewTableInfo(alias, db, tableName, false)

				if !pCtx.Tables.Contains(table) {
					tables = append(tables, table)
				}

			case *sqlparser.Subquery:

				if alias == "" {
					return nil, fmt.Errorf("Every derived table must have its own alias")
				}

				ctx, err := NewParseCtx(node.Select, pCtx.Config, pCtx.Database)
				if err != nil {
					return nil, fmt.Errorf("GetTableInfo Subquery ctx error: %v", err)
				}

				// apply derived table alias to context
				aliasedTables := []TableInfo{}
				for _, table := range ctx.Tables {
					table.Alias = alias
					table.Name = alias
					table.Derived = true
					aliasedTables = append(aliasedTables, table)
				}

				ctx.Tables = aliasedTables

				tables = append(tables, ctx.Tables...)

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

	// TODO: the current implementation assumes all queries get routed to a single
	// database and needs to be updated.
	//
	// If only one table exists within the context use the referenced database
	if len(tables) == 1 {
		pCtx.Database = tables[0].Db
	}

	return tables, nil
}

// hasStarExpr returns true if the select expression in the given statement
// is a star expression.
func hasStarExpr(ss sqlparser.SelectStatement) bool {

	switch e := ss.(type) {

	case *sqlparser.Select:

		for _, expr := range e.SelectExprs {
			switch expr.(type) {

			// TODO: validate no mixture of star and non-star expression
			case *sqlparser.StarExpr:
				return true

			}
		}

	default:
		return false
	}

	return false
}

// columnToCtx returns the column information given a parse context, the name
// of the field, and an optional alias.
func columnToCtx(pCtx *ParseCtx, expr *sqlparser.ColName) (*ColumnInfo, error) {

	columnInfo := &ColumnInfo{}

	columnName := string(expr.Name)
	tableName := string(expr.Qualifier)

	// TODO: check if column is star expression
	if columnName == "" {
		return nil, fmt.Errorf("column name can not be empty: %v", columnName)
	}

	table, err := pCtx.GetCurrentTable(pCtx.Database, tableName)
	if err != nil {
		return nil, err
	}

	if pCtx.Database == "" {
		pCtx.Database = table.Db
	}

	columnInfo.Table = table.Name

	// the column name itself could be an alias so handle this
	info, err := pCtx.ColumnInfo(columnName)

	if err != nil {
		// this is not an alias so it must be an actual column name
		// or a column name from a child context
		err := pCtx.CheckColumn(table.Name, columnName)
		if err != nil {
			return nil, err
		}
		columnInfo.Name = columnName
	} else {
		columnInfo.Name = info.Name
		columnInfo.Alias = info.Alias
	}

	// if the column isn't aliased, use the actual column name
	// as the alias
	if columnInfo.Alias == "" {
		columnInfo.Alias = columnInfo.Name
	}

	return columnInfo, nil
}
