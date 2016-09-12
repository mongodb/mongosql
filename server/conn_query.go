package server

import (
	"runtime/debug"
	"strings"
	"time"

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
			log.Logf(log.Always, "[conn%v] %s\n", c.connectionID, debug.Stack())
			err = mysqlerrors.Unknownf("execute %s error %v", sql, e)
			return
		}
	}()

	sql = strings.TrimRight(sql, ";")

	startTime := time.Now()

	log.Logf(log.DebugLow, `[conn%v] parsing "%s"`, c.connectionID, sql)

	var stmt parser.Statement

	stmt, err = parser.Parse(sql)
	if err != nil {
		// This is an ugly hack such that if someone tries to set some parameter to the default
		// ignore.  This is because the sql parser barfs.  We should probably fix there for reals.
		sqlUpper := strings.ToUpper(sql)
		if len(sqlUpper) > 3 && sqlUpper[0:4] == "SET " {
			if len(sqlUpper) > 7 && sqlUpper[len(sqlUpper)-8:] == "=DEFAULT" {
				// wow, this is ugly
				log.Logf(log.DebugLow, "[conn%v] done executing plan in %v", c.ConnectionId(), time.Now().Sub(startTime))
				return c.writeOK(nil)
			}
		}
		log.Logf(log.DebugLow, "[conn%v] done executing plan in %v", c.ConnectionId(), time.Now().Sub(startTime))
		return mysqlerrors.Newf(mysqlerrors.ER_PARSE_ERROR, `parse sql '%s' error: %s`, sql, err)
	}

	switch v := stmt.(type) {
	case *parser.Select:
		err = c.handleSelect(v, sql, nil)
		log.Logf(log.DebugLow, "[conn%v] done executing plan in %v", c.ConnectionId(), time.Now().Sub(startTime))
	case *parser.SimpleSelect:
		err = c.handleSimpleSelect(sql, v)
		log.Logf(log.DebugLow, "[conn%v] done executing plan in %v", c.ConnectionId(), time.Now().Sub(startTime))
	case *parser.Show:
		err = c.handleShow(sql, v)
	case *parser.DDL:
		err = c.handleDDL(v)
	case *parser.Kill, *parser.Set:
		err = c.handleCommand(stmt)
		log.Logf(log.DebugLow, "[conn%v] done executing plan in %v", c.ConnectionId(), time.Now().Sub(startTime))
	default:
		err = mysqlerrors.Unknownf("statement %T not supported", stmt)
	}

	return err
}
