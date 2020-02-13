// +build windows

package procutil

// GetUsageStats gets resource usage information for the current
// process and returns it as a formatted string.
func GetUsageStats() (string, error) {
	panic("this function should never be called on windows")
}
