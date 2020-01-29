package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	. "github.com/10gen/sqlproxy/evaluator/results"
	. "github.com/10gen/sqlproxy/evaluator/types"
	. "github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMemoryLimits(t *testing.T) {
	bgCtx := context.Background()
	execCfg := createExecutionCfg("evaluator_unit_test_dbname", 100, []uint8{4, 0, 0}, MySQLValueKind)
	execState := NewExecutionState()

	t.Run("group_by", func(t *testing.T) { testGroupByMemoryLimits(bgCtx, t, execCfg, execState) })
	t.Run("order_by", func(t *testing.T) { testOrderByMemoryLimits(bgCtx, t, execCfg, execState) })

	execCfg = createExecutionCfg("evaluator_unit_test_dbname", 500, []uint8{4, 0, 0}, MySQLValueKind)
	t.Run("join", func(t *testing.T) { testJoinMemoryLimits(bgCtx, t, execCfg, execState) })
}

func testGroupByMemoryLimits(ctx context.Context, t *testing.T, cfg *ExecutionConfig, st *ExecutionState) {
	req := require.New(t)

	runTest := func(projectedColumns ProjectedColumns, keys []SQLExpr, rows []bson.D) {
		bss := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		groupBy := NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(ctx, cfg, st)
		req.NoError(err)

		row := &Row{}

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	}

	data := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("_id", 1), bsonutil.NewDocElem("a", "a"), bsonutil.NewDocElem("b", 7)),
		bsonutil.NewD(bsonutil.NewDocElem("_id", 2), bsonutil.NewDocElem("a", "A"), bsonutil.NewDocElem("b", 8)),
		bsonutil.NewD(bsonutil.NewDocElem("_id", 3), bsonutil.NewDocElem("a", "b"), bsonutil.NewDocElem("b", 9)),
	)

	projectedColumns := ProjectedColumns{
		ProjectedColumn{
			Column: &Column{SelectID: 1, Table: tableOneName,
				OriginalTable: tableOneName, Database: BSONSourceDB, Name: "a",
				OriginalName: "a", MappingRegistryName: "",
				ColumnType: &ColumnType{
					EvalType:  EvalString,
					MongoType: schema.MongoInt,
				}, PrimaryKey: false},
			Expr: testSQLColumnExpr(1, BSONSourceDB, tableOneName, "a",
				EvalString, schema.MongoString, false),
		},
		ProjectedColumn{
			Column: &Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "",
				ColumnType: &ColumnType{
					EvalType:  EvalDouble,
					MongoType: schema.MongoNone,
				},
				PrimaryKey: false},
			Expr: NewSQLAggregationFunctionExpr(
				"sum",
				false,
				[]SQLExpr{
					testSQLColumnExpr(1, BSONSourceDB, tableOneName, "b",
						EvalInt64, schema.MongoInt, false),
				},
			),
		},
	}

	keys := []SQLExpr{testSQLColumnExpr(1, BSONSourceDB,
		tableOneName, "a", EvalString, schema.MongoString, false)}

	runTest(projectedColumns, keys, data)
}

func testOrderByMemoryLimits(ctx context.Context, t *testing.T, cfg *ExecutionConfig, st *ExecutionState) {
	req := require.New(t)

	runTest := func(terms []*OrderByTerm, rows []bson.D) {
		ts := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		orderby := NewOrderByStage(ts, terms...)
		iter, err := orderby.Open(ctx, cfg, st)
		req.NoError(err)

		row := &Row{}

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	}

	data := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("_id", 1), bsonutil.NewDocElem("a", "a"), bsonutil.NewDocElem("b", 7)),
		bsonutil.NewD(bsonutil.NewDocElem("_id", 2), bsonutil.NewDocElem("a", "A"), bsonutil.NewDocElem("b", 8)),
		bsonutil.NewD(bsonutil.NewDocElem("_id", 3), bsonutil.NewDocElem("a", "b"), bsonutil.NewDocElem("b", 8)),
		bsonutil.NewD(bsonutil.NewDocElem("_id", 4), bsonutil.NewDocElem("a", "B"), bsonutil.NewDocElem("b", 7)),
	)

	terms := []*OrderByTerm{
		NewOrderByTerm(testSQLColumnExpr(1, BSONSourceDB,
			tableOneName, "a", EvalString, schema.MongoString, false), true),
	}

	runTest(terms, data)
}

func testJoinMemoryLimits(ctx context.Context, t *testing.T, cfg *ExecutionConfig, st *ExecutionState) {

	criteria := NewSQLComparisonExpr(
		EQ,
		testSQLColumnExpr(
			1,
			BSONSourceDB,
			tableOneName,
			"orderid",
			EvalInt64,
			schema.MongoInt,
			false,
		),
		testSQLColumnExpr(
			1,
			BSONSourceDB,
			tableTwoName,
			"orderid",
			EvalInt64,
			schema.MongoInt,
			false,
		),
	)

	row := &Row{}

	t.Run("inner_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, InnerJoin)

		iter, err := operator.Open(ctx, cfg, st)
		req.NoError(err)

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	})

	t.Run("left_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, LeftJoin)

		iter, err := operator.Open(ctx, cfg, st)
		req.NoError(err)

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	})

	t.Run("right_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, RightJoin)

		iter, err := operator.Open(ctx, cfg, st)
		req.NoError(err)

		ok := iter.Next(ctx, row)
		req.False(ok)

		req.NoError(iter.Close())
		req.Error(iter.Err())
	})

	t.Run("cross_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(nil, RightJoin)

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
	criteria := NewSQLComparisonExpr(
		EQ,
		testSQLColumnExpr(
			1,
			BSONSourceDB,
			tableOneName,
			"orderid",
			EvalInt64,
			schema.MongoInt,
			false,
		),
		testSQLColumnExpr(
			1,
			BSONSourceDB,
			tableTwoName,
			"orderid",
			EvalInt64,
			schema.MongoInt,
			false,
		),
	)

	leftSize := valueSize(BSONSourceDB,
		tableOneName,
		"name",
		NewSQLVarchar(MySQLValueKind, "personA"),
	) + valueSize(BSONSourceDB, tableOneName, "orderid",
		NewSQLInt64(MySQLValueKind, 0)) +
		valueSize(BSONSourceDB, tableOneName,
			"_id", NewSQLInt64(MySQLValueKind, 0))

	rightSize := valueSize(BSONSourceDB, tableTwoName, "orderid",
		NewSQLInt64(MySQLValueKind, 0)) +
		valueSize(BSONSourceDB, tableTwoName, "amount",
			NewSQLInt64(MySQLValueKind, 0)) +
		valueSize(BSONSourceDB, tableTwoName, "_id",
			NewSQLInt64(MySQLValueKind, 0))

	t.Run("inner_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(criteria, InnerJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize + rightSize) * 4

		req.Equal(expected, actual)
	})

	t.Run("left_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(criteria, LeftJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize+rightSize)*4 +
			(leftSize + rightSize - 24) // nil right values

		req.Equal(expected, actual)
	})

	t.Run("right_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(criteria, RightJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize+rightSize)*4 +
			(leftSize - 23 + rightSize) // nil left values

		req.Equal(expected, actual)
	})

	t.Run("cross_join", func(t *testing.T) {
		req := require.New(t)
		operator := setupJoinOperator(nil, CrossJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := uint64(len(customers)) * uint64(len(orders)) * (leftSize + rightSize)

		req.Equal(expected, actual)
	})
}

func testRowGeneratorMemoryMonitor(t *testing.T) {
	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3)),
	)

	bss := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

	newColumn := NewColumn(0, "", "", "", "a", "", "a",
		EvalUint64, schema.MongoInt64, false, true)

	rg := NewRowGeneratorStage(bss, newColumn)

	outputMemory := getAllocatedMemorySizeAfterIteration(rg)

	// empty rows cost nothing
	require.Equal(t, uint64(0), outputMemory)
}

func testFilterMemoryMonitor(t *testing.T) {
	schema := MustLoadSchema(testSchema3)
	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", 4)),
	)

	matcher, err := GetSQLExpr(schema,
		BSONSourceDB,
		tableTwoName,
		"a = 6",
		false,
		nil)
	require.NoError(t, err)

	bss := NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
	filter := NewFilterStage(bss, matcher)

	actual := getAllocatedMemorySizeAfterIteration(filter)

	sizeA := valueSize(
		BSONSourceDB, tableTwoName, "a",
		NewSQLInt64(MySQLValueKind, 0),
	)
	sizeB := valueSize(
		BSONSourceDB, tableTwoName, "b",
		NewSQLInt64(MySQLValueKind, 0),
	)
	expected := sizeA + sizeB

	require.Equal(t, expected, actual)
}

func testGroupByMemoryMonitor(t *testing.T) {
	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", "a"), bsonutil.NewDocElem("b", 7)),
		bsonutil.NewD(bsonutil.NewDocElem("a", "a"), bsonutil.NewDocElem("b", 8)),
		bsonutil.NewD(bsonutil.NewDocElem("a", "b"), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", "b"), bsonutil.NewDocElem("b", 10)),
		bsonutil.NewD(bsonutil.NewDocElem("a", "b"), bsonutil.NewDocElem("b", 11)),
	)

	projectedColumns := ProjectedColumns{
		ProjectedColumn{
			Column: &Column{SelectID: 1, Table: tableOneName,
				OriginalTable: tableOneName, Database: BSONSourceDB, Name: "a",
				OriginalName: "a", MappingRegistryName: "",
				ColumnType: &ColumnType{
					EvalType:  EvalString,
					MongoType: schema.MongoString,
				},
				PrimaryKey: false},
			Expr: testSQLColumnExpr(1, BSONSourceDB, tableOneName, "a",
				EvalString, schema.MongoString, false),
		},
		ProjectedColumn{
			Column: &Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "",
				ColumnType: &ColumnType{
					EvalType:  EvalInt64,
					MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: NewSQLAggregationFunctionExpr(
				"sum",
				false,
				[]SQLExpr{
					testSQLColumnExpr(1, BSONSourceDB, tableOneName, "b",
						EvalInt64, schema.MongoInt, false),
				},
			),
		},
	}

	keys := []SQLExpr{testSQLColumnExpr(1, BSONSourceDB,
		tableOneName, "a", EvalString, schema.MongoString, false)}

	bss := NewBSONSourceStage(1, tableOneName, collation.Default, rows)
	groupBy := NewGroupByStage(bss, keys, projectedColumns)

	actual := getAllocatedMemorySizeAfterIteration(groupBy)

	sizeA := valueSize(
		BSONSourceDB, tableOneName, "a",
		NewSQLVarchar(MySQLValueKind, "a"),
	)
	sizeB := valueSize(
		BSONSourceDB, "", "sum(b)",
		NewSQLInt64(MySQLValueKind, 0),
	)
	expected := 2 * (sizeA + sizeB)

	require.Equal(t, expected, actual)
}

func testLimitMemoryMonitor(t *testing.T) {
	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", 4)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 0), bsonutil.NewDocElem("b", 13)),
		bsonutil.NewD(bsonutil.NewDocElem("a", -3), bsonutil.NewDocElem("b", 8)),
		bsonutil.NewD(bsonutil.NewDocElem("a", -6), bsonutil.NewDocElem("b", 17)),
		bsonutil.NewD(bsonutil.NewDocElem("a", -9), bsonutil.NewDocElem("b", 12)),
	)

	t.Run("non-blocking source", func(t *testing.T) {
		bss := NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		ls := NewLimitStage(bss, 2, 2)

		actual := getAllocatedMemorySizeAfterIteration(ls)

		sizeA := valueSize(
			BSONSourceDB, tableTwoName, "a",
			NewSQLInt64(MySQLValueKind, 0),
		)
		sizeB := valueSize(
			BSONSourceDB, tableTwoName, "b",
			NewSQLInt64(MySQLValueKind, 0),
		)
		expected := 2 * (sizeA + sizeB)

		require.Equal(t, expected, actual)
	})
	t.Run("blocking source", func(t *testing.T) {
		bss := NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		os := NewOrderByStage(bss,
			NewOrderByTerm(
				testSQLColumnExpr(
					1,
					BSONSourceDB,
					tableTwoName,
					"a",
					EvalInt64,
					schema.MongoInt,
					false),
				true))
		ls := NewLimitStage(os, 2, 2)

		actual := getAllocatedMemorySizeAfterIteration(ls)

		sizeA := valueSize(
			BSONSourceDB, tableTwoName, "a",
			NewSQLInt64(MySQLValueKind, 0),
		)
		sizeB := valueSize(
			BSONSourceDB, tableTwoName, "b",
			NewSQLInt64(MySQLValueKind, 0),
		)
		expected := 4 * (sizeA + sizeB)

		require.Equal(t, expected, actual)
	})
}

func testDynamicSourceMemoryMonitor(t *testing.T) {
	tableName := "foo"

	table := catalog.NewDynamicTable(tableName, catalog.BaseTable, func(string) RowIter {
		// Create an empty Channel
		rowChan := make(chan Row, DefaultRowChannelBufSize)
		done := make(chan struct{})
		rows := Rows{
			NewNamelessRow(NewSQLInt64(MongoSQLValueKind, 1), NewSQLInt64(MongoSQLValueKind, 2)),
			NewNamelessRow(NewSQLInt64(MongoSQLValueKind, 2), NewSQLInt64(MongoSQLValueKind, 3)),
			NewNamelessRow(NewSQLInt64(MongoSQLValueKind, 3), NewSQLInt64(MongoSQLValueKind, 4)),
		}
		go func() {
			defer close(rowChan)
			for _, row := range rows {
				select {
				case rowChan <- row:
				case <-done:
					return
				}
			}
		}()
		return NewRowChanIter(rowChan, done)
	})

	_, err := table.AddColumn(tableName, "one", EvalInt64)
	require.NoError(t, err)
	_, err = table.AddColumn(tableName, "two", EvalInt64)
	require.NoError(t, err)

	db, err := catalog.New("def").AddDatabase("db")
	require.NoError(t, err)

	source := NewDynamicSourceStage(db, table, 1, tableName)

	actual := getAllocatedMemorySizeAfterIteration(source)
	expected := uint64(0x60)

	require.Equal(t, expected, actual)
}

func testProjectMemoryMonitor(t *testing.T) {
	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", 4)),
	)

	bss := NewBSONSourceStage(1, tableOneName, collation.Default, rows)
	project := NewProjectStage(bss, ProjectedColumn{
		Column: &Column{SelectID: 1, Table: tableOneName, OriginalTable: tableOneName,
			Database: BSONSourceDB, Name: "a", OriginalName: "a",
			MappingRegistryName: "",
			ColumnType: &ColumnType{
				EvalType: EvalInt64, MongoType: schema.MongoInt,
			},
			PrimaryKey: false},
		Expr: testSQLColumnExpr(1, BSONSourceDB, tableOneName, "a",
			EvalInt64, schema.MongoInt, false),
	})

	actual := getAllocatedMemorySizeAfterIteration(project)
	expected := valueSize(
		BSONSourceDB,
		tableOneName,
		"a",
		NewSQLInt64(MySQLValueKind, 0)) * uint64(len(rows))

	require.Equal(t, expected, actual)
}

func testUnionMemoryMonitor(t *testing.T) {
	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", 4)),
	)

	u := NewUnionStage(UnionDistinct,
		NewBSONSourceStage(1, "foo", collation.Default, rows),
		NewBSONSourceStage(2, "bar", collation.Default, rows),
	)

	actual := getAllocatedMemorySizeAfterIteration(u)

	sizeA := valueSize(
		BSONSourceDB, tableOneName, "a",
		NewSQLInt64(MySQLValueKind, 0),
	)
	sizeB := valueSize(
		BSONSourceDB, tableOneName, "b",
		NewSQLInt64(MySQLValueKind, 0),
	)
	expected := uint64(len(rows)*2) * (sizeA + sizeB)

	require.Equal(t, expected, actual)
}
