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

should_parse!(select_all_lower, true, "select");
should_parse!(select_all_upper, true, "SELECT");
should_parse!(select_mixed_case, true, "SeLeCt");
should_parse!(select_star, false, "SeLeCt *");
should_parse!(use_stmt, false, "use foo");
