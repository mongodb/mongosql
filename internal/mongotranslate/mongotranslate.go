package mongotranslate

import (
	"context"
	"fmt"
	"log"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/variable"
	sqlLgr "github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
)

// TranslateSQLQuery takes a sql query (as a string) and returns a MongoDB
// aggregation pipeline as a string. This function also takes a defaultDB
// that is used as the default database for unqualified tables in the query,
// and an mdbVersion that is used as the MongoDB version for the aggregation
// language.
// If the showInferredSchema argument is true, this function will log the schema.
func TranslateSQLQuery(sqlQuery, defaultDB, mdbVersion string, showInferredSchema bool) (string, error) {
	// unconditionally prepend "explain" to the query. If the sqlQuery is
	// unexplainable (for example, a command), an error will be returned.
	sqlQuery = "explain " + sqlQuery

	is, err := InferSchemaFromQuery(sqlQuery, defaultDB)
	if err != nil {
		return "", fmt.Errorf("fatal error inferring schema: %v", err)
	}

	if showInferredSchema {
		log.Printf("*** Inferred Schema ***\n%s\n", is.String())
	}

	ctlg, err := getCatalog(mdbVersion, is)
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

// getCatalog copies the inferred schema into a Catalog and returns it.
func getCatalog(mdbVersion string, is InferredSchema) (catalog.Catalog, error) {
	gbl := variable.NewGlobalContainer(nil)
	gbl.MongoDBVersion = mdbVersion
	gbl.PolymorphicTypeConversionMode = string(variable.PolymorphicTypeConversionModeOff)
	gbl.SetSystemVariable(variable.TypeConversionMode, string(variable.MongoSQLTypeConversionMode))

	vars := variable.NewSessionContainer(gbl)
	vars.MongoDBVersion = mdbVersion
	vars.PolymorphicTypeConversionMode = string(variable.PolymorphicTypeConversionModeOff)
	vars.SetSystemVariable(variable.TypeConversionMode, string(variable.MongoSQLTypeConversionMode))

	ctlg := catalog.New("", vars)

	lgr := sqlLgr.GlobalLogger()

	var db catalog.Database
	var err error

	// Populate the catalog with the inferred schema
	for _, database := range is.Databases() {
		db, err = ctlg.AddDatabase(database)
		if err != nil {
			return nil, err
		}

		tableNames, _ := is.Tables(database)
		for _, tableName := range tableNames {
			table, err := schema.NewTable(lgr, tableName, tableName, nil, nil)
			if err != nil {
				return nil, err
			}

			columns, _ := is.Columns(database, tableName)
			for _, columnName := range columns {
				column := schema.NewColumn(columnName, schema.SQLInt, columnName, schema.MongoInt)
				table.AddColumn(lgr, column, false)
			}

			err = db.AddTable(catalog.NewMongoTable(table, catalog.BaseTable, collation.Default))
			if err != nil {
				return nil, err
			}
		}
	}

	return ctlg, nil
}
