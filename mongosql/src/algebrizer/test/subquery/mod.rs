use super::catalog;
use crate::{
    ast, map,
    mir::{binding_tuple::DatasourceName, schema::SchemaCache, *},
    multimap,
    schema::{Atomic, Document, Schema},
    set, unchecked_unique_linked_hash_map,
    usererror::UserError,
};
use lazy_static::lazy_static;

fn mir_array(scope: u16) -> Stage {
    Stage::Project(Project {
        is_add_fields: false,
        source: Box::new(Stage::Array(ArraySource {
            array: vec![Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => Expression::Literal(LiteralValue::Integer(1))
                }
                .into(),
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("arr", scope).into() => Expression::Reference(("arr", scope).into()),
        },
        cache: SchemaCache::new(),
    })
}
lazy_static! {
    static ref AST_ARRAY: ast::Datasource = ast::Datasource::Array(ast::ArraySource {
        array: vec![ast::Expression::Document(multimap! {
            "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
        },)],
        alias: "arr".into(),
    });
}
test_algebrize!(
    uncorrelated_exists,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
                        is_add_fields: false,
        source: Box::new(mir_array(1u16)),
        expression: map! {
            (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                "a".into() => Expression::Literal(LiteralValue::Integer(1))
            }.into())
        },
        cache: SchemaCache::new(),
    })).into())),
    input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                })
            )])
        },
        from_clause: Some(AST_ARRAY.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    correlated_exists,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
                        is_add_fields: false,
        source: Box::new(mir_array(2u16)),
        expression: map! {
            (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                "b_0".into() => Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                    field: "b".into(),
                    is_nullable: false,
                })
            }.into())
        },
        cache: SchemaCache::new(),
    })).into())),
    input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "b_0".into() => ast::Expression::Identifier("b".into())
                })
            )])
        },
        from_clause: Some(AST_ARRAY.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"b".to_string()},
            additional_properties: false,
            ..Default::default()
            }),
    },
);
test_algebrize!(
    exists_cardinality_gt_1,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Exists(Box::new(Stage::Project(Project {
                        is_add_fields: false,
        source: Box::new(Stage::Array(ArraySource {
            array: vec![
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1))}
                .into()),
                Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(2))}
                .into())
            ],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        })),
        expression: map! {
            ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into()),
        },
        cache: SchemaCache::new(),
    })).into())),
    input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
        },
        from_clause: Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                },),
                ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                },)
            ],
            alias: "arr".into(),
        })),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    exists_degree_gt_1,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Exists(
        Box::new(Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".to_string() => Expression::Literal(LiteralValue::Integer(1)),
                        "b".to_string() => Expression::Literal(LiteralValue::Integer(2))
                    }
                    .into(),
                )],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            })),
            expression: map! {
                ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into()),
            },
            cache: SchemaCache::new(),
        }))
        .into(),
    )),
    input = ast::Expression::Exists(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
        },
        from_clause: Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
            },),],
            alias: "arr".into(),
        })),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    uncorrelated_subquery_expr,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Subquery(SubqueryExpr {
        output_expr: Box::new(Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
            field: "a_0".to_string(),
            is_nullable: false,
        })),
        subquery: Box::new(Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(mir_array(1u16)),
            expression: map! {
                (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                    "a_0".into() => Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                        field: "a".into(),
                        is_nullable: false,
                    })
                }.into())
            },
            cache: SchemaCache::new(),
        })),
        is_nullable: false,
    })),
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "a_0".into() => ast::Expression::Identifier("a".into())
                })
            )])
        },
        from_clause: Some(AST_ARRAY.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    correlated_subquery_expr,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Subquery(SubqueryExpr {
        output_expr: Box::new(Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference((DatasourceName::Bottom, 2u16).into())),
            field: "b_0".to_string(),
            is_nullable: false,
        })),
        subquery: Box::new(Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(mir_array(2u16)),
            expression: map! {
                (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                    "b_0".into() => Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                        field: "b".into(),
                        is_nullable: false,
                    })
                }.into())
            },
            cache: SchemaCache::new(),
        })),
        is_nullable: false,
    })),
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "b_0".into() => ast::Expression::Identifier("b".into())
                })
            )])
        },
        from_clause: Some(AST_ARRAY.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"b".to_string()},
            additional_properties: false,
            ..Default::default()
            })
    },
);
test_algebrize!(
    degree_zero_unsat_output,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::InvalidSubqueryDegree),
    expected_error_code = 3022,
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
        },
        from_clause: Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![],
            alias: "arr".into(),
        })),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    substar_degree_eq_1,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Subquery(SubqueryExpr {
        output_expr: Box::new(Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
            field: "a".to_string(),
            is_nullable: false,
        })),
        subquery: Box::new(Stage::Project(Project {
            is_add_fields: false,
            source: Box::new(mir_array(1u16)),
            expression: map! {
                ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into())
            },
            cache: SchemaCache::new(),
        })),
        is_nullable: false,
    })),
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar(
                ast::SubstarExpr {
                    datasource: "arr".into(),
                }
            )])
        },
        from_clause: Some(AST_ARRAY.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    select_values_degree_gt_1,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::InvalidSubqueryDegree),
    expected_error_code = 3022,
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "a_0".into() => ast::Expression::Identifier("a".into()),
                    "b_0".into() => ast::Expression::Identifier("b".into())
                })
            ),])
        },
        from_clause: Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![
                ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::Literal(ast::Literal::Integer(1))
                },),
                ast::Expression::Document(multimap! {
                    "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
                },)
            ],
            alias: "arr".into(),
        })),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    star_degree_eq_1,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Subquery(SubqueryExpr {
        output_expr: Box::new(Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
            field: "a".to_string(),
            is_nullable: false,
        })),
        subquery: Box::new(mir_array(1u16)),
        is_nullable: false,
    })),
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
        },
        from_clause: Some(AST_ARRAY.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    select_star_degree_gt_1,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::InvalidSubqueryDegree),
    expected_error_code = 3022,
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
        },
        from_clause: Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
            })],
            alias: "arr".into(),
        })),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    substar_degree_gt_1,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::InvalidSubqueryDegree),
    expected_error_code = 3022,
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar(
                ast::SubstarExpr {
                    datasource: "arr".into(),
                }
            )])
        },
        from_clause: Some(ast::Datasource::Array(ast::ArraySource {
            array: vec![ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(1)),
                "b".into() => ast::Expression::Literal(ast::Literal::Integer(2))
            })],
            alias: "arr".into(),
        })),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: None,
        offset: None,
    },))),
);
test_algebrize!(
    uncorrelated_subquery_comparison_all,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
        operator: SubqueryComparisonOp::Eq,
        modifier: SubqueryModifier::All,
        argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
        subquery_expr: SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "a_0".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(mir_array(1u16)),
                expression: map! {
                (DatasourceName::Bottom,1u16).into() =>
                    Expression::Document(unchecked_unique_linked_hash_map!{
                        "a_0".into() =>
                            Expression::FieldAccess(FieldAccess{
                                expr:Box::new(Expression::Reference(("arr",1u16).into())),
                                field:"a".into(),
                                is_nullable:false,
                            })
                    }.into(),
                )},
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        },
        is_nullable: false,
    })),
    input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(5))),
        op: ast::ComparisonOp::Eq,
        quantifier: ast::SubqueryQuantifier::All,
        subquery: Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))
    }),
);
test_algebrize!(
    uncorrelated_subquery_comparison_any,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
        operator: SubqueryComparisonOp::Eq,
        modifier: SubqueryModifier::Any,
        argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
        subquery_expr: SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "a_0".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(mir_array(1u16)),
                expression: map! {
                    (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "a_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        },
        is_nullable: false,
    })),
    input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
        expr: Box::new(ast::Expression::Literal(ast::Literal::Integer(5))),
        op: ast::ComparisonOp::Eq,
        quantifier: ast::SubqueryQuantifier::Any,
        subquery: Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))
    }),
);
test_algebrize!(
    subquery_comparison_ext_json_arg_converted_if_subquery_is_not_string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
        operator: SubqueryComparisonOp::Eq,
        modifier: SubqueryModifier::Any,
        argument: Box::new(Expression::Literal(LiteralValue::Integer(5))),
        subquery_expr: SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "a_0".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(mir_array(1u16)),
                expression: map! {
                    (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "a_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        },
        is_nullable: false,
    })),
    input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
        expr: Box::new(ast::Expression::StringConstructor(
            "{\"$numberInt\": \"5\"}".to_string()
        )),
        op: ast::ComparisonOp::Eq,
        quantifier: ast::SubqueryQuantifier::Any,
        subquery: Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))
    }),
);
test_algebrize!(
    subquery_comparison_ext_json_arg_not_converted_if_subquery_is_string,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
        operator: SubqueryComparisonOp::Eq,
        modifier: SubqueryModifier::Any,
        argument: Box::new(Expression::Literal(LiteralValue::String("{\"$numberInt\": \"5\"}".to_string()))),
        subquery_expr: SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
                field: "a_0".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                        is_add_fields: false,
                source: Box::new(Stage::Project(Project {
                        is_add_fields: false,
                    source: Box::new(Stage::Array(ArraySource {
                        array: vec![Expression::Document(
                            unchecked_unique_linked_hash_map! {
                                "a".into() => Expression::Literal(LiteralValue::String("abc".to_string()))
                            }
                            .into(),
                        )],
                        alias: "arr".into(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        ("arr", 1u16).into() => Expression::Reference(("arr", 1u16).into()),
                    },
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "a_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("arr", 1u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        },
        is_nullable: false,
    })),
    input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
        expr: Box::new(ast::Expression::StringConstructor("{\"$numberInt\": \"5\"}".to_string())),
        op: ast::ComparisonOp::Eq,
        quantifier: ast::SubqueryQuantifier::Any,
        subquery: Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(ast::Datasource::Array(ast::ArraySource {
                array: vec![ast::Expression::Document(multimap! {
                    "a".into() => ast::Expression::StringConstructor("abc".to_string()),
                })],
                alias: "arr".into(),
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))
    }),
);
test_algebrize!(
    argument_from_super_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::SubqueryComparison(SubqueryComparison {
        operator: SubqueryComparisonOp::Eq,
        modifier: SubqueryModifier::All,
        argument: Box::new(Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            field: "b".to_string(),
            is_nullable: false,
        })),
        subquery_expr: SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((DatasourceName::Bottom, 2u16).into())),
                field: "a_0".to_string(),
                is_nullable: false,
            })),
            subquery: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(mir_array(2u16)),
                expression: map! {
                    (DatasourceName::Bottom, 2u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "a_0".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("arr", 2u16).into())),
                            field: "a".into(),
                            is_nullable: false,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            is_nullable: false,
        },
        is_nullable: false,
    })),
    input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
        expr: Box::new(ast::Expression::Identifier("b".into())),
        op: ast::ComparisonOp::Eq,
        quantifier: ast::SubqueryQuantifier::All,
        subquery: Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))
    }),
    env = map! {
        ("foo", 1u16).into() => Schema::Document( Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set!{"b".to_string()},
            additional_properties: false,
            ..Default::default()
            })
    },
);
test_algebrize!(
    argument_only_evaluated_in_super_scope,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Err(Error::FieldNotFound(
        "a".into(),
        None,
        ClauseType::Unintialized,
        0u16
    )),
    expected_error_code = 3008,
    input = ast::Expression::SubqueryComparison(ast::SubqueryComparisonExpr {
        expr: Box::new(ast::Expression::Identifier("a".into())),
        op: ast::ComparisonOp::Eq,
        quantifier: ast::SubqueryQuantifier::All,
        subquery: Box::new(ast::Query::Select(ast::SelectQuery {
            select_clause: ast::SelectClause {
                set_quantifier: ast::SetQuantifier::All,
                body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                    ast::Expression::Document(multimap! {
                        "a_0".into() => ast::Expression::Identifier("a".into())
                    })
                )])
            },
            from_clause: Some(AST_ARRAY.clone()),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        },))
    }),
);
test_algebrize!(
    potentially_missing_column,
    method = algebrize_expression,
    in_implicit_type_conversion_context = false,
    expected = Ok(Expression::Subquery(SubqueryExpr {
        output_expr: Box::new(Expression::FieldAccess(FieldAccess {
            expr: Box::new(Expression::Reference((DatasourceName::Bottom, 1u16).into())),
            field: "x".to_string(),
            is_nullable: true,
        })),
        subquery: Box::new(Stage::Limit(Limit {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(Stage::Project(Project {
                    is_add_fields: false,
                    source: Box::new(Stage::Collection(Collection {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        (DatasourceName::Named("bar".to_string()), 1u16).into() => Expression::Reference(("bar".to_string(), 1u16).into())
                    },
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (DatasourceName::Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map!{
                        "x".into() => Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(("bar", 1u16).into())),
                            field: "x".into(),
                            is_nullable: true,
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            limit: 1,
            cache: SchemaCache::new(),
        })),
        is_nullable: true,
    })),
    input = ast::Expression::Subquery(Box::new(ast::Query::Select(ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
                ast::Expression::Document(multimap! {
                    "x".into() => ast::Expression::Subpath(ast::SubpathExpr {
                        expr: Box::new(ast::Expression::Identifier("bar".into())),
                        subpath: "x".to_string()
                    })
                })
            )])
        },
        from_clause: Some(ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "bar".to_string(),
            alias: Some("bar".to_string()),
        })),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: Some(1),
        offset: None,
    }))),
    catalog = catalog(vec![("test", "bar")]),
);
