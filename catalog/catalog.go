package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

const (
	// InformationSchemaDatabase is the name of the MySQL information schema database.
	InformationSchemaDatabase = "information_schema"
)

// Name is the name of a catalog.
type Name string

// Catalog holds databases.
type Catalog struct {
	// Name is the name of the catalog.
	Name Name
	// containsAuthRestrictedNamespaces is true if there are namespaces that have
	// been sampled that are not visible to the current user. This can be used
	// to authorize flush sample as the user should only have permission to
	// resample if they can see all namespaces in the sample. Even though the
	// resample is actually performed as the admin user, this gives us a way to
	// restrict resampling in standalone mode.
	containsAuthRestrictedNamespaces bool

	databases   []*Database
	databaseMap map[string]*Database
}

// HasAuthRestrictedNamespaces returns true if the user
// cannot view the entire sampled namespace due to
// privilege restrictions.
func (c *Catalog) HasAuthRestrictedNamespaces() bool {
	return c.containsAuthRestrictedNamespaces
}

// New creates a new Catalog.
func New(name string) *Catalog {
	return &Catalog{
		Name:        Name(name),
		databases:   []*Database{},
		databaseMap: make(map[string]*Database),
	}
}

// AddDatabase adds the database to the Catalog.
func (c *Catalog) AddDatabase(name string) (*Database, error) {

	lowerName := strings.ToLower(name)
	_, ok := c.databaseMap[lowerName]
	if ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErDbCreateExists, name)
	}

	d := &Database{
		Name:     DatabaseName(name),
		tables:   []Table{},
		tableMap: make(map[string]Table),
	}

	c.databases = append(c.databases, d)
	c.databaseMap[lowerName] = d
	return d, nil
}

// Database gets the Database with the specified name.
func (c *Catalog) Database(name string) (*Database, error) {
	if d, ok := c.databaseMap[strings.ToLower(name)]; ok {
		return d, nil
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadDbError, name)
}

// Databases gets all the databases in the Catalog.
func (c *Catalog) Databases() []*Database {
	return c.databases
}

// DatabaseName is the name of a database.
type DatabaseName string

// Database is a container for tables.
type Database struct {
	// Name is the name of the database
	Name DatabaseName

	tables   []Table
	tableMap map[string]Table
}

// AddTable adds the table to the database.
func (d *Database) AddTable(t Table) error {
	if _, err := d.Table(string(t.Name())); err == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ErTableExistsError, t.Name())
	}

	d.tables = append(d.tables, t)
	d.tableMap[strings.ToLower(string(t.Name()))] = t
	return nil
}

// Table gets a Table from the Database.
func (d *Database) Table(name string) (Table, error) {
	if t, ok := d.tableMap[strings.ToLower(name)]; ok {
		return t, nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoSuchTable, string(d.Name), name)
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

// ColumnName is the name of a column.
type ColumnName string

// Column is the interface that wraps a SQL column.
type Column interface {
	// Name gets the name of the column.
	Name() ColumnName
	// Type is the type of the column.
	Type() schema.SQLType
	// Comments gets the comments for the column.
	Comments() string
}

// Columns is a slice of `Column`s.
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

// ForeignKey represents a foreign key in a SQL table.
// This generally allows for cross-referencing related
// data across tables. In our case, it primarily
// allows us to (from a child table) reference data in
// a parent table.
type ForeignKey struct {
	columns              []Column
	constraintName       string
	foreignDatabase      string
	foreignTable         string
	localToForeignColumn map[string]string
}

// NewForeignKey returns a new foreign key for the column c.
func NewForeignKey(c Column, name, db, tb, col string) ForeignKey {
	return ForeignKey{
		columns:         []Column{c},
		constraintName:  name,
		foreignDatabase: db,
		foreignTable:    tb,
		localToForeignColumn: map[string]string{
			string(c.Name()): col,
		},
	}
}

// Table is an interface describing an SQL table.
type Table interface {
	// Collation gets the collection for the table.
	Collation() *collation.Collation
	// Column gets the columns of the specified name.
	Column(string) (Column, error)
	// Columns gets the columns for the table.
	Columns() []Column
	// Comments gets the comments for the table.
	Comments() string
	// ForeignKeys returns the foreign keys for this table.
	ForeignKeys() []ForeignKey
	// Indexes return the indexes for this table.
	Indexes() []Index
	// Name gets the name of the table.
	Name() TableName
	// PrimaryKeys returns the primary keys
	// for this table.
	PrimaryKeys() []Column
	// Type is the type of the table.
	Type() TableType
}
