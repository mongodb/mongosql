use super::catalog;
use crate::{
    ast, map,
    mir::{self, binding_tuple::Key, schema::SchemaCache, Expression, Project, Stage},
    multimap,
    schema::ANY_DOCUMENT,
    unchecked_unique_linked_hash_map,
    usererror::UserError,
};

fn source() -> mir::Stage {
    mir::Stage::Collection(mir::Collection {
        db: "test".into(),
        collection: "baz".into(),
        cache: SchemaCache::new(),
    })
}

test_algebrize!(
    select_values_distinct,
    method = algebrize_select_clause,
    expected = Ok(Stage::Project(Project {
        is_add_fields: false,
        source: Box::new(Stage::Group(mir::Group {
            source: Box::new(Stage::Project(Project {
                is_add_fields: false,
                source: Box::new(source()),
                expression: map! {
                    ("foo", 1u16).into() => Expression::Reference(("foo", 0u16).into()),
                    ("bar", 1u16).into() => Expression::Reference(("bar", 0u16).into()),
                },
                cache: SchemaCache::new(),
            })),
            keys: vec![
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "__groupKey0".into(),
                    expr: Expression::Reference(("bar", 1u16).into()),
                }),
                mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                    alias: "__groupKey1".into(),
                    expr: Expression::Reference(("foo", 1u16).into()),
                }),
            ],
            aggregations: vec![],
            cache: SchemaCache::new(),
            scope: 1,
        })),
        expression: map! {
            ("bar", 1u16).into() => Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(Expression::Reference(Key::bot(1u16).into())),
            field: "__groupKey0".into(),
                is_nullable: true,
            }),
            ("foo", 1u16).into() => Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(Expression::Reference(Key::bot(1u16).into())),
            field: "__groupKey1".into(),
                is_nullable: true,
            }),
        },
        cache: SchemaCache::new(),
    })),
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::Distinct,
        body: ast::SelectBody::Values(vec![
            ast::SelectValuesExpression::Substar("foo".into()),
            ast::SelectValuesExpression::Substar("bar".into())
        ]),
    },
    source = source(),
    env = map! {
        ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        ("bar", 0u16).into() => ANY_DOCUMENT.clone(),
    },
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);

test_algebrize!(
    select_star_distinct,
    method = algebrize_select_clause,
    expected = Ok(Stage::Project(Project {
        is_add_fields: false,
        cache: SchemaCache::new(),
        source: Box::new(Stage::Group(mir::Group {
            source: Box::new(source()),
            keys: vec![mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                alias: "__groupKey0".into(),
                expr: Expression::Reference(("baz", 1u16).into()),
            }),],
            aggregations: vec![],
            cache: SchemaCache::new(),
            scope: 1,
        })),
        expression: map! {
            ("baz", 1u16).into() => Expression::FieldAccess(mir::FieldAccess {
                expr: Box::new(Expression::Reference(Key::bot(1u16).into())),
            field: "__groupKey0".into(),
                is_nullable: true,
            })
        },
    })),
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::Distinct,
        body: ast::SelectBody::Standard(vec![ast::SelectExpression::Star])
    },
    source = source(),
    env = map! {
        ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
    },
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);
test_algebrize!(
    select_duplicate_bot,
    method = algebrize_select_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(source()),
        expression: map! {
            Key::bot(1u16) => mir::Expression::Document(mir::DocumentExpr {
                document: unchecked_unique_linked_hash_map!{},
            })
        },
        cache: SchemaCache::new(),
    })),
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::All,
        body: ast::SelectBody::Values(vec![
            ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
            ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
        ]),
    },
    source = source(),
    env = map! {},
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);
test_algebrize!(
    select_duplicate_doc_key_a,
    method = algebrize_select_clause,
    expected = Err(Error::DuplicateDocumentKey("a".into())),
    expected_error_code = 3023,
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::All,
        body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
            ast::Expression::Document(multimap! {
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
                "a".into() => ast::Expression::Literal(ast::Literal::Integer(42)),
            },)
        ),]),
    },
    source = source(),
    env = map! {},
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);
test_algebrize!(
    select_bot_and_double_substar,
    method = algebrize_select_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(source()),
        expression: map! {
            Key::bot(1u16) => mir::Expression::Document(unchecked_unique_linked_hash_map!{}.into()),
            ("bar", 1u16).into() => mir::Expression::Reference(("bar", 1u16).into()),
            ("foo", 1u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })),
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::All,
        body: ast::SelectBody::Values(vec![
            ast::SelectValuesExpression::Substar("bar".into()),
            ast::SelectValuesExpression::Expression(ast::Expression::Document(multimap! {},)),
            ast::SelectValuesExpression::Substar("foo".into()),
        ]),
    },
    source = source(),
    env = map! {
        ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
        ("bar", 1u16).into() => ANY_DOCUMENT.clone(),
    },
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);
test_algebrize!(
    select_value_expression_must_be_document,
    method = algebrize_select_clause,
    expected = Err(Error::SchemaChecking(
        crate::mir::schema::Error::SchemaChecking {
            name: "project datasource",
            required: ANY_DOCUMENT.clone().into(),
            found: crate::schema::Schema::Atomic(crate::schema::Atomic::String).into(),
        }
    )),
    expected_error_code = 1002,
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::All,
        body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Expression(
            ast::Expression::StringConstructor("foo".into())
        ),]),
    },
    source = source(),
    env = map! {},
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);
test_algebrize!(
    select_duplicate_substar,
    method = algebrize_select_clause,
    expected = Err(Error::DuplicateKey(("foo", 1u16).into())),
    expected_error_code = 3020,
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::All,
        body: ast::SelectBody::Values(vec![
            ast::SelectValuesExpression::Substar("foo".into()),
            ast::SelectValuesExpression::Substar("foo".into()),
        ]),
    },
    source = source(),
    env = map! {
        ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
    },
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);
test_algebrize!(
    select_substar_body,
    method = algebrize_select_clause,
    expected = Ok(mir::Stage::Project(mir::Project {
        is_add_fields: false,
        source: Box::new(source()),
        expression: map! {
            ("foo", 1u16).into() => mir::Expression::Reference(("foo", 0u16).into()),
        },
        cache: SchemaCache::new(),
    })),
    input = ast::SelectClause {
        set_quantifier: ast::SetQuantifier::All,
        body: ast::SelectBody::Values(vec![ast::SelectValuesExpression::Substar("foo".into()),]),
    },
    source = source(),
    env = map! {
        ("foo", 0u16).into() => ANY_DOCUMENT.clone(),
    },
    catalog = catalog(vec![("test", "baz")]),
    is_add_fields = false,
);
