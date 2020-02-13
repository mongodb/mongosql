// +build !windows

package procutil

import (
	"fmt"
	"syscall"
)

// GetUsageStats gets resource usage information for the current
// process and returns it as a formatted string.
func GetUsageStats() (string, error) {
	usage := &syscall.Rusage{}
	err := syscall.Getrusage(syscall.RUSAGE_SELF, usage)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"utimens=%v\tstimens=%v\tmaxrss=%v\tminflt=%v\tmajflt=%v",
		usage.Utime.Nano(),
		usage.Stime.Nano(),
		usage.Maxrss,
		usage.Minflt,
		usage.Majflt,
	), nil
}
