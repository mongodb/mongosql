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

			pCtx.Phase = PhaseFrom

			for _, table := range stmt.From {

				err := algebrizeTableExpr(table, pCtx)
				if err != nil {
					return err
				}

			}
		}

		// algebrize 'SELECT EXPRESSION' clause
		pCtx.Phase = PhaseSelectExpr

		algebrizedSelectExprs, err := algebrizeSelectExprs(stmt.SelectExprs, pCtx)
		if err != nil {
			return err
		}

		stmt.SelectExprs = algebrizedSelectExprs

		// algebrize 'GROUP BY' clause
		if len(stmt.GroupBy) != 0 {

			pCtx.Phase = PhaseGroupBy

			var algebrizedValExprs sqlparser.ValExprs

			for _, valExpr := range stmt.GroupBy {

				pCtx.Expr = valExpr

				algebrizedValExpr, err := algebrizeExpr(valExpr, pCtx)
				if err != nil {
					return err
				}

				algebrizedValExprs = append(algebrizedValExprs, algebrizedValExpr.(sqlparser.ValExpr))
			}

			stmt.GroupBy = []sqlparser.ValExpr(algebrizedValExprs)
		}

		// algebrize 'WHERE' clause
		if stmt.Where != nil {

			pCtx.Phase = PhaseWhere

			pCtx.Expr = stmt.Where.Expr

			algebrizedStmt, err := algebrizeExpr(stmt.Where.Expr, pCtx)
			if err != nil {
				return err
			}

			stmt.Where.Expr = algebrizedStmt.(sqlparser.BoolExpr)
		}

		// algebrize 'HAVING' clause
		if stmt.Having != nil {

			pCtx.Phase = PhaseHaving

			pCtx.Expr = stmt.Having.Expr

			algebrizedStmt, err := algebrizeExpr(stmt.Having.Expr, pCtx)
			if err != nil {
				return err
			}

			stmt.Having.Expr = algebrizedStmt.(sqlparser.BoolExpr)
		}

		// algebrize 'ORDER BY' clause
		if len(stmt.OrderBy) != 0 {

			pCtx.Phase = PhaseOrderBy

			for _, orderBy := range stmt.OrderBy {

				pCtx.Expr = orderBy.Expr

				algebrizedStmt, err := algebrizeExpr(orderBy.Expr, pCtx)
				if err != nil {
					return err
				}

				orderBy.Expr = algebrizedStmt.(sqlparser.ValExpr)
			}
		}

		// algebrize 'LIMIT' clause
		if stmt.Limit != nil {

			pCtx.Phase = PhaseLimit

			if stmt.Limit.Offset != nil {

				pCtx.Expr = stmt.Limit.Offset

				offset, err := algebrizeExpr(stmt.Limit.Offset, pCtx)
				if err != nil {
					return err
				}

				stmt.Limit.Offset = offset.(sqlparser.ValExpr)
			}

			pCtx.Expr = stmt.Limit.Rowcount

			rowcount, err := algebrizeExpr(stmt.Limit.Rowcount, pCtx)
			if err != nil {
				return err
			}

			stmt.Limit.Rowcount = rowcount.(sqlparser.ValExpr)
		}

		// TODO: into?
		// algebrize group by -> having -> select
		// expressions -> INTO -> order by -> limit

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

		log.Logf(log.DebugLow, "select expr: %v (%T)\npCtx: %v\n\n", sqlparser.String(sExpr), sExpr, pCtx.String())

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

			// This replaces star expressions with actual column references if
			// possible.
			//
			// TODO: abstract configuration source to support databases like
			// the information schema

			if pCtx.Database != InformationSchema && !pCtx.InFuncExpr() {

				table, err := pCtx.GetCurrentTable(pCtx.Database, string(expr.TableName))
				if err != nil {
					algebrizeSelectExprs = append(algebrizeSelectExprs, expr)
					continue
				}

				schema := pCtx.TableSchema(table.Name)
				if schema != nil {
					for _, column := range schema.Columns {
						expr := &sqlparser.ColName{
							Name:      []byte(column.Name),
							Qualifier: []byte(table.Alias),
						}

						nonStarExpr := &sqlparser.NonStarExpr{
							Expr: expr,
						}
						algebrizeSelectExprs = append(algebrizeSelectExprs, nonStarExpr)
					}
				} else {
					if !table.Derived {
						return nil, fmt.Errorf("non-derived table '%v' does not exist", table.Name)
					}

					algebrizeSelectExprs = append(algebrizeSelectExprs, pCtx.GetTableColumns(table)...)
				}
				continue
			} else {
				algebrizeSelectExprs = append(algebrizeSelectExprs, expr)
			}

		case *sqlparser.NonStarExpr:

			nonStarExpr := expr

			pCtx.Expr = expr.Expr

			if pCtx.NonStarAlias == "" {
				pCtx.NonStarAlias = string(expr.As)
			}

			pCtx.State &^= StateRefColExpr

			nse, err := algebrizeExpr(expr.Expr, pCtx)
			if err != nil {
				return nil, err
			}

			// If this expression doesn't reference a column, add the
			// column information to the parse context.
			if !(pCtx.InRefColumn() || pCtx.InFuncExpr()) {

				nonStarAlias := string(nonStarExpr.As)
				nonStarName := sqlparser.String(nonStarExpr.Expr)

				if nonStarAlias == "" {
					nonStarAlias = nonStarName
				}

				var tableName string

				table, err := pCtx.GetCurrentTable(pCtx.Database, "")
				if err == nil {
					tableName = table.Alias
				}

				index := len(pCtx.Columns) + len(pCtx.ColumnReferences)

				column := ColumnInfo{nonStarName, nonStarAlias, tableName, index}

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
	log.Logf(log.DebugLow, "expr: %v (%T)\npCtx: %v\n\n", sqlparser.String(gExpr), gExpr, pCtx.String())

	switch expr := gExpr.(type) {

	case sqlparser.NumVal:

		return expr, nil

	case sqlparser.ValTuple:

		vals := sqlparser.ValExprs(expr)
		tuple := sqlparser.ValTuple{}

		for i, val := range vals {

			t, err := algebrizeExpr(val, pCtx)
			if err != nil {
				return nil, fmt.Errorf("can't handle ValExpr %v (%v): %v", i+1, sqlparser.String(val), err)
			}

			tuple = append(tuple, t.(sqlparser.ValExpr))

		}

		return tuple, nil

	case *sqlparser.NullVal:

		return expr, nil

		// TODO: regex lowercased
	case *sqlparser.ColName:

		return resolveColumnExpr(expr, pCtx)

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

		// algebrize subquery as a select expression

		nCtx, err := pCtx.ChildCtx(expr.Select)
		if err != nil {
			return nil, fmt.Errorf("error constructing subquery select expression context: %v", err)
		}

		nCtx.State |= StateSubQueryExpr

		err = algebrizeSelectStatement(expr.Select, nCtx)
		if err != nil {
			return nil, fmt.Errorf("can't algebrize select expression Subquery: %v", err)
		}

		return expr, nil

	case *sqlparser.FuncExpr:
		// set the current expression being parsed to this function to
		// prevent treating nested select expressions as top-level column
		// references

		if pCtx.NonStarAlias != "" {
			index := len(pCtx.Columns) + len(pCtx.ColumnReferences)
			ref := ColumnReference{pCtx.NonStarAlias, pCtx.DerivedTableName, pCtx.Expr, index}
			pCtx.ColumnReferences = append(pCtx.ColumnReferences, ref)
		}

		pCtx.State |= StateFuncExpr

		algebrizedSelectExprs, err := algebrizeSelectExprs(expr.Exprs, pCtx)
		if err != nil {
			return nil, err
		}

		pCtx.State &^= StateFuncExpr

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

		return algebrizeExpr(expr.Subquery, pCtx)

	case nil:

		return &sqlparser.NullVal{}, nil

	case sqlparser.ValArg:

		return nil, fmt.Errorf("can't handle ValArg type %T", expr)

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

		return nil

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

}

// algebrizeSimpleTableExpr takes a simple table expression and returns its algebrized nodes.
func algebrizeSimpleTableExpr(stExpr sqlparser.SimpleTableExpr, pCtx *ParseCtx) (sqlparser.SimpleTableExpr, error) {

	log.Logf(log.DebugLow, "simple table expr: %s (type is %T)\npCtx: %v\n\n", sqlparser.String(stExpr), stExpr, pCtx.String())

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

		// only perform data source algebrization when we're in the INIT phase
		if pCtx.Phase == PhaseInit {

			nCtx, err := pCtx.ChildCtx(expr.Select)
			if err != nil {
				return nil, fmt.Errorf("error constructing subquery source context: %v", err)
			}

			nCtx.State |= StateSubQueryExpr

			nCtx.DerivedTableName = pCtx.DerivedTableName

			if err = algebrizeSelectStatement(expr.Select, nCtx); err != nil {
				return nil, fmt.Errorf("can't algebrize Subquery: %v", err)
			}

		}

		return expr, nil

	default:

		return nil, fmt.Errorf("can't handle simple table expression type %T", expr)

	}

}

// algebrizeLRExpr takes two leaf expressions and returns them algebrized.
func algebrizeLRExpr(leftExpr, rightExpr sqlparser.Expr, pCtx *ParseCtx) (interface{}, interface{}, error) {

	left, err := algebrizeExpr(leftExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("leftExpr error: %v", err)
	}

	right, err := algebrizeExpr(rightExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("rightExpr error: %v", err)
	}

	return left, right, nil
}

// algebrizeLRTableExpr takes two leaf table expressions and returns their translations.
func algebrizeLRTableExpr(leftExpr, rightExpr sqlparser.TableExpr, pCtx *ParseCtx) (interface{}, interface{}, error) {

	err := algebrizeTableExpr(leftExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("lTableExpr error: %v", err)
	}

	err = algebrizeTableExpr(rightExpr, pCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("rTableExpr error: %v", err)
	}

	return leftExpr, rightExpr, nil
}

// GetTableInfo takes a select expression and returns the table information.
func GetTableInfo(tExprs sqlparser.TableExprs, pCtx *ParseCtx) ([]TableInfo, error) {

	log.Logf(log.DebugLow, "get table info: %s (type is %T)\npCtx: %v\n\n", sqlparser.String(tExprs), tExprs, pCtx.String())

	tables := []TableInfo{}

	for _, tExpr := range tExprs {

		switch expr := tExpr.(type) {

		case *sqlparser.AliasedTableExpr:

			alias := string(expr.As)

			pCtx.DerivedTableName = alias

			stExpr, err := algebrizeSimpleTableExpr(expr.Expr, pCtx)
			if err != nil {
				return nil, fmt.Errorf("GetTableInfo AliasedTableExpr error: %v", err)
			}

			pCtx.DerivedTableName = ""

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

				tables = append(tables, NewTableInfo(alias, db, tableName, false))

			case *sqlparser.Subquery:

				if alias == "" {
					return nil, fmt.Errorf("Every derived table must have its own alias")
				}

				ctx, err := NewParseCtx(node.Select, pCtx.Config, pCtx.Database)
				if err != nil {
					return nil, fmt.Errorf("GetTableInfo Subquery ctx error: %v", err)
				}

				// apply derived table alias to context
				for _, table := range ctx.Tables {
					table.Alias = alias
					table.Name = alias
					table.Derived = true
					tables = append(tables, table)
				}

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

// resolveColumnExpr takes a column name expression and resolves it within
// the given parse context.
func resolveColumnExpr(expr *sqlparser.ColName, pCtx *ParseCtx) (sqlparser.Expr, error) {

	defer func() {
		// reset the alias each time we resolve a column name and
		// indicate that a column expression was just algebrized
		pCtx.NonStarAlias = ""
		pCtx.State |= StateRefColExpr

	}()

	columnInfo, err := columnToCtx(pCtx, expr)
	if err != nil {
		return nil, err
	}

	if pCtx.NonStarAlias != "" {
		columnInfo.Alias = pCtx.NonStarAlias
	}

	// ensure all columns include a table name
	if len(expr.Qualifier) == 0 {
		expr.Qualifier = []byte(columnInfo.Table)
	}

	if pCtx.DerivedTableName != "" {
		columnInfo.Table = pCtx.DerivedTableName
	}

	// If we're not parsing a select expression, and we encounter a column
	// it could either be a schema table column or the alias referencing a
	// select expression
	if pCtx.Phase != PhaseSelectExpr {

		column := &Column{
			Table: string(expr.Qualifier),
			Name:  string(expr.Name),
			View:  string(expr.Name),
		}

		// If it's not a schema table column or a aliased
		// column reference, it could be an aliased column name.
		//
		// For example, the GROUP BY clause in:
		//
		// select a as x, sum (b) from foo group by x;
		//
		if !pCtx.IsSchemaColumn(column) && !pCtx.IsColumnReference(column) {
			column, err := pCtx.ColumnInfo(column.View)
			if err != nil {
				return pCtx.Expr, nil
			}

			expr := &sqlparser.ColName{
				Name:      []byte(column.Name),
				Qualifier: []byte(column.Table),
			}

			return expr, nil
		}

		// If the expression represents a table column
		// that is aliased, simply return the aliased
		// column expression.
		//
		// This will transform queries like:
		//
		// select a+b as f from foo order by f;
		//
		// to
		//
		// select a+b as f from foo order by a+b;
		//
		for _, ref := range pCtx.ColumnReferences {

			if ref.Name == column.Name {
				return ref.Expr, nil
			}
		}
	}

	// When parsing a select expression, the column expression
	// could either be of type *sqlparser.ColName or just
	// aliased as such. Add the expression to the parse context
	// either as a column or as a column reference.

	if !pCtx.InFuncExpr() {
		if _, ok := pCtx.Expr.(*sqlparser.ColName); ok {
			index := len(pCtx.Columns) + len(pCtx.ColumnReferences)
			columnInfo.Index = index
			pCtx.Columns = append(pCtx.Columns, *columnInfo)
		} else {
			if pCtx.NonStarAlias != "" {
				index := len(pCtx.Columns) + len(pCtx.ColumnReferences)
				ref := ColumnReference{columnInfo.Alias, columnInfo.Table, pCtx.Expr, index}
				pCtx.ColumnReferences = append(pCtx.ColumnReferences, ref)
			}
		}
	}

	expr.Name = []byte(columnInfo.Name)

	// ensure all columns include a table name
	if len(expr.Qualifier) == 0 {
		expr.Qualifier = []byte(columnInfo.Table)
	}

	return expr, nil
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

	columnInfo.Table = table.Alias

	// the column name itself could be an alias so handle this
	info, err := pCtx.ColumnInfo(columnName)

	if err != nil || table.Derived {
		// The column name given is not an alias so
		// it is either an actual column name or a
		// column name from a child context

		if err := pCtx.CheckColumn(table, columnName); err != nil {
			return nil, err
		}

		columnInfo.Name = columnName
	} else {
		columnInfo.Name = info.Name
		columnInfo.Alias = info.Alias
	}

	// If the column isn't aliased, use the actual
	// column name as the alias
	if columnInfo.Alias == "" {
		columnInfo.Alias = columnInfo.Name
	}

	return columnInfo, nil
}
