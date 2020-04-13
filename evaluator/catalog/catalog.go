package catalog

import (
	"context"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/strutil"
)

const (
	// InformationSchemaDatabase is the name of the MySQL information schema database.
	InformationSchemaDatabase = "INFORMATION_SCHEMA"
)

// Name is the name of a catalog.
type Name string

// Catalog is the interface that wraps methods for getting a SQL schema catalog.
type Catalog interface {
	// Databases returns all the databases in the catalog.
	Databases(ctx context.Context) ([]Database, error)
	// Database returns the database associated with the given name (case-insensitive).
	Database(ctx context.Context, databaseName string) (Database, error)
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

	databases   []Database
	databaseMap map[string]Database

	isCaseSensitive bool
}

// New creates a new Catalog.
func New(name string, isCaseSensitive bool) *SQLCatalog {
	return &SQLCatalog{
		Name:            Name(name),
		databases:       []Database{},
		databaseMap:     make(map[string]Database),
		isCaseSensitive: isCaseSensitive,
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
	lookupName := strutil.MaybeToLower(name, c.isCaseSensitive)
	if _, ok := c.databaseMap[lookupName]; ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErDbCreateExists, name)
	}

	d := &SQLDatabase{
		name:            DatabaseName(name),
		tables:          []Table{},
		tableMap:        make(map[string]Table),
		isCaseSensitive: c.isCaseSensitive,
	}

	c.databases = append(c.databases, d)
	c.databaseMap[lookupName] = d
	return d, nil
}

// Database gets the Database with the specified name.
func (c *SQLCatalog) Database(_ context.Context, name string) (Database, error) {
	if d, ok := c.databaseMap[strutil.MaybeToLower(name, c.isCaseSensitive)]; ok {
		return d, nil
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadDbError, name)
}

// Databases gets all the databases in the Catalog.
func (c *SQLCatalog) Databases(_ context.Context) ([]Database, error) {
	return c.databases, nil
}

// Database is an interface describing an SQL database.
type Database interface {
	// Name gets the name of the database.
	Name() DatabaseName
	// Tables gets the columns for the database.
	Tables(ctx context.Context) ([]Table, error)
	// Add a table to the database.
	AddTable(t Table) error
	// Lookup a table with the given name.
	Table(ctx context.Context, name string) (Table, error)
}

// DatabaseName is the name of a database.
type DatabaseName string

// SQLDatabase is a container for tables.
type SQLDatabase struct {
	name     DatabaseName
	tables   []Table
	tableMap map[string]Table

	isCaseSensitive bool
}

// AddTable adds the table to the database.
func (d *SQLDatabase) AddTable(t Table) error {
	if _, err := d.Table(context.Background(), t.Name()); err == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ErTableExistsError, t.Name())
	}

	d.tables = append(d.tables, t)
	d.tableMap[strutil.MaybeToLower(t.Name(), d.isCaseSensitive)] = t
	return nil
}

// Name gets the name of the Database.
func (d *SQLDatabase) Name() DatabaseName {
	return d.name
}

// Table gets a Table from the Database.
func (d *SQLDatabase) Table(_ context.Context, name string) (Table, error) {
	if t, ok := d.tableMap[strutil.MaybeToLower(name, d.isCaseSensitive)]; ok {
		return t, nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoSuchTable, string(d.name), name)
}

// Tables gets the tables in the Database.
func (d *SQLDatabase) Tables(_ context.Context) ([]Table, error) {
	return d.tables, nil
}

// Table is an interface describing an SQL table.
type Table interface {
	// Collation gets the collection for the table.
	Collation() *collation.Collation
	// Column gets the columns of the specified name.
	Column(string) (*results.Column, error)
	// Columns gets the columns for the table.
	Columns() results.Columns
	// Comments gets the comments for the table.
	Comments() string
	// ForeignKeys returns the foreign keys for this table.
	ForeignKeys() []ForeignKey
	// Indexes return the indexes for this table.
	Indexes() []Index
	// Name gets the name of the table.
	Name() string
	// PrimaryKeys returns the primary keys
	// for this table.
	PrimaryKeys() results.Columns
	// Type is the type of the table.
	Type() string
}

// MongoDBTable represents a table that exists as a MongoDB collection.
type MongoDBTable interface {
	Table
	// Collection returns the name of the underlying MongoDB collection.
	Collection() string
	// IsSharded returns true if this MongoDB table is sharded and false otherwise.
	IsSharded() bool
	// Pipeline returns the BSON pipeline to be prepended for this table.
	Pipeline() *ast.Pipeline
}

// Table type constants
const (
	BaseTable  = "BASE TABLE"
	SystemView = "SYSTEM VIEW"
	View       = "VIEW"
)

// information_schema databases table names.
const (
	CharacterSetsTable                      = "CHARACTER_SETS"
	CollationCharacterSetApplicabilityTable = "COLLATION_CHARACTER_SET_APPLICABILITY"
	CollationsTable                         = "COLLATIONS"
	ColumnPrivilegesTable                   = "COLUMN_PRIVILEGES"
	ColumnsTable                            = "COLUMNS"
	EnginesTable                            = "ENGINES"
	EventsTable                             = "EVENTS"
	GlobalStatusTable                       = "GLOBAL_STATUS"
	GlobalVariablesTable                    = "GLOBAL_VARIABLES"
	KeyColumnUsageTable                     = "KEY_COLUMN_USAGE"
	NdbTransidMysqlConnectionMapTable       = "ndb_transid_mysql_connection_map"
	PluginsTable                            = "PLUGINS"
	ParametersTable                         = "PARAMETERS"
	PartitionsTable                         = "PARTITIONS"
	ProfilingTable                          = "PROFILING"
	ReferentialConstraintsTable             = "REFERENTIAL_CONSTRAINTS"
	RoutinesTable                           = "ROUTINES"
	SchemaPrivilagesTable                   = "SCHEMA_PRIVILEGES"
	SchemataTable                           = "SCHEMATA"
	SessionStatusTable                      = "SESSION_STATUS"
	SessionVariablesTable                   = "SESSION_VARIABLES"
	StatisticsTable                         = "STATISTICS"
	TableConstraintsTable                   = "TABLE_CONSTRAINTS"
	TablePrivilagesTable                    = "TABLE_PRIVILEGES"
	TablespacesTable                        = "TABLESPACES"
	TablesTable                             = "TABLES"
	TriggersTable                           = "TRIGGERS"
	UserPrivilegesTable                     = "USER_PRIVILEGES"
)

// ForeignKey represents a foreign key in a SQL table.
// This generally allows for cross-referencing related
// data across tables. In our case, it primarily
// allows us to (from a child table) reference data in
// a parent table.
type ForeignKey struct {
	columns              results.Columns
	constraintName       string
	foreignDatabase      string
	foreignTable         string
	localToForeignColumn map[string]string
}

// NewForeignKey returns a new foreign key for the column c.
func NewForeignKey(c *results.Column, name, db, tb, col string) ForeignKey {
	return ForeignKey{
		columns:         results.Columns{c},
		constraintName:  name,
		foreignDatabase: db,
		foreignTable:    tb,
		localToForeignColumn: map[string]string{
			c.Name: col,
		},
	}
}

// InformationSchemaDual is a MongoTable that represents the "dual table"
// in the information_schema database. This is used for ADL translations.
// "dual" does not traditionally exist as a specific table in a specific
// database, however in ADL's execution engine, it is easiest to pretend
// it does. We chose to represent dual as a table in information_schema.
var InformationSchemaDual = MongoTable{
	name:           "DUAL",
	collation:      collation.Default,
	columns:        results.Columns{},
	columnMap:      map[string]*results.Column{},
	primaryKeys:    results.Columns{},
	indexes:        []Index{},
	foreignKeys:    []ForeignKey{},
	comments:       "",
	tableType:      BaseTable,
	isSharded:      false,
	collectionName: "DUAL",
	pipeline:       nil,
}
