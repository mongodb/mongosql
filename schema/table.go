package schema

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/mongo"

	"github.com/kr/pretty"
)

// Table represents a configuration for a table.
type Table struct {
	// sqlName is the name of this table in the sql schema.
	sqlName string
	// mongoName is the MongoDB collection name this table maps to.
	mongoName string
	// pipeline is a MongoDB aggregation pipeline to transform data before mapping
	// it to a SQL table.
	pipeline []bson.D
	// columns is a map of normalized column names to the columns in this database.
	columns map[normalizedName]*Column
	// primaryKey is a slice of all the columns that comprise the primary key.
	primaryKey map[normalizedName]struct{}
	// cachedSortedColumns is the cached result of the last call to
	// ColumnsSorted. If it is non-nil when ColumnsSorted is called, it will be
	// used to avoid duplicating a potentially expensive sort. cachedSort is
	// invalidated (set to nil) whenever the columns map is modified.
	cachedSortedColumns []*Column
	cacheLock           sync.RWMutex

	// The following fields are only used during mongo-to-relational schema translation

	// parent is a pointer to an array table's parent table.
	parent *Table
	// isPostProcessed tracks whether this table has already copied columns from
	// its parent table.
	isPostProcessed bool
	// unwindPath is the path unwound to generate this table, if it is an array table.
	unwindPath string
}

// NewTable creates a new table with the provided fields.
func NewTable(lg log.Logger, tbl, col string, pipeline []bson.D,
	cols []*Column) (*Table, error) {

	primaryKeys := map[normalizedName]struct{}{}

	table := &Table{
		pipeline:   []bson.D{},
		sqlName:    tbl,
		mongoName:  col,
		columns:    map[normalizedName]*Column{},
		primaryKey: primaryKeys,
	}

	for _, col := range cols {
		table.AddColumn(lg, col, false)
	}

	for _, stage := range pipeline {
		err := table.AddPipelineStage(stage)
		if err != nil {
			return nil, err
		}
	}

	return table, nil
}

// NewTableWithUnwindPath creates a new table with the provided fields.
// The unwindPath is the path in the original document that is unwound to
// generate this table, only needed for array tables.
func NewTableWithUnwindPath(lg log.Logger, tbl, col string, pipeline []bson.D,
	cols []*Column, unwindPath string) (*Table, error) {
	ret, err := NewTable(lg, tbl, col, pipeline, cols)
	if err != nil {
		return nil, err
	}
	ret.unwindPath = unwindPath
	return ret, nil
}

// NewTableFromDRDL returns a new Table that is built from the provided DRDL
// table. Each column in the DRDL table is converted to a *Column and then added
// to the schema in order.
func NewTableFromDRDL(lg log.Logger, drdlTbl *drdl.Table) (*Table, error) {
	cols := make([]*Column, len(drdlTbl.Columns))
	for i, drdlCol := range drdlTbl.Columns {
		col, err := NewColumnFromDRDL(drdlCol)

		if err != nil {
			return nil, fmt.Errorf(`unable to create column "%v" from drdl: %v`,
				drdlCol.MongoName, err)
		}

		cols[i] = col
	}
	return NewTable(
		lg,
		drdlTbl.SQLName, drdlTbl.MongoName,
		drdlTbl.Pipeline,
		cols,
	)
}

// AddColumn adds the provided column to the table. If the column's name
// conflicts with the name of an existing column, its name will be changed to
// something that is unique within the table. If the column is a Geo2D field,
// two separate columns will be added, one for the longitude and one for the
// latitude.
func (t *Table) AddColumn(lg log.Logger, c *Column, isPK bool) {
	if strings.Trim(c.MongoName(), " ") == "" {
		lg.Warnf(log.Admin, "omitting column %q with whitespace-only name in table %q",
			c.SQLName(), t.SQLName())
		return
	}

	if c.MongoType() == MongoGeo2D {
		t.addGeoColumn(lg, c, isPK)
		return
	}

	col := t.Column(c.SQLName())
	if col != nil {
		initName := col.SQLName()
		c.sqlName = t.uniqueColumnName(c.SQLName())
		if c.SQLName() != initName {
			lg.Warnf(log.Admin, "found 2 columns with the same case-insensitive "+
				"name in table %q: renamed %q to %q", t.SQLName(), initName, c.SQLName())
		}
	}

	t.addColumn(c, isPK)
}

// addColumn unconditionally adds the provided column to this table, not
// performing any validation of the Column's SQLName.
func (t *Table) addColumn(c *Column, isPK bool) {
	key := normalizeSQLName(c.SQLName())
	t.columns[key] = c

	if isPK {
		t.primaryKey[key] = struct{}{}
	}
	t.invalidateCachedSortedColumns()
}

// addGeoColumn is a helper function for adding Geo2D columns to a table.
func (t *Table) addGeoColumn(lg log.Logger, c *Column, isPK bool) {
	for i, suffix := range []string{"_longitude", "_latitude"} {
		newSQLName := c.SQLName() + suffix
		newMongoName := fmt.Sprintf("%v.%v", c.MongoName(), i)
		lg.Infof(
			log.Admin,
			"adding column %q for %s component of geo2d column %q in table %q",
			newSQLName, suffix[1:], c.SQLName(), t.SQLName(),
		)

		newColumn := NewColumnWithSampledTypes(
			newSQLName,
			SQLArrNumeric,
			newMongoName,
			MongoFloat,
			[]mongo.BSONType{mongo.Double},
		)

		t.AddColumn(lg, newColumn, isPK)
	}
	t.invalidateCachedSortedColumns()
}

// AddPipelineStage adds the provided BSON document to this table's pipeline.
// any extjson expressions used in the pipeline will be converted into proper
// BSON values.
func (t *Table) AddPipelineStage(doc bson.D) error {
	v, err := bsonutil.ConvertJSONValueToBSON(doc)
	if err != nil {
		return fmt.Errorf("unable to parse extended json: %v", err)
	}
	t.pipeline = append(t.pipeline, v.(bson.D))
	return nil
}

// cacheSortedColumns caches the provided sorted slice of columns.
func (t *Table) cacheSortedColumns(cols []*Column) {
	t.cacheLock.Lock()
	defer t.cacheLock.Unlock()

	t.cachedSortedColumns = make([]*Column, len(cols))
	copy(t.cachedSortedColumns, cols)
}

// ChangeColumnType changes the type of a schema.Column.
func (t *Table) ChangeColumnType(colName string, sqlType SQLType, mongoType MongoType) error {
	col := t.Column(colName)
	if col == nil {
		return fmt.Errorf("couldn't find column %q", colName)
	}

	// if the types are not changing, this is a no-op.
	if col.sqlType == sqlType && col.mongoType == mongoType {
		return nil
	}
	col.sqlType = sqlType
	col.mongoType = mongoType
	col.hasAlteredType = true
	return col.Validate()
}

// Column gets the column in this Table whose normalized SQLName matches the
// normalized form of the provided name. If no matching column exists in the
// table, nil is returned.
func (t *Table) Column(sqlName string) *Column {
	key := normalizeSQLName(sqlName)
	return t.columns[key]
}

// Columns returns a slice of all the columns in this Table.
func (t *Table) Columns() []*Column {
	cols := []*Column{}
	for _, col := range t.columns {
		cols = append(cols, col)
	}
	return cols
}

// NumColumns returns the number of columns in t.
func (t *Table) NumColumns() int64 {
	return int64(len(t.columns))
}

// ColumnsSorted returns a sorted slice of all the columns in this Table. _id columns
// will be sorted before all other columns, and the remaining columns will be
// sorted in ascending order by MongoName and SQLName (in that order).
func (t *Table) ColumnsSorted() []*Column {
	cache := t.getCachedSortedColumns()
	if cache != nil {
		return cache
	}

	keys := []normalizedName{}
	for key := range t.columns {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if t.columns[keys[i]].MongoName() == MongoPrimaryKey {
			return true
		}
		if t.columns[keys[j]].MongoName() == MongoPrimaryKey {
			return false
		}
		return keys[i] < keys[j]
	})

	cols := []*Column{}
	for _, key := range keys {
		cols = append(cols, t.columns[key])
	}

	t.cacheSortedColumns(cols)
	return cols
}

// DeepCopy returns a copy of this table. All fields except for the table's
// pipeline will be deep copies.
func (t *Table) DeepCopy() *Table {
	if t == nil {
		return nil
	}

	cols := map[normalizedName]*Column{}
	for key, col := range t.columns {
		cols[key] = col.DeepCopy()
	}

	pkCols := map[normalizedName]struct{}{}
	for colName := range t.primaryKey {
		pkCols[colName] = struct{}{}
	}

	parent := t.parent.DeepCopy()

	pipeline := bsonutil.DeepCopyDSlice(t.pipeline)

	return &Table{
		sqlName:    t.sqlName,
		mongoName:  t.mongoName,
		pipeline:   pipeline,
		columns:    cols,
		parent:     parent,
		primaryKey: pkCols,
		unwindPath: t.unwindPath,
	}
}

// Equals checks whether this Table is equal to the provided Table. The equality
// check ignores the parent and isPostProcessed fields, as those fields are only
// used for persisting some useful state during mongo-to-relational schema
// mapping.
func (t *Table) Equals(other *Table) error {
	if t == other {
		return nil
	}
	if t == nil {
		return fmt.Errorf("this table is nil, but other table is non-nil")
	}
	if other == nil {
		return fmt.Errorf("this table is non-nil, but other table is nil")
	}

	if t.sqlName != other.sqlName {
		return fmt.Errorf("sqlNames %q and %q do not match", t.sqlName, other.sqlName)
	}

	if t.mongoName != other.mongoName {
		return fmt.Errorf("mongoNames %q and %q do not match", t.mongoName, other.mongoName)
	}

	if len(t.pipeline) != len(other.pipeline) {
		return fmt.Errorf("pipeline lengths do not match:\nactual: %v\nexpected: %v",
			len(t.pipeline), len(other.pipeline))
	}

	if len(t.pipeline) > 0 && !bsonutil.PipelineEqual(t.pipeline, other.pipeline) {
		return fmt.Errorf("pipelines do not match:\nactual: %s\nexpected: %s",
			pretty.Formatter(t.pipeline), pretty.Formatter(other.pipeline))
	}

	if len(t.columns) != len(other.columns) {
		return fmt.Errorf(
			"this table has columns:\n%v\nother table has columns:\n%v",
			other.columns, t.columns,
		)
	}

	for key, column := range t.columns {
		otherColumn, ok := other.columns[key]
		if !ok {
			return fmt.Errorf("column %q missing from other table", column.SQLName())
		}
		err := column.Equals(otherColumn)
		if err != nil {
			return fmt.Errorf("columns with sqlName %q not equal: %v", column.SQLName(), err)
		}
	}

	return nil
}

// getCachedSortedColumns returns a shallow copy of this table's cached sort.
func (t *Table) getCachedSortedColumns() []*Column {
	t.cacheLock.RLock()
	defer t.cacheLock.RUnlock()

	if t.cachedSortedColumns == nil {
		return nil
	}
	cols := make([]*Column, len(t.cachedSortedColumns))
	copy(cols, t.cachedSortedColumns)
	return cols
}

// invalidateCachedSortedColumns invalidate's this table's currently cached
// sort.
func (t *Table) invalidateCachedSortedColumns() {
	t.cacheLock.Lock()
	defer t.cacheLock.Unlock()

	t.cachedSortedColumns = nil
}

// IsMongoNamePrimaryKey returns true if the provided MongoName is part of a
// primary key and false otherwise.
func (t *Table) IsMongoNamePrimaryKey(mongoName string) bool {
	if mongoName == MongoPrimaryKey {
		return true
	}

	for _, d := range t.pipeline {
		unwindVal, ok := d.Map()["$unwind"]
		if !ok {
			return false
		}

		unwind, ok := unwindVal.(bson.D)
		if !ok {
			return false
		}

		arrayIndexNameVal, ok := unwind.Map()["includeArrayIndex"]
		if !ok {
			continue
		}

		arrayIndexName, ok := arrayIndexNameVal.(string)
		if !ok {
			continue
		}

		if mongoName == arrayIndexName {
			return true
		}
	}

	return false
}

// IsSQLNamePrimaryKey returns whether the provided SQLName matches a
// primary-key column in this table.
func (t *Table) IsSQLNamePrimaryKey(sqlName string) bool {
	key := normalizeSQLName(sqlName)
	_, ok := t.primaryKey[key]
	return ok
}

// MongoName returns the name of this table's underlying collection.
func (t *Table) MongoName() string {
	return t.mongoName
}

// Parent returns this table's parent table.
func (t *Table) Parent() *Table {
	return t.parent
}

// Pipeline returns this table's pipeline.
func (t *Table) Pipeline() []bson.D {
	return t.pipeline
}

// PostProcess copies columns from this table's parent into the table. If
// preJoin is true, all of the parent columns will be copied; otherwise, only
// the primary key columns are included. If the parent is nil or if this table
// has already been post-processed, no action is taken.
func (t *Table) PostProcess(lg log.Logger, preJoin bool) {
	if t.parent == nil || t.isPostProcessed {
		return
	}

	// ensure parent is post-processed
	t.parent.PostProcess(lg, preJoin)

	// Add parent columns
	for _, c := range t.parent.ColumnsSorted() {
		col := c.DeepCopy()
		isPK := t.parent.IsSQLNamePrimaryKey(col.SQLName())
		if !isPK && !preJoin {
			continue
		}
		t.AddColumn(lg, col, isPK)
	}

	// prepend parent pipeline
	pipeline := bsonutil.DeepCopyDSlice(t.parent.Pipeline())
	pipeline = append(pipeline, t.pipeline...)
	t.pipeline = pipeline

	t.isPostProcessed = true
	t.invalidateCachedSortedColumns()
}

// RemoveColumnBySQLName looks for a column whose normalized SQLName matches the
// normalized form of the provided name. If the column is found, it is removed
// from the table. If not, an error is returned.
func (t *Table) RemoveColumnBySQLName(sqlName string) error {
	key := normalizeSQLName(sqlName)
	if _, ok := t.columns[key]; !ok {
		return fmt.Errorf("column %q not found in table", sqlName)
	}
	delete(t.columns, key)
	delete(t.primaryKey, key)
	t.invalidateCachedSortedColumns()
	return nil
}

// RenameColumn replaces the column with SQLName "oldName" with a new Column of
// SQLName "newName".
func (t *Table) RenameColumn(oldName, newName string) error {
	col := t.Column(newName)
	if col != nil {
		return fmt.Errorf("column with SQLName %q already exists", newName)
	}

	col = t.Column(oldName)
	if col == nil {
		return fmt.Errorf("could not find column %q", oldName)
	}

	newCol := col.DeepCopy()
	newCol.sqlName = newName

	err := newCol.Validate()
	if err != nil {
		return err
	}

	isPK := t.IsSQLNamePrimaryKey(col.SQLName())
	_ = t.RemoveColumnBySQLName(oldName)
	t.addColumn(newCol, isPK)
	return nil
}

// SetParent sets the provided table as this table's parent.
func (t *Table) SetParent(parent *Table) error {
	if t.parent != nil {
		return fmt.Errorf("table already has a parent")
	}
	t.parent = parent
	return nil
}

// SQLName returns this table's SQLName.
func (t *Table) SQLName() string {
	return t.sqlName
}

// uniqueColumnName returns a version of the provided SQLName that is unique
// within this table.
func (t *Table) uniqueColumnName(columnName string) string {
	retColumnName := columnName
	i := 0
	for {
		col := t.Column(retColumnName)
		if col != nil {
			retColumnName = fmt.Sprintf("%v_%v", columnName, i)
			i++
			continue
		}
		return retColumnName
	}
}

// UnwindPath returns the table unwindPath.
func (t *Table) UnwindPath() string {
	return t.unwindPath
}

// Validate checks whether this Table is valid, returning an error if not.
func (t *Table) Validate() error {
	if strings.Trim(t.sqlName, " ") == "" {
		return fmt.Errorf("invalid SQLName %q", t.sqlName)
	}

	haveMongoFilter := false

	cmap := make(map[string]struct{})

	for _, c := range t.Columns() {
		err := c.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate column '%s': %v", c.MongoName(), err)
		}

		if c.MongoType() == MongoFilter {
			if haveMongoFilter {
				return fmt.Errorf("cannot have more than one mongo filter")
			}
			haveMongoFilter = true
		}

		key := strings.ToLower(c.SQLName())
		if _, ok := cmap[key]; ok {
			return fmt.Errorf("duplicate SQL column: '%s'", c.SQLName())
		}
		cmap[key] = struct{}{}
	}

	return nil
}
