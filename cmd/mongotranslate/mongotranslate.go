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
	defaultDB := flag.String("defaultDB", defaultDbName, "The default database name to use for unqualified tables in the query.")
	mdbVersion := flag.String("mdbVersion", defaultMdbVersion, "The MongoDB version to which to translate the query.")

	showInferredSchema := flag.Bool("showSchema", false, "When true, the inferred schema will be printed before the translation.")

	flag.Parse()

	if *sqlQuery == "" {
		log.Fatalln("no query provided")
		return
	}

	explainPlan, err := mongotranslate.TranslateSQLQuery(*sqlQuery, *defaultDB, *mdbVersion, *showInferredSchema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(explainPlan)
}
