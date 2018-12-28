package catalog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// Build builds a catalog up from a schema and variables.
func Build(schema *schema.Schema, variables VariableContainer, info *mongodb.Info) (*SQLCatalog, error) {
	alteredSchema, err := schema.Altered()
	if err != nil {
		return nil, err
	}

	builder := &catalogBuilder{
		catalog:   New("def", variables),
		schema:    alteredSchema,
		variables: variables,
		info:      info,
	}

	err = builder.build()
	if err != nil {
		return nil, err
	}
	return builder.catalog, nil
}

type catalogBuilder struct {
	catalog   *SQLCatalog
	info      *mongodb.Info
	schema    *schema.Schema
	variables VariableContainer
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
	for _, dbConfig := range b.schema.DatabasesSorted() {
		if !info.IsVisibleDatabase(mongodb.DatabaseName(dbConfig.Name())) {
			b.catalog.containsAuthRestrictedNamespaces = true
			continue
		}

		d, err := b.catalog.Database(dbConfig.Name())
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

		for _, tblConfig := range dbConfig.TablesSorted() {
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

			t := NewMongoTable(tblConfig, tableType, col)

			t.isSharded = collection.IsSharded

			mongoNameToColumn := make(map[string]Column)

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

func (b *catalogBuilder) addCharsetTable(d Database) error {
	t := NewDynamicTable(CharacterSetsTable, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, c := range collation.GetAllCharsets() {
			rows = append(rows, NewDataRow(
				string(c.Name),
				string(c.DefaultCollationName),
				c.Description,
				int(c.MaxLen),
			))
		}
		return rows
	})

	t.AddColumns(
		"CHARACTER_SET_NAME", string(schema.SQLVarchar),
		"DEFAULT_COLLATE_NAME", string(schema.SQLVarchar),
		"DESCRIPTION", string(schema.SQLVarchar),
		"MAXLEN", string(schema.SQLInt),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addCollationTable(d Database) error {
	t := NewDynamicTable(CollationsTable, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, c := range collation.GetAll() {
			isDefault := "No"

			if c.Default {
				isDefault = "Yes"
			}

			rows = append(rows, NewDataRow(
				string(c.Name),
				string(c.CharsetName),
				int(c.ID),
				isDefault,
				"Yes",
				int(c.SortLen),
			))
		}
		return rows
	})

	t.AddColumns(
		"COLLATION_NAME", string(schema.SQLVarchar),
		"CHARACTER_SET_NAME", string(schema.SQLVarchar),
		"ID", string(schema.SQLInt),
		"IS_DEFAULT", string(schema.SQLVarchar),
		"IS_COMPILED", string(schema.SQLVarchar),
		"SORTLEN", string(schema.SQLInt),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addCollationCharacterSetApplicabilityTable(d Database) error {
	t := NewDynamicTable(CollationCharacterSetApplicabilityTable, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, c := range collation.GetAll() {
			rows = append(rows, NewDataRow(
				string(c.Name),
				string(c.CharsetName),
			))
		}
		return rows
	})

	t.AddColumns(
		"COLLATION_NAME", string(schema.SQLVarchar),
		"CHARACTER_SET_NAME", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addColumnsTable(d Database) error {
	c := b.catalog
	t := NewDynamicTable(ColumnsTable, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, db := range c.Databases() {
			for _, tbl := range db.Tables() {
				for i, col := range tbl.Columns() {
					columnKey := getIndexKey(col, tbl)
					maxVarcharLength := b.variables.GetUint16(variable.MongoDBMaxVarcharLength)
					columnType := translateColumnType(col.Type(), maxVarcharLength)
					dataType := columnType
					if idx := strings.Index(dataType, "("); idx >= 0 {
						dataType = dataType[:idx]
					}
					rows = append(rows, NewDataRow(
						string(c.Name),
						string(db.Name()),
						string(tbl.Name()),
						string(col.Name()),
						i+1,
						nil,
						"YES",
						dataType,
						nil,
						nil,
						nil,
						nil,
						nil,
						string(tbl.Collation().CharsetName),
						string(tbl.Collation().Name),
						columnType,
						columnKey,
						"",
						"select",
						col.Comments(),
						"",
					))
				}
			}
		}
		return rows
	})

	t.AddColumns(
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"COLUMN_NAME", string(schema.SQLVarchar),
		"ORDINAL_POSITION", string(schema.SQLInt),
		"COLUMN_DEFAULT", string(schema.SQLVarchar),
		"IS_NULLABLE", string(schema.SQLVarchar),
		"DATA_TYPE", string(schema.SQLVarchar),
		"CHARACTER_MAXIMUM_LENGTH", string(schema.SQLInt),
		"CHARACTER_OCTET_LENGTH", string(schema.SQLInt),
		"NUMERIC_PRECISION", string(schema.SQLInt),
		"NUMERIC_SCALE", string(schema.SQLInt),
		"DATETIME_PRECISION", string(schema.SQLInt),
		"CHARACTER_SET_NAME", string(schema.SQLVarchar),
		"COLLATION_NAME", string(schema.SQLVarchar),
		"COLUMN_TYPE", string(schema.SQLVarchar),
		"COLUMN_KEY", string(schema.SQLVarchar),
		"EXTRA", string(schema.SQLVarchar),
		"PRIVILEGES", string(schema.SQLVarchar),
		"COLUMN_COMMENT", string(schema.SQLVarchar),
		"GENERATION_EXPRESSION", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addColumnPrivilegesTable(d Database) error {
	t := NewDynamicTable(ColumnPrivilegesTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"GRANTEE", string(schema.SQLVarchar),
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"COLUMN_NAME", string(schema.SQLVarchar),
		"PRIVILEGE_TYPE", string(schema.SQLVarchar),
		"IS_GRANTABLE", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addEnginesTable(d Database) error {
	t := NewDynamicTable(EnginesTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"ENGINE", string(schema.SQLVarchar),
		"SUPPORT", string(schema.SQLVarchar),
		"COMMENT", string(schema.SQLVarchar),
		"TRANSACTIONS", string(schema.SQLVarchar),
		"XA", string(schema.SQLVarchar),
		"SAVEPOINTS", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addEventsTable(d Database) error {
	t := NewDynamicTable(EventsTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"EVENT_CATALOG", string(schema.SQLVarchar),
		"EVENT_SCHEMA", string(schema.SQLVarchar),
		"EVENT_NAME", string(schema.SQLVarchar),
		"DEFINER", string(schema.SQLVarchar),
		"TIME_ZONE", string(schema.SQLVarchar),
		"EVENT_BODY", string(schema.SQLVarchar),
		"EVENT_DEFINITION", string(schema.SQLVarchar),
		"EVENT_TYPE", string(schema.SQLVarchar),
		"EXECUTE_AT", string(schema.SQLVarchar),
		"INTERVAL_VALUE", string(schema.SQLVarchar),
		"INTERVAL_FIELD", string(schema.SQLVarchar),
		"SQL_MODE", string(schema.SQLVarchar),
		"STARTS", string(schema.SQLVarchar),
		"ENDS", string(schema.SQLVarchar),
		"STATUS", string(schema.SQLVarchar),
		"ON_COMPLETION", string(schema.SQLVarchar),
		"CREATED", string(schema.SQLVarchar),
		"LAST_ALTERED", string(schema.SQLVarchar),
		"LAST_EXECUTED", string(schema.SQLVarchar),
		"EVENT_COMMENT", string(schema.SQLVarchar),
		"ORIGINATOR", string(schema.SQLVarchar),
		"CHARACTER_SET_CLIENT", string(schema.SQLVarchar),
		"COLLATION_CONNECTION", string(schema.SQLVarchar),
		"DATABASE_COLLATION", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addFilesTable(d Database) error {
	t := NewDynamicTable("FILES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"FILE_ID", string(schema.SQLVarchar),
		"FILE_NAME", string(schema.SQLVarchar),
		"FILE_TYPE", string(schema.SQLVarchar),
		"TABLESPACE_NAME", string(schema.SQLVarchar),
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"LOGFILE_GROUP_NAME", string(schema.SQLVarchar),
		"LOGFILE_GROUP_NUMBER", string(schema.SQLVarchar),
		"ENGINE", string(schema.SQLVarchar),
		"FULLTEXT_KEYS", string(schema.SQLVarchar),
		"DELETED_ROWS", string(schema.SQLVarchar),
		"UPDATE_COUNT", string(schema.SQLVarchar),
		"FREE_EXTENTS", string(schema.SQLVarchar),
		"TOTAL_EXTENTS", string(schema.SQLVarchar),
		"EXTENT_SIZE", string(schema.SQLVarchar),
		"INITIAL_SIZE", string(schema.SQLVarchar),
		"MAXIMUM_SIZE", string(schema.SQLVarchar),
		"AUTOEXTEND_SIZE", string(schema.SQLVarchar),
		"CREATION_TIME", string(schema.SQLVarchar),
		"LAST_UPDATE_TIME", string(schema.SQLVarchar),
		"LAST_ACCESS_TIME", string(schema.SQLVarchar),
		"RECOVER_TIME", string(schema.SQLVarchar),
		"TRANSACTION_COUNTER", string(schema.SQLVarchar),
		"VERSION", string(schema.SQLVarchar),
		"ROW_FORMAT", string(schema.SQLVarchar),
		"TABLE_ROWS", string(schema.SQLVarchar),
		"AVG_ROW_LENGTH", string(schema.SQLVarchar),
		"DATA_LENGTH", string(schema.SQLVarchar),
		"MAX_DATA_LENGTH", string(schema.SQLVarchar),
		"INDEX_LENGTH", string(schema.SQLVarchar),
		"DATA_FREE", string(schema.SQLVarchar),
		"CREATE_TIME", string(schema.SQLVarchar),
		"UPDATE_TIME", string(schema.SQLVarchar),
		"CHECK_TIME", string(schema.SQLVarchar),
		"CHECKSUM", string(schema.SQLVarchar),
		"STATUS", string(schema.SQLVarchar),
		"EXTRA", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) getTableFromNamespace(ns namespace) (Table, error) {
	currentDb, err := b.catalog.Database(ns.database)
	if err != nil {
		return nil, err
	}

	return currentDb.Table(ns.table)
}

func (b *catalogBuilder) addKeyColumnUsageTable(d Database) error {
	t := NewDynamicTable(KeyColumnUsageTable, SystemView, func() []*DataRow {
		return b.getDataRowsForTableType(KeyColumnUsageTable)
	})

	t.AddColumns(
		"CONSTRAINT_CATALOG", string(schema.SQLVarchar),
		"CONSTRAINT_SCHEMA", string(schema.SQLVarchar),
		"CONSTRAINT_NAME", string(schema.SQLVarchar),
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"COLUMN_NAME", string(schema.SQLVarchar),
		"ORDINAL_POSITION", string(schema.SQLVarchar),
		"POSITION_IN_UNIQUE_CONSTRAINT", string(schema.SQLVarchar),
		"REFERENCED_TABLE_SCHEMA", string(schema.SQLVarchar),
		"REFERENCED_TABLE_NAME", string(schema.SQLVarchar),
		"REFERENCED_COLUMN_NAME", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addNdbTransidMysqlConnectionMapTable(d Database) error {
	t := NewDynamicTable(NdbTransidMysqlConnectionMapTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"mysql_connection_id", string(schema.SQLVarchar),
		"node_id", string(schema.SQLVarchar),
		"ndb_transid", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addParametersTable(d Database) error {
	t := NewDynamicTable(ParametersTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"SPECIFIC_CATALOG", string(schema.SQLVarchar),
		"SPECIFIC_SCHEMA", string(schema.SQLVarchar),
		"SPECIFIC_NAME", string(schema.SQLVarchar),
		"ORDINAL_POSITION", string(schema.SQLVarchar),
		"PARAMETER_MODE", string(schema.SQLVarchar),
		"PARAMETER_NAME", string(schema.SQLVarchar),
		"DATA_TYPE", string(schema.SQLVarchar),
		"CHARACTER_MAXIMUM_LENGTH", string(schema.SQLVarchar),
		"CHARACTER_OCTET_LENGTH", string(schema.SQLVarchar),
		"NUMERIC_PRECISION", string(schema.SQLVarchar),
		"NUMERIC_SCALE", string(schema.SQLVarchar),
		"DATETIME_PRECISION", string(schema.SQLVarchar),
		"CHARACTER_SET_NAME", string(schema.SQLVarchar),
		"COLLATION_NAME", string(schema.SQLVarchar),
		"DTD_IDENTIFIER", string(schema.SQLVarchar),
		"ROUTINE_TYPE", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addPartitionsTable(d Database) error {
	t := NewDynamicTable(PartitionsTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"PARTITION_NAME", string(schema.SQLVarchar),
		"SUBPARTITION_NAME", string(schema.SQLVarchar),
		"PARTITION_ORDINAL_POSITION", string(schema.SQLVarchar),
		"SUBPARTITION_ORDINAL_POSITION", string(schema.SQLVarchar),
		"PARTITION_METHOD", string(schema.SQLVarchar),
		"SUBPARTITION_METHOD", string(schema.SQLVarchar),
		"PARTITION_EXPRESSION", string(schema.SQLVarchar),
		"SUBPARTITION_EXPRESSION", string(schema.SQLVarchar),
		"PARTITION_DESCRIPTION", string(schema.SQLVarchar),
		"TABLE_ROWS", string(schema.SQLVarchar),
		"AVG_ROW_LENGTH", string(schema.SQLVarchar),
		"DATA_LENGTH", string(schema.SQLVarchar),
		"MAX_DATA_LENGTH", string(schema.SQLVarchar),
		"INDEX_LENGTH", string(schema.SQLVarchar),
		"DATA_FREE", string(schema.SQLVarchar),
		"CREATE_TIME", string(schema.SQLVarchar),
		"UPDATE_TIME", string(schema.SQLVarchar),
		"CHECK_TIME", string(schema.SQLVarchar),
		"CHECKSUM", string(schema.SQLVarchar),
		"PARTITION_COMMENT", string(schema.SQLVarchar),
		"NODEGROUP", string(schema.SQLVarchar),
		"TABLESPACE_NAME", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addPluginsTable(d Database) error {
	t := NewDynamicTable(PluginsTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"PLUGIN_NAME", string(schema.SQLVarchar),
		"PLUGIN_VERSION", string(schema.SQLVarchar),
		"PLUGIN_STATUS", string(schema.SQLVarchar),
		"PLUGIN_TYPE", string(schema.SQLVarchar),
		"PLUGIN_TYPE_VERSION", string(schema.SQLVarchar),
		"PLUGIN_LIBRARY", string(schema.SQLVarchar),
		"PLUGIN_LIBRARY_VERSION", string(schema.SQLVarchar),
		"PLUGIN_AUTHOR", string(schema.SQLVarchar),
		"PLUGIN_DESCRIPTION", string(schema.SQLVarchar),
		"PLUGIN_LICENSE", string(schema.SQLVarchar),
		"LOAD_OPTION", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addProfilingTable(d Database) error {
	t := NewDynamicTable(ProfilingTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"QUERY_ID", string(schema.SQLVarchar),
		"SEQ", string(schema.SQLVarchar),
		"STATE", string(schema.SQLVarchar),
		"DURATION", string(schema.SQLVarchar),
		"CPU_USER", string(schema.SQLVarchar),
		"CPU_SYSTEM", string(schema.SQLVarchar),
		"CONTEXT_VOLUNTARY", string(schema.SQLVarchar),
		"CONTEXT_INVOLUNTARY", string(schema.SQLVarchar),
		"BLOCK_OPS_IN", string(schema.SQLVarchar),
		"BLOCK_OPS_OUT", string(schema.SQLVarchar),
		"MESSAGES_SENT", string(schema.SQLVarchar),
		"MESSAGES_RECEIVED", string(schema.SQLVarchar),
		"PAGE_FAULTS_MAJOR", string(schema.SQLVarchar),
		"PAGE_FAULTS_MINOR", string(schema.SQLVarchar),
		"SWAPS", string(schema.SQLVarchar),
		"SOURCE_FUNCTION", string(schema.SQLVarchar),
		"SOURCE_FILE", string(schema.SQLVarchar),
		"SOURCE_LINE", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addReferentialConstraintsTable(d Database) error {
	t := NewDynamicTable(ReferentialConstraintsTable, SystemView, func() []*DataRow {
		return b.getDataRowsForTableType(ReferentialConstraintsTable)
	})

	t.AddColumns(
		"CONSTRAINT_CATALOG", string(schema.SQLVarchar),
		"CONSTRAINT_SCHEMA", string(schema.SQLVarchar),
		"CONSTRAINT_NAME", string(schema.SQLVarchar),
		"UNIQUE_CONSTRAINT_CATALOG", string(schema.SQLVarchar),
		"UNIQUE_CONSTRAINT_SCHEMA", string(schema.SQLVarchar),
		"UNIQUE_CONSTRAINT_NAME", string(schema.SQLVarchar),
		"MATCH_OPTION", string(schema.SQLVarchar),
		"UPDATE_RULE", string(schema.SQLVarchar),
		"DELETE_RULE", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"REFERENCED_TABLE_NAME", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addRoutinesTable(d Database) error {
	t := NewDynamicTable(RoutinesTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"SPECIFIC_NAME", string(schema.SQLVarchar),
		"ROUTINE_CATALOG", string(schema.SQLVarchar),
		"ROUTINE_SCHEMA", string(schema.SQLVarchar),
		"ROUTINE_NAME", string(schema.SQLVarchar),
		"ROUTINE_TYPE", string(schema.SQLVarchar),
		"DATA_TYPE", string(schema.SQLVarchar),
		"CHARACTER_MAXIMUM_LENGTH", string(schema.SQLVarchar),
		"CHARACTER_OCTET_LENGTH", string(schema.SQLVarchar),
		"NUMERIC_PRECISION", string(schema.SQLVarchar),
		"NUMERIC_SCALE", string(schema.SQLVarchar),
		"DATETIME_PRECISION", string(schema.SQLVarchar),
		"CHARACTER_SET_NAME", string(schema.SQLVarchar),
		"COLLATION_NAME", string(schema.SQLVarchar),
		"DTD_IDENTIFIER", string(schema.SQLVarchar),
		"ROUTINE_BODY", string(schema.SQLVarchar),
		"ROUTINE_DEFINITION", string(schema.SQLVarchar),
		"EXTERNAL_NAME", string(schema.SQLVarchar),
		"EXTERNAL_LANGUAGE", string(schema.SQLVarchar),
		"PARAMETER_STYLE", string(schema.SQLVarchar),
		"IS_DETERMINISTIC", string(schema.SQLVarchar),
		"SQL_DATA_ACCESS", string(schema.SQLVarchar),
		"SQL_PATH", string(schema.SQLVarchar),
		"SECURITY_TYPE", string(schema.SQLVarchar),
		"CREATED", string(schema.SQLVarchar),
		"LAST_ALTERED", string(schema.SQLVarchar),
		"SQL_MODE", string(schema.SQLVarchar),
		"ROUTINE_COMMENT", string(schema.SQLVarchar),
		"DEFINER", string(schema.SQLVarchar),
		"CHARACTER_SET_CLIENT", string(schema.SQLVarchar),
		"COLLATION_CONNECTION", string(schema.SQLVarchar),
		"DATABASE_COLLATION", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addSchemataTable(d Database) error {
	c := b.catalog
	t := NewDynamicTable(SchemataTable, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, db := range c.Databases() {
			rows = append(rows, NewDataRow(
				string(c.Name),
				string(db.Name()),
				string(collation.DefaultCharset.Name),
				string(collation.Default.Name),
				"",
			))
		}
		return rows
	})

	t.AddColumns(
		"CATALOG_NAME", string(schema.SQLVarchar),
		"SCHEMA_NAME", string(schema.SQLVarchar),
		"DEFAULT_CHARACTER_SET_NAME", string(schema.SQLVarchar),
		"DEFAULT_COLLATION_NAME", string(schema.SQLVarchar),
		"SQL_PATH", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addSchemaPrivileges(d Database) error {
	t := NewDynamicTable(SchemaPrivilagesTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"GRANTEE", string(schema.SQLVarchar),
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"PRIVILEGE_TYPE", string(schema.SQLVarchar),
		"IS_GRANTABLE", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addStatisticsTable(d Database) error {
	t := NewDynamicTable(StatisticsTable, SystemView, func() []*DataRow {
		return b.getDataRowsForTableType(StatisticsTable)
	})

	t.AddColumns(
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"NON_UNIQUE", string(schema.SQLVarchar),
		"INDEX_SCHEMA", string(schema.SQLVarchar),
		"INDEX_NAME", string(schema.SQLVarchar),
		"SEQ_IN_INDEX", string(schema.SQLVarchar),
		"COLUMN_NAME", string(schema.SQLVarchar),
		"COLLATION", string(schema.SQLVarchar),
		"CARDINALITY", string(schema.SQLVarchar),
		"SUB_PART", string(schema.SQLVarchar),
		"PACKED", string(schema.SQLVarchar),
		"NULLABLE", string(schema.SQLVarchar),
		"INDEX_TYPE", string(schema.SQLVarchar),
		"COMMENT", string(schema.SQLVarchar),
		"INDEX_COMMENT", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTablesTable(d Database) error {
	c := b.catalog
	t := NewDynamicTable(TablesTable, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, db := range c.Databases() {
			for _, tbl := range db.Tables() {
				rows = append(rows, NewDataRow(
					string(c.Name),
					string(db.Name()),
					string(tbl.Name()),
					string(tbl.Type()),
					"",
					"",
					"",
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					"NO",
					nil,
					nil,
					nil,
					string(tbl.Collation().Name),
					nil,
					nil,
					tbl.Comments(),
				))
			}
		}
		return rows
	})

	t.AddColumns(
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"TABLE_TYPE", string(schema.SQLVarchar),
		"ENGINE", string(schema.SQLVarchar),
		"VERSION", string(schema.SQLVarchar),
		"ROW_FORMAT", string(schema.SQLVarchar),
		"TABLE_ROWS", string(schema.SQLInt),
		"AVG_ROW_LENGTH", string(schema.SQLInt),
		"DATA_LENGTH", string(schema.SQLInt),
		"MAX_DATA_LENGTH", string(schema.SQLInt),
		"INDEX_LENGTH", string(schema.SQLInt),
		"DATA_FREE", string(schema.SQLInt),
		"AUTO_INCREMENT", string(schema.SQLVarchar),
		"CREATE_TIME", string(schema.SQLTimestamp),
		"UPDATE_TIME", string(schema.SQLTimestamp),
		"CHECK_TIME", string(schema.SQLTimestamp),
		"TABLE_COLLATION", string(schema.SQLVarchar),
		"CHECKSUM", string(schema.SQLVarchar),
		"CREATE_OPTIONS", string(schema.SQLVarchar),
		"TABLE_COMMENT", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTableSpacesTable(d Database) error {
	t := NewDynamicTable(TablespacesTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"TABLESPACE_NAME", string(schema.SQLVarchar),
		"ENGINE", string(schema.SQLVarchar),
		"TABLESPACE_TYPE", string(schema.SQLVarchar),
		"LOGFILE_GROUP_NAME", string(schema.SQLVarchar),
		"EXTENT_SIZE", string(schema.SQLVarchar),
		"AUTOEXTEND_SIZE", string(schema.SQLVarchar),
		"MAXIMUM_SIZE", string(schema.SQLVarchar),
		"NODEGROUP_ID", string(schema.SQLVarchar),
		"TABLESPACE_COMMENT", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTableConstraintsTable(d Database) error {
	t := NewDynamicTable(TableConstraintsTable, SystemView, func() []*DataRow {
		return b.getDataRowsForTableType(TableConstraintsTable)
	})

	t.AddColumns(
		"CONSTRAINT_CATALOG", string(schema.SQLVarchar),
		"CONSTRAINT_SCHEMA", string(schema.SQLVarchar),
		"CONSTRAINT_NAME", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"CONSTRAINT_TYPE", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTablePrivilegesTable(d Database) error {
	t := NewDynamicTable(TablePrivilagesTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"GRANTEE", string(schema.SQLVarchar),
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"PRIVILEGE_TYPE", string(schema.SQLVarchar),
		"IS_GRANTABLE", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTriggersTable(d Database) error {
	t := NewDynamicTable(TriggersTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"TRIGGER_CATALOG", string(schema.SQLVarchar),
		"TRIGGER_SCHEMA", string(schema.SQLVarchar),
		"TRIGGER_NAME", string(schema.SQLVarchar),
		"EVENT_MANIPULATION", string(schema.SQLVarchar),
		"EVENT_OBJECT_CATALOG", string(schema.SQLVarchar),
		"EVENT_OBJECT_SCHEMA", string(schema.SQLVarchar),
		"EVENT_OBJECT_TABLE", string(schema.SQLVarchar),
		"ACTION_ORDER", string(schema.SQLVarchar),
		"ACTION_CONDITION", string(schema.SQLVarchar),
		"ACTION_STATEMENT", string(schema.SQLVarchar),
		"ACTION_ORIENTATION", string(schema.SQLVarchar),
		"ACTION_TIMING", string(schema.SQLVarchar),
		"ACTION_REFERENCE_OLD_TABLE", string(schema.SQLVarchar),
		"ACTION_REFERENCE_NEW_TABLE", string(schema.SQLVarchar),
		"ACTION_REFERENCE_OLD_ROW", string(schema.SQLVarchar),
		"ACTION_REFERENCE_NEW_ROW", string(schema.SQLVarchar),
		"CREATED", string(schema.SQLVarchar),
		"SQL_MODE", string(schema.SQLVarchar),
		"DEFINER", string(schema.SQLVarchar),
		"CHARACTER_SET_CLIENT", string(schema.SQLVarchar),
		"COLLATION_CONNECTION", string(schema.SQLVarchar),
		"DATABASE_COLLATION", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addUserPrivilegesTable(d Database) error {
	t := NewDynamicTable(UserPrivilagesTable, SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"GRANTEE", string(schema.SQLVarchar),
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"PRIVILEGE_TYPE", string(schema.SQLVarchar),
		"IS_GRANTABLE", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) addViewsTable(d Database) error {
	t := NewDynamicTable("VIEWS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"TABLE_CATALOG", string(schema.SQLVarchar),
		"TABLE_SCHEMA", string(schema.SQLVarchar),
		"TABLE_NAME", string(schema.SQLVarchar),
		"VIEW_DEFINITION", string(schema.SQLVarchar),
		"CHECK_OPTION", string(schema.SQLVarchar),
		"IS_UPDATABLE", string(schema.SQLVarchar),
		"DEFINER", string(schema.SQLVarchar),
		"SECURITY_TYPE", string(schema.SQLVarchar),
		"CHARACTER_SET_CLIENT", string(schema.SQLVarchar),
		"COLLATION_CONNECTION", string(schema.SQLVarchar),
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

func (b *catalogBuilder) addVariableTable(d Database, name TableName,
	scope variable.Scope, kind variable.Kind) error {

	t := NewDynamicTable(name, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, v := range b.variables.List(scope, kind) {
			rows = append(rows, NewDataRow(v.Name, v.Value))
		}

		return rows
	})

	t.AddColumns(
		"VARIABLE_NAME", string(schema.SQLVarchar),
		"VARIABLE_VALUE", string(schema.SQLVarchar),
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
	t := NewDynamicTable("proc", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumns(
		"db", string(schema.SQLVarchar),
		"name", string(schema.SQLVarchar),
		"type", string(schema.SQLVarchar),
		"specific_name", string(schema.SQLVarchar),
		"language", string(schema.SQLVarchar),
		"sql_data_access", string(schema.SQLVarchar),
		"is_deterministic", string(schema.SQLVarchar),
		"security_type", string(schema.SQLVarchar),
		"param_list", string(schema.SQLVarchar),
		"returns", string(schema.SQLVarchar),
		"body", string(schema.SQLVarchar),
		"definer", string(schema.SQLVarchar),
		"created", string(schema.SQLTimestamp),
		"modified", string(schema.SQLTimestamp),
		"sql_mode", string(schema.SQLVarchar),
		"comment", string(schema.SQLVarchar),
		"character_set_client", string(schema.SQLVarchar),
		"collation_connection", string(schema.SQLVarchar),
		"db_collation", string(schema.SQLVarchar),
		"body_utf8", string(schema.SQLVarchar),
	)

	return d.AddTable(t)
}

func (b *catalogBuilder) getDataRowsForTableType(systemTableName TableName) []*DataRow {
	c := b.catalog

	var rows []*DataRow

	ck := key{
		catalog: string(c.Name),
	}

	for _, db := range c.databases {
		ck.database = string(db.Name())

		for _, tb := range db.Tables() {
			ck.table = string(tb.Name())

			primaryKeyRows := getDataRowsForPrimaryKey(systemTableName, ck, tb.PrimaryKeys())

			rows = append(rows, primaryKeyRows...)

			mongoTable, ok := tb.(*MongoTable)
			if !ok {
				continue
			}

			uniqueKeyRows := getDataRowsForUniqueIndexes(systemTableName, ck, mongoTable.indexes)

			rows = append(rows, uniqueKeyRows...)

			for _, fk := range mongoTable.foreignKeys {
				for position, col := range fk.columns {
					ck.column = string(col.Name())
					foreignKeyRow := getDataRowForForeignKey(systemTableName, ck, fk, position+1)
					rows = append(rows, foreignKeyRow)
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
	return rows
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
