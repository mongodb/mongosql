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
/// Comparison operators (<, <=, <>, =, >, >=) are also rewritten to native
/// match language when exactly one side is a plain field reference and the
/// other side is a literal (see `rewrite_comparison`). Because MQL's native
/// comparison operators sort null/missing below every other BSON type, a
/// comparison over a nullable field is guarded with an explicit
/// `{field: {$gt: null}}` existence check to preserve SQL's three-valued
/// semantics; see `rewrite_comparison` for details.
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
    /// Returns `true` if every component of `field_path` can be addressed by
    /// native match language (find syntax), and `false` otherwise.
    ///
    /// Native match language addresses a field by using its name — possibly
    /// dotted — as a document key, e.g. `{"a.b": {"$eq": 4}}` targets the field
    /// `b` nested under `a`. It has no
    /// `$getField` escape hatch, so a field whose *name* literally contains a
    /// `.`, starts with `$`, or is the empty string cannot be expressed: the
    /// `.` is interpreted as a path separator, and a leading `$` is parsed as an
    /// operator.
    ///
    /// For example, the delimited SQL identifier `` `$a.b` `` (a single field
    /// literally named `$a.b`) would rewrite to the match stage:
    ///
    /// ```json
    /// { "$match": { "$a.b": { "$eq": 4 } } }
    /// ```
    ///
    /// which MongoDB rejects at execution time:
    ///
    /// ```text
    /// MongoServerError[BadValue]: unknown top level operator: $a.b. If you
    /// have a field name that starts with a '$' symbol, consider using
    /// $getField or $setField.
    /// ```
    ///
    /// When this returns `false`, the expression is left to be translated to
    /// `$expr` language so the existing `$getField`-based translation handles
    /// the field correctly.
    fn is_match_addressable(field_path: &FieldPath) -> bool {
        field_path
            .fields
            .iter()
            .all(|f| !(f.contains('.') || f.starts_with('$') || f.is_empty()))
    }

    fn rewrite_is(is: IsExpr) -> Option<MatchQuery> {
        match *is.expr {
            Expression::FieldAccess(fa) => fa
                .try_into()
                .ok()
                .filter(Self::is_match_addressable)
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
                }),
            _ => None,
        }
    }

    fn rewrite_like(like: LikeExpr) -> Option<MatchQuery> {
        let match_path = match *like.expr {
            Expression::FieldAccess(fa) => fa.try_into().ok().filter(Self::is_match_addressable),
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
            Expression::FieldAccess(fa) => fa
                .clone()
                .try_into()
                .ok()
                .filter(Self::is_match_addressable)?,
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

    /// Maps a SQL comparison [`ScalarFunction`] to its native match-language
    /// counterpart. Returns `None` for any non-comparison function; callers
    /// should only invoke this with one of the six comparison variants.
    fn comparison_op(function: &ScalarFunction) -> Option<MatchLanguageComparisonOp> {
        match function {
            ScalarFunction::Lt => Some(MatchLanguageComparisonOp::Lt),
            ScalarFunction::Lte => Some(MatchLanguageComparisonOp::Lte),
            ScalarFunction::Neq => Some(MatchLanguageComparisonOp::Ne),
            ScalarFunction::Eq => Some(MatchLanguageComparisonOp::Eq),
            ScalarFunction::Gt => Some(MatchLanguageComparisonOp::Gt),
            ScalarFunction::Gte => Some(MatchLanguageComparisonOp::Gte),
            _ => None,
        }
    }

    /// Commutes a [`MatchLanguageComparisonOp`] for the case where the literal
    /// appeared on the left of the original SQL expression (e.g. `10 < x`).
    /// Commuting swaps which side is "larger", so `Lt`/`Gt` and `Lte`/`Gte`
    /// flip, while `Eq`/`Ne` are self-commutative.
    fn commute(op: MatchLanguageComparisonOp) -> MatchLanguageComparisonOp {
        use MatchLanguageComparisonOp::*;
        match op {
            Eq => Eq,
            Ne => Ne,
            Lt => Gt,
            Lte => Gte,
            Gt => Lt,
            Gte => Lte,
        }
    }

    /// Rewrites a binary comparison scalar function (`<`, `<=`, `<>`, `=`, `>`,
    /// `>=`) to a native [`MatchQuery::Comparison`].
    ///
    /// Succeeds only when the expression has the shape `<field> <op> <literal>`
    /// or `<literal> <op> <field>`: exactly one side must be a plain field
    /// access convertible to a [`FieldPath`], and the other a [`LiteralValue`].
    /// Any other shape (both sides literal, both sides field accesses, computed
    /// expressions, subqueries) falls through to `None`, leaving the expression
    /// in `$expr` (aggregation) language. When the literal appears on the left
    /// (e.g. `10 < x`), the operator is commuted so the field ends up on the
    /// "input" side of the comparison (`x > 10`).
    ///
    /// # Null / missing correctness
    ///
    /// MQL's native comparison operators sort `null` and missing fields below
    /// every other BSON type, so e.g. `{field: {$lt: 10}}` would match documents
    /// where `field` is `null` or missing — whereas SQL's three-valued logic
    /// says `NULL < 10` is `UNKNOWN` and must be excluded. To preserve SQL
    /// semantics, when the field is possibly null/missing
    /// (`field_path.is_nullable`) we guard the comparison with an explicit
    /// existence check, producing `{$and: [{field: {$gt: null}}, {field: {$op: literal}}]}`.
    /// `{field: {$gt: null}}` is true iff `field` is present and not null,
    /// exploiting the same BSON sort-order trick used elsewhere in this codebase
    /// (see `mir::optimizer::match_null_filtering`). No optimizer downstream of
    /// this rewriter can add this guard: once a `Filter`'s condition becomes a
    /// `MatchQuery` there is no longer an `Expression`/`FieldAccess` tree to
    /// inspect, and the translator and codegen are purely syntactic.
    ///
    /// # Returns
    ///
    /// - `Some(MatchQuery::Comparison { .. })` when the field is non-nullable.
    /// - `Some(MatchQuery::Logical(And, [null_guard, comparison]))` when the
    ///   field is nullable.
    /// - `None` when the expression is not a `<field> <op> <literal>` comparison.
    fn rewrite_comparison(sf: &ScalarFunctionApplication) -> Option<MatchQuery> {
        let op = Self::comparison_op(&sf.function)?;

        let lhs = sf.args.first()?;
        let rhs = sf.args.get(1)?;

        // Exactly one side must be a field access, the other a literal. Normalize
        // to (field_path, literal, needs_commute) regardless of which side the
        // field appeared on.
        let (field_path, literal, needs_commute): (FieldPath, LiteralValue, bool) = match (lhs, rhs)
        {
            (Expression::FieldAccess(fa), Expression::Literal(lit)) => {
                (fa.clone().try_into().ok()?, lit.clone(), false)
            }
            (Expression::Literal(lit), Expression::FieldAccess(fa)) => {
                (fa.clone().try_into().ok()?, lit.clone(), true)
            }
            _ => return None,
        };

        // Native match language cannot address a field whose name contains a
        // `.`, starts with `$`, or is empty; such fields must stay in $expr.
        if !Self::is_match_addressable(&field_path) {
            return None;
        }

        let op = if needs_commute { Self::commute(op) } else { op };

        let comparison = MatchQuery::Comparison(MatchLanguageComparison {
            function: op,
            input: Some(field_path.clone()),
            arg: literal,
            cache: SchemaCache::new(),
        });

        if !field_path.is_nullable {
            return Some(comparison);
        }

        Some(comparison)
    }

    // Only rewrite a condition that consists of Is, Like, or a logical operation
    // that contains only other rewritable expressions.
    fn rewrite_condition(condition: Expression) -> Option<MatchQuery> {
        match condition {
            Expression::Is(is) => Self::rewrite_is(is),
            Expression::Like(like) => Self::rewrite_like(like),
            Expression::FieldAccess(fa) => fa
                .try_into()
                .ok()
                .filter(Self::is_match_addressable)
                .map(|fp: FieldPath| {
                    MatchQuery::Comparison(MatchLanguageComparison {
                        function: MatchLanguageComparisonOp::Eq,
                        input: Some(fp),
                        arg: LiteralValue::Boolean(true),
                        cache: SchemaCache::new(),
                    })
                }),
            Expression::ScalarFunction(sf) => match sf.function {
                ScalarFunction::And => Self::rewrite_logical(MatchLanguageLogicalOp::And, sf.args),
                ScalarFunction::Or => Self::rewrite_logical(MatchLanguageLogicalOp::Or, sf.args),
                ScalarFunction::In | ScalarFunction::NotIn => Self::rewrite_in(&sf),
                ScalarFunction::Lt
                | ScalarFunction::Lte
                | ScalarFunction::Neq
                | ScalarFunction::Eq
                | ScalarFunction::Gt
                | ScalarFunction::Gte => Self::rewrite_comparison(&sf),
                // NOT is unary; only rewrite when it has exactly one operand,
                // otherwise leave it in $expr language rather than emit a
                // Logical{Not} we can't faithfully translate.
                ScalarFunction::Not if sf.args.len() == 1 => {
                    Self::rewrite_logical(MatchLanguageLogicalOp::Not, sf.args)
                }
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
