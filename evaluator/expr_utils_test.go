package evaluator_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"github.com/stretchr/testify/require"
)

func TestGetMongoDBInfo(t *testing.T) {

	req := require.New(t)

	versionArray := []uint8{3, 5, 6}
	sch := &schema.Schema{}
	privileges := mongodb.Privilege(0)

	info := GetMongoDBInfo(versionArray, sch, privileges)
	req.Equal(info.Version, "3.5.6")

	info = GetMongoDBInfo(nil, sch, privileges)
	req.Equal(info.Version, "3.4.0")
}

func TestIsFullyPushedDown(t *testing.T) {
	ms := createMongoSource(0, "foo", "foo")
	column := createProjectedColumn(0, ms, "foo", "a", "foo", "a", false)
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
