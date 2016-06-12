package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"

	"gopkg.in/mgo.v2/bson"
)

// mappingRegistry provides a way to get a field name from a table/column.
type mappingRegistry struct {
	columns []*Column
	fields  map[string]map[string]string
}

func (mr *mappingRegistry) addColumn(column *Column) {
	mr.columns = append(mr.columns, column)
}

func (mr *mappingRegistry) copy() *mappingRegistry {
	newMappingRegistry := &mappingRegistry{}
	newMappingRegistry.columns = make([]*Column, len(mr.columns))
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

// MongoSourceStage is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type MongoSourceStage struct {
	selectIDs       []int
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

func NewMongoSourceStage(selectID int, schema *schema.Schema, dbName, tableName, aliasName string) (*MongoSourceStage, error) {

	if dbName == "" {
		return nil, fmt.Errorf("dbName is empty")
	}

	if tableName == "" {
		return nil, fmt.Errorf("tableName is empty")
	}

	ms := &MongoSourceStage{
		selectIDs: []int{selectID},
		dbName:    dbName,
		tableName: tableName,
		aliasName: aliasName,
	}

	if ms.aliasName == "" {
		ms.aliasName = ms.tableName
	}

	database, ok := schema.Databases[ms.dbName]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_TABLE_ERROR, dbName+"."+tableName)
	}
	tableSchema, ok := database.Tables[ms.tableName]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_TABLE_ERROR, dbName+"."+tableName)
	}

	ms.collectionName = tableSchema.CollectionName

	ms.mappingRegistry = &mappingRegistry{}
	for _, c := range tableSchema.RawColumns {
		column := &Column{
			SelectID:  selectID,
			Table:     ms.aliasName,
			Name:      c.SqlName,
			SQLType:   c.SqlType,
			MongoType: c.MongoType,
		}
		ms.mappingRegistry.addColumn(column)
		ms.mappingRegistry.registerMapping(ms.aliasName, c.SqlName, c.Name)
	}

	ms.pipeline = tableSchema.Pipeline

	return ms, nil
}

func (ms *MongoSourceStage) clone() *MongoSourceStage {
	return &MongoSourceStage{
		selectIDs:       ms.selectIDs,
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

	mappedD := d.Map()

	for _, column := range ms.mappingRegistry.columns {

		mappedFieldName, ok := ms.mappingRegistry.lookupFieldName(column.Table, column.Name)
		if !ok {
			ms.err = fmt.Errorf("Unable to find mapping from %v.%v to a field name.", column.Table, column.Name)
			return false
		}

		extractedField, _ := extractFieldByName(mappedFieldName, mappedD)

		value := Value{
			SelectID: column.SelectID,
			Table:    column.Table,
			Name:     column.Name,
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
	return ms.mappingRegistry.columns
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
