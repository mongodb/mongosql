package mongotranslate

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/parser"
)

// InferSchemaFromQuery walks a query to infer the schema used in the query
// by looking for database names, table names, and column references.
// It is not possible to perfectly capture the schema from a query, so this function
// makes several assumptions when necessary.
//
// There are countless examples of queries that contain ambiguity, and in ambiguous cases
// the follow associations are made:
//   - unqualified tables are associated with the default database
//   - unqualified columns are associated with all tables (in all databases)
//
// For example, consider the query:
//     select a, b from t1, t2;
// This query does not contain any database name, therefore the defaultDB is used as the
// database for both tables t1 and t2. Additionally, this query does not qualify which
// tables the columns a and b belong to; therefore, both columns are associated with both
// tables. The resulting inferred schema is:
//   - Database: <default>
//     - Table: "t1"
//       - Columns: "a", "b"
//     - Table: "t2"
//       - Columns: "a", "b"
//
// On the other hand, qualified names and references will be directly used:
//     select t1.a, t2.b from db1.t1, db2.t1 as t2;
// This query qualifies every reference, so the inferred schema is:
//   - Database: "db1"
//     - Table: "t1"
//       - Columns: "a"
//   - Database: "db2"
//     - Table: "t1"       - (note that this is the real table name, not the alias)
//       - Columns: "b"
func InferSchemaFromQuery(query, defaultDB string) (InferredSchema, error) {
	stmt, err := parser.Parse(query)
	if err != nil {
		return nil, err
	}

	nw := newSchemaWalker(defaultDB)
	_, err = parser.Walk(nw, stmt)
	if err != nil {
		return nil, err
	}

	is := newInferredSchema()

	// Collect the tables found in the query.
	for _, t := range nw.tables {
		is.ensureTable(t.database, t.name)
	}

	// Infer the schema from the columns found in the query.
	for _, c := range nw.columns {
		switch {
		// The column is fully qualified (simplest case).
		case c.databaseIsKnown && c.tableIsKnown:
			is.ensureTable(c.database, c.table)
			is.schema[c.database][c.table][c.name] = struct{}{}

		// The column is only qualified by a table name.
		case c.tableIsKnown:
			if realTable, isAlias := nw.tableAliases[c.table]; isAlias {
				// if the column's table is an alias, lookup the real table.

				// if the "real table" is nil, this is an alias for a sub-query
				// and therefore no column associations should be made.
				if realTable == nil {
					continue
				}

				// otherwise, associate the column with the real table's data.
				is.schema[realTable.database][realTable.name][c.name] = struct{}{}

			} else if dbs, ok := nw.tableDatabases[c.table]; ok {
				// if the table is known to exist in some database(s), use them.
				for db := range dbs {
					is.schema[db][c.table][c.name] = struct{}{}
				}

			} else {
				// no databases are known for this table, use default.
				is.ensureTable(c.database, c.table)
				is.schema[c.database][c.table][c.name] = struct{}{}
			}

		// The column is not qualified at all; associate it with all tables.
		default:
			for db := range is.schema {
				for table := range is.schema[db] {
					is.schema[db][table][c.name] = struct{}{}
				}
			}
		}
	}

	return is, nil
}

// column is a simple struct storing information about the ColNames
// found in the query while walking.
type column struct {
	database, table, name         string
	databaseIsKnown, tableIsKnown bool
}

// table is a simple struct storing information about the TableNames
// and AliasedTableExprs found in the query while walking.
type table struct {
	database, name string
}

// schemaWalker is a walker implementation that collects the databases,
// tables, and columns referenced in a query.
type schemaWalker struct {
	defaultDB string
	columns   []*column
	tables    []*table

	// tableAliases maps from alias name to table pointer.
	//   - the produced schema will attribute columns referenced via an
	//     alias to the actual table.
	tableAliases map[string]*table

	// aliasedTables maps from database to a set of table names
	// which are aliased.
	aliasedTables map[string]map[string]struct{}

	// tableDatabases is a reverse lookup map from table name to set of
	// database name that contain a table with that name.
	tableDatabases map[string]map[string]struct{}
}

func newSchemaWalker(defaultDB string) *schemaWalker {
	return &schemaWalker{
		defaultDB:      defaultDB,
		columns:        make([]*column, 0),
		tables:         make([]*table, 0),
		tableAliases:   make(map[string]*table),
		aliasedTables:  make(map[string]map[string]struct{}),
		tableDatabases: make(map[string]map[string]struct{}),
	}
}

// addDatabaseToTable adds the argued database to the argued table's reverse
// mapping.
func (sw *schemaWalker) addDatabaseToTable(database, table string) {
	if _, ok := sw.tableDatabases[table]; !ok {
		sw.tableDatabases[table] = make(map[string]struct{})
	}

	sw.tableDatabases[table][database] = struct{}{}
}

// PreVisit is called for every node before its children are walked.
func (sw *schemaWalker) PreVisit(current parser.CST) (parser.CST, error) {
	// The schemaWalker uses PreVisit so that AliasedTableExprs are visited
	// before their children.

	switch t := current.(type) {
	case *parser.ColName:
		if t == nil {
			break
		}

		database := t.Database.Else(sw.defaultDB)
		tableName := t.Qualifier.Else("")

		sw.columns = append(sw.columns, &column{
			database:        database,
			table:           tableName,
			name:            t.Name,
			databaseIsKnown: t.Database.IsSome(),
			tableIsKnown:    t.Qualifier.IsSome(),
		})

	case *parser.TableName:
		if t == nil {
			break
		}

		database := t.Qualifier.Else(sw.defaultDB)
		tableName := t.Name

		if _, ok := sw.aliasedTables[database]; ok {
			if _, ok = sw.aliasedTables[database][tableName]; ok {
				break // if this table is aliased, do nothing.
			}
		}

		sw.addDatabaseToTable(database, tableName)
		sw.tables = append(sw.tables, &table{database: database, name: tableName})

	case *parser.AliasedTableExpr:
		// Only need to consider AliasedTableExprs where the sub-expression
		// is a TableName node.
		if tableExpr, ok := t.Expr.(*parser.TableName); ok {
			database := tableExpr.Qualifier.Else(sw.defaultDB)
			tableName := tableExpr.Name

			table := &table{database: database, name: tableName}

			if t.As.IsSome() {
				// if there is an alias, point it to the actual table.
				sw.tableAliases[t.As.Unwrap()] = table

				// also, mark the real table as aliased.
				if _, ok := sw.aliasedTables[database]; !ok {
					sw.aliasedTables[database] = make(map[string]struct{})
				}
				sw.aliasedTables[database][tableName] = struct{}{}
			} else {
				// if there is no alias in the expression, add this database
				// to the list of known databases for this table name.
				sw.addDatabaseToTable(database, tableName)
			}

			sw.tables = append(sw.tables, table)
		} else {
			// t.Expr is a sub-query.
			if t.As.IsSome() {
				// if there is an alias, mark that the alias
				// does not point to any named table.
				sw.tableAliases[t.As.Unwrap()] = nil
			}
		}
	}

	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (sw *schemaWalker) PostVisit(current parser.CST) (parser.CST, error) {
	return current, nil
}

// InferredSchema is an interface with methods for getting the databases,
// tables, and columns of a schema.
type InferredSchema interface {
	fmt.Stringer

	// Databases gets the list of database names in the schema.
	Databases() []string

	// Tables gets the list of table names for the provided database if it exists.
	Tables(database string) ([]string, error)

	// Columns gets the list of column names for the provided table and database
	// if they both exist.
	Columns(database, table string) ([]string, error)
}

type inferredSchema struct {
	databases []string
	schema    map[string]map[string]map[string]struct{}
}

func newInferredSchema() *inferredSchema {
	return &inferredSchema{
		databases: make([]string, 0),
		schema:    make(map[string]map[string]map[string]struct{}),
	}
}

// ensureDatabase ensures the provided database is initialized in
// the schema map.
func (is *inferredSchema) ensureDatabase(database string) {
	if _, ok := is.schema[database]; !ok {
		is.schema[database] = make(map[string]map[string]struct{})
		is.databases = append(is.databases, database)
	}
}

// ensureTable ensures the provided table is initialized in the
// schema map. It also ensures the database is initialized.
func (is *inferredSchema) ensureTable(database, table string) {
	is.ensureDatabase(database)

	if _, ok := is.schema[database][table]; !ok {
		is.schema[database][table] = make(map[string]struct{})
	}
}

// Databases gets the list of database names in the schema.
func (is *inferredSchema) Databases() []string {
	return is.databases
}

// Tables gets the list of table names for the provided database if it exists.
func (is *inferredSchema) Tables(database string) ([]string, error) {
	if tables, ok := is.schema[database]; ok {
		tableNames := make([]string, len(tables))

		i := 0
		for tableName := range tables {
			tableNames[i] = tableName
			i++
		}

		return tableNames, nil
	}

	return nil, fmt.Errorf("no such database: %s", database)
}

// Columns gets the list of column names for the provided table and database
// if they both exist.
func (is *inferredSchema) Columns(database, table string) ([]string, error) {
	if tables, ok := is.schema[database]; ok {
		if columns, ok := tables[table]; ok {
			columnNames := make([]string, len(columns))

			i := 0
			for columnName := range columns {
				columnNames[i] = columnName
				i++
			}

			return columnNames, nil
		}

		return nil, fmt.Errorf("no such table: %s.%s", database, table)
	}

	return nil, fmt.Errorf("no such database: %s", database)
}

func (is *inferredSchema) String() string {
	var sb strings.Builder

	dbs := is.Databases()

	sb.WriteString(fmt.Sprintf("Databases: %v\n", strings.Join(dbs, ",")))

	for _, db := range dbs {
		sb.WriteString(fmt.Sprintf("Database: %v\n", db))
		tables, _ := is.Tables(db)

		for _, table := range tables {
			sb.WriteString(fmt.Sprintf("\tTable: %v\n", table))
			columns, _ := is.Columns(db, table)
			sb.WriteString(fmt.Sprintf("\t\tColumns: %v\n", strings.Join(columns, ",")))
		}
	}

	return sb.String()
}
