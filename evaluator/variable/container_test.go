package variable_test

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/stretchr/testify/require"
)

func TestGlobalVariable(t *testing.T) {

	t.Run("default_value", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewGlobalContainer(nil)

		v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(true, v.Value)
	})

	t.Run("invalid_name", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewGlobalContainer(nil)

		_, err := subject.Get(variable.Name("test"), variable.GlobalScope, variable.SystemKind)
		req.Error(err)

		err = subject.Set(variable.Name("test"), variable.GlobalScope, variable.SystemKind, false)
		req.Error(err)
	})

	t.Run("invalid_scope", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewGlobalContainer(nil)

		f := func() {
			_ = subject.Set(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind, false)
		}
		req.Panics(f)

		_, err := subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
		req.Error(err)
	})

	t.Run("invalid_kind", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewGlobalContainer(nil)

		f := func() {
			_ = subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.UserKind, false)
		}
		req.Panics(f)

		f = func() {
			_, _ = subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.UserKind)
		}
		req.Panics(f)
	})

	t.Run("invalid_type", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewGlobalContainer(nil)

		err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, "yeahaehasdh")
		req.Error(err)
	})

	t.Run("valid_name_and_scope", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewGlobalContainer(nil)

		err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, false)
		req.NoError(err)

		v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(false, v.Value)
	})
}

func TestSessionVariable(t *testing.T) {
	t.Run("invalid_system_variable", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))
		_, err := subject.Get(variable.Name("test"), variable.SessionScope, variable.SystemKind)
		req.Error(err)
	})

	t.Run("default_value", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		v, err := subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(true, v.Value)
	})

	t.Run("fallback_to_parent", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(true, v.Value)
	})

	t.Run("unset_user_variable", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		v, err := subject.Get(variable.Name("test"), variable.SessionScope, variable.UserKind)
		req.NoError(err)
		req.Nil(v.Value)
	})

	t.Run("invalid_name", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		err := subject.Set(variable.Name("test"), variable.SessionScope, variable.SystemKind, false)
		req.Error(err)
	})

	t.Run("invalid_kind", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		f := func() {
			_ = subject.Set(variable.Name("autocommit"), variable.GlobalScope,
				variable.UserKind, false)
		}
		req.Panics(f)
	})

	t.Run("invalid_type", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, "yeahaehasdh")
		req.Error(err)
	})

	t.Run("parent_scope", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, false)
		req.NoError(err)

		v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(false, v.Value)

		v, err = subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(true, v.Value)
	})

	t.Run("current_scope", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		err := subject.Set(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind, false)
		req.NoError(err)

		v, err := subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(false, v.Value)

		v, err = subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
		req.NoError(err)
		req.Equal(true, v.Value)
	})

	t.Run("set user variable", func(t *testing.T) {
		req := require.New(t)
		subject := variable.NewSessionContainer(variable.NewGlobalContainer(nil))

		err := subject.Set(variable.Name("test"), variable.SessionScope, variable.UserKind, "yeah")
		req.NoError(err)

		v, err := subject.Get(variable.Name("test"), variable.SessionScope, variable.UserKind)
		req.NoError(err)
		req.Equal("yeah", v.Value)
	})

}
