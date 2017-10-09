package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

// Build builds a catalog up from a schema and variables.
func Build(schema *schema.Schema, variables *variable.Container) (*Catalog, error) {
	builder := &catalogBuilder{
		catalog:   New("def"),
		schema:    schema,
		variables: variables,
	}

	err := builder.build()
	if err != nil {
		return nil, err
	}
	return builder.catalog, nil
}

type catalogBuilder struct {
	catalog   *Catalog
	schema    *schema.Schema
	variables *variable.Container
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
	info := b.variables.MongoDBInfo
	for _, dbConfig := range b.schema.Databases {
		if !info.IsAnyAllowedDatabase(mongodb.DatabaseName(dbConfig.Name)) {
			continue
		}

		d, err := b.catalog.Database(dbConfig.Name)
		if err != nil {
			d, err = b.catalog.AddDatabase(dbConfig.Name)
			if err != nil {
				return err
			}
		}
		dInfo, ok := info.Databases[mongodb.DatabaseName(strings.ToLower(dbConfig.Name))]
		if !ok {
			continue
		}
		for _, tblConfig := range dbConfig.Tables {
			if !info.IsAnyAllowedCollection(mongodb.DatabaseName(dbConfig.Name), mongodb.CollectionName(tblConfig.CollectionName)) {
				continue
			}

			collection, ok := dInfo.Collections[mongodb.CollectionName(tblConfig.CollectionName)]
			if !ok {
				continue
			}

			col := collation.Default
			if collection.Collation != nil {
				col, err = collation.FromMongoDB(collection.Collation)
				if err != nil {
					return mysqlerrors.Newf(mysqlerrors.ER_UNKNOWN_COLLATION, "unable to translate MongoDB's collation for \"%s\".\"%s\": %v", dbConfig.Name, tblConfig.Name, err)
				}
			}
			tableType := BaseTable
			if collection.IsView {
				tableType = View
			}
			t := NewMongoTable(tblConfig, tableType, col)
			err = d.AddTable(t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *catalogBuilder) buildInformationSchemaDatabase() error {
	d, _ := b.catalog.AddDatabase(InformationSchemaDatabase)
	err := b.addCharsetTable(d)
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
	err = b.addProcessListTable(d)
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
	err = b.addVariableTables(d)
	if err != nil {
		return err
	}
	return nil
}

func (b *catalogBuilder) addCharsetTable(d *Database) error {
	t := NewDynamicTable("CHARACTER_SETS", SystemView, func() []*DataRow {
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

	t.AddColumn("CHARACTER_SET_NAME", schema.SQLVarchar)
	t.AddColumn("DEFAULT_COLLATE_NAME", schema.SQLVarchar)
	t.AddColumn("DESCRIPTION", schema.SQLVarchar)
	t.AddColumn("MAXLEN", schema.SQLInt)

	return d.AddTable(t)
}

func (b *catalogBuilder) addCollationTable(d *Database) error {
	t := NewDynamicTable("COLLATIONS", SystemView, func() []*DataRow {
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

	t.AddColumn("COLLATION_NAME", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_NAME", schema.SQLVarchar)
	t.AddColumn("ID", schema.SQLInt)
	t.AddColumn("IS_DEFAULT", schema.SQLVarchar)
	t.AddColumn("IS_COMPILED", schema.SQLVarchar)
	t.AddColumn("SORTLEN", schema.SQLInt)

	return d.AddTable(t)
}

func (b *catalogBuilder) addCollationCharacterSetApplicabilityTable(d *Database) error {
	t := NewDynamicTable("COLLATION_CHARACTER_SET_APPLICABILITY", SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, c := range collation.GetAll() {
			rows = append(rows, NewDataRow(
				string(c.Name),
				string(c.CharsetName),
			))
		}
		return rows
	})

	t.AddColumn("COLLATION_NAME", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_NAME", schema.SQLVarchar)
	return d.AddTable(t)
}

func (b *catalogBuilder) addColumnsTable(d *Database) error {
	c := b.catalog
	t := NewDynamicTable("COLUMNS", SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, db := range c.Databases() {
			for _, tbl := range db.Tables() {
				for i, col := range tbl.Columns() {
					columnType := translateColumnType(col.Type(), b.variables.MongoDBMaxVarcharLength)
					dataType := columnType
					if idx := strings.Index(dataType, "("); idx >= 0 {
						dataType = dataType[:idx]
					}
					rows = append(rows, NewDataRow(
						string(c.Name),
						string(db.Name),
						string(tbl.Name()),
						string(col.Name()),
						i,
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
						"",
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

	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("COLUMN_NAME", schema.SQLVarchar)
	t.AddColumn("ORDINAL_POSITION", schema.SQLInt)
	t.AddColumn("COLUMN_DEFAULT", schema.SQLVarchar)
	t.AddColumn("IS_NULLABLE", schema.SQLVarchar)
	t.AddColumn("DATA_TYPE", schema.SQLVarchar)
	t.AddColumn("CHARACTER_MAXIMUM_LENGTH", schema.SQLInt)
	t.AddColumn("CHARACTER_OCTET_LENGTH", schema.SQLInt)
	t.AddColumn("NUMERIC_PRECISION", schema.SQLInt)
	t.AddColumn("NUMERIC_SCALE", schema.SQLInt)
	t.AddColumn("DATETIME_PRECISION", schema.SQLInt)
	t.AddColumn("CHARACTER_SET_NAME", schema.SQLVarchar)
	t.AddColumn("COLLATION_NAME", schema.SQLVarchar)
	t.AddColumn("COLUMN_TYPE", schema.SQLVarchar)
	t.AddColumn("COLUMN_KEY", schema.SQLVarchar)
	t.AddColumn("EXTRA", schema.SQLVarchar)
	t.AddColumn("PRIVILEGES", schema.SQLVarchar)
	t.AddColumn("COLUMN_COMMENT", schema.SQLVarchar)
	t.AddColumn("GENERATION_EXPRESSION", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addColumnPrivilegesTable(d *Database) error {
	t := NewDynamicTable("COLUMN_PRIVILEGES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("GRANTEE", schema.SQLVarchar)
	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("COLUMN_NAME", schema.SQLVarchar)
	t.AddColumn("PRIVILEGE_TYPE", schema.SQLVarchar)
	t.AddColumn("IS_GRANTABLE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addEnginesTable(d *Database) error {
	t := NewDynamicTable("ENGINES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("ENGINE", schema.SQLVarchar)
	t.AddColumn("SUPPORT", schema.SQLVarchar)
	t.AddColumn("COMMENT", schema.SQLVarchar)
	t.AddColumn("TRANSACTIONS", schema.SQLVarchar)
	t.AddColumn("XA", schema.SQLVarchar)
	t.AddColumn("SAVEPOINTS", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addEventsTable(d *Database) error {
	t := NewDynamicTable("EVENTS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("EVENT_CATALOG", schema.SQLVarchar)
	t.AddColumn("EVENT_SCHEMA", schema.SQLVarchar)
	t.AddColumn("EVENT_NAME", schema.SQLVarchar)
	t.AddColumn("DEFINER", schema.SQLVarchar)
	t.AddColumn("TIME_ZONE", schema.SQLVarchar)
	t.AddColumn("EVENT_BODY", schema.SQLVarchar)
	t.AddColumn("EVENT_DEFINITION", schema.SQLVarchar)
	t.AddColumn("EVENT_TYPE", schema.SQLVarchar)
	t.AddColumn("EXECUTE_AT", schema.SQLVarchar)
	t.AddColumn("INTERVAL_VALUE", schema.SQLVarchar)
	t.AddColumn("INTERVAL_FIELD", schema.SQLVarchar)
	t.AddColumn("SQL_MODE", schema.SQLVarchar)
	t.AddColumn("STARTS", schema.SQLVarchar)
	t.AddColumn("ENDS", schema.SQLVarchar)
	t.AddColumn("STATUS", schema.SQLVarchar)
	t.AddColumn("ON_COMPLETION", schema.SQLVarchar)
	t.AddColumn("CREATED", schema.SQLVarchar)
	t.AddColumn("LAST_ALTERED", schema.SQLVarchar)
	t.AddColumn("LAST_EXECUTED", schema.SQLVarchar)
	t.AddColumn("EVENT_COMMENT", schema.SQLVarchar)
	t.AddColumn("ORIGINATOR", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_CLIENT", schema.SQLVarchar)
	t.AddColumn("COLLATION_CONNECTION", schema.SQLVarchar)
	t.AddColumn("DATABASE_COLLATION", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addFilesTable(d *Database) error {
	t := NewDynamicTable("FILES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("FILE_ID", schema.SQLVarchar)
	t.AddColumn("FILE_NAME", schema.SQLVarchar)
	t.AddColumn("FILE_TYPE", schema.SQLVarchar)
	t.AddColumn("TABLESPACE_NAME", schema.SQLVarchar)
	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("LOGFILE_GROUP_NAME", schema.SQLVarchar)
	t.AddColumn("LOGFILE_GROUP_NUMBER", schema.SQLVarchar)
	t.AddColumn("ENGINE", schema.SQLVarchar)
	t.AddColumn("FULLTEXT_KEYS", schema.SQLVarchar)
	t.AddColumn("DELETED_ROWS", schema.SQLVarchar)
	t.AddColumn("UPDATE_COUNT", schema.SQLVarchar)
	t.AddColumn("FREE_EXTENTS", schema.SQLVarchar)
	t.AddColumn("TOTAL_EXTENTS", schema.SQLVarchar)
	t.AddColumn("EXTENT_SIZE", schema.SQLVarchar)
	t.AddColumn("INITIAL_SIZE", schema.SQLVarchar)
	t.AddColumn("MAXIMUM_SIZE", schema.SQLVarchar)
	t.AddColumn("AUTOEXTEND_SIZE", schema.SQLVarchar)
	t.AddColumn("CREATION_TIME", schema.SQLVarchar)
	t.AddColumn("LAST_UPDATE_TIME", schema.SQLVarchar)
	t.AddColumn("LAST_ACCESS_TIME", schema.SQLVarchar)
	t.AddColumn("RECOVER_TIME", schema.SQLVarchar)
	t.AddColumn("TRANSACTION_COUNTER", schema.SQLVarchar)
	t.AddColumn("VERSION", schema.SQLVarchar)
	t.AddColumn("ROW_FORMAT", schema.SQLVarchar)
	t.AddColumn("TABLE_ROWS", schema.SQLVarchar)
	t.AddColumn("AVG_ROW_LENGTH", schema.SQLVarchar)
	t.AddColumn("DATA_LENGTH", schema.SQLVarchar)
	t.AddColumn("MAX_DATA_LENGTH", schema.SQLVarchar)
	t.AddColumn("INDEX_LENGTH", schema.SQLVarchar)
	t.AddColumn("DATA_FREE", schema.SQLVarchar)
	t.AddColumn("CREATE_TIME", schema.SQLVarchar)
	t.AddColumn("UPDATE_TIME", schema.SQLVarchar)
	t.AddColumn("CHECK_TIME", schema.SQLVarchar)
	t.AddColumn("CHECKSUM", schema.SQLVarchar)
	t.AddColumn("STATUS", schema.SQLVarchar)
	t.AddColumn("EXTRA", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addKeyColumnUsageTable(d *Database) error {
	t := NewDynamicTable("KEY_COLUMN_USAGE", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("CONSTRAINT_CATALOG", schema.SQLVarchar)
	t.AddColumn("CONSTRAINT_SCHEMA", schema.SQLVarchar)
	t.AddColumn("CONSTRAINT_NAME", schema.SQLVarchar)
	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("COLUMN_NAME", schema.SQLVarchar)
	t.AddColumn("ORDINAL_POSITION", schema.SQLVarchar)
	t.AddColumn("POSITION_IN_UNIQUE_CONSTRAINT", schema.SQLVarchar)
	t.AddColumn("REFERENCED_TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("REFERENCED_TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("REFERENCED_COLUMN_NAME", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addNdbTransidMysqlConnectionMapTable(d *Database) error {
	t := NewDynamicTable("ndb_transid_mysql_connection_map", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("mysql_connection_id", schema.SQLVarchar)
	t.AddColumn("node_id", schema.SQLVarchar)
	t.AddColumn("ndb_transid", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addParametersTable(d *Database) error {
	t := NewDynamicTable("PARAMETERS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("SPECIFIC_CATALOG", schema.SQLVarchar)
	t.AddColumn("SPECIFIC_SCHEMA", schema.SQLVarchar)
	t.AddColumn("SPECIFIC_NAME", schema.SQLVarchar)
	t.AddColumn("ORDINAL_POSITION", schema.SQLVarchar)
	t.AddColumn("PARAMETER_MODE", schema.SQLVarchar)
	t.AddColumn("PARAMETER_NAME", schema.SQLVarchar)
	t.AddColumn("DATA_TYPE", schema.SQLVarchar)
	t.AddColumn("CHARACTER_MAXIMUM_LENGTH", schema.SQLVarchar)
	t.AddColumn("CHARACTER_OCTET_LENGTH", schema.SQLVarchar)
	t.AddColumn("NUMERIC_PRECISION", schema.SQLVarchar)
	t.AddColumn("NUMERIC_SCALE", schema.SQLVarchar)
	t.AddColumn("DATETIME_PRECISION", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_NAME", schema.SQLVarchar)
	t.AddColumn("COLLATION_NAME", schema.SQLVarchar)
	t.AddColumn("DTD_IDENTIFIER", schema.SQLVarchar)
	t.AddColumn("ROUTINE_TYPE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addPartitionsTable(d *Database) error {
	t := NewDynamicTable("PARTITIONS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("PARTITION_NAME", schema.SQLVarchar)
	t.AddColumn("SUBPARTITION_NAME", schema.SQLVarchar)
	t.AddColumn("PARTITION_ORDINAL_POSITION", schema.SQLVarchar)
	t.AddColumn("SUBPARTITION_ORDINAL_POSITION", schema.SQLVarchar)
	t.AddColumn("PARTITION_METHOD", schema.SQLVarchar)
	t.AddColumn("SUBPARTITION_METHOD", schema.SQLVarchar)
	t.AddColumn("PARTITION_EXPRESSION", schema.SQLVarchar)
	t.AddColumn("SUBPARTITION_EXPRESSION", schema.SQLVarchar)
	t.AddColumn("PARTITION_DESCRIPTION", schema.SQLVarchar)
	t.AddColumn("TABLE_ROWS", schema.SQLVarchar)
	t.AddColumn("AVG_ROW_LENGTH", schema.SQLVarchar)
	t.AddColumn("DATA_LENGTH", schema.SQLVarchar)
	t.AddColumn("MAX_DATA_LENGTH", schema.SQLVarchar)
	t.AddColumn("INDEX_LENGTH", schema.SQLVarchar)
	t.AddColumn("DATA_FREE", schema.SQLVarchar)
	t.AddColumn("CREATE_TIME", schema.SQLVarchar)
	t.AddColumn("UPDATE_TIME", schema.SQLVarchar)
	t.AddColumn("CHECK_TIME", schema.SQLVarchar)
	t.AddColumn("CHECKSUM", schema.SQLVarchar)
	t.AddColumn("PARTITION_COMMENT", schema.SQLVarchar)
	t.AddColumn("NODEGROUP", schema.SQLVarchar)
	t.AddColumn("TABLESPACE_NAME", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addPluginsTable(d *Database) error {
	t := NewDynamicTable("PLUGINS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("PLUGIN_NAME", schema.SQLVarchar)
	t.AddColumn("PLUGIN_VERSION", schema.SQLVarchar)
	t.AddColumn("PLUGIN_STATUS", schema.SQLVarchar)
	t.AddColumn("PLUGIN_TYPE", schema.SQLVarchar)
	t.AddColumn("PLUGIN_TYPE_VERSION", schema.SQLVarchar)
	t.AddColumn("PLUGIN_LIBRARY", schema.SQLVarchar)
	t.AddColumn("PLUGIN_LIBRARY_VERSION", schema.SQLVarchar)
	t.AddColumn("PLUGIN_AUTHOR", schema.SQLVarchar)
	t.AddColumn("PLUGIN_DESCRIPTION", schema.SQLVarchar)
	t.AddColumn("PLUGIN_LICENSE", schema.SQLVarchar)
	t.AddColumn("LOAD_OPTION", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addProcessListTable(d *Database) error {
	t := NewDynamicTable("PROCESSLIST", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("ID", schema.SQLVarchar)
	t.AddColumn("USER", schema.SQLVarchar)
	t.AddColumn("HOST", schema.SQLVarchar)
	t.AddColumn("DB", schema.SQLVarchar)
	t.AddColumn("COMMAND", schema.SQLVarchar)
	t.AddColumn("TIME", schema.SQLVarchar)
	t.AddColumn("STATE", schema.SQLVarchar)
	t.AddColumn("INFO", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addProfilingTable(d *Database) error {
	t := NewDynamicTable("PROFILING", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("QUERY_ID", schema.SQLVarchar)
	t.AddColumn("SEQ", schema.SQLVarchar)
	t.AddColumn("STATE", schema.SQLVarchar)
	t.AddColumn("DURATION", schema.SQLVarchar)
	t.AddColumn("CPU_USER", schema.SQLVarchar)
	t.AddColumn("CPU_SYSTEM", schema.SQLVarchar)
	t.AddColumn("CONTEXT_VOLUNTARY", schema.SQLVarchar)
	t.AddColumn("CONTEXT_INVOLUNTARY", schema.SQLVarchar)
	t.AddColumn("BLOCK_OPS_IN", schema.SQLVarchar)
	t.AddColumn("BLOCK_OPS_OUT", schema.SQLVarchar)
	t.AddColumn("MESSAGES_SENT", schema.SQLVarchar)
	t.AddColumn("MESSAGES_RECEIVED", schema.SQLVarchar)
	t.AddColumn("PAGE_FAULTS_MAJOR", schema.SQLVarchar)
	t.AddColumn("PAGE_FAULTS_MINOR", schema.SQLVarchar)
	t.AddColumn("SWAPS", schema.SQLVarchar)
	t.AddColumn("SOURCE_FUNCTION", schema.SQLVarchar)
	t.AddColumn("SOURCE_FILE", schema.SQLVarchar)
	t.AddColumn("SOURCE_LINE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addReferentialConstraintsTable(d *Database) error {
	t := NewDynamicTable("REFERENTIAL_CONSTRAINTS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("CONSTRAINT_CATALOG", schema.SQLVarchar)
	t.AddColumn("CONSTRAINT_SCHEMA", schema.SQLVarchar)
	t.AddColumn("CONSTRAINT_NAME", schema.SQLVarchar)
	t.AddColumn("UNIQUE_CONSTRAINT_CATALOG", schema.SQLVarchar)
	t.AddColumn("UNIQUE_CONSTRAINT_SCHEMA", schema.SQLVarchar)
	t.AddColumn("UNIQUE_CONSTRAINT_NAME", schema.SQLVarchar)
	t.AddColumn("MATCH_OPTION", schema.SQLVarchar)
	t.AddColumn("UPDATE_RULE", schema.SQLVarchar)
	t.AddColumn("DELETE_RULE", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("REFERENCED_TABLE_NAME", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addRoutinesTable(d *Database) error {
	t := NewDynamicTable("ROUTINES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("SPECIFIC_NAME", schema.SQLVarchar)
	t.AddColumn("ROUTINE_CATALOG", schema.SQLVarchar)
	t.AddColumn("ROUTINE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("ROUTINE_NAME", schema.SQLVarchar)
	t.AddColumn("ROUTINE_TYPE", schema.SQLVarchar)
	t.AddColumn("DATA_TYPE", schema.SQLVarchar)
	t.AddColumn("CHARACTER_MAXIMUM_LENGTH", schema.SQLVarchar)
	t.AddColumn("CHARACTER_OCTET_LENGTH", schema.SQLVarchar)
	t.AddColumn("NUMERIC_PRECISION", schema.SQLVarchar)
	t.AddColumn("NUMERIC_SCALE", schema.SQLVarchar)
	t.AddColumn("DATETIME_PRECISION", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_NAME", schema.SQLVarchar)
	t.AddColumn("COLLATION_NAME", schema.SQLVarchar)
	t.AddColumn("DTD_IDENTIFIER", schema.SQLVarchar)
	t.AddColumn("ROUTINE_BODY", schema.SQLVarchar)
	t.AddColumn("ROUTINE_DEFINITION", schema.SQLVarchar)
	t.AddColumn("EXTERNAL_NAME", schema.SQLVarchar)
	t.AddColumn("EXTERNAL_LANGUAGE", schema.SQLVarchar)
	t.AddColumn("PARAMETER_STYLE", schema.SQLVarchar)
	t.AddColumn("IS_DETERMINISTIC", schema.SQLVarchar)
	t.AddColumn("SQL_DATA_ACCESS", schema.SQLVarchar)
	t.AddColumn("SQL_PATH", schema.SQLVarchar)
	t.AddColumn("SECURITY_TYPE", schema.SQLVarchar)
	t.AddColumn("CREATED", schema.SQLVarchar)
	t.AddColumn("LAST_ALTERED", schema.SQLVarchar)
	t.AddColumn("SQL_MODE", schema.SQLVarchar)
	t.AddColumn("ROUTINE_COMMENT", schema.SQLVarchar)
	t.AddColumn("DEFINER", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_CLIENT", schema.SQLVarchar)
	t.AddColumn("COLLATION_CONNECTION", schema.SQLVarchar)
	t.AddColumn("DATABASE_COLLATION", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addSchemataTable(d *Database) error {
	c := b.catalog
	t := NewDynamicTable("SCHEMATA", SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, db := range c.Databases() {
			rows = append(rows, NewDataRow(
				string(c.Name),
				string(db.Name),
				string(collation.DefaultCharset.Name),
				string(collation.Default.Name),
				"",
			))
		}
		return rows
	})

	t.AddColumn("CATALOG_NAME", schema.SQLVarchar)
	t.AddColumn("SCHEMA_NAME", schema.SQLVarchar)
	t.AddColumn("DEFAULT_CHARACTER_SET_NAME", schema.SQLVarchar)
	t.AddColumn("DEFAULT_COLLATION_NAME", schema.SQLVarchar)
	t.AddColumn("SQL_PATH", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addSchemaPrivileges(d *Database) error {
	t := NewDynamicTable("SCHEMA_PRIVILEGES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("GRANTEE", schema.SQLVarchar)
	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("PRIVILEGE_TYPE", schema.SQLVarchar)
	t.AddColumn("IS_GRANTABLE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addStatisticsTable(d *Database) error {
	t := NewDynamicTable("STATISTICS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("NON_UNIQUE", schema.SQLVarchar)
	t.AddColumn("INDEX_SCHEMA", schema.SQLVarchar)
	t.AddColumn("INDEX_NAME", schema.SQLVarchar)
	t.AddColumn("SEQ_IN_INDEX", schema.SQLVarchar)
	t.AddColumn("COLUMN_NAME", schema.SQLVarchar)
	t.AddColumn("COLLATION", schema.SQLVarchar)
	t.AddColumn("CARDINALITY", schema.SQLVarchar)
	t.AddColumn("SUB_PART", schema.SQLVarchar)
	t.AddColumn("PACKED", schema.SQLVarchar)
	t.AddColumn("NULLABLE", schema.SQLVarchar)
	t.AddColumn("INDEX_TYPE", schema.SQLVarchar)
	t.AddColumn("COMMENT", schema.SQLVarchar)
	t.AddColumn("INDEX_COMMENT", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTablesTable(d *Database) error {
	c := b.catalog
	t := NewDynamicTable("TABLES", SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, db := range c.Databases() {
			for _, tbl := range db.Tables() {
				rows = append(rows, NewDataRow(
					string(c.Name),
					string(db.Name),
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

	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("TABLE_TYPE", schema.SQLVarchar)
	t.AddColumn("ENGINE", schema.SQLVarchar)
	t.AddColumn("VERSION", schema.SQLVarchar)
	t.AddColumn("ROW_FORMAT", schema.SQLVarchar)
	t.AddColumn("TABLE_ROWS", schema.SQLInt64)
	t.AddColumn("AVG_ROW_LENGTH", schema.SQLInt64)
	t.AddColumn("DATA_LENGTH", schema.SQLInt64)
	t.AddColumn("MAX_DATA_LENGTH", schema.SQLInt64)
	t.AddColumn("INDEX_LENGTH", schema.SQLInt64)
	t.AddColumn("DATA_FREE", schema.SQLInt64)
	t.AddColumn("AUTO_INCREMENT", schema.SQLVarchar)
	t.AddColumn("CREATE_TIME", schema.SQLTimestamp)
	t.AddColumn("UPDATE_TIME", schema.SQLTimestamp)
	t.AddColumn("CHECK_TIME", schema.SQLTimestamp)
	t.AddColumn("TABLE_COLLATION", schema.SQLVarchar)
	t.AddColumn("CHECKSUM", schema.SQLVarchar)
	t.AddColumn("CREATE_OPTIONS", schema.SQLVarchar)
	t.AddColumn("TABLE_COMMENT", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTableSpacesTable(d *Database) error {
	t := NewDynamicTable("TABLESPACES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("TABLESPACE_NAME", schema.SQLVarchar)
	t.AddColumn("ENGINE", schema.SQLVarchar)
	t.AddColumn("TABLESPACE_TYPE", schema.SQLVarchar)
	t.AddColumn("LOGFILE_GROUP_NAME", schema.SQLVarchar)
	t.AddColumn("EXTENT_SIZE", schema.SQLVarchar)
	t.AddColumn("AUTOEXTEND_SIZE", schema.SQLVarchar)
	t.AddColumn("MAXIMUM_SIZE", schema.SQLVarchar)
	t.AddColumn("NODEGROUP_ID", schema.SQLVarchar)
	t.AddColumn("TABLESPACE_COMMENT", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTableConstraintsTable(d *Database) error {
	t := NewDynamicTable("TABLE_CONSTRAINTS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("CONSTRAINT_CATALOG", schema.SQLVarchar)
	t.AddColumn("CONSTRAINT_SCHEMA", schema.SQLVarchar)
	t.AddColumn("CONSTRAINT_NAME", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("CONSTRAINT_TYPE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTablePrivilegesTable(d *Database) error {
	t := NewDynamicTable("TABLE_PRIVILEGES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("GRANTEE", schema.SQLVarchar)
	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("PRIVILEGE_TYPE", schema.SQLVarchar)
	t.AddColumn("IS_GRANTABLE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addTriggersTable(d *Database) error {
	t := NewDynamicTable("TRIGGERS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("TRIGGER_CATALOG", schema.SQLVarchar)
	t.AddColumn("TRIGGER_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TRIGGER_NAME", schema.SQLVarchar)
	t.AddColumn("EVENT_MANIPULATION", schema.SQLVarchar)
	t.AddColumn("EVENT_OBJECT_CATALOG", schema.SQLVarchar)
	t.AddColumn("EVENT_OBJECT_SCHEMA", schema.SQLVarchar)
	t.AddColumn("EVENT_OBJECT_TABLE", schema.SQLVarchar)
	t.AddColumn("ACTION_ORDER", schema.SQLVarchar)
	t.AddColumn("ACTION_CONDITION", schema.SQLVarchar)
	t.AddColumn("ACTION_STATEMENT", schema.SQLVarchar)
	t.AddColumn("ACTION_ORIENTATION", schema.SQLVarchar)
	t.AddColumn("ACTION_TIMING", schema.SQLVarchar)
	t.AddColumn("ACTION_REFERENCE_OLD_TABLE", schema.SQLVarchar)
	t.AddColumn("ACTION_REFERENCE_NEW_TABLE", schema.SQLVarchar)
	t.AddColumn("ACTION_REFERENCE_OLD_ROW", schema.SQLVarchar)
	t.AddColumn("ACTION_REFERENCE_NEW_ROW", schema.SQLVarchar)
	t.AddColumn("CREATED", schema.SQLVarchar)
	t.AddColumn("SQL_MODE", schema.SQLVarchar)
	t.AddColumn("DEFINER", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_CLIENT", schema.SQLVarchar)
	t.AddColumn("COLLATION_CONNECTION", schema.SQLVarchar)
	t.AddColumn("DATABASE_COLLATION", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addUserPrivilegesTable(d *Database) error {
	t := NewDynamicTable("USER_PRIVILEGES", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("GRANTEE", schema.SQLVarchar)
	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("PRIVILEGE_TYPE", schema.SQLVarchar)
	t.AddColumn("IS_GRANTABLE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addViewsTable(d *Database) error {
	t := NewDynamicTable("VIEWS", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("TABLE_CATALOG", schema.SQLVarchar)
	t.AddColumn("TABLE_SCHEMA", schema.SQLVarchar)
	t.AddColumn("TABLE_NAME", schema.SQLVarchar)
	t.AddColumn("VIEW_DEFINITION", schema.SQLVarchar)
	t.AddColumn("CHECK_OPTION", schema.SQLVarchar)
	t.AddColumn("IS_UPDATABLE", schema.SQLVarchar)
	t.AddColumn("DEFINER", schema.SQLVarchar)
	t.AddColumn("SECURITY_TYPE", schema.SQLVarchar)
	t.AddColumn("CHARACTER_SET_CLIENT", schema.SQLVarchar)
	t.AddColumn("COLLATION_CONNECTION", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) addVariableTables(d *Database) error {

	err := b.addVariableTable(d, "GLOBAL_VARIABLES", variable.GlobalScope, variable.SystemKind)
	if err != nil {
		return err
	}
	err = b.addVariableTable(d, "SESSION_VARIABLES", variable.SessionScope, variable.SystemKind)
	if err != nil {
		return err
	}
	err = b.addVariableTable(d, "GLOBAL_STATUS", variable.GlobalScope, variable.StatusKind)
	if err != nil {
		return err
	}
	err = b.addVariableTable(d, "SESSION_STATUS", variable.SessionScope, variable.StatusKind)
	if err != nil {
		return err
	}

	return nil
}

func (b *catalogBuilder) addVariableTable(d *Database, name string, scope variable.Scope, kind variable.Kind) error {
	t := NewDynamicTable(name, SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, v := range b.variables.List(scope, kind) {
			rows = append(rows, NewDataRow(v.Name, v.Value))
		}

		return rows
	})

	t.AddColumn("VARIABLE_NAME", schema.SQLVarchar)
	t.AddColumn("VARIABLE_VALUE", schema.SQLVarchar)

	return d.AddTable(t)
}

func (b *catalogBuilder) buildMySQLDatabase() error {
	d, _ := b.catalog.AddDatabase("mysql")
	err := b.addMySQLProcTable(d)
	if err != nil {
		return err
	}

	return nil
}

func (b *catalogBuilder) addMySQLProcTable(d *Database) error {
	t := NewDynamicTable("proc", SystemView, func() []*DataRow {
		return []*DataRow{}
	})

	t.AddColumn("db", schema.SQLVarchar)
	t.AddColumn("name", schema.SQLVarchar)
	t.AddColumn("type", schema.SQLVarchar)
	t.AddColumn("specific_name", schema.SQLVarchar)
	t.AddColumn("language", schema.SQLVarchar)
	t.AddColumn("sql_data_access", schema.SQLVarchar)
	t.AddColumn("is_deterministic", schema.SQLVarchar)
	t.AddColumn("security_type", schema.SQLVarchar)
	t.AddColumn("param_list", schema.SQLVarchar)
	t.AddColumn("returns", schema.SQLVarchar)
	t.AddColumn("body", schema.SQLVarchar)
	t.AddColumn("definer", schema.SQLVarchar)
	t.AddColumn("created", schema.SQLTimestamp)
	t.AddColumn("modified", schema.SQLTimestamp)
	t.AddColumn("sql_mode", schema.SQLVarchar)
	t.AddColumn("comment", schema.SQLVarchar)
	t.AddColumn("character_set_client", schema.SQLVarchar)
	t.AddColumn("collation_connection", schema.SQLVarchar)
	t.AddColumn("db_collation", schema.SQLVarchar)
	t.AddColumn("body_utf8", schema.SQLVarchar)

	return d.AddTable(t)
}
