use crate::ast::{rewrites::tuples::SingleTupleRewriteVisitor, visitor::Visitor};
use crate::parser::Parser;

macro_rules! query_printer_test {
    ($func_name:ident, $should_print_as:expr, $input:expr) => {
        #[test]
        fn $func_name() {
            let res = Parser::new().parse_query($input).unwrap();
            let out = format!("{}", res);
            assert_eq!(out, $should_print_as);
        }
    };
}

query_printer_test!(union, "SELECT * UNION SELECT *", "select * unioN SElect *");
query_printer_test!(
    union_all,
    "SELECT * UNION ALL SELECT *",
    "select * unioN ALL SElect *"
);

query_printer_test!(select_star, "SELECT *", "select *");
query_printer_test!(select_sub_star, "SELECT foo.*", "select foo.*");
query_printer_test!(select_alias_expression, "SELECT 42 AS bar", "select 42 bar");
query_printer_test!(
    select_alias_expression_and_substar,
    "SELECT 42 AS bar, foo.*, 42 AS car, fuzz.*",
    "select 42 bar, foo.*, 42 car, fuzz.*"
);
query_printer_test!(
    select_distinct_alias_expression_and_substar,
    "SELECT DISTINCT 42 AS bar, foo.*, 42 AS car, fuzz.*",
    "select diStinCT 42 bar, foo.*, 42 car, fuzz.*"
);

query_printer_test!(
    select_values_sub_star,
    "SELECT VALUE foo.*",
    "select VALUES foo.*"
);
query_printer_test!(
    select_values_document_expression,
    "SELECT VALUE {'bar': 42}",
    "select VAlUEs {'bar': 42}"
);
query_printer_test!(
    select_values_alias_expression_and_substar,
    "SELECT VALUES {'bar': 42, 'car': 42}, foo.*, fuzz.*, {'hello': 'world'}",
    "select VaLuES {'bar': 42, 'car': 42}, foo.*, fuzz.*, {'hello': 'world'}"
);

query_printer_test!(
    select_from_array,
    "SELECT foo.* FROM [{'a': 42, 'b': 42}, {'a': 42, 'b': 43}] AS foo",
    "SeLeCT foo.* from [{'a':  42, 'b':   42},    {'a': 42, 'b': 43}] foo"
);

query_printer_test!(
    select_from_local_collection,
    "SELECT foo.* FROM foo",
    "SeLeCT foo.* from foo"
);
query_printer_test!(
    select_from_local_collection_with_alias,
    "SELECT foo.* FROM foo AS foo",
    "SeLeCT foo.* from foo foo"
);
query_printer_test!(
    select_from_qualified_collection,
    "SELECT foo.* FROM bar.foo",
    "SeLeCT foo.* from bar.foo"
);
query_printer_test!(
    select_from_qualified_collection_with_alias,
    "SELECT foo.* FROM bar.foo AS foo",
    "SeLeCT foo.* from bar.foo foo"
);

query_printer_test!(
    select_from_derived,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo",
    "SeLeCT foo.* from (SELECT * FROM bar) foo"
);

query_printer_test!(
    select_from_left_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar",
    "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_left_outer_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar",
    "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT OUTER JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_right_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo RIGHT JOIN [{'a': 32}] AS zar",
    "SeLeCT foo.* from (SELECT * FROM bar) foo RIGHT JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_right_outer_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo RIGHT JOIN [{'a': 32}] AS zar",
    "SeLeCT foo.* from (SELECT * FROM bar) foo RIGHT OUTER JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_inner_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo INNER JOIN [{'a': 32}] AS zar",
    "SeLeCT foo.* from (SELECT * FROM bar) foo INNER JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_cross_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo CROSS JOIN [{'a': 32}] AS zar",
    "SeLeCT foo.* from (SELECT * FROM bar) foo JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_explicit_cross_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo CROSS JOIN [{'a': 32}] AS zar",
    "SeLeCT foo.* from (SELECT * FROM bar) foo CROSS JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_left_join_with_condition,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar ON true",
    "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar On tRuE"
);
query_printer_test!(
    select_from_three_way_join,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar CROSS JOIN foo AS foo",
    "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar JOIN foo AS foo"
);
query_printer_test!(
    select_from_three_way_join_with_condition,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar CROSS JOIN foo AS foo ON 42 = 43",
    "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar JOIN foo AS foo ON 42 = 43"
);
query_printer_test!(
    select_from_three_way_join_with_two_conditions,
    "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar ON true CROSS JOIN foo AS foo ON 42 = 43",
    "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar oN TRUE JOIN foo AS foo ON 42 = 43"
);

query_printer_test!(
    select_where_1,
    "SELECT * FROM foo AS bar WHERE 1",
    "SELECT * FROM foo bar WHERE 1"
);
query_printer_test!(
    select_where_eq,
    "SELECT * FROM foo AS bar WHERE 1 = 2",
    "SELECT * FROM foo bar WHERE 1 = 2"
);

query_printer_test!(
    select_having_1,
    "SELECT * FROM foo AS bar HAVING 1",
    "SELECT * FROM foo bar HAVING 1"
);
query_printer_test!(
    select_having_eq,
    "SELECT * FROM foo AS bar HAVING 1 = 2",
    "SELECT * FROM foo bar HAVING 1 = 2"
);

query_printer_test!(
    select_where_1_having_1,
    "SELECT * FROM foo AS bar WHERE 1 HAVING 1",
    "SELECT * FROM foo bar WHERE 1 HAVING 1"
);
query_printer_test!(
    select_where_eq_having_eq,
    "SELECT * FROM foo AS bar WHERE 1 = 2 HAVING 1 = 2",
    "SELECT * FROM foo bar WHERE 1 = 2 HAVING 1 = 2"
);

query_printer_test!(
    select_group_by_null,
    "SELECT * FROM foo GROUP BY NULL",
    "SELECT * FROM foo GROUP BY NULL"
);
query_printer_test!(
    select_group_by_key,
    "SELECT * FROM foo GROUP BY a",
    "SELECT * FROM foo GROUP BY a"
);
query_printer_test!(
    select_group_by_keys,
    "SELECT * FROM foo GROUP BY a, b",
    "SELECT * FROM foo GROUP BY a, b"
);
query_printer_test!(
    select_group_by_keys_with_alias,
    "SELECT * FROM foo GROUP BY a, b AS b",
    "SELECT * FROM foo GROUP BY a, b b"
);
query_printer_test!(
    select_group_by_keys_with_aliases,
    "SELECT * FROM foo GROUP BY a AS c, b AS b",
    "SELECT * FROM foo GROUP BY a c, b b"
);
query_printer_test!(
    select_group_by_one_aggregates,
    "SELECT * FROM foo GROUP BY a AS c, b AS b AGGREGATE COUNT(*) AS agg1",
    "SELECT * FROM foo GROUP BY a c, b b AGGREGATE COUNT(*) agg1"
);
query_printer_test!(
    select_group_by_two_aggregates,
    "SELECT agg1, agg2 FROM foo GROUP BY a AS c, b AS b AGGREGATE COUNT(*) AS agg1, SUM(foo) AS agg2",
    "SELECT agg1, agg2 FROM foo GROUP BY a c, b b AGGREGATE COUNT(*) agg1, SUM(foo) agg2"
);

query_printer_test!(
    select_order_by_1_asc,
    "SELECT * FROM foo ORDER BY 1",
    "selECT * FROM foo ORDER BY 1 ASC"
);
query_printer_test!(
    select_order_by_1_desc,
    "SELECT * FROM foo ORDER BY 1 DESC",
    "selECT * FROM foo ORDER BY 1 DESC"
);
query_printer_test!(
    select_order_by_a_asc,
    "SELECT * FROM foo ORDER BY a",
    "selECT * FROM foo ORDER BY a ASC"
);
query_printer_test!(
    select_order_by_a_desc,
    "SELECT * FROM foo ORDER BY a DESC",
    "selECT * FROM foo ORDER BY a DESC"
);
query_printer_test!(
    select_order_by_1_asc_2_asc,
    "SELECT * FROM foo ORDER BY 1, 2",
    "selECT * FROM foo ORDER BY 1 ASC, 2"
);
query_printer_test!(
    select_order_by_1_desc_2_desc,
    "SELECT * FROM foo ORDER BY 1 DESC, 2 DESC",
    "selECT * FROM foo ORDER BY 1 DESC, 2 DESC"
);
query_printer_test!(
    select_order_by_a_asc_b_asc,
    "SELECT * FROM foo ORDER BY a, b",
    "selECT * FROM foo ORDER BY a ASC, b ASC"
);
query_printer_test!(
    select_order_by_a_desc_b_desc,
    "SELECT * FROM foo ORDER BY a DESC, b DESC",
    "selECT * FROM foo ORDER BY a DESC, b DESC"
);

query_printer_test!(
    select_limit_1,
    "SELECT * FROM foo AS bar LIMIT 1",
    "SELECT * FROM foo bar LIMIT 1"
);
query_printer_test!(
    select_offset_1,
    "SELECT * FROM foo AS bar OFFSET 1",
    "SELECT * FROM foo bar OFFSET 1"
);
query_printer_test!(
    select_limit_1_offset_1,
    "SELECT * FROM foo AS bar LIMIT 1 OFFSET 1",
    "SELECT * FROM foo bar LIMIT 1 OFFSET 1"
);
query_printer_test!(
    select_offset_1_limit_1,
    "SELECT * FROM foo AS bar LIMIT 1 OFFSET 1",
    "SELECT * FROM foo bar OFFSET 1 LIMIT 1"
);

query_printer_test!(
    select_all_clauses,
    "SELECT * FROM foo AS bar WHERE 1 GROUP BY a, b AGGREGATE COUNT(*) AS agg1, SUM(a) AS agg2 HAVING agg1 < agg2 ORDER BY agg1 LIMIT 100 OFFSET 10",
    "SELECT * FROM foo bar WHERE 1 GROUP BY a, b AGGREGATE COUNT(*) AS agg1, SUM(a) as agg2 HAVING agg1 < agg2 ORDER BY agg1 LIMIT 100 OFFSET 10"
);

macro_rules! expression_printer_test {
    ($func_name:ident, $should_print_as:expr, $input:expr) => {
        #[test]
        fn $func_name() {
            let res = Parser::new().parse_expression($input).unwrap();
            let res = SingleTupleRewriteVisitor {}.visit_expression(res);
            let out = format!("{}", res);
            assert_eq!(out, $should_print_as);
        }
    };
}

expression_printer_test!(identifier_non_latin_first_char, "`做`", "`做`");
expression_printer_test!(identifier_non_latin_subsequent_chars, "`_做`", "`_做`");
expression_printer_test!(delimited_identifier_containing_backtick, "````", "````");
expression_printer_test!(normal_identifiers, "(foo - bar) / car", "(foo - bar) / car");
expression_printer_test!(
    special_identifiers,
    "`fo.o` - bar / `$car`",
    "`fo.o` - (bar / `$car`)"
);
expression_printer_test!(starts_with_number, "`1foo`", "`1foo`");
expression_printer_test!(starts_with_underscore, "_foo", "`_foo`");
expression_printer_test!(regular_identifier_containing_number, "foo1", "foo1");
expression_printer_test!(empty_identifier, "``", "``");
expression_printer_test!(
    is_missing,
    "true AND (x IS MISSING)",
    "true AND (x IS MISSING)"
);
expression_printer_test!(is_type, "true AND (x IS INT)", "true AND (x IS int)");

expression_printer_test!(
    like_simple,
    "true AND (x LIKE '%hello%')",
    "true AND (x LIKE '%hello%')"
);
expression_printer_test!(
    like_escape,
    "true AND (x LIKE '%hello%' ESCAPE '@')",
    "true AND (x LIKE '%hello%' ESCAPE '@')"
);

expression_printer_test!(type_assert, "true AND x::!INT", "true AND (x::!INT)");

expression_printer_test!(tuple, "(a, b, c)", "(a, b, c)");

expression_printer_test!(
    cast_sigil_array,
    "CAST(4 + foo AS ARRAY)",
    "(4 + foo)::array"
);
expression_printer_test!(
    cast_sigil_bindata,
    "CAST(4 + foo AS BINDATA)",
    "(4 + foo)::bindata"
);
expression_printer_test!(
    cast_sigil_boolean,
    "CAST(4 + foo AS BOOLEAN)",
    "(4 + foo)::boolean"
);
expression_printer_test!(
    cast_sigil_datetime,
    "CAST(4 + foo AS BSON_DATE)",
    "(4 + foo)::bson_date"
);
expression_printer_test!(
    cast_sigil_dbpointer,
    "CAST(4 + foo AS DBPOINTER)",
    "(4 + foo)::dbpointer"
);
expression_printer_test!(
    cast_sigil_decimal_prec,
    "CAST(4 + foo AS DECIMAL)",
    "(4 + foo)::decimal(45)"
);
expression_printer_test!(
    cast_sigil_decimal,
    "CAST(4 + foo AS DECIMAL)",
    "(4 + foo)::decimal"
);
expression_printer_test!(
    cast_sigil_document,
    "CAST(4 + foo AS DOCUMENT)",
    "(4 + foo)::document"
);
expression_printer_test!(
    cast_sigil_float_prec,
    "CAST(4 + foo AS DOUBLE)",
    "(4 + foo)::float(10)"
);
expression_printer_test!(
    cast_sigil_double,
    "CAST(4 + foo AS DOUBLE)",
    "(4 + foo)::double"
);
expression_printer_test!(cast_sigil_int, "CAST(4 + foo AS INT)", "(4 + foo)::int");
expression_printer_test!(cast_sigil_long, "CAST(4 + foo AS LONG)", "(4 + foo)::long");
expression_printer_test!(
    cast_sigil_javascript,
    "CAST(4 + foo AS JAVASCRIPT)",
    "(4 + foo)::javascript"
);
expression_printer_test!(
    cast_sigil_javascriptwithscope,
    "CAST(4 + foo AS JAVASCRIPTWITHSCOPE)",
    "(4 + foo)::javascriptwithscope"
);
expression_printer_test!(
    cast_sigil_maxkey,
    "CAST(4 + foo AS MAXKEY)",
    "(4 + foo)::maxkey"
);
expression_printer_test!(
    cast_sigil_minkey,
    "CAST(4 + foo AS MINKEY)",
    "(4 + foo)::minkey"
);
expression_printer_test!(cast_sigil_null, "CAST(4 + foo AS NULL)", "(4 + foo)::null");
expression_printer_test!(
    cast_sigil_objectid,
    "CAST(4 + foo AS OBJECTID)",
    "(4 + foo)::objectid"
);
expression_printer_test!(
    cast_sigil_regularexpression,
    "CAST(4 + foo AS REGEX)",
    "(4 + foo)::regex"
);
expression_printer_test!(
    cast_sigil_string_prec,
    "CAST(4 + foo AS STRING)",
    "(4 + foo)::varchar(5)"
);
expression_printer_test!(
    cast_sigil_string,
    "CAST(4 + foo AS STRING)",
    "(4 + foo)::string"
);
expression_printer_test!(
    cast_sigil_symbol,
    "CAST(4 + foo AS SYMBOL)",
    "(4 + foo)::symbol"
);
expression_printer_test!(
    cast_sigil_timestamp,
    "CAST(4 + foo AS BSON_DATE)",
    "(4 + foo)::timestamp"
);
expression_printer_test!(
    cast_sigil_undefined,
    "CAST(4 + foo AS UNDEFINED)",
    "(4 + foo)::undefined"
);

expression_printer_test!(cast_int, "CAST(x AS INT)", "CAST(x as int)");
expression_printer_test!(
    cast_int_on_null,
    "CAST(x AS INT, 3 + 4 ON NULL)",
    "CAST(x as int, 3+4 ON NULL)"
);
expression_printer_test!(
    cast_int_on_error,
    "CAST(x AS INT, 3 + 4 ON ERROR)",
    "CAST(x as int, 3+4 ON ERROR)"
);
expression_printer_test!(
    cast_int_on_null_on_error,
    "CAST(x AS INT, 'bar' ON NULL, 'foo' ON ERROR)",
    "CAST(x as int, 'bar' ON NULL, 'foo' ON ERROR)"
);

expression_printer_test!(literal_null, "NULL", "nUlL");
expression_printer_test!(literal_true, "true", "tRuE");
expression_printer_test!(literal_false, "false", "fALse");
expression_printer_test!(literal_string, "'hellO'", "'hellO'");
expression_printer_test!(literal_string_with_single_quote, "'''hellO'", "'''hellO'");
expression_printer_test!(literal_int, "1", "1");
expression_printer_test!(literal_long, "8000000000", "8000000000");
expression_printer_test!(literal_double, "81111.1", "8.11111e4");
expression_printer_test!(literal_double_no_fraction, "8000.0", "8e3");

expression_printer_test!(unary_not, "NOT true", "not (true)");
expression_printer_test!(
    unary_not_parens,
    "NOT (true OR false)",
    "not (true OR false)"
);
expression_printer_test!(unary_neg, "-3", "- 3");
expression_printer_test!(unary_neg_parens, "-(3 + 4)", "-(3+4)");
expression_printer_test!(unary_pos, "3", "+ (3)");
expression_printer_test!(unary_pos_parens, "3 + 4", "+(3+4)");
expression_printer_test!(
    unary_negative_sub_function,
    "4 - SUM(bar)",
    "4 - + SUM(bar)"
);
expression_printer_test!(
    unary_positive_sub_function,
    "4 - -SUM(bar)",
    "4 - - SUM(bar)"
);
expression_printer_test!(
    unary_not_sub_function,
    "4 - NOT SUM(bar)",
    "4 - not SUM(bar)"
);

expression_printer_test!(binary_add_and_mul_no_parens, "3 + 4 * 6", "3 + 4 * 6");
expression_printer_test!(
    binary_add_and_mul_unncessary_parens,
    "3 + 4 * 6",
    "3 + (4 * 6)"
);
expression_printer_test!(binary_add_and_mul_parens, "(3 + 4) * 6", "(3 + 4) * 6");
expression_printer_test!(
    binary_add_and_mul_extra_parens,
    "(3 + 4) * 6",
    "((3 + 4)) * 6"
);

expression_printer_test!(
    binary_lt_and_and_parens,
    "(true AND 3) < 4",
    "(true AND 3) < 4"
);
expression_printer_test!(
    binary_lt_and_and_no_parens,
    "true AND 3 < 4",
    "true AND 3 < 4"
);

expression_printer_test!(
    binary_and_or_no_parens,
    "true AND false OR false",
    "true AND false OR false"
);
expression_printer_test!(
    binary_and_or_parens,
    "true AND (false OR false)",
    "true AND (false OR false)"
);

expression_printer_test!(binary_or, "hello OR world", "hello OR world");
expression_printer_test!(binary_concat, "hello || world", "hello || world");
expression_printer_test!(binary_lt, "hello < world", "hello < world");
expression_printer_test!(binary_lte, "hello <= world", "hello <= world");
expression_printer_test!(binary_gt, "hello > world", "hello > world");
expression_printer_test!(binary_gte, "hello >= world", "hello >= world");
expression_printer_test!(binary_eq, "hello = world", "hello = world");
expression_printer_test!(binary_neq, "hello <> world", "hello <> world");
expression_printer_test!(binary_neq2, "hello <> world", "hello != world");
expression_printer_test!(binary_add, "hello + world", "hello + world");
expression_printer_test!(binary_sub, "hello - world", "hello - world");
expression_printer_test!(binary_mul, "hello * world", "hello * world");
expression_printer_test!(binary_div, "hello / world", "hello / world");
expression_printer_test!(binary_in, "hello IN world", "hello in world");
expression_printer_test!(binary_not_in, "hello NOT IN world", "hello nOt in world");

expression_printer_test!(
    case_basic,
    "CASE WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END",
    "CASE WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END"
);
expression_printer_test!(
    case_basic_with_expr,
    "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END",
    "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END"
);
expression_printer_test!(
    case_basic_with_expr_and_else,
    "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false ELSE foo + bar END",
    "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false ELSE foo+bar END"
);
expression_printer_test!(
    case_as_expr,
    "true OR CASE foo * 3 WHEN x + 3 = y THEN true ELSE false END",
    "true OR (CASE foo * 3 WHEN x + 3 = y THEN true ELSE false END)"
);

expression_printer_test!(
    function_basic,
    "COALESCE(bar, foo + bar, hello * 3)",
    "COALESCE(bar, foo+bar, hello*3)"
);
expression_printer_test!(function_sub, "4 - SUM(bar)", "4 - SUM(bar)");
expression_printer_test!(
    function_extract_day,
    "EXTRACT(DAY FROM bar)",
    "extract(DAY FROM bar)"
);
expression_printer_test!(
    function_extract_hour,
    "EXTRACT(HOUR FROM bar)",
    "extract(HOUR FROM bar)"
);
expression_printer_test!(
    function_extract_minute,
    "EXTRACT(MINUTE FROM bar)",
    "extract(MINUTE FROM bar)"
);
expression_printer_test!(
    function_extract_month,
    "EXTRACT(MONTH FROM bar)",
    "extract(MONTH FROM bar)"
);
expression_printer_test!(
    function_extract_second,
    "EXTRACT(SECOND FROM bar)",
    "extract(SECOND FROM bar)"
);
expression_printer_test!(
    function_extract_year,
    "EXTRACT(YEAR FROM bar)",
    "extract(YEAR FROM bar)"
);

expression_printer_test!(
    position_ident_in_ident,
    "POSITION(x IN y)",
    "position(x in y)"
);
expression_printer_test!(
    position_string_in_ident,
    "POSITION('x' IN y)",
    "position('x' in y)"
);
expression_printer_test!(
    position_ident_in_string,
    "POSITION(x IN 'y')",
    "position(x in 'y')"
);
expression_printer_test!(
    position_string_in_string,
    "POSITION('x' IN 'y')",
    "position('x' in 'y')"
);

expression_printer_test!(function_upper, "UPPER('hello')", "UPPER('hello')");
expression_printer_test!(function_lower, "LOWER('hello')", "LOWER('hello')");

expression_printer_test!(function_trim_basic, "TRIM(BOTH FROM hello)", "TRIM(hello)");
expression_printer_test!(
    function_trim_basic_chars,
    "TRIM(BOTH 'asdf' FROM hello)",
    "TRIM('asdf' FROM hello)"
);
expression_printer_test!(
    function_trim_explicit_both,
    "TRIM(BOTH FROM hello)",
    "TRIM(BOTH FROM hello)"
);
expression_printer_test!(
    function_trim_leading,
    "TRIM(LEADING FROM hello)",
    "TRIM(LEADING FROM hello)"
);
expression_printer_test!(
    function_trim_trailing,
    "TRIM(TRAILING FROM hello)",
    "TRIM(TRAILING FROM hello)"
);
expression_printer_test!(
    function_trim_explicit_both_chars,
    "TRIM(BOTH 'asdf' FROM hello)",
    "TRIM(BOTH 'asdf' FROM hello)"
);
expression_printer_test!(
    function_trim_leading_chars,
    "TRIM(LEADING 'asdf' FROM hello)",
    "TRIM(LEADING 'asdf' FROM hello)"
);
expression_printer_test!(
    function_trim_trailing_chars,
    "TRIM(TRAILING 'asdf' FROM hello)",
    "TRIM(TRAILING 'asdf' FROM hello)"
);

expression_printer_test!(access_int, "[1, 2, 3][1]", "[1, 2, 3] [1]");
expression_printer_test!(access_string, "{'a': 3}['a']", "{'a': 3} ['a']");
expression_printer_test!(access_identifier_string, "a['a']", "a ['a']");
expression_printer_test!(access_string_with_parens, "(3 + 4)['a']", "(3+4) ['a']");
expression_printer_test!(access_access, "a[1][2]", "a[1][2]");

expression_printer_test!(subpath, "{'a': 3}.a", "{'a': 3} .a");
expression_printer_test!(subpath_identifier, "a.a", "a .a");
expression_printer_test!(subpath_with_parens, "(3 + 4).a", "(3+4) .a");

expression_printer_test!(subpath_access, "a[1].b", "a[1].b");
expression_printer_test!(access_subpath, "a[1].b", "a[1].b");
expression_printer_test!(subpath_subpath, "a.b.c", "a.b.c");

expression_printer_test!(
    sub_query,
    "1 + (SELECT * UNION SELECT *)",
    "1+(select * unioN SElect *)"
);
expression_printer_test!(
    exists,
    "true AND EXISTS(SELECT * UNION SELECT *)",
    "true and ExisTS(select * unioN SElect *)"
);
expression_printer_test!(
    subquery_eq_any,
    "x = ANY(SELECT * UNION SELECT *)",
    "x=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_eq_all,
    "x = ALL(SELECT * UNION SELECT *)",
    "x=ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lt_any,
    "x < ANY(SELECT * UNION SELECT *)",
    "x<ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lt_all,
    "x < ALL(SELECT * UNION SELECT *)",
    "x<ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lte_any,
    "x <= ANY(SELECT * UNION SELECT *)",
    "x<=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lte_all,
    "x <= ALL(SELECT * UNION SELECT *)",
    "x<=ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gt_any,
    "x > ANY(SELECT * UNION SELECT *)",
    "x>ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gt_all,
    "x > ALL(SELECT * UNION SELECT *)",
    "x>ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gte_any,
    "x >= ANY(SELECT * UNION SELECT *)",
    "x>=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gte_all,
    "x >= ALL(SELECT * UNION SELECT *)",
    "x>=ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq_any,
    "x <> ANY(SELECT * UNION SELECT *)",
    "x<>ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq_all,
    "x <> ALL(SELECT * UNION SELECT *)",
    "x<>ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq2_any,
    "x <> ANY(SELECT * UNION SELECT *)",
    "x!=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq2_all,
    "x <> ALL(SELECT * UNION SELECT *)",
    "x!=ALL(SELECT * UNION SELECT *)"
);
