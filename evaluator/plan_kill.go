package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/util"
)

type KillScope byte

const (
	KillConnection = iota
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

type KillExecutor struct {
	ID    SQLExpr
	Scope KillScope
	ctx   *ExecutionCtx
}

// NewKillCommand creates a new KillCommand.
func NewKillCommand(id SQLExpr, scope KillScope) *KillCommand {
	return &KillCommand{id, scope}
}

func (k *KillCommand) Execute(ctx *ExecutionCtx) Executor {
	return &KillExecutor{k.ID, k.Scope, ctx}
}

func (k *KillExecutor) Run() error {

	executorChan := make(chan error)

	var err error

	util.PanicSafeGo(func() {
		evalCtx := NewEvalCtx(k.ctx, collation.Default)

		eval, pErr := k.ID.Evaluate(evalCtx)
		if pErr != nil {
			executorChan <- pErr
		}

		id, pErr := util.ToInt(eval)
		if pErr != nil {
			executorChan <- mysqlerrors.Defaultf(
				mysqlerrors.ER_NO_SUCH_THREAD, eval)
		}

		executorChan <- evalCtx.Kill(uint32(id), k.Scope)
	}, func(err interface{}) {
		executorChan <- fmt.Errorf("%v", err)
	})

	select {
	case <-k.ctx.ConnectionCtx.Context().Done():
		err = k.ctx.ConnectionCtx.Context().Err()
	case err = <-executorChan:
	}

	return err
}
