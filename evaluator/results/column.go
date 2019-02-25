package results

import (
	"strings"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/schema"
)

// ColumnType is the type of a column.
type ColumnType struct {
	EvalType    types.EvalType
	MongoType   schema.MongoType
	UUIDSubType types.EvalType
}

// Column contains information used to select data
// from a PlanStage.
type Column struct {
	ColumnType
	SelectID            int
	Table               string
	OriginalTable       string
	Database            string
	Name                string
	OriginalName        string
	MappingRegistryName string
	MongoName           string
	PrimaryKey          bool
	Comments            string
	IsPolymorphic       bool
	HasAlteredType      bool
}

// NewColumn is a constructor for the Column struct.
func NewColumn(selectID int, table, originalTable, database, name,
	originalName, mappingRegistryName string, evalType types.EvalType,
	mongoType schema.MongoType, primaryKey bool) *Column {
	uuidSubType := types.EvalBinary
	if mongoType == schema.MongoUUIDJava {
		uuidSubType = types.EvalJavaUUID
	} else if mongoType == schema.MongoUUIDCSharp {
		uuidSubType = types.EvalCSharpUUID
	}
	return &Column{
		ColumnType: ColumnType{
			MongoType:   mongoType,
			EvalType:    evalType,
			UUIDSubType: uuidSubType,
		},
		SelectID:            selectID,
		Table:               table,
		OriginalTable:       originalTable,
		Database:            database,
		Name:                name,
		OriginalName:        originalName,
		MappingRegistryName: mappingRegistryName,
		PrimaryKey:          primaryKey,
	}
}

// NewColumnType returns a ColumnType with the specified types.EvalType and MongoType.
func NewColumnType(evalType types.EvalType, mongoType schema.MongoType) ColumnType {
	return ColumnType{
		EvalType:  evalType,
		MongoType: mongoType,
		// Because the need to set the UUIDSubType is so rare, we just use
		// the default EvalBinary encoding unless otherwise specified with the
		// other constructor.
		UUIDSubType: types.EvalBinary,
	}
}

// NewColumnTypeWithUUIDSubtype returns a ColumnType with the specified types.EvalType, MongoType, and
// UUIDSubType.
func NewColumnTypeWithUUIDSubtype(evalType types.EvalType,
	mongoType schema.MongoType,
	uuidSubType types.EvalType) ColumnType {
	return ColumnType{
		EvalType:    evalType,
		MongoType:   mongoType,
		UUIDSubType: uuidSubType,
	}
}

// Clone clones the Column.
func (c *Column) Clone() *Column {
	cb := NewColumnBuilder()
	cb.SetColumnType(NewColumnType(c.EvalType, c.MongoType))
	cb.SetSelectID(c.SelectID)
	cb.SetTable(c.Table)
	cb.SetOriginalTable(c.OriginalTable)
	cb.SetDatabase(c.Database)
	cb.SetName(c.Name)
	cb.SetOriginalName(c.OriginalName)
	cb.SetMappingRegistryName(c.MappingRegistryName)
	cb.SetMongoName(c.MongoName)
	cb.SetPrimaryKey(c.PrimaryKey)
	cb.SetComments(c.Comments)
	cb.SetIsPolymorphic(c.IsPolymorphic)
	cb.SetHasAlteredType(c.HasAlteredType)
	return cb.Build()
}

// Columns is a slice of Column pointers.
type Columns []*Column

// FindByName searches Columns for a column of a matching name.
func (cs Columns) FindByName(name string) (*Column, bool) {
	for _, c := range cs {
		if strings.EqualFold(name, c.Name) {
			return c, true
		}
	}

	return nil, false
}

// Unique ensures that only unique columns exist in the resulting slice.
func (cs Columns) Unique() Columns {
	var results Columns
	contains := func(column *Column) bool {
		for _, c := range results {
			if c.SelectID == column.SelectID &&
				c.Name == column.Name &&
				c.Table == column.Table &&
				c.Database == column.Database {
				return true
			}
		}

		return false
	}

	for _, c := range cs {
		if !contains(c) {
			results = append(results, c)
		}
	}

	return results
}
