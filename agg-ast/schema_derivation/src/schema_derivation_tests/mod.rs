macro_rules! test_derive_schema {
    // ref_schema and starting_schema are mutually exclusive. ref_schema should be used when only
    // one reference is needed, while starting_schema should be used when the schema needs multiple
    // fields.
    ($func_name:ident, expected = $expected:expr, input = $input:expr$(, starting_schema = $starting_schema:expr)?$(, ref_schema = $ref_schema:expr)?$(, variables = $variables:expr)?) => {
        #[test]
        fn $func_name() {
            let input: Expression = serde_json::from_str($input).unwrap();
            #[allow(unused_mut, unused_assignments)]
            let mut result_set_schema = Schema::Any;
            $(result_set_schema = Schema::Document(Document { keys: map! {"foo".to_string() => $ref_schema }, ..Default::default()});)?
            $(result_set_schema = $starting_schema;)?
            #[allow(unused_mut, unused_assignments)]
            let mut variables = BTreeMap::new();
            $(variables = $variables;)?
            let mut state = ResultSetState {
                catalog: &BTreeMap::new(),
                variables: &variables,
                result_set_schema
            };
            let result = input.derive_schema(&mut state);
            assert_eq!($expected, result);
        }
    };
}

#[cfg(test)]
mod bson;
#[cfg(test)]
mod expression;
#[cfg(test)]
mod match_stage;
#[cfg(test)]
mod stage;
#[cfg(test)]
mod tagged_ops;
#[cfg(test)]
mod untagged_ops;
