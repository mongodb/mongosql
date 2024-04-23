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

mod schema_ord {
    #[test]
    fn atomic_ord() {
        use crate::schema::Atomic::*;
        assert!(MinKey < Null);
        assert!(Null < Integer);
        assert!(Integer < Long);
        assert!(Long < Double);
        assert!(Double < Decimal);
        assert!(Decimal < Symbol);
        assert!(Symbol < String);
        assert!(String < BinData);
        assert!(BinData < ObjectId);
        assert!(ObjectId < Boolean);
        assert!(Boolean < Date);
        assert!(Date < Timestamp);
        assert!(Timestamp < Regex);
        assert!(Regex < DbPointer);
        assert!(DbPointer < Javascript);
        assert!(Javascript < JavascriptWithScope);
        assert!(JavascriptWithScope < MaxKey);
    }

    #[test]
    fn schema_ord() {
        use crate::{
            schema::{Document, Schema::*},
            set,
        };
        assert!(Unsat < Missing);
        assert!(Missing < Atomic(crate::schema::Atomic::MinKey));
        assert!(Atomic(crate::schema::Atomic::MaxKey) < Array(Box::new(Unsat)));
        assert!(Array(Box::new(Any)) < Document(Document::empty()));
        assert!(Document(Document::empty()) < AnyOf(set! {}));
        assert!(AnyOf(set! {}) < Any);
    }
}

mod union_schemata {
    use crate::{
        map,
        schema::{Atomic::*, Document, Schema::*},
        set,
    };
    macro_rules! test_union_schema {
        ($func_name:ident, expected = $expected:expr, left = $left:expr, right = $right:expr,) => {
            #[test]
            fn $func_name() {
                assert_eq!($expected, $left.union(&$right));
            }
        };
    }

    test_union_schema!(
        union_identical_schemata_is_no_op,
        expected = AnyOf(set! {Atomic(Integer), Atomic(Double), Atomic(Decimal)}),
        left = AnyOf(set! {Atomic(Integer), Atomic(Double), Atomic(Decimal)}),
        right = AnyOf(set! {Atomic(Integer), Atomic(Double), Atomic(Decimal)}),
    );
    test_union_schema!(
        union_two_divergent_schemata,
        expected = AnyOf(set! {Atomic(Integer), Missing}),
        left = Atomic(Integer),
        right = Missing,
    );
    test_union_schema!(
        union_any_x_is_any,
        expected = Any,
        left = Any,
        right = Atomic(Double),
    );
    test_union_schema!(
        union_unsat_x_is_x,
        expected = Atomic(Double),
        left = Atomic(Double),
        right = Unsat,
    );
    test_union_schema!(
        union_any_ofs_must_flatten,
        expected = AnyOf(set! {Atomic(Integer), Atomic(Double), Atomic(Decimal)}),
        left = AnyOf(set! {Atomic(Integer), Atomic(Double)}),
        right = AnyOf(set! {Atomic(Integer), Atomic(Decimal)}),
    );
    test_union_schema!(
        union_arrays_unions_items,
        expected = Array(Box::new(AnyOf(set! {Atomic(Integer), Atomic(Double)}))),
        left = Array(Box::new(Atomic(Integer))),
        right = Array(Box::new(Atomic(Double))),
    );
    // This case covers a pretty unusable schema. It's not worth trying
    // to provide more accurate types as the user won't be able to do anything
    // with the result anyway.
    test_union_schema!(
        union_arrays_of_arrays_of_multiple_types_is_any,
        expected = AnyOf(set! {Array(Box::new(Any))}),
        left = Array(Box::new(Atomic(Integer))),
        right = AnyOf(set! {Array(Box::new(Atomic(String))), Array(Box::new(Atomic(Double)))}),
    );
    test_union_schema!(
        document_union_required_empty_if_key_not_in_both,
        expected = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
              "b".to_string() => Atomic(Integer),
            },
            required: set! {},
            additional_properties: false,
            ..Default::default()
        }),
        left = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
            },
            required: set! {"a".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
        right = Document(Document {
            keys: map! {
              "b".to_string() => Atomic(Integer),
            },
            required: set! {"b".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
    );

    test_union_schema!(
        document_union_required_contains_key_in_both_documents,
        expected = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
              "b".to_string() => Atomic(Integer),
            },
            required: set! {"a".to_string(), "b".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
        left = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
              "b".to_string() => Atomic(Integer),
            },
            required: set! {"a".to_string(), "b".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
        right = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
              "b".to_string() => Atomic(Integer),
            },
            required: set! {"a".to_string(), "b".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
    );

    test_union_schema!(
        document_union_required_only_contains_required_union_from_both_documents,
        expected = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
              "b".to_string() => Atomic(Integer),
            },
            required: set! {"a".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
        left = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
              "b".to_string() => Atomic(Integer),
            },
            required: set! {"a".to_string(), "b".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
        right = Document(Document {
            keys: map! {
              "a".to_string() => Atomic(Double),
              "b".to_string() => Atomic(Integer),
            },
            required: set! {"a".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
    );

    test_union_schema!(
        union_documents_all_together,
        expected = Document(Document {
            keys: map! {
                "a".to_string() => Array(Box::new(
                    AnyOf(set! {
                        Atomic(Double),
                        Atomic(String),
                    })
                )),
                "b".to_string() => AnyOf(set! {
                    Atomic(Integer),
                    Atomic(String),
                }),
                "c".to_string() => Atomic(MaxKey),
            },
            required: set! {"a".to_string()},
            additional_properties: true,
            ..Default::default()
        }),
        left = Document(Document {
            keys: map! {
                "a".to_string() => Array(Box::new(Atomic(Double))),
                "b".to_string() => Atomic(Integer),
            },
            required: set! {"a".to_string(), "b".to_string()},
            additional_properties: false,
            ..Default::default()
        }),
        right = Document(Document {
            keys: map! {
                "a".to_string() => Array(Box::new(Atomic(String))),
                "b".to_string() => Atomic(String),
                "c".to_string() => Atomic(MaxKey),
            },
            required: set! {"a".to_string()},
            additional_properties: true,
            ..Default::default()
        }),
    );
}

mod to_bson {
    use crate::{
        map,
        schema::{Atomic, Document, Schema},
        set,
    };
    use bson::bson;
    macro_rules! test_schema_to_bson {
        ($func_name:ident, expected = $bson_doc:expr, input = $resultset_schema:expr) => {
            #[test]
            fn $func_name() {
                let b = bson::to_bson(
                    &crate::json_schema::Schema::try_from($resultset_schema).unwrap(),
                )
                .unwrap();
                assert_eq!($bson_doc, b);
            }
        };
    }
    test_schema_to_bson!(
        all_types_in_three_name_spaces,
        expected = bson! {{"bsonType": "object",
             "properties": {
                 "bar": {
                     "bsonType": "object",
                     "properties": {
                         "a": {},
                         "b": {"anyOf": []},
                         "c": {"bsonType":  "string"},
                         "e": {"bsonType":  "int"},
                         "f": {"bsonType":  "double"},
                         "g": {"bsonType":  "long"},
                         "h": {"bsonType":  "decimal"},
                         "i": {"bsonType":  "binData"},
                         "j": {"bsonType":  "objectId"},
                         "k": {"bsonType":  "bool"},
                         "l": {"bsonType":  "date"},
                         "m": {"bsonType":  "null"},
                         "n": {"bsonType":  "regex"},
                         "o": {"bsonType":  "dbPointer"},
                         "p": {"bsonType":  "javascript"},
                         "q": {"bsonType":  "symbol"},
                         "r": {"bsonType":  "javascriptWithScope"},
                         "s": {"bsonType":  "timestamp"},
                         "t": {"bsonType":  "minKey"},
                         "u": {"bsonType":  "maxKey"},
                         "v": {
                                 "bsonType": "array",
                                 "items": {
                                     "anyOf": [
                                         {"bsonType":  "null"},
                                         {"bsonType":  "string"},
                                     ]
                                 },
                         },
                         "w": {
                             "anyOf": [
                                 {"bsonType": "null"},
                                 {
                                     "bsonType": "object",
                                     "properties":  {},
                                     "additionalProperties": true,
                                 },
                             ]
                         },
                     },
                     "required": ["a", "b", "c"],
                     "additionalProperties": false,
                 },
             },
             "additionalProperties": false,
        }},
        input = Schema::Document(Document {
            keys: map! {
                "bar".into() => Schema::Document(Document {
                     keys: map!{
                         "a".into() => Schema::Any,
                         "b".into() => Schema::Unsat,
                         "c".into() => Schema::Atomic(Atomic::String),
                         "e".into() => Schema::Atomic(Atomic::Integer),
                         "f".into() => Schema::Atomic(Atomic::Double),
                         "g".into() => Schema::Atomic(Atomic::Long),
                         "h".into() => Schema::Atomic(Atomic::Decimal),
                         "i".into() => Schema::Atomic(Atomic::BinData),
                         "j".into() => Schema::Atomic(Atomic::ObjectId),
                         "k".into() => Schema::Atomic(Atomic::Boolean),
                         "l".into() => Schema::Atomic(Atomic::Date),
                         "m".into() => Schema::Atomic(Atomic::Null),
                         "n".into() => Schema::Atomic(Atomic::Regex),
                         "o".into() => Schema::Atomic(Atomic::DbPointer),
                         "p".into() => Schema::Atomic(Atomic::Javascript),
                         "q".into() => Schema::Atomic(Atomic::Symbol),
                         "r".into() => Schema::Atomic(Atomic::JavascriptWithScope),
                         "s".into() => Schema::Atomic(Atomic::Timestamp),
                         "t".into() => Schema::Atomic(Atomic::MinKey),
                         "u".into() => Schema::Atomic(Atomic::MaxKey),
                         "v".into() => Schema::Array(Box::new(Schema::AnyOf(set![
                                 Schema::Atomic(Atomic::String),
                                 Schema::Atomic(Atomic::Null)
                         ]))),
                         "w".into() => Schema::AnyOf(set![
                                 Schema::Document(Document {
                                     keys: map!{},
                                     required: set![],
                                     additional_properties: true,
                                     ..Default::default()
                                     }),
                                 Schema::Atomic(Atomic::Null),
                         ]),
                      },
                      required: set!["a".into(), "b".into(), "c".into()],
                      additional_properties: false,
                      ..Default::default()
                      }),
            },
            required: set![],
            additional_properties: false,
            ..Default::default()
        })
    );
}

// +-------------------+
// | JSON schema tests |
// +-------------------+

mod from_json {
    use crate::{
        json_schema::{self, BsonType, BsonTypeName, Items},
        map,
        schema::{self, Atomic::*, Document, Error, Schema::*},
        set,
    };
    macro_rules! test_from_json_schema {
        ($func_name:ident, schema_schema = $schema_schema:expr, json_schema = $json_schema:expr) => {
            #[test]
            fn $func_name() {
                let s = schema::Schema::try_from($json_schema);
                assert_eq!($schema_schema, s);
            }
        };
    }
    test_from_json_schema!(
        undefined_is_not_supported,
        schema_schema = Err(Error::UnsupportedBsonType("Error 1018: Unsupported BSON type: undefined. Consider updating the type of undefined fields to a non-deprecated BSON type, such as Null\n\tCaused by:\n\tDeprecated BSON type undefined".to_string())),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Undefined)),
            ..Default::default()
        }
    );
    test_from_json_schema!(
        convert_bson_single_to_atomic,
        schema_schema = Ok(Atomic(Integer)),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_empty_schema_to_any,
        schema_schema = Ok(Any),
        json_schema = json_schema::Schema::default()
    );

    test_from_json_schema!(
        with_no_bson_type_properties_extracted_into_any_of_document,
        schema_schema = Ok(AnyOf(set![
            Document(Document {
                keys: map! {"a".to_string() => Atomic(String)},
                required: set![],
                additional_properties: true,
                ..Default::default()
            }),
            schema::ANY_ARRAY.clone(),
            Atomic(String),
            Atomic(Integer),
            Atomic(Long),
            Atomic(Double),
            Atomic(Decimal),
            Atomic(BinData),
            Atomic(ObjectId),
            Atomic(Boolean),
            Atomic(Date),
            Atomic(Null),
            Atomic(Regex),
            Atomic(DbPointer),
            Atomic(Javascript),
            Atomic(Symbol),
            Atomic(JavascriptWithScope),
            Atomic(Timestamp),
            Atomic(MinKey),
            Atomic(MaxKey),
        ])),
        json_schema = json_schema::Schema {
            properties: Some(map! {"a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::String)),
                ..Default::default()
            }}),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        with_no_bson_type_items_extracted_into_any_of_array,
        schema_schema = Ok(AnyOf(set![
            Array(Atomic(String).into()),
            schema::ANY_DOCUMENT.clone(),
            Atomic(String),
            Atomic(Integer),
            Atomic(Long),
            Atomic(Double),
            Atomic(Decimal),
            Atomic(BinData),
            Atomic(ObjectId),
            Atomic(Boolean),
            Atomic(Date),
            Atomic(Null),
            Atomic(Regex),
            Atomic(DbPointer),
            Atomic(Javascript),
            Atomic(Symbol),
            Atomic(JavascriptWithScope),
            Atomic(Timestamp),
            Atomic(MinKey),
            Atomic(MaxKey),
        ])),
        json_schema = json_schema::Schema {
            items: Some(json_schema::Items::Single(
                json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::String)),
                    ..Default::default()
                }
                .into()
            )),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_bson_multiple_to_any_of,
        schema_schema = Ok(AnyOf(set![Atomic(Integer), Atomic(Null)])),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec![
                BsonTypeName::Int,
                BsonTypeName::Null
            ])),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_one_of_to_any_of,
        schema_schema = Ok(AnyOf(set![Atomic(Integer), Atomic(Null)])),
        json_schema = json_schema::Schema {
            one_of: Some(vec![
                json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                    ..Default::default()
                },
                json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Null)),
                    ..Default::default()
                }
            ]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        one_of_invalid_extra_fields,
        schema_schema = Err(Error::InvalidCombinationOfFields()),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            one_of: Some(vec![json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Null)),
                ..Default::default()
            }]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_any_of_to_any_of,
        schema_schema = Ok(AnyOf(set![Atomic(Integer), Atomic(Null)])),
        json_schema = json_schema::Schema {
            any_of: Some(vec![
                json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                    ..Default::default()
                },
                json_schema::Schema {
                    bson_type: Some(BsonType::Single(BsonTypeName::Null)),
                    ..Default::default()
                }
            ]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_properties_to_document,
        schema_schema = Ok(Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        })),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        document_bson_type_not_object,
        schema_schema = Ok(Atomic(Integer)),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        document_properties_not_set,
        schema_schema = Ok(Document(Document {
            keys: map![],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        })),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        document_additional_properties_not_set,
        schema_schema = Ok(Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        })),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Object)),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_array_to_any_of,
        schema_schema = Ok(Array(Box::new(Atomic(Integer)))),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Array)),
            items: Some(Items::Single(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }))),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        items_set_bson_type_not_array,
        schema_schema = Ok(Atomic(Integer)),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec![BsonTypeName::Int,])),
            items: Some(Items::Single(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }))),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        bson_type_array_set_missing_items_field,
        schema_schema = Ok(Array(Box::new(Any))),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Array)),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        bson_type_array_multiple_items_becomes_any,
        schema_schema = Ok(Array(Box::new(Any))),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Array)),
            items: Some(Items::Multiple(vec![json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }])),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        convert_array_and_document_fields,
        schema_schema = Ok(AnyOf(set![
            Array(Box::new(Atomic(Integer))),
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(Integer),
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: true,
                ..Default::default()
            })
        ])),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec![
                BsonTypeName::Array,
                BsonTypeName::Object
            ])),
            properties: Some(map! { "a".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }, "b".to_string() => json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }}),
            required: Some(vec!["a".to_string()]),
            additional_properties: Some(true),
            items: Some(Items::Single(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }))),
            ..Default::default()
        }
    );

    test_from_json_schema!(
        bson_type_object_set_missing_document_fields,
        schema_schema = Ok(AnyOf(set![
            Array(Box::new(Atomic(Integer))),
            Document(Document {
                keys: map![],
                required: set![],
                additional_properties: true,
                ..Default::default()
            })
        ])),
        json_schema = json_schema::Schema {
            bson_type: Some(BsonType::Multiple(vec![
                BsonTypeName::Array,
                BsonTypeName::Object
            ])),
            items: Some(Items::Single(Box::new(json_schema::Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }))),
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
        ($func_name:ident, expected = $expected:expr, _self = $self:expr, other = $other:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let res = $self.satisfies(&$other);
                assert_eq!($expected, res)
            }
        };
    }

    test_satisfies!(
        any_must_satisfy_any,
        expected = Must,
        _self = Any,
        other = Any
    );
    test_satisfies!(
        missing_must_satisfy_any,
        expected = Must,
        _self = Missing,
        other = Any
    );
    test_satisfies!(
        any_of_empty_must_satisfy_atomic,
        expected = Must,
        _self = AnyOf(set![]),
        other = Atomic(Integer)
    );
    test_satisfies!(
        any_of_empty_must_satisfy_any_of_empty,
        expected = Must,
        _self = AnyOf(set![]),
        other = AnyOf(set![]),
    );
    test_satisfies!(
        any_of_empty_must_satisfy_missing,
        expected = Must,
        _self = AnyOf(set![]),
        other = Missing,
    );
    test_satisfies!(
        missing_must_satisfy_missing,
        expected = Must,
        _self = Missing,
        other = Missing
    );
    test_satisfies!(
        missing_must_satisfy_any_of,
        expected = Must,
        _self = Missing,
        other = AnyOf(set![Missing])
    );
    test_satisfies!(
        any_of_missing_may_satisfy_missing,
        expected = May,
        _self = AnyOf(set![Atomic(Integer), Missing, Atomic(String)]),
        other = Missing
    );
    test_satisfies!(
        missing_must_not_satisfy_atomic,
        expected = Not,
        _self = Missing,
        other = Atomic(String)
    );
    test_satisfies!(
        missing_must_not_satisfy_array,
        expected = Not,
        _self = Missing,
        other = Array(Box::new(Any)),
    );
    test_satisfies!(
        missing_must_not_satisfy_document,
        expected = Not,
        _self = Missing,
        other = Document(Document {
            keys: map![],
            required: set![],
            additional_properties: true,
            ..Default::default()
        })
    );
    test_satisfies!(
        missing_must_not_satisfy_any_of,
        expected = Not,
        _self = Missing,
        other = AnyOf(set![Atomic(String), Atomic(Integer)])
    );
    test_satisfies!(
        atomic_must_satisfy_any,
        expected = Must,
        _self = Atomic(String),
        other = Any
    );
    test_satisfies!(
        any_may_satisfy_atomic,
        expected = May,
        _self = Any,
        other = Atomic(String)
    );
    test_satisfies!(
        array_of_any_does_not_satisfy_atomic,
        expected = Not,
        _self = Array(Box::new(Any)),
        other = Atomic(Integer),
    );
    test_satisfies!(
        missing_does_not_satisfy_atomic,
        expected = Not,
        _self = Missing,
        other = Atomic(String),
    );
    test_satisfies!(
        any_of_must_satisfy_any,
        expected = Must,
        _self = AnyOf(set![Atomic(String), Atomic(Integer)]),
        other = Any,
    );
    test_satisfies!(
        any_of_must_satisfy_when_any_of_contains_any,
        expected = Must,
        _self = AnyOf(set![Atomic(String), Atomic(Integer)]),
        other = AnyOf(set![Atomic(String), Atomic(Integer), Any]),
    );
    test_satisfies!(
        array_of_string_must_satisfy_any_of_array_of_int_or_array_of_string,
        expected = Must,
        _self = Array(Box::new(Atomic(String))),
        other = AnyOf(set![
            Array(Box::new(Atomic(String))),
            Array(Box::new(Atomic(Integer)))
        ]),
    );
    test_satisfies!(
        array_of_string_or_int_may_satisfy_any_of_array_of_int_or_array_of_string,
        expected = May,
        _self = Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer),]))),
        other = AnyOf(set![
            Array(Box::new(Atomic(String))),
            Array(Box::new(Atomic(Integer)))
        ]),
    );
    test_satisfies!(
        array_of_string_or_int_must_satisfy_array_of_string_or_int,
        expected = Must,
        _self = Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer),]))),
        other = Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer),]))),
    );
    test_satisfies!(
        document_must_satify_same_document,
        expected = Must,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_may_satify_with_more_permissive_key_schema,
        expected = May,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_must_not_satify_with_incompatable_key_schema,
        expected = Not,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_may_satify_with_fewer_required_keys,
        expected = May,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_must_not_satify_with_missing_required_key,
        expected = Not,
        _self = Document(Document {
            keys: map![
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_may_satify_with_missing_required_key,
        expected = May,
        _self = Document(Document {
            keys: map![
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: true,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_must_satify_with_more_required_keys,
        expected = Must,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_may_satify_due_to_possible_extra_keys,
        expected = May,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: true,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set![],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_satisfies!(
        document_satifies_multiple_any_of_results_in_must_satisfy,
        expected = Must,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map![
                    "b".to_string() => Atomic(Integer),
                ],
                required: set![],
                additional_properties: false,
                ..Default::default()
            }),
        ]),
    );
    test_satisfies!(
        document_satifies_any_of_any_of_results_must_satisfy,
        expected = Must,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map![
                    "e".to_string() => Atomic(Integer),
                ],
                required: set![],
                additional_properties: false,
                ..Default::default()
            }),
        ]),
    );
    test_satisfies!(
        document_may_satisfy_when_key_schema_may_satisfy,
        expected = May,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = Document(Document {
            keys: map![
                "a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_satisfies!(
        array_may_satisfy_when_array_item_schema_may_satisfy,
        expected = May,
        _self = Array(Box::new(Any)),
        other = Array(Box::new(Atomic(Integer))),
    );
    test_satisfies!(
        array_may_satisfy_when_array_item_schema_may_satisfy_multiple_any_of_array,
        expected = May,
        _self = Array(Box::new(Any)),
        other = AnyOf(set![
            Array(Box::new(Atomic(Integer))),
            Array(Box::new(Atomic(String))),
        ]),
    );
    test_satisfies!(
        array_may_satisfy_when_array_item_schema_may_satisfy_multiple_array_any_of,
        expected = May,
        _self = Array(Box::new(Any)),
        other = Array(Box::new(AnyOf(set![Atomic(Integer), Atomic(Double),]),)),
    );
    test_satisfies!(
        array_of_missing_does_not_satisfy_array_of_atomic,
        expected = Not,
        _self = Array(Box::new(Missing)),
        other = Array(Box::new(Atomic(Integer))),
    );
}

mod has_overlaping_keys_with {
    use crate::{
        map,
        schema::{Atomic, Document, Satisfaction, Schema, ANY_DOCUMENT, EMPTY_DOCUMENT},
        set,
    };
    macro_rules! test_has_overlapping_keys_with {
        ($func_name:ident, expected = $expected:expr, schema1 = $schema1:expr, schema2 = $schema2:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let out = $schema1.has_overlapping_keys_with($schema2);
                assert_eq!($expected, out);
            }
        };
    }

    test_has_overlapping_keys_with!(
        any_may_overlap_any_document,
        expected = Satisfaction::May,
        schema1 = &Schema::Any,
        schema2 = &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        any_overlap_may_any_document_symmetric,
        expected = Satisfaction::May,
        schema1 = &ANY_DOCUMENT,
        schema2 = &Schema::Any,
    );
    test_has_overlapping_keys_with!(
        atomic_has_no_keys_to_overlap,
        expected = Satisfaction::Not,
        schema1 = Schema::Atomic(Atomic::Integer),
        schema2 = &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        atomic_has_no_keys_to_overlap_symmetric,
        expected = Satisfaction::Not,
        schema1 = &ANY_DOCUMENT,
        schema2 = &Schema::Atomic(Atomic::Integer),
    );
    test_has_overlapping_keys_with!(
        any_document_may_overlap_keys_with_any_document,
        expected = Satisfaction::May,
        schema1 = &ANY_DOCUMENT,
        schema2 = &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        explicit_document_may_overlap_keys_with_any_document,
        expected = Satisfaction::May,
        schema1 = Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
            ..Default::default()
        }),
        schema2 = &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        any_document_does_not_overlap_with_empty_document,
        expected = Satisfaction::Not,
        schema1 = &EMPTY_DOCUMENT,
        schema2 = &ANY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        explicit_document_may_overlap_keys_with_any_document_symmetric,
        expected = Satisfaction::May,
        schema1 = &ANY_DOCUMENT,
        schema2 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_has_overlapping_keys_with!(
        any_document_does_not_overlap_with_empty_document_symmetric,
        expected = Satisfaction::Not,
        schema1 = &ANY_DOCUMENT,
        schema2 = &EMPTY_DOCUMENT,
    );
    test_has_overlapping_keys_with!(
        two_explicit_documents_without_required_keys_may_overlap,
        expected = Satisfaction::May,
        schema1 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
            ..Default::default()
        }),
        schema2 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {},
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_has_overlapping_keys_with!(
        two_explicit_documents_with_required_keys_may_overlap,
        expected = Satisfaction::May,
        schema1 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
        schema2 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"b".into()},
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_has_overlapping_keys_with!(
        two_explicit_documents_with_required_keys_must_overlap,
        expected = Satisfaction::Must,
        schema1 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
        schema2 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer)},
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_may_overlap,
        expected = Satisfaction::May,
        schema1 = &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"b".into()},
                additional_properties: false,
                ..Default::default()
            }),
        ]),
        schema2 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"c".into()},
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_must_overlap,
        expected = Satisfaction::Must,
        schema1 = &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
        ]),
        schema2 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_may_overlap_symmetric,
        expected = Satisfaction::May,
        schema1 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"c".into()},
            additional_properties: false,
            ..Default::default()
        }),
        schema2 = &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"b".into()},
                additional_properties: false,
                ..Default::default()
            }),
        ]),
    );
    test_has_overlapping_keys_with!(
        any_of_documents_with_required_keys_must_overlap_symmetric,
        expected = Satisfaction::Must,
        schema1 = &Schema::Document(Document {
            keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
            "b".into() => Schema::Atomic(Atomic::Integer),
            "c".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
        schema2 = &Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! { "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer)},
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
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
        ($func_name:ident, expected = $expected:expr, schema1 = $schema1:expr, schema2 = $schema2:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let out = $schema1.document_union($schema2);
                assert_eq!($expected, out);
            }
        };
    }

    test_document_union!(
        schema_does_not_satisfy_document_results_in_any,
        expected = EMPTY_DOCUMENT.clone(),
        schema1 = Schema::Atomic(Atomic::Integer),
        schema2 = ANY_DOCUMENT.clone(),
    );
    test_document_union!(
        schema_does_not_satisfy_document_results_in_any_symmetric,
        expected = EMPTY_DOCUMENT.clone(),
        schema1 = ANY_DOCUMENT.clone(),
        schema2 = Schema::Atomic(Atomic::Integer),
    );
    test_document_union!(
        document_union_of_two_documents_will_document_union_keys_and_intersect_required_wo_additional_properties,
        expected = Schema::Document(Document {
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
            ..Default::default()
            }),
        schema1 = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Decimal),
                "b".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
            }),
        schema2 = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "b".into()},
            additional_properties: false,
            ..Default::default()
            }),
    );
    test_document_union!(
        document_union_of_two_documents_will_document_union_keys_and_intersect_required_wo_additional_properties_symmetric,
        expected = Schema::Document(Document {
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
            ..Default::default()
            }),
        schema1 = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "b".into()},
            additional_properties: false,
            ..Default::default()
            }),
        schema2 = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Decimal),
                "b".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
            }),
    );
    test_document_union!(
        document_union_of_doc_with_empty_doc_is_doc_with_no_required,
        expected = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            ..Default::default()
        }),
        schema1 = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
            ..Default::default()
        }),
        schema2 = EMPTY_DOCUMENT.clone(),
    );
    test_document_union!(
        document_union_of_doc_with_empty_doc_is_doc_with_no_required_symmetric,
        expected = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            ..Default::default()
        }),
        schema1 = EMPTY_DOCUMENT.clone(),
        schema2 = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into(), "c".into()},
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_document_union!(
        document_union_of_any_of_recursively_applies_document_union,
        expected = Schema::Document(Document {
            keys: map! {
                "a".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Decimal),
                ]),
                "b".into() => Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Double),
                ]),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Double),
            },
            required: set! {"a".into()},
            additional_properties: false,
            ..Default::default()
        }),
        schema1 = Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"a".into(), "b".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Integer),
                    "b".into() => Schema::Atomic(Atomic::Integer),
                    "c".into() => Schema::Atomic(Atomic::Integer),
                },
                required: set! {"a".into(), "b".into()},
                additional_properties: false,
                ..Default::default()
            })
        ]),
        schema2 = Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Decimal),
                    "b".into() => Schema::Atomic(Atomic::Double),
                },
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
            }),
            Schema::Document(Document {
                keys: map! {
                    "a".into() => Schema::Atomic(Atomic::Decimal),
                    "d".into() => Schema::Atomic(Atomic::Double),
                },
                required: set! {"a".into()},
                additional_properties: false,
                ..Default::default()
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
        ($func_name:ident, expected = $expected:expr, _self = $self:expr, other = $other:expr $(,)?) => {
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
        expected = Not,
        _self = ANY_ARRAY,
        other = ANY_ARRAY,
    );
    test_is_comparable_with!(
        document_not_comparable_with_another_document,
        expected = Not,
        _self = ANY_DOCUMENT,
        other = ANY_DOCUMENT,
    );
    test_is_comparable_with!(
        array_not_comparable_with_document,
        expected = Not,
        _self = ANY_ARRAY,
        other = ANY_DOCUMENT,
    );

    test_is_comparable_with!(
        array_not_comparable_with_any,
        expected = Not,
        _self = ANY_ARRAY,
        other = Any,
    );
    test_is_comparable_with!(
        array_not_comparable_with_another_type,
        expected = Not,
        _self = ANY_ARRAY,
        other = Atomic(Integer),
    );
    test_is_comparable_with!(
        array_not_comparable_with_null,
        expected = Not,
        _self = ANY_ARRAY,
        other = Atomic(Null),
    );
    test_is_comparable_with!(
        array_not_comparable_with_missing,
        expected = Not,
        _self = ANY_ARRAY,
        other = Missing,
    );
    test_is_comparable_with!(
        array_not_comparable_with_unsat,
        expected = Not,
        _self = ANY_ARRAY,
        other = Unsat,
    );

    test_is_comparable_with!(
        document_not_comparable_with_any_type,
        expected = Not,
        _self = ANY_DOCUMENT,
        other = Any,
    );
    test_is_comparable_with!(
        document_not_comparable_with_a_type,
        expected = Not,
        _self = ANY_DOCUMENT,
        other = Atomic(Integer),
    );
    test_is_comparable_with!(
        document_not_comparable_with_null,
        expected = Not,
        _self = ANY_DOCUMENT,
        other = Atomic(Null),
    );
    test_is_comparable_with!(
        document_not_comparable_with_missing,
        expected = Not,
        _self = ANY_DOCUMENT,
        other = Missing,
    );
    test_is_comparable_with!(
        document_not_comparable_with_unsat,
        expected = Not,
        _self = ANY_DOCUMENT,
        other = Unsat,
    );

    // Any comparison tests.
    test_is_comparable_with!(
        any_type_may_be_comparable_with_another_type,
        expected = May,
        _self = Any,
        other = Atomic(Integer),
    );
    test_is_comparable_with!(
        any_type_may_be_comparable_with_null,
        expected = May,
        _self = Any,
        other = Atomic(Null),
    );
    test_is_comparable_with!(
        any_type_may_be_comparable_with_missing,
        expected = May,
        _self = Any,
        other = Missing,
    );
    test_is_comparable_with!(
        any_type_may_be_comparable_with_unsat,
        expected = May,
        _self = Any,
        other = Unsat,
    );
    // Missing comparability tests.
    test_is_comparable_with!(
        missing_must_be_comparable_with_missing,
        expected = Must,
        _self = Missing,
        other = Missing,
    );
    test_is_comparable_with!(
        missing_must_be_comparable_with_another_type,
        expected = Must,
        _self = Missing,
        other = Atomic(Integer),
    );
    test_is_comparable_with!(
        missing_must_be_comparable_with_null,
        expected = Must,
        _self = Missing,
        other = Atomic(Null),
    );
    test_is_comparable_with!(
        missing_must_be_comparable_with_unsat,
        expected = Must,
        _self = Missing,
        other = Unsat,
    );

    // Unsat comparability tests.
    test_is_comparable_with!(
        unsat_must_be_comparable_with_unsat,
        expected = Must,
        _self = Unsat,
        other = Unsat,
    );
    test_is_comparable_with!(
        unsat_must_be_comparable_with_another_type,
        expected = Must,
        _self = Unsat,
        other = Atomic(Integer),
    );
    test_is_comparable_with!(
        unsat_must_be_comparable_with_null,
        expected = Must,
        _self = Unsat,
        other = Atomic(Null),
    );

    // Atomic comparability tests.
    test_is_comparable_with!(
        null_must_be_comparable_with_null,
        expected = Must,
        _self = Atomic(Null),
        other = Atomic(Null),
    );
    test_is_comparable_with!(
        null_must_be_comparable_with_atomic_numeric,
        expected = Must,
        _self = Atomic(Null),
        other = Atomic(Integer),
    );
    test_is_comparable_with!(
        atomic_numeric_must_be_comparable_with_same_atomic_numeric,
        expected = Must,
        _self = Atomic(Integer),
        other = Atomic(Integer),
    );
    test_is_comparable_with!(
        atomic_numeric_must_be_comparable_with_different_atomic_numeric,
        expected = Must,
        _self = Atomic(Integer),
        other = Atomic(Double),
    );
    test_is_comparable_with!(
        non_numeric_atomic_must_be_comparable_with_same_non_numeric_atomic,
        expected = Must,
        _self = Atomic(String),
        other = Atomic(String),
    );
    test_is_comparable_with!(
        atomic_not_comparable_with_different_atomic,
        expected = Not,
        _self = Atomic(String),
        other = Atomic(Integer),
    );

    // AnyOf comparability tests (numeric).
    test_is_comparable_with!(
        numeric_atomic_must_be_comparable_with_a_set_of_numerics,
        expected = Must,
        _self = Atomic(Integer),
        other = AnyOf(set![Atomic(Integer), Atomic(Long)]),
    );
    test_is_comparable_with!(
        a_set_of_numerics_must_be_comparable_with_a_disjoint_set_of_numerics,
        expected = Must,
        _self = AnyOf(set![Atomic(Integer), Atomic(Long)]),
        other = AnyOf(set![Atomic(Double), Atomic(Decimal)]),
    );
    test_is_comparable_with!(
        numeric_must_be_comparable_with_different_numeric_or_null,
        expected = Must,
        _self = Atomic(Integer),
        other = AnyOf(set![Atomic(Long), Atomic(Null)]),
    );
    test_is_comparable_with!(
        numeric_or_null_must_be_comparable_with_different_numeric_or_null,
        expected = Must,
        _self = AnyOf(set![Atomic(Integer), Atomic(Null)]),
        other = AnyOf(set![Atomic(Long), Atomic(Null)]),
    );
    test_is_comparable_with!(
        numeric_atomic_may_be_comparable_with_potentially_same_numeric,
        expected = May,
        _self = Atomic(Integer),
        other = AnyOf(set![Atomic(Integer), Atomic(String)]),
    );
    test_is_comparable_with!(
        potential_numeric_may_be_comparable_with_potentially_same_numeric,
        expected = May,
        _self = AnyOf(set![Atomic(Integer), Atomic(String)]),
        other = AnyOf(set![Atomic(Integer), Atomic(String)]),
    );
    test_is_comparable_with!(
        potential_numeric_may_be_comparable_with_potentially_different_numeric,
        expected = May,
        _self = AnyOf(set![Atomic(Integer), Atomic(String)]),
        other = AnyOf(set![Atomic(Double), Atomic(String)]),
    );

    // AnyOf comparability tests (non-numeric).
    test_is_comparable_with!(
        atomic_must_be_comparable_with_same_atomic_or_null,
        expected = Must,
        _self = Atomic(String),
        other = AnyOf(set![Atomic(String), Atomic(Null)]),
    );
    test_is_comparable_with!(
        atomic_or_null_must_be_comparable_with_same_atomic_or_null,
        expected = Must,
        _self = AnyOf(set![Atomic(String), Atomic(Null)]),
        other = AnyOf(set![Atomic(String), Atomic(Null)]),
    );
    test_is_comparable_with!(
        atomic_may_be_comparable_with_potentially_same_atomic,
        expected = May,
        _self = Atomic(String),
        other = AnyOf(set![Atomic(String), Atomic(Integer)]),
    );
    test_is_comparable_with!(
        atomic_or_null_may_be_comparable_with_different_atomic_or_null,
        expected = May,
        _self = AnyOf(set![Atomic(String), Atomic(Null)]),
        other = AnyOf(set![Atomic(Integer), Atomic(Null)]),
    );
    test_is_comparable_with!(
        some_atomic_may_be_comparable_with_potentially_same_atomic,
        expected = May,
        _self = AnyOf(set![Atomic(String), Atomic(Boolean)]),
        other = AnyOf(set![Atomic(String), Atomic(Integer)]),
    );
    test_is_comparable_with!(
        a_set_of_atomics_not_comparable_with_disjoint_set_of_atomics,
        expected = Not,
        _self = AnyOf(set![Atomic(String), Atomic(Boolean)]),
        other = AnyOf(set![Atomic(Date), Atomic(Integer)]),
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
        ($func_name:ident, expected = $expected:expr, _self = $self:expr, other = $other:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let res = $self.contains_field($other);
                assert_eq!($expected, res)
            }
        };
    }

    test_contains_field!(
        any_may_contain_field,
        expected = May,
        _self = Any,
        other = "a"
    );
    test_contains_field!(
        missing_does_not_contain_field,
        expected = Not,
        _self = Missing,
        other = "a"
    );
    test_contains_field!(
        document_must_contain_field,
        expected = Must,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = "a",
    );
    test_contains_field!(
        document_may_contain_field,
        expected = May,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = "b",
    );
    test_contains_field!(
        document_may_contain_field_due_to_additional_properties,
        expected = May,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        }),
        other = "foo",
    );
    test_contains_field!(
        document_must_not_contain_field,
        expected = Not,
        _self = Document(Document {
            keys: map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Integer),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        other = "foo",
    );
    test_contains_field!(
        atomic_must_not_contain_field,
        expected = Not,
        _self = Atomic(String),
        other = "foo",
    );
    test_contains_field!(
        any_of_document_and_atomic_may_not_contain_field,
        expected = Not,
        _self = AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Atomic(String),
        ]),
        other = "c",
    );
    test_contains_field!(
        any_of_document_and_atomic_may_contain_field,
        expected = May,
        _self = AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Atomic(String),
        ]),
        other = "b",
    );
    test_contains_field!(
        any_of_document_and_document_must_contain_field,
        expected = Must,
        _self = AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(Integer),
                ],
                required: set!["b".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map![
                    "a".to_string() => Any,
                    "b".to_string() => Atomic(String),
                ],
                required: set!["b".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
        ]),
        other = "b",
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
        ($func_name:ident, expected = $expected:expr, input = $input:expr $(,)?) => {
            #[test]
            fn $func_name() {
                let res = schema::Schema::simplify(&$input);
                assert_eq!($expected, res)
            }
        };
    }

    test_simplify!(contains_empty_vec, expected = Unsat, input = AnyOf(set![]));
    test_simplify!(
        remove_any_of_duplicates,
        expected = AnyOf(set![Atomic(String), Atomic(Integer)]),
        input = AnyOf(set![Atomic(Integer), Atomic(Integer), Atomic(String)])
    );
    test_simplify!(
        remove_any_of_duplicates_not_consecutive,
        expected = AnyOf(set![Atomic(String), Atomic(Integer)]),
        input = AnyOf(set![
            Atomic(Integer),
            Atomic(Integer),
            Atomic(String),
            Atomic(Integer)
        ])
    );
    test_simplify!(flatten_any_is_flat, expected = Any, input = Any);
    test_simplify!(
        flatten_any_of_one_schema,
        expected = Atomic(Integer),
        input = AnyOf(set![Atomic(Integer)])
    );
    test_simplify!(
        flatten_any_of_any_schema,
        expected = Any,
        input = AnyOf(set!(Any, Missing))
    );
    test_simplify!(
        flatten_any_of_any_of,
        expected = AnyOf(set![Missing, Atomic(String), Atomic(Integer)]),
        input = AnyOf(set![AnyOf(set![Missing, Atomic(String)]), Atomic(Integer)]),
    );
    test_simplify!(
        flatten_any_of_and_remove_duplicates,
        expected = AnyOf(set![Atomic(String), Atomic(Integer), Atomic(Null)]),
        input = AnyOf(set![
            AnyOf(set![Atomic(Integer), Atomic(String)]),
            AnyOf(set![Atomic(Integer), Atomic(Null)])
        ])
    );
    test_simplify!(
        flatten_any_of_containing_array,
        expected = Array(Box::new(AnyOf(set![Atomic(String), Atomic(Integer)]))),
        input = AnyOf(set![Array(Box::new(AnyOf(set![
            Atomic(Integer),
            Atomic(String)
        ])))])
    );
    test_simplify!(
        flatten_any_of_and_return_single_element,
        expected = Atomic(Integer),
        input = AnyOf(set![Atomic(Integer), Atomic(Integer)])
    );
    test_simplify!(
        array,
        expected = Array(Box::new(AnyOf(set![
            Missing,
            Atomic(String),
            Atomic(Integer)
        ]))),
        input = Array(Box::new(AnyOf(set![
            AnyOf(set![Missing, Atomic(String)]),
            Atomic(Integer)
        ])))
    );
    test_simplify!(
        document,
        expected = Document(Document {
            keys: map![
                "a".to_string() => AnyOf(set![
                    Atomic(Null),
                    Atomic(String),
                    Atomic(Integer)
            ])
                ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        }),
        input = Document(Document {
            keys: map![
                "a".to_string() => AnyOf(set![
                AnyOf(set![Atomic(Null), Atomic(String)]),
                Atomic(Integer)
            ]),
                            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        })
    );
    test_simplify!(
        missing_in_documents,
        expected = Document(Document {
            keys: map![
                "a".to_string() => AnyOf(set![
                Atomic(String),
                Atomic(Integer)
            ]),
            "b".to_string() => Atomic(String),
                    "d".to_string() => Document(
                        Document {
                            keys: map![
                                "ia".to_string() => AnyOf(set![Atomic(String)]),
                            ],
                            required: set![],
                            additional_properties: false,
                            ..Default::default()
                            }
                    ),
                ],
            required: set!["b".to_string(), "d".to_string()],
            additional_properties: true,
            ..Default::default()
        }),
        input = Document(Document {
            keys: map![
                    "a".to_string() => AnyOf(set![
                        AnyOf(set![Missing, Atomic(String)]),
                        Atomic(Integer)
                    ]),
                    "b".to_string() => Atomic(String),
                    "c".to_string() => Missing,
                    "d".to_string() => Document(
                        Document {
                            keys: map![
                                "ia".to_string() => AnyOf(set![Missing, Atomic(String)]),
                            ],
                            required: set!["ia".to_string()],
                            additional_properties: false,
                            ..Default::default()
                            }
                    ),
            ],
            required: set![
                "a".to_string(),
                "b".to_string(),
                "c".to_string(),
                "d".to_string()
            ],
            additional_properties: true,
            ..Default::default()
        })
    );
}

mod get_single_field_name_and_schema {
    use crate::{
        map,
        schema::{Atomic::Integer, Atomic::String, Document, Schema::*},
        set,
    };
    macro_rules! test_get_single_field_name_and_schema {
        ($func_name:ident, expected = $expected:expr, schema = $schema:expr $(,)?) => {
            #[test]
            fn $func_name() {
                assert_eq!($expected, $schema.get_single_field_name_and_schema());
            }
        };
    }

    test_get_single_field_name_and_schema!(any, expected = None, schema = &Any);
    test_get_single_field_name_and_schema!(unsat, expected = None, schema = &Unsat);
    test_get_single_field_name_and_schema!(missing, expected = None, schema = &Missing);
    test_get_single_field_name_and_schema!(atomic, expected = None, schema = &Atomic(String));
    test_get_single_field_name_and_schema!(
        any_of_non_document,
        expected = None,
        schema = &AnyOf(set![Atomic(String)])
    );
    test_get_single_field_name_and_schema!(array, expected = None, schema = &Array(Box::new(Any)));
    test_get_single_field_name_and_schema!(empty_any_of, expected = None, schema = &AnyOf(set![]));
    test_get_single_field_name_and_schema!(
        empty_doc,
        expected = None,
        schema = &Document(Document {
            keys: map![],
            required: set![],
            additional_properties: false,
            ..Default::default()
        })
    );
    test_get_single_field_name_and_schema!(
        single_doc_no_required_keys,
        expected = Some(("a", Atomic(String))),
        schema = &Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
            ],
            required: set![],
            additional_properties: false,
            ..Default::default()
        })
    );
    test_get_single_field_name_and_schema!(
        single_doc_empty,
        expected = None,
        schema = &Document(Document {
            keys: map![],
            required: set![],
            additional_properties: false,
            ..Default::default()
        })
    );
    test_get_single_field_name_and_schema!(
        single_doc_multiple_required_keys,
        expected = None,
        schema = &Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
                "b".to_string() => Atomic(String),
            ],
            required: set!["a".to_string(), "b".to_string()],
            additional_properties: false,
            ..Default::default()
        })
    );
    test_get_single_field_name_and_schema!(
        single_doc_no_additional_properties,
        expected = Some(("a", Atomic(String))),
        schema = &Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        })
    );
    test_get_single_field_name_and_schema!(
        single_doc_with_additional_properties,
        expected = None,
        schema = &Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
            ],
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        })
    );
    test_get_single_field_name_and_schema!(
        possible_extra_keys,
        expected = None,
        schema = &Document(Document {
            keys: map![
                "a".to_string() => Atomic(String),
                "b".to_string() => Atomic(String),
            ],
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        })
    );
    test_get_single_field_name_and_schema!(
        two_docs_one_empty,
        expected = None,
        schema = &AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(String),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map![],
                required: set![],
                additional_properties: false,
                ..Default::default()
            })
        ])
    );
    test_get_single_field_name_and_schema!(
        two_docs_one_required_field_per_doc,
        expected = None,
        schema = &AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(String),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map![
                    "b".to_string() => Atomic(String),
                ],
                required: set!["b".to_string()],
                additional_properties: false,
                ..Default::default()
            })
        ])
    );
    test_get_single_field_name_and_schema!(
        duplicate_single_field_docs,
        expected = Some(("a", Atomic(String))),
        schema = &AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(String),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(String),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            })
        ])
    );
    test_get_single_field_name_and_schema!(
        duplicate_single_field_docs_with_diff_schema,
        expected = Some(("a", AnyOf(set![Atomic(String), Atomic(Integer)]))),
        schema = &AnyOf(set![
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(String),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(Integer),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            })
        ])
    );
    test_get_single_field_name_and_schema!(
        any_of_single_field_doc_and_unsat,
        expected = Some(("a", Atomic(String))),
        schema = &AnyOf(set![
            Unsat,
            Document(Document {
                keys: map![
                    "a".to_string() => Atomic(String),
                ],
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            })
        ])
    );
}

mod subtract_nullish {
    use crate::{
        schema::{Atomic::*, Document, Schema, Schema::*},
        set,
    };
    macro_rules! test_subtract_nullish {
        ($func_name:ident, expected = $expected:expr, _self = $self:expr) => {
            #[test]
            fn $func_name() {
                let res = $self.subtract_nullish();
                assert_eq!($expected, Schema::simplify(&res))
            }
        };
    }

    test_subtract_nullish!(
        remove_nullish_from_integer_or_nullish,
        expected = Atomic(Integer),
        _self = AnyOf(set![Atomic(Integer), Atomic(Null), Missing])
    );
    test_subtract_nullish!(
        remove_null_from_integer_or_null,
        expected = Atomic(Integer),
        _self = AnyOf(set![Atomic(Integer), Atomic(Null)])
    );
    test_subtract_nullish!(
        subtracting_nullish_from_null_yields_unsat,
        expected = Unsat,
        _self = Atomic(Null)
    );
    test_subtract_nullish!(
        remove_null_from_any,
        expected = AnyOf(set![
            Atomic(String),
            Atomic(Integer),
            Atomic(Long),
            Atomic(Double),
            Atomic(Decimal),
            Atomic(BinData),
            Atomic(ObjectId),
            Atomic(Boolean),
            Atomic(Date),
            // no Null
            Atomic(Regex),
            Atomic(DbPointer),
            Atomic(Javascript),
            Atomic(Symbol),
            Atomic(JavascriptWithScope),
            Atomic(Timestamp),
            Atomic(MinKey),
            Atomic(MaxKey),
            Array(Box::new(Any)),
            Document(Document::any()),
            // no Missing
        ]),
        _self = Any
    );
}

mod enumerate_field_paths {
    use crate::{
        map,
        schema::{Atomic::*, Document, Error, Schema, Schema::*, ANY_DOCUMENT},
        set,
    };
    use lazy_static::lazy_static;
    use std::collections::BTreeSet;

    lazy_static! {
        static ref A_B_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {"a".to_string() => Document(Document {
            keys: map! {"b".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
            })},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        });
    }

    macro_rules! test_enumerate_field_paths {
        ($func_name:ident, expected = $expected:expr, schema = $schema:expr, $(max_length = $max_length:expr,)?) => {
            #[test]
            fn $func_name() {
                #[allow(unused_mut, unused_assignments)]
                let mut max_length: Option<u32> = None;
                $(max_length = $max_length;)?

                let res = $schema.enumerate_field_paths(max_length);
                assert_eq!($expected, res)
            }
        };
    }

    test_enumerate_field_paths!(
        atomic,
        expected = Ok((set! {}, true)),
        schema = Atomic(Integer),
    );
    test_enumerate_field_paths!(
        array,
        expected = Ok((set! {}, true)),
        schema = Array(Box::new(Atomic(Integer))),
    );
    test_enumerate_field_paths!(
        any,
        expected = Err(Error::CannotEnumerateAllFieldPaths(Any)),
        schema = Any,
    );
    test_enumerate_field_paths!(unsat, expected = Ok((set! {}, true)), schema = Unsat,);
    test_enumerate_field_paths!(missing, expected = Ok((set! {}, true)), schema = Missing,);
    test_enumerate_field_paths!(
        document,
        expected = Ok((set! {vec!["a".to_string(), "b".to_string()]}, true)),
        schema = A_B_DOCUMENT_SCHEMA.clone(),
    );
    test_enumerate_field_paths!(
        any_of_documents,
        expected = Ok((
            set! {vec!["a".to_string(), "b".to_string()], vec!["x".to_string(), "y".to_string()]},
            true
        )),
        schema = Schema::AnyOf(set![
            A_B_DOCUMENT_SCHEMA.clone(),
            Document(Document {
                keys: map! {"x".to_string() => Document(Document {
                keys: map! {"y".to_string() => Atomic(Integer)},
                required: BTreeSet::new(),
                additional_properties: false,
                ..Default::default()
                })},
                required: BTreeSet::new(),
                additional_properties: false,
                ..Default::default()
            })
        ]),
    );
    test_enumerate_field_paths!(
        any_of_non_docs,
        expected = Ok((set! {}, true)),
        schema = AnyOf(set![Atomic(Integer), Atomic(String)]),
    );
    test_enumerate_field_paths!(
        nested_any_of_non_docs,
        expected = Ok((set! {}, true)),
        schema = AnyOf(set![
            AnyOf(set![Atomic(Integer), Atomic(Double)]),
            Atomic(String)
        ]),
    );
    test_enumerate_field_paths!(
        any_of_doc_and_non_doc_other_than_null_or_missing_returns_false,
        expected = Ok((set! {vec!["a".to_string(), "b".to_string()]}, false)),
        schema = AnyOf(set![A_B_DOCUMENT_SCHEMA.clone(), Atomic(String)]),
    );
    test_enumerate_field_paths!(
        any_of_doc_and_null_returns_true,
        expected = Ok((set! {vec!["a".to_string(), "b".to_string()]}, true)),
        schema = AnyOf(set![A_B_DOCUMENT_SCHEMA.clone(), Atomic(Null)]),
    );
    test_enumerate_field_paths!(
        any_of_doc_and_missing_returns_true,
        expected = Ok((set! {vec!["a".to_string(), "b".to_string()]}, true)),
        schema = AnyOf(set![A_B_DOCUMENT_SCHEMA.clone(), Missing]),
    );
    test_enumerate_field_paths!(
        any_of_doc_and_null_and_missing_returns_true,
        expected = Ok((set! {vec!["a".to_string(), "b".to_string()]}, true)),
        schema = AnyOf(set![A_B_DOCUMENT_SCHEMA.clone(), Atomic(Null), Missing]),
    );
    test_enumerate_field_paths!(
        nested_any_of_doc,
        expected = Ok((
            set! {vec!["a".to_string()], vec!["a".to_string(), "b".to_string()]},
            false
        )),
        schema = AnyOf(set! {
            Document(Document {
                keys: map!{
                    "a".to_string() => AnyOf(set!{
                        Atomic(Integer),
                        Document(Document {
                            keys: map! {"b".to_string() => Atomic(Integer)},
                            required: set!["b".to_string()],
                            additional_properties: false,
                            ..Default::default()
                            }),
                    }),
                },
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
                }),
            Atomic(Integer)
        }),
    );

    test_enumerate_field_paths!(
        any_of_identical_documents,
        expected = Ok((set! {vec!["a".to_string(), "b".to_string()]}, true)),
        schema = AnyOf(set![
            A_B_DOCUMENT_SCHEMA.clone(),
            A_B_DOCUMENT_SCHEMA.clone()
        ]),
    );
    test_enumerate_field_paths!(
        empty_any_of,
        expected = Ok((set! {}, true)),
        schema = AnyOf(set![]),
    );
    test_enumerate_field_paths!(
        any_of_enumerable_and_non_enumerable_documents,
        expected = Err(Error::CannotEnumerateAllFieldPaths(Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer), "b".to_string() => Atomic(Integer)},
            required: set![],
            additional_properties: true,
            ..Default::default()
        }))),
        schema = AnyOf(set![
            Document(Document {
                keys: map! {"a".to_string() => Atomic(Integer)},
                required: set!["a".to_string()],
                additional_properties: false,
                ..Default::default()
            }),
            Document(Document {
                keys: map! {"b".to_string() => Atomic(Integer)},
                required: set!["b".to_string()],
                additional_properties: true,
                ..Default::default()
            }),
        ]),
    );
    test_enumerate_field_paths!(
        any_document,
        expected = Err(Error::CannotEnumerateAllFieldPaths(Document(Document {
            keys: map! {},
            required: set![],
            additional_properties: true,
            ..Default::default()
        }))),
        schema = ANY_DOCUMENT,
    );
    test_enumerate_field_paths!(
        outer_additional_properties_true,
        expected = Err(Error::CannotEnumerateAllFieldPaths(Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer)},
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        }))),
        schema = Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer)},
            required: set!["a".to_string()],
            additional_properties: true,
            ..Default::default()
        }),
    );
    test_enumerate_field_paths!(
        inner_additional_properties_true,
        expected = Err(Error::CannotEnumerateAllFieldPaths(Document(Document {
            keys: map! {"b".to_string() => Atomic(Integer)},
            required: set!["b".to_string()],
            additional_properties: true,
            ..Default::default()
        }))),
        schema = Document(Document {
            keys: map! {"a".to_string() => Document(Document {
            keys: map!{"b".to_string() => Atomic(Integer)},
            required: set!["b".to_string()],
            additional_properties: true,
            ..Default::default()
            })},
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_enumerate_field_paths!(
        two_paths_share_a_prefix,
        expected = Ok((
            set! {vec!["a".to_string(), "b".to_string()], vec!["a".to_string(), "c".to_string()]},
            true
        )),
        schema = Document(Document {
            keys: map! {"a".to_string() => Document(Document {
            keys: map!{"b".to_string() => Atomic(Integer), "c".to_string() => Atomic(Integer)},
            required: set!["b".to_string(), "c".to_string()],
            additional_properties: false,
            ..Default::default()
            })},
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
    );
    test_enumerate_field_paths!(
        document_max_length_zero,
        expected = Ok((set! {}, true)),
        schema = A_B_DOCUMENT_SCHEMA.clone(),
        max_length = Some(0),
    );
    test_enumerate_field_paths!(
        document_max_length_less_than_max_nesting_depth,
        expected = Ok((set! {vec!["a".to_string()]}, true)),
        schema = Document(Document {
            keys: map! {"a".to_string() => Document(Document {
            keys: map!{"b".to_string() => Atomic(Integer)},
            required: set!["b".to_string()],
            additional_properties: false,
            ..Default::default()
            })},
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        max_length = Some(1),
    );
    test_enumerate_field_paths!(
        incomplete_subdocument_schema,
        expected = Ok((set! {vec!["a".to_string()]}, true)),
        schema = Document(Document {
            keys: map! {"a".to_string() => Document(Document {
            keys: map!{"b".to_string() => Document(Document {
                keys: map!{},
                required: set![],
                additional_properties: true,
                ..Default::default()
                })},
            required: set!["b".to_string()],
            additional_properties: false,
            ..Default::default()
            })},
            required: set!["a".to_string()],
            additional_properties: false,
            ..Default::default()
        }),
        max_length = Some(1),
    );
}

mod keys {
    use crate::{
        map,
        schema::{Atomic, Document, Schema::*},
        set,
    };

    macro_rules! test_keys {
        ($func_name:ident, expected = $expected:expr, schema = $schema:expr $(,)?) => {
            #[test]
            fn $func_name() {
                assert_eq!($expected, $schema.keys());
            }
        };
    }

    test_keys!(any, expected = Vec::<String>::new(), schema = &Any);

    test_keys!(unsat, expected = Vec::<String>::new(), schema = &Unsat);

    test_keys!(missing, expected = Vec::<String>::new(), schema = &Missing);

    test_keys!(
        atomic,
        expected = Vec::<String>::new(),
        schema = &Atomic(Atomic::Integer)
    );

    test_keys!(
        any_of,
        expected = vec!["a", "b", "c", "d", "bar", "baz", "foo"],
        schema = &AnyOf(set![
            Atomic(Atomic::Integer),
            Any,
            Document(Document {
                keys: map! {
                    "a".into() => Any,
                    "b".into() => Unsat,
                    "c".into() => Atomic(Atomic::String),
                    "d".into() => Document(Document{
                        keys: map!{
                            "a2".into() => Any,
                            "b2".into() => Unsat,
                            "c2".into() => Atomic(Atomic::String),
                            },
                        required: set![],
                        additional_properties: false,
                        ..Default::default()
                        }),
                },
                required: set![],
                additional_properties: true,
                ..Default::default()
            }),
            AnyOf(set![
                Document(Document {
                    keys: map! {
                    "foo".into() => Any,
                    "bar".into() => Unsat,
                    "baz".into() => Atomic(Atomic::String)},
                    required: set![],
                    additional_properties: false,
                    ..Default::default()
                }),
                Any,
                Missing,
            ]),
        ])
    );

    test_keys!(
        array,
        expected = Vec::<String>::new(),
        schema = &Array(Box::new(Atomic(Atomic::Integer)))
    );

    test_keys!(
        document,
        expected = vec!["a", "b", "c", "d", "e"],
        schema = &Document(Document {
            keys: map! {
                "a".into() => Any,
                "b".into() => Unsat,
                "c".into() => Atomic(Atomic::String),
                "d".into() => Document(Document{
                    keys: map!{
                        "a2".into() => Any,
                        "b2".into() => Unsat,
                        "c2".into() => Atomic(Atomic::String),

                        },
                    required: set![],
                    additional_properties: false,
                    ..Default::default()
                    }),
                "e".into() => Atomic(Atomic::Integer),
            },
            required: set![],
            additional_properties: false,
            ..Default::default()
        })
    );
}

mod display_trait {
    use crate::schema::{Atomic, Schema::*};

    macro_rules! test_display_trait {
        ($func_name:ident, expected = $expected:literal, schema = $schema:expr $(,)?) => {
            #[test]
            fn $func_name() {
                assert_eq!($expected, format!("{}", $schema));
            }
        };
    }

    test_display_trait!(non_atomic, expected = "any type", schema = &Any,);

    test_display_trait!(
        atomic,
        expected = "objectId",
        schema = &Atomic(Atomic::ObjectId),
    );

    test_display_trait!(unsat, expected = "Unsat", schema = &Unsat,);
}

mod collision_check {
    use crate::{
        map,
        mir::binding_tuple::Key,
        schema::{Atomic::*, Document, Error::*, Schema, Schema::*, SchemaEnvironment},
    };
    use lazy_static::lazy_static;
    use std::collections::BTreeSet;

    lazy_static! {
        static ref A_C_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer),
            "c".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        });
        static ref A_B_C_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
            "c".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        });
        static ref A_B_C_D_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer),
                "b".to_string() => Atomic(Integer),
                "c".to_string() => Atomic(Integer),
            "d".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        });
        static ref A_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        });
        static ref B_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {"b".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        });
        static ref D_DOCUMENT_SCHEMA: Schema = Schema::Document(Document {
            keys: map! {"d".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        });
        static ref A_DOCUMENT_SCHEMA_ADDITIONAL_PROPS_TRUE: Schema = Schema::Document(Document {
            keys: map! {"a".to_string() => Atomic(Integer)},
            required: BTreeSet::new(),
            additional_properties: true,
            ..Default::default()
        });
    }

    macro_rules! test_collisions_check {
        ($func_name:ident,
        expected = $expected:expr,
        instances = {$($key:expr => $value:expr),*}) => {
            #[test]
            fn $func_name() {
                let mut env_instance = SchemaEnvironment::new();
                $(
                    let key = Key::named($key, 0);
                    env_instance.insert(key, $value);
                )*
                let res = env_instance.check_for_non_namespaced_collisions();
                assert_eq!($expected, res);
            }
        };
    }

    test_collisions_check! {
        single_schema_can_enumerate_all_fields,
        expected = Ok(()),
        instances = {
            "schema" => A_C_DOCUMENT_SCHEMA.clone()
        }
    }

    test_collisions_check! {
        single_schema_cannot_enumerate_all_fields,
        expected = Err(
            CannotEnumerateAllFieldPaths(A_DOCUMENT_SCHEMA_ADDITIONAL_PROPS_TRUE.clone())),
        instances = {
            "schema" => A_DOCUMENT_SCHEMA_ADDITIONAL_PROPS_TRUE.clone()
        }
    }

    test_collisions_check! {
        multiple_schemas_cannot_enumerate_all_fields,
        expected = Err(
            CannotEnumerateAllFieldPaths(A_DOCUMENT_SCHEMA_ADDITIONAL_PROPS_TRUE.clone())),
        instances = {
            "schema_props_true" => A_DOCUMENT_SCHEMA_ADDITIONAL_PROPS_TRUE.clone(),
            "schema_props_false" => A_C_DOCUMENT_SCHEMA.clone()
        }
    }

    test_collisions_check! {
        conflict_on_one_field,
        expected = Err(FieldConflictInNonNamespacedResult("Error 4000: Consider aliasing \
            the following conflicting field(s) to unique names: a\n\tCaused by:\n\tCannot \
            return non-namespaced result set due to field name conflict(s), schema \
            [Document(Document { keys: {\"a\": Atomic(Integer), \"c\": Atomic(Integer)}, \
            required: {}, additional_properties: false }), Document(Document { keys: {\"a\": \
            Atomic(Integer)}, required: {}, additional_properties: false })]".to_string())),
        instances = {
        "schema" => A_C_DOCUMENT_SCHEMA.clone(),
        "schema_other" => A_DOCUMENT_SCHEMA.clone()
        }
    }

    test_collisions_check! {
        conflict_on_two_fields,
        expected = Err(FieldConflictInNonNamespacedResult("Error 4000: Consider aliasing \
            the following conflicting field(s) to unique names: a, c\n\tCaused by:\n\tCannot \
            return non-namespaced result set due to field name conflict(s), schema \
            [Document(Document { keys: {\"a\": Atomic(Integer), \"c\": Atomic(Integer)}, \
            required: {}, additional_properties: false }), Document(Document { keys: {\"a\": \
            Atomic(Integer), \"c\": Atomic(Integer)}, required: {}, \
            additional_properties: false })]".to_string())),
        instances = {
            "schema" => A_C_DOCUMENT_SCHEMA.clone(),
            "schema_other" => A_C_DOCUMENT_SCHEMA.clone()
        }
    }

    test_collisions_check! {
        conflict_on_three_fields,
        expected = Err(FieldConflictInNonNamespacedResult("Error 4000: Consider aliasing \
            the following conflicting field(s) to unique names: a, b, c\n\tCaused by:\n\tCannot \
            return non-namespaced result set due to field name conflict(s), schema \
            [Document(Document { keys: {\"a\": Atomic(Integer), \"b\": Atomic(Integer), \"c\": \
            Atomic(Integer)}, required: {}, additional_properties: false }), Document(Document \
            { keys: {\"a\": Atomic(Integer), \"b\": Atomic(Integer), \"c\": Atomic(Integer), \
            \"d\": Atomic(Integer)}, required: {}, additional_properties: false })]".to_string())),
        instances = {
            "schema" => A_B_C_DOCUMENT_SCHEMA.clone(),
            "schema_other" => A_B_C_D_DOCUMENT_SCHEMA.clone()
        }
    }

    test_collisions_check! {
        two_schemas_no_conflict,
        expected = Ok(()),
        instances = {
            "schema_a" => A_C_DOCUMENT_SCHEMA.clone(),
            "schema_b" => B_DOCUMENT_SCHEMA.clone()
        }
    }

    test_collisions_check! {
        three_schemas_no_conflict,
        expected = Ok(()),
        instances = {
            "schema_a" => A_C_DOCUMENT_SCHEMA.clone(),
            "schema_b" => B_DOCUMENT_SCHEMA.clone(),
            "schema_d" => D_DOCUMENT_SCHEMA.clone()
        }
    }

    test_collisions_check! {
        multiple_schemas_with_conflict,
        expected = Err(FieldConflictInNonNamespacedResult("Error 4000: Consider aliasing \
            the following conflicting field(s) to unique names: a, c, b\n\tCaused by:\n\tCannot \
            return non-namespaced result set due to field name conflict(s), schema \
            [Document(Document { keys: {\"a\": Atomic(Integer), \"c\": Atomic(Integer)}, \
            required: {}, additional_properties: false }), Document(Document { keys: {\"b\": \
            Atomic(Integer)}, required: {}, additional_properties: false }), Document(Document \
            { keys: {\"a\": Atomic(Integer), \"c\": Atomic(Integer)}, required: {}, \
            additional_properties: false }), Document(Document { keys: {\"b\": Atomic(Integer)}, \
            required: {}, additional_properties: false })]".to_string())),
        instances = {
            "schema_1" => A_C_DOCUMENT_SCHEMA.clone(),
            "schema_2" => B_DOCUMENT_SCHEMA.clone(),
            "schema_3" => D_DOCUMENT_SCHEMA.clone(),
            "schema_4" => A_C_DOCUMENT_SCHEMA.clone(),
            "schema_5" => B_DOCUMENT_SCHEMA.clone()
        }
    }
}
