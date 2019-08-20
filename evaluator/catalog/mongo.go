package catalog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// NewMongoTable creates a new MongoTable.
func NewMongoTable(databaseName string, t *schema.Table, tblType string,
	collation *collation.Collation, writeMode bool) *MongoTable {
	var columns results.Columns
	columnMap := make(map[string]*results.Column)
	var primaryKeys results.Columns
	for i, c := range t.ColumnsSorted() {
		tys := c.SampledTypes()
		hasNull := false
		for i := range tys {
			if tys[i] == "" {
				tys[i] = "null"
				hasNull = true
			}
		}
		isPolymorphic := false
		// NULLs do not count for being polymorphic from a SQL perspective,
		// as NULL is a value in all types, rather than a type as in MongoDB.
		// So if the number of sampled types is greater than two, or it is
		// exactly two and neither of them is NULL, this needs to be treated
		// as polymorphic.
		if len(tys) > 2 || (len(tys) == 2 && !hasNull) {
			isPolymorphic = true
		}
		sort.Strings(tys)

		// generate the defaultCommentStr that will be used if the user
		// did not specify a comment string.
		var defaultCommentStr string
		if writeMode {
			defaultCommentStr = ""
		} else {
			if len(tys) == 0 {
				defaultCommentStr = fmt.Sprintf(`{ "name": "%s" }`,
					c.MongoName())
			} else {
				defaultCommentStr = fmt.Sprintf(`{ "name": "%s", "sampledTypes": ["%v"] }`,
					c.MongoName(), strings.Join(tys, `", "`))
			}
		}
		isPrimaryKey := false
		if t.IsMongoNamePrimaryKey(c.MongoName()) {
			isPrimaryKey = true
		}

		cb := results.NewColumnBuilder()
		cb.SetColumnType(results.NewColumnType(types.SQLTypeToEvalType(c.SQLType()), c.MongoType()))
		cb.SetSelectID(i + 1)
		cb.SetTable(t.SQLName())
		cb.SetOriginalTable(t.MongoName())
		cb.SetDatabase(databaseName)
		cb.SetName(c.SQLName())
		cb.SetOriginalName(c.MongoName())
		cb.SetMappingRegistryName(c.SQLName())
		cb.SetMongoName(c.MongoName())
		cb.SetPrimaryKey(isPrimaryKey)
		cb.SetComments(c.Comment().Else(defaultCommentStr))
		cb.SetIsPolymorphic(isPolymorphic)
		cb.SetHasAlteredType(c.HasTypeAlteration())
		cb.SetNullable(c.Nullable())
		mc := cb.Build()
		if isPrimaryKey {
			primaryKeys = append(primaryKeys, mc)
		}
		columns = append(columns, mc)
		columnMap[strings.ToLower(c.SQLName())] = mc
	}

	var defaultComment string
	if writeMode {
		defaultComment = ""
	} else {
		if t.UnwindPath() != "" {
			defaultComment = fmt.Sprintf(`{ "collectionName": "%s", "unwoundFrom": "%s" }`,
				t.MongoName(), t.UnwindPath())
		} else {
			defaultComment = fmt.Sprintf(`{ "collectionName": "%s" }`, t.MongoName())
		}
	}

	indexes := make([]Index, len(t.Indexes()))
	for i, index := range t.Indexes() {
		cols := make(results.Columns, len(index.Parts()))
		for j, part := range index.Parts() {
			cols[j] = columnMap[part.SQLName()]
		}
		indexes[i] = Index{
			columns:        cols,
			unique:         index.Unique(),
			fullText:       index.FullText(),
			constraintName: index.SQLName(),
		}
	}

	pipeline, err := astutil.ParsePipeline(t.Pipeline())
	if err != nil {
		panic(fmt.Sprintf("failed to parse schema pipeline ([]bson.D) into evaluator pipeline (ast.Pipeline): %v", err))
	}

	return &MongoTable{
		name:           t.SQLName(),
		collation:      collation,
		columns:        columns,
		columnMap:      columnMap,
		tableType:      tblType,
		primaryKeys:    primaryKeys,
		collectionName: t.MongoName(),
		pipeline:       pipeline,
		indexes:        indexes,
		comments:       t.Comment().Else(defaultComment),
	}
}

// MongoTable is a table whose data comes from a MongoDB collection.
type MongoTable struct {
	name           string
	collation      *collation.Collation
	columns        results.Columns
	columnMap      map[string]*results.Column
	primaryKeys    results.Columns
	indexes        []Index
	foreignKeys    []ForeignKey
	comments       string
	tableType      string
	isSharded      bool
	collectionName string
	pipeline       *ast.Pipeline
}

// Name is the name of the MongoTable, t.
func (t *MongoTable) Name() string {
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
func (t *MongoTable) Column(name string) (*results.Column, error) {
	if c, ok := t.columnMap[strings.ToLower(name)]; ok {
		return c, nil
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, name, t.Name())
}

// Columns returns the columns in MongoTable, t.
func (t *MongoTable) Columns() results.Columns {
	var cols results.Columns
	cols = append(cols, t.columns...)
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

// Collection returns the name of the collection underlying MongoTable, t.
func (t *MongoTable) Collection() string {
	return t.collectionName
}

// Pipeline returns the BSON pipeline to be prepended for this table.
func (t *MongoTable) Pipeline() *ast.Pipeline {
	return t.pipeline
}

// PrimaryKeys returns the primary keys for
// the MongoTable, t.
func (t *MongoTable) PrimaryKeys() results.Columns {
	return t.primaryKeys
}

// Type returns the type of the MongoTable, t.
func (t *MongoTable) Type() string {
	return t.tableType
}
