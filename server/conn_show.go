package server

import (
	"github.com/10gen/sqlproxy/parser"
	"strings"
)

func (c *conn) handleShow(sql string, stmt *parser.Show) error {
	switch strings.ToLower(stmt.Section) {
	case "charset", "collation", "columns", "create database", "create table",
		"databases", "index", "indexes", "keys", "processlist", "schemas", "status", "tables", "variables":
		return c.handleSelect(sql, stmt)
	default:
		return c.handleShowNotImplemented(sql, stmt)
	}
}
