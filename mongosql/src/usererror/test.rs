use super::{UserError, UserErrorDisplay};

#[derive(Debug, PartialEq, UserErrorDisplay)]
enum Test {
    Foo,
    Bar,
    Baz,
}

impl UserError for Test {
    fn code(&self) -> u32 {
        match self {
            Test::Foo => 1,
            Test::Bar => 2,
            Test::Baz => 3,
        }
    }

    fn user_message(&self) -> Option<String> {
        match self {
            Test::Foo => Some(format!("{self:?}")),
            Test::Bar => None,
            Test::Baz => Some(format!("{self:?}")),
        }
    }

    fn technical_message(&self) -> String {
        match self {
            Test::Foo => "technically a foo".to_string(),
            Test::Bar => "technically a bar".to_string(),
            Test::Baz => "technically a baz".to_string(),
        }
    }
}

mod usererror {
    use super::{Test, UserError};
    #[test]
    fn user_error_codes() {
        assert_eq!(1, Test::Foo.code());
        assert_eq!(2, Test::Bar.code());
        assert_eq!(3, Test::Baz.code());
    }

    #[test]
    fn user_error_user_message() {
        assert_eq!("Foo", Test::Foo.user_message().unwrap());
        assert_eq!(None, Test::Bar.user_message());
        assert_eq!("Baz", Test::Baz.user_message().unwrap());
    }

    #[test]
    fn user_error_technical_message() {
        assert_eq!("technically a foo", Test::Foo.technical_message());
        assert_eq!("technically a bar", Test::Bar.technical_message());
        assert_eq!("technically a baz", Test::Baz.technical_message());
    }
}

mod usererrordisplay {
    use super::Test;

    #[test]
    fn usererrordisplay() {
        let expected_foo = "Error 1: Foo\n\tCaused by:\n\ttechnically a foo";
        let expected_bar = "Error 2: technically a bar";
        let expected_baz = "Error 3: Baz\n\tCaused by:\n\ttechnically a baz";

        assert_eq!(Test::Foo.to_string(), expected_foo);
        assert_eq!(Test::Bar.to_string(), expected_bar);
        assert_eq!(Test::Baz.to_string(), expected_baz);
    }
}
