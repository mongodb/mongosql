use crate::ast::{pretty_print::PrettyPrint, rewrites::*};

macro_rules! test_rewrite {
    ($func_name:ident, pass = $pass:expr, expected = $expected:expr, input = $input:expr,) => {
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

            let actual = pass.apply(query).map(|q| q.pretty_print().unwrap());

            assert_eq!(expected, actual);
        }
    };
}

mod positional_sort_key {
    use super::*;

    test_rewrite!(
        simple,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT a AS a FROM foo ORDER BY a ASC"),
        input = "SELECT a AS a FROM foo ORDER BY 1",
    );
    test_rewrite!(
        rewrite_in_derived_table,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT * FROM (SELECT a AS a FROM foo ORDER BY a ASC) AS sub"),
        input = "SELECT * FROM (SELECT a AS a FROM foo ORDER BY 1) sub",
    );
    test_rewrite!(
        rewrite_in_subquery_expr,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT (SELECT a AS a FROM foo ORDER BY a ASC)"),
        input = "SELECT (SELECT a AS a FROM foo ORDER BY 1)",
    );
    test_rewrite!(
        subquery_does_not_pollute_parent_state,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT b AS b, (SELECT a AS a FROM foo) FROM bar ORDER BY b ASC"),
        input = "SELECT b AS b, (SELECT a AS a FROM foo) FROM bar ORDER BY 1",
    );
    test_rewrite!(
        parent_does_not_pollute_subquery_state,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT b AS b, (SELECT a AS a FROM foo ORDER BY a ASC) FROM bar"),
        input = "SELECT b AS b, (SELECT a AS a FROM foo ORDER BY 1) FROM bar",
    );
    test_rewrite!(
        reference_aliases_not_modified,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT a AS a FROM foo ORDER BY a ASC, one ASC"),
        input = "SELECT a AS a FROM foo ORDER BY 1, one",
    );
    test_rewrite!(
        star_reference_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::PositionalSortKeyWithSelectStar),
        input = "SELECT * FROM foo ORDER BY 1",
    );
    test_rewrite!(
        substar_reference_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::NoAliasForSortKeyAtPosition(1)),
        input = "SELECT foo.* FROM foo ORDER BY 1",
    );
    test_rewrite!(
        unaliased_expression_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::NoAliasForSortKeyAtPosition(1)),
        input = "SELECT a FROM foo ORDER BY 1",
    );
    test_rewrite!(
        too_large_sort_key_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::PositionalSortKeyOutOfRange(2)),
        input = "SELECT a AS a FROM foo ORDER BY 2",
    );
    test_rewrite!(
        too_small_sort_key_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::PositionalSortKeyOutOfRange(0)),
        input = "SELECT a AS a FROM foo ORDER BY 0",
    );
    test_rewrite!(
        select_value_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::PositionalSortKeyWithSelectValue),
        input = "SELECT VALUE {'a': a} FROM foo ORDER BY 1",
    );
    test_rewrite!(
        select_star_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::PositionalSortKeyWithSelectStar),
        input = "SELECT * FROM foo ORDER BY 1",
    );
}

mod implicit_from {
    use super::*;

    test_rewrite!(
        simple_select_star,
        pass = ImplicitFromRewritePass,
        expected = Ok("SELECT * FROM [{}] AS _dual"),
        input = "SELECT *",
    );
    test_rewrite!(
        explicit_from_unmodified,
        pass = ImplicitFromRewritePass,
        expected = Ok("SELECT * FROM foo"),
        input = "SELECT * FROM foo",
    );
    test_rewrite!(
        rewrite_in_subquery,
        pass = ImplicitFromRewritePass,
        expected = Ok("SELECT * FROM (SELECT * FROM [{}] AS _dual) AS sub"),
        input = "SELECT * FROM (SELECT *) sub",
    );
}

mod aggregate {
    use super::*;

    // `SELECT` clause tests.
    test_rewrite!(
        one_func_in_select_clause,
        pass = AggregateRewritePass,
        expected =
            Ok("SELECT _agg1 FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1"),
        input = "SELECT SUM(x) FROM foo",
    );
    test_rewrite!(
        different_funcs_in_select_by_clause,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1, _agg2 FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1, COUNT(y) AS _agg2"),
        input = "SELECT SUM(x), COUNT(y) FROM foo",
    );
    test_rewrite!(
        identical_funcs_in_select_clause,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1, _agg2, _agg2, _agg1 FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1, SUM(x + 1) AS _agg2"),
        input = "SELECT SUM(x), SUM(x+1), SUM(x+1), SUM(x) FROM foo",
    );

    // `GROUP BY` clause tests.
    test_rewrite!(
        one_func_in_group_by_aggregate_not_modified,
        pass = AggregateRewritePass,
        expected = Ok("SELECT z FROM foo GROUP BY x AGGREGATE SUM(x) AS z"),
        input = "SELECT z FROM foo GROUP BY x AGGREGATE SUM(x) AS z",
    );

    // `HAVING` clause tests.
    test_rewrite!(
        one_func_in_having_clause_no_group_by,
        pass = AggregateRewritePass,
        expected =
            Ok("SELECT * FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1 HAVING _agg1 > 42"),
        input = "SELECT * FROM foo HAVING SUM(x) > 42",
    );
    test_rewrite!(
        one_func_in_having_clause_with_group_by_keys_preserved,
        pass = AggregateRewritePass,
        expected = Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS _agg1 HAVING _agg1 > 42"),
        input = "SELECT * FROM foo GROUP BY x HAVING SUM(x) > 42",
    );
    test_rewrite!(
        different_funcs_in_having_clause_with_group_by,
        pass = AggregateRewritePass,
        expected = Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS _agg1, COUNT(y) AS _agg2 HAVING _agg1 > 42 AND _agg2 < 42"),
        input = "SELECT * FROM foo GROUP BY x HAVING SUM(x) > 42 AND COUNT(y) < 42",
    );
    test_rewrite!(
        identical_funcs_in_having_clause_with_group_by,
        pass = AggregateRewritePass,
        expected = Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS _agg1 HAVING _agg1 < 42 AND _agg1 > 24"),
        input = "SELECT * FROM foo GROUP BY x HAVING SUM(x) < 42 AND SUM(x) > 24",
    );
    test_rewrite!(
        identical_funcs_in_having_clause_alias_order_dictated_by_select,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1, _agg2, _agg2, _agg1 FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1, SUM(x + 1) AS _agg2 HAVING _agg2 > 42 AND _agg1 < 42"),
        input = "SELECT SUM(x), SUM(x+1), SUM(x+1), SUM(x) FROM foo HAVING SUM(x+1) > 42 AND SUM(x) < 42",
    );

    // Subquery tests.
    test_rewrite!(
        top_level_select_and_subquery_select_different_funcs,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1 FROM (SELECT _agg1 GROUP BY NULL AS _groupKey1 AGGREGATE COUNT(y) AS _agg1) AS z GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1"),
        input = "SELECT SUM(x) FROM (SELECT COUNT(y)) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_select_identical_funcs,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1 FROM (SELECT _agg1 GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1) AS z GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1"),
        input = "SELECT SUM(x) FROM (SELECT SUM(x)) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_group_by_aggregate_not_modified,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1 FROM (SELECT * FROM foo GROUP BY x AGGREGATE COUNT(y) AS z) AS z GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1"),
        input = "SELECT SUM(x) FROM (SELECT * FROM foo GROUP BY x AGGREGATE COUNT(y) AS z) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_having,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1 FROM (SELECT * FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE COUNT(y) AS _agg1 HAVING _agg1 > 42) AS z GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1"),
        input = "SELECT SUM(x) FROM (SELECT * FROM foo HAVING COUNT(y) > 42) AS z",
    );
    test_rewrite!(
        top_level_select_and_subquery_exists,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1 FROM foo WHERE EXISTS(SELECT * FROM bar GROUP BY NULL AS _groupKey1 AGGREGATE COUNT(y) AS _agg1 HAVING _agg1 > 42) GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1"),
        input = "SELECT SUM(x) FROM foo WHERE EXISTS(SELECT * FROM bar HAVING COUNT(y) > 42)",
    );
    test_rewrite!(
        subquery_in_func_in_top_level_select,
        pass = AggregateRewritePass,
        expected = Ok("SELECT _agg1 FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x <> ANY(SELECT _agg1 FROM bar GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1)) AS _agg1"),
        input = "SELECT SUM(x <> ANY(SELECT SUM(x) FROM bar)) FROM foo",
    );
    test_rewrite!(
        subquery_in_func_in_group_by_agg_list,
        pass = AggregateRewritePass,
        expected = Ok("SELECT * FROM foo GROUP BY x AGGREGATE SUM(x <> ANY(SELECT _agg1 FROM bar GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1)) AS z"),
        input = "SELECT * FROM foo GROUP BY x AGGREGATE SUM(x <> ANY(SELECT SUM(x) FROM bar)) AS z",
    );

    // Error tests.

    // Error if an aggregation function is found in a `GROUP BY` key list.
    test_rewrite!(
        one_func_in_group_by_key_list_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByKeyList),
        input = "SELECT * FROM foo GROUP BY SUM(x)",
    );
    test_rewrite!(
        identical_funcs_in_select_clause_and_group_by_key_list_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByKeyList),
        input = "SELECT SUM(x) FROM foo GROUP BY SUM(x)",
    );
    test_rewrite!(
        one_func_in_group_by_key_list_in_subquery_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByKeyList),
        input = "SELECT SUM(x) FROM (SELECT * FROM foo GROUP BY SUM(x)) AS z GROUP BY x AGGREGATE SUM(x) AS z",
    );
    test_rewrite!(
        identical_funcs_in_group_by_key_list_and_agg_list_gives_key_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByKeyList),
        input = "SELECT * FROM foo GROUP BY SUM(x) AGGREGATE SUM(x) AS sumx",
    );

    // Error if an aggregation function is found after `AGGREGATE` and elsewhere in query.
    test_rewrite!(
        identical_funcs_in_select_clause_and_group_by_aggregate_clause_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        input = "SELECT SUM(x) FROM foo GROUP BY x AGGREGATE SUM(x) AS z",
    );
    test_rewrite!(
        different_funcs_in_select_clause_and_group_by_aggregate_clause_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        input = "SELECT SUM(x) FROM foo GROUP BY x AGGREGATE COUNT(x) AS z",
    );
    test_rewrite!(
        identical_funcs_in_group_by_aggregate_clause_and_having_clause_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        input = "SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS z HAVING SUM(x) > 42",
    );
    test_rewrite!(
        different_funcs_in_group_by_aggregate_clause_and_having_clause_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        input = "SELECT * FROM foo GROUP BY x AGGREGATE COUNT(x) AS z HAVING SUM(x) > 42",
    );

    // Error for subquery containing an aggregation function after `AGGREGATE` and elsewhere.
    test_rewrite!(
        funcs_in_subquery_select_clause_and_group_by_aggregate_clause_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        input = "SELECT * FROM (SELECT SUM(x) FROM foo GROUP BY x AGGREGATE COUNT(y) AS z) AS z",
    );
    test_rewrite!(
        funcs_in_subquery_group_by_aggregate_clause_and_having_clause_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        input = "SELECT * FROM (SELECT * FROM foo GROUP BY x AGGREGATE SUM(x) AS z HAVING SUM(x) > 42) AS z",
    );
    test_rewrite!(
        funcs_in_exists_select_clause_and_group_by_aggregate_clause_gives_error,
        pass = AggregateRewritePass,
        expected = Err(Error::AggregationFunctionInGroupByAggListAndElsewhere),
        input = "SELECT * FROM foo WHERE EXISTS(SELECT SUM(x) FROM foo GROUP BY x AGGREGATE COUNT(y) AS z)",
    );

    // ALL aggregation function test
    test_rewrite!(
        all_agg_becomes_unmodified_agg,
        pass = AggregateRewritePass,
        expected =
            Ok("SELECT _agg1 FROM foo GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1"),
        input = "SELECT SUM(ALL x) FROM foo",
    );
    test_rewrite!(
        nested_all,
        pass = AggregateRewritePass,
        expected =
            Ok("SELECT _agg2 GROUP BY NULL AS _groupKey1 AGGREGATE SUM(x) AS _agg1, SUM(_agg1) AS _agg2"),
        input = "SELECT SUM(ALL SUM(ALL x))",
    );
}

mod in_tuple {
    use super::*;

    test_rewrite!(
        one_element_tuple,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a = ANY(SELECT _1 FROM [{'_1': b}] AS _arr)"),
        input = "SELECT a IN (b)",
    );
    test_rewrite!(
        two_element_tuple,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a = ANY(SELECT _1 FROM [{'_1': b}, {'_1': c}] AS _arr)"),
        input = "SELECT a IN (b, c)",
    );
    test_rewrite!(
        one_element_tuple_not_in,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a <> ALL(SELECT _1 FROM [{'_1': b}] AS _arr)"),
        input = "SELECT a NOT IN (b)",
    );
    test_rewrite!(
        nested,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a = ANY(SELECT _1 FROM [{'_1': b = ANY(SELECT _1 FROM [{'_1': c}] AS _arr)}] AS _arr)"),
        input = "SELECT a IN (b IN (c))",
    );
    test_rewrite!(
        parenthesized_exprs_not_modified,
        pass = InTupleRewritePass,
        expected = Ok("SELECT (a + b) * c"),
        input = "SELECT (a + b) * c",
    );
    test_rewrite!(
        non_in_binary_op_not_modified,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a + (b, c)"),
        input = "SELECT a + (b, c)",
    );
    test_rewrite!(
        right_side_not_tuple_not_modified,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a IN b"),
        input = "SELECT a IN b",
    );
    test_rewrite!(
        position_in_argument_not_modified,
        pass = InTupleRewritePass,
        expected = Ok("SELECT POSITION(a IN (b))"),
        input = "SELECT POSITION(a IN (b))",
    );
}

mod select {
    use super::*;

    test_rewrite!(
        simple_ident_alias,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUE {'a1': a}"),
        input = "SELECT a as a1",
    );
    test_rewrite!(
        compound_ident_alias,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUE {'a1': a.b.c}"),
        input = "SELECT a.b.c as a1",
    );
    test_rewrite!(
        standalone_substar,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUE t.*"),
        input = "SELECT t.*",
    );
    test_rewrite!(
        multiple_substar,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUES t.*, a.*"),
        input = "SELECT t.*, a.*",
    );
    test_rewrite!(
        ident_substar_mix,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUES {'a': a}, t.*"),
        input = "SELECT a AS a, t.*",
    );
    test_rewrite!(
        multiple_ident_substar_mix,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUES {'a': a, 't': t}, a.*, t.*"),
        input = "SELECT a AS a, a.*, t AS t, t.*",
    );
    test_rewrite!(
        star_no_rewrite,
        pass = SelectRewritePass,
        expected = Ok("SELECT *"),
        input = "SELECT *",
    );
    test_rewrite!(
        select_value_no_rewrite,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUE {'a': a}"),
        input = "SELECT VALUE {'a': a}",
    );
    test_rewrite!(
        no_alias,
        pass = SelectRewritePass,
        expected = Err(Error::NoAliasForSelectExpression),
        input = "SELECT a",
    );
    test_rewrite!(
        subquery,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUE {'a': a, 'b': (SELECT VALUE {'c': c})}"),
        input = "SELECT a AS a, (SELECT c AS c) AS b",
    );
    test_rewrite!(
        select_values_subquery_top_level,
        pass = SelectRewritePass,
        expected = Err(Error::SubqueryWithSelectValue),
        input = "SELECT a AS a, (SELECT VALUES {'b': b}) AS sub",
    );
    test_rewrite!(
        select_values_subquery_datasource_only,
        pass = SelectRewritePass,
        expected = Ok("SELECT * FROM (SELECT VALUE {'c': d}) AS foo"),
        input = "SELECT * FROM (SELECT VALUE {'c': d}) AS foo",
    );
    test_rewrite!(
        select_values_subquery_not_top_level,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUE {'a': a, 'sub1': (SELECT VALUE {'b': b} FROM (SELECT VALUE {'c': d}) AS sub2)} FROM foo AS foo"),
        input = "SELECT a AS a, (SELECT b AS b FROM (SELECT VALUE {'c': d}) AS sub2) AS sub1 FROM foo AS foo",
    );
    test_rewrite!(
        select_values_exists_subquery,
        pass = SelectRewritePass,
        expected = Ok(
            "SELECT VALUE {'foo': (SELECT VALUE {'a': a, 'sub': EXISTS(SELECT VALUE {'c': d})})}"
        ),
        input = "SELECT (SELECT a AS a, EXISTS (SELECT VALUE {'c': d}) AS sub) AS foo",
    );
    test_rewrite!(
        select_values_nested_subquery_derived_datasource,
        pass = SelectRewritePass,
        expected = Err(Error::SubqueryWithSelectValue),
        input = "SELECT a AS a FROM (SELECT b AS b, (SELECT VALUES {'c': d}) AS sub) AS foo",
    );
}

mod add_alias {
    use super::*;

    test_rewrite!(
        simple_ident,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT a AS a"),
        input = "SELECT a",
    );
    test_rewrite!(
        compound_ident,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT a.b.c AS c"),
        input = "SELECT a.b.c",
    );
    test_rewrite!(
        generated_aliases,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT a + b AS _1, 123 AS _2"),
        input = "SELECT a + b, 123",
    );
    test_rewrite!(
        ident_and_generated_aliases,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT a AS a, 123 AS _2, b AS c, 456 AS _4"),
        input = "SELECT a, 123, b AS c, 456",
    );
    test_rewrite!(
        duplicate_aliases,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT 123 AS _1, 456 AS _1"),
        input = "SELECT 123, 456 AS _1",
    );
    test_rewrite!(
        group_by_no_alias_top_level_field_ref,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * GROUP BY a, foo.b, c"),
        input = "SELECT * GROUP BY a, foo.b, c",
    );
    test_rewrite!(
        group_by_skip_single_dot_ref,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * GROUP BY foo.bar.c AS c, bar.b"),
        input = "SELECT * GROUP BY foo.bar.c, bar.b",
    );
    test_rewrite!(
        group_by_non_ref,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * GROUP BY a + b AS _groupKey1, c * d AS _groupKey2"),
        input = "SELECT * GROUP BY a + b, c * d",
    );
    test_rewrite!(
        group_by_non_ref_explicit_alias,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * GROUP BY a * b AS a, c * d AS _groupKey2"),
        input = "SELECT * GROUP BY a * b AS a, c * d",
    );
    test_rewrite!(
        group_by_mix_non_ref_and_ref,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * GROUP BY a, a + b AS _groupKey2, c, c * d AS _groupKey4, e"),
        input = "SELECT * GROUP BY a, a + b, c, c * d, e",
    );
    test_rewrite!(
        mix_select_list_group_by,
        pass = AddAliasRewritePass,
        expected = Ok(
            "SELECT a + b AS _1, b AS b GROUP BY a, a + b AS _groupKey2, c, c * d AS _groupKey4, e"
        ),
        input = "SELECT a + b, b GROUP BY a, a + b, c, c * d, e",
    );
    test_rewrite!(
        collection_source_simple_ident,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * FROM foo AS foo"),
        input = "SELECT * FROM foo",
    );
    test_rewrite!(
        collection_source_compound_ident,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * FROM foo.bar AS bar"),
        input = "SELECT * FROM foo.bar",
    );
    test_rewrite!(
        collection_source_no_rewrite,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * FROM foo AS bar"),
        input = "SELECT * FROM foo AS bar",
    );
    test_rewrite!(
        from_join_simple_ident,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * FROM foo AS foo CROSS JOIN bar AS bar CROSS JOIN car AS car"),
        input = "SELECT * FROM foo JOIN bar JOIN car",
    );
    test_rewrite!(
        from_join_compound_ident,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT * FROM foo.bar AS bar CROSS JOIN bar AS bar CROSS JOIN car AS car"),
        input = "SELECT * FROM foo.bar JOIN bar JOIN car",
    );
    test_rewrite!(
        subquery_simple_ident,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT a AS a FROM foo AS foo WHERE a > (SELECT b AS b FROM baz AS baz)"),
        input = "SELECT a FROM foo WHERE a  > (SELECT b FROM baz)",
    );
    test_rewrite!(
        subquery_generated_alias,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT a AS a, 5 AS _2 FROM foo AS foo WHERE a > (SELECT 123 AS _1)"),
        input = "SELECT a, 5 FROM foo WHERE a  > (SELECT 123)",
    );
    test_rewrite!(
        counter_subquery,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT 1 + 2 AS _1, (SELECT 3 + a AS _1, 15 + b AS _2) AS _2, 4 + 5 AS _3"),
        input = "SELECT 1 + 2, (SELECT 3 + a, 15 + b), 4+5",
    );
    test_rewrite!(
        counter_multiple_nested_subqueries,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT 1 + 2 AS _1, (SELECT 3 + a AS _1, 4 + b AS _2 FROM (SELECT 5 + 6 AS _1) AS sub) AS _2, 7 + 8 AS _3"),
        input = "SELECT 1 + 2, (SELECT 3 + a, 4 + b FROM (SELECT 5+6) AS sub), 7+8",
    );
    test_rewrite!(
        group_by_in_subquery,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT 1 + 2 AS _1, (SELECT * FROM bar AS bar GROUP BY a, c + d AS _groupKey2) AS _2, b AS b FROM foo AS foo GROUP BY b + e AS _groupKey1, d"),
        input = "SELECT 1 + 2, (SELECT * FROM bar AS bar GROUP BY a, c + d), b FROM foo AS foo GROUP BY b + e, d",
    );
    test_rewrite!(
        group_by_with_subquery_key,
        pass = AddAliasRewritePass,
        expected = Ok("SELECT a + b AS _1, b AS b GROUP BY a, (SELECT a + b AS _1, c + d AS _2) AS _groupKey2, c, c * d AS _groupKey4, e"),
        input = "SELECT a + b, b GROUP BY a, (SELECT a + b, c + d), c, c * d, e",
    );
}

mod single_tuple {
    use super::*;

    test_rewrite!(
        one_element_tuple_unwrapped,
        pass = SingleTupleRewritePass,
        expected = Ok("SELECT a"),
        input = "SELECT (a)",
    );
    test_rewrite!(
        nested_one_element_tuple_unwrapped,
        pass = SingleTupleRewritePass,
        expected = Ok("SELECT a"),
        input = "SELECT (((a)))",
    );
    test_rewrite!(
        two_element_tuple_not_unwrapped,
        pass = SingleTupleRewritePass,
        expected = Ok("SELECT (a, b)"),
        input = "SELECT (a, b)",
    );
    test_rewrite!(
        nested_two_element_tuple_unwrapped,
        pass = SingleTupleRewritePass,
        expected = Ok("SELECT (a, b)"),
        input = "SELECT (((a, b)))",
    );
    test_rewrite!(
        subquery_one_element_tuple_unwrapped,
        pass = SingleTupleRewritePass,
        expected = Ok("SELECT * FROM (SELECT a) AS z"),
        input = "SELECT * FROM (SELECT (a)) AS z",
    );
    test_rewrite!(
        subquery_two_element_tuple_not_unwrapped,
        pass = SingleTupleRewritePass,
        expected = Ok("SELECT * FROM (SELECT (a, b)) AS z"),
        input = "SELECT * FROM (SELECT (a, b)) AS z",
    );
}

mod table_subquery {
    use super::*;

    test_rewrite!(
        in_to_eq_any,
        pass = TableSubqueryRewritePass,
        expected = Ok("SELECT * FROM table1 WHERE col1 = ANY(SELECT col1 FROM table2)"),
        input = "SELECT * FROM table1 WHERE col1 IN (SELECT col1 FROM table2)",
    );
    test_rewrite!(
        not_in_to_neq_all,
        pass = TableSubqueryRewritePass,
        expected = Ok("SELECT * FROM table1 WHERE col1 <> ALL(SELECT col1 FROM table2)"),
        input = "SELECT * FROM table1 WHERE col1 NOT IN (SELECT col1 FROM table2)",
    );
}
