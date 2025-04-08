mod field_not_found {
    test_user_error_messages! {
        no_found_fields,
        input = Error::FieldNotFound("x".into(), None, ClauseType::Select, 1u16),
        expected = "Field `x` of the `SELECT` clause at the 1 scope level not found.".to_string()
    }

    test_user_error_messages! {
        suggestions,
        input = Error::FieldNotFound("foo".into(), Some(vec!["feo".to_string(), "fooo".to_string(), "aaa".to_string(), "bbb".to_string()]), ClauseType::Where, 1u16),
        expected =  "Field `foo` not found in the `WHERE` clause at the 1 scope level. Did you mean: feo, fooo".to_string()
    }

    test_user_error_messages! {
        no_suggestions,
        input = Error::FieldNotFound("foo".into(), Some(vec!["aaa".to_string(), "bbb".to_string(), "ccc".to_string()]), ClauseType::Having, 16u16),
        expected = "Field `foo` in the `HAVING` clause at the 16 scope level not found.".to_string()
    }

    test_user_error_messages! {
        exact_match_found,
        input = Error::FieldNotFound("foo".into(), Some(vec!["foo".to_string()]), ClauseType::GroupBy, 0u16),
        expected = "Unexpected edit distance of 0 found with input: foo and expected: [\"foo\"]"
    }
}

mod derived_datasource_overlapping_keys {
    use crate::{
        map,
        schema::{Atomic, Document, Schema},
        set,
    };

    test_user_error_messages! {
    derived_datasource_overlapping_keys,
    input = Error::DerivedDatasourceOverlappingKeys(
        Schema::Document(Document {
            keys: map! {
                "bar".into() => Schema::Atomic(Atomic::Integer),
                "baz".into() => Schema::Atomic(Atomic::Integer),
                "foo1".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {
                "bar".into(),
                "baz".into(),
                "foo1".into(),
            },
            additional_properties: false,
            ..Default::default()
            }).into(),
        Schema::Document(Document {
            keys: map! {
                "bar".into() => Schema::Atomic(Atomic::Integer),
                "baz".into() => Schema::Atomic(Atomic::Integer),
                "foo2".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {
                "bar".into(),
                "baz".into(),
                "foo2".into(),
            },
        additional_properties: false,
        ..Default::default()
        }).into(),
        "foo".into(),
        crate::schema::Satisfaction::Must,
    ),
    expected = "Derived datasource `foo` has the following overlapping keys: bar, baz"
    }
}

mod ambiguous_field {
    test_user_error_messages! {
        ambiguous_field,
        input = Error::AmbiguousField("foo".into(), ClauseType::Select, 0u16),
        expected = "Field `foo` in the `SELECT` clause at the 0 scope level exists in multiple datasources and is ambiguous. Please qualify."
    }
}

mod cannot_enumerate_all_field_paths {
    test_user_error_messages! {
        cannot_enumerate_all_field_paths,
        input = Error::CannotEnumerateAllFieldPaths(crate::schema::Schema::Any.into()),
        expected = "Insufficient schema information."
    }
}
