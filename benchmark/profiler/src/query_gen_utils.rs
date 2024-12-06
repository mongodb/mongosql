use mongosql::json_schema;
use std::{collections::BTreeMap, fs::File, io::BufWriter, path::PathBuf};

/// query_gen_utils contains several utility types and functions that can
/// help a team member create large queries for memory testing. There are
/// functions for creating expressions and join clauses with specified
/// numbers of elements, as well as functions for creating large schemas.
///
/// Note that there is no function for writing/printing these queries
/// anywhere. That is up to the developer to decide what is best based
/// on what they are doing.
///
/// NameGenerator is an Iterator that generates 2-char-length names for use
/// in test SQL queries. The names are produced in ascending order following
/// the pattern: aa, ab, ac, ..., ay, az, ba, bb, ..., zy, zz.
///
/// After producing 'zz', the Iterator will be considered exhausted and will
/// return None for future calls to next().
///
/// Note that the names always come out in the same order, as described by the
/// pattern above.
#[derive(Copy, Clone)]
struct NameGenerator {
    first: char,
    second: char,
}

impl Iterator for NameGenerator {
    type Item = String;

    fn next(&mut self) -> Option<Self::Item> {
        if self.first > 'z' {
            return None;
        }

        let curr = format!("{}{}", self.first, self.second);

        self.second = (self.second as u8 + 1) as char;
        if self.second > 'z' {
            self.first = (self.first as u8 + 1) as char;
            self.second = 'a';
        }

        Some(curr)
    }
}

/// name_generator creates a NameGenerator at the start name, 'aa'.
fn name_generator() -> NameGenerator {
    NameGenerator {
        first: 'a',
        second: 'a',
    }
}

/// create_simple_expr creates an expression with num_operands number of
/// operands, joined using the separator. Note that the separator is
/// padded with whitespace. By providing a non-zero skip_num, the generated
/// names can start at an advanced position.
///
/// Examples:
///   assert_eq!(create_simple_expr(0, 2, "+"), "aa + ab".to_string())
///   assert_eq!(create_simple_expr(4, 2, "+"), "ae + af".to_string())
///
fn create_simple_expr(skip_num: usize, num_operands: usize, operator: &str) -> String {
    name_generator()
        .skip(skip_num)
        .take(num_operands)
        .map(|n| format!("`{n}`"))
        .collect::<Vec<String>>()
        .join(format!(" {operator} ").as_str())
}

/// create_complex_expr creates an expression with num_operands number of
/// operands, mapped using the operator_producer function, and joined using
/// the separator. Ensure that the operator_producer properly delimits fields
/// if `as` is going to be used since that is a keyword.
///
/// Example:
///   assert_eq!(
///     create_complex_expr(2, |n| format!("WHEN `{n}` THEN '{n}'"), " "),
///     "WHEN `aa` THEN 'aa' WHEN `ab` THEN 'ab'".to_string()
///   )
///
fn create_complex_expr(
    num_operands: usize,
    operator_producer: fn((usize, String)) -> String,
    separator: &str,
) -> String {
    name_generator()
        .take(num_operands)
        .enumerate()
        .map(operator_producer)
        .collect::<Vec<String>>()
        .join(separator)
}

/// create_join creates a JOIN string with num_collections number of
/// collections and the operator used for the ON expression between
/// each JOIN. The collections are joined in order of generation, and
/// each uses an ON expression. See the example.
///
/// Example:
///   assert_eq!(
///     create_join(3, "="),
///     "JOIN `ab` AS `ab` ON `aa`.a = `ab`.b JOIN `ac` AS `ac` ON `ab`.a = `ac`.b".to_string()
///   )
///
fn create_join(num_collections: usize, operator: &str) -> String {
    name_generator()
        .take(num_collections)
        .collect::<Vec<String>>()
        .windows(2)
        .map(|w| {
            format!(
                "JOIN `{1}` AS `{1}` ON `{0}`.a {operator} `{1}`.b",
                w[0], w[1]
            )
        })
        .collect::<Vec<String>>()
        .join(" ")
}

/// map! is a utility macro, copied from the main mongosql-rs/mongosql/src code.
/// It creates a Map based on key-value pairs.
macro_rules! map {
	($($key:expr => $val:expr),* $(,)?) => {
		std::iter::Iterator::collect([
			$({
				($key, $val)
			},)*
		].into_iter())
	};
}

/// Nullability is a utility enum for indicating whether and how
/// a field should be nullable.
enum Nullability {
    Always,
    Maybe,
    Never,
}

/// make_simple_schema creates a simple, single-type JSON schema
fn make_simple_schema(t: json_schema::BsonTypeName) -> json_schema::Schema {
    json_schema::Schema {
        bson_type: Some(json_schema::BsonType::Single(t)),
        ..Default::default()
    }
}

/// gen_schema_one_collection creates a JSON schema for a single collection,
/// "<db>.foo", with a schema with num_fields number of fields, each with the
/// primary_type as its type, and each being required. Depending on the value
/// of nullability, some, all, or none of the fields may be marked as
/// AnyOf(<primary>, null).
///
/// Example:
///   gen_schema_one_collection(
///     2,
///     json_schema::BsonTypeName::Bool,
///     Nullability::Never,
///     "test_db",
///     "src/config_loader/catalogs/test_db.json"
///   )
///   creates the JSON file:
///         {
///           "catalog_schema": {
///             "test_db": {
///               "foo": {
///                 "bsonType": "object",
///                 "properties": {
///                   "aa": {
///                     "bsonType": "bool"
///                   },
///                   "ab": {
///                     "bsonType": "bool"
///                   }
///                 },
///                 "required": [
///                   "aa",
///                   "ab"
///                 ],
///                 "additionalProperties": false
///               }
///             }
///           }
///         }
fn gen_schema_one_collection(
    num_fields: usize,
    primary_type: json_schema::BsonTypeName,
    nullability: Nullability,
    db: &str,
    target_file: &str,
) {
    let primary_type_schema = make_simple_schema(primary_type);

    let null_schema = make_simple_schema(json_schema::BsonTypeName::Null);

    let nullable_primary_type_schema = json_schema::Schema {
        any_of: Some(vec![primary_type_schema.clone(), null_schema]),
        ..Default::default()
    };

    let init_schema = json_schema::Schema {
        bson_type: Some(json_schema::BsonType::Single(
            json_schema::BsonTypeName::Object,
        )),
        properties: Some(map! {}),
        required: Some(vec![]),
        additional_properties: Some(false),
        ..Default::default()
    };

    let foo_schema =
        name_generator()
            .take(num_fields)
            .enumerate()
            .fold(init_schema, |acc, (i, n)| match acc {
                json_schema::Schema {
                    properties: Some(mut properties),
                    required: Some(mut required),
                    ..
                } => {
                    let field_schema = match nullability {
                        Nullability::Always => nullable_primary_type_schema.clone(),
                        Nullability::Never => primary_type_schema.clone(),
                        Nullability::Maybe => {
                            if i % 2 == 0 {
                                nullable_primary_type_schema.clone()
                            } else {
                                primary_type_schema.clone()
                            }
                        }
                    };

                    properties.insert(n.clone(), field_schema);
                    required.append(&mut vec![n]);

                    json_schema::Schema {
                        bson_type: Some(json_schema::BsonType::Single(
                            json_schema::BsonTypeName::Object,
                        )),
                        properties: Some(properties),
                        required: Some(required),
                        additional_properties: Some(false),
                        ..Default::default()
                    }
                }
                _ => panic!("invalid schema"),
            });

    let full_schema: BTreeMap<&str, BTreeMap<&str, BTreeMap<&str, json_schema::Schema>>> = map! {
        "catalog_schema" => map! {
            db => map! {
                "foo" => foo_schema
            }
        }
    };

    let file = File::create(PathBuf::from(target_file)).unwrap();
    let w = BufWriter::new(file);
    serde_json::to_writer_pretty(w, &full_schema).unwrap()
}

/// gen_schema_multiple_collections creates a JSON schema for num_collections
/// collections, each with required, non-null, integer "a" and "b" fields.
///
/// Example:
///   gen_schema_multiple_collections(
///     2,
///     "test_db",
///     "src/config_loader/catalogs/test_db.json"
///   )
///   creates the JSON file:
///         {
///           "catalog_schema": {
///             "test_db": {
///               "aa": {
///                 "bsonType": "object",
///                 "properties": {
///                   "a": {
///                     "bsonType": "int"
///                   },
///                   "b": {
///                     "bsonType": "a"
///                   }
///                 },
///                 "required": [
///                   "a",
///                   "b"
///                 ],
///                 "additionalProperties": false
///               },
///               "ab": {
///                 "bsonType": "object",
///                 "properties": {
///                   "a": {
///                     "bsonType": "int"
///                   },
///                   "b": {
///                     "bsonType": "a"
///                   }
///                 },
///                 "required": [
///                   "a",
///                   "b"
///                 ],
///                 "additionalProperties": false
///               }
///             }
///           }
///         }
fn gen_schema_multiple_collections(num_collections: usize, db: &str, target_file: &str) {
    let int_schema = make_simple_schema(json_schema::BsonTypeName::Int);

    let collection_schema = json_schema::Schema {
        bson_type: Some(json_schema::BsonType::Single(
            json_schema::BsonTypeName::Object,
        )),
        properties: Some(map! {
            "a".to_string() => int_schema.clone(),
            "b".to_string() => int_schema,
        }),
        required: Some(vec!["a".to_string(), "b".to_string()]),
        additional_properties: Some(false),
        ..Default::default()
    };

    let collection_schemas = name_generator()
        .take(num_collections)
        .map(|n| (n, collection_schema.clone()))
        .collect::<BTreeMap<String, json_schema::Schema>>();

    let full_schema: BTreeMap<&str, BTreeMap<&str, BTreeMap<String, json_schema::Schema>>> = map! {
        "catalog_schema" => map! {
            db => collection_schemas,
        },
    };

    let file = File::create(PathBuf::from(target_file)).unwrap();
    let w = BufWriter::new(file);
    serde_json::to_writer_pretty(w, &full_schema).unwrap()
}

/// gen_extraneous_schema_multiple_collections is similar to gen_schema_multiple_collections,
/// however it introduces extraneous data to the schema. Specifically, it doubles the number
/// of collections and includes extra fields in each collection. This is deterministic --
/// no randomness is used at all.
fn gen_extraneous_schema_multiple_collections(num_collections: usize, db: &str, target_file: &str) {
    let collection_schema = json_schema::Schema {
        bson_type: Some(json_schema::BsonType::Single(
            json_schema::BsonTypeName::Object,
        )),
        properties: Some(map! {
            // only "a" and "b" are "required"
            "a".to_string() => make_simple_schema(json_schema::BsonTypeName::Int),
            "b".to_string() => make_simple_schema(json_schema::BsonTypeName::Int),
            // extraneous fields
            "obj".to_string() => make_simple_schema(json_schema::BsonTypeName::Object),
            "arr".to_string() => make_simple_schema(json_schema::BsonTypeName::Array),
            "nul".to_string() => make_simple_schema(json_schema::BsonTypeName::Null),
            "str".to_string() => make_simple_schema(json_schema::BsonTypeName::String),
            "int".to_string() => make_simple_schema(json_schema::BsonTypeName::Int),
            "dbl".to_string() => make_simple_schema(json_schema::BsonTypeName::Double),
            "lon".to_string() => make_simple_schema(json_schema::BsonTypeName::Long),
            "dec".to_string() => make_simple_schema(json_schema::BsonTypeName::Decimal),
            "bin".to_string() => make_simple_schema(json_schema::BsonTypeName::BinData),
            "oid".to_string() => make_simple_schema(json_schema::BsonTypeName::ObjectId),
            "boo".to_string() => make_simple_schema(json_schema::BsonTypeName::Bool),
            "dat".to_string() => make_simple_schema(json_schema::BsonTypeName::Date),
            "reg".to_string() => make_simple_schema(json_schema::BsonTypeName::Regex),
            "dbp".to_string() => make_simple_schema(json_schema::BsonTypeName::DbPointer),
            "jav".to_string() => make_simple_schema(json_schema::BsonTypeName::Javascript),
            "sym".to_string() => make_simple_schema(json_schema::BsonTypeName::Symbol),
            "jws".to_string() => make_simple_schema(json_schema::BsonTypeName::JavascriptWithScope),
            "tim".to_string() => make_simple_schema(json_schema::BsonTypeName::Timestamp),
            "min".to_string() => make_simple_schema(json_schema::BsonTypeName::MinKey),
            "max".to_string() => make_simple_schema(json_schema::BsonTypeName::MaxKey),
            "und".to_string() => make_simple_schema(json_schema::BsonTypeName::Undefined),
        }),
        required: Some(vec!["a".to_string(), "b".to_string()]),
        additional_properties: Some(false),
        ..Default::default()
    };

    let collection_schemas = name_generator()
        .take(num_collections * 2) // creating twice as many as request
        .map(|n| (n, collection_schema.clone()))
        .collect::<BTreeMap<String, json_schema::Schema>>();

    let full_schema: BTreeMap<&str, BTreeMap<&str, BTreeMap<String, json_schema::Schema>>> = map! {
        "catalog_schema" => map! {
            db => collection_schemas,
        },
    };

    let file = File::create(PathBuf::from(target_file)).unwrap();
    let w = BufWriter::new(file);
    serde_json::to_writer_pretty(w, &full_schema).unwrap()
}
