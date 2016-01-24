package evaluator

import (
	"fmt"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"strconv"
	"strings"
)

var (
	InformationDatabase = "information_schema"
)

// PlanQuery constructs a query plan to satisfy the select statement.
func PlanQuery(ctx *ExecutionCtx, ss sqlparser.SelectStatement) (Operator, error) {

	switch ast := ss.(type) {

	case *sqlparser.Select:

		log.Logf(log.DebugLow, "Planning select query for %#v\n", sqlparser.String(ss))

		o, err := planQuery(ctx, ast)
		if err != nil {
			return nil, err
		}

		log.Logf(log.DebugLow, "Original query plan: \n%v\n", PrettyPrintPlan(o))

		o, err = OptimizeOperator(ctx, o)
		if err != nil {
			return nil, err
		}

		log.Logf(log.DebugLow, "Optimized query plan: \n%v\n", PrettyPrintPlan(o))

		return o, err

	case *sqlparser.SimpleSelect:

		log.Logf(log.DebugLow, "Planning simple select query for %#v\n", sqlparser.String(ss))

		o, err := planSimpleSelectExpr(ctx, ast)
		if err != nil {
			return nil, err
		}

		return OptimizeOperator(ctx, o)

	case *sqlparser.Union:

		return nil, fmt.Errorf("union select statement not yet implemented")

	default:

		return nil, fmt.Errorf("unknown select statement: %T", ast)

	}
}

// getColumnExpr returns the referenced Select Expression indicated by a column expression.
func getColumnExpr(sExprs SelectExpressions, expr *sqlparser.ColName) (*SelectExpression, error) {

	for i, sExpr := range sExprs {

		if sExpr.Table == string(expr.Qualifier) && sExpr.Name == string(expr.Name) {
			return &sExprs[i], nil
		}

		if len(sExpr.RefColumns) == 0 {
			// This handles column names that are aliases for actual
			// expressions. For example:
			//
			// SELECT a + b as c from foo GROUP by c ORDER by c;
			//
			// In this case, both the ORDER BY and GROUP BY terms will
			// be references to the underlying expresssion - (a + b)
			if sExpr.Name == string(expr.Name) {
				return &sExprs[i], nil
			}
		}
	}

	sqlExpr, err := NewSQLExpr(expr)
	if err != nil {
		return nil, err
	}

	column := &Column{
		Name: sqlExpr.String(),
		View: sqlExpr.String(),
	}

	return &SelectExpression{
		Expr:   sqlExpr,
		Column: column,
	}, nil

}

// getGroupByTerm returns the referenced expression in an GROUP BY clause if the
// expression is aliased by column name or position. Otherwise, it returns the
// expression supplied.
func getGroupByTerm(sExprs []SelectExpression, e sqlparser.Expr) (sExpr *SelectExpression, err error) {

	switch expr := e.(type) {

	case *sqlparser.ColName:

		sExpr, err = getColumnExpr(sExprs, expr)

	case sqlparser.NumVal:

		sExpr, err = getNumericExpr(sExprs, expr)

	}

	if err != nil {
		return nil, fmt.Errorf("Unknown column in 'GROUP BY' clause: '%v'", err)
	}

	if sExpr == nil {

		expr, err := NewSQLExpr(e)
		if err != nil {
			return nil, err
		}

		column := &Column{
			Name: expr.String(),
			View: expr.String(),
		}

		sExpr = &SelectExpression{
			Expr:   expr,
			Column: column,
		}

	}

	aggFuncs, err := getAggFunctions(sExpr.Expr)
	if err != nil {
		return nil, err
	}

	if len(aggFuncs) != 0 {
		return nil, fmt.Errorf("Invalid use of group function")
	}

	return sExpr, nil

}

// getNumericExpr returns the referenced Select Expression indicated numerically.
func getNumericExpr(sExprs SelectExpressions, expr sqlparser.Expr) (*SelectExpression, error) {

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
			return &sExpr, nil
		}
	}

	return nil, fmt.Errorf("%v", i)
}

// getOrderByTerm returns the referenced expression in an ORDER BY clause if the
// expression is aliased. Otherwise, it returns the expression supplied - provided
// it is valid as an ORDER BY term.
func getOrderByTerm(sExprs SelectExpressions, e sqlparser.Expr) (*SelectExpression, error) {

	switch expr := e.(type) {

	case *sqlparser.ColName:

		e, err := getColumnExpr(sExprs, expr)
		if err != nil {
			return nil, err
		}

		return e, nil

	case sqlparser.NumVal:

		e, err := getNumericExpr(sExprs, expr)
		if err != nil {
			return nil, fmt.Errorf("Unknown column in 'ORDER BY' clause: '%v'", err)
		}

		return e, nil

	}

	expr, err := NewSQLExpr(e)
	if err != nil {
		return nil, err
	}

	// check if we aleady have the term referenced
	// earlier in the statement
	for _, sExpr := range sExprs {
		if sExpr.Expr.String() == expr.String() {
			return &sExpr, nil
		}
	}

	column := &Column{
		Name: expr.String(),
		View: expr.String(),
	}

	sExpr := &SelectExpression{
		Expr:   expr,
		Column: column,
	}

	return sExpr, nil
}

// getReferencedExpressions parses the parts of the select AST and returns the
// referenced select expressions and all the expressions comprising the query.
func getReferencedExpressions(ast *sqlparser.Select, ctx *ExecutionCtx, sExprs SelectExpressions) (SelectExpressions, []SQLExpr, error) {

	var expressions SelectExpressions

	addReferencedColumns := func(columns []*Column) {
		for _, column := range columns {
			// only add basic columns present in table configuration
			hasColumn := (expressions.Contains(*column) || sExprs.Contains(*column))
			if ctx.ParseCtx.IsSchemaColumn(column) && !hasColumn && !hasStarExpr(ast) {
				sqlExpr := SQLFieldExpr{column.Table, column.Name}
				expression := SelectExpression{
					Column:     column,
					Referenced: true,
					Expr:       sqlExpr,
				}
				expressions = append(expressions, expression)
			}
		}
	}

	var exprs []SQLExpr

	// add any referenced columns in the GROUP BY clause
	for _, key := range ast.GroupBy {
		expr, err := getGroupByTerm(sExprs, key)
		if err != nil {
			return nil, nil, err
		}

		exprs = append(exprs, expr.Expr)
	}

	// add any referenced columns in the ORDER BY clause
	for _, key := range ast.OrderBy {
		expr, err := getOrderByTerm(sExprs, key.Expr)
		if err != nil {
			return nil, nil, err
		}

		exprs = append(exprs, expr.Expr)
	}

	// add any referenced columns in the WHERE clause
	if ast.Where != nil {
		expr, err := NewSQLExpr(ast.Where.Expr)
		if err != nil {
			return nil, nil, err
		}

		exprs = append(exprs, expr)
	}

	// add any referenced columns in the HAVING clause
	if ast.Having != nil {
		expr, err := NewSQLExpr(ast.Having.Expr)
		if err != nil {
			return nil, nil, err
		}
		exprs = append(exprs, expr)
	}

	for _, expr := range exprs {
		columns, err := referencedColumns(expr)
		if err != nil {
			return nil, nil, err
		}

		addReferencedColumns(columns)
	}

	return expressions, exprs, nil
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

// planGroupBy returns a query execution plan for a GROUP BY clause.
func planGroupBy(ast *sqlparser.Select, sExprs SelectExpressions) (*GroupBy, error) {

	groupBy := ast.GroupBy
	having := ast.Having

	gb := &GroupBy{
		sExprs: sExprs,
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
		gbExpr, err := getGroupByTerm(gb.sExprs, expr)
		if err != nil {
			return nil, err
		}

		gb.exprs = append(gb.exprs, *gbExpr)
	}

	return gb, nil
}

// planLimit returns a query execution plan for a LIMIT clause.
func planLimit(expr *sqlparser.Limit) (*Limit, error) {

	operator := &Limit{}

	// TODO: our MySQL parser only supports the LIMIT offset, row_count syntax
	// and not the equivalent LIMIT row_count OFFSET offset syntax.
	//
	// For compatibility with PostgreSQL, MySQL supports both so we should
	// update our parser.
	//
	eval, err := NewSQLExpr(expr.Offset)
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

	eval, err = NewSQLExpr(expr.Rowcount)
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
func planOrderBy(ast *sqlparser.Select, sExprs SelectExpressions) (*OrderBy, error) {

	//
	// TODO: doesn't make sense to allow ORDER BY aggregate expression without
	// a GROUP BY clause
	//

	orderBy := &OrderBy{}

	for _, i := range ast.OrderBy {

		expr, err := getOrderByTerm(sExprs, i.Expr)
		if err != nil {
			return nil, err
		}

		aggFuncs, err := getAggFunctions(expr.Expr)
		if err != nil {
			return nil, err
		}

		key := orderByKey{
			expr:      expr,
			isAggFunc: len(aggFuncs) != 0,
			ascending: i.Direction == sqlparser.AST_ASC,
		}

		orderBy.keys = append(orderBy.keys, key)
	}

	return orderBy, nil
}

// planQuery takes a select struct and returns a query execution plan for it.
func planQuery(ctx *ExecutionCtx, ast *sqlparser.Select) (operator Operator, err error) {

	if ast.From != nil {
		operator, err = planFromExpr(ctx, ast.From, ast.Where)
		if err != nil {
			return nil, err
		}
	}

	baseSExprs, err := refColsInSelectStmt(ast)
	if err != nil {
		return nil, err
	}

	refExprs, sqlExprs, err := getReferencedExpressions(ast, ctx, baseSExprs)
	if err != nil {
		return nil, err
	}

	sExprs := append(baseSExprs, refExprs...)

	var containsSubquery bool

	for _, e := range sqlExprs {

		containsSubquery, err = hasSubquery(e)
		if err != nil {
			return nil, err
		}

		if containsSubquery {
			break
		}

	}

	if containsSubquery {
		log.Logf(log.DebugLow, "hasSubQuery at depth %v\n", ctx.Depth)
	}

	operator = &SourceAppend{
		source:      operator,
		hasSubquery: containsSubquery,
	}

	if ast.Where != nil {
		matcher, err := NewSQLExpr(ast.Where.Expr)
		if err != nil {
			return nil, err
		}

		operator = &Filter{
			source:      operator,
			matcher:     matcher,
			hasSubquery: containsSubquery,
		}
	}

	var aggSelExprs SelectExpressions

	for _, expr := range sqlExprs {

		aggExprs, err := getAggFunctions(expr)
		if err != nil {
			return nil, err
		}

		for _, aggExpr := range aggExprs {

			column := &Column{
				Name: aggExpr.String(),
				View: aggExpr.String(),
			}

			aggSelExpr := SelectExpression{
				Expr:       aggExpr,
				Column:     column,
				Referenced: true,
			}

			aggSelExprs = append(aggSelExprs, aggSelExpr)
		}
	}

	sExprs = append(sExprs, aggSelExprs...)

	refAggFunction := len(aggSelExprs) != 0

	if !refAggFunction {
		// This handles aggregate function expressions with
		// select expressions; e.g. "select sum(a) from foo"
		aggFuncs, err := sExprs.AggFunctions()
		if err != nil {
			return nil, err
		}

		refAggFunction = len(*aggFuncs) != 0
	}

	// This handles GROUP BY expressions and/or aggregate functions in
	// other parts of the query - sinc aggregate functions with no
	// GROUP BY clause imply a single group
	needsGroupBy := len(ast.GroupBy) != 0 || refAggFunction

	if needsGroupBy {
		gb, err := planGroupBy(ast, sExprs)
		if err != nil {
			return nil, err
		}
		gb.source = operator
		operator = gb
	}

	// in planning the GROUP BY operator, we optimize
	// to include the filter (HAVING clause) within
	// that operator so we can push down more easily
	if ast.Having != nil && !needsGroupBy {
		matcher, err := NewSQLExpr(ast.Having.Expr)
		if err != nil {
			return nil, err
		}

		operator = &Filter{
			source:      operator,
			matcher:     matcher,
			hasSubquery: containsSubquery,
		}
	}

	needsDistinctGroup := !refAggFunction && (ast.Distinct == sqlparser.AST_DISTINCT)

	if needsDistinctGroup {
		operator = &GroupBy{
			exprs:   baseSExprs,
			matcher: SQLTrue,
			sExprs:  sExprs,
			grouped: len(ast.GroupBy) != 0,
			source:  operator,
		}
	}

	if len(ast.OrderBy) != 0 {
		ob, err := planOrderBy(ast, sExprs)
		if err != nil {
			return nil, err
		}
		ob.source = operator
		operator = ob
	}

	if ast.Limit != nil {
		lm, err := planLimit(ast.Limit)
		if err != nil {
			return nil, err
		}
		lm.source = operator
		operator = lm
	}

	operator = &Project{
		source: operator,
		sExprs: sExprs,
	}

	operator = &SourceRemove{
		source:      operator,
		hasSubquery: containsSubquery,
	}

	return operator, nil

}

// planSimpleSelectExpr takes a simple select expression and returns a query execution plan for it.
func planSimpleSelectExpr(ctx *ExecutionCtx, ss *sqlparser.SimpleSelect) (Operator, error) {

	sExprs, err := refColsInSelectStmt(ss)
	if err != nil {
		return nil, err
	}

	// ensure no columns are referenced in any of the select expressions
	for _, expr := range sExprs {
		if len(expr.RefColumns) != 0 {

			nonSystemVar := false

			// Check if the RefColumns actually reference table columns,
			// or if they only refer to system variables.
			for _, colRef := range expr.RefColumns {
				if !strings.HasPrefix(colRef.Name, "@@") {
					nonSystemVar = true
					break
				}
			}

			// If at least one of the RefColumns refers to anything other than
			// a system variable, the query is invalid since there is no table to reference.
			if nonSystemVar {
				return nil, fmt.Errorf("no column reference allowed in simple select statement")
			}
		}
	}

	// TODO: support distinct within the simple select query
	o := &Project{
		source: &Dual{sExprs: sExprs},
		sExprs: sExprs,
	}

	return o, nil

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

			sqlExpr, err := NewSQLExpr(expr.Expr)
			if err != nil {
				return nil, err
			}

			columns, err := referencedColumns(sqlExpr)
			if err != nil {
				return nil, err
			}

			column := &Column{View: string(expr.As)}

			selectExpression := SelectExpression{
				Column:     column,
				RefColumns: columns,
				Expr:       sqlExpr,
			}

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
		if len(s.As) != 0 {
			switch typedO := source.(type) {
			case *TableScan:
				typedO.aliasName = string(s.As)
			case *SchemaDataSource:
				typedO.aliasName = string(s.As)
			default:
				return nil, fmt.Errorf("planSimpleTableExpr returned unknown type %T", typedO)
			}
		}

		return source, nil

	case *sqlparser.Subquery:

		source, err := PlanQuery(c, expr.Select)
		if err != nil {
			return nil, err
		}

		as := &Subquery{
			source:    source,
			tableName: string(s.As),
		}

		return as, nil

	default:
		return nil, fmt.Errorf("can't yet handle simple table expression type %T", expr)
	}

}

// planTableExpr takes a table expression and returns an Operator that
// can iterate over its result set.
func planTableExpr(ctx *ExecutionCtx, tExpr sqlparser.TableExpr, w *sqlparser.Where) (Operator, error) {

	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:
		// this is a simple table to get data from
		return planSimpleTableExpr(ctx, expr, w)

	case *sqlparser.ParenTableExpr:
		return planTableExpr(ctx, expr.Expr, w)

	case *sqlparser.JoinTableExpr:
		left, err := planTableExpr(ctx, expr.LeftExpr, w)
		if err != nil {
			return nil, fmt.Errorf("error on left join node: %v", err)
		}

		right, err := planTableExpr(ctx, expr.RightExpr, w)
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
	isInformationDatabase := dbName == InformationDatabase || strings.ToLower(c.Db) == InformationDatabase

	if isInformationDatabase {

		// SchemaDataSource is a special table that handles queries against
		// the 'information_schema' database
		cds := &SchemaDataSource{
			tableName: strings.ToLower(string(t.Name)),
		}

		c.Db = InformationDatabase

		// if we got a valid filter/matcher, use it
		if err == nil {
			cds.matcher = matcher
		}

		return cds, nil
	}

	if c.Db == "" {
		c.Db = dbName
	}

	ts := &TableScan{
		tableName: string(t.Name),
		dbName:    dbName,
		matcher:   matcher,
	}

	return ts, nil
}

// refColsInSelectStmt returns any columns referenced in the select statement.
func refColsInSelectStmt(ss sqlparser.SelectStatement) (SelectExpressions, error) {

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
