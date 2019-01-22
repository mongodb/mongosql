package mongotranslate

import (
	"testing"
)

const testDefaultDB = "test"

type tcase struct {
	desc            string
	query           string
	defaultDB       string
	expectedDBs     []string
	expectedTables  map[string][]string
	expectedColumns map[string]map[string][]string
}

func TestInferSchemaFromQuery(t *testing.T) {
	simpleSelects := []tcase{
		{
			desc:        "simple, fully qualified parameters with explain",
			query:       "explain select test.foo.a from test.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "simple, fully qualified parameters without explain",
			query:       "select test.foo.a from test.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "single unqualified column with unqualified table",
			query:       "select a from foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "multiple unqualified columns with single unqualified table",
			query:       "select a, b, c from foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "single unqualified column with single qualified table",
			query:       "select a from db.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "multiple unqualified columns with single qualified table",
			query:       "select a, b, c from db.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "single partly qualified column with single unqualified table",
			query:       "select foo.a from foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns with single unqualified table",
			query:       "select foo.a, foo.b, foo.c from foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "single partly qualified column with single qualified table",
			query:       "select foo.a from db.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns with single qualified table",
			query:       "select foo.a, foo.b, foo.c from db.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "single fully qualified columns with single unqualified table",
			query:       "select test.foo.a from foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "multiple fully qualified columns with single unqualified table",
			query:       "select test.foo.a, test.foo.b, test.foo.c from foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "single fully qualified columns with single qualified table",
			query:       "select db.foo.a from db.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "multiple fully qualified columns with single qualified table",
			query:       "select db.foo.a, db.foo.b, db.foo.c from db.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple unqualified columns with multiple unqualified tables",
			query:       "select a, b, c from foo, bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
					"bar": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns (all same table) with multiple unqualified tables",
			query:       "select foo.a, foo.b, foo.c from foo, bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns (different tables) with multiple unqualified tables",
			query:       "select foo.a, bar.b, bar.c from foo, bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
					"bar": {"b", "c"},
				},
			},
		},
		{
			desc:        "multiple fully qualified columns (all same table) with multiple unqualified tables",
			query:       "select test.foo.a, test.foo.b, test.foo.c from foo, bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple unqualified columns with multiple qualified tables (same database)",
			query:       "select a, b, c from db.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a", "b", "c"},
					"bar": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple unqualified columns with multiple qualified tables (different databases)",
			query:       "select a, b, c from test.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
				"db": {
					"bar": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns (all same table) with multiple qualified tables (same database)",
			query:       "select foo.a, foo.b, foo.c from db.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns (all same table) with multiple qualified tables (different databases)",
			query:       "select foo.a, foo.b, foo.c from test.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns (different tables) with multiple qualified tables (same database)",
			query:       "select foo.a, bar.b, bar.c from db.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a"},
					"bar": {"b", "c"},
				},
			},
		},
		{
			desc:        "multiple partly qualified columns (different tables) with multiple qualified tables (different databases)",
			query:       "select foo.a, bar.b, bar.c from test.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
				"db": {
					"bar": {"b", "c"},
				},
			},
		},
		{
			desc:        "multiple fully qualified columns (all same table) with multiple qualified tables (same database)",
			query:       "select db.foo.a, db.foo.b, db.foo.c from db.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple fully qualified columns (all same table and database) with multiple qualified tables (different databases)",
			query:       "select db.bar.a, db.bar.b, db.bar.c from test.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"bar": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple fully qualified columns (different tables and databases) with multiple qualified tables (different databases)",
			query:       "select test.foo.a, db.bar.b, db.bar.c from test.foo, db.bar",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
				"db": {
					"bar": {"b", "c"},
				},
			},
		},
		{
			desc:        "mixed qualifications of columns and tables",
			query:       "select a, foo.b, db.foo.c, bar.d, bar.c, db2.baz.d, x.x from db.foo, bar, db2.baz, x",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db", "db2"},
			expectedTables: map[string][]string{
				testDefaultDB: {"bar", "x"},
				"db":          {"foo"},
				"db2":         {"baz"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"bar": {"a", "d", "c"},
					"x":   {"a", "x"},
				},
				"db": {
					"foo": {"a", "b", "c"},
				},
				"db2": {
					"baz": {"a", "d"},
				},
			},
		},
		{
			desc:        "partly and fully qualified columns with same table name and multiple tables with same name but different databases",
			query:       "select foo.a, db.foo.b from foo, db.foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
				"db": {
					"foo": {"a", "b"}, // it should assume a is also possibly in db.foo since this query is ambiguous
				},
			},
		},
	}

	nonsimpleQueries := []tcase{
		{
			desc:        "WHERE clause, multiple unqualified columns with single unqualified table",
			query:       "select a, b from foo where c = d",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c", "d"},
				},
			},
		},
		{
			desc:        "WHERE clause, multiple unqualified columns with single qualified table",
			query:       "select a, b from db.foo where c = d",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a", "b", "c", "d"},
				},
			},
		},
		{
			desc:        "WHERE clause, multiple unqualified columns with multiple unqualified tables",
			query:       "select a, b from foo, bar where c = d",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c", "d"},
					"bar": {"a", "b", "c", "d"},
				},
			},
		},
		{
			desc:        "WHERE clause, multiple unqualified columns with multiple qualified tables",
			query:       "select a, b from test.foo, db.bar where c = d",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c", "d"},
				},
				"db": {
					"bar": {"a", "b", "c", "d"},
				},
			},
		},
		{
			desc:        "WHERE clause, multiple partly qualified columns with multiple unqualified table",
			query:       "select foo.a, bar.b from foo, bar where foo.c = bar.d",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "c"},
					"bar": {"b", "d"},
				},
			},
		},
		{
			desc:        "WHERE clause, multiple fully qualified columns with multiple qualified table",
			query:       "select foo.a from test.foo, db.bar where db.bar.b > 1",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
				"db": {
					"bar": {"b"},
				},
			},
		},
		{
			desc:        "WHERE clause, same qualified columns in multiple clauses",
			query:       "select foo.a, bar.b from foo, bar where foo.a = bar.b",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
					"bar": {"b"},
				},
			},
		},
		{
			desc:        "WHERE clause, partly qualified column using fully qualified table",
			query:       "select foo.a, bar.b from foo, db.bar where foo.a = bar.b",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
				"db": {
					"bar": {"b"},
				},
			},
		},
		{
			desc:        "All components, single unqualified column with single unqualified table",
			query:       "select a from foo where a > 1 group by a having a > 2 order by a",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "All components, multiple unqualified columns with single unqualified table",
			query:       "select a from foo where b > 1 group by c having d > 2 order by e",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c", "d", "e"},
				},
			},
		},
		{
			desc:        "All components, multiple unqualified columns with single unqualified table",
			query:       "select a from foo where b > 1 group by c having d > 2 order by e",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c", "d", "e"},
				},
			},
		},
		{
			desc:        "All components, multiple unqualified columns with multiple unqualified tables",
			query:       "select a from foo, bar where b > 1 group by c having d > 2 order by e",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c", "d", "e"},
					"bar": {"a", "b", "c", "d", "e"},
				},
			},
		},
		{
			desc:        "All components, multiple partly qualified columns with multiple unqualified tables",
			query:       "select foo.a from foo, bar where bar.b > 1 group by foo.c having bar.d > 2 order by e",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "c", "e"},
					"bar": {"b", "d", "e"},
				},
			},
		},
		{
			desc:        "All components, mixed qualifications",
			query:       "select test.foo.a, db.bar.b, baz.c, d from foo, db.bar, baz where bar.b > 1 group by test.baz.c having foo.e > 2 order by f",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "baz"},
				"db":          {"bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "d", "e", "f"},
					"baz": {"c", "d", "f"},
				},
				"db": {
					"bar": {"b", "d", "f"},
				},
			},
		},
		{
			desc:        "JOIN with partly qualified columns",
			query:       "select * from foo join bar on foo.a = bar.b",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
					"bar": {"b"},
				},
			},
		},
	}

	queriesWithAliases := []tcase{
		{
			desc:        "single unqualified column with single unqualified aliased table",
			query:       "select a from foo as t1",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "single qualified column with single unqualified aliased table",
			query:       "select t1.a from foo as t1",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "single qualified column with single qualified aliased table",
			query:       "select t1.a from db.foo as t1",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db"},
			expectedTables: map[string][]string{
				"db": {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				"db": {
					"foo": {"a"},
				},
			},
		},
		{
			desc:        "multiple qualified columns with single unqualified aliased table",
			query:       "select t1.a, t1.b, t1.c from foo as t1",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple qualified columns with single unqualified aliased table with multiple aliases",
			query:       "select t1.a, t2.b, t2.c from foo as t1, foo as t2",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a", "b", "c"},
				},
			},
		},
		{
			desc:        "multiple qualified columns with mixed qualified aliased tables",
			query:       "select t1.a, t2.b, t2.c from foo as t1, db.foo as t2",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB, "db"},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo"},
				"db":          {"foo"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
				},
				"db": {
					"foo": {"b", "c"},
				},
			},
		},
		{
			desc:        "JOIN with partly qualified columns",
			query:       "select * from foo t1 join bar t2 on t1.a = t2.b",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
					"bar": {"b"},
				},
			},
		},
		{
			desc:        "partly qualified columns with tables of same name with different aliases",
			query:       "select t1.a, t2.b from db1.t1, db2.t1 as t2",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{"db1", "db2"},
			expectedTables: map[string][]string{
				"db1": {"t1"},
				"db2": {"t1"},
			},
			expectedColumns: map[string]map[string][]string{
				"db1": {
					"t1": {"a"},
				},
				"db2": {
					"t1": {"b"},
				},
			},
		},
		{
			desc:        "tables aliased with the other's name",
			query:       "select foo.a, bar.b from foo as bar, bar as foo",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "bar"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"b"},
					"bar": {"a"},
				},
			},
		},
		{
			desc:        "nested alias (with same name) in subquery",
			query:       "select t1.a, t2.a from foo as t1, (select foo.b as a from t2 as foo) as t2",
			defaultDB:   testDefaultDB,
			expectedDBs: []string{testDefaultDB},
			expectedTables: map[string][]string{
				testDefaultDB: {"foo", "t2"},
			},
			expectedColumns: map[string]map[string][]string{
				testDefaultDB: {
					"foo": {"a"},
					"t2":  {"b"},
				},
			},
		},
	}

	t.Run("Simple SELECT queries", func(t *testing.T) {
		testQueries(simpleSelects, t)
	})

	t.Run("Non-simple queries", func(t *testing.T) {
		testQueries(nonsimpleQueries, t)
	})

	t.Run("Queries with aliases", func(t *testing.T) {
		testQueries(queriesWithAliases, t)
	})
}

func testQueries(tcases []tcase, t *testing.T) {
	for _, tcase := range tcases {
		ns, err := InferSchemaFromQuery(tcase.query, tcase.defaultDB)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
		}

		actualDBs := ns.Databases()

		// ensure no unexpected databases were inferred.
		for _, actualDB := range actualDBs {
			if !contains(tcase.expectedDBs, actualDB) {
				t.Errorf("%s: unexpected database '%s' inferred from query", tcase.desc, actualDB)
			}
		}

		for _, expectedDB := range tcase.expectedDBs {
			// ensure all expected databases were inferred.
			if !contains(actualDBs, expectedDB) {
				t.Errorf("%s: expected database '%s' not inferred from query", tcase.desc, expectedDB)
			}

			actualTables, _ := ns.Tables(expectedDB)
			expectedTables := tcase.expectedTables[expectedDB]

			// ensure no unexpected tables were inferred.
			for _, actualTable := range actualTables {
				if !contains(expectedTables, actualTable) {
					t.Errorf("%s: unexpected table '%s.%s' inferred from query", tcase.desc, expectedDB, actualTable)
				}
			}

			for _, expectedTable := range expectedTables {
				// ensure all expected tables were inferred.
				if !contains(actualTables, expectedTable) {
					t.Errorf("%s: expected table '%s.%s' not inferred from query", tcase.desc, expectedDB, expectedTable)
				}

				actualColumns, _ := ns.Columns(expectedDB, expectedTable)
				expectedColumns := tcase.expectedColumns[expectedDB][expectedTable]

				// ensure no unexpected columns were inferred.
				for _, actualColumn := range actualColumns {
					if !contains(expectedColumns, actualColumn) {
						t.Errorf("%s: unexpected column '%s.%s.%s' inferred from query", tcase.desc, expectedDB, expectedTable, actualColumn)
					}
				}

				// ensure all expected columns were inferred.
				for _, expectedColumn := range expectedColumns {
					if !contains(actualColumns, expectedColumn) {
						t.Errorf("%s: expected column '%s.%s.%s' not inferred from query", tcase.desc, expectedDB, expectedTable, expectedColumn)
					}
				}
			}
		}
	}
}

func contains(a []string, s string) bool {
	for _, str := range a {
		if s == str {
			return true
		}
	}

	return false
}
