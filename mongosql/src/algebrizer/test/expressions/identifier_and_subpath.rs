use super::*;

test_algebrize!(
    qualified_ref_in_current_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
        field: "a".into(),
        is_nullable: true
    })),
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Identifier("foo".into())),
        subpath: "a".into(),
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
        ("foo", 0u16).into() => Schema::Document( Document {
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
    qualified_ref_in_super_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
        field: "a".into(),
        is_nullable: true
    })),
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Identifier("foo".into())),
        subpath: "a".into(),
    }),
    env = map! {
        ("bar", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
        ("foo", 0u16).into() => Schema::Document( Document {
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
    unqualified_ref_may_exist_in_current_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
        field: "a".into(),
        is_nullable: true
    })),
    input = ast::Expression::Identifier("a".into()),
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
    unqualified_ref_must_exist_in_current_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(("foo", 1u16).into())),
        field: "a".into(),
        is_nullable: false
    })),
    input = ast::Expression::Identifier("a".into()),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    unqualified_ref_may_exist_only_in_super_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
        field: "a".into(),
        is_nullable: true
    })),
    input = ast::Expression::Identifier("a".into()),
    env = map! {
        ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
        ("foo", 0u16).into() => Schema::Document( Document {
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
    unqualified_ref_must_exist_in_super_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
        field: "a".into(),
        is_nullable: false
    })),
    input = ast::Expression::Identifier("a".into()),
    env = map! {
        ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
        ("foo", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    unqualified_ref_must_exist_in_super_scope_bot_source,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(Key::bot(0u16).into())),
        field: "a".into(),
        is_nullable: false
    })),
    input = ast::Expression::Identifier("a".into()),
    env = map! {
        ("foo", 1u16).into() => Schema::Atomic(Atomic::Integer),
        Key::bot(0u16) => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    unqualified_ref_may_and_must_exist_in_two_sources_one_of_which_is_bot,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(Key::bot(1u16).into())),
        field: "a".into(),
        is_nullable: false
    })),
    input = ast::Expression::Identifier("a".into()),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
        Key::bot(1u16) => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    unqualified_ref_must_exist_in_two_non_bot_sources,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::AmbiguousField(
        "a".into(),
        ClauseType::Unintialized,
        1u16
    )),
    expected_error_code = 3009,
    input = ast::Expression::Identifier("a".into()),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
        ("bar", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    unqualified_subpath_in_current_and_super_must_exist_in_current,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
            expr: Box::new(mir::Expression::Reference(("test", 1u16).into())),
            field: "a".into(),
            is_nullable: false
        })),
        field: "c".into(),
        is_nullable: true
    })),
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Identifier("a".into())),
        subpath: "c".into(),
    }),
    env = map! {
        ("test", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
        ("super_test", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    unqualified_subpath_in_current_and_super_may_exist_is_ambiguous,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::AmbiguousField(
        "a".into(),
        ClauseType::Unintialized,
        1u16
    )),
    expected_error_code = 3009,
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Identifier("a".into())),
        subpath: "c".into(),
    }),
    env = map! {
        ("test", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
        ("super_test", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);
test_algebrize!(
    subpath_implicit_converts_ext_json,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            }
            .into(),
        )),
        field: "a".into(),
        is_nullable: false
    })),
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::StringConstructor("{\"a\": 1}".into())),
        subpath: "a".into(),
    }),
    env = map! {
        ("foo", 0u16).into() => Schema::Document( Document {
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
    unqualified_subpath_in_super_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
            expr: Box::new(mir::Expression::Reference(("super_test", 0u16).into())),
            field: "a".into(),
            is_nullable: false
        })),
        field: "c".into(),
        is_nullable: false
    })),
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Identifier("a".into())),
        subpath: "c".into(),
    }),
    env = map! {
        ("test", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "b".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
        ("super_test", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{"c".into()},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    qualified_ref_prefers_super_datasource_to_local_field,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(("foo", 0u16).into())),
        field: "a".into(),
        is_nullable: true
    })),
    // test MongoSql: SELECT (SELECT foo.a FROM bar) FROM foo => (foo.a)
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Identifier("foo".into())),
        subpath: "a".into(),
    }),
    env = map! {
        ("bar", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "foo".into() => Schema::Document( Document {
                    keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
        ("foo", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic( Atomic::Integer),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    qualified_ref_to_local_field,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
            expr: Box::new(mir::Expression::Reference(("bar", 1u16).into())),
            field: "foo".into(),
            is_nullable: true
        })),
        field: "a".into(),
        is_nullable: true
    })),
    //test MongoSql: SELECT (SELECT bar.foo.a FROM bar) FROM foo => (bar.foo.a)
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("bar".into())),
            subpath: "foo".into(),
        })),
        subpath: "a".into(),
    }),
    env = map! {
        ("bar", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "foo".into() => Schema::Document( Document {
                    keys: map! {"a".into() => Schema::Atomic(Atomic::Double)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
        ("foo", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic( Atomic::Integer),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    unqualified_reference_and_may_contain_sub_and_must_contain_outer_is_ambiguous,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::AmbiguousField(
        "a".into(),
        ClauseType::Unintialized,
        1u16
    )),
    expected_error_code = 3009,
    input = ast::Expression::Subpath(ast::SubpathExpr {
        expr: Box::new(ast::Expression::Identifier("a".into())),
        subpath: "c".into(),
    }),
    env = map! {
        ("test", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
        }),
        ("super_test", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Document( Document {
                    keys: map! {"c".into() => Schema::Atomic(Atomic::Integer)},
                    required: set!{},
                    additional_properties: false,
                    ..Default::default()
                }),
            },
            required: set!{"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    },
);

test_algebrize!(
    ref_does_not_exist,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::FieldNotFound(
        "bar".into(),
        None,
        ClauseType::Unintialized,
        0u16
    )),
    expected_error_code = 3008,
    input = ast::Expression::Identifier("bar".into()),
);
