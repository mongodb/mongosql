package evaluator

import (
	"bytes"
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
)

// MongoSourceStage is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type MongoSourceStage struct {
	collation       *collation.Collation
	selectIDs       []int
	dbName          string
	tableNames      []string
	aliasNames      []string
	collectionNames []string
	tableType       catalog.TableType
	mappingRegistry *mappingRegistry
	pipeline        []bson.D
}

// NewMongoSourceStage creates a new MongoSourceStage from a catalog.MongoTable.
func NewMongoSourceStage(db *catalog.Database, table *catalog.MongoTable, selectID int, aliasName string) *MongoSourceStage {

	ms := &MongoSourceStage{
		selectIDs:  []int{selectID},
		dbName:     string(db.Name),
		tableNames: []string{string(table.Name())},
		aliasNames: []string{aliasName},
		tableType:  table.Type(),
	}

	if len(ms.aliasNames) == 0 || ms.aliasNames[0] == "" {
		ms.aliasNames = ms.tableNames
	}

	ms.collation = table.Collation()
	ms.collectionNames = []string{table.CollectionName}

	ms.mappingRegistry = &mappingRegistry{}

	primaryKeys := catalog.Columns(table.PrimaryKeys())

	for _, c := range table.Columns() {
		mc := c.(*catalog.MongoColumn)
		column := &Column{
			SelectID:   selectID,
			Table:      ms.aliasNames[0],
			Name:       string(mc.Name()),
			SQLType:    mc.Type(),
			MongoType:  mc.MongoType,
			PrimaryKey: primaryKeys.Contains(mc.Name()),
		}

		ms.mappingRegistry.addColumn(column)
		ms.mappingRegistry.registerMapping(ms.aliasNames[0], string(mc.Name()), string(mc.MongoName))
	}

	ms.pipeline = make([]bson.D, len(table.Pipeline))
	copy(ms.pipeline, table.Pipeline)

	return ms
}

func (ms *MongoSourceStage) clone() *MongoSourceStage {
	pipeline := []bson.D{}
	for _, stage := range ms.pipeline {
		pipeline = append(pipeline, stage)
	}
	return &MongoSourceStage{
		selectIDs:       ms.selectIDs,
		dbName:          ms.dbName,
		tableNames:      ms.tableNames,
		aliasNames:      ms.aliasNames,
		collectionNames: ms.collectionNames,
		collation:       ms.collation,
		mappingRegistry: ms.mappingRegistry,
		pipeline:        pipeline,
	}
}

func (ms *MongoSourceStage) isView() bool {
	return ms.tableType == catalog.View
}

// Open establishes a connection to database collection for this table.
func (ms *MongoSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	errChan := make(chan error, 1)

	var iter mongodb.Cursor
	var err error

	util.PanicSafeGo(func() {
		iter, err = ctx.Session().Aggregate(ms.dbName,
			ms.collectionNames[0], ms.pipeline)
		errChan <- err
	}, func(err interface{}) {
		ctx.Logger(log.NetworkComponent).Errf(log.Always,
			"data access MongoDB session closed: %v", err)
	})

	select {
	case <-ctx.Context().Done():
		return nil, ctx.Context().Err()
	case err = <-errChan:
	}

	return &MongoSourceIter{
		mappingRegistry: ms.mappingRegistry,
		ctx:             ctx.Context(),
		iter:            iter,
		err:             nil,
	}, err
}

type MongoSourceIter struct {
	mappingRegistry *mappingRegistry
	ctx             context.Context
	iter            mongodb.Cursor
	err             error
}

func (ms *MongoSourceIter) Next(row *Row) bool {

	document := &bson.D{}
	if !ms.iter.Next(ms.ctx, document) {
		return false
	}

	mappedDocument := document.Map()

	for _, column := range ms.mappingRegistry.columns {

		mappedFieldName, ok := ms.mappingRegistry.lookupFieldName(column.Table, column.Name)
		if !ok {
			ms.err = fmt.Errorf("unable to find mapping from %v.%v to a field name", column.Table, column.Name)
			return false
		}

		extractedField, _ := extractFieldByName(mappedFieldName, mappedDocument)

		value := Value{
			SelectID: column.SelectID,
			Table:    column.Table,
			Name:     column.Name,
		}

		value.Data, ms.err = NewSQLValueFromSQLColumnExpr(extractedField, column.SQLType, column.MongoType)
		if ms.err != nil {
			ms.err = fmt.Errorf("column '%v': %v", unsanitizeFieldName(mappedFieldName), ms.err)
			return false
		}

		row.Data = append(row.Data, value)
	}

	return true
}

func (ms *MongoSourceStage) Columns() []*Column {
	return ms.mappingRegistry.columns
}

func (ms *MongoSourceStage) Collation() *collation.Collation {
	return ms.collation
}

func (ms *MongoSourceIter) Close() error {
	return ms.iter.Close(ms.ctx)
}

func (ms *MongoSourceIter) Err() error {
	if err := ms.iter.Err(); err != nil {
		return err
	}
	return ms.err
}

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

func (mr *mappingRegistry) isPrimaryKey(name string) bool {
	for _, column := range mr.columns {
		if column.Name == name {
			return column.PrimaryKey
		}
	}
	return false
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

func (mr *mappingRegistry) registerMapping(tbl, column, field string) {

	if mr.fields == nil {
		mr.fields = make(map[string]map[string]string)
	}

	if _, ok := mr.fields[tbl]; !ok {
		mr.fields[tbl] = make(map[string]string)
	}

	mr.fields[tbl][column] = field
}

// containsFieldName checks whether a field name exists across the entire registry
func (mr *mappingRegistry) containsFieldName(fieldName string) bool {

	for _, columns := range mr.fields {

		for _, field := range columns {
			if field == fieldName {
				return true
			}
		}
	}

	return false
}

func (mr *mappingRegistry) String() string {
	var b bytes.Buffer

	for table, entry := range mr.fields {
		for column, name := range entry {
			b.WriteString(fmt.Sprintf("%v.%v => %v\n", table, column, name))
		}
	}

	return b.String()
}
