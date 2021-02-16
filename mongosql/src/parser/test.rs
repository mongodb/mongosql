use crate::parser;

macro_rules! should_parse {
    ($func_name:ident, $should_parse:expr, $input:expr) => {
        #[test]
        fn $func_name() {
            let res = parser::parse($input);
            let should_parse = $should_parse;
            if should_parse {
                res.expect("expected input to parse, but it failed");
            } else {
                assert!(res.is_err());
            }
        }
    };
}

should_parse!(select_star, true, "select *");
should_parse!(select_star_upper, true, "SELECT *");
should_parse!(select_mixed_case, true, "SeLeCt *");
should_parse!(select_a_star, true, "select a.*");
should_parse!(select_underscore_id, true, "select _id");
should_parse!(select_contains_underscore, true, "select a_b");
should_parse!(select_multiple, true, "select a,b,c");
should_parse!(select_multiple_combo, true, "select a,b,*");
should_parse!(select_multiple_star, true, "select *,*");
should_parse!(select_multiple_dot_star, true, "select a.*,b.*");
should_parse!(select_all_lower, true, "select all *");
should_parse!(select_all_upper, true, "select ALL *");
should_parse!(select_all_mixed_case, true, "select aLl *");
should_parse!(select_distinct_lower, true, "select distinct *");
should_parse!(select_distinct_upper, true, "select DISTINCT *");
should_parse!(select_distinct_mixed_case, true, "select DiSTinCt *");
should_parse!(select_value_lower, true, "SELECT value foo.*");
should_parse!(select_value_upper, true, "SELECT VALUE foo.*");
should_parse!(select_value_mixed_case, true, "SELECT vAlUe foo.*");
should_parse!(select_values_lower, true, "SELECT values foo.*, bar.*");
should_parse!(select_values_upper, true, "SELECT VALUES foo.*, bar.*");
should_parse!(select_values_mixed_case, true, "SELECT vAluES foo.*, bar.*");
should_parse!(select_alias_lower, true, "SELECT foo as f");
should_parse!(select_alias_upper, true, "SELECT foo AS f");
should_parse!(select_alias_mixed_case, true, "SELECT foo aS f");
should_parse!(select_alias_compound_column, true, "SELECT a.b as a");
should_parse!(
    select_alias_multiple_combined,
    true,
    "SELECT a, b AS c, a.c"
);
should_parse!(select_long_compound, true, "SELECT a.b.c.d");
should_parse!(select_letter_number_ident, true, "SELECT a9");

should_parse!(use_stmt, false, "use foo");
should_parse!(select_compound_star, false, "SELECT a.b.c.*");
should_parse!(select_numerical_ident_prefix, false, "SELECT 9a");
should_parse!(select_value_star, false, "SELECT VALUE *");
should_parse!(select_value_alias, false, "SELECT VALUE foo AS f");
should_parse!(select_dangling_alias, false, "SELECT a.b AS");
should_parse!(select_compound_alias, false, "SELECT a AS b.c");
