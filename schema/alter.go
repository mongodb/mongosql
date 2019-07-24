package schema

import (
	"fmt"
	"strings"
)

// AlterationType is an enum that represents the type of change being made by a
// particular alteration (RenameColumn, DropColumn, or RenameTable).
type AlterationType string

// Possible values for AlterationType
const (
	RenameColumn AlterationType = "rename_column"
	DropColumn   AlterationType = "drop_column"
	ModifyColumn AlterationType = "modify_column"
	RenameTable  AlterationType = "rename_table"
)

// An Alteration specifies a change to be made to the schema of a SQL table.
type Alteration struct {
	Type          AlterationType `bson:"type"`
	Db            string         `bson:"db"`
	Table         string         `bson:"table"`
	NewTable      string         `bson:"new_table,omitempty"`
	Column        string         `bson:"column,omitempty"`
	NewColumn     string         `bson:"new_column,omitempty"`
	NewColumnType string         `bson:"new_column_type,omitempty"`
}

// DeepCopy returns a deep copy of this Alteration.
func (a *Alteration) DeepCopy() *Alteration {
	return &Alteration{
		Type:          a.Type,
		Db:            a.Db,
		Table:         a.Table,
		NewTable:      a.NewTable,
		Column:        a.Column,
		NewColumn:     a.NewColumn,
		NewColumnType: a.NewColumnType,
	}
}

func (a *Alteration) String() string {
	switch a.Type {
	case RenameColumn:
		return fmt.Sprintf("rename column \"%s.%s\" to \"%s.%s\"", a.Table, a.Column, a.Table,
			a.NewColumn)
	case ModifyColumn:
		return fmt.Sprintf("modify column \"%s.%s\" type %s", a.Table, a.Column, a.NewColumnType)
	case DropColumn:
		return fmt.Sprintf("drop column \"%s.%s\"", a.Table, a.Column)
	case RenameTable:
		return fmt.Sprintf("rename table %q to %q", a.Table, a.NewTable)
	}
	return "<unsupported alteration type>"
}

func (a *Alteration) alter(schema *Schema) error {
	db := schema.Database(a.Db)
	if db == nil {
		return fmt.Errorf("could not find database %q", a.Db)
	}

	table := db.Table(a.Table)
	if table == nil {
		return fmt.Errorf("could not find table %q in database %q", a.Table, a.Db)
	}

	switch a.Type {
	case RenameColumn:
		return table.RenameColumn(a.Column, a.NewColumn)
	case ModifyColumn:
		sqlType, err := getSQLTypeFromColumnType(a.NewColumnType)
		if err != nil {
			return fmt.Errorf("could not modify column %s: %v", a.Column, err)
		}
		mongoType := GetMongoTypeFromSQLType(sqlType)
		return table.ChangeColumnType(a.Column, sqlType, mongoType)
	case DropColumn:
		if len(table.Columns()) == 1 {
			return fmt.Errorf("cannot remove last column from table %q", a.Table)
		}
		if strings.Split(a.Column, ".")[0] == "_id" {
			return fmt.Errorf("cannot drop column %s: not allowed", a.Column)
		}
		return table.RemoveColumnBySQLName(a.Column)
	case RenameTable:
		return db.RenameTable(a.Table, a.NewTable)
	default:
		return fmt.Errorf("unknown alteration type %q", a.Type)
	}
}
