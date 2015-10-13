package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

var (
	ISTablesHeaders = []string{
		"TABLE_SCHEMA",
		"TABLE_NAME",
		"TABLE_TYPE",
		"TABLE_COMMENT",
	}

	ISColumnHeaders = []string{
		"TABLE_CATALOG",
		"TABLE_SCHEMA",
		"TABLE_NAME",
		"COLUMN_NAME",
		"COLUMN_TYPE",
		"ORDINAL_POSITION",
	}
)

type ConfigDataSource struct {
	tableName      string
	includeColumns bool
	filter         interface{}
	matcher        Matcher
	iter           FindResults
	err            error
	ctx            *ExecutionCtx
}

func (cds *ConfigDataSource) Open(ctx *ExecutionCtx) error {
	cds.ctx = ctx
	return cds.init(ctx)
}

func (cds *ConfigDataSource) init(ctx *ExecutionCtx) error {

	switch cds.tableName {
	case "key_column_usage":
		cds.iter = &EmptyFindResults{}
		return nil
	case "columns":
		cds.includeColumns = true
	case "tables":
	default:
		return fmt.Errorf("unknown information_schema table (%s)", cds.tableName)
	}

	cds.iter = cds.Find().Iter()
	return nil
}

func (cds *ConfigDataSource) Next(row *Row) bool {
	if cds.iter == nil {
		return false
	}

	data := &bson.D{}
	hasNext := cds.iter.Next(data)
	row.Data = []TableRow{{cds.tableName, bsonDToValues(*data), nil}}

	if !hasNext {
		cds.err = cds.iter.Err()
	}

	return hasNext
}

func (cds *ConfigDataSource) OpFields() []*Column {

	var columns []*Column

	headers := ISTablesHeaders

	if cds.includeColumns {
		headers = ISColumnHeaders
	}

	for _, c := range headers {
		column := &Column{
			Table: cds.tableName,
			Name:  c,
			View:  c,
		}
		columns = append(columns, column)
	}

	return columns
}

func (cds *ConfigDataSource) Close() error {
	return cds.iter.Close()
}

func (cds *ConfigDataSource) Err() error {
	return cds.iter.Err()
}

func (cds *ConfigDataSource) Find() FindQuery {
	return ConfigFindQuery{cds.ctx, cds.matcher, cds.includeColumns}
}

func (cds *ConfigDataSource) Insert(docs ...interface{}) error {
	return fmt.Errorf("cannot insert into config data source")
}

func (cds *ConfigDataSource) DropCollection() error {
	return fmt.Errorf("cannot drop config data source")
}

func _cfrNextHelper(result *bson.D, fieldName string, fieldValue interface{}) {
	*result = append(*result, bson.DocElem{fieldName, fieldValue})
}

// -------

type ConfigFindResults struct {
	ctx            *ExecutionCtx
	matcher        Matcher
	includeColumns bool

	dbOffset      int
	tableOffset   int
	columnsOffset int

	err error
}

func (cfr *ConfigFindResults) Next(result *bson.D) bool {
	if cfr.err != nil {
		return false
	}

	// are we in valid db space
	if cfr.dbOffset >= len(cfr.ctx.Config.RawSchemas) {
		// nope, we're done
		return false
	}

	db := cfr.ctx.Config.RawSchemas[cfr.dbOffset]

	// are we in valid table space
	if cfr.tableOffset >= len(db.RawTables) {
		cfr.dbOffset = cfr.dbOffset + 1
		cfr.tableOffset = 0
		cfr.columnsOffset = 0
		return cfr.Next(result)
	}

	table := db.RawTables[cfr.tableOffset]

	*result = bson.D{}

	tableName := "columns"

	if !cfr.includeColumns {
		_cfrNextHelper(result, ISTablesHeaders[0], db.DB)
		_cfrNextHelper(result, ISTablesHeaders[1], table.Table)

		_cfrNextHelper(result, ISTablesHeaders[2], "BASE TABLE")
		_cfrNextHelper(result, ISTablesHeaders[3], "d")

		cfr.tableOffset = cfr.tableOffset + 1
		tableName = "tables"
	} else {
		// are we in valid column space
		if cfr.columnsOffset >= len(table.Columns) {
			cfr.tableOffset = cfr.tableOffset + 1
			cfr.columnsOffset = 0
			return cfr.Next(result)
		}

		_cfrNextHelper(result, ISColumnHeaders[0], "def")

		_cfrNextHelper(result, ISColumnHeaders[1], db.DB)
		_cfrNextHelper(result, ISColumnHeaders[2], table.Table)

		col := table.Columns[cfr.columnsOffset]

		_cfrNextHelper(result, ISColumnHeaders[3], col.Name)
		_cfrNextHelper(result, ISColumnHeaders[4], col.MysqlType)

		_cfrNextHelper(result, ISColumnHeaders[5], cfr.columnsOffset+1)

		cfr.columnsOffset = cfr.columnsOffset + 1
	}

	evalCtx := &EvalCtx{[]Row{{[]TableRow{{tableName, bsonDToValues(*result), nil}}}}, cfr.ctx}
	if cfr.matcher != nil {
		m, err := cfr.matcher.Matches(evalCtx)
		if err != nil {
			cfr.err = err
			return false
		}
		if !m {
			return cfr.Next(result)
		}
	}
	return true
}

// -------

type ConfigFindQuery struct {
	ctx            *ExecutionCtx
	matcher        Matcher
	includeColumns bool
}

func (cfq ConfigFindQuery) Iter() FindResults {
	return &ConfigFindResults{cfq.ctx, cfq.matcher, cfq.includeColumns, 0, 0, 0, nil}
}

func (cfr *ConfigFindResults) Err() error {
	return cfr.err
}

func (cfr *ConfigFindResults) Close() error {
	return nil
}
