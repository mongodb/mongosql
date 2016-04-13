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

type SchemaDataSourceStage struct {
	tableName string
	aliasName string
	matcher   SQLExpr
}

type SchemaDataSourceIter struct {
	plan    *SchemaDataSourceStage
	iter    *SchemaFindResults
	err     error
	execCtx *ExecutionCtx
}

func (sds *SchemaDataSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	if sds.tableName == "key_column_usage" {
		return &EmptyIter{}, nil
	}

	it := &SchemaDataSourceIter{
		plan: sds,
		iter: &SchemaFindResults{execCtx: ctx},
	}
	switch sds.tableName {
	case "columns":
		it.iter.includeColumns = true
	case "tables":
		it.iter.includeColumns = false
	default:
		return nil, fmt.Errorf("unknown information_schema table (%s)", sds.tableName)
	}

	return it, nil
}

func (sds *SchemaDataSourceIter) Next(row *Row) bool {
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
	row.Data = TableRows{{sds.plan.tableName, values}}

	if !hasNext {
		sds.err = sds.iter.Err()
	}

	return hasNext
}

func (sds *SchemaDataSourceStage) OpFields() []*Column {

	var columns []*Column

	headers := ISTablesHeaders

	if sds.tableName == "columns" {
		headers = ISColumnHeaders
	}
	aliasName := sds.aliasName
	if aliasName == "" {
		aliasName = sds.tableName
	}

	for _, c := range headers {
		column := &Column{
			Table:     aliasName,
			Name:      c,
			View:      c,
			SQLType:   schema.SQLVarchar,
			MongoType: schema.MongoString,
		}
		columns = append(columns, column)
	}

	return columns
}

func (sds *SchemaDataSourceIter) Close() error {
	return sds.iter.Close()
}

func (sds *SchemaDataSourceIter) Err() error {
	return sds.iter.Err()
}

// func (sds *SchemaDataSource) Insert(docs ...interface{}) error {
// 	return fmt.Errorf("cannot insert into config data source")
// }
//
// func (sds *SchemaDataSource) DropCollection() error {
// 	return fmt.Errorf("cannot drop config data source")
// }

func _cfrNextHelper(result *bson.D, fieldName string, fieldValue interface{}) {
	*result = append(*result, bson.DocElem{fieldName, fieldValue})
}

// -------

type SchemaFindResults struct {
	execCtx        *ExecutionCtx
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
	if sfr.dbOffset >= len(sfr.execCtx.PlanCtx.Schema.RawDatabases) {
		// nope, we're done
		return false
	}

	db := sfr.execCtx.PlanCtx.Schema.RawDatabases[sfr.dbOffset]

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
		_cfrNextHelper(result, ISTablesHeaders[3], " ")

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
	evalCtx := &EvalCtx{Rows{{TableRows{{tableName, values}}}}, sfr.execCtx}
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

func (sfr *SchemaFindResults) Err() error {
	return sfr.err
}

func (sfr *SchemaFindResults) Close() error {
	return nil
}
