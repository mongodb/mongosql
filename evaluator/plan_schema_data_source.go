package evaluator

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/schema"
)

var (
	ISSchemataHeaders = []string{
		"CATALOG_NAME",
		"SCHEMA_NAME",
		"DEFAULT_CHARACTER_SET_NAME",
		"DEFAULT_COLLATION_NAME",
		"SQL_PATH",
	}

	ISTablesHeaders = []string{
		"TABLE_CATALOG",
		"TABLE_SCHEMA",
		"TABLE_NAME",
		"TABLE_TYPE",
		"TABLE_COMMENT",
	}

	ISColumnHeaders = []string{
		"TABLE_CATALOG",
		"TABLE_SCHEMA",
		"TABLE_NAME",
		"COLUMN_NAME",
		"ORDINAL_POSITION",
		"COLUMN_DEFAULT",
		"IS_NULLABLE",
		"DATA_TYPE",
		"CHARACTER_MAXIMUM_LENGTH",
		"CHARACTER_OCTET_LENGTH",
		"NUMERIC_PRECISION",
		"NUMERIC_SCALE",
		"DATETIME_PRECISION",
		"CHARACTER_SET_NAME",
		"COLLATION_NAME",
		"COLUMN_TYPE",
		"COLUMN_KEY",
		"EXTRA",
		"PRIVILEGES",
		"COLUMN_COMMENT",
	}
)

type SchemaDataSourceStage struct {
	tableName string
	aliasName string
}

func NewSchemaDataSourceStage(tableName, aliasName string) *SchemaDataSourceStage {
	if aliasName == "" {
		aliasName = tableName
	}

	return &SchemaDataSourceStage{
		tableName: tableName,
		aliasName: aliasName,
	}
}

func (sds *SchemaDataSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	if sds.tableName == "key_column_usage" {
		return &EmptyIter{}, nil
	}

	aliasName := sds.aliasName
	if aliasName == "" {
		aliasName = sds.tableName
	}

	it := &SchemaDataSourceIter{
		tableName: aliasName,
	}
	switch strings.ToLower(sds.tableName) {
	case "columns":
		it.rows = sds.gatherColumnRows(ctx)
	case "schemata":
		it.rows = sds.gatherSchemataRows(ctx)
	case "tables":
		it.rows = sds.gatherTableRows(ctx)
	default:
		return nil, fmt.Errorf("unsupported %q table %q", InformationDatabase, sds.tableName)
	}

	return it, nil
}

func (sds *SchemaDataSourceStage) OpFields() []*Column {

	var headers []string

	switch strings.ToLower(sds.tableName) {
	case "columns":
		headers = ISColumnHeaders
	case "schemata":
		headers = ISSchemataHeaders
	case "tables":
		headers = ISTablesHeaders
	}

	aliasName := sds.aliasName
	if aliasName == "" {
		aliasName = sds.tableName
	}

	var columns []*Column
	for _, c := range headers {
		column := &Column{
			Table:     aliasName,
			Name:      c,
			View:      c,
			SQLType:   schema.SQLVarchar,
			MongoType: schema.MongoString,
		}
		columns = append(columns, column)
	}

	return columns
}

func (sds *SchemaDataSourceStage) gatherColumnRows(ctx *ExecutionCtx) []Values {
	rows := []Values{}
	for _, db := range ctx.PlanCtx.Schema.RawDatabases {
		if !ctx.AuthProvider.IsDatabaseAllowed(db.Name) {
			continue
		}

		for _, table := range db.RawTables {
			if !ctx.AuthProvider.IsCollectionAllowed(db.Name, table.CollectionName) {
				continue
			}

			for i, column := range table.RawColumns {

				row := Values{
					Value{Name: ISColumnHeaders[0], View: ISColumnHeaders[0], Data: SQLVarchar("def")},
					Value{Name: ISColumnHeaders[1], View: ISColumnHeaders[1], Data: SQLVarchar(db.Name)},
					Value{Name: ISColumnHeaders[2], View: ISColumnHeaders[2], Data: SQLVarchar(table.Name)},
					Value{Name: ISColumnHeaders[3], View: ISColumnHeaders[3], Data: SQLVarchar(column.SqlName)},
					Value{Name: ISColumnHeaders[4], View: ISColumnHeaders[4], Data: SQLInt(i)},
					Value{Name: ISColumnHeaders[5], View: ISColumnHeaders[5], Data: SQLNull},
					Value{Name: ISColumnHeaders[6], View: ISColumnHeaders[6], Data: SQLVarchar("YES")},
					Value{Name: ISColumnHeaders[7], View: ISColumnHeaders[7], Data: SQLVarchar(string(column.SqlType))},
					Value{Name: ISColumnHeaders[8], View: ISColumnHeaders[8], Data: SQLNull},
					Value{Name: ISColumnHeaders[9], View: ISColumnHeaders[9], Data: SQLNull},
					Value{Name: ISColumnHeaders[10], View: ISColumnHeaders[10], Data: SQLNull},
					Value{Name: ISColumnHeaders[11], View: ISColumnHeaders[11], Data: SQLNull},
					Value{Name: ISColumnHeaders[12], View: ISColumnHeaders[12], Data: SQLNull},
					Value{Name: ISColumnHeaders[13], View: ISColumnHeaders[13], Data: SQLNull},
					Value{Name: ISColumnHeaders[14], View: ISColumnHeaders[14], Data: SQLVarchar("utf8_general_ci")},
					Value{Name: ISColumnHeaders[15], View: ISColumnHeaders[15], Data: SQLVarchar(string(column.SqlType))},
					Value{Name: ISColumnHeaders[16], View: ISColumnHeaders[16], Data: SQLNull},
					Value{Name: ISColumnHeaders[17], View: ISColumnHeaders[17], Data: SQLNull},
					Value{Name: ISColumnHeaders[18], View: ISColumnHeaders[18], Data: SQLNull},
					Value{Name: ISColumnHeaders[19], View: ISColumnHeaders[19], Data: SQLVarchar(fmt.Sprintf("{ \"name\": \"%s\" }", column.Name))},
				}

				rows = append(rows, row)
			}
		}
	}

	return rows
}

func (sds *SchemaDataSourceStage) gatherSchemataRows(ctx *ExecutionCtx) []Values {
	rows := []Values{}
	for _, db := range ctx.PlanCtx.Schema.RawDatabases {
		if !ctx.AuthProvider.IsDatabaseAllowed(db.Name) {
			continue
		}

		row := Values{
			Value{Name: ISSchemataHeaders[0], View: ISSchemataHeaders[0], Data: SQLVarchar("def")},
			Value{Name: ISSchemataHeaders[1], View: ISSchemataHeaders[1], Data: SQLVarchar(db.Name)},
			Value{Name: ISSchemataHeaders[2], View: ISSchemataHeaders[2], Data: SQLNull},
			Value{Name: ISSchemataHeaders[3], View: ISSchemataHeaders[3], Data: SQLVarchar("utf8_general_ci")},
			Value{Name: ISSchemataHeaders[4], View: ISSchemataHeaders[4], Data: SQLNull},
		}

		rows = append(rows, row)
	}

	return rows
}

func (sds *SchemaDataSourceStage) gatherTableRows(ctx *ExecutionCtx) []Values {
	rows := []Values{}
	for _, db := range ctx.PlanCtx.Schema.RawDatabases {
		if !ctx.AuthProvider.IsDatabaseAllowed(db.Name) {
			continue
		}

		for _, table := range db.RawTables {
			if !ctx.AuthProvider.IsCollectionAllowed(db.Name, table.CollectionName) {
				continue
			}

			row := Values{
				Value{Name: ISTablesHeaders[0], View: ISTablesHeaders[0], Data: SQLVarchar("def")},
				Value{Name: ISTablesHeaders[1], View: ISTablesHeaders[1], Data: SQLVarchar(db.Name)},
				Value{Name: ISTablesHeaders[2], View: ISTablesHeaders[2], Data: SQLVarchar(table.Name)},
				Value{Name: ISTablesHeaders[3], View: ISTablesHeaders[3], Data: SQLVarchar("VIEW")},
				Value{Name: ISTablesHeaders[4], View: ISTablesHeaders[4], Data: SQLVarchar(fmt.Sprintf("{ \"collectionName\": \"%s\" }", table.CollectionName))},
			}

			rows = append(rows, row)
		}

	}

	return rows
}

type SchemaDataSourceIter struct {
	tableName string
	rows      []Values
	index     int
}

func (sds *SchemaDataSourceIter) Next(row *Row) bool {
	if sds.index >= len(sds.rows) {
		return false
	}

	row.Data = TableRows{{sds.tableName, sds.rows[sds.index]}}
	sds.index++
	return true
}

func (sds *SchemaDataSourceIter) Close() error {
	return nil
}

func (sds *SchemaDataSourceIter) Err() error {
	return nil
}
