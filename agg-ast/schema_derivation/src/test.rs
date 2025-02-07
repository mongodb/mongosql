mod get_schema_for_path {
    use crate::{get_or_create_schema_for_path_mut, get_schema_for_path_mut};
    use mongosql::{
        map,
        schema::{Atomic, Document, Schema},
        set,
    };
    use std::collections::BTreeSet;

    macro_rules! test_schema_for_path {
        ($func_name:ident, expected = $expected:expr, input = $input:expr, path = $path:expr) => {
            #[test]
            fn $func_name() {
                let input_cloned = &mut $input.clone();
                let result = get_schema_for_path_mut(input_cloned, $path);
                assert_eq!($expected, result);
            }
        };
    }

    macro_rules! test_get_or_create_schema_for_path {
        ($func_name:ident, expected = $expected:expr, output = $output:expr, input = $input:expr, path = $path: expr) => {
            #[test]
            fn $func_name() {
                let input_cloned = &mut $input.clone();
                let result = get_or_create_schema_for_path_mut(input_cloned, $path);
                assert_eq!($expected, result);
                assert_eq!($output, *input_cloned);
            }
        };
        ($func_name:ident, expected = $expected:expr, input = $input:expr, path = $path:expr) => {
            #[test]
            fn $func_name() {
                let input_cloned = &mut $input.clone();
                let result = get_or_create_schema_for_path_mut(input_cloned, $path);
                assert_eq!($expected, result);
                assert_eq!($input, *input_cloned);
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

    test_get_or_create_schema_for_path!(
        get_or_create_get_two_levels,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                    },
                    additional_properties: false,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_two_levels,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Any,
                    },
                    additional_properties: true,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Any,
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_get_three_levels,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            additional_properties: false,
                            ..Default::default()
                        })
                    },
                    additional_properties: false,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string(), "c".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_three_levels,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Document(Document {
                            keys: map! {
                                "c".to_string() => Schema::Any,
                            },
                            additional_properties: true,
                            ..Default::default()
                        })
                    },
                    additional_properties: true,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Any,
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string(), "c".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_refine_two_levels_any_of,
        expected = Some(&mut Schema::Atomic(Atomic::Integer)),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Atomic(Atomic::Integer),
                    },
                    additional_properties: false,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                        Schema::Document(Document {
                            keys: map! {
                                "b".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            additional_properties: false,
                            ..Default::default()
                        }),
                        Schema::Atomic(Atomic::Integer)
                    )),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );

    test_get_or_create_schema_for_path!(
        get_or_create_create_and_refine_two_levels_any_of,
        expected = Some(&mut Schema::Any),
        output = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::Document(Document {
                    keys: map! {
                        "b".to_string() => Schema::Any,
                    },
                    additional_properties: true,
                    ..Default::default()
                })
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        input = Schema::Document(Document {
            keys: map! {
                "a".to_string() => Schema::AnyOf(set!(
                        Schema::Document(Document {
                            keys: map! {
                                "b".to_string() => Schema::Any,
                            },
                            additional_properties: true,
                            ..Default::default()
                        }),
                        Schema::Atomic(Atomic::Integer)
                    )),
            },
            required: BTreeSet::new(),
            additional_properties: false,
            ..Default::default()
        }),
        path = vec!["a".to_string(), "b".to_string()]
    );
}
