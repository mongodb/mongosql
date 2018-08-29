package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/collation"
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
				MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType: evaluator.EvalInt64, MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				evaluator.EvalInt64, schema.MongoInt),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "b", OriginalName: "b",
				MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType: evaluator.EvalInt64, MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
				evaluator.EvalInt64, schema.MongoInt),
		},
	}

	kind := evaluator.MySQLValueKind
	expected := []evaluator.Values{
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a",
			Data: evaluator.NewSQLInt64(kind, 6)}, {SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: "", Name: "b", Data: evaluator.NewSQLInt64(kind, 9)}},
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a",
			Data: evaluator.NewSQLInt64(kind, 3)}, {SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: "", Name: "b", Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 4)}},
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
			MappingRegistryName: "",
			ColumnType: evaluator.ColumnType{
				EvalType: evaluator.EvalInt64, MongoType: schema.MongoInt,
			},
			PrimaryKey: false},
		Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
			evaluator.EvalInt64, schema.MongoInt),
	})

	actual := getAllocatedMemorySizeAfterIteration(project)
	expected := valueSize(
		evaluator.BSONSourceDB,
		tableOneName,
		"a",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0)) * uint64(len(rows))

	require.Equal(t, expected, actual)
}
