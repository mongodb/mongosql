package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"

	"github.com/10gen/mongo-go-driver/bson"
)

const (
	// InformationSchemaDatabase is the name of the MySQL information schema database.
	InformationSchemaDatabase = "information_schema"
)

// Name is the name of a catalog.
type Name string

// Catalog is the interface that wraps methods for getting a SQL schema catalog.
type Catalog interface {
	// Databases returns all the databases in the catalog.
	Databases() []Database
	// Database returns the database associated with the given name (case-insensitive).
	Database(databaseName string) (Database, error)
	// HasAuthRestrictedNamespaces returns true if some namespaces require authentication for access
	// and false otherwise.
	HasAuthRestrictedNamespaces() bool
	// Variables returns an interface that wraps methods for getting variables' values.
	Variables() VariableContainer
}

// VariableContainer is the interface that wraps methods for accessing values of variables.
type VariableContainer interface {
	// Get returns the value of the variable, "name", of Kind "kind" in the given "scope".
	Get(name variable.Name, scope variable.Scope, kind variable.Kind) (variable.Value, error)
	// GetCollation gets the collation of the variable with the specified name.
	GetCollation(name variable.Name) *collation.Collation
	// GetInt64 returns the int64 value of the given system variable, "name".
	GetInt64(name variable.Name) int64
	// GetBool returns the bool value of the given system variable, "name".
	GetBool(name variable.Name) bool
	// GetString returns the string value of the given system variable, "name".
	GetString(name variable.Name) string
	// GetUint16 returns the uint64 value of the given system variable, "name".
	GetUint16(name variable.Name) uint16
	// GetUint64 returns the uint64 value of the given system variable, "name".
	GetUint64(name variable.Name) uint64
	// List returns the value of all variables of Kind "kind" in the given "scope".
	List(scope variable.Scope, kind variable.Kind) []variable.Value
}

// SQLCatalog holds databases.
type SQLCatalog struct {
	// Name is the name of the catalog.
	Name Name
	// containsAuthRestrictedNamespaces is true if there are namespaces that have
	// been sampled that are not visible to the current user. This can be used
	// to authorize flush sample as the user should only have permission to
	// resample if they can see all namespaces in the sample. Even though the
	// resample is actually performed as the admin user, this gives us a way to
	// restrict resampling in standalone mode.
	containsAuthRestrictedNamespaces bool

	// variables is a container of valid variables and their values, which
	// can be used to validate variable references and insert values during
	// algebrization.
	variables VariableContainer

	databases   []Database
	databaseMap map[string]Database
}

// New creates a new Catalog.
func New(name string, vars VariableContainer) *SQLCatalog {
	return &SQLCatalog{
		Name:        Name(name),
		databases:   []Database{},
		databaseMap: make(map[string]Database),
		variables:   vars,
	}
}

// HasAuthRestrictedNamespaces returns true if the user
// cannot view the entire sampled namespace due to
// privilege restrictions.
func (c *SQLCatalog) HasAuthRestrictedNamespaces() bool {
	return c.containsAuthRestrictedNamespaces
}

// AddDatabase adds the database to the Catalog.
func (c *SQLCatalog) AddDatabase(name string) (Database, error) {
	lowerName := strings.ToLower(name)
	_, ok := c.databaseMap[lowerName]
	if ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErDbCreateExists, name)
	}

	d := &SQLDatabase{
		name:     DatabaseName(name),
		tables:   []Table{},
		tableMap: make(map[string]Table),
	}

	c.databases = append(c.databases, d)
	c.databaseMap[lowerName] = d
	return d, nil
}

// Database gets the Database with the specified name.
func (c *SQLCatalog) Database(name string) (Database, error) {
	if d, ok := c.databaseMap[strings.ToLower(name)]; ok {
		return d, nil
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadDbError, name)
}

// Databases gets all the databases in the Catalog.
func (c *SQLCatalog) Databases() []Database {
	return c.databases
}

// Variables returns the variable.Container from the Catalog.
func (c *SQLCatalog) Variables() VariableContainer {
	return c.variables
}

// Database is an interface describing an SQL database.
type Database interface {
	// Name gets the name of the database.
	Name() DatabaseName
	// Tables gets the columns for the database.
	Tables() []Table
	// Add a table to the database.
	AddTable(t Table) error
	// Lookup a table with the given name.
	Table(name string) (Table, error)
}

// DatabaseName is the name of a database.
type DatabaseName string

// SQLDatabase is a container for tables.
type SQLDatabase struct {
	name     DatabaseName
	tables   []Table
	tableMap map[string]Table
}

// AddTable adds the table to the database.
func (d *SQLDatabase) AddTable(t Table) error {
	if _, err := d.Table(string(t.Name())); err == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ErTableExistsError, t.Name())
	}

	d.tables = append(d.tables, t)
	d.tableMap[strings.ToLower(string(t.Name()))] = t
	return nil
}

// Name gets the name of the Database.
func (d *SQLDatabase) Name() DatabaseName {
	return d.name
}

// Table gets a Table from the Database.
func (d *SQLDatabase) Table(name string) (Table, error) {
	if t, ok := d.tableMap[strings.ToLower(name)]; ok {
		return t, nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoSuchTable, string(d.name), name)
}

// Tables gets the tables in the Database.
func (d *SQLDatabase) Tables() []Table {
	return d.tables
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

type MongoDBTable interface {
	Table
	// Collection returns the name of the underlying MongoDB collection.
	Collection() string
	// IsSharded returns true if is a MongoDB table that is sharded and false otherwise.
	IsSharded() bool
	// Pipeline returns the BSON pipeline to be prepended for this table.
	Pipeline() []bson.D
}

// TableType is the type of a table.
type TableType string

// TableType constants.
const (
	BaseTable  TableType = "BASE TABLE"
	SystemView TableType = "SYSTEM VIEW"
	View       TableType = "VIEW"
)

// TableName is the name of a table.
type TableName string

const (
	CharacterSetsTable                      TableName = "CHARACTER_SETS"
	CollationCharacterSetApplicabilityTable TableName = "COLLATION_CHARACTER_SET_APPLICABILITY"
	CollationsTable                         TableName = "COLLATIONS"
	ColumnPrivilegesTable                   TableName = "COLUMN_PRIVILEGES"
	ColumnsTable                            TableName = "COLUMNS"
	EnginesTable                            TableName = "ENGINES"
	EventsTable                             TableName = "EVENTS"
	GlobalStatusTable                       TableName = "GLOBAL_STATUS"
	GlobalVariablesTable                    TableName = "GLOBAL_VARIABLES"
	KeyColumnUsageTable                     TableName = "KEY_COLUMN_USAGE"
	NdbTransidMysqlConnectionMapTable       TableName = "ndb_transid_mysql_connection_map"
	PluginsTable                            TableName = "PLUGINS"
	ParametersTable                         TableName = "PARAMETERS"
	PartitionsTable                         TableName = "PARTITIONS"
	ProfilingTable                          TableName = "PROFILING"
	ReferentialConstraintsTable             TableName = "REFERENTIAL_CONSTRAINTS"
	RoutinesTable                           TableName = "ROUTINES"
	SchemaPrivilagesTable                   TableName = "SCHEMA_PRIVILEGES"
	SchemataTable                           TableName = "SCHEMATA"
	SessionStatusTable                      TableName = "SESSION_STATUS"
	SessionVariablesTable                   TableName = "SESSION_VARIABLES"
	StatisticsTable                         TableName = "STATISTICS"
	TableConstraintsTable                   TableName = "TABLE_CONSTRAINTS"
	TablePrivilagesTable                    TableName = "TABLE_PRIVILEGES"
	TablespacesTable                        TableName = "TABLESPACES"
	TablesTable                             TableName = "TABLES"
	TriggersTable                           TableName = "TRIGGERS"
	UserPrivilagesTable                     TableName = "USER_PRIVILEGES"
)

// ColumnName is the name of a column.
type ColumnName string

// SQLType is human readable string representation of SQL types.
type SQLType string

// Column is the interface that wraps a SQL column.
type Column interface {
	// Name gets the name of the column.
	Name() ColumnName
	// Type is the type of the column.
	Type() SQLType
	// Comments gets the comments for the column.
	Comments() string
	// ShouldConvert returns true if this Column should
	// be wrapped in a SQLConvertExpr when referenced.
	// This occurs when there are sampled types that differ
	// from the consensus type of the column for MongoColumns.
	// It is always false for other types of Columns.
	ShouldConvert(polymorphicTypeConversionMode string) bool
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
