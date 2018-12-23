package memory_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/stretchr/testify/require"
)

func TestAquire_allocates_the_correct_amount_of_memory(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	require.Equal(t, m.Allocated(), uint64(0))

	err := m.Acquire(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Acquire(6)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(10))
}

func TestAquire_rolls_the_required_amount_up_to_the_parent(t *testing.T) {
	p := memory.NewMonitor("parent", 20)

	err := p.Acquire(4)
	require.NoError(t, err)

	m, _ := p.CreateChild("child", 10)

	require.Equal(t, p.Allocated(), uint64(4))
	require.Equal(t, m.Allocated(), uint64(0))

	err = m.Acquire(4)

	require.NoError(t, err)
	require.Equal(t, p.Allocated(), uint64(8))
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Acquire(6)

	require.NoError(t, err)
	require.Equal(t, p.Allocated(), uint64(14))
	require.Equal(t, m.Allocated(), uint64(10))
}

func TestAcquire_when_allocating_without_a_limit(t *testing.T) {
	m := memory.NewMonitor("test", 0)

	err := m.Acquire(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Acquire(8)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(12))
}

func TestAcquire_errors_when_allocating_more_memory_than_allowed(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	err := m.Acquire(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Acquire(8)

	require.Error(t, err)
	require.Equal(t, m.Allocated(), uint64(4))
}

func TestAcquire_errors_when_parent_fails_allocating_memory(t *testing.T) {
	p := memory.NewMonitor("parent", 20)
	m, _ := p.CreateChild("child", 10)

	err := p.Acquire(18)
	require.NoError(t, err)
	require.Equal(t, p.Allocated(), uint64(18))
	require.Equal(t, m.Allocated(), uint64(0))

	err = m.Acquire(8)

	require.Error(t, err)
	require.Equal(t, p.Allocated(), uint64(18))
	require.Equal(t, m.Allocated(), uint64(0))
}

func TestAcquire_errors_on_overflow(t *testing.T) {
	m := memory.NewMonitor("test", 0)

	err := m.Acquire(math.MaxUint64 - 10)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(math.MaxUint64-10))

	err = m.Acquire(13)
	fmt.Printf("%d\n", m.Allocated())

	require.Error(t, err)
	require.Equal(t, m.Allocated(), uint64(math.MaxUint64-10))
}

func TestAcquireReleaseGlobal(t *testing.T) {
	req := require.New(t)
	m := memory.NewMonitor("test", 0)

	err := m.AcquireGlobal(uint64(100))
	req.NoError(err)
	req.Equal(uint64(100), m.Allocated())
	req.Equal(uint64(100), m.GlobalAllocated())

	err = m.Acquire(10)
	req.NoError(err)
	req.Equal(uint64(110), m.Allocated())
	req.Equal(uint64(100), m.GlobalAllocated())

	err = m.ReleaseGlobal()
	req.NoError(err)
	req.Equal(uint64(10), m.Allocated())
	req.Equal(uint64(0), m.GlobalAllocated())
}

func TestClear(t *testing.T) {
	m := memory.NewMonitor("test", 100)

	require.NoError(t, m.Acquire(10))

	allocated, err := m.Clear()
	require.NoError(t, err)
	require.Equal(t, uint64(10), allocated)

	maxAllocated := m.PeakAllocated()
	require.Equal(t, uint64(0), maxAllocated)
}

func TestExclude_excludes_the_correct_amount_of_memory(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	err := m.Acquire(10)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(10))

	err = m.Exclude(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(6))

	err = m.Exclude(6)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(0))
}

func TestExclude_errors_when_too_much_memory_is_excluded(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	err := m.Exclude(1)

	require.Error(t, err)
	require.Equal(t, m.Allocated(), uint64(0))
}

func TestInclude_includes_the_correct_amount_of_memory(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	require.Equal(t, m.Allocated(), uint64(0))

	err := m.Include(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Include(6)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(10))
}

func TestInclude_when_including_without_a_limit(t *testing.T) {
	m := memory.NewMonitor("test", 0)

	err := m.Include(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Include(8)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(12))
}

func TestInclude_errors_when_including_more_memory_than_allowed(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	err := m.Acquire(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Acquire(8)

	require.Error(t, err)
	require.Equal(t, m.Allocated(), uint64(4))
}

func TestInclude_errors_on_overflow(t *testing.T) {
	m := memory.NewMonitor("test", 0)

	err := m.Acquire(math.MaxUint64 - 10)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(math.MaxUint64-10))

	err = m.Acquire(13)
	fmt.Printf("%d\n", m.Allocated())

	require.Error(t, err)
	require.Equal(t, m.Allocated(), uint64(math.MaxUint64-10))
}

func TestMaxAllocated(t *testing.T) {
	m := memory.NewMonitor("test", 100)

	require.NoError(t, m.Acquire(10))
	require.NoError(t, m.Acquire(20))
	require.NoError(t, m.Release(10))
	require.NoError(t, m.Acquire(30))
	require.NoError(t, m.Release(40))
	require.NoError(t, m.Acquire(10))

	maxAllocated := m.PeakAllocated()
	require.Equal(t, uint64(50), maxAllocated)
}

func TestRelease_returns_the_correct_amount_of_memory(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	err := m.Acquire(10)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(10))

	err = m.Release(4)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(6))

	err = m.Release(6)

	require.NoError(t, err)
	require.Equal(t, m.Allocated(), uint64(0))
}

func TestRelease_returns_the_correct_amount_of_memory_to_the_parent(t *testing.T) {
	p := memory.NewMonitor("parent", 20)
	m, err := p.CreateChild("child", 10)
	require.NoError(t, err)

	err = p.Acquire(10)

	require.NoError(t, err)
	require.Equal(t, p.Allocated(), uint64(10))
	require.Equal(t, m.Allocated(), uint64(0))

	err = m.Acquire(8)
	require.NoError(t, err)
	require.Equal(t, p.Allocated(), uint64(18))
	require.Equal(t, m.Allocated(), uint64(8))

	err = m.Release(4)

	require.NoError(t, err)
	require.Equal(t, p.Allocated(), uint64(14))
	require.Equal(t, m.Allocated(), uint64(4))

	err = m.Release(4)

	require.NoError(t, err)
	require.Equal(t, p.Allocated(), uint64(10))
	require.Equal(t, m.Allocated(), uint64(0))
}

func TestRelease_errors_when_too_much_memory_is_returned(t *testing.T) {
	m := memory.NewMonitor("test", 10)

	err := m.Release(1)

	require.Error(t, err)
	require.Equal(t, m.Allocated(), uint64(0))
}

func TestRelease_errors_when_too_much_memory_is_released_from_parent(t *testing.T) {
	p := memory.NewMonitor("parent", 20)
	m, _ := p.CreateChild("child", 10)

	err := m.Release(1)

	require.Error(t, err)
	require.Equal(t, p.Allocated(), uint64(0))
	require.Equal(t, m.Allocated(), uint64(0))
}

func TestCreateChild_does_not_error_when_limit_is_infinite(t *testing.T) {
	p := memory.NewMonitor("parent", 0)
	_, err := p.CreateChild("child", 20)
	require.NoError(t, err)
}

func TestCreateChild_errors_when_child_is_created_over_the_parent_limit(t *testing.T) {
	p := memory.NewMonitor("parent", 10)

	_, err := p.CreateChild("child", 20)
	require.Error(t, err)
}
