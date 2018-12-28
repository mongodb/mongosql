package procutil

import (
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
)

var profile *os.File
var profileLock sync.Mutex

// StartCPUProfile starts a CPU profile that will be written to the provided
// filename. If the profile cannot be started, the file cannot be created, or
// a profile is already in progress, an error will be returned.
func StartCPUProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %v", err)
	}

	profileLock.Lock()
	defer profileLock.Unlock()

	if profile != nil {
		return fmt.Errorf("CPU profile already running")
	}
	profile = f

	err = pprof.StartCPUProfile(f)
	if err != nil {
		profile = nil
		return fmt.Errorf("could not start CPU profile: %s", err)
	}

	return nil
}

// StopCPUProfile stops any cpu profile that is currently running.
func StopCPUProfile() {
	profileLock.Lock()
	defer profileLock.Unlock()

	pprof.StopCPUProfile()
	profile = nil
}
