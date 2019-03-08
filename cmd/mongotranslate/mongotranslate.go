package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/10gen/sqlproxy/internal/mongotranslate"
)

const (
	defaultDbName     = "test"
	defaultMdbVersion = "4.0.0"
)

func main() {
	sqlQuery := flag.String("query", "", "The sql query to translate into MongoDB aggregation language.")
	schema := flag.String("schema", "", "The path to a DRDL file or a directory containing DRDL files.")
	defaultDB := flag.String("defaultDB", defaultDbName, "The default database name to use for unqualified tables in the query.")
	mdbVersion := flag.String("mdbVersion", defaultMdbVersion, "The MongoDB version to which to translate the query.")

	flag.Parse()

	if *sqlQuery == "" {
		log.Fatalln("no query provided")
		return
	}
	if *schema == "" {
		log.Fatalln("no schema provided")
		return
	}

	explainPlan, err := mongotranslate.TranslateSQLQuery(*sqlQuery, *defaultDB, *mdbVersion, *schema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(explainPlan)
}
