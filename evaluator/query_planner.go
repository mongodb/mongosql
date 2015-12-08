package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"strconv"
	"strings"
)

var (
	InformationSchema = "information_schema"
)

// PlanQuery constructs a query plan to satisfy the select statement.
func PlanQuery(ctx *ExecutionCtx, ss sqlparser.SelectStatement) (Operator, error) {

	switch ast := ss.(type) {

	case *sqlparser.Select:

		log.Logf(log.DebugLow, "Planning select query for %#v\n", sqlparser.String(ss))

		return planSelectExpr(ctx, ast)

	case *sqlparser.SimpleSelect:

		log.Logf(log.DebugLow, "Planning simple select query for %#v\n", sqlparser.String(ss))

		return planSimpleSelectExpr(ctx, ast)

	case *sqlparser.Union:

		return nil, fmt.Errorf("union select statement not yet implemented")

	default:

		return nil, fmt.Errorf("unknown select statement: %T", ast)

	}

}

func getReferencedExpressions(ast *sqlparser.Select, ctx *ExecutionCtx, sExprs SelectExpressions) (SelectExpressions, error) {
	var expressions SelectExpressions

	addReferencedColumns := func(columns []*Column) {
		for _, column := range columns {
			// only add basic columns present in table configuration
			hasColumn := (expressions.Contains(*column) || sExprs.Contains(*column))
			if ctx.ParseCtx.IsSchemaColumn(column) && !hasColumn && !hasStarExpr(ast) {
				sqlExpr := &sqlparser.ColName{Name: []byte(column.Name), Qualifier: []byte(column.Table)}
				expression := SelectExpression{Column: *column, Referenced: true, Expr: sqlExpr}
				expressions = append(expressions, expression)
			}
		}
	}

	var exprs []sqlparser.Expr

	// add any referenced columns in the GROUP BY clause
	for _, key := range ast.GroupBy {
		expr, err := getGroupByTerm(sExprs, key)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}

	// add any referenced columns in the ORDER BY clause
	for _, key := range ast.OrderBy {
		expr, err := getOrderByTerm(sExprs, key.Expr)
		if err != nil {
			return nil, err
		}

		exprs = append(exprs, expr)
	}

	// add any referenced columns in the WHERE clause
	if ast.Where != nil {
		exprs = append(exprs, ast.Where.Expr)
	}

	// add any referenced columns in the HAVING clause
	if ast.Having != nil {
		exprs = append(exprs, ast.Having.Expr)
	}

	for _, expr := range exprs {
		columns, err := getReferencedColumns(expr)
		if err != nil {
			return nil, err
		}

		addReferencedColumns(columns)

	}

	return expressions, nil
}

// planSimpleSelectExpr takes a simple select expression and returns a query execution plan for it.
func planSimpleSelectExpr(ctx *ExecutionCtx, ss *sqlparser.SimpleSelect) (operator Operator, err error) {
	sExprs, err := refColsInSelectStmt(ss)
	if err != nil {
		return nil, err
	}

	// ensure no columns are referenced in any of the select expressions
	for _, expr := range sExprs {
		if len(expr.RefColumns) != 0 {
			nonSystemVar := false
			// Check if the RefColumns actually reference table columns, or if they only refer to
			// system variables.
			for _, colRef := range expr.RefColumns {
				if !strings.HasPrefix(colRef.Name, "@@") {
					nonSystemVar = true
					break
				}
			}
			// If at least one of the RefColumns refers to anything other than a system variable,
			// the query is invalid since there is no table to reference.
			if nonSystemVar {
				return nil, fmt.Errorf("no column reference allowed in simple select statement")
			}
		}
	}

	// TODO: support distinct within the simple select query
	queryPlan := &Select{sExprs: sExprs, source: &Dual{}}

	return queryPlan, nil
}

// planSelectExpr takes a select struct and returns a query execution plan for it.
func planSelectExpr(ctx *ExecutionCtx, ast *sqlparser.Select) (operator Operator, err error) {

	if ast.From != nil {
		operator, err = planFromExpr(ctx, ast.From, ast.Where)
		if err != nil {
			return nil, err
		}
	}

	appender := &SourceAppend{
		source: operator,
	}

	operator = appender

	selectOp := &Select{
		source: operator,
	}

	selectOp.sExprs, err = refColsInSelectStmt(ast)
	if err != nil {
		return nil, err
	}

	refExprs, err := getReferencedExpressions(ast, ctx, selectOp.sExprs)
	if err != nil {
		return nil, err
	}

	selectOp.sExprs = append(selectOp.sExprs, refExprs...)

	hasSubquery := selectOp.sExprs.HasSubquery()

	appender.hasSubquery = hasSubquery

	if ast.Where != nil {
		matcher, err := NewSQLExpr(ast.Where.Expr)
		if err != nil {
			return nil, err
		}

		selectOp.source = &Filter{
			source:      operator,
			matcher:     matcher,
			hasSubquery: hasSubquery,
		}
	}

	operator = selectOp

	// this handles queries like "select sum(a) from foo"
	hasAggSelectExpr := len(selectOp.sExprs.AggFunctions()) != 0

	// This handles GROUP BY expressions and/or aggregate functions in
	// select expressions - as aggregate functions with no GROUP BY
	// clause imply a single group
	if len(ast.GroupBy) != 0 || hasAggSelectExpr {

		operator, err = planGroupBy(ast, selectOp)
		if err != nil {
			return nil, err
		}
	}

	if ast.Having != nil && len(ast.GroupBy) == 0 {
		operator, err = planHaving(ast.Having, selectOp)
		if err != nil {
			return nil, err
		}
	}

	if len(ast.OrderBy) != 0 {
		operator, err = planOrderBy(ast, operator, selectOp.sExprs)
		if err != nil {
			return nil, err
		}
	}

	// Ensure that any supporting fields included to evaluate
	// various clauses are projected out in the final result set.
	operator = &Project{
		source: operator,
		sExprs: selectOp.sExprs,
	}

	if ast.Limit != nil {
		operator, err = planLimit(ast.Limit, operator)
		if err != nil {
			return nil, err
		}
	}

	if hasSubquery {
		log.Logf(log.DebugLow, "hasSubQuery at depth %v\n", ctx.Depth)
	}

	operator = &SourceRemove{
		source:      operator,
		hasSubquery: hasSubquery,
	}

	return operator, nil

}

// planLimit returns a query execution plan for a LIMIT clause.
func planLimit(expr *sqlparser.Limit, source Operator) (Operator, error) {

	operator := &Limit{
		source: source,
	}

	// TODO: our MySQL parser only supports the LIMIT offset, row_count syntax
	// and not the equivalent LIMIT row_count OFFSET offset syntax.
	//
	// For compatibility with PostgreSQL, MySQL supports both so we should
	// update our parser.
	//
	eval, err := NewSQLValue(expr.Offset)
	if err != nil {
		return nil, err
	}

	if expr.Offset != nil {
		offset, ok := eval.(SQLInt)
		if !ok {
			return nil, fmt.Errorf("LIMIT offset must be an integer")
		}
		operator.offset = int64(offset)

		if operator.offset < 0 {
			return nil, fmt.Errorf("LIMIT offset can not be negative")
		}
	}

	eval, err = NewSQLValue(expr.Rowcount)
	if err != nil {
		return nil, err
	}

	rowcount, ok := eval.(SQLInt)
	if !ok {
		return nil, fmt.Errorf("LIMIT row count must be an integer")
	}

	operator.rowcount = int64(rowcount)

	if operator.rowcount < 0 {
		return nil, fmt.Errorf("LIMIT row count can not be negative")
	}

	return operator, nil

}

// planOrderBy returns a query execution plan for an ORDER BY clause.
func planOrderBy(ast *sqlparser.Select, source Operator, sExprs SelectExpressions) (Operator, error) {

	//
	// TODO: doesn't make sense to allow ORDER BY aggregate expression without
	// a GROUP BY clause
	//
	orderBy := &OrderBy{
		source: source,
	}

	for _, i := range ast.OrderBy {

		expr, err := getOrderByTerm(sExprs, i.Expr)
		if err != nil {
			return nil, err
		}

		value, err := NewSQLValue(expr)
		if err != nil {
			return nil, err
		}

		key := orderByKey{
			value:     value,
			isAggFunc: hasAggFunctions(expr),
			ascending: i.Direction == sqlparser.AST_ASC,
		}

		orderBy.keys = append(orderBy.keys, key)
	}

	return orderBy, nil
}

// planGroupBy returns a query execution plan for a GROUP BY clause.
func planGroupBy(ast *sqlparser.Select, s *Select) (Operator, error) {
	groupBy := ast.GroupBy
	having := ast.Having

	gb := &GroupBy{
		sExprs: s.sExprs,
	}

	var expr sqlparser.Expr
	if having != nil {
		expr = having.Expr
	}

	// create a matcher that can evaluate the HAVING expression
	matcher, err := NewSQLExpr(expr)
	if err != nil {
		return nil, err
	}

	gb.matcher = matcher

	for _, valExpr := range groupBy {
		expr := sqlparser.Expr(valExpr)

		// a GROUP BY clause can't refer to non-aggregated columns in
		// the select list that are not named in the GROUP BY clause
		parsedExpr, err := getGroupByTerm(gb.sExprs, expr)
		if err != nil {
			return nil, err
		}

		gb.exprs = append(gb.exprs, parsedExpr)
	}

	// add any referenced columns to the select operator
	for _, expr := range s.sExprs {
		if len(expr.RefColumns) != 0 {
			for _, column := range expr.RefColumns {
				if !s.sExprs.Contains(*column) {
					newSExpr := SelectExpression{Column: *column, Referenced: true}
					s.sExprs = append(s.sExprs, newSExpr)
				}
			}
		}
	}

	gb.source = s

	return gb, nil
}

// planHaving returns a query execution plan for a HAVING clause.
func planHaving(having *sqlparser.Where, s *Select) (Operator, error) {

	// create a matcher that can evaluate the HAVING expression
	matcher, err := NewSQLExpr(having.Expr)
	if err != nil {
		return nil, err
	}

	// add any referenced columns to the select operator
	var newSExprs SelectExpressions
	for _, sExpr := range s.sExprs {
		if len(sExpr.RefColumns) != 0 {
			for _, column := range sExpr.RefColumns {
				newSExpr := SelectExpression{Column: *column, Referenced: true}
				newSExprs = append(newSExprs, newSExpr)
			}
		}
	}

	s.sExprs = append(s.sExprs, newSExprs...)

	hv := &Having{
		sExprs:  s.sExprs,
		source:  s,
		matcher: matcher,
	}

	return hv, nil
}

// getGroupByTerm returns the referenced expression in an GROUP BY clause if the
// expression is aliased by column name or position. Otherwise, it returns the
// expression supplied.
func getGroupByTerm(sExprs []SelectExpression, e sqlparser.Expr) (sqlparser.Expr, error) {

	switch expr := e.(type) {

	case *sqlparser.ColName:

		if term := getColumnTerm(sExprs, expr); term != nil {
			return term, nil
		}

		return expr, nil

	case sqlparser.NumVal:

		term, err := getNumericTerm(sExprs, expr)
		if err != nil {
			return nil, fmt.Errorf("Unknown column in 'GROUP BY' clause: '%v'", err)
		}

		return term, nil

	}

	return e, nil

}

// getOrderByTerm returns the referenced expression in an ORDER BY clause if the
// expression is aliased. Otherwise, it returns the expression supplied - provided
// it is valid as an ORDER BY term.
func getOrderByTerm(sExprs SelectExpressions, e sqlparser.Expr) (sqlparser.Expr, error) {
	switch expr := e.(type) {

	case *sqlparser.ColName:

		if term := getColumnTerm(sExprs, expr); term != nil {
			return term, nil
		}

		return expr, nil

	case sqlparser.NumVal:

		term, err := getNumericTerm(sExprs, e)
		if err != nil {
			return nil, fmt.Errorf("Unknown column in 'ORDER BY' clause: '%v'", err)
		}

		return term, nil

	case *sqlparser.FuncExpr:

		return expr, nil

	}

	return e, nil

}

func getColumnTerm(sExprs SelectExpressions, expr *sqlparser.ColName) sqlparser.Expr {

	for i, sExpr := range sExprs {

		if sExpr.Table == string(expr.Qualifier) && sExpr.Name == string(expr.Name) {
			return sExprs[i].Expr
		}

		if len(sExpr.RefColumns) != 0 {
			for _, column := range sExpr.RefColumns {
				if column.Table == string(expr.Qualifier) && column.Name == string(expr.Name) {
					return sExprs[i].Expr
				}
			}
		} else {
			// This handles column names that are aliases for actual
			// expressions. For example:
			//
			// SELECT a + b as c from foo GROUP by c ORDER by c;
			//
			// In this case, the both the ORDER BY and GROUP BY terms will
			// be references to the underlying expresssion - (a + b)
			if sExpr.Name == string(expr.Name) {
				return sExprs[i].Expr
			}
		}
	}

	return nil
}

func getNumericTerm(sExprs SelectExpressions, expr sqlparser.Expr) (sqlparser.Expr, error) {
	i, err := strconv.ParseInt(sqlparser.String(expr), 10, 64)
	if err != nil {
		return nil, err
	}

	var j int64

	for _, sExpr := range sExprs {
		if !sExpr.Referenced {
			j++
		}
		if j == i {
			return sExpr.Expr, nil
		}
	}
	return nil, fmt.Errorf("%v", i)
}

// refColsInSelectExpr returns a slice of select columns - each holding
// a non-star select expression and any columns referenced in it.
func refColsInSelectExpr(exprs sqlparser.SelectExprs) ([]SelectExpression, error) {
	sExprs := make([]SelectExpression, 0)

	for _, sExpr := range exprs {

		switch expr := sExpr.(type) {

		// TODO: validate no mixture of star and non-star expression
		case *sqlparser.StarExpr:
			continue

		case *sqlparser.NonStarExpr:

			columns, err := getReferencedColumns(expr.Expr)
			if err != nil {
				return nil, err
			}

			column := Column{View: string(expr.As)}

			selectExpression := SelectExpression{column, columns, expr.Expr, false}

			if c, ok := expr.Expr.(*sqlparser.ColName); ok {
				selectExpression.Table = string(c.Qualifier)
				selectExpression.Name = string(c.Name)

				if selectExpression.View == "" {
					selectExpression.View = string(c.Name)
				}
			} else {
				// get a string representation of the expression if
				// it isn't a column name e.g. sum(foo.a) for aggregate
				// expressions
				selectExpression.Name = sqlparser.String(expr.Expr)
			}

			if selectExpression.View == "" {
				// TODO: this currently includes the table name for
				// aggregate functions - e.g. if you have a query like:
				//
				// select a, sum(b) from foo;
				//
				// The headers in the result set are displayed as:
				//
				// a	sum(foo.b)
				//
				selectExpression.View = sqlparser.String(expr.Expr)
			}

			sExprs = append(sExprs, selectExpression)

		default:
			return nil, fmt.Errorf("unknown SelectExprs in refColsInSelectExpr: %T", expr)
		}
	}

	return sExprs, nil
}

// refColsInSelectStmt returns any columns referenced in the select statement.
func refColsInSelectStmt(ss sqlparser.SelectStatement) ([]SelectExpression, error) {

	switch stmt := ss.(type) {

	case *sqlparser.Select:

		return refColsInSelectExpr(stmt.SelectExprs)

	case *sqlparser.Union:
		l, err := refColsInSelectStmt(stmt.Left)
		if err != nil {
			return nil, err
		}

		r, err := refColsInSelectStmt(stmt.Right)
		if err != nil {
			return nil, err
		}

		return append(l, r...), nil

	case *sqlparser.SimpleSelect:

		return refColsInSelectExpr(stmt.SelectExprs)

	default:
		return nil, fmt.Errorf("unknown SelectStatement in refColsInSelectStmt: %T", stmt)
	}

}

// planFromExpr takes one or more table expressions and returns an Operator for the
// data source. If more than one table expression exists, it constructs a join
// operator (which is left deep for more than two table expressions).
func planFromExpr(ctx *ExecutionCtx, tExpr sqlparser.TableExprs, where *sqlparser.Where) (Operator, error) {

	if len(tExpr) == 0 {
		return nil, fmt.Errorf("can't plan table expression with no tables")
	} else if len(tExpr) == 1 {
		return planTableExpr(ctx, tExpr[0], where)
	}

	var left, right Operator
	var err error

	if len(tExpr) == 2 {

		left, err = planTableExpr(ctx, tExpr[0], where)
		if err != nil {
			return nil, fmt.Errorf("error planning left table expr: %v", err)
		}

		right, err = planTableExpr(ctx, tExpr[1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning right table expr: %v", err)
		}

	} else {

		left, err = planFromExpr(ctx, tExpr[:len(tExpr)-1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning left forest: %v", err)
		}

		right, err = planTableExpr(ctx, tExpr[len(tExpr)-1], where)
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
func planTableExpr(ctx *ExecutionCtx, tExpr sqlparser.TableExpr, where *sqlparser.Where) (Operator, error) {
	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:
		// this is a simple table to get data from
		return planSimpleTableExpr(ctx, expr, where)

	case *sqlparser.ParenTableExpr:
		return planTableExpr(ctx, expr.Expr, where)

	case *sqlparser.JoinTableExpr:
		left, err := planTableExpr(ctx, expr.LeftExpr, where)
		if err != nil {
			return nil, fmt.Errorf("error on left join node: %v", err)
		}

		right, err := planTableExpr(ctx, expr.RightExpr, where)
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

// planTableName takes a table name and returns an operator to get
// data from the appropriate source.
func planTableName(c *ExecutionCtx, t *sqlparser.TableName, w *sqlparser.Where) (Operator, error) {

	var matcher SQLExpr
	var err error

	dbName := strings.ToLower(string(t.Qualifier))
	isInformationSchema := dbName == InformationSchema || strings.ToLower(c.Db) == InformationSchema

	if isInformationSchema {

		// ConfigDataSource is a special table that handles queries against
		// the 'information_schema' database
		cds := &ConfigDataSource{
			tableName: strings.ToLower(string(t.Name)),
		}

		c.Db = InformationSchema

		// if we got a valid filter/matcher, use it
		if err == nil {
			cds.matcher = matcher
		}

		return cds, nil
	}

	ts := &TableScan{
		tableName: string(t.Name),
		dbName:    string(t.Qualifier),
		matcher:   matcher,
	}

	return ts, nil
}

// planSimpleTableExpr takes a simple table expression and returns an operator that can iterate
// over its result set.
func planSimpleTableExpr(c *ExecutionCtx, s *sqlparser.AliasedTableExpr, w *sqlparser.Where) (Operator, error) {
	switch expr := s.Expr.(type) {

	case *sqlparser.TableName:
		source, err := planTableName(c, expr, w)
		if err != nil {
			return nil, err
		}

		// use actual table name if table expression isn't aliased
		// otherwise, rename the table to its aliased form.
		if len(s.As) == 0 {
			return source, nil
		}

		as := &AliasedSource{
			source:    source,
			tableName: string(s.As),
		}

		return as, nil

	case *sqlparser.Subquery:

		source, err := PlanQuery(c, expr.Select)
		if err != nil {
			return nil, err
		}

		as := &AliasedSource{
			source:    source,
			tableName: string(s.As),
		}

		return as, nil

	default:
		return nil, fmt.Errorf("can't yet handle simple table expression type %T", expr)
	}

}

// getReferencedColumns accepts several exppressions and returns a slice
// of referenced columns in those expressions.
func getReferencedColumns(exprs ...sqlparser.Expr) ([]*Column, error) {

	log.Logf(log.DebugLow, "getReferencedColumns: %#v (type is %T)\n", exprs, exprs)

	columns := make([]*Column, 0)

	for _, e := range exprs {
		refColumns, err := referencedColumns(e)
		if err != nil {
			return nil, err
		}

		columns = append(columns, refColumns...)
	}

	return columns, nil
}

// referencedColumns returns a slice of referenced columns in an expression.
func referencedColumns(e sqlparser.Expr) ([]*Column, error) {

	log.Logf(log.DebugLow, "referencedColumns: %#v (type is %T)\n", e, e)

	switch expr := e.(type) {

	case sqlparser.ValTuple:
		columns := []*Column{}
		for _, valTuple := range expr {
			refCols, err := referencedColumns(valTuple)
			if err != nil {
				return nil, err
			}
			columns = append(columns, refCols...)
		}
		return columns, nil

	case sqlparser.NumVal, sqlparser.StrVal, *sqlparser.NullVal:
		return nil, nil

	case *sqlparser.ColName:

		c := &Column{
			Table: string(expr.Qualifier),
			Name:  string(expr.Name),
			View:  string(expr.Name),
		}
		return []*Column{c}, nil

	case *sqlparser.BinaryExpr:

		return getReferencedColumns(expr.Left, expr.Right)

	case *sqlparser.AndExpr:

		return getReferencedColumns(expr.Left, expr.Right)

	case *sqlparser.OrExpr:

		return getReferencedColumns(expr.Left, expr.Right)

	case *sqlparser.ParenBoolExpr:

		return referencedColumns(expr.Expr)

	case *sqlparser.ComparisonExpr:

		return getReferencedColumns(expr.Left, expr.Right)

	case *sqlparser.RangeCond:

		return getReferencedColumns(expr.From, expr.To, expr.Left)

	case *sqlparser.NullCheck:

		return referencedColumns(expr.Expr)

	case *sqlparser.UnaryExpr:

		return referencedColumns(expr.Expr)

	case *sqlparser.NotExpr:

		return referencedColumns(expr.Expr)

	case *sqlparser.Subquery:

		sc, err := refColsInSelectStmt(expr.Select)
		if err != nil {
			return nil, err
		}

		columns := SelectExpressions(sc).GetColumns()

		for _, column := range columns {
			column.InSubquery = true
		}

		return columns, nil

	case *sqlparser.FuncExpr:

		sc, err := refColsInSelectExpr(expr.Exprs)
		if err != nil {
			return nil, err
		}

		return SelectExpressions(sc).GetColumns(), nil

	case *sqlparser.CaseExpr:

		columns, err := getReferencedColumns(expr.Else, expr.Expr)
		if err != nil {
			return nil, err
		}

		for _, when := range expr.Whens {

			c, err := getReferencedColumns(when.Cond)
			if err != nil {
				return nil, err
			}

			columns = append(columns, c...)

			c, err = getReferencedColumns(when.Val)
			if err != nil {
				return nil, err
			}

			columns = append(columns, c...)
		}

		return columns, nil

		// TODO: fill these in
	case sqlparser.ValArg:

		return nil, fmt.Errorf("referenced columns for ValArg for NYI")

	case *sqlparser.ExistsExpr:

		return nil, fmt.Errorf("referenced columns for ExistsExpr for NYI")

	case nil:

		return nil, nil

	default:

		return nil, fmt.Errorf("referenced columns NYI for: %T", expr)

	}

	return nil, fmt.Errorf("referenced columns (on %T) reached an unreachable point", e)
}

// hasAggFunctions returns true if expression e contains an aggregate function and false otherwise.
func hasAggFunctions(e sqlparser.Expr) bool {

	log.Logf(log.DebugLow, "hasAggFunctions: %#v (type is %T)\n", e, e)

	switch expr := e.(type) {

	case sqlparser.ValTuple:

		for _, valTuple := range expr {
			if hasAggFunctions(valTuple) {
				return true
			}
		}

		return false

	case sqlparser.NumVal, sqlparser.StrVal, *sqlparser.NullVal, *sqlparser.ColName:

		return false

	case *sqlparser.BinaryExpr:

		if hasAggFunctions(expr.Left) || hasAggFunctions(expr.Right) {
			return true
		}

		return false

	case *sqlparser.AndExpr:

		if hasAggFunctions(expr.Left) || hasAggFunctions(expr.Right) {
			return true
		}

		return false

	case *sqlparser.OrExpr:

		if hasAggFunctions(expr.Left) || hasAggFunctions(expr.Right) {
			return true
		}

		return false

	case *sqlparser.ParenBoolExpr:

		return hasAggFunctions(expr.Expr)

	case *sqlparser.ComparisonExpr:

		if hasAggFunctions(expr.Left) || hasAggFunctions(expr.Right) {
			return true
		}

		return false

	case *sqlparser.RangeCond:

		if hasAggFunctions(expr.Left) || hasAggFunctions(expr.From) || hasAggFunctions(expr.To) {
			return true
		}

		return false

	case *sqlparser.NullCheck:

		return false

	case *sqlparser.UnaryExpr:

		return hasAggFunctions(expr.Expr)

	case *sqlparser.NotExpr:

		return hasAggFunctions(expr.Expr)

	case *sqlparser.FuncExpr:

		return isAggFunction(expr.Name)

	case *sqlparser.Subquery:

		return false

	case *sqlparser.CaseExpr:

		if hasAggFunctions(expr.Else) || hasAggFunctions(expr.Expr) {
			return true
		}

		for _, when := range expr.Whens {
			if hasAggFunctions(when.Cond) || hasAggFunctions(when.Val) {
				return true
			}
		}

		return false

	case nil:

		return false

	default:
		panic(fmt.Sprintf("hasAggFunctions NYI for: %T", expr))

	}

}

// isAggFunction returns true if the byte slice e contains the name of an aggregate function and false otherwise.
func isAggFunction(e []byte) bool {
	switch strings.ToLower(string(e)) {
	case "avg", "sum", "count", "max", "min":
		return true
	default:
		return false
	}
}
