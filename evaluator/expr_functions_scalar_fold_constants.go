package evaluator

import "github.com/10gen/sqlproxy/evaluator/values"

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
	valArgs := make([]values.SQLValue, len(f.args))
	allValues := true
	for i, arg := range f.args {
		if val, ok := arg.(SQLValueExpr); ok {
			valArgs[i] = val.Value
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
	return NewSQLValueExpr(val), true
}

func (f *baseScalarFunctionExpr) concatWsFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	valArgs := make([]values.SQLValue, len(f.args))
	allValues := true
	for i, arg := range f.args {
		if val, ok := arg.(SQLValueExpr); ok {
			valArgs[i] = val.Value
		} else {
			allValues = false
		}
	}
	if len(f.args) >= 2 {
		firstVal, ok := f.args[0].(SQLValueExpr)
		if ok && firstVal.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), true
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
	return NewSQLValueExpr(val), true
}

func (f *baseScalarFunctionExpr) locateFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	if hasNullExpr(f.args[:2]...) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), true
	}
	valArgs := make([]values.SQLValue, len(f.args))
	allValues := true
	for i, arg := range f.args {
		if val, ok := arg.(SQLValueExpr); ok {
			valArgs[i] = val.Value
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
	return NewSQLValueExpr(val), true
}

func (f *baseScalarFunctionExpr) nopushdownFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	return nil, false
}

func (f *baseScalarFunctionExpr) randFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	return nil, false
}

func (f *baseScalarFunctionExpr) substringIndexFoldConstants(cfg *OptimizerConfig) (SQLExpr, bool) {
	if hasNullExpr(f.args...) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), true
	}

	if v, ok := f.args[2].(SQLValueExpr); ok {
		if values.Int64(v.Value) == 0 {
			return NewSQLValueExpr(values.NewSQLVarchar(cfg.sqlValueKind, "")), true
		}
	}

	return nil, false
}
