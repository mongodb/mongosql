use crate::{
    ast, mir, mir::schema::SchemaCache, schema::Satisfaction, unchecked_unique_linked_hash_map,
    usererror::UserError,
};
use lazy_static::lazy_static;

// ARRAY DATASOURCE
// [{"a" : 1}] AS arr
fn mir_array_source() -> mir::Stage {
    mir::Stage::Array(mir::ArraySource {
        array: vec![mir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
            }
            .into(),
        )],
        alias: "arr".into(),
        cache: SchemaCache::new(),
    })
}
// GROUP BY KEYS
// arr.a AS key
fn mir_field_access() -> mir::OptionallyAliasedExpr {
    mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
        alias: "key".to_string(),
        expr: mir::Expression::FieldAccess(mir::FieldAccess {
            expr: Box::new(mir::Expression::Reference(("arr", 0u16).into())),
            field: "a".to_string(),
            is_nullable: false,
        }),
    })
}
// 1 AS literal
fn mir_literal_key() -> mir::OptionallyAliasedExpr {
    mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
        alias: "literal".into(),
        expr: mir::Expression::Literal(mir::LiteralValue::Integer(1)),
    })
}

// a + 1 as complex_expr
fn mir_field_access_complex_expr() -> mir::OptionallyAliasedExpr {
    mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
        alias: "complex_expr".into(),
        expr: mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
            function: mir::ScalarFunction::Add,
            args: vec![
                mir::Expression::FieldAccess(mir::FieldAccess {
                    expr: Box::new(mir::Expression::Reference(("arr", 0u16).into())),
                    field: "a".to_string(),
                    is_nullable: false,
                }),
                mir::Expression::Literal(mir::LiteralValue::Integer(1)),
            ],
            is_nullable: false,
        }),
    })
}
// AVG(DISTINCT arr.a) AS agg1
fn mir_agg_1_array() -> mir::AliasedAggregation {
    mir::AliasedAggregation {
        alias: "agg1".to_string(),
        agg_expr: mir::AggregationExpr::Function(mir::AggregationFunctionApplication {
            function: mir::AggregationFunction::Avg,
            arg: Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(mir::Expression::Reference(("arr", 0u16).into())),
                field: "a".to_string(),
                is_nullable: false,
            })),
            distinct: true,
            arg_is_possibly_doc: Satisfaction::Not,
        }),
    }
}
// COUNT(*) AS agg2
fn mir_agg_2() -> mir::AliasedAggregation {
    mir::AliasedAggregation {
        alias: "agg2".to_string(),
        agg_expr: mir::AggregationExpr::CountStar(false),
    }
}

lazy_static! {
    // GROUP BY KEYS
    static ref AST_SUBPATH: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
        expr: ast::Expression::Subpath(ast::SubpathExpr {
            expr: Box::new(ast::Expression::Identifier("arr".to_string())),
            subpath: "a".to_string()
        }),
        alias: "key".to_string(),
    });

    // 1 AS literal
    static ref AST_LITERAL_KEY: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
        expr: ast::Expression::Literal(ast::Literal::Integer(1)),
        alias: "literal".into(),
    });

    // a + 1 AS complex_expr
    static ref AST_SUBPATH_COMPLEX_EXPR: ast::OptionallyAliasedExpr = ast::OptionallyAliasedExpr::Aliased(ast::AliasedExpr {
        expr: ast::Expression::Binary(ast::BinaryExpr {
            left: Box::new(ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                subpath: "a".to_string()
            })),
            op: ast::BinaryOp::Add,
            right: Box::new(ast::Expression::Literal(ast::Literal::Integer(1)))
        }),
        alias: "complex_expr".into(),
    });

    // AGGREGATION FUNCTIONS

    // AVG(DISTINCT arr.a) AS agg1
    static ref AST_AGG_1_ARRAY: ast::AliasedExpr = ast::AliasedExpr {
        expr: ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Avg,
            args: ast::FunctionArguments::Args(vec![
                ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                    subpath: "a".to_string()
                })
            ]),
            set_quantifier: Some(ast::SetQuantifier::Distinct),
        }),
        alias: "agg1".to_string(),
    };

    // COUNT(*) AS agg2
    static ref AST_AGG_2: ast::AliasedExpr = ast::AliasedExpr {
        expr: ast::Expression::Function(ast::FunctionExpr {
            function: ast::FunctionName::Count,
            args: ast::FunctionArguments::Star,
            set_quantifier: None
        }),
        alias: "agg2".to_string(),
    };
}

// Successful tests.

// FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE AVG(DISTINCT arr.a) AS agg1, COUNT(*) AS agg2
test_algebrize!(
    group_by_key_with_aggregation_array_source,
    method = algebrize_group_by_clause,
    expected = Ok(mir::Stage::Group(mir::Group {
        source: Box::new(mir_array_source()),
        keys: vec![mir_field_access()],
        aggregations: vec![mir_agg_1_array(), mir_agg_2()],
        cache: SchemaCache::new(),
        scope: 0,
    })),
    input = Some(ast::GroupByClause {
        keys: vec![AST_SUBPATH.clone()],
        aggregations: vec![AST_AGG_1_ARRAY.clone(), AST_AGG_2.clone()],
    }),
    source = mir_array_source(),
);

// FROM [{"a": 1}] AS arr GROUP BY 1
test_algebrize!(
    group_by_key_is_literal,
    method = algebrize_group_by_clause,
    expected = Ok(mir::Stage::Group(mir::Group {
        source: Box::new(mir_array_source()),
        keys: vec![mir_literal_key()],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    })),
    input = Some(ast::GroupByClause {
        keys: vec![AST_LITERAL_KEY.clone()],
        aggregations: vec![],
    }),
    source = mir_array_source(),
);

// FROM [{"a": 1}] AS arr GROUP BY a + 1
test_algebrize!(
    group_by_key_is_complex_expression,
    method = algebrize_group_by_clause,
    expected = Ok(mir::Stage::Group(mir::Group {
        source: Box::new(mir_array_source()),
        keys: vec![mir_field_access_complex_expr()],
        aggregations: vec![],
        cache: SchemaCache::new(),
        scope: 0,
    })),
    input = Some(ast::GroupByClause {
        keys: vec![AST_SUBPATH_COMPLEX_EXPR.clone()],
        aggregations: vec![],
    }),
    source = mir_array_source(),
);

// Error tests.

// FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE 42 AS agg
test_algebrize!(
    group_by_key_with_non_function_aggregation_expression,
    method = algebrize_group_by_clause,
    expected = Err(Error::NonAggregationInPlaceOfAggregation(0)),
    expected_error_code = 3013,
    input = Some(ast::GroupByClause {
        keys: vec![AST_SUBPATH.clone()],
        aggregations: vec![ast::AliasedExpr {
            expr: ast::Expression::Literal(ast::Literal::Integer(42)),
            alias: "agg".to_string(),
        },],
    }),
    source = mir_array_source(),
);

// FROM [{"a": 1}] AS arr GROUP BY arr.a AS key, arr.a AS key
test_algebrize!(
    group_by_keys_must_have_unique_aliases,
    method = algebrize_group_by_clause,
    expected = Err(Error::DuplicateDocumentKey("key".into())),
    expected_error_code = 3023,
    input = Some(ast::GroupByClause {
        keys: vec![AST_SUBPATH.clone(), AST_SUBPATH.clone()],
        aggregations: vec![],
    }),
    source = mir_array_source(),
);

// FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE COUNT(*) AS a, COUNT(*) AS a
test_algebrize!(
    group_by_aggregations_must_have_unique_aliases,
    method = algebrize_group_by_clause,
    expected = Err(Error::DuplicateDocumentKey("a".into())),
    expected_error_code = 3023,
    input = Some(ast::GroupByClause {
        keys: vec![AST_SUBPATH.clone()],
        aggregations: vec![
            ast::AliasedExpr {
                expr: ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::Count,
                    args: ast::FunctionArguments::Star,
                    set_quantifier: None
                }),
                alias: "a".into(),
            },
            ast::AliasedExpr {
                expr: ast::Expression::Function(ast::FunctionExpr {
                    function: ast::FunctionName::Count,
                    args: ast::FunctionArguments::Star,
                    set_quantifier: None
                }),
                alias: "a".into(),
            },
        ],
    }),
    source = mir_array_source(),
);

// FROM [{"a": 1}] AS arr GROUP BY arr.a AS key AGGREGATE COUNT(*) AS key
test_algebrize!(
    group_by_aliases_must_be_unique_across_keys_and_aggregates,
    method = algebrize_group_by_clause,
    expected = Err(Error::DuplicateDocumentKey("key".into())),
    expected_error_code = 3023,
    input = Some(ast::GroupByClause {
        keys: vec![AST_SUBPATH.clone()],
        aggregations: vec![ast::AliasedExpr {
            expr: ast::Expression::Function(ast::FunctionExpr {
                function: ast::FunctionName::Count,
                args: ast::FunctionArguments::Star,
                set_quantifier: None
            }),
            alias: "key".into(),
        },],
    }),
    source = mir_array_source(),
);
