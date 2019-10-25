package eval

import (
	"fmt"
	"strconv"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/eval/bsoncompare"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var unaryFunctions = map[ast.UnaryOp]func(bsoncore.Value) (bsoncore.Value, error){
	ast.Abs:   bsonutil.Abs,
	ast.Ceil:  bsonutil.Ceil,
	ast.Floor: bsonutil.Floor,
	ast.Exp:   bsonutil.Exp,
	ast.Ln:    bsonutil.Ln,
	ast.Log10: bsonutil.Log10,
	ast.Sqrt:  bsonutil.Sqrt,
}

var binaryFunctions = map[ast.BinaryOp]func(bsoncore.Value, bsoncore.Value) (bsoncore.Value, error){
	ast.Divide:   bsonutil.Div,
	ast.Log:      bsonutil.Log,
	ast.Mod:      bsonutil.Mod,
	ast.Pow:      bsonutil.Pow,
	ast.Subtract: bsonutil.Sub,
	ast.Add:      bsonutil.Add,
	ast.Multiply: bsonutil.Mul,
}

// EvaluateExpr applies the expression to the doc.
func EvaluateExpr(expr ast.Expr, value bsoncore.Value) (bsoncore.Value, error) {
	return exprEvaluator{}.evalToConstant(expr, value)
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
		return bsoncore.Value{}, errors.New("expression could not be fully evaluated")
	}
}

// PartialEvaluateExpr evaluates as much of an expression as possible, returning a
// simplified expression that can be passed into mqlrun
func PartialEvaluateExpr(expr ast.Expr, value bsoncore.Value) (ast.Expr, error) {
	newExpr, err := exprEvaluator{}.eval(expr, value)
	if err != nil {
		return nil, errors.Wrap(err, "failed evaluating expression")
	}

	return newExpr, nil
}

type exprEvaluator struct {
	isMatchExpr bool
}

func (v exprEvaluator) evalToConstant(expr ast.Expr, value bsoncore.Value) (bsoncore.Value, error) {
	newExpr, err := v.eval(expr, value)
	if err == bsoncore.ErrElementNotFound {
		return bsoncore.Value{}, bsoncore.ErrElementNotFound
	} else if err != nil {
		return bsoncore.Value{}, errors.Wrap(err, "failed evaluating expression")
	}
	c, err := extractConstant(newExpr)
	if err != nil {
		return bsoncore.Value{}, err
	}

	return c, nil
}

func (v exprEvaluator) eval(n ast.Expr, value bsoncore.Value) (ast.Expr, error) {
	switch tn := n.(type) {
	case *ast.AggExpr:
		// A new expression evaluator is used here in order to get match semantics
		// instead of aggregation semantics inside of $expr. This is not a bug!
		an, err := exprEvaluator{}.eval(tn.Expr, value)
		if err != nil {
			return nil, err
		}
		if an != tn.Expr {
			return ast.NewAggExpr(an), nil
		}
		return n, nil
	case *ast.ArrayIndexRef:
		in, err := v.eval(tn.Index, value)
		if err != nil {
			return nil, err
		}
		inC, ok := in.(*ast.Constant)
		if !ok {
			if in != tn.Index {
				return ast.NewArrayIndexRef(in, tn.Parent), nil
			}
			return n, nil
		}
		index, ok := bsonutil.AsInt32OK(inC.Value)
		if !ok {
			return nil, errors.New("array index must be an integer")
		}

		if tn.Parent != nil {
			pn, err := v.eval(tn.Parent, value)
			if err != nil {
				return nil, err
			}

			switch tpn := pn.(type) {
			case *ast.Array:
				if index < 0 || index >= int32(len(tpn.Elements)) {
					return nil, errors.New("array index out of range")
				}
				return v.eval(tpn.Elements[index], value)
			case *ast.Constant:
				value = tpn.Value
			default:
				if pn != tn.Parent {
					return ast.NewArrayIndexRef(tn.Index, pn), nil
				}
				return n, nil
			}
		}

		switch value.Type {
		case bsontype.Array:
			value, err := value.Array().IndexErr(uint(index))
			if err != nil {
				return nil, err
			}
			return ast.NewConstant(value.Value()), nil
		default:
			return nil, bsoncore.ErrElementNotFound
		}
	case *ast.FieldOrArrayIndexRef:
		if tn.Parent != nil {
			pn, err := v.eval(tn.Parent, value)
			if err != nil {
				return nil, err
			}

			pc, ok := pn.(*ast.Constant)
			if !ok {
				if pn != tn.Parent {
					return ast.NewFieldOrArrayIndexRef(tn.Number, pn), nil
				}
				return n, nil
			}
			value = pc.Value
		}

		switch value.Type {
		case bsontype.Array:
			value, err := value.Array().IndexErr(uint(tn.Number))
			if err != nil {
				return nil, err
			}
			return ast.NewConstant(value.Value()), nil
		case bsontype.EmbeddedDocument:
			value, err := value.Document().LookupErr(strconv.Itoa(int(tn.Number)))
			if err != nil {
				return nil, err
			}
			return ast.NewConstant(value), nil
		default:
			return nil, bsoncore.ErrElementNotFound
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
		return tn, nil
	case *ast.Document:
		newElements := make([]*ast.DocumentElement, 0, len(tn.Elements))
		for _, e := range tn.Elements {
			en, err := v.eval(e.Expr, value)
			// If the bsontype of value is MinKey, it means this code has been
			// executed from ConstantPropagation, which requires returning an
			// err in the presence of a field reference.
			if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
				continue
			} else if err != nil {
				return nil, err
			}
			newElements = append(newElements, ast.NewDocumentElement(e.Name, en))
		}
		return ast.NewDocument(newElements...), nil
	case *ast.Array:
		newElements := make([]ast.Expr, len(tn.Elements))
		for i, e := range tn.Elements {
			en, err := v.eval(e, value)
			// If the bsontype of value is MinKey, it means this code has been
			// executed from ConstantPropagation, which requires returning an
			// err in the presence of a field reference.
			if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
				en = ast.NewConstant(bsonutil.Null())
			} else if err != nil {
				return nil, err
			}
			newElements[i] = en
		}
		return ast.NewArray(newElements...), nil
	case *ast.FieldRef:
		if tn.Parent != nil {

			pn, err := v.eval(tn.Parent, value)
			if err != nil {
				return nil, err
			}

			pc, ok := pn.(*ast.Constant)
			if !ok {
				if pn != tn.Parent {
					return ast.NewFieldRef(tn.Name, pn), nil
				}
				return n, nil
			}
			value = pc.Value
		}

		switch value.Type {
		case bsontype.EmbeddedDocument:
			value, err := value.Document().LookupErr(tn.Name)
			if err != nil {
				return nil, err
			}
			return ast.NewConstant(value), nil
		default:
			return nil, bsoncore.ErrElementNotFound
		}
	case *ast.Function:
		return v.evalFunction(tn, value)
	case *ast.Let:
		return v.evalLet(tn, value)
	case *ast.Exists:
		_, err := v.eval(tn.FieldRef, value)

		if err != nil && err != bsoncore.ErrElementNotFound {
			return nil, err
		}

		if tn.Exists == (err == nil) {
			return ast.NewConstant(bsonutil.True), nil
		}

		return ast.NewConstant(bsonutil.False), nil
	default:
		return n, nil
	}
}

func (v exprEvaluator) evalUnary(n *ast.Unary, value bsoncore.Value) (ast.Expr, error) {
	expr, err := v.eval(n.Expr, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		switch n.Op {
		case ast.Not:
			return ast.NewConstant(bsonutil.True), nil
		default:
			return ast.NewConstant(bsonutil.Null()), nil
		}
	} else if err != nil {
		return nil, err
	}

	exprc, ok := expr.(*ast.Constant)
	if !ok {
		if expr != n.Expr {
			return ast.NewUnary(n.Op, expr), nil
		}
		return n, nil
	}

	switch n.Op {
	case ast.Abs, ast.Ceil, ast.Floor, ast.Exp, ast.Ln, ast.Log10, ast.Sqrt:
		result, err := unaryFunctions[n.Op](exprc.Value)
		if err != nil {
			return nil, err
		}
		return ast.NewConstant(result), nil
	case ast.Not:
		if bsonutil.CoerceToBoolean(exprc.Value) {
			return ast.NewConstant(bsonutil.False), nil
		}
		return ast.NewConstant(bsonutil.True), nil
	default:
		if expr != n.Expr {
			return ast.NewUnary(n.Op, expr), nil
		}
		return n, nil
	}
}

func (v exprEvaluator) evalBinary(n *ast.Binary, value bsoncore.Value) (ast.Expr, error) {
	left, err := v.eval(n.Left, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.evalBinaryMissingField(n.Op, n.Right, value, false)
	} else if err != nil {
		return nil, err
	}

	right, err := v.eval(n.Right, value)
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.evalBinaryMissingField(n.Op, n.Left, value, true)
	} else if err != nil {
		return nil, err
	}

	lc, err1 := extractConstant(left)
	rc, err2 := extractConstant(right)
	if err1 != nil || err2 != nil {
		if left != n.Left || right != n.Right {
			return ast.NewBinary(n.Op, left, right), nil
		}
		return n, nil
	}

	switch n.Op {
	case ast.And:
		if !bsonutil.CoerceToBoolean(lc) || !bsonutil.CoerceToBoolean(rc) {
			return ast.NewConstant(bsonutil.False), nil
		}

		return ast.NewConstant(bsonutil.True), nil

	case ast.Compare:
		if cmp, err := bsoncompare.Compare(lc, rc); err != nil {
			return nil, err
		} else {
			return ast.NewConstant(bsonutil.Int32(int32(cmp))), nil
		}

	case ast.Equals, ast.GreaterThan, ast.GreaterThanOrEquals, ast.LessThan, ast.LessThanOrEquals, ast.NotEquals:
		if result, err := v.compare(n.Op, lc, rc); err != nil {
			return nil, err
		} else {
			return ast.NewConstant(bsonutil.Boolean(result)), nil
		}

	case ast.Or:
		if bsonutil.CoerceToBoolean(lc) || bsonutil.CoerceToBoolean(rc) {
			return ast.NewConstant(bsonutil.True), nil
		}

		return ast.NewConstant(bsonutil.False), nil
	case ast.Nor:
		if bsonutil.CoerceToBoolean(lc) || bsonutil.CoerceToBoolean(rc) {
			return ast.NewConstant(bsonutil.False), nil
		}

		return ast.NewConstant(bsonutil.True), nil

	case ast.Add, ast.Subtract, ast.Multiply, ast.Divide, ast.Mod, ast.Log, ast.Pow:
		result, err := binaryFunctions[n.Op](lc, rc)
		if err != nil {
			return nil, err
		}
		return ast.NewConstant(result), nil
	default:
		if left != n.Left || right != n.Right {
			return ast.NewBinary(n.Op, left, right), nil
		}
		return n, nil
	}
}

func (v exprEvaluator) evalTrunc(n *ast.Trunc, value bsoncore.Value) (ast.Expr, error) {
	number, err := v.eval(n.Number, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return ast.NewConstant(bsonutil.Null()), nil
	} else if err != nil {
		return nil, err
	}
	precision, err := v.eval(n.Precision, value)
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return ast.NewConstant(bsonutil.Null()), nil
	} else if err != nil {
		return nil, err
	}

	nc, ok1 := number.(*ast.Constant)
	pc, ok2 := precision.(*ast.Constant)

	if !ok1 || !ok2 {
		if number != n.Number || precision != n.Precision {
			return ast.NewTrunc(number, precision), nil
		}
		return n, nil
	}

	result, err := bsonutil.Trunc(nc.Value, pc.Value)
	if err != nil {
		return nil, err
	}

	return ast.NewConstant(result), nil
}

func (v exprEvaluator) compare(op ast.BinaryOp, left bsoncore.Value, right bsoncore.Value) (bool, error) {
	if !v.isMatchExpr {
		return v.compareScalar(op, left, right)
	}

	switch left.Type {
	case bsontype.Array:
		lvalues, _ := left.Array().Values()
		for _, lval := range lvalues {
			if result, err := v.compareScalar(op, lval, right); err != nil {
				return false, err
			} else if result {
				return true, nil
			}
		}

		return false, nil
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

func (v exprEvaluator) evalBinaryMissingField(op ast.BinaryOp, other ast.Expr, value bsoncore.Value, flip bool) (ast.Expr, error) {
	if flip {
		op = op.Flip()
	}

	otherValue, err := v.eval(other, value)
	// If the bsontype of value is MinKey, it means this code has been
	// executed from ConstantPropagation, which requires returning an
	// err in the presence of a field reference.
	// The bsontype.MinKey check is not strictly needed here because this function
	// should never be called in such a case. This check is maintained for future
	// compatability and completeness.
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.evalBinaryMissingBothFields(op)
	} else if err != nil {
		return nil, err
	}

	if v.isMatchExpr {
		if otherConst, ok := otherValue.(*ast.Constant); ok {
			switch op {
			case ast.Equals, ast.LessThanOrEquals, ast.GreaterThanOrEquals:
				if otherConst.Value.Type == bsontype.Null {
					return ast.NewConstant(bsonutil.True), nil
				}
				return ast.NewConstant(bsonutil.False), nil
			case ast.LessThan, ast.GreaterThan:
				return ast.NewConstant(bsonutil.False), nil
			case ast.NotEquals:
				if otherConst.Value.Type == bsontype.Null {
					return ast.NewConstant(bsonutil.False), nil
				}
				return ast.NewConstant(bsonutil.True), nil
			}
		}

		panic("unexpected match expressions")
	}

	switch op {
	case ast.Compare:
		if flip {
			return ast.NewConstant(bsonutil.Int32(1)), nil
		}
		return ast.NewConstant(bsonutil.Int32(-1)), nil
	case ast.LessThan, ast.LessThanOrEquals, ast.NotEquals:
		return ast.NewConstant(bsonutil.True), nil
	case ast.Or:
		if otherConst, ok := otherValue.(*ast.Constant); ok {
			return ast.NewConstant(
				bsonutil.Boolean(
					bsonutil.CoerceToBoolean(otherConst.Value),
				),
			), nil
		}
		return ast.NewBinary(
			ast.Or,
			otherValue,
			ast.NewConstant(bsonutil.False),
		), nil
	case ast.Divide, ast.Log, ast.Mod, ast.Pow, ast.Subtract, ast.Add, ast.Multiply:
		return ast.NewConstant(bsonutil.Null()), nil
	default:
		return ast.NewConstant(bsonutil.False), nil
	}
}

func (v exprEvaluator) evalBinaryMissingBothFields(op ast.BinaryOp) (ast.Expr, error) {
	switch op {
	case ast.Compare:
		return ast.NewConstant(bsonutil.Int32(0)), nil
	case ast.Equals, ast.LessThanOrEquals, ast.GreaterThanOrEquals:
		return ast.NewConstant(bsonutil.True), nil
	case ast.Divide, ast.Log, ast.Mod, ast.Pow, ast.Subtract, ast.Add, ast.Multiply:
		return ast.NewConstant(bsonutil.Null()), nil
	default:
		return ast.NewConstant(bsonutil.False), nil
	}
}

func (v exprEvaluator) evalConditional(n *ast.Conditional, value bsoncore.Value) (ast.Expr, error) {
	ifExpr, err := v.eval(n.If, value)
	if err == bsoncore.ErrElementNotFound && value.Type != bsontype.MinKey {
		return v.eval(n.Else, value)
	} else if err != nil {
		return nil, err
	}
	switch te := ifExpr.(type) {
	case *ast.Constant:
		if cmp, err := bsoncompare.Compare(te.Value, bsonutil.True); err != nil {
			return nil, err
		} else if cmp == 0 {
			// Failure to eval a branch of a cond is not a failure to simplify,
			// we can still simplify here by removing the cond because
			// the conditional was constant!
			evaled, err := v.eval(n.Then, value)
			if err != nil {
				return n.Then, nil
			}
			return evaled, nil
		}
		evaled, err := v.eval(n.Else, value)
		if err != nil {
			return n.Else, nil
		}
		return evaled, nil
	default:
		thenExpr, err := v.eval(n.Then, value)
		if err != nil {
			return nil, err
		}
		elseExpr, err := v.eval(n.Else, value)
		if err != nil {
			return nil, err
		}
		if ifExpr != n.If || thenExpr != n.Then || elseExpr != n.Else {
			return ast.NewConditional(ifExpr, thenExpr, elseExpr), nil
		}
		return n, nil
	}
}

func (v exprEvaluator) evalLet(n *ast.Let, value bsoncore.Value) (ast.Expr, error) {
	variables := make(map[string]ast.Expr)
	for _, letVar := range n.Variables {
		letVarExpr, err := v.eval(letVar.Expr, value)
		if err != nil {
			return nil, err
		}
		variables[letVar.Name] = letVarExpr
	}
	return v.eval(
		SubstituteVariables(n.Expr, variables).(ast.Expr),
		value,
	)
}
