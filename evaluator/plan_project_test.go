package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/stretchr/testify/require"
)

func TestProjectStage(t *testing.T) {
	ctx := createTestExecutionCtx(nil)

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTest := func(t *testing.T,
		projectedColumns evaluator.ProjectedColumns,
		optimize bool,
		rows []bson.D,
		expectedRows []evaluator.Values) {

		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan evaluator.PlanStage
		var err error

		project := evaluator.NewProjectStage(ts, projectedColumns...)

		plan = project
		if optimize {
			plan = evaluator.OptimizePlan(createTestConnectionCtx(testInfo), plan)
		}

		iter, err := plan.Open(ctx)
		require.NoError(t, err)

		i := 0
		row := &evaluator.Row{}

		for iter.Next(row) {
			require.Equal(t, len(row.Data), len(expectedRows[i]))
			require.Equal(t, row.Data, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		require.Equal(t, i, len(expectedRows))

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	}

	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
	}

	projectedColumns := evaluator.ProjectedColumns{
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "a", OriginalName: "a",
				MappingRegistryName: "", SQLType: schema.SQLInt, MongoType: schema.MongoInt,
				PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				schema.SQLInt, schema.MongoInt),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "b", OriginalName: "b",
				MappingRegistryName: "", SQLType: schema.SQLInt, MongoType: schema.MongoInt,
				PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
				schema.SQLInt, schema.MongoInt),
		},
	}

	expected := []evaluator.Values{
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a",
			Data: evaluator.SQLInt(6)}, {SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: "", Name: "b", Data: evaluator.SQLInt(9)}},
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a",
			Data: evaluator.SQLInt(3)}, {SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: "", Name: "b", Data: evaluator.SQLInt(4)}},
	}

	t.Run("correct results", func(t *testing.T) {
		runTest(t, projectedColumns, false, rows, expected)
	})
	t.Run("correct results after optimizations", func(t *testing.T) {
		runTest(t, projectedColumns, true, rows, expected)
	})
}

func TestProjectStageMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
	}
	bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)
	project := evaluator.NewProjectStage(bss, evaluator.ProjectedColumn{
		Column: &evaluator.Column{SelectID: 1, Table: tableOneName, OriginalTable: tableOneName,
			Database: evaluator.BSONSourceDB, Name: "a", OriginalName: "a",
			MappingRegistryName: "", SQLType: schema.SQLInt, MongoType: schema.MongoInt,
			PrimaryKey: false},
		Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
			schema.SQLInt, schema.MongoInt),
	})

	actual := getAllocatedMemorySizeAfterIteration(project)
	expected := valueSize(
		evaluator.BSONSourceDB,
		tableOneName,
		"a",
		evaluator.SQLInt(0)) * uint64(len(rows))

	require.Equal(t, expected, actual)
}
