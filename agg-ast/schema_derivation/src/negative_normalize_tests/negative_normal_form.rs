use crate::negative_normalize::NegativeNormalize;
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
