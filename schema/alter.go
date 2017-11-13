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
	db, ok := schema.Database(a.Db)
	if !ok {
		return fmt.Errorf("could not find database %q", a.Db)
	}

	table, ok := db.Table(a.Table)
	if !ok {
		return fmt.Errorf("could not find table %q in database %q", a.Table, a.Db)
	}

	switch a.Type {
	case RenameColumn:
		column, ok := table.Column(a.Column)
		if !ok {
			return fmt.Errorf("could not find column %q.%q in database %q", a.Table, a.Column, a.Db)
		}
		column.SQLName = a.NewColumn
	case DropColumn:
		if len(table.Columns) == 1 {
			return fmt.Errorf("cannot remove last column from table %q", a.Table)
		}
		if strings.Split(a.Column, ".")[0] == "_id" {
			return fmt.Errorf("cannot drop column %s: not allowed", a.Column)
		}
		for i, col := range table.Columns {
			if col.SQLName == a.Column {
				table.Columns = append(table.Columns[:i], table.Columns[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("could not find column %q.%q in database %q", a.Table, a.Column, a.Db)
	case RenameTable:
		table.Name = a.NewTable
		return nil
	default:
		return fmt.Errorf("unknown alteration type %q", a.Type)
	}

	return nil
}
