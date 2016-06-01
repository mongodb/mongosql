package server

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/deafgoat/mixer/sqlparser"
)

func (c *conn) handleShow(sql string, stmt *sqlparser.Show) error {
	switch strings.ToLower(stmt.Section) {
	case "columns":
		return c.handleShowColumns(sql, stmt)
	case "databases", "schemas":
		return c.handleShowDatabases(stmt)
	case "tables":
		return c.handleShowTables(sql, stmt)
	case "variables":
		return c.handleShowVariables(sql, stmt)
	default:
		return mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "no support for show (%s) for now", sql)
	}
}

func (c *conn) handleShowColumns(sql string, stmt *sqlparser.Show) error {

	translated :=
		"SELECT COLUMN_NAME AS `Field`" +
			", COLUMN_TYPE AS `Type`" +
			", IS_NULLABLE AS `Null`" +
			", COLUMN_KEY AS `Key`" +
			", COLUMN_DEFAULT AS `Default`" +
			", EXTRA AS `Extra`"

	if strings.ToLower(stmt.Modifier) == "full" {
		translated += ", COLLATION_NAME AS `Collation`" +
			", PRIVILEGES AS `Privileges`" +
			", COLUMN_COMMENT AS `Comment`"
	}

	translated += " FROM INFORMATION_SCHEMA.COLUMNS"

	var dbName string
	if c.currentDB != nil {
		dbName = c.currentDB.Name
	}

	table := ""

	switch f := stmt.From.(type) {
	case sqlparser.StrVal:
		table = string(f)
	case *sqlparser.ColName:
		if f.Qualifier != nil {
			dbName = string(f.Qualifier)
		}
		table = string(f.Name)
	default:
		return mysqlerrors.Defaultf(mysqlerrors.ER_ILLEGAL_VALUE_FOR_TYPE, "FROM", sqlparser.String(f))
	}

	if stmt.DBFilter != nil {
		switch f := stmt.DBFilter.(type) {
		case sqlparser.StrVal:
			dbName = string(f)
		case *sqlparser.ColName:
			dbName = string(f.Name)
		default:
			return mysqlerrors.Defaultf(mysqlerrors.ER_ILLEGAL_VALUE_FOR_TYPE, "FROM", sqlparser.String(f))
		}
	}

	translated += fmt.Sprintf(" WHERE TABLE_NAME = '%s'", table)

	if dbName != "" {
		translated += fmt.Sprintf(" AND TABLE_SCHEMA = '%s'", dbName)
	}

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case sqlparser.StrVal:
			translated += fmt.Sprintf(" AND COLUMN_NAME LIKE %s", sqlparser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" AND %s", sqlparser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY ORDINAL_POSITION"

	return c.handleQuery(translated)
}

func (c *conn) handleShowDatabases(stmt *sqlparser.Show) error {
	translated := "SELECT SCHEMA_NAME AS `Database`" +
		" FROM INFORMATION_SCHEMA.SCHEMATA"

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case sqlparser.StrVal:
			translated += fmt.Sprintf(" WHERE SCHEMA_NAME LIKE %s", sqlparser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" WHERE %s", sqlparser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY SCHEMA_NAME"

	return c.handleQuery(translated)
}

func (c *conn) handleShowTables(sql string, stmt *sqlparser.Show) error {
	dbName := ""
	if c.currentDB != nil {
		dbName = c.currentDB.Name
	}

	if stmt.From != nil {
		switch f := stmt.From.(type) {
		case sqlparser.StrVal:
			dbName = string(f)
		case *sqlparser.ColName:
			dbName = string(f.Name)
		default:
			return mysqlerrors.Defaultf(mysqlerrors.ER_ILLEGAL_VALUE_FOR_TYPE, "FROM", sqlparser.String(f))
		}
	}

	if dbName == "" {
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_DB_ERROR)
	}

	translated := fmt.Sprintf("SELECT TABLE_NAME AS `Tables_in_%s`", dbName)

	if strings.ToLower(stmt.Modifier) == "full" {
		translated += ", TABLE_TYPE AS `Type`"
	}

	translated += " FROM INFORMATION_SCHEMA.TABLES" +
		fmt.Sprintf(" WHERE TABLE_SCHEMA = '%s'", dbName)

	if stmt.LikeOrWhere != nil {
		switch stmt.LikeOrWhere.(type) {
		case sqlparser.StrVal:
			translated += fmt.Sprintf(" AND TABLE_NAME LIKE %s", sqlparser.String(stmt.LikeOrWhere))
		default:
			translated += fmt.Sprintf(" AND %s", sqlparser.String(stmt.LikeOrWhere))
		}
	}

	translated += " ORDER BY TABLE_NAME"

	return c.handleQuery(translated)
}

func (c *conn) handleShowVariables(sql string, stmt *sqlparser.Show) error {
	variables := []interface{}{}
	r, err := c.buildSimpleShowResultset(variables, "Variable")
	if err != nil {
		return err
	}
	return c.writeResultset(c.status, r)
}

func (c *conn) buildSimpleShowResultset(values []interface{}, name string) (*Resultset, error) {

	r := new(Resultset)

	field := &Field{}

	field.Name = Slice(name)
	field.Charset = 33
	field.Type = MYSQL_TYPE_VAR_STRING

	r.Fields = []*Field{field}

	var row []byte
	var err error

	for _, value := range values {
		row, err = formatValue(value)
		if err != nil {
			return nil, err
		}
		r.RowDatas = append(r.RowDatas, putLengthEncodedString(row))
	}

	return r, nil
}
