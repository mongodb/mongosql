use super::*;

/// Most tests use the same collection source and need to specify the
/// collection schema for the test to work. This helper allows easy
/// definition of that collection schema.
fn make_catalog(s: Schema) -> Catalog {
    Catalog::new(map! {
        Namespace {database: "test".into(), collection: "foo".into()} => s,
    })
}

test_algebrize!(
    simple,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Unwind(mir::Unwind {
        source: Box::new(mir_source_foo()),
        path: mir::FieldPath {
            key: ("foo", 0u16).into(),
            fields: vec!["arr".to_string()],
            is_nullable: false,
        },
        index: None,
        outer: false,
        cache: SchemaCache::new(),
        is_prefiltered: false,
    })),
    input = Some(ast::Datasource::Unwind(ast::UnwindSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![ast::UnwindOption::Path(ast::Expression::Identifier(
            "arr".into(),
        ))]
    })),
    catalog = make_catalog(Schema::Document(Document {
        keys: map! {
            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
        },
        required: set! {"arr".into()},
        additional_properties: false,
        ..Default::default()
    })),
);
test_algebrize!(
    all_opts,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Unwind(mir::Unwind {
        source: Box::new(mir_source_foo()),
        path: mir::FieldPath {
            key: ("foo", 0u16).into(),
            fields: vec!["arr".to_string()],
            is_nullable: false,
        },
        index: Some("i".into()),
        outer: true,
        cache: SchemaCache::new(),
        is_prefiltered: false,
    })),
    input = Some(ast::Datasource::Unwind(ast::UnwindSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![
            ast::UnwindOption::Path(ast::Expression::Identifier("arr".into())),
            ast::UnwindOption::Index("i".into()),
            ast::UnwindOption::Outer(true),
        ]
    })),
    catalog = make_catalog(Schema::Document(Document {
        keys: map! {
            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
        },
        required: set! {"arr".into()},
        additional_properties: false,
        ..Default::default()
    })),
);
test_algebrize!(
    compound_path,
    method = algebrize_from_clause,
    expected = Ok(mir::Stage::Unwind(mir::Unwind {
        source: Box::new(mir_source_foo()),
        path: mir::FieldPath {
            key: ("foo", 0u16).into(),
            fields: vec!["doc".to_string(), "arr".to_string()],
            is_nullable: false,
        },
        index: None,
        outer: false,
        cache: SchemaCache::new(),
        is_prefiltered: false,
    })),
    input = Some(ast::Datasource::Unwind(ast::UnwindSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
            ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("doc".into())),
                subpath: "arr".into(),
            }
        ))]
    })),
    catalog = make_catalog(Schema::Document(Document {
        keys: map! {
            "doc".into() => Schema::Document(Document {
                keys: map! {
                    "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
                },
                required: set!{"arr".into()},
                additional_properties: false,
                ..Default::default()
                }),
        },
        required: set! {"doc".into()},
        additional_properties: false,
        ..Default::default()
    })),
);
test_algebrize!(
    duplicate_opts,
    method = algebrize_from_clause,
    expected = Err(Error::DuplicateUnwindOption(ast::UnwindOption::Path(
        ast::Expression::Identifier("dup".into())
    ))),
    expected_error_code = 3027,
    input = Some(ast::Datasource::Unwind(ast::UnwindSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![
            ast::UnwindOption::Path(ast::Expression::Identifier("arr".into())),
            ast::UnwindOption::Path(ast::Expression::Identifier("dup".into())),
        ]
    })),
    catalog = make_catalog(ANY_DOCUMENT.clone()),
);
test_algebrize!(
    missing_path,
    method = algebrize_from_clause,
    expected = Err(Error::NoUnwindPath),
    expected_error_code = 3028,
    input = Some(ast::Datasource::Unwind(ast::UnwindSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![]
    })),
    catalog = make_catalog(ANY_DOCUMENT.clone()),
);
test_algebrize!(
    invalid_path,
    method = algebrize_from_clause,
    expected = Err(Error::InvalidUnwindPath),
    expected_error_code = 3029,
    input = Some(ast::Datasource::Unwind(ast::UnwindSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
            ast::SubpathExpr {
                expr: Box::new(ast::Expression::Document(vec![ast::DocumentPair {
                    key: "arr".into(),
                    value: ast::Expression::Array(vec![
                        ast::Expression::Literal(ast::Literal::Integer(1)),
                        ast::Expression::Literal(ast::Literal::Integer(2)),
                        ast::Expression::Literal(ast::Literal::Integer(3))
                    ])
                }])),
                subpath: "arr".into(),
            }
        )),]
    })),
    catalog = make_catalog(ANY_DOCUMENT.clone()),
);
test_algebrize!(
    correlated_path_disallowed,
    method = algebrize_from_clause,
    expected = Err(Error::FieldNotFound(
        "bar".into(),
        Some(vec!["arr".into()]),
        ClauseType::From,
        1u16,
    )),
    expected_error_code = 3008,
    input = Some(ast::Datasource::Unwind(ast::UnwindSource {
        datasource: Box::new(AST_SOURCE_FOO.clone()),
        options: vec![ast::UnwindOption::Path(ast::Expression::Subpath(
            ast::SubpathExpr {
                expr: Box::new(ast::Expression::Identifier("bar".into())),
                subpath: "arr".into(),
            }
        )),]
    })),
    env = map! {
        ("bar", 0u16).into() => Schema::Document( Document {
            keys: map! {
                "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
            },
            required: set!{ "arr".into() },
            additional_properties: false,
            ..Default::default()
            }),
    },
    catalog = make_catalog(Schema::Document(Document {
        keys: map! {
            "arr".into() => Schema::Array(Box::new(Schema::Atomic(Atomic::Integer))),
        },
        required: set! {"arr".into()},
        additional_properties: false,
        ..Default::default()
    })),
);
