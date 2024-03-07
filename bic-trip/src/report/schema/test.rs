use mongosql::{
    map,
    schema::{Atomic, Document, Schema},
    set,
};

use crate::schema::{process_schema, process_schemata, DatabaseAnalysis, SchemaAnalysis};

use super::CollectionAnalysis;

macro_rules! test_process_schema {
    ($name:ident, schema=$schema:expr, expected=$expected:expr) => {
        #[test]
        fn $name() {
            let mut analysis = CollectionAnalysis {
                collection_name: "test".to_string(),
                ..Default::default()
            };

            process_schema("test".to_string(), $schema, &mut analysis, 0);

            assert_eq!(analysis, $expected);
        }
    };
}

macro_rules! test_process_schemata {
    ($name:ident, schemata=$schemata:expr, expected=$expected:expr) => {
        #[test]
        fn $name() {
            let analysis = process_schemata($schemata);

            assert_eq!(analysis, $expected);
        }
    };
}

test_process_schema!(
    basic_analysis_simple_document_macro,
    schema = Schema::Document(Document {
        keys: map!(
            "a".into() => Schema::Atomic(Atomic::String),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
        ),
        ..Default::default()
    }),
    expected = CollectionAnalysis {
        collection_name: "test".to_string(),
        arrays: map! { "test.c".to_string() => 1 },
        ..Default::default()
    }
);

test_process_schema!(
    basic_analysis_one_embedded_object,
    schema = Schema::Document(Document {
        keys: map!(
            "a".into() => Schema::Atomic(Atomic::String),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Document(Document {
                keys: map!(
                    "d".into() => Schema::Atomic(Atomic::String),
                    "e".into() => Schema::Atomic(Atomic::Integer),
                ),
                required: set!("d".to_string()),
                ..Default::default()
            }),
        ),
        ..Default::default()
    }),
    expected = CollectionAnalysis {
        collection_name: "test".to_string(),
        documents: map! { "test.c".to_string() => 1 },
        ..Default::default()
    }
);

test_process_schema!(
    analysis_array_is_any_of,
    schema = Schema::Document(Document {
        keys: map!(
            "a".into() => Schema::Atomic(Atomic::String),
            "b".into() => Schema::Array(Box::new(Schema::AnyOf(set! {
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::Integer),
            }))),
        ),
        ..Default::default()
    }),
    expected = CollectionAnalysis {
        collection_name: "test".to_string(),
        arrays: map! { "test.b".to_string() => 1 },
        anyof: map! { "test.b".to_string() => 1 },
        ..Default::default()
    }
);
test_process_schema!(
    analysis_any_if_found_as_unstable,
    schema = Schema::Document(Document {
        keys: map!(
            "a".into() => Schema::Atomic(Atomic::String),
            "b".into() => Schema::Any,
        ),
        ..Default::default()
    }),
    expected = CollectionAnalysis {
        collection_name: "test".to_string(),
        unstable: map! { "test.b".to_string() => 1 },
        ..Default::default()
    }
);

test_process_schema!(
    analysis_all_together_and_nested,
    schema = Schema::Document(Document {
        keys: map!(
            "a".into() => Schema::Atomic(Atomic::String),
            "b".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            "c".into() => Schema::Document(Document {
                keys: map!(
                    "d".into() => Schema::Atomic(Atomic::String),
                    "e".into() => Schema::Array(Box::new(Schema::Document(Document {
                        keys: map!(
                            "f".into() => Schema::Atomic(Atomic::String),
                            "g".into() => Schema::Any,
                        ),
                        ..Default::default()
                    }))),
                ),
                required: set!("d".to_string()),
                ..Default::default()
            }),
            "h".into() => Schema::AnyOf(set! {
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::Integer),
            }),
            "i".into() => Schema::Any,
            "j".into() => Schema::Array(Box::new(Schema::Array(Box::new(Schema::Atomic(Atomic::String))))),
        ),
        ..Default::default()
    }),
    expected = CollectionAnalysis {
        collection_name: "test".to_string(),
        arrays: map! { "test.b".to_string() => 1, "test.j".to_string() => 1, "test.c.e".to_string() => 2, },
        arrays_of_arrays: map! { "test.j".to_string() => 1 },
        arrays_of_documents: map! { "test.c.e".to_string() => 2 },
        documents: map! { "test.c".to_string() => 1, "test.c.e".to_string() => 2},
        unstable: map! { "test.i".to_string() => 1, "test.c.e.g".to_string() => 3 },
        anyof: map! { "test.h".to_string() => 1 },
    }
);

test_process_schemata!(
    handles_many_schemas,
    schemata = map! {
        "test_db_a".to_string() => vec![
            map! {
                "test_coll_a".to_string() => Schema::Document(Document {
                    keys: map!(
                        "a".into() => Schema::Atomic(Atomic::String),
                        "b".into() => Schema::Atomic(Atomic::Integer),
                        "c".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                    ),
                    ..Default::default()
                }),
            },
            map! {
                "test_coll_b".to_string() => Schema::Document(Document {
                    keys: map!(
                        "a".into() => Schema::Atomic(Atomic::String),
                        "b".into() => Schema::Atomic(Atomic::Integer),
                        "c".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                    ),
                    ..Default::default()
                }),
            },
        ],
        "test_db_b".to_string() => vec![
            map! {
                "test_coll_a".to_string() => Schema::Document(Document {
                    keys: map!(
                        "a".into() => Schema::Atomic(Atomic::String),
                        "b".into() => Schema::Atomic(Atomic::Integer),
                        "c".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                    ),
                    ..Default::default()
                }),
            },
            map! {
                "test_coll_b".to_string() => Schema::Document(Document {
                    keys: map!(
                        "a".into() => Schema::Atomic(Atomic::String),
                        "b".into() => Schema::Atomic(Atomic::Integer),
                        "c".into() => Schema::Document(Document {
                            keys: map!(
                                "d".into() => Schema::Atomic(Atomic::String),
                                "e".into() => Schema::Atomic(Atomic::Integer),
                            ),
                            required: set!("d".to_string()),
                            ..Default::default()
                        }),
                    ),
                    ..Default::default()
                }),
            },
        ],
    },
    expected = SchemaAnalysis {
        database_analyses: map! {
            "test_db_a".to_string() => DatabaseAnalysis {
                database_name: "test_db_a".to_string(),
                collection_analyses: vec![
                    CollectionAnalysis {
                        collection_name: "test_coll_a".to_string(),
                        arrays: map! { "test_coll_a.c".to_string() => 1 },
                        ..Default::default()
                    },
                    CollectionAnalysis {
                        collection_name: "test_coll_b".to_string(),
                        arrays: map! { "test_coll_b.c".to_string() => 1 },
                        ..Default::default()
                    },
                ],
                ..Default::default()
            },
            "test_db_b".to_string() => DatabaseAnalysis {
                database_name: "test_db_b".to_string(),
                collection_analyses: vec![
                    CollectionAnalysis {
                        collection_name: "test_coll_a".to_string(),
                        arrays: map! { "test_coll_a.c".to_string() => 1 },
                        ..Default::default()
                    },
                    CollectionAnalysis {
                        collection_name: "test_coll_b".to_string(),
                        documents: map! { "test_coll_b.c".to_string() => 1 },
                        ..Default::default()
                    },
                ],
                ..Default::default()
            },
        },
    }
);
