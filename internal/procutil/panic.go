package procutil

// PanicSafeGo accepts a closure that may panic and another that is used
// to handle cases where such panics occur. It executes closureMayPanic in
// a goroutine and runs the panicRecovery closure if the former panics.
func PanicSafeGo(closureMayPanic func(), panicRecovery func(err interface{})) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				panicRecovery(err)
			}
		}()
		closureMayPanic()
	}()
}
