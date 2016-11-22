package server

import "github.com/10gen/sqlproxy/parser"

func (c *conn) handleUse(use *parser.Use) error {
	if err := c.useDB(string(use.DBName)); err != nil {
		return err
	}
	return c.writeOK(nil)
}
