package translator

import "fmt"

import "gopkg.in/mgo.v2/bson"

import "github.com/erh/mongo-sql-temp/config"

// ---

type ConfigFindResults struct {
	config *config.Config
	query interface{}

	dbOffset int
	tableOffset int
	columnsOffset int
}

func (cfr *ConfigFindResults) Next(result *bson.M) bool {

	// are we in valid db space
	if cfr.dbOffset >= len(cfr.config.RawSchemas) {
		// nope, we're done
		return false
	}

	db := cfr.config.RawSchemas[cfr.dbOffset]
	
	// are we in valid table space
	if cfr.tableOffset >= len(db.RawTables) {
		cfr.dbOffset = cfr.dbOffset + 1
		cfr.tableOffset = 0
		cfr.columnsOffset = 0
		return cfr.Next(result)
	}

	table := db.RawTables[cfr.tableOffset]
	
	// are we in valid column space
	if cfr.columnsOffset >= len(table.Columns) {
		cfr.tableOffset = cfr.tableOffset + 1
		cfr.columnsOffset = 0
		return cfr.Next(result)
	}

	*result = bson.M{}
	
	(*result)["TABLE_CATALOG"] = "def"

	(*result)["TABLE_SCHEMA"] = db.DB
	(*result)["TABLE_NAME"] = table.Table

	col := table.Columns[cfr.columnsOffset]
	
	(*result)["COLUMN_NAME"] = col.Name
	(*result)["COLUMN_TYPE"] = col.MysqlType

	cfr.columnsOffset = cfr.columnsOffset + 1
	return true
}

func (cfr *ConfigFindResults) Err() error {
	return nil
}

// -

type ConfigFindQuery struct {
	config *config.Config
	query interface{}
}

func (cfq ConfigFindQuery) Iter() FindResults {
	return &ConfigFindResults{cfq.config, cfq.query, 0, 0, 0}
}

// -

type ConfigDataSource struct {
	Config *config.Config
}

func (cds ConfigDataSource) Find(query interface{}) FindQuery {
	return ConfigFindQuery{cds.Config, query}
}

func (cds ConfigDataSource) Insert(docs ...interface{}) error {
	return fmt.Errorf("cannot insert into config data source")
}

func (cds ConfigDataSource) DropCollection() error {
	return fmt.Errorf("cannot drop config data source")
}
