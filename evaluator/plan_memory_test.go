package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestMemoryLimits(t *testing.T) {
	bgCtx := context.Background()
	execCfg := createExecutionCfg("evaluator_unit_test_dbname", 100, []uint8{4, 0, 0})
	execState := evaluator.NewExecutionState()

	t.Run("group_by", func(t *testing.T) { testGroupByMemoryLimits(bgCtx, t, execCfg, execState) })
	t.Run("order_by", func(t *testing.T) { testOrderByMemoryLimits(bgCtx, t, execCfg, execState) })

	execCfg = createExecutionCfg("evaluator_unit_test_dbname", 500, []uint8{4, 0, 0})
	t.Run("join", func(t *testing.T) { testJoinMemoryLimits(t, bgCtx, execCfg, execState) })
}

func testGroupByMemoryLimits(ctx context.Context, t *testing.T, cfg *evaluator.ExecutionConfig, st *evaluator.ExecutionState) {
	req := require.New(t)

	runTest := func(projectedColumns evaluator.ProjectedColumns, keys []evaluator.SQLExpr, rows []bson.D) {
		bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(ctx, cfg, st)
		req.NoError(err)

		row := &evaluator.Row{}

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
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
				OriginalName: "a", MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType:  evaluator.EvalString,
					MongoType: schema.MongoInt,
				}, PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				evaluator.EvalString, schema.MongoString),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType:  evaluator.EvalDouble,
					MongoType: schema.MongoNone,
				},
				PrimaryKey: false},
			Expr: evaluator.NewSQLAggregationFunctionExpr(
				"sum",
				false,
				[]evaluator.SQLExpr{
					evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
						evaluator.EvalInt64, schema.MongoInt),
				},
			),
		},
	}

	keys := []evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
		tableOneName, "a", evaluator.EvalString, schema.MongoString)}

	runTest(projectedColumns, keys, data)
}

func testOrderByMemoryLimits(ctx context.Context, t *testing.T, cfg *evaluator.ExecutionConfig, st *evaluator.ExecutionState) {
	req := require.New(t)

	runTest := func(terms []*evaluator.OrderByTerm, rows []bson.D) {
		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		orderby := evaluator.NewOrderByStage(ts, terms...)
		iter, err := orderby.Open(ctx, cfg, st)
		req.NoError(err)

		row := &evaluator.Row{}

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	}

	data := []bson.D{
		{{Name: "_id", Value: 1}, {Name: "a", Value: "a"}, {Name: "b", Value: 7}},
		{{Name: "_id", Value: 2}, {Name: "a", Value: "A"}, {Name: "b", Value: 8}},
		{{Name: "_id", Value: 3}, {Name: "a", Value: "b"}, {Name: "b", Value: 8}},
		{{Name: "_id", Value: 4}, {Name: "a", Value: "B"}, {Name: "b", Value: 7}},
	}

	terms := []*evaluator.OrderByTerm{
		evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
			tableOneName, "a", evaluator.EvalString, schema.MongoString), true),
	}

	runTest(terms, data)
}

func testJoinMemoryLimits(t *testing.T, ctx context.Context, cfg *evaluator.ExecutionConfig, st *evaluator.ExecutionState) {

	criteria := evaluator.NewSQLEqualsExpr(
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableOneName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableTwoName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
	)

	row := &evaluator.Row{}

	t.Run("inner_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, evaluator.InnerJoin)

		iter, err := operator.Open(ctx, cfg, st)
		req.NoError(err)

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	})

	t.Run("left_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, evaluator.LeftJoin)

		iter, err := operator.Open(ctx, cfg, st)
		req.NoError(err)

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	})

	t.Run("right_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, evaluator.RightJoin)

		iter, err := operator.Open(ctx, cfg, st)
		req.NoError(err)

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	})

	t.Run("cross_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(nil, evaluator.RightJoin)

		iter, err := operator.Open(ctx, cfg, st)
		req.NoError(err)

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	})
}

func TestMemoryMonitor(t *testing.T) {
	t.Run("row_generator", testRowGeneratorMemoryMonitor)
	t.Run("filter", testFilterMemoryMonitor)
	t.Run("group_by", testGroupByMemoryMonitor)
	t.Run("join", testJoinMemoryMonitor)
	t.Run("limit", testLimitMemoryMonitor)
	t.Run("dynamic_source", testDynamicSourceMemoryMonitor)
	t.Run("project", testProjectMemoryMonitor)
	t.Run("union", testUnionMemoryMonitor)
}

func testJoinMemoryMonitor(t *testing.T) {
	criteria := evaluator.NewSQLEqualsExpr(
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableOneName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableTwoName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
	)

	leftSize := valueSize(evaluator.BSONSourceDB,
		tableOneName,
		"name",
		evaluator.NewSQLVarchar(evaluator.MySQLValueKind, "personA"),
	) + valueSize(evaluator.BSONSourceDB, tableOneName, "orderid",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0)) +
		valueSize(evaluator.BSONSourceDB, tableOneName,
			"_id", evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0))

	rightSize := valueSize(evaluator.BSONSourceDB, tableTwoName, "orderid",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0)) +
		valueSize(evaluator.BSONSourceDB, tableTwoName, "amount",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0)) +
		valueSize(evaluator.BSONSourceDB, tableTwoName, "_id",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0))

	t.Run("inner_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(criteria, evaluator.InnerJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize + rightSize) * 4

		req.Equal(expected, actual)
	})

	t.Run("left_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(criteria, evaluator.LeftJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize+rightSize)*4 +
			(leftSize + rightSize - 24) // nil right values

		req.Equal(expected, actual)
	})

	t.Run("right_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(criteria, evaluator.RightJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize+rightSize)*4 +
			(leftSize - 23 + rightSize) // nil left values

		req.Equal(expected, actual)
	})

	t.Run("cross_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(nil, evaluator.CrossJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := uint64(len(customers)) * uint64(len(orders)) * (leftSize + rightSize)

		req.Equal(expected, actual)
	})
}

func testRowGeneratorMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: 6}},
		{{Name: "a", Value: 3}},
	}
	bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

	newColumn := evaluator.NewColumn(0, "", "", "", "a", "", "a",
		evaluator.EvalUint64, schema.MongoInt64, false)

	rg := evaluator.NewRowGeneratorStage(bss, newColumn)

	outputMemory := getAllocatedMemorySizeAfterIteration(rg)

	// empty rows cost nothing
	require.Equal(t, uint64(0), outputMemory)
}

func testFilterMemoryMonitor(t *testing.T) {
	schema := evaluator.MustLoadSchema(testSchema3)
	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
	}
	matcher, err := evaluator.GetSQLExpr(schema,
		evaluator.BSONSourceDB,
		tableTwoName,
		"a = 6")
	require.NoError(t, err)

	bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
	filter := evaluator.NewFilterStage(bss, matcher)

	actual := getAllocatedMemorySizeAfterIteration(filter)

	sizeA := valueSize(
		evaluator.BSONSourceDB, tableTwoName, "a",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)
	sizeB := valueSize(
		evaluator.BSONSourceDB, tableTwoName, "b",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)
	expected := sizeA + sizeB

	require.Equal(t, expected, actual)
}

func testGroupByMemoryMonitor(t *testing.T) {
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
				OriginalName: "a", MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType:  evaluator.EvalString,
					MongoType: schema.MongoString,
				},
				PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				evaluator.EvalString, schema.MongoString),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType:  evaluator.EvalInt64,
					MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: evaluator.NewSQLAggregationFunctionExpr(
				"sum",
				false,
				[]evaluator.SQLExpr{
					evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
						evaluator.EvalInt64, schema.MongoInt),
				},
			),
		},
	}

	keys := []evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
		tableOneName, "a", evaluator.EvalString, schema.MongoString)}

	bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)
	groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

	actual := getAllocatedMemorySizeAfterIteration(groupBy)

	sizeA := valueSize(
		evaluator.BSONSourceDB, tableOneName, "a",
		evaluator.NewSQLVarchar(evaluator.MySQLValueKind, "a"),
	)
	sizeB := valueSize(
		evaluator.BSONSourceDB, "", "sum(b)",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)
	expected := 2 * (sizeA + sizeB)

	require.Equal(t, expected, actual)
}

func testLimitMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
		{{Name: "a", Value: 0}, {Name: "b", Value: 13}},
		{{Name: "a", Value: -3}, {Name: "b", Value: 8}},
		{{Name: "a", Value: -6}, {Name: "b", Value: 17}},
		{{Name: "a", Value: -9}, {Name: "b", Value: 12}},
	}

	t.Run("non-blocking source", func(t *testing.T) {
		bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		ls := evaluator.NewLimitStage(bss, 2, 2)

		actual := getAllocatedMemorySizeAfterIteration(ls)

		sizeA := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "a",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		sizeB := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "b",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		expected := 2 * (sizeA + sizeB)

		require.Equal(t, expected, actual)
	})
	t.Run("blocking source", func(t *testing.T) {
		bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		os := evaluator.NewOrderByStage(bss,
			evaluator.NewOrderByTerm(
				evaluator.NewSQLColumnExpr(
					1,
					evaluator.BSONSourceDB,
					tableTwoName,
					"a",
					evaluator.EvalInt64,
					schema.MongoInt),
				true))
		ls := evaluator.NewLimitStage(os, 2, 2)

		actual := getAllocatedMemorySizeAfterIteration(ls)

		sizeA := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "a",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		sizeB := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "b",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		expected := 4 * (sizeA + sizeB)

		require.Equal(t, expected, actual)
	})
}

func testDynamicSourceMemoryMonitor(t *testing.T) {
	tableName := "foo"
	table := catalog.NewDynamicTable(tableName, catalog.BaseTable, func() []*catalog.DataRow {
		return []*catalog.DataRow{
			catalog.NewDataRow(1, 2),
			catalog.NewDataRow(2, 3),
			catalog.NewDataRow(3, 4),
		}
	})

	_, err := table.AddColumn("one", schema.SQLInt)
	require.NoError(t, err)
	_, err = table.AddColumn("two", schema.SQLInt)
	require.NoError(t, err)

	db := &catalog.Database{}

	source := evaluator.NewDynamicSourceStage(db, table, 1, tableName)

	actual := getAllocatedMemorySizeAfterIteration(source)
	expected := (valueSize(string(db.Name), tableName, "one",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0)) +
		valueSize(string(db.Name), tableName,
			"two", evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0))) * 3

	require.Equal(t, expected, actual)
}

func testProjectMemoryMonitor(t *testing.T) {
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

func testUnionMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
	}
	u := evaluator.NewUnionStage(evaluator.UnionDistinct,
		evaluator.NewBSONSourceStage(1, "foo", collation.Default, rows),
		evaluator.NewBSONSourceStage(2, "bar", collation.Default, rows),
	)

	actual := getAllocatedMemorySizeAfterIteration(u)

	sizeA := valueSize(
		evaluator.BSONSourceDB, tableOneName, "a",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)
	sizeB := valueSize(
		evaluator.BSONSourceDB, tableOneName, "b",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)
	expected := uint64(len(rows)*2) * (sizeA + sizeB)

	require.Equal(t, expected, actual)
}
