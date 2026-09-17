#[macro_export]
macro_rules! make_cond_expr {
    ($if:expr_2021, $then:expr_2021, $else:expr_2021) => {
        Expression::MqlSemanticOperator(MqlSemanticOperator {
            op: MqlOperator::Cond,
            args: vec![$if, $then, $else],
        })
    };
}
