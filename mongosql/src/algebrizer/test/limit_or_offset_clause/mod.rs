use super::{catalog, mir_source_foo, AST_SOURCE_FOO};
use crate::{ast, mir, mir::schema::SchemaCache};

test_algebrize!(
    limit_set,
    method = algebrize_limit_clause,
    expected = Ok(mir::Stage::Limit(mir::Limit {
        source: Box::new(mir_source_foo()),
        limit: 42_u64,
        cache: SchemaCache::new(),
    })),
    input = Some(42_u32),
    source = mir_source_foo(),
    catalog = catalog(vec![("test", "foo")]),
);
test_algebrize!(
    limit_unset,
    method = algebrize_limit_clause,
    expected = Ok(mir_source_foo()),
    input = None,
    source = mir_source_foo(),
);
test_algebrize!(
    offset_set,
    method = algebrize_offset_clause,
    expected = Ok(mir::Stage::Offset(mir::Offset {
        source: Box::new(mir_source_foo()),
        offset: 3,
        cache: SchemaCache::new(),
    })),
    input = Some(3_u32),
    source = mir_source_foo(),
    catalog = catalog(vec![("test", "foo")]),
);
test_algebrize!(
    offset_unset,
    method = algebrize_offset_clause,
    expected = Ok(mir_source_foo()),
    input = None,
    source = mir_source_foo(),
);
test_algebrize!(
    limit_and_offset,
    method = algebrize_select_query,
    expected = Ok(mir::Stage::Limit(mir::Limit {
        source: Box::new(mir::Stage::Offset(mir::Offset {
            source: Box::new(mir_source_foo()),
            offset: 3,
            cache: SchemaCache::new(),
        })),
        limit: 10,
        cache: SchemaCache::new(),
    })),
    input = ast::SelectQuery {
        select_clause: ast::SelectClause {
            set_quantifier: ast::SetQuantifier::All,
            body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
        },
        from_clause: Some(AST_SOURCE_FOO.clone()),
        where_clause: None,
        group_by_clause: None,
        having_clause: None,
        order_by_clause: None,
        limit: Some(10_u32),
        offset: Some(3_u32)
    },
    catalog = catalog(vec![("test", "foo")]),
);
