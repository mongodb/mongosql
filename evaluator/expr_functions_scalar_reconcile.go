package evaluator

import "fmt"

// This file contains custom reconciliation implementations for functions with
// non-standard reconciliation needs.
//
// Any scalar function in scalar_functions.yml defined with `custom_reconcile:
// true` will be generated with a `reconcile()` function that calls a custom
// function defined in this file instead of performing the default behavior. If
// a function marked as `custom_reconcile` does not have a reconciliation
// implementation provided in this file, the generated code will fail to
// compile.

func wrapAllArgs(funcName string) func([]SQLExpr) []SQLExpr {
	fn := func(args []SQLExpr) []SQLExpr {
		wrappedArgs := []SQLExpr{}
		for _, arg := range args {
			wrapped, err := NewSQLScalarFunctionExpr(funcName, []SQLExpr{arg})
			if err != nil {
				panic(err)
			}
			wrappedArgs = append(wrappedArgs, wrapped)
		}
		return wrappedArgs
	}
	return fn
}

func wrapArgsAtIndices(funcName string, indices ...int) func([]SQLExpr) []SQLExpr {
	fn := func(args []SQLExpr) []SQLExpr {
		newArgs := make([]SQLExpr, len(args))
		copy(newArgs, args)

		for _, idx := range indices {
			wrapped, err := NewSQLScalarFunctionExpr(funcName, []SQLExpr{newArgs[idx]})
			if err != nil {
				panic(fmt.Errorf("failed to construct scalar function %q", funcName))
			}
			newArgs[idx] = wrapped
		}

		return newArgs
	}
	return fn
}

var (
	lastDayReconcileArgs = wrapAllArgs("date")
	toDaysReconcileArgs  = wrapAllArgs("date")

	toSecondsReconcileArgs = wrapAllArgs("timestamp")

	dateAddReconcileArgs = wrapArgsAtIndices("timestamp", 0)
	dateSubReconcileArgs = wrapArgsAtIndices("timestamp", 0)

	weekWithDefaultModeReconcileArgs     = wrapArgsAtIndices("date", 0)
	weekWithModeReconcileArgs            = wrapArgsAtIndices("date", 0)
	yearWeekWithDefaultModeReconcileArgs = wrapArgsAtIndices("date", 0)
	yearWeekWithModeReconcileArgs        = wrapArgsAtIndices("date", 0)

	timestampAddReconcileArgs = wrapArgsAtIndices("timestamp", 2)

	timestampDiffReconcileArgs = wrapArgsAtIndices("timestamp", 1, 2)
)
