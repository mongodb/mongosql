package procutil

import "time"

// RepeatWithDelay with delay reruns the function every (duration) until the channel is done.
func RepeatWithDelay(cancel <-chan struct{}, every time.Duration, immediate bool, f func()) {
	RetryWithDelay(cancel, every, immediate, func() bool {
		f()
		return false
	})
}

// RetryWithDelay retries the function every (delay duration) until it indicates it is done.
func RetryWithDelay(cancel <-chan struct{}, delay time.Duration, immediate bool, f func() bool) {
	if immediate && f() {
		return
	}

	for {
		select {
		case <-cancel:
			return
		case <-time.After(delay):
			if f() {
				return
			}
		}
	}
}
