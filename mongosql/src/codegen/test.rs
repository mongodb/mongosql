macro_rules! test_codegen_plan {
    (
		$func_name:ident,
		Ok({
			database: $expected_db:expr,
			collection: $expected_collection:expr,
			pipeline: $expected_pipeline:expr,
		}),
		$input: expr,
	) => {
        #[test]
        fn $func_name() {
            use crate::codegen::{generate_mql, MqlTranslation};

            let input = $input;
            let expected_db = $expected_db;
            let expected_collection = $expected_collection;
            let expected_pipeline = $expected_pipeline;

            let MqlTranslation {
                database: db,
                collection: col,
                mapping_registry: _,
                pipeline: pipeline,
            } = generate_mql(input).expect("codegen failed");

            assert_eq!(expected_db, db);
            assert_eq!(expected_collection, col);
            assert_eq!(expected_pipeline, pipeline);
        }
    };

    ($func_name:ident, Err($expected_err:expr), $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::codegen::generate_mql;

            let input = $input;
            let expected = Err($expected_err);

            assert_eq!(expected, generate_mql(input));
        }
    };
}

macro_rules! test_codegen_expr {
    ($func_name:ident, $mapping_registry:expr, $expected:expr, $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::codegen::mql::MqlCodeGenerator;
            let mapping_registry = $mapping_registry;
            let expected = $expected;
            let input = $input;

            let gen = MqlCodeGenerator {
                mapping_registry,
                scope_level: 0u16,
            };
            assert_eq!(expected, gen.codegen_expression(input));
        }
    };

    ($func_name:ident, $expected:expr, $input:expr,) => {
        test_codegen_expr!(
            $func_name,
            crate::codegen::mql::MappingRegistry::default(),
            $expected,
            $input,
        );
    };
}

mod collection {
    use crate::ir::*;

    test_codegen_plan!(
        simple,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
            ],
        }),
        Stage::Collection(Collection {
            db: "mydb".to_string(),
            collection: "col".to_string(),
        }),
    );
}

mod array_stage {
    use crate::ir::*;

    test_codegen_plan!(
        empty,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
            ],
        }),
        Stage::Array(Array {
            array: vec![],
            alias: "arr".to_string(),
        }),
    );
    test_codegen_plan!(
        empty_id_alias,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": "$$ROOT"}},
            ],
        }),
        Stage::Array(Array {
            array: vec![],
            alias: "_id".to_string(),
        }),
    );
    test_codegen_plan!(
        non_empty,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": false}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
            ],
        }),
        Stage::Array(Array {
            array: vec![Expression::Literal(Literal::Boolean(false))],
            alias: "arr".to_string(),
        }),
    );
}

mod project {
    use crate::{
        codegen::Error,
        ir::{
            binding_tuple::{DatasourceName, Key},
            *,
        },
        map,
    };

    test_codegen_plan!(
        simple,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": 0, "a": {"$literal": 1}, "b": {"$literal": 2}, "c": {"$literal": 3}}},
            ],
        }),
        Stage::Project(Project {
            expression: map! {
                ("a", 0u16).into() => Expression::Literal(Literal::Integer(1)),
                ("b", 0u16).into() => Expression::Literal(Literal::Integer(2)),
                ("c", 0u16).into() => Expression::Literal(Literal::Integer(3)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        empty,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": 0}},
            ],
        }),
        Stage::Project(Project {
            expression: map! {},
            source: Stage::Array(Array {
                array: vec![Expression::Document(map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        source_bindings_available_in_project,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": 0, "foo": "$arr"}},
            ],
        }),
        Stage::Project(Project {
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("arr", 0u16).into()),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        user_defined_id_projection,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": {"$literal": 42.0}, "foo": {"$literal": 44.0}}},
            ],
        }),
        Stage::Project(Project {
            expression: map! {
                ("_id", 0u16).into() => Expression::Literal(Literal::Double(42.0)),
                ("foo", 0u16).into() => Expression::Literal(Literal::Double(44.0)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        user_bot_conflict,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"a": {"$literal": 42}}]},
                bson::doc!{"$project": {"_id": 0, "__bot": "$$ROOT"}},
                bson::doc!{"$project": {
                    "_id": 0,
                    "____bot": "$__bot",
                    // reordered by BTreeMap
                    "_____bot": {"$literal": 45.0},
                    "___bot": {"$literal": 44.0},
                    "__bot": {"$literal": 43.0},
                }},
            ],
        }),
        Stage::Project(Project {
            expression: map! {
                Key{ datasource: DatasourceName::Bottom, scope: 0u16 } => Expression::Reference(("__bot", 0u16).into()),
                ("__bot", 0u16).into() => Expression::Literal(Literal::Double(43.0)),
                ("___bot", 0u16).into() => Expression::Literal(Literal::Double(44.0)),
                ("_____bot", 0u16).into() => Expression::Literal(Literal::Double(45.0)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(map!{"a".into() => Expression::Literal(Literal::Integer(42))})],
                alias: "__bot".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        dot_project_field,
        Err(Error::DotsOrDollarsInProjectField),
        Stage::Project(Project {
            expression: map! {
                ("a.b", 0u16).into() => Expression::Literal(Literal::Integer(1)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(map! {})],
                alias: "arr".to_string(),
            })
            .into(),
        }),
    );

    test_codegen_plan!(
        dollar_project_field,
        Err(Error::DotsOrDollarsInProjectField),
        Stage::Project(Project {
            expression: map! {
                ("$a", 0u16).into() => Expression::Literal(Literal::Integer(1)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(map! {})],
                alias: "arr".to_string(),
            })
            .into(),
        }),
    );
}

mod filter {
    use crate::ir::*;

    test_codegen_plan!(
        simple,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$match": {"$expr": {"$literal": true}}},
            ],
        }),
        Stage::Filter(Filter {
            condition: Expression::Literal(Literal::Boolean(true)),
            source: Stage::Array(Array {
                array: vec![],
                alias: "arr".to_string(),
            }).into(),
        }),
    );
}

mod sort {
    use crate::{
        codegen::Error,
        ir::{Expression::Reference, SortSpecification::*, *},
        map,
    };

    test_codegen_plan!(
        empty,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$sort": {}},
            ],
        }),
        Stage::Sort(Sort {
            specs: vec![],
            source: Stage::Array(Array {
                array: vec![],
                alias: "arr".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        single_spec_asc,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col": 1}},
            ],
        }),
        Stage::Sort(Sort {
            specs: vec![Asc(Reference(("col", 0u16).into()).into())],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        single_spec_dsc,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col": -1}},
            ],
        }),
        Stage::Sort(Sort {
            specs: vec![Desc(Reference(("col", 0u16).into()).into())],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        multi_spec,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col": 1, "col.a": -1}},
            ],
        }),
        Stage::Sort(Sort {
            specs: vec![
                Asc(Reference(("col", 0u16).into()).into()),
                Desc(
                    Expression::FieldAccess(FieldAccess{
                        field: "a".to_string(),
                        expr: Reference(("col", 0u16).into()).into(),
                    }).into(),
                ),
            ],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        compound_ident,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col.f": 1}},
            ],
        }),
        Stage::Sort(Sort {
            specs: vec![Asc(
                Expression::FieldAccess(FieldAccess{
                    field: "f".to_string(),
                    expr: Reference(("col", 0u16).into()).into(),
                }).into(),
            )],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        other_expr,
        Err(Error::InvalidSortKey),
        Stage::Sort(Sort {
            specs: vec![Asc(Expression::Literal(Literal::Integer(1)).into())],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })
            .into(),
        }),
    );
    test_codegen_plan!(
        non_ident_field_reference,
        Err(Error::InvalidSortKey),
        Stage::Sort(Sort {
            specs: vec![Asc(Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    map! {"a".into() => Expression::Literal(Literal::Integer(1))}
                )
                .into(),
                field: "sub".to_string(),
            })
            .into(),)],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })
            .into(),
        }),
    );
}

mod limit_offset {
    use crate::ir::*;

    test_codegen_plan!(
        limit_simple,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$limit": 1u64},
            ],
        }),
        Stage::Limit(Limit {
            limit: 1,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        offset_simple,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$skip": 1u64},
            ],
        }),
        Stage::Offset(Offset {
            offset: 1,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
}

mod literal {
    use crate::ir::{Expression::*, Literal::*};
    use bson::{bson, Bson};

    test_codegen_expr!(null, Ok(bson!({ "$literal": Bson::Null })), Literal(Null),);
    test_codegen_expr!(bool, Ok(bson!({"$literal": true})), Literal(Boolean(true)),);
    test_codegen_expr!(
        string,
        Ok(bson!({"$literal": "abc"})),
        Literal(String("abc".into())),
    );
    test_codegen_expr!(int, Ok(bson!({"$literal": 5_i32})), Literal(Integer(5)),);
    test_codegen_expr!(long, Ok(bson!({"$literal": 6_i64})), Literal(Long(6)),);
    test_codegen_expr!(double, Ok(bson!({"$literal": 7.0})), Literal(Double(7.0)),);
}

mod reference {
    use crate::{
        codegen::{mql::MappingRegistry, Error},
        ir::Expression::*,
    };
    use bson::Bson;

    test_codegen_expr!(
        not_found,
        MappingRegistry::default(),
        Err(Error::ReferenceNotFound(("f", 0u16).into())),
        Reference(("f", 0u16).into()),
    );

    test_codegen_expr!(
        found,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(Bson::String("$f".into())),
        Reference(("f", 0u16).into()),
    );
}

mod array {
    use crate::ir::{Expression::*, Literal};
    use bson::bson;

    test_codegen_expr!(empty, Ok(bson!([])), Array(vec![]),);
    test_codegen_expr!(
        non_empty,
        Ok(bson!([{"$literal": "abc"}])),
        Array(vec![Literal(Literal::String("abc".into()))]),
    );
    test_codegen_expr!(
        nested,
        Ok(bson!([{ "$literal": null }, [{ "$literal": null }]])),
        Array(vec![
            Literal(Literal::Null),
            Array(vec![Literal(Literal::Null)])
        ]),
    );
}

mod document {
    use crate::{
        codegen::Error,
        ir::{Expression::*, Literal},
        map,
    };
    use bson::bson;

    test_codegen_expr!(empty, Ok(bson!({"$literal": {}})), Document(map! {}),);
    test_codegen_expr!(
        non_empty,
        Ok(bson!({"foo": {"$literal": 1}})),
        Document(map! {"foo".to_string() => Literal(Literal::Integer(1)),}),
    );
    test_codegen_expr!(
        nested,
        Ok(bson!({"foo": {"$literal": 1}, "bar": {"baz": {"$literal": 2}}})),
        Document(map! {
            "foo".to_string() => Literal(Literal::Integer(1)),
            "bar".to_string() => Document(map!{
                "baz".to_string() => Literal(Literal::Integer(2))
            }),
        }),
    );
    test_codegen_expr!(
        dollar_prefixed_key_disallowed,
        Err(Error::DotsOrDollarsInDocumentKey),
        Document(map! {"$foo".to_string() => Literal(Literal::Integer(1)),}),
    );
    test_codegen_expr!(
        key_containing_dot_disallowed,
        Err(Error::DotsOrDollarsInDocumentKey),
        Document(map! {"foo.bar".to_string() => Literal(Literal::Integer(1)),}),
    );
}

mod field_access {
    use crate::{codegen::mql::MappingRegistry, ir::*, map};
    use bson::Bson;

    test_codegen_expr!(
        reference,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(Bson::String("$f.sub".to_string())),
        Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "sub".to_string(),
        }),
    );
    test_codegen_expr!(
        field_access,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(Bson::String("$f.sub.sub".to_string())),
        Expression::FieldAccess(FieldAccess {
            field: "sub".to_string(),
            expr: Expression::FieldAccess(FieldAccess {
                expr: Expression::Reference(("f", 0u16).into()).into(),
                field: "sub".to_string(),
            })
            .into(),
        }),
    );
    test_codegen_expr!(
        expr,
        Ok(bson::bson!({"$getField": {
            "field": "sub",
            "input": {"a": {"$literal": 1}},
        }})),
        Expression::FieldAccess(FieldAccess {
            expr: Expression::Document(
                map! {"a".into() => Expression::Literal(Literal::Integer(1))}
            )
            .into(),
            field: "sub".to_string(),
        }),
    );

    test_codegen_expr!(
        success_on_non_reference_expr,
        Ok(bson::bson!({"$getField": {
            "field": "sub",
            "input": {"$literal": "f"},
        }})),
        Expression::FieldAccess(FieldAccess {
            expr: Expression::Literal(Literal::String("f".into())).into(),
            field: "sub".to_string(),
        }),
    );
    test_codegen_expr!(
        dollar_prefixed_field,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson!({"$getField": {"field": "$sub", "input": "$f"}})),
        Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "$sub".to_string(),
        }),
    );
    test_codegen_expr!(
        field_contains_dollar,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(Bson::String("$f.s$ub".to_string())),
        Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "s$ub".to_string(),
        }),
    );
    test_codegen_expr!(
        field_contains_dot,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson!({"$getField": {"field": "s.ub", "input": "$f"}})),
        Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "s.ub".to_string(),
        }),
    );
}

mod searched_case_expression {
    use crate::ir::*;
    test_codegen_expr!(
        one_case,
        Ok(bson::bson!({"$switch": 
            {"branches": [{"case": {"$literal": true}, 
                                            "then": {"$literal": "first case"}}],
            "default": {"$literal": "else case"}}})),
        Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(Literal::Boolean(true))),
                then: Box::new(Expression::Literal(Literal::String(
                    "first case".to_string()
                )))
            }],
            else_branch: Box::new(Expression::Literal(Literal::String(
                "else case".to_string()
            )))
        }),
    );

    test_codegen_expr!(
        multiple_cases,
        Ok(bson::bson!({"$switch": 
            {"branches": [{"case": {"$literal": false}, "then": {"$literal": "first case"}},
            {"case": {"$literal": true}, "then": {"$literal": "second case"}},
            {"case": {"$literal": true}, "then": {"$literal": "third case"}}],
            "default": {"$literal": "else case"}}})),
        Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(false))),
                    then: Box::new(Expression::Literal(Literal::String(
                        "first case".to_string()
                    )))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(true))),
                    then: Box::new(Expression::Literal(Literal::String(
                        "second case".to_string()
                    )))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(true))),
                    then: Box::new(Expression::Literal(Literal::String(
                        "third case".to_string()
                    )))
                }
            ],
            else_branch: Box::new(Expression::Literal(Literal::String(
                "else case".to_string()
            )))
        }),
    );
}

mod simple_case_expression {
    use crate::ir::*;
    test_codegen_expr!(
        one_case,
        Ok(
            bson::bson!({"$let": {"vars": {"target": {"$literal": "co"}}}, "in": {"$switch": 
            {"branches": [{"case": {"$sqlEq": ["$$target", {"$literal": true}]}, 
                                            "then": {"$literal": "true case"}}],
            "default": {"$literal": "else case"}}}})
        ),
        Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(Literal::String("co".to_string()))),
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(Literal::Boolean(true))),
                then: Box::new(Expression::Literal(Literal::String(
                    "true case".to_string()
                )))
            }],
            else_branch: Box::new(Expression::Literal(Literal::String(
                "else case".to_string()
            )))
        }),
    );
    test_codegen_expr!(
        multiple_cases,
        Ok(
            bson::bson!({"$let": {"vars": {"target": {"$literal": "co"}}}, "in": {"$switch": 
            {"branches": [{"case": {"$sqlEq": ["$$target", {"$literal": false}]}, 
                                            "then": {"$literal": "first case"}},
            {"case": {"$sqlEq": ["$$target", {"$literal": false}]}, 
                                            "then": {"$literal": "second case"}},
            {"case": {"$sqlEq": ["$$target", {"$literal": false}]}, 
                                            "then": {"$literal": "third case"}}],
            "default": {"$literal": "else case"}}}})
        ),
        Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(Literal::String("co".to_string()))),
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(false))),
                    then: Box::new(Expression::Literal(Literal::String(
                        "first case".to_string()
                    )))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(false))),
                    then: Box::new(Expression::Literal(Literal::String(
                        "second case".to_string()
                    )))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(Literal::Boolean(false))),
                    then: Box::new(Expression::Literal(Literal::String(
                        "third case".to_string()
                    )))
                }
            ],
            else_branch: Box::new(Expression::Literal(Literal::String(
                "else case".to_string()
            )))
        }),
    );
}

mod join {
    use crate::ir::*;

    test_codegen_plan!(
        join_simple,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}},
            bson::doc!{"$join": {"collection": "col2", "joinType": "inner", "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}]}}],
        }),
        Stage::Join(Join {
            condition: None,
            left: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col2".to_string(),
            }).into(),
            join_type: JoinType::Inner,
        }),
    );
    test_codegen_plan!(
        join_left_different_databases,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}}, bson::doc!{
                "$join":
                {"database": "mydb2",
                "collection": "col2",
                "joinType":"left",
                "let": {"col_0": "$col"},
                "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}],
                "condition": {"$match": {"$expr": {"$literal": true}}}
            }}],
        }),
        Stage::Join(Join {
            condition: Some(Expression::Literal(Literal::Boolean(true))),
            left: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb2".to_string(),
                collection: "col2".to_string(),
            }).into(),
            join_type: JoinType::Left,
        }),
    );
    test_codegen_plan!(
        join_on_array,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}}, bson::doc!{
                "$join":
                {"joinType":"left",
                "pipeline": [
                    {"$documents": [{"$literal": 1}, {"$literal": 1}]},
                    bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                ],
            }}],
        }),
        Stage::Join(Join {
            condition: None,
            left: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
            right: Stage::Array(Array {
                array: vec![Expression::Literal(Literal::Integer(1)),Expression::Literal(Literal::Integer(1))],
                alias: "arr".to_string(),
            }).into(),
            join_type: JoinType::Left,
        }),
    );
    test_codegen_plan!(
        join_condition_references_left,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}}, bson::doc!{
                "$join":
                {"database": "mydb2",
                "collection": "col2",
                "joinType":"left",
                "let": {"col_0": "$col"},
                "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}],
                "condition": {"$match": {"$expr": "$$col_0"}}
            }}],
        }),
        Stage::Join(Join {
            condition: Some(Expression::Reference(("col", 0u16).into())),
            left: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb2".to_string(),
                collection: "col2".to_string(),
            }).into(),
            join_type: JoinType::Left,
        }),
    );
    test_codegen_plan!(
        join_condition_references_right,
        Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}}, bson::doc!{
                "$join":
                {"database": "mydb2",
                "collection": "col2",
                "joinType":"left",
                "let": {"col_0": "$col"},
                "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}],
                "condition": {"$match": {"$expr": "$col2"}}
            }}],
        }),
        Stage::Join(Join {
            condition: Some(Expression::Reference(("col2", 0u16).into())),
            left: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb2".to_string(),
                collection: "col2".to_string(),
            }).into(),
            join_type: JoinType::Left,
        }),
    );
}

mod union {
    use crate::ir::*;

    test_codegen_plan!(
        collection_union_all_collection,
        Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "a": "$$ROOT"}},
                bson::doc!{"$unionWith": {"coll": "b", "pipeline": [{"$project": {"_id": 0, "b": "$$ROOT"}}]}}],
        }),
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "a".to_string(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "bar".to_string(),
                collection: "b".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        array_union_all_array,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": 1}]},
                bson::doc!{"$project": {"_id": 0, "arr1": "$$ROOT"}},
                bson::doc!{"$unionWith": {"pipeline": [
                    {"$documents": [{"$literal": 2}]},
                    {"$project": {"_id": 0, "arr2": "$$ROOT"}},
                ]}}],
        }),
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Stage::Array(Array {
                array: vec![Expression::Literal(Literal::Integer(1))],
                alias: "arr1".to_string(),
            }).into(),
            right: Stage::Array(Array {
                array: vec![Expression::Literal(Literal::Integer(2))],
                alias: "arr2".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        array_union_all_collection,
        Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": 1}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$unionWith": {"coll": "b", "pipeline": [{"$project": {"_id": 0, "b": "$$ROOT"}}]}}],
        }),
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Stage::Array(Array {
                array: vec![Expression::Literal(Literal::Integer(1))],
                alias: "arr".to_string(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "bar".to_string(),
                collection: "b".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        collection_union_all_array,
        Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "a": "$$ROOT"}},
                bson::doc!{"$unionWith": {"pipeline": [
                    {"$documents": [{"$literal": 1}]},
                    bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                ]}}],
        }),
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "a".to_string(),
            }).into(),
            right: Stage::Array(Array {
                array: vec![Expression::Literal(Literal::Integer(1))],
                alias: "arr".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        collection_union_all_with_nested_union_all,
        Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "a": "$$ROOT"}},
                bson::doc!{"$unionWith": {"coll": "b", "pipeline": [
                    {"$project": {"_id": 0, "b": "$$ROOT"}},
                    {"$unionWith": {"pipeline": [
                        {"$documents": [{"$literal": 1}]},
                        {"$project": {"_id": 0, "arr": "$$ROOT"}},
                    ]}}
                ]}}
            ],
        }),
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Stage::Collection(Collection {
                db: "foo".to_string(),
                collection: "a".to_string(),
            }).into(),
            right: Stage::Set(Set {
                operation: SetOperation::UnionAll,
                left: Stage::Collection(Collection {
                    db: "bar".to_string(),
                    collection: "b".to_string(),
                }).into(),
                right: Stage::Array(Array {
                    array: vec![Expression::Literal(Literal::Integer(1))],
                    alias: "arr".to_string(),
                }).into(),
            }).into(),
        }),
    );
}

mod function {
    use crate::{
        codegen::{mql::MappingRegistry, Error},
        ir::{definitions::*, Expression::*, Literal, ScalarFunction::*},
    };
    use bson::bson;

    test_codegen_expr!(
        concat_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$concat": ["$f", {"$literal": "bar"}]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Concat,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::String("bar".to_string())),
            ],
        }),
    );
    test_codegen_expr!(
        like_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("input", 0u16), "input");
            mr.insert(("pattern", 0u16), "pattern");
            mr
        },
        Ok(bson!({"$like": {
            "input": "$input",
            "pattern": "$pattern",
            "escape": "escape",
        }})),
        Like(LikeExpr {
            expr: Expression::Reference(("input", 0u16).into()).into(),
            pattern: Expression::Reference(("pattern", 0u16).into()).into(),
            escape: Some("escape".to_string()),
        }),
    );
    test_codegen_expr!(
        is_number_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$isNumber": "$f"})),
        Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Number,
        }),
    );
    test_codegen_expr!(
        is_null_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$or": [{"$eq": [{"$type": "$f"}, "missing"]},
                                {"$eq": [{"$type": "$f"}, "null"]}]})),
        Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Type(Type::Null),
        }),
    );
    test_codegen_expr!(
        is_missing_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$eq": [{"$type": "$f"}, "missing"]})),
        Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Missing,
        }),
    );
    test_codegen_expr!(
        is_type_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$eq": [{"$type": "$f"}, "double"]})),
        Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Type(Type::Double),
        }),
    );
    test_codegen_expr!(
        pos_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!("$f")),
        ScalarFunction(ScalarFunctionApplication {
            function: Pos,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        neg_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$multiply": [
            "$f", -1
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Neg,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        add_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
        Ok(bson::bson! ({"$add": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Add,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        sub_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$subtract": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Sub,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        mult_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
        Ok(bson::bson! ({"$multiply": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Mul,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        div_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlDivide": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Div,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        lt_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlLt": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Lt,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        lte_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlLte": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Lte,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        ne_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlNe": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Neq,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        eq_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlEq": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Eq,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        gt_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlGt": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Gt,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        gte_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlGte": [
            "$f", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Gte,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        between_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlBetween": [
            "$f", {"$literal": 1}, {"$literal": 10}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Between,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(10))
            ],
        }),
    );
    test_codegen_expr!(
        not_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlNot": ["$f"]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Not,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        and_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
        Ok(bson::bson! ({"$sqlAnd": [
            "$f1", "$f2", {"$literal": true}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: And,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
    );
    test_codegen_expr!(
        or_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
        Ok(bson::bson! ({"$sqlOr": [
            "$f1", "$f2", {"$literal": true}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Or,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
    );
    test_codegen_expr!(
        nullif_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
        Ok(bson::bson! ({"$nullIf": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: NullIf,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        coalesce_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
        Ok(bson::bson! ({"$coalesce": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Coalesce,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        slice_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson::bson! ({"$sqlSlice": [
            "$f", {"$literal": 1}, {"$literal": 1}
        ]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Slice,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
    );
    test_codegen_expr!(
        size_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlSize": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Size,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        position_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlIndexOfCP": [ "$f", {"$literal": "a"}, {"$literal": 1}, {"$literal": 10}]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Position,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(10)),
            ],
        }),
    );
    test_codegen_expr!(
        charlen_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlStrLenCP": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: CharLength,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        octetlen_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlStrLenBytes": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: OctetLength,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        bitlen_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$multiply": [{"$sqlStrLenBytes": "$f"}, 8]})),
        ScalarFunction(ScalarFunctionApplication {
            function: BitLength,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        substring_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlSubstrCP": ["$f", {"$literal": 1}, {"$literal": 10}]})),
        ScalarFunction(ScalarFunctionApplication {
            function: Substring,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(10)),
            ],
        }),
    );
    test_codegen_expr!(
        upper_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlToUpper": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Upper,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        lower_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$sqlToLower": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Lower,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        currenttimestamp_expr,
        Ok(bson!("$$NOW")),
        ScalarFunction(ScalarFunctionApplication {
            function: CurrentTimestamp,
            args: vec![],
        }),
    );
    test_codegen_expr!(
        ltrim_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$ltrim": {"input": "$f", "chars": {"$literal": "a"}}})),
        ScalarFunction(ScalarFunctionApplication {
            function: LTrim,
            args: vec![
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Reference(("f", 0u16).into()),
            ],
        }),
    );
    test_codegen_expr!(
        rtrim_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$rtrim": {"input": "$f", "chars": {"$literal": "a"}}})),
        ScalarFunction(ScalarFunctionApplication {
            function: RTrim,
            args: vec![
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Reference(("f", 0u16).into()),
            ],
        }),
    );
    test_codegen_expr!(
        trim_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$trim": {"input": "$f", "chars": {"$literal": "a"}}})),
        ScalarFunction(ScalarFunctionApplication {
            function: BTrim,
            args: vec![
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Reference(("f", 0u16).into()),
            ],
        }),
    );
    test_codegen_expr!(
        extract_year_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$year": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Year,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        extract_month_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$month": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Month,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        extract_day_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$day": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Day,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        extract_hour_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$hour": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Hour,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        extract_minute_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$minute": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Minute,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        extract_second_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(bson!({"$second": "$f"})),
        ScalarFunction(ScalarFunctionApplication {
            function: Second,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
    );
    test_codegen_expr!(
        computedfieldaccess_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Err(Error::UnsupportedFunction(ComputedFieldAccess)),
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: ComputedFieldAccess,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Long(42)),
            ],
        }),
    );
}

mod group_by {
    use crate::{codegen::Error, ir::definitions::*, map};
    use lazy_static::lazy_static;

    lazy_static! {
        static ref SOURCE_IR: Box<Stage> = Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
        }));
        static ref KEY_FOO_A: Vec<AliasedExpression> = vec![AliasedExpression {
            alias: Some("foo_a".to_string()),
            inner: Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "a".into()
            }),
        },];
    }

    test_codegen_plan!(
        simple,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"f": "$foo"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"f": "$_id.f"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![AliasedExpression {
                alias: Some("f".into()),
                inner: Expression::Reference(("foo", 0u16).into()),
            }],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        aliased_compound_ident,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        unaliased_compound_idents,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"__unaliasedKey1": "$foo.a", "__unaliasedKey2": "$foo.b"}}},
                bson::doc!{"$project": {"_id": 0, "foo": {"a": "$_id.__unaliasedKey1", "b": "$_id.__unaliasedKey2"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![AliasedExpression {
                alias: None,
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            },AliasedExpression {
                alias: None,
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into()
                }),
            },],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        mix_aliased_and_unaliased_keys,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a", "__unaliasedKey2": "$foo.b", "foo_c": "$foo.c"}}},
                bson::doc!{"$project": {"_id": 0, "foo": {"b": "$_id.__unaliasedKey2"}, "__bot": {"foo_a": "$_id.foo_a", "foo_c": "$_id.foo_c"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![AliasedExpression {
                alias: Some("foo_a".to_string()),
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            },
            AliasedExpression {
                alias: None,
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into()
                }),
            },
            AliasedExpression {
                alias: Some("foo_c".to_string()),
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "c".into()
                }),
            }],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        duplicate_generated_alias,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"__unaliasedKey2": "$foo.a", "___unaliasedKey2": "$foo.b"}}},
                bson::doc!{"$project": {"_id": 0, "foo": {"b": "$_id.___unaliasedKey2"}, "__bot": {"__unaliasedKey2": "$_id.__unaliasedKey2"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![AliasedExpression {
                alias: Some("__unaliasedKey2".to_string()),
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            },
            AliasedExpression {
                alias: None,
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into()
                }),
            }],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        invalid_group_key_field_access,
        Err(Error::InvalidGroupKey),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![AliasedExpression {
                alias: None,
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Document(
                        map! {"a".into() => Expression::Literal(Literal::Integer(1))}
                    )),
                    field: "sub".to_string(),
                })
            }],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        add_to_array,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$push": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::AddToArray,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        avg,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlAvg": {"var": "$foo.a", "distinct": false}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Avg,
                    distinct: false,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        count,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlCount": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Count,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        first,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$first": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::First,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        last,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlLast": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Last,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        max,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$max": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Max,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        merge_documents,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlMergeObjects": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::MergeDocuments,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        min,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$min": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Min,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        std_dev_pop,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlStdDevPop": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::StddevPop,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        std_dev_samp,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlStdDevSamp": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::StddevSamp,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        sum,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlSum": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
    test_codegen_plan!(
        count_star,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlCount": {"var": "$$ROOT", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                inner: AggregationExpr::CountStar(true)
            }],
        }),
    );
    test_codegen_plan!(
        agg_alias_underscore_id,
        Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"__id": "$foo.a"}, "__id": {"$sqlSum": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"__id": "$_id.__id", "_id": "$__id"}}},
            ],
        }),
        Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![AliasedExpression {
                alias: Some("__id".to_string()),
                inner: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            },],
            aggregations: vec![AliasedAggregation {
                alias: "_id".into(),
                inner: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    }))
                })
            }],
        }),
    );
}

mod subquery {
    use crate::{
        codegen::{mql::MappingRegistry, Error},
        ir::{binding_tuple::DatasourceName::Bottom, SubqueryModifier::*, *},
        map,
    };
    test_codegen_expr!(
        exists_uncorrelated,
        Ok(bson::bson!(
            {"$subqueryExists": {
                "db": "test",
                "collection": "foo",
                "let": {},
                "pipeline": [{"$project": {"_id": 0,"foo": "$$ROOT"}},]
            }}
        )),
        Expression::Exists(Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
        }))),
    );
    test_codegen_expr!(
        exists_correlated,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr
        },
        Ok(bson::bson!(
            {"$subqueryExists": {
                "db": "test",
                "collection": "bar",
                "let": {"foo_0": "$foo"},
                "pipeline": [
                    {"$project": {"_id": 0,"bar": "$$ROOT"}},
                    {"$project": {"_id": 0,"__bot": {"a": "$$foo_0.a"}}}
                ]
            }}
        )),
        Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "bar".into(),
            })),
            expression: map! {
                (Bottom, 1u16).into() => Expression::Document(map! {
                    "a".into() => Expression::FieldAccess(FieldAccess{
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    })
                })
            }
        }))),
    );
    test_codegen_expr!(
        subquery_expr_uncorrelated,
        Ok(bson::bson!(
            {"$subquery": {
                "db": "test",
                "collection": "foo",
                "let": {},
                "outputPath": ["foo"],
                "pipeline": [{"$project": {"_id": 0,"foo": "$$ROOT"}},]
            }}
        )),
        Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
        }),
    );
    test_codegen_expr!(
        subquery_expr_correlated,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr
        },
        Ok(bson::bson!(
            {"$subquery": {
                "db": "test",
                "collection": "bar",
                "let": {"foo_0": "$foo",},
                "outputPath": ["__bot", "a"],
                "pipeline": [
                    {"$project": {"_id": 0,"bar": "$$ROOT"}},
                    {"$project": {"_id": 0,"__bot": {"a": "$$foo_0.a"}}}
                ]
            }}
        )),
        Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into()
                        })
                    })
                }
            }))
        }),
    );
    test_codegen_expr!(
        subquery_expr_no_db_or_coll,
        Ok(bson::bson!(
            {"$subquery": {"let": {},"outputPath": ["arr"],"pipeline": [
                {"$documents": []},
                {"$project": {"_id": 0, "arr": "$$ROOT"}},
            ]}}
        )),
        Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("arr", 1u16).into())),
            subquery: Box::new(Stage::Array(Array {
                array: vec![],
                alias: "arr".into(),
            })),
        }),
    );
    test_codegen_expr!(
        output_expr_field_contains_dot,
        Ok(bson::bson!(
            {"$subquery": {
                "db": "test",
                "collection": "foo",
                "let": {},
                "outputPath": ["foo", "bar.a"],
                "pipeline": [
                    {"$project": {"_id": 0, "foo": "$$ROOT"}},
                    {"$project": {"_id": 0, "foo": "$foo"}}
                ]
            }}
        )),
        Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                field: "bar.a".into()
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into()
                })),
                expression: map! {
                    ("foo", 1u16).into() => Expression::Reference(("foo", 1u16).into())
                }
            }))
        }),
    );
    // This test verifies that we are using the datasource's MQL field name
    // in the output path. We create an MQL field name that doesn't match the
    // corresponding datasource by forcing a naming conflict in the project
    // stage. The translation engine could never actually produce a query like
    // this one, though, since a subquery expression's degree must be exactly 1.
    test_codegen_expr!(
        use_datasource_mql_name,
        Ok(bson::bson!(
            {"$subquery": {
                "db": "test",
                "collection": "__bot",
                "let": {},
                "outputPath": vec!["___bot"],
                "pipeline": [
                    {"$project": {"_id": 0, "__bot": "$$ROOT"}},
                    {"$project": {"_id": 0, "___bot": "$__bot", "__bot": {"$literal": 43.0}}}
                ]
            }}
        )),
        Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
            subquery: Box::new(Stage::Project(Project {
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Reference(("__bot", 1u16).into()),
                    ("__bot", 1u16).into() => Expression::Literal(Literal::Double(43.0)),
                },
                source: Stage::Collection(Collection {
                    db: "test".to_string(),
                    collection: "__bot".to_string(),
                })
                .into(),
            }))
        }),
    );
    test_codegen_expr!(
        subquery_comparison_uncorrelated,
        Ok(bson::bson!(
            {"$subqueryComparison": {
                "op": "eq",
                "modifier": "any",
                "arg": {"$literal": 5},
                "subquery": {
                    "db": "test",
                    "collection": "foo",
                    "let": {},
                    "outputPath": ["foo"],
                    "pipeline": [{"$project": {"_id": 0,"foo": "$$ROOT"}}]
                }
            }}
        )),
        Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: Any,
            argument: Box::new(Expression::Literal(Literal::Integer(5))),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                subquery: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                }))
            }
        }),
    );
    test_codegen_expr!(
        subquery_comparison_correlated,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr.insert(("x", 0u16), "x");
            mr
        },
        Ok(bson::bson!(
            {"$subqueryComparison": {
                "op": "eq",
                "modifier": "any",
                "arg": "$x",
                "subquery": {
                    "db": "test",
                    "collection": "bar",
                    "let": {"foo_0": "$foo","x_0": "$x",},
                    "outputPath": ["__bot", "a"],
                    "pipeline": [
                        {"$project": {"_id": 0,"bar": "$$ROOT"}},
                        {"$project": {"_id": 0,"__bot": {"a": "$$foo_0.a"}}}
                    ]
                }
            }}
        )),
        Expression::SubqueryComparison(SubqueryComparison {
            operator: SubqueryComparisonOp::Eq,
            modifier: Any,
            argument: Box::new(Expression::Reference(("x", 0u16).into())),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                    field: "a".into()
                })),
                subquery: Box::new(Stage::Project(Project {
                    source: Box::new(Stage::Collection(Collection {
                        db: "test".into(),
                        collection: "bar".into(),
                    })),
                    expression: map! {
                        (Bottom, 1u16).into() => Expression::Document(map! {
                            "a".into() => Expression::FieldAccess(FieldAccess{
                                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                                field: "a".into()
                            })
                        })
                    }
                }))
            }
        }),
    );
    test_codegen_expr!(
        invalid_output_expr,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr.insert(("x", 0u16), "x");
            mr
        },
        Err(Error::NoFieldPathForExpr),
        Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Literal(Literal::Integer(5))),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into()
                        })
                    })
                }
            }))
        }),
    );
}
