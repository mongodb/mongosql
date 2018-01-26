package util

import (
	"sync"
)

var (
	// a unique counter, anywhere a unique id is needed.
	uniqueCount      uint64
	uniqueCountMutex = &sync.Mutex{}
)

func GetUniqueID() uint64 {
	uniqueCountMutex.Lock()
	defer uniqueCountMutex.Unlock()
	i := uniqueCount
	// unint64 wraps around to 0 on overflow, which should be more than sufficient.
	// This will only fail if we have more than 2^64 expressions that need a uniqueId
	// in one query. Given memory constraints, such is infeasible, anyway.
	uniqueCount++
	return i
}
