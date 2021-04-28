use crate::ast::rewrites::InTupleRewritePass;

macro_rules! test_rewrite {
    ($func_name:ident, $passes:expr, $input:expr, $expected:expr) => {
        #[test]
        fn $func_name() {
            use crate::{ast::rewrites::Pass, parser::Parser};

            let passes = $passes;
            let input = $input;
            let expected = $expected;

            let query = Parser::new()
                .parse_query(input)
                .expect("input query failed to parse");
            let mut rewritten = query;
            for pass in passes {
                rewritten = pass.apply(rewritten).expect("rewrite pass failed");
            }
            let actual = format!("{}", rewritten);
            assert_eq!(actual, expected.to_string());
        }
    };
}

mod in_tuple {
    use super::*;

    test_rewrite!(
        one_element_tuple,
        vec![InTupleRewritePass],
        "SELECT a IN (b)",
        "SELECT a = ANY(SELECT _1 FROM [{'_1': b}] AS _arr)"
    );
    test_rewrite!(
        two_element_tuple,
        vec![InTupleRewritePass],
        "SELECT a IN (b, c)",
        "SELECT a = ANY(SELECT _1 FROM [{'_1': b}, {'_1': c}] AS _arr)"
    );
    test_rewrite!(
        nested,
        vec![InTupleRewritePass],
        "SELECT a IN (b IN (c))",
        "SELECT a = ANY(SELECT _1 FROM [{'_1': b = ANY(SELECT _1 FROM [{'_1': c}] AS _arr)}] AS _arr)"
    );
    test_rewrite!(
        parenthesized_exprs_not_modified,
        vec![InTupleRewritePass],
        "SELECT (a + b) * c",
        "SELECT (a + b) * c"
    );
    test_rewrite!(
        non_in_binary_op_not_modified,
        vec![InTupleRewritePass],
        "SELECT a + (b, c)",
        "SELECT a + (b, c)"
    );
    test_rewrite!(
        right_side_not_tuple_not_modified,
        vec![InTupleRewritePass],
        "SELECT a IN b",
        "SELECT a IN b"
    );
    test_rewrite!(
        position_in_argument_not_modified,
        vec![InTupleRewritePass],
        "SELECT POSITION(a IN (b))",
        "SELECT POSITION(a IN (b))"
    );
}
