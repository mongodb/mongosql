mod get_schema_for_path {
    use crate::get_schema_for_path_mut;
    use mongosql::{
        map,
        schema::{Atomic, Document, Schema},
    };
    use std::collections::BTreeSet;

    macro_rules! test_schema_for_path {
        ($func_name:ident, expected = $expected:expr, input = $input:expr, path = $path:expr) => {
            #[test]
            fn $func_name() {
                let input_cloned = &mut $input.clone();
                let result = get_schema_for_path_mut(input_cloned, $path);
                assert_eq!(result, $expected);
            }
        };
    }

    test_schema_for_path!(
        simple,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
                "c".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string()]
    );

    test_schema_for_path!(
        nested,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            ..Default::default()
                        })
                    },
                    ..Default::default()
                })
            },
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string(), "c".to_string()]
    );

    test_schema_for_path!(
        missing,
        expected = None,
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Atomic(Atomic::Integer),
                "b".to_string() => Schema::Atomic(Atomic::Integer),
                "c".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["d".to_string()]
    );
}
