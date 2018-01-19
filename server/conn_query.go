package server

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
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

	defer func() {
		if e := recover(); e != nil {
			c.logger.Errf(log.Dev, "query execution error\n%s\n", debug.Stack())
			err = mysqlerrors.Unknownf("execute %s error %v", sql, e)
			return
		}
	}()

	profile := c.server.cfg.Debug.ProfileScope
	if c.server.cfg.Debug.EnableProfiling == "cpu" && profile == "queries" {
		runtime.SetCPUProfileRate(100000)

		filename := fmt.Sprintf("query_%s.pprof", time.Now().Format("2006-01-02-15-04-05.000000"))
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %s", err)
		}
		defer f.Close()

		err = pprof.StartCPUProfile(f)
		if err != nil {
			return fmt.Errorf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	sql = strings.TrimRight(sql, ";")

	c.logger.Infof(log.Dev, `parsing "%s"`, sql)

	var stmt parser.Statement

	stmt, err = parser.Parse(sql)
	if err != nil {
		return mysqlerrors.Newf(mysqlerrors.ER_PARSE_ERROR, `parse sql '%s' error: %s`, sql, err)
	}

	startTime := time.Now()

	defer func() {
		c.logger.Infof(log.Admin, "done executing query in %vms",
			time.Now().Sub(startTime).Nanoseconds()/1000000)
	}()

	switch v := stmt.(type) {
	case *parser.Use:
		err = c.handleUse(v)
	case *parser.Select, *parser.SimpleSelect, *parser.Union:
		err = c.handleSelect(sql, v)
	case *parser.Show:
		err = c.handleShow(sql, v)
	case *parser.DropTable:
		err = c.handleDropTable(v)
	case *parser.AlterTable, *parser.Flush, *parser.Kill, *parser.RenameTable, *parser.Set:
		err = c.handleCommand(stmt)
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
