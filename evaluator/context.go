package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

// ServerCtx holds server context information
type ServerCtx interface {
	Alter(context.Context, []*schema.Alteration) (*schema.Schema, error)
	Resample(context.Context) (*schema.Schema, error)
	StartupInfo() []string
}

// ConnectionCtx holds connection context information.
type ConnectionCtx interface {
	ConnectionID() uint32
	DB() string
	Kill(uint32, KillScope) error
	LastInsertId() int64
	RowCount() int64
	Session() *mongodb.Session
	Logger(string) *log.Logger
	User() string
	Variables() *variable.Container
	Context() context.Context
	Server() ServerCtx

	Catalog() *catalog.Catalog
	UpdateCatalog(*schema.Schema) error
}

// ExecutionCtx holds execution context information
// used by each PlanStage Iter implementation.
type ExecutionCtx struct {
	ConnectionCtx

	// SrcRows is a row cache used when correlated subqueries
	// are in the tree.
	SrcRows []*Row
}

// NewExecutionCtx creates a new execution context.
func NewExecutionCtx(connCtx ConnectionCtx) *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: connCtx,
	}
}

// EvalCtx holds the current row to use when evaluating a SQLExpr.
type EvalCtx struct {
	*ExecutionCtx
	Rows []*Row

	Collation *collation.Collation
}

// NewEvalCtx creates a new evaluation context.
func NewEvalCtx(execCtx *ExecutionCtx, collation *collation.Collation, rows ...*Row) *EvalCtx {
	return &EvalCtx{
		ExecutionCtx: execCtx,
		Rows:         rows,
		Collation:    collation,
	}
}

// WithRows copies the EvalCtx but uses new rows.
func (ctx *EvalCtx) WithRows(rows ...*Row) *EvalCtx {
	return NewEvalCtx(ctx.ExecutionCtx, ctx.Collation, rows...)
}

// CreateChildExecutionCtx creates a child ExecutionCtx.
func (ctx *EvalCtx) CreateChildExecutionCtx() *ExecutionCtx {
	srcRows := make([]*Row, len(ctx.Rows), len(ctx.Rows)+len(ctx.ExecutionCtx.SrcRows))
	copy(srcRows, ctx.Rows)
	srcRows = append(srcRows, ctx.ExecutionCtx.SrcRows...)
	return &ExecutionCtx{
		ConnectionCtx: ctx.ExecutionCtx.ConnectionCtx,
		SrcRows:       srcRows,
	}
}

// RequiresEvalCtx is an interface a struct can implement to allow it to
// be queried for whether or not it requires an EvalCtx, or can be handled
// in memory.
type RequiresEvalCtx interface {
	RequiresEvalCtx() bool
}
