package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestGroupByPlanStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	runTest := func(t *testing.T,
		projectedColumns evaluator.ProjectedColumns,
		keys []evaluator.SQLExpr,
		rows []bson.D, expectedRows []evaluator.Values) {

		req := require.New(t)

		bss := evaluator.NewBSONSourceStage(1, tableOneName,
			collation.Must(collation.Get("utf8_general_ci")), rows)

		groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		row := &evaluator.Row{}

		i := 0

		for iter.Next(bgCtx, row) {
			req.Equal(len(row.Data), len(expectedRows[i]))
			req.Equal(row.Data, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		req.NoError(iter.Close())
		req.NoError(iter.Err())
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

	expected := []evaluator.Values{
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
			Data: evaluator.NewSQLVarchar(evaluator.MySQLValueKind, "a")}, {SelectID: 1,
			Database: evaluator.BSONSourceDB, Table: "", Name: "sum(b)",
			Data: evaluator.NewSQLFloat(evaluator.MySQLValueKind, 15)}},
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
			Data: evaluator.NewSQLVarchar(evaluator.MySQLValueKind, "b")}, {SelectID: 1,
			Database: evaluator.BSONSourceDB, Table: "", Name: "sum(b)",
			Data: evaluator.NewSQLFloat(evaluator.MySQLValueKind, 9)}},
	}

	runTest(t, projectedColumns, keys, data, expected)
}
