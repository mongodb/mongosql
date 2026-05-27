/// Optimizes IS,LIKE, IN, and NOT IN expressions such that they can be translated and
/// codegened using match language. If a Filter stage's condition contains
/// only IS and/or LIKE expressions, or disjunctions or conjunctions with
/// only IS and/or LIKE expressions, then the condition can be rewritten to
/// match language (mir::MatchQuery) and the Filter stage replaced with a
/// MatchFilter stage.
///
/// Also optimizes constant false conditions to a MatchFalse filter. This filter
/// can be then treated specially in codegen because certain versions of mongodb
/// require a bit of finesse to ensure that {$match: {$expr: false}} is not
/// treated as a COLLSCAN.
///
/// Note that although comparison operators _can_ be rewritten to use match
/// language, this optimization does not perform such rewrites. LIKE and IS
/// ultimately translate to $regexMatch and $eq/$type in aggregation language.
/// Neither of those can utilize indexes when used in $match stages. Comparison
/// operators can utilize indexes even when they use expr language in $match.
/// Therefore, this optimization is only concerned with rewriting LIKE and IS.
///
/// Also note, MatchSplitting should ensure we never have a conjunction at this
/// point, however we choose to make this optimization work independent of that
/// one.
///
#[cfg(test)]
mod test;

use crate::mir::ArrayExpr;
use crate::{
    mir::{
        optimizer::Optimizer,
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Expression, FieldPath, IsExpr, LikeExpr, LiteralValue, MatchFalse, MatchFilter,
        MatchLanguageComparison, MatchLanguageComparisonOp, MatchLanguageIn, MatchLanguageInOp,
        MatchLanguageLogical, MatchLanguageLogicalOp, MatchLanguageRegex, MatchLanguageType,
        MatchQuery, MqlStage, ScalarFunction, ScalarFunctionApplication, Stage, Type,
        TypeOrMissing,
    },
    util::{convert_sql_pattern, LIKE_OPTIONS},
    SchemaCheckingMode,
};

pub(crate) struct MatchLanguageRewriter;

impl Optimizer for MatchLanguageRewriter {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> (Stage, bool) {
        let mut v = MatchLanguageRewriterVisitor::default();
        let new_stage = v.visit_stage(st);
        (new_stage, v.changed)
    }
}

#[derive(Default)]
struct MatchLanguageRewriterVisitor {
    changed: bool,
}

impl MatchLanguageRewriterVisitor {
    fn rewrite_is(is: IsExpr) -> Option<MatchQuery> {
        match *is.expr {
            Expression::FieldAccess(fa) => fa
                .try_into()
                .map(|mp: FieldPath| {
                    if is.target_type == TypeOrMissing::Type(Type::Null) {
                        MatchQuery::Comparison(MatchLanguageComparison {
                            function: MatchLanguageComparisonOp::Eq,
                            input: Some(mp),
                            arg: LiteralValue::Null,
                            cache: SchemaCache::new(),
                        })
                    } else {
                        MatchQuery::Type(MatchLanguageType {
                            input: Some(mp),
                            target_type: is.target_type,
                            cache: SchemaCache::new(),
                        })
                    }
                })
                .ok(),
            _ => None,
        }
    }

    fn rewrite_like(like: LikeExpr) -> Option<MatchQuery> {
        let match_path = match *like.expr {
            Expression::FieldAccess(fa) => fa.try_into().ok(),
            _ => return None,
        };

        let pat = match *like.pattern {
            Expression::Literal(LiteralValue::String(p)) => p,
            _ => return None,
        };

        let mql_pattern = convert_sql_pattern(pat, like.escape);

        match_path.map(|mp| {
            MatchQuery::Regex(MatchLanguageRegex {
                input: Some(mp),
                regex: mql_pattern,
                options: LIKE_OPTIONS.clone(),
                cache: SchemaCache::new(),
            })
        })
    }

    fn rewrite_logical(op: MatchLanguageLogicalOp, args: Vec<Expression>) -> Option<MatchQuery> {
        args.into_iter()
            .map(Self::rewrite_condition)
            .collect::<Option<Vec<MatchQuery>>>()
            .map(|ma| {
                MatchQuery::Logical(MatchLanguageLogical {
                    op,
                    args: ma,
                    cache: SchemaCache::new(),
                })
            })
    }

    /// Rewrites an `IN` or `NOT IN` scalar function to a native [`MatchQuery::In`].
    ///
    /// Succeeds only when the expression has the shape `<field> IN (<literal>, …)`:
    /// the left-hand side must be a plain field access convertible to a [`FieldPath`],
    /// and every element on the right-hand side must be a [`LiteralValue`]. Any other
    /// shape — subqueries, field references inside the list, computed expressions — falls
    /// through to `None`, leaving the expression in `$expr` (aggregation) language.
    ///
    /// The distinction matters for query performance: the native `$in` / `$nin`
    /// operators can exploit indexes, whereas `$expr` cannot.
    ///
    /// # Returns
    ///
    /// - `Some(MatchQuery::In { … })` when the expression matches the required shape.
    /// - `None` when:
    ///   - `sf.args[0]` is not a [`FieldAccess`], or cannot be converted to a [`FieldPath`].
    ///   - `sf.args[1]` is not an [`ArrayExpr`].
    ///   - Any element of the array is not a [`LiteralValue`].
    ///   - `sf.function` is neither [`ScalarFunction::In`] nor [`ScalarFunction::NotIn`].
    fn rewrite_in(sf: &ScalarFunctionApplication) -> Option<MatchQuery> {
        // The left-hand side must be a plain field reference (e.g. `age`, `user.city`).
        // Anything more complex — a function call, a subquery, a literal — cannot be
        // mapped to the FieldPath that MatchLanguageIn requires, so we bail early.
        let field_path: FieldPath = match sf.args.first()? {
            Expression::FieldAccess(fa) => fa.clone().try_into().ok()?,
            _ => return None,
        };

        // Every value in the list must be a compile-time literal. We map each element
        // to `Some(literal)` or `None`, then collect into `Option<Vec<_>>`: if even one
        // element is `None` the entire collect() short-circuits to `None`, which the `?`
        // propagates. This all-or-nothing behaviour is intentional — a mixed list (e.g.
        // `age IN (18, min_age, 65)`) cannot use index-friendly match language and must
        // stay in $expr.
        let values: Vec<LiteralValue> = match sf.args.get(1)? {
            Expression::Array(ArrayExpr { array }) => array
                .iter()
                .map(|e| {
                    if let Expression::Literal(lit) = e {
                        Some(lit.clone())
                    } else {
                        None
                    }
                })
                .collect::<Option<Vec<_>>>()?,
            _ => return None,
        };

        // Map the SQL operator to its match-language counterpart. The wildcard arm is a
        // defensive guard; callers should only pass In/NotIn to this function.
        let op = match sf.function {
            ScalarFunction::In => MatchLanguageInOp::In,
            ScalarFunction::NotIn => MatchLanguageInOp::NotIn,
            _ => return None,
        };

        Some(MatchQuery::In(MatchLanguageIn {
            op,
            input: field_path,
            values,
            cache: SchemaCache::new(),
        }))
    }

    // Only rewrite a condition that consists of Is, Like, or a logical operation
    // that contains only other rewritable expressions.
    fn rewrite_condition(condition: Expression) -> Option<MatchQuery> {
        match condition {
            Expression::Is(is) => Self::rewrite_is(is),
            Expression::Like(like) => Self::rewrite_like(like),
            Expression::ScalarFunction(sf) => match sf.function {
                ScalarFunction::And => Self::rewrite_logical(MatchLanguageLogicalOp::And, sf.args),
                ScalarFunction::Or => Self::rewrite_logical(MatchLanguageLogicalOp::Or, sf.args),
                ScalarFunction::In | ScalarFunction::NotIn => Self::rewrite_in(&sf),
                _ => None,
            },
            // Note this relies on ConstantFolding to ensure that the a constant expression becomes
            // a LiteralValue.
            Expression::Literal(l) if l.is_falsy() => Some(MatchQuery::False(MatchFalse {
                cache: SchemaCache::new(),
            })),
            _ => None,
        }
    }
}

impl Visitor for MatchLanguageRewriterVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);

        match node.clone() {
            Stage::Filter(f) => {
                // If a Filter's condition can be rewritten to match language,
                // replace the Filter with an MqlIntrinsic MatchFilter with the
                // rewritten condition.
                Self::rewrite_condition(f.condition.clone()).map_or(node, |condition| {
                    self.changed = true;
                    Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
                        source: f.source,
                        condition,
                        cache: SchemaCache::new(),
                    })))
                })
            }
            _ => node,
        }
    }
}
