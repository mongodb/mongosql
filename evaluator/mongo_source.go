package evaluator

import (
	"fmt"
	"strings"

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

func (mr *mappingRegistry) lookupFieldName(tbl, column string) (string, bool) {
	if mr.fields == nil {
		return "", false
	}

	columnToField, ok := mr.fields[tbl]
	if !ok {
		return "", false
	}

	field, ok := columnToField[column]
	return field, ok
}

// MongoSource is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type MongoSource struct {
	dbName          string
	tableName       string
	aliasName       string
	fqns            string // the fully qualified namespace in MongoDB
	mappingRegistry *mappingRegistry
	pipeline        []bson.D
	matcher         SQLExpr
	ctx             *ExecutionCtx
	iter            FindResults
	err             error
}

func NewMongoSource(ctx *ExecutionCtx, dbName, tableName string, aliasName string) (*MongoSource, error) {

	if dbName == "" {
		return nil, fmt.Errorf("dbName is empty")
	}
	if tableName == "" {
		return nil, fmt.Errorf("tableName is empty")
	}
	ms := &MongoSource{
		dbName:    dbName,
		tableName: tableName,
		aliasName: aliasName,
	}

	if ms.aliasName == "" {
		ms.aliasName = ms.tableName
	}

	database, ok := ctx.Schema.Databases[ms.dbName]
	if !ok {
		return nil, fmt.Errorf("db (%s) doesn't exist - table (%s)", dbName, tableName)
	}
	tableSchema, ok := database.Tables[ms.tableName]
	if !ok {
		return nil, fmt.Errorf("table (%s) doesn't exist in db (%s)", tableName, dbName)
	}

	ms.fqns = tableSchema.FQNS

	ms.mappingRegistry = &mappingRegistry{}
	for _, c := range tableSchema.RawColumns {
		column := &Column{
			Table: ms.aliasName,
			Name:  c.SqlName,
			View:  c.SqlName,
			Type:  c.SqlType,
		}
		ms.mappingRegistry.addColumn(column)
		ms.mappingRegistry.registerMapping(ms.aliasName, c.SqlName, c.Name)
	}

	ms.pipeline = []bson.D{}

	return ms, nil
}

// WithPipeline creates a new TableScan operator by copying everything
// and changing only the pipeline.
func (ms *MongoSource) WithPipeline(pipeline []bson.D) *MongoSource {
	return &MongoSource{
		dbName:          ms.dbName,
		tableName:       ms.tableName,
		aliasName:       ms.aliasName,
		fqns:            ms.fqns,
		matcher:         ms.matcher,
		mappingRegistry: ms.mappingRegistry,
		pipeline:        pipeline,
	}
}

// WithMappingRegistry creates a new TableScan operator by copying everything
// and changing only the mappingRegistry.
func (ms *MongoSource) WithMappingRegistry(mappingRegistry *mappingRegistry) *MongoSource {

	return &MongoSource{
		dbName:          ms.dbName,
		tableName:       ms.tableName,
		aliasName:       ms.aliasName,
		fqns:            ms.fqns,
		matcher:         ms.matcher,
		pipeline:        ms.pipeline,
		mappingRegistry: mappingRegistry,
	}
}

// Open establishes a connection to database collection for this table.
func (ms *MongoSource) Open(ctx *ExecutionCtx) error {
	ms.ctx = ctx

	pcs := strings.SplitN(ms.fqns, ".", 2)

	ms.iter = MgoFindResults{ctx.Session.DB(pcs[0]).C(pcs[1]).Pipe(ms.pipeline).Iter()}

	return nil
}

func (ms *MongoSource) Next(row *Row) bool {
	if ms.iter == nil {
		return false
	}

	var hasNext bool

	for {
		d := &bson.D{}
		hasNext = ms.iter.Next(d)

		if !hasNext {
			break
		}

		values := make(map[string]Values)
		data := d.Map()

		var err error

		for _, column := range ms.mappingRegistry.columns {

			mappedFieldName, ok := ms.mappingRegistry.lookupFieldName(column.Table, column.Name)
			if !ok {
				ms.err = fmt.Errorf("Unable to find mapping from %v.%v to a field name.", column.Table, column.Name)
				return false
			}

			value := Value{
				Name: column.Name,
				View: column.View,
				Data: extractFieldByName(mappedFieldName, data),
			}

			value.Data, err = NewSQLValue(value.Data, column.Type)
			if err != nil {
				ms.err = err
				return false
			}

			tableName := column.Table
			if tableName == ms.tableName {
				tableName = ms.aliasName
			}

			if _, ok := values[tableName]; !ok {
				values[tableName] = Values{}
			}

			values[tableName] = append(values[tableName], value)
			delete(data, mappedFieldName)
		}

		tableRows := TableRows{}
		for k, v := range values {
			tableRows = append(tableRows, TableRow{k, v})
		}

		row.Data = tableRows

		evalCtx := &EvalCtx{Rows{*row}, ms.ctx}

		if ms.matcher != nil {
			m, err := Matches(ms.matcher, evalCtx)
			if err != nil {
				ms.err = err
				return false
			}
			if m {
				break
			}
		} else {
			break
		}
	}

	return hasNext
}

func (ms *MongoSource) OpFields() (columns []*Column) {
	return ms.mappingRegistry.columns
}

func (ms *MongoSource) Close() error {
	return ms.iter.Close()
}

func (ms *MongoSource) Err() error {
	if err := ms.iter.Err(); err != nil {
		return err
	}
	return ms.err
}
