use crate::{
    map,
    mir::{
        schema::{Error as mir_error, SchemaCache},
        *,
    },
    schema::{Atomic, Document, Schema},
    set, test_schema,
};

test_schema!(
    field_access_accessee_cannot_be_document,
    expected_error_code = 1002,
    expected = Err(mir_error::SchemaChecking {
        name: "FieldAccess",
        required: crate::schema::ANY_DOCUMENT.clone(),
        found: Schema::Atomic(Atomic::Long),
    }),
    input = Expression::FieldAccess(FieldAccess {
        expr: Box::new(Expression::Literal(LiteralValue::Long(1).into())),
        field: "foo".to_string(),
        cache: SchemaCache::new(),
        is_nullable: false,
    }),
);

test_schema!(
    field_access_field_must_not_exist_not_in_document,
    expected_error_code = 1007,
    expected = Err(mir_error::AccessMissingField(
        "foo".to_string(),
        Some(vec!["foof".to_string()])
    )),
    input = Expression::FieldAccess(FieldAccess {
        expr: Box::new(Expression::Reference(("bar", 0u16).into())),
        field: "foo".to_string(),
        cache: SchemaCache::new(),
        is_nullable: false,
    }),
    schema_env = map! {("bar", 0u16).into() => Schema::Document(
        Document {
            keys: map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
            required: set!{"foof".to_string()},
            additional_properties: false,
        }
    ),},
);

test_schema!(
    field_access_field_may_exist,
    expected = Ok(Schema::Any),
    input = Expression::FieldAccess(FieldAccess {
        expr: Box::new(Expression::Reference(("bar", 0u16).into())),
        field: "foo".to_string(),
        cache: SchemaCache::new(),
        is_nullable: false,
    }),
    schema_env = map! {("bar", 0u16).into() => Schema::Document(
        Document {
            keys: map!{"foof".to_string() => Schema::Atomic(Atomic::String)},
            required: set!{"foof".to_string()},
            additional_properties: true,
        }
    ),},
);

test_schema!(
    field_access_field_must_exist,
    expected = Ok(Schema::Atomic(Atomic::String)),
    input = Expression::FieldAccess(FieldAccess {
        expr: Box::new(Expression::Reference(("bar", 0u16).into())),
        field: "foo".to_string(),
        cache: SchemaCache::new(),
        is_nullable: false,
    }),
    schema_env = map! {("bar", 0u16).into() => Schema::Document(
        Document {
            keys: map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
            required: set!{"foo".to_string()},
            additional_properties: false,
        }
    ),},
);

test_schema!(
    field_access_field_must_any_of,
    expected = Ok(Schema::AnyOf(
        set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer)}
    )),
    input = Expression::FieldAccess(FieldAccess {
        expr: Box::new(Expression::Reference(("bar", 0u16).into())),
        field: "foo".to_string(),
        cache: SchemaCache::new(),
        is_nullable: false,
    }),
    schema_env = map! {("bar", 0u16).into() =>
        Schema::AnyOf(set!{
        Schema::Document(
            Document {
                keys: map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foo".to_string()},
                additional_properties: false,
            }
        ),
        Schema::Document(
            Document {
                keys: map!{"foo".to_string() => Schema::Atomic(Atomic::Integer)},
                required: set!{"foo".to_string()},
                additional_properties: false,
            }
        ),
    })},
);

test_schema!(
    field_access_field_must_any_of_with_missing,
    expected = Ok(Schema::AnyOf(
        set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Integer), Schema::Missing}
    )),
    input = Expression::FieldAccess(FieldAccess {
        expr: Box::new(Expression::Reference(("bar", 0u16).into())),
        field: "foo".to_string(),
        cache: SchemaCache::new(),
        is_nullable: false,
    }),
    schema_env = map! {("bar", 0u16).into() =>
        Schema::AnyOf(set!{
        Schema::Document(
            Document {
                keys: map!{"foo".to_string() => Schema::Atomic(Atomic::String)},
                required: set!{"foo".to_string()},
                additional_properties: false,
            }
        ),
        Schema::Document(
            Document {
                keys: map!{"foo".to_string() => Schema::Atomic(Atomic::Integer)},
                required: set!{"foo".to_string()},
                additional_properties: false,
            }
        ),
        Schema::Atomic(Atomic::Integer),
    })},
);
