package variable

import (
	"strings"

	"github.com/10gen/sqlproxy/schema"
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

// Value represents the value of a variable.
type Value struct {
	// Name is the name of the variable.
	Name Name
	// Kind is the kind of the variable.
	Kind Kind
	// SQLType is the schema type the value will be.
	SQLType schema.SQLType
	// Value is the value of the variable.
	Value interface{}
	// RawValue is the actual value of the variable.
	RawValue interface{}
}

// definition holds information used to manipulate variables.
type definition struct {
	Dummy            bool
	Name             Name
	Kind             Kind
	AllowedSetScopes Scope
	SQLType          schema.SQLType

	GetValue    func(container *Container) interface{}
	SetValue    func(container *Container, value interface{}) error
	GetRawValue func(container *Container) interface{}
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
