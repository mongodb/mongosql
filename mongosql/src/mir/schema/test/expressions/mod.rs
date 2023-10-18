mod aggregate; // mir::Aggregate
mod case; // mir::Expression::{SearchedCase, SimpleCase}
mod field_access; // mir::Expression::{FieldAccess, Reference (implicit)}
mod like; // mir::Expression::Like
mod literal; // mir::Expression::{Array, Document, Literal}
mod scalar_function; // mir::Expression::ScalarFunction
mod subquery; // mir::Expression::{Exists, Subquery, SubqueryComparison}
mod type_expr; // mir::Expression::{Cast, TypeAssertion}

mod mql_intrinsic {
    use crate::{
        map,
        mir::{schema::SchemaCache, *},
        schema::{Atomic, Document, Schema},
        set, test_schema,
    };

    test_schema!(
        field_existence_is_always_boolean,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = Expression::MQLIntrinsic(MQLExpression::FieldExistence(FieldExistence {
            field_access: FieldAccess {
                expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                field: "x".to_string(),
                cache: SchemaCache::new(),
                is_nullable: true,
            },
            cache: SchemaCache::new()
        })),
        schema_env = map! {
            ("foo", 0u16).into() => Schema::Document(Document {
                keys: map! {
                    "x".to_string() => Schema::AnyOf(set! {Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null), Schema::Missing}),
                },
                required: set! {},
                additional_properties: false,
            })
        },
    );
}
