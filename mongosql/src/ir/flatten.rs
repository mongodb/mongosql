use crate::ir::{visitor::Visitor, Expression, ScalarFunction, ScalarFunctionApplication, Stage};

#[derive(Default)]
struct ScalarFunctionApplicationVisitor;

/// Flatten nested binary functions into single, variadic function nodes.
/// For example, flatten nested binary additions like
///         Add                 Add
///       /     \            /   |   \
///     Add      z   into   x    y    z
///   /     \
/// x         y
///
/// Flattening applies to all associative operators in the ir, including
/// addition, multiplication, logical disjunction, logical conjunction, and
/// string concatenation.
pub fn flatten_variadic_functions(st: Stage) -> Stage {
    let mut v = ScalarFunctionApplicationVisitor::default();
    v.visit_stage(st)
}

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
