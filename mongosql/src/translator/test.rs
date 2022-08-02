macro_rules! test_translate_expression {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::{agg_ir, ir, translator};
            let translator = translator::MqlTranslator::new();
            let expected = $expected;
            let actual = translator.translate_expression($input);
            assert_eq!(actual, expected);
        }
    };
}

mod literal_expression {
    test_translate_expression!(
        null,
        expected = Ok(agg_ir::Expression::Literal(agg_ir::LiteralValue::Null)),
        input = ir::Expression::Literal(ir::LiteralValue::Null.into())
    );
    test_translate_expression!(
        boolean,
        expected = Ok(agg_ir::Expression::Literal(agg_ir::LiteralValue::Boolean(
            true
        ))),
        input = ir::Expression::Literal(ir::LiteralValue::Boolean(true).into())
    );
    test_translate_expression!(
        integer,
        expected = Ok(agg_ir::Expression::Literal(agg_ir::LiteralValue::Integer(
            1
        ))),
        input = ir::Expression::Literal(ir::LiteralValue::Integer(1).into())
    );
    test_translate_expression!(
        string,
        expected = Ok(agg_ir::Expression::Literal(agg_ir::LiteralValue::String(
            "foo".to_string()
        ))),
        input = ir::Expression::Literal(ir::LiteralValue::String("foo".to_string()).into())
    );
    test_translate_expression!(
        long,
        expected = Ok(agg_ir::Expression::Literal(agg_ir::LiteralValue::Long(2))),
        input = ir::Expression::Literal(ir::LiteralValue::Long(2).into())
    );
    test_translate_expression!(
        double,
        expected = Ok(agg_ir::Expression::Literal(agg_ir::LiteralValue::Double(
            3.0
        ))),
        input = ir::Expression::Literal(ir::LiteralValue::Double(3.0).into())
    );
}
