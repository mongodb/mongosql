use super::*;

test_algebrize!(
    null,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Literal(mir::LiteralValue::Null)),
    input = ast::Expression::Literal(ast::Literal::Null),
);

test_algebrize!(
    expr_true,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
    input = ast::Expression::Literal(ast::Literal::Boolean(true)),
);

test_algebrize!(
    expr_false,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Literal(mir::LiteralValue::Boolean(false))),
    input = ast::Expression::Literal(ast::Literal::Boolean(false)),
);

test_algebrize!(
    string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
        "hello!".into(),
    ))),
    input = ast::Expression::StringConstructor("hello!".into()),
);

test_algebrize!(
    int,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
    input = ast::Expression::Literal(ast::Literal::Integer(42)),
);

test_algebrize!(
    long,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Literal(mir::LiteralValue::Long(42))),
    input = ast::Expression::Literal(ast::Literal::Long(42)),
);

test_algebrize!(
    double,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Literal(mir::LiteralValue::Double(42f64))),
    input = ast::Expression::Literal(ast::Literal::Double(42f64)),
);

test_algebrize!(
    empty_array,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Array(vec![].into())),
    input = ast::Expression::Array(vec![]),
);

test_algebrize!(
    nested_array,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Array(
        vec![mir::Expression::Array(
            vec![
                mir::Expression::Literal(mir::LiteralValue::Long(42)),
                mir::Expression::Literal(mir::LiteralValue::Integer(42)),
            ]
            .into(),
        )]
        .into(),
    )),
    input = ast::Expression::Array(vec![ast::Expression::Array(vec![
        ast::Expression::Literal(ast::Literal::Long(42)),
        ast::Expression::Literal(ast::Literal::Integer(42)),
    ])]),
);

test_algebrize!(
    empty_document,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Document(
        unchecked_unique_linked_hash_map! {}.into(),
    )),
    input = ast::Expression::Document(multimap! {}),
);

test_algebrize!(
            nested_document,
            method = algebrize_expression,
            in_implicit_type_conversion_context = false,
            expected = Ok(mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "foo2".into() => mir::Expression::Document(
                        unchecked_unique_linked_hash_map!{"nested".into() => mir::Expression::Literal(mir::LiteralValue::Integer(52))} .into(),
                    ),
                    "bar2".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42))
                }.into(),
            )),
            input = ast::Expression::Document(multimap! {
                "foo2".into() => ast::Expression::Document(
                    multimap!{"nested".into() => ast::Expression::Literal(ast::Literal::Integer(52))},
                ),
                "bar2".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            }),
        );

test_algebrize!(
    document_with_keys_containing_dots_and_dollars,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::Document(
        unchecked_unique_linked_hash_map! {
            "a.b".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            "$c".into() => mir::Expression::Literal(mir::LiteralValue::Integer(42)),
        }
        .into(),
    )),
    input = ast::Expression::Document(multimap! {
        "a.b".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
        "$c".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
    }),
);

mod ext_json {
    use super::*;

    test_algebrize!(
        array_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Array(
            vec![
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                mir::Expression::Literal(mir::LiteralValue::String("hello".to_string()))
            ]
            .into(),
        )),
        input = ast::Expression::StringConstructor("[1, \"hello\"]".to_string()),
    );

    test_algebrize!(
        bindata_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Binary(
            bson::Binary {
                subtype: bson::spec::BinarySubtype::Uuid,
                bytes: vec![]
            }
        ))),
        input = ast::Expression::StringConstructor(
            "{ \"$binary\" : {\"base64\" : \"\", \"subType\" : \"04\"}}".to_string()
        ),
    );

    test_algebrize!(
        boolean_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
        input = ast::Expression::StringConstructor("true".to_string()),
    );

    test_algebrize!(
        datetime_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::DateTime(
            "2019-08-11T17:54:14.692Z"
                .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                .unwrap()
                .into(),
        ))),
        input = ast::Expression::StringConstructor(
            "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
        ),
    );

    test_algebrize!(
        decimal128_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Decimal128(
            "10.99".parse().unwrap()
        ))),
        input = ast::Expression::StringConstructor("{\"$numberDecimal\": \"10.99\"}".to_string()),
    );

    test_algebrize!(
                document_string_to_ext_json,
                method = algebrize_expression,
                in_implicit_type_conversion_context = true,
                expected = Ok(mir::Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "x".into() => mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                        "y".into() => mir::Expression::Literal(mir::LiteralValue::String("hello".to_string())),
                    }
                .into())),
                input = ast::Expression::StringConstructor("{\"x\": 3, \"y\": \"hello\"}".to_string()),
            );

    test_algebrize!(
        double_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Double(10.5))),
        input = ast::Expression::StringConstructor("10.5".to_string()),
    );

    test_algebrize!(
        int_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Long(3))),
        input = ast::Expression::StringConstructor("{\"$numberLong\": \"3\"}".to_string()),
    );

    test_algebrize!(
        javascript_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::JavaScriptCode(
            "code here".to_string()
        ))),
        input = ast::Expression::StringConstructor("{\"$code\": \"code here\"}".to_string()),
    );

    test_algebrize!(
        javascript_with_scope_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(
            mir::LiteralValue::JavaScriptCodeWithScope(bson::JavaScriptCodeWithScope {
                code: "code here".to_string(),
                scope: bson::doc! {}
            })
        )),
        input = ast::Expression::StringConstructor(
            "{\"$code\": \"code here\", \"$scope\": {}}".to_string()
        ),
    );

    test_algebrize!(
        maxkey_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::MaxKey)),
        input = ast::Expression::StringConstructor("{\"$maxKey\": 1}".to_string()),
    );

    test_algebrize!(
        minkey_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::MinKey)),
        input = ast::Expression::StringConstructor("{\"$minKey\": 1}".to_string()),
    );

    test_algebrize!(
        null_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Null)),
        input = ast::Expression::StringConstructor("null".to_string()),
    );

    test_algebrize!(
        objectid_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::ObjectId(
            bson::oid::ObjectId::parse_str("5ab9c3da31c2ab715d421285").unwrap()
        ))),
        input = ast::Expression::StringConstructor(
            "{\"$oid\": \"5ab9c3da31c2ab715d421285\"}".to_string()
        ),
    );

    test_algebrize!(
        regular_expression_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(
            mir::LiteralValue::RegularExpression(bson::Regex {
                pattern: "pattern".to_string(),
                options: "".to_string()
            })
        )),
        input = ast::Expression::StringConstructor(
            "{\"$regularExpression\":{\"pattern\": \"pattern\",\"options\": \"\"}}".to_string()
        ),
    );

    test_algebrize!(
        regular_string_stays_string_with_implicit_casting_true,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
            "abc".to_string()
        ))),
        input = ast::Expression::StringConstructor("abc".to_string()),
    );

    test_algebrize!(
        json_string_stays_string_with_implicit_casting_true,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
            "'{this_doc_is_actually_a_string: 1}'".to_string()
        ))),
        input =
            ast::Expression::StringConstructor("'{this_doc_is_actually_a_string: 1}'".to_string()),
    );

    test_algebrize!(
        symbol_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Symbol(
            "symbol".to_string()
        ))),
        input = ast::Expression::StringConstructor("{\"$symbol\": \"symbol\"}".to_string()),
    );

    test_algebrize!(
        timestamp_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Timestamp(
            bson::Timestamp {
                time: 1,
                increment: 2
            }
        ))),
        input = ast::Expression::StringConstructor(
            "{\"$timestamp\": {\"t\": 1, \"i\": 2}}".to_string()
        ),
    );

    test_algebrize!(
        undefined_string_to_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = true,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::Undefined)),
        input = ast::Expression::StringConstructor("{\"$undefined\": true}".to_string()),
    );

    test_algebrize!(
        ext_json_not_converted_when_conversion_false,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::Literal(mir::LiteralValue::String(
            "{\"$numberLong\": \"3\"}".to_string()
        ))),
        input = ast::Expression::StringConstructor("{\"$numberLong\": \"3\"}".to_string()),
    );
}
