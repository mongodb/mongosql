package eval

import (
	"fmt"
	"strconv"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/astprint"
	"github.com/10gen/mongoast/eval/bsoncompare"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/stringutil"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var (
	// ErrMemoryLimitExceeded is returned when the memory used during
	// expression evaluation exceeds the limit.
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")
	// ErrCouldNotBeFullyEvaluated is returned when an expression cannot be
	// fully evaluated to a constant.
	ErrCouldNotBeFullyEvaluated = errors.New("expression could not be fully evaluated")
)

var unaryFunctions = map[ast.UnaryOp]func(bsoncore.Value) (bsoncore.Value, error){
	ast.Abs:      bsonutil.Abs,
	ast.Ceil:     bsonutil.Ceil,
	ast.Exp:      bsonutil.Exp,
	ast.Floor:    bsonutil.Floor,
	ast.Ln:       bsonutil.Ln,
	ast.Log10:    bsonutil.Log10,
	ast.Not:      bsonutil.Not,
	ast.Sqrt:     bsonutil.Sqrt,
	ast.ToString: bsonutil.ToString,
}

var binaryFunctions = map[ast.BinaryOp]func(bsoncore.Value, bsoncore.Value) (bsoncore.Value, error){
	ast.Add: bsonutil.Add,
	ast.And: bsonutil.And,
	ast.Compare: func(left, right bsoncore.Value) (bsoncore.Value, error) {
		cmp, err := bsoncompare.Compare(left, right)
		if err != nil {
			return bsoncore.Value{}, err
		}
		return bsonutil.Int32(int32(cmp)), nil
	},
	ast.Concat:   bsonutil.Concat,
	ast.Divide:   bsonutil.Div,
	ast.Log:      bsonutil.Log,
	ast.Mod:      bsonutil.Mod,
	ast.Multiply: bsonutil.Mul,
	ast.Nor:      bsonutil.Nor,
	ast.Or:       bsonutil.Or,
	ast.Pow:      bsonutil.Pow,
	ast.Subtract: bsonutil.Sub,
}

var evaluableBinaryComparisonOperators = map[ast.BinaryOp]bool{
	ast.Equals:              true,
	ast.GreaterThan:         true,
	ast.GreaterThanOrEquals: true,
	ast.LessThan:            true,
	ast.LessThanOrEquals:    true,
	ast.NotEquals:           true,
}

// MongoDB system variables that we cannot evaluate here.
var unsupportedSystemVariables = stringutil.NewStringSet(
	"CLUSTER_TIME",
	"CURRENT",
	"DESCEND",
	"KEEP",
	"NOW",
	"PRUNE",
	"REMOVE",
	"ROOT",
)

// ValidateEvaluableExpr walks the provided expression and checks that this package
// is able to evaluate every expression node it contains. If this returns no error,
// then we know that evaluating the expression with EvaluateExpr should yield
// the exact same result as would MongoDB.
func ValidateEvaluableExpr(expr ast.Expr) error {
	v := &validateEvaluableExprVisitor{}
	v.Visit(expr)
	return v.err
}

// validateEvaluableExprRoot returns nil if this package is able to evaluate
// the root node of the provided expression, or an error if it is not.
func validateEvaluableExprRoot(expr ast.Expr) error {
	switch tn := expr.(type) {
	case *ast.AggExpr,
		*ast.Array,
		*ast.ArrayIndexRef,
		*ast.Conditional,
		*ast.Constant,
		*ast.Document,
		*ast.Exists,
		*ast.FieldOrArrayIndexRef,
		*ast.FieldRef,
		*ast.Trunc:
		return nil
	case *ast.Binary:
		if _, ok := binaryFunctions[tn.Op]; !ok && !evaluableBinaryComparisonOperators[tn.Op] {
			return fmt.Errorf("unsupported operator: %s", tn.Op)
		}
		return nil
	case *ast.Convert:
		// We currently only fully evaluate $convert expressions if the "to"
		// parameter resolves to an implemented type.
		toConst, toConstOK := tn.To.(*ast.Constant)
		if !toConstOK {
			return fmt.Errorf("$convert 'to' parameter must be a constant")
		}
		typ, err := convertToTypeFromConstant(toConst)
		if err != nil {
			return err
		}
		if !bsonutil.IsImplementedConvertToType(typ) {
			return fmt.Errorf("unsupported $convert 'to' parameter: %s", bsonutil.TypeToString(typ))
		}
		return nil
	case *ast.Let:
		// We currently are unable to fully evaluate expressions that modify
		// the CURRENT variable, which affects how fields are accessed (since
		// a field reference really accesses $$CURRENT.<field>).
		for _, v := range tn.Variables {
			if v.Name == "CURRENT" {
				return fmt.Errorf("unsupported $let variable: CURRENT")
			}
		}
		return nil
	case *ast.Unary:
		if _, ok := unaryFunctions[tn.Op]; !ok {
			return fmt.Errorf("unsupported operator: %s", tn.Op)
		}
		return nil
	case *ast.VariableRef:
		if unsupportedSystemVariables.Contains(tn.Name) {
			return fmt.Errorf("unsupported system variable reference: $$%s", tn.Name)
		}
		return nil
	}

	return fmt.Errorf("unsupported expression: %s", astprint.String(expr))
}

type validateEvaluableExprVisitor struct {
	err error
}

func (v *validateEvaluableExprVisitor) Visit(n ast.Node) ast.Node {
	if expr, ok := n.(ast.Expr); ok {
		if err := validateEvaluableExprRoot(expr); err != nil {
			v.err = err
		}
	}

	if v.err != nil {
		return n
	}

	return n.Walk(v)
}

// EvaluateExpr applies the expression to the doc.
func EvaluateExpr(expr ast.Expr, value bsoncore.Value, memoryLimit uint64) (bsoncore.Value, error) {
	return exprEvaluator{memoryLimit: memoryLimit}.evalToConstant(expr, value)
}

func extractConstant(expr ast.Expr) (bsoncore.Value, error) {
	switch te := expr.(type) {
	case *ast.AggExpr:
		return extractConstant(te.Expr)
	case *ast.Array:
		_, arr := bsoncore.AppendArrayStart(nil)
		for i, e := range te.Elements {
			v, err := extractConstant(e)
			if err != nil {
				return bsoncore.Value{}, err
			}
			arr = bsonutil.AppendValueElement(arr, strconv.Itoa(i), v)
		}
		arr, _ = bsoncore.AppendArrayEnd(arr, 0)
		return bsonutil.Array(arr), nil
	case *ast.Constant:
		return te.Value, nil
	case *ast.Document:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, e := range te.Elements {
			v, err := extractConstant(e.Expr)
			if err != nil {
				return bsoncore.Value{}, err
			}
			doc = bsonutil.AppendValueElement(doc, e.Name, v)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.VariableRef:
		return bsoncore.Value{}, errors.Errorf("use of undefined variable: %s", te.Name)
	default:
		return bsoncore.Value{}, ErrCouldNotBeFullyEvaluated
	}
}

// PartialEvaluateExpr evaluates as much of an expression as possible, returning a
// simplified expression that can be passed into mqlrun
func PartialEvaluateExpr(expr ast.Expr, value bsoncore.Value, memoryLimit uint64) (ast.Expr, error) {
	newExpr, _, err := exprEvaluator{memoryLimit: memoryLimit}.eval(expr, value)
	if err == ErrMemoryLimitExceeded {
		return nil, ErrMemoryLimitExceeded
	} else if err != nil {
		return nil, errors.Wrap(err, "failed evaluating expression")
	}

	return newExpr, nil
}

func makeConstant(v bsoncore.Value) (ast.Expr, uint64, error) {
	n := ast.NewConstant(v)
	return n, n.MemoryUsage(), nil
}

type exprEvaluator struct {
	isMatchLanguage bool
	memoryLimit     uint64
}

func (v exprEvaluator) evalToConstant(expr ast.Expr, value bsoncore.Value) (bsoncore.Value, error) {
	newExpr, _, err := v.eval(expr, value)
	if err == bsoncore.ErrElementNotFound || err == ErrMemoryLimitExceeded {
		return bsoncore.Value{}, err
	} else if err != nil {
		return bsoncore.Value{}, errors.Wrap(err, "failed evaluating expression")
	}
	c, err := extractConstant(newExpr)
	if err != nil {
		return bsoncore.Value{}, err
	}

	return c, nil
}

// NOTE: Every expression added here must also be added to the
// isExprRootEvaluable function in this package.
func (v exprEvaluator) eval(n ast.Expr, value bsoncore.Value) (ast.Expr, uint64, error) {
	// This simplest way to determine how much memory we're using would be to
	// call MemoryUsage() on each node. However, since MemoryUsage() has to
	// walk entire tree, calling it on every node would cause the algorithm to
	// run in quadratic time. In order to keep it linear, we return the memory
	// usage from this function, calculating it from the memory usages returned
	// from recursively evaluating the subnodes. We only call MemoryUsage() in
	// cases where we return a node from the original tree as is without
	// evaluating it.
	switch tn := n.(type) {
	case *ast.AggExpr:
		// A new expression evaluator is used here in order to get match semantics
		// instead of aggregation semantics inside of $expr. This is not a bug!
		an, amem, err := exprEvaluator{memoryLimit: v.memoryLimit}.eval(tn.Expr, value)
		if err != nil {
			return nil, 0, err
		}
		if an != tn.Expr {
			return ast.NewAggExpr(an), amem, nil
		}
		return n, amem, nil
	case *ast.ArrayIndexRef:
		in, imem, err := v.eval(tn.Index, value)
		if err != nil {
			return nil, 0, err
		}
		inC, ok := in.(*ast.Constant)
		if !ok {
			if in != tn.Index {
				return ast.NewArrayIndexRef(in, tn.Parent), imem + tn.Parent.MemoryUsage(), nil
			}
			return n, imem, nil
		}
		index, ok := bsonutil.AsInt32OK(inC.Value)
		if !ok {
			return nil, 0, errors.New("array index must be an integer")
		}

		if tn.Parent != nil {
			pn, pmem, err := v.eval(tn.Parent, value)
			if err != nil {
				return nil, 0, err
			}

			switch tpn := pn.(type) {
			case *ast.Array:
				if index < 0 || index >= int32(len(tpn.Elements)) {
					return nil, 0, errors.New("array index out of range")
				}
				return v.eval(tpn.Elements[index], value)
			case *ast.Constant:
				value = tpn.Value
			default:
				if pn != tn.Parent {
					return ast.NewArrayIndexRef(tn.Index, pn), tn.Index.MemoryUsage() + pmem, nil
				}
				return n, pmem + imem, nil
			}
		}

		switch value.Type {
		case bsontype.Array:
			elem, err := value.Array().IndexErr(uint(index))
			if err != nil {
				return nil, 0, err
			}
			value := elem.Value()
			return makeConstant(value)
		default:
			return nil, 0, bsoncore.ErrElementNotFound
		}
	case *ast.FieldOrArrayIndexRef:
		if tn.Parent != nil {
			pn, pmem, err := v.eval(tn.Parent, value)
			if err != nil {
				return nil, 0, err
			}

			pc, ok := pn.(*ast.Constant)
			if !ok {
				if pn != tn.Parent {
					return ast.NewFieldOrArrayIndexRef(tn.Number, pn), pmem, nil
				}
				return n, pmem, nil
			}
			value = pc.Value
		}

		switch value.Type {
		case bsontype.Array:
			elem, err := value.Array().IndexErr(uint(tn.Number))
			if err != nil {
				return nil, 0, err
			}
			value := elem.Value()
			return makeConstant(value)
		case bsontype.EmbeddedDocument:
			value, err := value.Document().LookupErr(strconv.Itoa(int(tn.Number)))
			if err != nil {
				return nil, 0, err
			}
			return makeConstant(value)
		default:
			return nil, 0, bsoncore.ErrElementNotFound
		}
	case *ast.Unary:
		return v.evalUnary(tn, value)
	case *ast.Binary:
		return v.evalBinary(tn, value)
	case *ast.Trunc:
		return v.evalTrunc(tn, value)
	case *ast.Conditional:
		return v.evalConditional(tn, value)
	case *ast.Constant:
		return makeConstant(tn.Value)
	case *ast.Document:
		newElements := make([]*ast.DocumentElement, 0, len(tn.Elements))
		var mem uint64
		for _, e := range tn.Elements {
			en, emem, err := v.eval(e.Expr, value)
			// If the bsontype of value is MinKey, it means this code has been
			// executed from ConstantPropagation, which requires returning an
			// err in the presence of a field reference.
			if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
				continue
			} else if err != nil {
				return nil, 0, err
			}
			newElements = append(newElements, ast.NewDocumentElement(e.Name, en))
			mem += uint64(len(e.Name)) + emem
			if v.memoryLimit > 0 && mem > v.memoryLimit {
				return nil, 0, ErrMemoryLimitExceeded
			}
		}
		return ast.NewDocument(newElements...), mem, nil
	case *ast.Array:
		newElements := make([]ast.Expr, len(tn.Elements))
		var mem uint64
		for i, e := range tn.Elements {
			en, emem, err := v.eval(e, value)
			// If the bsontype of value is MinKey, it means this code has been
			// executed from ConstantPropagation, which requires returning an
			// err in the presence of a field reference.
			if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
				en = ast.NewConstant(bsonutil.Null())
			} else if err != nil {
				return nil, 0, err
			}
			newElements[i] = en
			mem += emem
			if v.memoryLimit > 0 && mem > v.memoryLimit {
				return nil, 0, ErrMemoryLimitExceeded
			}
		}
		return ast.NewArray(newElements...), mem, nil
	case *ast.FieldRef:
		var pmem uint64
		if tn.Parent != nil {
			var pn ast.Expr
			var err error
			pn, pmem, err = v.eval(tn.Parent, value)
			if err != nil {
				return nil, 0, err
			}

			pc, ok := pn.(*ast.Constant)
			if !ok {
				if pn != tn.Parent {
					return ast.NewFieldRef(tn.Name, pn), uint64(len(tn.Name)) + pmem, nil
				}
				return n, uint64(len(tn.Name)) + pmem, nil
			}
			value = pc.Value
		}

		switch value.Type {
		case bsontype.Array:
			// In this case, re-evaluate the children for each element in the array and return a constant of the results
			// of each evaluation. In the case where the key doesn't exist, skip the element element completely.
			values, _ := value.Array().Values()
			_, newValues := bsoncore.AppendArrayStart(nil)
			childRef := ast.NewFieldRef(tn.Name, nil)
			i := 0
			for _, value := range values {
				pv, _, err := v.eval(childRef, value)
				switch {
				case err == bsoncore.ErrElementNotFound:
					continue
				case err != nil:
					return nil, 0, err
				}

				pc, ok := pv.(*ast.Constant)
				if !ok {
					return n, uint64(len(tn.Name)) + pmem, nil
				}

				newValues = bsonutil.AppendValueElement(newValues, strconv.Itoa(i), pc.Value)
				i++
			}

			newValues, _ = bsoncore.AppendArrayEnd(newValues, 0)
			return makeConstant(bsonutil.Array(newValues))
		case bsontype.EmbeddedDocument:
			value, err := value.Document().LookupErr(tn.Name)
			if err != nil {
				return nil, 0, err
			}
			return makeConstant(value)
		default:
			return nil, 0, bsoncore.ErrElementNotFound
		}
	case *ast.Function:
		return v.evalFunction(tn, value)
	case *ast.Let:
		return v.evalLet(tn, value)
	case *ast.Exists:
		_, _, err := v.eval(tn.Ref, value)

		if err != nil && err != bsoncore.ErrElementNotFound {
			return nil, 0, err
		}

		if tn.Exists == (err == nil) {
			return makeConstant(bsonutil.True)
		}

		return makeConstant(bsonutil.False)
	case *ast.Convert:
		return v.evalConvert(tn, value)
	default:
		return n, 0, nil
	}
}

func (v exprEvaluator) evalUnary(n *ast.Unary, value bsoncore.Value) (ast.Expr, uint64, error) {
	expr, mem, err := v.eval(n.Expr, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		switch n.Op {
		case ast.Not:
			return makeConstant(bsonutil.True)
		default:
			return makeConstant(bsonutil.Null())
		}
	} else if err != nil {
		return nil, 0, err
	}

	exprc, ok := expr.(*ast.Constant)
	if !ok {
		if expr != n.Expr {
			return ast.NewUnary(n.Op, expr), mem, nil
		}
		return n, mem, nil
	}

	unaryFunc, ok := unaryFunctions[n.Op]
	if !ok {
		if expr != n.Expr {
			return ast.NewUnary(n.Op, expr), mem, nil
		}
		return n, mem, nil
	}

	result, err := unaryFunc(exprc.Value)
	if err != nil {
		return nil, 0, err
	}
	return makeConstant(result)
}

func (v exprEvaluator) evalBinary(n *ast.Binary, value bsoncore.Value) (ast.Expr, uint64, error) {
	left, lmem, err := v.eval(n.Left, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.evalBinaryMissingField(n.Op, n.Right, value, false)
	} else if err != nil {
		return nil, 0, err
	}

	right, rmem, err := v.eval(n.Right, value)
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.evalBinaryMissingField(n.Op, n.Left, value, true)
	} else if err != nil {
		return nil, 0, err
	}

	lc, err1 := extractConstant(left)
	rc, err2 := extractConstant(right)
	if err1 != nil || err2 != nil {
		if left != n.Left || right != n.Right {
			return ast.NewBinary(n.Op, left, right), lmem + rmem, nil
		}
		return n, lmem + rmem, nil
	}

	if evaluableBinaryComparisonOperators[n.Op] {
		result, err := v.compare(n.Op, lc, rc)
		if err != nil {
			return nil, 0, err
		}
		return makeConstant(bsonutil.Boolean(result))
	}

	binaryFunc, ok := binaryFunctions[n.Op]
	if !ok {
		if left != n.Left || right != n.Right {
			return ast.NewBinary(n.Op, left, right), lmem + rmem, nil
		}
		return n, lmem + rmem, nil
	}

	result, err := binaryFunc(lc, rc)
	if err != nil {
		return nil, 0, err
	}
	return makeConstant(result)
}

func (v exprEvaluator) evalTrunc(n *ast.Trunc, value bsoncore.Value) (ast.Expr, uint64, error) {
	number, nmem, err := v.eval(n.Number, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return makeConstant(bsonutil.Null())
	} else if err != nil {
		return nil, 0, err
	}
	precision, pmem, err := v.eval(n.Precision, value)
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return makeConstant(bsonutil.Null())
	} else if err != nil {
		return nil, 0, err
	}

	nc, ok1 := number.(*ast.Constant)
	pc, ok2 := precision.(*ast.Constant)

	if !ok1 || !ok2 {
		if number != n.Number || precision != n.Precision {
			return ast.NewTrunc(number, precision), nmem + pmem, nil
		}
		return n, nmem + pmem, nil
	}

	result, err := bsonutil.Trunc(nc.Value, pc.Value)
	if err != nil {
		return nil, 0, err
	}

	return makeConstant(result)
}

func (v exprEvaluator) compare(op ast.BinaryOp, left bsoncore.Value, right bsoncore.Value) (bool, error) {
	isAggLanguage := !v.isMatchLanguage
	if isAggLanguage {
		return v.compareScalar(op, left, right)
	}

	leftIsArr := left.Type == bsontype.Array
	rightIsArr := right.Type == bsontype.Array

	compareArrToScalar := func(arrValues bsoncore.Value, scalar bsoncore.Value) (bool, error) {
		elems, _ := arrValues.Array().Values()
		opToUse := op
		if op == ast.NotEquals {
			opToUse = ast.Equals
		}
		for _, elem := range elems {
			if result, err := v.compareScalar(opToUse, elem, scalar); err != nil {
				return false, err
			} else if result {
				return op != ast.NotEquals, nil
			}
		}

		return op == ast.NotEquals, nil
	}

	switch {
	case leftIsArr == rightIsArr: // Compare the values holistically if they are both scalars or arrays.
		return v.compareScalar(op, left, right)
	case leftIsArr && !rightIsArr: // If its array to scalar, compare the scalar piece-wise to the array.
		return compareArrToScalar(left, right)
	case !leftIsArr && rightIsArr: // Finally, if it is scalar to array, it is automaticaly false, except in $ne.
		return (op == ast.NotEquals), nil
	default:
		return v.compareScalar(op, left, right)
	}
}

func (v exprEvaluator) compareScalar(op ast.BinaryOp, left bsoncore.Value, right bsoncore.Value) (bool, error) {
	switch op {
	case ast.Equals:
		if cmp, err := bsoncompare.Compare(left, right); err != nil {
			return false, err
		} else if cmp == 0 {
			return true, nil
		} else {
			return false, nil
		}
	case ast.GreaterThan:
		if cmp, err := bsoncompare.Compare(left, right); err != nil {
			return false, err
		} else if cmp > 0 {
			return true, nil
		} else {
			return false, nil
		}

	case ast.GreaterThanOrEquals:
		if cmp, err := bsoncompare.Compare(left, right); err != nil {
			return false, err
		} else if cmp >= 0 {
			return true, nil
		} else {
			return false, nil
		}

	case ast.LessThan:
		if cmp, err := bsoncompare.Compare(left, right); err != nil {
			return false, err
		} else if cmp < 0 {
			return true, nil
		} else {
			return false, nil
		}

	case ast.LessThanOrEquals:
		if cmp, err := bsoncompare.Compare(left, right); err != nil {
			return false, err
		} else if cmp <= 0 {
			return true, nil
		} else {
			return false, nil
		}

	case ast.NotEquals:
		if cmp, err := bsoncompare.Compare(left, right); err != nil {
			return false, err
		} else if cmp != 0 {
			return true, nil
		} else {
			return false, nil
		}
	default:
		panic(fmt.Sprintf("cannot call compareBinary with op %s", op))
	}
}

func (v exprEvaluator) evalBinaryMissingField(op ast.BinaryOp, other ast.Expr, value bsoncore.Value, flip bool) (ast.Expr, uint64, error) {
	if flip {
		op = op.Flip()
	}

	otherValue, otherMem, err := v.eval(other, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	// The bsontype.MinKey check is not strictly needed here because this function
	// should never be called in such a case. This check is maintained for future
	// compatability and completeness.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.evalBinaryMissingBothFields(op)
	} else if err != nil {
		return nil, 0, err
	}

	if v.isMatchLanguage {
		if otherConst, ok := otherValue.(*ast.Constant); ok {
			switch op {
			case ast.Equals, ast.LessThanOrEquals, ast.GreaterThanOrEquals:
				if otherConst.Value.Type == bsontype.Null {
					return makeConstant(bsonutil.True)
				}
				return makeConstant(bsonutil.False)
			case ast.LessThan, ast.GreaterThan:
				return makeConstant(bsonutil.False)
			case ast.NotEquals:
				if otherConst.Value.Type == bsontype.Null {
					return makeConstant(bsonutil.False)
				}
				return makeConstant(bsonutil.True)
			}
		}

		panic("unexpected match expressions")
	}

	switch op {
	case ast.Compare:
		if flip {
			return makeConstant(bsonutil.Int32(1))
		}
		return makeConstant(bsonutil.Int32(-1))
	case ast.LessThan, ast.LessThanOrEquals, ast.NotEquals:
		return makeConstant(bsonutil.True)
	case ast.Or:
		if otherConst, ok := otherValue.(*ast.Constant); ok {
			return makeConstant(
				bsonutil.Boolean(
					bsonutil.CoerceToBoolean(otherConst.Value),
				),
			)
		}
		return ast.NewBinary(
			ast.Or,
			otherValue,
			ast.NewConstant(bsonutil.False),
		), otherMem, nil
	case ast.Divide, ast.Log, ast.Mod, ast.Pow, ast.Subtract, ast.Add, ast.Multiply:
		return makeConstant(bsonutil.Null())
	default:
		return makeConstant(bsonutil.False)
	}
}

func (v exprEvaluator) evalBinaryMissingBothFields(op ast.BinaryOp) (ast.Expr, uint64, error) {
	switch op {
	case ast.Compare:
		return makeConstant(bsonutil.Int32(0))
	case ast.Equals, ast.LessThanOrEquals, ast.GreaterThanOrEquals:
		return makeConstant(bsonutil.True)
	case ast.Divide, ast.Log, ast.Mod, ast.Pow, ast.Subtract, ast.Add, ast.Multiply:
		return makeConstant(bsonutil.Null())
	default:
		return makeConstant(bsonutil.False)
	}
}

func (v exprEvaluator) evalConditional(n *ast.Conditional, value bsoncore.Value) (ast.Expr, uint64, error) {
	ifExpr, ifMem, err := v.eval(n.If, value)
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.eval(n.Else, value)
	} else if err != nil {
		return nil, 0, err
	}

	switch te := ifExpr.(type) {
	case *ast.Constant:
		if cmp := bsonutil.CoerceToBoolean(te.Value); cmp {
			evaled, evaledMem, err := v.eval(n.Then, value)
			if err != nil {
				return n.Then, n.Then.MemoryUsage(), nil
			}
			return evaled, evaledMem, nil
		}

		evaled, evaledMem, err := v.eval(n.Else, value)
		if err != nil {
			return n.Else, n.Else.MemoryUsage(), nil
		}
		return evaled, evaledMem, nil

	default:
		thenExpr, thenMem, err := v.eval(n.Then, value)
		if err != nil {
			return nil, 0, err
		}
		elseExpr, elseMem, err := v.eval(n.Else, value)
		if err != nil {
			return nil, 0, err
		}
		if ifExpr != n.If || thenExpr != n.Then || elseExpr != n.Else {
			return ast.NewConditional(ifExpr, thenExpr, elseExpr), ifMem + thenMem + elseMem, nil
		}
		return n, ifMem + thenMem + elseMem, nil
	}
}

func (v exprEvaluator) evalLet(n *ast.Let, value bsoncore.Value) (ast.Expr, uint64, error) {
	variables := make(map[string]ast.Expr)
	var mem uint64
	for _, letVar := range n.Variables {
		letVarExpr, letVarMem, err := v.eval(letVar.Expr, value)
		if err != nil {
			return nil, 0, err
		}
		variables[letVar.Name] = letVarExpr
		mem += uint64(len(letVar.Name)) + letVarMem
	}
	return v.eval(
		SubstituteVariables(n.Expr, variables).(ast.Expr),
		value,
	)
}

func (v exprEvaluator) evalConvert(n *ast.Convert, value bsoncore.Value) (ast.Expr, uint64, error) {
	evalOnNull := func() (ast.Expr, uint64, error) {
		if n.OnNull != nil {
			return v.eval(n.OnNull, value)
		}
		return makeConstant(bsonutil.Null())
	}

	input, inputMem, err := v.eval(n.Input, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		// Handle missing input value.
		return evalOnNull()
	} else if err != nil {
		return nil, 0, err
	}

	to, toMem, err := v.eval(n.To, value)
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		// Same MinKey situation as above.
		return makeConstant(bsonutil.Null())
	} else if err != nil {
		return nil, 0, err
	}

	partialEvalConvert := func() (ast.Expr, uint64, error) {
		mem := inputMem + toMem
		if n.OnError != nil {
			mem += n.OnError.MemoryUsage()
		}
		if n.OnNull != nil {
			mem += n.OnNull.MemoryUsage()
		}
		if input != n.Input && to != n.To {
			return ast.NewConvert(input, to, n.OnError, n.OnNull), mem, nil
		}
		return n, mem, nil
	}

	inputConst, inputOK := input.(*ast.Constant)
	toConst, toOK := to.(*ast.Constant)
	if !inputOK || !toOK {
		return partialEvalConvert()
	}

	toType, err := convertToTypeFromConstant(toConst)
	if err != nil {
		return nil, 0, err
	}

	// Handle null input value.
	if inputConst.Value.Type == bsontype.Null {
		return evalOnNull()
	}

	// We currently only support $convert with the string 'to' type. We should
	// return a partially evaluated expression for all other types.
	if !bsonutil.IsImplementedConvertToType(toType) {
		return partialEvalConvert()
	}

	result, err := bsonutil.Convert(inputConst.Value, toType)
	if err != nil {
		if n.OnError != nil {
			return v.eval(n.OnError, value)
		}
		return nil, 0, err
	}

	return makeConstant(result)
}

// convertToTypeFromConstant extracts a BSON type from a constant expression,
// which must be either a string value or an integer numeric value.
func convertToTypeFromConstant(n *ast.Constant) (bsontype.Type, error) {
	switch n.Value.Type {
	case bsontype.String:
		typ, ok := bsonutil.StringToTypeOK(n.Value.StringValue())
		if !ok {
			return 0, fmt.Errorf("unknown type name: %s", n.Value.StringValue())
		}
		return typ, nil
	case bsontype.Int32, bsontype.Int64, bsontype.Double, bsontype.Decimal128:
		typeInt, ok := bsonutil.AsInt64OK(n.Value)
		if !ok {
			return 0, fmt.Errorf("in $convert, numeric 'to' argument is not an integer")
		}
		typ, ok := bsonutil.Int64ToTypeOK(typeInt)
		if !ok {
			return 0, fmt.Errorf("in $convert, numeric value for 'to' does not correspond to a BSON type: %d", typeInt)
		}
		return typ, nil
	default:
		return 0, fmt.Errorf("$convert's 'to' argument must be a string or number, but is %s", bsonutil.TypeToString(n.Value.Type))
	}
}
