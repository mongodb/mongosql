use crate::{
    ast, map,
    mir::{self, binding_tuple::Key},
    multimap,
    schema::{
        Atomic, Document, Schema, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, NUMERIC_OR_NULLISH,
        STRING_OR_NULLISH,
    },
    set, test_algebrize, test_algebrize_expr_and_schema_check, unchecked_unique_linked_hash_map,
    usererror::UserError,
};

pub mod binary;
pub mod identifier_and_subpath;
pub mod literal;

mod unary {
    use super::*;

    test_algebrize!(
        neg_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Neg,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Neg,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize_expr_and_schema_check!(
        neg_wrong_type,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Neg",
            required: NUMERIC_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Boolean).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Neg,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
        }),
    );

    test_algebrize!(
        pos_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Pos,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Pos,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize_expr_and_schema_check!(
        pos_wrong_type,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Pos",
            required: NUMERIC_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Boolean).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Pos,
            expr: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
        }),
    );

    test_algebrize!(
        unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Neg,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(42)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Unary(ast::UnaryExpr {
            op: ast::UnaryOp::Neg,
            expr: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"42\"}".to_string()
            )),
        }),
    );
}

mod between {
    use super::*;

    test_algebrize!(
        simple,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(3)),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(1))),
            min: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
            max: Box::new(ast::Expression::Literal(ast::Literal::Integer(3))),
        }),
    );

    test_algebrize!(
        does_not_convert_ext_json_if_all_are_strings,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
        }),
    );

    test_algebrize!(
        convert_arg_and_min_ext_json_if_max_is_non_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
            max: Box::new(ast::Expression::Identifier("a".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        do_not_convert_arg_or_min_ext_json_if_max_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"2\"}".to_string()
                    )),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
            max: Box::new(ast::Expression::Identifier("a".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_arg_and_max_ext_json_if_min_is_non_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::Identifier("a".to_string())),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        do_not_convert_arg_or_max_ext_json_if_min_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"2\"}".to_string()
                    )),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::Identifier("a".to_string())),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_min_and_max_ext_json_if_arg_is_non_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        do_not_convert_min_or_max_ext_json_if_arg_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"2\"}".to_string()
                    )),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"2\"}".to_string()
            )),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_arg_ext_json_if_neither_min_nor_max_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::Identifier("a".to_string())),
            max: Box::new(ast::Expression::Identifier("b".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_arg_ext_json_if_one_of_min_or_max_is_non_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::Identifier("a".to_string())),
            max: Box::new(ast::Expression::Identifier("b".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        do_not_convert_arg_ext_json_if_both_min_and_max_are_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            min: Box::new(ast::Expression::Identifier("a".to_string())),
            max: Box::new(ast::Expression::Identifier("b".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                    "b".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_min_ext_json_if_neither_arg_nor_max_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            max: Box::new(ast::Expression::Identifier("b".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_min_ext_json_if_one_of_arg_or_max_is_non_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            max: Box::new(ast::Expression::Identifier("b".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        do_not_convert_min_ext_json_if_both_arg_and_max_are_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
            max: Box::new(ast::Expression::Identifier("b".to_string())),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                    "b".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_max_ext_json_if_neither_arg_nor_min_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::Identifier("b".to_string())),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        convert_max_ext_json_if_one_of_arg_or_min_is_non_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::Identifier("b".to_string())),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );

    test_algebrize!(
        do_not_convert_max_ext_json_if_both_arg_and_min_are_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Between,
                args: vec![
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "a".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: true,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Between(ast::BetweenExpr {
            arg: Box::new(ast::Expression::Identifier("a".to_string())),
            min: Box::new(ast::Expression::Identifier("b".to_string())),
            max: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )),
        }),
        env = map! {
            ("foo", 1u16).into() => Schema::Document( Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::String),
                    "b".into() => Schema::Atomic(Atomic::String),
                },
                required: set!{},
                additional_properties: false,
                ..Default::default()
            }),
        },
    );
}

mod scalar_function {
    use super::*;

    test_algebrize!(
        standard_scalar_function,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Lower,
                args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                    "hello".into(),
                )),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Lower,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "hello".into(),
            )]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        lower_scalar_function_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Lower,
                args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"1\"}".into(),
                )),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Lower,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".into(),
            )]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        replace,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Replace,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Replace,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor(" hello world ".to_string()),
                ast::Expression::StringConstructor("wo".to_string()),
                ast::Expression::StringConstructor("wowow".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        replace_null_one,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Replace,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Null),
                    mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Replace,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Null),
                ast::Expression::StringConstructor("wo".to_string()),
                ast::Expression::StringConstructor("wowow".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        replace_null_two,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Replace,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                    mir::Expression::Literal(mir::LiteralValue::Null),
                    mir::Expression::Literal(mir::LiteralValue::String("wowow".into())),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Replace,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor(" hello world ".to_string()),
                ast::Expression::Literal(ast::Literal::Null),
                ast::Expression::StringConstructor("wowow".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        replace_null_three,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Replace,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("wo".into())),
                    mir::Expression::Literal(mir::LiteralValue::Null),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Replace,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor(" hello world ".to_string()),
                ast::Expression::StringConstructor("wo".to_string()),
                ast::Expression::Literal(ast::Literal::Null),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        replace_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Replace,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"100\"}".into(),
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"100\"}".into(),
                    )),
                    mir::Expression::Literal(mir::LiteralValue::Null),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Replace,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"100\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"100\"}".to_string()),
                ast::Expression::Literal(ast::Literal::Null),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize_expr_and_schema_check!(
        replace_args_must_be_string_or_null,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Replace",
            required: STRING_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Replace,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(42)),
                ast::Expression::Literal(ast::Literal::Integer(42)),
                ast::Expression::Literal(ast::Literal::Integer(42)),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        log_bin_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Log,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(100)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Log,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(100)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        log_bin_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Log,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(100)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Log,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"100\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        round_bin_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Round,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Round,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(10)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        round_bin_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Round,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Round,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        cos_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Cos,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Cos,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        cos_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Cos,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Cos,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"10\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        sin_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Sin,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Sin,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        sin_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Sin,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Sin,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"10\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        tan_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Tan,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Tan,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        tan_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Tan,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Tan,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"10\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        radians_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Radians,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Radians,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        radians_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Radians,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Radians,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        sqrt_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Sqrt,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(4)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Sqrt,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(4)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        sqrt_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Sqrt,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(4)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Sqrt,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"4\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        abs_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Abs,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Abs,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(10)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        abs_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Abs,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(10)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Abs,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"10\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        ceil_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Ceil,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Ceil,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Double(1.5)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        ceil_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Ceil,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Ceil,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberDouble\": \"1.5\"}".to_string()
            )]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        degrees_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Degrees,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Degrees,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        degrees_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Degrees,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Degrees,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            )]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        floor_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Floor,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Floor,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Double(1.5)
            ),]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        floor_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Floor,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Double(1.5)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Floor,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberDouble\": \"1.5\"}".to_string()
            )]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        mod_bin_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Mod,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Mod,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(10)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        mod_bin_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Mod,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Mod,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string())
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        pow_bin_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Pow,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Pow,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(10)),
                ast::Expression::Literal(ast::Literal::Integer(10)),
            ]),
            set_quantifier: Some(ast::SetQuantifier::All),
        }),
    );

    test_algebrize!(
        pow_bin_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Pow,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(10)),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Pow,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"10\"}".to_string())
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        split_only_implicit_converts_third_arg,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Split,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Split,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        substring_implicit_converts_last_two_args,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Substring,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Substring,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        nullif_two_strings_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::NullIf,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\": \"1\"}".to_string()
                    )),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::NullIf,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        nullif_first_arg_string_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::NullIf,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::NullIf,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
                ast::Expression::Literal(ast::Literal::Integer(1)),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        nullif_second_arg_string_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::NullIf,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                ],
                is_nullable: true,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::NullIf,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::StringConstructor("{\"$numberInt\": \"1\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        coalesce,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication::new(
                mir::ScalarFunction::Coalesce,
                vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                ],
            )
        ),),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Coalesce,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
                ast::Expression::Literal(ast::Literal::Integer(2)),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        coalesce_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication::new(
                mir::ScalarFunction::Coalesce,
                vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(1)),
                    mir::Expression::Literal(mir::LiteralValue::Integer(2)),
                ],
            )
        ),),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Coalesce,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\":\"1\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\":\"2\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        size_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Size,
                args: vec![mir::Expression::Array(
                    vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
                ),],
                is_nullable: false,
            }
        ),),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Size,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Array(vec![
                ast::Expression::Literal(ast::Literal::Integer(1)),
            ])]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        size_unary_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Size,
                args: vec![mir::Expression::Array(
                    vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
                ),],
                is_nullable: false,
            }
        ),),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Size,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "[1]".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        slice_bin_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Slice,
                args: vec![
                    mir::Expression::Array(
                        vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
                    ),
                    mir::Expression::Literal(mir::LiteralValue::Integer(0)),
                ],

                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Slice,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Array(vec![ast::Expression::Literal(ast::Literal::Integer(1)),]),
                ast::Expression::Literal(ast::Literal::Integer(0)),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        slice_bin_op_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Slice,
                args: vec![
                    mir::Expression::Array(
                        vec![mir::Expression::Literal(mir::LiteralValue::Integer(1))].into(),
                    ),
                    mir::Expression::Literal(mir::LiteralValue::Integer(0)),
                ],

                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Slice,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("[1]".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\":\"0\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        day_of_week,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::DayOfWeek,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::DayOfWeek,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        day_of_week_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::DayOfWeek,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::DayOfWeek,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\":\"1\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        bit_length_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::BitLength,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::BitLength,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        bit_length_unary_op_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::BitLength,
                args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\":\"1\"}".to_string()
                )),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::BitLength,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\":\"1\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        char_length_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::CharLength,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CharLength,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        char_length_unary_op_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::CharLength,
                args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\":\"1\"}".to_string()
                )),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::CharLength,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\":\"1\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        octet_length_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::OctetLength,
                args: vec![mir::Expression::Literal(mir::LiteralValue::Integer(1)),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::OctetLength,
            args: ast::FunctionArguments::Args(vec![ast::Expression::Literal(
                ast::Literal::Integer(1)
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        octet_length_unary_op_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::OctetLength,
                args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\":\"1\"}".to_string()
                )),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::OctetLength,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\":\"1\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        position_binary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Position,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String("hello".to_string())),
                    mir::Expression::Literal(mir::LiteralValue::String("world".to_string())),
                ],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Position,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("hello".to_string()),
                ast::Expression::StringConstructor("world".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        position_binary_op_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Position,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\":\"1\"}".to_string()
                    )),
                    mir::Expression::Literal(mir::LiteralValue::String(
                        "{\"$numberInt\":\"2\"}".to_string()
                    )),
                ],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Position,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::StringConstructor("{\"$numberInt\":\"1\"}".to_string()),
                ast::Expression::StringConstructor("{\"$numberInt\":\"2\"}".to_string()),
            ]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        upper_unary_op,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Upper,
                args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                    "hello".to_string()
                )),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Upper,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "hello".to_string()
            ),]),
            set_quantifier: None,
        }),
    );

    test_algebrize!(
        upper_unary_op_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Upper,
                args: vec![mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\":\"1\"}".to_string()
                )),],
                is_nullable: false
            }
        )),
        input = ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Upper,
            args: ast::FunctionArguments::Args(vec![ast::Expression::StringConstructor(
                "{\"$numberInt\":\"1\"}".to_string()
            ),]),
            set_quantifier: None,
        }),
    );
}

mod trim {
    use super::*;

    test_algebrize!(
        ltrim,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::LTrim,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String("hello".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("hello world".into()))
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Leading,
            trim_chars: Box::new(ast::Expression::StringConstructor("hello".into())),
            arg: Box::new(ast::Expression::StringConstructor("hello world".into())),
        }),
    );

    test_algebrize!(
        rtrim,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::RTrim,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String("world".into())),
                    mir::Expression::Literal(mir::LiteralValue::String("hello world".into()))
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Trailing,
            trim_chars: Box::new(ast::Expression::StringConstructor("world".into())),
            arg: Box::new(ast::Expression::StringConstructor("hello world".into())),
        }),
    );

    test_algebrize!(
        btrim,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::BTrim,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::String(" ".into())),
                    mir::Expression::Literal(mir::LiteralValue::String(" hello world ".into()))
                ],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::StringConstructor(" ".into())),
            arg: Box::new(ast::Expression::StringConstructor(" hello world ".into())),
        }),
    );

    test_algebrize_expr_and_schema_check!(
        trim_arg_must_be_string_or_null,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "BTrim",
            required: STRING_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::StringConstructor(" ".into())),
            arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize_expr_and_schema_check!(
        trim_escape_must_be_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "BTrim",
            required: STRING_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Trim(ast::TrimExpr {
            trim_spec: ast::TrimSpec::Both,
            trim_chars: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            arg: Box::new(ast::Expression::StringConstructor(" ".into())),
        }),
    );
}

mod extract {
    use super::*;

    test_algebrize!(
        extract_year,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Year,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Year,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_month,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Month,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Month,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_day,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Day,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Day,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_hour,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Hour,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Hour,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_minute,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Minute,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Minute,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_second,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Second,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Second,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_millsecond,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Millisecond,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Millisecond,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_day_of_year,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::DayOfYear,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::DayOfYear,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_iso_week,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::IsoWeek,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::IsoWeek,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_day_of_week,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::DayOfWeek,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::DayOfWeek,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize!(
        extract_iso_weekday,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::IsoWeekday,
                args: vec![mir::Expression::ScalarFunction(
                    mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }
                ),],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::IsoWeekday,
            arg: Box::new(ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::CurrentTimestamp,
                args: ast::FunctionArguments::Args(vec![]),
                set_quantifier: Some(ast::SetQuantifier::All)
            })),
        }),
    );

    test_algebrize_expr_and_schema_check!(
        extract_must_be_date,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Second",
            required: DATE_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Second,
            arg: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
        }),
    );

    test_algebrize!(
        extract_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::Year,
                args: vec![mir::Expression::Literal(mir::LiteralValue::DateTime(
                    "2019-08-11T17:54:14.692Z"
                        .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                        .unwrap()
                        .into(),
                ))],
                is_nullable: false,
            }
        )),
        input = ast::Expression::Extract(ast::ExtractExpr {
            extract_spec: ast::DatePart::Year,
            arg: Box::new(ast::Expression::StringConstructor(
                "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
            )),
        }),
    );
}

mod date_function {
    use super::*;

    test_algebrize!(
        dateadd,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::DateFunction(
            mir::DateFunctionApplication {
                function: mir::DateFunction::Add,
                is_nullable: false,
                date_part: mir::DatePart::Quarter,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(5),),
                    mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }),
                ],
            }
        )),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Add,
            date_part: ast::DatePart::Quarter,
            args: vec![
                ast::Expression::Literal(ast::Literal::Integer(5)),
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                })
            ],
        }),
    );

    test_algebrize!(
        datediff,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::DateFunction(
            mir::DateFunctionApplication {
                function: mir::DateFunction::Diff,
                is_nullable: false,
                date_part: mir::DatePart::Week,
                args: vec![
                    mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }),
                    mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::String("sunday".to_string()),)
                ],
            }
        )),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Diff,
            date_part: ast::DatePart::Week,
            args: vec![
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                }),
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                }),
                ast::Expression::StringConstructor("sunday".to_string()),
            ],
        }),
    );

    test_algebrize!(
        datetrunc,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::DateFunction(
            mir::DateFunctionApplication {
                function: mir::DateFunction::Trunc,
                is_nullable: false,
                date_part: mir::DatePart::Year,
                args: vec![
                    mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                        function: mir::ScalarFunction::CurrentTimestamp,
                        args: vec![],
                        is_nullable: false,
                    }),
                    mir::Expression::Literal(mir::LiteralValue::String("sunday".to_string()),)
                ],
            }
        )),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Trunc,
            date_part: ast::DatePart::Year,
            args: vec![
                ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::CurrentTimestamp,
                    args: ast::FunctionArguments::Args(vec![]),
                    set_quantifier: Some(ast::SetQuantifier::All)
                }),
                ast::Expression::StringConstructor("sunday".to_string()),
            ],
        }),
    );

    test_algebrize!(
        date_function_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::DateFunction(
            mir::DateFunctionApplication {
                function: mir::DateFunction::Add,
                is_nullable: false,
                date_part: mir::DatePart::Quarter,
                args: vec![
                    mir::Expression::Literal(mir::LiteralValue::Integer(5),),
                    mir::Expression::Literal(mir::LiteralValue::DateTime(
                        "2019-08-11T17:54:14.692Z"
                            .parse::<chrono::DateTime<chrono::prelude::Utc>>()
                            .unwrap()
                            .into(),
                    ))
                ],
            }
        )),
        input = ast::Expression::DateFunction(ast::DateFunctionExpr {
            function: ast::DateFunctionName::Add,
            date_part: ast::DatePart::Quarter,
            args: vec![
                ast::Expression::StringConstructor("{\"$numberInt\": \"5\"}".to_string()),
                ast::Expression::StringConstructor(
                    "{\"$date\":\"2019-08-11T17:54:14.692Z\"}".to_string()
                )
            ],
        }),
    );
}

mod case {
    use super::*;

    test_algebrize!(
        searched_case,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "bar".into(),
                ))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "foo".into(),
            ))),
            is_nullable: false,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                then: Box::new(ast::Expression::StringConstructor("bar".into())),
            }],
            else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
        }),
    );

    test_algebrize!(
        searched_case_no_else,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "bar".into(),
                ))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                then: Box::new(ast::Expression::StringConstructor("bar".into())),
            }],
            else_branch: None,
        }),
    );

    test_algebrize_expr_and_schema_check!(
        searched_case_when_condition_is_not_bool,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "SearchedCase",
            required: BOOLEAN_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::String).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::StringConstructor("foo".into())),
                then: Box::new(ast::Expression::StringConstructor("bar".into())),
            }],
            else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
        }),
    );

    test_algebrize!(
        searched_case_nullable_then_expression_is_nullable,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: true,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "foo".into(),
            ))),
            is_nullable: true,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                then: Box::new(ast::Expression::Literal(ast::Literal::Null)),
            }],
            else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
        }),
    );

    test_algebrize!(
        searched_case_nullable_else_expression_is_nullable,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Boolean(true))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: true,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Boolean(true))),
                then: Box::new(ast::Expression::Literal(ast::Literal::Null)),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Null))),
        }),
    );

    test_algebrize!(
        simple_case,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "bar".into(),
                ))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "foo".into(),
            ))),
            is_nullable: false,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                then: Box::new(ast::Expression::StringConstructor("bar".into())),
            }],
            else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
        }),
    );

    test_algebrize!(
        simple_case_no_else,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "bar".into(),
                ))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                then: Box::new(ast::Expression::StringConstructor("bar".into())),
            }],
            else_branch: None,
        }),
    );

    test_algebrize_expr_and_schema_check!(
        simple_case_operand_and_when_operand_not_comparable,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(
            mir::schema::Error::InvalidComparison(
                "SimpleCase",
                Schema::Atomic(Atomic::Integer).into(),
                Schema::Atomic(Atomic::String).into(),
            )
        )),
        expected_error_code = 1005,
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::StringConstructor("foo".into())),
                then: Box::new(ast::Expression::StringConstructor("bar".into())),
            }],
            else_branch: Some(Box::new(ast::Expression::StringConstructor("baz".into()))),
        }),
    );

    test_algebrize!(
        simple_case_nullable_then_expression_is_nullable,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable: true,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "foo".into(),
            ))),
            is_nullable: true,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                then: Box::new(ast::Expression::Literal(ast::Literal::Null)),
            }],
            else_branch: Some(Box::new(ast::Expression::StringConstructor("foo".into()))),
        }),
    );

    test_algebrize!(
        simple_case_nullable_else_expression_is_nullable,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "bar".into(),
                ))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::Literal(ast::Literal::Integer(2))),
                then: Box::new(ast::Expression::StringConstructor("bar".into())),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Null))),
        }),
    );

    test_algebrize!(
        simple_case_expr_not_string_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(2))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"2\"}".to_string()
                ))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"3\"}".to_string()
            ))),
            is_nullable: false,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"2\"}".to_string()
                )),
                then: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"2\"}".to_string()
                )),
            }],
            else_branch: Some(Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"3\"}".to_string()
            ))),
        }),
    );

    test_algebrize!(
        simple_case_string_expr_does_not_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "hello".to_string()
            ))),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"2\"}".to_string()
                ))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(3))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(4))),
            is_nullable: false,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::StringConstructor(
                "hello".to_string()
            ))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"2\"}".to_string()
                )),
                then: Box::new(ast::Expression::Literal(ast::Literal::Integer(3))),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(4)))),
        }),
    );

    test_algebrize!(
        simple_case_where_all_strings_does_not_implicit_convert_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"1\"}".to_string()
            ))),
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"2\"}".to_string()
                ))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(3))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(4))),
            is_nullable: false,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: Some(Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"1\"}".to_string()
            ))),
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"2\"}".to_string()
                )),
                then: Box::new(ast::Expression::Literal(ast::Literal::Integer(3))),
            }],
            else_branch: Some(Box::new(ast::Expression::Literal(ast::Literal::Integer(4)))),
        }),
    );

    test_algebrize!(
        searched_case_implicit_converts_when_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
            when_branch: vec![mir::WhenBranch {
                when: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(1))),
                then: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                    "{\"$numberInt\": \"2\"}".to_string()
                ))),
                is_nullable: false,
            }],
            else_branch: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        })),
        input = ast::Expression::Case(ast::CaseExpr {
            expr: None,
            when_branch: vec![ast::WhenBranch {
                when: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"1\"}".to_string()
                )),
                then: Box::new(ast::Expression::StringConstructor(
                    "{\"$numberInt\": \"2\"}".to_string()
                )),
            }],
            else_branch: None,
        }),
    );
}

mod type_operators {
    use super::*;

    test_algebrize!(
        cast_full,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::Cast(mir::CastExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            to: mir::Type::String,
            on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "was_null".into(),
            ))),
            on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "was_error".into(),
            ))),
            is_nullable: false,
        })),
        input = ast::Expression::Cast(ast::CastExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            to: ast::Type::String,
            on_null: Some(Box::new(ast::Expression::StringConstructor(
                "was_null".into(),
            ))),
            on_error: Some(Box::new(ast::Expression::StringConstructor(
                "was_error".into(),
            ))),
        }),
    );

    test_algebrize!(
        cast_simple,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::Cast(mir::CastExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            to: mir::Type::String,
            on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
            is_nullable: true,
        })),
        input = ast::Expression::Cast(ast::CastExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            to: ast::Type::String,
            on_null: None,
            on_error: None,
        }),
    );

    test_algebrize!(
        type_assert_success,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            target_type: mir::Type::Int32,
        })),
        input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::Type::Int32,
        }),
    );

    test_algebrize_expr_and_schema_check!(
        type_assert_fail,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "::!",
            required: Schema::Atomic(Atomic::String).into(),
            found: Schema::Atomic(Atomic::Integer).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::Type::String,
        }),
    );

    test_algebrize!(
        type_assert_ext_json_string_does_not_convert_if_target_type_is_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "{\"$numberInt\": \"42\"}".to_string()
            ))),
            target_type: mir::Type::String,
        })),
        input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"42\"}".to_string()
            )),
            target_type: ast::Type::String,
        }),
    );

    test_algebrize!(
        type_assert_ext_json_string_converts_if_target_type_is_not_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            target_type: mir::Type::Int32,
        })),
        input = ast::Expression::TypeAssertion(ast::TypeAssertionExpr {
            expr: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"42\"}".to_string()
            )),
            target_type: ast::Type::Int32,
        }),
    );

    test_algebrize!(
        is_success,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            target_type: mir::TypeOrMissing::Type(mir::Type::Int32),
        })),
        input = ast::Expression::Is(ast::IsExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
        }),
    );

    test_algebrize_expr_and_schema_check!(
        is_recursive_failure,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Add",
            required: NUMERIC_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::String).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Is(ast::IsExpr {
            expr: Box::new(ast::Expression::Binary(ast::BinaryExpr {
                left: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
                op: ast::BinaryOp::Add,
                right: Box::new(ast::Expression::StringConstructor("a".into())),
            })),
            target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
        }),
    );

    test_algebrize!(
        is_implicit_converts_ext_json,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::Is(mir::IsExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::Integer(42))),
            target_type: mir::TypeOrMissing::Type(mir::Type::Int32),
        })),
        input = ast::Expression::Is(ast::IsExpr {
            expr: Box::new(ast::Expression::StringConstructor(
                "{\"$numberInt\": \"42\"}".to_string()
            )),
            target_type: ast::TypeOrMissing::Type(ast::Type::Int32),
        }),
    );
}

mod like {
    use super::*;

    test_algebrize!(
        like_success_with_pattern,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::Like(mir::LikeExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "42".into(),
            ))),
            pattern: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "42".into(),
            ))),
            escape: Some('f'),
        })),
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::StringConstructor("42".into())),
            pattern: Box::new(ast::Expression::StringConstructor("42".into())),
            escape: Some('f'),
        }),
    );

    test_algebrize!(
        like_success_no_pattern,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Ok(mir::Expression::Like(mir::LikeExpr {
            expr: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "42".into(),
            ))),
            pattern: Box::new(mir::Expression::Literal(mir::LiteralValue::String(
                "42".into(),
            ))),
            escape: None,
        })),
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::StringConstructor("42".into())),
            pattern: Box::new(ast::Expression::StringConstructor("42".into())),
            escape: None,
        }),
    );

    test_algebrize_expr_and_schema_check!(
        like_expr_must_be_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            pattern: Box::new(ast::Expression::StringConstructor("42".into())),
            escape: None,
        }),
    );

    test_algebrize_expr_and_schema_check!(
        like_pattern_must_be_string,
        method = algebrize_expression,
        in_implicit_type_conversion_context = false,
        expected = Err(Error::SchemaChecking(mir::schema::Error::SchemaChecking {
            name: "Like",
            required: STRING_OR_NULLISH.clone().into(),
            found: Schema::Atomic(Atomic::Integer).into(),
        })),
        expected_error_code = 1002,
        input = ast::Expression::Like(ast::LikeExpr {
            expr: Box::new(ast::Expression::StringConstructor("42".into())),
            pattern: Box::new(ast::Expression::Literal(ast::Literal::Integer(42))),
            escape: Some(' '),
        }),
    );
}
