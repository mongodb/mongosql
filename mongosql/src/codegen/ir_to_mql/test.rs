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
            use crate::codegen::{generate_mql_from_ir, ir_to_mql::MqlTranslation};

            let input = $input;
            let expected_db = $expected_db;
            let expected_collection = $expected_collection;
            let expected_pipeline = $expected_pipeline;

            let MqlTranslation {
                database: db,
                collection: col,
                mapping_registry: _,
                pipeline: pipeline,
            } = generate_mql_from_ir(input).expect("codegen failed");

            assert_eq!(expected_db, db);
            assert_eq!(expected_collection, col);
            assert_eq!(expected_pipeline, pipeline);
        }
    };

    ($func_name:ident, expected = Err($expected_err:expr), input = $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::codegen::generate_mql_from_ir;

            let input = $input;
            let expected = Err($expected_err);

            assert_eq!(expected, generate_mql_from_ir(input));
        }
    };
}

macro_rules! test_codegen_expr {
    ($func_name:ident, expected = $expected:expr, input = $input:expr, $(mapping_registry = $mapping_registry:expr,)?) => {
        #[test]
        fn $func_name() {
            use crate::codegen::ir_to_mql::{MqlCodeGenerator, MqlMappingRegistry};
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
    use crate::ir::{schema::SchemaCache, *};

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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
        }),
    );
}

mod array_stage {
    use crate::ir::{schema::SchemaCache, *};

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
        input = Stage::Array(ArraySource {
            array: vec![],
            alias: "arr".to_string(),
            cache: SchemaCache::new(),
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
        input = Stage::Array(ArraySource {
            array: vec![],
            alias: "_id".to_string(),
            cache: SchemaCache::new(),
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
        input = Stage::Array(ArraySource {
            array: vec![Expression::Literal(LiteralValue::Boolean(false).into())],
            alias: "arr".to_string(),
            cache: SchemaCache::new(),
        }),
    );
}

mod project {
    use crate::{
        codegen::ir_to_mql::Error,
        ir::{
            binding_tuple::{DatasourceName, Key},
            schema::SchemaCache,
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
                ("a", 0u16).into() => Expression::Literal(LiteralValue::Integer(1).into()),
                ("b", 0u16).into() => Expression::Literal(LiteralValue::Integer(2).into()),
                ("c", 0u16).into() => Expression::Literal(LiteralValue::Integer(3).into()),
            },
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{}.into())],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{}.into())],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{}.into())],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                ("_id", 0u16).into() => Expression::Literal(LiteralValue::Double(42.0).into()),
                ("foo", 0u16).into() => Expression::Literal(LiteralValue::Double(44.0).into()),
            },
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{}.into())],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                bson::doc! {
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "____bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$____bot"
                                }
                            }
                        }
                    }
                }
            ],
        }),
        input = Stage::Project(Project {
            expression: map! {
                Key{ datasource: DatasourceName::Bottom, scope: 0u16 } => Expression::Reference(("__bot", 0u16).into()),
                ("__bot", 0u16).into() => Expression::Literal(LiteralValue::Double(43.0).into()),
                ("___bot", 0u16).into() => Expression::Literal(LiteralValue::Double(44.0).into()),
                ("_____bot", 0u16).into() => Expression::Literal(LiteralValue::Double(45.0).into()),
            },
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{"a".into() => Expression::Literal(LiteralValue::Integer(42).into())}.into())],
                alias: "__bot".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
        }),
    );

    test_codegen_plan!(
        bot_test,
        expected = Ok({
            database: None,
            collection: None,
            pipeline: vec![
                bson::doc!{"$documents": [ {"$literal": {}}]},
                bson::doc!{"$project": {"_id": 0, "__bot": "$$ROOT"}},
                bson::doc!{"$project": {"_id": 0, "__bot": "$__bot", "a": {"$literal": 1}}},
                bson::doc!{"$replaceWith": {
                    "$unsetField": {
                        "field": "__bot", "input": {
                            "$setField": {
                                "input": "$$ROOT", "field": "", "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Project(Project {
            expression: map! {
                Key{ datasource: DatasourceName::Bottom, scope: 0u16 } => Expression::Reference(("__bot", 0u16).into()),
                ("a", 0u16).into() => Expression::Literal(LiteralValue::Integer(1).into()),
            },
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(unchecked_unique_linked_hash_map!{}.into())],
                alias: "__bot".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
        }),
    );

    test_codegen_plan!(
        dot_project_field,
        expected = Err(Error::InvalidProjectField),
        input = Stage::Project(Project {
            expression: map! {
                ("a.b", 0u16).into() => Expression::Literal(LiteralValue::Integer(1).into()),
            },
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {}.into()
                )],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
        }),
    );

    test_codegen_plan!(
        dollar_project_field,
        expected = Err(Error::InvalidProjectField),
        input = Stage::Project(Project {
            expression: map! {
                ("$a", 0u16).into() => Expression::Literal(LiteralValue::Integer(1).into()),
            },
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {}.into()
                )],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        empty_project_field,
        expected = Err(Error::InvalidProjectField),
        input = Stage::Project(Project {
            expression: map! {
                ("", 0u16).into() => Expression::Literal(LiteralValue::Integer(1).into()),
            },
            source: Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {}.into()
                )],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod filter {
    use crate::ir::{schema::SchemaCache, *};

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
            condition: Expression::Literal(LiteralValue::Boolean(true).into()),
            source: Stage::Array(ArraySource {
                array: vec![],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod sort {
    use crate::{
        codegen::ir_to_mql::Error,
        ir::{schema::SchemaCache, Expression::Reference, SortSpecification::*, *},
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
            source: Stage::Array(ArraySource {
                array: vec![],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                        cache: SchemaCache::new(),
                    }).into(),
                ),
            ],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                    cache: SchemaCache::new(),
                }).into(),
            )],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        other_expr,
        expected = Err(Error::InvalidSortKey),
        input = Stage::Sort(Sort {
            specs: vec![Asc(
                Expression::Literal(LiteralValue::Integer(1).into()).into()
            )],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        non_ident_field_reference,
        expected = Err(Error::InvalidSortKey),
        input = Stage::Sort(Sort {
            specs: vec![Asc(Expression::FieldAccess(FieldAccess {
                expr: Expression::Document(
                    unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                .into())
                .into(),
                field: "sub".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),)],
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod limit_offset {
    use crate::{
        codegen::ir_to_mql::Error,
        ir::{schema::SchemaCache, *},
    };

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
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod literal {
    use crate::ir::{Expression::*, LiteralValue::*};
    use bson::{bson, Bson};

    test_codegen_expr!(
        null,
        expected = Ok(bson!({ "$literal": Bson::Null })),
        input = Literal(Null.into()),
    );
    test_codegen_expr!(
        bool,
        expected = Ok(bson!({"$literal": true})),
        input = Literal(Boolean(true).into()),
    );
    test_codegen_expr!(
        string,
        expected = Ok(bson!({"$literal": "abc"})),
        input = Literal(String("abc".into()).into()),
    );
    test_codegen_expr!(
        int,
        expected = Ok(bson!({"$literal": 5_i32})),
        input = Literal(Integer(5).into()),
    );
    test_codegen_expr!(
        long,
        expected = Ok(bson!({"$literal": 6_i64})),
        input = Literal(Long(6).into()),
    );
    test_codegen_expr!(
        double,
        expected = Ok(bson!({"$literal": 7.0})),
        input = Literal(Double(7.0).into()),
    );
}

mod reference {
    use crate::{codegen::ir_to_mql::Error, ir::Expression::*};
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
    use crate::ir::{Expression::*, LiteralValue};
    use bson::bson;

    test_codegen_expr!(
        empty,
        expected = Ok(bson!([])),
        input = Array(vec![].into()),
    );
    test_codegen_expr!(
        non_empty,
        expected = Ok(bson!([{"$literal": "abc"}])),
        input = Array(vec![Literal(LiteralValue::String("abc".into()).into())].into()),
    );
    test_codegen_expr!(
        nested,
        expected = Ok(bson!([{ "$literal": null }, [{ "$literal": null }]])),
        input = Array(
            vec![
                Literal(LiteralValue::Null.into()),
                Array(vec![Literal(LiteralValue::Null.into())].into())
            ]
            .into()
        ),
    );
}

mod document {
    use crate::{
        ir::{Expression::*, LiteralValue},
        unchecked_unique_linked_hash_map,
    };
    use bson::bson;

    test_codegen_expr!(
        empty,
        expected = Ok(bson!({"$literal": {}})),
        input = Document(unchecked_unique_linked_hash_map! {}.into()),
    );
    test_codegen_expr!(
        non_empty,
        expected = Ok(bson!({"foo": {"$literal": 1}})),
        input = Document(
            unchecked_unique_linked_hash_map! {"foo".to_string() => Literal(LiteralValue::Integer(1).into()),}
        .into()),
    );
    test_codegen_expr!(
        nested,
        expected = Ok(bson!({"foo": {"$literal": 1}, "bar": {"baz": {"$literal": 2}}})),
        input = Document(
            unchecked_unique_linked_hash_map! {
                "foo".to_string() => Literal(LiteralValue::Integer(1).into()),
                "bar".to_string() => Document(unchecked_unique_linked_hash_map!{
                    "baz".to_string() => Literal(LiteralValue::Integer(2).into())
                }.into()),
            }
            .into()
        ),
    );
    test_codegen_expr!(
        dollar_prefixed_key_allowed,
        expected = Ok(bson!({"$getField": {"field": {"$literal": "$foo"}, "input": {"$literal":  1}}})),
        input = Document(
            unchecked_unique_linked_hash_map! {"$foo".to_string() => Literal(LiteralValue::Integer(1).into()),}
        .into()),
    );
    test_codegen_expr!(
        key_containing_dot_allowed,
        expected = Ok(bson!({"$getField": {"field": {"$literal": "foo.bar"}, "input": {"$literal":  1}}})),
        input = Document(
            unchecked_unique_linked_hash_map! {"foo.bar".to_string() => Literal(LiteralValue::Integer(1).into()),}
        .into()),
    );
    test_codegen_expr!(
        empty_key_allowed,
        expected = Ok(bson!({"$getField": {"field": {"$literal": ""}, "input": {"$literal":  1}}})),
        input = Document(
            unchecked_unique_linked_hash_map! {"".to_string() => Literal(LiteralValue::Integer(1).into()),}
        .into()),
    );
}

mod field_access {
    use crate::{
        ir::{schema::SchemaCache, *},
        unchecked_unique_linked_hash_map,
    };
    use bson::Bson;

    test_codegen_expr!(
        reference,
        expected = Ok(Bson::String("$f.sub".to_string())),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "sub".to_string(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            })
            .into(),
            cache: SchemaCache::new(),
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
                unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
            .into())
            .into(),
            field: "sub".to_string(),
            cache: SchemaCache::new(),
        }),
    );

    test_codegen_expr!(
        success_on_non_reference_expr,
        expected = Ok(bson::bson!({"$getField": {
            "field": "sub",
            "input": {"$literal": "f"},
        }})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Literal(LiteralValue::String("f".into()).into()).into(),
            field: "sub".to_string(),
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_expr!(
        dollar_prefixed_field,
        expected = Ok(bson::bson!({"$getField": {"field": {"$literal":"$sub"}, "input": "$f"}})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "$sub".to_string(),
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        field_contains_dot,
        expected = Ok(bson::bson!({"$getField": {"field": {"$literal": "s.ub"}, "input": "$f"}})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "s.ub".to_string(),
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        empty_field_in_field_access,
        expected = Ok(bson::bson!({"$getField": {"field": {"$literal": ""}, "input": "$f"}})),
        input = Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            field: "".to_string(),
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
}

mod searched_case_expression {
    use crate::ir::{schema::SchemaCache, *};
    test_codegen_expr!(
        one_case,
        expected = Ok(bson::bson!({"$switch":
            {"branches": [{"case": {"$literal": true},
                                            "then": {"$literal": "first case"}}],
            "default": {"$literal": "else case"}}})),
        input = Expression::SearchedCase(SearchedCaseExpr {
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
                then: Box::new(Expression::Literal(
                    LiteralValue::String("first case".to_string()).into()
                ))
            }],
            else_branch: Box::new(Expression::Literal(
                LiteralValue::String("else case".to_string()).into()
            )),
            cache: SchemaCache::new(),
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
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(false).into())),
                    then: Box::new(Expression::Literal(
                        LiteralValue::String("first case".to_string()).into()
                    ))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
                    then: Box::new(Expression::Literal(
                        LiteralValue::String("second case".to_string()).into()
                    ))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
                    then: Box::new(Expression::Literal(
                        LiteralValue::String("third case".to_string()).into()
                    ))
                }
            ],
            else_branch: Box::new(Expression::Literal(
                LiteralValue::String("else case".to_string()).into()
            )),
            cache: SchemaCache::new(),
        }),
    );
}

mod simple_case_expression {
    use crate::ir::{schema::SchemaCache, *};
    test_codegen_expr!(
        one_case,
        expected = Ok(
            bson::bson!({"$let": {"vars": {"target": {"$literal": "co"}}, "in": {"$switch":
            {"branches": [{"case": {"$sqlEq": ["$$target", {"$literal": true}]},
                                            "then": {"$literal": "true case"}}],
            "default": {"$literal": "else case"}}}}})
        ),
        input = Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(
                LiteralValue::String("co".to_string()).into()
            )),
            when_branch: vec![WhenBranch {
                when: Box::new(Expression::Literal(LiteralValue::Boolean(true).into())),
                then: Box::new(Expression::Literal(
                    LiteralValue::String("true case".to_string()).into()
                ))
            }],
            else_branch: Box::new(Expression::Literal(
                LiteralValue::String("else case".to_string()).into()
            )),
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_expr!(
        multiple_cases,
        expected = Ok(
            bson::bson!({"$let": {"vars": {"target": {"$literal": "co"}}, "in": {"$switch":
            {"branches": [{"case": {"$sqlEq": ["$$target", {"$literal": false}]},
                                            "then": {"$literal": "first case"}},
            {"case": {"$sqlEq": ["$$target", {"$literal": false}]},
                                            "then": {"$literal": "second case"}},
            {"case": {"$sqlEq": ["$$target", {"$literal": false}]},
                                            "then": {"$literal": "third case"}}],
            "default": {"$literal": "else case"}}}}})
        ),
        input = Expression::SimpleCase(SimpleCaseExpr {
            expr: Box::new(Expression::Literal(
                LiteralValue::String("co".to_string()).into()
            )),
            when_branch: vec![
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(false).into())),
                    then: Box::new(Expression::Literal(
                        LiteralValue::String("first case".to_string()).into()
                    ))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(false).into())),
                    then: Box::new(Expression::Literal(
                        LiteralValue::String("second case".to_string()).into()
                    ))
                },
                WhenBranch {
                    when: Box::new(Expression::Literal(LiteralValue::Boolean(false).into())),
                    then: Box::new(Expression::Literal(
                        LiteralValue::String("third case".to_string()).into()
                    ))
                }
            ],
            else_branch: Box::new(Expression::Literal(
                LiteralValue::String("else case".to_string()).into()
            )),
            cache: SchemaCache::new(),
        }),
    );
}

mod join {
    use crate::ir::{schema::SchemaCache, *};

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
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col2".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            join_type: JoinType::Inner,
            cache: SchemaCache::new(),
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
            condition: Some(Expression::Literal(LiteralValue::Boolean(true).into())),
            left: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb2".to_string(),
                collection: "col2".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            join_type: JoinType::Left,
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Array(ArraySource {
                array: vec![Expression::Literal(LiteralValue::Integer(1).into()),Expression::Literal(LiteralValue::Integer(1).into())],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            join_type: JoinType::Left,
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb2".to_string(),
                collection: "col2".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            join_type: JoinType::Left,
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb2".to_string(),
                collection: "col2".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            join_type: JoinType::Left,
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        let_binding_name_conflict_appends_underscores_for_uniqueness,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("Foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id" : 0, "Foo": "$$ROOT"}},
                bson::doc!{
                    "$join": {
                        "collection": "foo",
                        "joinType": "inner",
                        "pipeline": [{"$project": {"_id": 0, "foo": "$$ROOT"}}],
                    }
                },
                bson::doc!{
                    "$join": {
                        "collection": "bar",
                        "joinType": "inner",
                        "let": {"foo_0": "$Foo", "foo_0_": "$foo"},
                        "pipeline": [{"$project": {"_id": 0, "bar": "$$ROOT"}}],
                        "condition": {"$match": {"$expr": {"$literal": true}}}
                    }
                },
            ],
        }),
        input = Stage::Join(Join {
            condition: Some(Expression::Literal(LiteralValue::Boolean(true).into())),
            left: Stage::Join(Join {
                condition: None,
                left: Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "Foo".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "foo".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                join_type: JoinType::Inner,
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "bar".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            join_type: JoinType::Inner,
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        datasource_name_is_normalized_when_turned_into_var_binding,
        expected = Ok({
            database: Some("mydb".to_string()),
            collection: Some("Foo coll-ß".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id" : 0, "Foo coll-ß": "$$ROOT"}},
                bson::doc!{
                    "$join": {
                        "collection": "foo_coll_ß",
                        "joinType": "inner",
                        "pipeline": [{"$project": {"_id": 0, "foo_coll_ß": "$$ROOT"}}],
                    }
                },
                bson::doc!{
                    "$join": {
                        "collection": "bar",
                        "joinType": "inner",
                        "let": {"foo_coll_ß_0": "$Foo coll-ß", "foo_coll_ß_0_": "$foo_coll_ß"},
                        "pipeline": [{"$project": {"_id": 0, "bar": "$$ROOT"}}],
                        "condition": {"$match": {"$expr": {"$literal": true}}}
                    }
                },
            ],
        }),
        input = Stage::Join(Join {
            condition: Some(Expression::Literal(LiteralValue::Boolean(true).into())),
            left: Stage::Join(Join {
                condition: None,
                left: Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "Foo coll-ß".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Collection(Collection {
                    db: "mydb".to_string(),
                    collection: "foo_coll_ß".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                join_type: JoinType::Inner,
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "bar".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            join_type: JoinType::Inner,
            cache: SchemaCache::new(),
        }),
    );
}

mod union {
    use crate::ir::{schema::SchemaCache, *};

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
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "bar".to_string(),
                collection: "b".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
            left: Stage::Array(ArraySource {
                array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
                alias: "arr1".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Array(ArraySource {
                array: vec![Expression::Literal(LiteralValue::Integer(2).into())],
                alias: "arr2".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
            left: Stage::Array(ArraySource {
                array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Collection(Collection {
                db: "bar".to_string(),
                collection: "b".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Array(ArraySource {
                array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
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
                cache: SchemaCache::new(),
            }).into(),
            right: Stage::Set(Set {
                operation: SetOperation::UnionAll,
                left: Stage::Collection(Collection {
                    db: "bar".to_string(),
                    collection: "b".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                right: Stage::Array(ArraySource {
                    array: vec![Expression::Literal(LiteralValue::Integer(1).into())],
                    alias: "arr".to_string(),
                    cache: SchemaCache::new(),
                }).into(),
                cache: SchemaCache::new(),
            }).into(),
            cache: SchemaCache::new(),
        }),
    );
}

mod function {
    use crate::{
        codegen::ir_to_mql::Error,
        ir::{
            definitions::*, schema::SchemaCache, DateFunction, Expression::*, LiteralValue,
            ScalarFunction::*,
        },
    };
    use bson::bson;

    test_codegen_expr!(
        concat_expr,
        expected = Ok(bson::bson! ({"$concat": ["$f", {"$literal": "bar"}]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Concat,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(LiteralValue::String("bar".to_string()).into()),
            ],
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        is_null_expr,
        expected = Ok(bson!({"$or": [{"$eq": [{"$type": "$f"}, "null"]},
                                {"$eq": [{"$type": "$f"}, "missing"]}]})),
        input = Expression::Is(IsExpr {
            expr: Expression::Reference(("f", 0u16).into()).into(),
            target_type: TypeOrMissing::Type(Type::Null),
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
        expected = Ok(
            bson::bson! ({"$sqlDivide": {"dividend": "$f", "divisor": {"$literal": 1}, "onError": {"$literal": bson::Bson::Null}}})
        ),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Div,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(10).into())
            ],
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Boolean(true).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        position_expr,
        expected = Ok(bson!({"$sqlIndexOfCP": [ "$f", {"$literal": "a"}]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Position,
            args: vec![
                Expression::Literal(LiteralValue::String("a".to_string()).into()),
                Expression::Reference(("f", 0u16).into()),
            ],
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::Integer(1).into()),
                Expression::Literal(LiteralValue::Integer(10).into()),
            ],
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_expr!(
        ltrim_expr,
        expected = Ok(bson!({"$ltrim": {"input": "$f", "chars": {"$literal": "a"}}})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: LTrim,
            args: vec![
                Expression::Literal(LiteralValue::String("a".to_string()).into()),
                Expression::Reference(("f", 0u16).into()),
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::String("a".to_string()).into()),
                Expression::Reference(("f", 0u16).into()),
            ],
            cache: SchemaCache::new(),
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
                Expression::Literal(LiteralValue::String("a".to_string()).into()),
                Expression::Reference(("f", 0u16).into()),
            ],
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_day_expr,
        expected = Ok(bson!({"$dayOfMonth": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Day,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
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
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_week_expr,
        expected = Ok(bson!({"$week": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Week,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_day_of_year_expr,
        expected = Ok(bson!({"$dayOfYear": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: DayOfYear,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_iso_week_expr,
        expected = Ok(bson!({"$isoWeek": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: IsoWeek,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        extract_weekday_expr,
        expected = Ok(bson!({"$isoDayOfWeek": "$f"})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: IsoWeekday,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        dateadd,
        expected = Ok(
            bson!({"$dateAdd": {"startDate": "$$NOW", "unit": {"$literal": "year"}, "amount": {"$literal": 5}}})
        ),
        input = DateFunction(DateFunctionApplication {
            function: DateFunction::Add,
            date_part: DatePart::Year,
            args: vec![
                Expression::Literal(LiteralExpr {
                    value: LiteralValue::Integer(5),
                    cache: SchemaCache::new()
                }),
                ScalarFunction(ScalarFunctionApplication {
                    function: CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
            ],
            cache: SchemaCache::new()
        }),
    );
    test_codegen_expr!(
        datediff,
        expected = Ok(
            bson!({"$dateDiff": {"startDate": "$$NOW", "endDate": "$$NOW", "unit": {"$literal": "year"}, "startOfWeek": {"$literal": "sunday"}}})
        ),
        input = DateFunction(DateFunctionApplication {
            function: DateFunction::Diff,
            date_part: DatePart::Year,
            args: vec![
                ScalarFunction(ScalarFunctionApplication {
                    function: CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
                ScalarFunction(ScalarFunctionApplication {
                    function: CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
                Expression::Literal(LiteralExpr {
                    value: LiteralValue::String("sunday".to_string()),
                    cache: SchemaCache::new()
                }),
            ],
            cache: SchemaCache::new()
        }),
    );
    test_codegen_expr!(
        datetrunc,
        expected = Ok(
            bson!({"$dateTrunc": {"date": "$$NOW", "unit": {"$literal": "year"}, "startOfWeek": {"$literal": "sunday"}}})
        ),
        input = DateFunction(DateFunctionApplication {
            function: DateFunction::Trunc,
            date_part: DatePart::Year,
            args: vec![
                ScalarFunction(ScalarFunctionApplication {
                    function: CurrentTimestamp,
                    args: vec![],
                    cache: SchemaCache::new(),
                }),
                Expression::Literal(LiteralExpr {
                    value: LiteralValue::String("sunday".to_string()),
                    cache: SchemaCache::new()
                }),
            ],
            cache: SchemaCache::new()
        }),
    );
    test_codegen_expr!(
        computedfieldaccess_expr,
        expected = Err(Error::UnsupportedFunction(ComputedFieldAccess)),
        input = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ComputedFieldAccess,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(LiteralValue::Long(42).into()),
            ],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );

    test_codegen_expr!(
        abs_expr,
        expected = Ok(bson!({"$abs": "$f" })),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Abs,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new()
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        ceil_expr,
        expected = Ok(bson!({"$ceil": "$f" })),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Ceil,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new()
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        degrees_expr,
        expected = Ok(bson!({"$radiansToDegrees": "$f" })),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Degrees,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new()
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        floor_expr,
        expected = Ok(bson!({"$floor": "$f" })),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Floor,
            args: vec![Expression::Reference(("f", 0u16).into()),],
            cache: SchemaCache::new()
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        log_expr,
        expected = Ok(bson!({"$sqlLog": [
            "$f1", "$f2"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Log,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into())
            ],
            cache: SchemaCache::new()
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        mod_expr,
        expected = Ok(bson!({"$sqlMod": [
            "$f1", "$f2"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Mod,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into())
            ],
            cache: SchemaCache::new()
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        pow_expr,
        expected = Ok(bson!({"$pow": [
            "$f1", "$f2"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Pow,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into())
            ],
            cache: SchemaCache::new()
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        round_expr,
        expected = Ok(bson!({"$sqlRound": [
            "$f1", "$f2"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Round,
            args: vec![
                Expression::Reference(("f1", 0u16).into()),
                Expression::Reference(("f2", 0u16).into())
            ],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f1", 0u16), "f1");
            mr.insert(("f2", 0u16), "f2");
            mr
        },
    );
    test_codegen_expr!(
        split_expr,
        expected = Ok(bson::bson! ({"$sqlSplit": [
            "$f", {"$literal": "a"}, {"$literal": 1}
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Split,
            args: vec![
                Expression::Reference(("f", 0u16).into()),
                Expression::Literal(LiteralValue::String("a".to_string()).into()),
                Expression::Literal(LiteralValue::Integer(1).into())
            ],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        cos_expr,
        expected = Ok(bson!({"$sqlCos": [
            "$f"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Cos,
            args: vec![Expression::Reference(("f", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        sin_expr,
        expected = Ok(bson!({"$sqlSin": [
            "$f"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Sin,
            args: vec![Expression::Reference(("f", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        tan_expr,
        expected = Ok(bson!({"$sqlTan": [
            "$f"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Tan,
            args: vec![Expression::Reference(("f", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        radians_expr,
        expected = Ok(bson!({"$degreesToRadians":
            "$f"
        })),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Radians,
            args: vec![Expression::Reference(("f", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
    test_codegen_expr!(
        sqrt_expr,
        expected = Ok(bson!({"$sqlSqrt": [
            "$f"
        ]})),
        input = ScalarFunction(ScalarFunctionApplication {
            function: Sqrt,
            args: vec![Expression::Reference(("f", 0u16).into())],
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
    );
}

mod group_by {
    use crate::{
        codegen::ir_to_mql::Error, ir::definitions::*, ir::schema::SchemaCache, map,
        unchecked_unique_linked_hash_map,
    };

    fn source_ir() -> Box<Stage> {
        Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        }))
    }

    fn key_foo_a() -> Vec<OptionallyAliasedExpr> {
        vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
            alias: "foo_a".to_string(),
            expr: Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            }),
        })]
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "f".into(),
                expr: Expression::Reference(("foo", 0u16).into()),
            })],
            aggregations: vec![],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "_id".into(),
                expr: Expression::Reference(("foo", 0u16).into()),
            })],
            aggregations: vec![],
            cache: SchemaCache::new(),
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
            source: source_ir(),
            keys: vec![OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(
                    FieldAccess{
                        expr: Expression::Reference(("foo", 0u16).into()).into(),
                        field: "_id".into(),
                        cache: SchemaCache::new(),
                    }),
            )],
            aggregations: vec![],
            cache: SchemaCache::new(),
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
                source: source_ir(),
                expression: map! {
                    ("_id", 0u16).into() =>
                        Expression::Reference(("foo", 0u16).into())
                },
                cache: SchemaCache::new(),
            },)),
            keys: vec![OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(
                    FieldAccess{
                        expr: Expression::Reference(("_id", 0u16).into()).into(),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }
                )
            )],
            aggregations: vec![],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![],
            cache: SchemaCache::new(),
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
            source: source_ir(),
            keys: vec![OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into(),
                    cache: SchemaCache::new(),
                }),
            ), OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into(),
                    cache: SchemaCache::new(),
                }),
            )],
            aggregations: vec![],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "foo_a".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into(),
                    cache: SchemaCache::new(),
                }),
            }),
            OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into(),
                        cache: SchemaCache::new(),
                }),
            ),
            OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "foo_c".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "c".into(),
                        cache: SchemaCache::new(),
                }),
            })],
            aggregations: vec![],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: vec![
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "__unaliasedKey2".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into(),
                        cache: SchemaCache::new(),
                }),
            }),
            OptionallyAliasedExpr::Unaliased(
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "b".into(),
                        cache: SchemaCache::new(),
                }),
            )],
            aggregations: vec![],
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        invalid_group_key_field_access,
        expected = Err(Error::InvalidGroupKey),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: vec![OptionallyAliasedExpr::Unaliased(Expression::FieldAccess(
                FieldAccess {
                    expr: Box::new(Expression::Document(
                        unchecked_unique_linked_hash_map! {"a".into() => Expression::Literal(LiteralValue::Integer(1).into())}
                    .into())),
                    field: "sub".to_string(),
                    cache: SchemaCache::new(),
                }
            ))],
            aggregations: vec![],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::AddToArray,
                    distinct: false,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        add_to_set,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$addToSet": "$foo.a"}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::AddToArray,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Avg,
                    distinct: false,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Count,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::First,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_plan!(
        last,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$group": {"_id": {"foo_a": "$foo.a"}, "agg": {"$sqlLast": {"var": "$foo.a", "distinct": true}}}},
                bson::doc!{"$project": {"_id": 0, "__bot": {"foo_a": "$_id.foo_a", "agg": "$agg"}}},
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Last,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Max,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::MergeDocuments,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Min,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::StddevPop,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::StddevSamp,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: key_foo_a(),
            aggregations: vec![AliasedAggregation {
                alias: "agg".into(),
                agg_expr: AggregationExpr::CountStar(true)
            }],
            cache: SchemaCache::new(),
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
                bson::doc!{
                    "$replaceWith": {
                        "$unsetField": {
                            "field": "__bot",
                            "input": {
                                "$setField": {
                                    "input": "$$ROOT",
                                    "field": "",
                                    "value": "$__bot"
                                }
                            }
                        }
                    }
                },
            ],
        }),
        input = Stage::Group(Group {
            source: source_ir(),
            keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: "__id".to_string(),
                expr: Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    field: "a".into(),
                    cache: SchemaCache::new(),
                }),
            })],
            aggregations: vec![AliasedAggregation {
                alias: "_id".into(),
                agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                    function: AggregationFunction::Sum,
                    distinct: true,
                    arg: Box::new(Expression::FieldAccess(FieldAccess {
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    }))
                })
            }],
            cache: SchemaCache::new(),
        }),
    );
}

mod subquery {
    use crate::{
        codegen::ir_to_mql::Error,
        ir::{binding_tuple::DatasourceName::Bottom, schema::SchemaCache, SubqueryModifier::*, *},
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
        input = Expression::Exists(
            Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            }))
            .into()
        ),
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
                cache: SchemaCache::new(),
            })),
            expression: map! {
                (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                    "a".into() => Expression::FieldAccess(FieldAccess{
                        expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        field: "a".into(),
                        cache: SchemaCache::new(),
                    })
                }.into())
            },
            cache: SchemaCache::new(),
        })).into()),
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
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
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
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
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
            subquery: Box::new(Stage::Array(ArraySource {
                array: vec![],
                alias: "arr".into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
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
                field: "bar.a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    ("foo", 1u16).into() => Expression::Reference(("foo", 1u16).into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
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
                    ("__bot", 1u16).into() => Expression::Literal(LiteralValue::Double(43.0).into()),
                },
                source: Stage::Collection(Collection {
                    db: "test".to_string(),
                    collection: "__bot".to_string(),
                    cache: SchemaCache::new(),
                })
                .into(),
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
    );
    test_codegen_expr!(
        let_binding_name_conflict_appends_underscores_for_uniqueness,
        expected = Ok(bson::bson!(
            {"$subquery": {
                "db": "test",
                "collection": "bar",
                "let": {"foo_0": "$Foo", "foo_0_": "$foo"},
                "outputPath": ["__bot", "a"],
                "pipeline": [
                    {"$project": {"_id": 0, "bar": "$$ROOT"}},
                    {"$project": {"_id": 0, "__bot": {"a": {"$sqlEq": ["$$foo_0.a", "$$foo_0_.a"]}}}}
                ]
            }}
        )),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Eq,
                            args: vec![
                                Expression::FieldAccess(FieldAccess{
                                    expr: Box::new(Expression::Reference(("Foo", 0u16).into())),
                                    field: "a".into(),
                                    cache: SchemaCache::new(),
                                }),
                                Expression::FieldAccess(FieldAccess{
                                    expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                                    field: "a".into(),
                                    cache: SchemaCache::new(),
                                }),
                            ],
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("Foo", 0u16), "Foo");
            mr.insert(("foo", 0u16), "foo");
            mr
        },
    );
    test_codegen_expr!(
        datasource_name_is_normalized_when_turned_into_var_binding,
        expected = Ok(bson::bson!(
            {"$subquery": {
                "db": "test",
                "collection": "bar",
                "let": {"foo_coll_ß_0": "$Foo coll-ß", "foo_coll_ß_0_": "$foo_coll_ß"},
                "outputPath": ["__bot", "a"],
                "pipeline": [
                    {"$project": {"_id": 0, "bar": "$$ROOT"}},
                    {"$project": {"_id": 0, "__bot": {"a": {"$sqlEq": ["$$foo_coll_ß_0.a", "$$foo_coll_ß_0_.a"]}}}}
                ]
            }}
        )),
        input = Expression::Subquery(SubqueryExpr {
            output_expr: Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference((Bottom, 1u16).into())),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::ScalarFunction(ScalarFunctionApplication {
                            function: ScalarFunction::Eq,
                            args: vec![
                                Expression::FieldAccess(FieldAccess{
                                    expr: Box::new(Expression::Reference(("Foo coll-ß", 0u16).into())),
                                    field: "a".into(),
                                    cache: SchemaCache::new(),
                                }),
                                Expression::FieldAccess(FieldAccess{
                                    expr: Box::new(Expression::Reference(("foo_coll_ß", 0u16).into())),
                                    field: "a".into(),
                                    cache: SchemaCache::new(),
                                }),
                            ],
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("Foo coll-ß", 0u16), "Foo coll-ß");
            mr.insert(("foo_coll_ß", 0u16), "foo_coll_ß");
            mr
        },
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
            argument: Box::new(Expression::Literal(LiteralValue::Integer(5).into())),
            subquery_expr: SubqueryExpr {
                output_expr: Box::new(Expression::Reference(("foo", 1u16).into())),
                subquery: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new(),
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
                    field: "a".into(),
                    cache: SchemaCache::new(),
                })),
                subquery: Box::new(Stage::Project(Project {
                    source: Box::new(Stage::Collection(Collection {
                        db: "test".into(),
                        collection: "bar".into(),
                        cache: SchemaCache::new(),
                    })),
                    expression: map! {
                        (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                            "a".into() => Expression::FieldAccess(FieldAccess{
                                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                                field: "a".into(),
                                cache: SchemaCache::new(),
                            })
                        }.into())
                    },
                    cache: SchemaCache::new(),
                })),
                cache: SchemaCache::new(),
            },
            cache: SchemaCache::new(),
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
            output_expr: Box::new(Expression::Literal(LiteralValue::Integer(5).into())),
            subquery: Box::new(Stage::Project(Project {
                source: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                expression: map! {
                    (Bottom, 1u16).into() => Expression::Document(unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::FieldAccess(FieldAccess{
                            expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                            field: "a".into(),
                            cache: SchemaCache::new(),
                        })
                    }.into())
                },
                cache: SchemaCache::new(),
            })),
            cache: SchemaCache::new(),
        }),
        mapping_registry = {
            let mut mr = MqlMappingRegistry::default();
            mr.insert(("foo", 0u16), "foo");
            mr.insert(("x", 0u16), "x");
            mr
        },
    );
}

mod unwind {
    use crate::{
        codegen::ir_to_mql::Error,
        ir::{schema::SchemaCache, *},
        unchecked_unique_linked_hash_map,
    };

    fn make_unwind(path: Box<Expression>, index: Option<String>, outer: bool) -> Stage {
        Stage::Unwind(Unwind {
            source: Box::new(Stage::Collection(Collection {
                db: "test".to_string(),
                collection: "foo".to_string(),
                cache: SchemaCache::new(),
            })),
            path,
            index,
            outer,
            cache: SchemaCache::new(),
        })
    }

    test_codegen_plan!(
        simple,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$unwind": {"path": "$foo.arr"}},
            ],
        }),
        input = make_unwind(
            Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "arr".into(),
                cache: SchemaCache::new(),
            })),
            None,
            false,
        ),
    );

    test_codegen_plan!(
        all_opts,
        expected = Ok({
            database: Some("test".to_string()),
            collection: Some("foo".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "foo": "$$ROOT"}},
                bson::doc!{"$unwind": {"path": "$foo.arr", "includeArrayIndex": "foo.idx", "preserveNullAndEmptyArrays": true}},
            ],
        }),
        input = make_unwind(
            Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "arr".into(),
                cache: SchemaCache::new(),
            })),
            Some("idx".into()),
            true,
        ),
    );

    test_codegen_plan!(
        invalid,
        expected = Err(Error::InvalidUnwindPath),
        input = make_unwind(
            Box::new(Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(42).into())
                    }
                    .into()
                )),
                field: "a".into(),
                cache: SchemaCache::new(),
            })),
            Some("idx".into()),
            false,
        ),
    );
}
