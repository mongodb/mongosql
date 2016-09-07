package evaluator

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
)

const (
	isCatalogName             = "def"
	isCharacterSetName        = "utf8"
	isCollationName           = "utf8_bin"
	informationSchemaDatabase = "information_schema"
)

type isColumn struct {
	name    string
	sqlType schema.SQLType
}

var (
	isSchemataHeaders = []isColumn{
		{"CATALOG_NAME", schema.SQLVarchar},
		{"SCHEMA_NAME", schema.SQLVarchar},
		{"DEFAULT_CHARACTER_SET_NAME", schema.SQLVarchar},
		{"DEFAULT_COLLATION_NAME", schema.SQLVarchar},
		{"SQL_PATH", schema.SQLVarchar},
	}

	isTablesHeaders = []isColumn{
		{"TABLE_CATALOG", schema.SQLVarchar},
		{"TABLE_SCHEMA", schema.SQLVarchar},
		{"TABLE_NAME", schema.SQLVarchar},
		{"TABLE_TYPE", schema.SQLVarchar},
		{"TABLE_COMMENT", schema.SQLVarchar},
	}

	isColumnHeaders = []isColumn{
		{"TABLE_CATALOG", schema.SQLVarchar},
		{"TABLE_SCHEMA", schema.SQLVarchar},
		{"TABLE_NAME", schema.SQLVarchar},
		{"COLUMN_NAME", schema.SQLVarchar},
		{"ORDINAL_POSITION", schema.SQLInt64},
		{"COLUMN_DEFAULT", schema.SQLVarchar},
		{"IS_NULLABLE", schema.SQLVarchar},
		{"DATA_TYPE", schema.SQLVarchar},
		{"CHARACTER_MAXIMUM_LENGTH", schema.SQLInt64},
		{"CHARACTER_OCTET_LENGTH", schema.SQLInt64},
		{"NUMERIC_PRECISION", schema.SQLInt64},
		{"NUMERIC_SCALE", schema.SQLInt64},
		{"DATETIME_PRECISION", schema.SQLInt64},
		{"CHARACTER_SET_NAME", schema.SQLVarchar},
		{"COLLATION_NAME", schema.SQLVarchar},
		{"COLUMN_TYPE", schema.SQLVarchar},
		{"COLUMN_KEY", schema.SQLVarchar},
		{"EXTRA", schema.SQLVarchar},
		{"PRIVILEGES", schema.SQLVarchar},
		{"COLUMN_COMMENT", schema.SQLVarchar},
	}

	isVariableHeaders = []isColumn{
		{"VARIABLE_NAME", schema.SQLVarchar},
		{"VARIABLE_VALUE", schema.SQLVarchar},
	}

	isCollationHeaders = []isColumn{
		{"COLLATION_NAME", schema.SQLVarchar},
		{"CHARACTER_SET_NAME", schema.SQLVarchar},
		{"ID", schema.SQLVarchar},
		{"IS_DEFAULT", schema.SQLVarchar},
		{"IS_COMPILED", schema.SQLVarchar},
		{"SORTLEN", schema.SQLVarchar},
	}
)

type SchemaDataSourceStage struct {
	selectID  int
	tableName string
	aliasName string
	schema    *schema.Schema
}

func NewSchemaDataSourceStage(selectID int, schema *schema.Schema, tableName, aliasName string) *SchemaDataSourceStage {
	if aliasName == "" {
		aliasName = tableName
	}

	return &SchemaDataSourceStage{
		selectID:  selectID,
		schema:    schema,
		tableName: tableName,
		aliasName: aliasName,
	}
}

func (sds *SchemaDataSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	if sds.tableName == "key_column_usage" {
		return &EmptyIter{}, nil
	}

	it := &SchemaDataSourceIter{
		tableName: sds.aliasName,
	}
	switch strings.ToLower(sds.tableName) {
	case "collations":
		it.rows = sds.gatherCollationRows(ctx)
	case "columns":
		it.rows = sds.gatherColumnRows(ctx)
	case "schemata":
		it.rows = sds.gatherSchemataRows(ctx)
	case "tables":
		it.rows = sds.gatherTableRows(ctx)
	case "global_variables":
		it.rows = sds.gatherVariableRows(ctx, GlobalVariable)
	case "session_variables":
		it.rows = sds.gatherVariableRows(ctx, SessionVariable)
	default:
		return nil, fmt.Errorf("unsupported %q table %q", informationSchemaDatabase, sds.tableName)
	}

	return it, nil
}

func (sds *SchemaDataSourceStage) Columns() []*Column {

	var headers []isColumn

	switch strings.ToLower(sds.tableName) {
	case "collations":
		headers = isCollationHeaders
	case "columns":
		headers = isColumnHeaders
	case "schemata":
		headers = isSchemataHeaders
	case "tables":
		headers = isTablesHeaders
	case "global_variables", "session_variables":
		headers = isVariableHeaders
	}

	aliasName := sds.aliasName
	if aliasName == "" {
		aliasName = sds.tableName
	}

	var columns []*Column
	for _, c := range headers {
		column := &Column{
			SelectID:  sds.selectID,
			Table:     aliasName,
			Name:      c.name,
			SQLType:   c.sqlType,
			MongoType: schema.MongoNone,
		}
		columns = append(columns, column)
	}

	return columns
}

func (sds *SchemaDataSourceStage) getValue(c isColumn, data interface{}) Value {
	data, _ = NewSQLValueFromSQLColumnExpr(data, c.sqlType, schema.MongoNone)
	return Value{SelectID: sds.selectID, Table: sds.aliasName, Name: c.name, Data: data}
}

func (sds *SchemaDataSourceStage) gatherCollationRows(ctx *ExecutionCtx) []Values {

	collations := collation.GetAll()

	rows := []Values{}

	for _, c := range collations {
		isDefault := "No"

		if c.Default {
			isDefault = "Yes"
		}

		row := Values{
			sds.getValue(isCollationHeaders[0], c.Name),
			sds.getValue(isCollationHeaders[1], string(c.Charset.Name)),
			sds.getValue(isCollationHeaders[2], uint8(c.ID)),
			sds.getValue(isCollationHeaders[3], isDefault),
			sds.getValue(isCollationHeaders[4], "Yes"),
			sds.getValue(isCollationHeaders[5], c.SortLen),
		}

		rows = append(rows, row)
	}

	return rows
}

func (sds *SchemaDataSourceStage) gatherColumnRows(ctx *ExecutionCtx) []Values {
	rows := []Values{}

	appendRows := func(headers []isColumn, tableName string) {
		for i, c := range headers {
			row := Values{
				sds.getValue(isColumnHeaders[0], isCatalogName),
				sds.getValue(isColumnHeaders[1], informationSchemaDatabase),
				sds.getValue(isColumnHeaders[2], tableName),
				sds.getValue(isColumnHeaders[3], c.name),
				sds.getValue(isColumnHeaders[4], i),
				sds.getValue(isColumnHeaders[5], nil),
				sds.getValue(isColumnHeaders[6], "YES"),
				sds.getValue(isColumnHeaders[7], string(c.sqlType)),
				sds.getValue(isColumnHeaders[8], nil),
				sds.getValue(isColumnHeaders[9], nil),
				sds.getValue(isColumnHeaders[10], nil),
				sds.getValue(isColumnHeaders[11], nil),
				sds.getValue(isColumnHeaders[12], nil),
				sds.getValue(isColumnHeaders[13], isCharacterSetName),
				sds.getValue(isColumnHeaders[14], isCollationName),
				sds.getValue(isColumnHeaders[15], string(c.sqlType)),
				sds.getValue(isColumnHeaders[16], ""),
				sds.getValue(isColumnHeaders[17], ""),
				sds.getValue(isColumnHeaders[18], "select"),
				sds.getValue(isColumnHeaders[19], ""),
			}
			rows = append(rows, row)
		}
	}

	appendRows(isColumnHeaders, "COLUMNS")
	appendRows(isVariableHeaders, "GLOBAL_VARIABLES")
	appendRows(isSchemataHeaders, "SCHEMATA")
	appendRows(isVariableHeaders, "SESSION_VARIABLES")
	appendRows(isTablesHeaders, "TABLES")

	for _, db := range sds.schema.RawDatabases {
		if !ctx.AuthProvider.IsDatabaseAllowed(db.Name) {
			continue
		}

		for _, table := range db.RawTables {
			if !ctx.AuthProvider.IsCollectionAllowed(db.Name, table.CollectionName) {
				continue
			}

			for i, column := range table.RawColumns {
				if column.MongoType == schema.MongoFilter {
					continue
				}

				row := Values{
					sds.getValue(isColumnHeaders[0], isCatalogName),
					sds.getValue(isColumnHeaders[1], db.Name),
					sds.getValue(isColumnHeaders[2], table.Name),
					sds.getValue(isColumnHeaders[3], column.SqlName),
					sds.getValue(isColumnHeaders[4], i),
					sds.getValue(isColumnHeaders[5], nil),
					sds.getValue(isColumnHeaders[6], "YES"),
					sds.getValue(isColumnHeaders[7], string(column.SqlType)),
					sds.getValue(isColumnHeaders[8], nil),
					sds.getValue(isColumnHeaders[9], nil),
					sds.getValue(isColumnHeaders[10], nil),
					sds.getValue(isColumnHeaders[11], nil),
					sds.getValue(isColumnHeaders[12], nil),
					sds.getValue(isColumnHeaders[13], isCharacterSetName),
					sds.getValue(isColumnHeaders[14], isCollationName),
					sds.getValue(isColumnHeaders[15], string(column.SqlType)),
					sds.getValue(isColumnHeaders[16], ""),
					sds.getValue(isColumnHeaders[17], ""),
					sds.getValue(isColumnHeaders[18], "select"),
					sds.getValue(isColumnHeaders[19], fmt.Sprintf("{ \"name\": \"%s\" }", column.Name)),
				}

				rows = append(rows, row)
			}
		}
	}

	return rows
}

func (sds *SchemaDataSourceStage) gatherSchemataRows(ctx *ExecutionCtx) []Values {
	rows := []Values{
		Values{
			sds.getValue(isSchemataHeaders[0], isCatalogName),
			sds.getValue(isSchemataHeaders[1], informationSchemaDatabase),
			sds.getValue(isSchemataHeaders[2], isCharacterSetName),
			sds.getValue(isSchemataHeaders[3], isCollationName),
			sds.getValue(isSchemataHeaders[4], ""),
		},
	}

	for _, db := range sds.schema.RawDatabases {
		if !ctx.AuthProvider.IsDatabaseAllowed(db.Name) {
			continue
		}

		row := Values{
			sds.getValue(isSchemataHeaders[0], isCatalogName),
			sds.getValue(isSchemataHeaders[1], db.Name),
			sds.getValue(isSchemataHeaders[2], isCharacterSetName),
			sds.getValue(isSchemataHeaders[3], isCollationName),
			sds.getValue(isSchemataHeaders[4], ""),
		}

		rows = append(rows, row)
	}

	return rows
}

func (sds *SchemaDataSourceStage) gatherTableRows(ctx *ExecutionCtx) []Values {
	rows := []Values{
		Values{
			sds.getValue(isTablesHeaders[0], isCatalogName),
			sds.getValue(isTablesHeaders[1], informationSchemaDatabase),
			sds.getValue(isTablesHeaders[2], "COLUMNS"),
			sds.getValue(isTablesHeaders[3], "SYSTEM VIEW"),
			sds.getValue(isTablesHeaders[4], ""),
		},
		Values{
			sds.getValue(isTablesHeaders[0], isCatalogName),
			sds.getValue(isTablesHeaders[1], informationSchemaDatabase),
			sds.getValue(isTablesHeaders[2], "GLOBAL_VARIABLES"),
			sds.getValue(isTablesHeaders[3], "SYSTEM VIEW"),
			sds.getValue(isTablesHeaders[4], ""),
		},
		Values{
			sds.getValue(isTablesHeaders[0], isCatalogName),
			sds.getValue(isTablesHeaders[1], informationSchemaDatabase),
			sds.getValue(isTablesHeaders[2], "SCHEMATA"),
			sds.getValue(isTablesHeaders[3], "SYSTEM VIEW"),
			sds.getValue(isTablesHeaders[4], ""),
		},
		Values{
			sds.getValue(isTablesHeaders[0], isCatalogName),
			sds.getValue(isTablesHeaders[1], informationSchemaDatabase),
			sds.getValue(isTablesHeaders[2], "SESSION_VARIABLES"),
			sds.getValue(isTablesHeaders[3], "SYSTEM VIEW"),
			sds.getValue(isTablesHeaders[4], ""),
		},
		Values{
			sds.getValue(isTablesHeaders[0], isCatalogName),
			sds.getValue(isTablesHeaders[1], informationSchemaDatabase),
			sds.getValue(isTablesHeaders[2], "TABLES"),
			sds.getValue(isTablesHeaders[3], "SYSTEM VIEW"),
			sds.getValue(isTablesHeaders[4], ""),
		},
	}

	for _, db := range sds.schema.RawDatabases {
		if !ctx.AuthProvider.IsDatabaseAllowed(db.Name) {
			continue
		}

		for _, table := range db.RawTables {
			if !ctx.AuthProvider.IsCollectionAllowed(db.Name, table.CollectionName) {
				continue
			}

			row := Values{
				sds.getValue(isTablesHeaders[0], isCatalogName),
				sds.getValue(isTablesHeaders[1], db.Name),
				sds.getValue(isTablesHeaders[2], table.Name),
				sds.getValue(isTablesHeaders[3], "VIEW"),
				sds.getValue(isTablesHeaders[4], fmt.Sprintf("{ \"collectionName\": \"%s\" }", table.CollectionName)),
			}

			rows = append(rows, row)
		}
	}

	return rows
}

func (sds *SchemaDataSourceStage) gatherVariableRows(ctx *ExecutionCtx, variableKind VariableKind) []Values {
	rows := []Values{}

	scope, kind := variableKind.scopeAndKind()

	for _, value := range ctx.Variables().List(scope, kind) {
		row := Values{
			sds.getValue(isVariableHeaders[0], value.Name),
			sds.getValue(isVariableHeaders[1], value.Value),
		}

		rows = append(rows, row)
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

	row.Data = sds.rows[sds.index]
	sds.index++
	return true
}

func (sds *SchemaDataSourceIter) Close() error {
	return nil
}

func (sds *SchemaDataSourceIter) Err() error {
	return nil
}
