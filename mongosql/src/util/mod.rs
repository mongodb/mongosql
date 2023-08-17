use crate::air;
use lazy_static::lazy_static;
lazy_static! {
    pub static ref ROOT_NAME: String = "ROOT".to_string();
    pub static ref ROOT: air::Expression = air::Expression::Variable(air::Variable {
        parent: None,
        name: "ROOT".into()
    });
}

pub use mongosql_datastructures::unique_linked_hash_map;
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
