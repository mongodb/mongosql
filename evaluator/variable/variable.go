package variable

//go:generate go run testdata/generate.go variables.yml variables_generated.go

import (
	"strings"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
)

// Name is the name of a variable.
type Name string

// Kind represents the kind of a variable.
type Kind int

const (
	// SystemKind is a system variable.
	SystemKind Kind = iota
	// StatusKind is a status variable.
	StatusKind
	// UserKind is a user variable.
	UserKind
)

// Scope represents the scope of a variable.
type Scope int

const (
	// GlobalScope is the global scope.
	GlobalScope Scope = 1 << iota
	// SessionScope is the session scope.
	SessionScope
)

const (
	// GlobalScopeName is the global scope name.
	GlobalScopeName = "GLOBAL"
	// SessionScopeName is the session scope name.
	SessionScopeName = "SESSION"
)

// definition holds information used to manipulate variables.
type definition struct {
	Dummy            bool
	Name             Name
	Kind             Kind
	AllowedSetScopes Scope
	EvalType         types.EvalType

	GetValue func(container *Container) values.SQLValue
	SetValue func(container *Container, value values.SQLValue) error
}

var definitions = make(map[Name]*definition)

func init() {
	// Stub Variables
	for _, d := range stubVariableDefinitions {
		d.Dummy = true
		definitions[d.Name] = d
	}
}

// String returns the Scope as a string.
func (scope Scope) String() string {
	if (GlobalScope & scope) == GlobalScope {
		return GlobalScopeName
	} else if (SessionScope & scope) == SessionScope {
		return SessionScopeName
	}
	return ""
}

// ScopeFromString returns the Scope associated with
// the string form passed in.
func ScopeFromString(scope string) Scope {
	if strings.EqualFold(scope, GlobalScopeName) {
		return GlobalScope
	} else if strings.EqualFold(scope, SessionScopeName) {
		return SessionScope
	}
	return Scope(0)
}
