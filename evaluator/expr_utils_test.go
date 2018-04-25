package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	. "github.com/10gen/sqlproxy/evaluator"

	"github.com/stretchr/testify/require"
)

func TestIsFullyPushedDown(t *testing.T) {
	ms := createMongoSource(0, "foo", "foo")
	column := createProjectedColumn(0, ms, "foo", "a", "foo", "a")
	db, _ := testCatalog.Database("INFORMATION_SCHEMA")
	table, _ := db.Table("CHARACTER_SETS")

	tests := []struct {
		description string
		planStage   PlanStage
		err         error
	}{
		{
			"MongoSource",
			ms,
			nil,
		},
		{
			"Project -> MongoSource",
			NewProjectStage(ms, column),
			ErrNotFullyPushedDown,
		},
		{
			"Project -> RowGeneratorStage -> MongoSource",
			NewProjectStage(NewRowGeneratorStage(ms, nil), column),
			nil,
		},
		{
			"Empty",
			NewEmptyStage(nil, nil),
			nil,
		},
		{
			"DynamicSourceStage",
			NewDynamicSourceStage(db, table.(*catalog.DynamicTable), 0, ""),
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := require.New(t)
			if test.err != nil {
				req.NotNilf(IsFullyPushedDown(test.planStage), "expected error but got no error")
			} else {
				req.Nilf(IsFullyPushedDown(test.planStage), "expected no error but got error")
			}
		})
	}
}
