use crate::json_schema::*;

macro_rules! validate_json_schema {
    ($func_name:ident, expected_schema = $expected_schema:expr, $(expected_json = $expected_json:expr,)? input = $input:expr, ) => {
        #[test]
        fn $func_name() {
            let s: Schema = serde_json::from_str($input).unwrap();
            assert_eq!(s, *$expected_schema);

            #[allow(unused_mut, unused_assignments)]
            let mut expected_json = $input;
			$(expected_json = $expected_json;)?

            assert_eq!(
                serde_json::to_string($expected_schema).unwrap(),
                expected_json
            )
        }
    };
}

macro_rules! hashmap(
    { $($key:expr => $value:expr),+ } => {
      {
        let mut m = HashMap::new();
        $(
          m.insert($key, $value);
        )+
        m
      }
    };
  );

validate_json_schema!(
    empty_schema,
    expected_schema = &Schema::default(),
    input = "{}",
);
validate_json_schema!(
    schema_with_single_bson_type,
    expected_schema = &Schema {
        bson_type: Some(BsonType::Single(BsonTypeName::Int)),
        ..Default::default()
    },
    input = r#"{"bsonType":"int"}"#,
);
validate_json_schema!(
    schema_with_multiple_bson_types,
    expected_schema = &Schema {
        bson_type: Some(BsonType::Multiple(vec![
            BsonTypeName::Int,
            BsonTypeName::Null
        ])),
        ..Default::default()
    },
    input = r#"{"bsonType":["int","null"]}"#,
);
validate_json_schema!(
    schema_with_properties,
    expected_schema = &Schema {
        properties: Some(hashmap! {
            "a".to_string() => Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Schema::default()
            }
        }),
        ..Default::default()
    },
    input = r#"{"properties":{"a":{"bsonType":"int"}}}"#,
);
validate_json_schema!(
    schema_with_required,
    expected_schema = &Schema {
        required: Some(vec!["a".to_string(), "b".to_string()]),
        ..Default::default()
    },
    input = r#"{"required":["a","b"]}"#,
);
validate_json_schema!(
    schema_with_additional_properties_field,
    expected_schema = &Schema {
        additional_properties: Some(true),
        ..Default::default()
    },
    input = r#"{"additionalProperties":true}"#,
);
validate_json_schema!(
    schema_with_items,
    expected_schema = &Schema {
        items: Some(Items::Single(Box::new(Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            ..Default::default()
        }))),
        ..Default::default()
    },
    input = r#"{"items":{"bsonType":"int"}}"#,
);
validate_json_schema!(
    schema_with_any_of,
    expected_schema = &Schema {
        any_of: Some(vec![
            Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            },
            Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Null)),
                ..Default::default()
            }
        ]),
        ..Default::default()
    },
    input = r#"{"anyOf":[{"bsonType":"int"},{"bsonType":"null"}]}"#,
);
validate_json_schema!(
    schema_with_one_of,
    expected_schema = &Schema {
        one_of: Some(vec![
            Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            },
            Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Null)),
                ..Default::default()
            }
        ]),
        ..Default::default()
    },
    input = r#"{"oneOf":[{"bsonType":"int"},{"bsonType":"null"}]}"#,
);
validate_json_schema!(
    schema_with_extra_fields_ignored,
    expected_schema = &Schema {
        bson_type: Some(BsonType::Single(BsonTypeName::Int)),
        ..Default::default()
    },
    expected_json = r#"{"bsonType":"int"}"#,
    input = r#"{"extra1":"value1","bsonType":"int","extra2":"value2"}"#,
);
validate_json_schema!(
    schema_with_all_fields_non_default,
    expected_schema = &Schema {
        bson_type: Some(BsonType::Multiple(vec![
            BsonTypeName::Object,
            BsonTypeName::Array
        ])),
        properties: Some(hashmap! {
            "a".to_string() => Schema {
                bson_type: Some(BsonType::Single(BsonTypeName::Int)),
                ..Default::default()
            }
        }),
        required: Some(vec!["a".to_string()]),
        additional_properties: Some(true),
        items: Some(Items::Single(Box::new(Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            ..Default::default()
        }))),
        any_of: Some(vec![Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            ..Default::default()
        }]),
        one_of: Some(vec![Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            ..Default::default()
        }]),
    },
    input = r#"{"bsonType":["object","array"],"properties":{"a":{"bsonType":"int"}},"required":["a"],"additionalProperties":true,"items":{"bsonType":"int"},"anyOf":[{"bsonType":"int"}],"oneOf":[{"bsonType":"int"}]}"#,
);
validate_json_schema!(
    schema_with_items_as_array,
    expected_schema = &Schema {
        bson_type: Some(BsonType::Single(BsonTypeName::Array)),
        items: Some(Items::Multiple(vec![Schema::default()])),
        ..Default::default()
    },
    input = r#"{"bsonType":"array","items":[{}]}"#,
);

mod no_stack_overflows {
    use super::*;

    #[test]
    fn when_serializing() {
        let mut schema = Schema {
            bson_type: Some(BsonType::Single(BsonTypeName::Int)),
            ..Default::default()
        };
        for _ in 0..1000 {
            schema = Schema {
                any_of: Some(vec![schema]),
                ..Default::default()
            };
        }

        let _ = schema.to_bson();
    }

    #[test]
    fn when_deserializing() {
        let mut doc = bson::doc! { "bsonType": "int" };
        // This uses a smaller limit than the serialization test
        // since the mongodb::bson::Deserializer does not support
        // the disable_recursion_limit method required by serde_stacker
        // to handle greater depths. At 250 nesting depth, the standard
        // bson::from_document encounters a stack overflow, whereas the
        // manual serde_stacker deserialization technique does not.
        for _ in 0..250 {
            doc = bson::doc! { "anyOf": [doc] }
        }

        println!("start");

        let _ = Schema::from_document(&doc);
    }
}
