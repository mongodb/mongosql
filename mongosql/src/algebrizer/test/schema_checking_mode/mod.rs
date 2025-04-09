use super::catalog;
use crate::{
    ast,
    mir::{schema::SchemaCache, *},
    usererror::UserError,
    Schema,
};

test_algebrize!(
    comparison_fails_in_strict_mode,
    method = algebrize_order_by_clause,
    expected = Err(Error::SchemaChecking(
        schema::Error::SortKeyNotSelfComparable(0, Schema::Any.into())
    )),
    expected_error_code = 1010,
    input = Some(ast::OrderByClause {
        sort_specs: vec![ast::SortSpec {
            key: ast::SortKey::Simple(ast::Expression::Identifier("a".into())),
            direction: ast::SortDirection::Asc
        }]
    }),
    source = Stage::Collection(Collection {
        db: "".into(),
        collection: "test".into(),
        cache: SchemaCache::new(),
    }),
    catalog = catalog(vec![("", "test")]),
);

test_algebrize!(
    comparison_passes_in_relaxed_mode,
    method = algebrize_order_by_clause,
    expected_pat = Ok(_),
    input = Some(ast::OrderByClause {
        sort_specs: vec![ast::SortSpec {
            key: ast::SortKey::Simple(ast::Expression::Identifier("a".into())),
            direction: ast::SortDirection::Asc
        }]
    }),
    source = Stage::Collection(Collection {
        db: "".into(),
        collection: "foo".into(),
        cache: SchemaCache::new(),
    }),
    catalog = catalog(vec![("", "foo")]),
    schema_checking_mode = SchemaCheckingMode::Relaxed,
);
