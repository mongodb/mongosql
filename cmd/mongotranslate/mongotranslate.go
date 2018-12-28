package main

import (
	"context"
	"log"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/variable"
	sqlLgr "github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
)

var (
	namespaces = []string{
		"foo$a$b$c",
		"bar$c$d$e",
	}
)

const (
	defaultDbName = "test"
	mdbVersion    = "4.0.0"
	sql           = "explain select * from foo t1 join foo t2 on t1.a = t2.b"
)

func main() {
	ctlg, err := getTestCatalog(mdbVersion)
	if err != nil {
		log.Fatalf("fatal error creating catalog: %v", err)
	}

	qCfg := evaluator.NewDefaultQueryConfig(mdbVersion, ctlg)

	res, err := evaluator.ExecuteSQL(context.Background(), qCfg, sql)
	if err != nil {
		log.Fatalf("fatal error executing sql %q: %v", sql, err)
	}

	if len(res.Stats.Explain) != 1 {
		log.Fatalf("query not fully pushed down: %v", res.Stats.Explain[0].PushdownFailures[0])
	}

	log.Printf("explain plan is: \n%v", res.Stats.Explain[0].Pipeline.Else(""))
}

func getTestCatalog(mdbVersion string) (catalog.Catalog, error) {
	gbl := variable.NewGlobalContainer(nil)
	gbl.MongoDBVersion = mdbVersion
	gbl.PolymorphicTypeConversionMode = string(variable.PolymorphicTypeConversionModeOff)
	gbl.SetSystemVariable(variable.TypeConversionMode, string(variable.MySQLTypeConversionMode))

	vars := variable.NewSessionContainer(gbl)
	vars.MongoDBVersion = mdbVersion
	vars.PolymorphicTypeConversionMode = string(variable.PolymorphicTypeConversionModeOff)
	vars.SetSystemVariable(variable.TypeConversionMode, string(variable.MySQLTypeConversionMode))

	ctlg := catalog.New("", vars)

	db, err := ctlg.AddDatabase(defaultDbName)
	if err != nil {
		return nil, err
	}

	lgr := sqlLgr.GlobalLogger()

	for _, ns := range namespaces {
		split := strings.SplitN(ns, "$", 2)
		tableName, columnNames := split[0], split[1]

		table, err := schema.NewTable(lgr, tableName, tableName, nil, nil)
		if err != nil {
			return nil, err
		}

		columns := strings.Split(columnNames, "$")
		for _, columnName := range columns {
			column := schema.NewColumn(columnName, schema.SQLInt, columnName, schema.MongoInt)
			table.AddColumn(lgr, column, false)
		}

		err = db.AddTable(catalog.NewMongoTable(table, catalog.BaseTable, collation.Default))
		if err != nil {
			return nil, err
		}
	}

	return ctlg, nil
}
