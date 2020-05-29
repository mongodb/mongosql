package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/10gen/sqlproxy/mongosql"
)

const (
	defaultDbName       = "test"
	defaultMongoVersion = "latest"
	defaultFormat       = "multiline"
	currentVersion      = "1.0.0-beta1"
)

func main() {
	sqlQuery := flag.String("query", "", "The sql query to translate into MongoDB aggregation language.")
	schema := flag.String("schema", "", "The path to a DRDL file or a directory containing DRDL files.")
	dbName := flag.String("dbName", defaultDbName, "The database name to use for unqualified tables in the query.")
	mongoVersion := flag.String("mongoVersion", defaultMongoVersion, "The MongoDB version to which to translate the query.")
	format := flag.String("format", defaultFormat, `The desired formatting for the output. The flag can be set to "multiline" (formats output with one pipeline stage per line) or "none" (no formatting at all).`)
	explain := flag.Bool("explain", false, "Returns the explain output for the query rather than the pipeline output.")
	version := flag.Bool("version", false, "Prints the current version of mongotranslate.")
	queryFile := flag.String("queryFile", "", "The path to a text file containing the query to translate into MongoDB aggregation language.")

	flag.Parse()

	if *version {
		fmt.Println(currentVersion)
		return
	}
	if *sqlQuery == "" && *queryFile == "" {
		log.Fatalln("no query provided")
		return
	}
	if *sqlQuery != "" && *queryFile != "" {
		log.Fatalln("cannot supply both a query and queryFile flag")
	}
	if *schema == "" {
		log.Fatalln("no schema provided")
		return
	}
	// The adl format option is only for internal use, it should not be exposed to users.
	// The purpose is to aid in issues with ADL pushdown.
	if *format != "multiline" && *format != "none" && *format != "adl" {
		log.Fatalf("invalid value `%v` for option `--format`. Allowed values are: multiline, none.\n", *format)
		return
	}

	var explainPlan string
	var err error

	switch {
	case *format == "adl":
		explainPlan, err = mongosql.ADLTranslate(*sqlQuery, *dbName, *schema)
	case *sqlQuery != "":
		explainPlan, _, err = mongosql.TranslateSQLQuery(*sqlQuery, *dbName, *mongoVersion, *schema, *format, *explain, false)
	case *queryFile != "":
		explainPlan, _, err = mongosql.TranslateSQLQueryFile(*queryFile, *dbName, *mongoVersion, *schema, *format, *explain, false)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(explainPlan)
}
