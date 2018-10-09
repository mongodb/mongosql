package memory

import (
	"fmt"
	"math"
	"sync/atomic"
)

// NewMonitor creates a new monitor with the specified limit.
func NewMonitor(name string, limit uint64) *Monitor {
	return &Monitor{name: name, limit: limit}
}

// Monitor tracks memory usage. It exists as a hierarchy, potentially containing a parent Monitor.
// When memory is Acquire(d) or Release(d) from a child Monitor, it is also Acquire(d) or Release(d)
// respectively from the parent Monitor allowing us to roll up usage at a lower level to a higher
// level.
//
// In practice, this is a 2 or 3 level hierarchy, where the Server contains a Monitor,
// the Connection is a child of the Server, and each stage, if they need stage-level tracking, will
// contain a child of the Connection.
type Monitor struct {
	// name gives a context for the type of memory the Monitor is tracking.
	name string
	// parent references the Monitor which created this one. It exists to track memory hierarchicly.
	parent *Monitor
	// limit is the maximum amount of allocated memory allowed by this Monitor.
	limit uint64
	// allocated indicates the current amount of allocated memory. It must be accessed atomicly.
	allocated uint64
	// globalAllocated indicates the current amount of memory allocated with AcquireGlobal.
	// The memory included in this count is also tracked in the "allocated" field.
	globalAllocated uint64
	// peakAllocated is the peak amount of memory allocated by this memory during its lifetime.
	// It must be accessed atomically.
	peakAllocated uint64
}

// Allocated is the total memory in bytes allocated.
func (m *Monitor) Allocated() uint64 {
	return atomic.LoadUint64(&m.allocated)
}

// Acquire attempts to "checkout" the specified amount of
// memory. An error will be returned if the amount requested
// cannot be provided.
func (m *Monitor) Acquire(amount uint64) error {
	updated := atomic.AddUint64(&m.allocated, amount)
	if updated < amount || (m.limit > 0 && updated > m.limit) {
		atomic.AddUint64(&m.allocated, -amount)
		return m.wrapErrorf("an acquisition of %d bytes "+
			"pushes total allocation over the limit of %d", amount, m.limit)
	}

	if m.parent != nil {
		err := m.parent.Acquire(amount)
		if err != nil {
			atomic.AddUint64(&m.allocated, -amount)
			return m.wrapError(err)
		}
	}

	old := atomic.LoadUint64(&m.peakAllocated)
	if updated > old {
		// This is racy, but it's ok as this number
		// isn't required to be completely accurate.
		atomic.StoreUint64(&m.peakAllocated, updated)
	}

	return nil
}

// AcquireGlobal attempts to "checkout" the specified amount of
// memory. An error will be returned if the amount requested
// cannot be provided.
// AcquireGlobal should be used to track memory that will live for
// the lifetime of the memory monitor, so that it can easily be
// released at one time.
func (m *Monitor) AcquireGlobal(amount uint64) error {
	err := m.Acquire(amount)
	if err != nil {
		return err
	}
	atomic.AddUint64(&m.globalAllocated, amount)
	return nil
}

// Clear resets the monitor back to 0 levels.
func (m *Monitor) Clear() (uint64, error) {
	allocated := m.Allocated()
	if err := m.ReleaseGlobal(); err != nil {
		return allocated, m.wrapError(err)
	}
	allocated = m.Allocated()
	if err := m.Release(allocated); err != nil {
		return allocated, m.wrapError(err)
	}
	atomic.StoreUint64(&m.peakAllocated, 0)
	return allocated, nil
}

// Exclude decrements the allocated amount, but does not
// request release from the parent.
func (m *Monitor) Exclude(amount uint64) error {
	updated := atomic.AddUint64(&m.allocated, -amount)
	if math.MaxUint64-updated < amount {
		// underflow
		atomic.AddUint64(&m.allocated, amount)
		return m.wrapErrorf("memory excluded is more than has been allocated")
	}

	return nil
}

// GlobalAllocated is the total memory in bytes allocated.
func (m *Monitor) GlobalAllocated() uint64 {
	return atomic.LoadUint64(&m.globalAllocated)
}

// Include increments the allocated amount, but does not
// request allocation from the parent.
func (m *Monitor) Include(amount uint64) error {
	updated := atomic.AddUint64(&m.allocated, amount)
	if updated < amount || (m.limit > 0 && updated > m.limit) {
		atomic.AddUint64(&m.allocated, -amount)
		return m.wrapErrorf("an inclusion of %d bytes "+
			"pushes total allocation over the limit of %d", amount, m.limit)
	}

	old := atomic.LoadUint64(&m.peakAllocated)
	if updated > old {
		// This is racy, but it's ok as this number
		// isn't required to be completely accurate.
		atomic.StoreUint64(&m.peakAllocated, updated)
	}

	return nil
}

// Limit returns the maximum amount of memory that can be acquired from this Monitor.
func (m *Monitor) Limit() uint64 {
	return m.limit
}

// PeakAllocated is the maximum number of bytes allocated.
func (m *Monitor) PeakAllocated() uint64 {
	return atomic.LoadUint64(&m.peakAllocated)
}

// Release releases the specified amount of memory.
// It will decrease Allocated().
func (m *Monitor) Release(amount uint64) error {
	updated := atomic.AddUint64(&m.allocated, -amount)
	if math.MaxUint64-updated < amount {
		// underflow
		atomic.AddUint64(&m.allocated, amount)
		return m.wrapErrorf("memory released is more than has been acquired")
	}

	if m.parent != nil {
		if err := m.parent.Release(amount); err != nil {
			atomic.AddUint64(&m.allocated, amount)
			return m.wrapError(err)
		}
	}

	return nil
}

// ReleaseGlobal releases the amount of memory denoted by globalAllocated.
func (m *Monitor) ReleaseGlobal() error {
	toRelease := atomic.LoadUint64(&m.globalAllocated)
	err := m.Release(toRelease)
	if err != nil {
		return err
	}
	_ = atomic.SwapUint64(&m.globalAllocated, 0)
	return nil
}

// CreateChild creates a child with a limit on the
// total bytes it is allowed to allocate.
func (m *Monitor) CreateChild(name string, limit uint64) (*Monitor, error) {
	if m.limit > 0 && limit > m.limit {
		return nil, m.wrapErrorf("higher memory limit (%d) requested for"+
			" child(%s) than exists for parent (%d)", limit, name, m.limit)
	}
	return &Monitor{
		parent: m,
		name:   name,
		limit:  limit,
	}, nil
}

func (m *Monitor) wrapError(err error) error {
	return fmt.Errorf("%s - %v", m.name, err)
}

func (m *Monitor) wrapErrorf(format string, a ...interface{}) error {
	return m.wrapError(fmt.Errorf(format, a...))
}
