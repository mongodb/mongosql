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
        schema::{SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        Derived, ExistsExpr, Expression, FieldAccess, FieldExistence, Filter, LateralJoin,
        MQLExpression, ScalarFunction, ScalarFunctionApplication, Stage, SubqueryExpr,
    },
    SchemaCheckingMode,
};
use std::collections::BTreeMap;

pub(crate) struct MatchNullFilteringOptimizer;

impl Optimizer for MatchNullFilteringOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        _schema_state: &SchemaInferenceState,
    ) -> Stage {
        let mut v = MatchNullFilteringVisitor { scope: 0 };
        v.visit_stage(st)
    }
}

struct MatchNullFilteringVisitor {
    scope: u16,
}

impl MatchNullFilteringVisitor {
    /// create_null_filter_stage attempts to create a Filter stage that ensures
    /// possibly null FieldAccesses in the original_filter exist. If the
    /// original_filter does not contain any null-semantic operators or does
    /// not contain any nullable FieldAccesses, this method returns None.
    fn create_null_filter_stage(&self, original_filter: Filter) -> (Box<Stage>, Expression) {
        let Filter {
            source: original_source,
            condition: original_condition,
            ..
        } = original_filter;
        let (opt_cond, original_cond) = self.generate_null_filter_condition(original_condition);
        (
            match opt_cond {
                Some(opt_cond) => Box::new(Stage::Filter(Filter {
                    source: original_source,
                    condition: opt_cond,
                    cache: SchemaCache::new(),
                })),
                None => original_source,
            },
            original_cond,
        )
    }

    /// generate_null_filter_condition attempts to create an Expression that
    /// ensures possibly null FieldAccesses in the condition exist. If the
    /// condition does not contain any null-semantic operators or does not
    /// contain any nullable FieldAccesses, this method returns None.
    fn generate_null_filter_condition(
        &self,
        condition: Expression,
    ) -> (Option<Expression>, Expression) {
        let (fields, condition) = self.gather_fields_for_null_filter(condition);
        let mut optimized_exists_ops = fields
            .into_values()
            .map(|fa| {
                Expression::MQLIntrinsic(MQLExpression::FieldExistence(FieldExistence {
                    field_access: fa,
                    cache: SchemaCache::new(),
                }))
            })
            .collect::<Vec<Expression>>();

        match optimized_exists_ops.len() {
            0 => (None, condition),
            1 => (Some(optimized_exists_ops.swap_remove(0)), condition),
            _ => (
                Some(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::And,
                    is_nullable: false,
                    args: optimized_exists_ops,
                    cache: SchemaCache::new(),
                })),
                condition,
            ),
        }
    }

    /// gather_fields_for_null_filter collects unique "pure" FieldAccesses
    /// in the provided condition that appear as descendents of sql-semantic
    /// operators.
    fn gather_fields_for_null_filter(
        &self,
        condition: Expression,
    ) -> (BTreeMap<String, FieldAccess>, Expression) {
        let mut visitor = NullableFieldAccessGatherer {
            filter_fields: BTreeMap::new(),
            is_collecting: false,
            found_impure: false,
            scope: self.scope,
        };
        let condition = visitor.visit_expression(condition);
        (visitor.filter_fields, condition)
    }
}

impl Visitor for MatchNullFilteringVisitor {
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
        let (null_filter_stage, condition) = self.create_null_filter_stage(node);
        Filter {
            source: null_filter_stage,
            condition,
            cache: SchemaCache::new(),
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
            is_nullable: node.is_nullable,
            output_expr: node.output_expr,
            subquery: Box::new(node.subquery.walk(self)),
            cache: node.cache,
        };
        self.scope -= 1;
        node
    }
}

struct NullableSetter {}

impl Visitor for NullableSetter {
    // Do not recurse down subqueries
    fn visit_subquery_expr(&mut self, node: SubqueryExpr) -> SubqueryExpr {
        node
    }

    // Do not recurse down exists exprs
    fn visit_exists_expr(&mut self, node: ExistsExpr) -> ExistsExpr {
        node
    }

    fn visit_scalar_function_application(
        &mut self,
        node: ScalarFunctionApplication,
    ) -> ScalarFunctionApplication {
        let mut node = node.walk(self);
        // In the case of recusive ScalarFunctions, this will get set twice.
        // It is worth it for the improvement in code clarity.
        node.is_nullable = false;
        node
    }

    fn visit_expression(&mut self, node: Expression) -> Expression {
        let mut node = node.walk(self);
        node.set_is_nullable(false);
        node
    }
}

struct NullableFieldAccessGatherer {
    filter_fields: BTreeMap<String, FieldAccess>,
    is_collecting: bool,
    found_impure: bool,
    scope: u16,
}

impl Visitor for NullableFieldAccessGatherer {
    // Do not walk stages nested within expressions.
    fn visit_stage(&mut self, node: Stage) -> Stage {
        node
    }

    fn visit_expression(&mut self, node: Expression) -> Expression {
        match node {
            // If we encounter a "nullable" scalar function, meaning a scalar
            // function with SQL-style semantics, then we want to collect
            // nullable fields nested within that function's arguments.
            Expression::ScalarFunction(sf)
                if sf.is_nullable && sf.function.mql_null_semantics_diverge() =>
            {
                let old_is_collecting = self.is_collecting;
                let old_found_impure = self.found_impure;
                self.is_collecting = true;
                self.found_impure = false;
                let mut sf = sf.walk(self);
                // If no impure functions were found, we can set all the semantics
                // to MQL in the expression.
                if !self.found_impure {
                    let mut nullable_setter = NullableSetter {};
                    sf = nullable_setter.visit_scalar_function_application(sf);
                }
                self.is_collecting = old_is_collecting;
                self.found_impure = old_found_impure;

                Expression::ScalarFunction(sf)
            }
            Expression::FieldAccess(fa) => {
                let scope = fa.scope_if_pure();
                if self.is_collecting && scope == Some(self.scope) && fa.is_nullable {
                    // Only store nullable "pure" fields in the map.
                    match fa.to_string_if_pure() {
                        None => {
                            self.found_impure = true;
                        }
                        Some(s) => {
                            self.filter_fields.insert(s, fa.clone());
                        }
                    }
                }

                Expression::FieldAccess(fa)
            }
            _ => node.walk(self),
        }
    }
}

impl ScalarFunction {
    // This method defines those functions that have MQL NULL behavior that is divergent from SQL.
    // That is to say, in SQL the function will return NULL if any argument is NULL, but in MQL it
    // will not. For instance, NULL = 1 and NULL = NULL return false and true, respectively, in
    // MQL.
    pub fn mql_null_semantics_diverge(&self) -> bool {
        match self {
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

            // These functions all correctly return NULL in MQL, if there is a NULL argument.
            // NullIf is weird in that it can also return NULL if none of the arguments are NULL,
            // but that does not affect this optimization.
            // We include every variant instead of a wildcard in case new variants are added in the
            // future that could benefit
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
            | ScalarFunction::Millisecond
            | ScalarFunction::Week
            | ScalarFunction::DayOfWeek
            | ScalarFunction::DayOfYear
            | ScalarFunction::IsoWeek
            | ScalarFunction::IsoWeekday
            | ScalarFunction::MergeObjects => false,
        }
    }
}
