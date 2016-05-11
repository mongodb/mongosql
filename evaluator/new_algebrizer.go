package evaluator

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/kr/pretty"
)

// AlgebrizerContext holds information related to algebrization.
type AlgebrizerContext struct {
	dbName string
	schema *schema.Schema
}

// NewAlgebrizerContext creates a new algebrizer context
func NewAlgebrizerContext(dbName string, schema *schema.Schema) *AlgebrizerContext {
	return &AlgebrizerContext{
		dbName: dbName,
		schema: schema,
	}
}

// Algebrize takes a parsed SQL statement and returns an algebrized form of the query.
func Algebrize(selectStatement sqlparser.SelectStatement, ctx *AlgebrizerContext) (PlanStage, error) {

	fmt.Printf("\n%# v", pretty.Formatter(selectStatement))

	switch typedS := selectStatement.(type) {

	case *sqlparser.Select:
		return newAlgebrizeSelect(typedS, ctx)
	default:
		return nil, fmt.Errorf("can't handle select statement of type %T", selectStatement)
	}
}

func newAlgebrizeSelect(sel *sqlparser.Select, ctx *AlgebrizerContext) (PlanStage, error) {
	var plan PlanStage
	var err error
	if sel.From != nil {
		plan, err = newAlgebrizeTableExprs(sel.From, ctx)
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("\n%# v", pretty.Formatter(plan))

	return plan, nil
}

func newAlgebrizeTableExprs(tableExprs sqlparser.TableExprs, ctx *AlgebrizerContext) (PlanStage, error) {

	var plan PlanStage
	for i, tableExpr := range tableExprs {
		temp, err := newAlgebrizeTableExpr(tableExpr, ctx)
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

func newAlgebrizeTableExpr(tableExpr sqlparser.TableExpr, ctx *AlgebrizerContext) (PlanStage, error) {
	switch typedT := tableExpr.(type) {
	case *sqlparser.AliasedTableExpr:
		return newAlgebrizeSimpleTableExpr(typedT.Expr, string(typedT.As), ctx)
	case sqlparser.SimpleTableExpr:
		return newAlgebrizeSimpleTableExpr(typedT, "", ctx)
	default:
		return nil, fmt.Errorf("no support for %T", tableExpr)
	}
}

func newAlgebrizeSimpleTableExpr(tableExpr sqlparser.SimpleTableExpr, aliasName string, ctx *AlgebrizerContext) (PlanStage, error) {
	switch typedT := tableExpr.(type) {
	case *sqlparser.TableName:
		return newAlgebrizeTableName(typedT, aliasName, ctx)
	default:
		return nil, fmt.Errorf("no support for %T", tableExpr)
	}
}

func newAlgebrizeTableName(tableName *sqlparser.TableName, aliasName string, ctx *AlgebrizerContext) (PlanStage, error) {
	name := string(tableName.Name)
	if aliasName == "" {
		aliasName = name
	}

	dbName := strings.ToLower(string(tableName.Qualifier))
	if dbName == "" {
		dbName = ctx.dbName
	}

	if strings.EqualFold(dbName, InformationDatabase) {
		return NewSchemaDataSourceStage(name, aliasName), nil
	}

	return NewMongoSourceStage(ctx.schema, dbName, name, aliasName)
}
