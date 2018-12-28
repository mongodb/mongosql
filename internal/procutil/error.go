package procutil

import "context"

// CheckDeferredFunc manages errors returned from deferred functions.
func CheckDeferredFunc(close func() error, err *error) {
	cerr := close()
	if *err == nil {
		*err = cerr
	}
}

type closeFunc func(ctx context.Context) error

// CheckDeferredFuncWithContext manages errors returned from deferred functions
// that takes context as a parameter.
func CheckDeferredFuncWithContext(context context.Context, cf closeFunc, err *error) {
	cerr := cf(context)
	if *err == nil {
		*err = cerr
	}
}

// CheckForContextCancellationAndError checks to see if the context is done (timeout or cancelled).
// If the context has completed it returns the context's error, otherwise, returns possibleErr.
func CheckForContextCancellationAndError(ctx context.Context, possibleErr error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return possibleErr
	}
}
