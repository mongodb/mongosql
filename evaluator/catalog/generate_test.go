package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/stretchr/testify/require"
)

func TestGenerateCreateTable(t *testing.T) {
	req := require.New(t)

	drdlSchema, err := drdl.NewFromBytes(testSchema)
	req.NoError(err, "failed to load drdl")

	config, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema)
	req.NoError(err, "failed to create schema from drdl")

	db := config.Databases()[0]
	tbls := db.TablesSorted()

	tblConfig := tbls[1]
	tbl := catalog.NewMongoTable("", tblConfig, catalog.BaseTable, collation.Default)
	createTable := catalog.GenerateCreateTable(tbl, 0)
	req.Equal(testSchemaCreateTableFoo, createTable, "create table statement is incorrect")

	tblConfig = tbls[0]
	tbl = catalog.NewMongoTable("", tblConfig, catalog.BaseTable, collation.Default)
	createTable = catalog.GenerateCreateTable(tbl, 10)
	req.Equal(testSchemaCreateTableBar, createTable, "create table statement is incorrect")
}
