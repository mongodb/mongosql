macro_rules! test_codegen_air_expr {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::codegen::air_to_mql::MqlCodeGenerator;
            let expected = $expected;
            let input = $input;

            let gen = MqlCodeGenerator {};
            assert_eq!(expected, gen.codegen_air_expression(input));
        }
    };
}

mod date_function {
    use crate::air::{
        DateFunction::*, DateFunctionApplication, DatePart::*, Expression::*, LiteralValue::*,
        SQLOperator::*, SQLSemanticOperator,
    };
    use bson::bson;

    test_codegen_air_expr!(
        dateadd,
        expected = Ok(
            bson!({"$dateAdd": {"startDate": "$$NOW", "unit": {"$literal": "year"}, "amount": {"$literal": 5}}})
        ),
        input = DateFunction(DateFunctionApplication {
            function: Add,
            unit: Year,
            args: vec![
                Literal(Integer(5)),
                SQLSemanticOperator(SQLSemanticOperator {
                    op: CurrentTimestamp,
                    args: vec![],
                }),
            ],
        })
    );

    test_codegen_air_expr!(
        datediff,
        expected = Ok(
            bson!({"$dateDiff": {"startDate": "$$NOW", "endDate": "$$NOW", "unit": {"$literal": "year"}, "startOfWeek": {"$literal": "sunday"}}})
        ),
        input = DateFunction(DateFunctionApplication {
            function: Diff,
            unit: Year,
            args: vec![
                SQLSemanticOperator(SQLSemanticOperator {
                    op: CurrentTimestamp,
                    args: vec![],
                }),
                SQLSemanticOperator(SQLSemanticOperator {
                    op: CurrentTimestamp,
                    args: vec![],
                }),
                Literal(String("sunday".to_string())),
            ],
        })
    );

    test_codegen_air_expr!(
        datetrunc,
        expected = Ok(
            bson!({"$dateTrunc": {"date": "$$NOW", "unit": {"$literal": "year"}, "startOfWeek": {"$literal": "sunday"}}})
        ),
        input = DateFunction(DateFunctionApplication {
            function: Trunc,
            unit: Year,
            args: vec![
                SQLSemanticOperator(SQLSemanticOperator {
                    op: CurrentTimestamp,
                    args: vec![],
                }),
                Literal(String("sunday".to_string())),
            ],
        })
    );
}

mod switch {
    use crate::air::{
        Expression::*, LiteralValue::*, MQLOperator::*, MQLSemanticOperator, Switch, SwitchCase,
    };
    use bson::{bson, doc};

    test_codegen_air_expr!(
        one_case,
        expected = Ok(bson!({"$switch":
            {
                "branches": vec![doc!{
                    "case": {"$lt": [{ "$literal": 10}, { "$literal": 20}]},
                    "then": {"$literal": "first"}
                }],
                "default": {"$literal": "else"}
            }
        })),
        input = Switch(Switch {
            branches: vec![SwitchCase {
                case: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                    op: Lt,
                    args: vec![Literal(Integer(10)), Literal(Integer(20))],
                })),
                then: Box::new(Literal(String("first".to_string())))
            }],
            default: Box::new(Literal(String("else".to_string())))
        })
    );

    test_codegen_air_expr!(
        multiple_cases,
        expected = Ok(bson!({"$switch":
            {
                "branches": vec![doc!{
                    "case": {"$lt": [{ "$literal": 10}, { "$literal": 20}]},
                    "then": {"$literal": "first"}
                },
                doc!{
                    "case": {"$gt": [{ "$literal": 1}, { "$literal": 2}]},
                    "then": {"$literal": "second"}
                },
                doc!{
                    "case": {"$eq": [{ "$literal": 1}, { "$literal": 2}]},
                    "then": {"$literal": "third"}
                }],
                "default": {"$literal": "else"}
            }
        })),
        input = Switch(Switch {
            branches: vec![
                SwitchCase {
                    case: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                        op: Lt,
                        args: vec![Literal(Integer(10)), Literal(Integer(20))],
                    })),
                    then: Box::new(Literal(String("first".to_string())))
                },
                SwitchCase {
                    case: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                        op: Gt,
                        args: vec![Literal(Integer(1)), Literal(Integer(2))],
                    })),
                    then: Box::new(Literal(String("second".to_string())))
                },
                SwitchCase {
                    case: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                        op: Eq,
                        args: vec![Literal(Integer(1)), Literal(Integer(2))],
                    })),
                    then: Box::new(Literal(String("third".to_string())))
                }
            ],
            default: Box::new(Literal(String("else".to_string())))
        })
    );
}

mod literal {
    use crate::air::{Expression::*, LiteralValue::*};
    use bson::{bson, Bson};

    test_codegen_air_expr!(
        null,
        expected = Ok(bson!({ "$literal": Bson::Null })),
        input = Literal(Null)
    );

    test_codegen_air_expr!(
        boolean,
        expected = Ok(bson!({ "$literal": Bson::Boolean(true)})),
        input = Literal(Boolean(true))
    );

    test_codegen_air_expr!(
        string,
        expected = Ok(bson!({ "$literal": Bson::String("foo".to_string())})),
        input = Literal(String("foo".to_string()))
    );

    test_codegen_air_expr!(
        int,
        expected = Ok(bson!({ "$literal": 1_i32})),
        input = Literal(Integer(1))
    );

    test_codegen_air_expr!(
        long,
        expected = Ok(bson!({ "$literal": 2_i64})),
        input = Literal(Long(2))
    );

    test_codegen_air_expr!(
        double,
        expected = Ok(bson!({ "$literal": 3.0})),
        input = Literal(Double(3.0))
    );
}

mod trim {

    use crate::air::{Expression::*, LiteralValue::*, Trim, TrimOperator};
    use bson::bson;

    test_codegen_air_expr!(
        trim,
        expected =
            Ok(bson!({ "$trim": { "input": {"$literal": "foo"}, "chars": {"$literal": null }}})),
        input = Trim(Trim {
            op: TrimOperator::Trim,
            input: Box::new(Literal(String("foo".to_string()))),
            chars: Box::new(Literal(Null)),
        })
    );

    test_codegen_air_expr!(
        ltrim,
        expected =
            Ok(bson!({ "$ltrim": {"input": { "$literal": "foo"}, "chars": {"$literal": null }}})),
        input = Trim(Trim {
            op: TrimOperator::LTrim,
            input: Box::new(Literal(String("foo".to_string()))),
            chars: Box::new(Literal(Null)),
        })
    );

    test_codegen_air_expr!(
        rtrim,
        expected =
            Ok(bson!({ "$rtrim": {"input": { "$literal": "foo"}, "chars": {"$literal": null }}})),
        input = Trim(Trim {
            op: TrimOperator::RTrim,
            input: Box::new(Literal(String("foo".to_string()))),
            chars: Box::new(Literal(Null)),
        })
    );
}

mod mql_semantic_operator {
    use crate::air::{Expression::*, LiteralValue::*, MQLOperator::*, MQLSemanticOperator};
    use bson::bson;

    test_codegen_air_expr!(
        concat,
        expected = Ok(bson!({ "$concat": [{ "$literal": "foo"}, { "$literal": "bar"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Concat,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(String("bar".to_string()))
            ],
        })
    );

    test_codegen_air_expr!(
        add,
        expected = Ok(bson!({ "$add": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Add,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        subtract,
        expected = Ok(bson!({ "$subtract": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Subtract,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        multiply,
        expected = Ok(bson!({ "$multiply": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Multiply,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        divide,
        expected = Ok(bson!({ "$divide": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Divide,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        lt,
        expected = Ok(bson!({ "$lt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Lt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        lte,
        expected = Ok(bson!({ "$lte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Lte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        ne,
        expected = Ok(bson!({ "$ne": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Ne,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        eq,
        expected = Ok(bson!({ "$eq": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Eq,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        gt,
        expected = Ok(bson!({ "$gt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Gt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        gte,
        expected = Ok(bson!({ "$gte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Gte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        not,
        expected = Ok(bson!({ "$not": [{ "$literal": false}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Not,
            args: vec![Literal(Boolean(false))],
        })
    );

    test_codegen_air_expr!(
        and,
        expected = Ok(bson!({ "$and": [{ "$literal": true}, { "$literal": false}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: And,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );

    test_codegen_air_expr!(
        or,
        expected = Ok(bson!({ "$or": [{ "$literal": true}, { "$literal": false}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Or,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );

    test_codegen_air_expr!(
        slice,
        expected = Ok(
            bson!({ "$slice": [[{"$literal": 1}, {"$literal": 2}, {"$literal": 3}], {"$literal": 1}, { "$literal": 2}]})
        ),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Slice,
            args: vec![
                Array(vec![
                    Literal(Integer(1)),
                    Literal(Integer(2)),
                    Literal(Integer(3))
                ]),
                Literal(Integer(1)),
                Literal(Integer(2))
            ],
        })
    );

    test_codegen_air_expr!(
        size,
        expected = Ok(bson!({ "$size": [[{"$literal": 1}, {"$literal": 2}, {"$literal": 3}]]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Size,
            args: vec![Array(vec![
                Literal(Integer(1)),
                Literal(Integer(2)),
                Literal(Integer(3))
            ])],
        })
    );

    test_codegen_air_expr!(
        index_of_cp,
        expected = Ok(
            bson!({ "$indexOfCP": [{ "$literal": "foo"}, { "$literal": "bar"}, { "$literal": 1}, { "$literal": 2}]})
        ),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IndexOfCP,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(String("bar".to_string())),
                Literal(Integer(1)),
                Literal(Integer(2)),
            ],
        })
    );

    test_codegen_air_expr!(
        index_of_bytes,
        expected = Ok(
            bson!({ "$indexOfBytes": [{ "$literal": "foo"}, { "$literal": "bar"}, { "$literal": 1}, { "$literal": 2}]})
        ),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IndexOfBytes,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(String("bar".to_string())),
                Literal(Integer(1)),
                Literal(Integer(2)),
            ],
        })
    );

    test_codegen_air_expr!(
        str_len_cp,
        expected = Ok(bson!({ "$strLenCP": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: StrLenCP,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        str_len_bytes,
        expected = Ok(bson!({ "$strLenBytes": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: StrLenBytes,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        abs,
        expected = Ok(bson!({ "$abs": [{"$literal": -1}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Abs,
            args: vec![Literal(Integer(-1)),],
        })
    );

    test_codegen_air_expr!(
        ceil,
        expected = Ok(bson!({ "$ceil": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Ceil,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        cos,
        expected = Ok(bson!({ "$cos": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Cos,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        sin,
        expected = Ok(bson!({ "$sin": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Sin,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        tan,
        expected = Ok(bson!({ "$tan": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Tan,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        degrees_to_radians,
        expected = Ok(bson!({ "$degreesToRadians": [{"$literal": 180.0}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DegreesToRadians,
            args: vec![Literal(Double(180.0)),],
        })
    );

    test_codegen_air_expr!(
        radians_to_degrees,
        expected = Ok(bson!({ "$radiansToDegrees": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: RadiansToDegrees,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        floor,
        expected = Ok(bson!({ "$floor": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Floor,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        log,
        expected = Ok(bson!({ "$log": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Log,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        mod_op,
        expected = Ok(bson!({ "$mod": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Mod,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        pow,
        expected = Ok(bson!({ "$pow": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Pow,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        round,
        expected = Ok(bson!({ "$round": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Round,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        sqrt,
        expected = Ok(bson!({ "$sqrt": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Sqrt,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_air_expr!(
        substr_cp,
        expected =
            Ok(bson!({ "$substrCP": [{ "$literal": "foo"}, { "$literal": 1 }, { "$literal": 2 }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: SubstrCP,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(Integer(1)),
                Literal(Integer(2))
            ],
        })
    );

    test_codegen_air_expr!(
        substr_bytes,
        expected = Ok(
            bson!({ "$substrBytes": [{ "$literal": "foo"}, { "$literal": 1 }, { "$literal": 2 }]})
        ),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: SubstrBytes,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(Integer(1)),
                Literal(Integer(2))
            ],
        })
    );

    test_codegen_air_expr!(
        to_upper,
        expected = Ok(bson!({ "$toUpper": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: ToUpper,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        to_lower,
        expected = Ok(bson!({ "$toLower": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: ToLower,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        split,
        expected = Ok(bson!({ "$split": [{ "$literal": "foo" }, { "$literal": "o" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Split,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(String("o".to_string()))
            ],
        })
    );

    test_codegen_air_expr!(
        year,
        expected = Ok(bson!({ "$year": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Year,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        month,
        expected = Ok(bson!({ "$month": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Month,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        day_of_month,
        expected = Ok(bson!({ "$dayOfMonth": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DayOfMonth,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        hour,
        expected = Ok(bson!({ "$hour": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Hour,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        minute,
        expected = Ok(bson!({ "$minute": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Minute,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        second,
        expected = Ok(bson!({ "$second": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Second,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        week,
        expected = Ok(bson!({ "$week": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Week,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        day_of_year,
        expected = Ok(bson!({ "$dayOfYear": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DayOfYear,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        iso_week,
        expected = Ok(bson!({ "$isoWeek": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsoWeek,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        iso_day_of_week,
        expected = Ok(bson!({ "$isoDayOfWeek": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsoDayOfWeek,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        date_add,
        expected = Ok(bson!({ "$dateAdd": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DateAdd,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        date_diff,
        expected = Ok(bson!({ "$dateDiff": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DateDiff,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        date_trunc,
        expected = Ok(bson!({ "$dateTrunc": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DateTrunc,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        merge_object,
        expected = Ok(bson!({ "$mergeObjects": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: MergeObjects,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        is_number,
        expected = Ok(bson!({ "$isNumber": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsNumber,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        is_array,
        expected = Ok(bson!({ "$isArray": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsArray,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        if_null,
        expected = Ok(bson!({ "$ifNull": [{ "$literal": "foo" }, { "$literal": "bar" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IfNull,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(String("bar".to_string())),
            ],
        })
    );

    test_codegen_air_expr!(
        type_op,
        expected = Ok(bson!({ "$type": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Type,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        array_elem_at,
        expected = Ok(
            bson!({ "$arrayElemAt": [[{ "$literal": "foo" }, { "$literal": "bar" }], { "$literal": 1} ]})
        ),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: ElemAt,
            args: vec![
                Array(vec![
                    Literal(String("foo".to_string())),
                    Literal(String("bar".to_string()))
                ]),
                Literal(Integer(1)),
            ],
        })
    );

    test_codegen_air_expr!(
        cond,
        expected = Ok(
            bson!({ "$cond": [{ "$literal": false }, { "$literal": "foo" }, { "$literal": "bar"} ]})
        ),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Cond,
            args: vec![
                Literal(Boolean(false)),
                Literal(String("foo".to_string())),
                Literal(String("bar".to_string())),
            ],
        })
    );

    test_codegen_air_expr!(
        exists,
        expected = Ok(bson!({ "$exists": [{ "$literal": false }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Exists,
            args: vec![Literal(Boolean(false))],
        })
    );
}

mod sql_semantic_operator {
    use crate::{
        air::{Expression::*, LiteralValue::*, SQLOperator::*, SQLSemanticOperator},
        codegen::air_to_mql::Error,
    };
    use bson::bson;

    test_codegen_air_expr!(
        pos,
        expected = Ok(bson! ({ "$sqlPos": [{ "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Pos,
            args: vec![Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        neg,
        expected = Ok(bson! ({ "$sqlNeg": [{ "$literal": 1}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Neg,
            args: vec![Literal(Integer(1))],
        })
    );

    test_codegen_air_expr!(
        lt,
        expected = Ok(bson!({ "$sqlLt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Lt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        lte,
        expected = Ok(bson!({ "$sqlLte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Lte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        ne,
        expected = Ok(bson!({ "$sqlNe": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Ne,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        eq,
        expected = Ok(bson!({ "$sqlEq": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Eq,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        gt,
        expected = Ok(bson!({ "$sqlGt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Gt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );
    test_codegen_air_expr!(
        gte,
        expected = Ok(bson!({ "$sqlGte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Gte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_air_expr!(
        between,
        expected =
            Ok(bson!({ "$sqlBetween": [[{"$literal": 1}, {"$literal": 2}, {"$literal": 3}]]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Between,
            args: vec![Array(vec![
                Literal(Integer(1)),
                Literal(Integer(2)),
                Literal(Integer(3))
            ])],
        })
    );

    test_codegen_air_expr!(
        nullif_expr,
        expected = Ok(bson!({ "$nullIf": [{ "$literal": true}, { "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: NullIf,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );

    test_codegen_air_expr!(
        coalesce_expr,
        expected = Ok(bson!({ "$coalesce": [{ "$literal": 1}, { "$literal": 2}, {"$literal": 3}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Coalesce,
            args: vec![
                Literal(Integer(1)),
                Literal(Integer(2)),
                Literal(Integer(3))
            ],
        })
    );

    test_codegen_air_expr!(
        not,
        expected = Ok(bson!({ "$sqlNot": [{ "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Not,
            args: vec![Literal(Boolean(false))],
        })
    );

    test_codegen_air_expr!(
        and,
        expected = Ok(bson!({ "$sqlAnd": [{ "$literal": true}, { "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: And,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );
    test_codegen_air_expr!(
        or,
        expected = Ok(bson!({ "$sqlOr": [{ "$literal": true}, { "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Or,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );
    test_codegen_air_expr!(
        slice,
        expected = Ok(
            bson!({ "$sqlSlice": [[{"$literal": 1}, {"$literal": 2}, {"$literal": 3}], {"$literal": 1}, { "$literal": 2}]})
        ),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Slice,
            args: vec![
                Array(vec![
                    Literal(Integer(1)),
                    Literal(Integer(2)),
                    Literal(Integer(3))
                ]),
                Literal(Integer(1)),
                Literal(Integer(2))
            ],
        })
    );
    test_codegen_air_expr!(
        size,
        expected = Ok(bson!({ "$sqlSize": [{"$literal": 1}, {"$literal": 2}, {"$literal": 3}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Size,
            args: vec![Array(vec![
                Literal(Integer(1)),
                Literal(Integer(2)),
                Literal(Integer(3))
            ])],
        })
    );
    test_codegen_air_expr!(
        index_of_cp,
        expected = Ok(
            bson!({ "$sqlIndexOfCP": [{ "$literal": 2}, { "$literal": 1}, { "$literal": "bar"}, { "$literal": "foo"}]})
        ),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: IndexOfCP,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(String("bar".to_string())),
                Literal(Integer(1)),
                Literal(Integer(2)),
            ],
        })
    );

    test_codegen_air_expr!(
        str_len_cp,
        expected = Ok(bson!({ "$sqlStrLenCP": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: StrLenCP,
            args: vec![Literal(String("foo".to_string())),],
        })
    );
    test_codegen_air_expr!(
        str_len_bytes,
        expected = Ok(bson!({ "$sqlStrLenBytes": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: StrLenBytes,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        bit_length,
        expected = Ok(bson!({ "$sqlBitLength": [{ "$literal": "foo"}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: BitLength,
            args: vec![Literal(String("foo".to_string())),],
        })
    );
    test_codegen_air_expr!(
        cos,
        expected = Ok(bson!({ "$sqlCos": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Cos,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_air_expr!(
        log,
        expected = Ok(bson!({ "$sqlLog": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Log,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_air_expr!(
        mod_op,
        expected = Ok(bson!({ "$sqlMod": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Mod,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_air_expr!(
        round,
        expected = Ok(bson!({ "$sqlRound": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Round,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_air_expr!(
        sin,
        expected = Ok(bson!({ "$sqlSin": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Sin,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_air_expr!(
        sqrt,
        expected = Ok(bson!({ "$sqlSqrt": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Sqrt,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_air_expr!(
        tan,
        expected = Ok(bson!({ "$sqlTan": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Tan,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_air_expr!(
        substr_cp,
        expected = Ok(
            bson!({ "$sqlSubstrCP": [{ "$literal": "foo"}, { "$literal": 1 }, { "$literal": 2 }]})
        ),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: SubstrCP,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(Integer(1)),
                Literal(Integer(2))
            ],
        })
    );
    test_codegen_air_expr!(
        to_upper,
        expected = Ok(bson!({ "$sqlToUpper": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: ToUpper,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        to_lower,
        expected = Ok(bson!({ "$sqlToLower": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: ToLower,
            args: vec![Literal(String("foo".to_string())),],
        })
    );
    test_codegen_air_expr!(
        split,
        expected = Ok(bson!({ "$sqlSplit": [{ "$literal": "foo" }, { "$literal": "o" }]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Split,
            args: vec![
                Literal(String("foo".to_string())),
                Literal(String("o".to_string()))
            ],
        })
    );
    test_codegen_air_expr!(
        current_timestamp,
        expected = Ok(bson!("$$NOW")),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: CurrentTimestamp,
            args: vec![],
        })
    );
    test_codegen_air_expr!(
        computed_field_access,
        expected = Err(Error::UnsupportedOperator(ComputedFieldAccess)),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: ComputedFieldAccess,
            args: vec![]
        })
    );
}

mod document {
    use crate::{
        air::{Expression::*, LiteralValue::*},
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_air_expr!(
        empty,
        expected = Ok(bson!({"$literal": {}})),
        input = Document(unchecked_unique_linked_hash_map! {})
    );
    test_codegen_air_expr!(
        non_empty,
        expected = Ok(bson!({"foo": {"$literal": 1}})),
        input =
            Document(unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(Integer(1))})
    );
    test_codegen_air_expr!(
        nested,
        expected = Ok(bson!({"foo": {"$literal": 1}, "bar": {"baz": {"$literal": 2}}})),
        input = Document(unchecked_unique_linked_hash_map! {
            "foo".to_string() => Literal(Integer(1)),
            "bar".to_string() => Document(unchecked_unique_linked_hash_map!{
                "baz".to_string() => Literal(Integer(2))
            }),
        })
    );
}

mod array {
    use crate::air::{Expression::*, LiteralValue::*};
    use bson::bson;

    test_codegen_air_expr!(empty, expected = Ok(bson!([])), input = Array(vec![]));

    test_codegen_air_expr!(
        non_empty,
        expected = Ok(bson!([{"$literal": "abc"}])),
        input = Array(vec![Literal(String("abc".to_string()))])
    );

    test_codegen_air_expr!(
        nested,
        expected = Ok(bson!([{ "$literal": null }, [{ "$literal": null }]])),
        input = Array(vec![Literal(Null), Array(vec![Literal(Null)])])
    );
}

mod variable {
    use crate::air::{Expression::*, Variable};
    use bson::{bson, Bson};

    test_codegen_air_expr!(
        simple,
        expected = Ok(bson!(Bson::String("$$foo".to_string()))),
        input = Variable(Variable {
            parent: None,
            name: "foo".to_string()
        })
    );

    test_codegen_air_expr!(
        nested,
        expected = Ok(bson!(Bson::String("$$x.y.z".to_string()))),
        input = Variable(Variable {
            parent: Some(Box::new(Variable {
                parent: Some(Box::new(Variable {
                    parent: None,
                    name: "x".to_string()
                })),
                name: "y".to_string(),
            })),
            name: "z".to_string()
        })
    );
}

mod field_ref {
    use crate::air::{self, Expression::FieldRef};
    use bson::{bson, Bson};

    test_codegen_air_expr!(
        no_parent,
        expected = Ok(bson!(Bson::String("$foo".to_string()))),
        input = FieldRef(air::FieldRef {
            parent: None,
            name: "foo".to_string()
        })
    );
    test_codegen_air_expr!(
        parent,
        expected = Ok(bson!(Bson::String("$bar.foo".to_string()))),
        input = FieldRef(air::FieldRef {
            parent: Some(Box::new(air::FieldRef {
                parent: None,
                name: "bar".to_string()
            })),
            name: "foo".to_string()
        })
    );

    test_codegen_air_expr!(
        grandparent,
        expected = Ok(bson!(Bson::String("$baz.bar.foo".to_string()))),
        input = FieldRef(air::FieldRef {
            parent: Some(Box::new(air::FieldRef {
                parent: Some(Box::new(air::FieldRef {
                    parent: None,
                    name: "baz".to_string()
                })),
                name: "bar".to_string()
            })),
            name: "foo".to_string()
        })
    );
}

mod like {
    use crate::air::{self, Expression};
    use bson::bson;

    test_codegen_air_expr!(
        with_escape,
        expected = Ok(bson!({"$like": {
            "input": "$input",
            "pattern": "$pattern",
            "escape": "escape",
        }})),
        input = Expression::Like(air::Like {
            expr: Box::new(Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "input".to_string()
            })),
            pattern: Box::new(Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "pattern".to_string()
            })),
            escape: Some("escape".to_string()),
        })
    );

    test_codegen_air_expr!(
        without_escape,
        expected = Ok(bson!({"$like": {
            "input": "$input",
            "pattern": "$pattern",
        }})),
        input = Expression::Like(air::Like {
            expr: Box::new(Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "input".to_string()
            })),
            pattern: Box::new(Expression::FieldRef(air::FieldRef {
                parent: None,
                name: "pattern".to_string()
            })),
            escape: None,
        })
    );
}

mod is {
    use crate::air::{Expression::*, FieldRef, Is, Type, TypeOrMissing};
    use bson::bson;

    test_codegen_air_expr!(
        target_type_missing,
        expected = Ok(bson!({"$sqlIs": ["$x", {"$literal": "missing"}]})),
        input = Is(Is {
            expr: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "x".into(),
            })),
            target_type: TypeOrMissing::Missing,
        })
    );

    test_codegen_air_expr!(
        target_type_number,
        expected = Ok(bson!({"$sqlIs": ["$x", {"$literal": "number"}]})),
        input = Is(Is {
            expr: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "x".into(),
            })),
            target_type: TypeOrMissing::Number,
        })
    );

    test_codegen_air_expr!(
        target_type_type,
        expected = Ok(bson!({"$sqlIs": ["$x", {"$literal": "object"}]})),
        input = Is(Is {
            expr: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "x".into(),
            })),
            target_type: TypeOrMissing::Type(Type::Document),
        })
    );
}

mod get_field {
    use crate::{
        air::{
            self,
            Expression::{Document, GetField, Literal},
            LiteralValue,
        },
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_air_expr!(
        basic,
        expected = Ok(bson!({"$getField": {"field": "x", "input": {"x": {"$literal": 42}}}})),
        input = GetField(air::GetField {
            field: "x".to_string(),
            input: Document(unchecked_unique_linked_hash_map! {
                "x".to_string() => Literal(LiteralValue::Integer(42)),
            })
            .into(),
        })
    );

    test_codegen_air_expr!(
        with_dollar_sign,
        expected = Ok(
            bson!({"$getField": {"field": { "$literal": "$x"}, "input": {"$x": {"$literal": 42}}}})
        ),
        input = GetField(air::GetField {
            field: "$x".to_string(),
            input: Document(unchecked_unique_linked_hash_map! {
                "$x".to_string() => Literal(LiteralValue::Integer(42)),
            })
            .into(),
        })
    );
}

mod set_field {
    use crate::air::{Expression::*, FieldRef, SetField, Variable};
    use bson::bson;

    test_codegen_air_expr!(
        simple,
        expected = Ok(bson!({"$setField": {"field": "", "input": "$$ROOT", "value": "$__bot"}})),
        input = SetField(SetField {
            field: "".into(),
            input: Box::new(Variable(Variable {
                parent: None,
                name: "ROOT".into()
            })),
            value: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "__bot".into(),
            }))
        })
    );

    test_codegen_air_expr!(
        use_literal_when_needed,
        expected = Ok(
            bson!({"$setField": {"field": {"$literal": "$x_val"}, "input": "$$ROOT", "value": "$x"}})
        ),
        input = SetField(SetField {
            field: "$x_val".into(),
            input: Box::new(Variable(Variable {
                parent: None,
                name: "ROOT".into()
            })),
            value: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "x".into(),
            }))
        })
    );
}

mod unset_field {
    use crate::air::{Expression::*, FieldRef, UnsetField};
    use bson::bson;

    test_codegen_air_expr!(
        simple,
        expected = Ok(bson!({"$unsetField": {"field": "__bot", "input": "$doc"}})),
        input = UnsetField(UnsetField {
            field: "__bot".into(),
            input: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "doc".into(),
            }))
        })
    );

    test_codegen_air_expr!(
        use_literal_when_needed,
        expected = Ok(bson!({"$unsetField": {"field": {"$literal": "$x"}, "input": "$doc"}})),
        input = UnsetField(UnsetField {
            field: "$x".into(),
            input: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "doc".into(),
            }))
        })
    );
}

mod sql_convert {
    use crate::{
        air::{Expression::*, LiteralValue, SqlConvert, SqlConvertTargetType},
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_air_expr!(
        array,
        expected = Ok(bson!({ "$sqlConvert": {
          "input": [],
          "to": "array",
          "onNull": {"$literal": null},
          "onError": {"$literal": null}
        }})),
        input = SqlConvert(SqlConvert {
            input: Box::new(Array(vec![])),
            to: SqlConvertTargetType::Array,
            on_null: Box::new(Literal(LiteralValue::Null)),
            on_error: Box::new(Literal(LiteralValue::Null)),
        })
    );
    test_codegen_air_expr!(
        document,
        expected = Ok(bson!({ "$sqlConvert": {
          "input": {"$literal":{}},
          "to": "object",
          "onNull": {"$literal": null},
          "onError": {"$literal": null}
        }})),
        input = SqlConvert(SqlConvert {
            input: Box::new(Document(unchecked_unique_linked_hash_map! {})),
            to: SqlConvertTargetType::Document,
            on_null: Box::new(Literal(LiteralValue::Null)),
            on_error: Box::new(Literal(LiteralValue::Null)),
        })
    );
}

mod sql_divide {
    use crate::air::{Expression::*, LiteralValue, SqlDivide};

    test_codegen_air_expr!(
        simple,
        expected = Ok(
            bson::bson! ({"$sqlDivide": {"dividend": {"$literal": 1}, "divisor": {"$literal": 2}, "onError": {"$literal": null}}})
        ),
        input = SqlDivide(SqlDivide {
            dividend: Box::new(Literal(LiteralValue::Integer(1))),
            divisor: Box::new(Literal(LiteralValue::Integer(2))),
            on_error: Box::new(Literal(LiteralValue::Null)),
        })
    );
}

mod convert {
    use crate::{
        air::{Convert, Expression, LiteralValue::*, Type},
        codegen,
    };
    use bson::bson;

    macro_rules! test_codegen_working_convert {
        ($test_name:ident, expected = $expected_ty_str:expr, input = $input:expr) => {
        test_codegen_air_expr!(
            $test_name,
            expected = Ok(bson!({ "$convert":
                {
                        "input": {"$literal": "foo"},
                        "to": $expected_ty_str,
                        "onError": {"$literal": null},
                        "onNull": {"$literal": null},
                    }
                })),
            input = Expression::Convert(Convert {
                    input: Expression::Literal(String("foo".to_string())).into(),
                    to: $input,
                    on_null: Expression::Literal(Null).into(),
                    on_error: Expression::Literal(Null).into(),
                })
            );
        };
    }

    test_codegen_air_expr!(
        convert_array,
        expected = Err(codegen::air_to_mql::Error::ConvertToArray),
        input = Expression::Convert(Convert {
            input: Expression::Literal(String("foo".to_string())).into(),
            to: Type::Array,
            on_null: Expression::Literal(Null).into(),
            on_error: Expression::Literal(Null).into(),
        })
    );

    test_codegen_air_expr!(
        convert_document,
        expected = Err(codegen::air_to_mql::Error::ConvertToDocument),
        input = Expression::Convert(Convert {
            input: Expression::Literal(String("foo".to_string())).into(),
            to: Type::Document,
            on_null: Expression::Literal(Null).into(),
            on_error: Expression::Literal(Null).into(),
        })
    );

    test_codegen_working_convert!(
        convert_bin_data,
        expected = "binData",
        input = Type::BinData
    );

    test_codegen_working_convert!(convert_date, expected = "date", input = Type::Datetime);

    test_codegen_working_convert!(
        convert_db_pointer,
        expected = "dbPointer",
        input = Type::DbPointer
    );

    test_codegen_working_convert!(
        convert_decimal,
        expected = "decimal",
        input = Type::Decimal128
    );

    test_codegen_working_convert!(convert_double, expected = "double", input = Type::Double);

    test_codegen_working_convert!(convert_int, expected = "int", input = Type::Int32);

    test_codegen_working_convert!(convert_long, expected = "long", input = Type::Int64);

    test_codegen_working_convert!(
        convert_javascript,
        expected = "javascript",
        input = Type::Javascript
    );

    test_codegen_working_convert!(
        convert_javascript_with_scope,
        expected = "javascriptWithScope",
        input = Type::JavascriptWithScope
    );

    test_codegen_working_convert!(convert_max_key, expected = "maxKey", input = Type::MaxKey);

    test_codegen_working_convert!(convert_min_key, expected = "minKey", input = Type::MinKey);

    test_codegen_working_convert!(convert_null, expected = "null", input = Type::Null);

    test_codegen_working_convert!(
        convert_object_id,
        expected = "objectId",
        input = Type::ObjectId
    );

    test_codegen_working_convert!(
        convert_regular_expression,
        expected = "regex",
        input = Type::RegularExpression
    );

    test_codegen_working_convert!(convert_string, expected = "string", input = Type::String);

    test_codegen_working_convert!(convert_symbol, expected = "symbol", input = Type::Symbol);

    test_codegen_working_convert!(
        convert_timestamp,
        expected = "timestamp",
        input = Type::Timestamp
    );

    test_codegen_working_convert!(
        convert_undefined,
        expected = "undefined",
        input = Type::Undefined
    );
}

mod let_expr {
    use crate::air::{
        Expression::*, FieldRef, Let, LetVariable, MQLOperator, MQLSemanticOperator, Variable,
    };
    use bson::bson;

    test_codegen_air_expr!(
        no_variables,
        expected = Ok(bson!({
            "$let": {
                "vars": {},
                "in": {
                    "$add": ["$x", "$y"]
                }
            }
        })),
        input = Let(Let {
            vars: vec![],
            inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Add,
                args: vec![
                    FieldRef(FieldRef {
                        parent: None,
                        name: "x".to_string(),
                    }),
                    FieldRef(FieldRef {
                        parent: None,
                        name: "y".to_string(),
                    }),
                ],
            })),
        })
    );

    test_codegen_air_expr!(
        one_variable,
        expected = Ok(bson!({
            "$let": {
                "vars": {
                    "v_x": "$x"
                },
                "in": {
                    "$add": ["$$v_x", "$y"]
                }
            }
        })),
        input = Let(Let {
            vars: vec![LetVariable {
                name: "v_x".to_string(),
                expr: Box::new(FieldRef(FieldRef {
                    parent: None,
                    name: "x".to_string(),
                })),
            }],
            inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Add,
                args: vec![
                    Variable(Variable {
                        parent: None,
                        name: "v_x".to_string()
                    }),
                    FieldRef(FieldRef {
                        parent: None,
                        name: "y".to_string(),
                    }),
                ],
            })),
        })
    );

    test_codegen_air_expr!(
        multiple_variables,
        expected = Ok(bson!({
            "$let": {
                "vars": {
                    "v_x": "$x",
                    "v_y": "$y",
                    "v_z": "$z"
                },
                "in": {
                    "$add": ["$$v_x", "$$v_y", "$$v_z"]
                }
            }
        })),
        input = Let(Let {
            vars: vec![
                LetVariable {
                    name: "v_x".to_string(),
                    expr: Box::new(FieldRef(FieldRef {
                        parent: None,
                        name: "x".to_string(),
                    })),
                },
                LetVariable {
                    name: "v_y".to_string(),
                    expr: Box::new(FieldRef(FieldRef {
                        parent: None,
                        name: "y".to_string(),
                    })),
                },
                LetVariable {
                    name: "v_z".to_string(),
                    expr: Box::new(FieldRef(FieldRef {
                        parent: None,
                        name: "z".to_string(),
                    })),
                },
            ],
            inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Add,
                args: vec![
                    Variable(Variable {
                        parent: None,
                        name: "v_x".to_string()
                    }),
                    Variable(Variable {
                        parent: None,
                        name: "v_y".to_string()
                    }),
                    Variable(Variable {
                        parent: None,
                        name: "v_z".to_string()
                    }),
                ],
            })),
        })
    );
}

mod regex_match_expr {
    use crate::air::{Expression::*, LiteralValue, RegexMatch};
    use bson::bson;

    test_codegen_air_expr!(
        regex_match_no_options,
        expected = Ok(
            bson!({ "$regexMatch": {"input": {"$literal": "input"}, "regex": {"$literal": "regex"}}})
        ),
        input = RegexMatch(RegexMatch {
            input: Box::new(Literal(LiteralValue::String("input".to_string()))),
            regex: Box::new(Literal(LiteralValue::String("regex".to_string()))),
            options: None
        })
    );

    test_codegen_air_expr!(
        regex_match_options,
        expected = Ok(
            bson!({ "$regexMatch": {"input": {"$literal": "input"}, "regex": {"$literal": "regex"}, "options": {"$literal": "options"}}})
        ),
        input = RegexMatch(RegexMatch {
            input: Box::new(Literal(LiteralValue::String("input".to_string()))),
            regex: Box::new(Literal(LiteralValue::String("regex".to_string()))),
            options: Some(Box::new(Literal(LiteralValue::String(
                "options".to_string()
            ))),)
        })
    );
}

mod subquery_exists {
    use crate::{
        air::{
            Collection, Expression::*, FieldRef, LetVariable, Project, Stage::*, SubqueryExists,
            Variable,
        },
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_air_expr!(
        subquery_exists_uncorrelated,
        expected = Ok(bson!({
            "$subqueryExists": {
                "db": "test",
                "collection": "foo",
                "let": {},
                "pipeline": [
                    {"$project": {"_id": 0, "foo": "$$ROOT"}}
                ]
            }
        })),
        input = SubqueryExists(SubqueryExists {
            let_bindings: vec![],
            pipeline: Box::new(Project(Project {
                source: Box::new(Collection(Collection {
                    db: "test".to_string(),
                    collection: "foo".to_string(),
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "foo".to_string() => Variable(Variable{parent: None, name: "ROOT".to_string()}),
                },
            })),
        })
    );

    test_codegen_air_expr!(
        subquery_exists_correlated,
        expected = Ok(bson::bson!(
            {"$subqueryExists": {
                "db": "test",
                "collection": "bar",
                "let": {"vfoo_0": "$foo"},
                "pipeline": [
                    {"$project": {"_id": 0,"bar": "$$ROOT"}},
                    {"$project": {"_id": 0,"__bot": {"a": "$$vfoo_0.a"}}}
                ]
            }}
        )),
        input = SubqueryExists(SubqueryExists {
            let_bindings: vec![LetVariable {
                name: "vfoo_0".to_string(),
                expr: Box::new(FieldRef(FieldRef {
                    parent: None,
                    name: "foo".to_string(),
                })),
            },],
            pipeline: Box::new(Project(Project {
                source: Box::new(Project(Project {
                    source: Box::new(Collection(Collection {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "bar".to_string() => Variable(Variable{parent: None, name: "ROOT".to_string()}),
                    },
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => Variable(Variable {
                            parent: Some(Box::new(Variable {
                                parent: None,
                                name: "vfoo_0".to_string(),
                            })),
                            name: "a".to_string()
                        })
                    }),
                },
            })),
        })
    );
}

mod subquery {
    use crate::{
        air::{
            Collection, Documents, Expression::*, FieldRef, LetVariable, Project, Stage::*,
            Subquery, Variable,
        },
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_air_expr!(
        no_db_or_coll,
        expected = Ok(bson!({
            "$subquery": {
                "let": {},
                "outputPath": ["arr"],
                "pipeline": [
                    {"$documents": []},
                    {"$project": {"_id": 0, "arr": "$$ROOT"}}
                ]
            }
        })),
        input = Subquery(Subquery {
            let_bindings: vec![],
            output_path: vec!["arr".to_string()],
            pipeline: Box::new(Project(Project {
                source: Box::new(Documents(Documents { array: vec![] })),
                specifications: unchecked_unique_linked_hash_map! {
                    "arr".to_string() => Variable(Variable{parent: None, name: "ROOT".to_string()}),
                }
            })),
        })
    );

    test_codegen_air_expr!(
        fully_specified,
        expected = Ok(bson!({
            "$subquery": {
                "db": "test",
                "collection": "bar",
                "let": {
                    "vfoo_0": "$foo",
                    "vbaz_0": "$baz"
                },
                "outputPath": ["__bot", "a"],
                "pipeline": [
                    {"$project": {"_id": 0,"bar": "$$ROOT"}},
                    {"$project": {"_id": 0,"__bot": {"a": "$$vfoo_0.a"}}}
                ]
            }
        })),
        input = Subquery(Subquery {
            let_bindings: vec![
                LetVariable {
                    name: "vfoo_0".to_string(),
                    expr: Box::new(FieldRef(FieldRef {
                        parent: None,
                        name: "foo".to_string()
                    })),
                },
                LetVariable {
                    name: "vbaz_0".to_string(),
                    expr: Box::new(FieldRef(FieldRef {
                        parent: None,
                        name: "baz".to_string()
                    })),
                },
            ],
            output_path: vec!["__bot".to_string(), "a".to_string()],
            pipeline: Box::new(Project(Project {
                source: Box::new(Project(Project {
                    source: Box::new(Collection(Collection {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "bar".to_string() => Variable(Variable{parent: None, name: "ROOT".to_string()}),
                    }
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => Variable(Variable{
                            parent: Some(Box::new(Variable{parent: None, name: "vfoo_0".to_string()})),
                            name: "a".to_string(),
                        }),
                    }),
                }
            })),
        })
    );
}

mod subquery_comparison {
    use crate::{
        air::{
            Collection, Documents, Expression::*, FieldRef, LetVariable, Project, Stage::*,
            Subquery, SubqueryComparison, SubqueryComparisonOp, SubqueryModifier, Variable,
        },
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_air_expr!(
        no_db_or_coll,
        expected = Ok(bson!({
            "$subqueryComparison": {
                "op": "eq",
                "modifier": "any",
                "arg": "$x",
                "subquery": {
                    "let": {},
                    "outputPath": ["arr"],
                    "pipeline": [
                        {"$documents": []},
                        {"$project": {"_id": 0, "arr": "$$ROOT"}}
                    ]
                }
            }
        })),
        input = SubqueryComparison(SubqueryComparison {
            op: SubqueryComparisonOp::Eq,
            modifier: SubqueryModifier::Any,
            arg: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "x".to_string(),
            })),
            subquery: Box::new(Subquery {
                let_bindings: vec![],
                output_path: vec!["arr".to_string()],
                pipeline: Box::new(Project(Project {
                    source: Box::new(Documents(Documents { array: vec![] })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "arr".to_string() => Variable(Variable{parent: None, name: "ROOT".to_string()}),
                    }
                })),
            }),
        })
    );

    test_codegen_air_expr!(
        fully_specified,
        expected = Ok(bson!({
            "$subqueryComparison": {
                "op": "gt",
                "modifier": "all",
                "arg": "$x",
                "subquery": {
                    "db": "test",
                    "collection": "bar",
                    "let": {
                        "vfoo_0": "$foo",
                        "vbaz_0": "$baz"
                    },
                    "outputPath": ["__bot", "a"],
                    "pipeline": [
                        {"$project": {"_id": 0,"bar": "$$ROOT"}},
                        {"$project": {"_id": 0,"__bot": {"a": "$$vfoo_0.a"}}}
                    ]
                }
            }
        })),
        input = SubqueryComparison(SubqueryComparison {
            op: SubqueryComparisonOp::Gt,
            modifier: SubqueryModifier::All,
            arg: Box::new(FieldRef(FieldRef {
                parent: None,
                name: "x".to_string(),
            })),
            subquery: Box::new(Subquery {
                let_bindings: vec![
                    LetVariable {
                        name: "vfoo_0".to_string(),
                        expr: Box::new(FieldRef(FieldRef {
                            parent: None,
                            name: "foo".to_string()
                        })),
                    },
                    LetVariable {
                        name: "vbaz_0".to_string(),
                        expr: Box::new(FieldRef(FieldRef {
                            parent: None,
                            name: "baz".to_string()
                        })),
                    },
                ],
                output_path: vec!["__bot".to_string(), "a".to_string()],
                pipeline: Box::new(Project(Project {
                    source: Box::new(Project(Project {
                        source: Box::new(Collection(Collection {
                            db: "test".to_string(),
                            collection: "bar".to_string(),
                        })),
                        specifications: unchecked_unique_linked_hash_map! {
                            "bar".to_string() => Variable(Variable{parent: None, name: "ROOT".to_string()}),
                        }
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "__bot".to_string() => Document(unchecked_unique_linked_hash_map! {
                            "a".to_string() => Variable(Variable{
                                parent: Some(Box::new(Variable{parent: None, name: "vfoo_0".to_string()})),
                                name: "a".to_string(),
                            }),
                        }),
                    }
                })),
            }),
        })
    );
}
