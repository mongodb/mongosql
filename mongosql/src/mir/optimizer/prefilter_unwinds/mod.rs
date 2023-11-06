///
/// Prefilter Unwinds
///
/// The Prefilter Unwinds pass generates ElemMatch matches before Unwind stages
/// when there is a Filter after the Unwind that cannot move passed the Unwind
/// since it uses an Opaque field from the Unwind. The ElemMatch will potentially
/// not be entirely selective, but in conjunction with Stage Movement, it can possibly
/// allow the usage of more indexes that can massively improve query performance.
///
/// As an example of where this can really help, consider:
/// SELECT * FROM UNWIND(foo WITH path => v) WHERE v = 6
///
/// The ElemMatch on v = 6 will result in an index that removes any documents where the v array
/// does not contain 6. Note that the original Filter will be maintained after the Unwind so that
/// any value that is not explicitly 6 will be removed still.
///
/// In general, this can reduce a very large COL_SCANs into a possibly very selective IDX_SCANs
/// depending on the size of the arrays and volitility of data in the arrays.
///
#[cfg(test)]
mod test;

use super::Optimizer;
use crate::{
    mir::{
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        ElemMatch, Expression, FieldPath, Filter, LiteralValue, MQLStage, MatchFilter,
        MatchLanguageComparison, MatchLanguageComparisonOp, MatchLanguageLogical,
        MatchLanguageLogicalOp, MatchQuery, ScalarFunction, ScalarFunctionApplication, Stage,
        Unwind,
    },
    SchemaCheckingMode,
};

pub(crate) struct PrefilterUnwindsOptimizer {}

impl Optimizer for PrefilterUnwindsOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> Stage {
        let mut visitor = PrefilterUnwindsVisitor {};
        visitor.visit_stage(st)
    }
}

struct PrefilterUnwindsVisitor {}

impl Visitor for PrefilterUnwindsVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node {
            Stage::Filter(f) => match *f.source {
                Stage::Unwind(u) => {
                    let (field_uses, condition) = f.condition.field_uses();
                    let field_uses = if let Some(field_uses) = field_uses {
                        field_uses
                    } else {
                        return Stage::Filter(Filter {
                            source: Box::new(Stage::Unwind(u)),
                            condition,
                            ..f
                        });
                    };
                    if field_uses.len() == 1 {
                        let field_use = field_uses.into_iter().next().unwrap();
                        let opaque_field_defines = u.opaque_field_defines();
                        if !opaque_field_defines.contains(&field_use) {
                            return Stage::Filter(Filter {
                                source: Box::new(Stage::Unwind(u)),
                                condition,
                                ..f
                            });
                        }
                        if field_use.fields.get(0) == u.index.as_ref() {
                            return Stage::Filter(Filter {
                                source: Box::new(Stage::Unwind(u)),
                                condition,
                                ..f
                            });
                        }
                        Stage::Filter(Filter {
                            source: Box::new(Stage::Unwind(Unwind {
                                source: generate_prefilter(field_use, u.source, &condition),
                                ..u
                            })),
                            condition,
                            cache: f.cache,
                        })
                    } else {
                        Stage::Filter(Filter {
                            source: Box::new(Stage::Unwind(u)),
                            condition,
                            cache: f.cache,
                        })
                    }
                }
                _ => Stage::Filter(Filter { ..f }),
            },
            _ => node,
        }
    }
}

// generate_prefilter returns the original source Stage, if it fails to generate a useful ElemMatchLanguage
fn generate_prefilter(field: FieldPath, source: Box<Stage>, condition: &Expression) -> Box<Stage> {
    if let Expression::ScalarFunction(ScalarFunctionApplication {
        function: f,
        args: a,
        ..
    }) = condition
    {
        if matches!(f, ScalarFunction::Between) {
            let (lower_bound, upper_bound) =
                if let Some((lower_bound, upper_bound)) = get_between_literals(a) {
                    (lower_bound, upper_bound)
                } else {
                    return source;
                };
            return Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
                source,
                condition: generate_between_elem_match_query(field, lower_bound, upper_bound),
                cache: SchemaCache::new(),
            }))
            .into();
        }
        let lit = if let Some(lit) = get_comparison_literal(a) {
            lit
        } else {
            return source;
        };
        let function: MatchLanguageComparisonOp = if let Ok(op) = (*f).try_into() {
            op
        } else {
            return source;
        };
        return Stage::MQLIntrinsic(MQLStage::MatchFilter(MatchFilter {
            source,
            condition: generate_comparison_elem_match_query(function, field, lit),
            cache: SchemaCache::new(),
        }))
        .into();
    }
    source
}

fn generate_comparison_elem_match_query(
    function: MatchLanguageComparisonOp,
    field: FieldPath,
    lit: LiteralValue,
) -> MatchQuery {
    MatchQuery::ElemMatch(ElemMatch {
        input: field,
        condition: generate_match_comparison(function, lit).into(),
        cache: SchemaCache::new(),
    })
}

fn generate_between_elem_match_query(
    field: FieldPath,
    lower_bound: LiteralValue,
    upper_bound: LiteralValue,
) -> MatchQuery {
    MatchQuery::ElemMatch(ElemMatch {
        input: field,
        condition: MatchQuery::Logical(MatchLanguageLogical {
            op: MatchLanguageLogicalOp::And,
            args: vec![
                generate_match_comparison(MatchLanguageComparisonOp::Gte, lower_bound),
                generate_match_comparison(MatchLanguageComparisonOp::Lte, upper_bound),
            ],
            cache: SchemaCache::new(),
        })
        .into(),
        cache: SchemaCache::new(),
    })
}

fn generate_match_comparison(function: MatchLanguageComparisonOp, lit: LiteralValue) -> MatchQuery {
    MatchQuery::Comparison(MatchLanguageComparison {
        function,
        input: None,
        arg: lit,
        cache: SchemaCache::new(),
    })
}

fn get_comparison_literal(args: &Vec<Expression>) -> Option<LiteralValue> {
    if args.len() != 2 {
        return None;
    }
    match args.as_slice() {
        [Expression::Literal(ref lit), Expression::FieldAccess(_)] => Some(lit.clone()),
        [Expression::FieldAccess(_), Expression::Literal(ref lit)] => Some(lit.clone()),
        _ => None,
    }
}

fn get_between_literals(args: &Vec<Expression>) -> Option<(LiteralValue, LiteralValue)> {
    if args.len() != 3 {
        return None;
    }
    match args.as_slice() {
        [Expression::FieldAccess(_), Expression::Literal(ref lit1), Expression::Literal(ref lit2)] => {
            Some((lit1.clone(), lit2.clone()))
        }
        _ => None,
    }
}

impl TryFrom<ScalarFunction> for MatchLanguageComparisonOp {
    // We'll never use this error, at this time, so we just make it ()
    type Error = ();

    fn try_from(s: ScalarFunction) -> Result<Self, Self::Error> {
        Ok(match s {
            ScalarFunction::Lt => Self::Lt,
            ScalarFunction::Lte => Self::Lte,
            ScalarFunction::Neq => Self::Ne,
            ScalarFunction::Eq => Self::Eq,
            ScalarFunction::Gt => Self::Gt,
            ScalarFunction::Gte => Self::Gte,
            _ => return Err(()),
        })
    }
}
