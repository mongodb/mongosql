package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

const (
	// maxColumnBucketSize is maximum size of the bitmask
	// used to track - per row gotten from MongoDB
	// - which columns contained in the mapping
	// registry were returned from the database.
	maxColumnBucketSize = 63
)

// MongoSourceStage is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type MongoSourceStage struct {
	collation           *collation.Collation
	selectIDs           []int
	dbName              string
	tableNames          []string
	aliasNames          []string
	collectionNames     []string
	isShardedCollection map[string]bool
	tableType           catalog.TableType
	mappingRegistry     *mappingRegistry
	pipeline            []bson.D
}

// columnPosition associates a Column with its proper return position.
type columnPosition struct {
	column *Column
	index  int
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
	ms.isShardedCollection = map[string]bool{table.CollectionName: table.IsSharded()}
	ms.mappingRegistry = &mappingRegistry{}

	primaryKeys := catalog.Columns(table.PrimaryKeys())

	for _, c := range table.Columns() {
		mc := c.(*catalog.MongoColumn)
		column := NewColumn(selectID, ms.aliasNames[0], ms.tableNames[0], ms.dbName, string(mc.Name()),
			string(mc.Name()), "", mc.Type(), mc.MongoType, primaryKeys.Contains(mc.Name()))
		ms.mappingRegistry.addColumn(column)
		ms.mappingRegistry.registerMapping(ms.dbName, ms.aliasNames[0], string(mc.Name()), string(mc.MongoName))
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
		selectIDs:           ms.selectIDs,
		dbName:              ms.dbName,
		tableNames:          ms.tableNames,
		aliasNames:          ms.aliasNames,
		collectionNames:     ms.collectionNames,
		isShardedCollection: ms.isShardedCollection,
		collation:           ms.collation,
		mappingRegistry:     ms.mappingRegistry,
		pipeline:            pipeline,
	}
}

func (ms *MongoSourceStage) isView() bool {
	return ms.tableType == catalog.View
}

// getAggregationCursor get a cursor over MongoDB results from an Aggregation Pipeline.
func (ms *MongoSourceStage) getAggregationCursor(ctx *ExecutionCtx) (mongodb.Cursor, error) {
	errChan := make(chan error, 1)

	var iter mongodb.Cursor
	var err error

	util.PanicSafeGo(func() {
		iter, err = ctx.Session().Aggregate(ms.dbName,
			ms.collectionNames[0], ms.pipeline)
		errChan <- err
	}, func(err interface{}) {
		ctx.Logger(log.NetworkComponent).Errf(log.Admin,
			"MongoDB data access session closed: %v", err)
	})

	select {
	case <-ctx.Context().Done():
		return nil, ctx.Context().Err()
	case err = <-errChan:
	}

	return iter, err
}

// ColumnInfo keeps track of the data needed to correctly deserialize data from a MongoSourceStage.
type ColumnInfo struct {
	// Field is the name of the specific MongODB field.
	Field string
	// Type is the byte corresponding to the type
	// MongoDRDL specifies for the given column. The byte corresponds to the BSON kind byte, iff
	// the column type is a BSON type. Some Column types are not BSON types: e.g., Date, which needs
	// to drop the Time portions of a Timestamp for formatting purposes because BSON datetime objects
	// store both the date and the time.
	Type schema.BSONSpecType
	// UUIDSubtype is needed to handle UUIDs written by the Java and CSharp drivers, which store
	// UUIDs using different byte orders.
	UUIDSubtype schema.BSONSpecType
}

// FastMongoSourceIter implements FastIter. It is an Iterator over raw BSON Documents.
type FastMongoSourceIter struct {
	// ctx is the used to listen for any cancellation signals.
	ctx context.Context
	// iter is an implementation for getting data directly
	// from MongoDB.
	iter mongodb.Cursor
	// columnFields is a slice representing the field names,
	// in order, expected in the returned document.
	columnInfo []ColumnInfo
	// err holds any error that may occur during iteration.
	err error
}

// FastOpen opens a more optimized Iter over raw BSON documents returned from
// MongoDB in cases where no in-memory evaluation is needed to handle a query.
func (ms *MongoSourceStage) FastOpen(ctx *ExecutionCtx) (FastIter, error) {

	columns := ms.mappingRegistry.columns
	lenColumns := len(columns)
	uniqueFields := make(map[string]struct{}, lenColumns)

	columnInfo := make([]ColumnInfo, len(columns))
	for i, c := range columns {
		if c.MappingRegistryName == "" {
			c.MappingRegistryName = c.Name
		}

		mappedFieldName, ok := ms.mappingRegistry.lookupFieldName(c.Database,
			c.Table, c.MappingRegistryName)
		if !ok {
			return nil, fmt.Errorf("unable to find mapping from %v.%v.%v to "+
				"a field name %v", c.Database, c.Table,
				c.MappingRegistryName, ms.mappingRegistry.String())
		}

		if _, ok := uniqueFields[mappedFieldName]; ok {
			continue
		}
		uniqueFields[mappedFieldName] = struct{}{}
		uuidSubType := schema.BSONNone
		if c.MongoType == schema.MongoUUIDJava {
			uuidSubType = schema.BSONJavaUUID
		} else if c.MongoType == schema.MongoUUIDCSharp {
			uuidSubType = schema.BSONCSharpUUID
		}
		columnInfo[i] = ColumnInfo{Field: mappedFieldName, Type: schema.SQLTypeToBSONType[c.SQLType], UUIDSubtype: uuidSubType}
	}

	iter, err := ms.getAggregationCursor(ctx)
	if err != nil {
		return nil, err
	}

	return &FastMongoSourceIter{
		ctx:        ctx.Context(),
		iter:       iter,
		columnInfo: columnInfo,
		err:        nil,
	}, err
}

func (ms *FastMongoSourceIter) Next(doc *bson.RawD) bool {
	if !ms.iter.Next(ms.ctx, doc) {
		return false
	}

	return true
}

func (ms *FastMongoSourceIter) GetColumnInfo() []ColumnInfo {
	return ms.columnInfo
}

func (ms *FastMongoSourceIter) Err() error {
	if err := ms.iter.Err(); err != nil {
		return err
	}
	return ms.err
}

func (ms *FastMongoSourceIter) Close() error {
	return ms.iter.Close(ms.ctx)
}

// Open creates an Iter over rows returned from MongoDB.
func (ms *MongoSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	columns := ms.mappingRegistry.columns
	lenColumns := len(columns)

	numColumnMaskBuckets := (lenColumns / maxColumnBucketSize) + 1
	columnMaskBuckets := make([]uint64, numColumnMaskBuckets)
	lastColumnMaskBucket := (uint64(1) << uint64(lenColumns%maxColumnBucketSize)) - 1
	columnMaskBuckets[len(columnMaskBuckets)-1] = lastColumnMaskBucket

	for i := 0; i < len(columnMaskBuckets)-1; i++ {
		columnMaskBuckets[i] = (uint64(1) << maxColumnBucketSize) - 1
	}

	columnPositions := make(map[string]columnPosition, lenColumns)

	for i, c := range columns {
		if c.MappingRegistryName == "" {
			c.MappingRegistryName = c.Name
		}

		mappedFieldName, ok := ms.mappingRegistry.lookupFieldName(c.Database,
			c.Table, c.MappingRegistryName)
		if !ok {
			return nil, fmt.Errorf("unable to find mapping from %v.%v.%v to "+
				"a field name %v", c.Database, c.Table,
				c.MappingRegistryName, ms.mappingRegistry.String())
		}

		// future proofing situations where a field is redundantly mapped
		if _, ok := columnPositions[mappedFieldName]; ok {
			continue
		}

		columnPositions[mappedFieldName] = columnPosition{c, i}
	}

	iter, err := ms.getAggregationCursor(ctx)
	if err != nil {
		return nil, err
	}

	return &MongoSourceIter{
		mappingRegistry:   ms.mappingRegistry,
		ctx:               ctx.Context(),
		iter:              iter,
		columnPositions:   columnPositions,
		columnMaskBuckets: columnMaskBuckets,
		err:               nil,
	}, err
}

type MongoSourceIter struct {
	// mappingRegistry holds all columns that must be returned
	// from data gotten in the iterator.
	mappingRegistry *mappingRegistry
	// ctx is the used to listen for any cancellation signals.
	ctx context.Context
	// iter is an implementation forgetting data directly
	// from MongoDB.
	iter mongodb.Cursor
	// columnMaskBuckets is an expandable bit vector that is used
	// - per row - to track how many of the registry columns
	// were returned in the BSON document gotten from the
	// iterator.
	columnMaskBuckets []uint64
	// columnPositions maps the mapped field name for a
	// registry column to its ordinal position in the
	// returned table.
	columnPositions map[string]columnPosition
	// err holds any error that may occur during iteration.
	err error
}

func (ms *MongoSourceIter) Next(row *Row) bool {
	document := &bson.D{}
	if !ms.iter.Next(ms.ctx, document) {
		return false
	}

	row.Data = make([]Value, len(ms.mappingRegistry.columns))

	ms.mapDocumentToValues(row, *document)
	if ms.err != nil {
		return false
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
	if ms.ctx.Err() == nil {
		return ms.iter.Close(context.Background())
	}
	return nil
}

func (ms *MongoSourceIter) Err() error {
	if err := ms.iter.Err(); err != nil {
		return err
	}
	return ms.err
}

// mapDocumentToValues recursively traverses document and fills the
// values map with the keys and values it finds within the document.
func (ms *MongoSourceIter) mapDocumentToValues(row *Row, document interface{}) {
	// ensure all bit positions are set
	lenColumns := len(ms.mappingRegistry.columns)
	lastBucketMask := (uint64(1) << uint64(lenColumns%maxColumnBucketSize)) - 1
	ms.columnMaskBuckets[len(ms.columnMaskBuckets)-1] = lastBucketMask
	for i := 0; i < len(ms.columnMaskBuckets)-1; i++ {
		ms.columnMaskBuckets[i] = (uint64(1) << maxColumnBucketSize) - 1
	}

	ms.mapDocumentToValuesHelper([]string{}, row, document)

	// for any unset column key position, set the
	// value to be SQLNull
	for i := 0; i < len(ms.columnMaskBuckets); i++ {
		maskValue := ms.columnMaskBuckets[i]
		for j := 0; j < maxColumnBucketSize && maskValue != uint64(0); j++ {
			// check if bit position is unset; if not, unset it
			if ((maskValue >> uint64(j)) & 1) != 0 {
				idx := (maxColumnBucketSize * i) + j
				c := ms.mappingRegistry.columns[idx]
				row.Data[idx] = NewValue(c.SelectID, c.Database, c.Table, c.Name, SQLNull)
				maskValue &= ^(uint64(1) << uint64(j))
			}
		}
	}
}

func (ms *MongoSourceIter) mapDocumentToValuesHelper(frontier []string,
	row *Row, document interface{}) {
	if ms.err != nil {
		return
	}

	docValue := reflect.ValueOf(document)
	if !docValue.IsValid() {
		return
	}

	switch docValue.Type().Kind() {
	case reflect.Map:
		mapVal := docValue.MapIndex(reflect.ValueOf(docValue))
		if mapVal.Kind() == reflect.Invalid {
			return
		}
		for _, key := range mapVal.MapKeys() {
			frontier = append(frontier, fmt.Sprintf("%v", key))
			ms.mapDocumentToValuesHelper(frontier, row, mapVal.MapIndex(key).Interface())
			frontier = frontier[0 : len(frontier)-1]
		}
	case reflect.Slice:
		switch docValue.Type() {
		case bsonDType:
			bsonD := document.(bson.D)
			for _, d := range bsonD {
				frontier = append(frontier, d.Name)
				ms.mapDocumentToValuesHelper(frontier, row, d.Value)
				frontier = frontier[0 : len(frontier)-1]
			}
		default:
			// handle geo2d index fields
			for i := 0; i < docValue.Len(); i++ {
				frontier = append(frontier, fmt.Sprintf("%v", i))
				ms.mapDocumentToValuesHelper(frontier, row,
					docValue.Index(i).Interface())
				frontier = frontier[0 : len(frontier)-1]
			}
		}
	default:
		key := strings.Join(frontier, ".")
		if entry, ok := ms.columnPositions[key]; ok {
			c := entry.column
			value := NewValue(c.SelectID, c.Database,
				c.Table, c.Name, nil)
			value.Data, ms.err = NewSQLValueFromSQLColumnExpr(document,
				c.SQLType, c.MongoType)
			if ms.err != nil {
				ms.err = fmt.Errorf("column '%v': %v", unsanitizeFieldName(key), ms.err)
				return
			}
			row.Data[entry.index] = value
			columnBucket := entry.index / len(ms.columnPositions)
			columnPosition := entry.index % len(ms.columnPositions)
			// unset bit position
			ms.columnMaskBuckets[columnBucket] &= (^(uint64(1) << uint64(columnPosition)))
		}
	}
}

func (ms *MongoSourceStage) Pipeline() []bson.D {
	return ms.pipeline
}

func (ms *MongoSourceStage) Collection() string {
	return ms.collectionNames[0]
}

// mappingRegistry provides a way to get a field name from a table/column.
type mappingRegistry struct {
	columns []*Column
	fields  map[string]map[string]map[string]string
}

func (mr *mappingRegistry) addColumn(column *Column) {
	mr.columns = append(mr.columns, column)
}

func (mr *mappingRegistry) copy() *mappingRegistry {
	newMappingRegistry := &mappingRegistry{}
	newMappingRegistry.columns = make([]*Column, len(mr.columns))
	copy(newMappingRegistry.columns, mr.columns)
	if mr.fields != nil {
		for db, tables := range mr.fields {
			for tableName, columns := range tables {
				for columnName, fieldName := range columns {
					newMappingRegistry.registerMapping(db, tableName, columnName, fieldName)
				}
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

func (mr *mappingRegistry) lookupFieldName(dbName, tableName, columnName string) (string, bool) {
	if mr.fields == nil {
		return "", false
	}

	dbToColumn, ok := mr.fields[dbName]
	if !ok {
		return "", false
	}

	columnToField, ok := dbToColumn[tableName]
	if !ok {
		return "", false
	}

	field, ok := columnToField[columnName]

	return field, ok
}

func (mr *mappingRegistry) registerMapping(db, tbl, column, field string) bool {

	if mr.fields == nil {
		mr.fields = make(map[string]map[string]map[string]string)
	}
	if _, ok := mr.fields[db]; !ok {
		mr.fields[db] = make(map[string]map[string]string)
	}

	if _, ok := mr.fields[db][tbl]; !ok {
		mr.fields[db][tbl] = make(map[string]string)
	}

	if _, ok := mr.fields[db][tbl][column]; ok {
		return false
	}
	mr.fields[db][tbl][column] = field
	return true
}

// containsFieldName checks whether a field name exists across the entire registry
func (mr *mappingRegistry) containsFieldName(fieldName string) bool {

	for _, dbs := range mr.fields {
		for _, columns := range dbs {
			for _, field := range columns {
				if field == fieldName {
					return true
				}
			}
		}
	}
	return false
}

func (mr *mappingRegistry) String() string {
	var b bytes.Buffer
	for database, tables := range mr.fields {
		for table, entry := range tables {
			for column, name := range entry {
				b.WriteString(fmt.Sprintf("%v.%v.%v => %v\n", database, table, column, name))
			}
		}
	}
	return b.String()
}
