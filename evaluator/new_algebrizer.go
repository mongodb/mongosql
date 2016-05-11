package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
)

// AlgebrizerContext holds information related to algebrization.
type AlgebrizerContext struct {
	schema *schema.Schema
}

// NewAlgebrizerContext creates a new algebrizer context
func NewAlgebrizerContext(schema *schema.Schema) *AlgebrizerContext {
	return &AlgebrizerContext{
		schema: schema,
	}
}

// Algebrize takes a parsed SQL statement and returns an algebrized form of the query.
func Algebrize(ss sqlparser.SelectStatement, ctx *AlgebrizerContext) (PlanStage, error) {

	switch stmt := ss.(type) {

	case *sqlparser.Select:

		// algebrize 'FROM' clause
		if stmt.From != nil {

			// for _, table := range stmt.From {
			// 	// err := algebrizeTableExpr(table, pCtx)
			// 	// if err != nil {
			// 	// 	return nil, err
			// 	// }
			// }

		}
	}

	return nil, nil
}
