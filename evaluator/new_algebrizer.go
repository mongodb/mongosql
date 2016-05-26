package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
)

// Algebrize takes a parsed SQL statement and returns an algebrized form of the query.
func Algebrize(selectStatement sqlparser.SelectStatement, dbName string, schema *schema.Schema) (PlanStage, error) {
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
	columns                      []*Column         // all the columns in scope.
	correlated                   bool              // indicates whether this context is using columns in its parent.
	projectedColumns             SelectExpressions // columns to be projected from this scope.
	resolveProjectedColumnsFirst bool              // indicates whether to resolve a column using the projected columns first or second
}

func (a *algebrizer) lookupColumn(tableName, columnName string) (*Column, error) {
	var found *Column
	for _, column := range a.columns {
		if strings.EqualFold(column.Name, columnName) && (tableName == "" || strings.EqualFold(column.Table, tableName)) {
			if found != nil {
				if tableName != "" {
					return nil, fmt.Errorf("duplicate column name %q in table %q", columnName, tableName)
				}

				return nil, fmt.Errorf("column %q in the field list is ambiguous", columnName)
			}
			found = column
		}
	}

	if found == nil {
		if tableName == "" {
			return nil, fmt.Errorf("unknown column %q", columnName)
		}

		return nil, fmt.Errorf("unknown column %q in table %q", columnName, tableName)
	}

	return found, nil
}

func (a *algebrizer) lookupProjectedColumnExpr(columnName string) (*SelectExpression, bool) {
	for _, pc := range a.projectedColumns {
		if strings.EqualFold(pc.Name, columnName) {
			return &pc, true
		}
	}

	return nil, false
}

func (a *algebrizer) resolveColumnExpr(expr sqlparser.Expr, clause string) (SQLExpr, error) {
	if numVal, ok := expr.(sqlparser.NumVal); ok {
		n, err := strconv.ParseInt(sqlparser.String(numVal), 10, 64)
		if err != nil {
			return nil, err
		}

		if int(n) > len(a.projectedColumns) {
			return nil, fmt.Errorf("unknown column \"%v\" in %s", n, clause)
		}

		if n >= 0 {
			return a.projectedColumns[n-1].Expr, nil
		}
	}

	return a.translateExpr(expr)
}

func (a *algebrizer) registerColumns(columns []*Column) {
	a.columns = append(a.columns, columns...)
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

func (a *algebrizer) translateGroupBy(groupby sqlparser.GroupBy) ([]SQLExpr, error) {
	var keys []SQLExpr
	for _, g := range groupby {

		key, err := a.resolveColumnExpr(g, "group clause")
		if err != nil {
			return nil, err
		}

		afs, err := getAggFunctions(key)
		if len(afs) > 0 {
			return nil, fmt.Errorf("can't group on %q", afs[0].String())
		}

		keys = append(keys, key)
	}

	return keys, nil
}

func (a *algebrizer) translateLimit(limit *sqlparser.Limit) (SQLInt, SQLInt, error) {
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
			return 0, 0, fmt.Errorf("limit offset must be an integer")
		}

		if offset < 0 {
			return 0, 0, fmt.Errorf("limit offset cannot be negative")
		}
	}

	if limit.Rowcount != nil {
		eval, err := a.translateExpr(limit.Rowcount)
		if err != nil {
			return 0, 0, err
		}

		rowcount, ok = eval.(SQLInt)
		if !ok {
			return 0, 0, fmt.Errorf("limit rowcount must be an integer")
		}

		if rowcount < 0 {
			return 0, 0, fmt.Errorf("limit rowcount cannot be negative")
		}
	}

	return offset, rowcount, nil
}

func (a *algebrizer) translateNamedSelectStatement(selectStatement sqlparser.SelectStatement, sourceName string) (PlanStage, error) {
	algebrizer := &algebrizer{
		dbName:     a.dbName,
		schema:     a.schema,
		sourceName: sourceName,
	}
	return algebrizer.translateSelectStatement(selectStatement)
}

func (a *algebrizer) translateOrderBy(orderby sqlparser.OrderBy) ([]*orderByTerm, error) {
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

func (a *algebrizer) translateOrder(order *sqlparser.Order) (*orderByTerm, error) {
	ascending := !strings.EqualFold(order.Direction, sqlparser.AST_DESC)
	e, err := a.resolveColumnExpr(order.Expr, "order clause")
	if err != nil {
		return nil, err
	}

	return &orderByTerm{
		expr:      e,
		ascending: ascending,
	}, nil
}

func (a *algebrizer) translateSelectStatement(selectStatement sqlparser.SelectStatement) (PlanStage, error) {
	switch typedS := selectStatement.(type) {
	case *sqlparser.Select:
		return a.translateSelect(typedS)
	case *sqlparser.SimpleSelect:
		return a.translateSimpleSelect(typedS)
	default:
		return nil, fmt.Errorf("no support for %T", selectStatement)
	}
}

func (a *algebrizer) translateSimpleSelect(sel *sqlparser.SimpleSelect) (PlanStage, error) {
	projectedColumns, err := a.translateSelectExprs(sel.SelectExprs)
	if err != nil {
		return nil, err
	}

	return NewProjectStage(NewDualStage(), projectedColumns...), nil
}

func (a *algebrizer) translateSelect(sel *sqlparser.Select) (PlanStage, error) {
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
		// this list, where group by and having resolve from this list second, and
		// order by resolves from it first.
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

	builder.distinct = sel.Distinct == sqlparser.AST_DISTINCT

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

	// 3. Build the stages.
	return builder.build(), nil
}

func (a *algebrizer) translateSelectExprs(selectExprs sqlparser.SelectExprs) (SelectExpressions, error) {
	var projectedColumns SelectExpressions
	hasGlobalStar := false
	for _, selectExpr := range selectExprs {
		switch typedE := selectExpr.(type) {

		case *sqlparser.StarExpr:

			// validate tableName if present. Need to have a map of alias -> tableName -> schema
			tableName := ""
			if typedE.TableName != nil {
				tableName = string(typedE.TableName)
			} else {
				hasGlobalStar = true
			}

			for _, column := range a.columns {
				if tableName == "" || strings.EqualFold(tableName, column.Table) {
					projectedColumns = append(projectedColumns, SelectExpression{
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

		case *sqlparser.NonStarExpr:

			translatedExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			projectedColumn := SelectExpression{
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

			if sqlCol, ok := typedE.Expr.(*sqlparser.ColName); ok {
				projectedColumn.Name = string(sqlCol.Name)
			}

			if typedE.As != nil {
				projectedColumn.Name = string(typedE.As)
			} else if projectedColumn.Name == "" {
				projectedColumn.Name = sqlparser.String(typedE)
			}

			projectedColumns = append(projectedColumns, projectedColumn)
		}
	}

	if hasGlobalStar && len(selectExprs) > 1 {
		return nil, fmt.Errorf("cannot have a global * in the field list in conjunction with any other columns")
	}

	return projectedColumns, nil
}

func (a *algebrizer) translateTableExprs(tableExprs sqlparser.TableExprs) (PlanStage, error) {

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

func (a *algebrizer) translateTableExpr(tableExpr sqlparser.TableExpr) (PlanStage, error) {
	switch typedT := tableExpr.(type) {
	case *sqlparser.AliasedTableExpr:
		return a.translateSimpleTableExpr(typedT.Expr, string(typedT.As))
	case *sqlparser.ParenTableExpr:
		return a.translateTableExpr(typedT.Expr)
	case sqlparser.SimpleTableExpr:
		return a.translateSimpleTableExpr(typedT, "")
	case *sqlparser.JoinTableExpr:
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
		return nil, fmt.Errorf("no support for %T", tableExpr)
	}
}

func (a *algebrizer) translateSimpleTableExpr(tableExpr sqlparser.SimpleTableExpr, aliasName string) (PlanStage, error) {
	switch typedT := tableExpr.(type) {
	case *sqlparser.TableName:
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
		if strings.EqualFold(dbName, informationSchemaDatabase) {
			plan = NewSchemaDataSourceStage(a.schema, tableName, aliasName)
		} else {
			plan, err = NewMongoSourceStage(a.schema, dbName, tableName, aliasName)
			if err != nil {
				return nil, err
			}
		}

		a.registerColumns(plan.OpFields())

		return plan, nil
	case *sqlparser.Subquery:

		if aliasName == "" {
			return nil, fmt.Errorf("every derived table must have it's own alias")
		}

		plan, err := a.translateNamedSelectStatement(typedT.Select, aliasName)
		if err != nil {
			return nil, err
		}

		a.registerColumns(plan.OpFields())

		return plan, nil
	default:
		return nil, fmt.Errorf("no support for %T", tableExpr)
	}
}

func (a *algebrizer) translateExpr(expr sqlparser.Expr) (SQLExpr, error) {
	switch typedE := expr.(type) {
	case *sqlparser.AndExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
		if err != nil {
			return nil, err
		}

		return &SQLAndExpr{left, right}, nil
	case *sqlparser.BinaryExpr:
		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, true)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case sqlparser.AST_PLUS:
			return &SQLAddExpr{left, right}, nil
		case sqlparser.AST_MINUS:
			return &SQLSubtractExpr{left, right}, nil
		case sqlparser.AST_MULT:
			return &SQLMultiplyExpr{left, right}, nil
		case sqlparser.AST_DIV:
			return &SQLDivideExpr{left, right}, nil
		default:
			return nil, fmt.Errorf("no support for binary operator '%v'", typedE.Operator)
		}
	case *sqlparser.CaseExpr:
		return a.translateCaseExpr(typedE)
	case *sqlparser.ColName:
		tableName := ""
		if typedE.Qualifier != nil {
			tableName = string(typedE.Qualifier)
		}

		columnName := string(typedE.Name)

		if a.resolveProjectedColumnsFirst && tableName == "" {
			if expr, ok := a.lookupProjectedColumnExpr(columnName); ok {
				return expr.Expr, nil
			}
		}

		column, err := a.lookupColumn(tableName, columnName)
		if err != nil {
			if !a.resolveProjectedColumnsFirst && tableName == "" {
				if expr, ok := a.lookupProjectedColumnExpr(columnName); ok {
					return expr.Expr, nil
				}
			}

			if a.parent != nil {
				column, err = a.parent.lookupColumn(tableName, columnName)
				if err != nil {
					return nil, err
				}

				a.correlated = true
			} else {
				return nil, err
			}
		}

		return SQLColumnExpr{
			tableName:  column.Table,
			columnName: column.Name,
			columnType: schema.ColumnType{
				SQLType:   column.SQLType,
				MongoType: column.MongoType,
			},
		}, nil

	case *sqlparser.ComparisonExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, true)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case sqlparser.AST_EQ:
			return &SQLEqualsExpr{left, right}, nil
		case sqlparser.AST_LT:
			return &SQLLessThanExpr{left, right}, nil
		case sqlparser.AST_GT:
			return &SQLGreaterThanExpr{left, right}, nil
		case sqlparser.AST_LE:
			return &SQLLessThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_GE:
			return &SQLGreaterThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_NE:
			return &SQLNotEqualsExpr{left, right}, nil
		case sqlparser.AST_LIKE:
			// TODO: Might not want to reconcile expressions in this one...
			return &SQLLikeExpr{left, right}, nil
		case sqlparser.AST_IN:
			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLInExpr{left, right}, nil
		case sqlparser.AST_NOT_IN:
			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLNotExpr{&SQLInExpr{left, right}}, nil
		default:
			return nil, fmt.Errorf("no support for operator %q", typedE.Operator)
		}
	case *sqlparser.CtorExpr:
		// TODO: SQLCtorExpr contains reference to parse tree...

		// ctor := &SQLCtorExpr{Name: typedE.Name, Args: expr.Exprs}
		// return ctor.Evaluate(nil)
		return nil, fmt.Errorf("ctor expression not supported yet")
	case *sqlparser.ExistsExpr:
		subquery, err := a.translateSubqueryExpr(typedE.Subquery)
		if err != nil {
			return nil, err
		}
		return &SQLExistsExpr{nil, subquery}, nil
	case *sqlparser.FuncExpr:
		return a.translateFuncExpr(typedE)
	case *sqlparser.NotExpr:
		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		return &SQLNotExpr{child}, nil
	case *sqlparser.NullCheck:
		val, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		var child SQLExpr = &SQLNullCmpExpr{val}
		if typedE.Operator == sqlparser.AST_IS_NOT_NULL {
			child = &SQLNotExpr{child}
		}

		return child, nil
	case *sqlparser.NullVal:
		return SQLNull, nil
	case sqlparser.NumVal:
		// try to parse as int first
		if i, err := strconv.ParseInt(sqlparser.String(expr), 10, 64); err == nil {
			return SQLInt(i), nil
		}

		// if it's not a valid int, try parsing as float instead
		f, err := strconv.ParseFloat(sqlparser.String(expr), 64)
		if err != nil {
			return nil, err
		}

		return SQLFloat(f), nil
	case *sqlparser.OrExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
		if err != nil {
			return nil, err
		}

		return &SQLOrExpr{left, right}, nil
	case *sqlparser.ParenBoolExpr:
		return a.translateExpr(typedE.Expr)
	case *sqlparser.RangeCond:

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

		if typedE.Operator == sqlparser.AST_NOT_BETWEEN {
			return &SQLNotExpr{m}, nil
		}

		return m, nil
	case sqlparser.StrVal:
		return SQLVarchar(string(typedE)), nil
	case *sqlparser.Subquery:
		return a.translateSubqueryExpr(typedE)
	case *sqlparser.UnaryExpr:

		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case sqlparser.AST_UMINUS:
			return &SQLUnaryMinusExpr{child}, nil
		case sqlparser.AST_TILDA:
			return &SQLUnaryTildeExpr{child}, nil
		}

		return nil, fmt.Errorf("invalid unary operator - '%v'", string(typedE.Operator))

	case sqlparser.ValTuple:

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
		return nil, fmt.Errorf("no support for %T", expr)
	}
}

func (a *algebrizer) translateLeftRightExprs(left sqlparser.Expr, right sqlparser.Expr, reconcile bool) (SQLExpr, SQLExpr, error) {
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

func (a *algebrizer) translateCaseExpr(expr *sqlparser.CaseExpr) (SQLExpr, error) {
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

func (a *algebrizer) translateFuncExpr(expr *sqlparser.FuncExpr) (SQLExpr, error) {

	exprs := []SQLExpr{}
	name := string(expr.Name)

	if a.isAggFunction(name) {

		if len(expr.Exprs) != 1 {
			return nil, fmt.Errorf("aggregate function cannot contain tuples")
		}

		e := expr.Exprs[0]

		switch typedE := e.(type) {
		case *sqlparser.StarExpr:

			if name != "count" {
				return nil, fmt.Errorf(`%q aggregate function can not contain "*"`, name)
			}

			if expr.Distinct {
				return nil, fmt.Errorf(`count aggregate function can not have distinct "*"`)
			}

			exprs = append(exprs, SQLVarchar("*"))

		case *sqlparser.NonStarExpr:

			sqlExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, sqlExpr)
		default:
			return nil, fmt.Errorf("no support for %T", e)
		}

		return &SQLAggFunctionExpr{name, expr.Distinct, exprs}, nil
	}

	for _, e := range expr.Exprs {

		switch typedE := e.(type) {
		case *sqlparser.StarExpr:
			if !strings.EqualFold(name, "count") {
				return nil, fmt.Errorf(`argument to %q cannot contain "*"`, name)
			}
		case *sqlparser.NonStarExpr:
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
					return nil, fmt.Errorf("no support for %T", e)
				}
			}
		default:
			return nil, fmt.Errorf("no support for %T", expr)
		}

	}

	return &SQLScalarFunctionExpr{name, exprs}, nil
}

func (a *algebrizer) translateSubqueryExpr(expr *sqlparser.Subquery) (*SQLSubqueryExpr, error) {
	subqueryAlgebrizer := &algebrizer{
		parent: a,
		dbName: a.dbName,
		schema: a.schema,
	}

	plan, err := subqueryAlgebrizer.translateSelectStatement(expr.Select)
	if err != nil {
		return nil, err
	}

	return &SQLSubqueryExpr{
		plan:       plan,
		correlated: subqueryAlgebrizer.correlated,
	}, nil
}

type queryPlanBuilder struct {
	algebrizer    *algebrizer
	exprCollector *expressionCollector

	from     PlanStage
	where    SQLExpr
	groupBy  []SQLExpr
	having   SQLExpr
	distinct bool
	orderBy  []*orderByTerm
	project  SelectExpressions
	offset   int64
	rowcount int64
}

func (b *queryPlanBuilder) build() PlanStage {

	plan := b.buildWhere(b.from)
	plan = b.buildGroupBy(plan)
	plan = b.buildHaving(plan)
	plan = b.buildDistinct(plan)
	plan = b.buildOrderBy(plan)
	plan = b.buildLimit(plan)
	plan = b.buildProject(plan)

	return plan
}

func (b *queryPlanBuilder) buildDistinct(source PlanStage) PlanStage {
	plan := source
	if b.distinct {
		var keys SelectExpressions
		for _, c := range b.project {
			pc := projectedColumnFromExpr(c.Expr)
			keys = append(keys, *pc)
		}

		// projectedAggregates will include any column that is not an aggregate function.
		// as well as all the keys.
		projectedAggregates := keys
		for _, e := range b.exprCollector.allNonAggReferencedColumns.copyExprs() {
			pc := projectedColumnFromExpr(e)
			projectedAggregates = append(projectedAggregates, *pc)
		}

		plan = NewGroupByStage(plan, keys, projectedAggregates.Unique())

		// now we must replace all the project values with columns as
		// any that weren't already a column have now been computed.
		var projectedColumns SelectExpressions
		for i, pc := range b.project {
			b.exprCollector.Remove(pc.Expr)
			newExpr := SQLColumnExpr{
				tableName:  keys[i].Table,
				columnName: keys[i].Name,
				columnType: schema.ColumnType{
					SQLType:   keys[i].SQLType,
					MongoType: keys[i].MongoType,
				},
			}
			projectedColumns = append(projectedColumns, SelectExpression{
				Column: pc.Column,
				Expr:   newExpr,
			})
			b.exprCollector.Visit(newExpr)
		}

		b.project = projectedColumns
	}

	return plan
}

func (b *queryPlanBuilder) buildGroupBy(source PlanStage) PlanStage {
	plan := source
	if len(b.groupBy) > 0 || len(b.exprCollector.allAggFunctions.exprs) > 0 {
		var keys SelectExpressions
		for _, e := range b.groupBy {
			pc := projectedColumnFromExpr(e)
			keys = append(keys, *pc)
		}

		// do this now so it doesn't throw off the b.exprCollector.allNonAggReferencedColumns.
		b.exprCollector.RemoveAll(b.groupBy)

		// projectedAggregates will include all the aggregates as well
		// as any column that is not an aggregate function.
		var projectedAggregates SelectExpressions
		for _, e := range b.exprCollector.allNonAggReferencedColumns.copyExprs() {
			pc := projectedColumnFromExpr(e)
			projectedAggregates = append(projectedAggregates, *pc)
		}

		for _, e := range b.exprCollector.allAggFunctions.copyExprs() {
			pc := projectedColumnFromExpr(e)
			projectedAggregates = append(projectedAggregates, *pc)
		}

		plan = NewGroupByStage(plan, keys, projectedAggregates.Unique())

		// replace aggregation expressions with columns coming out of the GroupByStage
		// because they have already been aggregated and are now just columns.
		b.replaceAggFunctions()
	}

	return plan
}

func (b *queryPlanBuilder) buildHaving(source PlanStage) PlanStage {
	if b.having != nil {
		b.exprCollector.Remove(b.having)
		return NewFilterStage(source, b.having)
	}

	return source
}

func (b *queryPlanBuilder) buildLimit(source PlanStage) PlanStage {
	if b.offset > 0 || b.rowcount > 0 {
		return NewLimitStage(source, b.offset, b.rowcount)
	}

	return source
}

func (b *queryPlanBuilder) buildOrderBy(source PlanStage) PlanStage {
	if len(b.orderBy) > 0 {
		for _, obt := range b.orderBy {
			b.exprCollector.Remove(obt.expr)
		}

		return NewOrderByStage(source, b.orderBy...)
	}

	return source
}

func (b *queryPlanBuilder) buildProject(source PlanStage) PlanStage {
	if len(b.project) > 0 {
		return NewProjectStage(source, b.project...)
	}

	return source
}

func (b *queryPlanBuilder) buildWhere(source PlanStage) PlanStage {
	if b.where != nil {
		b.exprCollector.Remove(b.where)
		return NewFilterStage(source, b.where)
	}

	return source
}

func (b *queryPlanBuilder) replaceAggFunctions() error {

	// since we are replacing aggregates (which likely include columns) with other columns,
	// we need to update the exprCollection with the new information so that it continues
	// to be correct. Therefore, we'll be removing the old expressions and adding in
	// new ones.

	if len(b.project) > 0 {

		var projectedColumns SelectExpressions
		for _, pc := range b.project {
			b.exprCollector.Remove(pc.Expr)
			replaced, err := replaceAggFunctionsWithColumns("", pc.Expr)
			if err != nil {
				return err
			}
			b.exprCollector.Visit(replaced)

			projectedColumns = append(projectedColumns, SelectExpression{
				Expr:   replaced,
				Column: pc.Column,
			})
		}
		b.project = projectedColumns
	}

	if b.having != nil {
		b.exprCollector.Remove(b.having)
		having, err := replaceAggFunctionsWithColumns("", b.having)
		if err != nil {
			return err
		}
		b.exprCollector.Visit(having)

		b.having = having
	}

	if len(b.orderBy) > 0 {
		var orderBy []*orderByTerm
		for _, obt := range b.orderBy {
			b.exprCollector.Remove(obt.expr)
			replaced, err := replaceAggFunctionsWithColumns("", obt.expr)
			if err != nil {
				return err
			}
			b.exprCollector.Visit(replaced)

			orderBy = append(orderBy, &orderByTerm{
				ascending: obt.ascending,
				expr:      replaced,
			})
		}

		b.orderBy = orderBy
	}

	return nil
}

func (b *queryPlanBuilder) includeGroupBy(groupBy sqlparser.GroupBy) error {
	keys, err := b.algebrizer.translateGroupBy(groupBy)
	if err != nil {
		return err
	}

	b.exprCollector.VisitAll(keys)
	b.groupBy = keys
	return nil
}

func (b *queryPlanBuilder) includeHaving(having *sqlparser.Where) error {
	pred, err := b.algebrizer.translateExpr(having.Expr)
	if err != nil {
		return err
	}

	b.exprCollector.Visit(pred)
	b.having = pred
	return nil
}

func (b *queryPlanBuilder) includeLimit(limit *sqlparser.Limit) error {
	offset, rowcount, err := b.algebrizer.translateLimit(limit)
	if err != nil {
		return err
	}

	b.offset = int64(offset)
	b.rowcount = int64(rowcount)
	return nil
}

func (b *queryPlanBuilder) includeOrderBy(orderBy sqlparser.OrderBy) error {
	terms, err := b.algebrizer.translateOrderBy(orderBy)
	if err != nil {
		return err
	}

	for _, obt := range terms {
		b.exprCollector.Visit(obt.expr)
	}
	b.orderBy = terms
	return nil
}

func (b *queryPlanBuilder) includeSelect(selectExprs sqlparser.SelectExprs) error {
	project, err := b.algebrizer.translateSelectExprs(selectExprs)
	if err != nil {
		return err
	}

	for _, pc := range project {
		b.exprCollector.Visit(pc.Expr)
	}
	b.project = project
	return nil
}

func (b *queryPlanBuilder) includeWhere(where *sqlparser.Where) error {
	pred, err := b.algebrizer.translateExpr(where.Expr)
	if err != nil {
		return err
	}

	b.exprCollector.Visit(pred)
	b.where = pred
	return nil
}

type exprCountMap struct {
	counts map[string]int
	exprs  []SQLExpr
}

func newExprCountMap() *exprCountMap {
	return &exprCountMap{
		counts: make(map[string]int),
	}
}

func (m *exprCountMap) add(e SQLExpr) {
	s := e.String()
	if _, ok := m.counts[s]; ok {
		m.counts[s]++
	} else {
		m.counts[s] = 1
		m.exprs = append(m.exprs, e)
	}
}

func (m *exprCountMap) remove(e SQLExpr) {
	s := e.String()
	for i, expr := range m.exprs {
		if strings.EqualFold(s, expr.String()) {
			m.counts[s]--
			if m.counts[s] == 0 {
				m.exprs = append(m.exprs[:i], m.exprs[i+1:]...)
			}
			return
		}
	}
}

func (m *exprCountMap) copyExprs() []SQLExpr {
	exprs := make([]SQLExpr, len(m.exprs))
	copy(exprs, m.exprs)
	return exprs
}

type expressionCollector struct {
	allReferencedColumns       *exprCountMap
	allNonAggReferencedColumns *exprCountMap
	allAggFunctions            *exprCountMap

	inAggFunc  bool
	removeMode bool
}

func newExpressionCollector() *expressionCollector {
	return &expressionCollector{
		allReferencedColumns:       newExprCountMap(),
		allNonAggReferencedColumns: newExprCountMap(),
		allAggFunctions:            newExprCountMap(),
	}
}

func (c *expressionCollector) Remove(e SQLExpr) {
	c.removeMode = true
	c.Visit(e)
	c.removeMode = false
}

func (c *expressionCollector) RemoveAll(e []SQLExpr) {
	c.removeMode = true
	c.VisitAll(e)
	c.removeMode = false
}

func (c *expressionCollector) VisitAll(exprs []SQLExpr) {
	for _, e := range exprs {
		c.Visit(e)
	}
}

func (v *expressionCollector) Visit(e SQLExpr) (SQLExpr, error) {
	switch typedE := e.(type) {
	case *SQLAggFunctionExpr:
		v.inAggFunc = true
		if v.removeMode {
			v.allAggFunctions.remove(typedE)
		} else {
			v.allAggFunctions.add(typedE)
		}
		for _, a := range typedE.Exprs {
			v.Visit(a)
		}
		v.inAggFunc = false
		return typedE, nil
	case SQLColumnExpr:
		if v.removeMode {
			v.allReferencedColumns.remove(typedE)
		} else {
			v.allReferencedColumns.add(typedE)
		}
		if !v.inAggFunc {
			if v.removeMode {
				v.allNonAggReferencedColumns.remove(typedE)
			} else {
				v.allNonAggReferencedColumns.add(typedE)
			}
		}
		return typedE, nil
	default:
		return walk(v, e)
	}
}

type aggFunctionFinder struct {
	aggFuncs []*SQLAggFunctionExpr
}

// getAggFunctions will take an expression and return all
// aggregation functions it finds within the expression.
func getAggFunctions(e SQLExpr) ([]*SQLAggFunctionExpr, error) {
	af := &aggFunctionFinder{}
	_, err := af.Visit(e)
	if err != nil {
		return nil, err
	}

	return af.aggFuncs, nil
}

func (af *aggFunctionFinder) Visit(e SQLExpr) (SQLExpr, error) {
	switch typedE := e.(type) {
	case *SQLExistsExpr, SQLColumnExpr, SQLNullValue, SQLNumeric, SQLVarchar, *SQLSubqueryExpr:
		return e, nil
	case *SQLAggFunctionExpr:
		af.aggFuncs = append(af.aggFuncs, typedE)
	default:
		return walk(af, e)
	}

	return e, nil
}

type aggFunctionExprReplacer struct {
	tableName string
}

func replaceAggFunctionsWithColumns(tableName string, e SQLExpr) (SQLExpr, error) {
	v := &aggFunctionExprReplacer{tableName}
	return v.Visit(e)
}

func (v *aggFunctionExprReplacer) Visit(e SQLExpr) (SQLExpr, error) {
	switch typedE := e.(type) {
	case *SQLAggFunctionExpr:
		columnType := schema.ColumnType{
			SQLType:   typedE.Type(),
			MongoType: schema.MongoNone,
		}
		return SQLColumnExpr{v.tableName, typedE.String(), columnType}, nil
	default:
		return walk(v, e)
	}
}

func projectedColumnFromExpr(expr SQLExpr) *SelectExpression {
	pc := &SelectExpression{
		Column: &Column{
			SQLType: expr.Type(),
		},
		Expr: expr,
	}

	if sqlCol, ok := expr.(SQLColumnExpr); ok {
		pc.Name = sqlCol.columnName
		pc.Table = sqlCol.tableName
		pc.MongoType = sqlCol.columnType.MongoType
	} else {
		pc.Name = expr.String()
	}

	return pc
}
