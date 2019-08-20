package mongotranslate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
)

// TranslateSQLQuery takes a sql query (as a string or from a file) and
// returns a MongoDB aggregation pipeline as a string. This function also takes
// a dbName that is used as the default database for unqualified tables in the query,
// and a mongoVersion that is used as the MongoDB version for the aggregation language.
func TranslateSQLQuery(sqlQuery, queryFile, dbName, mongoVersion, schemaPath, format string, explain bool) (string, error) {
	if queryFile != "" {
		query, err := ioutil.ReadFile(queryFile)
		if err != nil {
			return "", fmt.Errorf("Could not open file %v", queryFile)
		}
		sqlQuery = string(query)
	}

	// unconditionally prepend "explain" to the query. If the sqlQuery is
	// unexplainable (for example, a command), an error will be returned.
	sqlQuery = "explain " + sqlQuery

	schema, err := loadSchema(schemaPath)
	if err != nil {
		return "", fmt.Errorf("fatal error getting schema from drdl: %v", err)
	}

	ctlg, err := getCatalog(mongoVersion, schema)
	if err != nil {
		return "", fmt.Errorf("fatal error creating catalog: %v", err)
	}

	qCfg := evaluator.NewDefaultQueryConfig(mongoVersion, dbName, ctlg)

	res, err := evaluator.ExecuteSQL(context.Background(), qCfg, sqlQuery)
	if err != nil {
		return "", fmt.Errorf("fatal error executing sql %q: %v", sqlQuery, err)
	}

	if explain {
		return getExplainOutput(format, res.Stats.Explain)
	}

	// check for any PushdownFailures.
	if !res.Stats.FullyPushedDown {
		return "", fmt.Errorf("query not fully pushed down; run with --explain for more details")
	}

	if format == "multiline" {
		return prettyFormat(res.Stats.Explain[0].Pipeline.Else("")), nil
	}
	return res.Stats.Explain[0].Pipeline.Else(""), nil
}

func jsonMarshal(t interface{}, indent string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func getExplainOutput(format string, explainRecords []*evaluator.ExplainRecord) (string, error) {
	var explainOutput string
	var jsonExplain []byte
	var err error
	if format == "multiline" {
		jsonExplain, err = jsonMarshal(explainRecords, "\t")
	} else {
		jsonExplain, err = jsonMarshal(explainRecords, "")
	}
	if err != nil {
		return "", fmt.Errorf("fatal error marshalling explain record to json: %v", err)
	}
	explainOutput = fmt.Sprintf("%v%+v", explainOutput, string(jsonExplain))
	return strings.TrimSuffix(explainOutput, "\n"), nil
}

func getStages(pipeline string) []string {
	var i, j, openCount, closedCount, nextBracket int
	var results []string

	pipeline = pipeline[1:]
	for {
		nextBracket = strings.IndexAny(pipeline[j:], "{}")
		if nextBracket == -1 {
			break
		}
		j += nextBracket

		if pipeline[j] == '{' {
			openCount++
		} else {
			closedCount++
			if openCount == closedCount {
				results = append(results, pipeline[i:j+1])
				i = j + 2
			}
		}
		j++
	}

	return results
}

func prettyFormat(pipeline string) string {
	formattedPipeline := fmt.Sprintf("[\n")
	stages := getStages(pipeline)
	for _, stage := range stages {
		formattedPipeline = fmt.Sprintf("%v\t%v,\n", formattedPipeline, stage)
	}
	return fmt.Sprintf("%v]", formattedPipeline)
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
func getCatalog(mongoVersion string, relationalSchema *schema.Schema) (catalog.Catalog, error) {
	gbl := variable.NewGlobalContainer(nil)
	if mongoVersion == "latest" {
		// set to an arbitrary large number so we don't have to update it every time we add features from a new MongoDB version
		mongoVersion = "100.0.0"
	}

	gbl.SetSystemVariable(variable.MongoDBVersion,
		values.NewSQLVarchar(values.VariableSQLValueKind, mongoVersion))
	gbl.SetSystemVariable(variable.PolymorphicTypeConversionMode,
		values.NewSQLVarchar(values.VariableSQLValueKind, variable.OffPolymorphicTypeConversionMode))
	gbl.SetSystemVariable(variable.TypeConversionMode,
		values.NewSQLVarchar(values.VariableSQLValueKind, variable.MongoSQLTypeConversionMode))

	vars := variable.NewSessionContainer(gbl)
	vars.SetSystemVariable(variable.MongoDBVersion,
		values.NewSQLVarchar(values.VariableSQLValueKind, mongoVersion))
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
			tbl, err := schema.NewTable(lgr, table.SQLName(), table.MongoName(),
				nil, nil, []schema.Index{}, option.NoneString())
			if err != nil {
				return nil, err
			}

			columns := table.Columns()
			for _, column := range columns {
				col := schema.NewColumn(column.SQLName(), column.SQLType(),
					column.MongoName(), column.MongoType(), false, option.NoneString())
				tbl.AddColumn(lgr, col, false)
			}

			err = db.AddTable(catalog.NewMongoTable(string(db.Name()), tbl, catalog.BaseTable, collation.Default, false))
			if err != nil {
				return nil, err
			}
		}
	}

	return ctlg, nil
}
