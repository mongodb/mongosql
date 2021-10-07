use crate::ir::{definitions::Collection, visitor::Visitor, Project, Stage};
use serde::{Deserialize, Serialize};

#[derive(Default)]
struct NamespaceVisitor {
    namespaces: Vec<Namespace>,
}

#[derive(Serialize, Deserialize, PartialEq, Clone, Debug)]
pub struct Namespace {
    pub database: String,
    pub collection: String,
}

pub fn get_namespaces(st: Stage) -> Vec<Namespace> {
    let mut v = NamespaceVisitor::default();
    v.visit_stage(st);
    v.namespaces
}

impl Visitor for NamespaceVisitor {
    fn visit_collection(&mut self, node: Collection) -> Collection {
        self.namespaces.push(Namespace {
            database: node.db.clone(),
            collection: node.collection.clone(),
        });
        node.walk(self)
    }

    fn visit_project(&mut self, node: Project) -> Project {
        let node = node.walk(self);
        node.expression.iter().for_each(|(_, expr)| {
            expr.clone().walk(self);
        });
        node
    }
}

#[cfg(test)]
mod test_get_namespaces {
    use crate::{ir::definitions::*, map};

    macro_rules! test_get_namespaces {
        ($func_name:ident, $expected:expr, $input:expr) => {
            #[test]
            fn $func_name() {
                use crate::ir::namespace::*;
                let input = $input;
                let expected = $expected;
                let actual = get_namespaces(input);
                assert_eq!(actual, expected);
            }
        };
    }

    test_get_namespaces!(
        no_collections,
        vec![],
        Stage::Join(Join {
            join_type: JoinType::Inner,
            left: Box::new(Stage::Array(Array {
                array: vec![Expression::Literal(Literal::Null)],
                alias: "foo".into()
            })),
            right: Box::new(Stage::Array(Array {
                array: vec![Expression::Literal(Literal::Null)],
                alias: "bar".into()
            })),
            condition: None,
        })
    );

    test_get_namespaces!(
        top_level_collection,
        vec![Namespace {
            database: "test".into(),
            collection: "foo".into()
        }],
        Stage::Collection(Collection {
            db: "test".into(),
            collection: "foo".into()
        })
    );

    test_get_namespaces!(
        stage_nested_collections,
        vec![
            Namespace {
                database: "test".into(),
                collection: "foo".into()
            },
            Namespace {
                database: "db2".into(),
                collection: "bar".into()
            },
            Namespace {
                database: "db2".into(),
                collection: "baz".into()
            }
        ],
        Stage::Set(Set {
            operation: SetOperation::UnionAll,
            left: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into()
            })),
            right: Box::new(Stage::Join(Join {
                join_type: JoinType::Inner,
                left: Box::new(Stage::Collection(Collection {
                    db: "db2".into(),
                    collection: "bar".into()
                })),
                right: Box::new(Stage::Collection(Collection {
                    db: "db2".into(),
                    collection: "baz".into()
                })),
                condition: None,
            }))
        })
    );

    test_get_namespaces!(
        expr_nested_collections,
        vec![
            Namespace {
                database: "test".into(),
                collection: "foo".into()
            },
            Namespace {
                database: "test".into(),
                collection: "bar".into()
            },
            Namespace {
                database: "test".into(),
                collection: "baz".into()
            },
            Namespace {
                database: "test".into(),
                collection: "xyz".into()
            }
        ],
        Stage::Project(Project {
            source: Box::new(Stage::Collection(Collection {
                db: "test".into(),
                collection: "foo".into()
            })),
            expression: map! {
                ("bar1", 0u16).into() => Expression::Subquery(SubqueryExpr {
                    output_expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                    subquery: Box::new(Stage::Collection(Collection {
                        db: "test".into(),
                        collection: "bar".into()
                    }))
                }),
                ("bar2", 0u16).into() => Expression::SubqueryComparison(SubqueryComparison {
                    operator: SubqueryComparisonOp::Lt,
                    modifier: SubqueryModifier::Any,
                    argument: Box::new(Expression::Reference(("foo", 0u16).into())),
                    subquery_expr: SubqueryExpr {
                        output_expr: Box::new(Expression::Reference(("foo", 0u16).into())),
                        subquery: Box::new(Stage::Collection(Collection {
                            db: "test".into(),
                            collection: "baz".into()
                        }))
                    }
                }),
                ("bar3", 0u16).into() => Expression::Exists(Box::new(Stage::Collection(Collection {
                    db: "test".into(),
                    collection: "xyz".into()
                }))),
            }
        })
    );
}
