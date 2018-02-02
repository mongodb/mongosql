package evaluator_test

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/stretchr/testify/require"
)

func TestGroupByPlanStage(t *testing.T) {
	ctx := createTestExecutionCtx(nil)

	runTest := func(t *testing.T,
		projectedColumns evaluator.ProjectedColumns,
		keys []evaluator.SQLExpr,
		rows []bson.D, expectedRows []evaluator.Values) {

		bss := evaluator.NewBSONSourceStage(1, tableOneName,
			collation.Must(collation.Get("utf8_general_ci")), rows)

		groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(ctx)
		require.NoError(t, err)

		row := &evaluator.Row{}

		i := 0

		for iter.Next(row) {
			require.Equal(t, len(row.Data), len(expectedRows[i]))
			require.Equal(t, row.Data, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	}

	data := []bson.D{
		{{Name: "_id", Value: 1}, {Name: "a", Value: "a"}, {Name: "b", Value: 7}},
		{{Name: "_id", Value: 2}, {Name: "a", Value: "A"}, {Name: "b", Value: 8}},
		{{Name: "_id", Value: 3}, {Name: "a", Value: "b"}, {Name: "b", Value: 9}},
	}

	projectedColumns := evaluator.ProjectedColumns{
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: tableOneName,
				OriginalTable: tableOneName, Database: evaluator.BSONSourceDB, Name: "a",
				OriginalName: "a", MappingRegistryName: "", SQLType: schema.SQLVarchar,
				MongoType: schema.MongoInt, PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				schema.SQLVarchar, schema.MongoString),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "", SQLType: schema.SQLFloat,
				MongoType: schema.MongoNone, PrimaryKey: false},
			Expr: &evaluator.SQLAggFunctionExpr{
				Name: "sum",
				Exprs: []evaluator.SQLExpr{
					evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
						schema.SQLInt, schema.MongoInt),
				},
			},
		},
	}

	keys := []evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
		tableOneName, "a", schema.SQLVarchar, schema.MongoString)}

	expected := []evaluator.Values{
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
			Data: evaluator.SQLVarchar("a")}, {SelectID: 1,
			Database: evaluator.BSONSourceDB, Table: "", Name: "sum(b)",
			Data: evaluator.SQLFloat(15)}},
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
			Data: evaluator.SQLVarchar("b")}, {SelectID: 1,
			Database: evaluator.BSONSourceDB, Table: "", Name: "sum(b)",
			Data: evaluator.SQLFloat(9)}},
	}

	runTest(t, projectedColumns, keys, data, expected)
}

func TestGroupByPlanStage_MemoryLimits(t *testing.T) {
	ctx := createTestExecutionCtx(nil)
	ctx.Variables().SetSystemVariable(variable.MongoDBMaxStageSize, 100)

	runTest := func(projectedColumns evaluator.ProjectedColumns, keys []evaluator.SQLExpr,
		rows []bson.D) {
		bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(ctx)
		require.NoError(t, err)

		row := &evaluator.Row{}

		ok := iter.Next(row)
		require.False(t, ok)

		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	}

	data := []bson.D{
		{{Name: "_id", Value: 1}, {Name: "a", Value: "a"}, {Name: "b", Value: 7}},
		{{Name: "_id", Value: 2}, {Name: "a", Value: "A"}, {Name: "b", Value: 8}},
		{{Name: "_id", Value: 3}, {Name: "a", Value: "b"}, {Name: "b", Value: 9}},
	}

	projectedColumns := evaluator.ProjectedColumns{
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: tableOneName,
				OriginalTable: tableOneName, Database: evaluator.BSONSourceDB, Name: "a",
				OriginalName: "a", MappingRegistryName: "", SQLType: schema.SQLVarchar,
				MongoType: schema.MongoInt, PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				schema.SQLVarchar, schema.MongoString),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "", SQLType: schema.SQLFloat, MongoType: schema.MongoNone,
				PrimaryKey: false},
			Expr: &evaluator.SQLAggFunctionExpr{
				Name: "sum",
				Exprs: []evaluator.SQLExpr{
					evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
						schema.SQLInt, schema.MongoInt),
				},
			},
		},
	}

	keys := []evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
		tableOneName, "a", schema.SQLVarchar, schema.MongoString)}

	runTest(projectedColumns, keys, data)
}

func TestGroupByStageMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: "a"}, {Name: "b", Value: 7}},
		{{Name: "a", Value: "a"}, {Name: "b", Value: 8}},
		{{Name: "a", Value: "b"}, {Name: "b", Value: 9}},
		{{Name: "a", Value: "b"}, {Name: "b", Value: 10}},
		{{Name: "a", Value: "b"}, {Name: "b", Value: 11}},
	}

	projectedColumns := evaluator.ProjectedColumns{
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: tableOneName,
				OriginalTable: tableOneName, Database: evaluator.BSONSourceDB, Name: "a",
				OriginalName: "a", MappingRegistryName: "", SQLType: schema.SQLVarchar,
				MongoType: schema.MongoString, PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				schema.SQLVarchar, schema.MongoString),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "", SQLType: schema.SQLInt, MongoType: schema.MongoInt,
				PrimaryKey: false},
			Expr: &evaluator.SQLAggFunctionExpr{
				Name: "sum",
				Exprs: []evaluator.SQLExpr{
					evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
						schema.SQLInt, schema.MongoInt),
				},
			},
		},
	}

	keys := []evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
		tableOneName, "a", schema.SQLVarchar, schema.MongoString)}

	bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)
	groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

	actual := getAllocatedMemorySizeAfterIteration(groupBy)
	expected := (valueSize(evaluator.BSONSourceDB, tableOneName, "a", evaluator.SQLVarchar("a")) +
		valueSize(evaluator.BSONSourceDB, "", "sum(b)", evaluator.SQLInt(0))) * 2

	require.Equal(t, expected, actual)
}
