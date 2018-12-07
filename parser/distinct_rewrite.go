package parser

import (
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/internal/util/option"
)

// RewriteDistinct tries to rewrite queries using
// distinct aggregation operators in terms of grouped subqueries.
func RewriteDistinct(stmt Statement) Statement {
	switch stmt.(type) {
	case *Select, *SimpleSelect, *Union:
	default:
		return stmt
	}
	distinctRw := NewDistinctRewriter()

	stmtCopy := stmt.Copy()
	// Attempt the rewrite.
	newStmt, err := walk(distinctRw, stmtCopy)
	if err != nil {
		panic(err)
	}

	return newStmt.(Statement)
}

type distinctFuncState struct {
	// distinctFuncExpr is the aggregation function with a distinct constraint.
	distinctFuncExpr *FuncExpr
	// expressionIndex is the index of the select expression that contains the
	// distinct group function.
	expressionIndex int
}

var _ walker = (*DistinctRewriter)(nil)

// DistinctRewriter tries to rewrite queries using GROUP BY in sub queries
// instead of as distinct aggregation functions.
type DistinctRewriter struct {
	distinctMaps  map[CST]*distinctFuncState
	uniqueIDCount int
	// We only want to rewrite selects in table contexts, that means
	// either the top level or in the FROM clause. We do not
	// want to rewrite in SELECT or WHERE clauses because MySQL
	// only allows correlating one level, and introducing a subquery,
	// as this transformation does, results in correlated columns
	// disappearing.
	inTableContext bool
}

// NewDistinctRewriter creates a new DistinctRewriter.
func NewDistinctRewriter() *DistinctRewriter {
	return &DistinctRewriter{
		distinctMaps:   make(map[CST]*distinctFuncState),
		uniqueIDCount:  -1,
		inTableContext: true,
	}
}

// PreVisit will collect information necessary to perform the
// rewrite, it keeps track of the single distinct FuncExpr
// allowed per Select statement.
func (d *DistinctRewriter) PreVisit(current CST) (CST, error) {
	switch typed := current.(type) {
	case TableExprs:
		d.inTableContext = true
		return current, nil
	case SelectExpr, *Where:
		d.inTableContext = false
		return current, nil
	case *Select:
		if !d.inTableContext {
			return current, nil
		}

		// If there is a distinct in having, the rewrite we perform
		// here will cause incorrect results. To properly support having
		// would require two subqueries: the first to group the distinct
		// expression in the having, the second to apply the having
		// to the group by. Since having clauses are relatively rare,
		// we decided not to support that at this time.
		if typed.Having != nil {
			f := &distinctFuncExprFinder{}
			_, err := walk(f, typed.Having)
			if err != nil {
				return nil, err
			}
			if len(f.distinctExprs) > 0 {
				return current, nil
			}
		}
		foundOneDistinct := false
		for i, expr := range typed.SelectExprs {
			f := &distinctFuncExprFinder{}
			_, err := walk(f, expr)
			if err != nil {
				return nil, err
			}

			// If there are any non-distinct Aggregation Functions,
			// our rewrite will be semantically incorrect, as the rewrite
			// will attempt to apply them to values with an extra grouping:
			//   select sum(distinct a) as sda, sum(a) as sa from foo
			// would rewrite to:
			//   select sum(sda) as sda, sum(sa) as sa from
			//      (select a as sda, a as sda group by 1) $___mongosqld_query_0
			// This would force both sums to be distinct, when the goal is for only
			// the first sum to be distinct.
			if f.foundNonDistinctAgg {
				// Set the entry for Select `typed` to nil,
				// because we might have given it a value
				// at a previous iteration of this loop.
				d.distinctMaps[typed] = nil
				return current, nil
			}

			// More than one distinct expression is an issue because we can only group
			// by one extra expression without changing the semantics:
			// The proper rewrite for
			//    select sum(distinct a) as sda, sum(distinct b) as sdb from foo
			// should rewrite to:
			//    select sum(sda) as sda, sum(sdb) as sdb from
			//        (select a as sda, NULL as sdb from foo group by 1
			//         union all
			//         select NULL as sda, b as sdb from foo group by 2)
			// but there is little utility to this as this point since we do not
			// have push down for union. This union is necessary because we want
			// each unique a, and each unique b, if we used one subquery grouped
			// by 1,2, we would get each unique *pair* of a,b, e.g.:
			// a=1,b=2 and a=2,b=2 are distinct, even though b is 2 twice.
			if len(f.distinctExprs) > 1 {
				// Set the entry for Select `typed` to nil,
				// because we might have given it a value
				// at a previous iteration of this loop.
				d.distinctMaps[typed] = nil
				return current, nil
			}
			if len(f.distinctExprs) == 1 {
				// We cannot rewrite, if there is a subquery in the selectExprs, because
				// we have no way to know at this phase whether or not it is correlated.
				// If it is correlated, the rewrite will break the query because mysql
				// only allows correlation one subquery up, and the rewrite will push
				// the correlation into the new subquery we generate, thus making
				// the correlated columns unknown:
				//    select (select sum(distinct b) as sdb from bar where b = foo.a) from foo
				// would rewrite to:
				//    select (select sum(sdb) as sdb from
				//              (select sdb from bar group by 1 where b = foo.a)) from foo
				// Because there is now an intervening subquery, foo.a is unknown.
				if f.foundSubquery {
					d.distinctMaps[typed] = nil
					return current, nil
				}

				// If we have already found a distinct expression before, we are again
				// in that situation where there is more than one distinct, which
				// does not work when rewritten, as explained above.
				if foundOneDistinct {
					d.distinctMaps[typed] = nil
					return current, nil
				}
				foundOneDistinct = true
				d.distinctMaps[typed] = &distinctFuncState{
					distinctFuncExpr: f.distinctExprs[0],
					expressionIndex:  i,
				}
			}
		}
	}
	return current, nil
}

var _ walker = (*distinctFuncExprFinder)(nil)

// distinctFuncExprFinder will find all distinct FuncExprs
// under a given root expression.
type distinctFuncExprFinder struct {
	distinctExprs       []*FuncExpr
	foundNonDistinctAgg bool
	foundSubquery       bool
}

// PreVisit for distinctFuncExprFinder finds all distinct FuncExprs.
// It also records if any non-distinct aggregation functions were found,
// as they would break our transformation.
func (f *distinctFuncExprFinder) PreVisit(current CST) (CST, error) {
	switch typed := current.(type) {
	case *FuncExpr:
		if typed.Distinct {
			f.distinctExprs = append(f.distinctExprs, typed)
		} else if _, ok := AggregationFunctions[strings.ToLower(typed.Name)]; ok {
			f.foundNonDistinctAgg = true
		}
	case *Subquery:
		f.foundSubquery = true
	}
	return current, nil
}

// PostVisit for distinctFuncExprFinder does nothing.
func (f *distinctFuncExprFinder) PostVisit(current CST) (CST, error) {
	return current, nil
}

// newQueryName generates hopefully unique query names.
// It is highly unlikely that a user would use names such
// as the name used here, and saves us the cost of checking
// every single name used in the query.
func (d *DistinctRewriter) newQueryName() string {
	d.uniqueIDCount++
	return "$___mongosqld_query_" + strconv.Itoa(d.uniqueIDCount)
}

// newAsName generates hopefully unique alias names.
// It is highly unlikely that a user would use names such
// as the name used here, and saves us the cost of checking
// every single name used in the query.
func (d *DistinctRewriter) newAsName() string {
	d.uniqueIDCount++
	return "$___mongosqld_as_" + strconv.Itoa(d.uniqueIDCount)
}

// getInnerAndOuterAlias generates the pairs of aliases needed for a given SelectExpr position
// in the inner and outer query.
func (d *DistinctRewriter) getInnerAndOuterAlias(nonStarExpr *NonStarExpr) (string, string) {
	var innerAs string
	var outerAs string
	if nonStarExpr.As != option.NoneString() {
		innerAs = nonStarExpr.As.Unwrap()
		outerAs = innerAs
	} else {
		// Otherwise, create an alias to refer to.
		innerAs = d.newAsName()
		buff := NewTrackedBuffer(nil)
		nonStarExpr.Expr.Format(buff)
		outerAs = buff.String()
	}
	return innerAs, outerAs
}

// PostVisit for DistinctRewriter actually performs the query transformation.
// To understand this transformation we will use the the following query as an
// example:
//
// select NULL, sum(distinct x) from foo union all select a, count(distinct b+c) from bar group by a;
//
func (d *DistinctRewriter) PostVisit(current CST) (CST, error) {
	switch query := current.(type) {
	case *Select:
		// When we traverse back up to a Select that contains a distinct
		// FuncExpr, it will be found in the following map. The state of the
		// map for our running example looks like this (note, the keys are not
		// actually strings):
		//
		// map{"select NULL, sum(distinct x) from foo" :
		//           distinctFuncState {
		// 	              distinctFuncExpr: "sum(distinct x)"
		//	              expressionIndex: 1
		//           },
		//     "select a, count(distinct b+c) from bar group by a" :
		//           distinctFuncState {
		// 	              distinctFuncExpr: "count(distinct b+c)"
		//	              expressionIndex: 1
		//           },
		//    }
		agg := d.distinctMaps[query]
		if agg == nil {
			return current, nil
		}

		// d.inTableContext must be set to true in order to optimize following
		// selects in a union.
		d.inTableContext = true

		// At this point, we know the Select in question contains
		// one and only one distinct aggregation function.
		// Create a new outer select expression, the original
		// query will be a subquery.
		outerQuery := &Select{
			QueryGlobals: &QueryGlobals{
				false, false,
			},
			SelectExprs: make(SelectExprs, len(query.SelectExprs)),
			GroupBy:     query.GroupBy.Copy().(GroupBy),
		}

		// Create proper selectExprs for the outer query based on
		// the selectExprs of what will now be a subquery.
		for i, selectExpr := range query.SelectExprs {
			// Consider the second Select in our example:
			//     select a, count(distinct b+c) from bar group by a
			// The selectExpr generated for the outer query for each
			// iteration will be:
			//  i   |  expr
			// -----+----------------------------
			//  0   |   $___mongosqld_as_1 as a
			//  1   |   count($___mongosqld_as_2) as count(distinct b+c)
			// -----+----------------------------
			if _, ok := selectExpr.(*StarExpr); ok {
				outerQuery.SelectExprs[i] = selectExpr
				continue
			}
			nonStarExpr := selectExpr.(*NonStarExpr)
			innerAs, outerAs := d.getInnerAndOuterAlias(nonStarExpr)
			nonStarExpr.As = option.SomeString(innerAs)
			if i == agg.expressionIndex {
				outerQuery.SelectExprs[i] = &NonStarExpr{
					Expr: agg.distinctFuncExpr,
					As:   option.SomeString(outerAs),
				}

				// This type assertion is safe because a StarExpr cannot exist in an
				// distinct aggregation FuncExpr.
				nonStarExpr.Expr = agg.distinctFuncExpr.Exprs[0].(*NonStarExpr).Expr
				agg.distinctFuncExpr.Exprs = SelectExprs{
					&NonStarExpr{Expr: &ColName{Name: innerAs}},
				}

				// Now that we have generated the as alias, we can remove the distinct.
				agg.distinctFuncExpr.Distinct = false
			} else {
				outerQuery.SelectExprs[i] = &NonStarExpr{
					Expr: &ColName{Name: innerAs},
					As:   option.SomeString(outerAs),
				}
			}
		}

		// At this point, we have generated the proper SelectExprs for the outquery.
		// We will consider the second Select in our example:
		//     select a, count(distinct b+c) from bar group by a
		// The outer query generated for this currently looks like the following:
		//     select $___mongosqld_as_1 as a, count($___mongosqld_as_2) as count(distinct a+b)
		//     from nil group by a
		// The nil exists because we have not yet set the original, soon to be
		// inner, query to be the subquery. The inner query looks like the following:
		//     select a, b+c as $___mongosqld_as_2 from bar group by a
		numVal := NumVal(strconv.Itoa(agg.expressionIndex + 1))

		// Add the groupby for the subquery.
		// After adding the groupby, the inner query will look like:
		//     select a as $___mongosqld_as_1, b+c as $___mongosqld_as_2 from bar group by a, 2
		if query.GroupBy == nil {
			query.GroupBy = []Expr{
				numVal,
			}
		} else {
			query.GroupBy = append(query.GroupBy, numVal)
		}

		// Now setting the From of the outer to be the original, now inner, results in:
		// select a, count($___mongosqld_as_2) from
		//   (select a, b+c as $___mongosqld_as_2 from bar group by a, 1) $___mongosqld_query_3
		// group by a
		outerQuery.From = []TableExpr{
			&AliasedTableExpr{
				Expr: &Subquery{Select: query},
				As:   option.SomeString(d.newQueryName()),
			},
		}
		return outerQuery, nil
	}
	return current, nil
}
