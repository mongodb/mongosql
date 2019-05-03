package eval

import (
	"strconv"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

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
		index, ok := bsonutil.AsInt32OK(in.(*ast.Constant).Value)
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
	case *ast.Binary:
		return v.evalBinary(tn, value)
	case *ast.Conditional:
		return v.evalConditional(tn, value)
	case *ast.Constant:
		return tn, nil
	case *ast.Document:
		newElements := make([]*ast.DocumentElement, 0, len(tn.Elements))
		for _, e := range tn.Elements {
			en, err := v.eval(e.Expr, value)
			if err == bsoncore.ErrElementNotFound {
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
			if err == bsoncore.ErrElementNotFound {
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
	default:
		return n, nil
	}
}

func (v exprEvaluator) evalBinary(n *ast.Binary, value bsoncore.Value) (ast.Expr, error) {
	left, err := v.eval(n.Left, value)
	if err == bsoncore.ErrElementNotFound {
		return v.evalBinaryMissingField(n.Op, n.Right, value)
	} else if err != nil {
		return nil, err
	}
	right, err := v.eval(n.Right, value)
	if err == bsoncore.ErrElementNotFound {
		return v.evalBinaryMissingField(n.Op.Flip(), n.Left, value)
	} else if err != nil {
		return nil, err
	}

	lc, ok1 := left.(*ast.Constant)
	rc, ok2 := right.(*ast.Constant)
	if !ok1 || !ok2 {
		if left != n.Left || right != n.Right {
			return ast.NewBinary(n.Op, left, right), nil
		}
		return n, nil
	}

	switch n.Op {
	case ast.And:
		if !bsonutil.CoerceToBoolean(lc.Value) || !bsonutil.CoerceToBoolean(rc.Value) {
			return ast.NewConstant(bsonutil.False), nil
		}

		return ast.NewConstant(bsonutil.True), nil

	case ast.Compare:
		if cmp, err := bsonutil.Compare(lc.Value, rc.Value); err != nil {
			return nil, err
		} else {
			return ast.NewConstant(bsonutil.Int32(int32(cmp))), nil
		}

	case ast.Equals:
		if cmp, err := bsonutil.Compare(lc.Value, rc.Value); err != nil {
			return nil, err
		} else if cmp == 0 {
			return ast.NewConstant(bsonutil.True), nil
		} else {
			return ast.NewConstant(bsonutil.False), nil
		}

	case ast.GreaterThan:
		if cmp, err := bsonutil.Compare(lc.Value, rc.Value); err != nil {
			return nil, err
		} else if cmp > 0 {
			return ast.NewConstant(bsonutil.True), nil
		} else {
			return ast.NewConstant(bsonutil.False), nil
		}

	case ast.GreaterThanOrEquals:
		if cmp, err := bsonutil.Compare(lc.Value, rc.Value); err != nil {
			return nil, err
		} else if cmp >= 0 {
			return ast.NewConstant(bsonutil.True), nil
		} else {
			return ast.NewConstant(bsonutil.False), nil
		}

	case ast.LessThan:
		if cmp, err := bsonutil.Compare(lc.Value, rc.Value); err != nil {
			return nil, err
		} else if cmp < 0 {
			return ast.NewConstant(bsonutil.True), nil
		} else {
			return ast.NewConstant(bsonutil.False), nil
		}

	case ast.LessThanOrEquals:
		if cmp, err := bsonutil.Compare(lc.Value, rc.Value); err != nil {
			return nil, err
		} else if cmp <= 0 {
			return ast.NewConstant(bsonutil.True), nil
		} else {
			return ast.NewConstant(bsonutil.False), nil
		}

	case ast.NotEquals:
		if cmp, err := bsonutil.Compare(lc.Value, rc.Value); err != nil {
			return nil, err
		} else if cmp != 0 {
			return ast.NewConstant(bsonutil.True), nil
		} else {
			return ast.NewConstant(bsonutil.False), nil
		}

	case ast.Or:
		if bsonutil.CoerceToBoolean(lc.Value) || bsonutil.CoerceToBoolean(rc.Value) {
			return ast.NewConstant(bsonutil.True), nil
		}

		return ast.NewConstant(bsonutil.False), nil

	default:
		if left != n.Left || right != n.Right {
			return ast.NewBinary(n.Op, left, right), nil
		}
		return n, nil
	}
}

func (v exprEvaluator) evalBinaryMissingField(op ast.BinaryOp, other ast.Expr, value bsoncore.Value) (ast.Expr, error) {
	otherValue, err := v.eval(other, value)
	if err == bsoncore.ErrElementNotFound {
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
	default:
		return ast.NewConstant(bsonutil.False), nil
	}
}

func (v exprEvaluator) evalBinaryMissingBothFields(op ast.BinaryOp) (ast.Expr, error) {
	switch op {
	case ast.Equals, ast.LessThanOrEquals, ast.GreaterThanOrEquals:
		return ast.NewConstant(bsonutil.True), nil
	default:
		return ast.NewConstant(bsonutil.False), nil
	}
}

func (v exprEvaluator) evalConditional(n *ast.Conditional, value bsoncore.Value) (ast.Expr, error) {
	ifExpr, err := v.eval(n.If, value)
	if err == bsoncore.ErrElementNotFound {
		return v.eval(n.Else, value)
	} else if err != nil {
		return nil, err
	}
	switch te := ifExpr.(type) {
	case *ast.Constant:
		if cmp, err := bsonutil.Compare(te.Value, bsonutil.True); err != nil {
			return nil, err
		} else if cmp == 0 {
			return v.eval(n.Then, value)
		}
		return v.eval(n.Else, value)
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
