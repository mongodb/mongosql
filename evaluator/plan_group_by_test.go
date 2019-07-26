package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/10gen/sqlproxy/evaluator"
	. "github.com/10gen/sqlproxy/evaluator/results"
	. "github.com/10gen/sqlproxy/evaluator/types"
	. "github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGroupByPlanStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg(MySQLValueKind)
	execState := NewExecutionState()

	runTest := func(t *testing.T,
		projectedColumns ProjectedColumns,
		keys []SQLExpr,
		rows []bson.D, expectedRows []RowValues) {

		req := require.New(t)

		bss := NewBSONSourceStage(1, tableOneName,
			collation.Must(collation.Get("utf8_general_ci")), rows)

		groupBy := NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		row := &Row{}

		i := 0

		for iter.Next(bgCtx, row) {
			req.Equal(len(row.Data), len(expectedRows[i]))
			req.Equal(row.Data, expectedRows[i])
			row = &Row{}
			i++
		}

		req.NoError(iter.Close())
		req.NoError(iter.Err())
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
				ColumnType: ColumnType{
					EvalType:  EvalString,
					MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: testSQLColumnExpr(1, BSONSourceDB, tableOneName, "a",
				EvalString, schema.MongoString, false),
		},
		ProjectedColumn{
			Column: &Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
				MappingRegistryName: "",
				ColumnType: ColumnType{
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

	expected := []RowValues{
		{{SelectID: 1, Database: BSONSourceDB, Table: tableOneName, Name: "a",
			Data: NewSQLVarchar(MySQLValueKind, "a")}, {SelectID: 1,
			Database: BSONSourceDB, Table: "", Name: "sum(b)",
			Data: NewSQLFloat(MySQLValueKind, 15)}},
		{{SelectID: 1, Database: BSONSourceDB, Table: tableOneName, Name: "a",
			Data: NewSQLVarchar(MySQLValueKind, "b")}, {SelectID: 1,
			Database: BSONSourceDB, Table: "", Name: "sum(b)",
			Data: NewSQLFloat(MySQLValueKind, 9)}},
	}

	runTest(t, projectedColumns, keys, data, expected)
}
