package server

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/metrics"
	"github.com/10gen/sqlproxy/parser"
)

const (
	// NoMemoryManagerFailpoint is the name of an environment variable that can be set to
	// instruct the BI Connector to return an error if memory is not precisely released
	// after a query completes execution.
	NoMemoryManagerFailpoint = "SQLPROXY_MEMORY_MANAGER_FAILPOINT_OFF"
)

func (c *conn) handleCommand(ctx context.Context, stmt parser.Statement) error {
	aCfg := c.getAlgebrizerConfig(parser.String(stmt), stmt)
	eCfg := c.getExecutionConfig()

	err := evaluator.EvaluateCommand(ctx, aCfg, eCfg)
	if err != nil {
		return err
	}

	return c.writeOK(nil)
}

func (c *conn) handleQuery(ctx context.Context, sql string) (err error) {
	if err = c.session.Validate(); err != nil {
		c.close(ctx)
		return err
	}

	var trackedStmt parser.Statement
	var planStats *evaluator.PlanStats

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
		latencyMs := time.Since(startTime).Nanoseconds() / 1000000
		c.logger.Infof(log.Admin, "done executing query in %vms", latencyMs)

		if trackedStmt != nil {

			mongoVersion := strings.Join(strings.Split(c.variables.GetString(variable.MongoDBVersion), ".")[:2], ".")
			biVersion := c.variables.GetString(variable.VersionComment)[10:]

			record, recErr := metrics.NewRecord(trackedStmt, mongoVersion, biVersion, planStats, latencyMs)
			if recErr != nil {
				c.logger.Errf(log.Always, "failed to build metrics record: %v", err)
			}

			c.server.enqueueRecord(record)
		}
	}()

	switch v := stmt.(type) {
	case *parser.Use:
		err = c.handleUse(v)
	case *parser.Select, *parser.SimpleSelect, *parser.Union:
		planStats, err = c.handleSelect(ctx, sql, v)
		if err == nil {
			trackedStmt = v
		} else if err == context.DeadlineExceeded {
			err = mysqlerrors.Defaultf(mysqlerrors.ErQueryTimeout)
		}
	case *parser.Show:
		err = c.handleShow(ctx, sql, v)
	case *parser.DropTable:
		err = c.handleDropTable(v)
	case *parser.AlterTable, *parser.Flush, *parser.Kill, *parser.RenameTable, *parser.Set:
		err = c.handleCommand(ctx, stmt)
	case *parser.Explain:
		err = c.handleExplain(ctx, sql, v)
	default:
		err = mysqlerrors.Unknownf("statement %T not supported", stmt)
	}

	if c.session.Err() != nil {
		c.close(ctx)
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

func (c *conn) getAlgebrizerConfig(sql string, stmt parser.Statement) *evaluator.AlgebrizerConfig {
	lg := c.Logger(log.AlgebrizerComponent)
	return evaluator.NewAlgebrizerConfig(lg, sql, stmt, c.DB(), c.catalog)
}

func (c *conn) getOptimizerConfig() *evaluator.OptimizerConfig {
	lg := c.Logger(log.OptimizerComponent)
	vars := c.variables
	eCfg := c.getExecutionConfig()
	return evaluator.NewOptimizerConfig(lg, vars, eCfg)
}

func (c *conn) getPushdownConfig() *evaluator.PushdownConfig {
	lg := c.Logger(log.OptimizerComponent)
	vars := c.variables
	return evaluator.NewPushdownConfig(lg, vars)
}

func (c *conn) getExecutionConfig() *evaluator.ExecutionConfig {
	lg := c.Logger(log.EvaluatorComponent)
	vars := c.variables
	cmds := c.getCommandHandler()
	mem := c.memoryMonitor
	dbName := c.DB()
	connID := uint64(c.connectionID)
	user := c.user
	remoteHost := c.remoteHost()
	return evaluator.NewExecutionConfig(lg, vars, cmds, mem, dbName, connID, user, remoteHost)
}
