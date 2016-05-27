package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/schema"

	"gopkg.in/mgo.v2/bson"
)

// mappingRegistry provides a way to get a field name from a table/column.
type mappingRegistry struct {
	columns []*mappedColumn
	fields  map[string]map[string]string
}

type mappedColumn struct {
	*Column
	alias string
}

func (mr *mappingRegistry) addColumn(column *Column, alias string) {
	if alias == "" {
		alias = column.Name
	}
	mr.columns = append(mr.columns, &mappedColumn{
		Column: column,
		alias:  alias,
	})
}

func (mr *mappingRegistry) copy() *mappingRegistry {
	newMappingRegistry := &mappingRegistry{}
	newMappingRegistry.columns = make([]*mappedColumn, len(mr.columns))
	copy(newMappingRegistry.columns, mr.columns)
	if mr.fields != nil {
		for tableName, columns := range mr.fields {
			for columnName, fieldName := range columns {
				newMappingRegistry.registerMapping(tableName, columnName, fieldName)
			}
		}
	}

	return newMappingRegistry
}

func (mr *mappingRegistry) registerMapping(tbl, column, field string) {

	if mr.fields == nil {
		mr.fields = make(map[string]map[string]string)
	}

	if _, ok := mr.fields[tbl]; !ok {
		mr.fields[tbl] = make(map[string]string)
	}

	mr.fields[tbl][column] = field
}

func (mr *mappingRegistry) lookupFieldName(tableName, columnName string) (string, bool) {
	if mr.fields == nil {
		return "", false
	}

	columnToField, ok := mr.fields[tableName]
	if !ok {
		return "", false
	}

	field, ok := columnToField[columnName]

	return field, ok
}

// MongoSource is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type MongoSourceStage struct {
	dbName          string
	tableName       string
	aliasName       string
	collectionName  string
	mappingRegistry *mappingRegistry
	pipeline        []bson.D
}

type MongoSourceIter struct {
	tableName       string
	aliasName       string
	mappingRegistry *mappingRegistry
	ctx             *ExecutionCtx
	iter            FindResults
	err             error
}

func NewMongoSourceStage(schema *schema.Schema, dbName, tableName string, aliasName string) (*MongoSourceStage, error) {

	if dbName == "" {
		return nil, fmt.Errorf("dbName is empty")
	}

	if tableName == "" {
		return nil, fmt.Errorf("tableName is empty")
	}

	ms := &MongoSourceStage{
		dbName:    dbName,
		tableName: tableName,
		aliasName: aliasName,
	}

	if ms.aliasName == "" {
		ms.aliasName = ms.tableName
	}

	database, ok := schema.Databases[ms.dbName]
	if !ok {
		return nil, fmt.Errorf("db %q doesn't exist - table %q", dbName, tableName)
	}
	tableSchema, ok := database.Tables[ms.tableName]
	if !ok {
		return nil, fmt.Errorf("table %q doesn't exist in db %q", tableName, dbName)
	}

	ms.collectionName = tableSchema.CollectionName

	ms.mappingRegistry = &mappingRegistry{}
	for _, c := range tableSchema.RawColumns {
		column := &Column{
			Table:     ms.aliasName,
			Name:      c.SqlName,
			SQLType:   c.SqlType,
			MongoType: c.MongoType,
		}
		ms.mappingRegistry.addColumn(column, "")
		ms.mappingRegistry.registerMapping(ms.aliasName, c.SqlName, c.Name)
	}

	ms.pipeline = tableSchema.Pipeline

	return ms, nil
}

func (ms *MongoSourceStage) clone() *MongoSourceStage {
	return &MongoSourceStage{
		dbName:          ms.dbName,
		tableName:       ms.tableName,
		aliasName:       ms.aliasName,
		collectionName:  ms.collectionName,
		mappingRegistry: ms.mappingRegistry,
		pipeline:        ms.pipeline,
	}
}

// Open establishes a connection to database collection for this table.
func (ms *MongoSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	mgoIter := MgoFindResults{ctx.Session().DB(ms.dbName).C(ms.collectionName).Pipe(ms.pipeline).AllowDiskUse().Iter()}

	return &MongoSourceIter{
		tableName:       ms.tableName,
		aliasName:       ms.aliasName,
		mappingRegistry: ms.mappingRegistry,
		ctx:             ctx,
		iter:            mgoIter,
		err:             nil}, nil
}

func (ms *MongoSourceIter) Next(row *Row) bool {
	if ms.iter == nil {
		return false
	}

	var hasNext bool

	d := &bson.D{}
	hasNext = ms.iter.Next(d)
	if !hasNext {
		return false
	}

	for _, column := range ms.mappingRegistry.columns {

		mappedFieldName, ok := ms.mappingRegistry.lookupFieldName(column.Table, column.alias)
		if !ok {
			ms.err = fmt.Errorf("Unable to find mapping from %v.%v to a field name.", column.Table, column.alias)
			return false
		}

		extractedField, _ := extractFieldByName(mappedFieldName, d.Map())

		value := Value{
			Table: ms.aliasName,
			Name:  column.Name,
		}

		value.Data, ms.err = NewSQLValue(extractedField, column.SQLType, column.MongoType)
		if ms.err != nil {
			return false
		}

		row.Data = append(row.Data, value)
	}

	return true
}

func (ms *MongoSourceStage) Columns() []*Column {
	var columns []*Column
	for _, c := range ms.mappingRegistry.columns {
		columns = append(columns, c.Column)
	}

	return columns
}

func (ms *MongoSourceIter) Close() error {
	return ms.iter.Close()
}

func (ms *MongoSourceIter) Err() error {
	if err := ms.iter.Err(); err != nil {
		return err
	}
	return ms.err
}
