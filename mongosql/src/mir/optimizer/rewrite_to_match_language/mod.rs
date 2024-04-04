/// Optimizes IS and LIKE expressions such that they can be translated and
/// codegenned using match language. If a Filter stage's condition contains
/// only IS and/or LIKE expressions, or disjunctions or conjunctions with
/// only IS and/or LIKE expressions, then the condition can be rewritten to
/// match language (mir::MatchQuery) and the Filter stage replaced with a
/// MatchFilter stage.
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

use crate::{
    mir::{
        optimizer::Optimizer,
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Expression, FieldPath, IsExpr, LikeExpr, LiteralValue, MQLStage, MatchFilter,
        MatchLanguageComparison, MatchLanguageComparisonOp, MatchLanguageLogical,
        MatchLanguageLogicalOp, MatchLanguageRegex, MatchLanguageType, MatchQuery, ScalarFunction,
        Stage, Type, TypeOrMissing,
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

    // Only rewrite a condition that consists of Is, Like, or a logical operation
    // that contains only other rewritable expressions.
    fn rewrite_condition(condition: Expression) -> Option<MatchQuery> {
        match condition {
            Expression::Is(is) => Self::rewrite_is(is),
            Expression::Like(like) => Self::rewrite_like(like),
            Expression::ScalarFunction(sf) => match sf.function {
                ScalarFunction::And => Self::rewrite_logical(MatchLanguageLogicalOp::And, sf.args),
                ScalarFunction::Or => Self::rewrite_logical(MatchLanguageLogicalOp::Or, sf.args),
                _ => None,
            },
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
                // replace the Filter with an MQLIntrinsic MatchFilter with the
                // rewritten condition.
                Self::rewrite_condition(f.condition.clone()).map_or(node, |condition| {
                    self.changed = true;
                    Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                        source: f.source,
                        condition,
                        cache: SchemaCache::new(),
                    }))
                })
            }
            _ => node,
        }
    }
}
