package catalog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// NewMongoTable creates a new MongoTable.
func NewMongoTable(t *schema.Table, tblType TableType, collation *collation.Collation) *MongoTable {
	var columns []*MongoColumn
	columnMap := make(map[string]*MongoColumn)
	var primaryKeys []Column
	for _, c := range t.ColumnsSorted() {
		types := c.SampledTypes()
		hasNull := false
		for i := range types {
			if types[i] == "" {
				types[i] = "null"
				hasNull = true
			}
		}
		isPolymorphic := false
		// NULLs do not count for being polymorphic from a SQL perspective,
		// as NULL is a value in all types, rather than a type as in MongoDB.
		// So if the number of sampled types is greater than two, or it is
		// exactly two and neither of them is NULL, this needs to be treated
		// as polymorphic.
		if len(types) > 2 || (len(types) == 2 && !hasNull) {
			isPolymorphic = true
		}
		sort.Strings(types)

		var commentStr string
		if len(types) == 0 {
			commentStr = fmt.Sprintf(`{ "name": "%s" }`,
				c.MongoName())
		} else {
			commentStr = fmt.Sprintf(`{ "name": "%s", "sampledTypes": ["%v"] }`,
				c.MongoName(), strings.Join(types, `", "`))
		}
		mc := &MongoColumn{
			name:           ColumnName(c.SQLName()),
			sqlType:        SQLType(c.SQLType()),
			MongoName:      c.MongoName(),
			MongoType:      c.MongoType(),
			comments:       commentStr,
			isPolymorphic:  isPolymorphic,
			hasAlteredType: c.HasTypeAlteration(),
		}
		columns = append(columns, mc)
		columnMap[strings.ToLower(c.SQLName())] = mc
		if t.IsMongoNamePrimaryKey(mc.MongoName) {
			primaryKeys = append(primaryKeys, mc)
		}
	}

	var comment string
	if t.UnwindPath() != "" {
		comment = fmt.Sprintf(`{ "collectionName": "%s", "unwoundFrom": "%s" }`,
			t.MongoName(), t.UnwindPath())
	} else {
		comment = fmt.Sprintf(`{ "collectionName": "%s" }`, t.MongoName())
	}
	return &MongoTable{
		name:           TableName(t.SQLName()),
		collation:      collation,
		columns:        columns,
		columnMap:      columnMap,
		tableType:      tblType,
		primaryKeys:    primaryKeys,
		collectionName: t.MongoName(),
		pipeline:       t.Pipeline(),
		comments:       comment,
	}
}

// MongoTable is a table whose data comes from a MongoDB collection.
type MongoTable struct {
	name           TableName
	collation      *collation.Collation
	columns        []*MongoColumn
	columnMap      map[string]*MongoColumn
	primaryKeys    []Column
	indexes        []Index
	foreignKeys    []ForeignKey
	comments       string
	tableType      TableType
	isSharded      bool
	collectionName string
	pipeline       []bson.D
}

// Name is the name of the MongoTable, t.
func (t *MongoTable) Name() TableName {
	return t.name
}

// IsSharded returns true if the MongoTable,
// t is in a sharded collection.
func (t *MongoTable) IsSharded() bool {
	return t.isSharded
}

// Collation gets the collation for the MongoTable, t.
func (t *MongoTable) Collation() *collation.Collation {
	return t.collation
}

// Column gets the column of the specified name.
func (t *MongoTable) Column(name string) (Column, error) {
	if c, ok := t.columnMap[strings.ToLower(name)]; ok {
		return c, nil
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, name, string(t.Name()))
}

// Columns returns the columns in MongoTable, t.
func (t *MongoTable) Columns() []Column {
	var cols []Column
	for _, c := range t.columns {
		cols = append(cols, c)
	}
	return cols
}

// Comments are the comments for the MongoTable, t.
func (t *MongoTable) Comments() string {
	return t.comments
}

// ForeignKeys returns the foreign keys for the MongoTable, t.
func (t *MongoTable) ForeignKeys() []ForeignKey {
	return t.foreignKeys
}

// Indexes returns the indexes for the MongoTable, t.
func (t *MongoTable) Indexes() []Index {
	return t.indexes
}

// IsMongoTable return true if this is a table from MongoDB.
func (t *MongoTable) IsMongoTable() bool {
	return true
}

// MongoName is the name of the collection underlying MongoTable, t.
func (t *MongoTable) MongoName() string {
	return t.collectionName
}

// Pipeline returns the BSON pipeline to be prepended for this table.
func (t *MongoTable) Pipeline() []bson.D {
	return t.pipeline
}

// PrimaryKeys returns the primary keys for
// the MongoTable, t.
func (t *MongoTable) PrimaryKeys() []Column {
	return t.primaryKeys
}

// Type returns the type of the MongoTable, t.
func (t *MongoTable) Type() TableType {
	return t.tableType
}

// MongoColumn is an mongo table column.
type MongoColumn struct {
	comments       string
	isPolymorphic  bool
	hasAlteredType bool

	name    ColumnName
	sqlType SQLType

	MongoName string
	MongoType schema.MongoType
}

// ShouldConvert returns true if `c` is a column must be wrapped in a convert expression.
// That means that either `c` contained polymorphic data types during sampling and the
// PolymorphicConversionMode is "PolymorphicConversionTypeModeFast", or simply if the
// PolymorphicConversionMode is "PolymorphicConversionModeSafe". "PolymorphicTypeConversionModeOff"
// always returns false.
func (c *MongoColumn) ShouldConvert(mode string) bool {
	if variable.PolymorphicTypeConversionModeType(mode) == variable.PolymorphicTypeConversionModeOff {
		return false
	}
	if variable.PolymorphicTypeConversionModeType(mode) == variable.PolymorphicTypeConversionModeSafe {
		return true
	}
	// In fast mode, we only want to introduce converts when we think they are
	// necessary to avoid query aggregation failures. Two places we know that
	// can definitely introduce aggregation failures are fields that were
	// sampled as polymorphic, and fields that have had their type altered with
	// the ALTER statement.
	return c.isPolymorphic || c.hasAlteredType
}

// Name returns the name of the column.
func (c *MongoColumn) Name() ColumnName {
	return c.name
}

// Comments returns the comments for the column.
func (c *MongoColumn) Comments() string {
	return c.comments
}

// Type returns the SQLType of the column, c.
func (c *MongoColumn) Type() SQLType {
	return c.sqlType
}
