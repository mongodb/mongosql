package option_test

import (
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/internal/option"
	"github.com/stretchr/testify/require"
)

func TestStringOption(t *testing.T) {
	req := require.New(t)

	some := option.SomeString("abc")
	req.True(some.IsSome())
	req.False(some.IsNone())
	req.Equal("abc", some.Unwrap())
	req.Equal("abc", some.Else("def"))

	someMapped := some.Map(strings.ToUpper)
	req.True(someMapped.IsSome())
	req.Equal("ABC", someMapped.Unwrap())

	some.Set("ghi")
	req.True(some.IsSome())
	req.False(some.IsNone())
	req.Equal("ghi", some.Unwrap())
	req.Equal("ghi", some.Else("def"))

	none := option.NoneString()
	req.False(none.IsSome())
	req.True(none.IsNone())
	req.Panics(func() { none.Unwrap() })
	req.Equal("def", none.Else("def"))

	noneMapped := none.Map(strings.ToUpper)
	req.True(noneMapped.IsNone())

	none.Set("mno")
	req.True(none.IsSome())
	req.False(none.IsNone())
	req.Equal("mno", none.Unwrap())
	req.Equal("mno", none.Else("def"))

	req.Equal(`Some("a")`, option.SomeString("a").String())
	req.Equal("None", option.NoneString().String())

	// test that == works for this implementation, since our code relies on that.
	// nolint: staticcheck
	req.True(option.SomeString("hello") == option.SomeString("hello"))
	// nolint: staticcheck
	req.True(option.SomeString("") == option.SomeString(""))
	req.True(option.SomeString("hello") != option.SomeString(""))
	req.True(option.SomeString("") != option.NoneString())
	// nolint: staticcheck
	req.True(option.NoneString() == option.NoneString())
}
