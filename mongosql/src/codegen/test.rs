macro_rules! test_codegen_plan {
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

    ($func_name:ident, expected = Err($expected_err:expr), input = $input:expr,) => {
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
    ($func_name:ident, expected = $expected:expr, input = $input:expr, $(mapping_registry = $mapping_registry:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::codegen::mql::{MqlCodeGenerator, MqlMappingRegistry};
            let expected = $expected;
            let input = $input;

            #[allow(unused_mut, unused_assignments)]
            let mut mapping_registry = MqlMappingRegistry::default();
            $(mapping_registry = $mapping_registry;)?

            let gen = MqlCodeGenerator {
                mapping_registry,
                scope_level: 0u16,
            };
            assert_eq!(expected, gen.codegen_expression(input));
        }
    };
}

mod collection {
    use crate::ir::*;

    test_codegen_plan!(
        simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
            ],
        }),
        input = Stage::Collection(Collection {
            db: "mydb".to_string(),
            collection: "col".to_string(),
        }),
    );
    test_codegen_plan!(
        collection_named_id,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("_id".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": "$$ROOT"}},
            ],
        }),
        input = Stage::Collection(Collection {
            db: "mydb".to_string(),
            collection: "_id".to_string(),
        }),
    );
}

mod array_stage {
    use crate::ir::*;

    test_codegen_plan!(
        empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
            ],
        }),
        input = Stage::Array(Array {
            array: vec![],
            alias: "arr".to_string(),
        }),
    );
    test_codegen_plan!(
        empty_id_alias,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": "$$ROOT"}},
            ],
        }),
        input = Stage::Array(Array {
            array: vec![],
            alias: "_id".to_string(),
        }),
    );
    test_codegen_plan!(
        non_empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": false}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
            ],
        }),
        input = Stage::Array(Array {
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
        map, unchecked_unique_linked_hash_map,
    };

    test_codegen_plan!(
        simple,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": 0, "a": {"$literal": 1}, "b": {"$literal": 2}, "c": {"$literal": 3}}},
            ],
        }),
        input = Stage::Project(Project {
            expression: map! {
                ("a", 0u16).into() => Expression::Literal(Literal::Integer(1)),
                ("b", 0u16).into() => Expression::Literal(Literal::Integer(2)),
                ("c", 0u16).into() => Expression::Literal(Literal::Integer(3)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": 0}},
            ],
        }),
        input = Stage::Project(Project {
            expression: map! {},
            source: Stage::Array(Array {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        source_bindings_available_in_project,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": 0, "foo": "$arr"}},
            ],
        }),
        input = Stage::Project(Project {
            expression: map! {
                ("foo", 0u16).into() => Expression::Reference(("arr", 0u16).into()),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        user_defined_id_projection,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$project": {"_id": {"$literal": 42.0}, "foo": {"$literal": 44.0}}},
            ],
        }),
        input = Stage::Project(Project {
            expression: map! {
                ("_id", 0u16).into() => Expression::Literal(Literal::Double(42.0)),
                ("foo", 0u16).into() => Expression::Literal(Literal::Double(44.0)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{})],
                alias: "arr".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        user_bot_conflict,
        expected = Ok({
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
        input = Stage::Project(Project {
            expression: map! {
                Key{ datasource: DatasourceName::Bottom, scope: 0u16 } => Expression::Reference(("__bot", 0u16).into()),
                ("__bot", 0u16).into() => Expression::Literal(Literal::Double(43.0)),
                ("___bot", 0u16).into() => Expression::Literal(Literal::Double(44.0)),
                ("_____bot", 0u16).into() => Expression::Literal(Literal::Double(45.0)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{"a".into() => Expression::Literal(Literal::Integer(42))})],
                alias: "__bot".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        dot_project_field,
        expected = Err(Error::DotsOrDollarsInProjectField),
        input = Stage::Project(Project {
            expression: map! {
                ("a.b", 0u16).into() => Expression::Literal(Literal::Integer(1)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map! {})],
                alias: "arr".to_string(),
            })
            .into(),
        }),
    );

    test_codegen_plan!(
        dollar_project_field,
        expected = Err(Error::DotsOrDollarsInProjectField),
        input = Stage::Project(Project {
            expression: map! {
                ("$a", 0u16).into() => Expression::Literal(Literal::Integer(1)),
            },
            source: Stage::Array(Array {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map! {})],
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
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$match": {"$expr": {"$literal": true}}},
            ],
        }),
        input = Stage::Filter(Filter {
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
        unchecked_unique_linked_hash_map,
    };

    test_codegen_plan!(
        empty,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": []},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$sort": {}},
            ],
        }),
        input = Stage::Sort(Sort {
            specs: vec![],
            source: Stage::Array(Array {
                array: vec![],
                alias: "arr".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        single_spec_asc,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col": 1}},
            ],
        }),
        input = Stage::Sort(Sort {
            specs: vec![Asc(Reference(("col", 0u16).into()).into())],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        single_spec_dsc,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col": -1}},
            ],
        }),
        input = Stage::Sort(Sort {
            specs: vec![Desc(Reference(("col", 0u16).into()).into())],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        multi_spec,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col": 1, "col.a": -1}},
            ],
        }),
        input = Stage::Sort(Sort {
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
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$sort": {"col.f": 1}},
            ],
        }),
        input = Stage::Sort(Sort {
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
        expected = Err(Error::InvalidSortKey),
        input = Stage::Sort(Sort {
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
        expected = Err(Error::InvalidSortKey),
        input = Stage::Sort(Sort {
            specs: vec![Asc(Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(Literal::Integer(1))}
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
    use crate::{codegen::Error, ir::*};

    test_codegen_plan!(
        limit_simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$limit": 1i64},
            ],
        }),
        input = Stage::Limit(Limit {
            limit: 1,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        limit_out_of_i64_range,
        expected = Err(Error::LimitOutOfI64Range(u64::MAX)),
        input = Stage::Limit(Limit {
            limit: u64::MAX,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })
            .into(),
        }),
    );
    test_codegen_plan!(
        offset_simple,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$skip": 1i64},
            ],
        }),
        input = Stage::Offset(Offset {
            offset: 1,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
    test_codegen_plan!(
        offset_out_of_i64_range,
        expected = Err(Error::OffsetOutOfI64Range(u64::MAX)),
        input = Stage::Offset(Offset {
            offset: u64::MAX,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            })
            .into(),
        }),
    );
}

mod literal {
    use crate::ir::{Expression::*, Literal::*};
    use bson::{bson, Bson};

    test_codegen_expr!(
        null,
        expected = Ok(bson!({ "$literal": Bson::Null })),
        input = Literal(Null),
    );
    test_codegen_expr!(
        bool,
        expected = Ok(bson!({"$literal": true})),
        input = Literal(Boolean(true)),
    );
    test_codegen_expr!(
        string,
        expected = Ok(bson!({"$literal": "abc"})),
        input = Literal(String("abc".into())),
    );
    test_codegen_expr!(
        int,
        expected = Ok(bson!({"$literal": 5_i32})),
        input = Literal(Integer(5)),
    );
    test_codegen_expr!(
        long,
        expected = Ok(bson!({"$literal": 6_i64})),
        input = Literal(Long(6)),
    );
    test_codegen_expr!(
        double,
        expected = Ok(bson!({"$literal": 7.0})),
        input = Literal(Double(7.0)),
    );
}

mod reference {
    use crate::{codegen::Error, ir::Expression::*};
    use bson::Bson;

    test_codegen_expr!(
        not_found,
        expected = Err(Error::ReferenceNotFound(("f", 0u16).into())),
        input = Reference(("f", 0u16).into()),
    );

    test_codegen_expr!(
        found,
        expected = Ok(Bson::String("$f".into())),
        input = Reference(("f", 0u16).into()),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
}

mod array {
    use crate::ir::{Expression::*, Literal};
    use bson::bson;

    test_codegen_expr!(empty, expected = Ok(bson!([])), input = Array(vec![]),);
    test_codegen_expr!(
        non_empty,
        expected = Ok(bson!([{"$literal": "abc"}])),
        input = Array(vec![Literal(Literal::String("abc".into()))]),
    );
    test_codegen_expr!(
        nested,
        expected = Ok(bson!([{ "$literal": null }, [{ "$literal": null }]])),
        input = Array(vec![
            Literal(Literal::Null),
            Array(vec![Literal(Literal::Null)])
        ]),
    );
}

mod document {
    use crate::{
        codegen::Error,
        ir::{Expression::*, Literal},
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_expr!(
        empty,
        expected = Ok(bson!({"$literal": {}})),
        input = Document(unchecked_unique_linked_hash_map! {}),
    );
    test_codegen_expr!(
        non_empty,
        expected = Ok(bson!({"foo": {"$literal": 1}})),
        input = Document(
            unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(Literal::Integer(1)),}
        ),
    );
    test_codegen_expr!(
        nested,
        expected = Ok(bson!({"foo": {"$literal": 1}, "bar": {"baz": {"$literal": 2}}})),
        input = Document(unchecked_unique_linked_hash_map! {
            "foo".to_string() => Literal(Literal::Integer(1)),
            "bar".to_string() => Document(unchecked_unique_linked_hash_map!{
                "baz".to_string() => Literal(Literal::Integer(2))
            }),
        }),
    );
    test_codegen_expr!(
        dollar_prefixed_key_disallowed,
        expected = Err(Error::DotsOrDollarsInDocumentKey),
        input = Document(
            unchecked_unique_linked_hash_map! {"$foo".to_string() => Literal(Literal::Integer(1)),}
        ),
    );
    test_codegen_expr!(
        key_containing_dot_disallowed,
        expected = Err(Error::DotsOrDollarsInDocumentKey),
        input = Document(
            unchecked_unique_linked_hash_map! {"foo.bar".to_string() => Literal(Literal::Integer(1)),}
        ),
    );
}

mod field_access {
    use crate::{ir::*, unchecked_unique_linked_hash_map};
    use bson::Bson;

    test_codegen_expr!(
        reference,
        expected = Ok(Bson::String("$f.sub".to_string())),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "sub".to_string(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        field_access,
        expected = Ok(Bson::String("$f.sub.sub".to_string())),
        input = Expression::FieldAccess(FieldAccess {
            field: "sub".to_string(),
            expr: Expression::FieldAccess(FieldAccess {
                expr: Expression::Reference(("f", 0u16).into()).into(),
                field: "sub".to_string(),
            })
            .into(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        expr,
        expected = Ok(bson::bson!({"$getField": {
            "field": "sub",
            "input": {"a": {"$literal": 1}},
        }})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Document(
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(Literal::Integer(1))}
            )
            .into(),
            field: "sub".to_string(),
        }),
    );

    test_codegen_expr!(
        success_on_non_reference_expr,
        expected = Ok(bson::bson!({"$getField": {
            "field": "sub",
            "input": {"$literal": "f"},
        }})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Literal(Literal::String("f".into())).into(),
            field: "sub".to_string(),
        }),
    );
    test_codegen_expr!(
        dollar_prefixed_field,
        expected = Ok(bson::bson!({"$getField": {"field": "$sub", "input": "$f"}})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "$sub".to_string(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        field_contains_dollar,
        expected = Ok(Bson::String("$f.s$ub".to_string())),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "s$ub".to_string(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        field_contains_dot,
        expected = Ok(bson::bson!({"$getField": {"field": "s.ub", "input": "$f"}})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "s.ub".to_string(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
}

mod searched_case_expression {
    use crate::ir::*;
    test_codegen_expr!(
        one_case,
        expected = Ok(bson::bson!({"$switch": 
            {"branches": [{"case": {"$literal": true}, 
                                            "then": {"$literal": "first case"}}],
            "default": {"$literal": "else case"}}})),
        input = Expression::SearchedCase(SearchedCaseExpr {
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
        expected = Ok(bson::bson!({"$switch": 
            {"branches": [{"case": {"$literal": false}, "then": {"$literal": "first case"}},
            {"case": {"$literal": true}, "then": {"$literal": "second case"}},
            {"case": {"$literal": true}, "then": {"$literal": "third case"}}],
            "default": {"$literal": "else case"}}})),
        input = Expression::SearchedCase(SearchedCaseExpr {
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
        expected = Ok(
            bson::bson!({"$let": {"vars": {"target": {"$literal": "co"}}}, "in": {"$switch": 
            {"branches": [{"case": {"$sqlEq": ["$$target", {"$literal": true}]}, 
                                            "then": {"$literal": "true case"}}],
            "default": {"$literal": "else case"}}}})
        ),
        input = Expression::SimpleCase(SimpleCaseExpr {
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
        expected = Ok(
            bson::bson!({"$let": {"vars": {"target": {"$literal": "co"}}}, "in": {"$switch": 
            {"branches": [{"case": {"$sqlEq": ["$$target", {"$literal": false}]}, 
                                            "then": {"$literal": "first case"}},
            {"case": {"$sqlEq": ["$$target", {"$literal": false}]}, 
                                            "then": {"$literal": "second case"}},
            {"case": {"$sqlEq": ["$$target", {"$literal": false}]}, 
                                            "then": {"$literal": "third case"}}],
            "default": {"$literal": "else case"}}}})
        ),
        input = Expression::SimpleCase(SimpleCaseExpr {
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
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("col".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "col": "$$ROOT"}},
            bson::doc!{"$join": {"collection": "col2", "joinType": "inner", "pipeline": [{"$project": {"_id": 0, "col2": "$$ROOT"}}]}}],
        }),
        input = Stage::Join(Join {
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
        expected = Ok({
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
        input = Stage::Join(Join {
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
        expected = Ok({
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
        input = Stage::Join(Join {
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
        expected = Ok({
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
        input = Stage::Join(Join {
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
        expected = Ok({
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
        input = Stage::Join(Join {
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
        expected = Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "a": "$$ROOT"}},
                bson::doc!{"$unionWith": {"coll": "b", "pipeline": [{"$project": {"_id": 0, "b": "$$ROOT"}}]}}],
        }),
        input = Stage::Set(Set {
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
        expected = Ok({
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
        input = Stage::Set(Set {
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
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [{"$literal": 1}]},
                bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                bson::doc!{"$unionWith": {"coll": "b", "pipeline": [{"$project": {"_id": 0, "b": "$$ROOT"}}]}}],
        }),
        input = Stage::Set(Set {
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
        expected = Ok({
            database: Some("foo".to_string()),
            collection: Some("a".to_string()),
            pipeline: vec![bson::doc!{"$project": {"_id" : 0, "a": "$$ROOT"}},
                bson::doc!{"$unionWith": {"pipeline": [
                    {"$documents": [{"$literal": 1}]},
                    bson::doc!{"$project": {"_id": 0, "arr": "$$ROOT"}},
                ]}}],
        }),
        input = Stage::Set(Set {
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
        expected = Ok({
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
        input = Stage::Set(Set {
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
        codegen::Error,
        ir::{definitions::*, Expression::*, Literal, ScalarFunction::*},
    };
    use bson::bson;

    test_codegen_expr!(
        concat_expr,
        expected = Ok(bson::bson! ({"$concat": ["$f", {"$literal": "bar"}]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Concat,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::String("bar".to_string())),
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        like_expr,
        expected = Ok(bson!({"$like": {
            "input": "$input",
            "pattern": "$pattern",
            "escape": "escape",
        }})),
        input = Like(LikeExpr {
            expr: Expression::Reference(("input", 0u16).into()).into(),
            pattern: Expression::Reference(("pattern", 0u16).into()).into(),
            escape: Some("escape".to_string()),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("input", 0u16), "input");
            mr.insert(("pattern", 0u16), "pattern");
            mr
        },
    );
    test_codegen_expr!(
        is_number_expr,
        expected = Ok(bson!({"$isNumber": "$f"})),
        input = Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Number,
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        is_null_expr,
        expected = Ok(bson!({"$or": [{"$eq": [{"$type": "$f"}, "missing"]},
                                {"$eq": [{"$type": "$f"}, "null"]}]})),
        input = Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Type(Type::Null),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        is_missing_expr,
        expected = Ok(bson!({"$eq": [{"$type": "$f"}, "missing"]})),
        input = Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Missing,
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        is_type_expr,
        expected = Ok(bson!({"$eq": [{"$type": "$f"}, "double"]})),
        input = Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Type(Type::Double),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        pos_expr,
        expected = Ok(bson!("$f")),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Pos,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        neg_expr,
        expected = Ok(bson::bson! ({"$multiply": [
            "$f", {"$literal": -1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Neg,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        add_expr,
        expected = Ok(bson::bson! ({"$add": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Add,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        sub_expr,
        expected = Ok(bson::bson! ({"$subtract": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Sub,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        mult_expr,
        expected = Ok(bson::bson! ({"$multiply": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Mul,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        div_expr,
        expected = Ok(bson::bson! ({"$sqlDivide": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Div,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        lt_expr,
        expected = Ok(bson::bson! ({"$sqlLt": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Lt,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        lte_expr,
        expected = Ok(bson::bson! ({"$sqlLte": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Lte,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        ne_expr,
        expected = Ok(bson::bson! ({"$sqlNe": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Neq,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        eq_expr,
        expected = Ok(bson::bson! ({"$sqlEq": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Eq,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        gt_expr,
        expected = Ok(bson::bson! ({"$sqlGt": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Gt,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        gte_expr,
        expected = Ok(bson::bson! ({"$sqlGte": [
            "$f", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Gte,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        between_expr,
        expected = Ok(bson::bson! ({"$sqlBetween": [
            "$f", {"$literal": 1}, {"$literal": 10}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Between,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(10))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        not_expr,
        expected = Ok(bson!({"$sqlNot": ["$f"]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Not,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        and_expr,
        expected = Ok(bson::bson! ({"$sqlAnd": [
            "$f1", "$f2", {"$literal": true}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: And,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        or_expr,
        expected = Ok(bson::bson! ({"$sqlOr": [
            "$f1", "$f2", {"$literal": true}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Or,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Boolean(true))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        nullif_expr,
        expected = Ok(bson::bson! ({"$nullIf": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: NullIf,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        coalesce_expr,
        expected = Ok(bson::bson! ({"$coalesce": [
            "$f1", "$f2", {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Coalesce,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into()),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        slice_expr,
        expected = Ok(bson::bson! ({"$sqlSlice": [
            "$f", {"$literal": 1}, {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Slice,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(1))
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        size_expr,
        expected = Ok(bson!({"$sqlSize": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Size,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        position_expr,
        expected = Ok(
            bson!({"$sqlIndexOfCP": [ "$f", {"$literal": "a"}, {"$literal": 1}, {"$literal": 10}]})
        ),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Position,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(10)),
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        charlen_expr,
        expected = Ok(bson!({"$sqlStrLenCP": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: CharLength,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        octetlen_expr,
        expected = Ok(bson!({"$sqlStrLenBytes": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: OctetLength,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        bitlen_expr,
        expected = Ok(bson!({"$multiply": [{"$sqlStrLenBytes": "$f"}, {"$literal": 8}]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: BitLength,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        substring_expr,
        expected = Ok(bson!({"$sqlSubstrCP": ["$f", {"$literal": 1}, {"$literal": 10}]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Substring,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Integer(1)),
                Expression::Literal(Literal::Integer(10)),
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        upper_expr,
        expected = Ok(bson!({"$sqlToUpper": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Upper,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        lower_expr,
        expected = Ok(bson!({"$sqlToLower": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Lower,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        currenttimestamp_expr,
        expected = Ok(bson!("$$NOW")),
        input = ScalarFunction(ScalarFunctionApplication {
            function: CurrentTimestamp,
            args: vec![],
        }),
    );
    test_codegen_expr!(
        ltrim_expr,
        expected = Ok(bson!({"$ltrim": {"input": "$f", "chars": {"$literal": "a"}}})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: LTrim,
            args: vec![
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Reference(("f", 0u16).into()),
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        rtrim_expr,
        expected = Ok(bson!({"$rtrim": {"input": "$f", "chars": {"$literal": "a"}}})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: RTrim,
            args: vec![
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Reference(("f", 0u16).into()),
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        trim_expr,
        expected = Ok(bson!({"$trim": {"input": "$f", "chars": {"$literal": "a"}}})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: BTrim,
            args: vec![
                Expression::Literal(Literal::String("a".to_string())),
                Expression::Reference(("f", 0u16).into()),
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_year_expr,
        expected = Ok(bson!({"$year": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Year,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_month_expr,
        expected = Ok(bson!({"$month": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Month,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_day_expr,
        expected = Ok(bson!({"$day": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Day,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_hour_expr,
        expected = Ok(bson!({"$hour": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Hour,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_minute_expr,
        expected = Ok(bson!({"$minute": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Minute,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_second_expr,
        expected = Ok(bson!({"$second": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Second,
            args: vec![Expression::Reference(("f", 0u16).into()),],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        computedfieldaccess_expr,
        expected = Err(Error::UnsupportedFunction(ComputedFieldAccess)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ComputedFieldAccess,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(Literal::Long(42)),
            ],
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
}

mod group_by {
    use crate::{codegen::Error, ir::definitions::*, map, unchecked_unique_linked_hash_map};
    use lazy_static::lazy_static;

    lazy_static! {
        static ref SOURCE_IR: Box<Stage> = Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
        }));
        static ref KEY_FOO_A: Vec<OptionallyAliasedExpr> =
            vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "foo_a".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            })];
    }

    test_codegen_plan!(
        simple,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"f": "$foo"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"f": "$_id.f"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "f".into(),
                expr: Expression::Reference(("foo", 0u16).into()),
            })],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        user_defined_id_group_should_work,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"_id": "$foo"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"_id": "$_id._id"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "_id".into(),
                expr: Expression::Reference(("foo", 0u16).into()),
            })],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        pass_through_id_group_should_work,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"__unaliasedKey1": "$foo._id"}}},
                bson::doc!{"$project": {"_id": 0, "foo": {"_id": "$_id.__unaliasedKey1"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(
                    FieldAccess{
                        expr: Expression::Reference(("foo", 0u16).into()).into(),
                        field: "_id".into(),
                    }),
            )],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        unaliased_id_datasource_group_key_should_not_remove_id_in_project,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$project": {"_id": "$foo"}},
                bson::doc!{"$group": {"_id": {"__unaliasedKey1": "$_id.a"}}},
                bson::doc!{"$project": {"_id": {"a": "$_id.__unaliasedKey1"}}},
            ],
        }),
        input = Stage::Group(Group {
            source:Box::new(Stage::Project(Project {
                source: SOURCE_IR.clone(),
                expression: map! {
                    ("_id", 0u16).into() =>
                        Expression::Reference(("foo", 0u16).into())
                }
            },)),
            keys: vec![OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(
                    FieldAccess{
                        expr: Expression::Reference(("_id", 0u16).into()).into(),
                        field: "a".into(),
                    }
                )
            )],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        aliased_compound_ident,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        unaliased_compound_idents,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"__unaliasedKey1": "$foo.a", "__unaliasedKey2": "$foo.b"}}},
                bson::doc!{"$project": {"_id": 0, "foo": {"a": "$_id.__unaliasedKey1", "b": "$_id.__unaliasedKey2"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            ), OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into()
                }),
            )],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        mix_aliased_and_unaliased_keys,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a", "__unaliasedKey2": "$foo.b", "foo_c": "$foo.c"}}},
                bson::doc!{"$project": {"_id": 0, "foo": {"b": "$_id.__unaliasedKey2"}, "__bot": {"foo_a": "$_id.foo_a", "foo_c": "$_id.foo_c"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "foo_a".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            }),
            OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into()
                }),
            ),
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "foo_c".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "c".into()
                }),
            })],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        duplicate_generated_alias,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"__unaliasedKey2": "$foo.a", "___unaliasedKey2": "$foo.b"}}},
                bson::doc!{"$project": {"_id": 0, "foo": {"b": "$_id.___unaliasedKey2"}, "__bot": {"__unaliasedKey2": "$_id.__unaliasedKey2"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "__unaliasedKey2".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            }),
            OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into()
                }),
            )],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        invalid_group_key_field_access,
        expected = Err(Error::InvalidGroupKey),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(
                FieldAccess {
                    expr: Box::new(Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(Literal::Integer(1))}
                    )),
                    field: "sub".to_string(),
                }
            ))],
            aggregations: vec![],
        }),
    );
    test_codegen_plan!(
        add_to_array,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$push": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlAvg": {"var": "$foo.a", "distinct": false}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlCount": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$first": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlLast": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$max": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlMergeObjects": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$min": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlStdDevPop": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlStdDevSamp": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlSum": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlCount": {"var": "$$ROOT", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: KEY_FOO_A.clone(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::CountStar(true)
            }],
        }),
    );
    test_codegen_plan!(
        agg_alias_underscore_id,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"__id": "$foo.a"}, "__id": {"$sqlSum": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"__id": "$_id.__id", "_id": "$__id"}}},
            ],
        }),
        input = Stage::Group(Group {
            source: SOURCE_IR.clone(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "__id".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into()
                }),
            })],
            aggregations: vec![AliasedAggregation {
                alias: "_id".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
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
        codegen::Error,
        ir::{binding_tuple::DatasourceName::Bottom, SubqueryModifier::*, *},
        map, unchecked_unique_linked_hash_map,
    };
    test_codegen_expr!(
        exists_uncorrelated,
        expected = Ok(bson::bson!(
            {"$subqueryExists": {
                "db": "test",
                "collection": "foo",
                "let": {},
                "pipeline": [{"$project": {"_id": 0,"foo": "$$ROOT"}},]
            }}
        )),
        input = Expression::Exists(Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
        }))),
    );
    test_codegen_expr!(
        exists_correlated,
        expected = Ok(bson::bson!(
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
        input = Expression::Exists(Box::new(Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "bar".into(),
            })),
            expression: map! {
                (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "a".into() => Expression::FieldAccess(FieldAccess{
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into()
                    })
                })
            }
        }))),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr
        },
    );
    test_codegen_expr!(
        subquery_expr_uncorrelated,
        expected = Ok(bson::bson!(
            {"$subquery": {
                "db": "test",
                "collection": "foo",
                "let": {},
                "outputPath": ["foo"],
                "pipeline": [{"$project": {"_id": 0,"foo": "$$ROOT"}},]
            }}
        )),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
            subquery: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
        }),
    );
    test_codegen_expr!(
        subquery_expr_correlated,
        expected = Ok(bson::bson!(
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
        input = Expression::Subquery(SubqueryExpr {
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
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into()
                        })
                    })
                }
            }))
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr
        },
    );
    test_codegen_expr!(
        subquery_expr_no_db_or_coll,
        expected = Ok(bson::bson!(
            {"$subquery": {"let": {},"outputPath": ["arr"],"pipeline": [
                {"$documents": []},
                {"$project": {"_id": 0, "arr": "$$ROOT"}},
            ]}}
        )),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Reference(("arr", 1u16).into())),
            subquery: Box::new(Stage::Array(Array {
                array: vec![],
                alias: "arr".into(),
            })),
        }),
    );
    test_codegen_expr!(
        output_expr_field_contains_dot,
        expected = Ok(bson::bson!(
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
        input = Expression::Subquery(SubqueryExpr {
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
        expected = Ok(bson::bson!(
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
        input = Expression::Subquery(SubqueryExpr {
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
        expected = Ok(bson::bson!(
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
        input = Expression::SubqueryComparison(SubqueryComparison {
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
        expected = Ok(bson::bson!(
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
        input = Expression::SubqueryComparison(SubqueryComparison {
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
                        (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::FieldAccess(FieldAccess{
                                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                                field: "a".into()
                            })
                        })
                    }
                }))
            }
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr.insert(("x", 0u16), "x");
            mr
        },
    );
    test_codegen_expr!(
        invalid_output_expr,
        expected = Err(Error::NoFieldPathForExpr),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::Literal(Literal::Integer(5))),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into()
                        })
                    })
                }
            }))
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr.insert(("x", 0u16), "x");
            mr
        },
    );
}
