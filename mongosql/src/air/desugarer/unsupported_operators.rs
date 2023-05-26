use crate::air::{
    self,
    desugarer::{Error, Pass, Result},
    util::sql_op_to_mql_op,
    visitor::Visitor,
    Expression, Is, Let, LetVariable, Like, LiteralValue, MQLOperator, MQLSemanticOperator,
    RegexMatch, SQLOperator, SQLSemanticOperator, SqlConvert, SqlDivide, Stage, Switch, SwitchCase,
    Type,
};

use crate::make_cond_expr;

/// Desugars any SQL operators that do not exist in MQL (e.g. Between, Like,
/// etc.) or that do not have the same semantics in MQL (e.g. Cos, Sin, Tan,
/// etc.) into appropriate, equivalent MQL operators.
pub struct UnsupportedOperatorsDesugarerPass;

impl Pass for UnsupportedOperatorsDesugarerPass {
    fn apply(&self, pipeline: Stage) -> Result<Stage> {
        let visitor = &mut UnsupportedOperatorsDesugarerVisitor { error: None };
        let stage = pipeline.walk(visitor);
        match visitor.error.clone() {
            Some(e) => Err(e),
            None => Ok(stage),
        }
    }
}
#[derive(Default)]
struct UnsupportedOperatorsDesugarerVisitor {
    error: Option<Error>,
}

const NAN_LITERAL: Expression = Expression::Literal(LiteralValue::Double(f64::NAN));
const INFINITY_LITERAL: Expression = Expression::Literal(LiteralValue::Double(f64::INFINITY));
const NEG_INFINITY_LITERAL: Expression =
    Expression::Literal(LiteralValue::Double(f64::NEG_INFINITY));
const NULL_LITERAL: Expression = Expression::Literal(LiteralValue::Null);
const NEG_TWENTY_LITERAL: Expression = Expression::Literal(LiteralValue::Integer(-20));
const NEG_ONE_LITERAL: Expression = Expression::Literal(LiteralValue::Integer(-1));
const ZERO_LITERAL: Expression = Expression::Literal(LiteralValue::Integer(0));
const ONE_LITERAL: Expression = Expression::Literal(LiteralValue::Integer(1));
const EIGHT_LITERAL: Expression = Expression::Literal(LiteralValue::Integer(8));
const ONE_HUNDRED_LITERAL: Expression = Expression::Literal(LiteralValue::Integer(100));

impl UnsupportedOperatorsDesugarerVisitor {
    fn convert_sql_pattern(pattern: String, escape: String) -> String {
        const REGEX_CHARS_TO_ESCAPE: [char; 12] =
            ['.', '^', '$', '*', '+', '?', '(', ')', '[', '{', '\\', '|'];
        let mut regex = "^".to_string();
        let mut escaped = false;
        for c in pattern.chars() {
            let c_string = c.to_string();
            if !escaped & (c_string == escape) {
                escaped = true;
                continue;
            }
            match c {
                '_' => {
                    let s = if escaped { '_' } else { '.' };
                    regex.push(s)
                }
                '%' => {
                    if escaped {
                        regex.push('%');
                    } else {
                        regex.push_str(".*");
                    }
                }
                _ => {
                    if REGEX_CHARS_TO_ESCAPE.contains(&c) {
                        regex.push('\\');
                    }
                    regex.push_str(c.to_string().as_str())
                }
            }
            escaped = false;
        }
        regex.push('$');
        regex.to_string()
    }

    fn wrap_in_infinity_check(&self, arg: Expression) -> Expression {
        let pos_inf_check = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![arg.clone(), INFINITY_LITERAL],
        });
        let neg_inf_check = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![arg, NEG_INFINITY_LITERAL],
        });
        Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Or,
            args: vec![pos_inf_check, neg_inf_check],
        })
    }

    fn desugar_trig_function(&mut self, trig_function: SQLSemanticOperator) -> Expression {
        let is_trig_arg_inf = self.wrap_in_infinity_check(trig_function.args[0].clone());
        let mql_op = sql_op_to_mql_op(trig_function.op).unwrap();
        let trig_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: mql_op,
            args: trig_function.args,
        });

        make_cond_expr!(is_trig_arg_inf, NULL_LITERAL, trig_expr)
    }

    fn desugar_like(&mut self, like: Like) -> Expression {
        let pattern = match *like.pattern {
            Expression::Literal(LiteralValue::String(l)) => {
                if let Some(e) = like.escape {
                    Self::convert_sql_pattern(l, e)
                } else {
                    format!("^{l}$")
                }
            }
            _ => {
                self.error = Some(Error::InvalidLikePattern);
                return Expression::Like(like);
            }
        };
        Expression::RegexMatch(RegexMatch {
            input: Box::new(*like.expr),
            regex: Box::new(Expression::Literal(LiteralValue::String(pattern))),
            options: Some(Box::new(Expression::Literal(LiteralValue::String(
                "is".to_string(),
            )))),
        })
    }

    fn desugar_is(&self, is: Is) -> Expression {
        match is.target_type {
            air::TypeOrMissing::Number => Expression::MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::IsNumber,
                args: vec![*is.expr],
            }),
            air::TypeOrMissing::Type(Type::Null) => {
                Expression::MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Or,
                    args: vec![
                        Expression::MQLSemanticOperator(MQLSemanticOperator {
                            op: MQLOperator::Eq,
                            args: vec![
                                Expression::MQLSemanticOperator(MQLSemanticOperator {
                                    op: MQLOperator::Type,
                                    args: vec![*is.expr.clone()],
                                }),
                                Expression::Literal(LiteralValue::String("null".to_string())),
                            ],
                        }),
                        Expression::MQLSemanticOperator(MQLSemanticOperator {
                            op: MQLOperator::Eq,
                            args: vec![
                                Expression::MQLSemanticOperator(MQLSemanticOperator {
                                    op: MQLOperator::Type,
                                    args: vec![*is.expr.clone()],
                                }),
                                Expression::Literal(LiteralValue::String("missing".to_string())),
                            ],
                        }),
                    ],
                })
            }
            air::TypeOrMissing::Type(Type::Array) => {
                Expression::MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::IsArray,
                    args: vec![*is.expr],
                })
            }
            t => Expression::MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Eq,
                args: vec![
                    Expression::MQLSemanticOperator(MQLSemanticOperator {
                        op: MQLOperator::Type,
                        args: vec![*is.expr.clone()],
                    }),
                    Expression::Literal(LiteralValue::String(String::from(t.to_str()))),
                ],
            }),
        }
    }

    fn desugar_sql_between(&self, between: SQLSemanticOperator) -> Expression {
        let input_var_name = "desugared_sqlBetween_input".to_string();
        let input_var_ref = Expression::Variable(input_var_name.clone().into());
        let let_vars = vec![LetVariable {
            name: input_var_name,
            expr: Box::new(between.args[0].clone()),
        }];
        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(Expression::SQLSemanticOperator(SQLSemanticOperator {
                op: SQLOperator::And,
                args: vec![
                    Expression::SQLSemanticOperator(SQLSemanticOperator {
                        op: SQLOperator::Gte,
                        args: vec![input_var_ref.clone(), between.args[1].clone()],
                    }),
                    Expression::SQLSemanticOperator(SQLSemanticOperator {
                        op: SQLOperator::Lte,
                        args: vec![input_var_ref, between.args[2].clone()],
                    }),
                ],
            })),
        })
    }

    fn desugar_mql_between(&self, between: MQLSemanticOperator) -> Expression {
        let input_var_name = "desugared_mqlBetween_input".to_string();
        let input_var_ref = Expression::Variable(input_var_name.clone().into());
        let let_vars = vec![LetVariable {
            name: input_var_name,
            expr: Box::new(between.args[0].clone()),
        }];
        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(Expression::MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::And,
                args: vec![
                    Expression::MQLSemanticOperator(MQLSemanticOperator {
                        op: MQLOperator::Gte,
                        args: vec![input_var_ref.clone(), between.args[1].clone()],
                    }),
                    Expression::MQLSemanticOperator(MQLSemanticOperator {
                        op: MQLOperator::Lte,
                        args: vec![input_var_ref, between.args[2].clone()],
                    }),
                ],
            })),
        })
    }

    fn desugar_sql_bit_length(&self, bit_length: SQLSemanticOperator) -> Expression {
        Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Multiply,
            args: vec![
                Expression::SQLSemanticOperator(SQLSemanticOperator {
                    op: SQLOperator::StrLenBytes,
                    args: bit_length.args,
                }),
                EIGHT_LITERAL,
            ],
        })
    }

    fn desugar_sql_coalesce(&self, coalesce: SQLSemanticOperator) -> Expression {
        let cases = coalesce
            .args
            .iter()
            .map(|branch| SwitchCase {
                case: Box::new(Expression::MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Gt,
                    args: vec![branch.clone(), NULL_LITERAL],
                })),
                then: Box::new(branch.clone()),
            })
            .collect();
        Expression::Switch(Switch {
            branches: cases,
            default: Box::new(NULL_LITERAL),
        })
    }

    fn desugar_sql_convert(&self, convert: SqlConvert) -> Expression {
        let input_var_name = "sqlConvert_input";
        let input_var_ref = Expression::Variable(input_var_name.to_string().into());

        let case_input_is_of_type = SwitchCase {
            case: Box::new(Expression::MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Eq,
                args: vec![
                    Expression::MQLSemanticOperator(MQLSemanticOperator {
                        op: MQLOperator::Type,
                        args: vec![input_var_ref.clone()],
                    }),
                    Expression::Literal(LiteralValue::String(convert.to.to_str().to_string())),
                ],
            })),
            then: Box::new(input_var_ref.clone()),
        };
        let case_input_is_null = SwitchCase {
            case: Box::new(Expression::MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Lte,
                args: vec![input_var_ref, NULL_LITERAL],
            })),
            then: convert.on_null,
        };

        let let_vars = vec![LetVariable {
            name: input_var_name.to_string(),
            expr: convert.input,
        }];
        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(Expression::Switch(Switch {
                branches: vec![case_input_is_of_type, case_input_is_null],
                default: convert.on_error,
            })),
        })
    }

    fn desugar_sql_divide(&self, divide: SqlDivide) -> Expression {
        let divisor_not_zero_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![*divide.divisor.clone(), ZERO_LITERAL],
        });
        let division_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Divide,
            args: vec![*divide.dividend, *divide.divisor],
        });

        make_cond_expr!(divisor_not_zero_expr, *divide.on_error, division_expr)
    }

    fn desugar_sql_log(&self, log: SQLSemanticOperator) -> Expression {
        let first_arg_nan = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![log.args[0].clone(), NAN_LITERAL],
        });
        let second_arg_nan = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![log.args[1].clone(), NAN_LITERAL],
        });
        let first_arg_negative = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Lte,
            args: vec![log.args[0].clone(), ZERO_LITERAL],
        });
        let second_arg_one = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![log.args[1].clone(), ONE_LITERAL],
        });
        let second_arg_negative = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Lte,
            args: vec![log.args[1].clone(), ZERO_LITERAL],
        });

        let nan_check_condition = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Or,
            args: vec![first_arg_nan, second_arg_nan],
        });
        let invalid_arg_condition = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Or,
            args: vec![first_arg_negative, second_arg_one, second_arg_negative],
        });
        let log_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Log,
            args: log.args,
        });

        let invalid_arg_conditional_with_log_expr =
            make_cond_expr!(invalid_arg_condition, NULL_LITERAL, log_expr);

        make_cond_expr!(
            nan_check_condition,
            NAN_LITERAL,
            invalid_arg_conditional_with_log_expr
        )
    }

    fn desugar_sql_mod(&self, mod_operator: SQLSemanticOperator) -> Expression {
        let dividend_var_name = "desugared_sqlMod_input0".to_string();
        let dividend_var_ref = Expression::Variable(dividend_var_name.clone().into());
        let divisor_var_name = "desugared_sqlMod_input1".to_string();
        let divisor_var_ref = Expression::Variable(divisor_var_name.clone().into());
        let divisor_is_zero_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![divisor_var_ref.clone(), ZERO_LITERAL],
        });
        let mod_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Mod,
            args: vec![dividend_var_ref, divisor_var_ref],
        });
        let mod_expr_conditional = make_cond_expr!(divisor_is_zero_expr, NULL_LITERAL, mod_expr);
        let let_vars = vec![
            LetVariable {
                name: dividend_var_name,
                expr: Box::new(mod_operator.args[0].clone()),
            },
            LetVariable {
                name: divisor_var_name,
                expr: Box::new(mod_operator.args[1].clone()),
            },
        ];

        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(mod_expr_conditional),
        })
    }

    fn desugar_sql_neg(&self, neg: SQLSemanticOperator) -> Expression {
        Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Multiply,
            args: vec![neg.args[0].clone(), NEG_ONE_LITERAL],
        })
    }

    fn desugar_sql_nullif(&self, nullif: SQLSemanticOperator) -> Expression {
        let expr_var_name = "expr1".to_string();
        let expr_var_ref = Expression::Variable(expr_var_name.clone().into());
        let nullif_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![expr_var_ref.clone(), nullif.args[1].clone()],
        });
        let let_vars = vec![LetVariable {
            name: expr_var_name,
            expr: Box::new(Expression::MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::IfNull,
                args: vec![nullif.args[0].clone(), NULL_LITERAL],
            })),
        }];
        let nullif_conditional_with_expr = make_cond_expr!(nullif_expr, NULL_LITERAL, expr_var_ref);

        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(nullif_conditional_with_expr),
        })
    }

    fn desugar_sql_pos(&self, pos: SQLSemanticOperator) -> Expression {
        pos.args[0].clone()
    }

    fn desugar_sql_round(&self, round: SQLSemanticOperator) -> Expression {
        let input_number_var_name = "desugared_sqlRound_input0".to_string();
        let input_place_var_name = "desugared_sqlRound_input1".to_string();
        let input_number_var_ref = Expression::Variable(input_number_var_name.clone().into());
        let input_place_var_ref = Expression::Variable(input_place_var_name.clone().into());
        let arg_is_nan = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![input_place_var_ref.clone(), NAN_LITERAL],
        });
        let input_place_within_range_check = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::And,
            args: vec![
                Expression::MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Gte,
                    args: vec![input_place_var_ref.clone(), NEG_TWENTY_LITERAL],
                }),
                Expression::MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Lte,
                    args: vec![input_place_var_ref.clone(), ONE_HUNDRED_LITERAL],
                }),
            ],
        });
        let round_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Round,
            args: vec![input_number_var_ref, input_place_var_ref],
        });
        let round_with_range_check_conditional =
            make_cond_expr!(input_place_within_range_check, round_expr, NULL_LITERAL);
        let round_if_valid_conditional =
            make_cond_expr!(arg_is_nan, NAN_LITERAL, round_with_range_check_conditional);
        let let_vars = vec![
            LetVariable {
                name: input_number_var_name,
                expr: Box::new(round.args[0].clone()),
            },
            LetVariable {
                name: input_place_var_name,
                expr: Box::new(round.args[1].clone()),
            },
        ];

        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(round_if_valid_conditional),
        })
    }

    fn desugar_sql_slice(&self, slice: SQLSemanticOperator) -> Expression {
        let slice_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Slice,
            args: slice.args.clone(),
        });

        // when there are three elements, they represent $slice: <number to skip, number to return>
        if slice.args.len() == 3 {
            let number_to_return_lte_zero_check =
                Expression::MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Lte,
                    args: vec![slice.args[2].clone(), ZERO_LITERAL],
                });
            make_cond_expr!(number_to_return_lte_zero_check, NULL_LITERAL, slice_expr)
        // otherwise, $slice: <number of elements to return>
        } else {
            slice_expr
        }
    }

    fn desugar_sql_split(&self, split: SQLSemanticOperator) -> Expression {
        let input_str_var_name = "desugared_sqlSplit_input0";
        let input_str_var_ref = Expression::Variable(input_str_var_name.to_string().into());
        let input_delim_var_name = "desugared_sqlSplit_input1";
        let input_delim_var_ref = Expression::Variable(input_delim_var_name.to_string().into());
        let input_token_num_var_name = "desugared_sqlSplit_input2";
        let input_token_num_var_ref =
            Expression::Variable(input_token_num_var_name.to_string().into());
        let split_expr_var_name = "desugared_sqlSplit_split_expr";
        let split_expr_var_ref = Expression::Variable(split_expr_var_name.to_string().into());
        let slice_expr_var_name = "desugared_sqlSplit_slice_expr";
        let slice_expr_var_ref = Expression::Variable(slice_expr_var_name.to_string().into());

        let split_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Split,
            args: vec![input_str_var_ref, input_delim_var_ref.clone()],
        });

        // If a token number's absolute value is greater than the length of the split
        // array, set the token number to that absolute value. This is done to
        // circumvent MongoDB's default behavior of setting the starting position of
        // $slice to the beginning of the array even when the absolute value of the
        // given position is larger than the array itself.
        let abs_token_num_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Abs,
            args: vec![input_token_num_var_ref.clone()],
        });
        let size_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Size,
            args: vec![split_expr_var_ref.clone()],
        });
        let abs_token_num_gt_size_check = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Gt,
            args: vec![abs_token_num_expr.clone(), size_expr],
        });
        let token_num_cond_expr = make_cond_expr!(
            abs_token_num_gt_size_check,
            abs_token_num_expr,
            input_token_num_var_ref
        );
        let slice_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Slice,
            args: vec![split_expr_var_ref.clone(), token_num_cond_expr, ONE_LITERAL],
        });

        // If $slice returns an empty vector, populate it with the empty string.
        // This is done to circumvent MongoDB's default behavior of $arrayElemAt[
        // [], 0] returning MISSING instead of the empty string.
        let slice_with_default_conditional = make_cond_expr!(
            Expression::MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Eq,
                args: vec![slice_expr_var_ref.clone(), Expression::Array(vec![])],
            }),
            Expression::Array(vec![Expression::Literal(LiteralValue::String(
                "".to_string(),
            ))]),
            slice_expr_var_ref
        );
        let array_elem_at_expr = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::ElemAt,
            args: vec![slice_with_default_conditional, ZERO_LITERAL],
        });

        // assembly of the desugared slice expr
        let array_elem_at_with_let_vars = Expression::Let(Let {
            vars: vec![LetVariable {
                name: slice_expr_var_name.to_string(),
                expr: Box::new(slice_expr),
            }],
            inside: Box::new(array_elem_at_expr),
        });
        let split_expr_null_check = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Lte,
            args: vec![split_expr_var_ref, NULL_LITERAL],
        });
        let split_expr_null_check_with_elem_at = Expression::Let(Let {
            vars: vec![LetVariable {
                name: split_expr_var_name.to_string(),
                expr: Box::new(split_expr),
            }],
            inside: Box::new(make_cond_expr!(
                split_expr_null_check,
                NULL_LITERAL,
                array_elem_at_with_let_vars
            )),
        });
        let input_delim_empty_check = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![
                input_delim_var_ref,
                Expression::Literal(LiteralValue::String("".to_string())),
            ],
        });
        let split_conditional_with_all_checks = make_cond_expr!(
            input_delim_empty_check,
            NULL_LITERAL,
            split_expr_null_check_with_elem_at
        );

        let let_vars = vec![
            LetVariable {
                name: input_str_var_name.to_string(),
                expr: Box::new(split.args[0].clone()),
            },
            LetVariable {
                name: input_delim_var_name.to_string(),
                expr: Box::new(split.args[1].clone()),
            },
            LetVariable {
                name: input_token_num_var_name.to_string(),
                expr: Box::new(split.args[2].clone()),
            },
        ];
        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(split_conditional_with_all_checks),
        })
    }

    fn desugar_sql_sqrt(&self, sqrt: SQLSemanticOperator) -> Expression {
        let input_var_name = "desugared_sqlSqrt_input".to_string();
        let input_var_ref = Expression::Variable(input_var_name.clone().into());
        let arg_is_nan = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Eq,
            args: vec![input_var_ref.clone(), NAN_LITERAL],
        });
        let arg_negative_check = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Lt,
            args: vec![input_var_ref.clone(), ZERO_LITERAL],
        });
        let arg_negative_check_with_sqrt = Expression::MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Cond,
            args: vec![
                arg_negative_check,
                NULL_LITERAL,
                Expression::MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Sqrt,
                    args: vec![input_var_ref],
                }),
            ],
        });

        let let_vars = vec![LetVariable {
            name: input_var_name,
            expr: Box::new(sqrt.args[0].clone()),
        }];
        Expression::Let(Let {
            vars: let_vars,
            inside: Box::new(make_cond_expr!(
                arg_is_nan,
                NAN_LITERAL,
                arg_negative_check_with_sqrt
            )),
        })
    }
}

impl Visitor for UnsupportedOperatorsDesugarerVisitor {
    fn visit_expression(&mut self, node: Expression) -> Expression {
        use Expression::*;
        let node = match node {
            Like(l) => self.desugar_like(l),
            Is(i) => self.desugar_is(i),
            SqlConvert(s) => self.desugar_sql_convert(s),
            SqlDivide(s) => self.desugar_sql_divide(s),
            SQLSemanticOperator(s) => match s.op {
                SQLOperator::Between => self.desugar_sql_between(s),
                SQLOperator::BitLength => self.desugar_sql_bit_length(s),
                SQLOperator::Coalesce => self.desugar_sql_coalesce(s),
                SQLOperator::Cos | SQLOperator::Sin | SQLOperator::Tan => {
                    self.desugar_trig_function(s)
                }
                SQLOperator::Log => self.desugar_sql_log(s),
                SQLOperator::Mod => self.desugar_sql_mod(s),
                SQLOperator::Neg => self.desugar_sql_neg(s),
                SQLOperator::NullIf => self.desugar_sql_nullif(s),
                SQLOperator::Pos => self.desugar_sql_pos(s),
                SQLOperator::Round => self.desugar_sql_round(s),
                SQLOperator::Slice => self.desugar_sql_slice(s),
                SQLOperator::Split => self.desugar_sql_split(s),
                SQLOperator::Sqrt => self.desugar_sql_sqrt(s),
                _ => SQLSemanticOperator(s),
            },
            MQLSemanticOperator(m) if m.op == MQLOperator::Between => self.desugar_mql_between(m),
            _ => node,
        };
        node.walk(self)
    }
}

#[cfg(test)]
mod test_helpers {
    use super::UnsupportedOperatorsDesugarerVisitor;
    macro_rules! test_convert_sql_pattern {
        ($func_name:ident, expected = $expected:expr, input = $input:expr, escape = $escape:expr) => {
            #[test]
            fn $func_name() {
                let input = $input;
                let expected = $expected;
                let escape = $escape;
                let actual = UnsupportedOperatorsDesugarerVisitor::convert_sql_pattern(
                    input.to_string(),
                    escape.to_string(),
                );
                assert_eq!(expected, actual)
            }
        };
    }

    test_convert_sql_pattern!(
        no_special_sql_characters,
        expected = "^abc$",
        input = "abc",
        escape = "\\"
    );

    test_convert_sql_pattern!(
        unescaped_special_sql_characters,
        expected = "^a.b.*c$",
        input = "a_b%c",
        escape = "\\"
    );

    test_convert_sql_pattern!(
        escaped_special_sql_characters,
        expected = "^a_b%c$",
        input = "a\\_b\\%c",
        escape = "\\"
    );

    test_convert_sql_pattern!(
        escaped_escape_character,
        expected = "^a\\\\.b%c$",
        input = "a\\\\_b\\%c",
        escape = "\\"
    );

    test_convert_sql_pattern!(
        special_mql_characters,
        expected = "^\\.\\^\\$\\*\\+\\?\\(\\)\\[\\{\\\\\\|$",
        input = ".^$*+?()[{\\|",
        escape = "e"
    );
}
