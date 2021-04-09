macro_rules! convert_to_snake_case_test {
    ($func_name:ident, $expected:expr, $input:expr) => {
        #[test]
        fn $func_name() {
            use crate::util::convert_to_snake_case;

            let res = convert_to_snake_case($input);
            assert_eq!($expected, res);
        }
    };
}

convert_to_snake_case_test!(single_word_snake_case_test, "hello", "Hello");
convert_to_snake_case_test!(double_word_snake_case_test, "hello_world", "HelloWorld");
convert_to_snake_case_test!(
    triple_word_snake_case_test,
    "hello_world_foo",
    "HelloWorldFoo"
);
convert_to_snake_case_test!(lowercase_snake_case_test, "helloworldfoo", "helloworldfoo");
convert_to_snake_case_test!(
    contains_underscore_snake_case_test,
    "hello__world",
    "Hello_World"
);
convert_to_snake_case_test!(
    starts_with_and_contains_underscore_snake_case_test,
    "__hello___world",
    "_Hello__World"
);
