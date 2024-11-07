macro_rules! test_derive_schema {
    ($func_name:ident, expected = $expected:expr, input = $input:expr$(, ref_schema = $ref_schema:expr)?$(, variables = $variables:expr)?) => {
        #[test]
        fn $func_name() {
            let input: Expression = serde_json::from_str($input).unwrap();
            #[allow(unused_mut, unused_assignments)]
            let mut result_set_schema = Schema::Any;
            $(result_set_schema = Schema::Document(Document { keys: map! {"foo".to_string() => $ref_schema }, ..Default::default()});)?
            #[allow(unused_mut, unused_assignments)]
            let mut variables = BTreeMap::new();
            $(variables = $variables;)?
            let mut state = ResultSetState {
                catalog: &BTreeMap::new(),
                variables: &variables,
                result_set_schema
            };
            let result = input.derive_schema(&mut state);
            assert_eq!(result, $expected);
        }
    };
}

#[cfg(test)]
mod expression;
#[cfg(test)]
mod stage;
#[cfg(test)]
mod tagged_ops;
#[cfg(test)]
mod untagged_ops;
