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
