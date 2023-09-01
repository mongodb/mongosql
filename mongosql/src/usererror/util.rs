use crate::schema::Schema;
use edit_distance::edit_distance;

type Result<T> = std::result::Result<T, String>;

pub fn generate_suggestion(input: &str, expected: &[String]) -> Result<Vec<String>> {
    let mut closest = expected
        .iter()
        .filter_map(|b| {
            let distance = edit_distance(input, b);
            if distance == 0 {
                return Some(Err(format!(
            "Unexpected edit distance of 0 found with input: {input} and expected: {expected:?}"
        )));
            }

            if distance <= 2 {
                Some(Ok((distance, b)))
            } else {
                None
            }
        })
        .collect::<Result<Vec<_>>>()?;

    closest.sort_by(|(a, _), (b, _)| a.cmp(b));

    Ok(closest.into_iter().map(|(_, s)| s.clone()).collect())
}

pub fn unsat_check(schema_vec: Vec<Schema>) -> Option<String> {
    let unsat_found = schema_vec.iter().any(|schema| schema.eq(&Schema::Unsat));

    if unsat_found {
        Some("Internal error: Unsat schema found. Report this to MongoDB".to_string())
    } else {
        None
    }
}

#[cfg(test)]
mod unsat_check_tests {
    use super::*;
    use crate::schema::Atomic;

    #[test]
    fn unsat_found() {
        assert_eq!(
            Some("Internal error: Unsat schema found. Report this to MongoDB".to_string()),
            unsat_check(vec![
                Schema::Unsat,
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Boolean)
            ])
        )
    }

    #[test]
    fn unsat_not_found() {
        assert_eq!(
            None,
            unsat_check(vec![Schema::Atomic(Atomic::Null), Schema::Missing])
        )
    }
}

#[cfg(test)]
mod suggestion {
    use super::generate_suggestion;
    #[test]
    fn suggestions_in_order() {
        assert_eq!(
            vec!["baz", "biz"],
            generate_suggestion(
                "bar",
                &[
                    "bizzle".to_string(), // edit distance 5 filtered out
                    "bazzle".to_string(), // edit distance 4 filtered out
                    "biz".to_string(),    // edit distance 2
                    "baz".to_string(),    // edit distance 1
                ]
            )
            .unwrap()
        );
    }

    #[test]
    fn edit_distance_0_is_err() {
        assert_eq!(
        "Unexpected edit distance of 0 found with input: bar and expected: [\"bizzle\", \"bazzle\", \"biz\", \"baz\", \"bar\"]",
            generate_suggestion(
                "bar",
                &[
                    "bizzle".to_string(),
                    "bazzle".to_string(),
                    "biz".to_string(),
                    "baz".to_string(),
                    "bar".to_string(),
                ]
            )
            .err()
            .unwrap()
        )
    }

    #[test]
    fn empty_input_no_matches() {
        assert_eq!(
            Vec::<String>::new(),
            generate_suggestion(
                "",
                &[
                    "bizzle".to_string(),
                    "bazzle".to_string(),
                    "biz".to_string(),
                    "baz".to_string(),
                ]
            )
            .unwrap()
        );
    }

    #[test]
    fn empty_expected_no_matches() {
        assert_eq!(
            Vec::<String>::new(),
            generate_suggestion("needle", &[]).unwrap()
        );
    }
}
