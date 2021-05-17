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

mod add_alias {
    use super::*;

    test_rewrite!(
        simple_ident,
        AddAliasRewritePass,
        Ok("SELECT a AS a"),
        "SELECT a",
    );
    test_rewrite!(
        compound_ident,
        AddAliasRewritePass,
        Ok("SELECT a.b.c AS c"),
        "SELECT a.b.c",
    );
    test_rewrite!(
        generated_aliases,
        AddAliasRewritePass,
        Ok("SELECT a + b AS _1, 123 AS _2"),
        "SELECT a + b, 123",
    );
    test_rewrite!(
        ident_and_generated_aliases,
        AddAliasRewritePass,
        Ok("SELECT a AS a, 123 AS _2, b AS c, 456 AS _4"),
        "SELECT a, 123, b AS c, 456",
    );
    test_rewrite!(
        duplicate_aliases,
        AddAliasRewritePass,
        Ok("SELECT 123 AS _1, 456 AS _1"),
        "SELECT 123, 456 AS _1",
    );
    test_rewrite!(
        group_by_no_alias_top_level_field_ref,
        AddAliasRewritePass,
        Ok("SELECT * GROUP BY a, foo.b, c"),
        "SELECT * GROUP BY a, foo.b, c",
    );
    test_rewrite!(
        group_by_skip_single_dot_ref,
        AddAliasRewritePass,
        Ok("SELECT * GROUP BY foo.bar.c AS c, bar.b"),
        "SELECT * GROUP BY foo.bar.c, bar.b",
    );
    test_rewrite!(
        group_by_non_ref,
        AddAliasRewritePass,
        Ok("SELECT * GROUP BY a + b AS _groupKey1, c * d AS _groupKey2"),
        "SELECT * GROUP BY a + b, c * d",
    );
    test_rewrite!(
        group_by_non_ref_explicit_alias,
        AddAliasRewritePass,
        Ok("SELECT * GROUP BY a * b AS a, c * d AS _groupKey2"),
        "SELECT * GROUP BY a * b AS a, c * d",
    );
    test_rewrite!(
        group_by_mix_non_ref_and_ref,
        AddAliasRewritePass,
        Ok("SELECT * GROUP BY a, a + b AS _groupKey2, c, c * d AS _groupKey4, e"),
        "SELECT * GROUP BY a, a + b, c, c * d, e",
    );
    test_rewrite!(
        mix_select_list_group_by,
        AddAliasRewritePass,
        Ok("SELECT a + b AS _1, b AS b GROUP BY a, a + b AS _groupKey2, c, c * d AS _groupKey4, e"),
        "SELECT a + b, b GROUP BY a, a + b, c, c * d, e",
    );
    test_rewrite!(
        collection_source_simple_ident,
        AddAliasRewritePass,
        Ok("SELECT * FROM foo AS foo"),
        "SELECT * FROM foo",
    );
    test_rewrite!(
        collection_source_compound_ident,
        AddAliasRewritePass,
        Ok("SELECT * FROM foo.bar AS bar"),
        "SELECT * FROM foo.bar",
    );
    test_rewrite!(
        collection_source_no_rewrite,
        AddAliasRewritePass,
        Ok("SELECT * FROM foo AS bar"),
        "SELECT * FROM foo AS bar",
    );
    test_rewrite!(
        from_join_simple_ident,
        AddAliasRewritePass,
        Ok("SELECT * FROM foo AS foo CROSS JOIN bar AS bar CROSS JOIN car AS car"),
        "SELECT * FROM foo JOIN bar JOIN car",
    );
    test_rewrite!(
        from_join_compound_ident,
        AddAliasRewritePass,
        Ok("SELECT * FROM foo.bar AS bar CROSS JOIN bar AS bar CROSS JOIN car AS car"),
        "SELECT * FROM foo.bar JOIN bar JOIN car",
    );
    test_rewrite!(
        subquery_simple_ident,
        AddAliasRewritePass,
        Ok("SELECT a AS a FROM foo AS foo WHERE a > (SELECT b AS b FROM baz AS baz)"),
        "SELECT a FROM foo WHERE a  > (SELECT b FROM baz)",
    );
    test_rewrite!(
        subquery_generated_alias,
        AddAliasRewritePass,
        Ok("SELECT a AS a, 5 AS _2 FROM foo AS foo WHERE a > (SELECT 123 AS _1)"),
        "SELECT a, 5 FROM foo WHERE a  > (SELECT 123)",
    );
    test_rewrite!(
        counter_subquery,
        AddAliasRewritePass,
        Ok("SELECT 1 + 2 AS _1, (SELECT 3 + a AS _1, 15 + b AS _2) AS _2, 4 + 5 AS _3"),
        "SELECT 1 + 2, (SELECT 3 + a, 15 + b), 4+5",
    );
    test_rewrite!(
        counter_multiple_nested_subqueries,
        AddAliasRewritePass,
        Ok("SELECT 1 + 2 AS _1, (SELECT 3 + a AS _1, 4 + b AS _2 FROM (SELECT 5 + 6 AS _1) AS sub) AS _2, 7 + 8 AS _3"),
        "SELECT 1 + 2, (SELECT 3 + a, 4 + b FROM (SELECT 5+6) AS sub), 7+8",
    );
    test_rewrite!(
        group_by_subquery,
        AddAliasRewritePass,
        Ok("SELECT 1 + 2 AS _1, (SELECT * FROM bar AS bar GROUP BY a, c + d AS _groupKey2) AS _2, b AS b FROM foo AS foo GROUP BY b + e AS _groupKey1, d"),
        "SELECT 1 + 2, (SELECT * FROM bar AS bar GROUP BY a, c + d), b FROM foo AS foo GROUP BY b + e, d",
    );
}
