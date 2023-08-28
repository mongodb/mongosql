use crate::air;
use lazy_static::lazy_static;
pub use mongosql_datastructures::unique_linked_hash_map;

lazy_static! {
    pub static ref ROOT_NAME: String = "ROOT".to_string();
    pub static ref ROOT: air::Expression = air::Expression::Variable(air::Variable {
        parent: None,
        name: ROOT_NAME.clone()
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
#[cfg(test)]
pub(crate) fn mir_collection(collection_name: &str) -> Box<mir::Stage> {
    Box::new(mir::Stage::Project(mir::Project {
        source: Box::new(mir::Stage::Collection(mir::Collection {
            db: "test_db".into(),
            collection: collection_name.into(),
            cache: mir::schema::SchemaCache::new(),
        })),
        expression: BindingTuple(map! {
            Key::named(collection_name, 0u16) => mir::Expression::Reference(mir::ReferenceExpr {
                key: Key::named(collection_name, 0u16),
                cache: mir::schema::SchemaCache::new(),
            }),
        }),
        cache: mir::schema::SchemaCache::new(),
    }))
}

#[cfg(test)]
pub(crate) fn air_collection(collection_name: &str) -> air::Collection {
    air::Collection {
        db: "test_db".into(),
        collection: collection_name.into(),
    }
}

#[cfg(test)]
pub(crate) fn air_db_collection(db_name: &str, collection_name: &str) -> air::Collection {
    air::Collection {
        db: db_name.into(),
        collection: collection_name.into(),
    }
}

#[cfg(test)]
pub(crate) fn air_pipeline_collection(collection_name: &str) -> Box<air::Stage> {
    air::Stage::Project(air::Project {
        source: air::Stage::Collection(air_collection(collection_name)).into(),
        specifications: unchecked_unique_linked_hash_map! {
            collection_name.to_string() => air::ProjectItem::Assignment(ROOT.clone()),
        },
    })
    .into()
}

#[cfg(test)]
pub(crate) fn mir_field_access(key_name: &str, field_name: &str) -> Box<mir::Expression> {
    Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
        expr: Box::new(mir::Expression::Reference(mir::ReferenceExpr {
            key: Key::named(key_name, 0u16),
            cache: mir::schema::SchemaCache::new(),
        })),
        field: field_name.to_string(),
        cache: mir::schema::SchemaCache::new(),
    }))
}

#[cfg(test)]
pub(crate) fn air_variable_from_root(rest: &str) -> air::Expression {
    let full_path = format!("{}.{}", *ROOT_NAME, rest);
    air::Expression::Variable(full_path.into())
}

const DEFAULT_ESCAPE: char = '\\';

pub(crate) fn convert_sql_pattern(pattern: String, escape: Option<String>) -> String {
    let escape = escape
        .map(|e| e.chars().next().unwrap())
        .unwrap_or(DEFAULT_ESCAPE);
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
    use super::convert_sql_pattern;
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
        escape = Some("\\".to_string())
    );

    test_convert_sql_pattern!(
        unescaped_special_sql_characters,
        expected = "^a.b.*c$",
        input = "a_b%c",
        escape = Some("\\".to_string())
    );

    test_convert_sql_pattern!(
        escaped_special_sql_characters,
        expected = "^a_b%c$",
        input = "a\\_b\\%c",
        escape = Some("\\".to_string())
    );

    test_convert_sql_pattern!(
        escaped_escape_character,
        expected = "^a\\\\.b%c$",
        input = "a\\\\_b\\%c",
        escape = Some("\\".to_string())
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
        escape = Some("e".to_string())
    );
}
