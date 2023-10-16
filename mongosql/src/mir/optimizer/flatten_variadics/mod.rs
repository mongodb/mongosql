///
/// Flatten Variadic
///
/// The Flatten Variadic Functions pass will flatten something such as OR(x, OR(y, z)) into
/// OR(x, y, z). This makes passes run after this pass easier to implement since they have
/// fewer equivalent forms to consider.
///
/// Any ScalarFunction that is variadic will be flattened. At this time the variadic functions
/// are:
///
/// * Add
/// * Mul
/// * And
/// * Or
/// * Concat
///
#[cfg(test)]
mod test;

use crate::{
    mir::{
        schema::SchemaInferenceState, visitor::Visitor, Expression, ScalarFunction,
        ScalarFunctionApplication, Stage,
    },
    SchemaCheckingMode,
};

use super::Optimizer;

pub(crate) struct FlattenVariadicFunctionsOptimizer {}

impl Optimizer for FlattenVariadicFunctionsOptimizer {
    fn optimize(&self, st: Stage, _sm: SchemaCheckingMode, _: &SchemaInferenceState) -> Stage {
        FlattenVariadicFunctionsOptimizer::flatten_variadic_functions(st)
    }
}

impl FlattenVariadicFunctionsOptimizer {
    /// Flatten nested binary functions into single, variadic function nodes.
    /// For example, flatten nested binary additions like
    ///         Add                 Add
    ///       /     \            /   |   \
    ///     Add      z   into   x    y    z
    ///   /     \
    /// x         y
    ///
    /// Flattening applies to all associative operators in the mir, including
    /// addition, multiplication, logical disjunction, logical conjunction, and
    /// string concatenation.
    fn flatten_variadic_functions(st: Stage) -> Stage {
        let mut v = ScalarFunctionApplicationVisitor;
        v.visit_stage(st)
    }
}
#[derive(Default)]
struct ScalarFunctionApplicationVisitor;

impl Visitor for ScalarFunctionApplicationVisitor {
    fn visit_scalar_function_application(
        &mut self,
        node: ScalarFunctionApplication,
    ) -> ScalarFunctionApplication {
        let node = node.walk(self);
        match node.function {
            ScalarFunction::Add
            | ScalarFunction::Mul
            | ScalarFunction::And
            | ScalarFunction::Or
            | ScalarFunction::Concat => ScalarFunctionApplication {
                function: node.function,
                is_nullable: node.is_nullable,
                args: node
                    .args
                    .iter()
                    .flat_map(|child| match child {
                        Expression::ScalarFunction(c) if node.function == c.function => {
                            c.args.clone()
                        }
                        _ => vec![child.clone()],
                    })
                    .collect(),
                cache: node.cache,
            },
            _ => node,
        }
    }
}
