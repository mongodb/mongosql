macro_rules! test_codegen_expression {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::codegen::MqlCodeGenerator;
            let expected = $expected;
            let input = $input;

            let gen = MqlCodeGenerator {};
            assert_eq!(expected, gen.codegen_expression(input));
        }
    };
}

mod date_function {
    use crate::air::{
        DateFunction::*, DateFunctionApplication, DatePart::*, Expression::*, LiteralValue::*,
        SQLOperator::*, SQLSemanticOperator,
    };
    use bson::bson;

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
        null,
        expected = Ok(bson!({ "$literal": Bson::Null })),
        input = Literal(Null)
    );

    test_codegen_expression!(
        boolean,
        expected = Ok(bson!({ "$literal": Bson::Boolean(true)})),
        input = Literal(Boolean(true))
    );

    test_codegen_expression!(
        string,
        expected = Ok(bson!({ "$literal": Bson::String("foo".to_string())})),
        input = Literal(String("foo".to_string()))
    );

    test_codegen_expression!(
        int,
        expected = Ok(bson!({ "$literal": 1_i32})),
        input = Literal(Integer(1))
    );

    test_codegen_expression!(
        long,
        expected = Ok(bson!({ "$literal": 2_i64})),
        input = Literal(Long(2))
    );

    test_codegen_expression!(
        double,
        expected = Ok(bson!({ "$literal": 3.0})),
        input = Literal(Double(3.0))
    );
}

mod trim {

    use crate::air::{Expression::*, LiteralValue::*, Trim, TrimOperator};
    use bson::bson;

    test_codegen_expression!(
        trim,
        expected =
            Ok(bson!({ "$trim": { "input": {"$literal": "foo"}, "chars": {"$literal": null }}})),
        input = Trim(Trim {
            op: TrimOperator::Trim,
            input: Box::new(Literal(String("foo".to_string()))),
            chars: Box::new(Literal(Null)),
        })
    );

    test_codegen_expression!(
        ltrim,
        expected =
            Ok(bson!({ "$ltrim": {"input": { "$literal": "foo"}, "chars": {"$literal": null }}})),
        input = Trim(Trim {
            op: TrimOperator::LTrim,
            input: Box::new(Literal(String("foo".to_string()))),
            chars: Box::new(Literal(Null)),
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
        add,
        expected = Ok(bson!({ "$add": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Add,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        subtract,
        expected = Ok(bson!({ "$subtract": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Subtract,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        multiply,
        expected = Ok(bson!({ "$multiply": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Multiply,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        divide,
        expected = Ok(bson!({ "$divide": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Divide,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        lt,
        expected = Ok(bson!({ "$lt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Lt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        lte,
        expected = Ok(bson!({ "$lte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Lte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        ne,
        expected = Ok(bson!({ "$ne": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Ne,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        eq,
        expected = Ok(bson!({ "$eq": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Eq,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        gt,
        expected = Ok(bson!({ "$gt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Gt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        gte,
        expected = Ok(bson!({ "$gte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Gte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        not,
        expected = Ok(bson!({ "$not": [{ "$literal": false}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Not,
            args: vec![Literal(Boolean(false))],
        })
    );

    test_codegen_expression!(
        and,
        expected = Ok(bson!({ "$and": [{ "$literal": true}, { "$literal": false}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: And,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );

    test_codegen_expression!(
        or,
        expected = Ok(bson!({ "$or": [{ "$literal": true}, { "$literal": false}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Or,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
        str_len_cp,
        expected = Ok(bson!({ "$strLenCP": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: StrLenCP,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        str_len_bytes,
        expected = Ok(bson!({ "$strLenBytes": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: StrLenBytes,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        abs,
        expected = Ok(bson!({ "$abs": [{"$literal": -1}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Abs,
            args: vec![Literal(Integer(-1)),],
        })
    );

    test_codegen_expression!(
        ceil,
        expected = Ok(bson!({ "$ceil": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Ceil,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        cos,
        expected = Ok(bson!({ "$cos": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Cos,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        sin,
        expected = Ok(bson!({ "$sin": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Sin,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        tan,
        expected = Ok(bson!({ "$tan": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Tan,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        degrees_to_radians,
        expected = Ok(bson!({ "$degreesToRadians": [{"$literal": 180.0}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DegreesToRadians,
            args: vec![Literal(Double(180.0)),],
        })
    );

    test_codegen_expression!(
        radians_to_degrees,
        expected = Ok(bson!({ "$radiansToDegrees": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: RadiansToDegrees,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        floor,
        expected = Ok(bson!({ "$floor": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Floor,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        log,
        expected = Ok(bson!({ "$log": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Log,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        mod_op,
        expected = Ok(bson!({ "$mod": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Mod,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        pow,
        expected = Ok(bson!({ "$pow": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Pow,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        round,
        expected = Ok(bson!({ "$round": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Round,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        sqrt,
        expected = Ok(bson!({ "$sqrt": [{"$literal": 3.5}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Sqrt,
            args: vec![Literal(Double(3.5)),],
        })
    );

    test_codegen_expression!(
        avg,
        expected = Ok(bson!({ "$avg": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Avg,
            args: vec![FieldRef("foo".to_string().into()),],
        })
    );

    test_codegen_expression!(
        max,
        expected = Ok(bson!({ "$max": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Max,
            args: vec![FieldRef("foo".to_string().into()),],
        })
    );

    test_codegen_expression!(
        min,
        expected = Ok(bson!({ "$min": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Min,
            args: vec![FieldRef("foo".to_string().into()),],
        })
    );

    test_codegen_expression!(
        sum,
        expected = Ok(bson!({ "$sum": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Sum,
            args: vec![FieldRef("foo".to_string().into()),],
        })
    );

    test_codegen_expression!(
        stddev_pop,
        expected = Ok(bson!({ "$stdDevPop": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: StddevPop,
            args: vec![FieldRef("foo".to_string().into()),],
        })
    );

    test_codegen_expression!(
        stddev_samp,
        expected = Ok(bson!({ "$stdDevSamp": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: StddevSamp,
            args: vec![FieldRef("foo".to_string().into()),],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
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

    test_codegen_expression!(
        to_upper,
        expected = Ok(bson!({ "$toUpper": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: ToUpper,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        to_lower,
        expected = Ok(bson!({ "$toLower": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: ToLower,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
        year,
        expected = Ok(bson!({ "$year": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Year,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        month,
        expected = Ok(bson!({ "$month": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Month,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        day_of_month,
        expected = Ok(bson!({ "$dayOfMonth": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DayOfMonth,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        hour,
        expected = Ok(bson!({ "$hour": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Hour,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        minute,
        expected = Ok(bson!({ "$minute": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Minute,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        second,
        expected = Ok(bson!({ "$second": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Second,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        week,
        expected = Ok(bson!({ "$week": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Week,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        day_of_year,
        expected = Ok(bson!({ "$dayOfYear": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DayOfYear,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        iso_week,
        expected = Ok(bson!({ "$isoWeek": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsoWeek,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        iso_day_of_week,
        expected = Ok(bson!({ "$isoDayOfWeek": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsoDayOfWeek,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        date_add,
        expected = Ok(bson!({ "$dateAdd": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DateAdd,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        date_diff,
        expected = Ok(bson!({ "$dateDiff": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DateDiff,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        date_trunc,
        expected = Ok(bson!({ "$dateTrunc": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: DateTrunc,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        merge_object,
        expected = Ok(bson!({ "$mergeObjects": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: MergeObjects,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        is_number,
        expected = Ok(bson!({ "$isNumber": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsNumber,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        is_array,
        expected = Ok(bson!({ "$isArray": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: IsArray,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
        type_op,
        expected = Ok(bson!({ "$type": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Type,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
        in_op,
        expected = Ok(bson!({ "$in": [{ "$literal": 1}, [{ "$literal": 1 }, { "$literal": 2 }]]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: In,
            args: vec![
                Literal(Integer(1)),
                Array(vec![Literal(Integer(1)), Literal(Integer(2))]),
            ],
        })
    );

    test_codegen_expression!(
        first,
        expected = Ok(bson!({ "$first": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: First,
            args: vec![FieldRef("foo".to_string().into())],
        })
    );

    test_codegen_expression!(
        last,
        expected = Ok(bson!({ "$last": ["$foo"]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Last,
            args: vec![FieldRef("foo".to_string().into())],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
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
        codegen::Error,
    };
    use bson::bson;

    test_codegen_expression!(
        pos,
        expected = Ok(bson! ({ "$sqlPos": [{ "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Pos,
            args: vec![Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        neg,
        expected = Ok(bson! ({ "$sqlNeg": [{ "$literal": 1}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Neg,
            args: vec![Literal(Integer(1))],
        })
    );

    test_codegen_expression!(
        lt,
        expected = Ok(bson!({ "$sqlLt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Lt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        lte,
        expected = Ok(bson!({ "$sqlLte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Lte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        ne,
        expected = Ok(bson!({ "$sqlNe": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Ne,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        eq,
        expected = Ok(bson!({ "$sqlEq": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Eq,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
        gt,
        expected = Ok(bson!({ "$sqlGt": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Gt,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );
    test_codegen_expression!(
        gte,
        expected = Ok(bson!({ "$sqlGte": [{ "$literal": 1}, { "$literal": 2}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Gte,
            args: vec![Literal(Integer(1)), Literal(Integer(2))],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
        nullif_expr,
        expected = Ok(bson!({ "$nullIf": [{ "$literal": true}, { "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: NullIf,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );

    test_codegen_expression!(
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

    test_codegen_expression!(
        not,
        expected = Ok(bson!({ "$sqlNot": [{ "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Not,
            args: vec![Literal(Boolean(false))],
        })
    );

    test_codegen_expression!(
        and,
        expected = Ok(bson!({ "$sqlAnd": [{ "$literal": true}, { "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: And,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );
    test_codegen_expression!(
        or,
        expected = Ok(bson!({ "$sqlOr": [{ "$literal": true}, { "$literal": false}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Or,
            args: vec![Literal(Boolean(true)), Literal(Boolean(false))],
        })
    );
    test_codegen_expression!(
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
    test_codegen_expression!(
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
    test_codegen_expression!(
        index_of_cp,
        expected = Ok(
            bson!({ "$sqlIndexOfCP": [{ "$literal": "foo"}, { "$literal": "bar"}, { "$literal": 1}, { "$literal": 2}]})
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

    test_codegen_expression!(
        str_len_cp,
        expected = Ok(bson!({ "$sqlStrLenCP": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: StrLenCP,
            args: vec![Literal(String("foo".to_string())),],
        })
    );
    test_codegen_expression!(
        str_len_bytes,
        expected = Ok(bson!({ "$sqlStrLenBytes": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: StrLenBytes,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        bit_length,
        expected = Ok(bson!({ "$sqlBitLength": [{ "$literal": "foo"}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: BitLength,
            args: vec![Literal(String("foo".to_string())),],
        })
    );
    test_codegen_expression!(
        cos,
        expected = Ok(bson!({ "$sqlCos": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Cos,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_expression!(
        log,
        expected = Ok(bson!({ "$sqlLog": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Log,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_expression!(
        mod_op,
        expected = Ok(bson!({ "$sqlMod": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Mod,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_expression!(
        round,
        expected = Ok(bson!({ "$sqlRound": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Round,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_expression!(
        sin,
        expected = Ok(bson!({ "$sqlSin": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Sin,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_expression!(
        sqrt,
        expected = Ok(bson!({ "$sqlSqrt": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Sqrt,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_expression!(
        tan,
        expected = Ok(bson!({ "$sqlTan": [{"$literal": 3.5}]})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: Tan,
            args: vec![Literal(Double(3.5)),],
        })
    );
    test_codegen_expression!(
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
    test_codegen_expression!(
        to_upper,
        expected = Ok(bson!({ "$sqlToUpper": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: ToUpper,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_expression!(
        to_lower,
        expected = Ok(bson!({ "$sqlToLower": { "$literal": "foo"}})),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: ToLower,
            args: vec![Literal(String("foo".to_string())),],
        })
    );
    test_codegen_expression!(
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
    test_codegen_expression!(
        current_timestamp,
        expected = Ok(bson!("$$NOW")),
        input = SQLSemanticOperator(SQLSemanticOperator {
            op: CurrentTimestamp,
            args: vec![],
        })
    );
    test_codegen_expression!(
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

    test_codegen_expression!(
        empty,
        expected = Ok(bson!({"$literal": {}})),
        input = Document(unchecked_unique_linked_hash_map! {})
    );
    test_codegen_expression!(
        non_empty,
        expected = Ok(bson!({"foo": {"$literal": 1}})),
        input =
            Document(unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(Integer(1))})
    );
    test_codegen_expression!(
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

    test_codegen_expression!(empty, expected = Ok(bson!([])), input = Array(vec![]));

    test_codegen_expression!(
        non_empty,
        expected = Ok(bson!([{"$literal": "abc"}])),
        input = Array(vec![Literal(String("abc".to_string()))])
    );

    test_codegen_expression!(
        nested,
        expected = Ok(bson!([{ "$literal": null }, [{ "$literal": null }]])),
        input = Array(vec![Literal(Null), Array(vec![Literal(Null)])])
    );
}

mod variable {
    use crate::air::Expression::*;
    use bson::{bson, Bson};

    test_codegen_expression!(
        simple,
        expected = Ok(bson!(Bson::String("$$foo".to_string()))),
        input = Variable("foo".to_string().into())
    );

    test_codegen_expression!(
        nested,
        expected = Ok(bson!(Bson::String("$$x.y.z".to_string()))),
        input = Variable("x.y.z".to_string().into())
    );
}

mod field_ref {
    use crate::air::Expression::FieldRef;
    use bson::{bson, Bson};

    test_codegen_expression!(
        no_parent,
        expected = Ok(bson!(Bson::String("$foo".to_string()))),
        input = FieldRef("foo".to_string().into())
    );
    test_codegen_expression!(
        parent,
        expected = Ok(bson!(Bson::String("$bar.foo".to_string()))),
        input = FieldRef("bar.foo".to_string().into())
    );

    test_codegen_expression!(
        grandparent,
        expected = Ok(bson!(Bson::String("$baz.bar.foo".to_string()))),
        input = FieldRef("baz.bar.foo".to_string().into())
    );
}

mod like {
    use crate::air::{self, Expression};
    use bson::bson;

    test_codegen_expression!(
        with_escape,
        expected = Ok(bson!({"$like": {
            "input": "$input",
            "pattern": "$pattern",
            "escape": "escape",
        }})),
        input = Expression::Like(air::Like {
            expr: Box::new(Expression::FieldRef("input".to_string().into())),
            pattern: Box::new(Expression::FieldRef("pattern".to_string().into())),
            escape: Some("escape".to_string()),
        })
    );

    test_codegen_expression!(
        without_escape,
        expected = Ok(bson!({"$like": {
            "input": "$input",
            "pattern": "$pattern",
        }})),
        input = Expression::Like(air::Like {
            expr: Box::new(Expression::FieldRef("input".to_string().into())),
            pattern: Box::new(Expression::FieldRef("pattern".to_string().into())),
            escape: None,
        })
    );
}

mod is {
    use crate::air::{Expression::*, Is, Type, TypeOrMissing};
    use bson::bson;

    test_codegen_expression!(
        target_type_missing,
        expected = Ok(bson!({"$sqlIs": ["$x", {"$literal": "missing"}]})),
        input = Is(Is {
            expr: Box::new(FieldRef("x".to_string().into())),
            target_type: TypeOrMissing::Missing,
        })
    );

    test_codegen_expression!(
        target_type_number,
        expected = Ok(bson!({"$sqlIs": ["$x", {"$literal": "number"}]})),
        input = Is(Is {
            expr: Box::new(FieldRef("x".to_string().into())),
            target_type: TypeOrMissing::Number,
        })
    );

    test_codegen_expression!(
        target_type_type,
        expected = Ok(bson!({"$sqlIs": ["$x", {"$literal": "object"}]})),
        input = Is(Is {
            expr: Box::new(FieldRef("x".to_string().into())),
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

    test_codegen_expression!(
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

    test_codegen_expression!(
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
    use crate::air::{Expression::*, SetField};
    use bson::bson;

    test_codegen_expression!(
        simple,
        expected = Ok(bson!({"$setField": {"field": "", "input": "$$ROOT", "value": "$__bot"}})),
        input = SetField(SetField {
            field: "".into(),
            input: Box::new(Variable("ROOT".to_string().into())),
            value: Box::new(FieldRef("__bot".to_string().into()))
        })
    );

    test_codegen_expression!(
        use_literal_when_needed,
        expected = Ok(
            bson!({"$setField": {"field": {"$literal": "$x_val"}, "input": "$$ROOT", "value": "$x"}})
        ),
        input = SetField(SetField {
            field: "$x_val".into(),
            input: Box::new(Variable("ROOT".to_string().into())),
            value: Box::new(FieldRef("x".to_string().into()))
        })
    );
}

mod unset_field {
    use crate::air::{Expression::*, UnsetField};
    use bson::bson;

    test_codegen_expression!(
        simple,
        expected = Ok(bson!({"$unsetField": {"field": "__bot", "input": "$doc"}})),
        input = UnsetField(UnsetField {
            field: "__bot".into(),
            input: Box::new(FieldRef("doc".to_string().into()))
        })
    );

    test_codegen_expression!(
        use_literal_when_needed,
        expected = Ok(bson!({"$unsetField": {"field": {"$literal": "$x"}, "input": "$doc"}})),
        input = UnsetField(UnsetField {
            field: "$x".into(),
            input: Box::new(FieldRef("doc".to_string().into()))
        })
    );
}

mod sql_convert {
    use crate::{
        air::{Expression::*, LiteralValue, SqlConvert, SqlConvertTargetType},
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_expression!(
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
    test_codegen_expression!(
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

    test_codegen_expression!(
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
        test_codegen_expression!(
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

    test_codegen_expression!(
        convert_array,
        expected = Err(codegen::Error::ConvertToArray),
        input = Expression::Convert(Convert {
            input: Expression::Literal(String("foo".to_string())).into(),
            to: Type::Array,
            on_null: Expression::Literal(Null).into(),
            on_error: Expression::Literal(Null).into(),
        })
    );

    test_codegen_expression!(
        convert_document,
        expected = Err(codegen::Error::ConvertToDocument),
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
    use crate::air::{Expression::*, Let, LetVariable, MQLOperator, MQLSemanticOperator};
    use bson::bson;

    test_codegen_expression!(
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
                    FieldRef("x".to_string().into()),
                    FieldRef("y".to_string().into()),
                ],
            })),
        })
    );

    test_codegen_expression!(
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
                expr: Box::new(FieldRef("x".to_string().into())),
            }],
            inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Add,
                args: vec![
                    Variable("v_x".to_string().into()),
                    FieldRef("y".to_string().into()),
                ],
            })),
        })
    );

    test_codegen_expression!(
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
                    expr: Box::new(FieldRef("x".to_string().into())),
                },
                LetVariable {
                    name: "v_y".to_string(),
                    expr: Box::new(FieldRef("y".to_string().into())),
                },
                LetVariable {
                    name: "v_z".to_string(),
                    expr: Box::new(FieldRef("z".to_string().into())),
                },
            ],
            inside: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::Add,
                args: vec![
                    Variable("v_x".to_string().into()),
                    Variable("v_y".to_string().into()),
                    Variable("v_z".to_string().into()),
                ],
            })),
        })
    );
}

mod regex_match {
    use crate::air::{Expression::*, LiteralValue, RegexMatch};
    use bson::bson;

    test_codegen_expression!(
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

    test_codegen_expression!(
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

mod reduce {
    use crate::air::{Expression::*, LiteralValue, Reduce};
    use bson::bson;

    test_codegen_expression!(
        simple,
        expected = Ok(
            bson!({ "$reduce": {"input": {"$literal": "input"}, "initialValue": {"$literal": "init"}, "in": {"$literal": "inside"}}})
        ),
        input = Reduce(Reduce {
            input: Box::new(Literal(LiteralValue::String("input".to_string()))),
            init_value: Box::new(Literal(LiteralValue::String("init".to_string()))),
            inside: Box::new(Literal(LiteralValue::String("inside".to_string()))),
        })
    );
}

mod subquery_exists {
    use crate::{
        air::{
            Collection, Expression::*, LetVariable, Project, ProjectItem, Stage::*, SubqueryExists,
        },
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_expression!(
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
                    "foo".to_string() => ProjectItem::Assignment(Variable("ROOT".to_string().into())),
                },
            })),
        })
    );

    test_codegen_expression!(
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
                expr: Box::new(FieldRef("foo".to_string().into())),
            },],
            pipeline: Box::new(Project(Project {
                source: Box::new(Project(Project {
                    source: Box::new(Collection(Collection {
                        db: "test".to_string(),
                        collection: "bar".to_string(),
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "bar".to_string() => ProjectItem::Assignment(Variable("ROOT".to_string().into())),
                    },
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => ProjectItem::Assignment(Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => Variable("vfoo_0.a".to_string().into()),
                    })),
                },
            })),
        })
    );
}

mod subquery {
    use crate::{
        air::{
            Collection, Documents, Expression::*, LetVariable, Project, ProjectItem, Stage::*,
            Subquery,
        },
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_expression!(
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
                    "arr".to_string() => ProjectItem::Assignment(Variable("ROOT".to_string().into())),
                }
            })),
        })
    );

    test_codegen_expression!(
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
                    expr: Box::new(FieldRef("foo".to_string().into())),
                },
                LetVariable {
                    name: "vbaz_0".to_string(),
                    expr: Box::new(FieldRef("baz".to_string().into())),
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
                        "bar".to_string() => ProjectItem::Assignment(Variable("ROOT".to_string().into())),
                    }
                })),
                specifications: unchecked_unique_linked_hash_map! {
                    "__bot".to_string() => ProjectItem::Assignment(Document(unchecked_unique_linked_hash_map! {
                        "a".to_string() => Variable("vfoo_0.a".to_string().into()),
                    })),
                }
            })),
        })
    );
}

mod subquery_comparison {
    use crate::air::ProjectItem;
    use crate::{
        air::{
            Collection, Documents, Expression::*, LetVariable, Project, Stage::*, Subquery,
            SubqueryComparison, SubqueryComparisonOp, SubqueryModifier,
        },
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_expression!(
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
            arg: Box::new(FieldRef("x".to_string().into())),
            subquery: Box::new(Subquery {
                let_bindings: vec![],
                output_path: vec!["arr".to_string()],
                pipeline: Box::new(Project(Project {
                    source: Box::new(Documents(Documents { array: vec![] })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "arr".to_string() => ProjectItem::Assignment(Variable("ROOT".to_string().into())),
                    }
                })),
            }),
        })
    );

    test_codegen_expression!(
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
            arg: Box::new(FieldRef("x".to_string().into())),
            subquery: Box::new(Subquery {
                let_bindings: vec![
                    LetVariable {
                        name: "vfoo_0".to_string(),
                        expr: Box::new(FieldRef("foo".to_string().into())),
                    },
                    LetVariable {
                        name: "vbaz_0".to_string(),
                        expr: Box::new(FieldRef("baz".to_string().into())),
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
                            "bar".to_string() => ProjectItem::Assignment(Variable("ROOT".to_string().into())),
                        }
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "__bot".to_string() => ProjectItem::Assignment(Document(unchecked_unique_linked_hash_map! {
                            "a".to_string() => Variable("vfoo_0.a".to_string().into()),
                        })),
                    }
                })),
            }),
        })
    );
}
