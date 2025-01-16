use crate::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Stage;
use mongosql::{
    map,
    schema::{Atomic, Document, Satisfaction, Schema},
    set,
};
use std::collections::BTreeMap;

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
                ..Default::default()
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
                            ..Default::default()
                        }),
                        Schema::Array(Box::new(Schema::AnyOf(
                            set!{Schema::Atomic(Atomic::Double), Schema::Atomic(Atomic::Decimal), Schema::Atomic(Atomic::Long), Schema::Atomic(Atomic::Integer)}
                        ))),
                    }),
                },
                required: set!{"foo".to_string()},
                ..Default::default()
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
                ..Default::default()
            }),
        },
        required: set!{"foo".to_string(), "bar".to_string()},
        ..Default::default()
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
        ..Default::default()
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_two_constraints,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::Null),
        },
        required: set!{"foo".to_string()},
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
    })),
    input = r#"{"$match": {"foo": {"$gt": 4}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_gte,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::Null),
        },
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
    })),
    input = r#"{"$match": {"foo": {"$exists": 42}}}"#,
    // Missing should be inferred since it's not in the required set
    starting_schema = Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        ..Default::default()
    })
}

test_derive_schema_for_match_stage! {
    derivation_for_exists_long_42,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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
        ..Default::default()
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

test_derive_schema_for_match_stage! {
    derivation_for_or_string_oid,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set!{
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::ObjectId),
            })
        },
        required: set!{"foo".to_string()},
        ..Default::default()
    })),
    input = r#"{"$match": {"$or": [{"foo": "hello"}, {"foo": {"$type": "objectId"}}]}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_and_over_or,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        required: set!{"foo".to_string()},
        ..Default::default()
    })),
    input = r#"{"$match": {"$and": [{"$or": [{"foo": "hello"}, {"foo": {"$type": "objectId"}}]}, {"foo": {"$type": "string"}}]}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_or_over_and_exists_true,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set![
                Schema::Atomic(Atomic::String),
                Schema::Atomic(Atomic::ObjectId),
            ]),
        },
        required: set!{"foo".to_string()},
        ..Default::default()
    })),
    input = r#"{"$match": {"$or": [{"$and": [{"foo": "hello"}, {"foo": {"$type": "string"}}]}, {"foo": {"$type": ["objectId"], "$exists": true}}]}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_or_over_and_exists_false,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::Atomic(Atomic::String),
        },
        ..Default::default()
    })),
    input = r#"{"$match": {"$or": [{"$and": [{"foo": "hello"}, {"foo": {"$type": "string"}}]}, {"foo": {"$exists": false}}]}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_nor,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set![
                 Schema::Atomic(Atomic::Double),
                 Schema::Document(Document::any()),
                 Schema::Array(Box::new(Schema::Any)),
                 Schema::Atomic(Atomic::BinData),
                 Schema::Atomic(Atomic::Undefined),
                 Schema::Atomic(Atomic::Boolean),
                 Schema::Atomic(Atomic::Date),
                 Schema::Atomic(Atomic::Null),
                 Schema::Atomic(Atomic::Regex),
                 Schema::Atomic(Atomic::DbPointer),
                 Schema::Atomic(Atomic::Javascript),
                 Schema::Atomic(Atomic::Symbol),
                 Schema::Atomic(Atomic::JavascriptWithScope),
                 Schema::Atomic(Atomic::Integer),
                 Schema::Atomic(Atomic::Timestamp),
                 Schema::Atomic(Atomic::Long),
                 Schema::Atomic(Atomic::Decimal),
                 Schema::Atomic(Atomic::MinKey),
                 Schema::Atomic(Atomic::MaxKey),
            ]),
        },
        ..Default::default()
    })),
    input = r#"{"$match": {"$nor": [{"foo": {"$type": "string"}}, {"foo": {"$type": "objectId"}}]}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_not,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set![
                 Schema::Atomic(Atomic::Double),
                 Schema::Document(Document::any()),
                 Schema::Array(Box::new(Schema::Any)),
                 Schema::Atomic(Atomic::BinData),
                 Schema::Atomic(Atomic::Undefined),
                 Schema::Atomic(Atomic::Boolean),
                 Schema::Atomic(Atomic::Date),
                 Schema::Atomic(Atomic::Null),
                 Schema::Atomic(Atomic::Regex),
                 Schema::Atomic(Atomic::DbPointer),
                 Schema::Atomic(Atomic::Javascript),
                 Schema::Atomic(Atomic::Symbol),
                 Schema::Atomic(Atomic::JavascriptWithScope),
                 Schema::Atomic(Atomic::Timestamp),
                 Schema::Atomic(Atomic::Long),
                 Schema::Atomic(Atomic::Decimal),
                 Schema::Atomic(Atomic::MinKey),
                 Schema::Atomic(Atomic::MaxKey),
            ]),
        },
        ..Default::default()
    })),
    input = r#"{"$match": {"foo": {"$not": {"$type": ["string", "int", "objectId"]}}}}"#,
    ref_schema = Schema::Any
}

test_derive_schema_for_match_stage! {
    derivation_for_not_multi,
    expected = Ok(Schema::Document(Document {
        keys: map! {
            "foo".to_string() => Schema::AnyOf(set![
                 Schema::Atomic(Atomic::Double),
                 Schema::Document(Document::any()),
                 Schema::Array(Box::new(Schema::Any)),
                 Schema::Atomic(Atomic::String),
                 Schema::Atomic(Atomic::BinData),
                 Schema::Atomic(Atomic::Undefined),
                 Schema::Atomic(Atomic::ObjectId),
                 Schema::Atomic(Atomic::Boolean),
                 Schema::Atomic(Atomic::Date),
                 Schema::Atomic(Atomic::Regex),
                 Schema::Atomic(Atomic::DbPointer),
                 Schema::Atomic(Atomic::Javascript),
                 Schema::Atomic(Atomic::Symbol),
                 Schema::Atomic(Atomic::JavascriptWithScope),
                 Schema::Atomic(Atomic::Timestamp),
                 Schema::Atomic(Atomic::Integer),
                 Schema::Atomic(Atomic::Long),
                 Schema::Atomic(Atomic::Decimal),
                 Schema::Atomic(Atomic::MinKey),
                 Schema::Atomic(Atomic::MaxKey),
            ]),
        },
        ..Default::default()
    })),
    input = r#"{"$match": {"foo": {"$not": {"$type": ["string", "int", "null"], "$eq": null}}}}"#,
    ref_schema = Schema::Any
}
