use super::{catalog, mir_source_foo};
use crate::{ast, mir, mir::schema::SchemaCache};

fn true_mir() -> mir::Expression {
    mir::Expression::Literal(mir::LiteralValue::Boolean(true))
}
const TRUE_AST: ast::Expression = ast::Expression::Literal(ast::Literal::Boolean(true));

test_algebrize!(
    simple,
    method = algebrize_filter_clause,
    expected = Ok(mir::Stage::Filter(mir::Filter {
        source: Box::new(mir_source_foo()),
        condition: true_mir(),
        cache: SchemaCache::new(),
    })),
    input = Some(TRUE_AST),
    source = mir_source_foo(),
    catalog = catalog(vec![("test", "foo")]),
);
test_algebrize!(
    none,
    method = algebrize_filter_clause,
    expected = Ok(mir_source_foo()),
    input = None,
    source = mir_source_foo(),
    catalog = catalog(vec![("test", "foo")]),
);
test_algebrize!(
    one_converted_to_bool,
    method = algebrize_filter_clause,
    expected = Ok(mir::Stage::Filter(mir::Filter {
        source: Box::new(mir_source_foo()),
        condition: true_mir(),
        cache: SchemaCache::new(),
    })),
    input = Some(ast::Expression::Literal(ast::Literal::Integer(1))),
    source = mir_source_foo(),
    catalog = catalog(vec![("test", "foo")]),
);
