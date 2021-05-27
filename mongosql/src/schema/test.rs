use crate::schema::{Atomic::*, Document, Satisfaction::*, Schema::*};
use common_macros::{b_tree_map, b_tree_set};

macro_rules! test_satisfies {
    ($func_name:ident, $expected:expr, $self:expr, $other:expr $(,)?) => {
        #[test]
        fn $func_name() {
            let res = $self.satisfies(&$other);
            assert_eq!($expected, res)
        }
    };
}

test_satisfies!(satisfies_any_must_satisfy_any, Must, Any, Any);
test_satisfies!(satisfies_missing_must_satisfy_any, Must, Missing, Any);
test_satisfies!(
    satisfies_missing_must_satisfy_missing,
    Must,
    Missing,
    Missing
);
test_satisfies!(
    satisfies_missing_must_satisfy_any_of,
    Must,
    Missing,
    AnyOf(vec![Missing])
);
test_satisfies!(
    satisfies_one_of_missing_may_satisfy_missing,
    May,
    OneOf(vec![Atomic(String), Missing, Atomic(Int)]),
    Missing
);
test_satisfies!(
    satisfies_any_of_missing_may_satisfy_missing,
    May,
    AnyOf(vec![Atomic(Int), Missing, Atomic(String)]),
    Missing
);
test_satisfies!(
    satisfies_missing_must_not_satisfy_atomic,
    Not,
    Missing,
    Atomic(String)
);
test_satisfies!(
    satisfies_missing_must_not_satisfy_array,
    Not,
    Missing,
    Array(Box::new(Any)),
);
test_satisfies!(
    satisfies_missing_must_not_satisfy_document,
    Not,
    Missing,
    Document(Document {
        keys: b_tree_map![],
        required: b_tree_set![],
        additional_properties: true,
    })
);
test_satisfies!(
    satisfies_missing_must_not_satisfy_any_of,
    Not,
    Missing,
    AnyOf(vec![Atomic(String), Atomic(Int)])
);
test_satisfies!(
    satisfies_missing_must_not_satisfy_one_of,
    Not,
    Missing,
    OneOf(vec![Atomic(String), Atomic(Int)])
);
test_satisfies!(satisfies_atomic_must_satisfy_any, Must, Atomic(String), Any);
test_satisfies!(satisfies_any_may_satisfy_atomic, May, Any, Atomic(String));
test_satisfies!(
    satisfies_array_of_any_does_not_satisfy_atomic,
    Not,
    Array(Box::new(Any)),
    Atomic(Int),
);
test_satisfies!(
    satisfies_missing_does_not_satisfy_atomic,
    Not,
    Missing,
    Atomic(String),
);
test_satisfies!(
    satisfies_any_of_must_satisfy_any,
    Must,
    AnyOf(vec![Atomic(String), Atomic(Int)]),
    Any,
);

test_satisfies!(
    satisfies_one_of_must_satisfy_one_of_when_equal,
    Must,
    OneOf(vec![Atomic(String), Atomic(Int)]),
    OneOf(vec![Atomic(String), Atomic(Int)]),
);
test_satisfies!(
    satisfies_one_of_may_satisfy_one_of_when_unequal,
    May,
    OneOf(vec![Atomic(Double), Atomic(Int)]),
    OneOf(vec![Atomic(String), Atomic(Int)]),
);
test_satisfies!(
    satisfies_one_of_may_satisfy_one_of_when_self_has_missing,
    May,
    OneOf(vec![Atomic(String), Missing]),
    OneOf(vec![Atomic(String), Atomic(Int)]),
);
test_satisfies!(
    satisfies_one_of_must_satisfy_one_of_when_other_has_missing,
    Must,
    OneOf(vec![Atomic(String), Atomic(Int)]),
    OneOf(vec![Atomic(String), Atomic(Int), Missing]),
);
test_satisfies!(
    satisfies_one_of_must_not_satisfy_when_one_of_contains_any,
    Not,
    OneOf(vec![Atomic(String), Atomic(Int)]),
    OneOf(vec![Atomic(String), Atomic(Int), Any]),
);
test_satisfies!(
    satisfies_any_of_must_satisfy_when_any_of_contains_any,
    Must,
    OneOf(vec![Atomic(String), Atomic(Int)]),
    AnyOf(vec![Atomic(String), Atomic(Int), Any]),
);
test_satisfies!(
    satisfies_one_of_must_may_satisfy_one_of_when_self_has_any,
    May,
    OneOf(vec![Atomic(String), Any]),
    OneOf(vec![Atomic(String), Atomic(Int)]),
);
test_satisfies!(
    satisfies_one_of_must_satisfy_subset_one_of,
    Must,
    OneOf(vec![Atomic(String), Atomic(Int)]),
    OneOf(vec![Atomic(String), Atomic(Int), Atomic(Double)]),
);
test_satisfies!(
    satisfies_one_of_may_satisfy_proper_superset_one_of,
    May,
    OneOf(vec![Atomic(String), Atomic(Int), Atomic(Double)]),
    OneOf(vec![Atomic(String), Atomic(Int)]),
);
test_satisfies!(
    satisfies_array_of_string_must_satisfy_one_of_array_of_int_or_array_of_string,
    Must,
    Array(Box::new(Atomic(String))),
    OneOf(vec![
        Array(Box::new(Atomic(String))),
        Array(Box::new(Atomic(Int)))
    ]),
);
test_satisfies!(
    satisfies_array_of_string_or_int_may_satisfy_one_of_array_of_int_or_array_of_string,
    May,
    Array(Box::new(OneOf(vec![Atomic(String), Atomic(Int),]))),
    OneOf(vec![
        Array(Box::new(Atomic(String))),
        Array(Box::new(Atomic(Int)))
    ]),
);
test_satisfies!(
    satisfies_array_of_string_or_int_must_satisfy_array_of_string_or_int,
    Must,
    Array(Box::new(OneOf(vec![Atomic(String), Atomic(Int),]))),
    Array(Box::new(OneOf(vec![Atomic(String), Atomic(Int),]))),
);
test_satisfies!(
    satisfies_document_must_satify_same_document,
    Must,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: true
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: true,
    }),
);
test_satisfies!(
    satisfies_document_may_satify_with_more_permissive_key_schema,
    May,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(String),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
);
test_satisfies!(
    satisfies_document_must_not_satify_with_incompatable_key_schema,
    Not,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(String),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
);
test_satisfies!(
    satisfies_document_may_satify_with_fewer_required_keys,
    May,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set![],
        additional_properties: false,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
);
test_satisfies!(
    satisfies_document_must_not_satify_with_missing_required_key,
    Not,
    Document(Document {
        keys: b_tree_map![
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set![],
        additional_properties: false,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
);
test_satisfies!(
    satisfies_document_may_satify_with_missing_required_key,
    May,
    Document(Document {
        keys: b_tree_map![
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set![],
        additional_properties: true,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: true,
    }),
);
test_satisfies!(
    satisfies_document_must_satify_with_more_required_keys,
    Must,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set![],
        additional_properties: false,
    }),
);
test_satisfies!(
    satisfies_document_may_satify_due_to_possible_extra_keys,
    May,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set![],
        additional_properties: true,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set![],
        additional_properties: false,
    }),
);
test_satisfies!(
    satisfies_document_satifies_multiple_one_of_results_in_not_satisfied,
    Not,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    OneOf(vec![
        Document(Document {
            keys: b_tree_map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set!["a".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: b_tree_map![
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set![],
            additional_properties: true,
        }),
    ]),
);
test_satisfies!(
    satisfies_document_satifies_multiple_any_of_results_in_must_satisfy,
    Must,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    AnyOf(vec![
        Document(Document {
            keys: b_tree_map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set!["a".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: b_tree_map![
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set![],
            additional_properties: false,
        }),
    ]),
);
test_satisfies!(
    satisfies_document_satifies_one_of_one_of_results_must_satisfy,
    Must,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    OneOf(vec![
        Document(Document {
            keys: b_tree_map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set!["a".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: b_tree_map![
                "e".to_string() => Atomic(Int),
            ],
            required: b_tree_set![],
            additional_properties: false,
        }),
    ]),
);
test_satisfies!(
    satisfies_document_may_satisfy_when_key_schema_may_satisfy,
    May,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Atomic(Int),
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
);
test_satisfies!(
    satisfies_array_may_satisfy_when_array_item_schema_may_satisfy,
    May,
    Array(Box::new(Any)),
    Array(Box::new(Atomic(Int))),
);
test_satisfies!(
    satisfies_array_may_satisfy_when_array_item_schema_may_satisfy_multiple_one_of_array,
    May,
    Array(Box::new(Any)),
    OneOf(vec![
        Array(Box::new(Atomic(Int))),
        Array(Box::new(Atomic(String))),
    ]),
);
test_satisfies!(
    satisfies_array_may_satisfy_when_array_item_schema_may_satisfy_multiple_array_one_of,
    May,
    Array(Box::new(Any)),
    Array(Box::new(OneOf(vec![Atomic(Int), Atomic(Double),]),)),
);
test_satisfies!(
    satisfies_array_of_missing_does_not_satisfy_array_of_atomic,
    Not,
    Array(Box::new(Missing)),
    Array(Box::new(Atomic(Int))),
);

macro_rules! test_contains_field {
    ($func_name:ident, $expected:expr, $self:expr, $other:expr $(,)?) => {
        #[test]
        fn $func_name() {
            let res = $self.contains_field($other);
            assert_eq!($expected, res)
        }
    };
}

test_contains_field!(contains_field_any_may_contain_field, May, Any, "a");
test_contains_field!(
    contains_field_missing_does_not_contain_field,
    Not,
    Missing,
    "a"
);
test_contains_field!(
    contains_field_document_must_contain_field,
    Must,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    "a",
);
test_contains_field!(
    contains_field_document_may_contain_field,
    May,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    "b",
);
test_contains_field!(
    contains_field_document_may_contain_field_due_to_additional_properties,
    May,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: true,
    }),
    "foo",
);
test_contains_field!(
    contains_field_document_must_not_contain_field,
    Not,
    Document(Document {
        keys: b_tree_map![
            "a".to_string() => Any,
            "b".to_string() => Atomic(Int),
        ],
        required: b_tree_set!["a".to_string()],
        additional_properties: false,
    }),
    "foo",
);
test_contains_field!(
    contains_field_atomic_must_not_contain_field,
    Not,
    Atomic(String),
    "foo",
);
test_contains_field!(
    contains_field_one_of_document_and_atomic_may_not_contain_field,
    Not,
    OneOf(vec![
        Document(Document {
            keys: b_tree_map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set!["a".to_string()],
            additional_properties: false,
        }),
        Atomic(String),
    ]),
    "c",
);
test_contains_field!(
    contains_field_one_of_document_and_atomic_may_contain_field,
    May,
    OneOf(vec![
        Document(Document {
            keys: b_tree_map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set!["a".to_string()],
            additional_properties: false,
        }),
        Atomic(String),
    ]),
    "b",
);
test_contains_field!(
    contains_field_one_of_document_and_document_must_contain_field,
    Must,
    OneOf(vec![
        Document(Document {
            keys: b_tree_map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(Int),
            ],
            required: b_tree_set!["b".to_string()],
            additional_properties: false,
        }),
        Document(Document {
            keys: b_tree_map![
                "a".to_string() => Any,
                "b".to_string() => Atomic(String),
            ],
            required: b_tree_set!["b".to_string()],
            additional_properties: false,
        }),
    ]),
    "b",
);
