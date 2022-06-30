use crate::ast::{
    pretty_print::PrettyPrint, rewrites::tuples::SingleTupleRewriteVisitor, visitor::Visitor,
};
use crate::parser::Parser;

macro_rules! query_printer_test {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let res = Parser::new().parse_query($input).unwrap();
            let out = res.pretty_print().unwrap();
            assert_eq!($expected, out);
        }
    };
}

macro_rules! expression_printer_test {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let res = Parser::new().parse_expression($input).unwrap();
            let res = SingleTupleRewriteVisitor {}.visit_expression(res);
            let out = res.pretty_print().unwrap();
            assert_eq!($expected, out);
        }
    };
}

query_printer_test!(
    union,
    expected = "SELECT * UNION SELECT *",
    input = "select * unioN SElect *"
);
query_printer_test!(
    union_all,
    expected = "SELECT * UNION ALL SELECT *",
    input = "select * unioN ALL SElect *"
);

query_printer_test!(select_star, expected = "SELECT *", input = "select *");
query_printer_test!(
    select_sub_star,
    expected = "SELECT foo.*",
    input = "select foo.*"
);
query_printer_test!(
    select_delimited_sub_star,
    expected = "SELECT `f\"o\"o`.*",
    input = "select `f\"o\"o`.*"
);
query_printer_test!(
    select_alias_expression,
    expected = "SELECT 42 AS bar",
    input = "select 42 bar"
);
query_printer_test!(
    select_delimited_alias_expression,
    expected = "SELECT 42 AS `b\"a\"r`",
    input = "select 42 `b\"a\"r`"
);
query_printer_test!(
    select_alias_expression_and_substar,
    expected = "SELECT 42 AS bar, foo.*, 42 AS car, fuzz.*",
    input = "select 42 bar, foo.*, 42 car, fuzz.*"
);
query_printer_test!(
    select_distinct_alias_expression_and_substar,
    expected = "SELECT DISTINCT 42 AS bar, foo.*, 42 AS car, fuzz.*",
    input = "select diStinCT 42 bar, foo.*, 42 car, fuzz.*"
);

query_printer_test!(
    select_values_sub_star,
    expected = "SELECT VALUE foo.*",
    input = "select VALUES foo.*"
);
query_printer_test!(
    select_values_document_expression,
    expected = "SELECT VALUE {'bar': 42}",
    input = "select VAlUEs {'bar': 42}"
);
query_printer_test!(
    select_values_alias_expression_and_substar,
    expected = "SELECT VALUES {'bar': 42, 'car': 42}, foo.*, fuzz.*, {'hello': 'world'}",
    input = "select VaLuES {'bar': 42, 'car': 42}, foo.*, fuzz.*, {'hello': 'world'}"
);

query_printer_test!(
    select_from_array,
    expected = "SELECT foo.* FROM [{'a': 42, 'b': 42}, {'a': 42, 'b': 43}] AS foo",
    input = "SeLeCT foo.* from [{'a':  42, 'b':   42},    {'a': 42, 'b': 43}] foo"
);
query_printer_test!(
    select_from_delimited_alias_array,
    expected = "SELECT foo.* FROM [{'a': 42, 'b': 42}, {'a': 42, 'b': 43}] AS `f\"o\"o`",
    input = "SeLeCT foo.* from [{'a':  42, 'b':   42},    {'a': 42, 'b': 43}] `f\"o\"o`"
);

query_printer_test!(
    select_from_local_collection,
    expected = "SELECT foo.* FROM foo",
    input = "SeLeCT foo.* from foo"
);
query_printer_test!(
    select_from_local_collection_with_alias,
    expected = "SELECT foo.* FROM foo AS foo",
    input = "SeLeCT foo.* from foo foo"
);
query_printer_test!(
    select_from_local_collection_with_delimited_alias,
    expected = "SELECT foo.* FROM foo AS `f\"o\"o`",
    input = "SeLeCT foo.* from foo `f\"o\"o`"
);
query_printer_test!(
    select_from_qualified_collection,
    expected = "SELECT foo.* FROM bar.foo",
    input = "SeLeCT foo.* from bar.foo"
);
query_printer_test!(
    select_from_collection_with_delimited_qualifier,
    expected = "SELECT foo.* FROM `bar%`.foo",
    input = "SeLeCT foo.* from `bar%`.foo"
);
query_printer_test!(
    select_from_qualified_collection_with_alias,
    expected = "SELECT foo.* FROM bar.foo AS foo",
    input = "SeLeCT foo.* from bar.foo foo"
);

query_printer_test!(
    select_from_delimited_db_collection_and_alias,
    expected = "SELECT * FROM `db%`.`coll%` AS `a%&`",
    input = "SELECT * FROM `db%`.`coll%` AS `a%&`"
);

query_printer_test!(
    select_from_delimited_collection,
    expected = "SELECT * FROM db.`coll%` AS a",
    input = "SELECT * FROM db.`coll%` AS a"
);

query_printer_test!(
    select_from_qualified_collection_with_delimited_alias,
    expected = "SELECT foo.* FROM bar.foo AS `f\"o\"o`",
    input = "SeLeCT foo.* from bar.foo `f\"o\"o`"
);

query_printer_test!(
    select_from_derived,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo"
);
query_printer_test!(
    select_from_derived_delimited_alias,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS `f\"o\"o`",
    input = "SeLeCT foo.* from (SELECT * FROM bar) `f\"o\"o`"
);

query_printer_test!(
    select_from_left_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_left_join_delimited_aliases,
    expected =
        "SELECT foo.* FROM (SELECT * FROM bar) AS `f\"o\"o` LEFT JOIN [{'a': 32}] AS `z\"a\"r`",
    input = "SeLeCT foo.* from (SELECT * FROM bar) `f\"o\"o` LEFT JOIN [{'a': 32}] `z\"a\"r`"
);
query_printer_test!(
    select_from_left_outer_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT OUTER JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_right_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo RIGHT JOIN [{'a': 32}] AS zar",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo RIGHT JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_right_outer_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo RIGHT JOIN [{'a': 32}] AS zar",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo RIGHT OUTER JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_inner_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo INNER JOIN [{'a': 32}] AS zar",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo INNER JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_cross_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo CROSS JOIN [{'a': 32}] AS zar",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_explicit_cross_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo CROSS JOIN [{'a': 32}] AS zar",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo CROSS JOIN [{'a': 32}] zar"
);
query_printer_test!(
    select_from_left_join_with_condition,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar ON true",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar On tRuE"
);
query_printer_test!(
    select_from_three_way_join,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar CROSS JOIN foo AS foo",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar JOIN foo AS foo"
);
query_printer_test!(
    select_from_three_way_join_with_condition,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar CROSS JOIN foo AS foo ON 42 = 43",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar JOIN foo AS foo ON 42 = 43"
);
query_printer_test!(
    select_from_three_way_join_with_two_conditions,
    expected = "SELECT foo.* FROM (SELECT * FROM bar) AS foo LEFT JOIN [{'a': 32}] AS zar ON true CROSS JOIN foo AS foo ON 42 = 43",
    input = "SeLeCT foo.* from (SELECT * FROM bar) foo LEFT JOIN [{'a': 32}] zar oN TRUE JOIN foo AS foo ON 42 = 43"
);
query_printer_test!(
    flatten_collection,
    expected = "SELECT * FROM FLATTEN(foo)",
    input = "SeLeCT * from FLATTEN(foo)"
);
query_printer_test!(
    flatten_collection_with_separator,
    expected = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '%')",
    input = "SeLeCT * from FLATTEN(foo with separator => '%')"
);
query_printer_test!(
    flatten_collection_with_depth,
    expected = "SELECT * FROM FLATTEN(foo WITH DEPTH => 1)",
    input = "SeLeCT * from FLATTEN(foo with depth => 1)"
);
query_printer_test!(
    flatten_comma_join_datasource_with_option,
    expected = "SELECT * FROM FLATTEN(foo CROSS JOIN bar WITH SEPARATOR => '%')",
    input = "SELECT * FRoM FLATTEN(foo, bar with separator => '%')"
);
query_printer_test!(
    flatten_collection_with_multiple_options_preserves_order_and_duplicates,
    expected =
        "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '%', DEPTH => 1, SEPARATOR => ':', DEPTH => 2)",
    input =
        "SeLeCT * from FLATTEN(foo with separator => '%', depth => 1, separator => ':', depth => 2)"
);
query_printer_test!(
    flatten_separator_contains_string_delimiter,
    expected = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '''')",
    input = "SeLeCT * from FLATTEN(foo with separator => '''')"
);
query_printer_test!(
    flatten_collection_duplicate_separator_should_be_preserved,
    expected = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_', SEPARATOR => '-')",
    input = "SeLeCT * from FLATTEN(foo with separator => '_', separator => '-')"
);
query_printer_test!(
    flatten_collection_explicit_default_separator_should_be_preserved,
    expected = "SELECT * FROM FLATTEN(foo WITH SEPARATOR => '_')",
    input = "SeLeCT * from FLATTEN(foo with separator => '_')"
);

query_printer_test!(
    unwind_collection,
    expected = "SELECT * FROM UNWIND(foo)",
    input = "SELECT * FROM UNWIND(foo)"
);

query_printer_test!(
    unwind_collection_with_path,
    expected = "SELECT * FROM UNWIND(foo WITH PATH => arr)",
    input = "SELECT * FROM UNWIND(foo with path => arr)"
);

query_printer_test!(
    unwind_collection_with_delimited_path,
    expected = "SELECT * FROM UNWIND(foo WITH PATH => `x.arr`)",
    input = "SELECT * FROM UNWIND(foo with path => `x.arr`)"
);

query_printer_test!(
    unwind_collection_with_index,
    expected = "SELECT * FROM UNWIND(foo WITH INDEX => i)",
    input = "SELECT * FROM UNWIND(foo with index => i)"
);

query_printer_test!(
    unwind_collection_with_delimited_index,
    expected = "SELECT * FROM UNWIND(foo WITH INDEX => `x.i`)",
    input = "SELECT * FROM UNWIND(foo with index => `x.i`)"
);

query_printer_test!(
    unwind_collection_with_outer_true,
    expected = "SELECT * FROM UNWIND(foo WITH OUTER => true)",
    input = "SELECT * FROM UNWIND(foo with outer => true)"
);

query_printer_test!(
    unwind_collection_explicit_outer_false_should_be_preserved,
    expected = "SELECT * FROM UNWIND(foo WITH OUTER => false)",
    input = "SELECT * FROM UNWIND(foo WITH outer => false)"
);

query_printer_test!(
    unwind_collection_with_multiple_options_preserves_order_and_duplicates,
    expected = "SELECT * FROM UNWIND(foo WITH PATH => arr, OUTER => false, OUTER => true, INDEX => i, PATH => a)",
    input = "SELECT * FROM UNWIND(foo with path => arr, outer => false, outer => true, index => i, path => a)"
);

query_printer_test!(
    select_where_1,
    expected = "SELECT * FROM foo AS bar WHERE 1",
    input = "SELECT * FROM foo bar WHERE 1"
);
query_printer_test!(
    select_where_eq,
    expected = "SELECT * FROM foo AS bar WHERE 1 = 2",
    input = "SELECT * FROM foo bar WHERE 1 = 2"
);

query_printer_test!(
    select_having_1,
    expected = "SELECT * FROM foo AS bar HAVING 1",
    input = "SELECT * FROM foo bar HAVING 1"
);
query_printer_test!(
    select_having_eq,
    expected = "SELECT * FROM foo AS bar HAVING 1 = 2",
    input = "SELECT * FROM foo bar HAVING 1 = 2"
);

query_printer_test!(
    select_where_1_having_1,
    expected = "SELECT * FROM foo AS bar WHERE 1 HAVING 1",
    input = "SELECT * FROM foo bar WHERE 1 HAVING 1"
);
query_printer_test!(
    select_where_eq_having_eq,
    expected = "SELECT * FROM foo AS bar WHERE 1 = 2 HAVING 1 = 2",
    input = "SELECT * FROM foo bar WHERE 1 = 2 HAVING 1 = 2"
);

query_printer_test!(
    select_group_by_null,
    expected = "SELECT * FROM foo GROUP BY NULL",
    input = "SELECT * FROM foo GROUP BY NULL"
);
query_printer_test!(
    select_group_by_key,
    expected = "SELECT * FROM foo GROUP BY a",
    input = "SELECT * FROM foo GROUP BY a"
);
query_printer_test!(
    select_group_by_keys,
    expected = "SELECT * FROM foo GROUP BY a, b",
    input = "SELECT * FROM foo GROUP BY a, b"
);
query_printer_test!(
    select_group_by_keys_with_alias,
    expected = "SELECT * FROM foo GROUP BY a, b AS b",
    input = "SELECT * FROM foo GROUP BY a, b b"
);
query_printer_test!(
    select_group_by_keys_with_aliases,
    expected = "SELECT * FROM foo GROUP BY a AS c, b AS b",
    input = "SELECT * FROM foo GROUP BY a c, b b"
);
query_printer_test!(
    select_group_by_one_aggregates,
    expected = "SELECT * FROM foo GROUP BY a AS c, b AS b AGGREGATE COUNT(*) AS agg1",
    input = "SELECT * FROM foo GROUP BY a c, b b AGGREGATE COUNT(*) agg1"
);
query_printer_test!(
    select_group_by_two_aggregates,
    expected = "SELECT agg1, agg2 FROM foo GROUP BY a AS c, b AS b AGGREGATE COUNT(*) AS agg1, SUM(foo) AS agg2",
    input = "SELECT agg1, agg2 FROM foo GROUP BY a c, b b AGGREGATE COUNT(*) agg1, SUM(foo) agg2"
);

query_printer_test!(
    select_order_by_1_asc,
    expected = "SELECT * FROM foo ORDER BY 1 ASC",
    input = "selECT * FROM foo ORDER BY 1 ASC"
);
query_printer_test!(
    select_order_by_1_desc,
    expected = "SELECT * FROM foo ORDER BY 1 DESC",
    input = "selECT * FROM foo ORDER BY 1 DESC"
);
query_printer_test!(
    select_order_by_a_asc,
    expected = "SELECT * FROM foo ORDER BY a ASC",
    input = "selECT * FROM foo ORDER BY a ASC"
);
query_printer_test!(
    select_order_by_a_desc,
    expected = "SELECT * FROM foo ORDER BY a DESC",
    input = "selECT * FROM foo ORDER BY a DESC"
);
query_printer_test!(
    select_order_by_1_asc_2_asc,
    expected = "SELECT * FROM foo ORDER BY 1 ASC, 2 ASC",
    input = "selECT * FROM foo ORDER BY 1 ASC, 2"
);
query_printer_test!(
    select_order_by_1_desc_2_desc,
    expected = "SELECT * FROM foo ORDER BY 1 DESC, 2 DESC",
    input = "selECT * FROM foo ORDER BY 1 DESC, 2 DESC"
);
query_printer_test!(
    select_order_by_a_asc_b_asc,
    expected = "SELECT * FROM foo ORDER BY a ASC, b ASC",
    input = "selECT * FROM foo ORDER BY a ASC, b ASC"
);
query_printer_test!(
    select_order_by_a_desc_b_desc,
    expected = "SELECT * FROM foo ORDER BY a DESC, b DESC",
    input = "selECT * FROM foo ORDER BY a DESC, b DESC"
);

query_printer_test!(
    select_limit_1,
    expected = "SELECT * FROM foo AS bar LIMIT 1",
    input = "SELECT * FROM foo bar LIMIT 1"
);
query_printer_test!(
    select_offset_1,
    expected = "SELECT * FROM foo AS bar OFFSET 1",
    input = "SELECT * FROM foo bar OFFSET 1"
);
query_printer_test!(
    select_limit_1_offset_1,
    expected = "SELECT * FROM foo AS bar LIMIT 1 OFFSET 1",
    input = "SELECT * FROM foo bar LIMIT 1 OFFSET 1"
);
query_printer_test!(
    select_offset_1_limit_1,
    expected = "SELECT * FROM foo AS bar LIMIT 1 OFFSET 1",
    input = "SELECT * FROM foo bar OFFSET 1 LIMIT 1"
);

query_printer_test!(
    select_all_clauses,
    expected = "SELECT * FROM foo AS bar WHERE 1 GROUP BY a, b AGGREGATE COUNT(*) AS agg1, SUM(a) AS agg2 HAVING agg1 < agg2 ORDER BY agg1 ASC LIMIT 100 OFFSET 10",
    input = "SELECT * FROM foo bar WHERE 1 GROUP BY a, b AGGREGATE COUNT(*) AS agg1, SUM(a) as agg2 HAVING agg1 < agg2 ORDER BY agg1 LIMIT 100 OFFSET 10"
);

query_printer_test!(
    doc_with_single_quote_in_key,
    expected = "SELECT {'''': 'single quote'}",
    input = "SELECT {'''': 'single quote'}"
);

expression_printer_test!(
    identifier_non_latin_first_char,
    expected = "`做`",
    input = "`做`"
);
expression_printer_test!(
    identifier_non_latin_subsequent_chars,
    expected = "`_做`",
    input = "`_做`"
);
expression_printer_test!(
    delimited_identifier_containing_backtick,
    expected = "````",
    input = "````"
);
expression_printer_test!(
    normal_identifiers,
    expected = "(foo - bar) / car",
    input = "(foo - bar) / car"
);
expression_printer_test!(
    special_identifiers,
    expected = "`fo.o` - bar / `$car`",
    input = "`fo.o` - (bar / `$car`)"
);
expression_printer_test!(starts_with_number, expected = "`1foo`", input = "`1foo`");
expression_printer_test!(starts_with_underscore, expected = "_foo", input = "`_foo`");
expression_printer_test!(
    regular_identifier_containing_number,
    expected = "foo1",
    input = "foo1"
);
expression_printer_test!(empty_identifier, expected = "``", input = "``");
expression_printer_test!(
    is_missing,
    expected = "true AND x IS MISSING",
    input = "true AND (x IS MISSING)"
);
expression_printer_test!(
    is_type,
    expected = "true AND x IS INT",
    input = "true AND (x IS int)"
);

expression_printer_test!(
    like_simple,
    expected = "true AND x LIKE '%hello%'",
    input = "true AND (x LIKE '%hello%')"
);
expression_printer_test!(
    like_escape,
    expected = "true AND x LIKE '%hello%' ESCAPE '@'",
    input = "true AND (x LIKE '%hello%' ESCAPE '@')"
);

expression_printer_test!(
    like_escape_with_single_quote,
    expected = "true AND x LIKE '%hello%' ESCAPE ''''",
    input = "true AND (x LIKE '%hello%' ESCAPE '''')"
);

expression_printer_test!(
    type_assert,
    expected = "true AND x::!INT",
    input = "true AND (x::!INT)"
);
expression_printer_test!(
    type_bson_date,
    expected = "x::!BSON_DATE",
    input = "x::!TIMESTAMP"
);
expression_printer_test!(
    type_bson_timestamp,
    expected = "x::!BSON_TIMESTAMP",
    input = "x::!BSON_TIMESTAMP"
);

expression_printer_test!(tuple, expected = "(a, b, c)", input = "(a, b, c)");

expression_printer_test!(
    cast_sigil_array,
    expected = "CAST(4 + foo AS ARRAY)",
    input = "(4 + foo)::array"
);
expression_printer_test!(
    cast_sigil_bindata,
    expected = "CAST(4 + foo AS BINDATA)",
    input = "(4 + foo)::bindata"
);
expression_printer_test!(
    cast_sigil_boolean,
    expected = "CAST(4 + foo AS BOOL)",
    input = "(4 + foo)::boolean"
);
expression_printer_test!(
    cast_sigil_bson_date,
    expected = "CAST(4 + foo AS BSON_DATE)",
    input = "(4 + foo)::bson_date"
);
expression_printer_test!(
    cast_sigil_dbpointer,
    expected = "CAST(4 + foo AS DBPOINTER)",
    input = "(4 + foo)::dbpointer"
);
expression_printer_test!(
    cast_sigil_decimal_prec,
    expected = "CAST(4 + foo AS DECIMAL)",
    input = "(4 + foo)::decimal(45)"
);
expression_printer_test!(
    cast_sigil_decimal,
    expected = "CAST(4 + foo AS DECIMAL)",
    input = "(4 + foo)::decimal"
);
expression_printer_test!(
    cast_sigil_document,
    expected = "CAST(4 + foo AS DOCUMENT)",
    input = "(4 + foo)::document"
);
expression_printer_test!(
    cast_sigil_float_prec,
    expected = "CAST(4 + foo AS DOUBLE)",
    input = "(4 + foo)::float(10)"
);
expression_printer_test!(
    cast_sigil_double,
    expected = "CAST(4 + foo AS DOUBLE)",
    input = "(4 + foo)::double"
);
expression_printer_test!(
    cast_sigil_int,
    expected = "CAST(4 + foo AS INT)",
    input = "(4 + foo)::int"
);
expression_printer_test!(
    cast_sigil_long,
    expected = "CAST(4 + foo AS LONG)",
    input = "(4 + foo)::long"
);
expression_printer_test!(
    cast_sigil_javascript,
    expected = "CAST(4 + foo AS JAVASCRIPT)",
    input = "(4 + foo)::javascript"
);
expression_printer_test!(
    cast_sigil_javascriptwithscope,
    expected = "CAST(4 + foo AS JAVASCRIPTWITHSCOPE)",
    input = "(4 + foo)::javascriptwithscope"
);
expression_printer_test!(
    cast_sigil_maxkey,
    expected = "CAST(4 + foo AS MAXKEY)",
    input = "(4 + foo)::maxkey"
);
expression_printer_test!(
    cast_sigil_minkey,
    expected = "CAST(4 + foo AS MINKEY)",
    input = "(4 + foo)::minkey"
);
expression_printer_test!(
    cast_sigil_null,
    expected = "CAST(4 + foo AS NULL)",
    input = "(4 + foo)::null"
);
expression_printer_test!(
    cast_sigil_objectid,
    expected = "CAST(4 + foo AS OBJECTID)",
    input = "(4 + foo)::objectid"
);
expression_printer_test!(
    cast_sigil_regularexpression,
    expected = "CAST(4 + foo AS REGEX)",
    input = "(4 + foo)::regex"
);
expression_printer_test!(
    cast_sigil_string_prec,
    expected = "CAST(4 + foo AS STRING)",
    input = "(4 + foo)::varchar(5)"
);
expression_printer_test!(
    cast_sigil_string,
    expected = "CAST(4 + foo AS STRING)",
    input = "(4 + foo)::string"
);
expression_printer_test!(
    cast_sigil_symbol,
    expected = "CAST(4 + foo AS SYMBOL)",
    input = "(4 + foo)::symbol"
);
expression_printer_test!(
    cast_sigil_bson_timestamp,
    expected = "CAST(4 + foo AS BSON_TIMESTAMP)",
    input = "(4 + foo)::bson_timestamp"
);
expression_printer_test!(
    cast_sigil_undefined,
    expected = "CAST(4 + foo AS UNDEFINED)",
    input = "(4 + foo)::undefined"
);

expression_printer_test!(
    cast_int,
    expected = "CAST(x AS INT)",
    input = "CAST(x as int)"
);
expression_printer_test!(
    cast_int_on_null,
    expected = "CAST(x AS INT, 3 + 4 ON NULL)",
    input = "CAST(x as int, 3+4 ON NULL)"
);
expression_printer_test!(
    cast_int_on_error,
    expected = "CAST(x AS INT, 3 + 4 ON ERROR)",
    input = "CAST(x as int, 3+4 ON ERROR)"
);
expression_printer_test!(
    cast_int_on_null_on_error,
    expected = "CAST(x AS INT, 'bar' ON NULL, 'foo' ON ERROR)",
    input = "CAST(x as int, 'bar' ON NULL, 'foo' ON ERROR)"
);

expression_printer_test!(literal_null, expected = "NULL", input = "nUlL");
expression_printer_test!(literal_true, expected = "true", input = "tRuE");
expression_printer_test!(literal_false, expected = "false", input = "fALse");
expression_printer_test!(literal_string, expected = "'hellO'", input = "'hellO'");
expression_printer_test!(
    literal_string_with_single_quote,
    expected = "'''hellO'",
    input = "'''hellO'"
);
expression_printer_test!(literal_int, expected = "1", input = "1");
expression_printer_test!(literal_long, expected = "8000000000", input = "8000000000");
expression_printer_test!(literal_double, expected = "81111.1", input = "8.11111e4");
expression_printer_test!(
    literal_double_no_fraction,
    expected = "8000.0",
    input = "8e3"
);

expression_printer_test!(unary_not, expected = "NOT true", input = "not (true)");
expression_printer_test!(
    unary_not_parens,
    expected = "NOT (true OR false)",
    input = "not (true OR false)"
);
expression_printer_test!(unary_neg, expected = "- 3", input = "- 3");
expression_printer_test!(unary_neg_parens, expected = "- (3 + 4)", input = "-(3+4)");
expression_printer_test!(unary_pos, expected = "+ 3", input = "+ (3)");
expression_printer_test!(unary_pos_parens, expected = "+ (3 + 4)", input = "+(3+4)");
expression_printer_test!(
    unary_negative_sub_function,
    expected = "4 - + SUM(bar)",
    input = "4 - + SUM(bar)"
);
expression_printer_test!(
    unary_positive_sub_function,
    expected = "4 - - SUM(bar)",
    input = "4 - - SUM(bar)"
);

expression_printer_test!(
    binary_add_and_mul_no_parens,
    expected = "3 + 4 * 6",
    input = "3 + 4 * 6"
);
expression_printer_test!(
    binary_add_and_mul_unncessary_parens,
    expected = "3 + 4 * 6",
    input = "3 + (4 * 6)"
);
expression_printer_test!(
    binary_add_and_mul_parens,
    expected = "(3 + 4) * 6",
    input = "(3 + 4) * 6"
);
expression_printer_test!(
    binary_add_and_mul_extra_parens,
    expected = "(3 + 4) * 6",
    input = "((3 + 4)) * 6"
);

expression_printer_test!(
    binary_lt_and_and_parens,
    expected = "(true AND 3) < 4",
    input = "(true AND 3) < 4"
);
expression_printer_test!(
    binary_lt_and_and_no_parens,
    expected = "true AND 3 < 4",
    input = "true AND 3 < 4"
);

expression_printer_test!(
    binary_and_or_no_parens,
    expected = "true AND false OR false",
    input = "true AND false OR false"
);
expression_printer_test!(
    binary_and_or_parens,
    expected = "true AND (false OR false)",
    input = "true AND (false OR false)"
);

expression_printer_test!(
    binary_or,
    expected = "hello OR world",
    input = "hello OR world"
);
expression_printer_test!(
    binary_concat,
    expected = "hello || world",
    input = "hello || world"
);
expression_printer_test!(
    binary_lt,
    expected = "hello < world",
    input = "hello < world"
);
expression_printer_test!(
    binary_lte,
    expected = "hello <= world",
    input = "hello <= world"
);
expression_printer_test!(
    binary_gt,
    expected = "hello > world",
    input = "hello > world"
);
expression_printer_test!(
    binary_gte,
    expected = "hello >= world",
    input = "hello >= world"
);
expression_printer_test!(
    binary_eq,
    expected = "hello = world",
    input = "hello = world"
);
expression_printer_test!(
    binary_neq,
    expected = "hello <> world",
    input = "hello <> world"
);
expression_printer_test!(
    binary_neq2,
    expected = "hello <> world",
    input = "hello != world"
);
expression_printer_test!(
    binary_add,
    expected = "hello + world",
    input = "hello + world"
);
expression_printer_test!(
    binary_sub,
    expected = "hello - world",
    input = "hello - world"
);
expression_printer_test!(
    binary_mul,
    expected = "hello * world",
    input = "hello * world"
);
expression_printer_test!(
    binary_div,
    expected = "hello / world",
    input = "hello / world"
);
expression_printer_test!(
    binary_in,
    expected = "hello IN world",
    input = "hello in world"
);
expression_printer_test!(
    binary_not_in,
    expected = "hello NOT IN world",
    input = "hello nOt in world"
);

expression_printer_test!(
    case_basic,
    expected = "CASE WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END",
    input = "CASE WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END"
);
expression_printer_test!(
    case_basic_with_expr,
    expected = "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END",
    input = "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false END"
);
expression_printer_test!(
    case_basic_with_expr_and_else,
    expected = "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false ELSE foo + bar END",
    input = "CASE foo * 3 WHEN x + 3 = y THEN true WHEN x + 4 = y THEN false ELSE foo+bar END"
);
expression_printer_test!(
    case_as_expr,
    expected = "true OR CASE foo * 3 WHEN x + 3 = y THEN true ELSE false END",
    input = "true OR (CASE foo * 3 WHEN x + 3 = y THEN true ELSE false END)"
);

expression_printer_test!(
    function_basic,
    expected = "COALESCE(bar, foo + bar, hello * 3)",
    input = "COALESCE(bar, foo+bar, hello*3)"
);
expression_printer_test!(
    function_distinct,
    expected = "AVG(DISTINCT x)",
    input = "AvG(DiSTINCT x)"
);
expression_printer_test!(function_all, expected = "AVG(ALL x)", input = "AvG(AlL x)");
expression_printer_test!(
    function_sub,
    expected = "4 - SUM(bar)",
    input = "4 - SUM(bar)"
);
expression_printer_test!(
    function_extract_day,
    expected = "EXTRACT(DAY FROM bar)",
    input = "extract(DAY FROM bar)"
);
expression_printer_test!(
    function_extract_hour,
    expected = "EXTRACT(HOUR FROM bar)",
    input = "extract(HOUR FROM bar)"
);
expression_printer_test!(
    function_extract_minute,
    expected = "EXTRACT(MINUTE FROM bar)",
    input = "extract(MINUTE FROM bar)"
);
expression_printer_test!(
    function_extract_month,
    expected = "EXTRACT(MONTH FROM bar)",
    input = "extract(MONTH FROM bar)"
);
expression_printer_test!(
    function_extract_second,
    expected = "EXTRACT(SECOND FROM bar)",
    input = "extract(SECOND FROM bar)"
);
expression_printer_test!(
    function_extract_year,
    expected = "EXTRACT(YEAR FROM bar)",
    input = "extract(YEAR FROM bar)"
);

expression_printer_test!(
    position_ident_in_ident,
    expected = "POSITION(x IN y)",
    input = "position(x in y)"
);
expression_printer_test!(
    position_string_in_ident,
    expected = "POSITION('x' IN y)",
    input = "position('x' in y)"
);
expression_printer_test!(
    position_ident_in_string,
    expected = "POSITION(x IN 'y')",
    input = "position(x in 'y')"
);
expression_printer_test!(
    position_string_in_string,
    expected = "POSITION('x' IN 'y')",
    input = "position('x' in 'y')"
);
expression_printer_test!(
    position_in_is_expr,
    expected = "POSITION('x' IN ('y' IS DOCUMENT))",
    input = "position('x' in ('y' is document))"
);

expression_printer_test!(
    function_upper,
    expected = "UPPER('hello')",
    input = "UPPER('hello')"
);
expression_printer_test!(
    function_lower,
    expected = "LOWER('hello')",
    input = "LOWER('hello')"
);

expression_printer_test!(
    function_trim_basic,
    expected = "TRIM(BOTH ' ' FROM hello)",
    input = "TRIM(hello)"
);
expression_printer_test!(
    function_trim_basic_chars,
    expected = "TRIM(BOTH 'asdf' FROM hello)",
    input = "TRIM('asdf' FROM hello)"
);
expression_printer_test!(
    function_trim_explicit_both,
    expected = "TRIM(BOTH ' ' FROM hello)",
    input = "TRIM(BOTH FROM hello)"
);
expression_printer_test!(
    function_trim_leading,
    expected = "TRIM(LEADING ' ' FROM hello)",
    input = "TRIM(LEADING FROM hello)"
);
expression_printer_test!(
    function_trim_trailing,
    expected = "TRIM(TRAILING ' ' FROM hello)",
    input = "TRIM(TRAILING FROM hello)"
);
expression_printer_test!(
    function_trim_explicit_both_chars,
    expected = "TRIM(BOTH 'asdf' FROM hello)",
    input = "TRIM(BOTH 'asdf' FROM hello)"
);
expression_printer_test!(
    function_trim_leading_chars,
    expected = "TRIM(LEADING 'asdf' FROM hello)",
    input = "TRIM(LEADING 'asdf' FROM hello)"
);
expression_printer_test!(
    function_trim_trailing_chars,
    expected = "TRIM(TRAILING 'asdf' FROM hello)",
    input = "TRIM(TRAILING 'asdf' FROM hello)"
);
expression_printer_test!(
    function_split,
    expected = "SPLIT(str, 'delim', 3)",
    input = "SPLIT(str, 'delim', 3)"
);

expression_printer_test!(function_abs, expected = "ABS(10)", input = "ABS(10)");
expression_printer_test!(function_ceil, expected = "CEIL(1.5)", input = "CEIL(1.5)");
expression_printer_test!(
    function_degrees,
    expected = "DEGREES(1)",
    input = "DEGREES(1)"
);
expression_printer_test!(
    function_floor,
    expected = "FLOOR(1.5)",
    input = "FLOOR(1.5)"
);
expression_printer_test!(
    function_log,
    expected = "LOG(100, 10)",
    input = "LOG(100, 10)"
);
expression_printer_test!(function_mod, expected = "MOD(80, 7)", input = "MOD(80, 7)");
expression_printer_test!(function_pow, expected = "POW(5, 2)", input = "POW(5, 2)");
expression_printer_test!(
    function_round,
    expected = "ROUND(100, 1)",
    input = "ROUND(100, 1)"
);
expression_printer_test!(function_cos, expected = "COS(1)", input = "COS(1)");
expression_printer_test!(function_sin, expected = "SIN(1)", input = "SIN(1)");
expression_printer_test!(function_tan, expected = "TAN(1)", input = "TAN(1)");
expression_printer_test!(
    function_radians,
    expected = "RADIANS(1)",
    input = "RADIANS(1)"
);
expression_printer_test!(function_sqrt, expected = "SQRT(1)", input = "SQRT(1)");

expression_printer_test!(
    access_int,
    expected = "[1, 2, 3][1]",
    input = "[1, 2, 3] [1]"
);
expression_printer_test!(
    access_string,
    expected = "{'a': 3}['a']",
    input = "{'a': 3} ['a']"
);
expression_printer_test!(
    access_identifier_string,
    expected = "a['a']",
    input = "a ['a']"
);
expression_printer_test!(
    access_string_with_parens,
    expected = "(3 + 4)['a']",
    input = "(3+4) ['a']"
);
expression_printer_test!(access_access, expected = "a[1][2]", input = "a[1][2]");

expression_printer_test!(subpath, expected = "{'a': 3}.a", input = "{'a': 3} .a");
expression_printer_test!(subpath_identifier, expected = "a.a", input = "a .a");
expression_printer_test!(
    subpath_identifier_needs_delimiter,
    expected = "a.`a\"b`",
    input = "a.`a\"b`"
);
expression_printer_test!(
    subpath_with_parens,
    expected = "(3 + 4).a",
    input = "(3+4) .a"
);

expression_printer_test!(subpath_access, expected = "a[1].b", input = "a[1].b");
expression_printer_test!(access_subpath, expected = "a[1].b", input = "a[1].b");
expression_printer_test!(subpath_subpath, expected = "a.b.c", input = "a.b.c");

expression_printer_test!(
    sub_query,
    expected = "1 + (SELECT * UNION SELECT *)",
    input = "1+(select * unioN SElect *)"
);
expression_printer_test!(
    exists,
    expected = "true AND EXISTS(SELECT * UNION SELECT *)",
    input = "true and ExisTS(select * unioN SElect *)"
);
expression_printer_test!(
    subquery_eq_any,
    expected = "x = ANY(SELECT * UNION SELECT *)",
    input = "x=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_eq_all,
    expected = "x = ALL(SELECT * UNION SELECT *)",
    input = "x=ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lt_any,
    expected = "x < ANY(SELECT * UNION SELECT *)",
    input = "x<ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lt_all,
    expected = "x < ALL(SELECT * UNION SELECT *)",
    input = "x<ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lte_any,
    expected = "x <= ANY(SELECT * UNION SELECT *)",
    input = "x<=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_lte_all,
    expected = "x <= ALL(SELECT * UNION SELECT *)",
    input = "x<=ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gt_any,
    expected = "x > ANY(SELECT * UNION SELECT *)",
    input = "x>ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gt_all,
    expected = "x > ALL(SELECT * UNION SELECT *)",
    input = "x>ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gte_any,
    expected = "x >= ANY(SELECT * UNION SELECT *)",
    input = "x>=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_gte_all,
    expected = "x >= ALL(SELECT * UNION SELECT *)",
    input = "x>=ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq_any,
    expected = "x <> ANY(SELECT * UNION SELECT *)",
    input = "x<>ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq_all,
    expected = "x <> ALL(SELECT * UNION SELECT *)",
    input = "x<>ALL(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq2_any,
    expected = "x <> ANY(SELECT * UNION SELECT *)",
    input = "x!=ANY(SELECT * UNION SELECT *)"
);
expression_printer_test!(
    subquery_neq2_all,
    expected = "x <> ALL(SELECT * UNION SELECT *)",
    input = "x!=ALL(SELECT * UNION SELECT *)"
);

// precedence tests
expression_printer_test!(
    unary_is_lower_prec_than_assert,
    expected = "(- x)::!INT",
    input = "(- x)::!INT"
);
expression_printer_test!(
    is_is_lower_prec_than_between,
    expected = "x BETWEEN y AND (z IS INT)",
    input = "x BETWEEN y AND (z IS INT)"
);
expression_printer_test!(
    in_is_lower_prec_than_between,
    expected = "x BETWEEN (y IN z) AND p",
    input = "x BETWEEN (y IN z) AND p"
);
expression_printer_test!(
    plus_is_higher_prec_than_between,
    expected = "x BETWEEN y + z AND p",
    input = "x BETWEEN (y + z) AND p"
);
expression_printer_test!(
    between_is_higher_prec_than_is,
    expected = "x BETWEEN y AND z IS INT",
    input = "(x BETWEEN y AND z) IS INT"
);
expression_printer_test!(
    is_is_lower_prec_than_like,
    expected = "x LIKE y IS BOOL",
    input = "(x LIKE y) IS BOOL"
);
expression_printer_test!(
    is_is_equal_prec_to_is,
    expected = "(x IS BOOL) IS BOOL",
    input = "(x IS BOOL) IS BOOL"
);
expression_printer_test!(
    like_is_lower_prec_than_concat,
    expected = "x LIKE 'bar' || 'foo'",
    input = "x LIKE ('bar' || 'foo')"
);
expression_printer_test!(
    subquery_comparison_is_higher_prec_than_and,
    expected = "true AND x = (SELECT * FROM foo)",
    input = "TRuE AND (x = (SELECT * FrOM foo))"
);
expression_printer_test!(
    subquery_comparison_is_lower_prec_than_add,
    expected = "42 + (x = (SELECT * FROM foo))",
    input = "42 + (x = (SELECT * FrOM foo))"
);
expression_printer_test!(
    assoc_bin_op_prints_without_explicit_parens,
    expected = "3 + 4 + 5",
    input = "(3 + 4) + 5"
);
expression_printer_test!(
    same_tier_bin_op_prints_without_explicit_parens,
    expected = "3 + 4 - 5",
    input = "(3 + 4) - 5"
);
expression_printer_test!(
    different_tier_bin_op_prints_with_explicit_parens,
    expected = "(3 + 4) * 5",
    input = "(3 + 4) * 5"
);
expression_printer_test!(
    associative_like_expression_should_not_need_explicit_parens,
    expected = "'hello' LIKE 'hello' LIKE 'world'",
    input = "'hello' LIKE 'hello' LIKE 'world'"
);
expression_printer_test!(
    associative_like_pattern_needs_explicit_parens,
    expected = "'hello' LIKE ('hello' LIKE 'world')",
    input = "'hello' LIKE ('hello' LIKE 'world')"
);
expression_printer_test!(
    associative_like_expression_with_escape_should_not_need_explicit_parens,
    expected = "'hello' LIKE 'hello' ESCAPE '_' LIKE 'world' ESCAPE '_'",
    input = "'hello' LIKE 'hello' ESCAPE '_' LIKE 'world' ESCAPE '_'"
);
expression_printer_test!(
    associative_like_pattern_with_escape_needs_explicit_parens,
    expected = "'hello' LIKE ('hello' LIKE 'world' ESCAPE '_') ESCAPE '_'",
    input = "'hello' LIKE ('hello' LIKE 'world' ESCAPE '_') ESCAPE '_'"
);
