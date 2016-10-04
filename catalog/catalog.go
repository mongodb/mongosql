package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// CatalogName is the name of a catalog.
type CatalogName string

// Catalog holds databases.
type Catalog struct {
	// Name is the name of the catalog.
	Name CatalogName

	databases []*Database
}

// New creates a new Catalog.
func New(name string) *Catalog {
	return &Catalog{
		Name:      CatalogName(name),
		databases: []*Database{},
	}
}

// AddDatabase adds the database to the Catalog.
func (c *Catalog) AddDatabase(name string) (*Database, error) {
	for _, d := range c.databases {
		if strings.ToLower(name) == strings.ToLower(string(d.Name)) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_DB_CREATE_EXISTS, name)
		}
	}

	d := &Database{
		Name:   DatabaseName(name),
		tables: []Table{},
	}

	c.databases = append(c.databases, d)
	return d, nil
}

// Database gets the Database with the specified name.
func (c *Catalog) Database(name string) (*Database, error) {
	for _, db := range c.databases {
		if strings.ToLower(name) == strings.ToLower(string(db.Name)) {
			return db, nil
		}
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_DB_ERROR, name)
}

// Databases gets all the databases in the Catalog.
func (c *Catalog) Databases() []*Database {
	return c.databases
}

// DatabaseName is the name of a database.
type DatabaseName string

// Database is a container for tables.
type Database struct {
	tables []Table

	// Name is the name of the database
	Name DatabaseName
}

// AddTable adds the table to the database.
func (d *Database) AddTable(t Table) error {
	if _, err := d.Table(string(t.Name())); err == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ER_TABLE_EXISTS_ERROR, t.Name())
	}

	d.tables = append(d.tables, t)
	return nil
}

// Table gets a Table from the Database.
func (d *Database) Table(name string) (Table, error) {
	for _, t := range d.tables {
		if strings.ToLower(name) == strings.ToLower(string(t.Name())) {
			return t, nil
		}
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_TABLE, string(d.Name), name)
}

// Tables gets the tables in the Database.
func (d *Database) Tables() []Table {
	return d.tables
}

// TableName is the name of a table.
type TableName string

// TableType is the type of a table.
type TableType string

// TableType constants.
const (
	BaseTable  TableType = "BASE TABLE"
	SystemView TableType = "SYSTEM VIEW"
	View       TableType = "VIEW"
)

// Table is the definition of a table.
type Table interface {
	// Name gets the name of the table.
	Name() TableName
	// Collation gets the collection for the table.
	Collation() *collation.Collation
	// Column gets the column of the specified name.
	Column(string) (Column, error)
	// Columns gets the columns for the table.
	Columns() []Column
	// PrimaryKeys returns the primary keys
	// for this table.
	PrimaryKeys() []Column
	// Comments gets the comments for the table.
	Comments() string
	// Type is the type of the table.
	Type() TableType
}

// ColumnName is the name of a column.
type ColumnName string

// Column is a column in a Table.
type Column interface {
	// Name gets the name of the column
	Name() ColumnName
	// Type is the type of the column.
	Type() schema.SQLType
	// Comments gets the comments for the column.
	Comments() string
}

type Columns []Column

// Contains returns true if the given name
// matches one of the columns in cols.
func (cols Columns) Contains(name ColumnName) bool {
	for _, c := range cols {
		if strings.ToLower(string(c.Name())) == strings.ToLower(string(name)) {
			return true
		}
	}
	return false
}
