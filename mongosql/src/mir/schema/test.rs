macro_rules! test_user_error_messages {
    ($func_name:ident, input = $input:expr, expected = $expected:expr) => {
        #[test]
        fn $func_name() {
            use crate::{mir::schema::Error, usererror::UserError};

            let user_message = $input.user_message();

            assert_eq!($expected, user_message.unwrap())
        }
    };
}

mod access_missing_field {
    test_user_error_messages! {
        access_missing_field_no_keys,
        input = Error::AccessMissingField("foo".to_string(), None),
        expected = "Cannot access field `foo` because it could not be found."
    }

    test_user_error_messages! {
        access_missing_field_no_close_keys,
        input = Error::AccessMissingField("foo".to_string(), Some(vec!["bar".to_string(), "baz".to_string()])),
        expected = "Cannot access field `foo` because it could not be found."
    }

    test_user_error_messages! {
        access_missing_field_some_close_keys,
        input = Error::AccessMissingField("foo".to_string(), Some(vec!["bar".to_string(), "baz".to_string(), "foof".to_string(), "fo".to_string()])),
        expected = "Cannot access field `foo` because it could not be found. Did you mean: foof, fo"
    }

    test_user_error_messages! {
        access_missing_field_all_close_keys,
        input = Error::AccessMissingField("foo".to_string(), Some(vec!["foo".to_string(), "foof".to_string(), "fo".to_string()])),
        expected = "Cannot access field `foo` because it could not be found. Internal error: Unexpected edit distance of 0 found with input: foo and expected: [\"foo\", \"foof\", \"fo\"]"
    }
}
