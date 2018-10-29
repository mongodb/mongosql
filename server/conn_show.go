package server

import (
	"context"
	"strings"

	"github.com/10gen/sqlproxy/parser"
)

func (c *conn) handleShow(ctx context.Context, sql string, stmt *parser.Show) error {
	switch strings.ToLower(stmt.Section) {
	case "charset", "collation", "columns", "create database", "create table",
		"databases", "index", "indexes", "keys", "processlist", "schemas", "status", "tables",
		"variables":
		return c.handleSelect(ctx, sql, stmt)
	default:
		return c.handleShowNotImplemented(sql, stmt)
	}
}
