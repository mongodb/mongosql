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

mod aggregate {
    use super::*;

    // `SELECT` clause tests.
    test_rewrite!(
        one_func_in_select_clause,
        AggregateRewritePass,
        Ok("SELECT _agg1 FROM foo GROUP BY NULL AGGREGATE SUM(x) AS _agg1"),
        "SELECT SUM(x) FROM foo",
    );
    test_rewrite!(
        different_funcs_in_select_by_clause,
        AggregateRewritePass,
        Ok("SELECT _agg1, _agg2 FROM foo GROUP BY NULL AGGREGATE SUM(x) AS _agg1, COUNT(y) AS _agg2"),
        "SELECT SUM(x), COUNT(y) FROM foo",
    );
    test_rewrite!(
        identical_funcs_in_select_clause,
        AggregateRewritePass,
        Ok("SELECT _agg1, _agg2, _agg2, _agg1 FROM foo GROUP BY NULL AGGREGATE SUM(x) AS _agg1, SUM(x + 1) AS _agg2"),
        "SELECT SUM(x), SUM(x+1), SUM(x+1), SUM(x) FROM foo",
    );

    // `GROUP BY` clause tests.
    test_rewrite!(
        one_func_in_group_by_aggregate_not_modified,
        AggregateRewritePass,
        Ok("SELECT z FROM foo GROUP BY x AGGREGATE SUM(x) AS z"),
        "SELECT z FROM foo GROUP BY x AGGREGATE SUM(x) AS z",
    );

    // `HAVING` clause tests.
    test_rewrite!(
        one_func_in_having_clause_no_group_by,
        AggregateRewritePass,
        Ok("SELECT * FROM foo GROUP BY NULL AGGREGATE SUM(x) AS _agg1 HAVING _agg1 > 42"),
        "SELECT * FROM foo HAVING SUM(x) > 42",
    );
    test_rewrite!(
        one_func_in_having_clause_with_group_by_keys_preserved,
        AggregateRewritePass,
        Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS _agg1 HAVING _agg1 > 42"),
        "SELECT * FROM foo GROUP BY x HAVING SUM(x) > 42",
    );
    test_rewrite!(
        different_funcs_in_having_clause_with_group_by,
        AggregateRewritePass,
        Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS _agg1, COUNT(y) AS _agg2 HAVING _agg1 > 42 AND _agg2 < 42"),
        "SELECT * FROM foo GROUP BY x HAVING SUM(x) > 42 AND COUNT(y) < 42",
    );
    test_rewrite!(
        identical_funcs_in_having_clause_with_group_by,
        AggregateRewritePass,
        Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS _agg1 HAVING _agg1 < 42 AND _agg1 > 24"),
        "SELECT * FROM foo GROUP BY x HAVING SUM(x) < 42 AND SUM(x) > 24",
    );
    test_rewrite!(
        identical_funcs_in_having_clause_alias_order_dictated_by_select,
        AggregateRewritePass,
        Ok("SELECT _agg1, _agg2, _agg2, _agg1 FROM foo GROUP BY NULL AGGREGATE SUM(x) AS _agg1, SUM(x + 1) AS _agg2 HAVING _agg2 > 42 AND _agg1 < 42"),
        "SELECT SUM(x), SUM(x+1), SUM(x+1), SUM(x) FROM foo HAVING SUM(x+1) > 42 AND SUM(x) < 42",
    );

    // Subquery tests.
    test_rewrite!(
        top_level_select_and_subquery_select_different_funcs,
        AggregateRewritePass,
        Ok("SELECT _agg1 FROM (SELECT _agg1 GROUP BY NULL AGGREGATE COUNT(y) AS _agg1) AS z GROUP BY NULL AGGREGATE SUM(x) AS _agg1"),
        "SELECT SUM(x) FROM (SELECT COUNT(y)) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_select_identical_funcs,
        AggregateRewritePass,
        Ok("SELECT _agg1 FROM (SELECT _agg1 GROUP BY NULL AGGREGATE SUM(x) AS _agg1) AS z GROUP BY NULL AGGREGATE SUM(x) AS _agg1"),
        "SELECT SUM(x) FROM (SELECT SUM(x)) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_group_by_aggregate_not_modified,
        AggregateRewritePass,
        Ok("SELECT _agg1 FROM (SELECT * FROM foo GROUP BY x AGGREGATE COUNT(y) AS z) AS z GROUP BY NULL AGGREGATE SUM(x) AS _agg1"),
        "SELECT SUM(x) FROM (SELECT * FROM foo GROUP BY x AGGREGATE COUNT(y) AS z) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_having,
        AggregateRewritePass,
        Ok("SELECT _agg1 FROM (SELECT * FROM foo GROUP BY NULL AGGREGATE COUNT(y) AS _agg1 HAVING _agg1 > 42) AS z GROUP BY NULL AGGREGATE SUM(x) AS _agg1"),
        "SELECT SUM(x) FROM (SELECT * FROM foo HAVING COUNT(y) > 42) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_exists,
        AggregateRewritePass,
        Ok("SELECT _agg1 FROM foo WHERE EXISTS(SELECT * FROM bar GROUP BY NULL AGGREGATE COUNT(y) AS _agg1 HAVING _agg1 > 42) GROUP BY NULL AGGREGATE SUM(x) AS _agg1"),
        "SELECT SUM(x) FROM foo WHERE EXISTS(SELECT * FROM bar HAVING COUNT(y) > 42)",
    );
    test_rewrite!(
        subquery_in_func_in_top_level_select,
        AggregateRewritePass,
        Ok("SELECT _agg1 FROM foo GROUP BY NULL AGGREGATE SUM(x <> ANY(SELECT _agg1 FROM bar GROUP BY NULL AGGREGATE SUM(x) AS _agg1)) AS _agg1"),
        "SELECT SUM(x <> ANY(SELECT SUM(x) FROM bar)) FROM foo",
    );
    test_rewrite!(
        subquery_in_func_in_group_by_agg_list,
        AggregateRewritePass,
        Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x <> ANY(SELECT _agg1 FROM bar GROUP BY NULL AGGREGATE SUM(x) AS _agg1)) AS z"),
        "SELECT * FROM foo GROUP BY x AGGREGATE SUM(x <> ANY(SELECT SUM(x) FROM bar)) AS z",
    );

    // Error tests.

    // Error if an aggregation function is found in a `GROUP BY` key list.
    test_rewrite!(
        one_func_in_group_by_key_list_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByKeyList),
        "SELECT * FROM foo GROUP BY SUM(x)",
    );
    test_rewrite!(
        identical_funcs_in_select_clause_and_group_by_key_list_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByKeyList),
        "SELECT SUM(x) FROM foo GROUP BY SUM(x)",
    );
    test_rewrite!(
        one_func_in_group_by_key_list_in_subquery_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByKeyList),
        "SELECT SUM(x) FROM (SELECT * FROM foo GROUP BY SUM(x)) AS z GROUP BY x AGGREGATE SUM(x) AS z",
    );
    test_rewrite!(
        identical_funcs_in_group_by_key_list_and_agg_list_gives_key_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByKeyList),
        "SELECT * FROM foo GROUP BY SUM(x) AGGREGATE SUM(x)",
    );

    // Error if an aggregation function is found after `AGGREGATE` without an alias.
    test_rewrite!(
        one_func_in_group_by_agg_list_with_no_alias_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListNotAliased),
        "SELECT * FROM foo GROUP BY x AGGREGATE SUM(x)",
    );
    test_rewrite!(
        identical_funcs_in_select_clause_and_group_by_agg_list_with_no_alias_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListNotAliased),
        "SELECT SUM(x) FROM foo GROUP BY x AGGREGATE SUM(x)",
    );
    test_rewrite!(
        one_func_in_group_by_agg_list_in_subquery_with_no_alias_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListNotAliased),
        "SELECT SUM(x) FROM (SELECT * FROM foo GROUP BY x AGGREGATE SUM(x)) AS z",
    );

    // Error if an aggregation function is found after `AGGREGATE` and elsewhere in query.
    test_rewrite!(
        identical_funcs_in_select_clause_and_group_by_aggregate_clause_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        "SELECT SUM(x) FROM foo GROUP BY x AGGREGATE SUM(x) AS z",
    );
    test_rewrite!(
        different_funcs_in_select_clause_and_group_by_aggregate_clause_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        "SELECT SUM(x) FROM foo GROUP BY x AGGREGATE COUNT(x) AS z",
    );
    test_rewrite!(
        identical_funcs_in_group_by_aggregate_clause_and_having_clause_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        "SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS z HAVING SUM(x) > 42",
    );
    test_rewrite!(
        different_funcs_in_group_by_aggregate_clause_and_having_clause_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        "SELECT * FROM foo GROUP BY x AGGREGATE COUNT(x) AS z HAVING SUM(x) > 42",
    );

    // Error for subquery containing an aggregation function after `AGGREGATE` and elsewhere.
    test_rewrite!(
        funcs_in_subquery_select_clause_and_group_by_aggregate_clause_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        "SELECT * FROM (SELECT SUM(x) FROM foo GROUP BY x AGGREGATE COUNT(y) AS z) AS z",
    );
    test_rewrite!(
        funcs_in_subquery_group_by_aggregate_clause_and_having_clause_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        "SELECT * FROM (SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS z HAVING SUM(x) > 42) AS z",
    );
    test_rewrite!(
        funcs_in_exists_select_clause_and_group_by_aggregate_clause_gives_error,
        AggregateRewritePass,
        Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        "SELECT * FROM foo WHERE EXISTS(SELECT SUM(x) FROM foo GROUP BY x AGGREGATE COUNT(y) AS z)",
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

mod select {
    use super::*;

    test_rewrite!(
        simple_ident_alias,
        SelectRewritePass,
        Ok("SELECT VALUE {'a1': a}"),
        "SELECT a as a1",
    );
    test_rewrite!(
        compound_ident_alias,
        SelectRewritePass,
        Ok("SELECT VALUE {'a1': a.b.c}"),
        "SELECT a.b.c as a1",
    );
    test_rewrite!(
        standalone_substar,
        SelectRewritePass,
        Ok("SELECT VALUE t.*"),
        "SELECT t.*",
    );
    test_rewrite!(
        multiple_substar,
        SelectRewritePass,
        Ok("SELECT VALUES t.*, a.*"),
        "SELECT t.*, a.*",
    );
    test_rewrite!(
        ident_substar_mix,
        SelectRewritePass,
        Ok("SELECT VALUES {'a': a}, t.*"),
        "SELECT a AS a, t.*",
    );
    test_rewrite!(
        multiple_ident_substar_mix,
        SelectRewritePass,
        Ok("SELECT VALUES {'a': a, 't': t}, a.*, t.*"),
        "SELECT a AS a, a.*, t AS t, t.*",
    );
    test_rewrite!(
        star_no_rewrite,
        SelectRewritePass,
        Ok("SELECT *"),
        "SELECT *",
    );
    test_rewrite!(
        select_value_no_rewrite,
        SelectRewritePass,
        Ok("SELECT VALUE {'a': a}"),
        "SELECT VALUE {'a': a}",
    );
    test_rewrite!(
        no_alias,
        SelectRewritePass,
        Err(Error::NoAliasForSelectExpression),
        "SELECT a",
    );
    test_rewrite!(
        subquery,
        SelectRewritePass,
        Ok("SELECT VALUE {'a': a, 'b': (SELECT VALUE {'c': c})}"),
        "SELECT a AS a, (SELECT c AS c) AS b",
    );
    test_rewrite!(
        select_values_subquery_top_level,
        SelectRewritePass,
        Err(Error::SubqueryWithSelectValue),
        "SELECT a AS a, (SELECT VALUES {'b': b}) AS sub",
    );
    test_rewrite!(
        select_values_subquery_datasource_only,
        SelectRewritePass,
        Ok("SELECT * FROM (SELECT VALUE {'c': d}) AS foo"),
        "SELECT * FROM (SELECT VALUE {'c': d}) AS foo",
    );
    test_rewrite!(
        select_values_subquery_not_top_level,
        SelectRewritePass,
        Ok("SELECT VALUE {'a': a, 'sub1': (SELECT VALUE {'b': b} FROM (SELECT VALUE {'c': d}) AS sub2)} FROM foo AS foo"),
        "SELECT a AS a, (SELECT b AS b FROM (SELECT VALUE {'c': d}) AS sub2) AS sub1 FROM foo AS foo",
    );
    test_rewrite!(
        select_values_exists_subquery,
        SelectRewritePass,
        Ok("SELECT VALUE {'foo': (SELECT VALUE {'a': a, 'sub': EXISTS(SELECT VALUE {'c': d})})}"),
        "SELECT (SELECT a AS a, EXISTS (SELECT VALUE {'c': d}) AS sub) AS foo",
    );
    test_rewrite!(
        select_values_nested_subquery_derived_datasource,
        SelectRewritePass,
        Err(Error::SubqueryWithSelectValue),
        "SELECT a AS a FROM (SELECT b AS b, (SELECT VALUES {'c': d}) AS sub) AS foo",
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

mod single_tuple {
    use super::*;

    test_rewrite!(
        one_element_tuple_unwrapped,
        SingleTupleRewritePass,
        Ok("SELECT a"),
        "SELECT (a)",
    );
    test_rewrite!(
        nested_one_element_tuple_unwrapped,
        SingleTupleRewritePass,
        Ok("SELECT a"),
        "SELECT (((a)))",
    );
    test_rewrite!(
        two_element_tuple_not_unwrapped,
        SingleTupleRewritePass,
        Ok("SELECT (a, b)"),
        "SELECT (a, b)",
    );
    test_rewrite!(
        nested_two_element_tuple_unwrapped,
        SingleTupleRewritePass,
        Ok("SELECT (a, b)"),
        "SELECT (((a, b)))",
    );
    test_rewrite!(
        subquery_one_element_tuple_unwrapped,
        SingleTupleRewritePass,
        Ok("SELECT * FROM (SELECT a) AS z"),
        "SELECT * FROM (SELECT (a)) AS z",
    );
    test_rewrite!(
        subquery_two_element_tuple_not_unwrapped,
        SingleTupleRewritePass,
        Ok("SELECT * FROM (SELECT (a, b)) AS z"),
        "SELECT * FROM (SELECT (a, b)) AS z",
    );
}
