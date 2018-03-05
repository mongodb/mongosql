package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/stretchr/testify/require"
)

func TestMongoTable(t *testing.T) {
	req := require.New(t)

	drdlSchema, err := drdl.NewFromBytes(testSchema)
	req.NoError(err, "failed to load drdl")

	config, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema)
	req.NoError(err, "failed to create schema from drdl")

	db := config.Databases()[0]
	tbl := db.TablesSorted()[1]

	mt := catalog.NewMongoTable(tbl, catalog.BaseTable, collation.Default)

	req.Equal("foo", string(mt.Name()), "incorrect sql name for table")
	req.Equal("fooCollection", mt.CollectionName, "incorrect collection name for table")
	req.Len(mt.Columns(), 4, "incorrect column count")

	column, err := mt.Column("id")
	req.NoError(err, "failed to get column")
	req.Equal("id", string(column.Name()), "incorrect column name")
	req.Equal("_id", column.(*catalog.MongoColumn).MongoName, "incorrect mongo field name")

	column, err = mt.Column("value")
	req.NoError(err, "failed to get column")
	req.Equal("value", string(column.Name()), "incorrect column name")
	req.Equal("a", column.(*catalog.MongoColumn).MongoName, "incorrect mongo field name")

	column, err = mt.Column("idx1")
	req.NoError(err, "failed to get column")
	req.Equal("idx1", string(column.Name()), "incorrect column name")
	req.Equal("a_idx", column.(*catalog.MongoColumn).MongoName, "incorrect mongo field name")

	column, err = mt.Column("idx2")
	req.NoError(err, "failed to get column")
	req.Equal("idx2", string(column.Name()), "incorrect column name")
	req.Equal("a_idx_1", column.(*catalog.MongoColumn).MongoName, "incorrect mongo field name")
}
