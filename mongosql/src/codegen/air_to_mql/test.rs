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

macro_rules! test_codegen_air_plan {
    (
		$func_name:ident,
		expected = Ok({
			database: $expected_db:expr,
			collection: $expected_collection:expr,
			pipeline: $expected_pipeline:expr,
		}),
		input = $input: expr,
	) => {
        #[test]
        fn $func_name() {
            use crate::codegen::{air_to_mql::MqlTranslation, generate_mql_from_air};

            let input = $input;
            let expected_db = $expected_db;
            let expected_collection = $expected_collection;
            let expected_pipeline = $expected_pipeline;

            let MqlTranslation {
                database: db,
                collection: col,
                pipeline: pipeline,
            } = generate_mql_from_air(input).expect("codegen failed");

            assert_eq!(expected_db, db);
            assert_eq!(expected_collection, col);
            assert_eq!(expected_pipeline, pipeline);
        }
    };

    ($func_name:ident, expected = Err($expected_err:expr), input = $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::codegen::generate_mql_from_air;

            let input = $input;
            let expected = Err($expected_err);

            assert_eq!(expected, generate_mql_from_air(input));
        }
    };
}

mod air_collection {
    use crate::air::*;

    test_codegen_air_plan!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: Vec::<bson::Document>::new(),
        }),
        input = Stage::Collection(Collection {
            db: "mydb".to_string(),
            collection: "col".to_string(),
        }),
    );
}

mod air_literal {
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

mod air_mql_semantic_operator {
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
        trim,
        expected = Ok(bson!({ "$trim": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: Trim,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        ltrim,
        expected = Ok(bson!({ "$ltrim": [{ "$literal": "foo"}]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: LTrim,
            args: vec![Literal(String("foo".to_string())),],
        })
    );

    test_codegen_air_expr!(
        rtrim,
        expected = Ok(bson!({ "$rtrim": [{ "$literal": "foo" }]})),
        input = MQLSemanticOperator(MQLSemanticOperator {
            op: RTrim,
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
}

mod air_document {
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

mod air_array {
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

mod air_variable {
    use crate::air::Expression::*;
    use bson::{bson, Bson};

    test_codegen_air_expr!(
        simple,
        expected = Ok(bson!(Bson::String("$$foo".to_string()))),
        input = Variable("foo".to_string())
    );
}

mod air_field_ref {
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
}

mod air_documents_stage {
    use crate::air::*;

    test_codegen_air_plan!(
        empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
            ],
        }),
        input = Stage::Documents(Documents {
            array: vec![],
        }),
    );
    test_codegen_air_plan!(
        non_empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": false}]},
            ],
        }),
        input = Stage::Documents(Documents {
            array: vec![Expression::Literal(LiteralValue::Boolean(false))],
        }),
    );
}

mod air_replace_root_stage {
    use crate::{air::*, unchecked_unique_linked_hash_map};

    test_codegen_air_plan!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {"$replaceRoot": {"newRoot": {"$literal": "$name"}}},
            ],
        }),
        input = Stage::ReplaceRoot(ReplaceRoot {
            source: Box::new(
                Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                }),
            ),
            new_root: Box::new(Expression::Literal(LiteralValue::String("$name".to_string()))),
        }),
    );
    test_codegen_air_plan!(
        document,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc! {
                   "$replaceRoot": {
                      "newRoot": {
                         "$mergeDocuments": [
                            {"$literal": "$name"},
                            {"_id": {"$literal": "$_id"}}
                         ]
                      }
                   }
                },
            ],
        }),
        input = Stage::ReplaceRoot(ReplaceRoot {
            source: Box::new(
                Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "col".to_string(),
                }),
            ),
            new_root: Box::new(
                Expression::Document(unchecked_unique_linked_hash_map! {
                    "$mergeDocuments".to_string() => Expression::Array(vec![
                        Expression::Literal(LiteralValue::String("$name".to_string())),
                        Expression::Document(unchecked_unique_linked_hash_map! {
                            "_id".to_string() => Expression::Literal(
                                LiteralValue::String("$_id".to_string())
                            )
                        })
                    ])
                })
            ),
        }),
    );
}
