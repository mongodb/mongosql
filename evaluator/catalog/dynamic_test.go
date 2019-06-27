package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/stretchr/testify/require"
)

func TestDynamic(t *testing.T) {
	req := require.New(t)

	testTable := catalog.NewDynamicTable("bar", "foo", func(string) results.Rows {
		out := make(results.Rows, 0)
		for i := int64(0); i < 10; i += 2 {
			out = append(out, results.NewNamedRow("bar", "foo",
				values.NewNamedSQLValue("one", values.NewSQLInt64(values.MongoSQLValueKind, i)),
				values.NewNamedSQLValue("two", values.NewSQLInt64(values.MongoSQLValueKind, i+1)),
			))
		}
		return out
	})

	_, err := testTable.AddColumn("foo", "one", types.EvalInt64)
	req.Nil(err)
	_, err = testTable.AddColumn("foo", "two", types.EvalInt64)
	req.Nil(err)

	req.Equal(testTable.Columns()[0].Name, "one")
	req.Equal(testTable.Columns()[0].SelectID, 1)
	req.Equal(testTable.Columns()[1].Name, "two")
	req.Equal(testTable.Columns()[1].SelectID, 2)

	rows := testTable.Rows("")
	for i := int64(0); i < int64(len(rows)); i += 2 {
		req.Equal(rows[i].Data[0].SelectID, 1)
		req.Equal(rows[i].Data[1].SelectID, 2)
		req.Equal(rows[i].Data[0].Name, "one")
		req.Equal(rows[i].Data[1].Name, "two")
	}
}
