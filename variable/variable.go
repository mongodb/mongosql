package variable

import "github.com/10gen/sqlproxy/schema"

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
}
