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

	return b.buildInformationSchemaDatabase()
}

func (b *catalogBuilder) buildFromSchema() error {
	info := b.variables.MongoDBInfo
	for _, dbConfig := range b.schema.RawDatabases {
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
		for _, tblConfig := range dbConfig.RawTables {
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
			t := NewMongoTable(tblConfig, col)
			err = d.AddTable(t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *catalogBuilder) buildInformationSchemaDatabase() error {
	d, _ := b.catalog.AddDatabase("information_schema")
	err := b.addCharsetTable(d)
	if err != nil {
		return err
	}
	err = b.addCollationTable(d)
	if err != nil {
		return err
	}
	err = b.addColumnsTable(d)
	if err != nil {
		return err
	}
	err = b.addSchemataTable(d)
	if err != nil {
		return err
	}
	err = b.addTablesTable(d)
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

func (b *catalogBuilder) addColumnsTable(d *Database) error {
	c := b.catalog
	t := NewDynamicTable("COLUMNS", SystemView, func() []*DataRow {
		var rows []*DataRow
		for _, db := range c.Databases() {
			for _, tbl := range db.Tables() {
				for i, col := range tbl.Columns() {
					rows = append(rows, NewDataRow(
						string(c.Name),
						string(db.Name),
						string(tbl.Name()),
						string(col.Name()),
						i,
						nil,
						"YES",
						string(col.Type()),
						nil,
						nil,
						nil,
						nil,
						nil,
						string(tbl.Collation().CharsetName),
						string(tbl.Collation().Name),
						string(col.Type()),
						"",
						"",
						"select",
						col.Comments(),
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
