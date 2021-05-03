use crate::json_schema::*;
use std::collections::HashMap;

macro_rules! validate_json_schema {
    ($func_name:ident, $input:expr, $expected_schema:expr, $expected_json:expr) => {
        #[test]
        fn $func_name() {
            let s: Schema = serde_json::from_str($input).unwrap();
            assert_eq!(s, *$expected_schema);

            let expected_json = if $expected_json == "" {
                $input
            } else {
                $expected_json
            };
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

validate_json_schema!(empty_schema, "{}", &Schema::default(), "");
validate_json_schema!(
    schema_with_single_bson_type,
    r#"{"bsonType":"int"}"#,
    &Schema {
        bson_type: BsonType::Single("int".to_string()),
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_multiple_bson_types,
    r#"{"bsonType":["int","null"]}"#,
    &Schema {
        bson_type: BsonType::Multiple(vec!["int".to_string(), "null".to_string()]),
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_properties,
    r#"{"properties":{"a":{"bsonType":"int"}}}"#,
    &Schema {
        properties: hashmap! {
            "a".to_string() => Schema {
                bson_type: BsonType::Single("int".to_string()),
                ..Default::default()
            }
        },
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_required,
    r#"{"required":["a","b"]}"#,
    &Schema {
        required: vec!["a".to_string(), "b".to_string()],
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_additional_properties_field,
    r#"{"additionalProperties":true}"#,
    &Schema {
        additional_properties: true,
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_items,
    r#"{"items":[{"bsonType":"int"},{"bsonType":"null"}]}"#,
    &Schema {
        items: vec![
            Schema {
                bson_type: BsonType::Single("int".to_string()),
                ..Default::default()
            },
            Schema {
                bson_type: BsonType::Single("null".to_string()),
                ..Default::default()
            }
        ],
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_any_of,
    r#"{"anyOf":[{"bsonType":"int"},{"bsonType":"null"}]}"#,
    &Schema {
        any_of: vec![
            Schema {
                bson_type: BsonType::Single("int".to_string()),
                ..Default::default()
            },
            Schema {
                bson_type: BsonType::Single("null".to_string()),
                ..Default::default()
            }
        ],
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_one_of,
    r#"{"oneOf":[{"bsonType":"int"},{"bsonType":"null"}]}"#,
    &Schema {
        one_of: vec![
            Schema {
                bson_type: BsonType::Single("int".to_string()),
                ..Default::default()
            },
            Schema {
                bson_type: BsonType::Single("null".to_string()),
                ..Default::default()
            }
        ],
        ..Default::default()
    },
    ""
);
validate_json_schema!(
    schema_with_extra_fields_ignored,
    r#"{"extra1":"value1","bsonType":"int","extra2":"value2"}"#,
    &Schema {
        bson_type: BsonType::Single("int".to_string()),
        ..Default::default()
    },
    r#"{"bsonType":"int"}"#
);
validate_json_schema!(
    schema_with_all_fields_non_default,
    r#"{"bsonType":["object","array"],"properties":{"a":{"bsonType":"int"}},"required":["a"],"additionalProperties":true,"items":[{"bsonType":"int"}],"anyOf":[{"bsonType":"int"}],"oneOf":[{"bsonType":"int"}]}"#,
    &Schema {
        bson_type: BsonType::Multiple(vec!["object".to_string(), "array".to_string()]),
        properties: hashmap! { "a".to_string() => Schema {
            bson_type: BsonType::Single("int".to_string()),
            ..Default::default()
        } },
        required: vec!["a".to_string()],
        additional_properties: true,
        items: vec![Schema {
            bson_type: BsonType::Single("int".to_string()),
            ..Default::default()
        }],
        any_of: vec![Schema {
            bson_type: BsonType::Single("int".to_string()),
            ..Default::default()
        }],
        one_of: vec![Schema {
            bson_type: BsonType::Single("int".to_string()),
            ..Default::default()
        }],
    },
    ""
);
