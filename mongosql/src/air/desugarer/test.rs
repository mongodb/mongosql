use crate::air::{self, agg_ast::ast_definitions as agg_ast, desugarer::Pass};
use serde::Deserialize;
use std::{fs, io::Read};
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

                assert_eq!(expected_air_pipeline, actual)
            }

            Ok(())
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

// TODO: SQL-1318 uncomment when agg_ast::Stage::Group is implemented
// mod accumulators {
//     use super::*;
//     use crate::air::desugarer::accumulators::AccumulatorsDesugarerPass;
//
//     test_desugarer!(
//         file = "desugar_accumulators.yml",
//         desugarer = AccumulatorsDesugarerPass
//     );
// }

// TODO: SQL-1318 uncomment when agg_ast::Stage::Join and agg_ast::Unwind is implemented
// mod joins {
//     use super::*;
//     use crate::air::desugarer::join::JoinDesugarerPass;
//
//     test_desugarer!(
//         file = "desugar_joins.yml",
//         desugarer = JoinDesugarerPass
//     );
// }

mod lookups {
    use super::*;
    use crate::air::desugarer::lookup::LookupDesugarerPass;

    test_desugarer!(
        file = "desugar_lookups.yml",
        desugarer = LookupDesugarerPass
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
        .fold(root, |acc, curr| (acc, curr).into())
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
                "_id".to_string() => air::Expression::Literal(air::LiteralValue::Integer(0)),
            }
        }),
        input = vec![agg_ast::Stage::Project(map! {
            "_id".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(0))
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
                        "a".to_string() => air::Expression::Literal(air::LiteralValue::Boolean(true)),
                        "b".to_string() => air::Expression::Literal(air::LiteralValue::Integer(2)),
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
                "a".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Boolean(true)),
                "b".to_string() => agg_ast::Expression::Literal(agg_ast::LiteralValue::Integer(2)),
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
