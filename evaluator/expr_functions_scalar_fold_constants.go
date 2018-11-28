package evaluator

// This file contains custom constant folding implementations for functions with
// non-standard constant folding needs.
//
// Any scalar function in scalar_functions.yml defined with `custom_fold_constants:
// true` will be generated with a `FoldConstants()` function that calls a custom
// function defined in this file instead of performing the default behavior
// (return NULL if any argument is NULL). If a function marked as
// `custom_fold_constants` does not have a FoldConstants implementation provided in this
// file, the generated code will fail to compile.

func (f *baseScalarFunctionExpr) charFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	valArgs := make([]SQLValue, len(f.args))
	allValues := true
	for i, arg := range f.args {
		if val, ok := arg.(SQLValue); ok {
			valArgs[i] = val
		} else {
			allValues = false
		}
	}
	if !allValues {
		return nil, false
	}
	val, err := f.charEvaluate(cfg.sqlValueKind, cfg.collation, valArgs)
	if err != nil {
		return nil, false
	}
	return val, true
}

func (f *baseScalarFunctionExpr) concatWsFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	valArgs := make([]SQLValue, len(f.args))
	allValues := true
	for i, arg := range f.args {
		if val, ok := arg.(SQLValue); ok {
			valArgs[i] = val
		} else {
			allValues = false
		}
	}
	if len(f.args) >= 2 {
		firstVal, ok := f.args[0].(SQLValue)
		if ok && firstVal.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), true
		}
	}
	if !allValues {
		return nil, false
	}
	// Call the function that contains the appropriate evaluation logic.
	val, err := f.concatWsEvaluate(cfg.sqlValueKind, cfg.collation, valArgs)
	if err != nil {
		return nil, false
	}
	return val, true
}

func (f *baseScalarFunctionExpr) locateFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	if hasNullExpr(f.args[:2]...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), true
	}
	valArgs := make([]SQLValue, len(f.args))
	allValues := true
	for i, arg := range f.args {
		if val, ok := arg.(SQLValue); ok {
			valArgs[i] = val
		} else {
			allValues = false
		}
	}
	if !allValues {
		return nil, false
	}
	val, err := f.locateEvaluate(cfg.sqlValueKind, cfg.collation, valArgs)
	if err != nil {
		return nil, false
	}
	return val, true
}

func (f *baseScalarFunctionExpr) nopushdownFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	return nil, false
}

func (f *baseScalarFunctionExpr) randFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	return nil, false
}

func (f *baseScalarFunctionExpr) substringIndexFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	if hasNullExpr(f.args...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), true
	}

	if v, ok := f.args[2].(SQLValue); ok {
		if Int64(v) == 0 {
			return NewSQLVarchar(cfg.sqlValueKind, ""), true
		}
	}

	return nil, false
}
