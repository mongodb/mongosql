package evaluator

// SetExecutor handles setting variables.
type SetExecutor struct {
	assignments []*SQLAssignmentExpr
}

// NewSetExecutor creates a new SetExecutor.
func NewSetExecutor(assignments []*SQLAssignmentExpr) *SetExecutor {
	return &SetExecutor{assignments}
}

func (s *SetExecutor) Execute(ctx *ExecutionCtx) error {
	evalCtx := NewEvalCtx(ctx)
	for _, a := range s.assignments {
		_, err := a.Evaluate(evalCtx)
		if err != nil {
			return err
		}
	}

	return nil
}
