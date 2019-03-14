package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/10gen/sqlproxy/internal/mongotranslate"
)

const (
	defaultDbName       = "test"
	defaultMongoVersion = "4.0.0"
	defaultFormat       = "multiline"
)

func main() {
	sqlQuery := flag.String("query", "", "The sql query to translate into MongoDB aggregation language.")
	schema := flag.String("schema", "", "The path to a DRDL file or a directory containing DRDL files.")
	dbName := flag.String("dbName", defaultDbName, "The database name to use for unqualified tables in the query.")
	mongoVersion := flag.String("mongoVersion", defaultMongoVersion, "The MongoDB version to which to translate the query.")
	format := flag.String("format", defaultFormat, `The desired formatting for the output. The flag can be set to "multiline" (formats output with one pipeline stage per line) or "none" (no formatting at all).`)

	flag.Parse()

	if *sqlQuery == "" {
		log.Fatalln("no query provided")
		return
	}
	if *schema == "" {
		log.Fatalln("no schema provided")
		return
	}
	if *format != "multiline" && *format != "none" {
		log.Fatalf("invalid value `%v` for option `--format`. Allowed values are: multiline, none.\n", *format)
		return
	}

	explainPlan, err := mongotranslate.TranslateSQLQuery(*sqlQuery, *dbName, *mongoVersion, *schema, *format)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(explainPlan)
}
