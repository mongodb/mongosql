package optimizer

import (
	"fmt"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/astprint"
)

// PartialRedundancyElimination removes redundant expressions by binding them in $let expressions.
func PartialRedundancyElimination(pipeline *ast.Pipeline) *ast.Pipeline {
	out, _ := ast.Visit(pipeline, func(v ast.Visitor, n ast.Node) ast.Node {
		switch typedN := n.(type) {
		case ast.Expr:
			return removeRedundancies(typedN)
		default:
			return n.Walk(v)
		}
	})
	return out.(*ast.Pipeline)
}

func removeRedundancies(e ast.Expr) ast.Expr {
	redundantExpressionCounts := make(map[string]uint32)
	addExpression := func(e ast.Expr) {
		exprStr := astprint.String(e)
		if exprCount, ok := redundantExpressionCounts[exprStr]; ok {
			redundantExpressionCounts[exprStr] = exprCount + 1
		} else {
			redundantExpressionCounts[exprStr] = 1
		}
	}
	// First count the redundant expressions.
	_, _ = ast.Visit(e, func(v ast.Visitor, n ast.Node) ast.Node {
		// There are a number of situations where we want to recursively call removeRedundancies
		//
		// If this is a AggExpr, we do not want to put $lets above it, as that would not work.
		// Once we have separate Match and Aggregation Expression trees, this check can go away.
		//
		// If this is an aggregate function:
		//	 $addToSet
		//	 $avg
		//	 $first
		//	 $last
		//	 $max
		//	 $mergeObjects
		//	 $min
		//	 $push
		//	 $stdDevPop
		//	 $stdDevSamp
		//	 $sum
		// We do not want to put $lets above it, as that would not work.
		//
		// If this is a variable definer, we need to recursively run the analysis on its bounded
		// expression, otherwise we can pull redundant variable uses out of scope of their
		// definitions. Technically, we could check for uses of the variable defined in the
		// potentially redundant expressions, but I feel that complicates the code significantly
		// for little benefit: generally any complex expressions in a variable defining function
		// are going to use that variable. This might be less true of $let, itself.
		//
		// If this has lazy argument semantics, we also cannot pull out redundant expressions
		// because that would force the expression to have strict semantics.
		switch typedExpr := n.(type) {
		case *ast.AggExpr:
			typedExpr.Expr = removeRedundancies(typedExpr.Expr)
			// There is no reason to add this as a possibly redundant expression.
			return typedExpr
		case *ast.Binary:
			if analyzer.HasLazyArgumentSemantics(typedExpr) {
				typedExpr.Left = removeRedundancies(typedExpr.Left)
				typedExpr.Right = removeRedundancies(typedExpr.Right)
				addExpression(typedExpr)
				return typedExpr
			}
		case *ast.Conditional:
			typedExpr.If = removeRedundancies(typedExpr.If)
			typedExpr.Then = removeRedundancies(typedExpr.Then)
			typedExpr.Else = removeRedundancies(typedExpr.Else)
			addExpression(typedExpr)
			return typedExpr
		case *ast.Map:
			typedExpr.Input = removeRedundancies(typedExpr.Input)
			typedExpr.In = removeRedundancies(typedExpr.In)
			addExpression(typedExpr)
			return typedExpr
		case *ast.Filter:
			typedExpr.Input = removeRedundancies(typedExpr.Input)
			typedExpr.Cond = removeRedundancies(typedExpr.Cond)
			addExpression(typedExpr)
			return typedExpr
		case *ast.Let:
			typedExpr.Expr = removeRedundancies(typedExpr.Expr)
			addExpression(typedExpr)
			return typedExpr
		case *ast.Function:
			if analyzer.IsVariableDefiningFunction(typedExpr) ||
				analyzer.HasLazyArgumentSemantics(typedExpr) ||
				analyzer.IsAccumulatorFunction(typedExpr) {
				switch typedArg := typedExpr.Arg.(type) {
				// If the argument is an array or document, we have
				// to handle it specially as e.g.:
				// {$func: {$let: {vars: ..., in: [1,2,...]}}} is not valid
				// mongo aggregation language, the $let bindings must be inside
				// the array or document. This is related to the non-referential transparency
				// called out in: https://jira.mongodb.org/browse/SERVER-40494
				case *ast.Array:
					elements := typedArg.Elements
					for i := range elements {
						elements[i] = removeRedundancies(elements[i])
					}
					addExpression(typedExpr)
					return typedExpr
				case *ast.Document:
					elements := typedArg.Elements

					// $switch and $zip are further special cases, as they have fields
					// (discussed below) which MUST be arrays, not expressions that
					// evaluate to arrays. Therefore, we cannot just removeRedundancies
					// from each DocumentElement.
					switch typedExpr.Name {
					case "$switch":
						// The $switch function should have just two fields in its
						// argument document: "branches" and "default".
						for i := range elements {
							switch elements[i].Name {
							// The branches field is an array of cases, which are documents
							// of the shape {"case": ..., "then": ...}. This structure must
							// be maintained while removing redundancies, as in branches
							// must still be an array (not a $let that evaluates to an array)
							// and each case document must still have the two fields from
							// above. Therefore, we iterate through each element of the
							// branches array, and for each of those elements (which are
							// documents) we iterate through each document-element and
							// remove redundancies from those.
							case "branches":
								if branchesArray, ok := elements[i].Expr.(*ast.Array); ok {
									branchesElements := branchesArray.Elements
									for j := range branchesElements {
										if branchElementDoc, ok := branchesElements[j].(*ast.Document); ok {
											branchElementDocElements := branchElementDoc.Elements
											for k := range branchElementDocElements {
												branchElementDocElements[k].Expr = removeRedundancies(branchElementDocElements[k].Expr)
											}
										}
									}
								}

							// The default field can be any expression, so it is fine
							// to call removeRedundancies as below.
							case "default":
								elements[i].Expr = removeRedundancies(elements[i].Expr)
							}
						}

					// The $zip function should have up to three fields in its
					// argument document: "inputs", "useLongestLength", and
					// "defaults". The "useLongestLength" argument should just
					// be a boolean so we ignore it below.
					case "$zip":
						for i := range elements {
							// Both the "inputs" and "defaults" fields must be
							// arrays, not expressions that evaluate to arrays.
							// Therefore, we remove redundancies from each of
							// their elements, but not from the arrays as a whole.
							if elementAsArray, ok := elements[i].Expr.(*ast.Array); ok {
								subElements := elementAsArray.Elements
								for j := range subElements {
									subElements[j] = removeRedundancies(subElements[j])
								}
							}
						}
					default:
						for i := range elements {
							elements[i].Expr = removeRedundancies(elements[i].Expr)
						}
					}
					addExpression(typedExpr)
					return typedExpr
				default:
					typedExpr.Arg = removeRedundancies(typedExpr.Arg)
				}
			}
		}

		// Now do a bottom up traversal of the current expression to find redundant sub-expressions.
		_ = n.Walk(v)
		e, ok := n.(ast.Expr)
		if !ok {
			return n
		}
		if isTrivialExpression(e) {
			return n
		}
		addExpression(e)
		return n
	})

	// If the current expression is $expr itself, we return early so that we will not try to deparse
	// it. Once we have separate Match and Aggregation Expression trees, this check can go away.
	if _, ok := e.(*ast.AggExpr); ok {
		return e
	}

	freshSuffix := -1
	freshName := func() string {
		freshSuffix++
		return fmt.Sprintf("mongoast__deduplicated__expr__%d", freshSuffix)
	}
	bottomUpSortedLetVariables := []*ast.LetVariable{}
	variableRefs := make(map[string]*ast.VariableRef)
	// Now that we have found the redundant expressions, replace the redundant expressions and collect the let bindings.
	outNode, _ := ast.Visit(e, func(v ast.Visitor, n ast.Node) ast.Node {
		e, ok := n.(ast.Expr)
		if !ok {
			return n
		}
		preImage := astprint.String(e)
		n = n.Walk(v)
		e = n.(ast.Expr)
		if isTrivialExpression(e) {
			return n
		}
		if exprCount, ok := redundantExpressionCounts[preImage]; ok && exprCount > 1 {
			if varRef, exists := variableRefs[preImage]; exists {
				return varRef
			}
			freshVarName := freshName()
			freshVar := ast.NewVariableRef(freshVarName)
			bottomUpSortedLetVariables = append(bottomUpSortedLetVariables, ast.NewLetVariable(freshVarName, e))
			variableRefs[preImage] = freshVar
			return freshVar
		}
		return n
	})
	// Now build the lets.
	out := outNode.(ast.Expr)
	for i := len(bottomUpSortedLetVariables) - 1; i >= 0; i-- {
		out = ast.NewLet([]*ast.LetVariable{bottomUpSortedLetVariables[i]}, out)
	}
	return out
}

// isTrivialExpression returns true if an expression is trivial. A trivial expression is one for
// which we do not need to remove redundancy, though in the case of arrays and documents,
// we can remove internal redundancy.
func isTrivialExpression(e ast.Expr) bool {
	switch e.(type) {
	case *ast.Array,
		*ast.ArrayIndexRef,
		*ast.Constant,
		*ast.Document,
		*ast.FieldOrArrayIndexRef,
		*ast.FieldRef,
		*ast.Unknown,
		*ast.VariableRef,
		*ast.MatchRegex:
		return true
	default:
		return false
	}
}
