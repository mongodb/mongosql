package evaluator_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/stretchr/testify/require"
)

func TestPushdownPlan(t *testing.T) {
	req := require.New(t)

	type test struct {
		name string
		sql  string
	}

	tests := []test{
		{"from_subquery", "select a, b from (select a, b from bar) b"},
		{"inner_join_and_predicate", "select * from bar a join foo b on a.a=b.a and a.a=b.f"},
		{"inner_join_equijoin", "select foo.a, bar.b from foo inner join bar on foo.a = bar.a"},
		{"inner_join_union", "select foo.a, bar.b from foo inner join bar on foo.a = bar.a " +
			"union select foo.a, bar.b from foo inner join bar on foo.a = bar.a"},
		{"inner_join_equijoin_filter", "select foo.a, bar.b from foo inner join bar on " +
			"foo.a = bar.a where foo.b = 10"},
		{"inner_join_equijoin_filter_and_predicate", "select foo.a, bar.b from foo inner join bar" +
			" on foo.a = bar.a where foo.b = 10 AND bar.b = 12"},
		{"inner_join_equijoin_filter_or_predicate", "select foo.a, bar.b from foo inner join bar" +
			" on foo.a = bar.a where foo.b = 10 OR bar.b = 12"},
		{"inner_join_equijoin_filter_and_or_predicate_0", "select foo.a, bar.b from foo inner " +
			"join bar on foo.a = bar.a where foo.b = 11 AND (foo.b = 10 OR bar.b = 12)"},
		{"inner_join_equijoin_filter_and_or_predicate_1", "select foo.a, bar.b from foo inner " +
			"join bar on foo.a = bar.a where (foo.b = 11 OR foo.b = 10) AND bar.b = 12"},
		{"inner_join_subquery", "select foo.a, b.b from foo join (select a, b from bar) b on " +
			"foo.a=b.a"},
		{"inner_join_subquery_with_predicate", "select foo.a, bar.b from foo inner join (select" +
			" bar.a, bar.b from bar where bar.b = 12) bar on foo.a = bar.a where bar.a = 10"},
		{"inner_join_select_same_column_name", "select foo.a, bar.a from foo inner join bar on " +
			"foo.a = bar.a"},
		{"inner_join_and_predicate_with_equijoin_part", "select foo.a, bar.b from foo inner join" +
			" bar on foo.a = bar.a AND foo.b > 10"},
		{"inner_join_and_predicate_with_equijoin_part_and_complex_other_part", "select foo.a, " +
			"bar.b from foo inner join bar on foo.a = bar.a AND foo.b > 10 AND (bar.b < 12 OR " +
			"bar.b > 10)"},
		{"inner_join_comma_operator", "select foo.a, bar.b from foo, bar where foo.a = bar.a"},
		{"inner_joins_subqueries_nested", "select * from (select foo.a from bar join (select " +
			"foo.a from foo) foo on foo.a=bar.b) x join (select g.a from bar join (select foo.a" +
			" from foo) g on g.a=bar.a) y on x.a=y.a"},
		{"left_join_inner_join_subqueries_nested", "select * from foo f left join (select b.b " +
			"from foo f join (select * from bar) b on f.a=b.a)  b on f.a=b.b"},
		{"right_join_inner_join_subqueries_nested", "select * from foo f right join (select b.b" +
			" from foo f join (select * from bar) b on f.a=b.a)  b on f.a=b.b"},
		{"join_nested_array_tables_0", "select * from foo f join merge m1 on f._id=m1._id join " +
			"(select * from foo) g on g.a=f.a join merge_d_a m2 on m2._id=m1._id and m2._id=g.a"},
		{"join_subqueries_where_limit", "select f.a from foo f join (select bar.a from bar) b " +
			"on f.a=b.a join (select foo.a from foo where foo.a > 4 limit 1) c on b.a=c.a and " +
			"f.a=c.a and f.b=b.a"},
		{"three_way_join_with_predicates_using_all_tables", "select * from foo f join merge m1 " +
			"on f._id=m1._id join merge_d_a m2 on m1._id=m2._id and f._id=m2._id"},
		{"three_way_join_with_predicates_using_some_tables", "select foo.c, bar.a, baz.b from " +
			"foo inner join bar on foo.a = bar.a inner join baz on bar.a = baz.a"},
		{"three_way_join_same_column_per_table", "select foo.a, bar.a, baz.a from foo inner " +
			"join bar on foo.a = bar.a inner join baz on bar.a = baz.a"},
		{"three_way_join_and_predicate_with_non_equality_part", "select foo.a, bar.a, baz.a " +
			"from bar inner join baz on baz.a = bar.a inner join foo on baz.a = foo.a and baz.a" +
			" > foo.c"},
		{"three_way_join_predicate_associativity", "select * from foo join (bar join baz on " +
			"bar.a = baz.a) on foo.a = bar.a"},
		{"join_nested_array_tables_1", "select * from foo f join merge m1 on f._id=m1._id join " +
			"merge_d_a m2 on m2._id=f._id and m2._id=m1._id"},
		{"nested_subquery_joins", "select f1.a, b1.b from foo f1 inner join (select b2.b, b2.a," +
			" b2._id from bar b2 join (select * from foo) f2 on f2._id = b2._id) b1 on b1._id " +
			"= f1._id"},
		{"inner_join_three_way_with_predicate_0", "select foo.a, bar.a, baz.a from foo inner " +
			"join bar on foo.a = bar.a inner join baz on bar.a = baz.a where foo.a = 10 AND bar.a" +
			" = 12 AND baz.a = 13"},
		{"inner_join_three_way_with_predicate_1", "select foo.a, bar.a, baz.a from foo inner " +
			"join bar on foo.a = bar.a inner join baz on bar.a = baz.a where (foo.a = 10 OR bar.a" +
			" = 11) AND bar.a = 12 AND baz.a = 13"},
		{"flip_join", "select * from foo r inner join merge_d_a a on r._id=a._id"},
		{"left_join_simple", "select foo.a, bar.b from foo left outer join bar on foo.a = bar.a"},
		{"left_join_with_filter", "select foo.a, bar.b from foo left outer join bar on foo.a = " +
			"bar.a where foo.a = 10 AND bar.b = 12"},
		{"left_join_with_non_eq_condition_on_left_table", "select foo.a, bar.b from foo left join" +
			" bar on foo.a = bar.a AND foo.b > 10"},
		{"left_join_with_non_eq_condition_on_right_table", "select foo.a, bar.b from foo left" +
			" join bar on foo.a = bar.a AND bar.b > 10"},
		{"left_join_three_way", "select foo.c, bar.a, baz.b from foo left join bar on foo.a = " +
			"bar.a left join baz on bar.a = baz.a"},
		{"right_non_equijoin_nopushdown", "select foo.a from foo right join bar on foo.a < bar.a"},
		{"right_join_simple", "select foo.a, bar.b from foo right outer join bar on foo.a = bar.a"},
		{"self_join_0", "select * from merge r left join merge_d_a a on r._id=a._id"},
		{"self_join_1", "select b._id, c._id from merge r inner join merge_b b on r._id=b._id " +
			"inner join merge_c c on b._id=c._id"},
		{"self_join_2", "select b._id, c._id from merge r left join merge_b b on r._id=b._id left" +
			" join merge_c c on b._id=c._id"},
		{"self_join_3", "select b._id, c._id from merge r left join merge_b b on r._id=b._id left" +
			" join merge_c c on r._id=c._id"},
		{"self_join_4", "select b._id, c._id from merge r left join merge_b b on r._id=b._id " +
			"inner join merge_c c on r._id=c._id left join merge_d_a a on r._id=a._id"},
		{"self_join_5", "select b._id, c._id from merge r inner join merge_b b on r._id=b._id " +
			"inner join merge_c c on r._id=c._id inner join merge_d_a a on r._id=a._id"},
		{"self_join_6", "select b._id, r._id from merge r inner join merge_d d on r._id=d._id " +
			"inner join merge_d_a a on r._id=a._id inner join merge_b b on r._id=b._id"},
		{"self_join_7", "select b._id, d._id from merge r inner join merge_b b on r._id=b._id " +
			"inner join merge_d d on r._id=d._id inner join merge_d_a a on r._id=a._id"},
		{"select_simple", "select a, b from foo"},
		{"select_correlated_subquery", "select a, (select foo.b from bar) from foo"},
		{"select_agg_from_subquery", "select count(*) from (select * from bar) foo"},
		{"where_simple", "select a from foo where a = 10"},
		{"where_and", "select a from foo where a = 10 AND b < c"},
		{"where_and_flipped", "select a from foo where b < c AND a = 10"},
		{"where_lt", "select a from foo where b < c"},
		{"where_nested_array_table", "select `d.a` from merge_d_a where `d.a` = 10"},
		{"where_nested_array_table_or", "select `d.a` from merge_d_a where `d.a` = 10 OR `d.a`" +
			" = 12"},
		{"where_array_table", "select c from merge_c where c = 10"},
		{"where_array_table_and", "select c from merge_c where c > 5 AND c < 10"},
		{"group_by_unprojected_column", "select a, b from foo group by c"},
		{"group_by_projected_column", "select a, b, c from foo group by c"},
		{"group_by_sum", "select a + b from foo group by a + b"},
		{"group_project_expression_on_group_key", "select a, b, c + a from foo group by c"},
		{"group_aggregate_0", "select max(a), max(b) from foo group by c"},
		{"group_aggregate_1", "select max(dt) from datetest"},
		{"group_aggregate_2", "select min(dt) from datetest"},
		{"group_aggregate_and_project_0", "select c, max(a), max(b) from foo group by c"},
		{"group_aggregate_and_project_1", "select a, max(b) from foo group by c"},
		{"group_aggregate_distinct_0", "select a, max(distinct b) from foo group by c"},
		{"group_aggregate_distinct_1", "select a, max(distinct b), c from foo group by c"},
		{"group_aggregate_in_expr_with_column_0", "select a + max(b) from foo group by c"},
		{"group_aggregate_in_expr_with_column_1", "select a + c + max(b) from foo group by c"},
		{"group_aggregate_distinct_with_expr_in_column_0", "select a + max(distinct b) from foo" +
			" group by c"},
		{"group_aggregate_distinct_with_expr_in_column_1", "select c + max(distinct b) from foo" +
			" group by c"},
		{"group_aggregate_expr_with_distinct_0", "select max(distinct a + b) from foo group by c"},
		{"group_aggregate_expr_with_distinct_1", "select a + max(distinct a + b) from foo group" +
			" by c"},
		{"aggregate_simple", "select sum(a) from foo"},
		{"count_star_optimized", "select count(*) from foo"},
		{"count_star_optimized_with_order", "select count(*) from foo order by 1"},
		{"count_star_non_optimized", "select count(*) from foo where true"},
		{"count_column", "select count(a) from foo"},
		{"count_distinct", "select count(distinct b) from foo"},
		{"group_having", "select max(a) from foo group by c having max(b) = 10"},
		{"order_simple", "select a from foo order by b"},
		{"order_inside_subquery", "(select a from foo order by b)"},
		{"order_subquery", "(select a from foo) order by a limit 1"},
		{"order_select_subquery_0", "select * from (select a from foo order by a limit 3) ut " +
			"order by a limit 1"},
		{"order_select_subquery_1", "select * from (select a from foo order by a limit 3) ut " +
			"order by a limit 1, 1"},
		{"order_multiple_desc", "select a from foo order by a, b desc"},
		{"order_by_aggregate_with_group", "select a from foo group by a order by max(b)"},
		{"testname", "select a from foo order by a > b"},
		{"limit_simple", "select a from foo limit 10"},
		{"limit_multiple", "select a from foo limit 10, 20"},
		{"limit_in_subquery", "(select a from foo limit 1)"},
		{"limit_subquery", "(select a from foo) limit 1"},
		{"mongo_filter_0", `select a from foo where filter='{"a": {"$gt": 3}}'`},
		{"mongo_filter_1", `select a from foo where filter='{"a": {"$elemMatch": {"$gte": 80,` +
			` "$lt": 85}}}' or b = 40`},
		{"no_column_ref_0", "select 1 from foo"},
		{"no_column_ref_1", "select 1 from foo where c>0"},
		{"no_column_ref_2", "select trim(concat(' Hi ', 'Ron ')) as tr, (1+(3*5))-4 as mt from " +
			"foo where c>0 order by tr"},
		{"no_column_ref_3", "select trim(concat(' Hi ', 'Ron ')) as tr, (1+(3*5))-4 as mt from" +
			" foo where c>0 group by tr"},
		{"no_column_ref_4", "select 1 from (select 1,2 from foo) as f"},
		{"no_column_ref_join_criteria_true", "select * from foo join bar on 1 = 1"},
		{"no_column_ref_join_criteria_false", "select * from foo join bar on 1 = 2"},
		{"no_column_ref_join_criteria_null", "select * from foo join bar on null"},
		{"no_column_ref_outer_join_criteria_false", "select * from foo left join bar on 1 = 2"},
		{"no_column_ref_outer_join_criteria_null", "select * from foo left join bar on null"},
		{"join_criteria_no_local_column_ref_left_join", "select * from foo left join bar on" +
			" bar.a = 1"},
		{"join_criteria_no_local_column_ref_right_join", "select * from foo right join bar on" +
			" foo.a = 1"},
		{"join_criteria_no_foreign_column_ref_right_join", "select * from foo right join bar on" +
			" bar.a = 1"},
		{"unique_field_gen_0", "select trim(''), ifnull(a, '') from foo"},
		{"unique_field_gen_1", "select trim(''), ifnull(a, ''), trim(' ') from foo"},
		{"unique_field_gen_2", "select a, b, trim('   ') from foo"},
		{"unique_field_gen_3", "select ifnull(a, ''), trim(''), a, trim(' ') from foo"},
		// these next five tests are not expected to push down, since they use no column refs
		{"unique_field_gen_4", "select trim('') from (select trim('') from foo) as subq"},
		{"unique_field_gen_5", "select trim('') from (select trim('') from (select trim('') from" +
			" foo) as subq1) as subq2"},
		{"unique_field_gen_6", "select trim('') from (select trim('') from (select trim('') from" +
			" (select trim('') from foo) as subq1) as subq2) as subq3"},
		{"unique_field_gen_7", "select trim(''), trim(' ') from foo inner join (select trim('')," +
			" trim(' ') from bar) as t2"},
		{"unique_field_gen_8", "select trim(''), trim(' '), trim('  ') from foo inner join " +
			"(select trim(''), trim(' '), trim('  ') from bar) as t2"},
		{"duplicate_pushdown_0", "select a, b as a from foo"},
		{"duplicate_pushdown_1", "select a, b as a, c as a from foo"},
		{"duplicate_pushdown_2", "select a, b as a, _id as a from foo"},
		{"duplicate_pushdown_3", "select a, b as a, e as a from foo"},
		{"optimal_cross_join_1", "select foo.a from foo cross join bar where foo.a = bar.b"},
		{"optimal_cross_join_2", "select foo.a from foo cross join bar where foo.a = bar.b and " +
			"foo.a = 4"},
		{"suboptimal_cross_join_1", "select foo.a from foo cross join bar cross join baz where " +
			"foo.a = baz.b"},
		{"suboptimal_cross_join_2", "select foo.a from foo cross join bar cross join baz where " +
			"foo.a = baz.b and foo.a = 4"},
		{"suboptimal_cross_join_3", "select foo.a from foo cross join (select bar.b from bar) " +
			"s cross join baz where foo.a = s.b and s.b = 4 and foo.a < 33"},
		{"suboptimal_cross_join_4", "select foo.a from foo cross join bar cross join baz where " +
			"foo.a < bar.b + 3 and foo.a < 5"},
		{"suboptimal_cross_join_5", "select foo.a from foo,bar,baz,merge where bar.b > merge.a " +
			"and foo.a = bar.b + merge.a and foo.a = baz.b"},
		{"suboptimal_cross_join_6", "select foo.a from foo cross join bar cross join baz where " +
			"foo.a = baz.b union select foo.a from foo cross join bar cross join baz where " +
			"foo.a = baz.b"},
		{"suboptimal_cross_join_7", "select * from (select foo.a from foo, bar, baz where " +
			"foo.a = baz.b) res"},
		{"suboptimal_cross_join_8", "select * from foo, bar, (select foo.a from foo, bar, baz " +
			"where foo.a = baz.b) res where foo.a = res.a"},
		{"suboptimal_cross_join_9", "select * from foo inner join bar, (select foo.a from foo, " +
			" bar, baz where foo.a = baz.b) res where foo.a = res.a"},
		{"suboptimal_cross_join_ultimate", "select * from " +
			"foo, " +
			"foo foo1, " +
			"foo foo2, " +
			"foo foo3, " +
			"foo foo4, " +
			"foo foo5, " +
			"foo foo6, " +
			"foo foo7, " +
			"foo foo8, " +
			"foo foo9, " +
			"foo foo10, " +
			"foo foo11, " +
			"foo foo12, " +
			"foo foo13, " +
			"foo foo14, " +
			"foo foo15, " +
			"foo foo16, " +
			"foo foo17, " +
			"foo foo18, " +
			"foo foo19, " +
			"foo foo20, " +
			"foo foo21, " +
			"foo foo22, " +
			"foo foo23, " +
			"foo foo24, " +
			"foo foo25, " +
			"foo foo26, " +
			"foo foo27, " +
			"foo foo28, " +
			"foo foo29, " +
			"foo foo30, " +
			"foo foo31, " +
			"foo foo32, " +
			"foo foo33, " +
			"foo foo34, " +
			"foo foo35, " +
			"foo foo36, " +
			"foo foo37, " +
			"foo foo38, " +
			"foo foo39, " +
			"foo foo40, " +
			"foo foo41, " +
			"foo foo42, " +
			"foo foo43, " +
			"foo foo44, " +
			"foo foo45, " +
			"foo foo46, " +
			"foo foo47, " +
			"foo foo48, " +
			"foo foo49, " +
			"foo foo50 " +
			"where " +
			"foo48.a = foo28.a AND " +
			"foo21.a = foo7.a AND " +
			"foo19.a = foo34.a AND " +
			"foo37.a = foo4.a AND " +
			"foo41.a = foo43.a AND " +
			"foo30.a = foo31.a AND " +
			"foo24.a = foo6.a AND " +
			"foo44.a = foo14.a AND " +
			"foo13.a = foo26.a AND " +
			"foo49.a = foo5.a AND " +
			"foo17.a = foo24.a AND " +
			"foo42.a = foo15.a AND " +
			"foo39.a = foo41.a AND " +
			"foo14.a = foo2.a AND " +
			"foo40.a = foo47.a AND " +
			"foo20.a = foo11.a AND " +
			"foo2.a = foo18.a AND " +
			"foo35.a = foo30.a AND " +
			"foo10.a = foo12.a AND " +
			"foo29.a = foo13.a AND " +
			"foo32.a = foo45.a AND " +
			"foo47.a = foo40.a AND " +
			"foo46.a = foo39.a AND " +
			"foo50.a = foo42.a AND " +
			"foo45.a = foo49.a AND " +
			"foo6.a = foo35.a AND " +
			"foo33.a = foo1.a AND " +
			"foo36.a = foo29.a AND " +
			"foo27.a = foo46.a AND " +
			"foo27.a = 4 AND " +
			"foo25.a = foo23.a"},
		{"non_equijoin_0", "select foo.a from foo inner join bar on foo.a < bar.b"},
		{"non_equijoin_1", "select foo.a from foo inner join bar on foo.a < foo.b"},
		{"non_equijoin_2", "select foo.a from foo, bar where foo.a < bar.a"},
		{"non_equijoin_3", "select foo.a from foo left join bar on foo.a < bar.a"},
		{"non_equijoin_4", "select foo.a from foo right join bar on foo.a < bar.a"},
		{"non_equijoin_5", "select foo.a from foo right join bar on foo.a < bar.a and " +
			"floor(foo.a) = ceil(bar.a)"},
		{"equijoin_subquery", "select foo.a, b.b from foo, (select a, b from bar) b where" +
			" foo.a = b.a"},
		{"repeat_arithmetic_0", "select a+(b+c)+a+(b+c+(a+a)) from foo"},
		{"repeat_arithmetic_1", "select a*(b*c)*a*(b*c*(a*a)) from foo"},
		{"repeat_arithmetic_2", "select a*(b*c)+a+(b*c*(a*a)) from foo"},
		{"repeat_arithmetic_3", "select (a+(b+c))*a*(b+c+(a+a)) from foo"},
		{"nse_join_criterion", "SELECT * FROM foo f JOIN bar b ON f.a <=> b.a"},
		{"nse_inner_join_criterion", "SELECT * FROM foo f INNER JOIN bar b ON f.a <=> b.a"},
		{"nse_left_join_criterion", "SELECT * FROM foo f LEFT JOIN bar b ON f.a <=> b.a"},
		{"nse_right_join_criterion", "SELECT * FROM foo f RIGHT JOIN bar b ON f.a <=> b.a"},
		{"nse_cross_join_criterion", "SELECT * FROM foo f RIGHT JOIN bar b ON f.a <=> b.a"},
	}

	// open the file with the cached test results
	cacheFile := "testdata/test_pushdown.json"
	file, err := os.Open(cacheFile)
	req.Nil(err)

	// read the contents of the cache file and close it
	data, err := ioutil.ReadAll(file)
	req.Nil(err, "failed to read cached results file")
	err = file.Close()
	req.Nil(err, "failed to close cached results file")

	// unmarshal the cached results into a two-dimensional
	// map, which is structured as follows:
	// {
	//   <mongodb_version>: {
	//     <testcase_name>: <pushdown_pipeline_as_json_string>,
	//   }
	// }
	cache := make(map[string]map[string]string)
	err = json.Unmarshal(data, &cache)
	req.Nil(err, "failed to unmarshal cached results json")

	// define the MongoDB versions for which we want to test pushdown
	versions := [][]uint8{
		{3, 2, 0},
		{3, 4, 0},
		{3, 6, 0},
	}

	// run a subtest for each version
	for _, version := range versions {
		v := formatVersion(version)
		t.Run(v, func(t *testing.T) {
			if cache[v] == nil {
				cache[v] = make(map[string]string)
			}

			// run a subtest for each query
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					req = require.New(t)
					actual := optimizePlan(t, version, test.sql)
					if *update {
						cache[v][test.name] = actual
						return
					}
					expected, ok := cache[v][test.name]
					req.True(ok, "test case not found in cache")
					if expected == "" || actual == "" {
						req.Equal(expected, actual, "result does not match cached result")
					} else {
						req.JSONEq(expected, actual, "result does not match cached result")
					}
				})
			}
		})
	}

	if *update {
		cacheBytes, err := json.MarshalIndent(cache, "", "    ")
		req.Nil(err)
		err = ioutil.WriteFile(cacheFile, cacheBytes, os.ModePerm)
		req.Nil(err)
	}
}

func optimizePlan(t *testing.T, version []uint8, sql string) string {
	req := require.New(t)

	testSchema := evaluator.MustLoadSchema(optimizerTestSchema)

	testInfo := evaluator.GetMongoDBInfo(version, testSchema, mongodb.AllPrivileges)
	testVariables := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	statement, err := parser.Parse(sql)
	req.Nil(err, "failed to parse statement")

	plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
	req.Nil(err, "failed to algebrize query")

	actualPlan := evaluator.OptimizePlan(createTestConnectionCtx(testInfo, version...), plan)

	var actual string
	ms, ok := actualPlan.(*evaluator.MongoSourceStage)
	if ok {
		converted, err := bsonutil.GetBSONValueAsJSON(ms.Pipeline(), true)
		req.Nil(err, "failed to get pipeline as json")

		actualBytes, err := json.Marshal(converted)
		req.Nil(err, "failed to marshal pipeline to json")

		actual = string(actualBytes)
	}

	return actual
}
