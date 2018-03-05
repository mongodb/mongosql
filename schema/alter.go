package schema

import (
	"fmt"
	"strings"
	"time"
)

// AlterationType is an enum that represents the type of change being made by a
// particular alteration (RenameColumn, DropColumn, or RenameTable).
type AlterationType string

// Possible values for AlterationType
const (
	RenameColumn AlterationType = "rename_column"
	DropColumn   AlterationType = "drop_column"
	RenameTable  AlterationType = "rename_table"
)

// An Alteration specifies a change to be made to the schema of a SQL table.
type Alteration struct {
	Timestamp time.Time      `bson:"timestamp"`
	Type      AlterationType `bson:"type"`
	Db        string         `bson:"db"`
	Table     string         `bson:"table"`
	NewTable  string         `bson:"new_table,omitempty"`
	Column    string         `bson:"column,omitempty"`
	NewColumn string         `bson:"new_column,omitempty"`
}

// DeepCopy returns a deep copy of this Alteration.
func (a *Alteration) DeepCopy() *Alteration {
	return &Alteration{
		Timestamp: a.Timestamp,
		Type:      a.Type,
		Db:        a.Db,
		Table:     a.Table,
		NewTable:  a.NewTable,
		Column:    a.Column,
		NewColumn: a.NewColumn,
	}
}

func (a *Alteration) String() string {
	switch a.Type {
	case RenameColumn:
		return fmt.Sprintf("rename column \"%s.%s\" to \"%s.%s\"", a.Table, a.Column, a.Table, a.NewColumn)
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
		column := table.Column(a.Column)
		if column == nil {
			return fmt.Errorf("could not find column %q.%q in database %q", a.Table, a.Column, a.Db)
		}
		column.sqlName = a.NewColumn
	case DropColumn:
		if len(table.Columns()) == 1 {
			return fmt.Errorf("cannot remove last column from table %q", a.Table)
		}
		if strings.Split(a.Column, ".")[0] == "_id" {
			return fmt.Errorf("cannot drop column %s: not allowed", a.Column)
		}
		return table.RemoveColumnBySQLName(a.Column)
	case RenameTable:
		table.sqlName = a.NewTable
		return nil
	default:
		return fmt.Errorf("unknown alteration type %q", a.Type)
	}

	return nil
}
