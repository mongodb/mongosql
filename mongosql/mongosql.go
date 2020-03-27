package mongosql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/10gen/mongoast/parser"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// TranslateSQLQuery translates a SQL query into a MongoDB aggregation pipeline.
// It also returns the name of the collection against which the pipeline would
// run. It accepts the sql query as a string, along with a dbName to use as the
// default database for unqualified tables in the query, a mongoVersion to use
// as the MongoDB version for the aggregation language, and a schemaPath which
// should be a path to a file or directory containing the DRDL schema to use for
// translation. If the format argument is "multiline", the pipeline string will
// have newlines between stages. If explain is true, the query's explain output
// is returned instead of just the aggregation pipeline.
func TranslateSQLQuery(sqlQuery, dbName, mongoVersion, schemaPath, format string, explain bool) (string, string, error) {
	sch, err := loadSchema(schemaPath)
	if err != nil {
		return "", "", fmt.Errorf("fatal error getting schema from drdl: %v", err)
	}

	// unconditionally prepend "explain" to the query. If the sqlQuery is
	// unexplainable (for example, a command), an error will be returned.
	sqlQuery = "explain " + sqlQuery

	ctlg, err := getCatalog(sch)
	if err != nil {
		return "", "", fmt.Errorf("fatal error creating catalog: %v", err)
	}

	mdbVersion, err := mongoDBVersionAsSlice(mongoVersion)
	if err != nil {
		return "", "", fmt.Errorf("invalid MongoDB version (%v): %v", mongoVersion, err)
	}

	tCfg := newTranslationConfig(ctlg, evaluator.NoOutputFormat, evaluator.NoOutputVersion,
		dbName, mdbVersion, false, false, true, true, true, false, false, false)
	qCfg := NewQueryConfigFromTranslationConfig(tCfg)

	res, err := evaluator.ExecuteSQL(context.Background(), qCfg, sqlQuery)
	if err != nil {
		return "", "", fmt.Errorf("fatal error executing sql %q: %v", sqlQuery, err)
	}

	if explain {
		return getExplainOutput(format, res.Stats.Explain)
	}

	// check for any PushdownFailures.
	if !res.Stats.FullyPushedDown {
		return "", "", errors.New("query not fully pushed down; run with --explain for more details")
	}

	pipeline := res.Stats.Explain[0].Pipeline.Else("")
	collection := getCollectionName(res.Stats.Explain[0])

	if format == "multiline" {
		return prettyFormat(pipeline), collection, nil
	}

	return pipeline, collection, nil
}

// TranslateSQLQueryFile translates a SQL query into a MongoDB aggregation pipeline.
// It also returns the name of the collection against which the pipeline would run.
// It accepts the sql query from a file, along with a dbName to use as the default
// database for unqualified tables in the query, a mongoVersion to use as the MongoDB
// version for the aggregation language, and a schemaPath which should be a path to a
// file or directory containing the DRDL schema to use for translation. If the format
// argument is "multiline", the pipeline string will have newlines between stages. If
// explain is true, the query's explain output is returned instead of just the
// aggregation pipeline.
func TranslateSQLQueryFile(queryFile, dbName, mongoVersion, schemaPath, format string, explain bool) (string, string, error) {
	query, err := ioutil.ReadFile(queryFile)
	if err != nil {
		return "", "", fmt.Errorf("could not open file %v", queryFile)
	}

	return TranslateSQLQuery(string(query), dbName, mongoVersion, schemaPath, format, explain)
}

// TranslateSQLQueryRaw translates a SQL query into a MongoDB aggregation pipeline
// as a raw bsoncore.Array. It also returns the name of the collection against which
// the pipeline would run. It accepts the sql query as a string, along with a dbName
// to use as the default database for unqualified tables in the query, a mongoVersion
// to use as the MongoDB version for the aggregation language, and a schema to use for
// translation.
func TranslateSQLQueryRaw(
	ctx context.Context,
	cfg *TranslationConfig,
	sqlQuery string,
) (bsoncore.Array, string, string, error) {
	qCfg := NewQueryConfigFromTranslationConfig(cfg)

	res, err := evaluator.ExecuteSQL(ctx, qCfg, sqlQuery)
	if err != nil {
		return nil, "", "", fmt.Errorf("fatal error executing sql %q: %v", sqlQuery, err)
	}

	if !res.Stats.FullyPushedDown {
		return nil, "", "", errors.New("query not fully pushed down")
	}

	ms, ok := res.PlanStage.(*evaluator.MongoSourceStage)
	if !ok {
		return nil, "", "", errors.New("query not fully pushed down")
	}

	pipeline := parser.DeparsePipeline(ms.Pipeline()).Array()
	database := ms.Database()
	collection := ms.Collection()

	return pipeline, database, collection, nil
}

func getExplainOutput(format string, explainRecords []*evaluator.ExplainRecord) (string, string, error) {
	var explainOutput string
	var jsonExplain []byte
	var err error
	if format == "multiline" {
		jsonExplain, err = jsonMarshal(explainRecords, "\t")
	} else {
		jsonExplain, err = jsonMarshal(explainRecords, "")
	}
	if err != nil {
		return "", "", fmt.Errorf("fatal error marshalling explain record to json: %v", err)
	}
	explainOutput = fmt.Sprintf("%v%+v", explainOutput, string(jsonExplain))

	collection := getCollectionName(explainRecords[len(explainRecords)-1])

	return strings.TrimSuffix(explainOutput, "\n"), collection, nil
}

func jsonMarshal(t interface{}, indent string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func prettyFormat(pipeline string) string {
	formattedPipeline := fmt.Sprintf("[\n")
	stages := getStages(pipeline)
	for _, stage := range stages {
		formattedPipeline = fmt.Sprintf("%v\t%v,\n", formattedPipeline, stage)
	}
	return fmt.Sprintf("%v]", formattedPipeline)
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

	relationalSchema, err := schema.NewFromDRDL(lgr, drdlSchema, false)
	if err != nil {
		return nil, err
	}

	return relationalSchema, nil
}

func mongoDBVersionAsSlice(version string) ([]uint8, error) {
	if version == "latest" {
		version = "100.0.0"
	}

	return procutil.VersionToSlice(version)
}

// getVariables constructs a variable container for mongosql.
func getVariables(mongoVersion string) *variable.Container {
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

	return vars
}

// getCatalog copies the schema into a Catalog and returns it.
func getCatalog(relationalSchema *schema.Schema) (catalog.Catalog, error) {
	ctlg := catalog.New("")

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
			err = db.AddTable(catalog.NewMongoTable(string(db.Name()), table, catalog.BaseTable, collation.Default, false))
			if err != nil {
				return nil, err
			}
		}
	}

	return ctlg, nil
}

func getCollectionName(er *evaluator.ExplainRecord) string {
	// Collections is of the shape "[collectionName<,collectionName>]"
	collectionNames := er.Collections.Else("[,]")
	return strings.Split(collectionNames[1:len(collectionNames)-1], ",")[0]
}
