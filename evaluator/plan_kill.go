package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
)

// KillScope is an enum that represents the scope of a kill command.
type KillScope byte

// These are the possible values for KillScope.
const (
	KillConnection KillScope = iota
	KillQuery
)

func (scope KillScope) String() string {
	if scope == KillConnection {
		return "connection"
	}
	return "query"
}

// KillCommand handles killing connection or queries.
type KillCommand struct {
	ID    SQLExpr
	Scope KillScope
}

// NewKillCommand creates a new KillCommand.
func NewKillCommand(id SQLExpr, scope KillScope) *KillCommand {
	return &KillCommand{id, scope}
}

// Children returns a slice of all the Node children of the Node.
func (k KillCommand) Children() []Node {
	return []Node{k.ID}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (k *KillCommand) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		k.ID = panicIfNotSQLExpr("KillCommand", n)
	default:
		panicWithInvalidIndex("KillCommand", i, 0)
	}
}

// Execute runs this command.
func (k *KillCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	eval, err := k.ID.Evaluate(ctx, cfg, st)
	if err != nil {
		return err
	}

	id, err := mathutil.ToInt(eval.Value())
	if err != nil {
		return mysqlerrors.Defaultf(mysqlerrors.ErNoSuchThread, eval)
	}

	return cfg.commandHandler.Kill(ctx, uint32(id), k.Scope)
}
