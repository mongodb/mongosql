package evaluator

type pushdownError struct {
	message string
	errors  map[PlanStage]string
}

func (e *pushdownError) Error() string {
	return e.message
}

func newPushdownError() *pushdownError {
	return &pushdownError{
		message: "Push down error: could not be fully pushed down",
		errors:  make(map[PlanStage]string),
	}
}

// IsPushdownError returns true if the provided error is a pushdownError.
func IsPushdownError(err error) bool {
	_, ok := err.(*pushdownError)
	return ok
}
