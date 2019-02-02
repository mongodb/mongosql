package evaluator

//go:generate go run testdata/generate_scalar_functions.go scalar_functions.yml expr_functions_scalar_generated.go

import (
	"fmt"
	"math"
	"strings"

	"github.com/10gen/sqlproxy/evaluator/types"
)

// SQLScalarFunctionExpr is the interface describing scalar function expressions.
type SQLScalarFunctionExpr interface {
	SQLExpr
	iSQLScalarFunctionExpr()
	invokedName() string
	getArgsPointer() *[]SQLExpr
}

type baseScalarFunctionExpr struct {
	// invokedAs contains the name that  was used to invoke this scalar function
	// (in some cases, one scalar function may be invoked via multiple names).
	invokedAs string
	// args contains the slice of arguments to this scalar function.
	args []SQLExpr
	// expectedTypes contains a slice of EvalTypes that indicates the correct
	// type for each argument to this scalar function. If the function is
	// variadic, the final EvalType in the slice represents the expected type
	// for all the variadic args.
	expectedTypes []types.EvalType
	// variadic indicates whether this function's final argument accepts 1 or
	// more parameters of the same type instead of just one.
	variadic bool
	// returnTypeFunc is a function that indicates the EvalType that this scalar function returns.
	returnTypeFunc func([]SQLExpr) types.EvalType
}

func (baseScalarFunctionExpr) iSQLScalarFunctionExpr() {}

func (sf baseScalarFunctionExpr) invokedName() string {
	return sf.invokedAs
}

func (sf baseScalarFunctionExpr) getArgsPointer() *[]SQLExpr {
	return &sf.args
}

// Children returns a slice of all the Node children of the Node.
func (sf baseScalarFunctionExpr) Children() []Node {
	out := make([]Node, len(sf.args))
	for i := range sf.args {
		out[i] = sf.args[i]
	}
	return out
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (sf *baseScalarFunctionExpr) ReplaceChild(i int, n Node) {
	if 0 <= i && i < len(sf.args) {
		sf.args[i] = panicIfNotSQLExpr("baseScalarFunctionExpr", n)
		return
	}
	panicWithInvalidIndex("baseScalarFunctionExpr", i, len(sf.args)-1)
}

// argTypes returns a slice of length len(sf.args) that contains the expected
// EvalType for each argument. This function assumes that the sf.args and
// sf.expectedTypes slices are correct, and does not perform any validation on
// those fields.
func (sf baseScalarFunctionExpr) argTypes() []types.EvalType {
	if len(sf.expectedTypes) == 0 {
		return nil
	}
	types := []types.EvalType{}
	lastIdx := len(sf.expectedTypes) - 1
	for i := range sf.args {
		idx := int(math.Min(float64(lastIdx), float64(i)))
		types = append(types, sf.expectedTypes[idx])
	}
	return types
}

// validateArgCount ensures that the slice of arguments in sf.args has the
// correct number of arguments.
func (sf baseScalarFunctionExpr) validateArgCount() error {
	if sf.variadic {
		if len(sf.args) < len(sf.expectedTypes) {
			return fmt.Errorf("expected at least %d arguments, but found %d", len(sf.expectedTypes), len(sf.args))
		}
	} else if len(sf.args) != len(sf.expectedTypes) {
		return fmt.Errorf("expected %d arguments, but found %d", len(sf.expectedTypes), len(sf.args))
	}
	return nil
}

// validateArgs ensures that the slice of arguments in sf.args is valid (i.e.
// that it has the correct number of arguments and each argument has the correct
// EvalType). If validation fails, an error is returned.
// nolint: megacheck
func (sf baseScalarFunctionExpr) validateArgs() error {
	err := sf.validateArgCount()
	if err != nil {
		return err
	}

	argTypes := sf.argTypes()
	for i, typ := range argTypes {
		// If the type is declared as polymorphic, there is nothing to check,
		// as all types are conforming.
		if typ == types.EvalPolymorphic {
			continue
		}
		if sf.args[i].EvalType() != typ {
			return fmt.Errorf(
				"expected EvalType %x at index %d, but got %x",
				typ, i, sf.args[i].EvalType(),
			)
		}
	}

	return nil
}

func (sf baseScalarFunctionExpr) String() string {
	var exprs []string
	for _, expr := range sf.args {
		exprs = append(exprs, expr.String())
	}
	return fmt.Sprintf("%s(%s)", sf.invokedAs, strings.Join(exprs, ","))
}

func (sf baseScalarFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLScalarFunctionExpr(%s)", sf.invokedAs)
}

func (sf baseScalarFunctionExpr) EvalType() types.EvalType {
	return sf.returnTypeFunc(sf.args)
}
