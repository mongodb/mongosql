package evaluator

import "gopkg.in/mgo.v2"

// ConnectionCtx holds connection context information.
type ConnectionCtx interface {
	LastInsertId() int64
	RowCount() int64
	ConnectionId() uint32
	DB() string
	Session() *mgo.Session
}

// ExecutionCtx holds execution context information
// used by each PlanStage Iter implementation.
type ExecutionCtx struct {
	ConnectionCtx

	// SrcRows is a row cache used when correlated subqueries
	// are in the tree.
	SrcRows []*Row

	AuthProvider AuthProvider
}

// NewExecutionCtx creates a new execution context.
func NewExecutionCtx(connCtx ConnectionCtx) *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: connCtx,
		AuthProvider:  NewAuthProvider(connCtx.Session()),
	}
}

// EvalCtx holds the current row to use when evaluating a SQLExpr.
type EvalCtx struct {
	*ExecutionCtx
	Rows []*Row
}

// NewEvalCtx creates a new evaluation context.
func NewEvalCtx(execCtx *ExecutionCtx, rows ...*Row) *EvalCtx {
	return &EvalCtx{
		ExecutionCtx: execCtx,
		Rows:         rows,
	}
}

// CreateChildExecutionCtx creates a child ExecutionCtx.
func (ctx *EvalCtx) CreateChildExecutionCtx() *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: ctx.ExecutionCtx.ConnectionCtx,
		AuthProvider:  ctx.ExecutionCtx.AuthProvider,
		SrcRows:       append(ctx.Rows, ctx.ExecutionCtx.SrcRows...),
	}
}
