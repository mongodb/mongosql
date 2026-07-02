macro_rules! test_codegen_match_query {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::{air, codegen::MqlCodeGenerator};
            use bson::bson;

            let expected = $expected;
            let input = $input;

            let gen = MqlCodeGenerator {};
            assert_eq!(expected, gen.codegen_match_query(input));
        }
    };
}

mod or {
    test_codegen_match_query!(
        empty,
        expected = Ok(bson!({"$or": []})),
        input = air::MatchQuery::Or(vec![])
    );

    test_codegen_match_query!(
        single,
        expected = Ok(bson!({"$or": [{"a": {"$gt": 1}}]})),
        input = air::MatchQuery::Or(vec![air::MatchQuery::Comparison(
            air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("a".to_string().into()),
                arg: air::LiteralValue::Integer(1)
            }
        )])
    );

    test_codegen_match_query!(
        multiple,
        expected = Ok(bson!({"$or": [{"a": {"$gt": 1}}, {"a": {"$lt": 10}}]})),
        input = air::MatchQuery::Or(vec![
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("a".to_string().into()),
                arg: air::LiteralValue::Integer(1)
            }),
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Lt,
                input: Some("a".to_string().into()),
                arg: air::LiteralValue::Integer(10)
            }),
        ])
    );
}

mod and {
    test_codegen_match_query!(
        empty,
        expected = Ok(bson!({"$and": []})),
        input = air::MatchQuery::And(vec![])
    );

    test_codegen_match_query!(
        single,
        expected = Ok(bson!({"$and": [{"a": {"$gt": 1}}]})),
        input = air::MatchQuery::And(vec![air::MatchQuery::Comparison(
            air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("a".to_string().into()),
                arg: air::LiteralValue::Integer(1)
            }
        )])
    );

    test_codegen_match_query!(
        multiple,
        expected = Ok(bson!({"$and": [{"a": {"$gt": 1}}, {"a": {"$lt": 10}}]})),
        input = air::MatchQuery::And(vec![
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("a".to_string().into()),
                arg: air::LiteralValue::Integer(1)
            }),
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Lt,
                input: Some("a".to_string().into()),
                arg: air::LiteralValue::Integer(10)
            }),
        ])
    );
}

mod type_op {
    test_codegen_match_query!(
        missing,
        expected = Ok(bson!({"a": {"$exists": false}})),
        input = air::MatchQuery::Type(air::MatchLanguageType {
            input: Some("a".to_string().into()),
            target_type: air::TypeOrMissing::Missing,
        })
    );

    test_codegen_match_query!(
        number,
        expected = Ok(bson!({"a": {"$type": "number"}})),
        input = air::MatchQuery::Type(air::MatchLanguageType {
            input: Some("a".to_string().into()),
            target_type: air::TypeOrMissing::Number,
        })
    );

    test_codegen_match_query!(
        atomic,
        expected = Ok(bson!({"a": {"$type": "string"}})),
        input = air::MatchQuery::Type(air::MatchLanguageType {
            input: Some("a".to_string().into()),
            target_type: air::TypeOrMissing::Type(air::Type::String),
        })
    );
}

mod regex {
    test_codegen_match_query!(
        simple,
        expected = Ok(bson!({"a": {"$regex": "abc", "$options": "ix"}})),
        input = air::MatchQuery::Regex(air::MatchLanguageRegex {
            input: Some("a".to_string().into()),
            regex: "abc".into(),
            options: "ix".into(),
        })
    );
}

mod elem_match {
    test_codegen_match_query!(
        simple,
        expected = Ok(bson!({"a": {"$elemMatch": {"$gte": 1}}})),
        input = air::MatchQuery::ElemMatch(air::ElemMatch {
            input: "a".to_string().into(),
            condition: Box::new(air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gte,
                input: None,
                arg: air::LiteralValue::Integer(1),
            })),
        })
    );

    // A condition nested inside $elemMatch may itself be a Not, e.g. `{a: {$elemMatch: {$not:
    // {$gt: 10}}}}`. The condition has no field ref of its own (the field is supplied by the
    // enclosing $elemMatch), which is the fieldless scenario `not::fieldless_comparison`
    // exercises directly on codegen_match_not; this test proves the full composition through
    // codegen_match_elem_match.
    test_codegen_match_query!(
        not_condition,
        expected = Ok(bson!({"a": {"$elemMatch": {"$not": {"$gt": 10}}}})),
        input = air::MatchQuery::ElemMatch(air::ElemMatch {
            input: "a".to_string().into(),
            condition: Box::new(air::MatchQuery::Not(Box::new(air::MatchQuery::Comparison(
                air::MatchLanguageComparison {
                    function: air::MatchLanguageComparisonOp::Gt,
                    input: None,
                    arg: air::LiteralValue::Integer(10),
                }
            )))),
        })
    );
}

mod comp {
    test_codegen_match_query!(
        lt,
        expected = Ok(bson!({"a": {"$lt": 1}})),
        input = air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Lt,
            input: Some("a".to_string().into()),
            arg: air::LiteralValue::Integer(1),
        })
    );

    test_codegen_match_query!(
        lte,
        expected = Ok(bson!({"a": {"$lte": 1}})),
        input = air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Lte,
            input: Some("a".to_string().into()),
            arg: air::LiteralValue::Integer(1),
        })
    );

    test_codegen_match_query!(
        ne,
        expected = Ok(bson!({"a": {"$ne": 1}})),
        input = air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Ne,
            input: Some("a".to_string().into()),
            arg: air::LiteralValue::Integer(1),
        })
    );

    test_codegen_match_query!(
        eq,
        expected = Ok(bson!({"a": {"$eq": 1}})),
        input = air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Eq,
            input: Some("a".to_string().into()),
            arg: air::LiteralValue::Integer(1),
        })
    );

    test_codegen_match_query!(
        gt,
        expected = Ok(bson!({"a": {"$gt": 1}})),
        input = air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Gt,
            input: Some("a".to_string().into()),
            arg: air::LiteralValue::Integer(1),
        })
    );

    test_codegen_match_query!(
        gte,
        expected = Ok(bson!({"a": {"$gte": 1}})),
        input = air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function: air::MatchLanguageComparisonOp::Gte,
            input: Some("a".to_string().into()),
            arg: air::LiteralValue::Integer(1),
        })
    );
}

mod match_constant_false {
    test_codegen_match_query!(
        constant_false,
        expected = Ok(bson!({"_id": bson::Bson::MinKey, "$expr": false})),
        input = air::MatchQuery::False
    );
}

// "$not" nests around the operator expression scoped to a field
// In practice, the NotComparisonRewritePass optimizes out NOT expressions for comparisons, some of these
// tests are illustrative, but not necessarily representative of what the optimizer will produce.
mod not {

    test_codegen_match_query!(
        single_comparison,
        expected = Ok(bson!({"age": {"$not": {"$gt": 10}}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Comparison(
            air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("age".to_string().into()),
                arg: air::LiteralValue::Integer(10),
            }
        )))
    );

    test_codegen_match_query!(
        binary_comparison,
        expected = Ok(bson!({"$not": {"$gt": 10}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Comparison(
            air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: None,
                arg: air::LiteralValue::Integer(10),
            }
        )))
    );

    // Negating an $and/$or that spans multiple different fields
    test_codegen_match_query!(
        not_and,
        expected =
            Ok(bson!({"$nor": [{"$and": [{"age": {"$gt": 10}}, {"name": {"$eq": "Alice"}}]}]})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::And(vec![
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("age".to_string().into()),
                arg: air::LiteralValue::Integer(10),
            }),
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Eq,
                input: Some("name".to_string().into()),
                arg: air::LiteralValue::String("Alice".to_string()),
            }),
        ])))
    );

    // Double negation is structurally valid; simplification is the optimizer's responsibility.
    test_codegen_match_query!(
        nested_not,
        expected = Ok(bson!({"age": {"$not": {"$not": {"$gt": 10}}}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Not(Box::new(
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("age".to_string().into()),
                arg: air::LiteralValue::Integer(10),
            })
        ))))
    );

    // NOT (x IN (1, 2, 3))
    test_codegen_match_query!(
        in_op,
        expected = Ok(bson!({"x": {"$not": {"$in": [1, 2, 3]}}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::In(air::MatchLanguageIn {
            op: air::MatchLanguageInOp::In,
            expression: air::FieldRef {
                parent: None,
                name: "x".to_string()
            },
            array_expression: vec![
                air::LiteralValue::Integer(1),
                air::LiteralValue::Integer(2),
                air::LiteralValue::Integer(3),
            ],
        })))
    );

    // NOT (x NOT IN (1, 2, 3)): double negation composed from Not wrapping In's own NotIn
    // form (which already codegens to "$not": {"$in": [...]}}).
    test_codegen_match_query!(
        not_in_op,
        expected = Ok(bson!({"x": {"$not": {"$not": {"$in": [1, 2, 3]}}}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::In(air::MatchLanguageIn {
            op: air::MatchLanguageInOp::NotIn,
            expression: air::FieldRef {
                parent: None,
                name: "x".to_string()
            },
            array_expression: vec![
                air::LiteralValue::Integer(1),
                air::LiteralValue::Integer(2),
                air::LiteralValue::Integer(3),
            ],
        })))
    );

    test_codegen_match_query!(
        regex,
        expected = Ok(bson!({"title": {"$not": {"$regex": "^T", "$options": ""}}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Regex(air::MatchLanguageRegex {
            input: Some("title".to_string().into()),
            regex: "^T".into(),
            options: "".into(),
        })))
    );

    test_codegen_match_query!(
        fieldless_regex,
        expected = Ok(bson!({"$not": {"$regex": "^T", "$options": ""}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Regex(air::MatchLanguageRegex {
            input: None,
            regex: "^T".into(),
            options: "".into(),
        })))
    );

    // Negating an $or that spans multiple different fields hits the same fallback path as
    // not_and.
    test_codegen_match_query!(
        not_or,
        expected = Ok(bson!({"$nor": [{"age": {"$gt": 10}}, {"name": {"$eq": "Alice"}}]})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Or(vec![
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gt,
                input: Some("age".to_string().into()),
                arg: air::LiteralValue::Integer(10),
            }),
            air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Eq,
                input: Some("name".to_string().into()),
                arg: air::LiteralValue::String("Alice".to_string()),
            }),
        ])))
    );

    test_codegen_match_query!(
        type_op,
        expected = Ok(bson!({"a": {"$not": {"$exists": false}}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Type(air::MatchLanguageType {
            input: Some("a".to_string().into()),
            target_type: air::TypeOrMissing::Missing,
        })))
    );

    test_codegen_match_query!(
        fieldless_type_op,
        expected = Ok(bson!({"$not": {"$type": "number"}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::Type(air::MatchLanguageType {
            input: None,
            target_type: air::TypeOrMissing::Number,
        })))
    );

    test_codegen_match_query!(
        elem_match,
        expected = Ok(bson!({"a": {"$not": {"$elemMatch": {"$gte": 1}}}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::ElemMatch(air::ElemMatch {
            input: "a".to_string().into(),
            condition: Box::new(air::MatchQuery::Comparison(air::MatchLanguageComparison {
                function: air::MatchLanguageComparisonOp::Gte,
                input: None,
                arg: air::LiteralValue::Integer(1),
            })),
        })))
    );

    test_codegen_match_query!(
        constant_false,
        expected = Ok(bson!({"$not": {"_id": bson::Bson::MinKey, "$expr": false}})),
        input = air::MatchQuery::Not(Box::new(air::MatchQuery::False))
    );
}

mod in_op {
    test_codegen_match_query!(
        in_single,
        expected = Ok(bson!({"a": {"$in": [1]}})),
        input = air::MatchQuery::In(air::MatchLanguageIn {
            op: air::MatchLanguageInOp::In,
            expression: air::FieldRef {
                parent: None,
                name: "a".to_string()
            },
            array_expression: vec![air::LiteralValue::Integer(1)],
        })
    );

    test_codegen_match_query!(
        in_multiple,
        expected = Ok(bson!({"a": {"$in": [1, 2, 3]}})),
        input = air::MatchQuery::In(air::MatchLanguageIn {
            op: air::MatchLanguageInOp::In,
            expression: air::FieldRef {
                parent: None,
                name: "a".to_string()
            },
            array_expression: vec![
                air::LiteralValue::Integer(1),
                air::LiteralValue::Integer(2),
                air::LiteralValue::Integer(3),
            ],
        })
    );

    // NotIn must emit { "$not": { "$in": [...] } }, NOT "$nin", which is not supported in
    // $match/$project expressions.
    test_codegen_match_query!(
        not_in_single,
        expected = Ok(bson!({"a": {"$not": {"$in": [1]}}})),
        input = air::MatchQuery::In(air::MatchLanguageIn {
            op: air::MatchLanguageInOp::NotIn,
            expression: air::FieldRef {
                parent: None,
                name: "a".to_string()
            },
            array_expression: vec![air::LiteralValue::Integer(1)],
        })
    );

    test_codegen_match_query!(
        not_in_multiple,
        expected = Ok(bson!({"a": {"$not": {"$in": [1, 2, 3]}}})),
        input = air::MatchQuery::In(air::MatchLanguageIn {
            op: air::MatchLanguageInOp::NotIn,
            expression: air::FieldRef {
                parent: None,
                name: "a".to_string()
            },
            array_expression: vec![
                air::LiteralValue::Integer(1),
                air::LiteralValue::Integer(2),
                air::LiteralValue::Integer(3),
            ],
        })
    );
}
