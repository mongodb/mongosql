package evaluator

import (
	"github.com/10gen/sqlproxy/collation"
)

// SubquerySourceStage handles taking sourced rows and projecting them into an alias.
type SubquerySourceStage struct {
	// aliasName is the alias for all the columns from the subquery.
	aliasName string
	// selectID is the selectID for the resulting columns.
	selectID int

	// source is the operator that provides the data to the subquery.
	source PlanStage
}

// NewSubquerySourceStage creates a new SubquerySourceStage.
func NewSubquerySourceStage(source PlanStage, selectID int, aliasName string) *SubquerySourceStage {
	return &SubquerySourceStage{
		source:    source,
		selectID:  selectID,
		aliasName: aliasName,
	}
}

func (s *SubquerySourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := s.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	var projectedColumns ProjectedColumns

	for _, column := range s.source.Columns() {
		projectedColumn := column.projectAs(column.Name)
		projectedColumn.SelectID = s.selectID
		projectedColumn.Table = s.aliasName
		projectedColumns = append(projectedColumns, projectedColumn)
	}

	return &ProjectIter{
		source:           sourceIter,
		projectedColumns: projectedColumns,
	}, nil
}

func (s *SubquerySourceStage) Columns() []*Column {
	var columns []*Column
	for _, column := range s.source.Columns() {
		column.SelectID = s.selectID
		column.Table = s.aliasName
		columns = append(columns, column.clone())
	}

	return columns
}

func (s *SubquerySourceStage) Collation() *collation.Collation {
	return s.source.Collation()
}

func (s *SubquerySourceStage) clone() *SubquerySourceStage {
	return &SubquerySourceStage{
		source:    s.source,
		selectID:  s.selectID,
		aliasName: s.aliasName,
	}
}
