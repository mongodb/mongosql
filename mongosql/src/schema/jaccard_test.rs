use lazy_static::lazy_static;

use crate::{
    map,
    schema::{Atomic, Document, JaccardIndex, Schema},
};

lazy_static! {
    static ref A_DOCUMENT: Document = Document {
        keys: map! {
            "a1".into() => Schema::Atomic(Atomic::Integer),
            "a2".into() => Schema::Atomic(Atomic::Integer),
            "a3".into() => Schema::Atomic(Atomic::Integer),
            "a4".into() => Schema::Atomic(Atomic::Integer),
            "a5".into() => Schema::Atomic(Atomic::Integer),
            "a6".into() => Schema::Atomic(Atomic::Integer),
            "a7".into() => Schema::Atomic(Atomic::Integer),
            "a8".into() => Schema::Atomic(Atomic::Integer),
            "a9".into() => Schema::Atomic(Atomic::Integer),
            "a10".into() => Schema::Atomic(Atomic::Integer),
            "a11".into() => Schema::Atomic(Atomic::Integer),
            "a12".into() => Schema::Atomic(Atomic::Integer),
            "a13".into() => Schema::Atomic(Atomic::Integer),
            "a14".into() => Schema::Atomic(Atomic::Integer),
            "a15".into() => Schema::Atomic(Atomic::Integer),
            "a16".into() => Schema::Atomic(Atomic::Integer),
            "a17".into() => Schema::Atomic(Atomic::Integer),
            "a18".into() => Schema::Atomic(Atomic::Integer),
            "a19".into() => Schema::Atomic(Atomic::Integer),
            "a20".into() => Schema::Atomic(Atomic::Integer),
            "a11".into() => Schema::Atomic(Atomic::Integer),
            "a12".into() => Schema::Atomic(Atomic::Integer),
            "a13".into() => Schema::Atomic(Atomic::Integer),
            "a14".into() => Schema::Atomic(Atomic::Integer),
            "a15".into() => Schema::Atomic(Atomic::Integer),
            "a16".into() => Schema::Atomic(Atomic::Integer),
            "a17".into() => Schema::Atomic(Atomic::Integer),
            "a18".into() => Schema::Atomic(Atomic::Integer),
            "a19".into() => Schema::Atomic(Atomic::Integer),
            "a20".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref AB_DOCUMENT: Document = Document {
        keys: map! {
            "a1".into() => Schema::Atomic(Atomic::Integer),
            "a2".into() => Schema::Atomic(Atomic::Integer),
            "a3".into() => Schema::Atomic(Atomic::Integer),
            "a4".into() => Schema::Atomic(Atomic::Integer),
            "a5".into() => Schema::Atomic(Atomic::Integer),
            "a6".into() => Schema::Atomic(Atomic::Integer),
            "a7".into() => Schema::Atomic(Atomic::Integer),
            "a8".into() => Schema::Atomic(Atomic::Integer),
            "a9".into() => Schema::Atomic(Atomic::Integer),
            "a10".into() => Schema::Atomic(Atomic::Integer),
            "b11".into() => Schema::Atomic(Atomic::Integer),
            "b12".into() => Schema::Atomic(Atomic::Integer),
            "b13".into() => Schema::Atomic(Atomic::Integer),
            "b14".into() => Schema::Atomic(Atomic::Integer),
            "b15".into() => Schema::Atomic(Atomic::Integer),
            "b16".into() => Schema::Atomic(Atomic::Integer),
            "b17".into() => Schema::Atomic(Atomic::Integer),
            "b18".into() => Schema::Atomic(Atomic::Integer),
            "b19".into() => Schema::Atomic(Atomic::Integer),
            "b20".into() => Schema::Atomic(Atomic::Integer),
            "b11".into() => Schema::Atomic(Atomic::Integer),
            "b12".into() => Schema::Atomic(Atomic::Integer),
            "b13".into() => Schema::Atomic(Atomic::Integer),
            "b14".into() => Schema::Atomic(Atomic::Integer),
            "b15".into() => Schema::Atomic(Atomic::Integer),
            "b16".into() => Schema::Atomic(Atomic::Integer),
            "b17".into() => Schema::Atomic(Atomic::Integer),
            "b18".into() => Schema::Atomic(Atomic::Integer),
            "b19".into() => Schema::Atomic(Atomic::Integer),
            "b20".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref B_DOCUMENT: Document = Document {
        keys: map! {
            "b1".into() => Schema::Atomic(Atomic::Integer),
            "b2".into() => Schema::Atomic(Atomic::Integer),
            "b3".into() => Schema::Atomic(Atomic::Integer),
            "b4".into() => Schema::Atomic(Atomic::Integer),
            "b5".into() => Schema::Atomic(Atomic::Integer),
            "b6".into() => Schema::Atomic(Atomic::Integer),
            "b7".into() => Schema::Atomic(Atomic::Integer),
            "b8".into() => Schema::Atomic(Atomic::Integer),
            "b9".into() => Schema::Atomic(Atomic::Integer),
            "b10".into() => Schema::Atomic(Atomic::Integer),
            "b11".into() => Schema::Atomic(Atomic::Integer),
            "b12".into() => Schema::Atomic(Atomic::Integer),
            "b13".into() => Schema::Atomic(Atomic::Integer),
            "b14".into() => Schema::Atomic(Atomic::Integer),
            "b15".into() => Schema::Atomic(Atomic::Integer),
            "b16".into() => Schema::Atomic(Atomic::Integer),
            "b17".into() => Schema::Atomic(Atomic::Integer),
            "b18".into() => Schema::Atomic(Atomic::Integer),
            "b19".into() => Schema::Atomic(Atomic::Integer),
            "b20".into() => Schema::Atomic(Atomic::Integer),
            "b11".into() => Schema::Atomic(Atomic::Integer),
            "b12".into() => Schema::Atomic(Atomic::Integer),
            "b13".into() => Schema::Atomic(Atomic::Integer),
            "b14".into() => Schema::Atomic(Atomic::Integer),
            "b15".into() => Schema::Atomic(Atomic::Integer),
            "b16".into() => Schema::Atomic(Atomic::Integer),
            "b17".into() => Schema::Atomic(Atomic::Integer),
            "b18".into() => Schema::Atomic(Atomic::Integer),
            "b19".into() => Schema::Atomic(Atomic::Integer),
            "b20".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref C_DOCUMENT: Document = Document {
        keys: map! {
            "c1".into() => Schema::Atomic(Atomic::Integer),
            "c2".into() => Schema::Atomic(Atomic::Integer),
            "c3".into() => Schema::Atomic(Atomic::Integer),
            "c4".into() => Schema::Atomic(Atomic::Integer),
            "c5".into() => Schema::Atomic(Atomic::Integer),
            "c6".into() => Schema::Atomic(Atomic::Integer),
            "c7".into() => Schema::Atomic(Atomic::Integer),
            "c8".into() => Schema::Atomic(Atomic::Integer),
            "c9".into() => Schema::Atomic(Atomic::Integer),
            "c10".into() => Schema::Atomic(Atomic::Integer),
            "c11".into() => Schema::Atomic(Atomic::Integer),
            "c12".into() => Schema::Atomic(Atomic::Integer),
            "c13".into() => Schema::Atomic(Atomic::Integer),
            "c14".into() => Schema::Atomic(Atomic::Integer),
            "c15".into() => Schema::Atomic(Atomic::Integer),
            "c16".into() => Schema::Atomic(Atomic::Integer),
            "c17".into() => Schema::Atomic(Atomic::Integer),
            "c18".into() => Schema::Atomic(Atomic::Integer),
            "c19".into() => Schema::Atomic(Atomic::Integer),
            "c20".into() => Schema::Atomic(Atomic::Integer),
            "c11".into() => Schema::Atomic(Atomic::Integer),
            "c12".into() => Schema::Atomic(Atomic::Integer),
            "c13".into() => Schema::Atomic(Atomic::Integer),
            "c14".into() => Schema::Atomic(Atomic::Integer),
            "c15".into() => Schema::Atomic(Atomic::Integer),
            "c16".into() => Schema::Atomic(Atomic::Integer),
            "c17".into() => Schema::Atomic(Atomic::Integer),
            "c18".into() => Schema::Atomic(Atomic::Integer),
            "c19".into() => Schema::Atomic(Atomic::Integer),
            "c20".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref D_DOCUMENT: Document = Document {
        keys: map! {
            "d1".into() => Schema::Atomic(Atomic::Integer),
            "d2".into() => Schema::Atomic(Atomic::Integer),
            "d3".into() => Schema::Atomic(Atomic::Integer),
            "d4".into() => Schema::Atomic(Atomic::Integer),
            "d5".into() => Schema::Atomic(Atomic::Integer),
            "d6".into() => Schema::Atomic(Atomic::Integer),
            "d7".into() => Schema::Atomic(Atomic::Integer),
            "d8".into() => Schema::Atomic(Atomic::Integer),
            "d9".into() => Schema::Atomic(Atomic::Integer),
            "d10".into() => Schema::Atomic(Atomic::Integer),
            "d11".into() => Schema::Atomic(Atomic::Integer),
            "d12".into() => Schema::Atomic(Atomic::Integer),
            "d13".into() => Schema::Atomic(Atomic::Integer),
            "d14".into() => Schema::Atomic(Atomic::Integer),
            "d15".into() => Schema::Atomic(Atomic::Integer),
            "d16".into() => Schema::Atomic(Atomic::Integer),
            "d17".into() => Schema::Atomic(Atomic::Integer),
            "d18".into() => Schema::Atomic(Atomic::Integer),
            "d19".into() => Schema::Atomic(Atomic::Integer),
            "d20".into() => Schema::Atomic(Atomic::Integer),
            "d11".into() => Schema::Atomic(Atomic::Integer),
            "d12".into() => Schema::Atomic(Atomic::Integer),
            "d13".into() => Schema::Atomic(Atomic::Integer),
            "d14".into() => Schema::Atomic(Atomic::Integer),
            "d15".into() => Schema::Atomic(Atomic::Integer),
            "d16".into() => Schema::Atomic(Atomic::Integer),
            "d17".into() => Schema::Atomic(Atomic::Integer),
            "d18".into() => Schema::Atomic(Atomic::Integer),
            "d19".into() => Schema::Atomic(Atomic::Integer),
            "d20".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref E_DOCUMENT: Document = Document {
        keys: map! {
            "e1".into() => Schema::Atomic(Atomic::Integer),
            "e2".into() => Schema::Atomic(Atomic::Integer),
            "e3".into() => Schema::Atomic(Atomic::Integer),
            "e4".into() => Schema::Atomic(Atomic::Integer),
            "e5".into() => Schema::Atomic(Atomic::Integer),
            "e6".into() => Schema::Atomic(Atomic::Integer),
            "e7".into() => Schema::Atomic(Atomic::Integer),
            "e8".into() => Schema::Atomic(Atomic::Integer),
            "e9".into() => Schema::Atomic(Atomic::Integer),
            "e10".into() => Schema::Atomic(Atomic::Integer),
            "e11".into() => Schema::Atomic(Atomic::Integer),
            "e12".into() => Schema::Atomic(Atomic::Integer),
            "e13".into() => Schema::Atomic(Atomic::Integer),
            "e14".into() => Schema::Atomic(Atomic::Integer),
            "e15".into() => Schema::Atomic(Atomic::Integer),
            "e16".into() => Schema::Atomic(Atomic::Integer),
            "e17".into() => Schema::Atomic(Atomic::Integer),
            "e18".into() => Schema::Atomic(Atomic::Integer),
            "e19".into() => Schema::Atomic(Atomic::Integer),
            "e20".into() => Schema::Atomic(Atomic::Integer),
            "e11".into() => Schema::Atomic(Atomic::Integer),
            "e12".into() => Schema::Atomic(Atomic::Integer),
            "e13".into() => Schema::Atomic(Atomic::Integer),
            "e14".into() => Schema::Atomic(Atomic::Integer),
            "e15".into() => Schema::Atomic(Atomic::Integer),
            "e16".into() => Schema::Atomic(Atomic::Integer),
            "e17".into() => Schema::Atomic(Atomic::Integer),
            "e18".into() => Schema::Atomic(Atomic::Integer),
            "e19".into() => Schema::Atomic(Atomic::Integer),
            "e20".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
    static ref F_DOCUMENT: Document = Document {
        keys: map! {
            "f1".into() => Schema::Atomic(Atomic::Integer),
            "f2".into() => Schema::Atomic(Atomic::Integer),
            "f3".into() => Schema::Atomic(Atomic::Integer),
            "f4".into() => Schema::Atomic(Atomic::Integer),
            "f5".into() => Schema::Atomic(Atomic::Integer),
            "f6".into() => Schema::Atomic(Atomic::Integer),
            "f7".into() => Schema::Atomic(Atomic::Integer),
            "f8".into() => Schema::Atomic(Atomic::Integer),
            "f9".into() => Schema::Atomic(Atomic::Integer),
            "f10".into() => Schema::Atomic(Atomic::Integer),
            "f11".into() => Schema::Atomic(Atomic::Integer),
            "f12".into() => Schema::Atomic(Atomic::Integer),
            "f13".into() => Schema::Atomic(Atomic::Integer),
            "f14".into() => Schema::Atomic(Atomic::Integer),
            "f15".into() => Schema::Atomic(Atomic::Integer),
            "f16".into() => Schema::Atomic(Atomic::Integer),
            "f17".into() => Schema::Atomic(Atomic::Integer),
            "f18".into() => Schema::Atomic(Atomic::Integer),
            "f19".into() => Schema::Atomic(Atomic::Integer),
            "f20".into() => Schema::Atomic(Atomic::Integer),
            "f11".into() => Schema::Atomic(Atomic::Integer),
            "f12".into() => Schema::Atomic(Atomic::Integer),
            "f13".into() => Schema::Atomic(Atomic::Integer),
            "f14".into() => Schema::Atomic(Atomic::Integer),
            "f15".into() => Schema::Atomic(Atomic::Integer),
            "f16".into() => Schema::Atomic(Atomic::Integer),
            "f17".into() => Schema::Atomic(Atomic::Integer),
            "f18".into() => Schema::Atomic(Atomic::Integer),
            "f19".into() => Schema::Atomic(Atomic::Integer),
            "f20".into() => Schema::Atomic(Atomic::Integer),
        },
        jaccard_index: JaccardIndex::default().into(),
        ..Default::default()
    };
}

mod jaccard {
    use crate::schema::JaccardIndex;

    use super::*;

    #[test]
    // https://en.wikipedia.org/wiki/Jaccard_index
    fn how_it_works() {
        let left = Document {
            keys: map! {
                "a".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::new(0.8).into(),
            ..Default::default()
        };
        let right = Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex {
                avg_ji: 0.5,
                num_unions: 1,
                ..Default::default()
            }
            .into(),
            ..Default::default()
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
    }

    #[test]
    fn union_of_empty_documents_is_safe() {
        let doc = Document {
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let new_doc = doc.clone().union(doc.clone());
        dbg!(new_doc.jaccard_index);
        assert_eq!(new_doc, doc);
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
        assert_eq!(new_right.jaccard_index.unwrap().avg_ji, 1.0);
    }

    #[test]
    fn four_unions_with_zero_jaccard_index_preserves_document() {
        let new_left = A_DOCUMENT.clone().union(B_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
        let new_left = new_left.union(C_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
        let new_left = new_left.union(D_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
        let new_left = new_left.union(E_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
    }

    #[test]
    fn five_unions_with_zero_jaccard_index_is_any_document() {
        let new_left = A_DOCUMENT.clone().union(B_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
        let new_left = new_left.union(C_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
        let new_left = new_left.union(D_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
        let new_left = new_left.union(E_DOCUMENT.clone());
        assert_ne!(new_left, Document::any());
        let new_left = new_left.union(F_DOCUMENT.clone());
        assert_eq!(new_left, Document::any());
    }

    #[test]
    fn breaking_rate_threshold_results_in_any_document() {
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
        let b_doc = Document {
            keys: map! {
                "b".into() => Schema::Atomic(Atomic::Integer),
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let c_doc = Document {
            keys: map! {
                "c".into() => Schema::Atomic(Atomic::Integer),
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let d_doc = Document {
            keys: map! {
                "d".into() => Schema::Atomic(Atomic::Integer),
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
                "g".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let e_doc = Document {
            keys: map! {
                "e".into() => Schema::Atomic(Atomic::Integer),
                "f".into() => Schema::Atomic(Atomic::Integer),
                "g".into() => Schema::Atomic(Atomic::Integer),
                "h".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };
        let f_doc = Document {
            keys: map! {
                "f".into() => Schema::Atomic(Atomic::Integer),
                "g".into() => Schema::Atomic(Atomic::Integer),
                "h".into() => Schema::Atomic(Atomic::Integer),
                "i".into() => Schema::Atomic(Atomic::Integer),
            },
            jaccard_index: JaccardIndex::default().into(),
            ..Default::default()
        };

        let new_left = a_doc
            .union(b_doc)
            .union(c_doc)
            .union(d_doc)
            .union(e_doc)
            .union(f_doc);
        assert_eq!(new_left, Document::any());
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
    fn some_insability_is_tolerated() {
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
            .union(A_DOCUMENT.clone())
            .union(B_DOCUMENT.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone());

        assert_ne!(new_left, Document::any());
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
            .union(A_DOCUMENT.clone())
            .union(B_DOCUMENT.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(a_doc.clone())
            .union(C_DOCUMENT.clone())
            .union(D_DOCUMENT.clone());

        assert_eq!(new_left, Document::any());
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

        assert_eq!(new_left, a_doc);
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

        assert_eq!(new_left, a_doc);
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

        assert_eq!(
            new_left,
            Document {
                keys: map! {
                    "a".into() => Schema::Document(Document {
                        keys: map! {
                            "b".into() => Schema::Document(Document::any()),
                        },
                        jaccard_index: JaccardIndex::default().into(),
                        ..Default::default()
                    }),
                },
                jaccard_index: JaccardIndex::default().into(),
                ..Default::default()
            }
        );
    }
}
