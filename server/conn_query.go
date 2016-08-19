package server

import (
	"runtime/debug"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/tomb.v2"
)

func (c *conn) handleCommand(stmt parser.Statement) error {
	executor, err := c.server.eval.EvaluateCommand(stmt, c)
	if err != nil {
		return err
	}

	err = executor.Run()

	log.Logf(log.DebugLow, "[conn%v] done executing plan", c.ConnectionId())

	if err != nil {
		return err
	}

	c.writeOK(nil)
	return nil
}

func (c *conn) handleQuery(sql string) (err error) {
	if !c.tomb.Alive() {
		c.tomb = &tomb.Tomb{}
	}

	log.Logf(log.DebugLow, "[conn%v] %s\n", c.connectionID, sql)

	defer func() {
		if e := recover(); e != nil {
			log.Logf(log.Always, "[conn%v] %s\n", c.connectionID, debug.Stack())
			err = mysqlerrors.Unknownf("execute %s error %v", sql, e)
			return
		}
	}()

	sql = strings.TrimRight(sql, ";")

	var stmt parser.Statement
	stmt, err = parser.Parse(sql)
	if err != nil {

		// This is an ugly hack such that if someone tries to set some parameter to the default
		// ignore.  This is because the sql parser barfs.  We should probably fix there for reals.
		sqlUpper := strings.ToUpper(sql)
		if len(sqlUpper) > 3 && sqlUpper[0:4] == "SET " {
			if len(sqlUpper) > 7 && sqlUpper[len(sqlUpper)-8:] == "=DEFAULT" {
				// wow, this is ugly
				return c.writeOK(nil)
			}
		}

		return mysqlerrors.Newf(mysqlerrors.ER_PARSE_ERROR, `parse sql '%s' error: %s`, sql, err)
	}

	switch v := stmt.(type) {
	case *parser.Select:
		return c.handleSelect(v, sql, nil)
	case *parser.SimpleSelect:
		return c.handleSimpleSelect(sql, v)
	case *parser.Show:
		return c.handleShow(sql, v)
	case *parser.DDL:
		return c.handleDDL(v)
	case *parser.Kill, *parser.Set:
		return c.handleCommand(stmt)
	default:
		return mysqlerrors.Unknownf("statement %T not supported", stmt)
	}
}
