package mongotranslate

import (
	"context"
	"fmt"
	"os"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
)

// TranslateSQLQuery takes a sql query (as a string) and returns a MongoDB
// aggregation pipeline as a string. This function also takes a defaultDB
// that is used as the default database for unqualified tables in the query,
// and an mdbVersion that is used as the MongoDB version for the aggregation
// language.
// If the showInferredSchema argument is true, this function will log the schema.
func TranslateSQLQuery(sqlQuery, defaultDB, mdbVersion, schemaPath string) (string, error) {
	// unconditionally prepend "explain" to the query. If the sqlQuery is
	// unexplainable (for example, a command), an error will be returned.
	sqlQuery = "explain " + sqlQuery

	schema, err := loadSchema(schemaPath)
	if err != nil {
		return "", fmt.Errorf("fatal error getting schema from drdl: %v", err)
	}

	ctlg, err := getCatalog(mdbVersion, schema)
	if err != nil {
		return "", fmt.Errorf("fatal error creating catalog: %v", err)
	}

	qCfg := evaluator.NewDefaultQueryConfig(mdbVersion, defaultDB, ctlg)

	res, err := evaluator.ExecuteSQL(context.Background(), qCfg, sqlQuery)
	if err != nil {
		return "", fmt.Errorf("fatal error executing sql %q: %v", sqlQuery, err)
	}

	// check for any PushdownFailures.
	for _, e := range res.Stats.Explain {
		// this is necessary for queries such as 'select t.a, s.b from db1.t, db2.s'
		// which produce res.Stats.Explain slices of length > 1 and have PushdownFailures
		// at res.Stats.Explain[ index > 0 ].
		if len(e.PushdownFailures) > 0 {
			return "", fmt.Errorf("query not fully pushed down: %v", e.PushdownFailures[0])
		}
	}

	return res.Stats.Explain[0].Pipeline.Else(""), nil
}

func loadSchema(path string) (*schema.Schema, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var drdlSchema *drdl.Schema

	if fi.IsDir() {
		drdlSchema, err = drdl.NewFromDir(path)
		if err != nil {
			return nil, err
		}
	} else {
		drdlSchema, err = drdl.NewFromFile(path)
		if err != nil {
			return nil, err
		}
	}

	lgr := log.GlobalLogger()

	relationalSchema, err := schema.NewFromDRDL(lgr, drdlSchema)
	if err != nil {
		return nil, err
	}

	return relationalSchema, nil
}

// getCatalog copies the schema into a Catalog and returns it.
func getCatalog(mdbVersion string, relationalSchema *schema.Schema) (catalog.Catalog, error) {
	gbl := variable.NewGlobalContainer(nil)
	gbl.SetSystemVariable(variable.MongoDBVersion,
		values.NewSQLVarchar(values.VariableSQLValueKind, mdbVersion))
	gbl.SetSystemVariable(variable.PolymorphicTypeConversionMode,
		values.NewSQLVarchar(values.VariableSQLValueKind, variable.OffPolymorphicTypeConversionMode))
	gbl.SetSystemVariable(variable.TypeConversionMode,
		values.NewSQLVarchar(values.VariableSQLValueKind, variable.MongoSQLTypeConversionMode))

	vars := variable.NewSessionContainer(gbl)
	vars.SetSystemVariable(variable.MongoDBVersion,
		values.NewSQLVarchar(values.VariableSQLValueKind, mdbVersion))
	vars.SetSystemVariable(variable.PolymorphicTypeConversionMode,
		values.NewSQLVarchar(values.VariableSQLValueKind, variable.OffPolymorphicTypeConversionMode))
	vars.SetSystemVariable(variable.TypeConversionMode,
		values.NewSQLVarchar(values.VariableSQLValueKind, variable.MongoSQLTypeConversionMode))

	ctlg := catalog.New("", vars)

	lgr := log.GlobalLogger()

	var db catalog.Database
	var err error

	// Populate the catalog with the schema
	for _, database := range relationalSchema.Databases() {
		db, err = ctlg.AddDatabase(database.Name())
		if err != nil {
			return nil, err
		}

		tables := database.Tables()
		for _, table := range tables {
			tbl, err := schema.NewTable(lgr, table.SQLName(), table.MongoName(), nil, nil)
			if err != nil {
				return nil, err
			}

			columns := table.Columns()
			for _, column := range columns {
				col := schema.NewColumn(column.SQLName(), column.SQLType(), column.MongoName(), column.MongoType())
				tbl.AddColumn(lgr, col, false)
			}

			err = db.AddTable(catalog.NewMongoTable(string(db.Name()), tbl, catalog.BaseTable, collation.Default))
			if err != nil {
				return nil, err
			}
		}
	}

	return ctlg, nil
}
