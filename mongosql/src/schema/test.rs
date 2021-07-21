// +---------------------------+
// | ORD for Satisfaction test |
// +---------------------------+

mod satisfaction_ord {
    #[test]
    fn satisfaction_ord() {
        use crate::schema::Satisfaction::*;
        assert!(Must > May);
        assert!(May > Not);
    }
}

// +-------------------+
// | JSON schema tests |
// +-------------------+

mod from_json {
    use crate::{
        ir::schema::Error,
        json_schema,
        json_schema::BsonType,
        map, schema,
        schema::{Atomic::*, Document, Schema::*},
        set,
    };
    use std::convert::TryFrom;

    macro_rules! test_from_json_schema {
        ($func_name:ident, $schema_schema:expr, $json_schema:expr) => {
            #[test]
            fn $func_name() {
                let s = schema::Schema::try_from($json_schema);
                assert_eq!(s, $schema_schema);
            }
        };
    }

    test_from_json_schema!(
        convert_bson_single_to_atomic,
        Ok(Atomic(Integer)),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("int".to_string())),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        invalid_bson_type,
        Err(Error::InvalidJsonSchema(
            "blah is not a valid BSON type".to_string()
        )),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("blah".to_string())),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_bson_multiple_to_any_of,
        Ok(AnyOf(set![Atomic(Integer), Atomic(Null)])),
        json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec![
                "int".to_string(),
                "null".to_string()
            ])),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_one_of_to_any_of,
        Ok(AnyOf(set![Atomic(Integer), Atomic(Null)])),
        json_schema::Schema {
            one_of: Some(vec![
                json_schema::Schema {
                    bson_type: Some(BsonType::Single("int".to_string())),
                    ..Default::default()
                },
                json_schema::Schema {
                    bson_type: Some(BsonType::Single("null".to_string())),
                    ..Default::default()
                }
            ]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        one_of_invalid_nested_bson,
        Err(Error::InvalidJsonSchema(
            "blah is not a valid BSON type".to_string()
        )),
        json_schema::Schema {
            one_of: Some(vec![
                json_schema::Schema {
                    bson_type: Some(BsonType::Single("blah".to_string())),
                    ..Default::default()
                },
                json_schema::Schema {
                    bson_type: Some(BsonType::Single("null".to_string())),
                    ..Default::default()
                }
            ]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        one_of_invalid_extra_fields,
        Err(Error::InvalidJsonSchema(
            "invalid combination of fields".to_string()
        )),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("int".to_string())),
            one_of: Some(vec![json_schema::Schema {
                bson_type: Some(BsonType::Single("null".to_string())),
                ..Default::default()
            }]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_any_of_to_any_of,
        Ok(AnyOf(set![Atomic(Integer), Atomic(Null)])),
        json_schema::Schema {
            any_of: Some(vec![
                json_schema::Schema {
                    bson_type: Some(BsonType::Single("int".to_string())),
                    ..Default::default()
                },
                json_schema::Schema {
                    bson_type: Some(BsonType::Single("null".to_string())),
                    ..Default::default()
                }
            ]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_properties_to_document,
        Ok(Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
        })),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("object".to_string())),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        document_bson_type_not_object,
        Ok(Atomic(Integer)),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("int".to_string())),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        document_properties_not_set,
        Ok(Document(Document {
            keys: map![],
            required: set!["a".to_string()],
            additional_properties: true
        })),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("object".to_string())),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        document_additional_properties_not_set,
        Ok(Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
        })),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("object".to_string())),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_array_to_any_of,
        Ok(Array(Box::new(Atomic(Integer)))),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("array".to_string())),
            items: Some(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            })),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        items_set_bson_type_not_array,
        Ok(AnyOf(set![Atomic(Integer)])),
        json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec!["int".to_string(),])),
            items: Some(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            })),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        bson_type_array_set_missing_items_field,
        Ok(Array(Box::new(Any))),
        json_schema::Schema {
            bson_type: Some(BsonType::Single("array".to_string())),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_array_and_document_fields,
        Ok(AnyOf(set![
            Array(Box::new(Atomic(Integer))),
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(Integer),
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: true,
            })
        ])),
        json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec![
                "array".to_string(),
                "object".to_string()
            ])),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            items: Some(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            })),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        bson_type_object_set_missing_document_fields,
        Ok(AnyOf(set![
            Array(Box::new(Atomic(Integer))),
            Document(Document {
                keys: map![],
                required: set![],
                additional_properties: true
            })
        ])),
        json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec![
                "array".to_string(),
                "object".to_string()
            ])),
            items: Some(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single("int".to_string())),
                ..Default::default()
            })),
            ..Default::default()
        }
    );
}

// +-----------------+
// | Satisfies tests |
// +-----------------+

mod satisfies {
    use crate::{
        map,
        schema::{Atomic::*, Document, Satisfaction::*, Schema::*},
        set,
    };

    macro_rules! test_satisfies {
        ($func_name:ident, $expected:expr, $self:expr, $other:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let res = $self.satisfies(&$other);
                assert_eq!($expected, res)
            }
        };
    }

    test_satisfies!(any_must_satisfy_any, Must, Any, Any);
    test_satisfies!(missing_must_satisfy_any, Must, Missing, Any);
    test_satisfies!(
        any_of_empty_must_satisfy_atomic,
        Must,
        AnyOf(set![]),
        Atomic(Integer)
    );
    test_satisfies!(
        any_of_empty_must_satisfy_any_of_empty,
        Must,
        AnyOf(set![]),
        AnyOf(set![]),
    );
    test_satisfies!(
        any_of_empty_must_satisfy_missing,
        Must,
        AnyOf(set![]),
        Missing,
    );
    test_satisfies!(missing_must_satisfy_missing, Must, Missing, Missing);
    test_satisfies!(
        missing_must_satisfy_any_of,
        Must,
        Missing,
        AnyOf(set![Missing])
    );
    test_satisfies!(
        any_of_missing_may_satisfy_missing,
        May,
        AnyOf(set![Atomic(Integer), Missing, Atomic(String)]),
        Missing
    );
    test_satisfies!(
        missing_must_not_satisfy_atomic,
        Not,
        Missing,
        Atomic(String)
    );
    test_satisfies!(
        missing_must_not_satisfy_array,
        Not,
        Missing,
        Array(Box::new(Any)),
    );
    test_satisfies!(
        missing_must_not_satisfy_document,
        Not,
        Missing,
        Document(Document {
            keys: map![],
            required: set![],
            additional_properties: true,
        })
    );
    test_satisfies!(
        missing_must_not_satisfy_any_of,
        Not,
        Missing,
        AnyOf(set![Atomic(String), Atomic(Integer)])
    );
    test_satisfies!(atomic_must_satisfy_any, Must, Atomic(String), Any);
    test_satisfies!(any_may_satisfy_atomic, May, Any, Atomic(String));
    test_satisfies!(
        array_of_any_does_not_satisfy_atomic,
        Not,
        Array(Box::new(Any)),
        Atomic(Integer),
    );
    test_satisfies!(
        missing_does_not_satisfy_atomic,
        Not,
        Missing,
        Atomic(String),
    );
    test_satisfies!(
        any_of_must_satisfy_any,
        Must,
        AnyOf(set![Atomic(String), Atomic(Integer)]),
        Any,
    );
    test_satisfies!(
        any_of_must_satisfy_when_any_of_contains_any,
        Must,
        AnyOf(set![Atomic(String), Atomic(Integer)]),
        AnyOf(set![Atomic(String), Atomic(Integer), Any]),
    );
    test_satisfies!(
        array_of_string_must_satisfy_any_of_array_of_int_or_array_of_string,
        Must,
        Array(Box::new(Atomic(String))),
        AnyOf(set![
            Array(Box::new(Atomic(String))),
            Array(Box::new(Atomic(Integer)))
        ]),
    );
    test_satisfies!(
        array_of_string_or_int_may_satisfy_any_of_array_of_int_or_array_of_string,
        May,
        Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer),]))),
        AnyOf(set![
            Array(Box::new(Atomic(String))),
            Array(Box::new(Atomic(Integer)))
        ]),
    );
    test_satisfies!(
        array_of_string_or_int_must_satisfy_array_of_string_or_int,
        Must,
        Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer),]))),
        Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer),]))),
    );
    test_satisfies!(
        document_must_satify_same_document,
        Must,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
        }),
    );
    test_satisfies!(
        document_may_satify_with_more_permissive_key_schema,
        May,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
    );
    test_satisfies!(
        document_must_not_satify_with_incompatable_key_schema,
        Not,
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
    );
    test_satisfies!(
        document_may_satify_with_fewer_required_keys,
        May,
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
    );
    test_satisfies!(
        document_must_not_satify_with_missing_required_key,
        Not,
        Document(Document {
            keys: map![
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
    );
    test_satisfies!(
        document_may_satify_with_missing_required_key,
        May,
        Document(Document {
            keys: map![
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: true,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
        }),
    );
    test_satisfies!(
        document_must_satify_with_more_required_keys,
        Must,
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
        }),
    );
    test_satisfies!(
        document_may_satify_due_to_possible_extra_keys,
        May,
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: true,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
        }),
    );
    test_satisfies!(
        document_satifies_multiple_any_of_results_in_must_satisfy,
        Must,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
            }),
            Document(Document {
                keys: map![
                    "b".to_string() => Atomic(Integer),
                ],
                required: set![],
                additional_properties: false,
            }),
        ]),
    );
    test_satisfies!(
        document_satifies_any_of_any_of_results_must_satisfy,
        Must,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
            }),
            Document(Document {
                keys: map![
                    "e".to_string() => Atomic(Integer),
                ],
                required: set![],
                additional_properties: false,
            }),
        ]),
    );
    test_satisfies!(
        document_may_satisfy_when_key_schema_may_satisfy,
        May,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
    );
    test_satisfies!(
        array_may_satisfy_when_array_item_schema_may_satisfy,
        May,
        Array(Box::new(Any)),
        Array(Box::new(Atomic(Integer))),
    );
    test_satisfies!(
        array_may_satisfy_when_array_item_schema_may_satisfy_multiple_any_of_array,
        May,
        Array(Box::new(Any)),
        AnyOf(set![
            Array(Box::new(Atomic(Integer))),
            Array(Box::new(Atomic(String))),
        ]),
    );
    test_satisfies!(
        array_may_satisfy_when_array_item_schema_may_satisfy_multiple_array_any_of,
        May,
        Array(Box::new(Any)),
        Array(Box::new(AnyOf(set![Atomic(Integer), Atomic(Double),]),)),
    );
    test_satisfies!(
        array_of_missing_does_not_satisfy_array_of_atomic,
        Not,
        Array(Box::new(Missing)),
        Array(Box::new(Atomic(Integer))),
    );
}

mod has_overlaping_keys_with {
    use crate::{
        map,
        schema::{Atomic, Document, Satisfaction, Schema, ANY_DOCUMENT, EMPTY_DOCUMENT},
        set,
    };

    macro_rules! test_has_overlapping_keys_with {
        ($func_name:ident, $expected:expr, $schema1:expr, $schema2:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let out = $schema1.has_overlapping_keys_with($schema2);
                assert_eq!($expected, out);
            }
        };
    }

    test_has_overlapping_keys_with!(
        any_may_overlap_any_document,
        Satisfaction::May,
        &Schema::Any,
        &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        any_overlap_may_any_document_symmetric,
        Satisfaction::May,
        &ANY_DOCUMENT,
        &Schema::Any,
    );
    test_has_overlapping_keys_with!(
        atomic_has_no_keys_to_overlap,
        Satisfaction::Not,
        Schema::Atomic(Atomic::Integer),
        &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        atomic_has_no_keys_to_overlap_symmetric,
        Satisfaction::Not,
        &ANY_DOCUMENT,
        &Schema::Atomic(Atomic::Integer),
    );
    test_has_overlapping_keys_with!(
        any_document_may_overlap_keys_with_any_document,
        Satisfaction::May,
        &ANY_DOCUMENT,
        &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        explicit_document_may_overlap_keys_with_any_document,
        Satisfaction::May,
        Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
        }),
        &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        any_document_does_not_overlap_with_empty_document,
        Satisfaction::Not,
        &EMPTY_DOCUMENT,
        &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        explicit_document_may_overlap_keys_with_any_document_symmetric,
        Satisfaction::May,
        &ANY_DOCUMENT,
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
        }),
    );
    test_has_overlapping_keys_with!(
        any_document_does_not_overlap_with_empty_document_symmetric,
        Satisfaction::Not,
        &ANY_DOCUMENT,
        &EMPTY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        two_explicit_documents_without_required_keys_may_overlap,
        Satisfaction::May,
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
        }),
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
        }),
    );
    test_has_overlapping_keys_with!(
        two_explicit_documents_with_required_keys_may_overlap,
        Satisfaction::May,
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"a".into()},
            additional_properties: false,
        }),
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"b".into()},
            additional_properties: false,
        }),
    );
    test_has_overlapping_keys_with!(
        two_explicit_documents_with_required_keys_must_overlap,
        Satisfaction::Must,
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"a".into()},
            additional_properties: false,
        }),
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"a".into()},
            additional_properties: false,
        }),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_may_overlap,
        Satisfaction::May,
        &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"b".into()},
                additional_properties: false,
            }),
        ]),
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"c".into()},
            additional_properties: false,
        }),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_must_overlap,
        Satisfaction::Must,
        &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
        ]),
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"a".into()},
            additional_properties: false,
        }),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_may_overlap_symmetric,
        Satisfaction::May,
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"c".into()},
            additional_properties: false,
        }),
        &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"b".into()},
                additional_properties: false,
            }),
        ]),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_must_overlap_symmetric,
        Satisfaction::Must,
        &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"a".into()},
            additional_properties: false,
        }),
        &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
            }),
        ]),
    );
}

mod document_union {
    use crate::{
        map,
        schema::{Atomic, Document, Schema, ANY_DOCUMENT, EMPTY_DOCUMENT},
        set,
    };

    macro_rules! test_document_union {
        ($func_name:ident, $expected:expr, $schema1:expr, $schema2:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let out = $schema1.document_union($schema2);
                assert_eq!($expected, out);
            }
        };
    }

    test_document_union!(
        schema_does_not_satisfy_document_results_in_any,
        EMPTY_DOCUMENT.clone(),
        Schema::Atomic(Atomic::Integer),
        ANY_DOCUMENT.clone(),
    );
    test_document_union!(
        schema_does_not_satisfy_document_results_in_any_symmetric,
        EMPTY_DOCUMENT.clone(),
        ANY_DOCUMENT.clone(),
        Schema::Atomic(Atomic::Integer),
    );
    test_document_union!(
        document_union_of_two_documents_will_document_union_keys_and_intersect_required_wo_additional_properties,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Decimal),
                    Schema::Atomic(Atomic::Integer),
                ]),
                "b".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Integer),
                ]),
            },
            required: set! {"a".into()},
            additional_properties: false,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Decimal),
                "b".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: false,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "b".into()},
            additional_properties: false,
        }),
    );
    test_document_union!(
        document_union_of_two_documents_will_document_union_keys_and_intersect_required_wo_additional_properties_symmetric,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                ]),
                "b".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Double),
                ]),
            },
            required: set! {"a".into()},
            additional_properties: false,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "b".into()},
            additional_properties: false,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Decimal),
                "b".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: false,
        }),
    );
    test_document_union!(
        document_union_of_two_documents_will_retain_keys_when_first_is_open,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                ]),
                "b".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: true,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Decimal),
                "b".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: true,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
        }),
    );
    test_document_union!(
        document_union_of_two_documents_will_retain_keys_when_second_is_open,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                ]),
                "c".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: true,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Decimal),
                "c".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: true,
        }),
    );
    test_document_union!(
        document_union_of_two_documents_will_intersect_keys_when_both_are_open,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                ]),
            },
            required: set! {"a".into()},
            additional_properties: true,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: true,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Decimal),
                "b".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: true,
        }),
    );
    test_document_union!(
        document_union_of_any_doc_with_doc_is_any_doc,
        ANY_DOCUMENT.clone(),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
        }),
        ANY_DOCUMENT.clone(),
    );
    test_document_union!(
        document_union_of_any_doc_with_doc_is_any_doc_symmetric,
        ANY_DOCUMENT.clone(),
        ANY_DOCUMENT.clone(),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
        }),
    );
    test_document_union!(
        document_union_of_doc_with_empty_doc_is_doc_with_no_required,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
        }),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
        }),
        EMPTY_DOCUMENT.clone(),
    );
    test_document_union!(
        document_union_of_doc_with_empty_doc_is_doc_with_no_required_symmetric,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
        }),
        EMPTY_DOCUMENT.clone(),
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
        }),
    );
    test_document_union!(
        document_union_of_any_of_recursively_applies_document_union,
        Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::AnyOf(set![
                        Schema::AnyOf(set![
                            Schema::Atomic(Atomic::Integer),
                            Schema::Atomic(Atomic::Decimal),
                        ]),
                        Schema::Atomic(Atomic::Decimal),
                    ]),
                    Schema::Atomic(Atomic::Integer),
                ]),
                "b".into() => Schema::AnyOf(set![
                    Schema::AnyOf(set![
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Double),
                    ]),
                    Schema::Atomic(Atomic::Integer),
                ]),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: false,
        }),
        Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"a".into(), "b".into()},
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                    "c".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"a".into(), "b".into()},
                additional_properties: false,
            })
        ]),
        Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Decimal),
                    "b".into() => Schema::Atomic(Atomic::Double),
                },
                required: set! {"a".into()},
                additional_properties: false,
            }),
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Decimal),
                    "d".into() => Schema::Atomic(Atomic::Double),
                },
                required: set! {"a".into()},
                additional_properties: false,
            }),
        ])
    );
}

// +---------------------+
// | Comparability tests |
// +---------------------+

mod is_comparable_with {
    use crate::{
        schema::{Atomic::*, Satisfaction::*, Schema::*, ANY_ARRAY, ANY_DOCUMENT},
        set,
    };

    macro_rules! test_is_comparable_with {
        ($func_name:ident, $expected:expr, $self:expr, $other:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let mut res = $self.is_comparable_with(&$other);
                assert_eq!($expected, res);
                res = $other.is_comparable_with(&$self);
                assert_eq!($expected, res)
            }
        };
    }

    // Disallowed comparability tests (arrays and documents).
    test_is_comparable_with!(
        array_not_comparable_with_another_array,
        Not,
        ANY_ARRAY,
        ANY_ARRAY,
    );
    test_is_comparable_with!(
        document_not_comparable_with_another_document,
        Not,
        ANY_DOCUMENT,
        ANY_DOCUMENT,
    );
    test_is_comparable_with!(
        array_not_comparable_with_document,
        Not,
        ANY_ARRAY,
        ANY_DOCUMENT,
    );

    test_is_comparable_with!(array_not_comparable_with_any, Not, ANY_ARRAY, Any,);
    test_is_comparable_with!(
        array_not_comparable_with_another_type,
        Not,
        ANY_ARRAY,
        Atomic(Integer),
    );
    test_is_comparable_with!(array_not_comparable_with_null, Not, ANY_ARRAY, Atomic(Null),);
    test_is_comparable_with!(array_not_comparable_with_missing, Not, ANY_ARRAY, Missing,);
    test_is_comparable_with!(array_not_comparable_with_unsat, Not, ANY_ARRAY, Unsat,);

    test_is_comparable_with!(
        document_not_comparable_with_any_type,
        Not,
        ANY_DOCUMENT,
        Any,
    );
    test_is_comparable_with!(
        document_not_comparable_with_a_type,
        Not,
        ANY_DOCUMENT,
        Atomic(Integer),
    );
    test_is_comparable_with!(
        document_not_comparable_with_null,
        Not,
        ANY_DOCUMENT,
        Atomic(Null),
    );
    test_is_comparable_with!(
        document_not_comparable_with_missing,
        Not,
        ANY_DOCUMENT,
        Missing,
    );
    test_is_comparable_with!(document_not_comparable_with_unsat, Not, ANY_DOCUMENT, Unsat,);

    // Any comparison tests.
    test_is_comparable_with!(
        any_type_may_be_comparable_with_another_type,
        May,
        Any,
        Atomic(Integer),
    );
    test_is_comparable_with!(any_type_may_be_comparable_with_null, May, Any, Atomic(Null),);
    test_is_comparable_with!(any_type_may_be_comparable_with_missing, May, Any, Missing,);
    test_is_comparable_with!(any_type_may_be_comparable_with_unsat, May, Any, Unsat,);

    // Missing comparability tests.
    test_is_comparable_with!(
        missing_must_be_comparable_with_missing,
        Must,
        Missing,
        Missing,
    );
    test_is_comparable_with!(
        missing_must_be_comparable_with_another_type,
        Must,
        Missing,
        Atomic(Integer),
    );
    test_is_comparable_with!(
        missing_must_be_comparable_with_null,
        Must,
        Missing,
        Atomic(Null),
    );
    test_is_comparable_with!(missing_must_be_comparable_with_unsat, Must, Missing, Unsat,);

    // Unsat comparability tests.
    test_is_comparable_with!(unsat_must_be_comparable_with_unsat, Must, Unsat, Unsat,);
    test_is_comparable_with!(
        unsat_must_be_comparable_with_another_type,
        Must,
        Unsat,
        Atomic(Integer),
    );
    test_is_comparable_with!(
        unsat_must_be_comparable_with_null,
        Must,
        Unsat,
        Atomic(Null),
    );

    // Atomic comparability tests.
    test_is_comparable_with!(
        null_must_be_comparable_with_null,
        Must,
        Atomic(Null),
        Atomic(Null),
    );
    test_is_comparable_with!(
        null_must_be_comparable_with_atomic_numeric,
        Must,
        Atomic(Null),
        Atomic(Integer),
    );
    test_is_comparable_with!(
        atomic_numeric_must_be_comparable_with_same_atomic_numeric,
        Must,
        Atomic(Integer),
        Atomic(Integer),
    );
    test_is_comparable_with!(
        atomic_numeric_must_be_comparable_with_different_atomic_numeric,
        Must,
        Atomic(Integer),
        Atomic(Double),
    );
    test_is_comparable_with!(
        non_numeric_atomic_must_be_comparable_with_same_non_numeric_atomic,
        Must,
        Atomic(String),
        Atomic(String),
    );
    test_is_comparable_with!(
        atomic_not_comparable_with_different_atomic,
        Not,
        Atomic(String),
        Atomic(Integer),
    );

    // AnyOf comparability tests (numeric).
    test_is_comparable_with!(
        numeric_atomic_must_be_comparable_with_a_set_of_numerics,
        Must,
        Atomic(Integer),
        AnyOf(set![Atomic(Integer), Atomic(Long)]),
    );
    test_is_comparable_with!(
        a_set_of_numerics_must_be_comparable_with_a_disjoint_set_of_numerics,
        Must,
        AnyOf(set![Atomic(Integer), Atomic(Long)]),
        AnyOf(set![Atomic(Double), Atomic(Decimal)]),
    );
    test_is_comparable_with!(
        numeric_must_be_comparable_with_different_numeric_or_null,
        Must,
        Atomic(Integer),
        AnyOf(set![Atomic(Long), Atomic(Null)]),
    );
    test_is_comparable_with!(
        numeric_or_null_must_be_comparable_with_different_numeric_or_null,
        Must,
        AnyOf(set![Atomic(Integer), Atomic(Null)]),
        AnyOf(set![Atomic(Long), Atomic(Null)]),
    );
    test_is_comparable_with!(
        numeric_atomic_may_be_comparable_with_potentially_same_numeric,
        May,
        Atomic(Integer),
        AnyOf(set![Atomic(Integer), Atomic(String)]),
    );
    test_is_comparable_with!(
        potential_numeric_may_be_comparable_with_potentially_same_numeric,
        May,
        AnyOf(set![Atomic(Integer), Atomic(String)]),
        AnyOf(set![Atomic(Integer), Atomic(String)]),
    );
    test_is_comparable_with!(
        potential_numeric_may_be_comparable_with_potentially_different_numeric,
        May,
        AnyOf(set![Atomic(Integer), Atomic(String)]),
        AnyOf(set![Atomic(Double), Atomic(String)]),
    );

    // AnyOf comparability tests (non-numeric).
    test_is_comparable_with!(
        atomic_must_be_comparable_with_same_atomic_or_null,
        Must,
        Atomic(String),
        AnyOf(set![Atomic(String), Atomic(Null)]),
    );
    test_is_comparable_with!(
        atomic_or_null_must_be_comparable_with_same_atomic_or_null,
        Must,
        AnyOf(set![Atomic(String), Atomic(Null)]),
        AnyOf(set![Atomic(String), Atomic(Null)]),
    );
    test_is_comparable_with!(
        atomic_may_be_comparable_with_potentially_same_atomic,
        May,
        Atomic(String),
        AnyOf(set![Atomic(String), Atomic(Integer)]),
    );
    test_is_comparable_with!(
        atomic_or_null_may_be_comparable_with_different_atomic_or_null,
        May,
        AnyOf(set![Atomic(String), Atomic(Null)]),
        AnyOf(set![Atomic(Integer), Atomic(Null)]),
    );
    test_is_comparable_with!(
        some_atomic_may_be_comparable_with_potentially_same_atomic,
        May,
        AnyOf(set![Atomic(String), Atomic(Boolean)]),
        AnyOf(set![Atomic(String), Atomic(Integer)]),
    );
    test_is_comparable_with!(
        a_set_of_atomics_not_comparable_with_disjoint_set_of_atomics,
        Not,
        AnyOf(set![Atomic(String), Atomic(Boolean)]),
        AnyOf(set![Atomic(Date), Atomic(Integer)]),
    );
}

// +----------------------+
// | Contains field tests |
// +----------------------+

mod contains_field {
    use crate::{
        map,
        schema::{Atomic::*, Document, Satisfaction::*, Schema::*},
        set,
    };

    macro_rules! test_contains_field {
        ($func_name:ident, $expected:expr, $self:expr, $other:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let res = $self.contains_field($other);
                assert_eq!($expected, res)
            }
        };
    }

    test_contains_field!(any_may_contain_field, May, Any, "a");
    test_contains_field!(missing_does_not_contain_field, Not, Missing, "a");
    test_contains_field!(
        document_must_contain_field,
        Must,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        "a",
    );
    test_contains_field!(
        document_may_contain_field,
        May,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        "b",
    );
    test_contains_field!(
        document_may_contain_field_due_to_additional_properties,
        May,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
        }),
        "foo",
    );
    test_contains_field!(
        document_must_not_contain_field,
        Not,
        Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
        }),
        "foo",
    );
    test_contains_field!(atomic_must_not_contain_field, Not, Atomic(String), "foo",);
    test_contains_field!(
        any_of_document_and_atomic_may_not_contain_field,
        Not,
        AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
            }),
            Atomic(String),
        ]),
        "c",
    );
    test_contains_field!(
        any_of_document_and_atomic_may_contain_field,
        May,
        AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
            }),
            Atomic(String),
        ]),
        "b",
    );
    test_contains_field!(
        any_of_document_and_document_must_contain_field,
        Must,
        AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["b".to_string()],
                additional_properties: false,
            }),
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(String),
                ],
                required: set!["b".to_string()],
                additional_properties: false,
            }),
        ]),
        "b",
    );
}

// +----------------+
// | Simplify tests |
// +----------------+

mod simplify {
    use crate::{
        map, schema,
        schema::{Atomic::*, Document, Schema::*},
        set,
    };

    macro_rules! test_simplify {
        ($func_name:ident, $expected:expr, $input:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let res = schema::Schema::simplify(&$input);
                assert_eq!($expected, res)
            }
        };
    }

    test_simplify!(contains_empty_vec, Unsat, AnyOf(set![]));
    test_simplify!(
        remove_any_of_duplicates,
        AnyOf(set![Atomic(String), Atomic(Integer)]),
        AnyOf(set![Atomic(Integer), Atomic(Integer), Atomic(String)])
    );
    test_simplify!(
        remove_any_of_duplicates_not_consecutive,
        AnyOf(set![Atomic(String), Atomic(Integer)]),
        AnyOf(set![
            Atomic(Integer),
            Atomic(Integer),
            Atomic(String),
            Atomic(Integer)
        ])
    );
    test_simplify!(flatten_any_is_flat, Any, Any);
    test_simplify!(
        flatten_any_of_one_schema,
        Atomic(Integer),
        AnyOf(set![Atomic(Integer)])
    );
    test_simplify!(flatten_any_of_any_schema, Any, AnyOf(set!(Any, Missing)));
    test_simplify!(
        flatten_any_of_any_of,
        AnyOf(set![Missing, Atomic(String), Atomic(Integer)]),
        AnyOf(set![AnyOf(set![Missing, Atomic(String)]), Atomic(Integer)]),
    );
    test_simplify!(
        flatten_any_of_and_remove_duplicates,
        AnyOf(set![Atomic(String), Atomic(Integer), Atomic(Null)]),
        AnyOf(set![
            AnyOf(set![Atomic(Integer), Atomic(String)]),
            AnyOf(set![Atomic(Integer), Atomic(Null)])
        ])
    );
    test_simplify!(
        flatten_any_of_containing_array,
        Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer)]))),
        AnyOf(set![Array(Box::new(AnyOf(set![
            Atomic(Integer),
            Atomic(String)
        ])))])
    );
    test_simplify!(
        flatten_any_of_and_return_single_element,
        Atomic(Integer),
        AnyOf(set![Atomic(Integer), Atomic(Integer)])
    );
    test_simplify!(
        array,
        Array(Box::new(AnyOf(set![
            Missing,
            Atomic(String),
            Atomic(Integer)
        ]))),
        Array(Box::new(AnyOf(set![
            AnyOf(set![Missing, Atomic(String)]),
            Atomic(Integer)
        ])))
    );
    test_simplify!(
        document,
        Document(Document {
            keys: map![
                "a".to_string() => AnyOf(set![
                Missing,
                Atomic(String),
                Atomic(Integer)
            ])
                ],
            required: set!["a".to_string()],
            additional_properties: true,
        }),
        Document(Document {
            keys: map![
                "a".to_string() => AnyOf(set![
                AnyOf(set![Missing, Atomic(String)]),
                Atomic(Integer)
            ]),
                            ],
            required: set!["a".to_string()],
            additional_properties: true,
        })
    );
}
