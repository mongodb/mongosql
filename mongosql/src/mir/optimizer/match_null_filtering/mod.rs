///
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
///
#[cfg(test)]
mod test;

use crate::{
    mir::{
        optimizer::Optimizer,
        schema::{CachedSchema, SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Derived, ExistsExpr, Expression, FieldAccess, FieldExistence, Filter, LateralJoin,
        MQLExpression, ScalarFunction, ScalarFunctionApplication, Stage, SubqueryExpr,
    },
    schema::{Satisfaction, NULLISH},
    SchemaCheckingMode,
};
use std::collections::BTreeMap;

pub(crate) struct MatchNullFilteringOptimizer;

impl Optimizer for MatchNullFilteringOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        schema_state: &SchemaInferenceState,
    ) -> Stage {
        let mut v = MatchNullFilteringVisitor {
            schema_state: schema_state.clone(),
            scope: 0,
            changed: false,
        };
        let st = v.visit_stage(st);

        // After adding these new Filter stages, we invalidate all of the
        // schema caches and recalculate. This is necessary since the new
        // Filter's FieldExistence expressions alter the schema of their
        // field references in subsequent stages. We cannot just call the
        // `schema` method on a per-expression or even per-stage basis
        // _during_ the walk because that would fail to incorporate
        // correlated schema information.
        if v.changed {
            let mut v = SchemaInvalidator;
            let st = v.visit_stage(st);
            let _ = st.schema(schema_state);
            st
        } else {
            st
        }
    }
}

struct MatchNullFilteringVisitor<'a> {
    schema_state: SchemaInferenceState<'a>,
    scope: u16,
    changed: bool,
}

impl<'a> MatchNullFilteringVisitor<'a> {
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
                Expression::MQLIntrinsic(MQLExpression::FieldExistence(FieldExistence {
                    field_access: fa,
                    cache: SchemaCache::new(),
                }))
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
            state: self.schema_state.clone(),
            scope: self.scope,
        };
        visitor.visit_expression(condition.clone());
        visitor.filter_fields
    }
}

impl<'a> Visitor for MatchNullFilteringVisitor<'a> {
    fn visit_derived(&mut self, node: Derived) -> Derived {
        self.scope += 1;
        let node = node.walk(self);
        self.scope -= 1;
        node
    }

    // SQL-1683 We skip applying this optimization to LateralJoins since the
    // correlated fields should not be filtered within the RHS of the Join.
    fn visit_lateral_join(&mut self, node: LateralJoin) -> LateralJoin {
        node
    }

    fn visit_filter(&mut self, node: Filter) -> Filter {
        let node = node.walk(self);
        match self.create_null_filter_stage(&node) {
            None => node,
            Some(null_filter_stage) => {
                self.changed = true;
                Filter {
                    source: Box::new(Stage::Filter(null_filter_stage)),
                    condition: node.condition,
                    cache: SchemaCache::new(),
                }
            }
        }
    }

    fn visit_exists_expr(&mut self, node: ExistsExpr) -> ExistsExpr {
        self.scope += 1;
        let node = node.walk(self);
        self.scope -= 1;
        node
    }

    fn visit_subquery_expr(&mut self, node: SubqueryExpr) -> SubqueryExpr {
        self.scope += 1;
        let node = SubqueryExpr {
            // do not walk subquery output_exprs
            output_expr: node.output_expr,
            subquery: Box::new(node.subquery.walk(self)),
            cache: node.cache,
        };
        self.scope -= 1;
        node
    }
}

struct NullableFieldAccessGatherer<'a> {
    filter_fields: BTreeMap<String, FieldAccess>,
    is_collecting: bool,
    state: SchemaInferenceState<'a>,
    scope: u16,
}

impl<'a> Visitor for NullableFieldAccessGatherer<'a> {
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
                let scope = fa.scope_if_pure();
                if self.is_collecting
                    && scope == Some(self.scope)
                    && self.schema_is_nullish(node.clone())
                {
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

impl<'a> NullableFieldAccessGatherer<'a> {
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
        match expr.schema(&self.state) {
            Err(_) => true, // assume null in absence of schema
            Ok(schema) => NULLISH.satisfies(&schema) != Satisfaction::Not,
        }
    }
}

struct SchemaInvalidator;

impl Visitor for SchemaInvalidator {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        node.invalidate_cache();
        node.walk(self)
    }

    fn visit_expression(&mut self, node: Expression) -> Expression {
        node.invalidate_cache();
        node.walk(self)
    }
}
