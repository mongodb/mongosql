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

	//fmt.Printf("\nSelect Statement: %# v", pretty.Formatter(selectStatement))

	algebrizer := &algebrizer{
		dbName: dbName,
		schema: schema,
	}
	return algebrizer.translateSelectStatement(selectStatement)
}

type algebrizer struct {
	parent           *algebrizer
	sourceName       string // the name of the output. This means we need all projected columns to use this as the table name.
	dbName           string // the default database name.
	schema           *schema.Schema
	columns          []*Column         // all the columns in scope.
	projectedColumns SelectExpressions // columns to be projected from this scope.
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
		if a.parent != nil {
			return a.parent.lookupColumn(tableName, columnName)
		}

		if tableName == "" {
			return nil, fmt.Errorf("unknown column %q", columnName)
		}

		return nil, fmt.Errorf("unknown column %q in table %q", columnName, tableName)
	}

	return found, nil
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
	ascending := !strings.EqualFold(order.Direction, "desc")
	if numVal, ok := order.Expr.(sqlparser.NumVal); ok {
		n, err := strconv.ParseInt(sqlparser.String(numVal), 10, 64)
		if err != nil {
			return nil, err
		}

		if int(n) > len(a.projectedColumns) {
			return nil, fmt.Errorf("unknown column \"%v\" in order clause", n)
		}

		if n >= 0 {
			return &orderByTerm{
				expr:      a.projectedColumns[n-1].Expr,
				ascending: ascending,
			}, nil
		}
	}

	e, err := a.translateExpr(order.Expr)
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
	default:
		return nil, fmt.Errorf("no support for %T", selectStatement)
	}
}

func (a *algebrizer) translateSelect(sel *sqlparser.Select) (PlanStage, error) {
	var plan PlanStage
	var err error
	if sel.From != nil {
		plan, err = a.translateTableExprs(sel.From)
		if err != nil {
			return nil, err
		}
	}

	// WHERE

	if sel.Where != nil {
		expr, err := a.translateExpr(sel.Where.Expr)
		if err != nil {
			return nil, err
		}

		if expr.Type() != schema.SQLBoolean {
			expr = &SQLConvertExpr{
				expr:     expr,
				convType: schema.SQLBoolean,
			}
		}

		plan = NewFilterStage(plan, expr)
	}

	projectedColumns, err := a.translateSelectExprs(sel.SelectExprs)
	if err != nil {
		return nil, err
	}

	// NOTE: at this point, we are now resolving from projected columns first, followed
	// by source columns.
	a.projectedColumns = projectedColumns

	// GROUP BY

	// HAVING

	// DISTINCT

	if sel.OrderBy != nil {
		terms, err := a.translateOrderBy(sel.OrderBy)
		if err != nil {
			return nil, err
		}

		plan = NewOrderByStage(plan, terms...)
	}

	if sel.Limit != nil {
		offset, rowcount, err := a.translateLimit(sel.Limit)
		if err != nil {
			return nil, err
		}

		plan = NewLimitStage(plan, int64(offset), int64(rowcount))
	}

	plan = NewProjectStage(plan, projectedColumns...)
	return plan, nil
}

func (a *algebrizer) translateSelectExprs(selectExprs sqlparser.SelectExprs) (SelectExpressions, error) {
	var projectedColumns SelectExpressions
	hasGlobalStar := false
	for _, selectExpr := range selectExprs {
		switch typedE := selectExpr.(type) {

		// TODO: validate no mixture of star and non-star expression
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
							View:      column.Name, // ???
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
				projectedColumn.Name = sqlCol.columnName
				projectedColumn.MongoType = sqlCol.columnType.MongoType
			}

			if typedE.As != nil {
				projectedColumn.Name = string(typedE.As)
			} else if projectedColumn.Name == "" {
				projectedColumn.Name = sqlparser.String(typedE)
			}

			// TODO: not sure we need View at all...
			projectedColumn.View = projectedColumn.Name

			projectedColumns = append(projectedColumns, projectedColumn)
		}
	}

	if hasGlobalStar && len(selectExprs) > 1 {
		return nil, fmt.Errorf("cannot have a global * in the field list conjunction with any other columns")
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
		} else {
			predicate = SQLBool(true)
		}
		if err != nil {
			return nil, err
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
		if strings.EqualFold(dbName, InformationDatabase) {
			plan = NewSchemaDataSourceStage(tableName, aliasName)
		} else {
			plan, err = NewMongoSourceStage(a.schema, dbName, tableName, aliasName)
			if err != nil {
				return nil, err
			}
		}

		a.registerColumns(plan.OpFields())

		return plan, nil
	case *sqlparser.Subquery:
		plan, err := a.translateNamedSelectStatement(typedT.Select, aliasName)
		if err != nil {
			return nil, err
		}

		plan = &SubqueryStage{
			source:    plan,
			tableName: aliasName,
		}

		a.registerColumns(plan.OpFields())

		return plan, nil
	default:
		return nil, fmt.Errorf("no support for %T", tableExpr)
	}
}

func (a *algebrizer) translateExpr(expr sqlparser.Expr) (SQLExpr, error) {
	switch typedE := expr.(type) {
	case *sqlparser.BinaryExpr:
		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		left, right, err = reconcileSQLExprs(left, right)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case '+':
			return &SQLAddExpr{left, right}, nil
		case '-':
			return &SQLSubtractExpr{left, right}, nil
		case '*':
			return &SQLMultiplyExpr{left, right}, nil
		case '/':
			return &SQLDivideExpr{left, right}, nil
		default:
			return nil, fmt.Errorf("no support for binary operator '%v'", typedE.Operator)
		}
	case *sqlparser.ColName:
		tableName := ""
		if typedE.Qualifier != nil {
			tableName = string(typedE.Qualifier)
		}

		columnName := string(typedE.Name)

		// certain stages resolve columns firt with the projected columns, disregarding
		// the table name.
		if a.projectedColumns != nil {
			for _, pc := range a.projectedColumns {
				if tableName == "" && strings.EqualFold(pc.Name, columnName) {
					return pc.Expr, nil
				}
			}
		}

		column, err := a.lookupColumn(tableName, columnName)
		if err != nil {
			return nil, err
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

		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case sqlparser.AST_EQ:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLEqualsExpr{left, right}, nil
		case sqlparser.AST_LT:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLLessThanExpr{left, right}, nil
		case sqlparser.AST_GT:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}

			return &SQLGreaterThanExpr{left, right}, nil
		case sqlparser.AST_LE:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLLessThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_GE:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLGreaterThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_NE:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLNotEqualsExpr{left, right}, nil
		case sqlparser.AST_LIKE:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLLikeExpr{left, right}, nil
		case sqlparser.AST_IN:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}

			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLInExpr{left, right}, nil
		case sqlparser.AST_NOT_IN:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}

			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLNotExpr{&SQLInExpr{left, right}}, nil
		default:
			return nil, fmt.Errorf("no support for operator %q", typedE.Operator)
		}
	case *sqlparser.FuncExpr:
		return a.translateFuncExpr(typedE)
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
	case sqlparser.StrVal:
		return SQLVarchar(string(typedE)), nil
	case *sqlparser.Subquery:
		return a.translateSubqueryExpr(typedE)
	default:
		return nil, fmt.Errorf("no support for %T", expr)
	}
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

			if !strings.EqualFold(name, "count") {
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

func (a *algebrizer) translateSubqueryExpr(expr *sqlparser.Subquery) (SQLExpr, error) {
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
		plan: plan,
	}, nil
}
