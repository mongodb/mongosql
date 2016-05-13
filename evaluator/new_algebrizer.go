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
	sourceName string // the name of the output. This means we need all projected columns to use this as the table name.
	dbName     string // the default database name.
	schema     *schema.Schema
	columns    []*Column // all the columns in scope.
}

func (a *algebrizer) registerColumns(columns []*Column) {
	a.columns = append(a.columns, columns...)
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

func (a *algebrizer) algebrizeNamedSource(selectStatement sqlparser.SelectStatement, sourceName string) (PlanStage, error) {
	algebrizer := &algebrizer{
		dbName:     a.dbName,
		schema:     a.schema,
		sourceName: sourceName,
	}
	return algebrizer.translateSelectStatement(selectStatement)
}

func (a *algebrizer) translateSelectStatement(selectStatement sqlparser.SelectStatement) (PlanStage, error) {
	switch typedS := selectStatement.(type) {

	case *sqlparser.Select:
		return a.translateSelect(typedS)
	default:
		return nil, fmt.Errorf("can't handle select statement of type %T", selectStatement)
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

	if sel.SelectExprs != nil {
		sExprs, err := a.translateSelectExprs(sel.SelectExprs)
		if err != nil {
			return nil, err
		}

		plan = &ProjectStage{
			source: plan,
			sExprs: sExprs,
		}
	}

	return plan, nil
}

func (a *algebrizer) translateSelectExprs(selectExprs sqlparser.SelectExprs) (SelectExpressions, error) {
	var results SelectExpressions
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
					results = append(results, SelectExpression{
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

			result := SelectExpression{
				Expr: translatedExpr,
				Column: &Column{
					Table:     a.sourceName,
					MongoType: schema.MongoNone,
					SQLType:   translatedExpr.Type(),
				},
			}

			if sqlCol, ok := translatedExpr.(SQLColumnExpr); ok {
				result.Name = sqlCol.columnName
				result.MongoType = sqlCol.columnType.MongoType
			}

			if typedE.As != nil {
				result.Name = string(typedE.As)
			} else if result.Name == "" {
				result.Name = sqlparser.String(typedE)
			}

			// TODO: not sure we need View at all...
			result.View = result.Name

			results = append(results, result)
		}
	}

	if hasGlobalStar && len(selectExprs) > 1 {
		return nil, fmt.Errorf("cannot have a global * in the field list conjunction with any other columns")
	}

	return results, nil
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
	case sqlparser.SimpleTableExpr:
		return a.translateSimpleTableExpr(typedT, "")
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
		plan, err := a.algebrizeNamedSource(typedT.Select, aliasName)
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
	case *sqlparser.ColName:
		tableName := ""
		if typedE.Qualifier != nil {
			tableName = string(typedE.Qualifier)
		}

		columnName := string(typedE.Name)

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

	default:
		return nil, fmt.Errorf("no support for %T", expr)
	}
}
