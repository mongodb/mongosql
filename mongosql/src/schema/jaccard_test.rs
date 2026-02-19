use lazy_static::lazy_static;

use crate::{
    map,
    schema::{Atomic, Document, JaccardIndex, Schema},
};

lazy_static! {
    static ref A_DOCUMENT: Document = Document {
        keys: map! {
            "a".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref B_DOCUMENT: Document = Document {
        keys: map! {
            "b".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref C_DOCUMENT: Document = Document {
        keys: map! {
            "c".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref D_DOCUMENT: Document = Document {
        keys: map! {
            "d".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref E_DOCUMENT: Document = Document {
        keys: map! {
            "e".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref F_DOCUMENT: Document = Document {
        keys: map! {
            "f".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref G_DOCUMENT: Document = Document {
        keys: map! {
            "g".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref H_DOCUMENT: Document = Document {
        keys: map! {
            "h".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
}

mod jaccard {
    use crate::{schema::JaccardIndex, set};

    use super::*;

    #[test]
    // https://en.wikipedia.org/wiki/Jaccard_index
    fn how_it_works() {
        let left = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 1.0,
                num_unions: 0,
                stability_limit: 0.8,
            }),
            unstable: false,
        };
        let right = Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.5,
                num_unions: 1,
                stability_limit: 0.8,
            }),
            unstable: false,
        };

        let new_left = left.union(right);
        let jaccard_index = new_left.jaccard_index.unwrap();

        // 1 existing union, plus 1 from the union operation
        assert_eq!(jaccard_index.num_unions, 2);

        // jaccard_index.avg_ji = (1.0 * 0 + 0.5 * 1) / (1 + 0) = 0.5
        // new_jaccard_index = a ∩ b / a ∪ b = 0 / 2 = 0
        // num_unions = 2 (see previous assertion)
        // let new_avg_ji = (ji.avg_ji * ji.num_unions + new_ji.avg_ji) / (ji.num_unions + 1)
        // new_avg_ji = (0.5 * 1 + 0) / (1 + 1) = 0.25
        assert_eq!(jaccard_index.avg_ji, 0.25);
    }

    #[test]
    fn document_default_does_not_have_jaccard_index() {
        let doc = Document::default();
        assert!(doc.jaccard_index.is_none());
        assert!(!doc.unstable);
    }

    #[test]
    fn union_of_empty_documents_is_safe() {
        let doc = Document {
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let new_doc = doc.clone().union(doc.clone());
        assert_eq!(new_doc.jaccard_index, doc.jaccard_index);
        assert!(!new_doc.unstable);
    }

    #[test]
    fn subsets_and_supersets_are_considered_identical() {
        let left = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let right = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };

        let new_left = left.clone().union(right.clone());
        let new_right = right.clone().union(left.clone());

        assert_eq!(new_left.jaccard_index.unwrap().avg_ji, 1.0);
        assert!(!new_left.unstable);
        assert_eq!(new_right.jaccard_index.unwrap().avg_ji, 1.0);
        assert!(!new_right.unstable);
    }

    #[test]
    fn four_unions_with_zero_jaccard_index_preserves_document() {
        let new_left = A_DOCUMENT.clone().union(B_DOCUMENT.clone());
        assert!(!new_left.unstable);
        let new_left = new_left.union(C_DOCUMENT.clone());
        assert!(!new_left.unstable);
        let new_left = new_left.union(D_DOCUMENT.clone());
        assert!(!new_left.unstable);
        let new_left = new_left.union(E_DOCUMENT.clone());
        assert!(new_left.eq_with_jaccard_index(&Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.0,
                num_unions: 4,
                stability_limit: 0.8,
            }),
            unstable: false,
        }));
    }

    #[test]
    fn five_unions_with_zero_jaccard_index_is_unstable() {
        let new_left = A_DOCUMENT.clone().union(B_DOCUMENT.clone());
        assert!(!new_left.unstable);
        let new_left = new_left.union(C_DOCUMENT.clone());
        assert!(!new_left.unstable);
        let new_left = new_left.union(D_DOCUMENT.clone());
        assert!(!new_left.unstable);
        let new_left = new_left.union(E_DOCUMENT.clone());
        assert!(!new_left.unstable);
        let new_left = new_left.union(F_DOCUMENT.clone());
        assert!(new_left.eq_with_jaccard_index(&Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.0,
                num_unions: 5,
                stability_limit: 0.8,
            }),
            unstable: true,
        }))
    }

    #[test]
    fn union_stable_and_unstable_schemas_returns_the_stable_schema_marked_as_unstable() {
        let stable_doc = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer)
            },
            required: set! {"a".into()},
            additional_properties: false,
            // In this example, we assume we've unioned 10 documents that all had the same schema.
            jaccard_index: Some(JaccardIndex {
                avg_ji: 1.0,
                num_unions: 10,
                stability_limit: 0.8,
            }),
            unstable: false,
        };

        let unstable_doc = Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
                "g".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            // In this example, we assume we've unioned 5 documents that all had unique schema.
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.0,
                num_unions: 5,
                stability_limit: 0.8,
            }),
            unstable: true,
        };

        let expected_doc = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into()},
            additional_properties: true,
            jaccard_index: Some(JaccardIndex {
                // ((weighted avg_ji for stable_doc and unstable_doc) * total # unions) / (total # unions + 1)
                avg_ji: (10f64 * 15f64) / 16f64,
                // 10 unions from stable_doc + 5 unions from unstable_doc + 1 union of them together
                num_unions: 16,
                stability_limit: 0.8,
            }),
            unstable: true,
        };

        // When we union a stable document and an unstable document, we want to retain the stable
        // schema but also mark it as unstable and set additional_properties to true.
        let actual_doc = stable_doc.clone().union(unstable_doc.clone());
        assert!(
            actual_doc.eq_with_jaccard_index(&expected_doc),
            "Incorrect document union result"
        );

        // Order should not impact this behavior
        let actual_doc = unstable_doc.union(stable_doc);
        assert!(
            actual_doc.eq_with_jaccard_index(&expected_doc),
            "Incorrect document union result"
        );
    }

    #[test]
    fn union_between_unstable_docs_prefers_schema_with_larger_avg_ji() {
        // Recall that the updated avg_ji is computed with the formula:
        //   new_avg_ji = (prev_avg_ji * num_unions + intersection_size / union_size) / (num_unions + 1)
        // Assume this document schema is the result of unioning the following 6 documents:
        // 1. { a: 1 }
        // 2. { a: 1, b: 1 }
        // 3. { a: 1, c: 1 }
        // 4. { a: 1, d: 1 }
        // 5. { a: 1, e: 1 }
        // 6. { a: 1, f: 1 }
        // That results in the Jaccard Index updating as follows as each document is unioned:
        // 1. num_unions = 0, avg_ji = 1
        // 2. num_unions = 1, avg_ji = 1 / 2
        // 3. num_unions = 2, avg_ji = (1/2 * 1 + 1/3) / 2 = 5 / 12
        // 4. num_unions = 3, avg_ji = (5/12 * 2 + 1/4) / 3 = 13 / 36
        // 5. num_unions = 4, avg_ji = (13/36 * 3 + 1/5) / 4 = 77 / 240
        // 6. num_unions = 5, avg_ji = (77/240 * 4 + 1/6) / 5 = 29 / 100
        let unstable_doc_1 = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into()},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.29,
                num_unions: 5,
                stability_limit: 0.8,
            }),
            unstable: true,
        };

        // Assume this document schema is the result of unioning the following 6 documents:
        // 1. { a: 1 }
        // 2. { a: 1, b: 1 }
        // 3. { a: 1, c: 1 }
        // 4. { a: 1, d: 1 }
        // 5. { a: 1, e: 1 }
        // 6. { f: 1 }
        // That results in the Jaccard Index updating as follows as each document is unioned:
        // 1. num_unions = 0, avg_ji = 1
        // 2. num_unions = 1, avg_ji = 1 / 2
        // 3. num_unions = 2, avg_ji = (1/2 * 1 + 1/3) / 2 = 5 / 12
        // 4. num_unions = 3, avg_ji = (5/12 * 2 + 1/4) / 3 = 13 / 36
        // 5. num_unions = 4, avg_ji = (13/36 * 3 + 1/5) / 4 = 77 / 240
        // 6. num_unions = 5, avg_ji = (77/240 * 4 + 0/6) / 5 = 77 / 300
        let unstable_doc_2 = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.256, // estimated value, the true value is .256666...
                num_unions: 5,
                stability_limit: 0.8,
            }),
            unstable: true,
        };

        // When we union an unstable document and an unstable document, we want to retain schema
        // with the higher avg_ji value and mark it with additional_properties true.
        let new_doc = unstable_doc_1.clone().union(unstable_doc_2.clone());
        assert!(new_doc.eq_with_jaccard_index(&Document {
            additional_properties: true,
            ..unstable_doc_1.clone()
        }));

        // Order should not impact this behavior
        let new_doc = unstable_doc_2.union(unstable_doc_1.clone());
        assert!(new_doc.eq_with_jaccard_index(&Document {
            additional_properties: true,
            ..unstable_doc_1
        }));
    }

    #[test]
    fn union_between_unstable_docs_with_equal_avg_ji_prefers_left() {
        let unstable_doc_1 = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"a".into()},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.29,
                num_unions: 5,
                stability_limit: 0.8,
            }),
            unstable: true,
        };

        let unstable_doc_2 = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"b".into()},
            additional_properties: false,
            jaccard_index: Some(JaccardIndex {
                avg_ji: 0.29,
                num_unions: 5,
                stability_limit: 0.8,
            }),
            unstable: true,
        };

        let new_doc = unstable_doc_1.clone().union(unstable_doc_2.clone());
        assert!(new_doc.eq_with_jaccard_index(&Document {
            additional_properties: true,
            ..unstable_doc_1.clone()
        }));

        let new_doc = unstable_doc_2.clone().union(unstable_doc_1);
        assert!(new_doc.eq_with_jaccard_index(&Document {
            additional_properties: true,
            ..unstable_doc_2
        }))
    }

    #[test]
    fn stable_docs() {
        let a_doc = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };

        let new_left = a_doc
            .clone()
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone());
        assert_eq!(new_left, a_doc);
        let jaccard_index = new_left.jaccard_index.unwrap();
        assert_eq!(jaccard_index.avg_ji, 1.0, "Incorrect avg_change_rate");
        assert_eq!(jaccard_index.num_unions, 5, "Incorrect num_unions");
        assert_eq!(
            jaccard_index.stability_limit, 0.8,
            "Incorrect instability_limit"
        );
    }

    #[test]
    fn some_instability_is_tolerated() {
        let a_doc = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let new_left = a_doc
            .clone()
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(E_DOCUMENT.clone())
            .union(F_DOCUMENT.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone());

        assert!(!new_left.unstable);
    }

    #[test]
    fn continued_instability_is_not_tolerated() {
        let a_doc = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let new_left = a_doc
            .clone()
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(E_DOCUMENT.clone())
            .union(F_DOCUMENT.clone())
            .union(G_DOCUMENT.clone())
            .union(H_DOCUMENT.clone());

        assert!(new_left.unstable);
    }

    #[test]
    fn nested_stable_documents() {
        let a_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(A_DOCUMENT.clone()),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let b_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(A_DOCUMENT.clone()),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let c_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(A_DOCUMENT.clone()),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let d_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(A_DOCUMENT.clone()),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let e_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(A_DOCUMENT.clone()),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let f_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(A_DOCUMENT.clone()),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };

        let new_left = a_doc
            .clone()
            .union(b_doc)
            .union(c_doc)
            .union(d_doc)
            .union(e_doc)
            .union(f_doc);

        assert!(new_left.eq_with_jaccard_index(&a_doc));
    }

    #[test]
    fn nested_nested_stable_documents() {
        let a_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(A_DOCUMENT.clone()),
                    },
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let b_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(A_DOCUMENT.clone()),
                    },
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let c_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(A_DOCUMENT.clone()),
                    },
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let d_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(A_DOCUMENT.clone()),
                    },
                ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let e_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(A_DOCUMENT.clone()),
                    },
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let f_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(A_DOCUMENT.clone()),
                    },
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };

        let new_left = a_doc
            .clone()
            .union(b_doc)
            .union(c_doc)
            .union(d_doc)
            .union(e_doc)
            .union(f_doc);

        assert!(new_left.eq_with_jaccard_index(&a_doc));
    }

    #[test]
    fn nested_nested_unstable_documents() {
        let a_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(A_DOCUMENT.clone()),
                    },
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let b_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(B_DOCUMENT.clone()),
                    },
                    jaccard_index: JaccardIndex::default().into(),
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let c_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(C_DOCUMENT.clone()),
                    },
                    jaccard_index: JaccardIndex::default().into(),
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let d_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(D_DOCUMENT.clone()),
                    },
                    jaccard_index: JaccardIndex::default().into(),
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let e_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(E_DOCUMENT.clone()),
                    },
                    jaccard_index: JaccardIndex::default().into(),
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let f_doc = Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(F_DOCUMENT.clone()),
                    },
                    jaccard_index: JaccardIndex::default().into(),
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };

        let new_left = a_doc
            .clone()
            .union(b_doc)
            .union(c_doc)
            .union(d_doc)
            .union(e_doc)
            .union(f_doc);

        assert!(new_left.eq_with_jaccard_index(&Document {
            keys: map! {
                "a".into() => Schema::Document(Document {
                    keys: map! {
                        "b".into() => Schema::Document(Document {
                            keys: map! {
                                "a".into() => Schema::Atomic(Atomic::Integer),
                                "b".into() => Schema::Atomic(Atomic::Integer),
                                "c".into() => Schema::Atomic(Atomic::Integer),
                                "d".into() => Schema::Atomic(Atomic::Integer),
                                "e".into() => Schema::Atomic(Atomic::Integer),
                                "f".into() => Schema::Atomic(Atomic::Integer),
                            },
                            required: set!{},
                            additional_properties: false,
                            jaccard_index: Some(JaccardIndex {
                                avg_ji: 0.0,
                                num_unions: 5,
                                stability_limit: 0.8,
                            }),
                            unstable: true,
                        }),
                    },
                    jaccard_index: JaccardIndex::default().into(),
                    ..Default::default()
                }),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        }));
    }
}
