package catalog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/astprint"
	"github.com/10gen/mongoast/parser"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
)

// NewMongoTable creates a new MongoTable.
func NewMongoTable(databaseName string, t *schema.Table, tblType string,
	collation *collation.Collation, writeMode bool) *MongoTable {
	var columns results.Columns
	columnMap := make(map[string]*results.Column)
	var primaryKeys results.Columns
	var colRange []*schema.Column
	// In writeMode, we preserve the column order specified by table creation.
	if writeMode {
		colRange = t.ColumnsDeclaredOrder()
	} else {
		colRange = t.ColumnsSorted()
	}
	for i, c := range colRange {
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

// MarshalBSON is a custom implementation for marshalling a MongoTable into
// raw BSON bytes. MongoTable deviates from the default BSON marshalling
// implementation by marshalling the `pipeline` field as a JSON string
// instead of BSON arrays and documents. This is necessary in order to
// store the table in MongoDB, since $-prefixed keys are not allowed. This
// function also omits some fields, such as columnMap, and simplifies some
// fields such as primaryKeys and the nested columns fields of indexes and
// foreignKeys.
func (t *MongoTable) MarshalBSON() ([]byte, error) {
	var indexes []marshalableIndex
	if t.indexes != nil {
		indexes = make([]marshalableIndex, len(t.indexes))
		for i, index := range t.indexes {
			indexes[i] = marshalableIndex{
				Columns:        index.columns.Names(),
				Unique:         index.unique,
				FullText:       index.fullText,
				ConstraintName: index.constraintName,
			}
		}
	}

	var foreignKeys []marshalableForeignKey
	if t.foreignKeys != nil {
		foreignKeys = make([]marshalableForeignKey, len(t.foreignKeys))
		for i, fk := range t.foreignKeys {
			foreignKeys[i] = marshalableForeignKey{
				Columns:              fk.columns.Names(),
				ConstraintName:       fk.constraintName,
				ForeignDatabase:      fk.foreignDatabase,
				ForeignTable:         fk.foreignTable,
				LocalToForeignColumn: fk.localToForeignColumn,
			}
		}
	}

	var primaryKeyNames []string
	if t.primaryKeys != nil {
		primaryKeyNames = t.primaryKeys.Names()
	}

	mt := marshalableMongoTable{
		CollectionName: t.collectionName,
		Name:           t.name,
		Collation:      string(t.collation.Name),
		Columns:        t.columns,
		PrimaryKeys:    primaryKeyNames,
		Indexes:        indexes,
		ForeignKeys:    foreignKeys,
		Comments:       t.comments,
		TableType:      t.tableType,
		Pipeline:       astprint.String(t.pipeline),
	}

	return bson.Marshal(&mt)
}

// UnmarshalBSON unmarshals the provided raw bytes into the MongoTable.
func (t *MongoTable) UnmarshalBSON(b []byte) error {
	mt := marshalableMongoTable{}
	err := bson.Unmarshal(b, &mt)
	if err != nil {
		return fmt.Errorf("failed to unmarshal MongoTable: %v", err)
	}

	// recover the collation
	mtCollation, err := collation.Get(collation.Name(mt.Collation))
	if err != nil {
		return fmt.Errorf("failed to unmarshal MongoTable: invalid collation name %q: %v", mt.Collation, err)
	}

	// recover the columnMap
	columnMap := make(map[string]*results.Column, len(mt.Columns))
	for _, col := range mt.Columns {
		// recover information not serialized by results.Column
		col.Table = mt.Name
		col.OriginalName = col.Name
		col.MappingRegistryName = col.Name

		// store in columnMap
		columnMap[col.Name] = col
	}

	// recover the primary keys
	var pks results.Columns
	if mt.PrimaryKeys != nil {
		pks = make(results.Columns, len(mt.PrimaryKeys))
		for i, pkName := range mt.PrimaryKeys {
			if col, ok := columnMap[pkName]; ok {
				pks[i] = col
			} else {
				return fmt.Errorf("failed to unmarshal MongoTable: unknown column %q for primary key", pkName)
			}
		}
	}

	// recover the columns for each index
	var indexes []Index
	if mt.Indexes != nil {
		indexes = make([]Index, len(mt.Indexes))
		for i, index := range mt.Indexes {
			columns := make(results.Columns, len(index.Columns))
			for j, colName := range index.Columns {
				if col, ok := columnMap[colName]; ok {
					columns[j] = col
				} else {
					return fmt.Errorf("failed to unmarshal MongoTable: unknown column %q for index", colName)
				}
			}

			indexes[i] = Index{
				columns:        columns,
				unique:         index.Unique,
				fullText:       index.FullText,
				constraintName: index.ConstraintName,
			}
		}
	}

	// recover the columns for each foreign key
	var fks []ForeignKey
	if mt.ForeignKeys != nil {
		fks = make([]ForeignKey, len(mt.ForeignKeys))
		for i, fk := range mt.ForeignKeys {
			columns := make(results.Columns, len(fk.Columns))
			for j, colName := range fk.Columns {
				if col, ok := columnMap[colName]; ok {
					columns[j] = col
				} else {
					return fmt.Errorf("failed to unmarshal MongoTable: unknown column %q for foreign key", colName)
				}
			}

			fks[i] = ForeignKey{
				columns:              columns,
				constraintName:       fk.ConstraintName,
				foreignDatabase:      fk.ForeignDatabase,
				foreignTable:         fk.ForeignTable,
				localToForeignColumn: fk.LocalToForeignColumn,
			}
		}
	}

	// recover pipeline
	pipeline, err := parser.ParsePipelineJSON(mt.Pipeline)
	if err != nil {
		return fmt.Errorf("failed to unmarshal MongoTable: failed to parse pipeline: %v", err)
	}

	// set the fields of the table
	t.name = mt.Name
	t.collation = mtCollation
	t.columns = mt.Columns
	t.columnMap = columnMap
	t.primaryKeys = pks
	t.indexes = indexes
	t.foreignKeys = fks
	t.comments = mt.Comments
	t.tableType = mt.TableType
	t.isSharded = true // default isSharded to true
	t.collectionName = mt.CollectionName
	t.pipeline = pipeline

	return nil
}

type marshalableIndex struct {
	Columns        []string `bson:"columns"`
	Unique         bool     `bson:"unique"`
	FullText       bool     `bson:"fullText"`
	ConstraintName string   `bson:"constraintName"`
}

type marshalableForeignKey struct {
	Columns              []string          `bson:"columns"`
	ConstraintName       string            `bson:"constraintName"`
	ForeignDatabase      string            `bson:"foreignDatabase"`
	ForeignTable         string            `bson:"foreignTable"`
	LocalToForeignColumn map[string]string `bson:"localToForeignColumn"`
}

type marshalableMongoTable struct {
	CollectionName string                  `bson:"collectionName"`
	Name           string                  `bson:"tableName"`
	Collation      string                  `bson:"collation"`
	Columns        results.Columns         `bson:"columns"`
	PrimaryKeys    []string                `bson:"primaryKeys"`
	Indexes        []marshalableIndex      `bson:"indexes"`
	ForeignKeys    []marshalableForeignKey `bson:"foreignKeys"`
	Comments       string                  `bson:"comments"`
	TableType      string                  `bson:"tableType"`
	Pipeline       string                  `bson:"pipeline"`
}
