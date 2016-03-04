package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/schema"
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

type SchemaDataSource struct {
	tableName      string
	aliasName      string
	includeColumns bool
	matcher        SQLExpr
	iter           FindResults
	err            error
	ctx            *ExecutionCtx
}

func (sds *SchemaDataSource) Open(ctx *ExecutionCtx) error {
	sds.ctx = ctx
	return sds.init(ctx)
}

func (sds *SchemaDataSource) init(ctx *ExecutionCtx) error {

	switch sds.tableName {
	case "key_column_usage":
		sds.iter = &EmptyFindResults{}
		return nil
	case "columns":
		sds.includeColumns = true
	case "tables":
	default:
		return fmt.Errorf("unknown information_schema table (%s)", sds.tableName)
	}

	if sds.aliasName == "" {
		sds.aliasName = sds.tableName
	}

	sds.iter = sds.Find().Iter()
	return nil
}

func (sds *SchemaDataSource) Next(row *Row) bool {
	if sds.iter == nil {
		return false
	}

	data := &bson.D{}
	hasNext := sds.iter.Next(data)
	values, err := bsonDToValues(*data)
	if err != nil {
		sds.err = err
		return false
	}
	row.Data = TableRows{{sds.tableName, values}}

	if !hasNext {
		sds.err = sds.iter.Err()
	}

	return hasNext
}

func (sds *SchemaDataSource) OpFields() []*Column {

	var columns []*Column

	headers := ISTablesHeaders

	if sds.includeColumns {
		headers = ISColumnHeaders
	}

	for _, c := range headers {
		column := &Column{
			Table:     sds.aliasName,
			Name:      c,
			View:      c,
			SQLType:   schema.SQLVarchar,
			MongoType: schema.MongoString,
		}
		columns = append(columns, column)
	}

	return columns
}

func (sds *SchemaDataSource) Close() error {
	return sds.iter.Close()
}

func (sds *SchemaDataSource) Err() error {
	return sds.iter.Err()
}

func (sds *SchemaDataSource) Find() FindQuery {
	return SchemaFindQuery{sds.ctx, sds.matcher, sds.includeColumns}
}

func (sds *SchemaDataSource) Insert(docs ...interface{}) error {
	return fmt.Errorf("cannot insert into config data source")
}

func (sds *SchemaDataSource) DropCollection() error {
	return fmt.Errorf("cannot drop config data source")
}

func _cfrNextHelper(result *bson.D, fieldName string, fieldValue interface{}) {
	*result = append(*result, bson.DocElem{fieldName, fieldValue})
}

// -------

type SchemaFindResults struct {
	ctx            *ExecutionCtx
	matcher        SQLExpr
	includeColumns bool

	dbOffset      int
	tableOffset   int
	columnsOffset int

	err error
}

func (sfr *SchemaFindResults) Next(result *bson.D) bool {
	if sfr.err != nil {
		return false
	}

	// are we in valid db space
	if sfr.dbOffset >= len(sfr.ctx.Schema.RawDatabases) {
		// nope, we're done
		return false
	}

	db := sfr.ctx.Schema.RawDatabases[sfr.dbOffset]

	// are we in valid table space
	if sfr.tableOffset >= len(db.RawTables) {
		sfr.dbOffset = sfr.dbOffset + 1
		sfr.tableOffset = 0
		sfr.columnsOffset = 0
		return sfr.Next(result)
	}

	table := db.RawTables[sfr.tableOffset]

	*result = bson.D{}

	tableName := "columns"

	if !sfr.includeColumns {
		_cfrNextHelper(result, ISTablesHeaders[0], db.Name)
		_cfrNextHelper(result, ISTablesHeaders[1], table.Name)

		_cfrNextHelper(result, ISTablesHeaders[2], "BASE TABLE")
		_cfrNextHelper(result, ISTablesHeaders[3], "d")

		sfr.tableOffset = sfr.tableOffset + 1
		tableName = "tables"
	} else {
		// are we in valid column space
		if sfr.columnsOffset >= len(table.RawColumns) {
			sfr.tableOffset = sfr.tableOffset + 1
			sfr.columnsOffset = 0
			return sfr.Next(result)
		}

		_cfrNextHelper(result, ISColumnHeaders[0], "def")

		_cfrNextHelper(result, ISColumnHeaders[1], db.Name)
		_cfrNextHelper(result, ISColumnHeaders[2], table.Name)

		col := table.RawColumns[sfr.columnsOffset]

		_cfrNextHelper(result, ISColumnHeaders[3], col.SqlName)
		_cfrNextHelper(result, ISColumnHeaders[4], string(col.SqlType))

		_cfrNextHelper(result, ISColumnHeaders[5], sfr.columnsOffset+1)

		sfr.columnsOffset = sfr.columnsOffset + 1
	}

	values, err := bsonDToValues(*result)
	if err != nil {
		sfr.err = err
		return false
	}
	evalCtx := &EvalCtx{Rows{{TableRows{{tableName, values}}}}, sfr.ctx}
	if sfr.matcher != nil {
		m, err := Matches(sfr.matcher, evalCtx)
		if err != nil {
			sfr.err = err
			return false
		}
		if !m {
			return sfr.Next(result)
		}
	}
	return true
}

type SchemaFindQuery struct {
	ctx            *ExecutionCtx
	matcher        SQLExpr
	includeColumns bool
}

func (cfq SchemaFindQuery) Iter() FindResults {
	return &SchemaFindResults{cfq.ctx, cfq.matcher, cfq.includeColumns, 0, 0, 0, nil}
}

func (sfr *SchemaFindResults) Err() error {
	return sfr.err
}

func (sfr *SchemaFindResults) Close() error {
	return nil
}
