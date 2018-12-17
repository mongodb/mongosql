package evaluator

// This file contains custom normalization implementations for functions with
// non-standard normalization needs.
//
// Any scalar function in scalar_functions.yml defined with `custom_normalize:
// true` will be generated with a `Normalize()` function that calls a custom
// function defined in this file instead of performing the default behavior
// (return NULL if any argument is NULL). If a function marked as
// `custom_normalize` does not have a normalize implementation provided in this
// file, the generated code will fail to compile.

func (f *baseScalarFunctionExpr) coalesceNormalize(kind SQLValueKind) (SQLExpr, bool) {
	return nil, false
}

func (f *baseScalarFunctionExpr) concatWsNormalize(kind SQLValueKind) (SQLExpr, bool) {
	if len(f.args) >= 2 {
		firstVal, ok := f.args[0].(SQLValue)
		if ok && firstVal.IsNull() {
			return NewSQLNull(kind, f.EvalType()), true
		}
	}

	return nil, false
}

func (f *baseScalarFunctionExpr) convNormalize(kind SQLValueKind) (SQLExpr, bool) {
	if hasNullExpr(f.args...) {
		return NewSQLNull(kind, f.EvalType()), true
	}

	if v, ok := f.args[1].(SQLValue); ok {
		if baseIsInvalid(absInt64(Int64(v))) {
			return NewSQLNull(kind, f.EvalType()), true
		}
	}

	if v, ok := f.args[2].(SQLValue); ok {
		if baseIsInvalid(absInt64(Int64(v))) {
			return NewSQLNull(kind, f.EvalType()), true
		}
	}

	return nil, false
}

func (f *baseScalarFunctionExpr) eltNormalize(kind SQLValueKind) (SQLExpr, bool) {
	if hasNullExpr(f.args[0]) {
		return NewSQLNull(kind, f.EvalType()), true
	}

	if v, ok := f.args[0].(SQLValue); ok {
		idx := Int64(v)
		if idx <= 0 || int(idx) > len(f.args) {
			return NewSQLNull(kind, f.EvalType()), true
		}
	}

	return nil, false
}

func (f *baseScalarFunctionExpr) fieldNormalize(kind SQLValueKind) (SQLExpr, bool) {
	if hasNullExpr(f.args...) {
		return NewSQLInt64(kind, 0), true
	}

	return nil, false
}

func (f *baseScalarFunctionExpr) intervalNormalize(kind SQLValueKind) (SQLExpr, bool) {
	sqlVal, ok := f.args[0].(SQLValue)
	if ok && sqlVal.IsNull() {
		return NewSQLInt64(kind, -1), true
	}
	return nil, false
}

func (f *baseScalarFunctionExpr) locateNormalize(kind SQLValueKind) (SQLExpr, bool) {
	if hasNullExpr(f.args[:2]...) {
		return NewSQLNull(kind, f.EvalType()), true
	}

	return nil, false
}

func (f *baseScalarFunctionExpr) nopushdownNormalize(kind SQLValueKind) (SQLExpr, bool) {
	return nil, false
}

func (f *baseScalarFunctionExpr) substringIndexNormalize(kind SQLValueKind) (SQLExpr, bool) {
	if hasNullExpr(f.args...) {
		return NewSQLNull(kind, f.EvalType()), true
	}

	if v, ok := f.args[2].(SQLValue); ok {
		if Int64(v) == 0 {
			return NewSQLVarchar(kind, ""), true
		}
	}

	return nil, false
}
