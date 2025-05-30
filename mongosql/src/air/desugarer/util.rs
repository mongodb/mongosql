#[macro_export]
macro_rules! make_cond_expr {
    ($if:expr, $then:expr, $else:expr) => {
        Expression::MqlSemanticOperator(MqlSemanticOperator {
            op: MqlOperator::Cond,
            args: vec![$if, $then, $else],
        })
    };
}
