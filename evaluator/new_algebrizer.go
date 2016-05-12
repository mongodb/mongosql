package evaluator

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/kr/pretty"
)

// Algebrize takes a parsed SQL statement and returns an algebrized form of the query.
func Algebrize(selectStatement sqlparser.SelectStatement, dbName string, schema *schema.Schema) (PlanStage, error) {

	fmt.Printf("\nSelect Statement: %# v", pretty.Formatter(selectStatement))

	algebrizer := &algebrizer{
		dbName: dbName,
		schema: schema,
	}

	return algebrizer.translateSelectStatement(selectStatement)
}

type algebrizer struct {
	dbName  string
	schema  *schema.Schema
	columns []*Column
}

func (a *algebrizer) registerColumns(columns []*Column) {
	a.columns = append(a.columns, columns...)
}

func (a *algebrizer) lookupColumn(tableName, columnName string) (*Column, error) {
	var found *Column
	for _, column := range a.columns {
		if strings.EqualFold(column.Name, columnName) && (tableName == "" || strings.EqualFold(column.Table, tableName)) {
			if found != nil {
				fullColumnName := columnName
				if tableName != "" {
					fullColumnName = tableName + "." + columnName
				}
				return nil, fmt.Errorf("column %q in the field list is ambiguous", fullColumnName)
			}
			found = column
		}
	}

	return found, nil
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
	for _, selectExpr := range selectExprs {
		switch typedE := selectExpr.(type) {

		// TODO: validate no mixture of star and non-star expression
		case *sqlparser.StarExpr:

			// validate tableName if present. Need to have a map of alias -> tableName -> schema

		case *sqlparser.NonStarExpr:

			translatedExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			result := SelectExpression{
				Expr: translatedExpr,
			}

			result.Column = &Column{
				MongoType: schema.MongoNone,
				SQLType:   translatedExpr.Type(),
				Table:     "", // this needs to come from the alias of the parent...?
			}

			if typedE.As != nil {
				result.Name = string(typedE.As)
			} else if sqlCol, ok := translatedExpr.(SQLColumnExpr); ok {
				result.Name = sqlCol.columnName
			}

			// TODO: not sure we need View at all...
			result.View = result.Name

			results = append(results, result)
		}
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

	default:
		return nil, fmt.Errorf("no support for %T", expr)
	}
}
