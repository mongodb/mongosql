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
func GenerateCreateTable(table Table, maxVarcharLength uint64) string {

	col := table.Collation()

	var buf bytes.Buffer

	switch table.Type() {
	case BaseTable:
		buf.WriteString("CREATE TABLE `" + table.Name() + "` (\n")
	default:
		buf.WriteString("CREATE TEMPORARY TABLE `" + table.Name() + "` (\n")
	}

	for i, column := range table.Columns() {
		if i > 0 {
			buf.WriteString(",\n")
		}

		buf.WriteString("  `" + column.Name + "`")
		colType := translateColumnType(column.EvalType, maxVarcharLength)
		buf.WriteString(" " + colType)
		if strings.HasPrefix(colType, "varchar") {
			buf.WriteString(" COLLATE " + string(col.Name))
		}
		buf.WriteString(" DEFAULT NULL")
		if column.Comments != "" {
			buf.WriteString(" COMMENT '" + strings.Replace(column.Comments, "'", "''", -1) + "'")
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
			buf.WriteString("`" + column.Name + "`")
		}
		buf.WriteString(")")
	}

	indexes := table.Indexes()
	if len(indexes) > 0 {
		for _, idx := range indexes {
			buf.WriteString(",\n")
			keyType := "KEY"
			if idx.unique {
				keyType = "UNIQUE " + keyType
			}
			buf.WriteString("  " + keyType + " `" + idx.constraintName + "` (")
			for j, col := range idx.columns {
				if j > 0 {
					buf.WriteString(",")
				}
				buf.WriteString("`" + col.Name + "`")
			}
			buf.WriteString(")")
		}
	}

	foreignKeys := table.ForeignKeys()
	if len(foreignKeys) > 0 {
		for _, fk := range foreignKeys {
			buf.WriteString(",\n")
			buf.WriteString("  CONSTRAINT `" + fk.constraintName + "` FOREIGN KEY" + " (")
			for j, col := range fk.columns {
				if j > 0 {
					buf.WriteString(",")
				}
				buf.WriteString("`" + col.Name + "`")
			}
			buf.WriteString(") REFERENCES `" + fk.foreignTable + "` (")
			for j, col := range fk.columns {
				if j > 0 {
					buf.WriteString(",")
				}
				buf.WriteString("`" + fk.localToForeignColumn[col.Name] + "`")
			}
			buf.WriteString(")")
			buf.WriteString(" ON DELETE CASCADE ON UPDATE CASCADE")
		}
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
