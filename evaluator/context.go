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
// used by each Iterator implemenation.
type ExecutionCtx struct {
	Depth int

	// GroupRows holds a set of rows used by each GROUP BY combination
	GroupRows []Row

	// SrcRows caches the data gotten from a table scan or join node
	SrcRows []*Row

	ConnectionCtx

	AuthProvider AuthProvider
}

// NewExecutionCtx creates a new execution context.
func NewExecutionCtx(connCtx ConnectionCtx) *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: connCtx,
		AuthProvider:  NewAuthProvider(connCtx.Session()),
	}
}

// EvalCtx holds a slice of rows used to evaluate a SQLValue.
type EvalCtx struct {
	Rows    []Row
	ExecCtx *ExecutionCtx
}
