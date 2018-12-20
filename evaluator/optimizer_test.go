package evaluator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/kr/pretty"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

// Fully pushed-down queries are covered by TestPushdownPlan in
// optimizer_pushdown_test.go. This test covers the remaining cases, testing
// how we push down queries that are only partially pushed down on certain
// MongoDB versions.
func TestOptimizePartialPushdown(t *testing.T) {

	type test struct {
		name     string
		sql      string
		expected [][]bson.D
		versions []string
	}

	tests := []test{

		{
			name:     "count_star",
			sql:      "select count(*) from foo",
			expected: [][]bson.D{},
		},
		{
			name:     "count_star_with_order",
			sql:      "select count(*) from foo order by 1",
			expected: [][]bson.D{},
		},
		{
			name:     "huge_limit",
			sql:      "select a from foo limit 18446744073709551614",
			expected: [][]bson.D{},
		},
		{
			name: "nopushdown_scalar_function_in_select",
			sql:  "select a+a, nopushdown(a+a) from foo",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(
						bsonutil.NewDocElem("$project", bsonutil.NewD(
							bsonutil.NewDocElem("_id", 0),

							// Only one add should be push down.
							bsonutil.NewDocElem("a", "$a"),
							bsonutil.NewDocElem("test_DOT_foo_DOT_a+test_DOT_foo_DOT_a", bsonutil.NewD(
								bsonutil.NewDocElem("$add", bsonutil.NewArray(
									"$a",
									"$a",
								)),
							)),
						)),
					),
				),
			},
		},
		{
			name: "nopushdown_scalar_function_in_where",
			sql:  "select a+a from foo where nopushdown(b=1)",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewD(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
						bsonutil.NewDocElem("test_DOT_foo_DOT_b", "$b"),
					)),
					),
				),
			},
		},
		{
			name: "nopushdown_scalar_function_in_orderby",
			sql:  "select a+a from foo order by nopushdown(a=1)",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project",
						bsonutil.NewD(
							bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"))))),
			},
		},
		{
			name: "nopushdown_scalar_function_in_orderby_after_where",
			sql:  "select a+a from foo where a > 3 order by nopushdown(a=1)",
			expected: [][]bson.D{
				bsonutil.NewDArray(bsonutil.NewD(
					bsonutil.NewDocElem("$match", bsonutil.NewD(
						bsonutil.NewDocElem("a", bsonutil.NewD(bsonutil.NewDocElem("$gt", int64(3))))))), bsonutil.NewD(
					bsonutil.NewDocElem("$project", bsonutil.NewD(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a")))))},
		},
		{
			name: "nopushdown_scalar_function_in_groupby",
			sql:  "select a+a from foo group by nopushdown(a)",
			expected: [][]bson.D{
				bsonutil.NewDArray(bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewD(
					bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a")))))},
		},
		{
			name: "nopushdown_scalar_function_in_join",
			sql: "select * from (select a+a as a from bar) a " +
				" inner join (select a+a as a, concat(b, b) from bar) b on nopushdown(a.a) = b.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(bsonutil.NewD(
					bsonutil.NewDocElem("$project", bsonutil.NewD(
						bsonutil.NewDocElem("test_DOT_bar_DOT_a+test_DOT_bar_DOT_a", bsonutil.NewD(
							bsonutil.NewDocElem("$add", bsonutil.NewArray(
								"$a",
								"$a",
							))))))), bsonutil.NewD(
					bsonutil.NewDocElem("$project", bsonutil.NewD(
						bsonutil.NewDocElem("test_DOT_a_DOT_a",
							"$test_DOT_bar_DOT_a+test_DOT_bar_DOT_a"),
					)),
				)), bsonutil.NewDArray(bsonutil.NewD(
					bsonutil.NewDocElem("$project", bsonutil.NewD(bsonutil.NewDocElem("b", "$b"),
						bsonutil.NewDocElem("test_DOT_bar_DOT_a+test_DOT_bar_DOT_a", bsonutil.NewD(
							bsonutil.NewDocElem("$add", bsonutil.NewArray(
								"$a",
								"$a",
							)))),
					)))),
			},
		},
		{
			name:     "inner_joins_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql: "select * from (select foo.a from bar join (select foo.a from foo) foo on" +
				" foo.a=bar.b) x join (select g.a from bar join (select foo.a from foo) g on " +
				"g.a=bar.a) y on x.a=y.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$match",
						bsonutil.NewM(
							bsonutil.NewDocElem("test_DOT_foo_DOT_a",
								bsonutil.NewM(bsonutil.NewDocElem("$ne", nil)))))),
					bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewM(
						bsonutil.NewDocElem("from", "bar"),
						bsonutil.NewDocElem("localField", "test_DOT_foo_DOT_a"),
						bsonutil.NewDocElem("foreignField", "b"),
						bsonutil.NewDocElem("as", "__joined_bar"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewM(
						bsonutil.NewDocElem("preserveNullAndEmptyArrays", false),
						bsonutil.NewDocElem("path", "$__joined_bar"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$test_DOT_foo_DOT_a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_x_DOT_a", "$test_DOT_foo_DOT_a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$match",
						bsonutil.NewM(bsonutil.NewDocElem("test_DOT_foo_DOT_a",
							bsonutil.NewM(bsonutil.NewDocElem("$ne", nil)))))),
					bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewM(
						bsonutil.NewDocElem("from", "bar"),
						bsonutil.NewDocElem("localField", "test_DOT_foo_DOT_a"),
						bsonutil.NewDocElem("foreignField", "a"),
						bsonutil.NewDocElem("as", "__joined_bar"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewM(
						bsonutil.NewDocElem("preserveNullAndEmptyArrays", false),
						bsonutil.NewDocElem("path", "$__joined_bar"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_g_DOT_a", "$test_DOT_foo_DOT_a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_y_DOT_a", "$test_DOT_g_DOT_a"),
					)),
					),
				),
			}},

		{
			name:     "left_join_inner_join_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql: "select * from foo f left join (select b.b from foo f join (select * from " +
				"bar) b on f.a=b.a)  b on f.a=b.b",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_f_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_f_DOT_a", "$a"),
						bsonutil.NewDocElem("test_DOT_f_DOT_b", "$b"),
						bsonutil.NewDocElem("test_DOT_f_DOT_c", "$c"),
						bsonutil.NewDocElem("test_DOT_f_DOT_e", "$d.e"),
						bsonutil.NewDocElem("test_DOT_f_DOT_f", "$d.f"),
						bsonutil.NewDocElem("test_DOT_f_DOT_g", "$g"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_bar_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
						bsonutil.NewDocElem("test_DOT_bar_DOT_b", "$b"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$match",
						bsonutil.NewM(bsonutil.NewDocElem("test_DOT_bar_DOT_a",
							bsonutil.NewM(bsonutil.NewDocElem("$ne", nil)))))),
					bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewM(
						bsonutil.NewDocElem("from", "foo"),
						bsonutil.NewDocElem("localField", "test_DOT_bar_DOT_a"),
						bsonutil.NewDocElem("foreignField", "a"),
						bsonutil.NewDocElem("as", "__joined_f"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewM(
						bsonutil.NewDocElem("path", "$__joined_f"),
						bsonutil.NewDocElem("preserveNullAndEmptyArrays", false),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_b_DOT_b", "$test_DOT_bar_DOT_b"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_b_DOT_b", "$test_DOT_b_DOT_b"),
					)),
					),
				),
			}},

		{
			name: "join_nested_array_tables_0",
			sql: "select * from foo f join merge m1 on f._id=m1._id join (select * from foo) g" +
				" on g.a=f.a join merge_d_a m2 on m2._id=m1._id and m2._id=g.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(
						bsonutil.NewDocElem("includeArrayIndex", "d_idx"),
						bsonutil.NewDocElem("path", "$d"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(
						bsonutil.NewDocElem("includeArrayIndex", "d.a_idx"),
						bsonutil.NewDocElem("path", "$d.a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$match",
						bsonutil.NewM(bsonutil.NewDocElem("_id",
							bsonutil.NewM(bsonutil.NewDocElem("$ne", nil)))))),
					bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewM(
						bsonutil.NewDocElem("from", "foo"),
						bsonutil.NewDocElem("localField", "_id"),
						bsonutil.NewDocElem("foreignField", "_id"),
						bsonutil.NewDocElem("as", "__joined_f"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewM(
						bsonutil.NewDocElem("path", "$__joined_f"),
						bsonutil.NewDocElem("preserveNullAndEmptyArrays", false),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_f_DOT__id", "$__joined_f._id"),
						bsonutil.NewDocElem("test_DOT_f_DOT_a", "$__joined_f.a"),
						bsonutil.NewDocElem("test_DOT_f_DOT_b", "$__joined_f.b"),
						bsonutil.NewDocElem("test_DOT_f_DOT_c", "$__joined_f.c"),
						bsonutil.NewDocElem("test_DOT_f_DOT_e", "$__joined_f.d.e"),
						bsonutil.NewDocElem("test_DOT_f_DOT_f", "$__joined_f.d.f"),
						bsonutil.NewDocElem("test_DOT_f_DOT_g", "$__joined_f.g"),
						bsonutil.NewDocElem("test_DOT_m1_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_m1_DOT_a", "$a"),
						bsonutil.NewDocElem("test_DOT_m2_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_m2_DOT_d_DOT_a", "$d.a"),
						bsonutil.NewDocElem("test_DOT_m2_DOT_d_DOT_a_idx", "$d.a_idx"),
						bsonutil.NewDocElem("test_DOT_m2_DOT_d_idx", "$d_idx"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
						bsonutil.NewDocElem("test_DOT_foo_DOT_b", "$b"),
						bsonutil.NewDocElem("test_DOT_foo_DOT_c", "$c"),
						bsonutil.NewDocElem("test_DOT_foo_DOT_e", "$d.e"),
						bsonutil.NewDocElem("test_DOT_foo_DOT_f", "$d.f"),
						bsonutil.NewDocElem("test_DOT_foo_DOT_g", "$g"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_g_DOT__id", "$test_DOT_foo_DOT__id"),
						bsonutil.NewDocElem("test_DOT_g_DOT_a", "$test_DOT_foo_DOT_a"),
						bsonutil.NewDocElem("test_DOT_g_DOT_b", "$test_DOT_foo_DOT_b"),
						bsonutil.NewDocElem("test_DOT_g_DOT_c", "$test_DOT_foo_DOT_c"),
						bsonutil.NewDocElem("test_DOT_g_DOT_e", "$test_DOT_foo_DOT_e"),
						bsonutil.NewDocElem("test_DOT_g_DOT_f", "$test_DOT_foo_DOT_f"),
						bsonutil.NewDocElem("test_DOT_g_DOT_g", "$test_DOT_foo_DOT_g"),
					)),
					),
				),
			}},

		{
			name:     "join_subqueries_where_limit",
			versions: []string{"3.2", "3.4"},
			sql: "select f.a from foo f join (select bar.a from bar) b on f.a=b.a join " +
				"(select foo.a from foo where foo.a > 4 limit 1) c on b.a=c.a and f.a=c.a and " +
				"f.b=b.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$match",
						bsonutil.NewM(bsonutil.NewDocElem("a",
							bsonutil.NewM(bsonutil.NewDocElem("$gt", int64(4))))))),
					bsonutil.NewD(bsonutil.NewDocElem("$limit", int64(1))),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_c_DOT_a", "$test_DOT_foo_DOT_a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_b_DOT_a", "$test_DOT_bar_DOT_a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_f_DOT_a", "$a"),
						bsonutil.NewDocElem("test_DOT_f_DOT_b", "$b"),
					)),
					),
				),
			}},
		{
			name:     "right_non_equijoin",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
					)),
					),
				),
			},
		},

		{
			name:     "self_join_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select * from merge r left join merge_d_a a on r._id=a._id",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_r_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_r_DOT_a", "$a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(
						bsonutil.NewDocElem("includeArrayIndex", "d_idx"),
						bsonutil.NewDocElem("path", "$d"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(
						bsonutil.NewDocElem("includeArrayIndex", "d.a_idx"),
						bsonutil.NewDocElem("path", "$d.a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_a_DOT_d_idx", "$d_idx"),
						bsonutil.NewDocElem("test_DOT_a_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_a_DOT_d_DOT_a", "$d.a"),
						bsonutil.NewDocElem("test_DOT_a_DOT_d_DOT_a_idx", "$d.a_idx"),
					)),
					),
				),
			},
		},

		{
			name:     "self_join_4",
			versions: []string{"3.4"},
			sql: "select b._id, c._id from merge r left join merge_b b on r._id=b._id inner" +
				" join merge_c c on r._id=c._id left join merge_d_a a on r._id=a._id",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$addFields", bsonutil.NewM(
						bsonutil.NewDocElem("_id_0", bsonutil.NewD(bsonutil.NewDocElem("$cond", bsonutil.NewArray(
							bsonutil.NewD(bsonutil.NewDocElem("$or", bsonutil.NewArray(
								bsonutil.NewD(bsonutil.NewDocElem("$lte", bsonutil.NewArray(
									"$b",
									interface{}(nil),
								)),
								),
								bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
									"$b",
									bsonutil.NewArray(),
								)),
								),
							))),
							interface{}(nil),
							"$_id",
						)))))),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(bsonutil.NewDocElem("includeArrayIndex", "b_idx"),
						bsonutil.NewDocElem("path", "$b"),
						bsonutil.NewDocElem("preserveNullAndEmptyArrays",
							true),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(bsonutil.NewDocElem("includeArrayIndex", "c_idx"),
						bsonutil.NewDocElem("path", "$c"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(bsonutil.NewDocElem("test_DOT_b_DOT__id", "$_id_0"),
						bsonutil.NewDocElem("test_DOT_c_DOT__id", "$_id"),
						bsonutil.NewDocElem("test_DOT_r_DOT__id", "$_id"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(bsonutil.NewDocElem("includeArrayIndex", "d_idx"),
						bsonutil.NewDocElem("path", "$d"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewD(bsonutil.NewDocElem("includeArrayIndex", "d.a_idx"),
						bsonutil.NewDocElem("path", "$d.a"),
					)),
					),
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(bsonutil.NewDocElem("test_DOT_a_DOT__id", "$_id")))),
				),
			},
		},

		{
			name:     "non_equijoin_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo inner join bar on foo.a < bar.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
					)),
					),
				),
			}},
		{
			name:     "non_equijoin_2",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo, bar where foo.a < bar.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
				),
			}},
		{
			name:     "non_equijoin_3",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo left join bar on foo.a < bar.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
					)),
					),
				),
			}},
		{
			name:     "non_equijoin_4",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: [][]bson.D{
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
					)),
					),
				),
				bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
						bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
					)),
					),
				),
			}},
	}

	versionByStr := map[string][]uint8{
		"3.2": {3, 2, 0},
		"3.4": {3, 4, 0},
		"3.6": {3, 6, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			versions := test.versions
			if len(versions) == 0 {
				versions = []string{"3.2", "3.4", "3.6"}
			}

			for _, version := range versions {
				t.Run(version, func(t *testing.T) {
					req := require.New(t)

					testSchema := evaluator.MustLoadSchema(optimizerTestSchema)

					testInfo := evaluator.GetMongoDBInfo(versionByStr[version], testSchema, mongodb.AllPrivileges)
					testVariables := evaluator.CreateTestVariables(testInfo)
					testSchemaCatalog := evaluator.GetCatalog(testSchema, testVariables, testInfo)
					defaultDbName := "test"

					statement, err := parser.Parse(test.sql)
					req.Nil(err, "failed to parse statement")

					rCfg := evaluator.NewRewriterConfig(log.GlobalLogger(), false)

					rewritten, err := evaluator.RewriteQuery(rCfg, statement)
					req.Nil(err, "failed to rewrite query")

					aCfg := createAlgebrizerCfg(defaultDbName, testSchemaCatalog)
					plan, err := evaluator.AlgebrizeQuery(aCfg, rewritten)

					req.Nil(err, "failed to algebrize query")

					eCfg := createExecutionCfg("test_db_name", 0, versionByStr[version])
					oCfg := createOptimizerCfg(collation.Default, eCfg)
					optimizedPlan, err := evaluator.OptimizePlan(context.Background(), oCfg, plan)
					req.Nil(err, "failed to optimize plan")

					pCfg := createPushdownCfg(versionByStr[version])
					pushedDown, err := evaluator.PushdownPlan(pCfg, optimizedPlan)

					var actualPlan evaluator.PlanStage
					if err != nil && !evaluator.IsNonFatalPushdownError(err) {
						actualPlan = optimizedPlan
					} else {
						actualPlan = pushedDown
					}

					actual := evaluator.GetNodePipeline(actualPlan)
					actual = bsonutil.NormalizeBSON(actual).([][]bson.D)
					expected := bsonutil.NormalizeBSON(test.expected).([][]bson.D)

					req.Equalf(len(expected), len(actual),
						"expected %d pipelines in query plan, found %d\nexpected pipelines: "+
							"%#v\nactual pipelines: %#v\nactual plan:\n%s",
						len(expected), len(actual), test.expected, actual,
						evaluator.PrettyPrintPlan(actualPlan))

					diff := ShouldResembleDiffed(actual, expected)
					req.Emptyf(diff, "expected pipeline diff to be empty\nexpected: %#v\nactual:"+
						" %#v\n", expected, actual)

				})
			}
		})
	}

}

var optimizerTestSchema = []byte(`
schema:
-
  db: test
  tables:
  -
     table: datetest
     collection: datetest
     columns:
     -
        Name: dt
        MongoType: date
        SqlName: dt
        SqlType: date

  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: c
        MongoType: int
        SqlType: int
     -
        Name: d.e
        MongoType: int
        SqlName: e
        SqlType: int
     -
        Name: d.f
        MongoType: int
        SqlName: f
        SqlType: int
     -
        Name: g
        MongoType: bool
        SqlName: g
        SqlType: boolean
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
     -
        Name: filter
        MongoType: mongo.Filter
        SqlName: filter
        SqlType: varchar
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
     table: baz
     collection: baz
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
    table: merge
    collection: merge
    pipeline: []
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: a
      MongoType: float64
      SqlName: a
      SqlType: float64
  - table: merge_b
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: b_idx
        path: $b
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: b
      MongoType: float64
      SqlName: b
      SqlType: float64
    - Name: b_idx
      MongoType: int
      SqlName: b_idx
      SqlType: int
  - table: merge_c
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: c_idx
        path: $c
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: c
      MongoType: float64
      SqlName: c
      SqlType: float64
    - Name: c_idx
      MongoType: int
      SqlName: c_idx
      SqlType: int
  - table: merge_d
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: d_idx
        path: $d
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: d_idx
      MongoType: int
      SqlName: d_idx
      SqlType: int
  - table: merge_d_a
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: d_idx
        path: $d
    - $unwind:
        includeArrayIndex: d.a_idx
        path: $d.a
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: d.a
      MongoType: float64
      SqlName: d.a
      SqlType: float64
    - Name: d.a_idx
      MongoType: int
      SqlName: d.a_idx
      SqlType: int
    - Name: d_idx
      MongoType: int
      SqlName: d_idx
      SqlType: int
`)

func TestPushdownSharding(t *testing.T) {
	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := getMongoDBInfoWithShardedCollection(nil, testSchema, mongodb.AllPrivileges, "foo")
	testVariables := evaluator.CreateTestVariables(testInfo)
	testSchemaCatalog := evaluator.GetCatalog(testSchema, testVariables, testInfo)
	defaultDbName := "test"
	test := func(sql string, expected ...[]bson.D) {
		t.Run(sql, func(t *testing.T) {
			req := require.New(t)

			statement, err := parser.Parse(sql)
			req.NoError(err)

			rCfg := evaluator.NewRewriterConfig(log.GlobalLogger(), false)

			rewritten, err := evaluator.RewriteQuery(rCfg, statement)
			req.NoError(err, "failed to rewrite query")

			aCfg := createAlgebrizerCfg(defaultDbName, testSchemaCatalog)
			plan, err := evaluator.AlgebrizeQuery(aCfg, rewritten)

			req.NoError(err)

			version := []uint8{3, 4, 0}

			eCfg := createExecutionCfg("test_db", 0, version)
			oCfg := createOptimizerCfg(collation.Default, eCfg)
			optimized, err := evaluator.OptimizePlan(context.Background(), oCfg, plan)
			req.NoError(err)

			pCfg := createPushdownCfg(version)
			pushedDown, err := evaluator.PushdownPlan(pCfg, optimized)

			var actualPlan evaluator.PlanStage
			if err != nil && !evaluator.IsNonFatalPushdownError(err) {
				actualPlan = optimized
			} else {
				actualPlan = pushedDown
			}

			actual := evaluator.GetNodePipeline(actualPlan)
			actual, expected = bsonutil.NormalizeBSON(actual).([][]bson.D),
				bsonutil.NormalizeBSON(expected).([][]bson.D)

			v := ShouldResembleDiffed(actual, expected)
			if v != "" {
				fmt.Printf("\n SQL: %v", sql)
				fmt.Printf("\n ACTUAL: %#v", pretty.Formatter(actual))
				fmt.Printf("\n EXPECTED: %#v", pretty.Formatter(expected))
			}
			req.Zero(v)
		})
	}

	// should not push down because the from collection is sharded.
	test("select * from bar left join foo on bar.a=foo.a and bar.a=foo.f", bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
			bsonutil.NewDocElem("test_DOT_bar_DOT_b", "$b"),
			bsonutil.NewDocElem("test_DOT_bar_DOT__id", "$_id"),
			bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
		)),
		),
	), bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("$project", bsonutil.NewM(
				bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_b", "$b"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_c", "$c"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_e", "$d.e"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_g", "$g"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_f", "$d.f"),
				bsonutil.NewDocElem("test_DOT_foo_DOT__id", "$_id"),
			)),
		),
	))
	test("select * from bar right join foo on bar.a=foo.a and bar.a=foo.f", bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewM(
			bsonutil.NewDocElem("from", "bar"),
			bsonutil.NewDocElem("localField", "a"),
			bsonutil.NewDocElem("foreignField", "a"),
			bsonutil.NewDocElem("as", "__joined_bar"),
		)),
		),
		bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
			bsonutil.NewDocElem("c", 1),
			bsonutil.NewDocElem("d.f", 1),
			bsonutil.NewDocElem("g", 1),
			bsonutil.NewDocElem("_id", 1),
			bsonutil.NewDocElem("filter", 1),
			bsonutil.NewDocElem("__joined_bar", bsonutil.NewM(
				bsonutil.NewDocElem("$cond", bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
						bsonutil.NewM(bsonutil.NewDocElem("$ifNull", bsonutil.NewArray(
							"$a",
							nil,
						))),
						nil,
					)),
					),
					bsonutil.NewM(bsonutil.NewDocElem("$literal", bsonutil.NewArray())),
					"$__joined_bar",
				)),
			)),
			bsonutil.NewDocElem("a", 1),
			bsonutil.NewDocElem("b", 1),
			bsonutil.NewDocElem("d.e", 1),
		)),
		),
		bsonutil.NewD(bsonutil.NewDocElem("$addFields", bsonutil.NewM(bsonutil.NewDocElem("__joined_bar", bsonutil.NewM(
			bsonutil.NewDocElem("$filter", bsonutil.NewM(
				bsonutil.NewDocElem("cond", bsonutil.NewM(
					bsonutil.NewDocElem("$let", bsonutil.NewM(
						bsonutil.NewDocElem("vars", bsonutil.NewM(
							bsonutil.NewDocElem("left", "$$this.a"), bsonutil.NewDocElem("right", "$d.f"))),
						bsonutil.NewDocElem("in", bsonutil.NewM(
							bsonutil.NewDocElem("$cond", bsonutil.NewArray(
								bsonutil.NewM(
									bsonutil.NewDocElem("$or", bsonutil.NewArray(
										bsonutil.NewM(
											bsonutil.NewDocElem("$eq", bsonutil.NewArray(
												bsonutil.NewM(
													bsonutil.NewDocElem("$ifNull", bsonutil.NewArray(
														"$$left",
														nil,
													)),
												),
												nil,
											)),
										),
										bsonutil.NewM(
											bsonutil.NewDocElem("$eq", bsonutil.NewArray(
												bsonutil.NewM(
													bsonutil.NewDocElem("$ifNull", bsonutil.NewArray(
														"$$right",
														nil,
													))),
												nil,
											))),
									))),
								nil,
								bsonutil.NewM(
									bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$left",
										"$$right",
									))),
							)))),
					)))),
				bsonutil.NewDocElem("input", "$__joined_bar"),
				bsonutil.NewDocElem("as", "this"),
			))))))),
		bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewM(
			bsonutil.NewDocElem("path", "$__joined_bar"),
			bsonutil.NewDocElem("preserveNullAndEmptyArrays", true),
		)),
		),
		bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
			bsonutil.NewDocElem("test_DOT_bar_DOT_b", "$__joined_bar.b"),
			bsonutil.NewDocElem("test_DOT_foo_DOT_f", "$d.f"),
			bsonutil.NewDocElem("test_DOT_foo_DOT_c", "$c"),
			bsonutil.NewDocElem("test_DOT_foo_DOT_e", "$d.e"),
			bsonutil.NewDocElem("test_DOT_foo_DOT_g", "$g"),
			bsonutil.NewDocElem("test_DOT_foo_DOT__id", "$_id"),
			bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$__joined_bar.a"),
			bsonutil.NewDocElem("test_DOT_bar_DOT__id", "$__joined_bar._id"),
			bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
			bsonutil.NewDocElem("test_DOT_foo_DOT_b", "$b"),
			bsonutil.NewDocElem("_id", 0),
		)),
		),
	),
	)

	// after flipping, the from collection, foo is sharded and it should not push down.
	test("select * from foo right join bar on foo.a=bar.a and foo.f=bar.a", bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("$project", bsonutil.NewM(
				bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_b", "$b"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_c", "$c"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_e", "$d.e"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_g", "$g"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_f", "$d.f"),
				bsonutil.NewDocElem("test_DOT_foo_DOT__id", "$_id"),
			)),
		),
	), bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
			bsonutil.NewDocElem("test_DOT_bar_DOT_b", "$b"),
			bsonutil.NewDocElem("test_DOT_bar_DOT__id", "$_id"),
			bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$a"),
		)),
		),
	),
	)

	// should flip after not being able to be pushed down the first time due to foo being
	// sharded and then push down.
	test("select * from bar inner join foo on bar.a=foo.a and bar.a=foo.f",
		bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$match", bsonutil.NewM(bsonutil.NewDocElem("a", bsonutil.NewM(bsonutil.NewDocElem("$ne", nil)))))),
			bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewM(
				bsonutil.NewDocElem("from", "bar"),
				bsonutil.NewDocElem("localField", "a"),
				bsonutil.NewDocElem("foreignField", "a"),
				bsonutil.NewDocElem("as", "__joined_bar"),
			))),
			bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewM(
				bsonutil.NewDocElem("path", "$__joined_bar"),
				bsonutil.NewDocElem("preserveNullAndEmptyArrays", false),
			))),
			bsonutil.NewD(bsonutil.NewDocElem("$addFields", bsonutil.NewM(
				bsonutil.NewDocElem("__predicate", bsonutil.NewD(
					bsonutil.NewDocElem("$let", bsonutil.NewD(
						bsonutil.NewDocElem("vars", bsonutil.NewM(
							bsonutil.NewDocElem("predicate", bsonutil.NewM(
								bsonutil.NewDocElem("$let", bsonutil.NewM(
									bsonutil.NewDocElem("vars", bsonutil.NewM(
										bsonutil.NewDocElem("right", "$d.f"),
										bsonutil.NewDocElem("left", "$__joined_bar.a"),
									)),
									bsonutil.NewDocElem("in", bsonutil.NewM(
										bsonutil.NewDocElem("$cond", bsonutil.NewArray(
											bsonutil.NewM(
												bsonutil.NewDocElem("$or", bsonutil.NewArray(
													bsonutil.NewM(
														bsonutil.NewDocElem("$eq", bsonutil.NewArray(
															bsonutil.NewM(
																bsonutil.NewDocElem("$ifNull", bsonutil.NewArray(
																	"$$left",
																	nil,
																)),
															),
															nil,
														)),
													),
													bsonutil.NewM(
														bsonutil.NewDocElem("$eq", bsonutil.NewArray(
															bsonutil.NewM(
																bsonutil.NewDocElem("$ifNull", bsonutil.NewArray(
																	"$$right",
																	nil,
																)),
															),
															nil,
														)),
													),
												)),
											),
											nil,
											bsonutil.NewM(
												bsonutil.NewDocElem("$eq", bsonutil.NewArray(
													"$$left",
													"$$right",
												)),
											),
										)),
									)),
								)),
							)),
						)),
						bsonutil.NewDocElem("in", bsonutil.NewD(
							bsonutil.NewDocElem("$cond", bsonutil.NewArray(
								bsonutil.NewD(bsonutil.NewDocElem("$or", bsonutil.NewArray(
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										false,
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										0,
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"-0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"0.0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"-0.0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										nil,
									)),
									),
								)),
								),
								false,
								true,
							)),
						)),
					)),
				)),
			)),
			),
			bsonutil.NewD(bsonutil.NewDocElem("$match", bsonutil.NewM(bsonutil.NewDocElem("__predicate", true)))),
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
				bsonutil.NewDocElem("test_DOT_bar_DOT_a", "$__joined_bar.a"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_c", "$c"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_g", "$g"),
				bsonutil.NewDocElem("test_DOT_bar_DOT_b", "$__joined_bar.b"),
				bsonutil.NewDocElem("test_DOT_bar_DOT__id", "$__joined_bar._id"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_a", "$a"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_b", "$b"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_e", "$d.e"),
				bsonutil.NewDocElem("test_DOT_foo_DOT_f", "$d.f"),
				bsonutil.NewDocElem("test_DOT_foo_DOT__id", "$_id"),
				bsonutil.NewDocElem("_id", 0),
			)),
			),
		))
}

func TestOptimizeEvaluations(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   evaluator.SQLExpr
	}

	runTests := func(tests []test) {
		schema := evaluator.MustLoadSchema(testSchema3)
		for _, tst := range tests {
			tName := fmt.Sprintf("%q should be optimized to %q", tst.sql, tst.expected)
			t.Run(tName, func(t *testing.T) {
				req := require.New(t)

				e, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, tst.sql)
				req.NoError(err)

				eCfg := createTestExecutionCfg()
				oCfg := createOptimizerCfg(collation.Default, eCfg)
				result, err := evaluator.OptimizeEvaluations(oCfg, e)
				req.NoError(err)

				expectedVal, ok := tst.result.(evaluator.SQLValue)
				if ok && expectedVal.IsNull() {
					actualVal, ok := result.(evaluator.SQLValue)
					req.True(ok)
					req.True(actualVal.IsNull())
				} else {
					req.Zero(convey.ShouldResemble(result, tst.result))
				}
			})
		}
	}

	tests := []test{
		{"3 / '3'", "1", evaluator.NewSQLFloat(valKind, 1)},
		{"3 * '3'", "9", evaluator.NewSQLInt64(valKind, 9)},
		{"3 + '3'", "6", evaluator.NewSQLInt64(valKind, 6)},
		{"3 - '3'", "0", evaluator.NewSQLInt64(valKind, 0)},
		{"3 div '3'", "1", evaluator.NewSQLInt64(valKind, 1)},
		{"3 = '3'", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 <= '3'", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 >= '3'", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 < '3'", "false", evaluator.NewSQLBool(valKind, false)},
		{"3 > '3'", "false", evaluator.NewSQLBool(valKind, false)},
		{"3 <=> '3'", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 = a", "a = 3", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLInt64(valKind, 3),
		)},
		{"3 < a", "a > 3", evaluator.NewSQLGreaterThanExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLInt64(valKind, 3),
		)},
		{"3 <= a", "a >= 3", evaluator.NewSQLGreaterThanOrEqualExpr(evaluator.NewSQLColumnExpr(
			1, "test", "bar", "a", evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLInt64(valKind, 3))},
		{"3 > a", "a < 3", evaluator.NewSQLLessThanExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLInt64(valKind, 3),
		)},
		{"3 >= a", "a <= 3", evaluator.NewSQLLessThanOrEqualExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLInt64(valKind, 3),
		)},
		{"3 <> a", "a <> 3", evaluator.NewSQLNotEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLInt64(valKind, 3),
		)},
		{"3 + 3 = 6", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 <=> 3", "true", evaluator.NewSQLBool(valKind, true)},
		{"NULL <=> 3", "false", evaluator.NewSQLBool(valKind, false)},
		{"3 <=> NULL", "false", evaluator.NewSQLBool(valKind, false)},
		{"NULL <=> NULL", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 / (3 - 2) = a", "a = 3", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLFloat(valKind, 3),
		)},
		{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "a = 3", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt), evaluator.NewSQLInt64(valKind, 3))},
		{"3 / (3 - 2) = a AND 4 - 2 = b", "a = 3 AND b = 2",
			evaluator.NewSQLAndExpr(
				evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
					evaluator.EvalInt64, schema.MongoInt), evaluator.NewSQLFloat(valKind, 3)),
				evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "b",
					evaluator.EvalInt64, schema.MongoInt), evaluator.NewSQLInt64(valKind, 2)))},
		{"3 + 3 = 6 OR a = 3", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 + 3 = 5 OR a = 3", "a = 3", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLInt64(valKind, 3),
		)},
		{"0 OR NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"1 OR NULL", "true", evaluator.NewSQLBool(valKind, true)},
		{"NULL OR NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"0 AND 6+1 = 6", "false", evaluator.NewSQLBool(valKind, false)},
		{"3 + 3 = 5 AND a = 3", "false", evaluator.NewSQLBool(valKind, false)},
		{"0 AND NULL", "false", evaluator.NewSQLBool(valKind, false)},
		{"1 AND NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"1 AND 6+0 = 6", "true", evaluator.NewSQLBool(valKind, true)},
		{"3 + 3 = 6 AND a = 3", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(
			1, "test", "bar", "a", evaluator.EvalInt64,
			schema.MongoInt), evaluator.NewSQLInt64(valKind, 3))},
		{"(3 + 3 = 5) XOR a = 3", "a = 3", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt), evaluator.NewSQLInt64(valKind, 3))},
		{"(3 + 3 = 6) XOR a = 3", "a <> 3", evaluator.NewSQLNotExpr(evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt), evaluator.NewSQLInt64(valKind, 3)))},
		{"(13 + 9 > 6) XOR (a = 4)", "a <> 4", evaluator.NewSQLNotExpr(
			evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt), evaluator.NewSQLInt64(valKind, 4)))},
		{"(8 / 5 = 9) XOR (a = 5)", "a = 5", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a", evaluator.EvalInt64,
				schema.MongoInt), evaluator.NewSQLInt64(valKind, 5))},
		{"false XOR 23", "true", evaluator.NewSQLBool(valKind, true)},
		{"true XOR 23", "false", evaluator.NewSQLBool(valKind, false)},
		{"a = 23 XOR true", "a <> 23", evaluator.NewSQLNotExpr(evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a", evaluator.EvalInt64,
				schema.MongoInt), evaluator.NewSQLInt64(valKind, 23)))},
		{"!3", "0", evaluator.NewSQLBool(valKind, false)},
		{"!NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"a = ~1", "a = 18446744073709551614", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLUint64(valKind, uint64(18446744073709551614)))},
		{"a = ~2398238912332232323", "a = 16048505161377319292", evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				evaluator.EvalInt64, schema.MongoInt),
			evaluator.NewSQLUint64(valKind, uint64(16048505161377319292)))},
		{"DAYNAME('2016-1-1')", "Friday", evaluator.NewSQLVarchar(valKind, "Friday")},
		{"(8-7)", "1", evaluator.NewSQLInt64(valKind, 1)},
		{"a LIKE NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"4 LIKE NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"a = NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"a > NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"a >= NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"a < NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"a <= NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"a != NULL", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"(1, 3) > (3, 4)", "SQLFalse", evaluator.NewSQLBool(valKind, false)},
		{"(4, 3) > (3, 4)", "SQLTrue", evaluator.NewSQLBool(valKind, true)},
		{"(4, 31) > (4, 4)", "SQLTrue", evaluator.NewSQLBool(valKind, true)},

		{"abs(NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"abs(-10)", "10", evaluator.NewSQLFloat(valKind, 10)},
		{"ascii(NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"ascii('a')", "97", evaluator.NewSQLInt64(valKind, 97)},
		{"char_length(NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"character_length(NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"concat(NULL, a)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"concat(a, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"concat('go', 'lang')", "golang", evaluator.NewSQLVarchar(valKind, "golang")},
		{"concat_ws(NULL, a)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"convert(NULL, SIGNED)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"elt(NULL, 'a', 'b')", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"elt(4, 'a', 'b')", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"exp(NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"exp(2)", "7.38905609893065", evaluator.NewSQLFloat(valKind, 7.38905609893065)},
		{"greatest(a, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"greatest(2, 3)", "3", evaluator.NewSQLInt64(valKind, 3)},
		{"ifnull(NULL, 10)", "10", evaluator.NewSQLInt64(valKind, 10)},
		{"ifnull(10, 1)", "10", evaluator.NewSQLInt64(valKind, 10)},
		{"interval(NULL, a)", "-1", evaluator.NewSQLInt64(valKind, -1)},
		{"interval(0, 1)", "0", evaluator.NewSQLInt64(valKind, 0)},
		{"interval(1, 2, 3, 4)", "1", evaluator.NewSQLInt64(valKind, 0)},
		{"interval(1, 1, 2, 3)", "1", evaluator.NewSQLInt64(valKind, 1)},
		{"interval(-1, NULL, NULL, -0.5, 3, 4)", "1", evaluator.NewSQLInt64(valKind, 2)},
		{"interval(-3.4, -4, -3.6, -3.4, -3, 1, 2)", "3", evaluator.NewSQLInt64(valKind, 3)},
		{"interval(8, -4, 0, 7, 8)", "4", evaluator.NewSQLInt64(valKind, 4)},
		{"interval(8, -3, 1, 7, 7)", "1", evaluator.NewSQLInt64(valKind, 4)},
		{"interval(7.7, -3, 1, 7, 7)", "1", evaluator.NewSQLInt64(valKind, 4)},
		{"least(a, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"least(2, 3)", "2", evaluator.NewSQLInt64(valKind, 2)},
		{"locate('bar', 'foobar', NULL)", "0", evaluator.NewSQLInt64(valKind, 0)},
		{"locate('bar', 'foobar')", "4", evaluator.NewSQLInt64(valKind, 4)},
		{"makedate(2000, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"makedate(NULL, 10)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"mid('foobar', NULL, 2)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"mod(10, 2)", "0", evaluator.NewSQLFloat(valKind, 0)},
		{"mod(NULL, 2)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"mod(10, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"nullif(1, 1)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"nullif(1, null)", "1", evaluator.NewSQLInt64(valKind, 1)},
		{"pow(a, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"pow(NULL, a)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"pow(2,2)", "4", evaluator.NewSQLFloat(valKind, 4)},
		{"round(NULL, 2)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"round(2, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"round(2, 2)", "2", evaluator.NewSQLFloat(valKind, 2)},
		{"repeat('a', NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"repeat(NULL, 3)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"substring(NULL, 2)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"substring(NULL, 2, 3)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"substring('foobar', NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"substring('foobar', NULL, 2)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"substring('foobar', 2, NULL)", "null", evaluator.NewSQLNullUntyped(valKind)},
		{"substring('foobar', 2, 3)", "oob", evaluator.NewSQLVarchar(valKind, "oob")},
		{"substring_index(NULL, 'o', 0)", "", evaluator.NewSQLNullUntyped(valKind)},
		{"substring_index('foobar', 'o', 0)", "", evaluator.NewSQLVarchar(valKind, "")},
	}

	runTests(tests)
}

func TestOptimizeEvaluationFailures(t *testing.T) {

	type test struct {
		sql string
		err error
	}

	runTests := func(tests []test) {
		schema := evaluator.MustLoadSchema(testSchema3)
		for _, tst := range tests {
			tName := fmt.Sprintf("%q should fail with error %q", tst.sql, tst.err)
			t.Run(tName, func(t *testing.T) {
				req := require.New(t)

				e, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, tst.sql)
				req.NoError(err)

				eCfg := createTestExecutionCfg()
				oCfg := createOptimizerCfg(collation.Default, eCfg)
				_, err = evaluator.OptimizeEvaluations(oCfg, e)
				req.Zero(convey.ShouldResemble(err, tst.err))
			})
		}
	}

	tests := []test{
		{"pow(-2,2.2)", mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange, "DOUBLE",
			"pow(-2,2.2)")},
		{"pow(0,-2.2)", mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange, "DOUBLE",
			"pow(0,-2.2)")},
		{"pow(0,-5)", mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange, "DOUBLE",
			"pow(0,-5)")},
	}

	runTests(tests)
}
