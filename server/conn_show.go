package server

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

func (c *conn) handleShow(sql string, stmt *parser.Show) error {
	switch strings.ToLower(stmt.Section) {
	case "charset":
		return c.handleShowCharset(stmt)
	case "columns":
		return c.handleShowColumns(sql, stmt)
	case "collation":
		return c.handleShowCollation(stmt)
	case "databases", "schemas":
		return c.handleShowDatabases(stmt)
	case "status":
		return c.handleShowStatus(stmt)
	case "tables":
		return c.handleShowTables(sql, stmt)
	case "variables":
		return c.handleShowVariables(stmt)
	default:
		return mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "no support for show (%s) for now", sql)
	}
}

func (c *conn) handleShowCharset(stmt *parser.Show) error {

	translated := "SELECT * FROM (" +
		" SELECT CHARACTER_SET_NAME AS `Charset`" +
		", DESCRIPTION AS `Description`" +
		", DEFAULT_COLLATE_NAME AS `Default collation`" +
		", MAXLEN AS `Maxlen`"

	translated += " FROM INFORMATION_SCHEMA.CHARACTER_SETS) `is`"

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case parser.StrVal:
			translated += fmt.Sprintf(" WHERE `Charset` LIKE %s", parser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" WHERE %s", parser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY `Charset`"

	return c.handleQuery(translated)
}

func (c *conn) handleShowCollation(stmt *parser.Show) error {

	translated := "SELECT * FROM (" +
		" SELECT COLLATION_NAME AS `Collation`" +
		", CHARACTER_SET_NAME AS `Charset`" +
		", ID AS `Id`" +
		", IS_DEFAULT AS `Default`" +
		", IS_COMPILED AS `Compiled`" +
		", SORTLEN AS `Sortlen`"

	translated += " FROM INFORMATION_SCHEMA.COLLATIONS) `is`"

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case parser.StrVal:
			translated += fmt.Sprintf(" WHERE `Collation` LIKE %s", parser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" WHERE %s", parser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY `Collation`"

	return c.handleQuery(translated)
}

func (c *conn) handleShowColumns(sql string, stmt *parser.Show) error {

	translated := "SELECT `Field`, `Type`"
	if strings.EqualFold(stmt.Modifier, "full") {
		translated += ", `Collation`"
	}

	translated += ", `Null`, `Key`, `Default`, `Extra`"

	if strings.EqualFold(stmt.Modifier, "full") {
		translated += ", `Privileges`, `Comment`"
	}

	translated += " FROM (" +
		" SELECT COLUMN_NAME AS `Field`" +
		", COLUMN_TYPE AS `Type`" +
		", IS_NULLABLE AS `Null`" +
		", COLUMN_KEY AS `Key`" +
		", COLUMN_DEFAULT AS `Default`" +
		", EXTRA AS `Extra`" +
		", COLLATION_NAME AS `Collation`" +
		", PRIVILEGES AS `Privileges`" +
		", COLUMN_COMMENT AS `Comment`" +
		", TABLE_NAME" +
		", TABLE_SCHEMA" +
		", ORDINAL_POSITION"

	translated += " FROM INFORMATION_SCHEMA.COLUMNS) `is`"

	var dbName string
	if c.currentDB != nil {
		dbName = c.currentDB.Name
	}

	table := ""

	switch f := stmt.From.(type) {
	case parser.StrVal:
		table = string(f)
	case *parser.ColName:
		if f.Qualifier != nil {
			dbName = string(f.Qualifier)
		}
		table = string(f.Name)
	default:
		return mysqlerrors.Defaultf(mysqlerrors.ER_ILLEGAL_VALUE_FOR_TYPE, "FROM", parser.String(f))
	}

	if stmt.DBFilter != nil {
		switch f := stmt.DBFilter.(type) {
		case parser.StrVal:
			dbName = string(f)
		case *parser.ColName:
			dbName = string(f.Name)
		default:
			return mysqlerrors.Defaultf(mysqlerrors.ER_ILLEGAL_VALUE_FOR_TYPE, "FROM", parser.String(f))
		}
	}

	translated += fmt.Sprintf(" WHERE TABLE_NAME = '%s'", table)

	if dbName != "" {
		translated += fmt.Sprintf(" AND TABLE_SCHEMA = '%s'", dbName)
	}

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case parser.StrVal:
			translated += fmt.Sprintf(" AND `Field` LIKE %s", parser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" AND %s", parser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY ORDINAL_POSITION"

	return c.handleQuery(translated)
}

func (c *conn) handleShowDatabases(stmt *parser.Show) error {
	translated := "SELECT * FROM (SELECT SCHEMA_NAME AS `Database`" +
		" FROM INFORMATION_SCHEMA.SCHEMATA) `is`"

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case parser.StrVal:
			translated += fmt.Sprintf(" WHERE `Database` LIKE %s", parser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" WHERE %s", parser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY `Database`"

	return c.handleQuery(translated)
}

func (c *conn) handleShowTables(sql string, stmt *parser.Show) error {
	dbName := ""
	if c.currentDB != nil {
		dbName = c.currentDB.Name
	}

	if stmt.From != nil {
		switch f := stmt.From.(type) {
		case parser.StrVal:
			dbName = string(f)
		case *parser.ColName:
			dbName = string(f.Name)
		default:
			return mysqlerrors.Defaultf(mysqlerrors.ER_ILLEGAL_VALUE_FOR_TYPE, "FROM", parser.String(f))
		}
	}

	if dbName == "" {
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_DB_ERROR)
	}

	columnName := fmt.Sprintf("`Tables_in_%s`", dbName)

	translated := fmt.Sprintf("SELECT %s", columnName)
	if strings.EqualFold(stmt.Modifier, "full") {
		translated += ", `Type`"
	}

	translated += " FROM (" +
		fmt.Sprintf(" SELECT TABLE_NAME AS %s, TABLE_TYPE AS `Type`", columnName) +
		" FROM INFORMATION_SCHEMA.TABLES " +
		fmt.Sprintf(" WHERE TABLE_SCHEMA = '%s') `is`", dbName)

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case parser.StrVal:
			translated += fmt.Sprintf(" WHERE %s LIKE %s", columnName, parser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" WHERE %s", parser.String(stmt.LikeOrWhere))
		}
	}

	translated += fmt.Sprintf(" ORDER BY %s", columnName)

	return c.handleQuery(translated)
}

func (c *conn) handleShowVariableKind(stmt *parser.Show, kind string) error {
	tableName := strings.ToUpper(stmt.Modifier)
	kind = strings.ToUpper(kind)

	translated := fmt.Sprintf("SELECT * FROM (SELECT VARIABLE_NAME AS `Variable_name`, VARIABLE_VALUE AS `Value`"+
		" FROM INFORMATION_SCHEMA.%s_%s) `is`", tableName, kind)

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case parser.StrVal:
			translated += fmt.Sprintf(" WHERE `Variable_name` LIKE %s", parser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" WHERE %s", parser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY `Variable_name`"

	return c.handleQuery(translated)
}

func (c *conn) handleShowStatus(stmt *parser.Show) error {
	return c.handleShowVariableKind(stmt, "STATUS")
}

func (c *conn) handleShowVariables(stmt *parser.Show) error {
	return c.handleShowVariableKind(stmt, "VARIABLES")
}
