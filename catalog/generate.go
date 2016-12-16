package catalog

import (
	"bytes"

	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
)

// GenerateCreateTable generates a create table statement for the table.
func GenerateCreateTable(table Table) string {

	col := table.Collation()

	var buf bytes.Buffer

	switch table.Type() {
	case BaseTable:
		buf.WriteString("CREATE TABLE `" + string(table.Name()) + "` (\n")
	default:
		buf.WriteString("CREATE TEMPORARY TABLE `" + string(table.Name()) + "` (\n")
	}

	for i, column := range table.Columns() {
		if i > 0 {
			buf.WriteString(",\n")
		}

		buf.WriteString("  `" + string(column.Name()) + "`")
		buf.WriteString(" " + translateColumnType(col, column.Type()))
		buf.WriteString(" DEFAULT NULL")
		if column.Comments() != "" {
			buf.WriteString(" COMMENT '" + strings.Replace(column.Comments(), "'", "''", -1) + "'")
		}
	}

	pks := table.PrimaryKeys()
	if len(pks) > 0 {
		buf.WriteString(",\n")
		buf.WriteString("  PRIMARY KEY (")
		for i, column := range pks {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("`" + string(column.Name()) + "`")
		}
		buf.WriteString(")")
	}

	buf.WriteString("\n)")

	switch table.Type() {
	case BaseTable:
		buf.WriteString(" ENGINE=MongoDB")
	default:
		buf.WriteString(" ENGINE=MEMORY")
	}

	buf.WriteString(" DEFAULT CHARSET=" + string(col.CharsetName) + " COLLATE=" + string(col.Name))

	if table.Comments() != "" {
		buf.WriteString(" COMMENT='" + strings.Replace(table.Comments(), "'", "''", -1) + "'")
	}

	return buf.String()
}

func translateColumnType(collation *collation.Collation, sqlType schema.SQLType) string {
	switch sqlType {
	case schema.SQLBoolean:
		return "tinyint(1)"
	case schema.SQLDate:
		return "date"
	case schema.SQLDecimal128:
		return "decimal"
	case schema.SQLFloat, schema.SQLArrNumeric:
		return "double"
	case schema.SQLInt, schema.SQLInt64:
		return "bigint(20)"
	case schema.SQLObjectID:
		return "varchar(24) COLLATE " + string(collation.Name)
	case schema.SQLTimestamp:
		return "datetime"
	case schema.SQLUint64:
		return "bigint(20) unsigned"
	case schema.SQLUUID:
		return "varchar(36) COLLATE " + string(collation.Name)
	case schema.SQLVarchar:
		return "varchar(65535) COLLATE " + string(collation.Name)
	default:
		return "<unknown>"
	}
}
