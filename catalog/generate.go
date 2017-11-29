package catalog

import (
	"bytes"

	"strings"
)

// GenerateCreateDatabase generates a create database statement for the database with its modifier.
func GenerateCreateDatabase(database, modifier string) string {
	prefix := "CREATE DATABASE `"
	if modifier != "" {
		prefix = "CREATE DATABASE /*!32312 " + modifier + "*/ `"
	}
	return prefix + database + "` /*!40100 DEFAULT CHARACTER SET utf8 */"
}

// GenerateCreateTable generates a create table statement for the table.
func GenerateCreateTable(table Table, maxVarcharLength uint16) string {

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
		colType := translateColumnType(column.Type(), maxVarcharLength)
		buf.WriteString(" " + colType)
		if strings.HasPrefix(colType, "varchar") {
			buf.WriteString(" COLLATE " + string(col.Name))
		}
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
