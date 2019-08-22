package parser

import (
	"regexp"
	"testing"

	"github.com/10gen/sqlproxy/internal/versionutil"
	"github.com/stretchr/testify/require"
)

const versionCode = "50712"

func TestRewriteAndFormat(t *testing.T) {
	// All of these tests call .Copy() on their inputs to give coverage for our
	// Copy() implementations.
	// These tests also test the output of Format.
	t.Run("constant scalar functions", testRewriteAndFormatConstantScalarFunctions)
	t.Run("distinct", testRewriteAndFormatDistinct)
	t.Run("command", testRewriteAndFormatCommand)
	t.Run("ignored", testRewriteAndFormatIgnored)
	t.Run("namer", testNamer)
	t.Run("query", testRewriteAndFormatQuery)
}

func testRewriteAndFormatConstantScalarFunctions(t *testing.T) {
	tcases := []struct {
		desc       string
		query      string
		expected   string
		checkRegex bool
	}{
		{
			desc:       "rewrite pi()",
			query:      "select pi() from foo",
			expected:   "select 3.141592653589793E0 from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite connection_id()",
			query:      "select connection_id() from foo",
			expected:   "select cast(42, unsigned) from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite database()",
			query:      "select database() from foo",
			expected:   "select 'test_db_name' from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite schema()",
			query:      "select schema() from foo",
			expected:   "select 'test_db_name' from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite version()",
			query:      "select version() from foo",
			expected:   "select 'test_version' from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite user()",
			query:      "select user() from foo",
			expected:   "select 'test_user@test_remoteHost' from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite current_user()",
			query:      "select current_user() from foo",
			expected:   "select 'test_user@test_remoteHost' from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite session_user()",
			query:      "select session_user() from foo",
			expected:   "select 'test_user@test_remoteHost' from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite system_user()",
			query:      "select system_user() from foo",
			expected:   "select 'test_user@test_remoteHost' from foo",
			checkRegex: false,
		},
		{
			desc:       "rewrite curtime()",
			query:      "select curtime() from foo",
			expected:   `select datetime \'\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.\d\d\d\d\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite current_time()",
			query:      "select current_time() from foo",
			expected:   `select datetime \'\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.\d\d\d\d\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite utc_time()",
			query:      "select utc_time() from foo",
			expected:   `select datetime \'\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.\d\d\d\d\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite current_timestamp()",
			query:      "select current_timestamp() from foo",
			expected:   `select datetime \'\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.\d\d\d\d\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite now()",
			query:      "select now() from foo",
			expected:   `select datetime \'\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.\d\d\d\d\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite curdate()",
			query:      "select curdate() from foo",
			expected:   `select date \'\d\d\d\d-\d\d-\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite current_date()",
			query:      "select current_date() from foo",
			expected:   `select date \'\d\d\d\d-\d\d-\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite utc_timestamp()",
			query:      "select utc_timestamp() from foo",
			expected:   `select datetime \'\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.\d\d\d\d\d\d\' from foo`,
			checkRegex: true,
		},
		{
			desc:       "rewrite utc_date()",
			query:      "select utc_date() from foo",
			expected:   `select date \'\d\d\d\d-\d\d-\d\d\' from foo`,
			checkRegex: true,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			req := require.New(t)

			tree, err := Parse(tcase.query)
			req.NoError(err)

			newTree, err := RewriteConstantScalarFunctions(tree.Copy().(Statement), 42, "test_db_name", "test_version", "test_remoteHost", "test_user")
			req.NoError(err)
			buf := NewTrackedBuffer(nil)
			newTree.Format(buf)
			newTreeStr := buf.String()

			if tcase.checkRegex {
				req.Regexp(regexp.MustCompile(tcase.expected), newTreeStr)
			} else {
				req.Equal(tcase.expected, newTreeStr)
			}

		})
	}
}

func testRewriteAndFormatDistinct(t *testing.T) {
	tcases := []struct {
		desc     string
		query    string
		expected string
	}{
		{
			desc:     "nothing distinct to rewrite",
			query:    "select * from foo",
			expected: "select * from foo",
		},
		{
			desc:     "too many distincts to rewrite",
			query:    "select sum(distinct a), sum(distinct b) from foo",
			expected: "select sum(distinct a), sum(distinct b) from foo",
		},
		{
			desc:  "show that derived tables can be rewritten",
			query: "select sum(distinct x) from (select sum(distinct y) as x from foo)",
			expected: "select sum($___mongosqld_as_1) as sum(distinct x) from " +
				"(select x as $___mongosqld_as_1 from (select sum(x) as x from " +
				"(select y as x from foo group by 1) as $___mongosqld_query_0) group by 1)" +
				" as $___mongosqld_query_2",
		},
		{
			desc:     "show that we do not rewrite when subqueries are in select exprs",
			query:    "select (select sum(distinct y) from DUAL) as y from foo",
			expected: "select (select sum(distinct y) from DUAL) as y from foo",
		},
		{
			desc:     "show that we do not rewrite when subqueries are in where clauses",
			query:    "select x from foo where (select sum(distinct y) from DUAL)",
			expected: "select x from foo where (select sum(distinct y) from DUAL)",
		},
		{
			desc:     "having disables rewrite",
			query:    "select x from foo having sum(distinct y) = 3",
			expected: "select x from foo having sum(distinct y) = 3",
		},
		{
			desc:     "having still disables rewrite",
			query:    "select sum(distinct x) from foo having sum(distinct y) = 3",
			expected: "select sum(distinct x) from foo having sum(distinct y) = 3",
		},
		{
			desc:  "simple distinct aggregation function query",
			query: "select sum(distinct x) from foo",
			expected: "select sum($___mongosqld_as_0) as sum(distinct x) from " +
				"(select x as $___mongosqld_as_0 from foo group by 1) as $___mongosqld_query_1",
		},
		{
			desc:  "simple distinct aggregation function query with where",
			query: "select sum(distinct x) from foo where x = 1",
			expected: "select sum($___mongosqld_as_0) as sum(distinct x) from " +
				"(select x as $___mongosqld_as_0 from foo where x = 1 group by 1) as" +
				" $___mongosqld_query_1",
		},
		{
			desc:  "simple distinct aggregation function query with where, order, and limit",
			query: "select sum(distinct x) from foo where x = 1 order by 1 limit 10",
			expected: "select sum($___mongosqld_as_0) as sum(distinct x) from " +
				"(select x as $___mongosqld_as_0 from foo where x = 1 group by 1 " +
				"order by 1 asc limit 10) as $___mongosqld_query_1",
		},
		{
			desc:  "simple distinct aggregation function query with alias",
			query: "select sum(distinct x) as bar from foo as foo2",
			expected: "select sum(bar) as bar from (select x as bar from foo as foo2 group by 1)" +
				" as $___mongosqld_query_0",
		},
		{
			desc:     "non distinct agg func disables rewrite",
			query:    "select sum(distinct x), count(y) from foo",
			expected: "select sum(distinct x), count(y) from foo",
		},
		{
			desc: "distinct union query",
			query: "select a, sum(distinct x) from foo group by a " +
				"union all select NULL, count(distinct a+b) from bar",
			expected: "select $___mongosqld_as_0 as a, sum($___mongosqld_as_1) as sum(distinct x)" +
				" from (select a as $___mongosqld_as_0, x as $___mongosqld_as_1" +
				" from foo group by a, 2)" +
				" as $___mongosqld_query_2 group by a union all " +
				"select $___mongosqld_as_3 as null, count($___mongosqld_as_4) as " +
				"count(distinct a+b) from (select null as $___mongosqld_as_3, a+b " +
				"as $___mongosqld_as_4 from bar group by 2) as $___mongosqld_query_5",
		},
		{
			desc: "nested subqueries distinct union",
			query: "select sum(distinct x) from (select sum(distinct y) as x from foo) " +
				"union all select x, sum(distinct y) from bar group by x",
			expected: "select sum($___mongosqld_as_1) as sum(distinct x) from " +
				"(select x as $___mongosqld_as_1 from (select sum(x) as x from " +
				"(select y as x from foo group by 1) as $___mongosqld_query_0) group by 1)" +
				" as $___mongosqld_query_2 union all " +
				"select $___mongosqld_as_3 as x, sum($___mongosqld_as_4) as sum(distinct y) from" +
				" (select x as $___mongosqld_as_3, y as $___mongosqld_as_4 from" +
				" bar group by x, 2)" +
				" as $___mongosqld_query_5 group by x",
		},
		{
			desc: "nested subqueries distinct union both sides",
			query: "select sum(distinct x) from (select sum(distinct y) as x from foo) " +
				"union all select x, sum(distinct y) from bar group by x",
			expected: "select sum($___mongosqld_as_1) as sum(distinct x) from " +
				"(select x as $___mongosqld_as_1 from (select sum(x) as x from " +
				"(select y as x from foo group by 1) as $___mongosqld_query_0) group by 1)" +
				" as $___mongosqld_query_2 union all select $___mongosqld_as_3 as x, " +
				"sum($___mongosqld_as_4) as sum(distinct y) from " +
				"(select x as $___mongosqld_as_3, y as $___mongosqld_as_4" +
				" from bar group by x, 2) as $___mongosqld_query_5 group by x",
		},
		{
			desc: "distinct join query",
			query: "select sum(distinct l.x + r.x) from foo l " +
				"inner join bar r on l._id = r._id",
			expected: "select sum($___mongosqld_as_0) as sum(distinct l.x+r.x) from " +
				"(select l.x+r.x as $___mongosqld_as_0 from foo as l " +
				"join bar as r on l._id = r._id group by 1) as $___mongosqld_query_1",
		},
		{
			desc: "distinct rewrite above join with expression in agg function",
			query: "select sum(distinct a + b) from (select a.a as a, b.b as b" +
				" from groupby a inner join groupby b) g",
			expected: "select sum($___mongosqld_as_0) as sum(distinct a+b) from " +
				"(select a+b as $___mongosqld_as_0 from (select a.a as a, b.b as b from groupby" +
				" as a join groupby as b) as g group by 1) as $___mongosqld_query_1",
		},
		{
			desc:  "duplicate distinct_rewrite_join from integration tests",
			query: "select sum(distinct a.a+b.b) from groupby a inner join groupby b",
			expected: "select sum($___mongosqld_as_0) as sum(distinct a.a+b.b) from " +
				"(select a.a+b.b as $___mongosqld_as_0 from groupby as a " +
				"join groupby as b group by 1) as $___mongosqld_query_1",
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			req := require.New(t)

			tree, err := Parse(tcase.query)
			req.NoError(err)

			newTree := RewriteDistinct(tree)
			buf := NewTrackedBuffer(nil)
			newTree.Format(buf)
			newTreeStr := buf.String()
			req.Equal(tcase.expected, newTreeStr)
		})
	}
}

func testRewriteAndFormatCommand(t *testing.T) {
	tcases := []struct {
		desc     string
		command  string
		expected string
	}{
		{
			desc:     "rewrite bool types",
			command:  "create table foo(a bool, b bit, c boolean)",
			expected: "create table foo(a boolean, b boolean, c boolean)",
		},
		{
			desc:     "rewrite date types",
			command:  "create table foo(a datetime, b timestamp)",
			expected: "create table foo(a timestamp, b timestamp)",
		},
		{
			desc:     "rewrite int types",
			command:  "create table foo(a tinyint, b smallint, c int, d integer, e bigint, f tinyint(1))",
			expected: "create table foo(a int, b int, c int, d int, e int, f int(1))",
		},
		{
			desc:     "rewrite char types",
			command:  "create table foo(a tinytext, b text, c mediumtext, d longtext, e char, f varchar)",
			expected: "create table foo(a varchar, b varchar, c varchar, d varchar, e varchar, f varchar)",
		},
		{
			desc:     "rewrite float types",
			command:  "create table foo(a double, b float)",
			expected: "create table foo(a float, b float)",
		},
		{
			desc:     "rewrite decimal types",
			command:  "create table foo(a decimal)",
			expected: "create table foo(a decimal)",
		},
		{
			desc: "test full create table formatting",
			command: `create table if not exists foo(
			                           a int not null unique,
			                           b int null unique comment 'hello',
									   fulltext index idx(a, b),
									   c int comment 'world',
									   unique index(c)) comment = 'tbl!'`,
			expected: "create table if not exists foo(a int not null unique, " +
				"b int unique comment 'hello', fulltext index `idx`(a asc, b asc), " +
				"c int comment 'world', unique index(c asc)) comment = 'tbl!'",
		},
		{
			desc:     "test create database formatting",
			command:  "create database if not exists foo",
			expected: "create database if not exists `foo`",
		},
		{
			desc:     "test drop database formatting",
			command:  "drop database if exists foo",
			expected: "drop database if exists `foo`",
		},
		{
			desc:     "test drop table formatting",
			command:  "drop table if exists foo",
			expected: "drop table if exists foo",
		},
		{
			desc:     "insert missing column list",
			command:  "insert into foo value(1)",
			expected: "insert into foo values(1)",
		},
		{
			desc:     "insert empty column list",
			command:  "insert into foo() value(1)",
			expected: "insert into foo values(1)",
		},
		{
			desc:     "insert non-empty column list",
			command:  "insert into foo(x) value(1)",
			expected: "insert into foo(x) values(1)",
		},
		{
			desc:     "insert missing column list without into",
			command:  "insert foo value(1)",
			expected: "insert into foo values(1)",
		},
		{
			desc:     "insert empty column list without into",
			command:  "insert foo() value(1)",
			expected: "insert into foo values(1)",
		},
		{
			desc:     "insert non-empty column list without into",
			command:  "insert foo(x) value(1)",
			expected: "insert into foo(x) values(1)",
		},
		{
			desc: "insert many column row missing column list",
			command: `insert into foo value(1,2,3,4,'hello', true, false, NULL, date '2012-01-01',
			               time '01:03:03', timestamp '2012-01-01T01:03:03')`,
			expected: `insert into foo values(1, 2, 3, 4, 'hello', ` +
				`true, false, null, date '2012-01-01', ` +
				`time '01:03:03', timestamp '2012-01-01T01:03:03')`,
		},
		{
			desc: "insert many column row empty column list",
			command: `insert into foo() value(1,2,3,4,'hello', true, false, NULL, date '2012-01-01',
			               time '01:03:03', timestamp '2012-01-01T01:03:03')`,
			expected: `insert into foo values(1, 2, 3, 4, 'hello', ` +
				`true, false, null, date '2012-01-01', ` +
				`time '01:03:03', timestamp '2012-01-01T01:03:03')`,
		},
		{
			desc: "insert multi column row values with column specifier",
			command: `insert into foo(x,y,z) value(1,2,3,4,
			               'hello', true, false, NULL, date '2012-01-01',
			               time '01:03:03', timestamp '2012-01-01T01:03:03')`,
			expected: `insert into foo(x, y, z) values(1, 2, 3, 4, ` +
				`'hello', true, false, null, date '2012-01-01', ` +
				`time '01:03:03', timestamp '2012-01-01T01:03:03')`,
		},
		{
			desc: "insert many column row with multiple rows with column specifier",
			command: `insert into foo(x,y,z) values
							(1,2,3,4,
			               'hello', true, false, NULL, date '2012-01-01',
			               time '01:03:03', timestamp '2012-01-01T01:03:03'),
							(41,42,43,44,
			               'world', false, true, NULL, date '2013-01-01',
			               time '01:03:03', timestamp '2013-01-01T01:03:03')`,
			expected: `insert into foo(x, y, z) values` +
				`(1, 2, 3, 4, ` +
				`'hello', true, false, null, date '2012-01-01', ` +
				`time '01:03:03', timestamp '2012-01-01T01:03:03'), ` +
				`(41, 42, 43, 44, ` +
				`'world', false, true, null, date '2013-01-01', ` +
				`time '01:03:03', timestamp '2013-01-01T01:03:03')`,
		},
		{
			desc:     "insert with default value",
			command:  `insert into foo() value(1,2,DEFAULT)`,
			expected: `insert into foo values(1, 2, default)`,
		},
		{
			desc:     "insert with several default values",
			command:  `insert into foo() value(DeFault,2,DEFAULT, deFAULT)`,
			expected: `insert into foo values(default, 2, default, default)`,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			req := require.New(t)

			tree, err := Parse(tcase.command)
			req.NoError(err)

			newTree, err := DesugarStatement(tree.Copy().(Statement), versionCode)
			req.Nil(err)
			buf := NewTrackedBuffer(nil)
			newTree.Format(buf)
			newTreeStr := buf.String()
			req.Equal(tcase.expected, newTreeStr)
		})
	}

	t.Run("test that bit(n) for n > 1 fails", func(t *testing.T) {
		req := require.New(t)
		tree, err := Parse("create table foo(x bit(12))")
		req.NoError(err)

		_, err = DesugarStatement(tree.Copy().(Statement), versionCode)
		req.NotNil(err)
		req.Equal("bit(n) for n > 1 is not allowed at this time, found n = 12", err.Error())
	})
}

func testNamer(t *testing.T) {
	tcases := []struct {
		desc     string
		query    string
		expected string
	}{
		{
			desc:     "no non-star exprs to name",
			query:    "select * from foo",
			expected: "select * from foo",
		},
		{
			desc:     "rename simple column ref",
			query:    "select a from foo",
			expected: "select a as a from foo",
		},
		{
			desc:     "rename literal column",
			query:    "select 2 from foo",
			expected: "select 2 as 2 from foo",
		},
		{
			desc:     "rename expr column",
			query:    "select 2+2 from foo",
			expected: "select 2+2 as 2+2 from foo",
		},
		{
			desc:     "expr column string not perfectly preserved",
			query:    "select 2 + 2 from foo",
			expected: "select 2+2 as 2+2 from foo",
		},
		{
			desc:     "value should not be backticked",
			query:    "select sum(value) from foo",
			expected: "select sum(value) as sum(value) from foo",
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			req := require.New(t)

			tree, err := Parse(tcase.query)
			req.NoError(err)

			newTree := NameColumns(tree)
			buf := NewTrackedBuffer(nil)
			newTree.Format(buf)
			newTreeStr := buf.String()
			req.Equal(tcase.expected, newTreeStr)
		})
	}
}

func testRewriteAndFormatQuery(t *testing.T) {
	tcases := []struct {
		desc     string
		query    string
		expected string
	}{
		{
			desc:     "nothing to desugar",
			query:    "select * from foo",
			expected: "select * from foo",
		},
		{
			desc:     "unwrap single tuples",
			query:    "select (2+2) from foo",
			expected: "select 2+2 from foo",
		},
		{
			desc:     "rewrite = subquery comparisons to non-subqueries",
			query:    "select (select * from foo) = 1 from DUAL",
			expected: "select (select * from foo) = (select 1 from DUAL) from DUAL",
		},
		{
			desc:     "rewrite <=> subquery comparisons to non-subqueries",
			query:    "select (select * from foo) <=> 1 from DUAL",
			expected: "select (select * from foo) <=> (select 1 from DUAL) from DUAL",
		},
		{
			desc:     "rewrite < subquery comparisons to non-subqueries",
			query:    "select (select * from foo) < 1 from DUAL",
			expected: "select (select * from foo) < (select 1 from DUAL) from DUAL",
		},
		{
			desc:     "rewrite > subquery comparisons to non-subqueries",
			query:    "select (select * from foo) > 1 from DUAL",
			expected: "select (select * from foo) > (select 1 from DUAL) from DUAL",
		},
		{
			desc:     "rewrite <= subquery comparisons to non-subqueries",
			query:    "select (select * from foo) <= 1 from DUAL",
			expected: "select (select * from foo) <= (select 1 from DUAL) from DUAL",
		},
		{
			desc:     "rewrite >= subquery comparisons to non-subqueries",
			query:    "select (select * from foo) >= 1 from DUAL",
			expected: "select (select * from foo) >= (select 1 from DUAL) from DUAL",
		},
		{
			desc:     "rewrite = nested subquery",
			query:    "select * from (select (select * from foo) = 1 from DUAL)",
			expected: "select * from (select (select * from foo) = (select 1 from DUAL) from DUAL)",
		},
		{
			desc:     "replace between with conjunction",
			query:    "select x between 1 and 20 from DUAL",
			expected: "select x >= 1 and x <= 20 from DUAL",
		},
		{
			desc:     "replace not between with disjunction",
			query:    "select x not between 1 and 20 from DUAL",
			expected: "select x < 1 or x > 20 from DUAL",
		},
		{
			desc:     "replace field with case",
			query:    "select field(x, 12, 22, 23, 24, 25) from foo",
			expected: "select case when x = 12 then 1 when x = 22 then 2 when x = 23 then 3 when x = 24 then 4 when x = 25 then 5  else 0 end from foo",
		},
		{
			desc:     "replace is not with not is",
			query:    "select a is not true from foo",
			expected: "select not a is true from foo",
		},
		{
			desc:     "replace coalesce with case",
			query:    "select coalesce(x, 12, 22, 23, 24, 25) from foo",
			expected: "select case when not x is null then x when not 12 is null then 12 when not 22 is null then 22 when not 23 is null then 23 when not 24 is null then 24 when not 25 is null then 25  end from foo",
		},
		{
			desc:     "replace elt with case",
			query:    "select elt(x, 12, 22, 23, 24, 25) from foo",
			expected: "select case when x = 1 then 12 when x = 2 then 22 when x = 3 then 23 when x = 4 then 24 when x = 5 then 25  end from foo",
		},
		{
			desc:     "replace nested field/elt with case",
			query:    "select field(x, 12, elt(y, 22, 23, 24), 23, 24, 25) from foo",
			expected: "select case when x = 12 then 1 when x = case when y = 1 then 22 when y = 2 then 23 when y = 3 then 24  end then 2 when x = 23 then 3 when x = 24 then 4 when x = 25 then 5  else 0 end from foo",
		},
		{
			desc:     "replace if with case",
			query:    "select if(x>10,2,3) from DUAL",
			expected: "select case when x > 10 then 2  else 3 end from DUAL",
		},
		{
			desc:     "replace nested if with case",
			query:    "select if(x>10,if(x<2,2,3),3) from DUAL",
			expected: "select case when x > 10 then case when x < 2 then 2  else 3 end  else 3 end from DUAL",
		},
		{
			desc:     "replace ifnull with case",
			query:    "select ifnull(x,'hello') from DUAL",
			expected: "select case when x is null then 'hello'  else x end from DUAL",
		},
		{
			desc:     "replace interval with case",
			query:    "select interval(x, 1, 75, 17, 30, 56, 175) from foo",
			expected: "select case when x is null then -1 when x < 1 then 0 when x < 75 then 1 when x < 17 then 2 when x < 30 then 3 when x < 56 then 4 when x < 175 then 5  else 6 end from foo",
		},
		{
			desc:     "replace nullif with case",
			query:    "select nullif('hello','hello') from DUAL",
			expected: "select case when 'hello' = 'hello' then null  else 'hello' end from DUAL",
		},
		{
			desc:     "in list",
			query:    "select a in (x, y, z) from DUAL",
			expected: "select a = x or a = y or a = z from DUAL",
		},
		{
			desc:     "not in list",
			query:    "select a not in (x, y, z) from DUAL",
			expected: "select not a = x or a = y or a = z from DUAL",
		},
		{
			desc:     "left tuple to subquery",
			query:    "select (a, b, c) = (select d, e, f from bar) from foo",
			expected: "select (select a, b, c from DUAL) = (select d, e, f from bar) from foo",
		},
		{
			desc:     "right tuple gt to subquery",
			query:    "select (select a, b, c from bar) > (d, e, f) from foo",
			expected: "select (select a, b, c from bar) > (select d, e, f from DUAL) from foo",
		},
		{
			desc:     "right tuple gte to subquery",
			query:    "select (select a, b, c from bar) >= (d, e, f) from foo",
			expected: "select (select a, b, c from bar) >= (select d, e, f from DUAL) from foo",
		},
		{
			desc:     "right tuple lt to subquery",
			query:    "select (select a, b, c from bar) < (d, e, f) from foo",
			expected: "select (select a, b, c from bar) < (select d, e, f from DUAL) from foo",
		},
		{
			desc:     "right tuple lte to subquery",
			query:    "select (select a, b, c from bar) <= (d, e, f) from foo",
			expected: "select (select a, b, c from bar) <= (select d, e, f from DUAL) from foo",
		},
		{
			desc:     "tuple comparison",
			query:    "select (a, b) < (c, d) from foo",
			expected: "select a < c or a = c and b < d from foo",
		},
		{
			desc:     "nested tuple comparison",
			query:    "select ((a), (b)) < ((c), (d)) from foo",
			expected: "select a < c or a = c and b < d from foo",
		},
		{
			desc:     "non-uniform depth destructuring comparison",
			query:    "select ((((a))), (b)) > ((((c))), (d)) from foo",
			expected: "select a > c or a = c and b > d from foo",
		},
		{
			desc:     "subquery operator some",
			query:    "select a < some (select b, c from bar) from foo",
			expected: "select (select a from DUAL) < any (select b, c from bar) from foo",
		},
		{
			desc:     "subquery operator any",
			query:    "select a < any (select b, c from bar) from foo",
			expected: "select (select a from DUAL) < any (select b, c from bar) from foo",
		},
		{
			desc:     "subquery operator all",
			query:    "select a < all (select b, c from bar) from foo",
			expected: "select (select a from DUAL) < all (select b, c from bar) from foo",
		},
		{
			desc:     "subquery operator none",
			query:    "select a < (select b, c from bar) from foo",
			expected: "select (select a from DUAL) < (select b, c from bar) from foo",
		},
		{
			desc:     "make implicit reference to dual table explicit",
			query:    "select 2",
			expected: "select 2 from DUAL",
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			req := require.New(t)

			tree, err := Parse(tcase.query)
			req.NoError(err)

			newTree, err := desugarStatementNoNaming(tree.Copy().(Statement), versionCode)
			req.NoError(err)

			buf := NewTrackedBuffer(nil)
			newTree.Format(buf)
			newTreeStr := buf.String()
			req.Equal(tcase.expected, newTreeStr)
		})
	}
}

// The naming pass makes it harder to write test cases. This is a testing only hack function
// that skips the naming pass, but performs the other command and query desugarings. It does not
// test constant scalar rewriting or distinct group rewriting.
func desugarStatementNoNaming(statement Statement, versionCode versionutil.MySQLFixedWidthVersionCode) (Statement, error) {
	desugarers := []Walker{
		&evaluateConditionalComment{versionCode},
		&createTableTypeDesugarer{},
		&isNotDesugarer{},
		&unwrapSingleTuples{},
		&someToAnyDesugarer{},
		&betweenDesugarer{},
		&ifToCaseDesugarer{},
		&inSubqueryDesugarer{},
		&inListConverter{},
		&subqueryComparisonConverter{},
		&tupleComparisonDesugarer{},
		&makeDualExplicit{},
	}

	result := statement.(CST)
	var err error
	for _, pass := range desugarers {
		result, err = Walk(pass, result)
		if err != nil {
			return nil, err
		}
	}

	return result.(Statement), nil
}

func testRewriteAndFormatIgnored(t *testing.T) {
	tcases := []struct {
		desc     string
		command  string
		expected string
	}{
		{
			desc:     "complex table locks",
			command:  "lock tables scalar READ lOcAL, scalar2 as sc WRITE, scalar3 as sc3 READ, scalar4 as sc4 low_priority wRiTe",
			expected: "/* IGNORED */ lock tables scalar as `scalar` read, scalar2 as `sc` write, scalar3 as `sc3` read, scalar4 as `sc4` write",
		},
		{
			desc:     "unlock tables",
			command:  "UnLoCK TABLES",
			expected: "/* IGNORED */ unlock tables",
		},
		{
			desc:     "enable keys",
			command:  "EnABLE KeYS",
			expected: "/* IGNORED */ enable keys",
		},
		{
			desc:     "disable keys",
			command:  "DiSABLE KeYS",
			expected: "/* IGNORED */ disable keys",
		},
		{
			desc:     "alter table enable keys",
			command:  "Alter table foo EnABLE KeYS",
			expected: "/* IGNORED */ enable keys",
		},
		{
			desc:     "alter table disable keys",
			command:  "AltEr table bar DiSABLE KeYS",
			expected: "/* IGNORED */ disable keys",
		},
		{
			desc:     "conditional comment that is executable",
			command:  "/*!50712 set @foo ='hello'*/",
			expected: "set @foo = 'hello'",
		},
		{
			desc:     "conditional comment with a query that is executable",
			command:  "/*!50712 select * from bar*/",
			expected: "select * from bar",
		},
		{
			desc:     "conditional comment that is not executable",
			command:  "/*!99999 set @foo='hello'*/",
			expected: "/* IGNORED */ /*!99999 set @foo='hello'*/",
		},
		{
			desc:     "conditional comment with a query that is not executable",
			command:  "/*!99999 select * from bar */",
			expected: "/* IGNORED */ /*!99999 select * from bar */",
		},
		{
			desc:     "conditional comment with a query that is executable and namer",
			command:  "/*!50712 select a+b from bar*/",
			expected: "select a+b as a+b from bar",
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			req := require.New(t)

			tree, err := Parse(tcase.command)
			req.NoError(err)

			newTree, err := DesugarStatement(tree.Copy().(Statement), versionCode)
			req.NoError(err)
			buf := NewTrackedBuffer(nil)
			newTree.Format(buf)
			newTreeStr := buf.String()
			req.Equal(tcase.expected, newTreeStr)
		})
	}
}
