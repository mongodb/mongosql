use crate::parser::ast::*;
use crate::parser::Parser;

macro_rules! should_parse {
    ($func_name:ident, $should_parse:expr, $input:expr) => {
        #[test]
        fn $func_name() {
            let res = Parser::new().parse_query($input);
            let should_parse = $should_parse;
            if should_parse {
                res.expect("expected input to parse, but it failed");
            } else {
                assert!(res.is_err());
            }
        }
    };
}

macro_rules! validate_query_ast {
    ($func_name:ident, $input:expr, $ast:expr) => {
        #[test]
        fn $func_name() {
            let p = Parser::new();
            assert_eq!(p.parse_query($input).unwrap(), $ast)
        }
    };
}

macro_rules! validate_expression_ast {
    ($func_name:ident, $input:expr, $ast:expr) => {
        #[test]
        fn $func_name() {
            let p = Parser::new();
            assert_eq!(p.parse_expression($input).unwrap(), $ast)
        }
    };
}

// Select tests
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
should_parse!(select_alias_no_as, true, "SELECT foo f");
should_parse!(select_alias_compound_column, true, "SELECT a.b as a");
should_parse!(
    select_alias_multiple_combined,
    true,
    "SELECT a, b AS c, a.c"
);
should_parse!(select_long_compound, true, "SELECT a.b.c.d");
should_parse!(select_letter_number_ident, true, "SELECT a9");
should_parse!(select_delimited_ident_quotes, true, r#"SELECT "foo""#);
should_parse!(select_delimited_ident_backticks, true, "SELECT `foo`");
should_parse!(select_delimited_quote_empty, true, r#"SELECT """#);
should_parse!(select_delimited_backtick_empty, true, "SELECT ``");
should_parse!(
    select_delimited_escaped_quote,
    true,
    r#"SELECT "fo""o""""""#
);
should_parse!(
    select_delimited_escaped_backtick,
    true,
    "SELECT `f``oo`````"
);

should_parse!(use_stmt, false, "use foo");
should_parse!(select_compound_star, false, "SELECT a.b.c.*");
should_parse!(select_numerical_ident_prefix, false, "SELECT 9ae");
should_parse!(select_value_star, false, "SELECT VALUE *");
should_parse!(select_value_alias, false, "SELECT VALUE foo AS f");
should_parse!(select_dangling_alias, false, "SELECT a.b AS");
should_parse!(select_compound_alias, false, "SELECT a AS b.c");

should_parse!(
    select_delimited_extra_quote_outer,
    false,
    r#"SELECT ""foo"""#
);
should_parse!(
    select_delimited_extra_backtick_outer,
    false,
    "SELECT ``foo``"
);
should_parse!(
    select_delimited_escaped_quote_odd,
    false,
    r#"SELECT "f"oo"""#
);
should_parse!(
    select_delimited_escaped_backtick_odd,
    false,
    "SELECT `foo````"
);
should_parse!(
    select_delimited_backslash_escape,
    false,
    r#"SELECT "fo\"\"o""#
);
should_parse!(select_unescaped_quotes_in_ident, false, r#"SELECT fo""o"#);
should_parse!(select_unescaped_backticks_in_ident, false, "SELECT fo``o");

validate_query_ast!(
    ident,
    "SELECT foo",
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                expression: Expression::Identifier(Identifier::Simple("foo".to_string())),
                alias: None
            })])
        },
        order_by_clause: None,
    })
);
validate_query_ast!(
    delimited_quote,
    r#"SELECT "foo""#,
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                expression: Expression::Identifier(Identifier::Simple("foo".to_string())),
                alias: None
            })])
        },
        order_by_clause: None,
    })
);
validate_query_ast!(
    delimited_backtick,
    "select `foo`",
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                expression: Expression::Identifier(Identifier::Simple("foo".to_string())),
                alias: None
            })])
        },
        order_by_clause: None,
    })
);
validate_query_ast!(
    delimited_escaped_backtick,
    "SELECT `fo``o`````",
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                expression: Expression::Identifier(Identifier::Simple("fo`o``".to_string())),
                alias: None
            })])
        },
        order_by_clause: None,
    })
);
validate_query_ast!(
    delimited_escaped_quote,
    r#"SELECT "fo""o""""""#,
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                expression: Expression::Identifier(Identifier::Simple(r#"fo"o"""#.to_string())),
                alias: None
            })])
        },
        order_by_clause: None,
    })
);
validate_query_ast!(
    backtick_delimiter_escaped_quote,
    r#"SELECT `fo""o`"#,
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                expression: Expression::Identifier(Identifier::Simple(r#"fo""o"#.to_string())),
                alias: None
            })])
        },
        order_by_clause: None,
    })
);
validate_query_ast!(
    quote_delimiter_escaped_backtick,
    r#"SELECT "fo``o""#,
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                expression: Expression::Identifier(Identifier::Simple("fo``o".to_string())),
                alias: None
            })])
        },
        order_by_clause: None,
    })
);

// Set query tests
should_parse!(select_union_simple, true, "SELECT a UNION SELECT b");
should_parse!(
    select_union_multiple,
    true,
    "SELECT a UNION SELECT b UNION SELECT c"
);
should_parse!(
    select_union_all_multiple,
    true,
    "SELECT a UNION ALL SELECT b UNION ALL SELECT c"
);

validate_query_ast!(
    union_is_left_associative,
    "select a union select b union all select c",
    Query::Set(SetQuery {
        left: Box::new(Query::Set(SetQuery {
            left: Box::new(Query::Select(SelectQuery {
                select_clause: SelectClause {
                    set_quantifier: SetQuantifier::All,
                    body: SelectBody::Standard(vec![SelectExpression::Aliased(
                        AliasedExpression {
                            expression: Expression::Identifier(Identifier::Simple("a".to_string())),
                            alias: None
                        }
                    )])
                },
                order_by_clause: None,
            })),
            op: SetOperator::Union,
            right: Box::new(Query::Select(SelectQuery {
                select_clause: SelectClause {
                    set_quantifier: SetQuantifier::All,
                    body: SelectBody::Standard(vec![SelectExpression::Aliased(
                        AliasedExpression {
                            expression: Expression::Identifier(Identifier::Simple("b".to_string())),
                            alias: None
                        }
                    )])
                },
                order_by_clause: None,
            }))
        })),
        op: SetOperator::UnionAll,
        right: Box::new(Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Aliased(AliasedExpression {
                    expression: Expression::Identifier(Identifier::Simple("c".to_string())),
                    alias: None
                })])
            },
            order_by_clause: None,
        }))
    })
);

// Operator tests
should_parse!(unary_pos, true, "select +a");
should_parse!(unary_neg, true, "select -a");
should_parse!(unary_not, true, "select NOT a");
should_parse!(binary_add, true, "select a+b+c+d+e");
should_parse!(binary_sub, true, "select a-b-c-d-e");
should_parse!(binary_mul, true, "select a*b*c*d*e");
should_parse!(binary_mul_add, true, "select a*b+c*d+e");
should_parse!(binary_div_add, true, "select a/b+c/d+e");
should_parse!(binary_div_mul, true, "select a/b*c");
should_parse!(binary_lt, true, "select a<b<c<d<e");
should_parse!(binary_lte, true, "select a<=b");
should_parse!(binary_gt, true, "select a>b");
should_parse!(binary_gte, true, "select a>=b");
should_parse!(binary_neq, true, "select a<>b");
should_parse!(binary_eq, true, "select a=b");
should_parse!(binary_string_concat, true, "select a || b");
should_parse!(binary_or, true, "select a OR b");
should_parse!(binary_and, true, "select a AND b");
should_parse!(binary_compare_and_add, true, "select b<a+c and b>d+e");
should_parse!(binary_compare_or_mul, true, "select b<a*c or b>d*e");
should_parse!(binary_lt_and_neq, true, "select a<b and c<>e");
should_parse!(between, true, "select a BETWEEN b AND c");
should_parse!(case, true, "select CASE WHEN a=b THEN a ELSE c END");
should_parse!(
    case_multiple_when_clauses,
    true,
    "select CASE WHEN a or b THEN a WHEN c=d THEN c ELSE e END"
);
should_parse!(
    case_multiple_exprs,
    true,
    "select CASE a WHEN a <> b THEN a WHEN c and d THEN c ELSE e END"
);

should_parse!(between_invalid_binary_op, false, "select a BETWEEN b + c");
should_parse!(
    case_non_bool_conditions,
    false,
    "select case a when a+b then a else c-d"
);

validate_expression_ast!(
    binary_sub_unary_neg_ast,
    "b--a",
    Expression::Binary(BinaryExpr {
        left: Box::new(Expression::Identifier(Identifier::Simple("b".to_string()))),
        op: BinaryOp::Sub,
        right: Box::new(Expression::Unary(UnaryExpr {
            op: UnaryOp::Neg,
            expr: Box::new(Expression::Identifier(Identifier::Simple("a".to_string())))
        }))
    })
);

validate_expression_ast!(
    binary_mul_add_ast,
    "c*a+b",
    Expression::Binary(BinaryExpr {
        left: Box::new(Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier(Identifier::Simple("c".to_string()))),
            op: BinaryOp::Mul,
            right: Box::new(Expression::Identifier(Identifier::Simple("a".to_string())))
        })),
        op: BinaryOp::Add,
        right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
    })
);

validate_expression_ast!(
    binary_add_concat_ast,
    "a+b||c",
    Expression::Binary(BinaryExpr {
        left: Box::new(Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
            op: BinaryOp::Add,
            right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
        })),
        op: BinaryOp::Concat,
        right: Box::new(Expression::Identifier(Identifier::Simple("c".to_string())))
    })
);

validate_expression_ast!(
    binary_concat_compare_ast,
    "c>a||b",
    Expression::Binary(BinaryExpr {
        left: Box::new(Expression::Identifier(Identifier::Simple("c".to_string()))),
        op: BinaryOp::Gt,
        right: Box::new(Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
            op: BinaryOp::Concat,
            right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
        }))
    })
);

validate_expression_ast!(
    binary_compare_and_ast,
    "a<b AND c",
    Expression::Binary(BinaryExpr {
        left: Box::new(Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
            op: BinaryOp::Lt,
            right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
        })),
        op: BinaryOp::And,
        right: Box::new(Expression::Identifier(Identifier::Simple("c".to_string())))
    })
);

validate_expression_ast!(
    binary_and_or_ast,
    "a AND b OR b",
    Expression::Binary(BinaryExpr {
        left: Box::new(Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
            op: BinaryOp::And,
            right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
        })),
        op: BinaryOp::Or,
        right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
    })
);

validate_expression_ast!(
    between_ast,
    "a between b and c",
    Expression::Between(BetweenExpr {
        expr: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
        min: Box::new(Expression::Identifier(Identifier::Simple("b".to_string()))),
        max: Box::new(Expression::Identifier(Identifier::Simple("c".to_string()))),
    })
);

validate_expression_ast!(
    not_between_ast,
    "a not between b and c",
    Expression::Unary(UnaryExpr {
        op: UnaryOp::Not,
        expr: Box::new(Expression::Between(BetweenExpr {
            expr: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
            min: Box::new(Expression::Identifier(Identifier::Simple("b".to_string()))),
            max: Box::new(Expression::Identifier(Identifier::Simple("c".to_string()))),
        }))
    })
);

validate_expression_ast!(
    case_multiple_when_branches_ast,
    "case when a=b then a when c=d then c else e end",
    Expression::Case(CaseExpr {
        expr: None,
        when_branch: vec![
            WhenBranch {
                when: Box::new(Expression::Binary(BinaryExpr {
                    left: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
                    op: BinaryOp::Eq,
                    right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
                })),
                then: Box::new(Expression::Identifier(Identifier::Simple("a".to_string())))
            },
            WhenBranch {
                when: Box::new(Expression::Binary(BinaryExpr {
                    left: Box::new(Expression::Identifier(Identifier::Simple("c".to_string()))),
                    op: BinaryOp::Eq,
                    right: Box::new(Expression::Identifier(Identifier::Simple("d".to_string())))
                })),
                then: Box::new(Expression::Identifier(Identifier::Simple("c".to_string())))
            }
        ],
        else_branch: Some(Box::new(Expression::Identifier(Identifier::Simple(
            "e".to_string()
        ))))
    })
);

validate_expression_ast!(
    case_multiple_exprs_ast,
    "case a when a=b then a else c end",
    Expression::Case(CaseExpr {
        expr: Some(Box::new(Expression::Identifier(Identifier::Simple(
            "a".to_string()
        )))),
        when_branch: vec![WhenBranch {
            when: Box::new(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier(Identifier::Simple("a".to_string()))),
                op: BinaryOp::Eq,
                right: Box::new(Expression::Identifier(Identifier::Simple("b".to_string())))
            })),
            then: Box::new(Expression::Identifier(Identifier::Simple("a".to_string())))
        }],
        else_branch: Some(Box::new(Expression::Identifier(Identifier::Simple(
            "c".to_string()
        ))))
    })
);

// Order by tests
should_parse!(order_by_simple, true, "select * order by a");
should_parse!(order_by_asc, true, "select * order by a ASC");
should_parse!(order_by_desc, true, "select * order by a DESC");
should_parse!(order_by_multiple, true, "select a, b, c order by a, b");
should_parse!(
    order_by_multiple_directions,
    true,
    "select * order by a DESC, b ASC, c"
);
should_parse!(order_by_positional_sort, true, "select a, b order by 1, 2");
should_parse!(
    order_by_positional_sort_with_star,
    true,
    "select * order by 1"
);

validate_query_ast!(
    order_by_default_direction,
    "select * order by a",
    Query::Select(SelectQuery {
        select_clause: SelectClause {
            set_quantifier: SetQuantifier::All,
            body: SelectBody::Standard(vec![SelectExpression::Star])
        },
        order_by_clause: Some(OrderByClause {
            sort_specs: vec![SortSpec {
                key: SortKey::Simple(Identifier::Simple("a".to_string())),
                direction: SortDirection::Asc
            }]
        })
    })
);

should_parse!(
    order_by_neg_positional_sort,
    false,
    "select a, b order by -1, 2"
);
should_parse!(
    order_by_positional_sort_too_big,
    false,
    "select a, b order by 9223372036854775808"
);

// Literals tests
should_parse!(null_literal, true, "select null");
should_parse!(null_literal_mixed_case, true, "select nULL");
should_parse!(unsigned_int_literal, true, "select 123");
should_parse!(neg_int_literal, true, "select -123");
should_parse!(pos_int_literal, true, "select +123");
should_parse!(int_literal_leading_zeros, true, "select 008");
should_parse!(convert_to_long, true, "select 2147483648");
should_parse!(string_literal, true, "select 'foo'");
should_parse!(string_literal_special_characters, true, "select 'αβγ'");
should_parse!(empty_string_literal, true, "select ''");
should_parse!(unsigned_double_literal, true, "select 0.5");
should_parse!(unsigned_double_literal_no_fraction, true, "select 1.");
should_parse!(unsigned_double_literal_no_whole_num, true, "select .6");
should_parse!(neg_double_literal, true, "select -4.089015");
should_parse!(pos_double_literal, true, "select +0.0");
should_parse!(double_literal_exponent_lowercase, true, "select 1e2");
should_parse!(double_literal_exponent_uppercase, true, "select 2E3");
should_parse!(double_literal_exponent_beg_fraction, true, "select 8.e+23");
should_parse!(double_literal_exponent_mid_fraction, true, "select 9.07e-2");
should_parse!(double_literal_exponent_no_whole_num, true, "select .2E3");
should_parse!(double_literal_exponent_signed, true, "select -7.2E3");
should_parse!(boolean_literal_true, true, "select true");
should_parse!(boolean_literal_false, true, "select false");
should_parse!(
    boolean_literal_binary,
    true,
    "select true AND false OR false"
);

should_parse!(string_literal_single_quote, false, "select '''");
should_parse!(double_exponent_no_exp, false, "select 1e");
should_parse!(long_too_big, false, "select 9223372036854775808");

validate_expression_ast!(
    string_escaped_quote_no_chars,
    "''''",
    Expression::Literal(Literal::String(r#"'"#.to_string()))
);

validate_expression_ast!(
    string_escaped_quote,
    "'foo''s'",
    Expression::Literal(Literal::String(r#"foo's"#.to_string()))
);

validate_expression_ast!(
    double_neg_no_decimal,
    "-2E+3",
    Expression::Unary(UnaryExpr {
        op: UnaryOp::Neg,
        expr: Box::new(Expression::Literal(Literal::Double(2000.0)))
    })
);

validate_expression_ast!(
    double_no_whole,
    ".2E-3",
    Expression::Literal(Literal::Double(0.0002))
);

validate_expression_ast!(
    double_no_frac_or_sign,
    "234.E6",
    Expression::Literal(Literal::Double(234000000.0))
);

validate_expression_ast!(
    double_all_components,
    "234.2E-3",
    Expression::Literal(Literal::Double(0.2342))
);

validate_expression_ast!(
    double_binary_add,
    "2E3 + 5.16E-8",
    Expression::Binary(BinaryExpr {
        left: Box::new(Expression::Literal(Literal::Double(2000.0))),
        op: BinaryOp::Add,
        right: Box::new(Expression::Literal(Literal::Double(0.0000000516)))
    })
);

// Parenthesized expressions tests
should_parse!(parens_multiple_binary_ops, true, "SELECT ((a+b)-(d/c))*7");
should_parse!(
    parens_case_expr,
    true,
    "select (CASE WHEN a=b THEN 1 ELSE 2 END)*4"
);

should_parse!(unbalanced_parens, false, "SELECT ((a+b)");
should_parse!(empty_parens, false, "SELECT ()");

// Scalar function tests
should_parse!(null_if, true, "select nullif(a, b)");
should_parse!(coalesce, true, "select coalesce(a, b, c, d)");
should_parse!(size, true, "select size(a)");
should_parse!(position, true, "select position('b' IN 'abc')");
should_parse!(char_length, true, "select char_length('foo')");
should_parse!(character_length, true, "select character_length('bar')");
should_parse!(octet_length, true, "select octet_length(a)");
should_parse!(bit_length, true, "select bit_length(a)");
should_parse!(
    extract_timezone_hour,
    true,
    "select extract(timezone_hour from a)"
);
should_parse!(
    extract_timezone_minute,
    true,
    "select extract(timezone_minute from a)"
);
should_parse!(extract_year, true, "select extract(year from a)");
should_parse!(extract_month, true, "select extract(month from a)");
should_parse!(extract_day, true, "select extract(day from a)");
should_parse!(extract_hour, true, "select extract(hour from a)");
should_parse!(extract_minute, true, "select extract(minute from a)");
should_parse!(extract_second, true, "select extract(second from a)");
should_parse!(
    substring_from,
    true,
    "select SUBSTRING(str FROM start FOR length)"
);
should_parse!(
    substring_comma,
    true,
    "select SUBSTRING(str, start, length)"
);
should_parse!(
    substring_comma_no_length,
    true,
    "select SUBSTRING(str, start)"
);
should_parse!(fold_upper, true, "select upper(a)");
should_parse!(fold_lower, true, "select lower(a)");
should_parse!(trim_leading, true, "select trim(LEADING substr FROM str)");
should_parse!(trim_trailing, true, "select trim(TRAILING substr FROM str)");
should_parse!(trim_both, true, "select trim(BOTH substr FROM str)");
should_parse!(current_timestamp_no_args, true, "select current_timestamp");
should_parse!(
    current_timestamp_with_args,
    true,
    "select current_timestamp(a)"
);
should_parse!(create_func_no_args, true, "select brand_new_func()");
should_parse!(
    create_func_with_args,
    true,
    "select brand_new_func(a, b, c)"
);
should_parse!(nested_scalar_func, true, "select nullif(coalesce(a, b), c)");
validate_expression_ast!(
    extract_ast,
    "extract(year from a)",
    Expression::Function(FunctionExpr {
        function: FunctionName("EXTRACT".to_string()),
        args: vec![
            FunctionArg::Extract(ExtractSpec::Year),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple("a".to_string())))
        ]
    })
);
validate_expression_ast!(
    fold_ast,
    "upper(a)",
    Expression::Function(FunctionExpr {
        function: FunctionName("FOLD".to_string()),
        args: vec![
            FunctionArg::Fold(Casing::Upper),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple("a".to_string())))
        ]
    })
);
validate_expression_ast!(
    trim_default_spec,
    "trim(substr FROM str)",
    Expression::Function(FunctionExpr {
        function: FunctionName("TRIM".to_string()),
        args: vec![
            FunctionArg::Trim(TrimSpec::Both),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple(
                "substr".to_string()
            ))),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple(
                "str".to_string()
            )))
        ]
    })
);
validate_expression_ast!(
    trim_default_substr,
    "trim(leading FROM str)",
    Expression::Function(FunctionExpr {
        function: FunctionName("TRIM".to_string()),
        args: vec![
            FunctionArg::Trim(TrimSpec::Leading),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple(" ".to_string()))),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple(
                "str".to_string()
            )))
        ]
    })
);
validate_expression_ast!(
    trim_default_spec_and_substr,
    "trim(str)",
    Expression::Function(FunctionExpr {
        function: FunctionName("TRIM".to_string()),
        args: vec![
            FunctionArg::Trim(TrimSpec::Both),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple(" ".to_string()))),
            FunctionArg::Expr(Expression::Identifier(Identifier::Simple(
                "str".to_string()
            )))
        ]
    })
);
