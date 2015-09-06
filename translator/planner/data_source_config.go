package planner

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2/bson"
)

type ConfigFindResults struct {
	config         *config.Config
	query          interface{}
	includeColumns bool

	dbOffset      int
	tableOffset   int
	columnsOffset int

	err error
}

func _cfrNextHelper(result *bson.D, fieldName string, fieldValue interface{}) {
	*result = append(*result, bson.DocElem{fieldName, fieldValue})
}

func (cfr *ConfigFindResults) Next(result *bson.D) bool {
	if cfr.err != nil {
		return false
	}
	
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

	*result = bson.D{}

	if !cfr.includeColumns {
		_cfrNextHelper(result, "TABLE_SCHEMA", db.DB)
		_cfrNextHelper(result, "TABLE_NAME", table.Table)

		_cfrNextHelper(result, "TABLE_TYPE", "BASE TABLE")
		_cfrNextHelper(result, "TABLE_COMMENT", "d")

		cfr.tableOffset = cfr.tableOffset + 1
	} else {
		// are we in valid column space
		if cfr.columnsOffset >= len(table.Columns) {
			cfr.tableOffset = cfr.tableOffset + 1
			cfr.columnsOffset = 0
			return cfr.Next(result)
		}

		_cfrNextHelper(result, "TABLE_CATALOG", "def")

		_cfrNextHelper(result, "TABLE_SCHEMA", db.DB)
		_cfrNextHelper(result, "TABLE_NAME", table.Table)

		col := table.Columns[cfr.columnsOffset]

		_cfrNextHelper(result, "COLUMN_NAME", col.Name)
		_cfrNextHelper(result, "COLUMN_TYPE", col.MysqlType)

		_cfrNextHelper(result, "ORDINAL_POSITION", cfr.columnsOffset + 1)

		cfr.columnsOffset = cfr.columnsOffset + 1
	}

	matches, err := Matches(cfr.query, result)
	if err != nil {
		cfr.err = err
		return false
	}
	if !matches {
		return cfr.Next(result)
	}
	
	return true
}

func (cfr *ConfigFindResults) Err() error {
	return cfr.err
}

func (cfr *ConfigFindResults) Close() error {
	return nil
}

// -

type ConfigFindQuery struct {
	config         *config.Config
	query          interface{}
	includeColumns bool
}

func (cfq ConfigFindQuery) Iter() FindResults {
	return &ConfigFindResults{cfq.config, cfq.query, cfq.includeColumns, 0, 0, 0, nil}
}

// -

type ConfigDataSource struct {
	Config         *config.Config
	IncludeColumns bool
}

func (cds ConfigDataSource) Find(query interface{}) FindQuery {
	return ConfigFindQuery{cds.Config, query, cds.IncludeColumns}
}

func (cds ConfigDataSource) Insert(docs ...interface{}) error {
	return fmt.Errorf("cannot insert into config data source")
}

func (cds ConfigDataSource) DropCollection() error {
	return fmt.Errorf("cannot drop config data source")
}

func (cds ConfigDataSource) GetColumns() []config.Column {
	return []config.Column{}
}
