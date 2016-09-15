package evaluator

import (
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/variable"
	"gopkg.in/mgo.v2"
	"gopkg.in/tomb.v2"
)

// ConnectionCtx holds connection context information.
type ConnectionCtx interface {
	ConnectionId() uint32
	DB() string
	Kill(uint32, KillScope) error
	LastInsertId() int64
	RowCount() int64
	Session() *mgo.Session
	Logger(string) *log.Logger
	Tomb() *tomb.Tomb
	User() string
	Variables() *variable.Container
}

// ExecutionCtx holds execution context information
// used by each PlanStage Iter implementation.
type ExecutionCtx struct {
	ConnectionCtx

	// SrcRows is a row cache used when correlated subqueries
	// are in the tree.
	SrcRows []*Row

	// CacheRows is a row cache used to minimize the number of pushdowns
	// resulting from non-correlated subqueries.
	CacheRows map[string]interface{}
}

// NewExecutionCtx creates a new execution context.
func NewExecutionCtx(connCtx ConnectionCtx) *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: connCtx,
		CacheRows:     make(map[string]interface{}),
	}
}

// EvalCtx holds the current row to use when evaluating a SQLExpr.
type EvalCtx struct {
	*ExecutionCtx
	Rows []*Row

	collation *collation.Collation
}

// NewEvalCtx creates a new evaluation context.
func NewEvalCtx(execCtx *ExecutionCtx, collation *collation.Collation, rows ...*Row) *EvalCtx {
	return &EvalCtx{
		ExecutionCtx: execCtx,
		Rows:         rows,
		collation:    collation,
	}
}

// WithRows copies the EvalCtx but uses new rows.
func (ctx *EvalCtx) WithRows(rows ...*Row) *EvalCtx {
	return NewEvalCtx(ctx.ExecutionCtx, ctx.collation, rows...)
}

// CreateChildExecutionCtx creates a child ExecutionCtx.
func (ctx *EvalCtx) CreateChildExecutionCtx() *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: ctx.ExecutionCtx.ConnectionCtx,
		SrcRows:       append(ctx.Rows, ctx.ExecutionCtx.SrcRows...),
	}
}

// RequiresEvalCtx is an interface a struct can implement to allow it to
// be queried for whether or not it requires an EvalCtx, or can be handled
// in memory.
type RequiresEvalCtx interface {
	RequiresEvalCtx() bool
}
