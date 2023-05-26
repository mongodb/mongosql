#[cfg(test)]
mod test;

use super::Optimizer;
use crate::{
    mir::{
        schema::{CachedSchema, SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Expression, FieldAccess, Filter, OptimizedMatchExists, ScalarFunction,
        ScalarFunctionApplication, Stage,
    },
    schema::{Satisfaction, NULLISH},
    SchemaCheckingMode,
};
use std::collections::BTreeMap;

/// Optimizes Filter stages by prepending them, when necessary, with Filter
/// stages that assert nullable FieldAccess expressions exist and are not null.
///
/// If a Filter stage's condition contains SQL-semantic ScalarFunctions, then
/// for all FieldAccess expressions that MAY or MUST be NULL or MISSING that
/// are descendants of those ScalarFunctions:
///   1. Find all FieldAccess expressions that MAY or MUST be NULL or MISSING
///      that are descendants of those Scalar Functions.
///   2. Insert a _new_ Filter stage before it that asserts those FieldAccess
///      expressions exist and are not null.
///   3. Update those FieldAccesses in the _original_ Filter's condition to
///      indicate they are _NOT_ NULL or MISSING.
///
/// This will help the final MQL translations utilize indexes since it will
/// ensure Filter stages can always use MQL semantics for condition operators.
pub(crate) struct MatchNullFilteringOptimizer;

impl Optimizer for MatchNullFilteringOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_env: &SchemaInferenceState,
    ) -> Stage {
        let mut v = MatchNullFilteringVisitor;
        v.visit_stage(st)
    }
}

struct MatchNullFilteringVisitor;

impl MatchNullFilteringVisitor {
    /// create_null_filter_stage attempts to create a Filter stage that ensures
    /// possibly null FieldAccesses in the original_filter exist. If the
    /// original_filter does not contain any null-semantic operators or does
    /// not contain any nullable FieldAccesses, this method returns None.
    fn create_null_filter_stage(&self, original_filter: &Filter) -> Option<Filter> {
        self.generate_null_filter_condition(&original_filter.condition)
            .map(|null_filter_condition| Filter {
                source: original_filter.source.clone(),
                condition: null_filter_condition,
                cache: SchemaCache::new(),
            })
    }

    /// generate_null_filter_condition attempts to create an Expression that
    /// ensures possibly null FieldAccesses in the condition exist. If the
    /// condition does not contain any null-semantic operators or does not
    /// contain any nullable FieldAccesses, this method returns None.
    fn generate_null_filter_condition(&self, condition: &Expression) -> Option<Expression> {
        let optimized_exists_ops = self
            .gather_fields_for_null_filter(condition)
            .into_values()
            .map(|fa| {
                Expression::OptimizedMatchExists(OptimizedMatchExists {
                    field_access: fa,
                    cache: SchemaCache::new(),
                })
            })
            .collect::<Vec<Expression>>();

        match optimized_exists_ops.len() {
            0 => None,
            1 => optimized_exists_ops.first().cloned(),
            _ => Some(Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::And,
                args: optimized_exists_ops,
                cache: SchemaCache::new(),
            })),
        }
    }

    /// gather_fields_for_null_filter collects unique "pure" FieldAccesses
    /// in the provided condition that appear as descendents of sql-semantic
    /// operators.
    fn gather_fields_for_null_filter(
        &self,
        condition: &Expression,
    ) -> BTreeMap<String, FieldAccess> {
        let mut visitor = NullableFieldAccessGatherer {
            filter_fields: BTreeMap::new(),
            is_collecting: false,
        };
        visitor.visit_expression(condition.clone());
        visitor.filter_fields
    }
}

impl Visitor for MatchNullFilteringVisitor {
    fn visit_filter(&mut self, node: Filter) -> Filter {
        let node = node.walk(self);
        match self.create_null_filter_stage(&node) {
            None => node,
            Some(null_filter_stage) => Filter {
                source: Box::new(Stage::Filter(null_filter_stage)),
                condition: node.condition,
                cache: SchemaCache::new(),
            },
        }
    }
}

struct NullableFieldAccessGatherer {
    filter_fields: BTreeMap<String, FieldAccess>,
    is_collecting: bool,
}

impl Visitor for NullableFieldAccessGatherer {
    // Do not walk stages nested within expressions.
    fn visit_stage(&mut self, node: Stage) -> Stage {
        node
    }

    fn visit_expression(&mut self, node: Expression) -> Expression {
        match node.clone() {
            // If we encounter a "nullable" scalar function, meaning a scalar
            // function with SQL-style semantics, then we want to collect
            // nullable fields nested within that function's arguments.
            Expression::ScalarFunction(sf) if self.is_nullable_function(sf.function) => {
                let old_is_collecting = self.is_collecting;
                self.is_collecting = true;
                sf.walk(self);
                self.is_collecting = old_is_collecting;

                node
            }
            Expression::FieldAccess(fa) => {
                if self.is_collecting && self.schema_is_nullish(node.clone()) {
                    // Only store nullable "pure" fields in the map.
                    match fa.to_string_if_pure() {
                        None => (),
                        Some(s) => {
                            self.filter_fields.insert(s, fa);
                        }
                    }
                }

                node
            }
            _ => node.walk(self),
        }
    }
}

impl NullableFieldAccessGatherer {
    fn is_nullable_function(&self, f: ScalarFunction) -> bool {
        match f {
            ScalarFunction::Lt
            | ScalarFunction::Lte
            | ScalarFunction::Neq
            | ScalarFunction::Eq
            | ScalarFunction::Gt
            | ScalarFunction::Gte
            | ScalarFunction::Between
            | ScalarFunction::Not
            | ScalarFunction::And
            | ScalarFunction::Or => true,

            // We include every variant instead of a wildcard in case
            // new variants are added in the future that could benefit
            // from this optimization.
            ScalarFunction::Concat
            | ScalarFunction::Pos
            | ScalarFunction::Neg
            | ScalarFunction::Add
            | ScalarFunction::Sub
            | ScalarFunction::Mul
            | ScalarFunction::Div
            | ScalarFunction::ComputedFieldAccess
            | ScalarFunction::NullIf
            | ScalarFunction::Coalesce
            | ScalarFunction::Slice
            | ScalarFunction::Size
            | ScalarFunction::Position
            | ScalarFunction::CharLength
            | ScalarFunction::OctetLength
            | ScalarFunction::BitLength
            | ScalarFunction::Abs
            | ScalarFunction::Ceil
            | ScalarFunction::Cos
            | ScalarFunction::Degrees
            | ScalarFunction::Floor
            | ScalarFunction::Log
            | ScalarFunction::Mod
            | ScalarFunction::Pow
            | ScalarFunction::Radians
            | ScalarFunction::Round
            | ScalarFunction::Sin
            | ScalarFunction::Sqrt
            | ScalarFunction::Tan
            | ScalarFunction::Substring
            | ScalarFunction::Upper
            | ScalarFunction::Lower
            | ScalarFunction::BTrim
            | ScalarFunction::LTrim
            | ScalarFunction::RTrim
            | ScalarFunction::Split
            | ScalarFunction::CurrentTimestamp
            | ScalarFunction::Year
            | ScalarFunction::Month
            | ScalarFunction::Day
            | ScalarFunction::Hour
            | ScalarFunction::Minute
            | ScalarFunction::Second
            | ScalarFunction::Week
            | ScalarFunction::DayOfYear
            | ScalarFunction::IsoWeek
            | ScalarFunction::IsoWeekday
            | ScalarFunction::MergeObjects => false,
        }
    }

    fn schema_is_nullish(&self, expr: Expression) -> bool {
        // We cannot use the SchemaInferenceState provided to the optimize
        // function here to calculate the schema because the Expression may
        // include a Reference that is not guaranteed to exist in the state's
        // SchemaEnvironment. Accessing the the cache is the best way to get
        // this Expression's schema.
        match expr.get_cached_schema() {
            None => true, // assume null in absence of schema
            Some(schema_result) => match schema_result {
                Err(_) => true, // assume null in absence of schema
                Ok(schema) => NULLISH.satisfies(&schema) != Satisfaction::Not,
            },
        }
    }
}
