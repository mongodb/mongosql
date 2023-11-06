use crate::{
    catalog::{Catalog, Namespace},
    map,
    mir::{
        schema::{Atomic, Document, Error, Join, JoinType, ResultSet, SchemaCache},
        Collection, FieldPath, Stage, Unwind,
    },
    schema::Schema,
    set, test_schema,
};

/// Most tests use the same source, path, and cache for the Unwind stage.
/// Typically, only the index and outer parameters change, so this helper
/// allows us to easily and concisely construct test Unwinds.
fn make_unwind(index: Option<String>, outer: bool) -> Stage {
    Stage::Unwind(Unwind {
        source: Box::new(Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into(),
            cache: SchemaCache::new(),
        })),
        path: FieldPath {
            key: ("foo", 0u16).into(),
            fields: vec!["arr".to_string()],
            is_nullable: true,
        },
        index,
        outer,
        cache: SchemaCache::new(),
    })
}

/// Most tests use the same collection source and need to specify the
/// collection schema for the test to work. This helper allows easy
/// definition of that collection schema.
fn make_catalog(s: Schema) -> Catalog {
    Catalog::new(map! {
        Namespace {db: "test".into(), collection: "foo".into()} => s,
    })
}

mod no_index {
    use super::*;

    test_schema!(
        path_result_schema_is_inner_schema_when_path_must_be_array,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = make_unwind(None, false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
            },
            required: set! { "arr".into() },
            additional_properties: false,
        })),
    );

    test_schema!(
        path_result_schema_is_any_of_inner_schemas_and_other_schemas_when_path_may_be_array_but_never_null_or_missing,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::AnyOf(set! [
                            Schema::Atomic(Atomic::Double),
                            Schema::Atomic(Atomic::Long),
                            Schema::Atomic(Atomic::Integer),
                        ]),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = make_unwind(None, false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::AnyOf(set! [
                    Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                    Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                    Schema::Atomic(Atomic::Integer),
                ]),
            },
            required: set! { "arr".into() },
            additional_properties: false,
        })),
    );

    test_schema!(
        path_result_schema_is_inner_schema_when_path_may_be_array_null_or_missing,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = make_unwind(None, false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::AnyOf(set! [
                    Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                    Schema::Atomic(Atomic::Null),
                    Schema::Missing,
                ]),
            },
            required: set! {},
            additional_properties: false,
        })),
    );

    test_schema!(
        path_result_schema_is_unchanged_when_path_not_array_and_not_nullish,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = make_unwind(None, false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! { "arr".into() },
            additional_properties: false,
        })),
    );

    test_schema!(
        path_result_schema_is_non_nullish_when_path_not_array_but_may_be_nullish,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Atomic(Atomic::Integer),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = make_unwind(None, false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::AnyOf(set! [
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Null),
                ]),
            },
            required: set! {},
            additional_properties: false,
        })),
    );

    test_schema!(
        path_result_schema_retains_path_schema_nullability_from_within_array_regardless_of_outer_false,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::AnyOf(set![
                            Schema::Atomic(Atomic::String),
                            Schema::Atomic(Atomic::Null)
                        ]),
                    },
                    required: set! { "arr".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = make_unwind(None, false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::AnyOf(set![
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Null)
                ]))),
            },
            required: set! { "arr".into() },
            additional_properties: false,
        })),
    );

    test_schema!(
        correlated_path_disallowed,
        expected_error_code = 1000,
        expected = Err(Error::DatasourceNotFoundInSchemaEnv(("bar", 0u16).into())),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            path: FieldPath {
                key: ("bar", 0u16).into(),
                fields: vec!["arr".to_string()],
                is_nullable: false,
            },
            index: None,
            outer: false,
            cache: SchemaCache::new(),
        }),
        schema_env = map! {
            ("bar", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            }),
        },
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
            },
            required: set! { "arr".into() },
            additional_properties: false,
        })),
    );

    mod outer_true {
        use super::*;

        test_schema!(
            path_result_schema_is_any_of_inner_schema_or_missing_when_path_must_be_array,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::String),
                                Schema::Missing,
                            ]),
                        },
                        required: set! { },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(None, true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );

        test_schema!(
            path_result_schema_is_any_of_inner_schemas_or_other_schemas_or_missing_when_path_may_be_array_but_never_null_or_missing,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::Double),
                                Schema::Atomic(Atomic::Long),
                                Schema::Atomic(Atomic::Integer),
                                Schema::Missing,
                            ]),
                        },
                        required: set! { },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(None, true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::AnyOf(set! [
                        Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                        Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                        Schema::Atomic(Atomic::Integer),
                    ]),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );

        test_schema!(
            path_result_schema_is_any_of_inner_schema_null_or_missing_when_path_may_be_array_null_or_missing,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::String),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing,
                            ]),
                        },
                        required: set! { },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(None, true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::AnyOf(set! [
                        Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    ]),
                },
                required: set! { },
                additional_properties: false,
            })),
        );

        test_schema!(
            path_result_schema_is_unchanged_when_path_not_array,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::Integer),
                                Schema::Missing,
                            ]),
                        },
                        required: set! { },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(None, true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {},
                additional_properties: false,
            })),
        );
    }
}

mod index_does_not_conflict {
    use super::*;

    mod outer_false {
        use super::*;

        test_schema!(
            index_result_schema_is_not_nullable_when_path_must_be_array,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::String),
                            "idx".into() => Schema::Atomic(Atomic::Long),
                        },
                        required: set! { "arr".into(), "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), false),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );

        test_schema!(
            index_result_schema_is_nullable_when_path_may_be_array_but_never_null_or_missing,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set![
                                Schema::Atomic(Atomic::Double),
                                Schema::Atomic(Atomic::Long),
                                Schema::Atomic(Atomic::Integer),
                            ]),
                            "idx".into() => Schema::AnyOf(set![
                                Schema::Atomic(Atomic::Long),
                                Schema::Atomic(Atomic::Null)
                            ]),
                        },
                        required: set! { "arr".into(), "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), false),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::AnyOf(set! [
                        Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                        Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                        Schema::Atomic(Atomic::Integer),
                    ]),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );

        test_schema!(
            index_result_schema_is_not_nullable_when_path_may_be_array_null_or_missing,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::String),
                            "idx".into() => Schema::Atomic(Atomic::Long),
                        },
                        required: set! { "arr".into(), "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), false),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::AnyOf(set! [
                        Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    ]),
                },
                required: set! {},
                additional_properties: false,
            })),
        );

        test_schema!(
            index_result_schema_is_exactly_null_when_path_not_array,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::Integer),
                            "idx".into() => Schema::Atomic(Atomic::Null),
                        },
                        required: set! { "arr".into(), "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), false),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );
    }

    mod outer_true {
        use super::*;

        test_schema!(
            index_result_schema_is_nullable_when_path_must_be_array,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::String),
                                Schema::Missing,
                            ]),
                            "idx".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::Long),
                                Schema::Atomic(Atomic::Null),
                            ]),
                        },
                        required: set! { "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );

        test_schema!(
            index_result_schema_is_nullable_when_path_may_be_array_but_never_null_or_missing,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set![
                                Schema::Atomic(Atomic::Double),
                                Schema::Atomic(Atomic::Long),
                                Schema::Atomic(Atomic::Integer),
                                Schema::Missing,
                            ]),
                            "idx".into() => Schema::AnyOf(set![
                                Schema::Atomic(Atomic::Long),
                                Schema::Atomic(Atomic::Null)
                            ]),
                        },
                        required: set! { "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::AnyOf(set! [
                        Schema::Array(Box::new(Schema::Atomic(Atomic::Double))),
                        Schema::Array(Box::new(Schema::Atomic(Atomic::Long))),
                        Schema::Atomic(Atomic::Integer),
                    ]),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );

        test_schema!(
            index_result_schema_is_nullable_when_path_may_be_array_null_or_missing,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::String),
                                Schema::Atomic(Atomic::Null),
                                Schema::Missing,
                            ]),
                            "idx".into() => Schema::AnyOf(set! [
                                Schema::Atomic(Atomic::Long),
                                Schema::Atomic(Atomic::Null),
                            ]),
                        },
                        required: set! { "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::AnyOf(set! [
                        Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    ]),
                },
                required: set! {},
                additional_properties: false,
            })),
        );

        test_schema!(
            index_result_schema_is_exactly_null_when_path_not_array,
            expected = Ok(ResultSet {
                schema_env: map! {
                    ("foo", 0u16).into() => Schema::Document(Document {
                        keys: map! {
                            "arr".into() => Schema::Atomic(Atomic::Integer),
                            "idx".into() => Schema::Atomic(Atomic::Null),
                        },
                        required: set! { "arr".into(), "idx".into() },
                        additional_properties: false,
                    }),
                },
                min_size: 0,
                max_size: None,
            }),
            input = make_unwind(Some("idx".into()), true),
            catalog = make_catalog(Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            })),
        );
    }

    test_schema!(
        index_is_put_at_top_level_even_when_path_is_nested,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "a".into() => Schema::Document(Document {
                            keys: map! {
                                "b".into() => Schema::Document(Document {
                                    keys: map! {
                                        "arr".into() => Schema::Atomic(Atomic::String),
                                    },
                                    required: set! { "arr".into() },
                                    additional_properties: false,
                                }),
                            },
                            required: set! { "b".into() },
                            additional_properties: false,
                        }),
                        "idx".into() => Schema::Atomic(Atomic::Long),
                    },
                    required: set! { "a".into(), "idx".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into(),
                cache: SchemaCache::new(),
            })),
            path: FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["a".to_string(), "b".to_string(), "arr".to_string()],
                is_nullable: false,
            },
            index: Some("idx".into()),
            outer: false,
            cache: SchemaCache::new(),
        }),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(Document {
                            keys: map! {
                                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                            },
                            required: set! { "arr".into() },
                            additional_properties: false,
                        })
                    },
                    required: set! { "b".into() },
                    additional_properties: false,
                }),
            },
            required: set! { "a".into() },
            additional_properties: false,
        })),
    );

    test_schema!(
        index_conflicts_are_only_checked_against_path_datasource,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Atomic(Atomic::String),
                        "idx".into() => Schema::Atomic(Atomic::Long),
                    },
                    required: set! { "arr".into(), "idx".into() },
                    additional_properties: false,
                }),
                ("bar", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "idx".into() => Schema::Atomic(Atomic::String),
                    },
                    required: set! { "idx".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = Stage::Unwind(Unwind {
            source: Box::new(Stage::Join(Join {
                join_type: JoinType::Inner,
                left: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "foo".into(),
                    cache: SchemaCache::new(),
                })),
                right: Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "bar".into(),
                    cache: SchemaCache::new(),
                })),
                condition: None,
                cache: SchemaCache::new(),
            })),
            path: FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["arr".to_string()],
                is_nullable: false,
            },
            index: Some("idx".into()),
            outer: false,
            cache: SchemaCache::new(),
        }),
        catalog = Catalog::new(map! {
            Namespace {db: "test".into(), collection: "foo".into()} => Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                },
                required: set! { "arr".into() },
                additional_properties: false,
            }),
            Namespace {db: "test".into(), collection: "bar".into()} => Schema::Document(Document {
                keys: map! {
                    "idx".into() => Schema::Atomic(Atomic::String),
                },
                required: set! { "idx".into() },
                additional_properties: false,
            }),
        }),
    );
}

mod index_may_conflict {
    use super::*;

    test_schema!(
        error_in_strict_mode,
        expected_error_code = 1014,
        expected = Err(Error::UnwindIndexNameConflict("idx".into())),
        input = make_unwind(Some("idx".into()), false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                "idx".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "arr".into(), "idx".into() },
            additional_properties: false,
        })),
    );

    test_schema!(
        succeed_in_relaxed_mode,
        expected = Ok(ResultSet {
            schema_env: map! {
                ("foo", 0u16).into() => Schema::Document(Document {
                    keys: map! {
                        "arr".into() => Schema::Atomic(Atomic::String),
                        "idx".into() => Schema::Atomic(Atomic::Long),
                    },
                    required: set! { "arr".into(), "idx".into() },
                    additional_properties: false,
                }),
            },
            min_size: 0,
            max_size: None,
        }),
        input = make_unwind(Some("idx".into()), false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                "idx".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "arr".into() },
            additional_properties: false,
        })),
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );
}

mod index_must_conflict {
    use super::*;

    test_schema!(
        error_in_strict_mode,
        expected_error_code = 1014,
        expected = Err(Error::UnwindIndexNameConflict("idx".into())),
        input = make_unwind(Some("idx".into()), false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                "idx".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "arr".into(), "idx".into() },
            additional_properties: false,
        })),
    );

    test_schema!(
        error_in_relaxed_mode,
        expected_error_code = 1014,
        expected = Err(Error::UnwindIndexNameConflict("idx".into())),
        input = make_unwind(Some("idx".into()), false),
        catalog = make_catalog(Schema::Document(Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                "idx".into() => Schema::Atomic(Atomic::String),
            },
            required: set! { "arr".into(), "idx".into() },
            additional_properties: false,
        })),
        schema_checking_mode = SchemaCheckingMode::Relaxed,
    );
}
