package catalog

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// Build builds a catalog up from a schema and variables.
func Build(schema *schema.Schema, variables *variable.Container, info *mongodb.Info, writeMode bool) (*SQLCatalog, error) {
	builder := &catalogBuilder{
		catalog:   New("def"),
		schema:    schema,
		variables: variables,
		info:      info,
		writeMode: writeMode,
	}

	err := builder.build()
	if err != nil {
		return nil, err
	}
	builder.schema = nil
	return builder.catalog, nil
}

// BuildFromSchema builds a catalog from a schema.
func BuildFromSchema(schema *schema.Schema, info *mongodb.Info, writeMode bool) (*SQLCatalog, error) {
	builder := &catalogBuilder{
		catalog:   New("def"),
		schema:    schema,
		info:      info,
		writeMode: writeMode,
	}

	err := builder.buildFromSchema()
	if err != nil {
		return nil, err
	}
	builder.schema = nil
	return builder.catalog, nil
}

type catalogBuilder struct {
	catalog   *SQLCatalog
	info      *mongodb.Info
	schema    *schema.Schema
	variables *variable.Container
	writeMode bool
}

func (b *catalogBuilder) build() error {
	err := b.buildFromSchema()
	if err != nil {
		return err
	}

	err = b.buildInformationSchemaDatabase()
	if err != nil {
		return err
	}

	return b.buildMySQLDatabase()
}

func (b *catalogBuilder) buildFromSchema() error {
	info := b.info

	dbs := b.schema.DatabasesSorted()
	for _, dbConfig := range dbs {
		if !info.IsVisibleDatabase(mongodb.DatabaseName(dbConfig.Name())) {
			b.catalog.containsAuthRestrictedNamespaces = true
			continue
		}

		d, err := b.catalog.Database(context.Background(), dbConfig.Name())
		if err != nil {
			d, err = b.catalog.AddDatabase(dbConfig.Name())
			if err != nil {
				return err
			}
		}

		dInfo, ok := info.Databases[mongodb.DatabaseName(strings.ToLower(dbConfig.Name()))]
		if !ok {
			continue
		}

		tbls := dbConfig.TablesSorted()
		for _, tblConfig := range tbls {
			allowed := info.IsVisibleCollection(
				mongodb.DatabaseName(dbConfig.Name()),
				mongodb.CollectionName(tblConfig.MongoName()),
			)
			if !allowed {
				b.catalog.containsAuthRestrictedNamespaces = true
				continue
			}

			collection, ok := dInfo.Collections[mongodb.CollectionName(tblConfig.MongoName())]
			if !ok {
				continue
			}

			col := collation.Default
			if collection.Collation != nil {
				col, err = collationFromMongoDB(collection.Collation)
				if err != nil {
					return mysqlerrors.Newf(
						mysqlerrors.ErUnknownCollation,
						"unable to translate MongoDB's collation for %q.%q: %v",
						dbConfig.Name(), tblConfig.SQLName(), err,
					)
				}
			}

			tableType := BaseTable
			if collection.IsView {
				tableType = View
			}

			t := NewMongoTable(dbConfig.Name(), tblConfig, tableType, col, b.writeMode)

			t.isSharded = collection.IsSharded

			// If we are not in --writeMode build the indexes the old way.
			// In --writeMode we will get the indexes from the schema.
			if !b.writeMode {
				mongoNameToColumn := make(map[string]*results.Column, len(t.columns))

				for _, c := range t.columns {
					mongoNameToColumn[c.MongoName] = c
				}

				idx := 1
				for _, i := range collection.Indexes {
					index := addColumnToIndex(i, mongoNameToColumn)
					if index != nil {
						if i.Unique {
							index.constraintName = createUniqueIndexName(
								dbConfig.Name(),
								tblConfig.SQLName(),
								idx,
							)
							index.unique = true
							idx++
						}
						t.indexes = append(t.indexes, *index)
					}
				}
			}

			if err = d.AddTable(t); err != nil {
				return err
			}
		}
	}

	// Adding foreign keys occurs after visiting all namespaces since such keys
	// can only be identified if all tables within a collection are known.
	b.addForeignKeys()
	return nil
}

func (b *catalogBuilder) buildInformationSchemaDatabase() error {
	d, err := b.catalog.AddDatabase(InformationSchemaDatabase)
	if err != nil {
		return err
	}
	err = b.addCharsetTable(d)
	if err != nil {
		return err
	}
	err = b.addCollationTable(d)
	if err != nil {
		return err
	}
	err = b.addCollationCharacterSetApplicabilityTable(d)
	if err != nil {
		return err
	}
	err = b.addColumnsTable(d)
	if err != nil {
		return err
	}
	err = b.addColumnPrivilegesTable(d)
	if err != nil {
		return err
	}
	err = b.addEnginesTable(d)
	if err != nil {
		return err
	}
	err = b.addEventsTable(d)
	if err != nil {
		return err
	}
	err = b.addFilesTable(d)
	if err != nil {
		return err
	}
	err = b.addKeyColumnUsageTable(d)
	if err != nil {
		return err
	}
	err = b.addNdbTransidMysqlConnectionMapTable(d)
	if err != nil {
		return err
	}
	err = b.addParametersTable(d)
	if err != nil {
		return err
	}
	err = b.addPartitionsTable(d)
	if err != nil {
		return err
	}
	err = b.addPluginsTable(d)
	if err != nil {
		return err
	}
	err = b.addProfilingTable(d)
	if err != nil {
		return err
	}
	err = b.addReferentialConstraintsTable(d)
	if err != nil {
		return err
	}
	err = b.addRoutinesTable(d)
	if err != nil {
		return err
	}
	err = b.addSchemataTable(d)
	if err != nil {
		return err
	}
	err = b.addSchemaPrivileges(d)
	if err != nil {
		return err
	}
	err = b.addStatisticsTable(d)
	if err != nil {
		return err
	}
	err = b.addTablesTable(d)
	if err != nil {
		return err
	}
	err = b.addTableSpacesTable(d)
	if err != nil {
		return err
	}
	err = b.addTableConstraintsTable(d)
	if err != nil {
		return err
	}
	err = b.addTablePrivilegesTable(d)
	if err != nil {
		return err
	}
	err = b.addTriggersTable(d)
	if err != nil {
		return err
	}
	err = b.addUserPrivilegesTable(d)
	if err != nil {
		return err
	}
	err = b.addViewsTable(d)
	if err != nil {
		return err
	}
	return b.addVariableTables(d)
}

func newInfoRow(aliasName string, vs ...values.NamedSQLValue) results.Row {
	return results.NewNamedRow("information_schema", aliasName, vs...)
}

// getValueCreators returns three helper functions for creating NameSQLValues of type
// SQLNull, SQLVarchar, and SQLInt64 respectively, each with the proper SQLValueKind.
func getValueCreators(variables *variable.Container) (func(string) values.NamedSQLValue,
	func(string, string) values.NamedSQLValue,
	func(string, int64) values.NamedSQLValue) {

	kind := values.MongoSQLValueKind
	if variables.GetString(variable.TypeConversionMode) == variable.MySQLTypeConversionMode {
		kind = values.MySQLValueKind
	}
	nullv := func(name string) values.NamedSQLValue {
		return values.NewNamedSQLValue(name, values.NewSQLNull(kind))
	}
	strv := func(name, s string) values.NamedSQLValue {
		return values.NewNamedSQLValue(name, values.NewSQLVarchar(kind, s))
	}
	intv := func(name string, i int64) values.NamedSQLValue {
		return values.NewNamedSQLValue(name, values.NewSQLInt64(kind, i))
	}
	return nullv, strv, intv
}

func (b *catalogBuilder) addCharsetTable(d Database) error {
	characterSetName := "CHARACTER_SET_NAME"
	defaultCollateName := "DEFAULT_COLLATE_NAME"
	description := "DESCRIPTION"
	maxLen := "MAXLEN"
	t := NewDynamicTable(CharacterSetsTable, SystemView, func(aliasName string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})
		_, strv, intv := getValueCreators(b.variables)
		go func() {
			defer close(rowChan)
			for _, c := range collation.GetAllCharsets() {
				select {
				case rowChan <- newInfoRow(aliasName,
					strv(characterSetName, string(c.Name)),
					strv(defaultCollateName, string(c.DefaultCollationName)),
					strv(description, c.Description),
					intv(maxLen, int64(c.MaxLen)),
				):
				case <-done:
					return
				}
			}
		}()
		return results.NewRowChanIter(rowChan, done)
	})

	t.AddColumns(CharacterSetsTable,
		NewDynamicColumnDeclaration(characterSetName, types.EvalString),
		NewDynamicColumnDeclaration(defaultCollateName, types.EvalString),
		NewDynamicColumnDeclaration(description, types.EvalString),
		NewDynamicColumnDeclaration(maxLen, types.EvalInt64),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addCollationTable(d Database) error {
	collationName := "COLLATION_NAME"
	characterSetName := "CHARACTER_SET_NAME"
	id := "ID"
	isDefault := "IS_DEFAULT"
	isCompiled := "IS_COMPILED"
	sortlen := "SORTLEN"
	t := NewDynamicTable(CollationsTable, SystemView, func(aliasName string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})
		_, strv, intv := getValueCreators(b.variables)
		go func() {
			defer close(rowChan)
			for _, c := range collation.GetAll() {
				def := "No"

				if c.Default {
					def = "Yes"
				}

				select {
				case rowChan <- newInfoRow(aliasName,
					strv(collationName, string(c.Name)),
					strv(characterSetName, string(c.CharsetName)),
					intv(id, int64(c.ID)),
					strv(isDefault, def),
					strv(isCompiled, "Yes"),
					intv(sortlen, int64(c.SortLen)),
				):
				case <-done:
					return
				}
			}
		}()
		return results.NewRowChanIter(rowChan, done)
	})

	t.AddColumns(CollationsTable,
		NewDynamicColumnDeclaration(collationName, types.EvalString),
		NewDynamicColumnDeclaration(characterSetName, types.EvalString),
		NewDynamicColumnDeclaration(id, types.EvalInt64),
		NewDynamicColumnDeclaration(isDefault, types.EvalString),
		NewDynamicColumnDeclaration(isCompiled, types.EvalString),
		NewDynamicColumnDeclaration(sortlen, types.EvalInt64),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addCollationCharacterSetApplicabilityTable(d Database) error {
	collationName := "COLLATION_NAME"
	characterSetName := "CHARACTER_SET_NAME"

	t := NewDynamicTable(CollationCharacterSetApplicabilityTable, SystemView,
		func(aliasName string) results.RowIter {
			rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
			done := make(chan struct{})
			_, strv, _ := getValueCreators(b.variables)
			go func() {
				defer close(rowChan)
				for _, c := range collation.GetAll() {
					select {
					case rowChan <- newInfoRow(
						aliasName,
						strv(collationName, string(c.Name)),
						strv(characterSetName, string(c.CharsetName)),
					):
					case <-done:
						return
					}
				}
			}()

			return results.NewRowChanIter(rowChan, done)
		})

	t.AddColumns(CollationCharacterSetApplicabilityTable,
		NewDynamicColumnDeclaration(collationName, types.EvalString),
		NewDynamicColumnDeclaration(characterSetName, types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addColumnsTable(d Database) error {
	c := b.catalog
	tableName := ColumnsTable
	columnDecls := []DynamicColumnDeclaration{
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLUMN_NAME", types.EvalString),
		NewDynamicColumnDeclaration("ORDINAL_POSITION", types.EvalInt64),
		NewDynamicColumnDeclaration("COLUMN_DEFAULT", types.EvalString),
		NewDynamicColumnDeclaration("IS_NULLABLE", types.EvalString),
		NewDynamicColumnDeclaration("DATA_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_MAXIMUM_LENGTH", types.EvalUint64),
		NewDynamicColumnDeclaration("CHARACTER_OCTET_LENGTH", types.EvalUint64),
		NewDynamicColumnDeclaration("NUMERIC_PRECISION", types.EvalUint64),
		NewDynamicColumnDeclaration("NUMERIC_SCALE", types.EvalUint64),
		NewDynamicColumnDeclaration("DATETIME_PRECISION", types.EvalUint64),
		NewDynamicColumnDeclaration("CHARACTER_SET_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLUMN_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("COLUMN_KEY", types.EvalString),
		NewDynamicColumnDeclaration("EXTRA", types.EvalString),
		NewDynamicColumnDeclaration("PRIVILEGES", types.EvalString),
		NewDynamicColumnDeclaration("COLUMN_COMMENT", types.EvalString),
		NewDynamicColumnDeclaration("GENERATION_EXPRESSION", types.EvalString),
	}

	columnNames := make([]string, len(columnDecls))
	for i := range columnNames {
		columnNames[i] = columnDecls[i].columnName
	}

	t := NewDynamicTable(tableName, SystemView, func(aliasName string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})
		nullv, strv, intv := getValueCreators(b.variables)

		go func() {
			defer close(rowChan)
			dbs, _ := c.Databases(context.Background())
			for _, db := range dbs {
				tbls, _ := db.Tables(context.Background())
				for _, tbl := range tbls {
					for i, col := range tbl.Columns() {
						columnKey := getIndexKey(col, tbl)
						maxVarcharLength := b.variables.GetUint64(variable.MongoDBMaxVarcharLength)
						columnType := translateColumnType(col.ColumnType.EvalType, maxVarcharLength)
						dataType := columnType
						if idx := strings.Index(dataType, "("); idx >= 0 {
							dataType = dataType[:idx]
						}
						nullable := "NO"
						if col.Nullable {
							nullable = "YES"
						}
						select {
						case rowChan <- newInfoRow(
							aliasName,
							strv(columnNames[0], string(c.Name)),
							strv(columnNames[1], string(db.Name())),
							strv(columnNames[2], tbl.Name()),
							strv(columnNames[3], col.Name),
							intv(columnNames[4], int64(i+1)),
							nullv(columnNames[5]),
							strv(columnNames[6], nullable),
							strv(columnNames[7], dataType),
							nullv(columnNames[8]),
							nullv(columnNames[9]),
							nullv(columnNames[10]),
							nullv(columnNames[11]),
							nullv(columnNames[12]),
							strv(columnNames[13], string(tbl.Collation().CharsetName)),
							strv(columnNames[14], string(tbl.Collation().Name)),
							strv(columnNames[15], columnType),
							strv(columnNames[16], columnKey),
							strv(columnNames[17], ""),
							strv(columnNames[18], "select"),
							strv(columnNames[19], col.Comments),
							strv(columnNames[20], ""),
						):
						case <-done:
							return
						}
					}
				}
			}
		}()
		return results.NewRowChanIter(rowChan, done)
	})

	t.AddColumns(tableName,
		columnDecls...,
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addColumnPrivilegesTable(d Database) error {
	t := NewDynamicTable(ColumnPrivilegesTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(ColumnPrivilegesTable,
		NewDynamicColumnDeclaration("GRANTEE", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLUMN_NAME", types.EvalString),
		NewDynamicColumnDeclaration("PRIVILEGE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("IS_GRANTABLE", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addEnginesTable(d Database) error {
	t := NewDynamicTable(EnginesTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(EnginesTable,
		NewDynamicColumnDeclaration("ENGINE", types.EvalString),
		NewDynamicColumnDeclaration("SUPPORT", types.EvalString),
		NewDynamicColumnDeclaration("COMMENT", types.EvalString),
		NewDynamicColumnDeclaration("TRANSACTIONS", types.EvalString),
		NewDynamicColumnDeclaration("XA", types.EvalString),
		NewDynamicColumnDeclaration("SAVEPOINTS", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addEventsTable(d Database) error {
	t := NewDynamicTable(EventsTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(EventsTable,
		NewDynamicColumnDeclaration("EVENT_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_NAME", types.EvalString),
		NewDynamicColumnDeclaration("DEFINER", types.EvalString),
		NewDynamicColumnDeclaration("TIME_ZONE", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_BODY", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_DEFINITION", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("EXECUTE_AT", types.EvalString),
		NewDynamicColumnDeclaration("INTERVAL_VALUE", types.EvalString),
		NewDynamicColumnDeclaration("INTERVAL_FIELD", types.EvalString),
		NewDynamicColumnDeclaration("SQL_MODE", types.EvalString),
		NewDynamicColumnDeclaration("STARTS", types.EvalString),
		NewDynamicColumnDeclaration("ENDS", types.EvalString),
		NewDynamicColumnDeclaration("STATUS", types.EvalString),
		NewDynamicColumnDeclaration("ON_COMPLETION", types.EvalString),
		NewDynamicColumnDeclaration("CREATED", types.EvalString),
		NewDynamicColumnDeclaration("LAST_ALTERED", types.EvalString),
		NewDynamicColumnDeclaration("LAST_EXECUTED", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_COMMENT", types.EvalString),
		NewDynamicColumnDeclaration("ORIGINATOR", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_SET_CLIENT", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION_CONNECTION", types.EvalString),
		NewDynamicColumnDeclaration("DATABASE_COLLATION", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addFilesTable(d Database) error {
	files := "FILES"
	t := NewDynamicTable(files, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(files,
		NewDynamicColumnDeclaration("FILE_ID", types.EvalString),
		NewDynamicColumnDeclaration("FILE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("FILE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("TABLESPACE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("LOGFILE_GROUP_NAME", types.EvalString),
		NewDynamicColumnDeclaration("LOGFILE_GROUP_NUMBER", types.EvalString),
		NewDynamicColumnDeclaration("ENGINE", types.EvalString),
		NewDynamicColumnDeclaration("FULLTEXT_KEYS", types.EvalString),
		NewDynamicColumnDeclaration("DELETED_ROWS", types.EvalString),
		NewDynamicColumnDeclaration("UPDATE_COUNT", types.EvalString),
		NewDynamicColumnDeclaration("FREE_EXTENTS", types.EvalString),
		NewDynamicColumnDeclaration("TOTAL_EXTENTS", types.EvalString),
		NewDynamicColumnDeclaration("EXTENT_SIZE", types.EvalString),
		NewDynamicColumnDeclaration("INITIAL_SIZE", types.EvalString),
		NewDynamicColumnDeclaration("MAXIMUM_SIZE", types.EvalString),
		NewDynamicColumnDeclaration("AUTOEXTEND_SIZE", types.EvalString),
		NewDynamicColumnDeclaration("CREATION_TIME", types.EvalString),
		NewDynamicColumnDeclaration("LAST_UPDATE_TIME", types.EvalString),
		NewDynamicColumnDeclaration("LAST_ACCESS_TIME", types.EvalString),
		NewDynamicColumnDeclaration("RECOVER_TIME", types.EvalString),
		NewDynamicColumnDeclaration("TRANSACTION_COUNTER", types.EvalString),
		NewDynamicColumnDeclaration("VERSION", types.EvalString),
		NewDynamicColumnDeclaration("ROW_FORMAT", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_ROWS", types.EvalString),
		NewDynamicColumnDeclaration("AVG_ROW_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("DATA_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("MAX_DATA_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("INDEX_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("DATA_FREE", types.EvalString),
		NewDynamicColumnDeclaration("CREATE_TIME", types.EvalString),
		NewDynamicColumnDeclaration("UPDATE_TIME", types.EvalString),
		NewDynamicColumnDeclaration("CHECK_TIME", types.EvalString),
		NewDynamicColumnDeclaration("CHECKSUM", types.EvalString),
		NewDynamicColumnDeclaration("STATUS", types.EvalString),
		NewDynamicColumnDeclaration("EXTRA", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addKeyColumnUsageTable(d Database) error {
	tableName := KeyColumnUsageTable
	columnDecls := []DynamicColumnDeclaration{
		NewDynamicColumnDeclaration("CONSTRAINT_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("CONSTRAINT_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("CONSTRAINT_NAME", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLUMN_NAME", types.EvalString),
		NewDynamicColumnDeclaration("ORDINAL_POSITION", types.EvalInt64),
		NewDynamicColumnDeclaration("POSITION_IN_UNIQUE_CONSTRAINT", types.EvalInt64),
		NewDynamicColumnDeclaration("REFERENCED_TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("REFERENCED_TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("REFERENCED_COLUMN_NAME", types.EvalString),
	}

	columnNames := make([]string, len(columnDecls))
	for i := range columnNames {
		columnNames[i] = columnDecls[i].columnName
	}

	t := NewDynamicTable(tableName, SystemView, func(aliasName string) results.RowIter {
		return b.getRowsForTableType(tableName, aliasName, columnNames)
	})

	t.AddColumns(tableName,
		columnDecls...,
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addNdbTransidMysqlConnectionMapTable(d Database) error {
	t := NewDynamicTable(NdbTransidMysqlConnectionMapTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(NdbTransidMysqlConnectionMapTable,
		NewDynamicColumnDeclaration("mysql_connection_id", types.EvalString),
		NewDynamicColumnDeclaration("node_id", types.EvalString),
		NewDynamicColumnDeclaration("ndb_transid", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addParametersTable(d Database) error {
	t := NewDynamicTable(ParametersTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(ParametersTable,
		NewDynamicColumnDeclaration("SPECIFIC_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("SPECIFIC_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("SPECIFIC_NAME", types.EvalString),
		NewDynamicColumnDeclaration("ORDINAL_POSITION", types.EvalString),
		NewDynamicColumnDeclaration("PARAMETER_MODE", types.EvalString),
		NewDynamicColumnDeclaration("PARAMETER_NAME", types.EvalString),
		NewDynamicColumnDeclaration("DATA_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_MAXIMUM_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_OCTET_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("NUMERIC_PRECISION", types.EvalString),
		NewDynamicColumnDeclaration("NUMERIC_SCALE", types.EvalString),
		NewDynamicColumnDeclaration("DATETIME_PRECISION", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_SET_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION_NAME", types.EvalString),
		NewDynamicColumnDeclaration("DTD_IDENTIFIER", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_TYPE", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addPartitionsTable(d Database) error {
	t := NewDynamicTable(PartitionsTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(PartitionsTable,
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("PARTITION_NAME", types.EvalString),
		NewDynamicColumnDeclaration("SUBPARTITION_NAME", types.EvalString),
		NewDynamicColumnDeclaration("PARTITION_ORDINAL_POSITION", types.EvalString),
		NewDynamicColumnDeclaration("SUBPARTITION_ORDINAL_POSITION", types.EvalString),
		NewDynamicColumnDeclaration("PARTITION_METHOD", types.EvalString),
		NewDynamicColumnDeclaration("SUBPARTITION_METHOD", types.EvalString),
		NewDynamicColumnDeclaration("PARTITION_EXPRESSION", types.EvalString),
		NewDynamicColumnDeclaration("SUBPARTITION_EXPRESSION", types.EvalString),
		NewDynamicColumnDeclaration("PARTITION_DESCRIPTION", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_ROWS", types.EvalString),
		NewDynamicColumnDeclaration("AVG_ROW_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("DATA_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("MAX_DATA_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("INDEX_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("DATA_FREE", types.EvalString),
		NewDynamicColumnDeclaration("CREATE_TIME", types.EvalString),
		NewDynamicColumnDeclaration("UPDATE_TIME", types.EvalString),
		NewDynamicColumnDeclaration("CHECK_TIME", types.EvalString),
		NewDynamicColumnDeclaration("CHECKSUM", types.EvalString),
		NewDynamicColumnDeclaration("PARTITION_COMMENT", types.EvalString),
		NewDynamicColumnDeclaration("NODEGROUP", types.EvalString),
		NewDynamicColumnDeclaration("TABLESPACE_NAME", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addPluginsTable(d Database) error {
	t := NewDynamicTable(PluginsTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(PluginsTable,
		NewDynamicColumnDeclaration("PLUGIN_NAME", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_VERSION", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_STATUS", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_TYPE_VERSION", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_LIBRARY", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_LIBRARY_VERSION", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_AUTHOR", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_DESCRIPTION", types.EvalString),
		NewDynamicColumnDeclaration("PLUGIN_LICENSE", types.EvalString),
		NewDynamicColumnDeclaration("LOAD_OPTION", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addProfilingTable(d Database) error {
	t := NewDynamicTable(ProfilingTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(ProfilingTable,
		NewDynamicColumnDeclaration("QUERY_ID", types.EvalString),
		NewDynamicColumnDeclaration("SEQ", types.EvalString),
		NewDynamicColumnDeclaration("STATE", types.EvalString),
		NewDynamicColumnDeclaration("DURATION", types.EvalString),
		NewDynamicColumnDeclaration("CPU_USER", types.EvalString),
		NewDynamicColumnDeclaration("CPU_SYSTEM", types.EvalString),
		NewDynamicColumnDeclaration("CONTEXT_VOLUNTARY", types.EvalString),
		NewDynamicColumnDeclaration("CONTEXT_INVOLUNTARY", types.EvalString),
		NewDynamicColumnDeclaration("BLOCK_OPS_IN", types.EvalString),
		NewDynamicColumnDeclaration("BLOCK_OPS_OUT", types.EvalString),
		NewDynamicColumnDeclaration("MESSAGES_SENT", types.EvalString),
		NewDynamicColumnDeclaration("MESSAGES_RECEIVED", types.EvalString),
		NewDynamicColumnDeclaration("PAGE_FAULTS_MAJOR", types.EvalString),
		NewDynamicColumnDeclaration("PAGE_FAULTS_MINOR", types.EvalString),
		NewDynamicColumnDeclaration("SWAPS", types.EvalString),
		NewDynamicColumnDeclaration("SOURCE_FUNCTION", types.EvalString),
		NewDynamicColumnDeclaration("SOURCE_FILE", types.EvalString),
		NewDynamicColumnDeclaration("SOURCE_LINE", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addReferentialConstraintsTable(d Database) error {
	tableName := ReferentialConstraintsTable
	// All of the columns in this table are type varchar, so we can just
	// declare the names, and generate the decls from the names rather than
	// vice versa.
	columnNames := []string{
		"CONSTRAINT_CATALOG",
		"CONSTRAINT_SCHEMA",
		"CONSTRAINT_NAME",
		"UNIQUE_CONSTRAINT_CATALOG",
		"UNIQUE_CONSTRAINT_SCHEMA",
		"UNIQUE_CONSTRAINT_NAME",
		"MATCH_OPTION",
		"UPDATE_RULE",
		"DELETE_RULE",
		"TABLE_NAME",
		"REFERENCED_TABLE_NAME",
	}

	t := NewDynamicTable(tableName, SystemView, func(aliasName string) results.RowIter {
		return b.getRowsForTableType(tableName, aliasName, columnNames)
	})

	columnDecls := make([]DynamicColumnDeclaration, len(columnNames))
	for i := range columnNames {
		columnDecls[i] = NewDynamicColumnDeclaration(columnNames[i], types.EvalString)
	}

	t.AddColumns(tableName,
		columnDecls...,
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addRoutinesTable(d Database) error {
	t := NewDynamicTable(RoutinesTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(RoutinesTable,
		NewDynamicColumnDeclaration("SPECIFIC_NAME", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("DATA_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_MAXIMUM_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_OCTET_LENGTH", types.EvalString),
		NewDynamicColumnDeclaration("NUMERIC_PRECISION", types.EvalString),
		NewDynamicColumnDeclaration("NUMERIC_SCALE", types.EvalString),
		NewDynamicColumnDeclaration("DATETIME_PRECISION", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_SET_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION_NAME", types.EvalString),
		NewDynamicColumnDeclaration("DTD_IDENTIFIER", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_BODY", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_DEFINITION", types.EvalString),
		NewDynamicColumnDeclaration("EXTERNAL_NAME", types.EvalString),
		NewDynamicColumnDeclaration("EXTERNAL_LANGUAGE", types.EvalString),
		NewDynamicColumnDeclaration("PARAMETER_STYLE", types.EvalString),
		NewDynamicColumnDeclaration("IS_DETERMINISTIC", types.EvalString),
		NewDynamicColumnDeclaration("SQL_DATA_ACCESS", types.EvalString),
		NewDynamicColumnDeclaration("SQL_PATH", types.EvalString),
		NewDynamicColumnDeclaration("SECURITY_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("CREATED", types.EvalString),
		NewDynamicColumnDeclaration("LAST_ALTERED", types.EvalString),
		NewDynamicColumnDeclaration("SQL_MODE", types.EvalString),
		NewDynamicColumnDeclaration("ROUTINE_COMMENT", types.EvalString),
		NewDynamicColumnDeclaration("DEFINER", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_SET_CLIENT", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION_CONNECTION", types.EvalString),
		NewDynamicColumnDeclaration("DATABASE_COLLATION", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addSchemataTable(d Database) error {
	c := b.catalog
	tableName := SchemataTable
	columnDecls := []DynamicColumnDeclaration{
		NewDynamicColumnDeclaration("CATALOG_NAME", types.EvalString),
		NewDynamicColumnDeclaration("SCHEMA_NAME", types.EvalString),
		NewDynamicColumnDeclaration("DEFAULT_CHARACTER_SET_NAME", types.EvalString),
		NewDynamicColumnDeclaration("DEFAULT_COLLATION_NAME", types.EvalString),
		NewDynamicColumnDeclaration("SQL_PATH", types.EvalString),
	}

	columnNames := make([]string, len(columnDecls))
	for i := range columnNames {
		columnNames[i] = columnDecls[i].columnName
	}

	t := NewDynamicTable(tableName, SystemView, func(aliasName string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})
		_, strv, _ := getValueCreators(b.variables)

		go func() {
			defer close(rowChan)
			dbs, _ := c.Databases(context.Background())
			for _, db := range dbs {
				select {
				case rowChan <- newInfoRow(aliasName,
					strv(columnNames[0], string(c.Name)),
					strv(columnNames[1], string(db.Name())),
					strv(columnNames[2], string(collation.DefaultCharset.Name)),
					strv(columnNames[3], string(collation.Default.Name)),
					strv(columnNames[4], ""),
				):
				case <-done:
					return
				}
			}
		}()
		return results.NewRowChanIter(rowChan, done)
	})

	t.AddColumns(tableName,
		columnDecls...,
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addSchemaPrivileges(d Database) error {
	t := NewDynamicTable(SchemaPrivilagesTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(SchemaPrivilagesTable,
		NewDynamicColumnDeclaration("GRANTEE", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("PRIVILEGE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("IS_GRANTABLE", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addStatisticsTable(d Database) error {
	tableName := StatisticsTable
	columnDecls := []DynamicColumnDeclaration{
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("NON_UNIQUE", types.EvalBoolean),
		NewDynamicColumnDeclaration("INDEX_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("INDEX_NAME", types.EvalString),
		NewDynamicColumnDeclaration("SEQ_IN_INDEX", types.EvalInt64),
		NewDynamicColumnDeclaration("COLUMN_NAME", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION", types.EvalString),
		NewDynamicColumnDeclaration("CARDINALITY", types.EvalString),
		NewDynamicColumnDeclaration("SUB_PART", types.EvalString),
		NewDynamicColumnDeclaration("PACKED", types.EvalString),
		NewDynamicColumnDeclaration("NULLABLE", types.EvalString),
		NewDynamicColumnDeclaration("INDEX_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("COMMENT", types.EvalString),
		NewDynamicColumnDeclaration("INDEX_COMMENT", types.EvalString),
	}

	columnNames := make([]string, len(columnDecls))
	for i := range columnNames {
		columnNames[i] = columnDecls[i].columnName
	}

	t := NewDynamicTable(tableName, SystemView, func(aliasName string) results.RowIter {
		return b.getRowsForTableType(tableName, aliasName, columnNames)
	})

	t.AddColumns(tableName,
		columnDecls...,
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTablesTable(d Database) error {
	c := b.catalog

	tableName := TablesTable
	columnDecls := []DynamicColumnDeclaration{
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("ENGINE", types.EvalString),
		NewDynamicColumnDeclaration("VERSION", types.EvalUint64),
		NewDynamicColumnDeclaration("ROW_FORMAT", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_ROWS", types.EvalUint64),
		NewDynamicColumnDeclaration("AVG_ROW_LENGTH", types.EvalUint64),
		NewDynamicColumnDeclaration("DATA_LENGTH", types.EvalUint64),
		NewDynamicColumnDeclaration("MAX_DATA_LENGTH", types.EvalUint64),
		NewDynamicColumnDeclaration("INDEX_LENGTH", types.EvalUint64),
		NewDynamicColumnDeclaration("DATA_FREE", types.EvalUint64),
		NewDynamicColumnDeclaration("AUTO_INCREMENT", types.EvalUint64),
		NewDynamicColumnDeclaration("CREATE_TIME", types.EvalDatetime),
		NewDynamicColumnDeclaration("UPDATE_TIME", types.EvalDatetime),
		NewDynamicColumnDeclaration("CHECK_TIME", types.EvalDatetime),
		NewDynamicColumnDeclaration("TABLE_COLLATION", types.EvalString),
		NewDynamicColumnDeclaration("CHECKSUM", types.EvalUint64),
		NewDynamicColumnDeclaration("CREATE_OPTIONS", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_COMMENT", types.EvalString),
	}

	columnNames := make([]string, len(columnDecls))
	for i := range columnNames {
		columnNames[i] = columnDecls[i].columnName
	}

	t := NewDynamicTable(TablesTable, SystemView, func(aliasName string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})

		nullv, strv, _ := getValueCreators(b.variables)
		go func() {
			defer close(rowChan)
			dbs, _ := c.Databases(context.Background())
			for _, db := range dbs {
				tbls, _ := db.Tables(context.Background())
				for _, tbl := range tbls {
					select {
					case rowChan <- newInfoRow(aliasName,
						strv(columnNames[0], string(c.Name)),
						strv(columnNames[1], string(db.Name())),
						strv(columnNames[2], tbl.Name()),
						strv(columnNames[3], tbl.Type()),
						strv(columnNames[4], ""),
						strv(columnNames[5], ""),
						strv(columnNames[6], ""),
						nullv(columnNames[7]),
						nullv(columnNames[8]),
						nullv(columnNames[9]),
						nullv(columnNames[10]),
						nullv(columnNames[11]),
						nullv(columnNames[12]),
						strv(columnNames[13], "NO"),
						nullv(columnNames[14]),
						nullv(columnNames[15]),
						nullv(columnNames[16]),
						strv(columnNames[17], string(tbl.Collation().Name)),
						nullv(columnNames[18]),
						nullv(columnNames[19]),
						strv(columnNames[20], tbl.Comments()),
					):
					case <-done:
						return
					}
				}
			}
		}()
		return results.NewRowChanIter(rowChan, done)
	})

	t.AddColumns(tableName,
		columnDecls...,
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTableSpacesTable(d Database) error {
	t := NewDynamicTable(TablespacesTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(TablespacesTable,
		NewDynamicColumnDeclaration("TABLESPACE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("ENGINE", types.EvalString),
		NewDynamicColumnDeclaration("TABLESPACE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("LOGFILE_GROUP_NAME", types.EvalString),
		NewDynamicColumnDeclaration("EXTENT_SIZE", types.EvalString),
		NewDynamicColumnDeclaration("AUTOEXTEND_SIZE", types.EvalString),
		NewDynamicColumnDeclaration("MAXIMUM_SIZE", types.EvalString),
		NewDynamicColumnDeclaration("NODEGROUP_ID", types.EvalString),
		NewDynamicColumnDeclaration("TABLESPACE_COMMENT", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTableConstraintsTable(d Database) error {
	tableName := TableConstraintsTable
	// All of the columns in this table are type varchar, so we can just
	// declare the names, and generate the decls from the names rather than
	// vice versa.
	columnNames := []string{
		"CONSTRAINT_CATALOG",
		"CONSTRAINT_SCHEMA",
		"CONSTRAINT_NAME",
		"TABLE_SCHEMA",
		"TABLE_NAME",
		"CONSTRAINT_TYPE",
	}

	t := NewDynamicTable(tableName, SystemView, func(aliasName string) results.RowIter {
		return b.getRowsForTableType(tableName, aliasName, columnNames)
	})

	columnDecls := make([]DynamicColumnDeclaration, len(columnNames))
	for i := range columnNames {
		columnDecls[i] = NewDynamicColumnDeclaration(columnNames[i], types.EvalString)
	}

	t.AddColumns(tableName,
		columnDecls...,
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTablePrivilegesTable(d Database) error {
	t := NewDynamicTable(TablePrivilagesTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(TablePrivilagesTable,
		NewDynamicColumnDeclaration("GRANTEE", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("PRIVILEGE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("IS_GRANTABLE", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTriggersTable(d Database) error {
	t := NewDynamicTable(TriggersTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(TriggersTable,
		NewDynamicColumnDeclaration("TRIGGER_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TRIGGER_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TRIGGER_NAME", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_MANIPULATION", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_OBJECT_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_OBJECT_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("EVENT_OBJECT_TABLE", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_ORDER", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_CONDITION", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_STATEMENT", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_ORIENTATION", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_TIMING", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_REFERENCE_OLD_TABLE", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_REFERENCE_NEW_TABLE", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_REFERENCE_OLD_ROW", types.EvalString),
		NewDynamicColumnDeclaration("ACTION_REFERENCE_NEW_ROW", types.EvalString),
		NewDynamicColumnDeclaration("CREATED", types.EvalString),
		NewDynamicColumnDeclaration("SQL_MODE", types.EvalString),
		NewDynamicColumnDeclaration("DEFINER", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_SET_CLIENT", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION_CONNECTION", types.EvalString),
		NewDynamicColumnDeclaration("DATABASE_COLLATION", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addUserPrivilegesTable(d Database) error {
	t := NewDynamicTable(UserPrivilegesTable, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(UserPrivilegesTable,
		NewDynamicColumnDeclaration("GRANTEE", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("PRIVILEGE_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("IS_GRANTABLE", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addViewsTable(d Database) error {
	views := "VIEWS"
	t := NewDynamicTable(views, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(views,
		NewDynamicColumnDeclaration("TABLE_CATALOG", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_SCHEMA", types.EvalString),
		NewDynamicColumnDeclaration("TABLE_NAME", types.EvalString),
		NewDynamicColumnDeclaration("VIEW_DEFINITION", types.EvalString),
		NewDynamicColumnDeclaration("CHECK_OPTION", types.EvalString),
		NewDynamicColumnDeclaration("IS_UPDATABLE", types.EvalString),
		NewDynamicColumnDeclaration("DEFINER", types.EvalString),
		NewDynamicColumnDeclaration("SECURITY_TYPE", types.EvalString),
		NewDynamicColumnDeclaration("CHARACTER_SET_CLIENT", types.EvalString),
		NewDynamicColumnDeclaration("COLLATION_CONNECTION", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addVariableTables(d Database) error {
	err := b.addVariableTable(d, GlobalVariablesTable, variable.GlobalScope, variable.SystemKind)
	if err != nil {
		return err
	}
	err = b.addVariableTable(d, GlobalStatusTable, variable.GlobalScope, variable.StatusKind)
	if err != nil {
		return err
	}

	err = b.addVariableTable(d, SessionVariablesTable, variable.SessionScope, variable.SystemKind)
	if err != nil {
		return err
	}
	return b.addVariableTable(d, SessionStatusTable, variable.SessionScope, variable.StatusKind)
}

func (b *catalogBuilder) addVariableTable(d Database, name string,
	scope variable.Scope, kind variable.Kind) error {
	variableName := "VARIABLE_NAME"
	variableValue := "VARIABLE_VALUE"

	t := NewDynamicTable(name, SystemView, func(aliasName string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})
		_, strv, _ := getValueCreators(b.variables)
		go func() {
			defer close(rowChan)
			for _, v := range b.variables.List(scope, kind) {
				select {
				// TODO: display actual value instead of nullv!!!
				case rowChan <- newInfoRow(aliasName,
					strv(variableName, v.Name),
					values.NewNamedSQLValue(variableValue, v.Value)):
				case <-done:
					return
				}
			}
		}()
		return results.NewRowChanIter(rowChan, done)
	})

	t.AddColumns(name,
		NewDynamicColumnDeclaration(variableName, types.EvalString),
		NewDynamicColumnDeclaration(variableValue, types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) buildMySQLDatabase() error {
	d, err := b.catalog.AddDatabase("mysql")
	if err != nil {
		return err
	}
	return b.addMySQLProcTable(d)
}

func (b *catalogBuilder) addMySQLProcTable(d Database) error {
	proc := "proc"
	t := NewDynamicTable(proc, SystemView, results.NewEmptyRowChanIter)

	t.AddColumns(proc,
		NewDynamicColumnDeclaration("db", types.EvalString),
		NewDynamicColumnDeclaration("name", types.EvalString),
		NewDynamicColumnDeclaration("type", types.EvalString),
		NewDynamicColumnDeclaration("specific_name", types.EvalString),
		NewDynamicColumnDeclaration("language", types.EvalString),
		NewDynamicColumnDeclaration("sql_data_access", types.EvalString),
		NewDynamicColumnDeclaration("is_deterministic", types.EvalString),
		NewDynamicColumnDeclaration("security_type", types.EvalString),
		NewDynamicColumnDeclaration("param_list", types.EvalString),
		NewDynamicColumnDeclaration("returns", types.EvalString),
		NewDynamicColumnDeclaration("body", types.EvalString),
		NewDynamicColumnDeclaration("definer", types.EvalString),
		NewDynamicColumnDeclaration("created", types.EvalDatetime),
		NewDynamicColumnDeclaration("modified", types.EvalDatetime),
		NewDynamicColumnDeclaration("sql_mode", types.EvalString),
		NewDynamicColumnDeclaration("comment", types.EvalString),
		NewDynamicColumnDeclaration("character_set_client", types.EvalString),
		NewDynamicColumnDeclaration("collation_connection", types.EvalString),
		NewDynamicColumnDeclaration("db_collation", types.EvalString),
		NewDynamicColumnDeclaration("body_utf8", types.EvalString),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) getRowsForTableType(systemTableName string, aliasName string, columnNames []string) results.RowIter {
	c := b.catalog

	ck := key{
		catalog: string(c.Name),
	}

	rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
	done := make(chan struct{})

	go func() {
		defer close(rowChan)
		for _, db := range c.databases {
			ck.database = string(db.Name())

			tbls, _ := db.Tables(context.Background())
			for _, tb := range tbls {
				ck.table = tb.Name()

				primaryKeyRowIter := b.getRowIterForPrimaryKey(systemTableName, aliasName, ck, tb.PrimaryKeys(), columnNames)
				row := results.Row{}
				for primaryKeyRowIter.Next(context.Background(), &row) {
					select {
					case rowChan <- row:
					case <-done:
						_ = primaryKeyRowIter.Close()
						return
					}
				}
				_ = primaryKeyRowIter.Close()

				mongoTable, ok := tb.(*MongoTable)
				if !ok {
					continue
				}

				uniqueKeyRowIter := b.getRowIterForUniqueIndexes(systemTableName, aliasName, ck, mongoTable.indexes, columnNames)
				for uniqueKeyRowIter.Next(context.Background(), &row) {
					select {
					case rowChan <- row:
					case <-done:
						_ = uniqueKeyRowIter.Close()
						return
					}
				}
				_ = uniqueKeyRowIter.Close()

				for _, fk := range mongoTable.foreignKeys {
					for position, col := range fk.columns {
						ck.column = col.Name
						foreignKeyRow := b.getRowForForeignKey(systemTableName, aliasName, ck, fk, position+1, columnNames)
						select {
						case rowChan <- foreignKeyRow:
						case <-done:
							return
						}
						// The break below occurs because only the name of the
						// constraint needs to be recorded in the table constraints;
						// it does not need to include every column as a row.
						if systemTableName == TableConstraintsTable || systemTableName == ReferentialConstraintsTable {
							break
						}
					}
				}
			}
		}
	}()
	return results.NewRowChanIter(rowChan, done)
}

// collationFromMongoDB creates a Collation from a mongodb.Collation.
func collationFromMongoDB(mc *mongodb.Collation) (*collation.Collation, error) {
	t, err := language.Parse(mc.Locale)
	if err != nil {
		return nil, fmt.Errorf("unable to translate locale (%s): %v", mc.Locale, err)
	}

	if mc.Alternate != "non-ignorable" {
		// Go's collate package doesn't support non-ignorable

		t, err = t.SetTypeForKey("ka", mc.Alternate)
		if err != nil {
			return nil, fmt.Errorf("unable to translate alternate (%s): %v", mc.Alternate, err)
		}
	}

	t, err = t.SetTypeForKey("kb", strconv.FormatBool(mc.Backwards))
	if err != nil {
		return nil, fmt.Errorf("unable to translate backwards (%v): %v", mc.Backwards, err)
	}

	t, err = t.SetTypeForKey("kc", strconv.FormatBool(mc.CaseLevel))
	if err != nil {
		return nil, fmt.Errorf("unable to translate caseLevel (%v): %v", mc.CaseLevel, err)
	}

	t, err = t.SetTypeForKey("kf", mc.CaseFirst)
	if err != nil {
		return nil, fmt.Errorf("unable to translate caseFirst (%s): %v", mc.CaseFirst, err)
	}

	t, err = t.SetTypeForKey("kk", strconv.FormatBool(mc.Normalization))
	if err != nil {
		return nil, fmt.Errorf("unable to translate normalization (%v): %v", mc.Normalization, err)
	}

	t, err = t.SetTypeForKey("kn", strconv.FormatBool(mc.NumericOrdering))
	if err != nil {
		return nil, fmt.Errorf(
			"unable to translate numeric ordering (%v): %v",
			mc.NumericOrdering, err,
		)
	}

	t, err = t.SetTypeForKey("kv", mc.MaxVariable)
	if err != nil {
		return nil, fmt.Errorf("unable to translate maxVariable (%s): %v", mc.MaxVariable, err)
	}

	if mc.Strength > 0 {
		var value string
		switch mc.Strength {
		case 1:
			value = "level1"
		case 2:
			value = "level2"
		case 0, 3:
			value = "level3"
		case 4:
			value = "level4"
		case 5:
			value = "identic"
		}
		t, err = t.SetTypeForKey("ks", value)
		if err != nil {
			return nil, fmt.Errorf("unable to translate strength (%v): %v", mc.Strength, err)
		}
	}

	return collation.NewCollation(
		t,
		collate.New(t),
		mc.Strength < 3 && mc.CaseLevel,
		mc.Strength == 1,
		mc.Strength == 3 && mc.CaseLevel,
		collation.CharsetName("utf8"),
		8,
	), nil
}
