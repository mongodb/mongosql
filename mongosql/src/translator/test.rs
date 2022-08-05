macro_rules! test_translate_expression {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::translator;
            let translator = translator::MqlTranslator::new();
            let expected = $expected;
            let actual = translator.translate_expression($input);
            assert_eq!(actual, expected);
        }
    };
}

#[allow(unused_macros)]
macro_rules! test_translate_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::{agg_ir, ir, translator};
            let translator = translator::MqlTranslator::new();
            let expected = $expected;
            let actual = translator.translate_stage($input);
            assert_eq!(actual, expected);
        }
    };
}

mod literal_expression {
    use crate::{agg_ir, ir};
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

mod document_expression {
    use crate::unchecked_unique_linked_hash_map;
    use crate::{agg_ir, ir, translator::Error};
    test_translate_expression!(
        empty,
        expected = Ok(agg_ir::Expression::Document(
            unchecked_unique_linked_hash_map! {}
        )),
        input = ir::Expression::Document(unchecked_unique_linked_hash_map! {}.into())
    );
    test_translate_expression!(
        non_empty,
        expected = Ok(agg_ir::Expression::Document(
            unchecked_unique_linked_hash_map! {"foo".to_string() => agg_ir::Expression::Literal(agg_ir::LiteralValue::Integer(1))}
        )),
        input = ir::Expression::Document(
            unchecked_unique_linked_hash_map! {"foo".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),}
        .into())
    );
    test_translate_expression!(
        nested,
        expected = Ok(agg_ir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "foo".to_string() => agg_ir::Expression::Literal(agg_ir::LiteralValue::Integer(1)),
                "bar".to_string() => agg_ir::Expression::Document(unchecked_unique_linked_hash_map!{
                    "baz".to_string() => agg_ir::Expression::Literal(agg_ir::LiteralValue::Integer(2))
                }),
            }
        )),
        input = ir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "foo".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into()),
                "bar".to_string() => ir::Expression::Document(unchecked_unique_linked_hash_map!{
                    "baz".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(2).into())
                }.into()),
            }
            .into()
        )
    );
    test_translate_expression!(
        dollar_prefixed_key_disallowed,
        expected = Err(Error::InvalidDocumentKey("$foo".to_string())),
        input = ir::Expression::Document(
            unchecked_unique_linked_hash_map! {"$foo".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())}.into())
    );
    test_translate_expression!(
        key_containing_dot_disallowed,
        expected = Err(Error::InvalidDocumentKey("foo.bar".to_string())),
        input = ir::Expression::Document(
            unchecked_unique_linked_hash_map! {"foo.bar".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())}.into()
        )
    );
    test_translate_expression!(
        empty_key_disallowed,
        expected = Err(Error::InvalidDocumentKey("".to_string())),
        input = ir::Expression::Document(
            unchecked_unique_linked_hash_map! {"".to_string() => ir::Expression::Literal(ir::LiteralValue::Integer(1).into())}.into())
    );
}

mod array_expression {
    use crate::{agg_ir, ir};
    test_translate_expression!(
        empty,
        expected = Ok(agg_ir::Expression::Array(vec![])),
        input = ir::Expression::Array(vec![].into())
    );
    test_translate_expression!(
        non_empty,
        expected = Ok(agg_ir::Expression::Array(vec![
            agg_ir::Expression::Literal(agg_ir::LiteralValue::String("abc".to_string()))
        ])),
        input = ir::Expression::Array(
            vec![ir::Expression::Literal(
                ir::LiteralValue::String("abc".into()).into()
            )]
            .into()
        )
    );
    test_translate_expression!(
        nested,
        expected = Ok(agg_ir::Expression::Array(vec![
            agg_ir::Expression::Literal(agg_ir::LiteralValue::Null),
            agg_ir::Expression::Array(vec![agg_ir::Expression::Literal(
                agg_ir::LiteralValue::Null
            )])
        ])),
        input = ir::Expression::Array(
            vec![
                ir::Expression::Literal(ir::LiteralValue::Null.into()),
                ir::Expression::Array(
                    vec![ir::Expression::Literal(ir::LiteralValue::Null.into())].into()
                )
            ]
            .into()
        )
    );
}
