mod negation {
    use crate::match_stage::NegativeNormalize;
    use agg_ast::definitions::{Expression, MatchExpression};

    macro_rules! test_negation {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let input: MatchExpression = serde_json::from_str($input).unwrap();
                let expected: MatchExpression = serde_json::from_str($expected).unwrap();
                let result = input.get_negation();
                assert_eq!(result, expected);
            }
        };
    }

    macro_rules! test_expression_negation {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let input: Expression = serde_json::from_str($input).unwrap();
                let expected: Expression = serde_json::from_str($expected).unwrap();
                let result = input.get_negation();
                assert_eq!(result, expected);
            }
        };
    }

    test_negation!(
        lt,
        expected = r#"{"$expr": {"$gte": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$lt": ["$foo", true]}}"#
    );

    test_negation!(
        lte,
        expected = r#"{"$expr": {"$gt": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$lte": ["$foo", true]}}"#
    );

    test_negation!(
        gt,
        expected = r#"{"$expr": {"$lte": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$gt": ["$foo", true]}}"#
    );

    test_negation!(
        gte,
        expected = r#"{"$expr": {"$lt": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$gte": ["$foo", true]}}"#
    );

    test_negation!(
        eq,
        expected = r#"{"$expr": {"$ne": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$eq": ["$foo", true]}}"#
    );

    test_negation!(
        ne,
        expected = r#"{"$expr": {"$eq": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$ne": ["$foo", true]}}"#
    );

    test_negation!(
        and,
        expected = r#"{"$expr": {"$or": [{"$gt": ["$foo", 10]}, {"$lt": ["$foo", 5]}]}}"#,
        input = r#"{"$expr": {"$and": [{"$lte": ["$foo", 10]}, {"$gte": ["$foo", 5]}]}}"#
    );

    test_negation!(
        or,
        expected = r#"{"$expr": {"$and": [{"$gte": ["$foo", 10]}, {"$lte": ["$foo", 5]}]}}"#,
        input = r#"{"$expr": {"$or": [{"$lt": ["$foo", 10]}, {"$gt": ["$foo", 5]}]}}"#
    );

    test_expression_negation!(literal, expected = r#"2"#, input = r#"2"#);

    test_expression_negation!(array, expected = r#"[2]"#, input = r#"[2]"#);

    test_expression_negation!(document, expected = r#"{"a": 2}"#, input = r#"{"a": 2}"#);

    test_expression_negation!(
        field_ref,
        expected = r#"{"$or": [{"$lte": ["$foo", null]}, {"$eq": ["$foo", 0]}]}"#,
        input = r#""$foo""#
    );
}

mod negative_normal_form {
    use crate::match_stage::NegativeNormalize;
    use agg_ast::definitions::MatchExpression;

    macro_rules! test_nnf {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let input: MatchExpression = serde_json::from_str($input).unwrap();
                let expected: MatchExpression = serde_json::from_str($expected).unwrap();
                let result = input.get_negative_normal_form();
                assert_eq!(result, expected);
            }
        };
    }

    test_nnf!(
        expr_func_noop,
        expected = r#"{"$expr": {"$eq": ["$foo", 0]}}"#,
        input = r#"{"$expr": {"$eq": ["$foo", 0]}}"#
    );
}
