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
