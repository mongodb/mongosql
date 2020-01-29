package results

import (
	"fmt"
	"strings"
)

var columnFieldIndexToColumnName = [14]string{
	"ColumnType",
	"SelectID",
	"Table",
	"OriginalTable",
	"Database",
	"Name",
	"OriginalName",
	"MappingRegistryName",
	"MongoName",
	"PrimaryKey",
	"Comments",
	"IsPolymorphic",
	"HasAlteredType",
	"Null",
}

// ColumnBuilder builds Column objects.
type ColumnBuilder struct {
	column    *Column
	setFields [14]bool
}

// NewColumnBuilder builds a new ColumnBuilder.
func NewColumnBuilder() *ColumnBuilder {
	return &ColumnBuilder{column: &Column{}}
}

// Build attempts to build a Column.
func (cb *ColumnBuilder) Build() *Column {
	missingFields := make([]string, 0)
	for i, b := range cb.setFields {
		if !b {
			missingFields = append(missingFields, columnFieldIndexToColumnName[i])
		}
	}
	if len(missingFields) > 0 {
		panic(fmt.Sprintf("could not Build values.Column as the following fields have not been set: %s",
			strings.Join(missingFields, ", ")))
	}
	return cb.column
}

// SetColumnType sets the ColumnType for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetColumnType(ct *ColumnType) *ColumnBuilder {
	cb.column.ColumnType = ct
	cb.setFields[0] = true
	return cb
}

// SetSelectID sets the SelectID for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetSelectID(selectID int) *ColumnBuilder {
	cb.column.SelectID = selectID
	cb.setFields[1] = true
	return cb
}

// SetTable sets the Table for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetTable(table string) *ColumnBuilder {
	cb.column.Table = table
	cb.setFields[2] = true
	return cb
}

// SetOriginalTable sets the OriginalTable for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetOriginalTable(originalTable string) *ColumnBuilder {
	cb.column.OriginalTable = originalTable
	cb.setFields[3] = true
	return cb
}

// SetDatabase sets the Database for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetDatabase(database string) *ColumnBuilder {
	cb.column.Database = database
	cb.setFields[4] = true
	return cb
}

// SetName sets the Name for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetName(name string) *ColumnBuilder {
	cb.column.Name = name
	cb.setFields[5] = true
	return cb
}

// SetOriginalName sets the OriginalName for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetOriginalName(originalName string) *ColumnBuilder {
	cb.column.OriginalName = originalName
	cb.setFields[6] = true
	return cb
}

// SetMappingRegistryName sets the MappingRegistryName for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetMappingRegistryName(mappingRegistryName string) *ColumnBuilder {
	cb.column.MappingRegistryName = mappingRegistryName
	cb.setFields[7] = true
	return cb
}

// SetMongoName sets the MongoName for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetMongoName(mongoName string) *ColumnBuilder {
	cb.column.MongoName = mongoName
	cb.setFields[8] = true
	return cb
}

// SetPrimaryKey sets the PrimaryKey for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetPrimaryKey(primaryKey bool) *ColumnBuilder {
	cb.column.PrimaryKey = primaryKey
	cb.setFields[9] = true
	return cb
}

// SetComments sets the Comments for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetComments(comments string) *ColumnBuilder {
	cb.column.Comments = comments
	cb.setFields[10] = true
	return cb
}

// SetIsPolymorphic sets the IsPolymorphic for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetIsPolymorphic(isPolymorphic bool) *ColumnBuilder {
	cb.column.IsPolymorphic = isPolymorphic
	cb.setFields[11] = true
	return cb
}

// SetHasAlteredType sets the HasAlteredType for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetHasAlteredType(hasAlteredType bool) *ColumnBuilder {
	cb.column.HasAlteredType = hasAlteredType
	cb.setFields[12] = true
	return cb
}

// SetNullable sets the Null field for the Column in the ColumnBuilder
func (cb *ColumnBuilder) SetNullable(nullable bool) *ColumnBuilder {
	cb.column.Nullable = nullable
	cb.setFields[13] = true
	return cb
}
