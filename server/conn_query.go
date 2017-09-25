package server

import (
	"runtime/debug"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

func (c *conn) handleCommand(stmt parser.Statement) error {
	executor, err := evaluator.EvaluateCommand(stmt, c)
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
	if err := c.session.Validate(); err != nil {
		c.close()
		return err
	}

	select {
	case <-c.ctx.Done():
		c.refreshContext()
	default:
	}

	defer func() {
		if e := recover(); e != nil {
			c.logger.Errf(log.Dev, "query execution error\n%s\n", debug.Stack())
			err = mysqlerrors.Unknownf("execute %s error %v", sql, e)
			return
		}
	}()

	sql = strings.TrimRight(sql, ";")

	startTime := time.Now()

	logTimeTaken := func() {
		c.logger.Infof(log.Admin, "done executing query in %vms", time.Now().Sub(startTime).Nanoseconds()/1000000)
	}

	c.logger.Infof(log.Dev, `parsing "%s"`, sql)

	var stmt parser.Statement

	stmt, err = parser.Parse(sql)
	if err != nil {
		return mysqlerrors.Newf(mysqlerrors.ER_PARSE_ERROR, `parse sql '%s' error: %s`, sql, err)
	}

	switch v := stmt.(type) {
	case *parser.Use:
		err = c.handleUse(v)
	case *parser.Select:
		err = c.handleSelect(sql, v)
		logTimeTaken()
	case *parser.SimpleSelect:
		err = c.handleSelect(sql, v)
		logTimeTaken()
	case *parser.Union:
		err = c.handleSelect(sql, v)
		logTimeTaken()
	case *parser.Show:
		err = c.handleShow(sql, v)
	case *parser.DropTable:
		err = c.handleDropTable(v)
	case *parser.Flush, *parser.Kill, *parser.Set:
		err = c.handleCommand(stmt)
		logTimeTaken()
	case *parser.Explain:
		err = c.handleExplain(sql, v)
	default:
		err = mysqlerrors.Unknownf("statement %T not supported", stmt)
	}

	if c.session.Err() != nil {
		c.close()
	}

	return err
}
