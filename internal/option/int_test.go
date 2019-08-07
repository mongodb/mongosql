package option_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/option"
	"github.com/stretchr/testify/require"
)

func TestIntOption(t *testing.T) {
	inc := func(i int) int {
		return i + 1
	}
	req := require.New(t)

	some := option.SomeInt(42)
	req.True(some.IsSome())
	req.False(some.IsNone())
	req.Equal(42, some.Unwrap())
	req.Equal(42, some.Else(43))

	someMapped := some.Map(inc)
	req.True(someMapped.IsSome())
	req.Equal(43, someMapped.Unwrap())

	some.Set(45)
	req.True(some.IsSome())
	req.False(some.IsNone())
	req.Equal(45, some.Unwrap())
	req.Equal(45, some.Else(43))

	none := option.NoneInt()
	req.False(none.IsSome())
	req.True(none.IsNone())
	req.Panics(func() { none.Unwrap() })
	req.Equal(43, none.Else(43))

	noneMapped := none.Map(inc)
	req.True(noneMapped.IsNone())

	none.Set(53)
	req.True(none.IsSome())
	req.False(none.IsNone())
	req.Equal(53, none.Unwrap())
	req.Equal(53, none.Else(43))

	constantlyHello := func(int) string {
		return "hello"
	}
	req.Equal("hello", option.SomeInt(42).MapString(constantlyHello).Unwrap())
	req.Equal("", option.NoneInt().MapString(constantlyHello).Else(""))

	req.Equal(`Some(42)`, option.SomeInt(42).String())
	req.Equal("None", option.NoneInt().String())

	// test that == works for this implementation, since our code relies on that.
	// nolint: staticcheck
	req.True(option.SomeInt(42) == option.SomeInt(42))
	// nolint: staticcheck
	req.True(option.SomeInt(0) == option.SomeInt(0))
	req.True(option.SomeInt(42) != option.SomeInt(0))
	req.True(option.SomeInt(0) != option.NoneInt())
	// nolint: staticcheck
	req.True(option.NoneInt() == option.NoneInt())
}
