package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/stretchr/testify/require"
)

func TestMongoTable(t *testing.T) {
	req := require.New(t)

	drdlSchema, err := drdl.NewFromBytes(testSchema)
	req.NoError(err, "failed to load drdl")

	config, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, false)
	req.NoError(err, "failed to create schema from drdl")

	db := config.Databases()[0]
	tbl := db.TablesSorted()[1]

	tests := []struct {
		desc            string
		isCaseSensitive bool
	}{
		{"case insensitive", false},
		{"case sensitive", true},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			mt := catalog.NewMongoTable(db.Name(), tbl, catalog.BaseTable, collation.Default, false, test.isCaseSensitive)

			req.Equal("foo", mt.Name(), "incorrect sql name for table")
			req.Equal("fooCollection", mt.Collection(), "incorrect collection name for table")
			req.Len(mt.Columns(), 4, "incorrect column count")

			var column *results.Column
			column, err = mt.Column("id")
			req.NoError(err, "failed to get column")
			req.Equal("id", column.Name, "incorrect column name")
			req.Equal("_id", column.MongoName, "incorrect mongo field name")

			column, err = mt.Column("value")
			req.NoError(err, "failed to get column")
			req.Equal("value", column.Name, "incorrect column name")
			req.Equal("a", column.MongoName, "incorrect mongo field name")

			column, err = mt.Column("idx1")
			req.NoError(err, "failed to get column")
			req.Equal("idx1", column.Name, "incorrect column name")
			req.Equal("a_idx", column.MongoName, "incorrect mongo field name")

			column, err = mt.Column("IDX2")
			if test.isCaseSensitive {
				req.Error(err)
			} else {
				req.NoError(err, "failed to get column")
				req.Equal("idx2", column.Name, "incorrect column name")
				req.Equal("a_idx_1", column.MongoName, "incorrect mongo field name")
			}
		})
	}
}
