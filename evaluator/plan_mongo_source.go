package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

// nullField represents a null bson.Raw, we use this in the MongoSourceIter
// Next method in order to represent missing values. If a value is missing
// we will lookup nullField in the map.
var nullField = bson.Raw{Kind: byte(EvalNull), Data: []byte{}}

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

	// A MongoSourceStage may require evaluation of one or more
	// NonCorrelatedSubqueryFutures before its evaluation can commence.
	// This is empty if a MongoSourceStage has no piecewise dependencies.
	piecewiseDeps []*NonCorrelatedSubqueryFuture

	// A MongoSourceStage representing a portion of a subquery may reference
	// columns from parent scopes. When pushed down, these columns are stored
	// as CorrelatedSubqueryColumnFutures, and must be evaluated before the
	// pipeline can be marshalled.
	correlatedColumns []*CorrelatedSubqueryColumnFuture
}

// NewMongoSourceStage creates a new MongoSourceStage from a catalog.MongoTable.
func NewMongoSourceStage(db *catalog.Database,
	table *catalog.MongoTable,
	selectID int,
	aliasName string) *MongoSourceStage {

	ms := &MongoSourceStage{
		selectIDs:         []int{selectID},
		dbName:            string(db.Name),
		tableNames:        []string{string(table.Name())},
		aliasNames:        []string{aliasName},
		tableType:         table.Type(),
		piecewiseDeps:     []*NonCorrelatedSubqueryFuture{},
		correlatedColumns: []*CorrelatedSubqueryColumnFuture{},
	}

	if len(ms.aliasNames) == 0 || ms.aliasNames[0] == "" {
		ms.aliasNames = ms.tableNames
	}

	ms.collation = table.Collation()
	ms.collectionNames = []string{table.CollectionName}
	ms.isShardedCollection = map[string]bool{table.CollectionName: table.IsSharded()}
	ms.mappingRegistry = newMappingRegistry()

	primaryKeys := catalog.Columns(table.PrimaryKeys())

	for _, c := range table.Columns() {
		mc := c.(*catalog.MongoColumn)
		column := NewColumn(selectID,
			ms.aliasNames[0],
			ms.tableNames[0],
			ms.dbName,
			string(mc.Name()),
			string(mc.Name()),
			"",
			SQLTypeToEvalType(mc.Type()),
			mc.MongoType,
			primaryKeys.Contains(mc.Name()),
		)
		ms.mappingRegistry.addColumn(column)
		ms.mappingRegistry.registerMapping(ms.dbName,
			ms.aliasNames[0],
			string(mc.Name()),
			mc.MongoName)
	}

	ms.pipeline = bsonutil.DeepCopyPipeline(table.Pipeline)
	return ms
}

func (ms *MongoSourceStage) clone() PlanStage {
	deps := make([]*NonCorrelatedSubqueryFuture, len(ms.piecewiseDeps))
	copy(deps, ms.piecewiseDeps)
	corr := make([]*CorrelatedSubqueryColumnFuture, len(ms.correlatedColumns))
	copy(corr, ms.correlatedColumns)

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
		pipeline:            bsonutil.DeepCopyPipeline(ms.pipeline),
		piecewiseDeps:       deps,
		correlatedColumns:   corr,
	}
}

func (ms *MongoSourceStage) isView() bool {
	return ms.tableType == catalog.View
}

// getAggregationCursor get a cursor over MongoDB results from an Aggregation Pipeline.
func (ms *MongoSourceStage) getAggregationCursor(ctx context.Context, cfg *ExecutionConfig) (mongodb.Cursor, error) {
	errChan := make(chan error, 1)

	var iter mongodb.Cursor
	var err error

	util.PanicSafeGo(func() {
		iter, err = cfg.commandHandler.Aggregate(ms.dbName, ms.collectionNames[0], ms.pipeline)
		errChan <- err
	}, func(err interface{}) {
		cfg.lg.Errf(log.Admin, "MongoDB data access session closed: %v", err)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err = <-errChan:
	}

	return iter, err
}

// ColumnInfo keeps track of the data needed to correctly deserialize data from
// a MongoSourceStage.
type ColumnInfo struct {
	// Field is the name of the specific MongoDB field.
	Field string
	// Type is the byte corresponding to the type MongoDRDL specifies for
	// the given column. The byte corresponds to the BSON kind byte, iff
	// the column type is a BSON type. Some Column types are not BSON
	// types: e.g., Date, which needs to drop the Time portions of a
	// Timestamp for formatting purposes because BSON datetime objects
	// store both the date and the time. This is represented using
	// the type alias EvalType.
	Type EvalType
	// UUIDSubtype is needed to handle UUIDs written by the Java and CSharp
	// drivers, which store UUIDs using different byte orders.
	UUIDSubtype EvalType
}

// FastMongoSourceIter implements FastIter. It is an Iterator over raw BSON
// Documents.
type FastMongoSourceIter struct {
	// iter is an implementation for getting data directly
	// from MongoDB.
	iter mongodb.Cursor
	// columnInfo is a slice representing the field names,
	// in order, expected in the returned document.
	columnInfo []ColumnInfo
	// err holds any error that may occur during iteration.
	err error
}

// FastOpen opens a more optimized Iter over raw BSON documents returned from
// MongoDB in cases where no in-memory evaluation is needed to handle a query.
func (ms *MongoSourceStage) FastOpen(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (FastIter, error) {
	err := ms.resolveDeps(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

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
		uuidSubType := EvalNone
		if c.MongoType == schema.MongoUUIDJava {
			uuidSubType = EvalJavaUUID
		} else if c.MongoType == schema.MongoUUIDCSharp {
			uuidSubType = EvalCSharpUUID
		}
		columnInfo[i] = ColumnInfo{
			Field:       mappedFieldName,
			Type:        c.EvalType,
			UUIDSubtype: uuidSubType,
		}
	}

	iter, err := ms.getAggregationCursor(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &FastMongoSourceIter{
		iter:       iter,
		columnInfo: columnInfo,
		err:        nil,
	}, err
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (ms *FastMongoSourceIter) Next(ctx context.Context, doc *bson.RawD) bool {
	return ms.iter.Next(ctx, doc)
}

// GetColumnInfo returns the slice of ColumnInfo necessary for streaming the results.
func (ms *FastMongoSourceIter) GetColumnInfo() []ColumnInfo {
	return ms.columnInfo
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (ms *FastMongoSourceIter) Err() error {
	if err := ms.iter.Err(); err != nil {
		return err
	}
	return ms.err
}

// Close closes the iterator, returning any error encountered while doing so.
func (ms *FastMongoSourceIter) Close() error {
	return ms.iter.Close(context.Background())
}

// buildProjectBodyForMongoSource builds a $project/$addFields body to flatten
// embedded documents, it also returns the updated fields and whether or not any
// embedded documents actually occurred. The interface is busy, but pulling this
// function out of Open allows for us to unit test this functionality.
func buildProjectBodyForMongoSource(fields []string,
	fieldNamesSet map[string]struct{}, columns Columns,
	isAtLeast34 bool) (bson.D, []string, bool) {
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
	projectBody := bson.D{}
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
				bson.DocElem{
					Name:  flattenedFieldName,
					Value: getProjectedFieldName(mappedFieldName, columns[i].EvalType),
				})
		} else if !isAtLeast34 {
			// We have to add this column if !asAtLeast34 or we will drop fields.
			// Note that even though we are building the projectBody, it will not
			// actually be added to the pipeline unless hasEmbeddedDocs is true.
			projectBody = append(projectBody, bson.DocElem{Name: mappedFieldName,
				Value: true})
		}
	}
	return projectBody, fields, hasEmbeddedDocs
}

func (ms *MongoSourceStage) resolveDeps(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	for i, dep := range ms.piecewiseDeps {
		cfg.lg.Debugf(log.Dev,
			"executing piecewise dependency %d:\n%s",
			i, PrettyPrintPlan(dep.plan),
		)
		err := dep.Evaluate(ctx, cfg, st)
		if err != nil {
			if _, ok := err.(*mysqlerrors.MySQLError); ok {
				return err
			}
			return fmt.Errorf("error evaluating piecewise dependency: %v", err)
		}
	}
	for i, col := range ms.correlatedColumns {
		cfg.lg.Debugf(log.Dev,
			"resolving pushed down correlated column %d:\n%s",
			i, col.String,
		)
		err := col.Evaluate(cfg, st)
		if err != nil {
			return fmt.Errorf("error resolving pushed down correlated column: %v", err)
		}
	}
	return nil
}

// Open creates an Iter over rows returned from MongoDB.
func (ms *MongoSourceStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (Iter, error) {
	err := ms.resolveDeps(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

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
	if util.VersionAtLeast(cfg.mongoDBVersion, []uint8{3, 4, 0}) {
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
		stageName := "$project"
		if isAtLeast34 {
			stageName = "$addFields"
		}
		project := bson.D{{Name: stageName, Value: projectBody}}
		ms.pipeline = append(ms.pipeline, project)
	}

	iter, err := ms.getAggregationCursor(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// set the fieldValueMap for each field to be the nullField
	// for the first call to Next. Subsequent calls to Next
	// will set the values back to the nullField. This
	// allows us to catch missing values.
	fieldValueMap := make(map[string]*bson.Raw, len(columns))
	for _, field := range fields {
		fieldValueMap[field] = &nullField
	}
	return &MongoSourceIter{
		cfg:             cfg,
		err:             nil,
		fields:          fields,
		iter:            iter,
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
	// iter is an implementation forgetting data directly
	// from MongoDB.
	iter mongodb.Cursor
	// mappingRegistry holds all columns that must be returned
	// from data gotten in the iterator.
	mappingRegistry *mappingRegistry
	// fieldValueMap stores the mapping from field names to bson.Raw values.
	fieldValueMap map[string]*bson.Raw
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (ms *MongoSourceIter) Next(ctx context.Context, row *Row) bool {
	document := &bson.RawD{}
	if !ms.iter.Next(ctx, document) {
		return false
	}

	lenCols := len(ms.mappingRegistry.columns)
	if len(row.Data) != lenCols {
		row.Data = make([]Value, lenCols)
	}

	values := *document
	for i := range values {
		ms.fieldValueMap[values[i].Name] = &(values[i].Value)
	}

	for i, col := range ms.mappingRegistry.columns {
		fieldName := ms.fields[i]
		field := ms.fieldValueMap[fieldName]
		// Set ms.fieldValueMap to have the nullField for the next call to Next.
		ms.fieldValueMap[fieldName] = &nullField
		sqlValue, err := BSONValueToSQLValue(
			ms.cfg.sqlValueKind,
			EvalType(field.Kind),
			col.UUIDSubType,
			field.Data,
		)
		if err != nil {
			ms.err = err
			return false
		}
		converted := ConvertTo(sqlValue, col.EvalType)
		row.Data[i] = NewValue(col.SelectID, col.Database, col.Table, col.Name, converted)
	}
	if ms.err != nil {
		return false
	}

	ms.err = ms.cfg.memoryMonitor.Acquire(row.Data.Size())
	return ms.err == nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (ms *MongoSourceStage) Columns() []*Column {
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
	if len(ms.pipeline) > 0 {
		prettyPipeline, err := pipelineJSON(ms.pipeline, 0, false)
		if err != nil { // marshaling as json failed, fall back to Sprintf
			prettyPipeline = pipelineString(ms.pipeline, 0)
		}
		b.Write(prettyPipeline)
	}
	b.WriteRune(']')
	return b.String()
}

// Close closes the iterator, returning any error encountered while doing so.
func (ms *MongoSourceIter) Close() error {
	return ms.iter.Close(context.Background())
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (ms *MongoSourceIter) Err() error {
	if err := ms.iter.Err(); err != nil {
		return err
	}
	return ms.err
}

// Pipeline returns the aggregation pipeline used by this MongoSourceStage.
func (ms *MongoSourceStage) Pipeline() []bson.D {
	return ms.pipeline
}

// Collection gets the name of the collection against which this MongoSourceStage
// will run its aggregation pipeline.
func (ms *MongoSourceStage) Collection() string {
	return ms.collectionNames[0]
}

// mappingRegistry provides a way to get a field name from a table/column.
type mappingRegistry struct {
	columns    []*Column
	fields     map[string]map[string]map[string]string
	fieldNames map[string]struct{}
}

func (mr *mappingRegistry) addColumn(column *Column) {
	mr.columns = append(mr.columns, column)
}

// newMappingRegistry returns an initialized registry.
func newMappingRegistry() *mappingRegistry {
	return &mappingRegistry{
		fieldNames: make(map[string]struct{}),
	}
}

func (mr *mappingRegistry) copy() *mappingRegistry {
	newMappingRegistry := newMappingRegistry()
	newMappingRegistry.columns = cloneColumns(mr.columns)

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
					newMappingRegistry.registerMapping(database, tableName, columnName,
						prefix+"."+fieldName)
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
	mr.fieldNames[field] = struct{}{}
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

// NonCorrelatedSubqueryFuture represents a value that must be obtained by executing a query
// plan before it can be used. A piece should be placed into an
// aggregation pipeline when pushing down an expression containing
// a non-correlated SQLSubqueryExpr into a MongoSourceStage.
type NonCorrelatedSubqueryFuture struct {
	evaluated bool
	value     interface{}
	plan      PlanStage
}

// NewNonCorrelatedSubqueryFuture returns a new NonCorrelatedSubqueryFuture based on the provided query plan.
func NewNonCorrelatedSubqueryFuture(p PlanStage) *NonCorrelatedSubqueryFuture {
	return &NonCorrelatedSubqueryFuture{
		evaluated: false,
		value:     nil,
		plan:      p,
	}
}

// Evaluate executes this piece's query plan and sets its value to the result.
func (p *NonCorrelatedSubqueryFuture) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	if p.evaluated {
		panic("cannot evaluate piece twice")
	}

	iter, err := p.plan.Open(ctx, cfg, st)
	if err != nil {
		return err
	}
	iter = newMemoryIter(cfg, iter)

	row := &Row{}
	ok := iter.Next(ctx, row)
	if !ok {
		// if the iter failed with an error, return it
		err = iter.Err()
		if err != nil {
			return err
		}

		// otherwise, there are no results in the result set, so we return NULL
		p.evaluated = true
		p.value = nil
		return nil
	}

	nullRow := &Row{}
	if iter.Next(ctx, nullRow) {
		return mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
	}

	if len(row.Data) != 1 {
		// we shouldn't get here, because the algebrizer should have
		// rejected any query returning multiple columns
		panic("too many columns in correlated subquery expr result")
	}

	p.value = row.Data[0].Data.Value()
	p.evaluated = true

	return nil
}

// GetBSON returns this piece's cached result for BSON marshalling.
// This function panics if the piece has not yet been evaluated.
func (p *NonCorrelatedSubqueryFuture) GetBSON() (interface{}, error) {
	if !p.evaluated {
		panic("cannot marshal pipeline with unevaluated piece")
	}
	return p.value, nil
}

// CorrelatedSubqueryColumnFuture represents a value that must be obtained from a
// correlated column expr before it can be used. A CorrelatedSubqueryColumnFuture
// should be placed into an aggregation pipeline when pushing down a
// correlated SQLSubqueryExpr's internal query plan.
type CorrelatedSubqueryColumnFuture struct {
	evaluated  bool
	value      interface{}
	selectID   int
	database   string
	table      string
	column     string
	columnType ColumnType
}

// NewCorrelatedSubqueryColumnFuture returns a new CorrelatedSubqueryColumnFuture based on
// the provided SQLColumnExpr.
func NewCorrelatedSubqueryColumnFuture(expr *SQLColumnExpr) *CorrelatedSubqueryColumnFuture {
	return &CorrelatedSubqueryColumnFuture{
		evaluated:  false,
		value:      nil,
		selectID:   expr.selectID,
		database:   expr.databaseName,
		table:      expr.tableName,
		column:     expr.columnName,
		columnType: expr.columnType,
	}
}

// Evaluate resolves this CorrelatedSubqueryColumnFuture to a value.
func (cc *CorrelatedSubqueryColumnFuture) Evaluate(cfg *ExecutionConfig, st *ExecutionState) error {
	for _, row := range st.correlatedRows {
		if result, ok := row.GetField(cc.selectID, cc.database, cc.table, cc.column); ok {
			cc.value = ConvertTo(result, cc.columnType.EvalType)
			return nil
		}
	}

	// TODO BI-1883
	cc.value = NewSQLNull(cfg.sqlValueKind, cc.columnType.EvalType)
	return nil
}

// GetBSON returns the correlated column's cached result for BSON marshalling.
// This function panics if the column's value has not yet been resolved.
func (cc *CorrelatedSubqueryColumnFuture) GetBSON() (interface{}, error) {
	if !cc.evaluated {
		panic("cannot marshal pipeline with unresolved correlated column")
	}
	return cc.value, nil
}

func (cc *CorrelatedSubqueryColumnFuture) String() string {
	if cc.database != "" {
		return fmt.Sprintf("%v.%v.%v", cc.database, cc.table, cc.column)
	} else if cc.table != "" {
		return fmt.Sprintf("%v.%v", cc.table, cc.column)
	} else {
		return fmt.Sprintf("%v", cc.column)
	}
}

// MarshalJSON returns the JSON representation of this CorrelatedSubqueryColumnFuture.
func (cc *CorrelatedSubqueryColumnFuture) MarshalJSON() ([]byte, error) {
	return []byte(cc.String()), nil
}
