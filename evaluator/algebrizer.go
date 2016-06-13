package evaluator

import (
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

// Algebrize takes a parsed SQL statement and returns an algebrized form of the query.
func Algebrize(selectStatement parser.SelectStatement, dbName string, schema *schema.Schema) (PlanStage, error) {
	algebrizer := &algebrizer{
		dbName: dbName,
		schema: schema,
	}
	return algebrizer.translateSelectStatement(selectStatement)
}

type algebrizer struct {
	parent                       *algebrizer
	sourceName                   string // the name of the output. This means we need all projected columns to use this as the table name.
	dbName                       string // the default database name.
	schema                       *schema.Schema
	columns                      []*Column        // all the columns in scope.
	tableNames                   []string         // all the table names in scope.
	correlated                   bool             // indicates whether this context is using columns in its parent.
	hasCorrelatedSubquery        bool             // indicates whether this context has a correlated subquery.
	projectedColumns             ProjectedColumns // columns to be projected from this scope.
	resolveProjectedColumnsFirst bool             // indicates whether to resolve a column using the projected columns first or second
}

func (a *algebrizer) fullName(tableName, columnName string) string {
	fn := columnName
	if tableName != "" {
		fn = tableName + "." + fn
	}

	return fn
}

func (a *algebrizer) lookupColumn(tableName, columnName string) (*Column, error) {
	var found *Column
	for _, column := range a.columns {
		if strings.EqualFold(column.Name, columnName) && (tableName == "" || strings.EqualFold(column.Table, tableName)) {
			if found != nil {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NON_UNIQ_ERROR, a.fullName(tableName, columnName), "field list")
			}
			found = column
		}
	}

	if found == nil {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_FIELD_ERROR, a.fullName(tableName, columnName), "field list")
	}

	return found, nil
}

func (a *algebrizer) lookupProjectedColumn(columnName string) (*ProjectedColumn, bool) {
	for _, pc := range a.projectedColumns {
		if strings.EqualFold(pc.Name, columnName) {
			return &pc, true
		}
	}

	return nil, false
}

func (a *algebrizer) resolveColumnExpr(tableName, columnName string) (SQLExpr, error) {

	if a.resolveProjectedColumnsFirst && tableName == "" {
		if expr, ok := a.lookupProjectedColumn(columnName); ok {
			return expr.Expr, nil
		}
	}

	column, err := a.lookupColumn(tableName, columnName)
	if err == nil {
		return SQLColumnExpr{
			tableName:  column.Table,
			columnName: column.Name,
			columnType: schema.ColumnType{
				SQLType:   column.SQLType,
				MongoType: column.MongoType,
			},
		}, nil
	}

	if !a.resolveProjectedColumnsFirst && tableName == "" {
		if expr, ok := a.lookupProjectedColumn(columnName); ok {
			return expr.Expr, nil
		}
	}

	// we didn't find it in the current scope, so we need to search our parent,
	// and let it search its parent, etc...
	if a.parent != nil {
		expr, parentErr := a.parent.resolveColumnExpr(tableName, columnName)
		if parentErr == nil {
			a.correlated = true
			return expr, nil
		}
	}

	return nil, err
}

func (a *algebrizer) registerColumns(columns []*Column) error {
	contains := func(c *Column) bool {
		for _, c2 := range a.columns {
			if strings.EqualFold(c.Name, c2.Name) && strings.EqualFold(c.Table, c2.Table) {
				return true
			}
		}

		return false
	}

	// this ensures that we have no duplicate columns. We have to check duplicates
	// against the existing columns as well as against itself.
	for _, c := range columns {
		if contains(c) {
			return mysqlerrors.Defaultf(mysqlerrors.ER_DUP_FIELDNAME, a.fullName(c.Table, c.Name))
		}
		a.columns = append(a.columns, c)
	}

	return nil
}

// registerTable ensures that we have no duplicate table names or aliases.
func (a *algebrizer) registerTable(tableName string) error {
	for _, registeredName := range a.tableNames {
		if strings.EqualFold(tableName, registeredName) {
			return mysqlerrors.Defaultf(mysqlerrors.ER_NONUNIQ_TABLE, tableName)
		}
	}

	a.tableNames = append(a.tableNames, tableName)

	return nil
}

// isAggFunction returns true if the byte slice e contains the name of an aggregate function and false otherwise.
func (a *algebrizer) isAggFunction(name string) bool {
	switch strings.ToLower(name) {
	case "avg", "sum", "count", "max", "min":
		return true
	default:
		return false
	}
}

func (a *algebrizer) translateGroupBy(groupby parser.GroupBy) ([]SQLExpr, error) {
	var keys []SQLExpr
	for _, g := range groupby {

		key, err := a.translatePossibleColumnRefExpr(g, "group clause")
		if err != nil {
			return nil, err
		}

		afs, err := getAggFunctions(key)
		if len(afs) > 0 {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_GROUP_FIELD, afs[0].String())
		}

		keys = append(keys, key)
	}

	return keys, nil
}

func (a *algebrizer) translateLimit(limit *parser.Limit) (SQLInt, SQLInt, error) {
	var rowcount SQLInt
	var offset SQLInt
	var ok bool
	if limit.Offset != nil {
		eval, err := a.translateExpr(limit.Offset)
		if err != nil {
			return 0, 0, err
		}

		offset, ok = eval.(SQLInt)
		if !ok {
			return 0, 0, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_SPVAR_TYPE_IN_LIMIT)
		}

		if offset < 0 {
			return 0, 0, mysqlerrors.Newf(mysqlerrors.ER_SYNTAX_ERROR, "Offset cannot be negative")
		}
	}

	if limit.Rowcount != nil {
		eval, err := a.translateExpr(limit.Rowcount)
		if err != nil {
			return 0, 0, err
		}

		rowcount, ok = eval.(SQLInt)
		if !ok {
			return 0, 0, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_SPVAR_TYPE_IN_LIMIT)
		}

		if rowcount < 0 {
			return 0, 0, mysqlerrors.Newf(mysqlerrors.ER_SYNTAX_ERROR, "Rowcount cannot be negative")
		}
	}

	return offset, rowcount, nil
}

func (a *algebrizer) translateNamedSelectStatement(selectStatement parser.SelectStatement, sourceName string) (PlanStage, error) {
	algebrizer := &algebrizer{
		dbName:     a.dbName,
		schema:     a.schema,
		sourceName: sourceName,
	}
	return algebrizer.translateSelectStatement(selectStatement)
}

func (a *algebrizer) translateOrderBy(orderby parser.OrderBy) ([]*orderByTerm, error) {
	var terms []*orderByTerm
	for _, o := range orderby {
		term, err := a.translateOrder(o)
		if err != nil {
			return nil, err
		}

		terms = append(terms, term)
	}

	return terms, nil
}

func (a *algebrizer) translateOrder(order *parser.Order) (*orderByTerm, error) {
	ascending := !strings.EqualFold(order.Direction, parser.AST_DESC)
	e, err := a.translatePossibleColumnRefExpr(order.Expr, "order clause")
	if err != nil {
		return nil, err
	}

	return &orderByTerm{
		expr:      e,
		ascending: ascending,
	}, nil
}

func (a *algebrizer) translateSelectStatement(selectStatement parser.SelectStatement) (PlanStage, error) {
	switch typedS := selectStatement.(type) {
	case *parser.Select:
		return a.translateSelect(typedS)
	case *parser.SimpleSelect:
		return a.translateSimpleSelect(typedS)
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_YET, parser.String(selectStatement))
	}
}

func (a *algebrizer) translateSimpleSelect(sel *parser.SimpleSelect) (PlanStage, error) {
	projectedColumns, err := a.translateSelectExprs(sel.SelectExprs)
	if err != nil {
		return nil, err
	}

	return NewProjectStage(NewDualStage(), projectedColumns...), nil
}

func (a *algebrizer) translateSelect(sel *parser.Select) (PlanStage, error) {
	builder := &queryPlanBuilder{
		algebrizer: a,
	}

	// 1. Translate all the tables, subqueries, and joins in the FROM clause.
	// This establishes all the columns which are in scope.
	if sel.From != nil {
		plan, err := a.translateTableExprs(sel.From)
		if err != nil {
			return nil, err
		}

		builder.from = plan

		// TODO: probably add allowed tableNames in order to filter out subquery columns from
		// unrelated tables
		builder.exprCollector = newExpressionCollector()
	}

	// 2. Translate all the other clauses from this scope. We aren't going to create the plan stages
	// yet because the expressions may need to be substituted if a group by exists.
	// Also, in the future, since we are collecting what columns are required at each stage, we'll be
	// able to add a RequiredColumns() function to PlanStage that will let us push down a $project
	// before the first PlanStage we have to execute in memory so as to only pull back the columns
	// we'll actually need.
	if sel.Where != nil {
		err := builder.includeWhere(sel.Where)
		if err != nil {
			return nil, err
		}
	}

	if sel.SelectExprs != nil {
		err := builder.includeSelect(sel.SelectExprs)
		if err != nil {
			return nil, err
		}

		// set projected columns globally because column resolution depends on
		// this list from which GROUP BY and HAVING resolve from it second, and
		// ORDER BY resolves from it first.
		a.projectedColumns = builder.project
	}

	if sel.GroupBy != nil {
		err := builder.includeGroupBy(sel.GroupBy)
		if err != nil {
			return nil, err
		}
	}

	if sel.Having != nil {
		err := builder.includeHaving(sel.Having)
		if err != nil {
			return nil, err
		}
	}

	builder.distinct = sel.Distinct == parser.AST_DISTINCT

	// order by resolves from the projected columns first
	a.resolveProjectedColumnsFirst = true

	if sel.OrderBy != nil {
		err := builder.includeOrderBy(sel.OrderBy)
		if err != nil {
			return nil, err
		}
	}

	if sel.Limit != nil {
		err := builder.includeLimit(sel.Limit)
		if err != nil {
			return nil, err
		}
	}

	builder.hasCorrelatedSubquery = a.hasCorrelatedSubquery

	// 3. Build the stages.
	return builder.build(), nil
}

func (a *algebrizer) translateSelectExprs(selectExprs parser.SelectExprs) (ProjectedColumns, error) {
	var projectedColumns ProjectedColumns
	hasGlobalStar := false
	for _, selectExpr := range selectExprs {
		switch typedE := selectExpr.(type) {

		case *parser.StarExpr:

			// validate tableName if present. Need to have a map of alias -> tableName -> schema
			tableName := ""
			if typedE.TableName != nil {
				tableName = string(typedE.TableName)
			} else {
				hasGlobalStar = true
			}

			for _, column := range a.columns {
				if tableName == "" || strings.EqualFold(tableName, column.Table) {
					projectedColumns = append(projectedColumns, ProjectedColumn{
						Column: &Column{
							Table:     a.sourceName,
							Name:      column.Name,
							SQLType:   column.SQLType,
							MongoType: column.MongoType,
						},
						Expr: SQLColumnExpr{
							tableName:  column.Table,
							columnName: column.Name,
							columnType: schema.ColumnType{
								SQLType:   column.SQLType,
								MongoType: column.MongoType,
							},
						},
					})
				}
			}

		case *parser.NonStarExpr:

			translatedExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			projectedColumn := ProjectedColumn{
				Expr: translatedExpr,
				Column: &Column{
					Table:     a.sourceName,
					MongoType: schema.MongoNone,
					SQLType:   translatedExpr.Type(),
				},
			}

			if sqlCol, ok := translatedExpr.(SQLColumnExpr); ok {
				projectedColumn.MongoType = sqlCol.columnType.MongoType
			}

			if sqlCol, ok := typedE.Expr.(*parser.ColName); ok {
				projectedColumn.Name = string(sqlCol.Name)
			}

			if typedE.As != nil {
				projectedColumn.Name = string(typedE.As)
			} else if projectedColumn.Name == "" {
				projectedColumn.Name = parser.String(typedE)
			}

			projectedColumns = append(projectedColumns, projectedColumn)
		}
	}

	if hasGlobalStar && len(selectExprs) > 1 {
		return nil, mysqlerrors.Newf(mysqlerrors.ER_SYNTAX_ERROR, "Cannot have a '*' in conjunction with any other columns")
	}

	return projectedColumns, nil
}

func (a *algebrizer) translateTableExprs(tableExprs parser.TableExprs) (PlanStage, error) {

	var plan PlanStage
	for i, tableExpr := range tableExprs {
		temp, err := a.translateTableExpr(tableExpr)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			plan = temp
		} else {
			plan = &JoinStage{
				left:  plan,
				right: temp,
				kind:  CrossJoin,
			}
		}
	}

	return plan, nil
}

func (a *algebrizer) translateTableExpr(tableExpr parser.TableExpr) (PlanStage, error) {
	switch typedT := tableExpr.(type) {
	case *parser.AliasedTableExpr:
		return a.translateSimpleTableExpr(typedT.Expr, string(typedT.As))
	case *parser.ParenTableExpr:
		return a.translateTableExpr(typedT.Expr)
	case parser.SimpleTableExpr:
		return a.translateSimpleTableExpr(typedT, "")
	case *parser.JoinTableExpr:
		left, err := a.translateTableExpr(typedT.LeftExpr)
		if err != nil {
			return nil, err
		}
		right, err := a.translateTableExpr(typedT.RightExpr)
		if err != nil {
			return nil, err
		}

		var predicate SQLExpr
		if typedT.On != nil {
			predicate, err = a.translateExpr(typedT.On)
			if err != nil {
				return nil, err
			}
		} else {
			predicate = SQLBool(true)
		}

		return NewJoinStage(JoinKind(typedT.Join), left, right, predicate), nil
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_YET, parser.String(tableExpr))
	}
}

func (a *algebrizer) translateSimpleTableExpr(tableExpr parser.SimpleTableExpr, aliasName string) (PlanStage, error) {
	switch typedT := tableExpr.(type) {
	case *parser.TableName:
		tableName := string(typedT.Name)
		if aliasName == "" {
			aliasName = tableName
		}

		dbName := strings.ToLower(string(typedT.Qualifier))
		if dbName == "" {
			dbName = a.dbName
		}

		var plan PlanStage
		var err error
		if strings.EqualFold(tableName, "DUAL") {
			plan = NewDualStage()
		} else if strings.EqualFold(dbName, informationSchemaDatabase) {
			plan = NewSchemaDataSourceStage(a.schema, tableName, aliasName)
		} else {
			plan, err = NewMongoSourceStage(a.schema, dbName, tableName, aliasName)
			if err != nil {
				return nil, err
			}
		}

		err = a.registerTable(aliasName)
		if err != nil {
			return nil, err
		}

		err = a.registerColumns(plan.Columns())
		if err != nil {
			return nil, err
		}

		return plan, nil
	case *parser.Subquery:

		if aliasName == "" {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_DERIVED_MUST_HAVE_ALIAS)
		}

		plan, err := a.translateNamedSelectStatement(typedT.Select, aliasName)
		if err != nil {
			return nil, err
		}

		err = a.registerColumns(plan.Columns())
		if err != nil {
			return nil, err
		}

		return plan, nil
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_YET, parser.String(tableExpr))
	}
}

func (a *algebrizer) translateSubqueryExpr(expr *parser.Subquery) (*SQLSubqueryExpr, error) {
	subqueryAlgebrizer := &algebrizer{
		parent: a,
		dbName: a.dbName,
		schema: a.schema,
	}

	plan, err := subqueryAlgebrizer.translateSelectStatement(expr.Select)
	if err != nil {
		return nil, err
	}

	if subqueryAlgebrizer.correlated {
		a.hasCorrelatedSubquery = true
	}

	return &SQLSubqueryExpr{
		plan:       plan,
		correlated: subqueryAlgebrizer.correlated,
	}, nil
}

func (a *algebrizer) translateExpr(expr parser.Expr) (SQLExpr, error) {
	switch typedE := expr.(type) {
	case *parser.AndExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
		if err != nil {
			return nil, err
		}

		return &SQLAndExpr{left, right}, nil
	case *parser.BinaryExpr:
		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, true)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case parser.AST_PLUS:
			return &SQLAddExpr{left, right}, nil
		case parser.AST_MINUS:
			return &SQLSubtractExpr{left, right}, nil
		case parser.AST_MULT:
			return &SQLMultiplyExpr{left, right}, nil
		case parser.AST_DIV:
			return &SQLDivideExpr{left, right}, nil
		case parser.AST_IDIV:
			return &SQLIDivideExpr{left, right}, nil
		case parser.AST_MOD:
			return &SQLModExpr{left, right}, nil
		default:
			return nil, mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "No support for binary operator '%v'", typedE.Operator)
		}
	case *parser.CaseExpr:
		return a.translateCaseExpr(typedE)
	case *parser.ColName:
		tableName := ""
		if typedE.Qualifier != nil {
			tableName = string(typedE.Qualifier)
		}

		columnName := string(typedE.Name)

		if tableName == "" && strings.HasPrefix(columnName, "@") {
			variableName := strings.TrimPrefix(columnName, "@")
			variableType := UserDefinedVariable
			if strings.HasPrefix(variableName, "@") {
				variableName = strings.TrimPrefix(variableName, "@")
				variableType = SystemVariable
			}
			return &SQLVariableExpr{
				Name:         variableName,
				VariableType: variableType,
			}, nil
		}

		return a.resolveColumnExpr(tableName, columnName)
	case *parser.ComparisonExpr:

		reconcile := typedE.Operator != parser.AST_LIKE

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, reconcile)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case parser.AST_EQ:
			return &SQLEqualsExpr{left, right}, nil
		case parser.AST_LT:
			return &SQLLessThanExpr{left, right}, nil
		case parser.AST_GT:
			return &SQLGreaterThanExpr{left, right}, nil
		case parser.AST_LE:
			return &SQLLessThanOrEqualExpr{left, right}, nil
		case parser.AST_GE:
			return &SQLGreaterThanOrEqualExpr{left, right}, nil
		case parser.AST_NE:
			return &SQLNotEqualsExpr{left, right}, nil
		case parser.AST_LIKE:
			// TODO: Might not want to reconcile expressions in this one...
			return &SQLLikeExpr{left, right}, nil
		case parser.AST_IN:
			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLInExpr{left, right}, nil
		case parser.AST_NOT_IN:
			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLNotExpr{&SQLInExpr{left, right}}, nil
		default:
			return nil, mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "No support for operator '%v'", typedE.Operator)
		}
	case *parser.CtorExpr:
		// TODO: currently only supports single argument constructors
		strVal, ok := typedE.Exprs[0].(parser.StrVal)
		if !ok {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_ILLEGAL_VALUE_FOR_TYPE, "parameter", parser.String(typedE.Exprs[0]))
		}

		arg := string(strVal)

		switch typedE.Name {
		case parser.AST_DATE:
			return NewSQLValue(arg, schema.SQLDate, schema.MongoNone)
		case parser.AST_DATETIME:
			return NewSQLValue(arg, schema.SQLTimestamp, schema.MongoNone)
		case parser.AST_TIMESTAMP:
			return NewSQLValue(arg, schema.SQLTimestamp, schema.MongoNone)
		default:
			return nil, mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "No support for constructor '%v'", string(typedE.Name))
		}
	case *parser.ExistsExpr:
		subquery, err := a.translateSubqueryExpr(typedE.Subquery)
		if err != nil {
			return nil, err
		}
		return &SQLExistsExpr{subquery}, nil
	case *parser.FalseVal:
		return SQLFalse, nil
	case *parser.FuncExpr:
		return a.translateFuncExpr(typedE)
	case *parser.NotExpr:
		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		return &SQLNotExpr{child}, nil
	case *parser.NullCheck:
		val, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		var child SQLExpr = &SQLNullCmpExpr{val}
		if typedE.Operator == parser.AST_IS_NOT_NULL {
			child = &SQLNotExpr{child}
		}

		return child, nil
	case *parser.NullVal:
		return SQLNull, nil
	case parser.NumVal:
		// try to parse as int first
		if i, err := strconv.ParseInt(parser.String(expr), 10, 64); err == nil {
			return SQLInt(i), nil
		}

		// if it's not a valid int, try parsing as float instead
		f, err := strconv.ParseFloat(parser.String(expr), 64)
		if err != nil {
			return nil, err
		}

		return SQLFloat(f), nil
	case *parser.OrExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
		if err != nil {
			return nil, err
		}

		return &SQLOrExpr{left, right}, nil
	case *parser.ParenBoolExpr:
		return a.translateExpr(typedE.Expr)
	case *parser.RangeCond:

		from, err := a.translateExpr(typedE.From)
		if err != nil {
			return nil, err
		}

		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		to, err := a.translateExpr(typedE.To)
		if err != nil {
			return nil, err
		}

		left, from, err = reconcileSQLExprs(left, from)
		if err != nil {
			return nil, err
		}

		lower := &SQLGreaterThanOrEqualExpr{left, from}

		left, to, err = reconcileSQLExprs(left, to)
		if err != nil {
			return nil, err
		}

		upper := &SQLLessThanOrEqualExpr{left, to}

		var m SQLExpr = &SQLAndExpr{lower, upper}

		if typedE.Operator == parser.AST_NOT_BETWEEN {
			return &SQLNotExpr{m}, nil
		}

		return m, nil
	case parser.StrVal:
		return SQLVarchar(string(typedE)), nil
	case *parser.Subquery:
		return a.translateSubqueryExpr(typedE)
	case *parser.TrueVal:
		return SQLTrue, nil
	case *parser.UnaryExpr:

		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case parser.AST_UMINUS:
			return &SQLUnaryMinusExpr{child}, nil
		case parser.AST_TILDA:
			return &SQLUnaryTildeExpr{child}, nil
		}

		return nil, mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "No support for operator '%v'", typedE.Operator)

	case parser.ValTuple:

		var exprs []SQLExpr

		for _, e := range typedE {
			newExpr, err := a.translateExpr(e)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, newExpr)
		}

		if len(exprs) == 1 {
			// TODO: remove this check from ast_factories.go and add test.
			return exprs[0], nil
		}

		return &SQLTupleExpr{exprs}, nil
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "No support for '%v'", parser.String(typedE))
	}
}

func (a *algebrizer) translatePossibleColumnRefExpr(expr parser.Expr, clause string) (SQLExpr, error) {
	if numVal, ok := expr.(parser.NumVal); ok {
		n, err := strconv.ParseInt(parser.String(numVal), 10, 64)
		if err != nil {
			return nil, err
		}

		if int(n) > len(a.projectedColumns) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_FIELD_ERROR, strconv.Itoa(int(n)), clause)
		}

		if n >= 0 {
			return a.projectedColumns[n-1].Expr, nil
		}
	}

	return a.translateExpr(expr)
}

func (a *algebrizer) translateLeftRightExprs(left, right parser.Expr, reconcile bool) (SQLExpr, SQLExpr, error) {
	leftEval, err := a.translateExpr(left)
	if err != nil {
		return nil, nil, err
	}

	rightEval, err := a.translateExpr(right)
	if err != nil {
		return nil, nil, err
	}

	if reconcile {
		leftEval, rightEval, err = reconcileSQLExprs(leftEval, rightEval)
	}

	return leftEval, rightEval, err
}

func (a *algebrizer) translateCaseExpr(expr *parser.CaseExpr) (SQLExpr, error) {
	// There are two kinds of case expression.
	//
	// 1. For simple case expressions, we create an equality matcher that compares
	// the expression against each value in the list of cases.
	//
	// 2. For searched case expressions, we create a matcher based on the boolean
	// expression in each when condition.

	var e SQLExpr
	var err error

	if expr.Expr != nil {
		e, err = a.translateExpr(expr.Expr)
		if err != nil {
			return nil, err
		}
	}

	var conditions []caseCondition
	var matcher SQLExpr

	for _, when := range expr.Whens {

		// searched case
		if expr.Expr == nil {
			matcher, err = a.translateExpr(when.Cond)
			if err != nil {
				return nil, err
			}
		} else {
			// TODO: support simple case in parser
			c, err := a.translateExpr(when.Cond)
			if err != nil {
				return nil, err
			}

			matcher = &SQLEqualsExpr{e, c}
		}

		then, err := a.translateExpr(when.Val)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, caseCondition{matcher, then})
	}

	var elseValue SQLExpr
	if expr.Else == nil {
		elseValue = SQLNull
	} else if elseValue, err = a.translateExpr(expr.Else); err != nil {
		return nil, err
	}

	value := &SQLCaseExpr{
		elseValue:      elseValue,
		caseConditions: conditions,
	}

	// TODO: You cannot specify the literal NULL for every return expr
	// and the else expr.
	return value, nil
}

func (a *algebrizer) translateFuncExpr(expr *parser.FuncExpr) (SQLExpr, error) {

	exprs := []SQLExpr{}
	name := string(expr.Name)

	if a.isAggFunction(name) {

		if len(expr.Exprs) != 1 {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_ARGUMENTS, name)
		}

		e := expr.Exprs[0]

		switch typedE := e.(type) {
		case *parser.StarExpr:

			if name != "count" {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_ARGUMENTS, name)
			}

			if expr.Distinct {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_ARGUMENTS, name)
			}

			exprs = append(exprs, SQLVarchar("*"))

		case *parser.NonStarExpr:

			sqlExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, sqlExpr)
		default:
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_YET, parser.String(e))
		}

		return &SQLAggFunctionExpr{name, expr.Distinct, exprs}, nil
	}

	for _, e := range expr.Exprs {

		switch typedE := e.(type) {
		case *parser.StarExpr:
			if !strings.EqualFold(name, "count") {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_ARGUMENTS, name)
			}
		case *parser.NonStarExpr:
			sqlExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, sqlExpr)

			if typedE.As != nil {
				as := string(typedE.As)
				switch strings.ToLower(as) {
				case "cast":
					exprs = append(exprs, SQLVarchar(as))
				default:
					return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_YET, parser.String(e))
				}
			}
		default:
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_YET, parser.String(expr))
		}

	}

	return &SQLScalarFunctionExpr{name, exprs}, nil
}
