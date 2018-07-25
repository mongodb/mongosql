package evaluator

type pushDownError struct {
	message string
	errors  map[PlanStage]string
}

func (e *pushDownError) Error() string {
	return e.message
}

func newPushDownError() *pushDownError {
	return &pushDownError{
		message: "Push down error: could not be fully pushed down",
		errors:  make(map[PlanStage]string),
	}
}
