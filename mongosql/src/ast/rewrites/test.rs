use crate::ast::{pretty_print::PrettyPrint, rewrites::*};

macro_rules! test_rewrite {
    ($func_name:ident, pass = $pass:expr, expected = $expected:expr, input = $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::{ast::rewrites::Pass, parser};

            let pass = $pass;
            let input = $input;
            let expected: Result<&str> = $expected;
            let expected = expected.map(String::from);

            let query = parser::parse_query(input).expect("input query failed to parse");

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
        recursively_rewrite_in_subquery_expr,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT b AS b, (SELECT a AS a FROM foo ORDER BY a ASC LIMIT 1) AS c FROM bar ORDER BY c ASC"),
        input = "SELECT b AS b, (SELECT a AS a FROM foo ORDER BY 1 LIMIT 1) AS c FROM bar ORDER BY 2",
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
        subquery_in_different_clause_does_not_pollute_parent_state,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT b AS b FROM bar WHERE (SELECT a AS a FROM foo) > 1 ORDER BY b ASC"),
        input = "SELECT b AS b FROM bar WHERE (SELECT a AS a FROM foo) > 1 ORDER BY 1",
    );
    test_rewrite!(
        parent_does_not_pollute_subquery_in_different_clause_state,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT b AS b FROM bar WHERE (SELECT a AS a FROM foo ORDER BY a ASC) > 1"),
        input = "SELECT b AS b FROM bar WHERE (SELECT a AS a FROM foo ORDER BY 1) > 1",
    );
    test_rewrite!(
        rewrite_both_parent_and_subquery,
        pass = PositionalSortKeyRewritePass,
        expected =
            Ok("SELECT b AS b, (SELECT a AS a FROM foo ORDER BY a ASC) FROM bar ORDER BY b ASC"),
        input = "SELECT b AS b, (SELECT a AS a FROM foo ORDER BY 1) FROM bar ORDER BY 1",
    );
    test_rewrite!(
        derived_table_does_not_pollute_parent_state,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT b AS b FROM (SELECT a AS a, b AS b FROM foo) AS sub ORDER BY b ASC"),
        input = "SELECT b AS b FROM (SELECT a AS a, b AS b FROM foo) AS sub ORDER BY 1",
    );
    test_rewrite!(
        parent_does_not_pollute_derived_table_state,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT b AS b FROM (SELECT a AS a, b AS b FROM foo ORDER BY a ASC) AS sub"),
        input = "SELECT b AS b FROM (SELECT a AS a, b AS b FROM foo ORDER BY 1) AS sub",
    );
    test_rewrite!(
        rewrite_both_parent_and_derived_table,
        pass = PositionalSortKeyRewritePass,
        expected =
            Ok("SELECT b AS b FROM (SELECT a AS a, b AS b FROM foo ORDER BY a ASC) AS sub ORDER BY b ASC"),
        input = "SELECT b AS b FROM (SELECT a AS a, b AS b FROM foo ORDER BY 1) AS sub ORDER BY 1",
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
        select_star_from_derived_outer_order_fails,
        pass = PositionalSortKeyRewritePass,
        expected = Err(Error::PositionalSortKeyWithSelectStar),
        input = "SELECT * FROM (SELECT a AS a FROM foo) AS sub ORDER BY 1",
    );
    test_rewrite!(
        select_star_from_derived_inner_order_rewrites,
        pass = PositionalSortKeyRewritePass,
        expected = Ok("SELECT * FROM (SELECT a AS a FROM foo ORDER BY a ASC) AS sub"),
        input = "SELECT * FROM (SELECT a AS a FROM foo ORDER BY 1) AS sub",
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

    // multi-arg COUNT tests
    test_rewrite!(
        one_multi_arg_count_rewritten_to_doc_arg,
        pass = AggregateRewritePass,
        expected = Ok(
            "SELECT * FROM foo GROUP BY NULL AS n AGGREGATE COUNT({'_arg0': a, '_arg1': b}) AS c"
        ),
        input = "SELECT * FROM foo GROUP BY NULL AS n AGGREGATE COUNT(a, b) AS c",
    );
    test_rewrite!(
        multiple_multi_arg_counts_rewritten_to_doc_arg,
        pass = AggregateRewritePass,
        expected = Ok("SELECT * FROM foo GROUP BY NULL AS n AGGREGATE COUNT({'_arg0': a, '_arg1': b}) AS c1, COUNT({'_arg0': b, '_arg1': c, '_arg2': e.f}) AS c2"),
        input = "SELECT * FROM foo GROUP BY NULL AS n AGGREGATE COUNT(a, b) AS c1, COUNT(b, c, e.f) AS c2",
    );
    test_rewrite!(
        multi_arg_count_in_different_clause_full_rewritten,
        pass = AggregateRewritePass,
        expected = Ok(
            "SELECT _agg1 FROM foo GROUP BY NULL AS n AGGREGATE COUNT(DISTINCT {'_arg0': a, '_arg1': b}) AS _agg1"
        ),
        input = "SELECT COUNT(DISTINCT a, b) FROM foo GROUP BY NULL AS n",
    );
}

mod in_tuple {
    use super::*;

    test_rewrite!(
        one_element_tuple,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a = b"),
        input = "SELECT a IN (b)",
    );
    test_rewrite!(
        three_element_tuple,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a = b OR a = c OR a = d"),
        input = "SELECT a IN (b, c, d)",
    );
    test_rewrite!(
        one_element_tuple_not_in,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a <> b"),
        input = "SELECT a NOT IN (b)",
    );
    test_rewrite!(
        three_element_tuple_not_in,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a <> b AND a <> c AND a <> d"),
        input = "SELECT a NOT IN (b, c, d)",
    );
    test_rewrite!(
        nested,
        pass = InTupleRewritePass,
        expected = Ok("SELECT a = (b = c)"),
        input = "SELECT a IN (b IN (c))",
    );
    test_rewrite!(
        one_element_tuple_no_simple_field_ref,
        pass = InTupleRewritePass,
        expected = Ok("SELECT SUM(a) IN (SELECT _1 FROM [{'_1': b}] AS _arr)"),
        input = "SELECT SUM(a) IN (b)",
    );
    test_rewrite!(
        one_element_tuple_not_in_no_simple_field_ref,
        pass = InTupleRewritePass,
        expected = Ok("SELECT SUM(a) NOT IN (SELECT _1 FROM [{'_1': b}] AS _arr)"),
        input = "SELECT SUM(a) NOT IN (b)",
    );
    test_rewrite!(
        three_element_tuple_no_simple_field_ref,
        pass = InTupleRewritePass,
        expected =
            Ok("SELECT SUM(a) IN (SELECT _1 FROM [{'_1': b}, {'_1': c}, {'_1': d}] AS _arr)"),
        input = "SELECT SUM(a) IN (b, c, d)",
    );
    test_rewrite!(
        three_element_tuple_not_in_no_simple_field_ref,
        pass = InTupleRewritePass,
        expected =
            Ok("SELECT SUM(a) NOT IN (SELECT _1 FROM [{'_1': b}, {'_1': c}, {'_1': d}] AS _arr)"),
        input = "SELECT SUM(a) NOT IN (b, c, d)",
    );
    test_rewrite!(
        nested_no_simple_field_ref,
        pass = InTupleRewritePass,
        expected = Ok("SELECT SUM(a) IN (SELECT _1 FROM [{'_1': SUM(b) IN (SELECT _1 FROM [{'_1': c}] AS _arr)}] AS _arr)"),
        input = "SELECT SUM(a) IN (SUM(b) IN (c))",
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
        multiple_ident_unordered,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUE {'c': c, 'a': a, 'b': b}"),
        input = "SELECT c AS c, a AS a, b AS b",
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
        expected = Ok("SELECT VALUES {'a': a}, a.*, {'t': t}, t.*"),
        input = "SELECT a AS a, a.*, t AS t, t.*",
    );
    test_rewrite!(
        multiple_ident_substar_mix_groups_idents,
        pass = SelectRewritePass,
        expected = Ok("SELECT VALUES {'b': b, 'a': a}, a.*, {'z': y, 'y': z}"),
        input = "SELECT b as b, a AS a, a.*, y AS z, z as y",
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

mod group_by_select_alias {
    use super::*;

    test_rewrite!(
        simple,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT b FROM foo GROUP BY a AS b"),
        input = "SELECT a AS b FROM foo GROUP BY b",
    );

    test_rewrite!(
        same_alias_as_field_name,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT b FROM foo GROUP BY b AS b"),
        input = "SELECT b AS b FROM foo GROUP BY b",
    );

    test_rewrite!(
        subquery,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT * FROM (SELECT b FROM foo GROUP BY a AS b) AS sub"),
        input = "SELECT * FROM (SELECT a AS b FROM foo GROUP BY b) AS sub",
    );

    test_rewrite!(
        do_not_rewrite_group_by_alias_in_subquery,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT * FROM (SELECT b AS a FROM mytbl) AS sub GROUP BY a"),
        input = "SELECT * FROM (SELECT b AS a FROM mytbl) AS sub GROUP BY a",
    );

    test_rewrite!(
        rewrite_to_alias_in_same_level_query_not_subquery,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT a FROM (SELECT b AS a FROM mytbl) AS sub GROUP BY sub.a AS a"),
        input = "SELECT sub.a AS a FROM (SELECT b AS a FROM mytbl) AS sub GROUP BY a",
    );

    test_rewrite!(
        do_not_rewrite_if_select_values,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT VALUE {'b': a} FROM foo GROUP BY b"),
        input = "SELECT VALUE {'b': a} FROM foo GROUP BY b",
    );

    test_rewrite!(
        do_not_rewrite_if_aggregation_exists,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT a AS b FROM foo GROUP BY b AGGREGATE SUM(b) AS sumb",),
        input = "SELECT a AS b FROM foo GROUP BY b AGGREGATE SUM(b) AS sumb",
    );

    test_rewrite!(
        tableau_generated_query,
        pass = GroupBySelectAliasRewritePass,
        expected = Ok("SELECT str2 FROM Calcs WHERE ((NOT (Calcs.str2 IN ('eight', 'eleven', 'fifteen', 'five'))) OR (Calcs.str2 IS NULL)) GROUP BY Calcs.str2 AS str2"),
        input = "SELECT Calcs.str2 AS str2 FROM Calcs WHERE ((NOT (Calcs.str2 IN ('eight', 'eleven', 'fifteen', 'five'))) OR (Calcs.str2 IS NULL)) GROUP BY str2",
    );
}

mod optional_parameters {
    use super::*;

    mod flatten {
        use super::*;

        test_rewrite!(
            remove_explicit_underscore_separator,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM FLATTEN(foo)"),
            input = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_')",
        );

        test_rewrite!(
            remove_explicit_underscore_separator_when_other_options_present,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM FLATTEN(foo WITH DEPTH => 5)"),
            input = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_', DEPTH => 5)",
        );

        test_rewrite!(
            do_not_modify_duplicate_default_separator_opts,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_', SEPARATOR => '_')"),
            input = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_', SEPARATOR => '_')",
        );

        test_rewrite!(
            do_not_modify_duplicate_non_default_separator_opts,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_', SEPARATOR => '-')"),
            input = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_', SEPARATOR => '-')",
        );

        test_rewrite!(
            recursively_apply_rewrite,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM FLATTEN(FLATTEN(foo))"),
            input =
                "SELECT * FROM FLATTEN(FLATTEN(foo WITH SEPARATOR => '_') WITH SEPARATOR => '_')",
        );
    }

    mod unwind {
        use super::*;

        test_rewrite!(
            remove_explicit_default_false,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM UNWIND(foo)"),
            input = "SELECT * FROM UNWIND(foo WITH OUTER => false)",
        );

        test_rewrite!(
            remove_explicit_default_false_when_other_options_present,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM UNWIND(foo WITH INDEX => idx)"),
            input = "SELECT * FROM UNWIND(foo WITH OUTER => false, INDEX => idx)",
        );

        test_rewrite!(
            do_not_modify_duplicate_default_outer_opts,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM UNWIND(foo WITH OUTER => false, OUTER => false)"),
            input = "SELECT * FROM UNWIND(foo WITH OUTER => false, OUTER => false)",
        );

        test_rewrite!(
            do_not_modify_duplicate_non_default_outer_opts,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM UNWIND(foo WITH OUTER => false, OUTER => true)"),
            input = "SELECT * FROM UNWIND(foo WITH OUTER => false, OUTER => true)",
        );

        test_rewrite!(
            recursively_apply_rewrite,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT * FROM UNWIND(UNWIND(foo))"),
            input = "SELECT * FROM UNWIND(UNWIND(foo WITH OUTER => false) WITH OUTER => false)",
        );
    }

    mod case {
        use super::*;

        test_rewrite!(
            simple_case_make_else_null_explicit,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT CASE num WHEN 1 THEN true ELSE NULL END FROM foo"),
            input = "SELECT CASE num WHEN 1 THEN true END FROM foo",
        );

        test_rewrite!(
            simple_case_do_not_modify_existing_else,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT CASE num WHEN 1 THEN true ELSE false END FROM foo"),
            input = "SELECT CASE num WHEN 1 THEN true ELSE false END FROM foo",
        );

        test_rewrite!(
            searched_case_make_else_null_explicit,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT CASE WHEN true THEN true ELSE NULL END FROM foo"),
            input = "SELECT CASE WHEN true THEN true END FROM foo",
        );

        test_rewrite!(
            searched_case_do_not_modify_existing_else,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT CASE WHEN true THEN true ELSE false END FROM foo"),
            input = "SELECT CASE WHEN true THEN true ELSE false END FROM foo",
        );

        test_rewrite!(
            recursively_apply_rewrite,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT CASE WHEN (CASE a WHEN 1 THEN true ELSE NULL END) THEN true ELSE NULL END FROM foo"),
            input = "SELECT CASE WHEN (CASE a WHEN 1 THEN true END) THEN true END FROM foo",
        );
    }

    mod function {
        use super::*;

        test_rewrite!(
            add_substring_optional_parameter_if_missing,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT SUBSTRING('abc' FROM 1 FOR -1) FROM foo"),
            input = "SELECT SUBSTRING('abc', 1) FROM foo",
        );

        test_rewrite!(
            do_not_modify_substring_with_optional_parameter,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT SUBSTRING('abc' FROM 1 FOR 1) FROM foo"),
            input = "SELECT SUBSTRING('abc', 1, 1) FROM foo",
        );

        test_rewrite!(
            add_current_timestamp_optional_parameter_if_missing,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT CURRENT_TIMESTAMP(6) FROM foo"),
            input = "SELECT CURRENT_TIMESTAMP FROM foo",
        );

        test_rewrite!(
            do_not_modify_current_timestamp_with_optional_parameter,
            pass = OptionalParameterRewritePass,
            expected = Ok("SELECT CURRENT_TIMESTAMP(1) FROM foo"),
            input = "SELECT CURRENT_TIMESTAMP(1) FROM foo",
        );

        test_rewrite!(
            recursively_apply_rewrite,
            pass = OptionalParameterRewritePass,
            expected =
                Ok("SELECT SUBSTRING(CAST(CURRENT_TIMESTAMP(6) AS STRING) FROM 1 FOR -1) FROM foo"),
            input = "SELECT SUBSTRING(CAST(CURRENT_TIMESTAMP AS STRING), 1) FROM foo",
        );
    }
}

mod not {
    use super::*;

    test_rewrite!(
        simple,
        pass = NotComparisonRewritePass,
        expected = Ok("SELECT a <> b FROM foo"),
        input = "SELECT NOT a = b FROM foo",
    );

    test_rewrite!(
        recursive,
        pass = NotComparisonRewritePass,
        expected = Ok("SELECT (SELECT a >= 1 FROM bar LIMIT 1) <> (SELECT b <= 1 FROM baz LIMIT 1) FROM foo"),
        input = "SELECT NOT (SELECT NOT a < 1 FROM bar LIMIT 1) = (SELECT NOT b > 1 FROM baz LIMIT 1) FROM foo",
    );
}

mod scalar_functions {
    use super::*;

    test_rewrite!(
        log_two_args_removes_escape,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT LOG(10, 5)"),
        input = "SELECT {fn LOG(10, 5)}",
    );
    test_rewrite!(
        log_three_args_is_error,
        pass = ScalarFunctionsRewritePass,
        expected = Err(Error::IncorrectArgumentCount {
            name: "LOG",
            required: "1 or 2",
            found: 3
        }),
        input = "SELECT { fn LOG(10, 5, 10) }",
    );
    test_rewrite!(
        log_one_arg_is_base_e,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT LOG(10, 2.718281828459045)"),
        input = "SELECT LOG(10)",
    );
    test_rewrite!(
        log10,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT LOG(10, 10.0)"),
        input = "SELECT LOG10(10)",
    );
    test_rewrite!(
        ltrim,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT TRIM(LEADING ' ' FROM ' stuff ')"),
        input = "SELECT LTRIM(' stuff ')",
    );
    test_rewrite!(
        ltrim_two_args_is_error,
        pass = ScalarFunctionsRewritePass,
        expected = Err(Error::IncorrectArgumentCount {
            name: "LTRIM",
            required: "1",
            found: 2
        }),
        input = "SELECT LTRIM(' stuff ', 'more stuff')",
    );
    test_rewrite!(
        rtrim,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT TRIM(TRAILING ' ' FROM ' stuff ')"),
        input = "SELECT RTRIM(' stuff ')",
    );
    test_rewrite!(
        rtrim_two_args_is_error,
        pass = ScalarFunctionsRewritePass,
        expected = Err(Error::IncorrectArgumentCount {
            name: "RTRIM",
            required: "1",
            found: 2
        }),
        input = "SELECT RTRIM(' stuff ', 'more stuff')",
    );
    test_rewrite!(
        timestamp_add,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT DATEADD(YEAR, 2, d)"),
        input = "SELECT TIMESTAMPADD(YEAR, 2, d)",
    );
    test_rewrite!(
        date_add_incorrect_arg_count,
        pass = ScalarFunctionsRewritePass,
        expected = Err(Error::IncorrectArgumentCount {
            name: "DATEADD",
            required: "3",
            found: 2
        }),
        input = "SELECT DATEADD(YEAR, 2)",
    );
    test_rewrite!(
        timestamp_diff,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT DATEDIFF(QUARTER, d1, d2, 'sunday')"),
        input = "SELECT TIMESTAMPDIFF(QUARTER, d1, d2)",
    );
    test_rewrite!(
        timestamp_diff_4_args,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT DATEDIFF(MILLISECOND, d1, d2, 'monday')"),
        input = "SELECT TIMESTAMPDIFF(SQL_TSI_FRAC_SECOND, d1, d2, 'monday')",
    );
    test_rewrite!(
        date_diff_incorrect_arg_count,
        pass = ScalarFunctionsRewritePass,
        expected = Err(Error::IncorrectArgumentCount {
            name: "DATEDIFF",
            required: "3 or 4",
            found: 2
        }),
        input = "SELECT DATEDIFF(QUARTER, d1)",
    );
    test_rewrite!(
        timestamp_trunc,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT DATETRUNC(YEAR, d, 'sunday')"),
        input = "SELECT TIMESTAMPTRUNC(YEAR, d)",
    );
    test_rewrite!(
        timestamp_trunc_3_args,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT DATETRUNC(YEAR, d, 'monday')"),
        input = "SELECT TIMESTAMPTRUNC(YEAR, d, 'monday')",
    );
    test_rewrite!(
        date_trunc_incorrect_arg_count,
        pass = ScalarFunctionsRewritePass,
        expected = Err(Error::IncorrectArgumentCount {
            name: "DATETRUNC",
            required: "2 or 3",
            found: 1
        }),
        input = "SELECT DATETRUNC(YEAR)",
    );
    test_rewrite!(
        year,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(YEAR FROM d)"),
        input = "SELECT YEAR(d)",
    );
    test_rewrite!(
        month,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MONTH FROM d)"),
        input = "SELECT MONTH(d)",
    );
    test_rewrite!(
        week,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(WEEK FROM d)"),
        input = "SELECT WEEK(d)",
    );
    test_rewrite!(
        dayofmonth,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY FROM d)"),
        input = "SELECT DAYOFMONTH(d)",
    );
    test_rewrite!(
        dayofyear,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY_OF_YEAR FROM d)"),
        input = "SELECT DAYOFYEAR(d)",
    );
    test_rewrite!(
        dayofweek,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY_OF_WEEK FROM d)"),
        input = "SELECT DAYOFWEEK(d)",
    );
    test_rewrite!(
        hour,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(HOUR FROM d)"),
        input = "SELECT HOUR(d)",
    );
    test_rewrite!(
        second,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(SECOND FROM d)"),
        input = "SELECT SECOND(d)",
    );
    test_rewrite!(
        millisecond,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MILLISECOND FROM d)"),
        input = "SELECT MILLISECOND(d)",
    );
    test_rewrite!(
        extract_sql_tsi_frac_second,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MILLISECOND FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_FRAC_SECOND FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_second,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(SECOND FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_SECOND FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_minute,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MINUTE FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_MINUTE FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_hour,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(HOUR FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_HOUR FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_day,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_DAY FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_dayofyear,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY_OF_YEAR FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_DAYOFYEAR FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_week,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(WEEK FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_WEEK FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_month,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MONTH FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_MONTH FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_quarter,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(QUARTER FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_QUARTER FROM d)",
    );
    test_rewrite!(
        extract_sql_tsi_year,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(YEAR FROM d)"),
        input = "SELECT EXTRACT(SQL_TSI_YEAR FROM d)",
    );
    test_rewrite!(
        extract_millisecond,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MILLISECOND FROM d)"),
        input = "SELECT EXTRACT(MILLISECOND FROM d)",
    );
    test_rewrite!(
        extract_second,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(SECOND FROM d)"),
        input = "SELECT EXTRACT(SECOND FROM d)",
    );
    test_rewrite!(
        extract_minute,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MINUTE FROM d)"),
        input = "SELECT EXTRACT(MINUTE FROM d)",
    );
    test_rewrite!(
        extract_hour,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(HOUR FROM d)"),
        input = "SELECT EXTRACT(HOUR FROM d)",
    );
    test_rewrite!(
        extract_day,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY FROM d)"),
        input = "SELECT EXTRACT(DAY FROM d)",
    );
    test_rewrite!(
        extract_day_of_year,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY_OF_YEAR FROM d)"),
        input = "SELECT EXTRACT(DAY_OF_YEAR FROM d)",
    );
    test_rewrite!(
        extract_dayofyear,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(DAY_OF_YEAR FROM d)"),
        input = "SELECT EXTRACT(DAYOFYEAR FROM d)",
    );
    test_rewrite!(
        extract_iso_weekday,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(ISO_WEEKDAY FROM d)"),
        input = "SELECT EXTRACT(ISO_WEEKDAY FROM d)",
    );
    test_rewrite!(
        extract_week,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(WEEK FROM d)"),
        input = "SELECT EXTRACT(WEEK FROM d)",
    );
    test_rewrite!(
        extract_iso_week,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(ISO_WEEK FROM d)"),
        input = "SELECT EXTRACT(ISO_WEEK FROM d)",
    );
    test_rewrite!(
        extract_month,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(MONTH FROM d)"),
        input = "SELECT EXTRACT(MONTH FROM d)",
    );
    test_rewrite!(
        extract_quarter,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(QUARTER FROM d)"),
        input = "SELECT EXTRACT(QUARTER FROM d)",
    );
    test_rewrite!(
        extract_year,
        pass = ScalarFunctionsRewritePass,
        expected = Ok("SELECT EXTRACT(YEAR FROM d)"),
        input = "SELECT EXTRACT(YEAR FROM d)",
    );
}
