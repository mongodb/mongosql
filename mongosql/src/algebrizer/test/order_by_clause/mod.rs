use super::catalog;
use crate::{
    ast, map, mir,
    mir::schema::SchemaCache,
    schema::{Atomic, Document, Schema},
    set, unchecked_unique_linked_hash_map,
};

fn source() -> mir::Stage {
    mir::Stage::Collection(mir::Collection {
        db: "test".into(),
        collection: "baz".into(),
        cache: SchemaCache::new(),
    })
}

test_algebrize!(
    asc_and_desc,
    method = algebrize_order_by_clause,
    expected = Ok(mir::Stage::Sort(mir::Sort {
        source: Box::new(source()),
        specs: vec![
            mir::SortSpecification::Asc(mir::FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["a".to_string()],
                is_nullable: false,
            }),
            mir::SortSpecification::Desc(mir::FieldPath {
                key: ("foo", 0u16).into(),
                fields: vec!["b".to_string()],
                is_nullable: false,
            })
        ],
        cache: SchemaCache::new(),
    })),
    input = Some(ast::OrderByClause {
        sort_specs: vec![
            ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".to_string())),
                    subpath: "a".to_string()
                })),
                direction: ast::SortDirection::Asc
            },
            ast::SortSpec {
                key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                    expr: Box::new(ast::Expression::Identifier("foo".to_string())),
                    subpath: "b".to_string()
                })),
                direction: ast::SortDirection::Desc
            }
        ],
    }),
    source = source(),
    env = map! {
        ("foo", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::String),
            },
            required: set!{},
            additional_properties: false,
            ..Default::default()
            }),
    },
    catalog = catalog(vec![("test", "baz")]),
);

test_algebrize!(
    sort_key_from_source,
    method = algebrize_order_by_clause,
    expected = Ok(mir::Stage::Sort(mir::Sort {
        source: Box::new(mir::Stage::Array(mir::ArraySource {
            array: vec![mir::Expression::Document(
                unchecked_unique_linked_hash_map! {
                    "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
                }
                .into(),
            )],
            alias: "arr".into(),
            cache: SchemaCache::new(),
        })),
        specs: vec![mir::SortSpecification::Asc(mir::FieldPath {
            key: ("arr", 0u16).into(),
            fields: vec!["a".to_string()],
            is_nullable: false,
        }),],
        cache: SchemaCache::new(),
    })),
    input = Some(ast::OrderByClause {
        sort_specs: vec![ast::SortSpec {
            key: ast::SortKey::Simple(ast::Expression::Subpath(ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("arr".to_string())),
                subpath: "a".to_string()
            })),
            direction: ast::SortDirection::Asc
        },],
    }),
    source = mir::Stage::Array(mir::ArraySource {
        array: vec![mir::Expression::Document(
            unchecked_unique_linked_hash_map! {
                "a".into() => mir::Expression::Literal(mir::LiteralValue::Integer(1))
            }
            .into(),
        )],
        alias: "arr".into(),
        cache: SchemaCache::new(),
    }),
);
