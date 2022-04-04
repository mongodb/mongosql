macro_rules! parsable {
    ($func_name:ident, expected = $expected:expr, $(expected_error = $expected_error:expr,)? input = $input:expr) => {
        #[test]
        fn $func_name() {
            let res = Parser::new().parse_query($input);
            let expected = $expected;
            if expected {
                res.expect("expected input to parse, but it failed");
            } else {
                assert!(res.is_err());
                #[allow(unused_variables)]
                match res {
                    Ok(_) => panic!("expected parse error, but parsing succeeded"),
                    Err(Error::Lalrpop(s)) => {$(assert_eq!($expected_error.to_string(), s))?},
                }
            }
        }
    };
}
macro_rules! validate_ast {
    ($func_name:ident, method = $method:ident, expected = $expected:expr, input = $input:expr,) => {
        #[test]
        fn $func_name() {
            let p = Parser::new();
            assert_eq!(p.$method($input).unwrap(), $expected)
        }
    };
}

mod select {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(star, expected = true, input = "select *");
    parsable!(star_upper, expected = true, input = "SELECT *");
    parsable!(mixed_case, expected = true, input = "SeLeCt *");
    parsable!(a_star, expected = true, input = "select a.*");
    parsable!(underscore_id, expected = true, input = "select _id");
    parsable!(contains_underscore, expected = true, input = "select a_b");
    parsable!(multiple, expected = true, input = "select a,b,c");
    parsable!(multiple_combo, expected = true, input = "select a,b,*");
    parsable!(multiple_star, expected = true, input = "select *,*");
    parsable!(multiple_dot_star, expected = true, input = "select a.*,b.*");
    parsable!(all_lower, expected = true, input = "select all *");
    parsable!(all_upper, expected = true, input = "select ALL *");
    parsable!(all_mixed_case, expected = true, input = "select aLl *");
    parsable!(distinct_lower, expected = true, input = "select distinct *");
    parsable!(distinct_upper, expected = true, input = "select DISTINCT *");
    parsable!(
        distinct_mixed_case,
        expected = true,
        input = "select DiSTinCt *"
    );
    parsable!(value_lower, expected = true, input = "SELECT value foo.*");
    parsable!(value_upper, expected = true, input = "SELECT VALUE foo.*");
    parsable!(
        value_mixed_case,
        expected = true,
        input = "SELECT vAlUe foo.*"
    );
    parsable!(
        values_lower,
        expected = true,
        input = "SELECT values foo.*, bar.*"
    );
    parsable!(
        values_upper,
        expected = true,
        input = "SELECT VALUES foo.*, bar.*"
    );
    parsable!(
        values_mixed_case,
        expected = true,
        input = "SELECT vAluES foo.*, bar.*"
    );
    parsable!(alias_lower, expected = true, input = "SELECT foo as f");
    parsable!(alias_upper, expected = true, input = "SELECT foo AS f");
    parsable!(alias_mixed_case, expected = true, input = "SELECT foo aS f");
    parsable!(alias_no_as, expected = true, input = "SELECT foo f");
    parsable!(
        alias_compound_column,
        expected = true,
        input = "SELECT a.b as a"
    );
    parsable!(
        alias_multiple_combined,
        expected = true,
        input = "SELECT a, b AS c, a.c"
    );
    parsable!(long_compound, expected = true, input = "SELECT a.b.c.d");
    parsable!(letter_number_ident, expected = true, input = "SELECT a9");
    parsable!(
        delimited_ident_quotes,
        expected = true,
        input = r#"SELECT "foo""#
    );
    parsable!(
        delimited_ident_backticks,
        expected = true,
        input = "SELECT `foo`"
    );
    parsable!(
        delimited_quote_empty,
        expected = true,
        input = r#"SELECT """#
    );
    parsable!(
        delimited_backtick_empty,
        expected = true,
        input = "SELECT ``"
    );
    parsable!(
        delimited_escaped_quote,
        expected = true,
        input = r#"SELECT "fo""o""""""#
    );
    parsable!(
        delimited_escaped_backtick,
        expected = true,
        input = "SELECT `f``oo`````"
    );

    parsable!(use_stmt, expected = false, input = "use foo");
    parsable!(compound_star, expected = false, input = "SELECT a.b.c.*");
    parsable!(
        numerical_ident_prefix,
        expected = false,
        input = "SELECT 9ae"
    );
    parsable!(value_star, expected = false, input = "SELECT VALUE *");
    parsable!(
        value_alias,
        expected = false,
        input = "SELECT VALUE foo AS f"
    );
    parsable!(dangling_alias, expected = false, input = "SELECT a.b AS");
    parsable!(compound_alias, expected = false, input = "SELECT a AS b.c");

    parsable!(
        delimited_extra_quote_outer,
        expected = false,
        input = r#"SELECT ""foo"""#
    );
    parsable!(
        delimited_extra_backtick_outer,
        expected = false,
        input = "SELECT ``foo``"
    );
    parsable!(
        delimited_escaped_quote_odd,
        expected = false,
        input = r#"SELECT "f"oo"""#
    );
    parsable!(
        delimited_escaped_backtick_odd,
        expected = false,
        input = "SELECT `foo````"
    );
    parsable!(
        delimited_backslash_escape,
        expected = false,
        input = r#"SELECT "fo\"\"o""#
    );
    parsable!(
        unescaped_quotes_in_ident,
        expected = false,
        input = r#"SELECT fo""o"#
    );
    parsable!(
        unescaped_backticks_in_ident,
        expected = false,
        input = "SELECT fo``o"
    );

    validate_ast!(
        ident,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()))
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT foo",
    );
    validate_ast!(
        delimited_quote,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = r#"SELECT "foo""#,
    );
    validate_ast!(
        delimited_backtick,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "select `foo`",
    );
    validate_ast!(
        delimited_escaped_backtick_ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("fo`o``".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT `fo``o`````",
    );
    validate_ast!(
        delimited_escaped_quote_ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier(
                        r#"fo"o"""#.to_string()
                    ),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = r#"SELECT "fo""o""""""#,
    );
    validate_ast!(
        backtick_delimiter_escaped_quote,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier(
                        r#"fo""o"#.to_string()
                    ),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = r#"SELECT `fo""o`"#,
    );
    validate_ast!(
        quote_delimiter_escaped_backtick,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("fo``o".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = r#"SELECT "fo``o""#,
    );
}

mod query {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(
        select_union_simple,
        expected = true,
        input = "SELECT a UNION SELECT b"
    );
    parsable!(
        select_union_multiple,
        expected = true,
        input = "SELECT a UNION SELECT b UNION SELECT c"
    );
    parsable!(
        select_union_all_multiple,
        expected = true,
        input = "SELECT a UNION ALL SELECT b UNION ALL SELECT c"
    );

    validate_ast!(
        union_is_left_associative,
        method = parse_query,
        expected = Query::Set(SetQuery {
            left: Box::new(Query::Set(SetQuery {
                left: Box::new(Query::Select(SelectQuery {
                    select_clause: SelectClause {
                        set_quantifier: SetQuantifier::All,
                        body: SelectBody::Standard(vec![SelectExpression::Expression(
                            OptionallyAliasedExpr::Unaliased(Expression::Identifier(
                                "a".to_string()
                            ),)
                        )])
                    },
                    from_clause: None,
                    where_clause: None,
                    group_by_clause: None,
                    having_clause: None,
                    order_by_clause: None,
                    limit: None,
                    offset: None,
                })),
                op: SetOperator::Union,
                right: Box::new(Query::Select(SelectQuery {
                    select_clause: SelectClause {
                        set_quantifier: SetQuantifier::All,
                        body: SelectBody::Standard(vec![SelectExpression::Expression(
                            OptionallyAliasedExpr::Unaliased(Expression::Identifier(
                                "b".to_string()
                            ),)
                        )])
                    },
                    from_clause: None,
                    where_clause: None,
                    group_by_clause: None,
                    having_clause: None,
                    order_by_clause: None,
                    limit: None,
                    offset: None,
                }))
            })),
            op: SetOperator::UnionAll,
            right: Box::new(Query::Select(SelectQuery {
                select_clause: SelectClause {
                    set_quantifier: SetQuantifier::All,
                    body: SelectBody::Standard(vec![SelectExpression::Expression(
                        OptionallyAliasedExpr::Unaliased(Expression::Identifier("c".to_string()),)
                    )])
                },
                from_clause: None,
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None,
            }))
        }),
        input = "select a union select b union all select c",
    );
}

mod operator {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(unary_pos, expected = true, input = "select +a");
    parsable!(unary_neg, expected = true, input = "select -a");
    parsable!(unary_not, expected = true, input = "select NOT a");
    parsable!(binary_add, expected = true, input = "select a+b+c+d+e");
    parsable!(binary_sub, expected = true, input = "select a-b-c-d-e");
    parsable!(binary_mul, expected = true, input = "select a*b*c*d*e");
    parsable!(binary_mul_add, expected = true, input = "select a*b+c*d+e");
    parsable!(binary_div_add, expected = true, input = "select a/b+c/d+e");
    parsable!(binary_div_mul, expected = true, input = "select a/b*c");
    parsable!(binary_num, expected = true, input = "select 3+4-5/7*8");
    parsable!(binary_lt, expected = true, input = "select a<b<c<d<e");
    parsable!(binary_lte, expected = true, input = "select a<=b");
    parsable!(binary_gt, expected = true, input = "select a>b");
    parsable!(binary_gte, expected = true, input = "select a>=b");
    parsable!(binary_neq_1, expected = true, input = "select a!=b");
    parsable!(binary_neq_2, expected = true, input = "select a<>b");
    parsable!(binary_eq, expected = true, input = "select a=b");
    parsable!(
        binary_string_concat,
        expected = true,
        input = "select a || b"
    );
    parsable!(binary_or, expected = true, input = "select a OR b");
    parsable!(binary_and, expected = true, input = "select a AND b");
    parsable!(
        binary_compare_and_add,
        expected = true,
        input = "select b<a+c and b>d+e"
    );
    parsable!(
        binary_compare_or_mul,
        expected = true,
        input = "select b<a*c or b>d*e"
    );
    parsable!(
        binary_lt_and_neq,
        expected = true,
        input = "select a<b and c<>e"
    );
    parsable!(between, expected = true, input = "select a BETWEEN b AND c");
    parsable!(
        binary_between,
        expected = true,
        input = "SELECT 1 between 0 and 3 and 2 between 1 and 3"
    );
    parsable!(
        case,
        expected = true,
        input = "select CASE WHEN a=b THEN a ELSE c END"
    );
    parsable!(
        case_multiple_when_clauses,
        expected = true,
        input = "select CASE WHEN a or b THEN a WHEN c=d THEN c ELSE e END"
    );
    parsable!(
        case_multiple_exprs,
        expected = true,
        input = "select CASE a WHEN a <> b THEN a WHEN c and d THEN c ELSE e END"
    );
    parsable!(
        binary_case,
        expected = true,
        input = "select CASE WHEN a=b THEN a ELSE c END + CASE WHEN c=d THEN c ELSE e END"
    );
    parsable!(
        case_between,
        expected = true,
        input = "select CASE when a BETWEEN b and c THEN a ELSE b END"
    );
    parsable!(
        in_op_in_between_expr,
        expected = true,
        input = "SELECT a BETWEEN (b IN c) AND d"
    );

    parsable!(is_type, expected = true, input = "select a IS STRING");
    parsable!(
        is_not_type,
        expected = true,
        input = "select a-8 IS NOT DECIMAL(1)"
    );
    parsable!(is_missing, expected = true, input = "select a IS MISSING");
    parsable!(
        is_not_missing,
        expected = true,
        input = "select (a+b) IS NOT MISSING"
    );
    parsable!(is_number, expected = true, input = "select a IS NUMBER");
    parsable!(
        is_not_number,
        expected = true,
        input = "select a IS NOT NUMBER"
    );
    parsable!(like, expected = true, input = "select col1 LIKE 'A%'");
    parsable!(
        not_like_multiple_spaces,
        expected = true,
        input = "select col1 NOT   LIKE '[a-z][a-z]'"
    );
    parsable!(
        like_escape,
        expected = true,
        input = "select col1 LIKE '%a!% b' ESCAPE '!'"
    );
    parsable!(
        like_escape_multibyte_unicode,
        expected = true,
        input = "select col1 LIKE '%a!% b' ESCAPE '山'"
    );
    parsable!(
        like_escape_empty,
        expected = false,
        input = "select col1 LIKE 'blah' ESCAPE ''"
    );
    parsable!(
        like_escape_multichar,
        expected = false,
        input = "select col1 LIKE 'blah' ESCAPE 'blah'"
    );
    parsable!(
        not_like_escape,
        expected = true,
        input = "select col1 NOT LIKE '%a!% b' ESCAPE '!'"
    );
    parsable!(
        not_like_escape_empty,
        expected = false,
        input = "select col1 NOT LIKE 'blah' ESCAPE ''"
    );
    parsable!(
        not_like_escape_multichar,
        expected = false,
        input = "select col1 NOT LIKE 'blah' ESCAPE 'blah'"
    );
    parsable!(
        where_is,
        expected = true,
        input = "select * where a IS NULL"
    );
    parsable!(
        where_like,
        expected = true,
        input = "select * where col1 LIKE 'A%'"
    );

    validate_ast!(
        is_missing_ast,
        method = parse_expression,
        expected = Expression::Is(IsExpr {
            expr: Box::new(Expression::Identifier("a".to_string())),
            target_type: TypeOrMissing::Missing,
        }),
        input = "a IS MISSING",
    );
    validate_ast!(
        is_number_expr,
        method = parse_expression,
        expected = Expression::Is(IsExpr {
            expr: Box::new(Expression::Literal(Literal::Integer(1))),
            target_type: TypeOrMissing::Number,
        }),
        input = "1 IS NUMBER",
    );
    validate_ast!(
        is_not_number_expr,
        method = parse_expression,
        expected = Expression::Unary(UnaryExpr {
            op: UnaryOp::Not,
            expr: Box::new(Expression::Is(IsExpr {
                expr: Box::new(Expression::Literal(Literal::Integer(1))),
                target_type: TypeOrMissing::Number,
            })),
        }),
        input = "1 IS NOT NUMBER",
    );
    parsable!(
        between_invalid_binary_op,
        expected = false,
        input = "select a BETWEEN b + c"
    );
    parsable!(
        not_between_invalid_binary_op,
        expected = false,
        input = "select a NOT BETWEEN b / c"
    );
    parsable!(
        case_non_bool_conditions,
        expected = false,
        input = "select case a when a+b then a else c-d"
    );

    validate_ast!(
        binary_sub_unary_neg_ast,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier("b".to_string())),
            op: BinaryOp::Sub,
            right: Box::new(Expression::Unary(UnaryExpr {
                op: UnaryOp::Neg,
                expr: Box::new(Expression::Identifier("a".to_string()))
            }))
        }),
        input = "b- -a",
    );

    validate_ast!(
        binary_mul_add_ast,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier("c".to_string())),
                op: BinaryOp::Mul,
                right: Box::new(Expression::Identifier("a".to_string()))
            })),
            op: BinaryOp::Add,
            right: Box::new(Expression::Identifier("b".to_string()))
        }),
        input = "c*a+b",
    );

    validate_ast!(
        binary_add_concat_ast,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier("a".to_string())),
                op: BinaryOp::Add,
                right: Box::new(Expression::Identifier("b".to_string()))
            })),
            op: BinaryOp::Concat,
            right: Box::new(Expression::Identifier("c".to_string()))
        }),
        input = "a+b||c",
    );

    validate_ast!(
        binary_concat_compare_ast,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier("c".to_string())),
            op: BinaryOp::Comparison(ComparisonOp::Gt),
            right: Box::new(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier("a".to_string())),
                op: BinaryOp::Concat,
                right: Box::new(Expression::Identifier("b".to_string()))
            }))
        }),
        input = "c>a||b",
    );

    validate_ast!(
        binary_compare_and_ast,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier("a".to_string())),
                op: BinaryOp::Comparison(ComparisonOp::Lt),
                right: Box::new(Expression::Identifier("b".to_string()))
            })),
            op: BinaryOp::And,
            right: Box::new(Expression::Identifier("c".to_string()))
        }),
        input = "a<b AND c",
    );

    validate_ast!(
        binary_and_or_ast,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier("a".to_string())),
                op: BinaryOp::And,
                right: Box::new(Expression::Identifier("b".to_string()))
            })),
            op: BinaryOp::Or,
            right: Box::new(Expression::Identifier("b".to_string()))
        }),
        input = "a AND b OR b",
    );

    validate_ast!(
        between_ast,
        method = parse_expression,
        expected = Expression::Between(BetweenExpr {
            expr: Box::new(Expression::Identifier("a".to_string())),
            min: Box::new(Expression::Identifier("b".to_string())),
            max: Box::new(Expression::Identifier("c".to_string())),
        }),
        input = "a between b and c",
    );

    validate_ast!(
        not_between_ast,
        method = parse_expression,
        expected = Expression::Unary(UnaryExpr {
            op: UnaryOp::Not,
            expr: Box::new(Expression::Between(BetweenExpr {
                expr: Box::new(Expression::Identifier("a".to_string())),
                min: Box::new(Expression::Identifier("b".to_string())),
                max: Box::new(Expression::Identifier("c".to_string())),
            }))
        }),
        input = "a not between b and c",
    );

    validate_ast!(
        case_multiple_when_branches_ast,
        method = parse_expression,
        expected = Expression::Case(CaseExpr {
            expr: None,
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Binary(BinaryExpr {
                        left: Box::new(Expression::Identifier("a".to_string())),
                        op: BinaryOp::Comparison(ComparisonOp::Eq),
                        right: Box::new(Expression::Identifier("b".to_string()))
                    })),
                    then: Box::new(Expression::Identifier("a".to_string()))
                },
                WhenBranch {
                    when: Box::new(Expression::Binary(BinaryExpr {
                        left: Box::new(Expression::Identifier("c".to_string())),
                        op: BinaryOp::Comparison(ComparisonOp::Eq),
                        right: Box::new(Expression::Identifier("d".to_string()))
                    })),
                    then: Box::new(Expression::Identifier("c".to_string()))
                }
            ],
            else_branch: Some(Box::new(Expression::Identifier("e".to_string())))
        }),
        input = "case when a=b then a when c=d then c else e end",
    );

    validate_ast!(
        case_multiple_exprs_ast,
        method = parse_expression,
        expected = Expression::Case(CaseExpr {
            expr: Some(Box::new(Expression::Identifier("a".to_string()))),
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Binary(BinaryExpr {
                    left: Box::new(Expression::Identifier("a".to_string())),
                    op: BinaryOp::Comparison(ComparisonOp::Eq),
                    right: Box::new(Expression::Identifier("b".to_string()))
                })),
                then: Box::new(Expression::Identifier("a".to_string()))
            }],
            else_branch: Some(Box::new(Expression::Identifier("c".to_string())))
        }),
        input = "case a when a=b then a else c end",
    );
}

mod group_by {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(simple, expected = true, input = "select * group by a");
    parsable!(compound, expected = true, input = "select * group by a.b");
    parsable!(alias, expected = true, input = "select * group by a as b");
    parsable!(
        aggregate_count_star,
        expected = true,
        input = "select * group by a aggregate count(*) as b"
    );
    parsable!(
        aggregate_alias,
        expected = true,
        input = "select * group by a aggregate sum(a) as b"
    );
    parsable!(
        aggregate_all,
        expected = true,
        input = "select * group by a aggregate sum(all a) as b"
    );
    parsable!(
        aggregate_distinct,
        expected = true,
        input = "select * group by a aggregate sum(distinct a) as b"
    );
    parsable!(
        aggregate_distinct_alias,
        expected = true,
        input = "select * group by a aggregate sum(distinct a) as b"
    );
    parsable!(
        aggregate_distinct_all,
        expected = false,
        input = "select * group by a aggregate sum(distinct all a) as b"
    );
    parsable!(none, expected = false, input = "select * group by");
    parsable!(
        aggregate_none,
        expected = false,
        input = "select * group by a aggregate"
    );
    parsable!(
        aggregate_no_args,
        expected = true,
        input = "select * group by a aggregate sum() as b"
    );
    parsable!(
        aggregate_no_alias,
        expected = false,
        input = "select * group by a aggregate sum()"
    );
    parsable!(
        aggregate_second_no_alias,
        expected = false,
        input = "select * group by a aggregate sum(a) AS a, AVG(b), sum(c)"
    );

    validate_ast!(
        aggregate_distinct_with_alias,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: Some(GroupByClause {
                keys: vec![
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("a".to_string())),
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("b".to_string()))
                ],
                aggregations: vec![AliasedExpr {
                    expr: Expression::Function(FunctionExpr {
                        function: FunctionName::Sum,
                        args: FunctionArguments::Args(vec![Expression::Identifier(
                            "b".to_string()
                        )]),
                        set_quantifier: Some(SetQuantifier::Distinct),
                    }),
                    alias: "c".to_string(),
                }]
            }),
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "select * group by a, b aggregate sum(distinct b) as c",
    );
}

mod having {
    use crate::ast::*;
    use crate::parser::{Error, Parser};

    parsable!(simple, expected = true, input = "select * having y");
    parsable!(
        with_group_by,
        expected = true,
        input = "select * group by a having y"
    );
    parsable!(
        with_aggregation_function,
        expected = true,
        input = "select * group by a having sum(a) > 0 "
    );

    validate_ast!(
        with_aggregation_distinct_and_group_by,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: Some(GroupByClause {
                keys: vec![OptionallyAliasedExpr::Unaliased(Expression::Identifier(
                    "a".to_string()
                ),)],
                aggregations: vec![]
            }),
            having_clause: Some(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Function(FunctionExpr {
                    function: FunctionName::Sum,
                    args: FunctionArguments::Args(vec![Expression::Identifier("a".to_string())]),
                    set_quantifier: Some(SetQuantifier::Distinct),
                })),
                op: BinaryOp::Comparison(ComparisonOp::Gt),
                right: Box::new(Expression::Literal(Literal::Integer(0)))
            })),
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "select * group by a having sum(distinct a) > 0",
    );
}

mod order_by {
    use crate::ast::*;
    use crate::parser::{Error, Parser};

    parsable!(simple_ident, expected = true, input = "select * order by a");
    parsable!(
        compound_ident,
        expected = true,
        input = "select * order by a.b"
    );
    parsable!(asc, expected = true, input = "select * order by a ASC");
    parsable!(desc, expected = true, input = "select * order by a DESC");
    parsable!(
        multiple,
        expected = true,
        input = "select a, b, c order by a, b"
    );
    parsable!(
        multiple_directions,
        expected = true,
        input = "select * order by a DESC, b ASC, c"
    );
    parsable!(
        positional_sort,
        expected = true,
        input = "select a, b order by 1, 2"
    );
    parsable!(
        positional_sort_with_star,
        expected = true,
        input = "select * order by 1"
    );

    validate_ast!(
        default_direction,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: Some(OrderByClause {
                sort_specs: vec![SortSpec {
                    key: SortKey::Simple(Expression::Identifier("a".to_string())),
                    direction: SortDirection::Asc
                }]
            }),
            limit: None,
            offset: None,
        }),
        input = "select * order by a",
    );
}

mod limit_offset {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(limit_simple, expected = true, input = "select * limit 42");
    parsable!(offset_simple, expected = true, input = "select * offset 42");
    parsable!(
        limit_comma_offset,
        expected = true,
        input = "select * limit 42, 24"
    );
    parsable!(
        limit_then_offset,
        expected = true,
        input = "select * limit 42 offset 24"
    );
    parsable!(
        offset_then_limit,
        expected = true,
        input = "select * offset 42 limit 24"
    );
    parsable!(
        offset_twice,
        expected = false,
        input = "select * limit 42, 24 offset 24"
    );
    parsable!(
        limit_alphabetic,
        expected = false,
        input = "select * limit a"
    );
    parsable!(
        limit_non_integer,
        expected = false,
        input = "select * limit 42.0"
    );
    parsable!(
        limit_negative,
        expected = false,
        input = "select * limit -42"
    );
    parsable!(
        limit_overflow,
        expected = false,
        input = "select * limit 4294967296"
    ); // 2^32

    validate_ast!(
        limit_one_value,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(42_u32),
            offset: None,
        }),
        input = "select * limit 42",
    );

    validate_ast!(
        limit_two_values,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(42_u32),
            offset: Some(24_u32),
        }),
        input = "select * limit 42, 24",
    );

    validate_ast!(
        limit_with_offset,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(42_u32),
            offset: Some(24_u32),
        }),
        input = "select * limit 42 offset 24",
    );
}

mod fetch_first {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(
        simple,
        expected = true,
        input = "select * fetch first 42 rows only"
    );
    parsable!(
        then_offset,
        expected = true,
        input = "select * fetch first 42 rows only offset 24"
    );
    parsable!(
        offset_then_fetch_first,
        expected = true,
        input = "select * offset 42 fetch first 24 rows only"
    );
    parsable!(
        row_synonym,
        expected = true,
        input = "select * fetch first 42 row only"
    );
    parsable!(
        fetch_next_synonym,
        expected = true,
        input = "select * fetch next 42 rows only"
    );
    parsable!(
        fetch_next_row_synonym,
        expected = true,
        input = "select * fetch next 42 row only"
    );
    parsable!(
        comma_offset,
        expected = false,
        input = "select * fetch first 24 rows only, 24"
    );
    parsable!(
        alphabetic,
        expected = false,
        input = "select * fetch first a rows only"
    );
    parsable!(
        non_integer,
        expected = false,
        input = "select * fetch first 2.0 rows only"
    );
    parsable!(
        negative,
        expected = false,
        input = "select * fetch first -42 rows only"
    );
    parsable!(
        overflow,
        expected = false,
        input = "select * fetch first 4294967296 rows only"
    );

    validate_ast!(
        no_offset,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(42_u32),
            offset: None,
        }),
        input = "select * fetch first 42 rows only",
    );

    validate_ast!(
        offset,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(42_u32),
            offset: Some(24_u32),
        }),
        input = "select * fetch first 42 rows only offset 24",
    );

    validate_ast!(
        synonyms,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: Some(42_u32),
            offset: Some(24_u32),
        }),
        input = "select * fetch next 42 row only offset 24",
    );

    parsable!(
        neg_positional_sort,
        expected = false,
        input = "select a, b order by -1, 2"
    );
    parsable!(
        positional_sort_too_big,
        expected = false,
        input = "select a, b order by 9223372036854775808"
    );
}

mod literals {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(null, expected = true, input = "select null");
    parsable!(null_mixed_case, expected = true, input = "select nULL");
    parsable!(unsigned_int, expected = true, input = "select 123");
    parsable!(neg_int, expected = true, input = "select -123");
    parsable!(pos_int, expected = true, input = "select +123");
    parsable!(int_leading_zeros, expected = true, input = "select 008");
    parsable!(
        convert_to_long,
        expected = true,
        input = "select 2147483648"
    );
    parsable!(string, expected = true, input = "select 'foo'");
    parsable!(
        string_special_characters,
        expected = true,
        input = "select 'αβγ'"
    );
    parsable!(empty_string, expected = true, input = "select ''");
    parsable!(unsigned_double, expected = true, input = "select 0.5");
    parsable!(
        unsigned_double_no_fraction,
        expected = true,
        input = "select 1."
    );
    parsable!(
        unsigned_double_no_whole_num,
        expected = true,
        input = "select .6"
    );
    parsable!(neg_double, expected = true, input = "select -4.089015");
    parsable!(pos_double, expected = true, input = "select +0.0");
    parsable!(
        double_exponent_lowercase,
        expected = true,
        input = "select 1e2"
    );
    parsable!(
        double_exponent_uppercase,
        expected = true,
        input = "select 2E3"
    );
    parsable!(
        double_exponent_beg_fraction,
        expected = true,
        input = "select 8.e+23"
    );
    parsable!(
        double_exponent_mid_fraction,
        expected = true,
        input = "select 9.07e-2"
    );
    parsable!(
        double_exponent_no_whole_num,
        expected = true,
        input = "select .2E3"
    );
    parsable!(
        double_exponent_signed,
        expected = true,
        input = "select -7.2E3"
    );
    parsable!(
        boolean_true,
        expected = true,
        input = "select expected = true"
    );
    parsable!(
        boolean_false,
        expected = true,
        input = "select expected = false"
    );
    parsable!(
        boolean_binary,
        expected = true,
        input = "select expected = true AND expected = false OR expected = false"
    );

    parsable!(string_single_quote, expected = false, input = "select '''");
    parsable!(
        double_exponent_no_exp,
        expected = false,
        input = "select 1e"
    );
    parsable!(
        long_too_big,
        expected = false,
        input = "select 9223372036854775808"
    );

    validate_ast!(
        string_escaped_quote_no_chars,
        method = parse_expression,
        expected = Expression::Literal(Literal::String(r#"'"#.to_string())),
        input = "''''",
    );

    validate_ast!(
        string_escaped_quote,
        method = parse_expression,
        expected = Expression::Literal(Literal::String(r#"foo's"#.to_string())),
        input = "'foo''s'",
    );

    validate_ast!(
        double_neg_no_decimal,
        method = parse_expression,
        expected = Expression::Unary(UnaryExpr {
            op: UnaryOp::Neg,
            expr: Box::new(Expression::Literal(Literal::Double(2000.0)))
        }),
        input = "-2E+3",
    );

    validate_ast!(
        double_no_whole,
        method = parse_expression,
        expected = Expression::Literal(Literal::Double(0.0002)),
        input = ".2E-3",
    );

    validate_ast!(
        double_no_frac_or_sign,
        method = parse_expression,
        expected = Expression::Literal(Literal::Double(234000000.0)),
        input = "234.E6",
    );

    validate_ast!(
        double_all_components,
        method = parse_expression,
        expected = Expression::Literal(Literal::Double(0.2342)),
        input = "234.2E-3",
    );

    validate_ast!(
        double_binary_add,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Literal(Literal::Double(2000.0))),
            op: BinaryOp::Add,
            right: Box::new(Expression::Literal(Literal::Double(0.0000000516)))
        }),
        input = "2E3 + 5.16E-8",
    );
}

mod array {
    use crate::parser::{Error, Parser};
    parsable!(empty, expected = true, input = "select []");
    parsable!(homogeneous, expected = true, input = "select [1, 2, 3]");
    parsable!(
        heterogeneous,
        expected = true,
        input = "select [1, 'a', expected = true, -42]"
    );
    parsable!(indexing, expected = true, input = "select [1, 2, 3][0]");
}

mod parenthized_expression {
    use crate::parser::{Error, Parser};
    parsable!(
        multiple_binary_ops,
        expected = true,
        input = "SELECT ((a+b)-(d/c))*7"
    );
    parsable!(
        case_expr,
        expected = true,
        input = "select (CASE WHEN a=b THEN 1 ELSE 2 END)*4"
    );
    parsable!(unbalanced, expected = false, input = "SELECT ((a+b)");
    parsable!(empty, expected = false, input = "SELECT ()");
}

mod scalar_function {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(null_if, expected = true, input = "select nullif(a, b)");
    parsable!(
        coalesce,
        expected = true,
        input = "select coalesce(a, b, c, d)"
    );
    parsable!(size, expected = true, input = "select size(a)");
    parsable!(
        position,
        expected = true,
        input = "select position('b' IN 'abc')"
    );
    parsable!(
        position_binary_mul,
        expected = true,
        input = "select position(1*2 IN 2)"
    );
    parsable!(
        position_is_op,
        expected = true,
        input = "SELECT POSITION(x IN (`foo` IS DOCUMENT)) from bar"
    );
    parsable!(
        position_unary_neg,
        expected = true,
        input = "select position(-2 IN -2)"
    );
    parsable!(
        position_between_parens,
        expected = true,
        input = "select position((b BETWEEN c AND c) IN expected = true)"
    );
    parsable!(
        position_binary_or,
        expected = true,
        input = "select position(a OR b IN expected = true)"
    );
    parsable!(
        position_binary_compare,
        expected = true,
        input = "select position(expected = true IN a < b)"
    );
    parsable!(
        char_length,
        expected = true,
        input = "select char_length('foo')"
    );
    parsable!(
        character_length,
        expected = true,
        input = "select character_length('bar')"
    );
    parsable!(
        octet_length,
        expected = true,
        input = "select octet_length(a)"
    );
    parsable!(bit_length, expected = true, input = "select bit_length(a)");
    parsable!(
        extract_year,
        expected = true,
        input = "select extract(year from a)"
    );
    parsable!(
        extract_month,
        expected = true,
        input = "select extract(month from a)"
    );
    parsable!(
        extract_day,
        expected = true,
        input = "select extract(day from a)"
    );
    parsable!(
        extract_hour,
        expected = true,
        input = "select extract(hour from a)"
    );
    parsable!(
        extract_minute,
        expected = true,
        input = "select extract(minute from a)"
    );
    parsable!(
        extract_second,
        expected = true,
        input = "select extract(second from a)"
    );
    parsable!(
        slice_arr,
        expected = true,
        input = "select slice([42, 43, 44])"
    );
    parsable!(
        slice_arr_length,
        expected = true,
        input = "select slice([42, 43, 44], 1)"
    );
    parsable!(
        slice_arr_start_length,
        expected = true,
        input = "select slice([42, 43, 44], 0, 1)"
    );
    parsable!(
        substring_from,
        expected = true,
        input = "select SUBSTRING(str FROM start FOR length)"
    );
    parsable!(
        substring_comma,
        expected = true,
        input = "select SUBSTRING(str, start, length)"
    );
    parsable!(
        substring_comma_no_length,
        expected = true,
        input = "select SUBSTRING(str, start)"
    );
    parsable!(fold_upper, expected = true, input = "select upper(a)");
    parsable!(fold_lower, expected = true, input = "select lower(a)");
    parsable!(
        trim_leading,
        expected = true,
        input = "select trim(LEADING substr FROM str)"
    );
    parsable!(
        trim_trailing,
        expected = true,
        input = "select trim(TRAILING substr FROM str)"
    );
    parsable!(
        trim_both,
        expected = true,
        input = "select trim(BOTH substr FROM str)"
    );
    parsable!(
        current_timestamp_no_args,
        expected = true,
        input = "select current_timestamp"
    );
    parsable!(
        current_timestamp_with_args,
        expected = true,
        input = "select current_timestamp(a)"
    );
    parsable!(
        create_func_no_args,
        expected = false,
        input = "select brand_new_func()"
    );
    parsable!(
        create_func_with_args,
        expected = false,
        input = "select brand_new_func(a, b, c)"
    );
    parsable!(
        nested_scalar_func,
        expected = true,
        input = "select nullif(coalesce(a, b), c)"
    );
    parsable!(
        position_invalid_binary_op,
        expected = false,
        input = "select position(x OR y)"
    );
    parsable!(
        scalar_function_binary_op,
        expected = true,
        input = "select char_length('foo') + 5"
    );
    parsable!(
        user_defined_function_not_allowed,
        expected = false,
        expected_error = "unknown function myFunc",
        input = "select myFunc(x)"
    );

    validate_ast!(
        position_ast,
        method = parse_expression,
        expected = Expression::Function(FunctionExpr {
            function: FunctionName::Position,
            args: FunctionArguments::Args(vec![
                Expression::Tuple(vec![Expression::Binary(BinaryExpr {
                    left: Box::new(Expression::Identifier("a".to_string())),
                    op: BinaryOp::Add,
                    right: Box::new(Expression::Binary(BinaryExpr {
                        left: Box::new(Expression::Identifier("b".to_string())),
                        op: BinaryOp::Mul,
                        right: Box::new(Expression::Identifier("c".to_string()))
                    }))
                })]),
                Expression::Identifier("d".to_string()),
            ]),
            set_quantifier: None,
        }),
        input = "position((a+b*c) IN d)",
    );
    validate_ast!(
        extract_ast,
        method = parse_expression,
        expected = Expression::Extract(ExtractExpr {
            extract_spec: ExtractSpec::Year,
            arg: Box::new(Expression::Identifier("a".to_string()))
        }),
        input = "extract(year from a)",
    );
    validate_ast!(
        trim_default_spec,
        method = parse_expression,
        expected = Expression::Trim(TrimExpr {
            trim_spec: TrimSpec::Both,
            trim_chars: Box::new(Expression::Identifier("substr".into())),
            arg: Box::new(Expression::Identifier("str".to_string())),
        }),
        input = "trim(substr FROM str)",
    );
    validate_ast!(
        trim_default_substr,
        method = parse_expression,
        expected = Expression::Trim(TrimExpr {
            trim_spec: TrimSpec::Leading,
            trim_chars: Box::new(Expression::Literal(Literal::String(" ".into()))),
            arg: Box::new(Expression::Identifier("str".to_string())),
        }),
        input = "trim(leading FROM str)",
    );
    validate_ast!(
        trim_default_spec_and_substr,
        method = parse_expression,
        expected = Expression::Trim(TrimExpr {
            trim_spec: TrimSpec::Both,
            trim_chars: Box::new(Expression::Literal(Literal::String(" ".into()))),
            arg: Box::new(Expression::Identifier("str".to_string())),
        }),
        input = "trim(str)",
    );
    validate_ast!(
        fold_ast,
        method = parse_expression,
        expected = Expression::Function(FunctionExpr {
            function: FunctionName::Upper,
            args: FunctionArguments::Args(vec![Expression::Identifier("a".to_string())]),
            set_quantifier: None,
        }),
        input = "upper(a)",
    );
}

mod from {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(no_qualifier, expected = true, input = "SELECT * FROM foo");
    parsable!(qualifier, expected = true, input = "SELECT * FROM bar.foo");
    parsable!(
        no_qualifier_with_alias,
        expected = true,
        input = "SELECT * FROM foo car"
    );
    parsable!(
        qualifier_with_alias,
        expected = true,
        input = "SELECT * FROM bar.foo car"
    );
    parsable!(
        no_qualifier_with_as_alias,
        expected = true,
        input = "SELECT * FROM foo AS car"
    );
    parsable!(
        qualifier_with_as_alias,
        expected = true,
        input = "SELECT * FROM bar.foo AS car"
    );
    parsable!(
        array_with_alias,
        expected = true,
        input = "SELECT * FROM [{'a': 1}, {'b': 2}] arr"
    );
    parsable!(
        array_with_as_alias,
        expected = true,
        input = "SELECT * FROM [{'a': 1}, {'b': 2}] AS arr"
    );

    parsable!(
        two_comma_join_second_alias,
        expected = true,
        input = "SELECT * FROM foo, bar AS bar"
    );
    parsable!(
        two_comma_join_first_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo, bar"
    );
    parsable!(
        two_comma_join_both_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo, bar AS bar"
    );
    parsable!(
        two_comma_join,
        expected = true,
        input = "SELECT * FROM foo, bar"
    );
    parsable!(
        three_comma_join,
        expected = true,
        input = "SELECT * FROM foo, bar AS bar, car"
    );

    parsable!(
        two_inner_join_second_alias,
        expected = true,
        input = "SELECT * FROM foo JOIN bar AS bar"
    );
    parsable!(
        two_inner_join_first_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo INNER JOIN bar"
    );
    parsable!(
        two_inner_join_both_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo INNER JOIN bar AS bar"
    );
    parsable!(
        two_inner_join,
        expected = true,
        input = "SELECT * FROM foo JOIN bar"
    );

    parsable!(
        two_cross_join_second_alias,
        expected = true,
        input = "SELECT * FROM foo CROSS JOIN bar AS bar"
    );
    parsable!(
        two_cross_join_first_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo CROSS JOIN bar"
    );
    parsable!(
        two_cross_join_both_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo CROSS JOIN bar AS bar"
    );
    parsable!(
        two_cross_join,
        expected = true,
        input = "SELECT * FROM foo CROSS JOIN bar"
    );

    parsable!(
        two_left_join_second_alias,
        expected = true,
        input = "SELECT * FROM foo LEFT JOIN bar AS bar"
    );
    parsable!(
        two_left_join_first_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo LEFT OUTER JOIN bar"
    );
    parsable!(
        two_left_join_both_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo LEFT OUTER JOIN bar AS bar"
    );
    parsable!(
        two_left_join,
        expected = true,
        input = "SELECT * FROM foo LEFT JOIN bar"
    );

    parsable!(
        two_left_join_second_alias_with_on,
        expected = true,
        input = "SELECT * FROM foo LEFT JOIN bar AS bar ON 1 = 2"
    );
    parsable!(
        two_left_join_first_alias_with_on,
        expected = true,
        input = "SELECT * FROM foo AS foo LEFT OUTER JOIN bar ON 1 = 2"
    );
    parsable!(
        two_left_join_both_alias_with_on,
        expected = true,
        input = "SELECT * FROM foo AS foo LEFT OUTER JOIN bar AS bar ON 1 = 2"
    );

    parsable!(
        three_inner_join_with_ons,
        expected = true,
        input = "SELECT * FROM foo JOIN bar ON 1 = 2 JOIN car ON 3 = 4"
    );

    parsable!(
        two_right_join_second_alias,
        expected = true,
        input = "SELECT * FROM foo RIGHT JOIN bar AS bar"
    );
    parsable!(
        two_right_join_first_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo RIGHT OUTER JOIN bar"
    );
    parsable!(
        two_right_join_both_alias,
        expected = true,
        input = "SELECT * FROM foo AS foo RIGHT OUTER JOIN bar AS bar"
    );
    parsable!(
        two_right_join,
        expected = true,
        input = "SELECT * FROM foo RIGHT JOIN bar"
    );

    parsable!(
        derived_with_alias,
        expected = true,
        input = "SELECT * FROM (SELECT * FROM foo) bar"
    );
    parsable!(
        derived_with_as_alias,
        expected = true,
        input = "SELECT * FROM (SELECT * FROM foo) AS bar"
    );
    parsable!(
        derived_must_have_alias,
        expected = false,
        input = "SELECT * FROM (SELECT * FROM foo)"
    );

    validate_ast!(
        comma_join_is_cross_join,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Join(JoinSource {
                join_type: JoinType::Cross,
                left: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "foo".to_string(),
                    alias: None
                })),
                right: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "bar".to_string(),
                    alias: None
                })),
                condition: None
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM foo, bar",
    );
    validate_ast!(
        cross_join_is_cross_join,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Join(JoinSource {
                join_type: JoinType::Cross,
                left: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "foo".to_string(),
                    alias: None
                })),
                right: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "bar".to_string(),
                    alias: None
                })),
                condition: None
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM foo CROSS JOIN bar",
    );
    validate_ast!(
        join_is_cross_join,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Join(JoinSource {
                join_type: JoinType::Cross,
                left: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "foo".to_string(),
                    alias: None
                })),
                right: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "bar".to_string(),
                    alias: None
                })),
                condition: None
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM foo JOIN bar",
    );
    validate_ast!(
        left_join_is_left_join,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Join(JoinSource {
                join_type: JoinType::Left,
                left: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "foo".to_string(),
                    alias: None
                })),
                right: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "bar".to_string(),
                    alias: None
                })),
                condition: None
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM foo LEFT JOIN bar",
    );
    validate_ast!(
        right_join_is_right_join,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Join(JoinSource {
                join_type: JoinType::Right,
                left: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "foo".to_string(),
                    alias: None
                })),
                right: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "bar".to_string(),
                    alias: None
                })),
                condition: None
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM foo RIGHT JOIN bar",
    );
    validate_ast!(
        join_is_left_associative,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Join(JoinSource {
                join_type: JoinType::Cross,
                left: Box::new(Datasource::Join(JoinSource {
                    join_type: JoinType::Cross,
                    left: Box::new(Datasource::Collection(CollectionSource {
                        database: None,
                        collection: "foo".to_string(),
                        alias: None,
                    })),
                    right: Box::new(Datasource::Collection(CollectionSource {
                        database: None,
                        collection: "bar".to_string(),
                        alias: None,
                    })),
                    condition: None
                })),
                right: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "car".to_string(),
                    alias: None
                })),
                condition: None
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM foo JOIN bar JOIN car",
    );

    parsable!(
        cannot_have_more_than_one_qualifier,
        expected = false,
        expected_error =
            "collection datasources can only have database qualification, found: car.bar.foo",
        input = "SELECT * FROM car.bar.foo"
    );
    parsable!(
        cannot_be_document,
        expected = false,
        expected_error = "found unsupported expression used as datasource: {'foo': 3 + 4}",
        input = "SELECT * FROM {'foo': 3+4}"
    );
    parsable!(
        cannot_be_literal,
        expected = false,
        expected_error = "found unsupported expression used as datasource: 3",
        input = "SELECT * FROM 3"
    );
    parsable!(
        cannot_be_binary_op,
        expected = false,
        expected_error = "found unsupported expression used as datasource: 3 + 4",
        input = "SELECT * FROM 3 + 4"
    );
    parsable!(
        array_must_have_alias,
        expected = false,
        expected_error = "array datasources must have aliases",
        input = "SELECT * FROM [{'a': 1}]"
    );
    parsable!(
        natural_join_not_allowed,
        expected = false,
        input = "SELECT * FROM foo NATURAL JOIN bar"
    );
}
mod where_test {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(
        single_condition,
        expected = true,
        input = "select * WHERE a >= 2"
    );
    parsable!(
        single_column_expr,
        expected = true,
        input = "select * WHERE a"
    );
    parsable!(
        multiple_conditions,
        expected = true,
        input = "select * WHERE a > 1 AND b > 1"
    );
    parsable!(
        case_expr,
        expected = true,
        input = "select * WHERE CASE WHEN a = expected = true THEN a ELSE expected = false END"
    );
    parsable!(null, expected = true, input = "select * WHERE NULL");

    validate_ast!(
        ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: None,
            where_clause: Some(Expression::Binary(BinaryExpr {
                left: Box::new(Expression::Identifier("a".to_string())),
                op: BinaryOp::Comparison(ComparisonOp::Gte),
                right: Box::new(Expression::Literal(Literal::Integer(2)))
            })),
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * WHERE a >= 2",
    );
}

mod type_conversion {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(
        cast_to_double,
        expected = true,
        input = "select CAST(v AS DOUBLE)"
    );
    parsable!(
        cast_to_double_precision,
        expected = true,
        input = "select CAST(v AS DOUBLE PRECISION)"
    );
    parsable!(
        cast_to_double_shorthand,
        expected = true,
        input = "select v::DOUBLE PRECISION"
    );
    parsable!(
        cast_to_real,
        expected = true,
        input = "select CAST(v AS REAL)"
    );
    parsable!(
        cast_to_real_shorthand,
        expected = true,
        input = "select v::REAL"
    );
    parsable!(
        cast_to_float,
        expected = true,
        input = "select CAST(v AS FLOAT)"
    );
    parsable!(
        cast_to_float_int,
        expected = true,
        input = "select CAST(v AS FLOAT(25))"
    );
    parsable!(
        cast_to_float_shorthand,
        expected = true,
        input = "select v::FLOAT(25)"
    );
    parsable!(
        cast_to_string,
        expected = true,
        input = "select CAST(v AS STRING)"
    );
    parsable!(
        cast_to_string_shorthand,
        expected = true,
        input = "select v::STRING"
    );
    parsable!(
        cast_to_varchar,
        expected = true,
        input = "select CAST(v AS VARCHAR)"
    );
    parsable!(
        cast_to_varchar_int,
        expected = true,
        input = "select CAST(v AS VARCHAR(1))"
    );
    parsable!(
        cast_to_varchar_shorthand,
        expected = true,
        input = "select v::VARCHAR(1)"
    );
    parsable!(
        cast_to_char,
        expected = true,
        input = "select CAST(v AS CHAR)"
    );
    parsable!(
        cast_to_char_int,
        expected = true,
        input = "select CAST(v AS CHAR(1))"
    );
    parsable!(
        cast_to_char_shorthand,
        expected = true,
        input = "select v::CHAR(1)"
    );
    parsable!(
        cast_to_character,
        expected = true,
        input = "select CAST(v AS CHARACTER)"
    );
    parsable!(
        cast_to_character_int,
        expected = true,
        input = "select CAST(v AS CHARACTER(1))"
    );
    parsable!(
        cast_to_character_shorthand,
        expected = true,
        input = "select v::CHARACTER(1)"
    );
    parsable!(
        cast_to_char_varying,
        expected = true,
        input = "select CAST(v AS CHAR VARYING)"
    );
    parsable!(
        cast_to_char_varying_int,
        expected = true,
        input = "select CAST(v AS CHAR VARYING(1))"
    );
    parsable!(
        cast_to_char_varying_shorthand,
        expected = true,
        input = "select v::CHAR VARYING(1)"
    );
    parsable!(
        cast_to_character_varying,
        expected = true,
        input = "select CAST(v AS CHARACTER VARYING)"
    );
    parsable!(
        cast_to_character_varying_int,
        expected = true,
        input = "select CAST(v AS CHARACTER VARYING(1))"
    );
    parsable!(
        cast_to_character_varying_shorthand,
        expected = true,
        input = "select v::CHARACTER VARYING(1)"
    );
    parsable!(
        cast_to_document,
        expected = true,
        input = "select CAST(v AS DOCUMENT)"
    );
    parsable!(
        cast_to_document_shorthand,
        expected = true,
        input = "select v::DOCUMENT"
    );
    parsable!(
        cast_to_array,
        expected = true,
        input = "select CAST(v AS ARRAY)"
    );
    parsable!(
        cast_to_array_shorthand,
        expected = true,
        input = "select v::ARRAY"
    );
    parsable!(
        cast_to_bindata,
        expected = true,
        input = "select CAST(v AS BINDATA)"
    );
    parsable!(
        cast_to_bindata_shorthand,
        expected = true,
        input = "select v::BINDATA"
    );
    parsable!(
        cast_to_undefined,
        expected = true,
        input = "select CAST(v AS UNDEFINED)"
    );
    parsable!(
        cast_to_undefined_shorthand,
        expected = true,
        input = "select v::UNDEFINED"
    );
    parsable!(
        cast_to_object_id,
        expected = true,
        input = "select CAST(v AS OBJECTID)"
    );
    parsable!(
        cast_to_object_id_shorthand,
        expected = true,
        input = "select v::OBJECTID"
    );
    parsable!(
        cast_to_bool,
        expected = true,
        input = "select CAST(v AS BOOL)"
    );
    parsable!(
        cast_to_bool_shorthand,
        expected = true,
        input = "select v::BOOL"
    );
    parsable!(
        cast_to_bit,
        expected = true,
        input = "select CAST(v AS BIT)"
    );
    parsable!(
        cast_to_bit_shorthand,
        expected = true,
        input = "select v::BIT"
    );
    parsable!(
        cast_to_boolean,
        expected = true,
        input = "select CAST(v AS BOOLEAN)"
    );
    parsable!(
        cast_to_boolean_shorthand,
        expected = true,
        input = "select v::BOOLEAN"
    );
    parsable!(
        cast_to_bson_date,
        expected = true,
        input = "select CAST(v AS BSON_DATE)"
    );
    parsable!(
        cast_to_bson_date_shorthand,
        expected = true,
        input = "select v::BSON_DATE"
    );
    parsable!(
        cast_to_timestamp,
        expected = true,
        input = "select CAST(v AS TIMESTAMP)"
    );
    parsable!(
        cast_to_timestamp_shorthand,
        expected = true,
        input = "select v::TIMESTAMP"
    );
    parsable!(
        cast_to_null,
        expected = true,
        input = "select CAST(v AS NULL)"
    );
    parsable!(
        cast_to_null_shorthand,
        expected = true,
        input = "select v::NULL"
    );
    parsable!(
        cast_to_regex,
        expected = true,
        input = "select CAST(v AS REGEX)"
    );
    parsable!(
        cast_to_regex_shorthand,
        expected = true,
        input = "select v::REGEX"
    );
    parsable!(
        cast_to_dbpointer,
        expected = true,
        input = "select CAST(v AS DBPOINTER)"
    );
    parsable!(
        cast_to_dbpointer_shorthand,
        expected = true,
        input = "select v::DBPOINTER"
    );
    parsable!(
        cast_to_javascript,
        expected = true,
        input = "select CAST(v AS JAVASCRIPT)"
    );
    parsable!(
        cast_to_javascript_shorthand,
        expected = true,
        input = "select v::JAVASCRIPT"
    );
    parsable!(
        cast_to_symbol,
        expected = true,
        input = "select CAST(v AS SYMBOL)"
    );
    parsable!(
        cast_to_symbol_shorthand,
        expected = true,
        input = "select v::SYMBOL"
    );
    parsable!(
        cast_to_javascriptwithscope,
        expected = true,
        input = "select CAST(v AS JAVASCRIPTWITHSCOPE)"
    );
    parsable!(
        cast_to_javascriptwithscope_shorthand,
        expected = true,
        input = "select v::JAVASCRIPTWITHSCOPE"
    );
    parsable!(
        cast_to_int,
        expected = true,
        input = "select CAST(v AS INT)"
    );
    parsable!(
        cast_to_int_shorthand,
        expected = true,
        input = "select v::INT"
    );
    parsable!(
        cast_to_integer,
        expected = true,
        input = "select CAST(v AS INTEGER)"
    );
    parsable!(
        cast_to_integer_shorthand,
        expected = true,
        input = "select v::INTEGER"
    );
    parsable!(
        cast_to_small_int,
        expected = true,
        input = "select CAST(v AS SMALLINT)"
    );
    parsable!(
        cast_to_small_int_shorthand,
        expected = true,
        input = "select v::SMALLINT"
    );
    parsable!(
        cast_to_bson_timestamp,
        expected = true,
        input = "select CAST(v AS BSON_TIMESTAMP)"
    );
    parsable!(
        cast_to_bson_timestamp_shorthand,
        expected = true,
        input = "select v::BSON_TIMESTAMP"
    );
    parsable!(
        cast_to_long,
        expected = true,
        input = "select CAST(v AS LONG)"
    );
    parsable!(
        cast_to_long_shorthand,
        expected = true,
        input = "select v::LONG"
    );
    parsable!(
        cast_to_decimal,
        expected = true,
        input = "select CAST(v AS DECIMAL)"
    );
    parsable!(
        cast_to_decimal_int,
        expected = true,
        input = "select CAST(v AS DECIMAL(1))"
    );
    parsable!(
        cast_to_decimal_shorthand,
        expected = true,
        input = "select v::DECIMAL(1)"
    );
    parsable!(
        cast_to_dec,
        expected = true,
        input = "select CAST(v AS DEC)"
    );
    parsable!(
        cast_to_dec_int,
        expected = true,
        input = "select CAST(v AS DEC(1))"
    );
    parsable!(
        cast_to_dec_shorthand,
        expected = true,
        input = "select v::DEC"
    );
    parsable!(
        cast_to_numeric,
        expected = true,
        input = "select CAST(v AS NUMERIC)"
    );
    parsable!(
        cast_to_numeric_int,
        expected = true,
        input = "select CAST(v AS NUMERIC(1))"
    );
    parsable!(
        cast_to_numeric_shorthand,
        expected = true,
        input = "select v::NUMERIC"
    );
    parsable!(
        cast_to_minkey,
        expected = true,
        input = "select CAST(v AS MINKEY)"
    );
    parsable!(
        cast_to_minkey_shorthand,
        expected = true,
        input = "select v::MINKEY"
    );
    parsable!(
        cast_to_maxkey,
        expected = true,
        input = "select CAST(v AS MAXKEY)"
    );
    parsable!(
        cast_to_maxkey_shorthand,
        expected = true,
        input = "select v::MAXKEY"
    );
    parsable!(
        cast_on_null,
        expected = true,
        input = "select CAST(v AS BOOL, 'null' ON NULL)"
    );
    parsable!(
        cast_on_error,
        expected = true,
        input = "select CAST(v AS INT, 'null' ON ERROR)"
    );
    parsable!(
        cast_on_null_on_error,
        expected = true,
        input = "select CAST(v AS INT, 'null' ON NULL, 'error' ON ERROR)"
    );
    parsable!(type_assert, expected = true, input = "select a::!STRING");
    parsable!(
        type_assert_in_func,
        expected = true,
        input = "select SUBSTRING(foo::!STRING, 1, 1)"
    );

    validate_ast!(
        cast_to_decimal_ast,
        method = parse_expression,
        expected = Expression::Cast(CastExpr {
            expr: Box::new(Expression::Identifier("v".to_string())),
            to: Type::Decimal128,
            on_null: Some(Box::new(Expression::Literal(Literal::String(
                "null".to_string()
            )))),
            on_error: Some(Box::new(Expression::Literal(Literal::String(
                "error".to_string()
            )))),
        }),
        input = "CAST(v AS DECIMAL(1), 'null' ON NULL, 'error' ON ERROR)",
    );
    validate_ast!(
        cast_precedence_binary,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier("a".to_string())),
            op: BinaryOp::Mul,
            right: Box::new(Expression::Cast(CastExpr {
                expr: Box::new(Expression::Identifier("b".to_string())),
                to: Type::Int32,
                on_null: None,
                on_error: None,
            }))
        }),
        input = "a * b::int",
    );
    validate_ast!(
        cast_precedence_unary,
        method = parse_expression,
        expected = Expression::Unary(UnaryExpr {
            op: UnaryOp::Not,
            expr: Box::new(Expression::Cast(CastExpr {
                expr: Box::new(Expression::Identifier("a".to_string())),
                to: Type::Boolean,
                on_null: None,
                on_error: None,
            }))
        }),
        input = "NOT a::bool",
    );
}

mod subquery {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(
        simple_subquery,
        expected = true,
        input = "SELECT VALUE (SELECT a)"
    );
    parsable!(
        multiple_nested_subqueries,
        expected = true,
        input = "SELECT (SELECT a)"
    );
    parsable!(
        exists_subquery,
        expected = true,
        input = "SELECT EXISTS (SELECT a)"
    );
    parsable!(
        not_exists_subquery,
        expected = true,
        input = "SELECT NOT EXISTS (SELECT a)"
    );
    parsable!(
        any_subquery,
        expected = true,
        input = "SELECT x <> ANY (SELECT a)"
    );
    parsable!(
        all_subquery,
        expected = true,
        input = "SELECT x = ALL (SELECT a)"
    );
    parsable!(
        in_tuple_subquery,
        expected = true,
        input = "SELECT X IN (A, B, C)"
    );
    parsable!(
        not_in_tuple_subquery,
        expected = true,
        input = "SELECT X NOT IN (A, B, C)"
    );

    parsable!(in_test, expected = true, input = "SELECT X NOT IN (1+2-3)");

    parsable!(empty_tuple, expected = false, input = "SELECT X NOT IN ()");
    parsable!(
        tuple_with_dangling_comma,
        expected = false,
        input = "SELECT X NOT IN (A,)"
    );

    validate_ast!(
        some_subquery,
        method = parse_expression,
        expected = Expression::SubqueryComparison(SubqueryComparisonExpr {
            expr: Box::new(Expression::Identifier("x".to_string())),
            op: ComparisonOp::Neq,
            quantifier: SubqueryQuantifier::Any,
            subquery: Box::new(Query::Select(SelectQuery {
                select_clause: SelectClause {
                    set_quantifier: SetQuantifier::All,
                    body: SelectBody::Standard(vec![SelectExpression::Expression(
                        OptionallyAliasedExpr::Unaliased(Expression::Identifier("a".to_string()),)
                    )])
                },
                from_clause: None,
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None
            }))
        }),
        input = "x <> SOME (SELECT a)",
    );

    validate_ast!(
        in_subquery,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier("x".to_string())),
            op: BinaryOp::In,
            right: Box::new(Expression::Subquery(Box::new(Query::Select(SelectQuery {
                select_clause: SelectClause {
                    set_quantifier: SetQuantifier::All,
                    body: SelectBody::Standard(vec![SelectExpression::Expression(
                        OptionallyAliasedExpr::Unaliased(Expression::Identifier("a".to_string()),)
                    )])
                },
                from_clause: None,
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None
            }))))
        }),
        input = "x IN (SELECT a)",
    );

    validate_ast!(
        not_in_subquery,
        method = parse_expression,
        expected = Expression::Binary(BinaryExpr {
            left: Box::new(Expression::Identifier("x".to_string())),
            op: BinaryOp::NotIn,
            right: Box::new(Expression::Subquery(Box::new(Query::Select(SelectQuery {
                select_clause: SelectClause {
                    set_quantifier: SetQuantifier::All,
                    body: SelectBody::Standard(vec![SelectExpression::Expression(
                        OptionallyAliasedExpr::Unaliased(Expression::Identifier("a".to_string()),)
                    )])
                },
                from_clause: None,
                where_clause: None,
                group_by_clause: None,
                having_clause: None,
                order_by_clause: None,
                limit: None,
                offset: None
            }))))
        }),
        input = "x NOT IN (SELECT a)",
    );
}

mod document {
    use crate::{
        ast::*,
        multimap,
        parser::{Error, Parser},
    };

    parsable!(empty_doc_literal, expected = true, input = "select {}");
    parsable!(
        doc_literal,
        expected = true,
        input = "select {'a':1, 'b':2}"
    );
    parsable!(
        doc_mixed_field_binary_op,
        expected = true,
        input = "select {'a':3+4}"
    );
    parsable!(
        doc_field_access_bracket,
        expected = true,
        input = "select doc['a']"
    );
    parsable!(
        doc_literal_field_access_dot,
        expected = true,
        input = "select {'a': 1}.a"
    );
    parsable!(
        doc_literal_field_access_multi_level_dot,
        expected = true,
        input = "select {'a': {'b': {'c': 100}}}.a.b.c"
    );
    parsable!(
        doc_literal_field_dot_star_field,
        expected = true,
        input = "select {'a': {'*': 100, 'b': 10, 'c': 1}}.a.`*`"
    );
    parsable!(
        doc_literal_field_access_bracket,
        expected = true,
        input = "select {'a': 1, 'b': 2}['a']"
    );
    parsable!(
        doc_literal_field_access_bracket_one_level,
        expected = true,
        input = "select a['b']"
    );
    parsable!(
        doc_literal_field_access_bracket_multi_level,
        expected = true,
        input = "select a['b']['c']"
    );
    parsable!(
        doc_literal_field_bracket_star_field,
        expected = true,
        input = "select {'a': {'*': 100, 'b': 10, 'c': 1}}['a']['*']"
    );

    parsable!(
        doc_literal_non_string_keys,
        expected = false,
        input = "select {1:1, 2:2}"
    );
    parsable!(non_doc_field_access, expected = false, input = "select 1.a");

    validate_ast!(
        doc_literal_field_access_ast,
        method = parse_expression,
        expected = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Subpath(SubpathExpr {
                    expr: Box::new(Expression::Document(
                        multimap! {"a".to_string() => Expression::Document(
                            multimap!{"b".to_string() => Expression::Document(
                                multimap!{"c".to_string() => Expression::Literal(Literal::Integer(100))}
                            )}
                        )}
                    )),
                    subpath: "a".to_string()
                })),
                subpath: "b".to_string()
            })),
            subpath: "c".to_string()
        }),
        input = "{'a': {'b': {'c': 100}}}.a.b.c",
    );
    validate_ast!(
        doc_mixed_field_access,
        method = parse_expression,
        expected = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Access(AccessExpr {
                expr: Box::new(Expression::Subpath(SubpathExpr {
                    expr: Box::new(Expression::Identifier("a".to_string())),
                    subpath: "b".to_string()
                })),
                subfield: Box::new(Expression::Literal(Literal::String("c".to_string()))),
            })),
            subpath: "d".to_string()
        }),
        input = "a.b['c'].d",
    );
}

mod comments {
    use crate::ast::*;
    use crate::parser::{Error, Parser};
    parsable!(
        standard_parse,
        expected = true,
        input = "SELECT a FROM foo -- This is a standard comment"
    );
    parsable!(
        standard_conflict_with_minus_minus,
        expected = true,
        input = "SELECT a -- b"
    );
    parsable!(
        standard_single_line_parse,
        expected = true,
        input = "-- This is a standard single line comment
    SELECT a FROM foo"
    );
    parsable!(
        inline_parse,
        expected = true,
        input = "SELECT a /* This is an inline comment */ FROM foo"
    );
    parsable!(
        multiline_parse,
        expected = true,
        input = "/* This is a multiline comment
    This is a multiline comment */
    SELECT a FROM foo"
    );
    parsable!(
        multiline_inline_parse,
        expected = true,
        input = "SELECT a /* This is an inline
    comment */ FROM foo"
    );
    parsable!(
        multiline_nesting,
        expected = true,
        input = "/* This is a multiline comment
    * with nesting: /* nested block comment */
    */
    SELECT a FROM foo"
    );

    validate_ast!(
        standard_ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT foo -- This is a standard comment",
    );

    validate_ast!(
        ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "-- This is a standard single line comment
    SELECT foo",
    );

    validate_ast!(
        inline_ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT /* This is an inline comment */ foo",
    );

    validate_ast!(
        multiline_ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "/* This is a multiline comment
    This is a multiline comment */
    SELECT foo",
    );

    validate_ast!(
        inline_multiline_ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT /* This is an inline
    comment */ foo",
    );

    validate_ast!(
        multiline_nesting_ast,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Expression(
                    OptionallyAliasedExpr::Unaliased(Expression::Identifier("foo".to_string()),)
                )])
            },
            from_clause: None,
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "/* This is a multiline comment
    * with nesting: /* nested block comment */
    */
    SELECT foo",
    );
}

mod flatten {
    use crate::{
        ast::*,
        parser::{Error, Parser},
    };
    parsable!(
        standard_parse,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo)"
    );
    parsable!(
        datasource_explicitly_aliased,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo as f)"
    );
    parsable!(
        depth,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo, depth => 1)"
    );
    parsable!(
        separator,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo, separator => '%')"
    );
    parsable!(
        depth_and_separator,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo, depth => 1, separator => '%')"
    );
    parsable!(
        separator_and_depth,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo, separator => '%', depth => 1)"
    );
    parsable!(
        separator_len_gt_1,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo, separator => 'hello')"
    );
    parsable!(
        array_datasource,
        expected = true,
        input = "SELECT * FROM FLATTEN([{'a': 1}] as arr)"
    );
    parsable!(
        join_datasource,
        expected = true,
        input = "SELECT * FROM FLATTEN(foo JOIN bar)"
    );
    parsable!(
        derived_datasource,
        expected = true,
        input = "SELECT * FROM FLATTEN((SELECT * FROM foo) as derived)"
    );
    parsable!(
        flatten_datasource,
        expected = true,
        input = "SELECT * FROM FLATTEN(FLATTEN(foo))"
    );
    parsable!(
        unwind_datasource,
        expected = true,
        input = "SELECT * FROM FLATTEN(UNWIND(foo, PATH => arr))"
    );
    parsable!(
        depth_neg,
        expected = false,
        input = "SELECT * FROM FLATTEN(foo, depth => -1)"
    );
    parsable!(
        depth_not_int,
        expected = false,
        input = "SELECT * FROM FLATTEN(foo, depth => 1.2)"
    );
    parsable!(
        no_datasource,
        expected = false,
        input = "SELECT * FROM FLATTEN()"
    );
    parsable!(
        missing_comma,
        expected = false,
        input = "SELECT * FROM FLATTEN(foo depth => 1)"
    );
    parsable!(
        extra_comma,
        expected = false,
        input = "SELECT * FROM FLATTEN(foo,)"
    );
    validate_ast!(
        duplicate_options,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Flatten(FlattenSource {
                datasource: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "foo".to_string(),
                    alias: None
                })),
                options: vec![
                    FlattenOption::Depth(1),
                    FlattenOption::Separator('%'.to_string()),
                    FlattenOption::Depth(2)
                ]
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM FLATTEN(foo, depth => 1, separator => '%', depth => 2)",
    );
}

mod unwind {
    use crate::{
        ast::*,
        parser::{Error, Parser},
    };
    // Note it is syntactically valid to omit PATH even though that is semantically invalid
    parsable!(
        standard_parse,
        expected = true,
        input = "SELECT * FROM UNWIND(foo)"
    );
    parsable!(
        datasource_explicitly_aliased,
        expected = true,
        input = "SELECT * FROM UNWIND(foo as f)"
    );
    parsable!(
        path,
        expected = true,
        input = "SELECT * FROM UNWIND(foo, PATH => arr)"
    );
    parsable!(
        index,
        expected = true,
        input = "SELECT * FROM UNWIND(foo, PATH => arr, INDEX => i)"
    );
    parsable!(
        outer,
        expected = true,
        input = "SELECT * FROM UNWIND(foo, PATH => arr, OUTER => true)"
    );
    parsable!(
        path_and_index_and_outer,
        expected = true,
        input = "SELECT * FROM UNWIND(foo, PATH => arr, INDEX => i, OUTER => true)"
    );
    parsable!(
        outer_and_path_and_index,
        expected = true,
        input = "SELECT * FROM UNWIND(foo, OUTER => true, PATH => arr, INDEX => i)"
    );
    parsable!(
        index_and_outer_and_path,
        expected = true,
        input = "SELECT * FROM UNWIND(foo, INDEX => i, OUTER => true, PATH => arr)"
    );
    parsable!(
        multi_part_path,
        expected = true,
        input = "SELECT * FROM UNWIND(foo, PATH => a.b.c)"
    );
    parsable!(
        array_datasource,
        expected = true,
        input = "SELECT * FROM UNWIND([{'a': [1]}] AS arr, PATH => a)"
    );
    parsable!(
        join_datasource,
        expected = true,
        input = "SELECT * FROM UNWIND(foo JOIN bar, PATH => bar.arr)"
    );
    parsable!(
        derived_datasource,
        expected = true,
        input = "SELECT * FROM UNWIND((SELECT * FROM foo) AS derived, PATH => arr)"
    );
    parsable!(
        flatten_datasource,
        expected = true,
        input = "SELECT * FROM UNWIND(FLATTEN(foo), PATH => arr)"
    );
    parsable!(
        unwind_datasource,
        expected = true,
        input = "SELECT * FROM UNWIND(UNWIND(foo, PATH => arr), PATH => arr)"
    );
    parsable!(
        path_not_ident,
        expected = false,
        input = "SELECT * FROM UNWIND(foo, PATH => 'arr')"
    );
    parsable!(
        index_not_ident,
        expected = false,
        input = "SELECT * FROM UNWIND(foo, INDEX => 'i')"
    );
    parsable!(
        outer_not_bool,
        expected = false,
        input = "SELECT * FROM UNWIND(foo, OUTER => 1)"
    );
    parsable!(
        no_datasource_or_options,
        expected = false,
        input = "SELECT * FROM UNWIND()"
    );
    parsable!(
        missing_comma,
        expected = false,
        input = "SELECT * FROM UNWIND(foo PATH => arr)"
    );
    parsable!(
        extra_comma,
        expected = false,
        input = "SELECT * FROM UNWIND(foo,)"
    );
    validate_ast!(
        duplicate_options,
        method = parse_query,
        expected = Query::Select(SelectQuery {
            select_clause: SelectClause {
                set_quantifier: SetQuantifier::All,
                body: SelectBody::Standard(vec![SelectExpression::Star])
            },
            from_clause: Some(Datasource::Unwind(UnwindSource {
                datasource: Box::new(Datasource::Collection(CollectionSource {
                    database: None,
                    collection: "foo".to_string(),
                    alias: None
                })),
                options: vec![
                    UnwindOption::Path(Expression::Identifier("arr".into())),
                    UnwindOption::Index("i".into()),
                    UnwindOption::Index("idx".into()),
                    UnwindOption::Path(Expression::Identifier("a".into())),
                    UnwindOption::Outer(false),
                ]
            })),
            where_clause: None,
            group_by_clause: None,
            having_clause: None,
            order_by_clause: None,
            limit: None,
            offset: None,
        }),
        input = "SELECT * FROM UNWIND(foo, PATH => arr, INDEX => i, INDEX => idx, PATH => a, OUTER => false)",
    );
}
