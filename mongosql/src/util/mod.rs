use crate::air;
use lazy_static::lazy_static;
pub use mongosql_datastructures::unique_linked_hash_map;

pub const ROOT_NAME: &str = "ROOT";

lazy_static! {
    pub static ref ROOT: air::Expression = air::Expression::Variable(air::Variable {
        parent: None,
        name: ROOT_NAME.to_string()
    });
    // https://www.mongodb.com/docs/manual/reference/operator/query/regex/#mongodb-query-op.-options
    // `s` allows '.' to match all characters including newline characters
    // `i` denotes case insensitivity
    pub static ref LIKE_OPTIONS: String = "si".to_string();
}

#[macro_export]
macro_rules! map {
	($($key:expr => $val:expr),* $(,)?) => {
		std::iter::Iterator::collect([
			$({
				($key, $val)
			},)*
		].into_iter())
	};
}

#[macro_export]
macro_rules! set {
	($($val:expr),* $(,)?) => {
		std::iter::Iterator::collect([
			$({
				($val)
			},)*
		].into_iter())
	};
}

// The unchecked version unwraps insertions. This should only be used for testing.
#[cfg(test)]
#[macro_export]
macro_rules! unchecked_unique_linked_hash_map {
	($($key:expr => $val:expr),* $(,)?) => {{
            #[allow(unused_mut)]
            let mut out = mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap::new();
            $(
                out.insert($key, $val).unwrap();
            )*
            out
	}};
}

#[cfg(test)]
use crate::mir;
#[cfg(test)]
use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

/// mir_project_collection creates a mir::Project stage with a mir::Collection
/// source. A collection_name must be specified; database name, expected name,
/// and scope are optional. If no database name is provided, the default name
/// is "test_db". If no expected name is provided, the default name is the
/// collection name. If no scope is provided, the default scope is 0. The
/// Project stage's expression contains one mapping -- the collection at the
/// scope mapped to a Reference with the same data.
#[cfg(test)]
pub(crate) fn mir_project_collection(
    db_name: Option<&str>,
    collection_name: &str,
    expected_name: Option<&str>,
    scope: Option<u16>,
) -> Box<mir::Stage> {
    let db_name = db_name.unwrap_or("test_db");
    let expected_name = expected_name.unwrap_or(collection_name);
    let scope = scope.unwrap_or(0u16);
    Box::new(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Collection(mir::Collection {
            db: db_name.into(),
            collection: collection_name.into(),
            cache: mir::schema::SchemaCache::new(),
        })),
        expression: BindingTuple(map! {
            Key::named(expected_name, scope) => mir::Expression::Reference(mir::ReferenceExpr {
                key: Key::named(collection_name, scope),
            }),
        }),
        cache: mir::schema::SchemaCache::new(),
    }))
}

/// mir_collection creates a mir::Collection with the specified db and
/// collection names. It does not wrap the Collection in a Project stage.
#[cfg(test)]
pub(crate) fn mir_collection(db_name: &str, collection_name: &str) -> Box<mir::Stage> {
    Box::new(mir::Stage::Collection(mir::Collection {
        db: db_name.to_string(),
        collection: collection_name.to_string(),
        cache: mir::schema::SchemaCache::new(),
    }))
}

#[cfg(test)]
pub(crate) fn mir_project_bot_collection(collection_name: &str) -> Box<mir::Stage> {
    Box::new(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: collection_name.into(),
            cache: mir::schema::SchemaCache::new(),
        })),
        expression: BindingTuple(map! {
            Key::bot(0u16) => mir::Expression::Reference((collection_name, 0u16).into()),
        }),
        cache: mir::schema::SchemaCache::new(),
    }))
}

#[cfg(test)]
pub(crate) fn mir_field_access(
    key_name: &str,
    field_name: &str,
    is_nullable: bool,
) -> Box<mir::Expression> {
    Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
            key: make_key(key_name),
        })),
        field: field_name.to_string(),
        is_nullable,
    }))
}

#[cfg(test)]
pub(crate) fn mir_field_access_multi_part(
    key_name: &str,
    field_names: Vec<&str>,
    is_nullable: bool,
) -> Box<mir::Expression> {
    field_names.into_iter().fold(
        Box::new(mir::Expression::Reference(mir::ReferenceExpr {
            key: make_key(key_name),
        })),
        |acc, field_name| {
            Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                expr: acc,
                field: field_name.to_string(),
                is_nullable,
            }))
        },
    )
}

#[cfg(test)]
pub(crate) fn mir_field_path(datasource_name: &str, field_names: Vec<&str>) -> mir::FieldPath {
    mir::FieldPath::new(
        make_key(datasource_name),
        field_names.into_iter().map(String::from).collect(),
    )
}

#[cfg(test)]
fn make_key(key_name: &str) -> Key {
    if key_name == "__bot__" {
        Key::bot(0)
    } else {
        Key::named(key_name, 0)
    }
}

/// air_project_collection creates an air::Project stage with an air::Collection
/// source. A collection_name must be specified; database name and expected name
/// are optional. If no database name is provided, the default name is "test_db"
/// and if no expected name is provided, the default is the collection name. The
/// Project stage's specifications contain one mapping -- the expected name
/// mapped to the ROOT variable.
#[cfg(test)]
pub(crate) fn air_project_collection(
    db_name: Option<&str>,
    collection_name: &str,
    expected_name: Option<&str>,
) -> Box<air::Stage> {
    let db_name = db_name.unwrap_or("test_db");
    let expected_name = expected_name.unwrap_or(collection_name);
    Box::new(air::Stage::Project(air::Project {
        source: air_collection_stage(db_name, collection_name),
        specifications: unchecked_unique_linked_hash_map! {
            expected_name.to_string() => air::ProjectItem::Assignment(ROOT.clone()),
        },
    }))
}

/// air_collection_raw creates an air::Collection with the specified db and
/// collection names. It does not wrap the Collection in a Project stage. It
/// is returned as an air::Collection, not an air::Stage or Box<air::Stage>.
#[cfg(test)]
pub(crate) fn air_collection_raw(db_name: &str, collection_name: &str) -> air::Collection {
    air::Collection {
        db: db_name.to_string(),
        collection: collection_name.to_string(),
    }
}

/// air_collection_stage creates an air::Collection with the specified db and
/// collection names. It does not wrap the Collection in Project stage. It is
/// returned as a Box<air::Stage>.
#[cfg(test)]
pub(crate) fn air_collection_stage(db_name: &str, collection_name: &str) -> Box<air::Stage> {
    Box::new(air::Stage::Collection(air_collection_raw(
        db_name,
        collection_name,
    )))
}

/// air_documents_stage creates an air::Documents with the specified array
/// vector. It is returned as a Box<air::Stage>.
#[cfg(test)]
pub(crate) fn air_documents_stage(array: Vec<air::Expression>) -> Box<air::Stage> {
    Box::new(air::Stage::Documents(air::Documents { array }))
}

#[cfg(test)]
pub(crate) fn air_variable_from_root(rest: &str) -> air::Expression {
    let full_path = format!("{}.{}", ROOT_NAME, rest);
    air::Expression::Variable(full_path.into())
}

#[cfg(test)]
pub(crate) fn sql_options_exclude_namespaces() -> crate::options::SqlOptions {
    crate::options::SqlOptions {
        exclude_namespaces: crate::options::ExcludeNamespacesOption::ExcludeNamespaces,
        ..Default::default()
    }
}

const DEFAULT_ESCAPE: char = '\\';

pub(crate) fn convert_sql_pattern(pattern: String, escape: Option<char>) -> String {
    let escape = escape.unwrap_or(DEFAULT_ESCAPE);
    const REGEX_CHARS_TO_ESCAPE: [char; 12] =
        ['.', '^', '$', '*', '+', '?', '(', ')', '[', '{', '\\', '|'];
    let mut regex = "^".to_string();
    let mut escaped = false;
    for c in pattern.chars() {
        if !escaped & (c == escape) {
            escaped = true;
            continue;
        }
        match c {
            '_' => {
                let s = if escaped { '_' } else { '.' };
                regex.push(s)
            }
            '%' => {
                if escaped {
                    regex.push('%');
                } else {
                    regex.push_str(".*");
                }
            }
            _ => {
                if REGEX_CHARS_TO_ESCAPE.contains(&c) {
                    regex.push('\\');
                }
                regex.push_str(c.to_string().as_str())
            }
        }
        escaped = false;
    }
    regex.push('$');
    regex.to_string()
}

#[cfg(test)]
mod test_convert_sql_pattern {
    use super::{convert_sql_pattern, DEFAULT_ESCAPE};
    macro_rules! test_convert_sql_pattern {
        ($func_name:ident, expected = $expected:expr, input = $input:expr, escape = $escape:expr) => {
            #[test]
            fn $func_name() {
                let input = $input;
                let expected = $expected;
                let escape = $escape;
                let actual = convert_sql_pattern(input.to_string(), escape);
                assert_eq!(expected, actual)
            }
        };
    }

    test_convert_sql_pattern!(
        no_special_sql_characters,
        expected = "^abc$",
        input = "abc",
        escape = Some(DEFAULT_ESCAPE)
    );

    test_convert_sql_pattern!(
        unescaped_special_sql_characters,
        expected = "^a.b.*c$",
        input = "a_b%c",
        escape = Some(DEFAULT_ESCAPE)
    );

    test_convert_sql_pattern!(
        escaped_special_sql_characters,
        expected = "^a_b%c$",
        input = "a\\_b\\%c",
        escape = Some(DEFAULT_ESCAPE)
    );

    test_convert_sql_pattern!(
        escaped_escape_character,
        expected = "^a\\\\.b%c$",
        input = "a\\\\_b\\%c",
        escape = Some(DEFAULT_ESCAPE)
    );

    test_convert_sql_pattern!(
        default_escape_character,
        expected = "^a\\\\.b%c$",
        input = "a\\\\_b\\%c",
        escape = None
    );

    test_convert_sql_pattern!(
        special_mql_characters,
        expected = "^\\.\\^\\$\\*\\+\\?\\(\\)\\[\\{\\\\\\|$",
        input = ".^$*+?()[{\\|",
        escape = Some('e')
    );
}
