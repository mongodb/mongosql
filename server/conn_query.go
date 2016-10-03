package server

import (
	"runtime/debug"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"gopkg.in/tomb.v2"
)

func (c *conn) handleCommand(stmt parser.Statement) error {
	executor, err := c.server.eval.EvaluateCommand(stmt, c)
	if err != nil {
		return err
	}

	if err = executor.Run(); err != nil {
		return err
	}

	c.writeOK(nil)
	return nil
}

func (c *conn) handleQuery(sql string) (err error) {
	if !c.tomb.Alive() {
		c.tomb = &tomb.Tomb{}
	}

	defer func() {
		if e := recover(); e != nil {
			c.logger.Logf(log.Info, "%s\n", debug.Stack())
			err = mysqlerrors.Unknownf("execute %s error %v", sql, e)
			return
		}
	}()

	sql = strings.TrimRight(sql, ";")

	startTime := time.Now()

	logTimeTaken := func() {
		c.logger.Logf(log.Info, "done executing query in %vms", time.Now().Sub(startTime).Nanoseconds()/1000000)
	}

	c.logger.Logf(log.DebugLow, `parsing "%s"`, sql)

	var stmt parser.Statement

	stmt, err = parser.Parse(sql)
	if err != nil {
		// This is an ugly hack such that if someone tries to set some parameter to the default
		// ignore.  This is because the sql parser barfs.  We should probably fix there for reals.
		sqlUpper := strings.ToUpper(sql)
		if len(sqlUpper) > 3 && sqlUpper[0:4] == "SET " {
			if len(sqlUpper) > 7 && sqlUpper[len(sqlUpper)-8:] == "=DEFAULT" {
				// wow, this is ugly
				logTimeTaken()
				return c.writeOK(nil)
			}
		}
		logTimeTaken()
		return mysqlerrors.Newf(mysqlerrors.ER_PARSE_ERROR, `parse sql '%s' error: %s`, sql, err)
	}

	switch v := stmt.(type) {
	case *parser.Select:
		err = c.handleSelect(v, sql, nil)
		logTimeTaken()
	case *parser.SimpleSelect:
		err = c.handleSimpleSelect(sql, v)
		logTimeTaken()
	case *parser.Show:
		err = c.handleShow(sql, v)
	case *parser.DDL:
		err = c.handleDDL(v)
	case *parser.Kill, *parser.Set:
		err = c.handleCommand(stmt)
		logTimeTaken()
	case *parser.Explain:
		err = c.handleExplain(sql, v)
	default:
		err = mysqlerrors.Unknownf("statement %T not supported", stmt)
	}

	return err
}
