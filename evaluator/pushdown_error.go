package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PushdownError represents an error returned by PushdownPlan. It can be fatal,
// which indicates that pushdown failed because some unexpected error was
// encountered during translation and a valid query plan could not be generated,
// or non-fatal, which indicates that one or more PlanStages could not be pushed
// down, but a valid, not-fully-pushed-down plan was generated.
type PushdownError interface {
	error
	IPushdownError()
	Failures() map[PlanStage][]PushdownFailure
}

// pushdownError represents an error encountered while trying to push down a query plan.
type pushdownError struct {
	// If fatal is true, pushdown failed because some unexpected error was
	// encountered during translation. Fatal errors indicate that a valid plan
	// could not be generated. Non-fatal errors indicate that one or more
	// PlanStages could not be pushed down, but a valid, not-fully-pushed-down
	// plan was generated.
	fatal bool
	// If fatal is true, fatalErr contains the error that caused pushdown to
	// fail. Otherwise, fatalErr is false.
	fatalErr error
	// If fatal is false, failures contains a pushdownFailure that explains
	// why each remaining stage in the query plan could not be pushed down.
	failures map[PlanStage][]PushdownFailure
}

func fatalPushdownError(err error) PushdownError {
	return &pushdownError{
		fatal:    true,
		fatalErr: err,
		failures: nil,
	}
}

func nonFatalPushdownError(failures map[PlanStage][]PushdownFailure) PushdownError {
	return &pushdownError{
		fatal:    false,
		fatalErr: nil,
		failures: failures,
	}
}

func (*pushdownError) IPushdownError() {}

func (e *pushdownError) Failures() map[PlanStage][]PushdownFailure {
	return e.failures
}

func (e *pushdownError) Error() string {
	if e.fatal {
		return e.fatalErr.Error()
	}
	return "failed to fully push down query plan"
}

// IsFatalPushdownError returns true if the provided error is a fatal pushdownError.
func IsFatalPushdownError(err error) bool {
	pde, ok := err.(*pushdownError)
	if !ok {
		return false
	}
	return pde.fatal
}

// IsNonFatalPushdownError returns true if the provided error is a non-fatal pushdownError.
func IsNonFatalPushdownError(err error) bool {
	pde, ok := err.(*pushdownError)
	if !ok {
		return false
	}
	return !pde.fatal
}

// PushdownFailure represents is an indication that (and explanation why) a
// SQLExpr or PlanStage could not be translated to aggregation.
type PushdownFailure interface {
	error
	IPushdownFailure()
}

// pushdownFailure contains information that explains why a given PlanStage
// could not be pushed down.
type pushdownFailure struct {
	// name is the name of the thing (usually a SQLExpr or a PlanStage) that
	// could not be pushed down.
	name string
	// reason is a short summary of why pushdown was impossible. This string
	// should not contain any potentially sensitive information like field names
	// or literal values.
	reason string
	// metadata contains extra metadata that will be printed any time this
	// failure is logged, but will not be included in any anonymous metrics
	// collection. This means it is safe to include potentially sensitive
	// information in these fields.
	metadata map[string]string
}

func newPushdownFailure(name, msg string, meta ...string) PushdownFailure {
	if len(meta)%2 != 0 {
		panic("must provide an even number of meta strings (key-value pairs)")
	}

	metaMap := make(map[string]string)
	i := 0
	for i < len(meta) {
		k := meta[i]
		v := meta[i+1]
		metaMap[k] = v
		i += 2
	}

	return &pushdownFailure{
		name:     name,
		reason:   msg,
		metadata: metaMap,
	}
}

func newUntranslatableExprFailure(e SQLExpr) PushdownFailure {
	return &pushdownFailure{
		name:   e.ExprName(),
		reason: "expression is not translatable to the aggregation language",
		metadata: map[string]string{
			"expr": e.String(),
		},
	}
}

func (p *pushdownFailure) IPushdownFailure() {}

func (p *pushdownFailure) Error() string {
	msg := fmt.Sprintf("failed to push down %s: %s.", p.name, p.reason)
	for k, v := range p.metadata {
		msg += fmt.Sprintf(" %s=%s", k, v)
	}
	return msg
}

func (p *pushdownFailure) MarshalJSON() ([]byte, error) {
	val := struct {
		Name   string `json:"name"`
		Reason string `json:"reason"`
	}{
		p.name,
		p.reason,
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(val)

	return buffer.Bytes(), err
}
