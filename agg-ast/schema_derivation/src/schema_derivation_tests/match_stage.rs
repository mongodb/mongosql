use crate::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Stage;
use mongosql::{
    map,
    schema::{Atomic, Document, Schema},
    set,
};
use std::collections::BTreeMap;

macro_rules! test_derive_schema_for_match_stage {
    // ref_schema and starting_schema are mutually exclusive. ref_schema should be used when only
    // one reference is needed, while starting_schema should be used when the schema needs multiple
    // fields.
    ($func_name:ident, expected = $expected:expr, input = $input:expr$(, starting_schema = $starting_schema:expr)?$(, ref_schema = $ref_schema:expr)?$(, variables = $variables:expr)?) => {
        #[test]
        fn $func_name() {
            let input: Stage = serde_json::from_str($input).unwrap();
            #[allow(unused_mut, unused_assignments)]
            let mut result_set_schema = Schema::Any;
            $(result_set_schema = Schema::Document(Document {
                keys: map! {"foo".to_string() => $ref_schema },
                required: set!{"foo".to_string()},
                ..Default::default()
            });)?
            $(result_set_schema = $starting_schema;)?
            #[allow(unused_mut, unused_assignments)]
            let mut variables = BTreeMap::new();
            $(variables = $variables;)?
            let mut state = ResultSetState {
                catalog: &BTreeMap::new(),
                variables: &variables,
                result_set_schema
            };
            let result = input.derive_schema(&mut state);
            let result = result.as_ref().map(Schema::simplify);
            assert_eq!($expected, result);
        }
    };
}

macro_rules! test_derivation_for_bits_ops {
    (
        $name:ident,
        operator = $op:expr,
    ) => {
        test_derive_schema_for_match_stage! {
            $name,
            expected = Ok(Schema::Document(Document {
                keys: map! {
                    "foo".to_string() => Schema::AnyOf(
                        set!{
                             Schema::Atomic(Atomic::BinData),
                             Schema::Atomic(Atomic::Double),
                             Schema::Atomic(Atomic::Decimal),
                             Schema::Atomic(Atomic::Long),
                             Schema::Atomic(Atomic::Integer)
                        }
                    ),
                },
                required: set!{"foo".to_string()},
                additional_properties: false,
                jaccard_index: None,
            })),
            input = format!(r#"{{"$match": {{"foo": {{"{}": [1, 5]}}}}}}"#, $op).as_str(),
            ref_schema = Schema::Any
        }
    };
}

macro_rules! test_derivation_for_geo_ops {
    (
        $name:ident,
        operator = $op:expr,
    ) => {
        test_derive_schema_for_match_stage! {
            $name,
            expected = Ok(Schema::Document(Document {
                keys: map! {
                    "foo".to_string() => Schema::AnyOf(set!{
                        Schema::Document(Document {
                            keys: map! {
                                "type".to_string() => Schema::Atomic(Atomic::String),
                                "coordinates".to_string() => Schema::Array(Box::new(Schema::AnyOf(
                                    set!{Schema::Atomic(Atomic::Double), Schema::Atomic(Atomic::Decimal), Schema::Atomic(Atomic::Long), Schema::Atomic(Atomic::Integer)}
                                ))),
                            },
                            required: set!{"coordinates".to_string()},
                            additional_properties: false,
                            jaccard_index: None,
                        }),
                        Schema::Array(Box::new(Schema::AnyOf(
                            set!{Schema::Atomic(Atomic::Double), Schema::Atomic(Atomic::Decimal), Schema::Atomic(Atomic::Long), Schema::Atomic(Atomic::Integer)}
                        ))),
                    }),
                },
                required: set!{"foo".to_string()},
                additional_properties: false,
                jaccard_index: None,
            })),
            input = format!(r#"{{"$match": {{"foo": {{"{}": [1, 5]}}}}}}"#, $op).as_str(),
            ref_schema = Schema::Any
        }
    };
}

test_derive_schema_for_match_stage! {
    derivation_for_multi_match_exprs,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set!{
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
            }),
            "bar".to_string() => Schema::Document(Document {
                keys: map! {
                    "car".to_string() => Schema::Atomic(Atomic::String),
                },
                required: set!{"car".to_string()},
                additional_properties: false,
                jaccard_index: None,
            }),
        },
        required: set!{"foo".to_string(), "bar".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$eq": 4}, "bar.car": {"$eq": "hello"}}}"#,
    starting_schema = Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Any,
            "bar".to_string() => Schema::Document(Document {
                keys: map! {
                    "car".to_string() => Schema::AnyOf(set!{
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Null),
                    }),
                },
                required: set!{"car".to_string()},
                additional_properties: false,
                jaccard_index: None,
            }),
        },
        required: set!{"foo".to_string(), "bar".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_two_constraints,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::Null),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$gte": null, "$exists": true}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_eq,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set!{
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
            }),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": 4}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_gt,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set!{
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
            }),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$gt": 4}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_gte,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set!{
                Schema::Atomic(Atomic::Null),
            }),
        },
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$gte": null}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_lt,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set!{
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
            }),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$lt": 4}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_lte,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$lte": "hello"}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_ne,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$ne": null}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_42,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": 42}}}"#,
    // Missing should be inferred since it's not in the required set
    starting_schema = Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_long_42,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": {"$numberLong": "42"}}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_42p5,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": 42.5}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_decimal_42p5,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": {"$numberDecimal": "42.5"}}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_true,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": true}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_0,
    expected = Ok(Schema::Document(Document {
        keys: map! {},
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": 0}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_long_0,
    expected = Ok(Schema::Document(Document {
        keys: map! {},
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": {"$numberLong": "0"}}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_0p0,
    expected = Ok(Schema::Document(Document {
        keys: map! {        },
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": 0.0}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_decimal_0p0,
    expected = Ok(Schema::Document(Document {
        keys: map! {},
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": {"$numberDecimal": "0.0"}}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_false,
    expected = Ok(Schema::Document(Document {
        keys: map! {},
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": false}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_null,
    expected = Ok(Schema::Document(Document {
        keys: map! {},
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$exists": null}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_type_string,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$type": "string"}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Missing,
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_type_multi,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() =>
                Schema::AnyOf(set!{
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Integer),
                })
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$type": ["string", "int"]}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Atomic(Atomic::Integer),
        Schema::Missing,
        Schema::Atomic(Atomic::Null),
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_in,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() =>
                Schema::AnyOf(set!{
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::String),
                    Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
                })
        },
        required: set!{},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$in": [["hello", "world"], "hello", null]}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Atomic(Atomic::Integer),
        Schema::Missing,
        Schema::Atomic(Atomic::Null),
        Schema::Array(Box::new(Schema::Any)),
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_nin,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() =>
                Schema::AnyOf(set!{
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::String),
                    Schema::Array(Box::new(Schema::Any)),
                })
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$nin": [["hello", "world"], "hello", null]}}}"#,
    ref_schema = Schema::AnyOf(set!{
        Schema::Atomic(Atomic::String),
        Schema::Atomic(Atomic::Integer),
        Schema::Missing,
        Schema::Atomic(Atomic::Null),
        Schema::Array(Box::new(Schema::Any)),
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_mod,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set!{
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
            }),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$mod": [2,0]}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_size,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Array(Box::new(Schema::Any)),
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$size": 2}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_all,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!{
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::Null),
            })))
        },
        required: set!{"foo".to_string()},
        additional_properties: false,
        jaccard_index: None,
    })),
    input = r#"{"$match": {"foo": {"$all": ["hello", "world", null]}}}"#,
    ref_schema = Schema::Any
}

test_derivation_for_bits_ops! { derivation_for_bits_all_clear, operator = "$bitsAllClear", }
test_derivation_for_bits_ops! { derivation_for_bits_any_clear, operator = "$bitsAnyClear", }
test_derivation_for_bits_ops! { derivation_for_bits_all_set, operator = "$bitsAllSet", }
test_derivation_for_bits_ops! { derivation_for_bits_any_set, operator = "$bitsAnySet", }

test_derivation_for_geo_ops! { derivation_for_geo_near, operator = "$near", }
test_derivation_for_geo_ops! { derivation_for_geo_near_sphere, operator = "$nearSphere", }
test_derivation_for_geo_ops! { derivation_for_geo_geo_within, operator = "$geoWithin", }
test_derivation_for_geo_ops! { derivation_for_geo_geo_intersects, operator = "$geoIntersects", }
