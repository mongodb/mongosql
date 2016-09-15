package evaluator

import (
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
)

// PlanStage represents a single a node in the Plan tree.
type PlanStage interface {
	node

	// Open returns an iterator that returns results from executing this plan stage with the given
	// ExecutionContext.
	Open(*ExecutionCtx) (Iter, error)

	// Columns returns the ordered set of columns that are contained in results from this plan.
	Columns() []*Column

	// Collation returns the collation to use for comparisons.
	Collation() *collation.Collation
}

// Iter represents an object that can iterate through a set of rows.
type Iter interface {
	// Next retrieves the next row from this iterator. It returns true if it has
	// additional data and false if there is no more data or if an error occurred
	// during processing.
	//
	// When Next returns false, the Err method should be called to verify if
	// there was an error during processing.
	//
	// For example:
	//    iter, err := plan.Open(ctx);
	//
	//    if err != nil {
	//        return err
	//    }
	//
	//    for iter.Next(&row) {
	//        fmt.Printf("Row: %v\n", row)
	//    }
	//
	//    if err := iter.Close(); err != nil {
	//        return err
	//    }
	//
	//    if err := iter.Err(); err != nil {
	//        return err
	//    }
	//
	Next(*Row) bool

	// Close frees up any resources in use by this iterator. Callers should always
	// call the Close method once they are finished with an iterator.
	Close() error

	// Err returns nil if no errors happened during processing, or the actual
	// error otherwise. Callers should always call the Err method to check whether
	// any error was encountered during processing they are finished with an iterator.
	Err() error
}

// Executor represents an object that can run a command.
type Executor interface {
	Run() error
}

// Column contains information used to select data
// from a PlanStage.
type Column struct {
	SelectID  int
	Table     string
	Name      string
	SQLType   schema.SQLType
	MongoType schema.MongoType
}

type Columns []*Column

// Unique ensures that only unique columns exist in the resulting slice.
func (cs Columns) Unique() Columns {
	var results Columns
	contains := func(column *Column) bool {
		for _, c := range results {
			if c.SelectID == column.SelectID && c.Name == column.Name && c.Table == column.Table {
				return true
			}
		}

		return false
	}

	for _, c := range cs {
		if !contains(c) {
			results = append(results, c)
		}
	}

	return results
}

// ProjectedColumn is a column projection. It contains the SQLExpr for the column
// as well as the column information that will be projected.
type ProjectedColumn struct {
	// Column holds the projection information.
	*Column

	// Expr holds the expression to be evaluated.
	Expr SQLExpr
}

func (se *ProjectedColumn) clone() *ProjectedColumn {
	return &ProjectedColumn{
		Column: se.Column,
		Expr:   se.Expr,
	}
}

// ProjectedColumns is a slice of ProjectedColumn.
type ProjectedColumns []ProjectedColumn

// Unique ensures that only unique projected columns exist in the resulting slice.
func (pcs ProjectedColumns) Unique() ProjectedColumns {
	var results ProjectedColumns
	contains := func(column *ProjectedColumn) bool {
		for _, expr := range results {
			if expr.Column.SelectID == column.SelectID && expr.Column.Name == column.Name && expr.Column.Table == column.Table {
				return true
			}
		}

		return false
	}

	for _, c := range pcs {
		if !contains(&c) {
			results = append(results, c)
		}
	}

	return results
}
