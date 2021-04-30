use crate::ast::rewrites::*;

macro_rules! test_rewrite {
    ($func_name:ident, $pass:expr, $expected:expr, $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::{ast::rewrites::Pass, parser::Parser};

            let pass = $pass;
            let input = $input;
            let expected: Result<&str> = $expected;
            let expected = expected.map(String::from);

            let query = Parser::new()
                .parse_query(input)
                .expect("input query failed to parse");

            let actual = pass.apply(query).map(|q| format!("{}", q));

            assert_eq!(actual, expected);
        }
    };
}

mod positional_sort_key {
    use super::*;

    test_rewrite!(
        simple,
        PositionalSortKeyRewritePass,
        Ok("SELECT a AS a FROM foo ORDER BY a"),
        "SELECT a AS a FROM foo ORDER BY 1",
    );
    test_rewrite!(
        rewrite_in_derived_table,
        PositionalSortKeyRewritePass,
        Ok("SELECT * FROM (SELECT a AS a FROM foo ORDER BY a) AS sub"),
        "SELECT * FROM (SELECT a AS a FROM foo ORDER BY 1) sub",
    );
    test_rewrite!(
        rewrite_in_subquery_expr,
        PositionalSortKeyRewritePass,
        Ok("SELECT (SELECT a AS a FROM foo ORDER BY a)"),
        "SELECT (SELECT a AS a FROM foo ORDER BY 1)",
    );
    test_rewrite!(
        subquery_does_not_pollute_parent_state,
        PositionalSortKeyRewritePass,
        Ok("SELECT b AS b, (SELECT a AS a FROM foo) FROM bar ORDER BY b"),
        "SELECT b AS b, (SELECT a AS a FROM foo) FROM bar ORDER BY 1",
    );
    test_rewrite!(
        parent_does_not_pollute_subquery_state,
        PositionalSortKeyRewritePass,
        Ok("SELECT b AS b, (SELECT a AS a FROM foo ORDER BY a) FROM bar"),
        "SELECT b AS b, (SELECT a AS a FROM foo ORDER BY 1) FROM bar",
    );
    test_rewrite!(
        reference_aliases_not_modified,
        PositionalSortKeyRewritePass,
        Ok("SELECT a AS a FROM foo ORDER BY a, one"),
        "SELECT a AS a FROM foo ORDER BY 1, one",
    );
    test_rewrite!(
        star_reference_fails,
        PositionalSortKeyRewritePass,
        Err(Error::NoAliasForSortKeyAtPosition(1)),
        "SELECT * FROM foo ORDER BY 1",
    );
    test_rewrite!(
        substar_reference_fails,
        PositionalSortKeyRewritePass,
        Err(Error::NoAliasForSortKeyAtPosition(1)),
        "SELECT foo.* FROM foo ORDER BY 1",
    );
    test_rewrite!(
        unaliased_expression_fails,
        PositionalSortKeyRewritePass,
        Err(Error::NoAliasForSortKeyAtPosition(1)),
        "SELECT a FROM foo ORDER BY 1",
    );
    test_rewrite!(
        too_large_sort_key_fails,
        PositionalSortKeyRewritePass,
        Err(Error::PositionalSortKeyOutOfRange(2)),
        "SELECT a AS a FROM foo ORDER BY 2",
    );
    test_rewrite!(
        too_small_sort_key_fails,
        PositionalSortKeyRewritePass,
        Err(Error::PositionalSortKeyOutOfRange(0)),
        "SELECT a AS a FROM foo ORDER BY 0",
    );
    test_rewrite!(
        select_value_fails,
        PositionalSortKeyRewritePass,
        Err(Error::PositionalSortKeyWithSelectValue),
        "SELECT VALUE {'a': a} FROM foo ORDER BY 1",
    );
}

mod implicit_from {
    use super::*;

    test_rewrite!(
        simple_select_star,
        ImplicitFromRewritePass,
        Ok("SELECT * FROM [{}] AS _dual"),
        "SELECT *",
    );
    test_rewrite!(
        explicit_from_unmodified,
        ImplicitFromRewritePass,
        Ok("SELECT * FROM foo"),
        "SELECT * FROM foo",
    );
    test_rewrite!(
        rewrite_in_subquery,
        ImplicitFromRewritePass,
        Ok("SELECT * FROM (SELECT * FROM [{}] AS _dual) AS sub"),
        "SELECT * FROM (SELECT *) sub",
    );
}

mod in_tuple {
    use super::*;

    test_rewrite!(
        one_element_tuple,
        InTupleRewritePass,
        Ok("SELECT a = ANY(SELECT _1 FROM [{'_1': b}] AS _arr)"),
        "SELECT a IN (b)",
    );
    test_rewrite!(
        two_element_tuple,
        InTupleRewritePass,
        Ok("SELECT a = ANY(SELECT _1 FROM [{'_1': b}, {'_1': c}] AS _arr)"),
        "SELECT a IN (b, c)",
    );
    test_rewrite!(
        nested,
        InTupleRewritePass,
        Ok("SELECT a = ANY(SELECT _1 FROM [{'_1': b = ANY(SELECT _1 FROM [{'_1': c}] AS _arr)}] AS _arr)"),
        "SELECT a IN (b IN (c))",
    );
    test_rewrite!(
        parenthesized_exprs_not_modified,
        InTupleRewritePass,
        Ok("SELECT (a + b) * c"),
        "SELECT (a + b) * c",
    );
    test_rewrite!(
        non_in_binary_op_not_modified,
        InTupleRewritePass,
        Ok("SELECT a + (b, c)"),
        "SELECT a + (b, c)",
    );
    test_rewrite!(
        right_side_not_tuple_not_modified,
        InTupleRewritePass,
        Ok("SELECT a IN b"),
        "SELECT a IN b",
    );
    test_rewrite!(
        position_in_argument_not_modified,
        InTupleRewritePass,
        Ok("SELECT POSITION(a IN (b))"),
        "SELECT POSITION(a IN (b))",
    );
}
