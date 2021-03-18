use crate::parser::ast::*;
use crate::parser::Parser;

macro_rules! should_parse {
    ($func_name:ident, $should_parse:expr, $input:expr) => {
        #[test]
        fn $func_name() {
            let p = Parser::new();
            let res = p.parse_query($input);
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
should_parse!(select_numerical_ident_prefix, false, "SELECT 9a");
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
                key: SortKey::Key(Identifier::Simple("a".to_string())),
                direction: SortDirection::Asc
            }]
        })
    })
);
