use crate::{
    module::submodule::{ast, visitor::Visitor},
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};

use linked_hash_map::LinkedHashMap;

struct AtomVisitor {
    atom_names: Vec<String>,
}

impl Visitor for AtomVisitor {
    fn visit_atom(&mut self, node: ast::Atom) -> ast::Atom {
        self.atom_names.push(node.name.clone());
        node
    }
}

#[test]
fn simple_atom_visitor_test() {
    use ast::*;

    let mut v = AtomVisitor { atom_names: vec![] };
    v.visit_atom(Atom {
        name: "hello".to_string(),
    });
    assert_eq!(vec!["hello".to_string()], v.atom_names);
}

#[test]
fn tree_atom_visitor_test() {
    use ast::*;
    use std::collections::BTreeMap;

    let mut v = AtomVisitor { atom_names: vec![] };

    let l = Box::new(Expression::Plus(Plus {
        left: Box::new(Expression::Atoms(vec![
            Atom {
                name: "a1".to_string(),
            },
            Atom {
                name: "a2".to_string(),
            },
        ])),
        right: Box::new(Expression::Null),
    }));

    let r = Box::new(Expression::Tree(Tree {
        branch_b1: Box::new("b1".to_string()),
        branch_b2: Box::new(Expression::Atom(Atom {
            name: "a3".to_string(),
        })),
        branch_b3: Box::new(Some(Expression::Atom(Atom {
            name: "a4".to_string(),
        }))),
        branch_b4: Box::new(vec![
            Expression::Atom(Atom {
                name: "a5".to_string(),
            }),
            Expression::Atom(Atom {
                name: "a6".to_string(),
            }),
        ]),
        branch_b5: Box::new({
            let mut m = BTreeMap::new();
            m.insert(
                Atom {
                    name: "a7".to_string(),
                },
                Expression::Literal("l1".to_string()),
            );
            m.insert(
                Atom {
                    name: "a8".to_string(),
                },
                Expression::Atom(Atom {
                    name: "a9".to_string(),
                }),
            );
            m
        }),
        branch_b6: Box::new({
            let mut m = BTreeMap::new();
            m.insert(
                Box::new(Atom {
                    name: "a10".to_string(),
                }),
                Box::new(Expression::Literal("l2".to_string())),
            );
            m
        }),
        branch_o1: Some("o1".to_string()),
        branch_o2: Some(Box::new(Expression::Atom(Atom {
            name: "a11".to_string(),
        }))),
        branch_o3: Some(vec![
            Expression::Literal("l3".to_string()),
            Expression::Atom(Atom {
                name: "a12".to_string(),
            }),
        ]),
        branch_o4: Some({
            let mut m = BTreeMap::new();
            m.insert(
                Atom {
                    name: "a13".to_string(),
                },
                Expression::Literal("l4".to_string()),
            );
            m
        }),
        branch_o5: Some({
            let mut m = BTreeMap::new();
            m.insert(
                Box::new(Atom {
                    name: "a14".to_string(),
                }),
                Box::new(Expression::Literal("l5".to_string())),
            );
            m
        }),

        branch_v1: vec!["v1".to_string()],
        branch_v2: vec![
            Box::new(Expression::Literal("l6".to_string())),
            Box::new(Expression::Atom(Atom {
                name: "a15".to_string(),
            })),
        ],
        branch_v3: vec![
            vec![
                Expression::Literal("l7".to_string()),
                Expression::Atom(Atom {
                    name: "a16".to_string(),
                }),
            ],
            vec![Expression::Atom(Atom {
                name: "a17".to_string(),
            })],
        ],
        branch_v4: vec![
            {
                let mut m = BTreeMap::new();
                m.insert(
                    Atom {
                        name: "a18".to_string(),
                    },
                    Expression::Literal("l8".to_string()),
                );
                m
            },
            {
                let mut m = BTreeMap::new();
                m.insert(
                    Atom {
                        name: "a19".to_string(),
                    },
                    Expression::Literal("l9".to_string()),
                );
                m
            },
        ],
        branch_v5: vec![
            {
                let mut m = BTreeMap::new();
                m.insert(
                    Box::new(Atom {
                        name: "a20".to_string(),
                    }),
                    Box::new(Expression::Literal("l10".to_string())),
                );
                m
            },
            {
                let mut m = BTreeMap::new();
                m.insert(
                    Box::new(Atom {
                        name: "a21".to_string(),
                    }),
                    Box::new(Expression::Literal("l11".to_string())),
                );
                m
            },
        ],

        branch_m1: {
            let mut m = BTreeMap::new();
            m.insert(
                Box::new(vec![
                    Atom {
                        name: "a22".to_string(),
                    },
                    Atom {
                        name: "a23".to_string(),
                    },
                ]),
                Box::new(vec![
                    Expression::Atom(Atom {
                        name: "a24".to_string(),
                    }),
                    Expression::Literal("l12".to_string()),
                ]),
            );
            m
        },
    }));

    let e = Expression::Plus(Plus { left: l, right: r });

    v.visit_expression(e);

    assert_eq!(
        vec![
            "a1".to_string(),
            "a2".to_string(),
            "a3".to_string(),
            "a4".to_string(),
            "a5".to_string(),
            "a6".to_string(),
            "a7".to_string(),
            "a8".to_string(),
            "a9".to_string(),
            "a10".to_string(),
            "a11".to_string(),
            "a12".to_string(),
            "a13".to_string(),
            "a14".to_string(),
            "a15".to_string(),
            "a16".to_string(),
            "a17".to_string(),
            "a18".to_string(),
            "a19".to_string(),
            "a20".to_string(),
            "a21".to_string(),
            "a22".to_string(),
            "a23".to_string(),
            "a24".to_string(),
        ],
        v.atom_names,
    );
}

#[test]
fn hash_tree_atom_visitor_test() {
    use ast::*;
    use std::collections::HashMap;

    let mut v = AtomVisitor { atom_names: vec![] };

    let e = HashTree {
        branch_m1: {
            let mut m = HashMap::new();
            m.insert("hello1".to_string(), "world1".to_string());
            m
        },
        branch_m2: {
            let mut m = HashMap::new();
            m.insert(
                Box::new(Atom {
                    name: "hello2".to_string(),
                }),
                "world2".to_string(),
            );
            m
        },
        branch_m3: {
            let mut m = HashMap::new();
            m.insert(
                "hello3".to_string(),
                Box::new(Atom {
                    name: "world3".to_string(),
                }),
            );
            m
        },
        branch_m4: {
            let mut m = HashMap::new();
            m.insert(
                Box::new(Atom {
                    name: "hello4".to_string(),
                }),
                Box::new(Atom {
                    name: "world4".to_string(),
                }),
            );
            m
        },

        branch_l1: {
            let mut m = LinkedHashMap::new();
            m.insert("linked_hello1".to_string(), "linked_world1".to_string());
            m
        },
        branch_l2: {
            let mut m = LinkedHashMap::new();
            m.insert(
                Box::new(Atom {
                    name: "linked_hello2".to_string(),
                }),
                "linked_world2".to_string(),
            );
            m
        },
        branch_l3: {
            let mut m = LinkedHashMap::new();
            m.insert(
                "linked_hello3".to_string(),
                Box::new(Atom {
                    name: "linked_world3".to_string(),
                }),
            );
            m
        },
        branch_l4: {
            let mut m = LinkedHashMap::new();
            m.insert(
                Box::new(Atom {
                    name: "linked_hello4".to_string(),
                }),
                Box::new(Atom {
                    name: "linked_world4".to_string(),
                }),
            );
            m
        },

        branch_ul1: {
            let mut m = UniqueLinkedHashMap::new();
            m.insert(
                "unique_linked_hello1".to_string(),
                "unique_linked_world1".to_string(),
            )
            .unwrap();
            m
        },
        branch_ul2: {
            let mut m = UniqueLinkedHashMap::new();
            m.insert(
                Box::new(Atom {
                    name: "unique_linked_hello2".to_string(),
                }),
                "unique_linked_world2".to_string(),
            )
            .unwrap();
            m
        },
        branch_ul3: {
            let mut m = UniqueLinkedHashMap::new();
            m.insert(
                "unique_linked_hello3".to_string(),
                Box::new(Atom {
                    name: "unique_linked_world3".to_string(),
                }),
            )
            .unwrap();
            m
        },
        branch_ul4: {
            let mut m = UniqueLinkedHashMap::new();
            m.insert(
                Box::new(Atom {
                    name: "unique_linked_hello4".to_string(),
                }),
                Box::new(Atom {
                    name: "unique_linked_world4".to_string(),
                }),
            )
            .unwrap();
            m
        },
    };

    v.visit_hash_tree(e);

    assert_eq!(
        vec![
            "hello2".to_string(),
            "world3".to_string(),
            "hello4".to_string(),
            "world4".to_string(),
            "linked_hello2".to_string(),
            "linked_world3".to_string(),
            "linked_hello4".to_string(),
            "linked_world4".to_string(),
            "unique_linked_hello2".to_string(),
            "unique_linked_world3".to_string(),
            "unique_linked_hello4".to_string(),
            "unique_linked_world4".to_string(),
        ],
        v.atom_names
    );
}
