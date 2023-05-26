use crate::air::{
    self,
    agg_ast::ast_definitions as agg_ast,
    desugarer::{self, Pass},
    visitor::Visitor,
    Stage,
};
use chrono::prelude::*;
use itertools::Itertools;
use lazy_static::lazy_static;
use serde::Deserialize;
use std::{fs, io::Read, str::FromStr};
use thiserror::Error;

macro_rules! test_desugarer {
    (file = $file:expr, desugarer = $desugarer:ident) => {
        #[test]
        fn test() -> Result<(), Error> {
            let file_path = format!("src/air/desugarer/testdata/{}", $file);

            let test_file = parse_test_yaml(file_path.as_str())?;

            for test in test_file.tests {
                if test.skip_reason.is_some() {
                    continue;
                }

                let input_air_pipeline = to_air_pipeline(test.input);
                let expected_air_pipeline = to_air_pipeline(test.expected);

                let actual = $desugarer
                    .apply(input_air_pipeline)
                    .map_err(Error::CannotDesugar)?;

                assert_eq!(expected_air_pipeline, actual, "{}", test.name)
            }

            Ok(())
        }
    };
}

macro_rules! test_desugar_manual {
    (name = $name:ident, desugarer = $desugarer:ident, input = $input:expr, expected = $expected:expr) => {
        #[test]
        fn $name() {
            let input = $input;
            let expected = $expected;

            let actual: Result<Stage, Error> =
                $desugarer.apply(input).map_err(Error::CannotDesugar);

            assert_eq!(expected, actual);
        }
    };
}

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("failed to read file '{0}': {1}")]
    InvalidFile(String, String),
    #[error("failed to read file '{0}' to string: {1}")]
    CannotReadFileToString(String, String),
    #[error("failed to deserialize YAML file '{0}': {1}")]
    CannotDeserializeYaml(String, String),
    #[error("failed to desugar: {0:?}")]
    CannotDesugar(crate::air::desugarer::Error),
}

#[derive(Debug, PartialEq, Deserialize)]
struct TestFile {
    tests: Vec<TestCase>,
}

#[derive(Debug, PartialEq, Deserialize)]
struct TestCase {
    name: String,
    skip_reason: Option<String>,
    input: Vec<agg_ast::Stage>,
    expected: Vec<agg_ast::Stage>,
}

mod accumulators {
    use super::*;
    use crate::air::desugarer::accumulators::AccumulatorsDesugarerPass;

    test_desugarer!(
        file = "desugar_accumulators.yml",
        desugarer = AccumulatorsDesugarerPass
    );
}

mod joins {
    use super::*;
    use crate::air::desugarer::join::JoinDesugarerPass;

    test_desugarer!(file = "desugar_joins.yml", desugarer = JoinDesugarerPass);
}

mod root_references {
    use super::*;
    use crate::air::desugarer::root_references::RootReferenceDesugarerPass;

    test_desugarer!(
        file = "desugar_root_references.yml",
        desugarer = RootReferenceDesugarerPass
    );
}

mod sql_match_null_semantics {
    use super::*;
    use crate::air::desugarer::match_null_semantics::MatchDesugarerPass;

    test_desugarer!(
        file = "desugar_sql_match_null_semantics.yml",
        desugarer = MatchDesugarerPass
    );
}

mod sql_null_semantics {
    use super::*;
    use crate::air::desugarer::sql_null_semantics_operators::SQLNullSemanticsOperatorsDesugarerPass;

    test_desugarer!(
        file = "desugar_sql_null_semantics.yml",
        desugarer = SQLNullSemanticsOperatorsDesugarerPass
    );
}

mod fold_converts {
    use super::*;
    use crate::{
        air::desugarer::fold_converts::FoldConvertsDesugarerPass, unchecked_unique_linked_hash_map,
    };

    lazy_static! {
        static ref OID: bson::oid::ObjectId =
            bson::oid::ObjectId::parse_str("507f1f77bcf86cd799439011").unwrap();
        static ref DATE: bson::DateTime = {
            let chrono_dt: chrono::DateTime<Utc> = "2014-01-01T08:15:39.736Z".parse().unwrap();
            chrono_dt.into()
        };
        static ref INT_DEC: bson::Decimal128 = bson::Decimal128::from_str("3").unwrap();
        static ref DOUBLE_DEC: bson::Decimal128 =
            bson::Decimal128::from_str(&format!("{}", std::f64::consts::PI)).unwrap();
        static ref STRING_DEC: bson::Decimal128 = bson::Decimal128::from_str("3.14159265").unwrap();
    }

    test_desugar_manual!(
        name = no_convert,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::ObjectId(*OID)))}
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::ObjectId(*OID)))}
        }))
    );
    test_desugar_manual!(
        name = convert_valid_string_to_date,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String("2014-01-01T08:15:39.736Z".into())).into(),
                    to: air::Type::Datetime,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::DateTime(*DATE)))}
        }))
    );
    test_desugar_manual!(
        name = convert_invalid_string_to_date,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String("2014-01-56T08:15:39.736Z".into())).into(),
                    to: air::Type::Datetime,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Err::<Stage, Error>(Error::CannotDesugar(
            desugarer::Error::InvalidConstantConvert(air::Type::Datetime)
        ))
    );
    test_desugar_manual!(
        name = convert_valid_string_to_oid,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String("507f1f77bcf86cd799439011".into())).into(),
                    to: air::Type::ObjectId,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::ObjectId(*OID)))}
        }))
    );
    test_desugar_manual!(
        name = convert_invalid_string_to_oid,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String("57f1f77bcf86cd799439011".into())).into(),
                    to: air::Type::ObjectId,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Err::<Stage, Error>(Error::CannotDesugar(
            desugarer::Error::InvalidConstantConvert(air::Type::ObjectId)
        ))
    );
    test_desugar_manual!(
        name = convert_valid_string_to_decimal128,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String("3.14159265".into())).into(),
                    to: air::Type::Decimal128,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Decimal128(*STRING_DEC)))}
        }))
    );
    test_desugar_manual!(
        name = convert_valid_nested_string_convert_to_decimal128,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Convert(air::Convert {
                        input: air::Expression::Literal(air::LiteralValue::String("3.14159265".into())).into(),
                        to: air::Type::String,
                        on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                        on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    }).into(),
                    to: air::Type::Decimal128,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                    }))
            }
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Decimal128(*STRING_DEC)))}
        }))
    );
    test_desugar_manual!(
        name = convert_invalid_string_to_decimal128,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::String("hello".into())).into(),
                    to: air::Type::Decimal128,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Err::<Stage, Error>(Error::CannotDesugar(
            desugarer::Error::InvalidConstantConvert(air::Type::Decimal128)
        ))
    );
    test_desugar_manual!(
        name = convert_double_to_decimal128,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::Double(std::f64::consts::PI)).into(),
                    to: air::Type::Decimal128,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Decimal128(*DOUBLE_DEC)))}
        }))
    );
    test_desugar_manual!(
        name = convert_int_to_decimal128,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::Integer(3)).into(),
                    to: air::Type::Decimal128,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Decimal128(*INT_DEC)))}
        }))
    );
    test_desugar_manual!(
        name = convert_long_to_decimal128,
        desugarer = FoldConvertsDesugarerPass,
        input = air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(
                air::Expression::Convert(air::Convert {
                    input: air::Expression::Literal(air::LiteralValue::Long(3)).into(),
                    to: air::Type::Decimal128,
                    on_null: air::Expression::Literal(air::LiteralValue::Null).into(),
                    on_error: air::Expression::Literal(air::LiteralValue::Null).into(),
                })
            )}
        }),
        expected = Ok::<Stage, desugarer::test::Error>(air::Stage::Project(air::Project {
            source: air::Stage::Collection(air::Collection {
                db: "test".into(),
                collection: "default".into()
            })
            .into(),
            specifications: unchecked_unique_linked_hash_map! {"expr".into() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Decimal128(*INT_DEC)))}
        }))
    );
}

mod subquery_expressions {
    use super::*;
    use crate::air::desugarer::subquery::SubqueryExprDesugarerPass;

    test_desugarer!(
        file = "desugar_subquery_expressions.yml",
        desugarer = SubqueryExprDesugarerPass
    );
}

mod unsupported_operators {
    use super::*;
    use crate::air::desugarer::unsupported_operators::UnsupportedOperatorsDesugarerPass;

    test_desugarer!(
        file = "desugar_unsupported_operators.yml",
        desugarer = UnsupportedOperatorsDesugarerPass
    );
}

mod all_desugarer_passes {
    use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;

    use super::*;
    use crate::air::{desugarer::desugar_pipeline, Project};

    struct ProjectKeySortVisitor;
    impl Visitor for ProjectKeySortVisitor {
        fn visit_project(&mut self, node: Project) -> Project {
            let mut node = node.walk(self);
            let mut new_specs = UniqueLinkedHashMap::new();
            for (k, v) in node
                .specifications
                .into_iter()
                // sort alphabetically for testing purposes
                .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
            {
                new_specs.insert(k, v).unwrap();
            }
            node.specifications = new_specs;
            node
        }
    }

    #[test]
    fn test() -> Result<(), Error> {
        let file_path = "src/air/desugarer/testdata/desugar_all.yml";

        let test_file = parse_test_yaml(file_path)?;

        for test in test_file.tests {
            if test.skip_reason.is_some() {
                continue;
            }

            let input_air_pipeline = to_air_pipeline(test.input);
            let expected_air_pipeline = to_air_pipeline(test.expected);

            let mut sorter = ProjectKeySortVisitor;
            let actual = sorter
                .visit_stage(desugar_pipeline(input_air_pipeline).map_err(Error::CannotDesugar)?);

            assert_eq!(expected_air_pipeline, actual, "{}", test.name)
        }

        Ok(())
    }
}

fn parse_test_yaml(path: &str) -> Result<TestFile, Error> {
    let mut f =
        fs::File::open(path).map_err(|e| Error::InvalidFile(path.to_string(), format!("{e:?}")))?;
    let mut contents = String::new();
    f.read_to_string(&mut contents)
        .map_err(|e| Error::CannotReadFileToString(path.to_string(), format!("{e:?}")))?;
    let yaml: TestFile = serde_yaml::from_str(&contents)
        .map_err(|e| Error::CannotDeserializeYaml(path.to_string(), format!("{e:?}")))?;
    Ok(yaml)
}

fn to_air_pipeline(pipeline: Vec<agg_ast::Stage>) -> air::Stage {
    let root = air::Stage::Collection(air::Collection {
        db: "test".to_string(),
        collection: "default".to_string(),
    });

    pipeline
        .into_iter()
        .fold(root, |acc, curr| (Some(acc), curr).into())
}

#[cfg(test)]
mod to_air_pipeline_test {
    use super::to_air_pipeline;
    use crate::{
        air::{self, agg_ast::ast_definitions as agg_ast},
        map, unchecked_unique_linked_hash_map,
    };

    macro_rules! test_to_air_pipeline {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let input = $input;
                let expected = $expected;

                let actual = to_air_pipeline(input);

                assert_eq!(expected, actual)
            }
        };
    }

    test_to_air_pipeline!(
        empty,
        expected = air::Stage::Collection(air::Collection {
            db: "test".to_string(),
            collection: "default".to_string()
        }),
        input = vec![]
    );

    test_to_air_pipeline!(
        singleton,
        expected = air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Collection(air::Collection {
                db: "test".to_string(),
                collection: "default".to_string()
            })),
            specifications: unchecked_unique_linked_hash_map! {
                "x".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(42))),
            }
        }),
        input = vec![agg_ast::Stage::Project(map! {
            "x".to_string() => agg_ast::ProjectItem::Assignment(agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(42))),
        }),]
    );

    test_to_air_pipeline!(
        multiple_stages,
        expected = air::Stage::Sort(air::Sort {
            source: Box::new(air::Stage::Match(air::Match {
                source: Box::new(air::Stage::Project(air::Project {
                    source: Box::new(air::Stage::Collection(air::Collection {
                        db: "test".to_string(),
                        collection: "default".to_string(),
                    })),
                    specifications: unchecked_unique_linked_hash_map! {
                        "a".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Boolean(true))),
                        "b".to_string() => air::ProjectItem::Assignment(air::Expression::Literal(air::LiteralValue::Integer(2))),
                    },
                })),
                expr: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            })),
            specs: vec![
                air::SortSpecification::Asc("a".to_string()),
                air::SortSpecification::Desc("b".to_string()),
            ]
        }),
        input = vec![
            agg_ast::Stage::Project(map! {
                "a".to_string() => agg_ast::ProjectItem::Assignment(agg_ast::Expression::Literal(agg_ast::LiteralValue::Boolean(true))),
                "b".to_string() => agg_ast::ProjectItem::Assignment(agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(2))),
            }),
            agg_ast::Stage::Match(agg_ast::MatchExpression::NonExpr(
                agg_ast::Expression::Literal(agg_ast::LiteralValue::Null),
            )),
            agg_ast::Stage::Sort(map! {
                "a".to_string() => 1,
                "b".to_string() => -1,
            }),
        ]
    );
}
