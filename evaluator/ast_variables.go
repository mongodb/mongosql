package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/schema"
)

var systemVars = map[string]SQLValue{
	"max_allowed_packet": SQLInt(4194304),
}

// SQLVariableType indicates the type of variable being referenced.
type SQLVariableType int

const (
	UserDefinedVariable SQLVariableType = iota
	SystemVariable
)

//
// SQLVariableExpr represents a variable lookup.
//
type SQLVariableExpr struct {
	Name         string
	VariableType SQLVariableType
}

func (v *SQLVariableExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	switch v.VariableType {
	case SystemVariable:
		return systemVars[v.Name], nil
	}

	return nil, fmt.Errorf("unknown variable %s", v.String())
}

func (v *SQLVariableExpr) String() string {
	switch v.VariableType {
	case UserDefinedVariable:
		return "@" + v.Name
	case SystemVariable:
		return "@@" + v.Name
	default:
		return v.Name
	}
}

func (v *SQLVariableExpr) Type() schema.SQLType {
	switch v.VariableType {
	case SystemVariable:
		if value, ok := systemVars[v.Name]; ok {
			return value.Type()
		}
	}

	return schema.SQLNone
}
