use super::*;
use mongosql::{
    schema::{Atomic, Schema},
    set,
};

#[test]
fn schema_for_empty_bson_array() {
    assert_eq!(
        Schema::Array(Schema::Atomic(Atomic::Null).into()),
        schema_for_bson(&Bson::Array(vec![])),
    );
}

#[test]
fn schema_for_bson_array_single() {
    let schema = schema_for_bson(&Bson::Array(vec![Bson::Double(1.0)]));
    assert_eq!(
        Schema::Array(Schema::Atomic(Atomic::Double).into(),),
        schema,
    );
}

#[test]
fn schema_for_bson_array_multi_same() {
    let schema = schema_for_bson(&Bson::Array(vec![Bson::Double(1.0), Bson::Double(42.0)]));
    assert_eq!(
        Schema::Array(Schema::Atomic(Atomic::Double).into(),),
        schema,
    );
}

#[test]
fn schema_for_bson_array_multi_differ() {
    let schema = schema_for_bson(&Bson::Array(vec![Bson::Double(1.0), Bson::Null]));
    assert_eq!(
        Schema::Array(
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Null),
            ])
            .into(),
        ),
        schema,
    );
}

#[test]
fn schema_for_nested_bson_array() {
    let b = Bson::Array(vec![
        Bson::Array(vec![Bson::Double(1.0), Bson::Null]),
        Bson::Array(vec![Bson::String("foo".into())]),
    ]);
    let schema = schema_for_bson(&b);
    assert_eq!(
        Schema::Array(
            Schema::Array(
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::String),
                ])
                .into(),
            )
            .into(),
        ),
        schema,
    );
}
