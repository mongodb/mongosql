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
	"github.com/10gen/sqlproxy/evaluator/metrics"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

const (
	// NoMemoryManagerFailpoint is the name of an environment variable that can be set to
	// instruct the BI Connector to return an error if memory is not precisely released
	// after a query completes execution.
	NoMemoryManagerFailpoint = "SQLPROXY_MEMORY_MANAGER_FAILPOINT_OFF"
)

func (c *conn) handleQuery(ctx context.Context, sql string) (err error) {

	var res *evaluator.QueryResult
	var trackedStmt parser.Statement
	var planStats *evaluator.PlanStats

	if err = c.session.Validate(ctx); err != nil {
		c.close(ctx)
		return err
	}

	defer func() {
		if c.session.Err() != nil {
			c.close(ctx)
		}
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

	startTime := time.Now()

	defer func() {
		if err == context.DeadlineExceeded {
			err = mysqlerrors.Defaultf(mysqlerrors.ErQueryTimeout)
		}
		cErr := c.cleanupMemory()
		if err == nil {
			err = cErr
		}
		latencyMs := time.Since(startTime).Nanoseconds() / 1000000
		c.logger.Infof(log.Admin, "done executing query in %vms", latencyMs)

		if trackedStmt != nil {

			mongoVersion := strings.Join(strings.Split(c.variables.GetString(variable.MongoDBVersion), ".")[:2], ".")
			biVersion := c.variables.GetString(variable.MongosqldVersion)

			if c.variables.GetBool(variable.AnonymizeMetrics) {
				trackedStmt = parser.AnonymizeStatement(trackedStmt)
			}

			record, recErr := metrics.NewRecord(trackedStmt, mongoVersion, biVersion, planStats, latencyMs)
			if recErr != nil {
				c.logger.Errf(log.Always, "failed to build metrics record: %v", err)
			}

			c.server.enqueueRecord(record)
		}
	}()

	lg := c.Logger(log.EvaluatorComponent)

	rCfg := c.getRewriterConfig()
	aCfg := c.getAlgebrizerConfig()
	oCfg := c.getOptimizerConfig()
	pCfg := c.getPushdownConfig()
	eCfg := c.getExecutionConfig()
	qCfg := evaluator.NewQueryConfig(lg, rCfg, aCfg, oCfg, pCfg, eCfg)

	var queryCtx context.Context
	maxTimeMS := c.variables.GetInt64(variable.MaxTimeMS)
	// When the user has supplied a max execution time we create a time bounded context for
	// the query so that the query will be cancelled if the time deadline is reached.
	// A MaxTimeMS of `0` means no max time set.
	if maxTimeMS > 0 {
		var cancelQueryCtx context.CancelFunc
		queryCtx, cancelQueryCtx = context.WithTimeout(ctx, time.Duration(maxTimeMS*int64(time.Millisecond)))
		defer cancelQueryCtx()
	} else {
		queryCtx = ctx
	}

	res, err = evaluator.ExecuteSQL(queryCtx, qCfg, sql)

	if err != nil {
		return err
	}

	streamMemoryIter := func() error {
		iter, err := res.GetRowIter(ctx, eCfg, evaluator.NewExecutionState())
		if err != nil {
			return err
		}
		memIter := evaluator.NewMemoryIter(eCfg, iter)
		return c.streamRowResultset(queryCtx, res.Columns, memIter)
	}

	streamRowIter := func() error {
		iter, err := res.GetRowIter(ctx, eCfg, evaluator.NewExecutionState())
		if err != nil {
			return err
		}
		return c.streamRowResultset(queryCtx, res.Columns, iter)
	}

	switch res.Op {
	case evaluator.COMMAND:
		return c.writeOK(nil)
	case evaluator.SHOWNOTIMPL:
		v := res.Stmt.(*parser.Show)
		return c.handleShowNotImplemented(sql, v)
	case evaluator.SHOW:
		planStats = res.Stats
		return streamRowIter()
	case evaluator.QUERY:
		planStats = res.Stats
		trackedStmt = res.Stmt
		docIter, _ := res.GetDocIter(ctx, eCfg, evaluator.NewExecutionState())
		if docIter != nil {
			return c.streamDocResultset(queryCtx, res.Columns, docIter)
		}
		return streamMemoryIter()
	case evaluator.EXPLAIN:
		return streamRowIter()
	default:
		return mysqlerrors.Unknownf("query result type %T not supported", res.Op)
	}
}

func (c *conn) cleanupMemory() error {
	peakAllocatedDuringQuery := c.memoryMonitor.PeakAllocated()
	c.logger.Debugf(log.Admin, "%s peak allocated", strutil.ByteString(peakAllocatedDuringQuery))

	allocated, memErr := c.memoryMonitor.Clear()
	if memErr != nil {
		c.logger.Debugf(log.Admin, "%v", memErr)
	}
	if allocated > 0 {
		c.logger.Debugf(log.Admin, "%s released", strutil.ByteString(allocated))
	}
	if allocated = c.memoryMonitor.Allocated(); allocated != 0 {
		if os.Getenv(NoMemoryManagerFailpoint) != "" {
			return fmt.Errorf("didn't release %s of memory", strutil.ByteString(allocated))
		}
	}
	return nil
}

func (c *conn) getRewriterConfig() *evaluator.RewriterConfig {
	connID := uint64(c.connectionID)
	lg := c.Logger(log.RewriterComponent)
	dbName := c.DB()
	rewriteDistinctAsGroup := c.variables.GetBool(variable.RewriteDistinctAsGroup)
	mysqlVersion := c.variables.GetString(variable.Version)
	remoteHost := c.remoteHost()
	user := c.user

	return evaluator.NewRewriterConfig(connID, dbName, lg,
		rewriteDistinctAsGroup, mysqlVersion, remoteHost, user)
}

func (c *conn) getAlgebrizerConfig() *evaluator.AlgebrizerConfig {
	lg := c.Logger(log.AlgebrizerComponent)
	dbName := c.DB()
	ctlg := c.catalog
	vars := c.variables
	mongoDBToplogy := c.variables.GetString(variable.MongoDBTopology)
	isWriteMode := c.server.cfg.Schema.WriteMode
	sqlValueKind := evaluator.GetSQLValueKind(c.variables)
	sqlSelectLimit := c.variables.GetUint64(variable.SQLSelectLimit)
	mongoDBMaxVarcharLength := c.variables.GetUint64(variable.MongoDBMaxVarcharLength)
	groupConcatMaxLen := c.variables.GetInt64(variable.GroupConcatMaxLen)
	polymorphicTypeConversionMode := c.variables.GetString(variable.PolymorphicTypeConversionMode)
	mdbVersion := evaluator.GetMongoDBVersion(c.variables)

	return evaluator.NewAlgebrizerConfig(lg, dbName, ctlg, vars, mongoDBToplogy, isWriteMode,
		sqlValueKind, sqlSelectLimit, mongoDBMaxVarcharLength, groupConcatMaxLen,
		polymorphicTypeConversionMode, mdbVersion, true, false)
}

func (c *conn) getOptimizerConfig() *evaluator.OptimizerConfig {
	lg := c.Logger(log.OptimizerComponent)
	collation := c.variables.GetCollation(variable.CollationConnection)
	sqlValueKind := evaluator.GetSQLValueKind(c.variables)
	optimizeCrossJoins := c.variables.GetBool(variable.OptimizeCrossJoins)
	optimizeEvaluations := c.variables.GetBool(variable.OptimizeEvaluations)
	optimizeFiltering := c.variables.GetBool(variable.OptimizeFiltering)
	optimizeInnerJoins := c.variables.GetBool(variable.OptimizeInnerJoins)
	reconcileArithmeticAggFunctions := c.variables.GetBool(variable.ReconcileArithmeticAggFunctions)

	return evaluator.NewOptimizerConfig(lg, collation, sqlValueKind, optimizeCrossJoins,
		optimizeEvaluations, optimizeFiltering, optimizeInnerJoins, reconcileArithmeticAggFunctions)
}

func (c *conn) getPushdownConfig() *evaluator.PushdownConfig {
	lg := c.Logger(log.OptimizerComponent)
	mdbVersion := evaluator.GetMongoDBVersion(c.variables)
	shouldPushDown := c.variables.GetBool(variable.Pushdown)
	pushDownSelfJoins := c.variables.GetBool(variable.OptimizeSelfJoins)
	sqlValueKind := evaluator.GetSQLValueKind(c.variables)
	format := evaluator.NoOutputFormat
	formatVersion := evaluator.NoOutputVersion

	return evaluator.NewPushdownConfig(lg, mdbVersion, false, false, true, shouldPushDown,
		pushDownSelfJoins, sqlValueKind, format, formatVersion)
}

func (c *conn) getExecutionConfig() *evaluator.ExecutionConfig {
	lg := c.Logger(log.ExecutorComponent)
	dbName := c.DB()
	mdbVersion := evaluator.GetMongoDBVersion(c.variables)
	fullPushdownOnly := c.variables.GetBool(variable.FullPushdownExecMode)
	maxStageSize := c.variables.GetUint64(variable.MongoDBMaxStageSize)
	sqlValueKind := evaluator.GetSQLValueKind(c.variables)
	cmds := c.getCommandHandler()
	mem := c.memoryMonitor

	return evaluator.NewExecutionConfig(lg, dbName, mdbVersion, fullPushdownOnly,
		maxStageSize, sqlValueKind, cmds, mem)
}
