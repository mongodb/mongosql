#[macro_export]
macro_rules! make_cond_expr {
    ($if:expr, $then:expr, $else:expr) => {
        Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Cond,
            args: vec![$if, $then, $else],
        })
    };
}
