package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnonymize(t *testing.T) {
	tests := []struct {
		name, in, out string
	}{
		{
			"table_simple",
			"select * from foo",
			"select * from tbl_1",
		},
		{
			"table_column",
			"select a from foo",
			"select col_1 from tbl_1",
		},
		{
			"two_column",
			"select a, b from foo",
			"select col_1, col_2 from tbl_1",
		},
		{
			"two_table_two_column",
			"select b, a from foo",
			"select col_1, col_2 from tbl_1",
		},
		{
			"two_table_join",
			"select * from foo join bar",
			"select * from tbl_1 join tbl_2",
		},
		{
			"two_table_join_opposite_order",
			"select * from bar join foo",
			"select * from tbl_1 join tbl_2",
		},
		{
			"two_table_join_qualified_column_0",
			"select foo.a from foo, bar",
			"select tbl_1.col_1 from tbl_1, tbl_2",
		},
		{
			"two_table_join_qualified_column_1",
			"select bar.a from foo, bar",
			"select tbl_1.col_1 from tbl_2, tbl_1",
		},
		{
			"two_table_join_same_column_name_different_tables",
			"select foo.a, bar.a from foo, bar",
			"select tbl_1.col_1, tbl_2.col_1 from tbl_1, tbl_2",
		},
		{
			"many_tables",
			"select * from foo, bar, baz, barbaz",
			"select * from tbl_1, tbl_2, tbl_3, tbl_4",
		},
		{
			"many_tables_repeat",
			"select * from foo, bar, foo, bar",
			"select * from tbl_1, tbl_2, tbl_1, tbl_2",
		},
		{
			"literals_int",
			"select 1, 2, 3, 4096, -10",
			"select 123, 123, 123, 123, 123",
		},
		{
			"literals_float",
			"select 1.0, 2.3, 3.6, -0.4",
			"select 4.56, 4.56, 4.56, 4.56",
		},
		{
			"literals_string",
			"select '123', 'hello', '22 10th Street'",
			"select 'abc', 'abc', 'abc'",
		},
		{
			"literals_date",
			"select date '2005-08-09', date '2006-04-22'",
			"select date '2012-03-04', date '2012-03-04'",
		},
		{
			"literals_null",
			"select * from foo where a <=> null",
			"select * from tbl_1 where col_1 <=> null",
		},
		{
			"literals_unknown",
			"select * from foo where a <=> unknown",
			"select * from tbl_1 where col_1 <=> null",
		},
		{
			"literals_bool_0",
			"select * from foo where a = true and b = false",
			"select * from tbl_1 where col_1 = false and col_2 = false",
		},
		{
			"literals_bool_1",
			"select * from foo where a = false and b = true",
			"select * from tbl_1 where col_1 = false and col_2 = false",
		},
		{
			"name_collision_0",
			"select a, b from a join b",
			"select col_1, col_2 from tbl_1 join tbl_2",
		},
		{
			"name_collision_1",
			"select b.a, a.b from a join b",
			"select tbl_1.col_1, tbl_2.col_2 from tbl_2 join tbl_1",
		},
		{
			"name_collision_2",
			"select b.a, a.b from a.a join b.b",
			"select tbl_1.col_1, tbl_2.col_2 from db_1.tbl_2 join db_2.tbl_1",
		},
		{
			"name_collision_2",
			"select a.b.a, a.a.b from a.a join a.b join b.b",
			"select db_1.tbl_1.col_1, db_1.tbl_2.col_2 from db_1.tbl_2 join db_1.tbl_1 join db_2.tbl_1",
		},
		{
			"comments_0",
			"select /* SENSITIVE */ * from foo",
			"select * from tbl_1",
		},
		{
			"comments_1",
			"select /* SENSITIVE */ * from /* SENSITIVE */ foo",
			"select * from tbl_1",
		},
		{
			"comments_2",
			"select * from foo -- SENSITIVE",
			"select * from tbl_1",
		},
		{
			"comments_3",
			"select * from foo # SENSITIVE",
			"select * from tbl_1",
		},
		{
			"cte_0",
			"with cte1 as (select * from foo) select a from cte1",
			"with tbl_1 as (select * from tbl_2) select col_1 from tbl_1",
		},
		{
			"cte_1",
			"with cte1 as (with nestedcte as (select * from foo) select a from nestedcte) select a from cte1",
			"with tbl_1 as (with tbl_2 as (select * from tbl_3) select col_1 from tbl_2) select col_1 from tbl_1",
		},
		{
			"cte_2",
			"with cte1 (c, d) as (select a, b from foo) select d from cte1",
			"with tbl_1 (col_1, col_2) as (select col_3, col_4 from tbl_2) select col_2 from tbl_1",
		},
		{
			"alias_column",
			"select a as b, c as d from foo where d",
			"select col_2 as col_1, col_4 as col_3 from tbl_1 where col_3",
		},
		{
			"alias_table",
			"select f.id, b.id from foo as f join bar as b",
			"select tbl_1.col_1, tbl_2.col_1 from tbl_3 as tbl_1 join tbl_4 as tbl_2",
		},
		{
			"alias_derived_table",
			"select sub.a from (select * from foo) as sub",
			"select tbl_1.col_1 from (select * from tbl_2) as tbl_1",
		},
		{
			"qualified_star_0",
			"select foo.*, bar.a from foo join bar",
			"select tbl_1.*, tbl_2.col_1 from tbl_1 join tbl_2",
		},
		{
			"qualified_star_0",
			"select foo.*, bar.a from foo join bar",
			"select tbl_1.*, tbl_2.col_1 from tbl_1 join tbl_2",
		},
		{
			"qualified_star_1",
			"select test.foo.*, test.bar.a from foo join bar",
			"select db_1.tbl_1.*, db_1.tbl_2.col_1 from tbl_1 join tbl_2",
		},
		{
			"qualified_star_2",
			"select dbA.foo.*, dbB.bar.a from dbB.foo join dbA.bar",
			"select db_1.tbl_1.*, db_2.tbl_2.col_1 from db_2.tbl_1 join db_1.tbl_2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			stmt, err := Parse(test.in)
			req.NoError(err)
			res := String(anonymizeStatement(stmt))
			req.Equal(test.out, res)
		})
	}
}
