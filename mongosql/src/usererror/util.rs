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
