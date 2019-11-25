package evaluator

import (
	"context"
	"errors"
	"math/rand"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
)

// QueryConfig is a container for all the values needed to process a SQL query.
type QueryConfig struct {
	lg   log.Logger
	rCfg *RewriterConfig
	aCfg *AlgebrizerConfig
	oCfg *OptimizerConfig
	pCfg *PushdownConfig
	eCfg *ExecutionConfig
}

// NewQueryConfigFromCatalog returns a new default QueryConfig.
func NewQueryConfigFromCatalog(defaultDbName string, ctlg catalog.Catalog, unflattenResults bool) *QueryConfig {
	lgr := log.GlobalLogger()
	vars := ctlg.Variables()
	mySQLVersion := getMySQLVersion(vars)
	rCfg := NewRewriterConfig(uint64(0), defaultDbName, lgr, false, mySQLVersion, "localhost", "user")
	// Default config will be writeMode = false
	aCfg := NewAlgebrizerConfig(lgr, defaultDbName, ctlg, false)
	eCfg := NewExecutionConfig(lgr, vars, errCommandHandler{}, nil, defaultDbName)
	oCfg := NewOptimizerConfig(lgr, vars)
	pCfg := NewPushdownConfig(lgr, vars, unflattenResults)

	return NewQueryConfig(lgr, rCfg, aCfg, oCfg, pCfg, eCfg)
}

// NewQueryConfig returns a new QueryConfig.
func NewQueryConfig(lg log.Logger, rCfg *RewriterConfig, aCfg *AlgebrizerConfig,
	oCfg *OptimizerConfig, pCfg *PushdownConfig, eCfg *ExecutionConfig) *QueryConfig {
	return &QueryConfig{
		lg:   lg,
		rCfg: rCfg,
		aCfg: aCfg,
		oCfg: oCfg,
		pCfg: pCfg,
		eCfg: eCfg,
	}
}

// ExecuteSQL parses a query or command, evaluates it, and returns a *QueryResult to the server.
func ExecuteSQL(ctx context.Context, qCfg *QueryConfig, sql string) (*QueryResult, error) {
	qCfg.lg.Infof(log.Dev, "parsing %q", sql)

	stmt, err := parser.Parse(sql)
	if err != nil {
		return nil, mysqlerrors.Newf(mysqlerrors.ErParseError, `parse sql '%s' error: %s`, sql, err)
	}
	return executeSQLStatement(ctx, qCfg, stmt)
}

// executeSQLStatement executes a preparsed sql Statement.
func executeSQLStatement(ctx context.Context, qCfg *QueryConfig, stmt parser.Statement) (*QueryResult, error) {
	qCfg.lg.Infof(log.Dev, "generating plan for sql...")

	rewritten, err := RewriteStatement(qCfg.rCfg, stmt)
	if err = procutil.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	switch v := rewritten.(type) {
	case *parser.Select, *parser.Union:
		return EvaluateQuery(ctx, qCfg, v)
	case *parser.Show:
		return EvaluateShow(ctx, qCfg, v)
	case *parser.AlterTable, *parser.Flush, *parser.Kill,
		*parser.RenameTable, *parser.Set, *parser.Use,
		*parser.DropTable, *parser.DropDatabase,
		*parser.CreateTable, *parser.CreateDatabase,
		*parser.Insert:
		return EvaluateCommand(ctx, qCfg.rCfg, qCfg.aCfg, qCfg.eCfg, v)
	case *parser.Explain:
		switch strings.ToLower(v.Section) {
		case "plan":
			return handleExplainPlan(ctx, qCfg, v)
		default:
			// unreachable
			buf := parser.NewTrackedBuffer(nil)
			stmt.Format(buf)
			return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet, "no support for explain (%s) "+
				"for now", buf.String())
		}
	case *parser.IgnoredStatement:
		return handleIgnoredStatement(v)
	default:
		return nil, mysqlerrors.Unknownf("statement %T not supported", stmt)
	}
}

func handleExplainPlan(ctx context.Context, qCfg *QueryConfig, stmt *parser.Explain) (*QueryResult, error) {
	res, err := EvaluateExplain(ctx, qCfg, stmt.Statement)
	if err != nil {
		return nil, err
	}
	return res, err
}

func handleIgnoredStatement(statement *parser.IgnoredStatement) (*QueryResult, error) {
	return NewQueryResult(statement.Statement, nil, nil, nil, COMMAND), nil
}

// ExecutionConfig is a container for all the values needed to execute
// queries and perform in-memory evaluation.
type ExecutionConfig struct {
	lg               log.Logger
	dbName           string
	mongoDBVersion   []uint8
	fullPushdownOnly bool
	maxStageSize     uint64
	sqlValueKind     values.SQLValueKind

	commandHandler CommandHandler
	memoryMonitor  memory.Monitor
}

// NewExecutionConfig returns a new ExecutionConfig constructed from the
// provided values. ExecutionConfigs should always be constructed via this
// function instead of via a struct literal.
func NewExecutionConfig(lg log.Logger, vars catalog.VariableContainer, cmds CommandHandler, mem memory.Monitor, dbName string) *ExecutionConfig {
	return &ExecutionConfig{
		lg:               lg,
		commandHandler:   cmds,
		dbName:           dbName,
		mongoDBVersion:   getMongoDBVersion(vars),
		fullPushdownOnly: vars.GetBool(variable.FullPushdownExecMode),
		memoryMonitor:    mem,
		sqlValueKind:     GetSQLValueKind(vars),
		maxStageSize:     vars.GetUint64(variable.MongoDBMaxStageSize),
	}
}

// A CommandHandler provides implementations for any external actions
// that must be taken while evaluating queries and commands (querying
// MongoDB, rotating logs, etc).
type CommandHandler interface {
	// Aggregate runs the provided aggregation pipeline against the
	// specified database and collection.
	Aggregate(ctx context.Context, db, col string, pipeline []bson.D) (mongodb.Cursor, error)
	// Count runs a count command against the specified database and collection.
	Count(ctx context.Context, db, col string) (int, error)
	// DropTable supports dropping tables.
	DropTable(ctx context.Context, db, tbl string) error
	// DropDatabase drops databases.
	DropDatabase(ctx context.Context, db string) error
	// CreateTable supports creating tables.
	CreateTable(ctx context.Context, db string, table *schema.Table) error
	// CreateDatabase creates Databases.
	CreateDatabase(ctx context.Context, db string) error
	// Insert inserts documents into the specified namespace.
	Insert(ctx context.Context, db, table string, docs []interface{}) error
	// Kill kills a Connection or Query (the KillScope). The targetConnID is the
	// ID of the connection that is to be killed. The targetConnID may be the
	// current connection id.
	Kill(ctx context.Context, targetConnID uint32, ks KillScope) error
	// Resample forces a sample refresh. It must occur in the server
	// as that is where the schemata are maintained.
	Resample(context.Context) error
	// RotateLogs rotates the log file.
	RotateLogs() error
	// Set sets the value of the specified variable to the provided value.
	Set(variable.Name, variable.Scope, variable.Kind, values.SQLValue) error
	// SetDatabase sets the current database.
	SetDatabase(db string) error
	// SetScopeAuthorized returns an error if the user is not authorized to
	// set variables in the provided scope.
	SetScopeAuthorized(variable.Scope) error
	// UnsetDatabase unsets the current database.
	UnsetDatabase() error
}

// ExecutionState is a container for state that has to be shared between
// multiple parts of query execution. We want this struct to be as small
// and simple as possible, and should avoid adding to it.
type ExecutionState struct {
	// When an expression is being evaluated, rows contains the row(s) that
	// should be used to resolve SQLColumnExprs to values.
	// There will usually be just one row at a time, but there may be multiple
	// rows when a join or union is involved.
	rows []*results.Row
	// When evaluating a correlated subquery expression, correlatedRows
	// contains the rows from parent queries that should be used to
	// resolve correlated SQLColumnExprs to values.
	correlatedRows []*results.Row
	// Collation may differ from table to table. This field specifies which
	// collation should be used during evaluation.
	collation *collation.Collation
	// The RAND scalar function needs to be able to store and reuse random number
	// generators while evaluating a query. This is where they are stored.
	randomExprs map[uint64]*rand.Rand
}

// NewExecutionState returns a new ExecutionState initialized with default values.
func NewExecutionState() *ExecutionState {
	return &ExecutionState{
		rows:        []*results.Row{},
		randomExprs: make(map[uint64]*rand.Rand),
		collation:   collation.Default,
	}
}

// Random returns the random number generator with the provided id.
// If no random number generator with the provided id exists, it is created.
func (st *ExecutionState) Random(id uint64) *rand.Rand {
	r, ok := st.randomExprs[id]
	if ok {
		return r
	}

	src := rand.NewSource(rand.Int63())
	r = rand.New(src)
	st.randomExprs[id] = r
	return r
}

// RandomWithSeed returns the random number generator with the provided id if it exists.
// If no random number generator with that id exists, it is created with the provided seed.
func (st *ExecutionState) RandomWithSeed(id uint64, seed int64) *rand.Rand {
	r, ok := st.randomExprs[id]
	if ok {
		return r
	}

	src := rand.NewSource(seed)
	r = rand.New(src)
	st.randomExprs[id] = r
	return r
}

// SubqueryState returns a new ExecutionState with rows scoped
// for executing a correlated subquery.
func (st *ExecutionState) SubqueryState() *ExecutionState {
	cRows := []*results.Row{}
	cRows = append(cRows, st.rows...)
	cRows = append(cRows, st.correlatedRows...)
	return &ExecutionState{
		rows:           []*results.Row{},
		correlatedRows: cRows,
		randomExprs:    st.randomExprs,
		collation:      st.collation,
	}
}

// WithCollation returns a new ExecutionState with the provided collation.
func (st *ExecutionState) WithCollation(c *collation.Collation) *ExecutionState {
	return &ExecutionState{
		rows:           st.rows,
		randomExprs:    st.randomExprs,
		correlatedRows: st.correlatedRows,
		collation:      c,
	}
}

// WithRows returns a new ExecutionState with the provided rows.
func (st *ExecutionState) WithRows(rows ...*results.Row) *ExecutionState {
	return &ExecutionState{
		rows:           rows,
		randomExprs:    st.randomExprs,
		correlatedRows: st.correlatedRows,
		collation:      st.collation,
	}
}

// CheckPlanExecution is a watchdog for query execution that also possibly optimizes (as a hack)
// to a FastPlanStage in cases where the top level stage is not a FastPlanStage but could be:
// see the documentation for getFastPlanStage for more information on what stages can be
// optimized as FastPlanStages. It returns an error if, for some reason, we should not execute
// this query, e.g., fullPushdownOnly is set and this query is not pushedDown.
func CheckPlanExecution(ctx context.Context, cfg *ExecutionConfig, plan PlanStage) (PlanStage, error) {

	// If possible, return a fast iterator for this plan.
	mongodb32 := procutil.VersionExactly(cfg.mongoDBVersion, []uint8{3, 2})
	fastPlan, ok := getFastPlanStage(plan, mongodb32, false)
	if ok {

		cfg.lg.Debugf(log.Admin,
			"executing query plan with fast iterator: \n%v",
			PrettyPrintPlan(fastPlan),
		)
		return fastPlan, nil
	}

	// If full pushdown exec mode is enabled, don't execute this query unless
	// it is fully pushed down. We don't need this check above because a fast
	// query plan is considered to be fully pushed down.
	if cfg.fullPushdownOnly {
		err := IsFullyPushedDown(plan)
		if err != nil {
			return nil, err
		}
	}

	cfg.lg.Debugf(log.Admin,
		"executing query plan: \n%v",
		PrettyPrintPlan(plan),
	)

	return plan, nil
}

// getFastPlanStage returns a FastPlanStage and true if possible,
// otherwise nil and false. Also, remove any unncessary UnionDistincts,
// which are any UnionDistincts other another UnionDistinct.
// The parameter underDistinct tells us if we are below a UnionDistinct in
// the Plan, in which case all UnionDisticts should be replaced with UnionAll
// in order to improve performance: there is no reason to remove duplicates
// twice.
//
// is32 is true if the server versions is 3.2.x.
func getFastPlanStage(plan PlanStage, is32 bool, underDistinct bool) (FastPlanStage, bool) {
	if fastPlan, ok := plan.(*MongoSourceStage); ok {
		return fastPlan, true
	} else if projectPlan, ok := plan.(*ProjectStage); ok {
		if groupPlan, ok := projectPlan.source.(*GroupByStage); ok {
			if unionPlan, ok := groupPlan.source.(*UnionStage); ok {
				// The presence of a UnionDistinct under a GroupByStage
				// tells us the GroupByStage is just being used for uniqueness.
				// A GroupByStage above a UnionAll could have other uses.
				if unionPlan.kind == UnionDistinct {
					if left, ok := getFastPlanStage(unionPlan.left, is32, true); ok {
						if right, ok := getFastPlanStage(unionPlan.right, is32, true); ok {
							unionType := UnionDistinct
							localIs32 := is32
							if underDistinct {
								localIs32 = false
								unionType = UnionAll
							}
							// Note that we remove the project stages, which means
							// we need to create a new stage here just in case we
							// ultimately end up not able to generate a complete
							// FastPlanStage. If we modified the plan in place,
							// such a situation would result in an unusable plan.
							ret := NewUnionStage(unionType, left, right)
							ret.is32 = localIs32
							return ret, true
						}
					}
				}
			}
		} else if unionPlan, ok := projectPlan.source.(*UnionStage); ok {
			// A UnionDistinct should always be under a GroupByStage under
			// the way we currently generated plan stages, but this check
			// protects us against future changes.
			if unionPlan.kind != UnionAll {
				return nil, false
			}
			if left, ok := getFastPlanStage(unionPlan.left, is32, underDistinct); ok {
				if right, ok := getFastPlanStage(unionPlan.right, is32, underDistinct); ok {
					return NewUnionStage(UnionAll, left, right), true
				}
			}
		}
	}
	return nil, false
}

type errCommandHandler struct{}

// Aggregate runs the provided aggregation pipeline against the
// specified database and collection.
func (errCommandHandler) Aggregate(ctx context.Context, db, col string, pipeline []bson.D) (mongodb.Cursor, error) {
	return nil, errors.New("command not supported")
}

// Count runs a count command against the specified database and collection.
func (errCommandHandler) Count(ctx context.Context, db, col string) (int, error) {
	return -1, errors.New("command not supported")
}

// DropTable supports dropping tables.
func (errCommandHandler) DropTable(ctx context.Context, db, tbl string) error {
	return errors.New("command not supported")
}

// DropDatabase drops databases.
func (errCommandHandler) DropDatabase(ctx context.Context, db string) error {
	return errors.New("command not supported")
}

// CreateTable supports creating tables.
func (errCommandHandler) CreateTable(ctx context.Context, db string, table *schema.Table) error {
	return errors.New("command not supported")
}

// CreateDatabase creates Databases.
func (errCommandHandler) CreateDatabase(ctx context.Context, db string) error {
	return errors.New("command not supported")
}

// Insert inserts documents into the specified namespace.
func (errCommandHandler) Insert(ctx context.Context, db, table string, docs []interface{}) error {
	return errors.New("command not supported")
}

// Kill kills a Connection or Query (the KillScope). The targetConnID is the
// ID of the connection that is to be killed. The targetConnID may be the
// current connection id.
func (errCommandHandler) Kill(ctx context.Context, targetConnID uint32, ks KillScope) error {
	return errors.New("command not supported")
}

// Resample forces a sample refresh. It must occur in the server
// as that is where the schemata are maintained.
func (errCommandHandler) Resample(context.Context) error {
	return errors.New("command not supported")
}

// RotateLogs rotates the log file.
func (errCommandHandler) RotateLogs() error {
	return errors.New("command not supported")
}

// Set sets the value of the specified variable to the provided value.
func (errCommandHandler) Set(variable.Name, variable.Scope, variable.Kind, values.SQLValue) error {
	return errors.New("command not supported")
}

// SetDatabase sets the current database.
func (errCommandHandler) SetDatabase(db string) error {
	return errors.New("command not supported")
}

// SetScopeAuthorized returns an error if the user is not authorized to
// set variables in the provided scope.
func (errCommandHandler) SetScopeAuthorized(variable.Scope) error {
	return errors.New("command not supported")
}

// UnsetDatabase unsets the current database.
func (errCommandHandler) UnsetDatabase() error {
	return errors.New("command not supported")
}
