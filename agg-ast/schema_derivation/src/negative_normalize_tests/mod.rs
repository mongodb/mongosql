macro_rules! test_negation {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::negative_normalize::NegativeNormalize;
            use agg_ast::definitions::MatchExpression;

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
            use crate::negative_normalize::NegativeNormalize;
            use agg_ast::definitions::Expression;

            let input: Expression = serde_json::from_str($input).unwrap();
            let expected: Expression = serde_json::from_str($expected).unwrap();
            let result = input.get_negation();
            assert_eq!(result, expected);
        }
    };
}

#[cfg(test)]
mod match_field;
#[cfg(test)]
mod match_logical;
#[cfg(test)]
mod negative_normal_form;
#[cfg(test)]
mod tagged_ops;
#[cfg(test)]
mod untagged_ops;

#[cfg(test)]
mod constants {
    test_expression_negation!(literal, expected = r#"2"#, input = r#"2"#);

    test_expression_negation!(array, expected = r#"[2]"#, input = r#"[2]"#);

    test_expression_negation!(document, expected = r#"{"a": 2}"#, input = r#"{"a": 2}"#);
}

#[cfg(test)]
mod field_ref {
    test_expression_negation!(
        field_ref,
        expected = r#"{"$or": [{"$lte": ["$foo", null]}, {"$eq": ["$foo", 0]}, {"$eq": ["$foo", false]}]}"#,
        input = r#""$foo""#
    );
}
