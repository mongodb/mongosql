package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/option"
)

// DesugarQuery is a compiler phase that occurs after parsing and before
// algebrization. This phase converts a CST from its input form to an equivalent
// simpler from. Constructs that exist in the input can be wholly removed in the
// output. Operations in this phase should be simple. CSTs leave the deeper
// structure of the query obfuscated and no attempt to uncover it should be made.
func DesugarQuery(statement Statement) (Statement, error) {
	type desugarPass struct {
		pass Walker
		// prePassDebuggingMessage will be printed before the pass, if it is not NoneString.
		prePassDebuggingMessage option.String
		// prePassDebuggingMessage will be printed after the pass, if it is not NoneString.
		postPassDebuggingMessage option.String
	}

	desugarers := []desugarPass{
		{&isNotDesugarer{}, option.NoneString(), option.NoneString()},
		{&unwrapSingleTuples{}, option.NoneString(), option.NoneString()},
		{&someToAnyDesugarer{}, option.NoneString(), option.NoneString()},
		{&betweenDesugarer{}, option.NoneString(), option.NoneString()},
		{&ifToCaseDesugarer{}, option.NoneString(), option.NoneString()},
		{&inSubqueryDesugarer{}, option.NoneString(), option.NoneString()},
		{&inListConverter{}, option.NoneString(), option.NoneString()},
		{&subqueryComparisonConverter{}, option.NoneString(), option.NoneString()},
		{&tupleComparisonDesugarer{}, option.NoneString(), option.NoneString()},
		{&makeDualExplicit{}, option.NoneString(), option.NoneString()},
	}

	result := statement.(CST)
	var err error
	for _, pass := range desugarers {
		if pass.prePassDebuggingMessage != option.NoneString() {
			fmt.Printf(pass.prePassDebuggingMessage.Unwrap(), result)
		}
		result, err = Walk(pass.pass, result)
		if err != nil {
			return nil, err
		}
		if pass.postPassDebuggingMessage != option.NoneString() {
			fmt.Printf(pass.postPassDebuggingMessage.Unwrap(), result)
		}
	}

	return result.(Statement), nil
}

var _ Walker = (*isNotDesugarer)(nil)

// isNotDesugarer replaces `x IS NOT y` with `NOT(x IS y)`
type isNotDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*isNotDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*isNotDesugarer) PostVisit(current CST) (CST, error) {
	cmp, ok := current.(*ComparisonExpr)
	if !ok {
		return current, nil
	}

	switch cmp.Operator {
	case AST_IS_NOT:
		return &NotExpr{
			Expr: &ComparisonExpr{
				Left:     cmp.Left,
				Right:    cmp.Right,
				Operator: AST_IS,
			},
		}, nil
	default:
		return current, nil
	}
}

// makeDualExplicit sets the name of the From field in a Select statement
// to "DUAL" when the user does not explicitly name the DUAL table.
type makeDualExplicit struct{}

// PreVisit is called for every node before its children are walked.
func (*makeDualExplicit) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*makeDualExplicit) PostVisit(current CST) (CST, error) {
	if node, isSelect := current.(*Select); isSelect {
		if node.From == nil {
			node.From = TableExprs{&DualTableExpr{}}
		}
	}
	return current, nil
}

var _ Walker = (*makeDualExplicit)(nil)

// betweenDesugarer replaces BETWEEN (and NOT BETWEEN) with comparisons.
type betweenDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*betweenDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*betweenDesugarer) PostVisit(current CST) (CST, error) {
	if node, isRangeCond := current.(*RangeCond); isRangeCond {
		switch node.Operator {
		case AST_BETWEEN:
			return &AndExpr{
				Left: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_GE,
					Right:    node.From,
				},
				Right: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_LE,
					Right:    node.To,
				},
			}, nil
		case AST_NOT_BETWEEN:
			return &OrExpr{
				Left: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_LT,
					Right:    node.From,
				},
				Right: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_GT,
					Right:    node.To,
				},
			}, nil
		}
	}
	return current, nil
}

var _ Walker = (*betweenDesugarer)(nil)

// ifToCaseDesugarer replaces IF scalar functions with CaseExprs.
type ifToCaseDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*ifToCaseDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

func desugarCoalesce(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 1 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs))
	for i, expr := range node.Exprs {
		caseConditions[i] = &When{
			Cond: &NotExpr{
				Expr: &ComparisonExpr{
					Operator: AST_IS,
					Left:     expr.(*NonStarExpr).Expr,
					Right:    &NullVal{},
				},
			},
			Val: expr.(*NonStarExpr).Expr,
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  nil,
	}, nil
}

func desugarElt(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 2 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs)-1)
	for i, expr := range node.Exprs[1:] {
		caseConditions[i] = &When{
			Cond: &ComparisonExpr{
				Operator: AST_EQ,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    NumVal(strconv.Itoa(i + 1)),
			},
			Val: expr.(*NonStarExpr).Expr,
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  nil,
	}, nil
}

func desugarField(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 2 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs)-1)
	for i, expr := range node.Exprs[1:] {
		caseConditions[i] = &When{
			Cond: &ComparisonExpr{
				Operator: AST_EQ,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    expr.(*NonStarExpr).Expr,
			},
			Val: NumVal(strconv.Itoa(i + 1)),
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  NumVal("0"),
	}, nil
}

func desugarIf(node *FuncExpr) (CST, error) {
	if len(node.Exprs) != 3 {
		return node, nil
	}
	return &CaseExpr{
		Expr: nil,
		Whens: []*When{
			{Cond: node.Exprs[0].(*NonStarExpr).Expr,
				Val: node.Exprs[1].(*NonStarExpr).Expr,
			},
		},
		Else: node.Exprs[2].(*NonStarExpr).Expr,
	}, nil
}

func desugarIfNull(node *FuncExpr) (CST, error) {
	if len(node.Exprs) != 2 {
		return node, nil
	}
	return &CaseExpr{
		Expr: nil,
		Whens: []*When{
			{Cond: &ComparisonExpr{
				Operator: AST_IS,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    &NullVal{},
			},
				Val: node.Exprs[1].(*NonStarExpr).Expr,
			},
		},
		Else: node.Exprs[0].(*NonStarExpr).Expr,
	}, nil
}

func desugarInterval(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 2 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs))
	caseConditions[0] = &When{
		Cond: &ComparisonExpr{
			Operator: AST_IS,
			Left:     node.Exprs[0].(*NonStarExpr).Expr,
			Right:    &NullVal{},
		},
		Val: NumVal("-1"),
	}
	for i, expr := range node.Exprs[1:] {
		caseConditions[i+1] = &When{
			Cond: &ComparisonExpr{
				Operator: AST_LT,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    expr.(*NonStarExpr).Expr,
			},
			Val: NumVal(strconv.Itoa(i)),
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  NumVal(strconv.Itoa(len(caseConditions) - 1)),
	}, nil
}

func desugarNullIf(node *FuncExpr) (CST, error) {
	if len(node.Exprs) != 2 {
		return node, nil
	}
	return &CaseExpr{
		Expr: nil,
		Whens: []*When{
			{Cond: &ComparisonExpr{
				Operator: AST_EQ,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    node.Exprs[1].(*NonStarExpr).Expr,
			},
				Val: &NullVal{},
			},
		},
		Else: node.Exprs[0].(*NonStarExpr).Expr,
	}, nil
}

// PostVisit is called for every node after its children are walked.
func (*ifToCaseDesugarer) PostVisit(current CST) (CST, error) {
	if node, isFunc := current.(*FuncExpr); isFunc {
		switch strings.ToLower(node.Name) {
		case "coalesce":
			return desugarCoalesce(node)
		case "elt":
			return desugarElt(node)
		case "field":
			return desugarField(node)
		case "if":
			return desugarIf(node)
		case "interval":
			return desugarInterval(node)
		case "ifnull":
			return desugarIfNull(node)
		case "nullif":
			return desugarNullIf(node)
		}
	}
	return current, nil
}

var _ Walker = (*ifToCaseDesugarer)(nil)

// unwrapSingleTuples is a desugarer that removes single-element tuples
// generated by the parser. Desugarers orchestrates a desugaring phase on the
// CST by implementing the Walker interface.
type unwrapSingleTuples struct{}

// PreVisit is called for every node before its children are walked.
func (*unwrapSingleTuples) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*unwrapSingleTuples) PostVisit(current CST) (CST, error) {
	return detupleWrappedExpr(current), nil
}

var _ Walker = (*unwrapSingleTuples)(nil)

// detupleWrappedExpr removes tuples that were placed around expressions in the
// parser where parentheses existed.
//
// Note: This should not be necessary. However, our parser interprets every set
// of parentheses (even those in arithmetic exprs, for example) as a tuple, so
// we need to get rid of them.
func detupleWrappedExpr(node CST) CST {
	if tuple, isTuple := node.(ValTuple); isTuple && len(tuple) == 1 {
		return tuple[0]
	}
	return node
}

// someToAnyDesugarer replaces SOME with ANY, as they are aliases.
type someToAnyDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*someToAnyDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*someToAnyDesugarer) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		if node.SubqueryOperator == AST_SOME {
			return &ComparisonExpr{
				Operator:         node.Operator,
				Left:             node.Left,
				Right:            node.Right,
				SubqueryOperator: AST_ANY,
			}, nil
		}
	}
	return current, nil
}

var _ Walker = (*someToAnyDesugarer)(nil)

// inSubqueryDesugarer replaces IN (subquery) with = ANY (subquery)
// and NOT IN (subquery) with <> ALL (subquery).
type inSubqueryDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*inSubqueryDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*inSubqueryDesugarer) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		switch node.SubqueryOperator {
		case AST_NOT_IN:
			return &ComparisonExpr{
				AST_NE,
				node.Left,
				node.Right,
				AST_ALL,
			}, nil
		case AST_IN:
			return &ComparisonExpr{
				AST_EQ,
				node.Left,
				node.Right,
				AST_ANY,
			}, nil
		}
	}
	return current, nil
}

var _ Walker = (*inSubqueryDesugarer)(nil)

// inListConverter is a desugarer that breakes IN lists into boolean
// comparisons.
// Desugarers orchestrates a desugaring phase on the CST by implementing
// the Walker interface.
type inListConverter struct{}

// PreVisit is called for every node before its children are walked. PreVisit
// desugars NOT IN nodes to IN nodes, which will themselves be desugared further
// in the PostVisit function.
func (*inListConverter) PreVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		if node.Operator == AST_NOT_IN {
			// ignore NOT IN subquery, that is a different expression
			if _, isSub := node.Right.(*Subquery); !isSub {
				current = breakUpNotIn(node)
			}
		}
	}
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*inListConverter) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		if node.Operator == AST_IN {
			// ignore IN subquery, that is a different expression
			if _, isSub := node.Right.(*Subquery); !isSub {
				if tuple, isTuple := node.Right.(ValTuple); isTuple {
					current = inListToDisjunction(node.Left, tuple)
				} else {
					current = inListToDisjunction(node.Left, []Expr{node.Right})
				}
			}
		}
	}
	return current, nil
}

var _ Walker = (*inListConverter)(nil)

// breakUpNotIn rewrites a NOT IN list expression as a boolean NOT of an IN list
// expression. NOT IN list is of the form a NOT IN (b, c...) and is expressable
// as NOT (a IN (b, c...))
func breakUpNotIn(node *ComparisonExpr) Expr {
	node.Operator = AST_IN
	return &NotExpr{node}
}

// detupleEquality rewrites an IN list expression as a disjunction by enumerating every
// equality comparison.
// IN list is of the form a in (b, c...) and is expressable as
// a = b OR a = c OR ...
func inListToDisjunction(leftExpr Expr, rightExprs ValTuple) Expr {
	var makeDisjunction func(leftExpr Expr, rightExprs ValTuple) Expr
	makeDisjunction = func(leftExpr Expr, rightExprs ValTuple) Expr {
		if len(rightExprs) == 1 {
			return &ComparisonExpr{
				AST_EQ,
				leftExpr,
				rightExprs[0],
				"",
			}
		}
		return &OrExpr{
			&ComparisonExpr{
				AST_EQ,
				leftExpr,
				rightExprs[0],
				"",
			},
			makeDisjunction(leftExpr.Copy().(Expr), rightExprs[1:]),
		}
	}
	return makeDisjunction(leftExpr, rightExprs)
}

// tupleComparisonDesugarer is a desugarer that rewrites multi-element tuples into other
// expressions. For example, the tupleComparisonDesugarer will rewrite `select (a, b) <
// (c, d) from foo` as `select a < c or a = c and b < d from foo`.
type tupleComparisonDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*tupleComparisonDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*tupleComparisonDesugarer) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		_, leftIsTuple := node.Left.(ValTuple)
		_, rightIsTuple := node.Right.(ValTuple)

		if leftIsTuple || rightIsTuple {
			var err error
			current, err = detupleToBooleanCompare(node)
			if err != nil {
				return nil, err
			}
		}
	}
	return current, nil
}

var _ Walker = (*tupleComparisonDesugarer)(nil)

// Tuples are only legal in comparisons.
// Tuple comparisons are either equivalent to conjunctions or disjunctions
// of comparisons or subquery comparisons.
// This function handles the first case.
func detupleToBooleanCompare(node *ComparisonExpr) (Expr, error) {
	// Check outermost left side to ensure it is a tuple.
	left, leftIsTuple := node.Left.(ValTuple)
	if !leftIsTuple {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
	}

	// Ensure both sides are uniform. And extract values from nests.
	var checkLengthAndExtract func(left ValTuple, rightUnchecked Expr) (Exprs, Exprs, error)
	checkLengthAndExtract = func(left ValTuple, rightUnchecked Expr) (Exprs, Exprs, error) {
		var leftExprs Exprs
		var rightExprs Exprs

		// Make sure the right hand side is a tuple first.
		right, rightIsTuple := rightUnchecked.(ValTuple)
		if !rightIsTuple {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(left))
		}

		// Make sure the lengths are the same.
		if len(left) != len(right) {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(left))
		}

		// Collect values in turn from each tuple.
		for i, leftExpr := range left {
			rightExpr := right[i]

			// Recursively descend into left side if a tuple is located.
			leftElem, innerLeftIsTuple := leftExpr.(ValTuple)
			if innerLeftIsTuple {
				leftAdditions, rightAdditions, err := checkLengthAndExtract(leftElem, rightExpr)
				if err != nil {
					return nil, nil, err
				}
				leftExprs = append(leftExprs, leftAdditions...)
				rightExprs = append(rightExprs, rightAdditions...)
			} else {
				// If the right hand side is a tuple here, it doesn't match the expression
				// on the left.
				if _, innerRightIsTuple := rightExpr.(ValTuple); innerRightIsTuple {
					return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
				}
				leftExprs = append(leftExprs, leftExpr)
				rightExprs = append(rightExprs, rightExpr)
			}
		}

		return leftExprs, rightExprs, nil
	}

	leftExprs, rightExprs, err := checkLengthAndExtract(left, node.Right)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case string(AST_EQ), string(AST_NSE):
		return detupleEquality(node.Operator, leftExprs, rightExprs), nil
	case string(AST_NE):
		eq := detupleEquality(string(AST_EQ), leftExprs, rightExprs)
		return &NotExpr{eq}, nil
	default:
		return detupleInequality(node.Operator, leftExprs, rightExprs), nil
	}
}

// detupleEquality rewrites a tuple as a conjunction by moving from the former to the later:
// Tuple equality expressions of the form (a1, b1, ...) op (a2, b2, ...) are expressable as
// a1 op a2 AND b1 op b2 AND ...
func detupleEquality(operator string, leftExprs, rightExprs Exprs) Expr {
	var makeConjunction func(operator string, leftExprs, rightExprs Exprs) Expr
	makeConjunction = func(operator string, leftExprs, rightExprs Exprs) Expr {
		if len(leftExprs) == 1 {
			return &ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			}
		}
		return &AndExpr{
			&ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			},
			makeConjunction(operator, leftExprs[1:], rightExprs[1:]),
		}
	}
	return makeConjunction(operator, leftExprs, rightExprs)
}

// detupleInequality rewrites a tuple as a disjunction by moving from the former to the later:
// Tuple inequality expressions of the form (a1, b1, ...) op (a2, b2, ...) are expressable as
// a1 op a2 OR (a1 = a2 AND b1 op b2 OR (...))
func detupleInequality(operator string, leftExprs, rightExprs Exprs) Expr {
	var makeDisjunction func(operator string, leftExprs, rightExprs Exprs) Expr
	makeDisjunction = func(operator string, leftExprs, rightExprs Exprs) Expr {
		if len(leftExprs) == 1 {
			return &ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			}
		}
		return &OrExpr{
			&ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			},
			&AndExpr{
				&ComparisonExpr{
					AST_EQ,
					leftExprs[0],
					rightExprs[0],
					"",
				},
				makeDisjunction(operator, leftExprs[1:], rightExprs[1:]),
			},
		}
	}

	switch operator {
	case string(AST_LE):
		operator = string(AST_LT)
	case string(AST_GE):
		operator = string(AST_GT)
	}

	return makeDisjunction(operator, leftExprs, rightExprs)
}

// detupleToSubquery rewrites a tuple to a subquery. Tuples are only legal in
// comparisons. Tuple comparisons are either equivalent to conjunctions or
// disjunctions of comparisons or subquery comparisons. This function handles
// the second case.
func detupleToSubquery(node ValTuple) *Subquery {
	selExprs := make(SelectExprs, len(node))
	for i, expr := range node {
		selExprs[i] = &NonStarExpr{Expr: expr}
	}
	return &Subquery{
		&Select{SelectExprs: selExprs, QueryGlobals: &QueryGlobals{}},
		false,
	}
}

// toTuple returns the Expr as a tuple. If it already is one, it is returned
// unchanged; if it is not, it is returned as a single-value tuple.
func toTuple(e Expr) ValTuple {
	if tuple, isTuple := e.(ValTuple); isTuple {
		return tuple
	}
	return ValTuple{e}
}

// subqueryComparisonConverter is a desugarer that rewrites comparison
// expressions that have a subquery on one side and not the other to
// have subqueries on both sides.
type subqueryComparisonConverter struct{}

var _ Walker = (*subqueryComparisonConverter)(nil)

// PreVisit is called for every node before its children are walked.
func (*subqueryComparisonConverter) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*subqueryComparisonConverter) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		_, leftIsSubquery := node.Left.(*Subquery)
		_, rightIsSubquery := node.Right.(*Subquery)

		// Even non-tuple comparison operands are converted into subqueries,
		// as this allows us to always produce full subquery comparisons
		// where the left and right operands are both subqueries.
		if !leftIsSubquery && rightIsSubquery {
			node.Left = detupleToSubquery(toTuple(node.Left))
		} else if !rightIsSubquery && leftIsSubquery {
			node.Right = detupleToSubquery(toTuple(node.Right))
		}
	}
	return current, nil
}
