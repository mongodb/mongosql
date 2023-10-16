use crate::{
    catalog::Namespace,
    map,
    mir::{
        schema::{Error as mir_error, SchemaCache},
        *,
    },
    schema::{Atomic, ResultSet, Schema, ANY_DOCUMENT},
    set, test_schema, unchecked_unique_linked_hash_map,
};

fn true_mir() -> Expression {
    Expression::Literal(LiteralValue::Boolean(true).into())
}

fn test_source() -> Stage {
    Stage::Collection(Collection {
        db: "test".into(),
        collection: "foo".into(),
        cache: schema::SchemaCache::new(),
    })
}

test_schema!(
    boolean_condition_allowed,
    expected_pat = Ok(ResultSet { .. }),
    input = Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: true_mir(),
        cache: SchemaCache::new(),
    }),
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    null_condition_allowed,
    expected_pat = Ok(ResultSet { .. }),
    input = Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: Expression::Literal(LiteralValue::Null.into()),
        cache: SchemaCache::new(),
    }),
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    missing_condition_allowed,
    expected_pat = Ok(ResultSet { .. }),
    input = Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: Expression::Reference(("m", 0u16).into()),
        cache: SchemaCache::new(),
    }),
    schema_env = map! {("m", 0u16).into() => Schema::Missing},
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    non_boolean_condition_disallowed,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "filter condition",
        required: Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
        ]),
        found: Schema::Atomic(Atomic::Integer),
    }),
    input = Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: Expression::Literal(LiteralValue::Integer(123).into()),
        cache: SchemaCache::new(),
    }),
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    possible_non_boolean_condition_disallowed,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "filter condition",
        required: Schema::AnyOf(set![
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
        ]),
        found: Schema::Any,
    }),
    input = Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: Expression::FieldAccess(FieldAccess {
            expr: Expression::Reference(("foo", 0u16).into()).into(),
            field: "bar".into(),
            cache: SchemaCache::new(),
            is_nullable: false,
        }),
        cache: SchemaCache::new(),
    }),
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    source_fails_schema_check,
    expected_pat = Err(mir_error::SchemaChecking {
        name: "array datasource items",
        ..
    }),
    input = Stage::Filter(Filter {
        source: Stage::Array(ArraySource {
            alias: "arr".into(),
            array: vec![Expression::Literal(LiteralValue::Null.into())],
            cache: SchemaCache::new(),
        })
        .into(),
        condition: true_mir(),
        cache: SchemaCache::new(),
    }),
);

test_schema!(
    condition_fails_schema_check,
    expected_error_code = 1000,
    expected = Err(mir_error::DatasourceNotFoundInSchemaEnv(
        ("abc", 0u16).into()
    )),
    input = Stage::Filter(Filter {
        source: Box::new(test_source()),
        condition: Expression::Reference(("abc", 0u16).into()),
        cache: SchemaCache::new(),
    }),
    catalog = Catalog::new(map! {
        Namespace {db: "test".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
    }),
);

test_schema!(
    min_size_reduced_to_zero_max_size_preserved,
    expected_pat = Ok(ResultSet{
        min_size: 0,
        max_size: Some(1),
        ..
    }),
    input = Stage::Filter(Filter {
        condition: true_mir(),
        source: Stage::Array(ArraySource {
            alias: "arr".into(),
            array: vec![Expression::Document(unchecked_unique_linked_hash_map!{"a".into() => Expression::Literal(LiteralValue::Null.into()),}.into())],
            cache: SchemaCache::new(),
        }).into(),
        cache: SchemaCache::new(),
    }),
);
