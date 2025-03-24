macro_rules! test_serde_stage {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use super::TestStage;

            let input = $input;
            let s: TestStage = serde_yaml::from_str(&input).unwrap();

            assert_eq!($expected, s.stage, "failed to deserialize");

            // test roundtrip by serializing to string and then deserializing
            // again (so that we do not need to worry about white space, as we would
            // if we were comparing the original input string to the output string)
            let output = serde_json::to_string(&s).unwrap();
            let s: TestStage = serde_yaml::from_str(&output).unwrap();

            // we output the failed serialization as json since it's a bit easier to read
            // in that it looks more like agg
            assert_eq!($expected, s.stage, "failed to serialize: {}", output);
        }
    };
}

macro_rules! test_serde_expr {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::serde_test::expression_test::TestExpr;

            let input = $input;
            let e: TestExpr = serde_yaml::from_str(&input).unwrap();

            assert_eq!($expected, e.expr, "failed to deserialize");

            // test roundtrip by serializing to string and then deserializing
            // again (so that we do not need to worry about white space, as we would
            // if we were comparing the original input string to the output string)
            let output = serde_json::to_string(&e).unwrap();
            let e: TestExpr = serde_yaml::from_str(&output).unwrap();

            // we output the failed serialization as json since it's a bit easier to read
            // in that it looks more like agg
            assert_eq!($expected, e.expr, "failed to serialize: {}", output);
        }
    };
}

macro_rules! test_match_bin_op {
    ($func_name:ident, string_op = $string_op:expr, expected_op = $expected_op:expr) => {
        test_serde_stage!(
            $func_name,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Field(MatchField {
                    field: Ref::FieldRef("a".to_string()),
                    ops: map! { $expected_op =>  bson::Bson::Int32(1) }
                })]
            }),
            input = format!(r#"stage: {{"$match": {{"a": {{"{}": 1}}}}}}"#, $string_op)
        );
    };
}

macro_rules! test_match_logical_vararg {
    ($func_name:ident, string_op = $string_op:expr, expected_op = $expected_op:expr) => {
        test_serde_stage!(
            $func_name,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Logical($expected_op(vec![
                    MatchExpression::Field(MatchField {
                        field: Ref::FieldRef("a".to_string()),
                        ops: map! { MatchBinaryOp::Gt =>  bson::Bson::Int32(1) }
                    }),
                    MatchExpression::Field(MatchField {
                        field: Ref::FieldRef("a".to_string()),
                        ops: map! { MatchBinaryOp::Lt =>  bson::Bson::Int32(1) }
                    }),
                    MatchExpression::Field(MatchField {
                        field: Ref::FieldRef("a".to_string()),
                        ops: map! { MatchBinaryOp::Lte =>  bson::Bson::Int32(1) }
                    }),
                ])),]
            }),
            input = format!(
                r#"stage: {{"$match": {{"{}":
                [{{"a": {{"$gt": 1}}}}, {{"a": {{"$lt": 1}}}}, {{"a": {{"$lte": 1}}}}]
            }}}}"#,
                $string_op
            )
        );
    };
}

macro_rules! test_serde_date_operator {
    ($func_name:ident, string_op = $string_op:expr, expected_op = $expected_op:expr) => {
        test_serde_expr!(
            $func_name,
            expected = Expression::TaggedOperator($expected_op(DateExpression {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                ))))
            })),
            input = format!(
                r#"expr: {{"{}": {{"date": "$date", "timezone": "$timezone"}}}}"#,
                $string_op
            )
        );
    };
}

mod stage_test {
    use crate::definitions::Stage;
    use serde::{Deserialize, Serialize};

    #[derive(Debug, PartialEq, Serialize, Deserialize)]
    struct TestStage {
        stage: Stage,
    }

    mod documents {
        use crate::{
            definitions::{Expression, LiteralValue, Stage},
            map,
        };

        test_serde_stage!(
            empty,
            expected = Stage::Documents(vec![]),
            input = r#"stage: {"$documents": []}"#
        );

        test_serde_stage!(
            singleton,
            expected = Stage::Documents(vec![
                map! {"a".to_string() => Expression::Literal(LiteralValue::Int32(1)) }
            ]),
            input = r#"stage: {"$documents": [{"a": 1}]}"#
        );

        test_serde_stage!(
            multiple_elements,
            expected = Stage::Documents(vec![
                map! {
                    "a".to_string() => Expression::Literal(LiteralValue::Int32(1)),
                    "b".to_string() => Expression::Literal(LiteralValue::Int32(2)),
                },
                map! {
                    "a".to_string() => Expression::Literal(LiteralValue::String("yes".to_string())),
                    "b".to_string() => Expression::Literal(LiteralValue::Null),
                },
                map! {
                    "a".to_string() => Expression::Document(map! {
                        "b".to_string() => Expression::Document(map! {
                            "c".to_string() => Expression::Literal(LiteralValue::Boolean(true)),
                        }),
                    }),
                },
            ]),
            input = r#"stage: {"$documents": [
                                {"a": 1, "b": 2},
                                {"a": "yes", "b": null},
                                {"a": {"b": {"c": true}}}
            ]}"#
        );
    }

    mod project {
        use crate::{
            definitions::{
                Expression, LiteralValue, ProjectItem, ProjectStage, Ref, Stage, UntaggedOperator,
                UntaggedOperatorName,
            },
            map, ROOT_NAME,
        };

        test_serde_stage!(
            empty,
            expected = Stage::Project(ProjectStage { items: map! {} }),
            input = r#"stage: {"$project": {}}"#
        );

        test_serde_stage!(
            singleton_exclusion,
            expected = Stage::Project(ProjectStage {
                items: map! { "_id".to_string() => ProjectItem::Exclusion }
            }),
            input = r#"stage: {"$project": {"_id": 0}}"#
        );

        test_serde_stage!(
            singleton_inclusion,
            expected = Stage::Project(ProjectStage {
                items: map! { "_id".to_string() => ProjectItem::Inclusion }
            }),
            input = r#"stage: {"$project": {"_id": 1}}"#
        );

        test_serde_stage!(
            singleton_assignment,
            expected = Stage::Project(ProjectStage {
                items: map! { "_id".to_string() => ProjectItem::Assignment(Expression::Literal(LiteralValue::String("hello".to_string()))) }
            }),
            input = r#"stage: {"$project": {"_id": "hello"}}"#
        );

        test_serde_stage!(
            multiple_elements,
            expected = Stage::Project(ProjectStage {
                items: map! {
                    "_id".to_string() => ProjectItem::Exclusion,
                    "foo".to_string() => ProjectItem::Assignment(Expression::Ref(Ref::VariableRef(ROOT_NAME.to_string()))),
                    "bar".to_string() => ProjectItem::Assignment(Expression::Ref(Ref::FieldRef("bar".to_string()))),
                    "a".to_string() => ProjectItem::Assignment(Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Add,
                        args: vec![
                            Expression::Literal(LiteralValue::Int32(1)),
                            Expression::Literal(LiteralValue::Int32(2)),
                        ]
                    })),
                    "x".to_string() => ProjectItem::Assignment(Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Literal,
                        args: vec![
                            Expression::Literal(LiteralValue::Int32(0)),
                        ]
                    })),
                    "y".to_string() => ProjectItem::Assignment(Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Literal,
                        args: vec![
                            Expression::Literal(LiteralValue::Int32(1)),
                        ]
                    })),
                }
            }),
            input = r#"stage: {"$project": {
                                "_id": 0,
                                "foo": "$$ROOT",
                                "bar": "$bar",
                                "a": {"$add": [1, 2]},
                                "x": { "$literal": 0 },
                                "y": { "$literal": 1 },
            }}"#
        );
    }

    mod replace_with {
        use crate::{
            definitions::{
                Expression, Ref, ReplaceStage, Stage, UntaggedOperator, UntaggedOperatorName,
            },
            ROOT_NAME,
        };

        test_serde_stage!(
            simple,
            expected = Stage::ReplaceWith(ReplaceStage::Expression(Expression::Ref(
                Ref::FieldRef("a".to_string())
            ))),
            input = r#"stage: {"$replaceWith": "$a"}"#
        );

        test_serde_stage!(
            complex,
            expected = Stage::ReplaceWith(ReplaceStage::Expression(Expression::UntaggedOperator(
                UntaggedOperator {
                    op: UntaggedOperatorName::MergeObjects,
                    args: vec![
                        Expression::Ref(Ref::VariableRef(ROOT_NAME.to_string())),
                        Expression::Ref(Ref::FieldRef("as".to_string())),
                    ]
                }
            ))),
            input = r#"stage: {"$replaceWith": {"$mergeObjects": ["$$ROOT", "$as"]}}"#
        );

        test_serde_stage!(
            replace_root,
            expected = Stage::ReplaceWith(ReplaceStage::NewRoot(Expression::Ref(Ref::FieldRef(
                "n".to_string()
            )))),
            input = r#"stage: {"$replaceRoot": {"newRoot": "$n"}}"#
        );
    }

    mod elem_match {
        use crate::{
            definitions::{
                MatchArrayExpression, MatchArrayQuery, MatchBinaryOp, MatchElement,
                MatchExpression, MatchField, MatchLogical, MatchMisc, MatchStage, Ref, Stage,
            },
            map,
        };

        test_serde_stage!(
            elem_match_value,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::Element(MatchElement {
                    field: Ref::FieldRef("x".to_string()),
                    query: MatchArrayExpression::Value(map! {
                        MatchBinaryOp::Gt => bson::Bson::Int32(1),
                        MatchBinaryOp::Lt => bson::Bson::Int32(3)
                    })
                }))]
            }),
            input = r#"stage: {$match: {x: {$elemMatch: {$gt: 1, $lt: 3}}}}"#
        );

        test_serde_stage!(
            elem_match_fields_binaries,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::Element(MatchElement {
                    field: Ref::FieldRef("x".to_string()),
                    query: MatchArrayExpression::Query(MatchArrayQuery {
                        query: vec![
                            MatchExpression::Field(MatchField {
                                field: Ref::FieldRef("bar".to_string()),
                                ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(1) }
                            }),
                            MatchExpression::Field(MatchField {
                                field: Ref::FieldRef("foo".to_string()),
                                ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(1) }
                            })
                        ]
                    })
                }))]
            }),
            input = r#"stage: {$match: {x: {$elemMatch: {'bar': {'$eq': 1}, 'foo': 1}}}}"#
        );

        test_serde_stage!(
            elem_match_fields_or,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::Element(MatchElement {
                    field: Ref::FieldRef("x".to_string()),
                    query: MatchArrayExpression::Query(MatchArrayQuery {
                        query: vec![MatchExpression::Logical(MatchLogical::Or(vec![
                            MatchExpression::Field(MatchField {
                                field: Ref::FieldRef("foo".to_string()),
                                ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(1) }
                            }),
                            MatchExpression::Field(MatchField {
                                field: Ref::FieldRef("bar".to_string()),
                                ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(32) }
                            })
                        ]))]
                    })
                }))]
            }),
            input = r#"stage: {$match: {x: {$elemMatch: {$or: [{foo: {$eq: 1}}, {bar: 32}]}}}}"#
        );
    }

    mod match_stage {
        use crate::{
            definitions::{
                Expression, LiteralValue, MatchArrayExpression, MatchBinaryOp, MatchComment,
                MatchExpr, MatchExpression, MatchField, MatchJsonSchema, MatchLogical, MatchMisc,
                MatchNot, MatchNotExpression, MatchRegex, MatchStage, MatchText, MatchTextContents,
                MatchWhere, Ref, Stage, UntaggedOperator, UntaggedOperatorName,
            },
            map,
        };

        test_serde_stage!(
            expr,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Expr(MatchExpr {
                    expr: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::SQLEq,
                        args: vec![
                            Expression::Ref(Ref::FieldRef("a".to_string())),
                            Expression::Ref(Ref::FieldRef("b".to_string())),
                        ]
                    }))
                })]
            }),
            input = r#"stage: {"$match": {"$expr": {"$sqlEq": ["$a", "$b"]}}}"#
        );

        test_serde_stage!(
            implicit_eq,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Field(MatchField {
                    field: Ref::FieldRef("a".to_string()),
                    ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(1) }
                })]
            }),
            input = r#"stage: {"$match": {"a": 1}}"#
        );

        test_match_bin_op!(
            explicit_eq,
            string_op = "$eq",
            expected_op = MatchBinaryOp::Eq
        );
        test_match_bin_op!(gt, string_op = "$gt", expected_op = MatchBinaryOp::Gt);
        test_match_bin_op!(gte, string_op = "$gte", expected_op = MatchBinaryOp::Gte);
        test_match_bin_op!(in_op, string_op = "$in", expected_op = MatchBinaryOp::In);
        test_match_bin_op!(lt, string_op = "$lt", expected_op = MatchBinaryOp::Lt);
        test_match_bin_op!(lte, string_op = "$lte", expected_op = MatchBinaryOp::Lte);
        test_match_bin_op!(ne, string_op = "$ne", expected_op = MatchBinaryOp::Ne);
        test_match_bin_op!(nin, string_op = "$nin", expected_op = MatchBinaryOp::Nin);
        test_match_bin_op!(
            exists,
            string_op = "$exists",
            expected_op = MatchBinaryOp::Exists
        );
        test_match_bin_op!(
            type_op,
            string_op = "$type",
            expected_op = MatchBinaryOp::Type
        );
        test_match_bin_op!(
            size_op,
            string_op = "$size",
            expected_op = MatchBinaryOp::Size
        );
        test_match_bin_op!(mod_op, string_op = "$mod", expected_op = MatchBinaryOp::Mod);
        test_match_bin_op!(
            bits_any_set,
            string_op = "$bitsAnySet",
            expected_op = MatchBinaryOp::BitsAnySet
        );
        test_match_bin_op!(
            bits_any_clear,
            string_op = "$bitsAnyClear",
            expected_op = MatchBinaryOp::BitsAnyClear
        );
        test_match_bin_op!(
            bits_all_set,
            string_op = "$bitsAllSet",
            expected_op = MatchBinaryOp::BitsAllSet
        );
        test_match_bin_op!(
            bits_all_clear,
            string_op = "$bitsAllClear",
            expected_op = MatchBinaryOp::BitsAllClear
        );
        test_match_bin_op!(all, string_op = "$all", expected_op = MatchBinaryOp::All);
        test_match_bin_op!(
            geo_intersects,
            string_op = "$geoIntersects",
            expected_op = MatchBinaryOp::GeoIntersects
        );
        test_match_bin_op!(
            geo_within,
            string_op = "$geoWithin",
            expected_op = MatchBinaryOp::GeoWithin
        );
        test_match_bin_op!(near, string_op = "$near", expected_op = MatchBinaryOp::Near);
        test_match_bin_op!(
            near_sphere,
            string_op = "$nearSphere",
            expected_op = MatchBinaryOp::NearSphere
        );

        test_serde_stage!(
            multi_conditions_on_field,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Field(MatchField {
                    field: Ref::FieldRef("a".to_string()),
                    ops: map! {
                        MatchBinaryOp::Eq =>  bson::Bson::Int32(1),
                        MatchBinaryOp::Ne =>  bson::Bson::String("hello".to_string())
                    }
                }),]
            }),
            input = r#"stage: {"$match": {"a": {"$eq": 1, "$ne": "hello"}}}"#
        );

        test_serde_stage!(
            multi_fields_in_match_stage,
            expected = Stage::Match(MatchStage {
                expr: vec![
                    MatchExpression::Field(MatchField {
                        field: Ref::FieldRef("a".to_string()),
                        ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(1) }
                    }),
                    MatchExpression::Field(MatchField {
                        field: Ref::FieldRef("b".to_string()),
                        ops: map! { MatchBinaryOp::Ne =>  bson::Bson::String("hello".to_string()) }
                    }),
                ]
            }),
            input = r#"stage: {"$match": {"a": {"$eq": 1}, "b": {"$ne": "hello"}}}"#
        );

        test_match_logical_vararg!(or, string_op = "$or", expected_op = MatchLogical::Or);
        test_match_logical_vararg!(and, string_op = "$and", expected_op = MatchLogical::And);
        test_match_logical_vararg!(nor, string_op = "$nor", expected_op = MatchLogical::Nor);

        test_serde_stage!(
            not_element,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Logical(MatchLogical::Not(MatchNot {
                    field: Ref::FieldRef("a".to_string()),
                    expr: MatchNotExpression::Element(MatchArrayExpression::Value(map! {
                        MatchBinaryOp::Gt => bson::Bson::Int32(3),
                        MatchBinaryOp::Eq => bson::Bson::Int32(5),
                    }))
                })),]
            }),
            input = r#"stage: {"$match": {"a": {"$not": {"$elemMatch": {"$gt": 3, "$eq": 5}}}}}"#
        );

        test_serde_stage!(
            not_query,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Logical(MatchLogical::Not(MatchNot {
                    field: Ref::FieldRef("bar".to_string()),
                    expr: MatchNotExpression::Query(map! {
                        MatchBinaryOp::Ne => bson::Bson::Int32(42),
                        MatchBinaryOp::Gt => bson::Bson::Int32(50)
                    })
                })),]
            }),
            input = r#"stage: {"$match": {"bar": {"$not": {$ne: 42, $gt: 50}}}}"#
        );

        test_serde_stage!(
            not_regex,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Logical(MatchLogical::Not(MatchNot {
                    field: Ref::FieldRef("bar".to_string()),
                    expr: MatchNotExpression::Regex(bson::Bson::String("hello world!".to_string())),
                })),]
            }),
            input = r#"stage: {"$match": {"bar": {"$not": "hello world!"}}}"#
        );

        test_serde_stage!(
            where_expr,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::Where(MatchWhere {
                    code: bson::Bson::String("function() { return this.isGood == 42 }".to_string()),
                })),]
            }),
            input = r#"stage: {"$match": {"$where": "function() { return this.isGood == 42 }"}}"#
        );

        test_serde_stage!(
            text,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::Text(MatchText {
                    expr: MatchTextContents {
                        search: "Coffee".to_string(),
                        language: Some("English".to_string()),
                        case_sensitive: Some(true),
                        diacritic_sensitive: Some(false),
                    }
                }))]
            }),
            input = r#"stage: {"$match":  {"$text": {"$search": "Coffee", "$language": "English", "$caseSensitive": true, "$diacriticSensitive": false }}}"#
        );

        test_serde_stage!(
            json_schema,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::JsonSchema(
                    MatchJsonSchema {
                        schema: bson::Bson::Document(
                            map! { "bsonType".to_string() => bson::Bson::String("object".to_string()) }
                        ),
                    }
                )),]
            }),
            input = r#"stage: {"$match": {"$jsonSchema": {"bsonType": "object"}}}"#
        );

        test_serde_stage!(
            regex_no_options,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::Regex(MatchRegex {
                    field: Ref::FieldRef("x".to_string()),
                    pattern: bson::Bson::String("hello".to_string()),
                    options: None,
                })),]
            }),
            input = r#"stage: {"$match": {"x": {"$regex": "hello"}}}"#
        );

        test_serde_stage!(
            regex_options,
            expected = Stage::Match(MatchStage {
                expr: vec![MatchExpression::Misc(MatchMisc::Regex(MatchRegex {
                    field: Ref::FieldRef("x".to_string()),
                    pattern: bson::Bson::String("hello".to_string()),
                    options: Some(bson::Bson::String("i".to_string())),
                })),]
            }),
            input = r#"stage: {"$match": {"x": {"$regex": "hello", "$options": "i"}}}"#
        );

        test_serde_stage!(
            mixed_match_top_level,
            expected = Stage::Match(MatchStage {
                expr: vec![
                    MatchExpression::Misc(MatchMisc::Comment(MatchComment {
                        comment: "hello!".to_string()
                    })),
                    MatchExpression::Misc(MatchMisc::Where(MatchWhere {
                        code: bson::Bson::String(
                            "function() { return this.isGood == 42 }".to_string()
                        ),
                    })),
                    MatchExpression::Field(MatchField {
                        field: Ref::FieldRef("foo".to_string()),
                        ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(1) }
                    }),
                    MatchExpression::Logical(MatchLogical::Not(MatchNot {
                        field: Ref::FieldRef("bar".to_string()),
                        expr: MatchNotExpression::Query(map! {
                            MatchBinaryOp::Ne => bson::Bson::Int32(42)
                        })
                    })),
                    MatchExpression::Logical(MatchLogical::Or(vec![
                        MatchExpression::Expr(MatchExpr {
                            expr: Expression::UntaggedOperator(UntaggedOperator {
                                op: UntaggedOperatorName::Eq,
                                args: vec![
                                    Expression::Ref(Ref::FieldRef("a".to_string())),
                                    Expression::Literal(LiteralValue::Int32(1))
                                ]
                            })
                            .into()
                        }),
                        MatchExpression::Field(MatchField {
                            field: Ref::FieldRef("b".to_string()),
                            ops: map! { MatchBinaryOp::Eq =>  bson::Bson::Int32(2) }
                        }),
                    ])),
                ]
            }),
            input = r#"stage: {"$match": {"$comment": "hello!", "$where": "function() { return this.isGood == 42 }", "foo": 1, "bar": {"$not": {$ne: 42}},  "$or": [{$expr: {$eq: [$a, 1]}}, {b: 2}]}}"#
        );
    }

    mod const_stages {
        use crate::definitions::Stage;

        test_serde_stage!(
            limit,
            expected = Stage::Limit(10),
            input = r#"stage: {"$limit": 10}"#
        );

        test_serde_stage!(
            skip,
            expected = Stage::Skip(100),
            input = r#"stage: {"$skip": 100}"#
        );

        test_serde_stage!(
            count,
            expected = Stage::Count("a".to_string()),
            input = r#"stage: {"$count": "a"}"#
        );
    }

    mod sort {
        use crate::{
            definitions::{Expression, Ref, Stage},
            map,
        };

        test_serde_stage!(
            empty,
            expected = Stage::Sort(map! {}),
            input = r#"stage: {"$sort": {}}"#
        );

        test_serde_stage!(
            singleton,
            expected = Stage::Sort(map! { "a".to_string() => 1 }),
            input = r#"stage: {"$sort": {"a": 1}}"#
        );

        test_serde_stage!(
            multiple_elements,
            expected = Stage::Sort(map! { "a".to_string() => 1, "b".to_string() => -1 }),
            input = r#"stage: {"$sort": {"a": 1, "b": -1}}"#
        );

        test_serde_stage!(
            sort_by_count,
            expected =
                Stage::SortByCount(Box::new(Expression::Ref(Ref::FieldRef("f".to_string())))),
            input = r#"stage: {"$sortByCount": "$f"}"#
        );
    }

    mod unwind {
        use crate::definitions::{Expression, Ref, Stage, Unwind, UnwindExpr};

        test_serde_stage!(
            unwind_field_ref,
            expected = Stage::Unwind(Unwind::FieldPath(Expression::Ref(Ref::FieldRef(
                "eca58228-b657-498a-b76e-f48a9161a404".to_string()
            )))),
            input = r#"stage: { "$unwind": "$eca58228-b657-498a-b76e-f48a9161a404" }"#
        );

        test_serde_stage!(
            unwind_document_no_options,
            expected = Stage::Unwind(Unwind::Document(UnwindExpr {
                path: Box::new(Expression::Ref(Ref::FieldRef("array".to_string()))),
                include_array_index: None,
                preserve_null_and_empty_arrays: None
            })),
            input = r#"stage: {"$unwind": {"path": "$array"}}"#
        );

        test_serde_stage!(
            unwind_document_all_options,
            expected = Stage::Unwind(Unwind::Document(UnwindExpr {
                path: Box::new(Expression::Ref(Ref::FieldRef("array".to_string()))),
                include_array_index: Some("i".to_string()),
                preserve_null_and_empty_arrays: Some(true)
            })),
            input = r#"stage: {"$unwind": {"path": "$array", "includeArrayIndex": "i", "preserveNullAndEmptyArrays": true }}"#
        );
    }

    mod join {
        use crate::{
            definitions::{
                Expression, Join, JoinType, LiteralValue, ProjectItem, ProjectStage, Ref, Stage,
                UntaggedOperator, UntaggedOperatorName,
            },
            map,
        };

        test_serde_stage!(
            inner_join,
            expected = Stage::Join(Box::new(Join {
                database: None,
                collection: Some("bar".to_string()),
                let_body: None,
                join_type: JoinType::Inner,
                pipeline: vec![],
                condition: None
            })),
            input =
                r#"stage: {"$join": {"collection": "bar", "joinType": "inner", "pipeline": [] }}"#
        );

        test_serde_stage!(
            left_join_with_db,
            expected = Stage::Join(Box::new(Join {
                database: Some("db".to_string()),
                collection: Some("bar".to_string()),
                let_body: None,
                join_type: JoinType::Left,
                pipeline: vec![],
                condition: None
            })),
            input = r#"stage: { "$join":
                  {
                    "database": "db",
                    "collection": "bar",
                    "joinType": "left",
                    "pipeline": [],
                  },
              }"#
        );

        test_serde_stage!(
            join_with_no_collection_and_pipeline,
            expected = Stage::Join(Box::new(Join {
                database: None,
                collection: None,
                let_body: None,
                join_type: JoinType::Inner,
                pipeline: vec![Stage::Documents(vec![
                    map! {"a".to_string() => Expression::Literal(LiteralValue::Int32(1)) },
                    map! {"a".to_string() => Expression::Literal(LiteralValue::Int32(2)) },
                    map! {"a".to_string() => Expression::Literal(LiteralValue::Int32(3)) },
                ])],
                condition: None
            })),
            input = r#"stage: {
                "$join":
                  {
                    "joinType": "inner",
                    "pipeline":
                      [{ "$documents": [{ "a": 1 }, { "a": 2 }, { "a": 3 }] }],
                  },
              }"#
        );

        test_serde_stage!(
            join_with_let_vars_and_condition,
            expected = Stage::Join(Box::new(Join {
                database: None,
                collection: Some("bar".to_string()),
                let_body: Some(map! {
                    "x".to_string() => Expression::Ref(Ref::FieldRef("x".to_string()))
                }),
                join_type: JoinType::Inner,
                pipeline: vec![Stage::Project(ProjectStage {
                    items: map! {
                        "_id".to_string() => ProjectItem::Exclusion,
                        "x".to_string() => ProjectItem::Inclusion,
                    }
                })],
                condition: Some(Expression::UntaggedOperator(UntaggedOperator {
                    op: UntaggedOperatorName::SQLEq,
                    args: vec![
                        Expression::Ref(Ref::VariableRef("x".to_string())),
                        Expression::Ref(Ref::FieldRef("x".to_string())),
                    ]
                })),
            })),
            input = r#"stage: {
                "$join":
                  {
                    "collection": "bar",
                    "joinType": "inner",
                    "let": { "x": "$x" },
                    "pipeline": [{ "$project": { "_id": 0, "x": 1 } }],
                    "condition":
                      { "$sqlEq": ["$$x", "$x"] },
                  },
              }"#
        );

        test_serde_stage!(
            nested_join,
            expected = Stage::Join(Box::new(Join {
                database: None,
                collection: Some("bar".to_string()),
                let_body: None,
                join_type: JoinType::Inner,
                pipeline: vec![Stage::Join(Box::new(Join {
                    database: None,
                    collection: Some("baz".to_string()),
                    join_type: JoinType::Inner,
                    let_body: None,
                    pipeline: vec![Stage::Join(Box::new(Join {
                        database: None,
                        collection: Some("car".to_string()),
                        join_type: JoinType::Inner,
                        let_body: None,
                        pipeline: vec![],
                        condition: None
                    }))],
                    condition: None
                }))],
                condition: None
            })),
            input = r#"stage: {
                "$join":
                  {
                    "collection": "bar",
                    "joinType": "inner",
                    "pipeline":
                      [
                        {
                          "$join":
                            {
                              "collection": "baz",
                              "joinType": "inner",
                              "pipeline":
                                [
                                  {
                                    "$join":
                                      {
                                        "collection": "car",
                                        "joinType": "inner",
                                        "pipeline": [],
                                      },
                                  },
                                ],
                            },
                        },
                      ],
                  },
              }"#
        );
    }

    mod lookup_test {
        use crate::{
            definitions::{
                ConciseSubqueryLookup, EqualityLookup, Expression, LiteralValue, Lookup,
                LookupFrom, MatchExpr, MatchExpression, MatchStage, Namespace, ProjectItem,
                ProjectStage, Ref, Stage, SubqueryLookup, UntaggedOperator, UntaggedOperatorName,
            },
            map,
        };

        test_serde_stage!(
            subquery_lookup_with_no_optional_fields,
            expected = Stage::Lookup(Lookup::Subquery(SubqueryLookup {
                from: None,
                let_body: None,
                pipeline: vec![],
                as_var: "as_var".to_string()
            })),
            input = r#"stage: {"$lookup": {"pipeline": [], "as": "as_var"}}"#
        );
        test_serde_stage!(
            subquery_lookup_from_collection,
            expected = Stage::Lookup(Lookup::Subquery(SubqueryLookup {
                from: Some(LookupFrom::Collection("from_coll".to_string())),
                let_body: None,
                pipeline: vec![],
                as_var: "as_var".to_string()
            })),
            input = r#"stage: {"$lookup": {"from": "from_coll", "pipeline": [], "as": "as_var"}}"#
        );

        test_serde_stage!(
            subquery_lookup_from_namespace,
            expected = Stage::Lookup(Lookup::Subquery(SubqueryLookup {
                from: Some(LookupFrom::Namespace(Namespace {
                    db: "from_db".to_string(),
                    coll: "from_coll".to_string()
                })),
                let_body: None,
                pipeline: vec![],
                as_var: "as_var".to_string()
            })),
            input = r#"stage: {"$lookup": {"from": {"db": "from_db", "coll": "from_coll"}, "pipeline": [], "as": "as_var"}}"#
        );

        test_serde_stage!(
            subquery_lookup_with_single_let_var,
            expected = Stage::Lookup(Lookup::Subquery(SubqueryLookup {
                from: Some(LookupFrom::Namespace(Namespace {
                    db: "from_db".to_string(),
                    coll: "from_coll".to_string()
                })),
                let_body: Some(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Int32(9))
                }),
                pipeline: vec![],
                as_var: "as_var".to_string()
            })),
            input = r#"stage: {"$lookup": {
                "from": {"db": "from_db", "coll": "from_coll"},
                "let": {"x": 9},
                "pipeline": [],
                "as": "as_var"
            }}"#
        );

        test_serde_stage!(
            subquery_lookup_with_multiple_let_vars,
            expected = Stage::Lookup(Lookup::Subquery(SubqueryLookup {
                from: Some(LookupFrom::Namespace(Namespace {
                    db: "from_db".to_string(),
                    coll: "from_coll".to_string()
                })),
                let_body: Some(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Int32(9)),
                    "y".to_string() => Expression::Ref(Ref::FieldRef("z".to_string())),
                }),
                pipeline: vec![],
                as_var: "as_var".to_string()
            })),
            input = r#"stage: {"$lookup": {
                "from": {"db": "from_db", "coll": "from_coll"},
                "let": {
                    "x": 9,
                    "y": "$z"
                },
                "pipeline": [],
                "as": "as_var"
            }}"#
        );

        test_serde_stage!(
            subquery_lookup_with_pipeline,
            expected = Stage::Lookup(Lookup::Subquery(SubqueryLookup {
                from: Some(LookupFrom::Namespace(Namespace {
                    db: "db".to_string(),
                    coll: "bar".to_string()
                })),
                let_body: Some(map! {
                    "foo_b_0".to_string() => Expression::Ref(Ref::FieldRef("b".to_string())),
                }),
                pipeline: vec![
                    Stage::Match(MatchStage {
                        expr: vec![MatchExpression::Expr(MatchExpr {
                            expr: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                                op: UntaggedOperatorName::Eq,
                                args: vec![
                                    Expression::Ref(Ref::VariableRef("foo_b_0".to_string())),
                                    Expression::Ref(Ref::FieldRef("b".to_string()))
                                ]
                            }))
                        })]
                    }),
                    Stage::Project(ProjectStage {
                        items: map! {
                            "_id".to_string() => ProjectItem::Exclusion,
                            "a".to_string() => ProjectItem::Inclusion,
                        }
                    })
                ],
                as_var: "__subquery_result_0".to_string()
            })),
            input = r#"stage: {
                "$lookup":
                  {
                    "from": { "db": "db", "coll": "bar" },
                    "let": { "foo_b_0": "$b" },
                    "pipeline":
                      [
                        { "$match": { "$expr": { "$eq": ["$$foo_b_0", "$b"] } } },
                        { "$project": { "_id": 0, "a": 1 } },
                      ],
                    "as": "__subquery_result_0"
              }}"#
        );

        test_serde_stage!(
            concise_subquery_lookup_with_no_optional_fields,
            expected = Stage::Lookup(Lookup::ConciseSubquery(ConciseSubqueryLookup {
                from: None,
                let_body: None,
                pipeline: vec![],
                as_var: "as_var".to_string(),
                local_field: "foo".to_string(),
                foreign_field: "bar".to_string()
            })),
            input = r#"stage: {"$lookup": {"pipeline": [], "as": "as_var", "localField": "foo", "foreignField": "bar"}}"#
        );

        test_serde_stage!(
            concise_subquery_lookup_fully_specified,
            expected = Stage::Lookup(Lookup::ConciseSubquery(ConciseSubqueryLookup {
                from: Some(LookupFrom::Collection("coll".to_string())),
                let_body: Some(map! {
                    "foo_b_0".to_string() => Expression::Ref(Ref::FieldRef("b".to_string())),
                }),
                pipeline: vec![
                    Stage::Match(MatchStage {
                        expr: vec![MatchExpression::Expr(MatchExpr {
                            expr: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                                op: UntaggedOperatorName::Eq,
                                args: vec![
                                    Expression::Ref(Ref::VariableRef("foo_b_0".to_string())),
                                    Expression::Ref(Ref::FieldRef("b".to_string()))
                                ]
                            }))
                        })]
                    }),
                    Stage::Project(ProjectStage {
                        items: map! {
                            "_id".to_string() => ProjectItem::Exclusion,
                            "a".to_string() => ProjectItem::Inclusion,
                        }
                    })
                ],
                as_var: "__subquery_result_0".to_string(),
                local_field: "foo".to_string(),
                foreign_field: "bar".to_string()
            })),
            input = r#"stage: {
                "$lookup":
                  {
                    "from": "coll",
                    "let": { "foo_b_0": "$b" },
                    "pipeline":
                      [
                        { "$match": { "$expr": { "$eq": ["$$foo_b_0", "$b"] } } },
                        { "$project": { "_id": 0, "a": 1 } },
                      ],
                    "as": "__subquery_result_0",
                    "localField": "foo",
                    "foreignField": "bar"
              }}"#
        );

        test_serde_stage!(
            equality_lookup,
            expected = Stage::Lookup(Lookup::Equality(EqualityLookup {
                from: LookupFrom::Collection("coll".to_string()),
                as_var: "__subquery_result_0".to_string(),
                local_field: "foo".to_string(),
                foreign_field: "bar".to_string()
            })),
            input = r#"stage: { "$lookup": { "from": "coll", "as": "__subquery_result_0", "localField": "foo", "foreignField": "bar" }}"#
        );
    }

    mod group_test {
        use crate::{
            definitions::{
                Expression, Group, LiteralValue, Ref, SQLAccumulator, Stage, TaggedOperator,
                UntaggedOperator, UntaggedOperatorName,
            },
            map,
        };

        test_serde_stage!(
            group_null_id_no_acc,
            expected = Stage::Group(Group {
                keys: Expression::Literal(LiteralValue::Null),
                aggregations: map! {}
            }),
            input = r#"stage: {"$group": {
                "_id": null,
            }}"#
        );

        test_serde_stage!(
            group_with_single_acc,
            expected = Stage::Group(Group {
                keys: Expression::Literal(LiteralValue::Null),
                aggregations: map! {
                "acc".to_string() => Expression::TaggedOperator(TaggedOperator::SQLSum(
                    SQLAccumulator {
                        distinct: true,
                        var: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                        arg_is_possibly_doc: Some("not".to_string()),
                    }))
                }
            }),
            input = r#"stage: {
                "$group":
                  {
                    "_id": null,
                    "acc": { "$sqlSum": { "var": "$a", "distinct": true, "arg_is_possibly_doc": "not" } },
                  }
              }"#
        );

        test_serde_stage!(
            group_with_keys_and_multiple_acc,
            expected = Stage::Group(Group {
                keys: Expression::Document(map! {
                    "a".to_string() => Expression::Ref(Ref::FieldRef("a".to_string()))
                },),
                aggregations: map! {
                    "acc_one".to_string() => Expression::TaggedOperator(TaggedOperator::SQLSum(
                        SQLAccumulator {
                            distinct: true,
                            var: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                            arg_is_possibly_doc: None,
                        })
                    ),
                    "acc_two".to_string() => Expression::TaggedOperator(TaggedOperator::SQLAvg(
                        SQLAccumulator {
                            distinct: true,
                            var: Box::new(Expression::Ref(Ref::FieldRef("b".to_string()))),
                            arg_is_possibly_doc: None,
                        })
                    ),
                }
            }),
            input = r#"stage: {
                "$group":
                {
                    "_id": {"a": "$a"},
                    "acc_one": { "$sqlSum": { "var": "$a", "distinct": true } },
                    "acc_two": { "$sqlAvg": { "var": "$b", "distinct": true } },
                }
            }"#
        );

        test_serde_stage!(
            group_with_non_sql_acc,
            expected = Stage::Group(Group {
                keys: Expression::Literal(LiteralValue::Null),
                aggregations: map! {
                    "acc".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::AddToSet,
                        args: vec![Expression::Ref(Ref::FieldRef("a".to_string()))],
                    })
                }
            }),
            input = r#"stage: { "$group": { "_id": null, "acc": { "$addToSet": "$a" } } }"#
        );
    }

    mod add_fields {
        use crate::{
            definitions::{Expression, LiteralValue, Stage},
            map,
        };

        test_serde_stage!(
            empty,
            expected = Stage::AddFields(map! {}),
            input = r#"stage: {"$addFields": {}}"#
        );

        test_serde_stage!(
            single_field,
            expected = Stage::AddFields(map! {
                "a".to_string() => Expression::Literal(LiteralValue::Int32(1)),
            }),
            input = r#"stage: {"$addFields": {"a": 1}}"#
        );

        test_serde_stage!(
            multiple_fields,
            expected = Stage::AddFields(map! {
                "a".to_string() => Expression::Literal(LiteralValue::Int32(1)),
                "b".to_string() => Expression::Literal(LiteralValue::Boolean(false)),
                "c".to_string() => Expression::Literal(LiteralValue::Double(2.4)),
            }),
            input = r#"stage: {"$addFields": {"a": 1, "b": false, "c": 2.4}}"#
        );

        test_serde_stage!(
            set_alias,
            expected = Stage::AddFields(map! {
                "a".to_string() => Expression::Literal(LiteralValue::Int32(1)),
                "b".to_string() => Expression::Literal(LiteralValue::Boolean(false)),
                "c".to_string() => Expression::Literal(LiteralValue::Double(2.4)),
            }),
            input = r#"stage: {"$set": {"a": 1, "b": false, "c": 2.4}}"#
        );
    }

    mod redact {
        use crate::{
            definitions::{Expression, Ref, Stage},
            PRUNE_NAME,
        };

        test_serde_stage!(
            empty,
            expected = Stage::Redact(Box::new(Expression::Ref(Ref::VariableRef(
                PRUNE_NAME.to_string()
            )))),
            input = r#"stage: {"$redact": "$$PRUNE"}"#
        );
    }

    mod unset {
        use crate::definitions::{Stage, Unset};

        test_serde_stage!(
            single,
            expected = Stage::Unset(Unset::Single("foo".to_string())),
            input = r#"stage: {"$unset": "foo"}"#
        );

        test_serde_stage!(
            multiple,
            expected = Stage::Unset(Unset::Multiple(vec![
                "foo".to_string(),
                "bar".to_string(),
                "baz".to_string()
            ])),
            input = r#"stage: {"$unset": ["foo", "bar", "baz"]}"#
        );
    }

    mod set_window_fields {
        use crate::{
            definitions::{
                Derivative, EmptyDoc, Expression, LiteralValue, SetWindowFields,
                SetWindowFieldsOutput, Stage, TaggedOperator, UntaggedOperator,
                UntaggedOperatorName, Window,
            },
            map,
        };
        use bson::Bson;

        test_serde_stage!(
            only_output_empty,
            expected = Stage::SetWindowFields(SetWindowFields {
                partition_by: None,
                sort_by: None,
                output: map! {},
            }),
            input = r#"stage: {"$setWindowFields": {"output": {}}}"#
        );

        test_serde_stage!(
            only_output_single,
            expected = Stage::SetWindowFields(SetWindowFields {
                partition_by: None,
                sort_by: None,
                output: map! {
                    "o1".to_string() => SetWindowFieldsOutput {
                        window_func: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                            op: UntaggedOperatorName::Sum,
                            args: vec![Expression::Literal(LiteralValue::Int32(1))],
                        })),
                        window: None,
                    }
                },
            }),
            input = r#"stage: {"$setWindowFields": {
                                    "output": {
                                        "o1": {
                                            "$sum": 1,
                                        },
                                    }
            }}"#
        );

        // This test covers the various forms of outputs:
        //   - Without "window"
        //   - With "window.documents"
        //   - With "window.range"
        //   - With "window.unit"
        test_serde_stage!(
            only_output_multiple,
            expected = Stage::SetWindowFields(SetWindowFields {
                partition_by: None,
                sort_by: None,
                output: map! {
                    "documents".to_string() => SetWindowFieldsOutput {
                        window_func: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                            op: UntaggedOperatorName::Sum,
                            args: vec![Expression::Literal(LiteralValue::Int32(1))],
                        })),
                        window: Some(Window {
                            documents: Some([Bson::Int64(-1), Bson::Int32(1)]),
                            range: None,
                            unit: None,
                        }),
                    },
                    "no_window".to_string() => SetWindowFieldsOutput {
                        window_func: Box::new(Expression::TaggedOperator(TaggedOperator::Derivative(Derivative {
                            input: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                            unit: Some("seconds".to_string()),
                        }))),
                        window: None,
                    },
                    "range_and_unit".to_string() => SetWindowFieldsOutput {
                        window_func: Box::new(Expression::TaggedOperator(TaggedOperator::DenseRank(EmptyDoc {}))),
                        window: Some(Window {
                            documents: None,
                            range: Some([Bson::Int64(-10), Bson::Int32(10)]),
                            unit: Some("seconds".to_string()),
                        }),
                    },
                },
            }),
            input = r#"stage: {"$setWindowFields": {
                                    "output": {
                                        "documents": {
                                            "$sum": 1,
                                            "window": {
                                                "documents": [-1, 1],
                                            },
                                        },
                                        "no_window": {
                                            "$derivative": {
                                                "input": 1,
                                                "unit": "seconds",
                                            },
                                        },
                                        "range_and_unit": {
                                            "$denseRank": {},
                                            "window": {
                                                "range": [-10, 10],
                                                "unit": "seconds",
                                            },
                                        },
                                    }
            }}"#
        );

        test_serde_stage!(
            fully_specified,
            expected = Stage::SetWindowFields(SetWindowFields {
                partition_by: Some(Box::new(Expression::Literal(LiteralValue::Int32(1)))),
                sort_by: Some(map! {
                    "a".to_string() => 1,
                    "b".to_string() => -1,
                }),
                output: map! {
                    "o1".to_string() => SetWindowFieldsOutput {
                        window_func: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                            op: UntaggedOperatorName::Sum,
                            args: vec![Expression::Literal(LiteralValue::Int32(1))],
                        })),
                        window: Some(Window {
                            documents: Some([Bson::Int32(1), Bson::Int32(2)]),
                            range: None,
                            unit: Some("seconds".to_string()),
                        })
                    }
                }
            }),
            input = r#"stage: {"$setWindowFields": {
                                    "partitionBy": 1,
                                    "sortBy": {
                                        "a": 1,
                                        "b": -1,
                                    },
                                    "output": {
                                        "o1": {
                                            "$sum": 1,
                                            "window": {
                                                "documents": [1, 2],
                                                "unit": "seconds"
                                            },
                                        },
                                    },
            }}"#
        );
    }

    mod bucket {
        use crate::{
            definitions::{
                Bucket, Expression, LiteralValue, Stage, UntaggedOperator, UntaggedOperatorName,
            },
            map,
        };
        use bson::Bson;

        test_serde_stage!(
            only_group_by_and_boundaries,
            expected = Stage::Bucket(Bucket {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                boundaries: vec![Bson::Int32(0), Bson::Int32(5)],
                default: None,
                output: None,
            }),
            input = r#"stage: {"$bucket": {
                "groupBy": 1,
                "boundaries": [0, 5],
            }}"#
        );

        test_serde_stage!(
            with_default,
            expected = Stage::Bucket(Bucket {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                boundaries: vec![Bson::Int32(0), Bson::Int32(5)],
                default: Some(Bson::Int32(10)),
                output: None,
            }),
            input = r#"stage: {"$bucket": {
                "groupBy": 1,
                "boundaries": [0, 5],
                "default": 10,
            }}"#
        );

        test_serde_stage!(
            fully_specified_with_one_output,
            expected = Stage::Bucket(Bucket {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                boundaries: vec![Bson::Int32(0), Bson::Int32(5)],
                default: Some(Bson::Int32(10)),
                output: Some(map! {
                    "o1".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Sum,
                        args: vec![Expression::Literal(LiteralValue::Int32(1))]
                    })
                }),
            }),
            input = r#"stage: {"$bucket": {
                "groupBy": 1,
                "boundaries": [0, 5],
                "default": 10,
                "output": {
                    "o1": { "$sum": 1 },
                }
            }}"#
        );

        test_serde_stage!(
            fully_specified_with_multiple_output,
            expected = Stage::Bucket(Bucket {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                boundaries: vec![Bson::Int32(0), Bson::Int32(5)],
                default: Some(Bson::Int32(10)),
                output: Some(map! {
                    "o1".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Sum,
                        args: vec![Expression::Literal(LiteralValue::Int32(1))]
                    }),
                    "o2".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Avg,
                        args: vec![Expression::Literal(LiteralValue::Int32(2))]
                    })
                }),
            }),
            input = r#"stage: {"$bucket": {
                "groupBy": 1,
                "boundaries": [0, 5],
                "default": 10,
                "output": {
                    "o1": { "$sum": 1 },
                    "o2": { "$avg": 2 },
                }
            }}"#
        );
    }

    mod bucket_auto {
        use crate::{
            definitions::{
                BucketAuto, Expression, LiteralValue, Stage, UntaggedOperator, UntaggedOperatorName,
            },
            map,
        };

        test_serde_stage!(
            only_group_by_and_buckets,
            expected = Stage::BucketAuto(BucketAuto {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                buckets: 5,
                output: None,
                granularity: None,
            }),
            input = r#"stage: {"$bucketAuto": {
                "groupBy": 1,
                "buckets": 5,
            }}"#
        );

        test_serde_stage!(
            with_granularity,
            expected = Stage::BucketAuto(BucketAuto {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                buckets: 10,
                output: None,
                granularity: Some("R5".to_string()),
            }),
            input = r#"stage: {"$bucketAuto": {
                "groupBy": 1,
                "buckets": 10,
                "granularity": "R5",
            }}"#
        );

        test_serde_stage!(
            fully_specified_with_one_output,
            expected = Stage::BucketAuto(BucketAuto {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                buckets: 2,
                output: Some(map! {
                    "o1".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Sum,
                        args: vec![Expression::Literal(LiteralValue::Int32(1))]
                    })
                }),
                granularity: Some("R40".to_string()),
            }),
            input = r#"stage: {"$bucketAuto": {
                "groupBy": 1,
                "buckets": 2,
                "output": {
                    "o1": { "$sum": 1 },
                },
                "granularity": "R40",
            }}"#
        );

        test_serde_stage!(
            fully_specified_with_multiple_output,
            expected = Stage::BucketAuto(BucketAuto {
                group_by: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                buckets: 3,
                output: Some(map! {
                    "o1".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Sum,
                        args: vec![Expression::Literal(LiteralValue::Int32(1))]
                    }),
                    "o2".to_string() => Expression::UntaggedOperator(UntaggedOperator {
                        op: UntaggedOperatorName::Avg,
                        args: vec![Expression::Literal(LiteralValue::Int32(2))]
                    })
                }),
                granularity: Some("E6".to_string()),
            }),
            input = r#"stage: {"$bucketAuto": {
                "groupBy": 1,
                "buckets": 3,
                "output": {
                    "o1": { "$sum": 1 },
                    "o2": { "$avg": 2 },
                },
                "granularity": "E6",
            }}"#
        );
    }

    mod densify {
        use crate::definitions::{Densify, DensifyRange, DensifyRangeBounds, Stage};
        use bson::Bson;

        test_serde_stage!(
            bounds_full,
            expected = Stage::Densify(Densify {
                field: "f".to_string(),
                partition_by_fields: None,
                range: DensifyRange {
                    step: Bson::Int32(1),
                    bounds: DensifyRangeBounds::Full,
                    unit: None,
                },
            }),
            input = r#"stage: {"$densify": {
                "field": "f",
                "range": {
                    "step": 1,
                    "bounds": "full",
                },
            }}"#
        );

        test_serde_stage!(
            bounds_partition,
            expected = Stage::Densify(Densify {
                field: "f".to_string(),
                partition_by_fields: None,
                range: DensifyRange {
                    step: Bson::Int32(1),
                    bounds: DensifyRangeBounds::Partition,
                    unit: None,
                },
            }),
            input = r#"stage: {"$densify": {
                "field": "f",
                "range": {
                    "step": 1,
                    "bounds": "partition",
                },
            }}"#
        );

        test_serde_stage!(
            bounds_array,
            expected = Stage::Densify(Densify {
                field: "f".to_string(),
                partition_by_fields: None,
                range: DensifyRange {
                    step: Bson::Int32(1),
                    bounds: DensifyRangeBounds::Array(Box::new([Bson::Int32(1), Bson::Int32(2)])),
                    unit: None,
                },
            }),
            input = r#"stage: {"$densify": {
                "field": "f",
                "range": {
                    "step": 1,
                    "bounds": [1, 2],
                },
            }}"#
        );

        test_serde_stage!(
            fully_specified,
            expected = Stage::Densify(Densify {
                field: "d".to_string(),
                partition_by_fields: Some(vec!["x".to_string(), "y".to_string(), "z".to_string()]),
                range: DensifyRange {
                    step: Bson::Int32(1),
                    bounds: DensifyRangeBounds::Full,
                    unit: Some("second".to_string()),
                },
            }),
            input = r#"stage: {"$densify": {
                "field": "d",
                partitionByFields: ["x", "y", "z"],
                "range": {
                    "step": 1,
                    "bounds": "full",
                    "unit": "second",
                },
            }}"#
        );
    }

    mod facet {
        use crate::{
            definitions::{ProjectItem, ProjectStage, Stage},
            map,
        };

        test_serde_stage!(
            empty,
            expected = Stage::Facet(map! {}),
            input = r#"stage: {"$facet": {}}"#
        );

        test_serde_stage!(
            single,
            expected = Stage::Facet(map! {
                "outputField1".to_string() => vec![Stage::Count("x".to_string())]
            }),
            input = r#"stage: {"$facet": {
                "outputField1": [{"$count": "x"}]
            }}"#
        );

        test_serde_stage!(
            multiple,
            expected = Stage::Facet(map! {
                "o1".to_string() => vec![Stage::Limit(10)],
                "outputField2".to_string() => vec![
                    Stage::Project(ProjectStage {
                        items: map! {
                            "_id".to_string() => ProjectItem::Exclusion,
                        },
                    }),
                    Stage::Count("x".to_string()),
                ],
            }),
            input = r#"stage: {"$facet": {
                "o1": [{"$limit": 10}],
                "outputField2": [{"$project": {"_id": 0}}, {"$count": "x"}],
            }}"#
        );
    }

    mod fill {
        use crate::{
            definitions::{Expression, Fill, FillOutput, FillOutputMethod, LiteralValue, Stage},
            map,
        };

        test_serde_stage!(
            with_partition_by,
            expected = Stage::Fill(Fill {
                partition_by: Some(Box::new(Expression::Literal(LiteralValue::Int32(10)))),
                partition_by_fields: None,
                sort_by: None,
                output: map! {
                    "x".to_string() => FillOutput::Value(Expression::Literal(LiteralValue::Int32(1))),
                }
            }),
            input = r#"stage: {"$fill": {
                "partitionBy": 10,
                "output": {
                    "x": {"value": 1},
                },
            }}"#
        );

        test_serde_stage!(
            with_partition_by_fields,
            expected = Stage::Fill(Fill {
                partition_by: None,
                partition_by_fields: Some(vec!["x".to_string(), "y".to_string()]),
                sort_by: None,
                output: map! {
                    "x".to_string() => FillOutput::Value(Expression::Literal(LiteralValue::Int32(1))),
                }
            }),
            input = r#"stage: {"$fill": {
                "partitionByFields": ["x", "y"],
                "output": {
                    "x": {"value": 1},
                },
            }}"#
        );

        test_serde_stage!(
            with_sort_by_linear,
            expected = Stage::Fill(Fill {
                partition_by: None,
                partition_by_fields: None,
                sort_by: Some(map! {
                    "x".to_string() => -1i8,
                }),
                output: map! {
                    "x".to_string() => FillOutput::Method(FillOutputMethod::Linear),
                }
            }),
            input = r#"stage: {"$fill": {
                "sortBy": {"x": -1},
                "output": {
                    "x": {"method": "linear"},
                },
            }}"#
        );

        test_serde_stage!(
            with_sort_by_locf,
            expected = Stage::Fill(Fill {
                partition_by: None,
                partition_by_fields: None,
                sort_by: Some(map! {
                    "x".to_string() => -1i8,
                }),
                output: map! {
                    "x".to_string() => FillOutput::Method(FillOutputMethod::Locf),
                }
            }),
            input = r#"stage: {"$fill": {
                "sortBy": {"x": -1},
                "output": {
                    "x": {"method": "locf"},
                },
            }}"#
        );
    }

    mod geo_near {
        use crate::{
            definitions::{
                GeoJSON, GeoNear, GeoNearPoint, MatchBinaryOp, MatchExpression, MatchField, Ref,
                Stage,
            },
            map,
        };
        use bson::Bson;

        test_serde_stage!(
            with_no_optional_fields,
            expected = Stage::GeoNear(GeoNear {
                distance_field: "f".to_string(),
                distance_multiplier: None,
                include_locs: None,
                key: None,
                max_distance: None,
                min_distance: None,
                near: GeoNearPoint::GeoJSON(GeoJSON {
                    r#type: "Point".to_string(),
                    coordinates: [Bson::Double(-73.856077), Bson::Double(40.848447)],
                }),
                query: None,
                spherical: None,
            }),
            input = r#"stage: {"$geoNear": {
                "distanceField": "f",
                "near": {
                    "type": "Point",
                    "coordinates": [-73.856077, 40.848447],
                }
            }}"#
        );

        test_serde_stage!(
            fully_specified,
            expected = Stage::GeoNear(GeoNear {
                distance_field: "f".to_string(),
                distance_multiplier: Some(Bson::Int32(3)),
                include_locs: Some("locs".to_string()),
                key: Some("idx".to_string()),
                max_distance: Some(Bson::Int32(100)),
                min_distance: Some(Bson::Int32(10)),
                near: GeoNearPoint::Legacy([Bson::Double(-51.634855), Bson::Double(51.959558)]),
                query: Some(MatchExpression::Field(MatchField {
                    field: Ref::FieldRef("x".to_string()),
                    ops: map! {
                        MatchBinaryOp::Eq => Bson::Int32(42),
                    },
                })),
                spherical: Some(true),
            }),
            input = r#"stage: {"$geoNear": {
                "distanceField": "f",
                "distanceMultiplier": 3,
                "includeLocs": "locs",
                "key": "idx",
                "maxDistance": 100,
                "minDistance": 10,
                "near": [-51.634855, 51.959558],
                "query": {"x": 42},
                "spherical": true,
            }}"#
        );
    }

    mod sample {
        use crate::definitions::{Sample, Stage};

        test_serde_stage!(
            simple,
            expected = Stage::Sample(Sample { size: 500 }),
            input = r#"stage: {"$sample": {"size": 500}}"#
        );
    }

    mod union_with {
        use crate::definitions::{Stage, UnionWith, UnionWithPipeline};

        test_serde_stage!(
            empty_pipeline,
            expected = Stage::UnionWith(UnionWith::Pipeline(UnionWithPipeline {
                coll: Some("empty".to_string()),
                pipeline: Some(vec![]),
            })),
            input = r#"stage: {"$unionWith": {
                "coll": "empty",
                "pipeline": []
            }}"#
        );

        test_serde_stage!(
            singleton_pipeline,
            expected = Stage::UnionWith(UnionWith::Pipeline(UnionWithPipeline {
                coll: Some("single".to_string()),
                pipeline: Some(vec![Stage::Limit(10)]),
            })),
            input = r#"stage: {"$unionWith": {
                "coll": "single",
                "pipeline": [{"$limit": 10}]
            }}"#
        );

        test_serde_stage!(
            multiple_element_pipeline,
            expected = Stage::UnionWith(UnionWith::Pipeline(UnionWithPipeline {
                coll: Some("multiple".to_string()),
                pipeline: Some(vec![Stage::Skip(5), Stage::Limit(10)]),
            })),
            input = r#"stage: {"$unionWith": {
                "coll": "multiple",
                "pipeline": [{"$skip": 5}, {"$limit": 10}]
            }}"#
        );

        test_serde_stage!(
            collection,
            expected = Stage::UnionWith(UnionWith::Collection("coll".to_string())),
            input = r#"stage: {"$unionWith": "coll"}"#
        );
    }

    mod search_stages {
        use crate::{
            definitions::{AtlasSearchStage, Expression, GraphLookup, LiteralValue, Ref, Stage},
            map,
        };

        test_serde_stage!(
            graph_lookup,
            expected = Stage::GraphLookup(GraphLookup {
                from: "start".to_string(),
                start_with: Box::new(Expression::Ref(Ref::FieldRef("start".to_string()))),
                connect_from_field: "start".to_string(),
                connect_to_field: "end".to_string(),
                as_var: "path".to_string(),
                max_depth: Some(5),
                depth_field: Some("depth".to_string()),
                restrict_search_with_match: Some(Box::new(Expression::Document(
                    map! { "sql".to_string() => Expression::Literal(LiteralValue::Boolean(true)) }
                ))),
            }),
            input = r#"stage: {"$graphLookup": {
                "from": "start",
                "startWith": "$start",
                "connectFromField": "start",
                "connectToField": "end",
                "as": "path",
                "maxDepth": 5,
                "depthField": "depth",
                "restrictSearchWithMatch": { "sql": true }
            }}"#
        );

        // testing every possible permutation of $search, $searchMeta, and $vectorSearch isn't something
        // we need to do since we do not have plans to inspect them
        test_serde_stage!(
            search,
            expected = Stage::AtlasSearchStage(AtlasSearchStage::Search(Box::new(
                Expression::Document(
                    map! {"autocomplete".to_string() => Expression::Document(map! {"query".to_string() => Expression::Literal(LiteralValue::String("off".to_string())), "path".to_string() => Expression::Literal(LiteralValue::String("title".to_string()))})}
                )
            ))),
            input = r#"stage: {"$search": {"autocomplete": {"query": "off", "path": "title"}}}"#
        );

        test_serde_stage!(
            search_meta,
            expected = Stage::AtlasSearchStage(AtlasSearchStage::SearchMeta(Box::new(
                Expression::Document(
                    map! {"searchTerm".to_string() => Expression::Literal(LiteralValue::String("off".to_string()))}
                )
            ))),
            input = r#"stage: {"$searchMeta": {"searchTerm": "off"}}"#
        );

        test_serde_stage!(
            vector_search,
            expected = Stage::AtlasSearchStage(AtlasSearchStage::VectorSearch(Box::new(
                Expression::Document(
                    map! {"vectorSearch".to_string() => Expression::Document(map! {"query".to_string() => Expression::Literal(LiteralValue::String("off".to_string())), "path".to_string() => Expression::Literal(LiteralValue::String("title".to_string()))})}
                )
            ))),
            input =
                r#"stage: {"$vectorSearch": {"vectorSearch": {"query": "off", "path": "title"}}}"#
        );
    }
}

mod expression_test {
    use crate::definitions::Expression;
    use serde::{Deserialize, Serialize};

    #[derive(Debug, PartialEq, Deserialize, Serialize)]
    struct TestExpr {
        expr: Expression,
    }

    mod literal {
        use crate::definitions::{Expression, LiteralValue};
        // These tests are complete for the bson types actually supported by the bson crate's
        // extended json parser. Other ones come back as untagged operators or, bizarrely for
        // generic binary, as an array of bytes.
        //
        // We may want to consider adding tests that test actual bson bytes, but there's no
        // reason to assume these won't work.

        test_serde_expr!(
            null,
            expected = Expression::Literal(LiteralValue::Null),
            input = r#"expr: null"#
        );

        test_serde_expr!(
            boolean_true,
            expected = Expression::Literal(LiteralValue::Boolean(true)),
            input = r#"expr: true"#
        );

        test_serde_expr!(
            boolean_false,
            expected = Expression::Literal(LiteralValue::Boolean(false)),
            input = r#"expr: false"#
        );

        test_serde_expr!(
            int,
            expected = Expression::Literal(LiteralValue::Int32(1)),
            input = r#"expr: 1"#
        );

        test_serde_expr!(
            long,
            expected = Expression::Literal(LiteralValue::Int64(2147483648)),
            input = r#"expr: 2147483648"#
        );

        test_serde_expr!(
            double,
            expected = Expression::Literal(LiteralValue::Double(1.5)),
            input = r#"expr: 1.5"#
        );

        test_serde_expr!(
            string,
            expected = Expression::Literal(LiteralValue::String("yes".to_string())),
            input = r#"expr: "yes""#
        );

        test_serde_expr!(
            oid,
            expected = Expression::Literal(LiteralValue::ObjectId(
                bson::oid::ObjectId::parse_str("5d505646cf6d4fe581014ab2").unwrap()
            )),
            input = r#"expr: {"$oid": "5d505646cf6d4fe581014ab2"}"#
        );

        test_serde_expr!(
            max_key,
            expected = Expression::Literal(LiteralValue::MaxKey),
            input = r#"expr: {"$maxKey": 1}"#
        );

        test_serde_expr!(
            min_key,
            expected = Expression::Literal(LiteralValue::MinKey),
            input = r#"expr: {"$minKey": 1}"#
        );

        test_serde_expr!(
            uuid,
            expected = Expression::Literal(LiteralValue::Binary(bson::Binary {
                subtype: bson::spec::BinarySubtype::Uuid,
                bytes: vec![
                    147, 109, 160, 31, 154, 189, 77, 157, 128, 199, 2, 175, 133, 200, 34, 168
                ],
            })),
            input = r#"expr: {"$uuid": "936da01f-9abd-4d9d-80c7-02af85c822a8"}"#
        );
    }

    mod string_or_ref {
        use crate::definitions::{Expression, LiteralValue, Ref};

        test_serde_expr!(
            string,
            expected = Expression::Literal(LiteralValue::String("yes".to_string())),
            input = r#"expr: "yes""#
        );

        test_serde_expr!(
            simple_field_ref,
            expected = Expression::Ref(Ref::FieldRef("a".to_string())),
            input = r#"expr: "$a""#
        );

        test_serde_expr!(
            nested_field_ref,
            expected = Expression::Ref(Ref::FieldRef("a.b.c".to_string())),
            input = r#"expr: "$a.b.c""#
        );

        test_serde_expr!(
            variable,
            expected = Expression::Ref(Ref::VariableRef("v".to_string())),
            input = r#"expr: "$$v""#
        );
    }

    mod array {
        use crate::definitions::{Expression, LiteralValue};

        test_serde_expr!(
            empty,
            expected = Expression::Array(vec![]),
            input = r#"expr: []"#
        );

        test_serde_expr!(
            singleton,
            expected = Expression::Array(vec![Expression::Literal(LiteralValue::Int32(1))]),
            input = r#"expr: [1]"#
        );

        test_serde_expr!(
            multiple_elements,
            expected = Expression::Array(vec![
                Expression::Literal(LiteralValue::Int32(1)),
                Expression::Literal(LiteralValue::String("yes".to_string())),
                Expression::Array(vec![
                    Expression::Literal(LiteralValue::Boolean(true)),
                    Expression::Literal(LiteralValue::Double(4.1)),
                ]),
            ]),
            input = r#"expr: [1, "yes", [true, 4.1]]"#
        );
    }

    mod document {
        use crate::{
            definitions::{Expression, LiteralValue},
            map,
        };

        test_serde_expr!(
            empty,
            expected = Expression::Document(map! {}),
            input = r#"expr: {}"#
        );

        test_serde_expr!(
            singleton,
            expected = Expression::Document(
                map! {"a".to_string() => Expression::Literal(LiteralValue::Int32(1))}
            ),
            input = r#"expr: {"a": 1}"#
        );

        test_serde_expr!(
            multiple_elements,
            expected = Expression::Document(map! {
                "a".to_string() => Expression::Literal(LiteralValue::Int32(1)),
                "b".to_string() => Expression::Literal(LiteralValue::String("two".to_string())),
                "c".to_string() => Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Boolean(false))
                }),
            }),
            input = r#"expr: {"a": 1, "b": "two", "c": {"x": false}}"#
        );

        test_serde_expr!(
            similar_to_op_but_no_dollarx,
            expected = Expression::Document(map! {
                "notOp".to_string() => Expression::Array(vec![
                    Expression::Literal(LiteralValue::Int32(1)),
                    Expression::Literal(LiteralValue::Int32(2)),
                    Expression::Literal(LiteralValue::Int32(3)),
                ])
            }),
            input = r#"expr: {"notOp": [1, 2, 3]}"#
        );
    }

    mod tagged_operators {
        use crate::{
            definitions::{
                Accumulator, Bottom, BottomN, Convert, DateAdd, DateDiff, DateExpression,
                DateFromParts, DateFromString, DateSubtract, DateToParts, DateToString, DateTrunc,
                Expression, Filter, Function, GetField, Let, Like, LiteralValue, Map, Median,
                NArrayOp, Percentile, ProjectItem, ProjectStage, Reduce, Ref, RegexAggExpression,
                Replace, SQLConvert, SQLDivide, SetField, SortArray, SortArraySpec, Stage,
                Subquery, SubqueryComparison, SubqueryExists, Switch, SwitchCase, TaggedOperator,
                Top, TopN, Trim, UnsetField, UntaggedOperator, UntaggedOperatorName, Zip,
            },
            map,
        };

        test_serde_expr!(
            accumulator_all_args,
            expected = Expression::TaggedOperator(TaggedOperator::Accumulator(Accumulator {
                init: Box::new(Expression::Literal(LiteralValue::String(
                    "function (y) { return y; }".to_string()
                ))),
                init_args: Some(vec![Expression::Literal(LiteralValue::Int32(42))]),
                accumulate: Box::new(Expression::Literal(LiteralValue::String(
                    "function (acc, curr) { return acc + curr; }".to_string()
                ))),
                accumulate_args: vec![Expression::Ref(Ref::FieldRef("a".to_string()))],
                merge: Box::new(Expression::Literal(LiteralValue::String(
                    "function (a, b) { return a + b; }".to_string()
                ))),
                finalize: Some(Box::new(Expression::Literal(LiteralValue::String(
                    "function (_x) { return 42; }".to_string()
                )))),
                lang: "js".to_string()
            })),
            input = r#"expr: {"$accumulator": {
                                "init": "function (y) { return y; }",
                                "initArgs": [42],
                                "accumulate": "function (acc, curr) { return acc + curr; }",
                                "accumulateArgs": ["$a"],
                                "merge": "function (a, b) { return a + b; }",
                                "finalize": "function (_x) { return 42; }",
                                "lang": "js"
            }}"#
        );

        test_serde_expr!(
            accumulator_no_optional_args,
            expected = Expression::TaggedOperator(TaggedOperator::Accumulator(Accumulator {
                init: Box::new(Expression::Literal(LiteralValue::String(
                    "function (y) { return y; }".to_string()
                ))),
                init_args: None,
                accumulate: Box::new(Expression::Literal(LiteralValue::String(
                    "function (acc, curr) { return acc + curr; }".to_string()
                ))),
                accumulate_args: vec![Expression::Ref(Ref::FieldRef("a".to_string()))],
                merge: Box::new(Expression::Literal(LiteralValue::String(
                    "function (a, b) { return a + b; }".to_string()
                ))),
                finalize: None,
                lang: "js".to_string()
            })),
            input = r#"expr: {"$accumulator": {
                                "init": "function (y) { return y; }",
                                "accumulate": "function (acc, curr) { return acc + curr; }",
                                "accumulateArgs": ["$a"],
                                "merge": "function (a, b) { return a + b; }",
                                "lang": "js"
            }}"#
        );

        test_serde_expr!(
            function,
            expected = Expression::TaggedOperator(TaggedOperator::Function(Function {
                body: Box::new(Expression::Literal(LiteralValue::String(
                    "function (x) { return x + 1; }".to_string()
                ))),
                args: vec![
                    Expression::Literal(LiteralValue::Int32(1)),
                    Expression::Literal(LiteralValue::Int32(2)),
                ],
                lang: "js".to_string()
            })),
            input = r#"expr: {"$function": {
                                "body": "function (x) { return x + 1; }",
                                "args": [1, 2],
                                "lang": "js"
            }}"#
        );

        test_serde_expr!(
            get_field,
            expected = Expression::TaggedOperator(TaggedOperator::GetField(GetField {
                field: "x".to_string(),
                input: Box::new(Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Int32(1))
                }))
            })),
            input = r#"expr: {"$getField": {"field": "x", "input": {"x": 1}}}"#
        );

        test_serde_expr!(
            set_field,
            expected = Expression::TaggedOperator(TaggedOperator::SetField(SetField {
                field: "x".to_string(),
                input: Box::new(Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Int32(1))
                })),
                value: Box::new(Expression::Literal(LiteralValue::String("new".to_string())))
            })),
            input = r#"expr: {"$setField": {"field": "x", "input": {"x": 1}, "value": "new"}}"#
        );

        test_serde_expr!(
            unset_field,
            expected = Expression::TaggedOperator(TaggedOperator::UnsetField(UnsetField {
                field: "x".to_string(),
                input: Box::new(Expression::Document(map! {
                    "x".to_string() => Expression::Literal(LiteralValue::Int32(1))
                }))
            })),
            input = r#"expr: {"$unsetField": {"field": "x", "input": {"x": 1}}}"#
        );

        test_serde_expr!(
            switch,
            expected = Expression::TaggedOperator(TaggedOperator::Switch(Switch {
                branches: vec![
                    SwitchCase {
                        case: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                        then: Box::new(Expression::Literal(LiteralValue::Int32(10))),
                    },
                    SwitchCase {
                        case: Box::new(Expression::Ref(Ref::FieldRef("b".to_string()))),
                        then: Box::new(Expression::Literal(LiteralValue::Int32(20))),
                    },
                ],
                default: Box::new(Expression::Literal(LiteralValue::Null))
            })),
            input = r#"expr: {"$switch": {
                                "branches": [
                                    {"case": "$a", "then": 10},
                                    {"case": "$b", "then": 20},
                                ],
                                "default": null
            }}"#
        );

        test_serde_expr!(
            let_expr,
            expected = Expression::TaggedOperator(TaggedOperator::Let(Let {
                vars: map! {
                    "a".to_string() => Expression::Literal(LiteralValue::Int32(1)),
                    "b".to_string() => Expression::Literal(LiteralValue::Int32(2)),
                },
                inside: Box::new(Expression::Literal(LiteralValue::String(
                    "body".to_string()
                )))
            })),
            input = r#"expr: {"$let": {
                                "vars": {"a": 1, "b": 2},
                                "in": "body"
            }}"#
        );

        test_serde_expr!(
            sql_convert,
            expected = Expression::TaggedOperator(TaggedOperator::SQLConvert(SQLConvert {
                input: Box::new(Expression::Literal(LiteralValue::String("1".to_string()))),
                to: "int".to_string(),
                on_null: Box::new(Expression::Literal(LiteralValue::Null)),
                on_error: Box::new(Expression::Literal(LiteralValue::Null)),
            })),
            input = r#"expr: {"$sqlConvert": {
                                "input": "1",
                                "to": "int",
                                "onNull": null,
                                "onError": null
            }}"#
        );

        test_serde_expr!(
            convert_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::Convert(Convert {
                input: Box::new(Expression::Literal(LiteralValue::String("1".to_string()))),
                to: Box::new(Expression::Literal(LiteralValue::String("int".to_string()))),
                format: None,
                on_null: None,
                on_error: None,
            })),
            input = r#"expr: {"$convert": {
                                "input": "1",
                                "to": "int"
            }}"#
        );

        test_serde_expr!(
            convert_subtype,
            expected = Expression::TaggedOperator(TaggedOperator::Convert(Convert {
                input: Box::new(Expression::Literal(LiteralValue::Int32(123))),
                to: Box::new(Expression::Document(map! {
                    "type".to_string() => Expression::Literal(LiteralValue::String("binData".to_string())),
                    "subtype".to_string() => Expression::Literal(LiteralValue::Int32(0)),
                })),
                format: None,
                on_null: None,
                on_error: None,
            })),
            input = r#"expr: {"$convert": {
                                "input": 123,
                                "to": {"type": "binData", "subtype": 0},
            }}"#
        );

        test_serde_expr!(
            convert_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::Convert(Convert {
                input: Box::new(Expression::Literal(LiteralValue::String("1".to_string()))),
                to: Box::new(Expression::Literal(LiteralValue::String("int".to_string()))),
                format: Some("hi".to_string()),
                on_null: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                on_error: Some(Box::new(Expression::Literal(LiteralValue::Null))),
            })),
            input = r#"expr: {"$convert": {
                                "input": "1",
                                "to": "int",
                                "format": "hi",
                                "onNull": null,
                                "onError": null
            }}"#
        );

        test_serde_expr!(
            convert_null_options,
            expected = Expression::TaggedOperator(TaggedOperator::Convert(Convert {
                input: Box::new(Expression::Literal(LiteralValue::String("1".to_string()))),
                to: Box::new(Expression::Literal(LiteralValue::String(
                    "binData".to_string()
                ))),
                format: Some("format".to_string()),
                on_null: Some(Box::new(Expression::Literal(LiteralValue::Int32(1)))),
                on_error: Some(Box::new(Expression::Literal(LiteralValue::Int32(2)))),
            })),
            input = r#"expr: {"$convert": {
                                "input": "1",
                                "to": "binData",
                                "format": "format",
                                "onNull": 1,
                                "onError": 2
            }}"#
        );

        test_serde_expr!(
            like_with_escape,
            expected = Expression::TaggedOperator(TaggedOperator::Like(Like {
                input: Box::new(Expression::Literal(LiteralValue::String(
                    "x*yz".to_string()
                ))),
                pattern: Box::new(Expression::Literal(LiteralValue::String(
                    "x!*.*".to_string()
                ))),
                escape: Some('!')
            })),
            input = r#"expr: {"$like": {
                                "input": "x*yz",
                                "pattern": "x!*.*",
                                "escape": "!"
            }}"#
        );

        test_serde_expr!(
            like_without_escape,
            expected = Expression::TaggedOperator(TaggedOperator::Like(Like {
                input: Box::new(Expression::Literal(LiteralValue::String(
                    "x*yz".to_string()
                ))),
                pattern: Box::new(Expression::Literal(LiteralValue::String(
                    "x!*.*".to_string()
                ))),
                escape: None
            })),
            input = r#"expr: {"$like": {
                                "input": "x*yz",
                                "pattern": "x!*.*"
            }}"#
        );

        test_serde_expr!(
            sql_divide,
            expected = Expression::TaggedOperator(TaggedOperator::SQLDivide(SQLDivide {
                dividend: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                divisor: Box::new(Expression::Literal(LiteralValue::Int32(2))),
                on_error: Box::new(Expression::Literal(LiteralValue::Null)),
            })),
            input = r#"expr: {"$sqlDivide": {
                                "dividend": "$a",
                                "divisor": 2,
                                "onError": null
            }}"#
        );

        test_serde_expr!(
            regex_find_with_options,
            expected = Expression::TaggedOperator(TaggedOperator::RegexFind(RegexAggExpression {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                regex: Box::new(Expression::Literal(LiteralValue::String(
                    "pattern".to_string()
                ))),
                options: Some(Box::new(Expression::Literal(LiteralValue::String(
                    "imxs".to_string()
                )))),
            })),
            input = r#"expr: {"$regexFind": {
                                "input": "$a",
                                "regex": "pattern",
                                "options": "imxs"
            }}"#
        );

        test_serde_expr!(
            regex_find_without_options,
            expected = Expression::TaggedOperator(TaggedOperator::RegexFind(RegexAggExpression {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                regex: Box::new(Expression::Literal(LiteralValue::String(
                    "/pattern/i".to_string()
                ))),
                options: None,
            })),
            input = r#"expr: {"$regexFind": {
                                "input": "$a",
                                "regex": "/pattern/i"
            }}"#
        );

        test_serde_expr!(
            regex_find_all_with_options,
            expected =
                Expression::TaggedOperator(TaggedOperator::RegexFindAll(RegexAggExpression {
                    input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                    regex: Box::new(Expression::Literal(LiteralValue::String(
                        "pattern".to_string()
                    ))),
                    options: Some(Box::new(Expression::Literal(LiteralValue::String(
                        "imxs".to_string()
                    )))),
                })),
            input = r#"expr: {"$regexFindAll": {
                                "input": "$a",
                                "regex": "pattern",
                                "options": "imxs"
            }}"#
        );

        test_serde_expr!(
            regex_find_all_without_options,
            expected =
                Expression::TaggedOperator(TaggedOperator::RegexFindAll(RegexAggExpression {
                    input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                    regex: Box::new(Expression::Literal(LiteralValue::String(
                        "/pattern/i".to_string()
                    ))),
                    options: None,
                })),
            input = r#"expr: {"$regexFindAll": {
                                "input": "$a",
                                "regex": "/pattern/i"
            }}"#
        );

        test_serde_expr!(
            replace_all,
            expected = Expression::TaggedOperator(TaggedOperator::ReplaceAll(Replace {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                find: Box::new(Expression::Literal(LiteralValue::String(
                    "pattern".to_string()
                ))),
                replacement: Box::new(Expression::Literal(LiteralValue::String("new".to_string()))),
            })),
            input = r#"expr: {"$replaceAll": {
                                "input": "$a",
                                "find": "pattern",
                                "replacement": "new"
            }}"#
        );

        test_serde_expr!(
            replace_one,
            expected = Expression::TaggedOperator(TaggedOperator::ReplaceOne(Replace {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                find: Box::new(Expression::Literal(LiteralValue::String(
                    "pattern".to_string()
                ))),
                replacement: Box::new(Expression::Literal(LiteralValue::String("new".to_string()))),
            })),
            input = r#"expr: {"$replaceOne": {
                                "input": "$a",
                                "find": "pattern",
                                "replacement": "new"
            }}"#
        );

        test_serde_expr!(
            sql_subquery,
            expected = Expression::TaggedOperator(TaggedOperator::Subquery(Subquery {
                db: Some("foo".to_string()),
                collection: Some("bar".to_string()),
                let_bindings: None,
                output_path: Some(vec!["x".to_string()]),
                pipeline: vec![Stage::Project(ProjectStage {
                    items: map! {"x".to_string() => ProjectItem::Inclusion}
                })]
            })),
            input = r#"expr: {"$subquery": {
                            "db": "foo",
                            "collection": "bar",
                            "outputPath": ["x"],
                            "pipeline": [
                              {
                                "$project": {
                                  "x": 1
                                }
                              }
                            ]
                          }}"#
        );

        test_serde_expr!(
            sql_subquery_comparison,
            expected = Expression::TaggedOperator(TaggedOperator::SubqueryComparison(
                SubqueryComparison {
                    op: "eq".to_string(),
                    modifier: "all".to_string(),
                    arg: Box::new(Expression::Literal(LiteralValue::Int32(42))),
                    subquery: Subquery {
                        db: Some("foo".to_string()),
                        collection: Some("bar".to_string()),
                        let_bindings: None,
                        output_path: Some(vec!["x".to_string()]),
                        pipeline: vec![
                            Stage::Documents(vec![]),
                            Stage::Project(ProjectStage {
                                items: map! {"x".to_string() => ProjectItem::Inclusion}
                            })
                        ]
                    }
                    .into()
                }
            )),
            input = r#"expr: {"$subqueryComparison": {
                            "op": "eq",
                            "modifier": "all",
                            "arg": 42,
                            "subquery": {
                                "db": "foo",
                                "collection": "bar",
                                "outputPath": ["x"],
                                "pipeline": [
                                    {"$documents": []},
                                    {
                                        "$project": {
                                            "x": 1
                                        }
                                    }
                                ]
                          }}}"#
        );

        test_serde_expr!(
            sql_subquery_exists,
            expected = Expression::TaggedOperator(TaggedOperator::SubqueryExists(SubqueryExists {
                db: Some("foo".to_string()),
                collection: Some("bar".to_string()),
                let_bindings: None,
                pipeline: vec![Stage::Project(ProjectStage {
                    items: map! {"x".to_string() => ProjectItem::Inclusion}
                })]
            })),
            input = r#"expr: {"$subqueryExists": {
                            "db": "foo",
                            "collection": "bar",
                            "pipeline": [
                              {
                                "$project": {
                                  "x": 1
                                }
                              }
                            ]
                          }}"#
        );

        // accumulator operators
        test_serde_expr!(
            bottom,
            expected = Expression::TaggedOperator(TaggedOperator::Bottom(Bottom {
                output: Box::new(Expression::Array(vec![
                    Expression::Ref(Ref::FieldRef("playerId".to_string())),
                    Expression::Ref(Ref::FieldRef("score".to_string()))
                ])),
                sort_by: Box::new(Expression::Document(map!(
                    "score".to_string() => Expression::Literal(LiteralValue::Int64(-1))
                )))
            })),
            input = r#"expr: { $bottom: {
                                output: [ "$playerId", "$score" ],
                                sortBy: { "score": -1 }
            }}"#
        );

        test_serde_expr!(
            bottom_n,
            expected = Expression::TaggedOperator(TaggedOperator::BottomN(BottomN {
                output: Box::new(Expression::Array(vec![
                    Expression::Ref(Ref::FieldRef("playerId".to_string())),
                    Expression::Ref(Ref::FieldRef("score".to_string()))
                ])),
                sort_by: Box::new(Expression::Document(map!(
                    "score".to_string() => Expression::Literal(LiteralValue::Int64(-1))
                ))),
                n: 3,
            })),
            input = r#"expr: { $bottomN: {
                                output: [ "$playerId", "$score" ],
                                sortBy: { "score": -1 },
                                n: 3
            }}"#
        );

        test_serde_expr!(
            median_numeric_input,
            expected = Expression::TaggedOperator(TaggedOperator::Median(Median {
                method: "approximate".to_string(),
                input: Box::new(Expression::Ref(Ref::FieldRef("test01".to_string())))
            })),
            input = r#"expr: { $median: {
                                input: "$test01",
                                method: 'approximate'
            }}"#
        );

        test_serde_expr!(
            median_vec_input,
            expected = Expression::TaggedOperator(TaggedOperator::Median(Median {
                method: "approximate".to_string(),
                input: Box::new(Expression::Array(vec![
                    Expression::Ref(Ref::FieldRef("test01".to_string())),
                    Expression::Ref(Ref::FieldRef("test02".to_string())),
                    Expression::Ref(Ref::FieldRef("test03".to_string())),
                ]))
            })),
            input = r#"expr: { $median: {
                                input: [ "$test01", "$test02", "$test03" ],
                                method: 'approximate'
            }}"#
        );

        test_serde_expr!(
            percentile_numeric_input,
            expected = Expression::TaggedOperator(TaggedOperator::Percentile(Percentile {
                method: "approximate".to_string(),
                input: Box::new(Expression::Ref(Ref::FieldRef("test01".to_string()))),
                p: vec![
                    Expression::Literal(LiteralValue::Double(0.9)),
                    Expression::Literal(LiteralValue::Double(0.5)),
                    Expression::Literal(LiteralValue::Double(0.75)),
                    Expression::Literal(LiteralValue::Double(0.95)),
                ]
            })),
            input = r#"expr: { $percentile: {
                                input: "$test01",
                                p: [ 0.9, 0.5, 0.75, 0.95 ],
                                method: 'approximate'
            }}"#
        );

        test_serde_expr!(
            percentile_vec_input,
            expected = Expression::TaggedOperator(TaggedOperator::Percentile(Percentile {
                method: "approximate".to_string(),
                input: Box::new(Expression::Array(vec![
                    Expression::Ref(Ref::FieldRef("test01".to_string())),
                    Expression::Ref(Ref::FieldRef("test02".to_string())),
                    Expression::Ref(Ref::FieldRef("test03".to_string())),
                ])),
                p: vec![
                    Expression::Literal(LiteralValue::Double(0.5)),
                    Expression::Literal(LiteralValue::Double(0.95)),
                ]
            })),
            input = r#"expr: { $percentile: {
                                input: [ "$test01", "$test02", "$test03" ],
                                p: [ 0.5, 0.95 ],
                                method: 'approximate'
            }}"#
        );

        test_serde_expr!(
            top,
            expected = Expression::TaggedOperator(TaggedOperator::Top(Top {
                output: Box::new(Expression::Array(vec![
                    Expression::Ref(Ref::FieldRef("playerId".to_string())),
                    Expression::Ref(Ref::FieldRef("score".to_string()))
                ])),
                sort_by: Box::new(Expression::Document(map!(
                    "score".to_string() => Expression::Literal(LiteralValue::Int64(-1))
                )))
            })),
            input = r#"expr: { $top: {
                                output: [ "$playerId", "$score" ],
                                sortBy: { "score": -1 }
            }}"#
        );

        test_serde_expr!(
            top_n,
            expected = Expression::TaggedOperator(TaggedOperator::TopN(TopN {
                output: Box::new(Expression::Array(vec![
                    Expression::Ref(Ref::FieldRef("playerId".to_string())),
                    Expression::Ref(Ref::FieldRef("score".to_string()))
                ])),
                sort_by: Box::new(Expression::Document(map!(
                    "score".to_string() => Expression::Literal(LiteralValue::Int64(-1))
                ))),
                n: 3,
            })),
            input = r#"expr: { $topN: {
                                output: [ "$playerId", "$score" ],
                                sortBy: { "score": -1 },
                                n: 3
            }}"#
        );

        // Array Operators
        test_serde_expr!(
            first_n,
            expected = Expression::TaggedOperator(TaggedOperator::FirstN(NArrayOp {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                n: Box::new(Expression::Literal(LiteralValue::Int32(3))),
            })),
            input = r#"expr: {"$firstN": {
                                "input": "$a",
                                "n": 3
            }}"#
        );

        test_serde_expr!(
            last_n,
            expected = Expression::TaggedOperator(TaggedOperator::LastN(NArrayOp {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                n: Box::new(Expression::Literal(LiteralValue::Int32(3))),
            })),
            input = r#"expr: {"$lastN": {
                                "input": "$a",
                                "n": 3
            }}"#
        );

        test_serde_expr!(
            filter_with_as,
            expected = Expression::TaggedOperator(TaggedOperator::Filter(Filter {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                _as: Some("x".to_string()),
                cond: Box::new(Expression::Literal(LiteralValue::Int32(2))),
                limit: None,
            })),
            input = r#"expr: {"$filter": {
                                "input": "$a",
                                "as": "x",
                                "cond": 2
            }}"#
        );

        test_serde_expr!(
            filter_with_limit,
            expected = Expression::TaggedOperator(TaggedOperator::Filter(Filter {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                _as: Some("x".to_string()),
                cond: Box::new(Expression::Literal(LiteralValue::Int32(2))),
                limit: Some(Box::new(Expression::UntaggedOperator(UntaggedOperator {
                    op: UntaggedOperatorName::Add,
                    args: vec![
                        Expression::Literal(LiteralValue::Int32(1)),
                        Expression::Literal(LiteralValue::Int32(2)),
                    ]
                }))),
            })),
            input = r#"expr: {"$filter": {
                                "input": "$a",
                                "as": "x",
                                "cond": 2,
                                limit: {"$add": [1 ,2]}
            }}"#
        );

        test_serde_expr!(
            map_with_as,
            expected = Expression::TaggedOperator(TaggedOperator::Map(Map {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                _as: Some("x".to_string()),
                inside: Box::new(Expression::Literal(LiteralValue::Int32(2))),
            })),
            input = r#"expr: {"$map": {
                                "input": "$a",
                                "as": "x",
                                "in": 2
            }}"#
        );

        test_serde_expr!(
            map_without_as,
            expected = Expression::TaggedOperator(TaggedOperator::Map(Map {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                _as: None,
                inside: Box::new(Expression::Literal(LiteralValue::Int32(2))),
            })),
            input = r#"expr: {"$map": {
                                "input": "$a",
                                "in": 2
            }}"#
        );

        test_serde_expr!(
            max_n_array_element,
            expected = Expression::TaggedOperator(TaggedOperator::MaxNArrayElement(NArrayOp {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                n: Box::new(Expression::Literal(LiteralValue::Int32(2))),
            })),
            input = r#"expr: {"$maxN": {
                                "input": "$a",
                                "n": 2
            }}"#
        );

        test_serde_expr!(
            min_n_array_element,
            expected = Expression::TaggedOperator(TaggedOperator::MinNArrayElement(NArrayOp {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                n: Box::new(Expression::Literal(LiteralValue::Int32(2))),
            })),
            input = r#"expr: {"$minN": {
                                "input": "$a",
                                "n": 2
            }}"#
        );

        test_serde_expr!(
            reduce,
            expected = Expression::TaggedOperator(TaggedOperator::Reduce(Reduce {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                initial_value: Box::new(Expression::Literal(LiteralValue::Int32(2))),
                inside: Box::new(Expression::UntaggedOperator(UntaggedOperator {
                    op: UntaggedOperatorName::Add,
                    args: vec![
                        Expression::Ref(Ref::VariableRef("this".to_string())),
                        Expression::Literal(LiteralValue::Int32(2)),
                    ],
                })),
            })),
            input = r#"expr: {"$reduce": {
                                "input": "$a",
                                "initialValue": 2,
                                "in": {$add: ["$$this", 2]}
            }}"#
        );

        test_serde_expr!(
            sort_array,
            expected = Expression::TaggedOperator(TaggedOperator::SortArray(SortArray {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                sort_by: SortArraySpec::Keys(map! {
                    "x".to_string() => -1,
                    "y".to_string() => 1,
                }),
            })),
            input = r#"expr: {"$sortArray": {
                                "input": "$a",
                                "sortBy": {"x": -1, "y": 1}
            }}"#
        );

        test_serde_expr!(
            sort_array_with_limit,
            expected = Expression::TaggedOperator(TaggedOperator::SortArray(SortArray {
                input: Box::new(Expression::Ref(Ref::FieldRef("a".to_string()))),
                sort_by: SortArraySpec::Value(-1),
            })),
            input = r#"expr: {"$sortArray": {
                                "input": "$a",
                                "sortBy": -1
            }}"#
        );

        test_serde_expr!(
            zip,
            expected = Expression::TaggedOperator(TaggedOperator::Zip(Zip {
                inputs: Box::new(Expression::Array(vec![Expression::Ref(Ref::FieldRef(
                    "a".to_string()
                ))])),
                use_longest_length: true,
                defaults: Some(Box::new(Expression::Array(vec![
                    Expression::Literal(LiteralValue::Int32(1)),
                    Expression::Literal(LiteralValue::Int32(2)),
                    Expression::Literal(LiteralValue::Int32(3)),
                ]))),
            })),
            input = r#"expr: {"$zip": {
                                "inputs": ["$a"],
                                "useLongestLength": true,
                                "defaults": [1, 2, 3]
            }}"#
        );

        test_serde_expr!(
            zip_default_false,
            expected = Expression::TaggedOperator(TaggedOperator::Zip(Zip {
                inputs: Box::new(Expression::Array(vec![Expression::Ref(Ref::FieldRef(
                    "a".to_string()
                ))])),
                use_longest_length: false,
                defaults: None,
            })),
            input = r#"expr: {"$zip": {
                                "inputs": ["$a"],
            }}"#
        );
        // date operators
        test_serde_expr!(
            hour_no_timezone,
            expected = Expression::TaggedOperator(TaggedOperator::Hour(DateExpression {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                timezone: None
            })),
            input = r#"expr: {"$hour": "$date" }"#
        );

        test_serde_expr!(
            hour_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::Hour(DateExpression {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                ))))
            })),
            input = r#"expr: {"$hour": {date: "$date", timezone: "$timezone" } }"#
        );

        test_serde_expr!(
            hour_document_no_timezone,
            expected = Expression::TaggedOperator(TaggedOperator::Hour(DateExpression {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                timezone: None
            })),
            input = r#"expr: {"$hour": {date: "$date" } }"#
        );

        test_serde_expr!(
            hour_document_timezone_null,
            expected = Expression::TaggedOperator(TaggedOperator::Hour(DateExpression {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                timezone: Some(Box::new(Expression::Literal(LiteralValue::Null)))
            })),
            input = r#"expr: {"$hour": {date: "$date", timezone: null } }"#
        );

        test_serde_date_operator!(
            minute_fully_specified,
            string_op = "$minute",
            expected_op = TaggedOperator::Minute
        );

        test_serde_date_operator!(
            second_fully_specified,
            string_op = "$second",
            expected_op = TaggedOperator::Second
        );

        test_serde_date_operator!(
            millisecond_fully_specified,
            string_op = "$millisecond",
            expected_op = TaggedOperator::Millisecond
        );

        test_serde_date_operator!(
            week_fully_specified,
            string_op = "$week",
            expected_op = TaggedOperator::Week
        );

        test_serde_date_operator!(
            month_fully_specified,
            string_op = "$month",
            expected_op = TaggedOperator::Month
        );

        test_serde_date_operator!(
            year_fully_specified,
            string_op = "$year",
            expected_op = TaggedOperator::Year
        );

        test_serde_date_operator!(
            day_of_week_fully_specified,
            string_op = "$dayOfWeek",
            expected_op = TaggedOperator::DayOfWeek
        );

        test_serde_date_operator!(
            day_of_month_fully_specified,
            string_op = "$dayOfMonth",
            expected_op = TaggedOperator::DayOfMonth
        );

        test_serde_date_operator!(
            day_of_year_fully_specified,
            string_op = "$dayOfYear",
            expected_op = TaggedOperator::DayOfYear
        );

        test_serde_date_operator!(
            iso_day_of_week_fully_specified,
            string_op = "$isoDayOfWeek",
            expected_op = TaggedOperator::IsoDayOfWeek
        );

        test_serde_date_operator!(
            iso_week_fully_specified,
            string_op = "$isoWeek",
            expected_op = TaggedOperator::IsoWeek
        );

        test_serde_date_operator!(
            iso_week_year_fully_specified,
            string_op = "$isoWeekYear",
            expected_op = TaggedOperator::IsoWeekYear
        );

        test_serde_expr!(
            date_to_parts_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::DateToParts(DateToParts {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                timezone: None,
                iso8601: None
            })),
            input = r#"expr: {"$dateToParts": {date: "$date" } }"#
        );

        test_serde_expr!(
            date_to_parts_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateToParts(DateToParts {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
                iso8601: Some(true)
            })),
            input = r#"expr: {"$dateToParts": {date: "$date", timezone: "$timezone", iso8601: true } }"#
        );

        test_serde_expr!(
            date_from_parts_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateFromParts(DateFromParts {
                year: Some(Box::new(Expression::Ref(Ref::FieldRef("year".to_string())))),
                month: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "month".to_string()
                )))),
                day: Some(Box::new(Expression::Ref(Ref::FieldRef("day".to_string())))),
                hour: Some(Box::new(Expression::Ref(Ref::FieldRef("hour".to_string())))),
                minute: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "minute".to_string()
                )))),
                second: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "second".to_string()
                )))),
                millisecond: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "millisecond".to_string()
                )))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
                iso_day_of_week: None,
                iso_week: None,
                iso_week_year: None,
            })),
            input = r#"expr: {"$dateFromParts": {year: "$year", timezone: "$timezone", month: "$month", day: "$day", hour: "$hour", minute: "$minute", second: "$second", millisecond: "$millisecond" } }"#
        );

        test_serde_expr!(
            date_from_iso_parts,
            expected = Expression::TaggedOperator(TaggedOperator::DateFromParts(DateFromParts {
                year: None,
                month: None,
                day: None,
                hour: Some(Box::new(Expression::Ref(Ref::FieldRef("hour".to_string())))),
                minute: None,
                second: None,
                millisecond: None,
                timezone: None,
                iso_day_of_week: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "isoDayOfWeek".to_string()
                )))),
                iso_week: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "isoWeek".to_string()
                )))),
                iso_week_year: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "isoWeekYear".to_string()
                )))),
            })),
            input = r#"expr: {"$dateFromParts": {isoWeekYear: "$isoWeekYear", isoWeek: "$isoWeek", isoDayOfWeek: "$isoDayOfWeek", hour: "$hour" } }"#
        );

        test_serde_expr!(
            date_from_parts_all_specified_null,
            expected = Expression::TaggedOperator(TaggedOperator::DateFromParts(DateFromParts {
                year: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                month: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                day: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                hour: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                minute: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                second: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                millisecond: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                timezone: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                iso_day_of_week: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                iso_week: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                iso_week_year: Some(Box::new(Expression::Literal(LiteralValue::Null))),
            })),
            input = r#"expr: {"$dateFromParts": {year: null, month: null, day: null, isoWeekYear: null, isoWeek: null, isoDayOfWeek: null, hour: null, minute: null, second: null, millisecond: null, timezone: null} }"#
        );

        test_serde_expr!(
            date_from_string_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::DateFromString(DateFromString {
                date_string: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                format: None,
                timezone: None,
                on_error: None,
                on_null: None,
            })),
            input = r#"expr: {"$dateFromString": {dateString: "$date" } }"#
        );

        test_serde_expr!(
            date_from_string_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateFromString(DateFromString {
                date_string: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                format: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "format".to_string()
                )))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
                on_error: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "onError".to_string()
                )))),
                on_null: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "onNull".to_string()
                )))),
            })),
            input = r#"expr: {"$dateFromString": { dateString: "$date", timezone: "$timezone", format: "$format", onError: "$onError", onNull: "$onNull" } }"#
        );

        test_serde_expr!(
            date_from_string_optional_params_explicitly_null,
            expected = Expression::TaggedOperator(TaggedOperator::DateFromString(DateFromString {
                date_string: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                format: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                timezone: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                on_error: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                on_null: Some(Box::new(Expression::Literal(LiteralValue::Null))),
            })),
            input = r#"expr: {"$dateFromString": { dateString: "$date", timezone: null, format: null, onError: null, onNull: null } }"#
        );

        test_serde_expr!(
            date_to_string_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::DateToString(DateToString {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                format: None,
                timezone: None,
                on_null: None,
            })),
            input = r#"expr: {"$dateToString": {date: "$date" } }"#
        );

        test_serde_expr!(
            date_to_string_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateToString(DateToString {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                format: Some(Box::new(Expression::Literal(LiteralValue::String(
                    "format".to_string()
                )))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
                on_null: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "onNull".to_string()
                )))),
            })),
            input = r#"expr: {"$dateToString": {date: "$date", timezone: "$timezone", format: "format", onNull: "$onNull" } }"#
        );

        test_serde_expr!(
            date_to_string_optional_params_explicitly_null,
            expected = Expression::TaggedOperator(TaggedOperator::DateToString(DateToString {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                format: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                timezone: Some(Box::new(Expression::Literal(LiteralValue::Null))),
                on_null: Some(Box::new(Expression::Literal(LiteralValue::Null))),
            })),
            input = r#"expr: {"$dateToString": {date: "$date", "format": null, "timezone": null, "onNull": null } }"#
        );

        test_serde_expr!(
            date_add_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::DateAdd(DateAdd {
                start_date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                amount: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                timezone: None,
            })),
            input = r#"expr: {"$dateAdd": {startDate: "$date", unit: "year", amount: 1 } }"#
        );

        test_serde_expr!(
            date_add_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateAdd(DateAdd {
                start_date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                amount: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
            })),
            input = r#"expr: {"$dateAdd": {startDate: "$date", unit: "year", amount: 1, timezone: "$timezone" } }"#
        );

        test_serde_expr!(
            date_subtract_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::DateSubtract(DateSubtract {
                start_date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                amount: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                timezone: None,
            })),
            input = r#"expr: {"$dateSubtract": {startDate: "$date", unit: "year", amount: 1 } }"#
        );

        test_serde_expr!(
            date_subtract_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateSubtract(DateSubtract {
                start_date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                amount: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
            })),
            input = r#"expr: {"$dateSubtract": {startDate: "$date", unit: "year", amount: 1, timezone: "$timezone" } }"#
        );

        test_serde_expr!(
            date_diff_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::DateDiff(DateDiff {
                start_date: Box::new(Expression::Ref(Ref::FieldRef("startDate".to_string()))),
                end_date: Box::new(Expression::Ref(Ref::FieldRef("endDate".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                timezone: None,
                start_of_week: None
            })),
            input = r#"expr: {"$dateDiff": {startDate: "$startDate", endDate: "$endDate", unit: "year" } }"#
        );

        test_serde_expr!(
            date_diff_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateDiff(DateDiff {
                start_date: Box::new(Expression::Ref(Ref::FieldRef("startDate".to_string()))),
                end_date: Box::new(Expression::Ref(Ref::FieldRef("endDate".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
                start_of_week: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "startOfWeek".to_string()
                ))))
            })),
            input = r#"expr: {"$dateDiff": {startDate: "$startDate", endDate: "$endDate", unit: "year", timezone: "$timezone", startOfWeek: "$startOfWeek" } }"#
        );

        test_serde_expr!(
            date_trunc_no_options,
            expected = Expression::TaggedOperator(TaggedOperator::DateTrunc(DateTrunc {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                timezone: None,
                start_of_week: None,
                bin_size: None
            })),
            input = r#"expr: {"$dateTrunc": {date: "$date", unit: "year" } }"#
        );

        test_serde_expr!(
            date_trunc_fully_specified,
            expected = Expression::TaggedOperator(TaggedOperator::DateTrunc(DateTrunc {
                date: Box::new(Expression::Ref(Ref::FieldRef("date".to_string()))),
                unit: Box::new(Expression::Literal(LiteralValue::String(
                    "year".to_string()
                ))),
                timezone: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "timezone".to_string()
                )))),
                start_of_week: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "startOfWeek".to_string()
                )))),
                bin_size: Some(Box::new(Expression::Ref(Ref::FieldRef(
                    "binSize".to_string()
                ))))
            })),
            input = r#"expr: {"$dateTrunc": {date: "$date", unit: "year", timezone: "$timezone", startOfWeek: "$startOfWeek", binSize: "$binSize" } }"#
        );

        test_serde_expr!(
            trim_no_chars,
            expected = Expression::TaggedOperator(TaggedOperator::Trim(Trim {
                input: Box::new(Expression::Literal(LiteralValue::String(
                    "hello".to_string()
                ))),
                chars: None
            })),
            input = r#"expr: {"$trim": {input: "hello" } }"#
        );

        test_serde_expr!(
            trim_chars_null,
            expected = Expression::TaggedOperator(TaggedOperator::Trim(Trim {
                input: Box::new(Expression::Literal(LiteralValue::String(
                    "hello".to_string()
                ))),
                chars: Some(Box::new(Expression::Literal(LiteralValue::Null)))
            })),
            input = r#"expr: {"$trim": {input: "hello", "chars": null } }"#
        );

        test_serde_expr!(
            trim_chars,
            expected = Expression::TaggedOperator(TaggedOperator::Trim(Trim {
                input: Box::new(Expression::Literal(LiteralValue::String(
                    "hello".to_string()
                ))),
                chars: Some(Box::new(Expression::Literal(LiteralValue::String(
                    "world".to_string()
                ))))
            })),
            input = r#"expr: {"$trim": {input: "hello", chars: "world" } }"#
        );

        mod window_functions {
            use crate::definitions::{
                Derivative, EmptyDoc, ExpMovingAvg, ExpMovingAvgOpt, Expression, Integral,
                LiteralValue, Shift, TaggedOperator,
            };

            test_serde_expr!(
                dense_rank,
                expected = Expression::TaggedOperator(TaggedOperator::DenseRank(EmptyDoc {})),
                input = r#"expr: {"$denseRank": {}}"#
            );

            test_serde_expr!(
                derivative_no_unit,
                expected = Expression::TaggedOperator(TaggedOperator::Derivative(Derivative {
                    input: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    unit: None,
                })),
                input = r#"expr: {"$derivative": {
                                    "input": 1,
                }}"#
            );

            test_serde_expr!(
                derivative_unit,
                expected = Expression::TaggedOperator(TaggedOperator::Derivative(Derivative {
                    input: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    unit: Some("day".to_string()),
                })),
                input = r#"expr: {"$derivative": {
                                    "input": 1,
                                    "unit": "day",
                }}"#
            );

            test_serde_expr!(
                document_number,
                expected = Expression::TaggedOperator(TaggedOperator::DocumentNumber(EmptyDoc {})),
                input = r#"expr: {"$documentNumber": {}}"#
            );

            test_serde_expr!(
                exp_moving_avg_n,
                expected = Expression::TaggedOperator(TaggedOperator::ExpMovingAvg(ExpMovingAvg {
                    input: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    opt: ExpMovingAvgOpt::N(1),
                })),
                input = r#"expr: {"$expMovingAvg": {
                                    "input": 1,
                                    "N": 1,
                }}"#
            );

            test_serde_expr!(
                exp_moving_avg_alpha,
                expected = Expression::TaggedOperator(TaggedOperator::ExpMovingAvg(ExpMovingAvg {
                    input: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    opt: ExpMovingAvgOpt::Alpha(1.5),
                })),
                input = r#"expr: {"$expMovingAvg": {
                                    "input": 1,
                                    "alpha": 1.5,
                }}"#
            );

            test_serde_expr!(
                integral_no_unit,
                expected = Expression::TaggedOperator(TaggedOperator::Integral(Integral {
                    input: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    unit: None,
                })),
                input = r#"expr: {"$integral": {
                                    "input": 1,
                }}"#
            );

            test_serde_expr!(
                integral_unit,
                expected = Expression::TaggedOperator(TaggedOperator::Integral(Integral {
                    input: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    unit: Some("day".to_string()),
                })),
                input = r#"expr: {"$integral": {
                                    "input": 1,
                                    "unit": "day",
                }}"#
            );

            test_serde_expr!(
                rank,
                expected = Expression::TaggedOperator(TaggedOperator::Rank(EmptyDoc {})),
                input = r#"expr: {"$rank": {}}"#
            );

            test_serde_expr!(
                shift_no_default,
                expected = Expression::TaggedOperator(TaggedOperator::Shift(Shift {
                    output: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    by: 1,
                    default: None,
                })),
                input = r#"expr: {"$shift": {
                                    "output": 1,
                                    "by": 1,
                }}"#
            );

            test_serde_expr!(
                shift_default,
                expected = Expression::TaggedOperator(TaggedOperator::Shift(Shift {
                    output: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    by: 1,
                    default: Some(Box::new(Expression::Literal(LiteralValue::Int32(1)))),
                })),
                input = r#"expr: {"$shift": {
                                    "output": 1,
                                    "by": 1,
                                    "default": 1,
                }}"#
            );
        }

        mod cond {
            use crate::definitions::{Cond, Expression, LiteralValue, TaggedOperator};

            test_serde_expr!(
                tagged_input,
                expected = Expression::TaggedOperator(TaggedOperator::Cond(Cond {
                    r#if: Box::new(Expression::Literal(LiteralValue::Boolean(true))),
                    then: Box::new(Expression::Literal(LiteralValue::Int32(1))),
                    r#else: Box::new(Expression::Literal(LiteralValue::Null)),
                })),
                input = r#"expr: {"$cond": {
                                    "if": true,
                                    "then": 1,
                                    "else": null,
                }}"#
            );

            test_serde_expr!(
                untagged_input,
                expected = Expression::TaggedOperator(TaggedOperator::Cond(Cond {
                    r#if: Box::new(Expression::Literal(LiteralValue::Boolean(false))),
                    then: Box::new(Expression::Literal(LiteralValue::Int32(0))),
                    r#else: Box::new(Expression::Literal(LiteralValue::String("x".to_string()))),
                })),
                input = r#"expr: {"$cond": [false, 0, "x"]}"#
            );
        }
    }

    mod untagged_operators {
        use crate::{
            definitions::{Expression, LiteralValue, Ref, UntaggedOperator, UntaggedOperatorName},
            map,
        };

        test_serde_expr!(
            one_argument_non_array,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: UntaggedOperatorName::SQLSqrt,
                args: vec![Expression::Ref(Ref::FieldRef("x".to_string()))]
            }),
            input = r#"expr: {"$sqlSqrt": "$x"}"#
        );

        test_serde_expr!(
            one_argument,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: UntaggedOperatorName::SQLSqrt,
                args: vec![Expression::Ref(Ref::FieldRef("x".to_string()))]
            }),
            input = r#"expr: {"$sqlSqrt": ["$x"]}"#
        );

        test_serde_expr!(
            multiple_arguments,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: UntaggedOperatorName::Add,
                args: vec![
                    Expression::Ref(Ref::FieldRef("x".to_string())),
                    Expression::Ref(Ref::FieldRef("y".to_string())),
                    Expression::Ref(Ref::FieldRef("z".to_string())),
                ]
            }),
            input = r#"expr: {"$add": ["$x", "$y", "$z"]}"#
        );

        test_serde_expr!(
            literal,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: UntaggedOperatorName::Literal,
                args: vec![Expression::Literal(LiteralValue::Int32(1))]
            }),
            input = r#"expr: {"$literal": 1}"#
        );

        test_serde_expr!(
            empty_document_argument,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: UntaggedOperatorName::Count,
                args: vec![Expression::Document(map!())]
            }),
            input = r#"expr: {"$count": {}}"#
        );

        test_serde_expr!(
            empty_vec_argument,
            expected = Expression::UntaggedOperator(UntaggedOperator {
                op: UntaggedOperatorName::Rand,
                args: vec![]
            }),
            input = r#"expr: {"$rand": []}"#
        );
    }
}
