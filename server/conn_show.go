package server

import (
	"strings"

	"github.com/10gen/sqlproxy/parser"
)

func (c *conn) handleShow(sql string, stmt *parser.Show) error {
	switch strings.ToLower(stmt.Section) {
	case "charset", "collation", "columns", "create table",
		"databases", "schemas", "status", "tables", "variables":
		fields, iter, err := c.server.eval.EvaluateQuery(stmt, c)
		if err != nil {
			return err
		}
		return c.streamResultset(fields, iter)
	default:
		return c.handleShowNotImplemented(sql, stmt)
	}
}
