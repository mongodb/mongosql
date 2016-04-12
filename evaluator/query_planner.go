package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
)

var (
	InformationDatabase = "information_schema"
)

// PlanQuery constructs a query plan to satisfy the select statement.
func PlanQuery(ctx *PlanCtx, ss sqlparser.SelectStatement) (PlanStage, error) {

	switch ast := ss.(type) {

	case *sqlparser.Select:

		log.Logf(log.DebugLow, "Planning select query for %#v\n", sqlparser.String(ss))

		o, err := planQuery(ctx, ast)
		if err != nil {
			return nil, err
		}

		log.Logf(log.DebugLow, "Original query plan: \n%v\n", PrettyPrintPlan(o))

		o, err = OptimizePlan(ctx, o)
		if err != nil {
			return nil, err
		}

		return o, err

	case *sqlparser.SimpleSelect:

		log.Logf(log.DebugLow, "Planning simple select query for %#v\n", sqlparser.String(ss))

		o, err := planSimpleSelectExpr(ctx, ast)
		if err != nil {
			return nil, err
		}

		return OptimizePlan(ctx, o)

	case *sqlparser.Union:

		return nil, fmt.Errorf("union select statement not yet implemented")

	default:

		return nil, fmt.Errorf("unknown select statement: %T", ast)

	}
}

// getColumnExpr returns the referenced Select Expression indicated by a column expression.
func getColumnExpr(sExprs SelectExpressions, expr *sqlparser.ColName, tables map[string]*schema.Table) (*SelectExpression, error) {

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

	sqlExpr, err := NewSQLExpr(expr, tables)
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
func getGroupByTerm(sExprs []SelectExpression, e sqlparser.Expr, tables map[string]*schema.Table) (sExpr *SelectExpression, err error) {

	switch expr := e.(type) {

	case *sqlparser.ColName:

		sExpr, err = getColumnExpr(sExprs, expr, tables)

	case sqlparser.NumVal:

		sExpr, err = getNumericExpr(sExprs, expr)

	}

	if err != nil {
		return nil, fmt.Errorf("Unknown column in 'GROUP BY' clause: '%v'", err)
	}

	if sExpr == nil {

		expr, err := NewSQLExpr(e, tables)
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
func getOrderByTerm(sExprs SelectExpressions, e sqlparser.Expr, tables map[string]*schema.Table) (*SelectExpression, error) {

	switch expr := e.(type) {

	case *sqlparser.ColName:

		e, err := getColumnExpr(sExprs, expr, tables)
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

	expr, err := NewSQLExpr(e, tables)
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
func getReferencedExpressions(ast *sqlparser.Select, planCtx *PlanCtx, sExprs SelectExpressions) (SelectExpressions, []SQLExpr, error) {

	var expressions SelectExpressions

	tables := getDBTables(planCtx)

	addReferencedColumns := func(columns []*Column) {
		for _, column := range columns {
			// only add basic columns present in table configuration
			hasColumn := (expressions.Contains(*column) || sExprs.Contains(*column))
			if planCtx.ParseCtx.IsSchemaColumn(column) && !hasColumn && !hasStarExpr(ast) {
				columnType := getColumnType(tables, column.Table, column.Name)
				sqlExpr := SQLColumnExpr{column.Table, column.Name, *columnType}
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
		expr, err := getGroupByTerm(sExprs, key, tables)
		if err != nil {
			return nil, nil, err
		}

		exprs = append(exprs, expr.Expr)
	}

	// add any referenced columns in the ORDER BY clause
	for _, key := range ast.OrderBy {
		expr, err := getOrderByTerm(sExprs, key.Expr, tables)
		if err != nil {
			return nil, nil, err
		}

		exprs = append(exprs, expr.Expr)
	}

	// add any referenced columns in the WHERE clause
	if ast.Where != nil {
		expr, err := NewSQLExpr(ast.Where.Expr, tables)
		if err != nil {
			return nil, nil, err
		}

		exprs = append(exprs, expr)
	}

	// add any referenced columns in the HAVING clause
	if ast.Having != nil {
		expr, err := NewSQLExpr(ast.Having.Expr, tables)
		if err != nil {
			return nil, nil, err
		}
		exprs = append(exprs, expr)
	}

	for _, expr := range exprs {
		columns, err := referencedColumns(expr, tables)
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

// planFromExpr takes one or more table expressions and returns an PlanStage for the
// data source. If more than one table expression exists, it constructs a join
// operator (which is left deep for more than two table expressions).
func planFromExpr(planCtx *PlanCtx, tExpr sqlparser.TableExprs, where *sqlparser.Where) (PlanStage, error) {

	if len(tExpr) == 0 {
		return nil, fmt.Errorf("can't plan table expression with no tables")
	} else if len(tExpr) == 1 {
		return planTableExpr(planCtx, tExpr[0], where)
	}

	var left, right PlanStage
	var err error

	if len(tExpr) == 2 {

		left, err = planTableExpr(planCtx, tExpr[0], where)
		if err != nil {
			return nil, fmt.Errorf("error planning left table expr: %v", err)
		}

		right, err = planTableExpr(planCtx, tExpr[1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning right table expr: %v", err)
		}

	} else {

		left, err = planFromExpr(planCtx, tExpr[:len(tExpr)-1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning left forest: %v", err)
		}

		right, err = planTableExpr(planCtx, tExpr[len(tExpr)-1], where)
		if err != nil {
			return nil, fmt.Errorf("error planning right table leaf: %v", err)
		}

	}

	join := &JoinStage{
		left:  left,
		right: right,
		kind:  sqlparser.AST_CROSS_JOIN,
	}

	return join, nil

}

// planGroupBy returns a query execution plan for a GROUP BY clause.
func planGroupBy(ast *sqlparser.Select, sExprs SelectExpressions, tables map[string]*schema.Table) (*GroupByStage, error) {

	groupBy := ast.GroupBy

	gb := &GroupByStage{
		selectExprs: sExprs,
	}

	for _, valExpr := range groupBy {

		expr := sqlparser.Expr(valExpr)

		// a GROUP BY clause can't refer to non-aggregated columns in
		// the select list that are not named in the GROUP BY clause
		gbExpr, err := getGroupByTerm(gb.selectExprs, expr, tables)
		if err != nil {
			return nil, err
		}

		gb.keyExprs = append(gb.keyExprs, *gbExpr)
	}

	return gb, nil
}

// planLimit returns a query execution plan for a LIMIT clause.
func planLimit(expr *sqlparser.Limit, tables map[string]*schema.Table) (*LimitStage, error) {

	limitPlan := &LimitStage{}

	// TODO: our MySQL parser only supports the LIMIT offset, row_count syntax
	// and not the equivalent LIMIT row_count OFFSET offset syntax.
	//
	// For compatibility with PostgreSQL, MySQL supports both so we should
	// update our parser.
	//
	eval, err := NewSQLExpr(expr.Offset, tables)
	if err != nil {
		return nil, err
	}

	if expr.Offset != nil {
		offset, ok := eval.(SQLInt)
		if !ok {
			return nil, fmt.Errorf("LIMIT offset must be an integer")
		}
		limitPlan.offset = int64(offset)

		if limitPlan.offset < 0 {
			return nil, fmt.Errorf("LIMIT offset can not be negative")
		}
	}

	eval, err = NewSQLExpr(expr.Rowcount, tables)
	if err != nil {
		return nil, err
	}

	limitCount, ok := eval.(SQLInt)
	if !ok {
		return nil, fmt.Errorf("LIMIT row count must be an integer")
	}

	limitPlan.limit = int64(limitCount)

	if limitPlan.limit < 0 {
		return nil, fmt.Errorf("LIMIT row count can not be negative")
	}

	return limitPlan, nil

}

// planOrderBy returns a query execution plan for an ORDER BY clause.
func planOrderBy(ast *sqlparser.Select, sExprs SelectExpressions, tables map[string]*schema.Table) (*OrderByStage, error) {

	//
	// TODO: doesn't make sense to allow ORDER BY aggregate expression without
	// a GROUP BY clause
	//

	orderBy := &OrderByStage{}

	for _, i := range ast.OrderBy {

		expr, err := getOrderByTerm(sExprs, i.Expr, tables)
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

func getDBTables(planCtx *PlanCtx) map[string]*schema.Table {
	isInformationDatabase := strings.ToLower(planCtx.Db) == InformationDatabase
	if isInformationDatabase || planCtx.Db == "" {
		return nil
	}
	return planCtx.Schema.Databases[planCtx.Db].Tables
}

// planQuery takes a select struct and returns a query execution plan for it.
func planQuery(planCtx *PlanCtx, ast *sqlparser.Select) (plan PlanStage, err error) {

	tables := getDBTables(planCtx)

	if ast.From != nil {
		plan, err = planFromExpr(planCtx, ast.From, ast.Where)
		if err != nil {
			return nil, err
		}
	}

	projectedSelectExprs, err := referencedSelectExpressions(ast, tables)
	if err != nil {
		return nil, err
	}

	refExprs, sqlExprs, err := getReferencedExpressions(ast, planCtx, projectedSelectExprs)
	if err != nil {
		return nil, err
	}

	allSelectExprs := append(projectedSelectExprs, refExprs...)

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
		log.Logf(log.DebugLow, "hasSubQuery is true\n") //at depth %v\n", planCtx.ParseCtx.Depth)
		plan = &SourceAppendStage{source: plan}
	}

	if ast.Where != nil {
		matcher, err := NewSQLExpr(ast.Where.Expr, tables)
		if err != nil {
			return nil, err
		}

		plan = &FilterStage{source: plan, matcher: matcher, hasSubquery: containsSubquery}
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

	allSelectExprs = append(allSelectExprs, aggSelExprs...)

	refAggFunction := len(aggSelExprs) != 0

	if !refAggFunction {
		// This handles aggregate function expressions with
		// select expressions; e.g. "select sum(a) from foo"
		aggFuncs, err := allSelectExprs.AggFunctions()
		if err != nil {
			return nil, err
		}

		refAggFunction = len(*aggFuncs) != 0
	}

	// This handles GROUP BY expressions and/or aggregate functions in
	// other parts of the query - since aggregate functions with no
	// GROUP BY clause imply a single group
	needsGroupBy := len(ast.GroupBy) != 0 || refAggFunction

	if needsGroupBy {
		gb, err := planGroupBy(ast, allSelectExprs, tables)
		if err != nil {
			return nil, err
		}
		gb.source = plan
		plan = gb

		// at this point, all aggregations, computations, etc... have been done in the group
		// by, so we need to make sure that these are replaced by what is now a column.
		projectedSelectExprs = replaceSelectExpressionsWithColumns(groupTempTable, projectedSelectExprs)
		allSelectExprs = replaceSelectExpressionsWithColumns(groupTempTable, allSelectExprs)
	}

	if ast.Having != nil {
		matcher, err := NewSQLExpr(ast.Having.Expr, tables)
		if err != nil {
			return nil, err
		}

		// we need to replace all the SQLAggFunctionExpr inside matcher with fields,
		// because, regardless of whether push down will occur, all aggregations
		// have already been evaluated.
		matcher, err = replaceAggFunctionsWithColumns(groupTempTable, matcher)
		if err != nil {
			return nil, err
		}

		plan = &FilterStage{source: plan, matcher: matcher, hasSubquery: containsSubquery}
	}

	if ast.Distinct == sqlparser.AST_DISTINCT {

		plan = &GroupByStage{
			keyExprs:    projectedSelectExprs,
			selectExprs: allSelectExprs,
			source:      plan,
		}
	}

	if len(ast.OrderBy) != 0 {
		ob, err := planOrderBy(ast, allSelectExprs, tables)
		if err != nil {
			return nil, err
		}
		ob.source = plan
		plan = ob
	}

	if ast.Limit != nil {
		lm, err := planLimit(ast.Limit, tables)
		if err != nil {
			return nil, err
		}
		lm.source = plan
		plan = lm
	}

	plan = &ProjectStage{
		source: plan,
		sExprs: projectedSelectExprs,
	}

	if containsSubquery {
		plan = &SourceRemoveStage{
			source: plan,
		}
	}

	return plan, nil
}

// planSimpleSelectExpr takes a simple select expression and returns a query execution plan for it.
func planSimpleSelectExpr(planCtx *PlanCtx, ss *sqlparser.SimpleSelect) (PlanStage, error) {

	tables := getDBTables(planCtx)

	sExprs, err := referencedSelectExpressions(ss, tables)
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
	o := &ProjectStage{
		source: &DualStage{sExprs: sExprs},
		sExprs: sExprs,
	}

	return o, nil

}

// refColsInSelectExpr returns a slice of select columns - each holding
// a non-star select expression and any columns referenced in it.
func refColsInSelectExpr(exprs sqlparser.SelectExprs, tables map[string]*schema.Table) ([]SelectExpression, error) {

	sExprs := make([]SelectExpression, 0)

	for _, sExpr := range exprs {

		switch expr := sExpr.(type) {

		// TODO: validate no mixture of star and non-star expression
		case *sqlparser.StarExpr:

			continue

		case *sqlparser.NonStarExpr:

			sqlExpr, err := NewSQLExpr(expr.Expr, tables)
			if err != nil {
				return nil, err
			}

			columns, err := referencedColumns(sqlExpr, tables)
			if err != nil {
				return nil, err
			}

			column := &Column{View: string(expr.As), SQLType: sqlExpr.Type()}

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
				selectExpression.Name = sqlExpr.String()
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
				selectExpression.SQLType = sqlExpr.Type()
				selectExpression.View = sqlExpr.String()
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
func planSimpleTableExpr(c *PlanCtx, s *sqlparser.AliasedTableExpr, w *sqlparser.Where) (PlanStage, error) {

	switch expr := s.Expr.(type) {

	case *sqlparser.TableName:

		return planTableName(c, expr, string(s.As), w)

	case *sqlparser.Subquery:

		source, err := PlanQuery(c, expr.Select)
		if err != nil {
			return nil, err
		}

		as := &SubqueryStage{
			source:    source,
			tableName: string(s.As),
		}

		return as, nil

	default:
		return nil, fmt.Errorf("can't yet handle simple table expression type %T", expr)
	}

}

// planTableExpr takes a table expression and returns an PlanStage that
// can iterate over its result set.
func planTableExpr(planCtx *PlanCtx, tExpr sqlparser.TableExpr, w *sqlparser.Where) (PlanStage, error) {

	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:
		// this is a simple table to get data from
		return planSimpleTableExpr(planCtx, expr, w)

	case *sqlparser.ParenTableExpr:
		return planTableExpr(planCtx, expr.Expr, w)

	case *sqlparser.JoinTableExpr:
		left, err := planTableExpr(planCtx, expr.LeftExpr, w)
		if err != nil {
			return nil, fmt.Errorf("error on left join node: %v", err)
		}

		right, err := planTableExpr(planCtx, expr.RightExpr, w)
		if err != nil {
			return nil, fmt.Errorf("error on right join node: %v", err)
		}

		tables := getDBTables(planCtx)

		on, err := NewSQLExpr(expr.On, tables)
		if err != nil {
			return nil, err
		}

		join := &JoinStage{
			left:    left,
			right:   right,
			matcher: on,
			kind:    JoinKind(expr.Join),
		}

		return join, nil

	default:
		return nil, fmt.Errorf("can't handle table expression type %T", expr)
	}
}

// planTableName takes a table name and returns an operator to get
// data from the appropriate source.
func planTableName(planCtx *PlanCtx, t *sqlparser.TableName, aliasName string, w *sqlparser.Where) (PlanStage, error) {

	dbName := strings.ToLower(string(t.Qualifier))
	isInformationDatabase := dbName == InformationDatabase || strings.ToLower(planCtx.Db) == InformationDatabase

	if isInformationDatabase {
		planCtx.Db = InformationDatabase
		return NewSchemaDataSourceStage(string(t.Name), aliasName), nil
	}

	if dbName == "" {
		dbName = planCtx.Db
	}

	return NewMongoSourceStage(planCtx, dbName, string(t.Name), aliasName)
}

// referencedSelectExpressions returns any columns referenced in the select statement.
func referencedSelectExpressions(ss sqlparser.SelectStatement, tables map[string]*schema.Table) (SelectExpressions, error) {

	switch stmt := ss.(type) {

	case *sqlparser.Select:

		return refColsInSelectExpr(stmt.SelectExprs, tables)

	case *sqlparser.Union:

		leftSelectExprs, err := referencedSelectExpressions(stmt.Left, tables)
		if err != nil {
			return nil, err
		}

		rightSelectExprs, err := referencedSelectExpressions(stmt.Right, tables)
		if err != nil {
			return nil, err
		}

		return append(leftSelectExprs, rightSelectExprs...), nil

	case *sqlparser.SimpleSelect:

		return refColsInSelectExpr(stmt.SelectExprs, tables)

	default:
		return nil, fmt.Errorf("unknown SelectStatement in referencedSelectExpressions: %T", stmt)
	}

}

func replaceSelectExpressionsWithColumns(tableName string, sExprs SelectExpressions) SelectExpressions {

	newSExprs := SelectExpressions{}

	for _, sExpr := range sExprs {

		expr, ok := sExpr.Expr.(SQLColumnExpr)
		if !ok {
			expr = SQLColumnExpr{tableName, sExpr.Expr.String(), expr.columnType}
		} else {
			sExpr.Column.SQLType = expr.Type()
		}

		newSExpr := SelectExpression{
			Column:     sExpr.Column,
			RefColumns: sExpr.RefColumns,
			Expr:       expr,
		}

		newSExprs = append(newSExprs, newSExpr)
	}

	return newSExprs
}
