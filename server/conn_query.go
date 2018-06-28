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
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

const (
	// NoMemoryManagerFailpoint is the name of an environment variable that can be set to
	// instruct the BI Connector to return an error if memory is not precisely released
	// after a query completes execution.
	NoMemoryManagerFailpoint = "SQLPROXY_MEMORY_MANAGER_FAILPOINT_OFF"
)

func (c *conn) handleCommand(stmt parser.Statement) error {
	executor, err := evaluator.EvaluateCommand(stmt, c)
	if err != nil {
		return err
	}

	if err = executor.Run(); err != nil {
		return err
	}

	return c.writeOK(nil)
}

func (c *conn) handleQuery(sql string) (err error) {
	if err = c.session.Validate(); err != nil {
		c.close()
		return err
	}

	defer func() {
		if e := recover(); e != nil {
			c.logger.Errf(log.Dev, "query execution error: %s\n%s\n", e, debug.Stack())
			err = mysqlerrors.Unknownf("execute %s error %v", sql, e)
			return
		}
	}()

	profile := c.server.cfg.Debug.ProfileScope
	if c.server.cfg.Debug.EnableProfiling == "cpu" && profile == "queries" {
		runtime.SetCPUProfileRate(100000)

		filename := fmt.Sprintf("query_%s.pprof", time.Now().Format("2006-01-02-15-04-05.000000"))
		var f *os.File
		f, err = os.Create(filename)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %s", err)
		}

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
		return mysqlerrors.Newf(mysqlerrors.ErParseError, `parse sql '%s' error: %s`, sql, err)
	}

	startTime := time.Now()

	defer func() {
		cErr := c.cleanupMemory()
		if err == nil {
			err = cErr
		}
		c.logger.Infof(log.Admin, "done executing query in %vms",
			time.Since(startTime).Nanoseconds()/1000000)
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

func (c *conn) cleanupMemory() error {
	peakAllocatedDuringQuery := c.memoryMonitor.PeakAllocated()
	c.logger.Debugf(log.Admin, "%s peak allocated", util.ByteString(peakAllocatedDuringQuery))

	allocated, memErr := c.memoryMonitor.Clear()
	if memErr != nil {
		c.logger.Debugf(log.Admin, "%v", memErr)
	}
	if allocated > 0 {
		c.logger.Debugf(log.Admin, "%s released", util.ByteString(allocated))
	}
	if allocated = c.memoryMonitor.Allocated(); allocated != 0 {
		if os.Getenv(NoMemoryManagerFailpoint) != "" {
			return fmt.Errorf("didn't release %s of memory", util.ByteString(allocated))
		}
	}
	return nil
}
