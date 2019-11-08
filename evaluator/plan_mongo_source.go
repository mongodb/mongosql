package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
)

// nullField represents a null bson.RawValue, we use this in the
// MongoSourceIter Next method in order to represent missing values.
// If a value is missing we will lookup nullField in the map.
var nullField = bson.RawValue{Type: bson.TypeNull, Value: []byte{}}

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
	tableType           string
	mappingRegistry     *mappingRegistry
	pipeline            *ast.Pipeline
	isDual              bool

	// A MongoSourceStage may be constrained by a LIMIT N. This field denotes the
	// value of N, so that there is some ability to guarantee the number of rows
	// that may be returned from this MongoSourceStage alone. Also, if this
	// MongoSourceStage is a dual, then LimitRowCount will be 1 because a dual
	// can have at most one row. If there exists no LIMIT N, then this field has
	// value -1.
	LimitRowCount int
}

// Children returns a slice of all the Node children of the Node.
func (MongoSourceStage) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (MongoSourceStage) ReplaceChild(i int, e Node) {
	panicWithInvalidIndex("MongoSourceStage", i, -1)
}

func newMongoSourceStage(db catalog.Database, table catalog.MongoDBTable, selectID int, aliasName string) *MongoSourceStage {
	ms := &MongoSourceStage{
		selectIDs:     []int{selectID},
		dbName:        string(db.Name()),
		tableNames:    []string{table.Name()},
		aliasNames:    []string{aliasName},
		tableType:     table.Type(),
		LimitRowCount: -1,
	}

	if len(ms.aliasNames) == 0 || ms.aliasNames[0] == "" {
		ms.aliasNames = ms.tableNames
	}

	ms.collation = table.Collation()
	collectionName := table.Collection()
	ms.collectionNames = []string{collectionName}
	ms.isShardedCollection = map[string]bool{collectionName: table.IsSharded()}
	ms.mappingRegistry = newMappingRegistry()

	for _, c := range table.Columns() {
		newColumn := c.Clone()
		newColumn.SelectID = selectID
		newColumn.Table = ms.aliasNames[0]
		newColumn.OriginalTable = ms.tableNames[0]
		newColumn.MappingRegistryName = ""

		ms.mappingRegistry.addColumn(newColumn)
		ms.mappingRegistry.registerMapping(ms.dbName, ms.aliasNames[0], c.Name, c.MongoName, false)
	}
	return ms
}

// NewMongoSourceStage creates a new MongoSourceStage from a catalog.MongoDBTable.
func NewMongoSourceStage(db catalog.Database, table catalog.MongoDBTable, selectID int, aliasName string) *MongoSourceStage {
	ms := newMongoSourceStage(db, table, selectID, aliasName)
	ms.pipeline = astutil.DeepCopyPipeline(table.Pipeline())
	return ms
}

// NewMongoSourceDualStage creates a new MongoSourceStage that represents a dual stage from a given catalog.MongoDBTable.
// Do not call if MongoDB version is less than 3.4, this function relies on the $collStats aggregation stage.
// Do not call if the connected server is a mongos, as $collStats will return 0 documents if called on a nonexistent table.
func NewMongoSourceDualStage(db catalog.Database, table catalog.MongoDBTable, selectID int, aliasName string) PlanStage {
	ms := newMongoSourceStage(db, table, selectID, aliasName)
	ms.isDual = true
	ms.LimitRowCount = 1

	// We use $collstats to get a guaranteed single document back, which is then used to house the fields from dual.
	ms.pipeline = ast.NewPipeline(
		ast.NewCollStatsStage(nil, nil, nil),
		ast.NewLimitStage(1), // Avoid getting more than one document back in sharded case.
		// By projecting a field that does not exist, we create an empty document.
		// This is a small optimization to throw out the output of $collStats.
		ast.NewProjectStage(ast.NewIncludeProjectItem(ast.NewFieldRef("newField", nil))),
	)

	return ms
}

func (ms *MongoSourceStage) clone() PlanStage {
	return &MongoSourceStage{
		selectIDs:           ms.selectIDs,
		dbName:              ms.dbName,
		tableNames:          ms.tableNames,
		aliasNames:          ms.aliasNames,
		collectionNames:     ms.collectionNames,
		tableType:           ms.tableType,
		isShardedCollection: ms.isShardedCollection,
		collation:           ms.collation,
		mappingRegistry:     ms.mappingRegistry.copy(),
		pipeline:            astutil.DeepCopyPipeline(ms.pipeline),
		isDual:              ms.isDual,
		LimitRowCount:       ms.LimitRowCount,
	}
}

func (ms *MongoSourceStage) isView() bool {
	return ms.tableType == catalog.View
}

// IsDual will return whether the MongoSourceStage represents a dual stage.
func (ms *MongoSourceStage) IsDual() bool {
	return ms.isDual
}

// getAggregationCursor get a cursor over MongoDB results from an Aggregation Pipeline.
func (ms *MongoSourceStage) getAggregationCursor(ctx context.Context, cfg *ExecutionConfig) (mongodb.Cursor, error) {
	errChan := make(chan error, 1)

	var cursor mongodb.Cursor
	var bsonPipeline []bson.D
	var err error

	procutil.PanicSafeGo(func() {
		// Convert from ast.Pipeline to []bson.D
		bsonPipeline, err = astutil.DeparsePipeline(ms.pipeline)
		if err != nil {
			errChan <- err
			return
		}
		cursor, err = cfg.commandHandler.Aggregate(ctx, ms.dbName, ms.collectionNames[0], bsonPipeline)
		errChan <- err
	}, func(err interface{}) {
		cfg.lg.Errf(log.Admin, "MongoDB data access session closed: %v", err)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err = <-errChan:
	}

	return cursor, err
}

// FastMongoSourceIter implements DocIter. It is an Iterator over raw BSON
// Documents.
type FastMongoSourceIter struct {
	// cursor is an implementation for getting data directly
	// from MongoDB.
	cursor mongodb.Cursor
	// columnInfo is a slice representing the field names,
	// in order, expected in the returned document.
	columnInfo []results.ColumnInfo
	// err holds any error that may occur during iteration.
	err error
}

// FastOpen opens a more optimized Iter over raw BSON documents returned from
// MongoDB in cases where no in-memory evaluation is needed to handle a query.
func (ms *MongoSourceStage) FastOpen(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (results.DocIter, error) {
	columns := ms.mappingRegistry.columns
	lenColumns := len(columns)
	uniqueFields := make(map[string]struct{}, lenColumns)

	columnInfo := make([]results.ColumnInfo, len(columns))
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
		uuidSubType := types.EvalBinary
		if c.MongoType == schema.MongoUUIDJava {
			uuidSubType = types.EvalJavaUUID
		} else if c.MongoType == schema.MongoUUIDCSharp {
			uuidSubType = types.EvalCSharpUUID
		}
		columnInfo[i] = results.ColumnInfo{
			Field:       mappedFieldName,
			Type:        c.EvalType,
			UUIDSubtype: uuidSubType,
		}
	}

	cursor, err := ms.getAggregationCursor(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &FastMongoSourceIter{
		cursor:     cursor,
		columnInfo: columnInfo,
		err:        nil,
	}, err
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (ms *FastMongoSourceIter) Next(ctx context.Context, doc *bson.Raw) bool {
	return ms.cursor.NextRaw(ctx, doc)
}

// GetColumnInfo returns the slice of ColumnInfo necessary for streaming the results.
func (ms *FastMongoSourceIter) GetColumnInfo() []results.ColumnInfo {
	return ms.columnInfo
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (ms *FastMongoSourceIter) Err() error {
	if err := ms.cursor.Err(); err != nil {
		return err
	}
	return ms.err
}

// Close closes the iterator, returning any error encountered while doing so.
func (ms *FastMongoSourceIter) Close() error {
	return ms.cursor.Close(context.Background())
}

// buildProjectBodyForMongoSource builds a $project/$addFields body to flatten
// embedded documents, it also returns the updated fields and whether or not any
// embedded documents actually occurred. The interface is busy, but pulling this
// function out of Open allows for us to unit test this functionality.
func buildProjectBodyForMongoSource(
	fields []string,
	fieldNamesSet map[string]struct{},
	columns results.Columns,
	isAtLeast34 bool,
) ([]*ast.AddFieldsItem, []string, bool) {
	// getUniqueFieldName will be used when creating $project/$addFields
	// body names for embedded documents.
	getUniqueFieldName := func(fieldName string) string {
		ret := fieldName
		for i := 0; ; i++ {
			if _, ok := fieldNamesSet[ret]; !ok {
				fieldNamesSet[ret] = struct{}{}
				return ret
			}
			ret = fieldName + strconv.Itoa(i)
		}
	}
	projectBody := make([]*ast.AddFieldsItem, 0, len(fields))
	hasEmbeddedDocs := false
	for i, mappedFieldName := range fields {
		fieldIsEmbedded := strings.Contains(mappedFieldName, ".")
		hasEmbeddedDocs = hasEmbeddedDocs || fieldIsEmbedded
		// If the field is embedded, it needs to be added to the $addFields body,
		// or if the server version < 3.4, we need to add it because all fields
		// must be project, however, the $project will only be added if at least
		// one embedded field is found, this just allows us to accomplish this
		// within one loop.
		if fieldIsEmbedded {
			flattenedFieldName := getUniqueFieldName(sanitizeFieldName(mappedFieldName))
			// Now we overwrite the previous non-flattened name.
			fields[i] = flattenedFieldName
			// In most cases getProjectedFieldName will simply add a "$" to the beginning
			// of the mappedFieldName. However, if the field is a 2d geo array
			// (EvalArrNumeric), it will properly add array indexing to get the
			// two values out.
			projectBody = append(projectBody,
				ast.NewAddFieldsItem(
					flattenedFieldName,
					getProjectedFieldName(mappedFieldName, columns[i].EvalType),
				),
			)
		} else if !isAtLeast34 {
			// We have to add this column if !asAtLeast34 or we will drop fields.
			// Note that even though we are building the projectBody, it will not
			// actually be added to the pipeline unless hasEmbeddedDocs is true.
			projectBody = append(projectBody,
				ast.NewAddFieldsItem(mappedFieldName, astutil.TrueLiteral),
			)
		}
	}
	return projectBody, fields, hasEmbeddedDocs
}

// Open creates an Iter over rows returned from MongoDB.
func (ms *MongoSourceStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (results.RowIter, error) {
	columns := ms.mappingRegistry.columns

	// We need to add a last $project or $addFields stage to flatten any embedded
	// documents/array indices. This might be redundant, but note that this does not occur in
	// FastOpen, which should always be called when we have full
	// pushdown (or nearly full pushdown with top Union stages).
	// We only add the $project or $addFields stage if we see an embedded doc or array index,
	// to keep simple queries simple. In the case of $addFields, only the embedded fields
	// are touched, in a $project we must make sure to keep all the normal fields as well.

	// $addFields was introduced in 3.4, only used $addFields if >= 3.4.
	isAtLeast34 := false
	if procutil.VersionAtLeast(cfg.mongoDBVersion, []uint8{3, 4, 0}) {
		isAtLeast34 = true
	}
	fields := make([]string, len(columns))
	fieldNamesSet := make(map[string]struct{}, len(columns))
	// first collect all the fieldNames
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
		fields[i] = mappedFieldName
		fieldNamesSet[mappedFieldName] = struct{}{}
	}

	// Now we potentially build a $project/$addFields.
	projectBody, fields, hasEmbeddedDocs := buildProjectBodyForMongoSource(fields,
		fieldNamesSet, columns, isAtLeast34)
	// If the there are embedded documents we will add a project to flatten them,
	// otherwise we will not change the pipeline.
	if hasEmbeddedDocs {
		if isAtLeast34 {
			ms.pipeline.Stages = append(ms.pipeline.Stages,
				ast.NewAddFieldsStage(projectBody...),
			)
		} else {
			projectItems := make([]ast.ProjectItem, len(projectBody))
			for i, afi := range projectBody {
				projectItems[i] = ast.NewAssignProjectItem(afi.Name, afi.Expr)
			}
			ms.pipeline.Stages = append(ms.pipeline.Stages,
				ast.NewProjectStage(projectItems...),
			)
		}
	}

	cursor, err := ms.getAggregationCursor(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// set the fieldValueMap for each field to be the nullField
	// for the first call to Next. Subsequent calls to Next
	// will set the values back to the nullField. This
	// allows us to catch missing values.
	fieldValueMap := make(map[string]*bson.RawValue, len(columns))
	for _, field := range fields {
		fieldValueMap[field] = &nullField
	}
	return &MongoSourceIter{
		cfg:             cfg,
		err:             nil,
		fields:          fields,
		cursor:          cursor,
		mappingRegistry: ms.mappingRegistry,
		fieldValueMap:   fieldValueMap,
	}, err
}

// MongoSourceIter returns rows sourced from MongoDB documents.
type MongoSourceIter struct {
	cfg *ExecutionConfig
	// err holds any error that may occur during iteration.
	err error
	// fields keeps track of the field name for each column.
	fields []string
	// cursor is an implementation for getting data directly
	// from MongoDB.
	cursor mongodb.Cursor
	// mappingRegistry holds all columns that must be returned
	// from data gotten in the iterator.
	mappingRegistry *mappingRegistry
	// fieldValueMap stores the mapping from field names to bson.RawValues.
	fieldValueMap map[string]*bson.RawValue
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (ms *MongoSourceIter) Next(ctx context.Context, row *results.Row) bool {
	d := bson.Raw{}
	if !ms.cursor.NextRaw(ctx, &d) {
		return false
	}

	lenCols := len(ms.mappingRegistry.columns)
	if len(row.Data) != lenCols {
		row.Data = make(results.RowValues, lenCols)
	}

	vs, err := d.Elements()
	if err != nil {
		return false
	}
	for _, v := range vs {
		val := v.Value()
		ms.fieldValueMap[v.Key()] = &val
	}

	for i, col := range ms.mappingRegistry.columns {
		fieldName := ms.fields[i]
		field := ms.fieldValueMap[fieldName]
		// Set ms.fieldValueMap to have the nullField for the next call to Next.
		ms.fieldValueMap[fieldName] = &nullField
		sqlValue, err := values.BSONValueToSQLValue(
			ms.cfg.sqlValueKind,
			types.EvalType(field.Type),
			col.UUIDSubType,
			field.Value,
		)
		if err != nil {
			ms.err = err
			return false
		}
		converted := values.ConvertTo(sqlValue, col.EvalType)
		row.Data[i] = results.NewRowValue(col.SelectID, col.Database, col.Table, col.Name, converted)
	}
	if ms.err != nil {
		return false
	}

	ms.err = ms.cfg.memoryMonitor.Acquire(row.Data.Size())
	return ms.err == nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (ms *MongoSourceStage) Columns() []*results.Column {
	return ms.mappingRegistry.columns
}

// Collation returns the collation to use for comparisons.
func (ms *MongoSourceStage) Collation() *collation.Collation {
	return ms.collation
}

// PipelineString returns the string representation of the stage's pipeline.
func (ms *MongoSourceStage) PipelineString() string {
	b := bytes.NewBufferString("")
	b.WriteRune('[')
	if len(ms.pipeline.Stages) > 0 {
		prettyPipeline, err := astutil.PipelineJSON(ms.pipeline, 0, false)
		if err != nil { // marshaling as json failed, fall back to Sprintf
			prettyPipeline = astutil.PipelineString(ms.pipeline, 0)
		}
		b.Write(prettyPipeline)
	}
	b.WriteRune(']')
	return b.String()
}

// Close closes the iterator, returning any error encountered while doing so.
func (ms *MongoSourceIter) Close() error {
	return ms.cursor.Close(context.Background())
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (ms *MongoSourceIter) Err() error {
	if err := ms.cursor.Err(); err != nil {
		return err
	}
	return ms.err
}

// Pipeline returns the aggregation pipeline used by this MongoSourceStage.
func (ms *MongoSourceStage) Pipeline() *ast.Pipeline {
	return ms.pipeline
}

// Collection gets the name of the collection against which this MongoSourceStage
// will run its aggregation pipeline.
func (ms *MongoSourceStage) Collection() string {
	return ms.collectionNames[0]
}

// mappingRegistry provides a way to get a field name from a table/column.
type mappingRegistry struct {
	columns []*results.Column
	fields  map[string]map[string]map[string]string

	// fieldNames maps from field name to a bool
	//   - true indicates the field name refers to a variable reference
	//   - false indicates the field name refers to a field reference
	fieldNames map[string]bool
}

func (mr *mappingRegistry) addColumn(column *results.Column) {
	mr.columns = append(mr.columns, column)
}

// newMappingRegistry returns an initialized registry.
func newMappingRegistry() *mappingRegistry {
	return &mappingRegistry{
		fieldNames: make(map[string]bool),
	}
}

func (mr *mappingRegistry) copy() *mappingRegistry {
	newMappingRegistry := newMappingRegistry()
	newMappingRegistry.columns = cloneColumns(mr.columns)

	if mr.fields != nil {
		for db, tables := range mr.fields {
			for tableName, columns := range tables {
				for columnName, fieldName := range columns {
					newMappingRegistry.registerMapping(db, tableName, columnName, fieldName, mr.fieldNames[fieldName])
				}
			}
		}
	}
	return newMappingRegistry
}

// merge returns a new mapping registry that consists of all entries from
// mr and foreign. A prefix is used for all column names on the foreign side.
func (mr *mappingRegistry) merge(foreign *mappingRegistry, prefix string) *mappingRegistry {
	newMappingRegistry := mr.copy()

	newMappingRegistry.columns = append(newMappingRegistry.columns,
		foreign.columns...)
	if foreign.fields != nil {
		for database, tables := range foreign.fields {
			for tableName, columns := range tables {
				for columnName, fieldName := range columns {
					newMappingRegistry.registerMapping(database, tableName, columnName, prefix+"."+fieldName, false)
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

func (mr *mappingRegistry) lookupFieldRef(dbName, tableName, columnName string) (ast.Ref, bool) {
	fieldName, ok := mr.lookupFieldName(dbName, tableName, columnName)
	if !ok {
		return nil, false
	}

	if mr.fieldNames[fieldName] {
		return ast.NewVariableRef(fieldName), true
	}

	return astutil.FieldRefFromFieldName(fieldName), true
}

func (mr *mappingRegistry) registerMapping(db, tbl, column, field string, isVariable bool) bool {
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
	mr.fieldNames[field] = isVariable
	return true
}

// containsFieldName checks whether a field name exists across the entire registry
func (mr *mappingRegistry) containsFieldName(fieldName string) bool {
	_, ok := mr.fieldNames[fieldName]
	return ok
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
