package evaluator

import (
	"context"

	"math/rand"

	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

// ServerCtx holds server context information
type ServerCtx interface {
	// Alter executes schema alterations. It must occur in the server
	// as that is where the schemata are maintained.
	Alter(context.Context, []*schema.Alteration) (*schema.Schema, error)
	// IsProcessOwner returns true if the passed user owns the given processID.
	// Only the server can have this information.
	IsProcessOwner(user string, processID uint32) (bool, error)
	// IsAdminUser tells if the user name and source is the admin user.
	// The source is necessary because MongoDB supports users with the same
	// name in separate databases.
	IsAdminUser(user string, source string) bool
	// Kill kills a Connection or Query (the killscope). The requestingConnID is
	// the id of the connection requesting the kill, the targetConnID is the ID
	// of the connection that is to be killed. They may be the same ID.
	Kill(requestingConnID uint32, targetConnID uint32, killScope KillScope) error
	// Resample forces a sample refresh. It must occur in the server
	// as that is where the schemata are maintained.
	Resample(context.Context) (*schema.Schema, error)
	// RotateLogs rotates the log file.
	RotateLogs() error
}

// TranslationCtx holds the context information used to perform translation to
// MongoDB query language.
type TranslationCtx interface {
	// VersionAtLeast returns if the current TranslationCtx has a version
	// at least that of the one specified as a sequence of uint8. This is
	// used primarily to know what pushdown targets are available.
	VersionAtLeast(...uint8) bool
	// Logger retrieves a reference to a logger. If a string is passed
	// it retrieves the logger for the component specified by that string.
	// If no string is passed, it retrieves the root logger.
	Logger(...string) log.Logger
	// Variables returns the variable container for the current connection.
	Variables() *variable.Container
}

// ConnectionCtx holds connection context information.
type ConnectionCtx interface {
	TranslationCtx
	// Catalog retrieves the catalog of visible SQL namespaces,
	// which maps a given sql namespace to the proper mongodb
	// namespace.
	Catalog() *catalog.Catalog
	// ConnectionID returns the current connection ID.
	ConnectionID() uint32
	// Context returns the underlying context struct.
	Context() context.Context
	// DB returns the current database name.
	DB() string
	// LastInsertId returns the last insert id.
	LastInsertId() int64
	// RemoteHost returns the name of the host from where this connection
	// originates.
	RemoteHost() string
	// RowCount returns the number of rows affected by the last statement
	// from this connection.
	RowCount() int64
	// Server returns the server underlying this connection.
	Server() ServerCtx
	// Session returns a new mongodb.Session connected to MongoDB.
	Session() *mongodb.Session
	// AuthenticationDatabase returns authentication database for the
	// current connection's user, i.e., the database in which the
	// current user was created.
	AuthenticationDatabase() string
	// UpdateCatalog updates the catalog to utilize the new schema.
	UpdateCatalog(*schema.Schema) error
	// User returns the name of the user who owns this connection.
	User() string
	// MemoryMonitor returns the memory monitor for this connection.
	MemoryMonitor() *memory.Monitor
}

// ExecutionCtx holds execution context information
// used by each PlanStage Iter implementation.
type ExecutionCtx struct {
	ConnectionCtx
	// A map from uint64 to go ptr to Rand structs.
	// These are needed because each RAND() in a SQL expression has its
	// own separate sequence. We count each rand with a global uint64.
	RandomExprs map[uint64]*rand.Rand
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

func (ctx *ExecutionCtx) valueKind() SQLValueKind {
	return GetSQLValueKind(ctx.Variables())
}

// EvalCtx holds the current row to use when evaluating a SQLExpr.
type EvalCtx struct {
	*ExecutionCtx
	Rows      []*Row
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
